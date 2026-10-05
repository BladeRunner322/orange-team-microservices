package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
)

// Check — функция проверки одной зависимости.
// Возвращает nil, если зависимость здорова.
type Check func(ctx context.Context) error

// ReadinessHandler проверяет готовность Gateway обслуживать трафик:
// доступность Redis и gRPC-соединений с downstream-сервисами.
type ReadinessHandler struct {
	redis  *redis.Client
	checks map[string]Check
}

// NewReadinessHandler создаёт handler с зависимостями.
//
// checks — карта «имя зависимости → функция проверки».
// Имена попадают в JSON-ответ (например, {"auth":"ok","profiles":"ok"}).
func NewReadinessHandler(redisClient *redis.Client, checks map[string]Check) *ReadinessHandler {
	return &ReadinessHandler{
		redis:  redisClient,
		checks: checks,
	}
}

// Handle отдаёт 200, если все зависимости доступны, иначе 503.
func (h *ReadinessHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	results := map[string]string{}
	ready := true

	// Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		results["redis"] = "fail: " + err.Error()
		ready = false
	} else {
		results["redis"] = "ok"
	}

	// gRPC-зависимости
	for name, check := range h.checks {
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
