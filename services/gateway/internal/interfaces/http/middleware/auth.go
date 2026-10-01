package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/authctx"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

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

			info, err := authClient.ValidateToken(r.Context(), token)
			if err != nil || info.UserID == "" {
				httputil.SendError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := authctx.WithUserID(r.Context(), info.UserID)
			ctx = authctx.WithRole(ctx, info.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	return authctx.UserIDFromContext(ctx)
}

func GetRole(ctx context.Context) (string, bool) {
	return authctx.RoleFromContext(ctx)
}
