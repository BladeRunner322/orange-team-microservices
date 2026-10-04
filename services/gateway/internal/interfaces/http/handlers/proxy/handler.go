package proxy

import (
	"encoding/json"
	"net/http"
)

// Handler — заглушка для ещё не реализованных эндпоинтов.
// Возвращает 501 Not Implemented. Используется для Habits, Workouts,
// Leaderboard — пока соответствующие сервисы не готовы.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "endpoint not implemented yet",
	})
}
