package httputil

import (
	"encoding/json"
	"net/http"
)

// SendJSON отправляет JSON-ответ с указанным статусом.
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// SendError отправляет JSON-ошибку с указанным статусом.
func SendError(w http.ResponseWriter, status int, message string) {
	SendJSON(w, status, map[string]string{"error": message})
}
