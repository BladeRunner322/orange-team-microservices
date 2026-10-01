package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
)

// ReadinessHandler проверяет готовность Auth обслуживать трафик:
// доступность PostgreSQL и Redis.
type ReadinessHandler struct {
	pool  *postgres.PgxPool
	redis *redis.Client
}

// NewReadinessHandler создаёт handler с зависимостями.
func NewReadinessHandler(pool *postgres.PgxPool, redisClient *redis.Client) *ReadinessHandler {
	return &ReadinessHandler{
		pool:  pool,
		redis: redisClient,
	}
}

// Handle отдаёт 200, если Postgres и Redis доступны, иначе 503.
func (h *ReadinessHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	ready := true

	if err := h.pool.Ping(ctx); err != nil {
		checks["postgres"] = "fail: " + err.Error()
		ready = false
	} else {
		checks["postgres"] = "ok"
	}

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
