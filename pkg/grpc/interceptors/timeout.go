package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// TimeoutInterceptor ограничивает время выполнения каждого gRPC-вызова.
// Если вызов не завершится за timeout — вернётся context deadline exceeded.
//
// Если timeout <= 0 — интерсептор не применяется (вызов идёт без дедлайна).
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if timeout <= 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
