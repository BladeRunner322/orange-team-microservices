package interceptors

import (
	"context"
	"time"

	retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
)

// RetryInterceptor ретраит gRPC-вызовы только для методов из whitelist.
//
// Whitelist — полные имена методов в формате "/package.Service/Method",
// например "/auth.AuthService/ValidateToken". Методы вне списка вызываются
// напрямую, без retry (см. ADR-020, п.3).
//
// Ретраятся только транспортные ошибки: codes.Unavailable и
// codes.DeadlineExceeded. Логические (NotFound, InvalidArgument,
// PermissionDenied и т.п.) возвращаются клиенту сразу.
//
// target — имя downstream ("auth-service:50051"), используется как label
// в метрике grpc_client_retry_attempts_total.
//
// maxAttempts — максимум попыток включая первую. Внутри библиотеки
// grpc_retry это количество retries (maxAttempts - 1), потому что
// первая попытка за retry не считается.
func RetryInterceptor(
	target string,
	methods []string,
	maxAttempts uint32,
	baseDelay time.Duration,
	maxDelay time.Duration,
) grpc.UnaryClientInterceptor {
	whitelist := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		whitelist[m] = struct{}{}
	}

	// retryInterceptor создаётся один раз. Он не знает про method
	// и target — эти данные приходят либо через контекст (метрики
	// считаем сами), либо через параметры вызова.
	//
	// WithMax в go-grpc-middleware/retry задаёт МАКСИМАЛЬНОЕ КОЛИЧЕСТВО
	// ПОПЫТОК, а не retry (несмотря на название параметра maxRetries).
	// WithMax(3) = 3 попытки: 1 initial + 2 retry.
	// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2.3.1/interceptors/retry/options.go#L29-L31
	retryInterceptor := retry.UnaryClientInterceptor(
		retry.WithMax(uint(maxAttempts)),
		retry.WithBackoff(retry.BackoffExponentialWithJitter(baseDelay, 0.2)),
		// DeadlineExceeded НЕ ретраится автоматически (см. комментарий
		// к WithCodes в исходниках библиотеки). Оставляем только Unavailable.
		retry.WithCodes(codes.Unavailable),
	)

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Метод не в whitelist — вызов напрямую.
		if _, ok := whitelist[method]; !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Оборачиваем invoker, чтобы посчитать количество реальных
		// вызовов downstream. Разница между calls и 1 = число retry.
		var calls int
		wrappedInvoker := func(
			ctx context.Context,
			method string,
			req, reply any,
			cc *grpc.ClientConn,
			opts ...grpc.CallOption,
		) error {
			calls++
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		err := retryInterceptor(ctx, method, req, reply, cc, wrappedInvoker, opts...)

		// Считаем retry только если они реально были (calls > 1).
		// Иначе метрика превратится в шум: каждый вызов писал бы
		// success, даже когда retry не потребовался.
		if calls > 1 {
			retryCount := float64(calls - 1)
			metrics.GRPCClientRetryAttempts.
				WithLabelValues(target, method, "retry").
				Add(retryCount)

			if err != nil {
				metrics.GRPCClientRetryAttempts.
					WithLabelValues(target, method, "failure").
					Inc()
			} else {
				metrics.GRPCClientRetryAttempts.
					WithLabelValues(target, method, "success").
					Inc()
			}
		}

		return err
	}
}