package interceptors

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
)

// CircuitBreakerInterceptor оборачивает gRPC-вызов в Circuit Breaker.
//
// Если breaker в состоянии OPEN — вызов не доходит до downstream,
// возвращается codes.Unavailable немедленно (fail-fast). Это защищает
// от каскадного отказа: Gateway не копит горутины, ожидая таймаут
// от недоступного сервиса (см. ADR-020).
//
// При переходе состояния (closed → open и т.д.) вызывается
// OnStateChange, который обновляет метрики и логирует событие.
func CircuitBreakerInterceptor(cb *gobreaker.CircuitBreaker[any]) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		_, err := cb.Execute(func() (any, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})

		// gobreaker в состоянии OPEN возвращает ErrOpenState,
		// в состоянии HALF-OPEN с превышением MaxRequests — ErrTooManyRequests.
		// Мапим оба в codes.Unavailable, чтобы Gateway вернул клиенту 503,
		// а не 500 с непонятным сообщением.
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return status.Error(codes.Unavailable, "circuit breaker is open")
		}

		return err
	}
}

// NewCircuitBreaker создаёт Circuit Breaker для одного downstream-target.
//
// name — имя breaker'а (обычно Target, например "auth-service:50051").
// Используется как label метрики grpc_client_circuit_breaker_state.
//
// Параметры соответствуют pkg/grpc/client.Config (CB*).
// При переходе состояний обновляются метрики:
//   - grpc_client_circuit_breaker_state{target}
//   - grpc_client_circuit_breaker_transitions_total{target, from, to}
func NewCircuitBreaker(
	name string,
	maxRequests uint32,
	interval time.Duration,
	timeout time.Duration,
	minRequests uint32,
	errorRate float64,
) *gobreaker.CircuitBreaker[any] {
	// Инициализируем gauge в closed-state (0). Иначе до первого
	// перехода метрика не существует, и Grafana-панель пустая.
	metrics.GRPCClientCircuitBreakerState.WithLabelValues(name).Set(0)

	return gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < minRequests {
				return false
			}
			rate := float64(counts.TotalFailures) / float64(counts.Requests)
			return rate > errorRate
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			metrics.GRPCClientCircuitBreakerState.WithLabelValues(name).Set(stateToFloat(to))
			metrics.GRPCClientCircuitBreakerTransitions.
				WithLabelValues(name, stateToString(from), stateToString(to)).
				Inc()
		},
		IsSuccessful: func(err error) bool {
			// Не считаем бизнес-ошибки "провалами" для CB.
			// CB должен реагировать на транспортные проблемы, а не
			// на логические (NotFound, InvalidArgument, Forbidden и т.п.).
			// По умолчанию (IsSuccessful==nil) любая ошибка = failure.
			if err == nil {
				return true
			}
			st, ok := status.FromError(err)
			if !ok {
				// Не gRPC-status (транспортная ошибка) — failure.
				return false
			}
			switch st.Code() {
			case codes.InvalidArgument,
				codes.NotFound,
				codes.AlreadyExists,
				codes.PermissionDenied,
				codes.Unauthenticated,
				codes.FailedPrecondition,
				codes.OutOfRange,
				codes.Aborted:
				// Бизнес-ошибка — не failure для CB.
				return true
			}
			// Unavailable, DeadlineExceeded, Internal, Unknown,
			// ResourceExhausted, DataLoss — failure.
			return false
		},
	})
}

// stateToFloat конвертирует gobreaker.State в числовое значение
// для метрики: closed=0, half-open=1, open=2.
func stateToFloat(s gobreaker.State) float64 {
	switch s {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateHalfOpen:
		return 1
	case gobreaker.StateOpen:
		return 2
	default:
		return -1
	}
}

// stateToString конвертирует gobreaker.State в строку для метки.
func stateToString(s gobreaker.State) string {
	switch s {
	case gobreaker.StateClosed:
		return "closed"
	case gobreaker.StateHalfOpen:
		return "half-open"
	case gobreaker.StateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}
