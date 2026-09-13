package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("нет Authorization header — 401", func(t *testing.T) {
		mw := AuthMiddleware(&mockAuthClient{})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "missing Authorization header")
	})

	t.Run("неверный формат — 401", func(t *testing.T) {
		mw := AuthMiddleware(&mockAuthClient{})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "NotBearer abc")
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid Authorization header format")
	})

	t.Run("только один токен без префикса — 401", func(t *testing.T) {
		mw := AuthMiddleware(&mockAuthClient{})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "abc")
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("невалидный токен (пустой UserID) — 401", func(t *testing.T) {
		client := &mockAuthClient{
			validateFunc: func(ctx context.Context, token string) (ports.UserInfo, error) {
				return ports.UserInfo{}, nil
			},
		}
		mw := AuthMiddleware(client)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid")
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid or expired token")
	})

	t.Run("ошибка от Auth — 401", func(t *testing.T) {
		client := &mockAuthClient{
			validateFunc: func(ctx context.Context, token string) (ports.UserInfo, error) {
				return ports.UserInfo{}, errors.New("unavailable")
			},
		}
		mw := AuthMiddleware(client)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer token")
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("валидный токен — кладём user_id и role в контекст", func(t *testing.T) {
		client := &mockAuthClient{
			validateFunc: func(ctx context.Context, token string) (ports.UserInfo, error) {
				assert.Equal(t, "my-token", token)
				return ports.UserInfo{UserID: "user-123", Role: "admin"}, nil
			},
		}
		mw := AuthMiddleware(client)

		var gotUserID, gotRole string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUserID, _ = GetUserID(r.Context())
			gotRole, _ = GetRole(r.Context())
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer my-token")
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, "user-123", gotUserID)
		assert.Equal(t, "admin", gotRole)
	})

	t.Run("Bearer в любом регистре — работает", func(t *testing.T) {
		client := &mockAuthClient{
			validateFunc: func(ctx context.Context, token string) (ports.UserInfo, error) {
				return ports.UserInfo{UserID: "u", Role: "user"}, nil
			},
		}
		mw := AuthMiddleware(client)

		for _, prefix := range []string{"Bearer", "bearer", "BEARER"} {
			t.Run(prefix, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", prefix+" my-token")
				rec := httptest.NewRecorder()

				mw(nextHandler(t)).ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
			})
		}
	})

	t.Run("GetUserID на пустом контексте — false", func(t *testing.T) {
		_, ok := GetUserID(context.Background())
		assert.False(t, ok)
	})

	t.Run("GetRole на пустом контексте — false", func(t *testing.T) {
		_, ok := GetRole(context.Background())
		assert.False(t, ok)
	})
}
