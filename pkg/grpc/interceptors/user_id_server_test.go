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

// echoHandler сохраняет context наружу через замыкание.
func echoHandler(captured *context.Context) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		*captured = ctx
		return nil, nil
	}
}

func TestUserIDServerInterceptor(t *testing.T) {
	interceptor := UserIDServerInterceptor()
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

		_, ok := ctxkeys.UserIDFromContext(captured)
		assert.False(t, ok)
	})

	t.Run("user_id и role из metadata", func(t *testing.T) {
		md := metadata.Pairs(
			ctxkeys.MetadataUserID, "user-123",
			ctxkeys.MetadataRole, "admin",
		)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		var captured context.Context
		_, err := interceptor(ctx, nil, info, echoHandler(&captured))
		require.NoError(t, err)

		userID, ok := ctxkeys.UserIDFromContext(captured)
		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)

		role, ok := ctxkeys.RoleFromContext(captured)
		assert.True(t, ok)
		assert.Equal(t, "admin", role)
	})

	t.Run("только user_id, без role", func(t *testing.T) {
		md := metadata.Pairs(ctxkeys.MetadataUserID, "user-456")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		var captured context.Context
		_, err := interceptor(ctx, nil, info, echoHandler(&captured))
		require.NoError(t, err)

		userID, ok := ctxkeys.UserIDFromContext(captured)
		assert.True(t, ok)
		assert.Equal(t, "user-456", userID)

		_, ok = ctxkeys.RoleFromContext(captured)
		assert.False(t, ok)
	})
}
