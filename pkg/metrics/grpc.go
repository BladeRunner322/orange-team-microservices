package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// GRPCRequestsTotal — общее количество gRPC-запросов по методу и статусу.
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	// GRPCRequestDuration — длительность gRPC-запросов в миллисекундах.
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_ms",
			Help:    "Duration of gRPC requests in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
		[]string{"method"},
	)

	// GRPCRequestsInFlight — количество gRPC-запросов в обработке.
	GRPCRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "grpc_requests_in_flight",
			Help: "Number of gRPC requests currently being processed",
		},
	)

	// ============================================================
	//  Client-side метрики (Gateway → downstream).
	//  Закрывают gap из ADR-004: раньше инструментированы были
	//  только входящие gRPC-вызовы.
	// ============================================================

	// GRPCClientRequestsTotal — общее количество исходящих gRPC-запросов
	// по target (адрес downstream), методу и статусу.
	GRPCClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "Total number of outgoing gRPC requests",
		},
		[]string{"target", "method", "status"},
	)

	// GRPCClientRequestDuration — длительность исходящих gRPC-запросов
	// в миллисекундах. Buckets совпадают с server-side, чтобы метрики
	// были сопоставимы.
	GRPCClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_client_request_duration_ms",
			Help:    "Duration of outgoing gRPC requests in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
		[]string{"target", "method"},
	)

	// GRPCClientRequestsInFlight — количество исходящих gRPC-запросов
	// в обработке сейчас, по target.
	GRPCClientRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_client_requests_in_flight",
			Help: "Number of outgoing gRPC requests currently being processed",
		},
		[]string{"target"},
	)

	// GRPCClientCircuitBreakerState — текущее состояние Circuit Breaker
	// для target: 0=closed, 1=half-open, 2=open.
	GRPCClientCircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_client_circuit_breaker_state",
			Help: "Current state of the gRPC client circuit breaker (0=closed, 1=half-open, 2=open)",
		},
		[]string{"target"},
	)

	// GRPCClientCircuitBreakerTransitions — количество переходов
	// состояния Circuit Breaker по target и направлению (from, to).
	GRPCClientCircuitBreakerTransitions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_circuit_breaker_transitions_total",
			Help: "Total number of circuit breaker state transitions",
		},
		[]string{"target", "from", "to"},
	)

	// GRPCClientRetryAttempts — количество retry-попыток по target,
	// методу и результату (result: success|failure).
	GRPCClientRetryAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_retry_attempts_total",
			Help: "Total number of retry attempts for outgoing gRPC requests",
		},
		[]string{"target", "method", "result"},
	)
)
