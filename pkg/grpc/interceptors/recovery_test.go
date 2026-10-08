package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
)

func TestRecoveryInterceptor(t *testing.T) {
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("паника превращается в codes.Internal", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			panic("boom")
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		require.Error(t, err, "после паники клиент должен получить ошибку, а не nil")

		st, ok := status.FromError(err)
		require.True(t, ok, "ошибка должна быть gRPC-статусом")
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("успешный handler — ошибки нет", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger.NewTestLogger())

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("обычная ошибка handler пробрасывается как есть", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger.NewTestLogger())

		sentinel := status.Error(codes.NotFound, "not found")
		handler := func(_ context.Context, _ any) (any, error) {
			return nil, sentinel
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)
	})
}
