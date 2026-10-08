package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
)

func TestRequestIDServerInterceptor(t *testing.T) {
	interceptor := RequestIDServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("нет metadata — context без изменений", func(t *testing.T) {
		var captured context.Context
		_, err := interceptor(
			context.Background(),
			nil,
			info,
			echoHandler(&captured),
		)
		require.NoError(t, err)

		_, ok := ctxkeys.RequestIDFromContext(captured)
		assert.False(t, ok)
	})

	t.Run("request_id из metadata", func(t *testing.T) {
		md := metadata.Pairs(ctxkeys.MetadataRequestID, "req-123")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		var captured context.Context
		_, err := interceptor(ctx, nil, info, echoHandler(&captured))
		require.NoError(t, err)

		requestID, ok := ctxkeys.RequestIDFromContext(captured)
		assert.True(t, ok)
		assert.Equal(t, "req-123", requestID)
	})

	t.Run("несколько значений в metadata — берётся первое", func(t *testing.T) {
		md := metadata.Pairs(
			ctxkeys.MetadataRequestID, "first",
			ctxkeys.MetadataRequestID, "second",
		)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		var captured context.Context
		_, err := interceptor(ctx, nil, info, echoHandler(&captured))
		require.NoError(t, err)

		requestID, ok := ctxkeys.RequestIDFromContext(captured)
		assert.True(t, ok)
		assert.Equal(t, "first", requestID)
	})
}