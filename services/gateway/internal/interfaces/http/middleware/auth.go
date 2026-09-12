package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(authClient ports.AuthClientInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.SendError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				httputil.SendError(w, http.StatusUnauthorized, "invalid Authorization header format")
				return
			}
			token := parts[1]

			userID, err := authClient.ValidateToken(r.Context(), token)
			if err != nil || userID == "" {
				httputil.SendError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
