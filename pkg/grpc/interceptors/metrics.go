package interceptors

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsInterceptor собирает метрики для gRPC-запросов.
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Увеличиваем счётчик in-flight
		metrics.RequestsInFlight.Inc()
		defer metrics.RequestsInFlight.Dec()

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Milliseconds()

		// Определяем статус ответа
		statusCode := "ok"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = "unknown"
			}
		}

		// Обновляем метрики
		metrics.RequestsTotal.WithLabelValues(info.FullMethod, statusCode).Inc()
		metrics.RequestDuration.WithLabelValues(info.FullMethod).Observe(float64(duration))

		return resp, err
	}
}
