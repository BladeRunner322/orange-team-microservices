package handlers

import (
	"encoding/json"
	"net/http"
)

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "endpoint not implemented yet",
	})
}
