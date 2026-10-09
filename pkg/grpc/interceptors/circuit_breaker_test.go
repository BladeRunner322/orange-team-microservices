package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCircuitBreakerInterceptor(t *testing.T) {
	const method = "/test.Service/Method"

	// newTestCB создаёт CB с агрессивными порогами для быстрого открытия.
	newTestCB := func() *gobreaker.CircuitBreaker[any] {
		return NewCircuitBreaker(
			"test-target",
			1,                    // MaxRequests в half-open
			10*time.Second,       // Interval (не важен для теста)
			100*time.Millisecond, // Timeout в open
			2,                    // MinRequests
			0.5,                  // ErrorRate 50%
		)
	}

	// errInvoker возвращает invoker, который всегда падает с данной ошибкой.
	errInvoker := func(err error) grpc.UnaryInvoker {
		return func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return err
		}
	}

	// okInvoker возвращает invoker, который всегда успешен.
	okInvoker := func() grpc.UnaryInvoker {
		return func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return nil
		}
	}

	// codeOf извлекает gRPC-код из ошибки.
	codeOf := func(t *testing.T, err error) codes.Code {
		t.Helper()
		st, ok := status.FromError(err)
		require.True(t, ok, "ошибка должна быть gRPC-статусом")
		return st.Code()
	}

	t.Run("closed: успешный вызов проходит", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		err := interceptor(context.Background(), method, nil, nil, nil, okInvoker())
		assert.NoError(t, err)
	})

	t.Run("closed: ошибка пробрасывается как есть", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		// NotFound — бизнес-ошибка. CB пропустит её как есть,
		// без трансформации (не мапится в ErrOpenState).
		sentinel := status.Error(codes.NotFound, "not found")
		err := interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))

		require.Error(t, err)
		assert.Equal(t, codes.NotFound, codeOf(t, err))
		assert.Equal(t, "not found", status.Convert(err).Message())
	})

	t.Run("открывается после превышения порога ошибок", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		sentinel := status.Error(codes.Unavailable, "down")

		// MinRequests=2, ErrorRate=0.5. 2 ошибки → 100% > 50% → open.
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))

		// Breaker открыт. Следующий вызов — fail-fast, invoker не вызывается.
		invoked := false
		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			invoked = true
			return nil
		}

		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		require.Error(t, err)
		assert.False(t, invoked, "invoker не должен вызываться в OPEN-состоянии")

		assert.Equal(t, codes.Unavailable, codeOf(t, err))
		assert.Equal(t, "circuit breaker is open", status.Convert(err).Message())
	})

	t.Run("half-open: успешный запрос закрывает breaker", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		sentinel := status.Error(codes.Unavailable, "down")

		// Открываем.
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))

		// Ждём Timeout (100ms), чтобы перейти в half-open.
		time.Sleep(150 * time.Millisecond)

		// Успешный запрос в half-open — breaker вернётся в closed.
		err := interceptor(context.Background(), method, nil, nil, nil, okInvoker())
		require.NoError(t, err)

		// Проверяем: следующий запрос снова проходит (closed).
		err = interceptor(context.Background(), method, nil, nil, nil, okInvoker())
		assert.NoError(t, err)
	})

	t.Run("half-open: ошибка снова открывает breaker", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		sentinel := status.Error(codes.Unavailable, "down")

		// Открываем.
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))

		// Ждём Timeout.
		time.Sleep(150 * time.Millisecond)

		// В half-open снова ошибка — breaker вернётся в open.
		_ = interceptor(context.Background(), method, nil, nil, nil, errInvoker(sentinel))

		// Следующий вызов — снова fail-fast.
		invoked := false
		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			invoked = true
			return nil
		}
		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		require.Error(t, err)
		assert.False(t, invoked)
		assert.Equal(t, codes.Unavailable, codeOf(t, err))
	})

	t.Run("бизнес-ошибки не открывают breaker", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		// 5 раз NotFound — бизнес-ошибка. Breaker должен остаться closed,
		// несмотря на то, что порог (2) давно превышен.
		for i := 0; i < 5; i++ {
			err := interceptor(context.Background(), method, nil, nil, nil,
				errInvoker(status.Error(codes.NotFound, "not found")))
			require.Error(t, err)
		}

		// Breaker закрыт — успешный вызов проходит.
		err := interceptor(context.Background(), method, nil, nil, nil, okInvoker())
		assert.NoError(t, err)
	})

	t.Run("транспортные ошибки открывают breaker", func(t *testing.T) {
		cb := newTestCB()
		interceptor := CircuitBreakerInterceptor(cb)

		// Unavailable — транспортная ошибка.
		for i := 0; i < 2; i++ {
			_ = interceptor(context.Background(), method, nil, nil, nil,
				errInvoker(status.Error(codes.Unavailable, "down")))
		}

		// Breaker открыт — следующий вызов fail-fast.
		err := interceptor(context.Background(), method, nil, nil, nil, okInvoker())
		require.Error(t, err)
		assert.Equal(t, codes.Unavailable, codeOf(t, err))
	})
}

func TestStateToFloat(t *testing.T) {
	assert.Equal(t, 0.0, stateToFloat(gobreaker.StateClosed))
	assert.Equal(t, 1.0, stateToFloat(gobreaker.StateHalfOpen))
	assert.Equal(t, 2.0, stateToFloat(gobreaker.StateOpen))
}

func TestStateToString(t *testing.T) {
	assert.Equal(t, "closed", stateToString(gobreaker.StateClosed))
	assert.Equal(t, "half-open", stateToString(gobreaker.StateHalfOpen))
	assert.Equal(t, "open", stateToString(gobreaker.StateOpen))
}