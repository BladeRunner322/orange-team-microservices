package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

var validate = validator.New()

// RegisterHandler возвращает обработчик POST /register.
func RegisterHandler(authClient ports.AuthClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := authClient.Register(r.Context(), req.Email, req.Password, req.FullName)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusCreated, resp)
	}
}

// LoginHandler возвращает обработчик POST /login.
func LoginHandler(authClient ports.AuthClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := authClient.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, resp)
	}
}

// RefreshHandler возвращает обработчик POST /refresh.
func RefreshHandler(authClient ports.AuthClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := authClient.RefreshToken(r.Context(), req.RefreshToken)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, resp)
	}
}

// LogoutHandler возвращает обработчик POST /logout.
func LogoutHandler(authClient ports.AuthClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LogoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := authClient.Logout(r.Context(), req.RefreshToken); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
	}
}
