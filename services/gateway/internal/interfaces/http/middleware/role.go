package middleware

import (
	"net/http"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

// RequireRole проверяет, что роль пользователя входит в список разрешённых.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetRole(r.Context())
			if !ok || !allowed[role] {
				httputil.SendError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
