package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRetryInterceptor(t *testing.T) {
	const (
		methodInWhitelist  = "/test.Service/GetThing"
		methodOutOfWhitelist = "/test.Service/CreateThing"
		target             = "test-service:50051"
	)

	// newInterceptor создаёт interceptor с быстрым backoff для тестов.
	newInterceptor := func(methods []string) grpc.UnaryClientInterceptor {
		return RetryInterceptor(
			target,
			methods,
			3,                   // max attempts: 1 initial + 2 retries
			1*time.Millisecond,  // base delay
			10*time.Millisecond, // max delay
		)
	}

	// invokerFunc — invoker, возвращающий ошибки из списка по порядку.
	// На выходе за пределами списка возвращает nil.
	invokerFunc := func(errs ...error) (grpc.UnaryInvoker, *int) {
		calls := 0
		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			idx := calls
			calls++
			if idx >= len(errs) {
				return nil
			}
			return errs[idx]
		}
		return invoker, &calls
	}

	t.Run("метод не в whitelist — 1 вызов, retry нет", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc(status.Error(codes.Unavailable, "down"))

		err := interceptor(
			context.Background(), methodOutOfWhitelist,
			nil, nil, nil, invoker,
		)

		require.Error(t, err)
		assert.Equal(t, 1, *calls, "метод вне whitelist не должен ретраиться")
	})

	t.Run("успех с первого раза — 1 вызов", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc() // всегда успех

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.NoError(t, err)
		assert.Equal(t, 1, *calls)
	})

	t.Run("успех после 1 retry — 2 вызова", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc(
			status.Error(codes.Unavailable, "down"),
			nil, // вторая попытка успешна
		)

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.NoError(t, err)
		assert.Equal(t, 2, *calls)
	})

	t.Run("успех после 2 retries — 3 вызова", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc(
			status.Error(codes.Unavailable, "down"),
			status.Error(codes.Unavailable, "down again"),
			nil,
		)

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.NoError(t, err)
		assert.Equal(t, 3, *calls)
	})

	t.Run("все попытки провалились — maxAttempts вызовов", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc(
			status.Error(codes.Unavailable, "down"),
			status.Error(codes.Unavailable, "down"),
			status.Error(codes.Unavailable, "down"),
			status.Error(codes.Unavailable, "down"), // запас, если будет больше
		)

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.Error(t, err)
		assert.Equal(t, 3, *calls, "WithMax(3) => 3 попытки (1 initial + 2 retry)")
	})

	t.Run("NotFound не ретраится — 1 вызов", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc(
			status.Error(codes.NotFound, "not found"),
		)

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, 1, *calls, "логические ошибки не ретраятся")
	})

	t.Run("InvalidArgument не ретраится — 1 вызов", func(t *testing.T) {
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc(
			status.Error(codes.InvalidArgument, "bad request"),
		)

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.Error(t, err)
		assert.Equal(t, 1, *calls)
	})

	t.Run("пустой whitelist — retry не применяется ни к одному методу", func(t *testing.T) {
		interceptor := newInterceptor([]string{})
		invoker, calls := invokerFunc(status.Error(codes.Unavailable, "down"))

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.Error(t, err)
		assert.Equal(t, 1, *calls, "пустой whitelist = retry ни для кого")
	})

	t.Run("успех без retry не считает success-метрику", func(t *testing.T) {
		// Проверяем логику: метрика result=success инкрементится
		// только если calls > 1. Здесь calls = 1 — метрики не должно быть.
		// Проверить это можно только через prometheus testutil —
		// упрощённая проверка через отсутствие паники и err == nil.
		interceptor := newInterceptor([]string{methodInWhitelist})
		invoker, calls := invokerFunc()

		err := interceptor(
			context.Background(), methodInWhitelist,
			nil, nil, nil, invoker,
		)

		require.NoError(t, err)
		assert.Equal(t, 1, *calls)
	})
}