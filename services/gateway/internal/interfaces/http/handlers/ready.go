package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
)

// ReadinessHandler проверяет готовность Gateway обслуживать трафик:
// доступность Redis (нужен для rate limiting).
type ReadinessHandler struct {
	redis *redis.Client
}

func NewReadinessHandler(redisClient *redis.Client) *ReadinessHandler {
	return &ReadinessHandler{redis: redisClient}
}

func (h *ReadinessHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	ready := true

	if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = "fail: " + err.Error()
		ready = false
	} else {
		checks["redis"] = "ok"
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
		"checks": checks,
	})
}
