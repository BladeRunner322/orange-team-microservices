package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
)

// RequestIDMiddleware добавляет request_id в контекст и заголовок ответа.
//
// Если клиент прислал X-Request-ID — используем его значение как есть.
// Это даёт сквозной трейс: клиент или внешний балансировщик может задать
// request_id, и он пройдёт через Gateway → gRPC metadata → downstream.
//
// Если заголовка нет (или он пустой) — генерируем UUID.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(ctxkeys.MetadataRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := ctxkeys.WithRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID извлекает request_id из контекста.
// Возвращает пустую строку, если request_id не установлен.
//
// Обёртка над ctxkeys.RequestIDFromContext для удобства вызывающих
// в HTTP-слое Gateway (logger.go).
func GetRequestID(ctx context.Context) string {
	requestID, _ := ctxkeys.RequestIDFromContext(ctx)
	return requestID
}
