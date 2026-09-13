package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
)

// HTTPMetricsMiddleware собирает HTTP-метрики Prometheus:
// количество запросов, длительность и число в обработке.
//
// Путь нормализуется через chi RoutePattern, чтобы /workouts/123
// и /workouts/456 не создавали отдельные метрики (кардинальность).
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		start := time.Now()

		rw := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		path := normalizedPath(r)

		metrics.HTTPRequestsTotal.
			WithLabelValues(r.Method, path, strconv.Itoa(rw.status)).
			Inc()
		metrics.HTTPRequestDuration.
			WithLabelValues(r.Method, path).
			Observe(duration)
	})
}

// normalizedPath возвращает шаблон маршрута chi (например, "/workouts/{workoutId}").
// Если паттерн недоступен — возвращает сырой путь.
func normalizedPath(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}

// metricsResponseWriter перехватывает статус-код ответа.
type metricsResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
