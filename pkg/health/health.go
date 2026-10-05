// Package health — HTTP-обработчики /health, /ready и /metrics
// для сервисов проекта.
//
// Отдаёт отдельные http.HandlerFunc, чтобы каждый сервис мог
// повесить их на свой роутер (ServeMux, chi, что угодно).
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Check — функция проверки одной зависимости.
// Возвращает nil, если зависимость здорова.
type Check func(ctx context.Context) error

// HealthHandler возвращает обработчик /health (liveness).
// Отвечает всегда 200, если процесс жив.
func HealthHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": serviceName,
		})
	}
}

// ReadyHandler возвращает обработчик /ready (readiness).
// Прогоняет все checks с общим таймаутом 2 секунды.
// Если хоть одна проверка упала — возвращает 503.
func ReadyHandler(checks map[string]Check) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		results := map[string]string{}
		ready := true

		for name, check := range checks {
			if err := check(ctx); err != nil {
				results[name] = "fail: " + err.Error()
				ready = false
			} else {
				results[name] = "ok"
			}
		}

		status := http.StatusOK
		overall := "ok"
		if !ready {
			status = http.StatusServiceUnavailable
			overall = "not_ready"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": overall,
			"checks": results,
		})
	}
}

// MetricsHandler возвращает обработчик /metrics (Prometheus).
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
