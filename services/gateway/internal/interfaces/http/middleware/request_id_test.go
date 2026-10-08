package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("генерирует request_id и кладёт в context и header", func(t *testing.T) {
		mw := RequestIDMiddleware

		var gotID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID = GetRequestID(r.Context())
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		// Заголовок ответа установлен
		headerID := rec.Header().Get("X-Request-ID")
		require.NotEmpty(t, headerID)

		// В context тот же самый ID
		assert.Equal(t, headerID, gotID)
	})

	t.Run("request_id — валидный UUID", func(t *testing.T) {
		mw := RequestIDMiddleware

		var gotID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID = GetRequestID(r.Context())
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		_, err := uuid.Parse(gotID)
		assert.NoError(t, err, "request_id должен быть валидным UUID")
	})

	t.Run("каждый запрос получает уникальный request_id", func(t *testing.T) {
		mw := RequestIDMiddleware

		var firstID, secondID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if firstID == "" {
				firstID = GetRequestID(r.Context())
			} else {
				secondID = GetRequestID(r.Context())
			}
		})
		handler := mw(next)

		// Первый запрос
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))

		// Второй запрос
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.NotEqual(t, firstID, secondID, "ID должны отличаться")
	})

	t.Run("GetRequestID на пустом context — пустая строка", func(t *testing.T) {
		got := GetRequestID(context.Background())
		assert.Empty(t, got)
	})

	t.Run("подхватывает X-Request-ID от клиента", func(t *testing.T) {
		// Сквозной трейс: если клиент прислал X-Request-ID —
		// используем его значение как есть, не генерируем новый.
		mw := RequestIDMiddleware

		var gotID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID = GetRequestID(r.Context())
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "client-provided-id")
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, "client-provided-id", gotID)
		assert.Equal(t, "client-provided-id", rec.Header().Get("X-Request-ID"))
	})

	t.Run("пустой X-Request-ID от клиента — генерируем UUID", func(t *testing.T) {
		mw := RequestIDMiddleware

		var gotID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID = GetRequestID(r.Context())
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "")
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		_, err := uuid.Parse(gotID)
		assert.NoError(t, err, "при пустом заголовке должен быть сгенерирован UUID")
		assert.Equal(t, gotID, rec.Header().Get("X-Request-ID"))
	})
}
