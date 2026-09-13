package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	t.Run("пустой контекст — 403", func(t *testing.T) {
		mw := RequireRole("admin")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "forbidden")
	})

	t.Run("роль user — 403 на admin", func(t *testing.T) {
		mw := RequireRole("admin")

		ctx := context.WithValue(context.Background(), RoleKey, "user")
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("роль admin — 200, next вызван", func(t *testing.T) {
		mw := RequireRole("admin")
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.WithValue(context.Background(), RoleKey, "admin")
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called)
	})

	t.Run("несколько разрешённых ролей", func(t *testing.T) {
		mw := RequireRole("admin", "moderator")

		for _, role := range []string{"admin", "moderator"} {
			t.Run(role, func(t *testing.T) {
				called := false
				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
				})

				ctx := context.WithValue(context.Background(), RoleKey, role)
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
				rec := httptest.NewRecorder()

				mw(next).ServeHTTP(rec, req)

				assert.True(t, called, "роль %s должна быть пропущена", role)
			})
		}
	})

	t.Run("пустая роль — 403", func(t *testing.T) {
		mw := RequireRole("admin")

		ctx := context.WithValue(context.Background(), RoleKey, "")
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("без списка ролей — всё запрещено", func(t *testing.T) {
		mw := RequireRole()

		ctx := context.WithValue(context.Background(), RoleKey, "admin")
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(nextHandler(t)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
