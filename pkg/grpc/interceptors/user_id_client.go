package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/authctx"
)

// UserIDClientInterceptor кладёт user_id и role из context
// в исходящие gRPC metadata, если они там есть.
//
// Используется Gateway при вызове downstream-сервисов.
// Если в context нет user_id (публичный запрос до аутентификации),
// interceptor ничего не добавляет.
func UserIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var pairs []string

		if userID, ok := authctx.UserIDFromContext(ctx); ok && userID != "" {
			pairs = append(pairs, authctx.MetadataUserID, userID)
		}
		if role, ok := authctx.RoleFromContext(ctx); ok && role != "" {
			pairs = append(pairs, authctx.MetadataRole, role)
		}

		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
