package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
)

// UserIDServerInterceptor извлекает user_id и роль из входящих gRPC metadata
// и кладёт их в context.Context.
//
// Gateway передаёт эти данные в metadata после валидации JWT.
// Downstream-сервисы используют ctxkeys.UserIDFromContext для доступа к ним.
//
// Если metadata отсутствует (сервис вызван напрямую без Gateway),
// интерсептор пропускает запрос — context остаётся без user_id.
func UserIDServerInterceptor() grpc.UnaryServerInterceptor {
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

		if userIDs := md.Get(ctxkeys.MetadataUserID); len(userIDs) > 0 {
			ctx = ctxkeys.WithUserID(ctx, userIDs[0])
		}
		if roles := md.Get(ctxkeys.MetadataRole); len(roles) > 0 {
			ctx = ctxkeys.WithRole(ctx, roles[0])
		}

		return handler(ctx, req)
	}
}
