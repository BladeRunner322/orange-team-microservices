package authgrpc

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor логирует каждый gRPC-запрос.
func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)

		entry := log.With(
			"method", info.FullMethod,
			"duration_ms", elapsed.Milliseconds(),
		)
		if err != nil {
			st, _ := status.FromError(err)
			entry.Warn("gRPC request failed", "code", st.Code().String(), "error", st.Message())
		} else {
			entry.Info("gRPC request succeeded")
		}
		return resp, err
	}
}
