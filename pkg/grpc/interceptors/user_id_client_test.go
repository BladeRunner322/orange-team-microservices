package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/authctx"
)

// captureInvoker сохраняет context, с которым вызывается invoker,
// чтобы можно было проверить, что interceptor положил в outgoing metadata.
func captureInvoker(captured *context.Context) grpc.UnaryInvoker {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		*captured = ctx
		return nil
	}
}

// outgoingValues достаёт значения ключа из outgoing metadata.
func outgoingValues(t *testing.T, ctx context.Context, key string) []string {
	t.Helper()
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return nil
	}
	return md.Get(key)
}

func TestUserIDClientInterceptor(t *testing.T) {
	interceptor := UserIDClientInterceptor()
	method := "/test.Service/Method"

	t.Run("context без user_id — metadata пустая", func(t *testing.T) {
		var captured context.Context
		err := interceptor(
			context.Background(),
			method,
			nil, nil, nil,
			captureInvoker(&captured),
		)
		require.NoError(t, err)

		assert.Empty(t, outgoingValues(t, captured, authctx.MetadataUserID))
		assert.Empty(t, outgoingValues(t, captured, authctx.MetadataRole))
	})

	t.Run("только user_id", func(t *testing.T) {
		ctx := authctx.WithUserID(context.Background(), "user-123")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		values := outgoingValues(t, captured, authctx.MetadataUserID)
		require.Len(t, values, 1)
		assert.Equal(t, "user-123", values[0])

		assert.Empty(t, outgoingValues(t, captured, authctx.MetadataRole))
	})

	t.Run("user_id и role", func(t *testing.T) {
		ctx := authctx.WithUserID(context.Background(), "user-123")
		ctx = authctx.WithRole(ctx, "admin")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		userIDs := outgoingValues(t, captured, authctx.MetadataUserID)
		require.Len(t, userIDs, 1)
		assert.Equal(t, "user-123", userIDs[0])

		roles := outgoingValues(t, captured, authctx.MetadataRole)
		require.Len(t, roles, 1)
		assert.Equal(t, "admin", roles[0])
	})

	t.Run("пустой user_id не добавляется", func(t *testing.T) {
		ctx := authctx.WithUserID(context.Background(), "")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		assert.Empty(t, outgoingValues(t, captured, authctx.MetadataUserID))
	})

	t.Run("пустой role не добавляется", func(t *testing.T) {
		ctx := authctx.WithRole(context.Background(), "")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		assert.Empty(t, outgoingValues(t, captured, authctx.MetadataRole))
	})

	t.Run("дополняет существующую outgoing metadata", func(t *testing.T) {
		// Предзаполняем metadata чем-то другим (например, трассировка)
		existing := metadata.Pairs("x-trace-id", "trace-abc")
		ctx := metadata.NewOutgoingContext(context.Background(), existing)
		ctx = authctx.WithUserID(ctx, "user-42")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		// user_id добавился
		userIDs := outgoingValues(t, captured, authctx.MetadataUserID)
		require.Len(t, userIDs, 1)
		assert.Equal(t, "user-42", userIDs[0])

		// существующая metadata не потерялась
		traces := outgoingValues(t, captured, "x-trace-id")
		require.Len(t, traces, 1)
		assert.Equal(t, "trace-abc", traces[0])
	})

	t.Run("invoker вызывается ровно один раз", func(t *testing.T) {
		ctx := authctx.WithUserID(context.Background(), "user-1")

		calls := 0
		invoker := func(
			ctx context.Context,
			method string,
			req, reply any,
			cc *grpc.ClientConn,
			opts ...grpc.CallOption,
		) error {
			calls++
			return nil
		}

		err := interceptor(ctx, method, nil, nil, nil, invoker)
		require.NoError(t, err)

		assert.Equal(t, 1, calls)
	})
}
