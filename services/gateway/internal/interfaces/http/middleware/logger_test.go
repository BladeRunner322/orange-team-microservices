package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
)

func TestLoggerMiddleware(t *testing.T) {
	t.Run("логирует запрос без user_id", func(t *testing.T) {
		mw := LoggerMiddleware(logger.NewTestLogger())

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("логирует с user_id из контекста", func(t *testing.T) {
		mw := LoggerMiddleware(logger.NewTestLogger())

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := ctxkeys.WithUserID(httptest.NewRequest(http.MethodGet, "/test", nil).Context(), "user-123")
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("логирует с request_id из контекста", func(t *testing.T) {
		mw := LoggerMiddleware(logger.NewTestLogger())

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		ctx := ctxkeys.WithRequestID(httptest.NewRequest(http.MethodGet, "/test", nil).Context(), "req-abc")
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("логирует статус 500", func(t *testing.T) {
		mw := LoggerMiddleware(logger.NewTestLogger())

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		req := httptest.NewRequest(http.MethodPost, "/fail", nil)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("логирует с user_id и request_id одновременно", func(t *testing.T) {
		mw := LoggerMiddleware(logger.NewTestLogger())

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ctx := ctxkeys.WithUserID(httptest.NewRequest(http.MethodGet, "/test", nil).Context(), "user-abc")
		ctx = ctxkeys.WithRequestID(ctx, "req-xyz")
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
