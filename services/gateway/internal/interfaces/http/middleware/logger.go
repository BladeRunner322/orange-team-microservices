package middleware

import (
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
)

// LoggerMiddleware логирует каждый запрос с контекстными полями.
func LoggerMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := GetRequestID(r.Context())

			entry := log.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			entry.Info("request started")

			// Обёртка для записи статуса
			rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			// После обработки проверяем, есть ли user_id
			if userID, ok := GetUserID(r.Context()); ok {
				entry = entry.With("user_id", userID)
			}

			entry.With(
				"status", rw.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
			).Info("request finished")
		})
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
