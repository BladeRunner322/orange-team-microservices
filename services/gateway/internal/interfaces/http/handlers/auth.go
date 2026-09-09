package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/utils"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2,max=50"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func RegisterHandler(authClient ports.AuthClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := authClient.Register(r.Context(), req.Email, req.Password, req.FullName)
		if err != nil {
			code, msg := utils.GrpcErrorToHTTP(err)
			utils.SendError(w, code, msg)
			return
		}

		utils.SendJSON(w, http.StatusCreated, resp)
	}
}

func LoginHandler(authClient ports.AuthClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := authClient.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			code, msg := utils.GrpcErrorToHTTP(err)
			utils.SendError(w, code, msg)
			return
		}

		utils.SendJSON(w, http.StatusOK, resp)
	}
}
