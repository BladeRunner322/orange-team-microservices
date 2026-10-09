package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
)

// ClientMetricsInterceptor собирает метрики для исходящих gRPC-вызовов.
//
// Закрывает gap из ADR-004: раньше были инструментированы только
// входящие вызовы (MetricsInterceptor). Теперь видно, сколько времени
// Gateway ждёт downstream и какие коды ответов получает.
//
// target — адрес downstream ("auth-service:50051"), используется как
// label. Один и тот же физический сервис, вызванный с разных клиентов,
// попадёт в одну серию.
//
// Метрики:
//   - grpc_client_requests_total{target, method, status}
//   - grpc_client_request_duration_ms{target, method}
//   - grpc_client_requests_in_flight{target}
func ClientMetricsInterceptor(target string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		metrics.GRPCClientRequestsInFlight.WithLabelValues(target).Inc()
		defer metrics.GRPCClientRequestsInFlight.WithLabelValues(target).Dec()

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start).Milliseconds()

		statusCode := "ok"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = "unknown"
			}
		}

		metrics.GRPCClientRequestsTotal.
			WithLabelValues(target, method, statusCode).
			Inc()
		metrics.GRPCClientRequestDuration.
			WithLabelValues(target, method).
			Observe(float64(duration))

		return err
	}
}