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

func TestRequestIDClientInterceptor(t *testing.T) {
	interceptor := RequestIDClientInterceptor()
	method := "/test.Service/Method"

	t.Run("context без request_id — metadata пустая", func(t *testing.T) {
		var captured context.Context
		err := interceptor(
			context.Background(),
			method,
			nil, nil, nil,
			captureInvoker(&captured),
		)
		require.NoError(t, err)

		assert.Empty(t, outgoingValues(t, captured, ctxkeys.MetadataRequestID))
	})

	t.Run("request_id из context", func(t *testing.T) {
		ctx := ctxkeys.WithRequestID(context.Background(), "req-123")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		values := outgoingValues(t, captured, ctxkeys.MetadataRequestID)
		require.Len(t, values, 1)
		assert.Equal(t, "req-123", values[0])
	})

	t.Run("пустой request_id не добавляется", func(t *testing.T) {
		ctx := ctxkeys.WithRequestID(context.Background(), "")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		assert.Empty(t, outgoingValues(t, captured, ctxkeys.MetadataRequestID))
	})

	t.Run("дополняет существующую outgoing metadata", func(t *testing.T) {
		existing := metadata.Pairs("x-trace-id", "trace-abc")
		ctx := metadata.NewOutgoingContext(context.Background(), existing)
		ctx = ctxkeys.WithRequestID(ctx, "req-42")

		var captured context.Context
		err := interceptor(ctx, method, nil, nil, nil, captureInvoker(&captured))
		require.NoError(t, err)

		requestIDs := outgoingValues(t, captured, ctxkeys.MetadataRequestID)
		require.Len(t, requestIDs, 1)
		assert.Equal(t, "req-42", requestIDs[0])

		traces := outgoingValues(t, captured, "x-trace-id")
		require.Len(t, traces, 1)
		assert.Equal(t, "trace-abc", traces[0])
	})

	t.Run("invoker вызывается ровно один раз", func(t *testing.T) {
		ctx := ctxkeys.WithRequestID(context.Background(), "req-1")

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
