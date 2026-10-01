package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
)

// ReadinessHandler проверяет готовность Profiles обслуживать трафик:
// доступность PostgreSQL.
type ReadinessHandler struct {
	pool *postgres.PgxPool
}

// NewReadinessHandler создаёт handler с пулом соединений.
func NewReadinessHandler(pool *postgres.PgxPool) *ReadinessHandler {
	return &ReadinessHandler{pool: pool}
}

// Handle отдаёт 200, если Postgres доступен, иначе 503.
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
