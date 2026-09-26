// Package health — HTTP-эндпоинт /health для Docker healthcheck.
//
// Отвечает {"status":"ok","service":"profiles"} на GET /health.
// По образцу services/auth/internal/interfaces/http/health/health.go.
package health

// TODO: Handler() http.HandlerFunc
