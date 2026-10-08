package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
)

// RequestIDServerInterceptor извлекает request_id из входящих gRPC metadata
// и кладёт его в context.Context.
//
// Gateway передаёт request_id в metadata. Downstream-сервисы используют
// ctxkeys.RequestIDFromContext для доступа к нему — например,
// LoggingInterceptor логирует с этим полем.
//
// Если metadata отсутствует (сервис вызван напрямую без Gateway),
// interceptor пропускает запрос — context остаётся без request_id.
func RequestIDServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		if requestIDs := md.Get(ctxkeys.MetadataRequestID); len(requestIDs) > 0 {
			ctx = ctxkeys.WithRequestID(ctx, requestIDs[0])
		}

		return handler(ctx, req)
	}
}
