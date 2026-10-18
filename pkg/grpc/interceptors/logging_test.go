package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
)

func TestLoggingInterceptor(t *testing.T) {
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("успешный вызов", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("вызов с ошибкой", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		require.Error(t, err)
	})

	t.Run("вызов с request_id в context", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		ctx := ctxkeys.WithRequestID(context.Background(), "req-123")
		resp, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("вызов с пустым request_id — не логируется", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		ctx := ctxkeys.WithRequestID(context.Background(), "")
		resp, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("вызов с не-gRPC ошибкой", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			return nil, errors.New("plain error")
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		require.Error(t, err)
	})
}
