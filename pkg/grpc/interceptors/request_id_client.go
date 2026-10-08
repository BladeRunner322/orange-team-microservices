package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
)

// RequestIDClientInterceptor кладёт request_id из context
// в исходящие gRPC metadata, если он там есть.
//
// Используется Gateway при вызове downstream-сервисов.
// Если в context нет request_id — interceptor ничего не добавляет.
func RequestIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if requestID, ok := ctxkeys.RequestIDFromContext(ctx); ok && requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, ctxkeys.MetadataRequestID, requestID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
