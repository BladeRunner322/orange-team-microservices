package interceptors

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor логирует каждый gRPC-запрос.
//
// Если в context установлен request_id (RequestIDServerInterceptor выше
// в цепочке) — добавляет его в лог. Это даёт сквозной трейс: один
// request_id виден в логах Gateway и всех downstream-сервисов.
func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)

		entry := log.With(
			"method", info.FullMethod,
			"duration_ms", elapsed.Milliseconds(),
		)

		if requestID, ok := ctxkeys.RequestIDFromContext(ctx); ok && requestID != "" {
			entry = entry.With("request_id", requestID)
		}

		if err != nil {
			st, _ := status.FromError(err)
			entry.Warn("gRPC request failed", "code", st.Code().String(), "error", st.Message())
		} else {
			entry.Info("gRPC request succeeded")
		}
		return resp, err
	}
}
