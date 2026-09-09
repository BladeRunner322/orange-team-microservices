package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/middleware"
)

// GetUserHandler — временная заглушка для /users/me
// Возвращает user_id из контекста (после аутентификации)
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"user not authenticated"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"message": "temporary stub, real Users service will be added later",
	})
}
