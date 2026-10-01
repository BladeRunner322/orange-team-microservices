package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestTimeoutInterceptor(t *testing.T) {
	method := "/test.Service/Method"

	t.Run("устанавливает deadline при положительном timeout", func(t *testing.T) {
		interceptor := TimeoutInterceptor(2 * time.Second)

		var captured context.Context
		invoker := func(
			ctx context.Context,
			_ string,
			_, _ any,
			_ *grpc.ClientConn,
			_ ...grpc.CallOption,
		) error {
			captured = ctx
			return nil
		}

		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		require.NoError(t, err)

		deadline, ok := captured.Deadline()
		require.True(t, ok, "deadline должен быть установлен")
		assert.WithinDuration(t, time.Now().Add(2*time.Second), deadline, time.Second)
	})

	t.Run("не устанавливает deadline при timeout=0", func(t *testing.T) {
		interceptor := TimeoutInterceptor(0)

		var captured context.Context
		invoker := func(
			ctx context.Context,
			_ string,
			_, _ any,
			_ *grpc.ClientConn,
			_ ...grpc.CallOption,
		) error {
			captured = ctx
			return nil
		}

		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		require.NoError(t, err)

		_, ok := captured.Deadline()
		assert.False(t, ok, "deadline не должен быть установлен")
	})

	t.Run("не устанавливает deadline при отрицательном timeout", func(t *testing.T) {
		interceptor := TimeoutInterceptor(-1 * time.Second)

		var captured context.Context
		invoker := func(
			ctx context.Context,
			_ string,
			_, _ any,
			_ *grpc.ClientConn,
			_ ...grpc.CallOption,
		) error {
			captured = ctx
			return nil
		}

		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		require.NoError(t, err)

		_, ok := captured.Deadline()
		assert.False(t, ok)
	})

	t.Run("invoker вызван ровно один раз", func(t *testing.T) {
		interceptor := TimeoutInterceptor(time.Second)

		calls := 0
		invoker := func(
			_ context.Context,
			_ string,
			_, _ any,
			_ *grpc.ClientConn,
			_ ...grpc.CallOption,
		) error {
			calls++
			return nil
		}

		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		require.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("пробрасывает ошибку от invoker", func(t *testing.T) {
		interceptor := TimeoutInterceptor(time.Second)
		sentinel := assert.AnError

		invoker := func(
			_ context.Context,
			_ string,
			_, _ any,
			_ *grpc.ClientConn,
			_ ...grpc.CallOption,
		) error {
			return sentinel
		}

		err := interceptor(context.Background(), method, nil, nil, nil, invoker)
		assert.ErrorIs(t, err, sentinel)
	})
}
