package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegisterHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockAuthClient{
			registerFunc: func(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error) {
				return &auth.RegisterResponse{Id: "123", Email: email, FullName: fullName}, nil
			},
		}
		handler := RegisterHandler(mock)

		reqBody := RegisterRequest{Email: "test@example.com", Password: "password123", FullName: "Test User"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp auth.RegisterResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "123", resp.Id)
		assert.Equal(t, "test@example.com", resp.Email)
	})

	t.Run("invalid email", func(t *testing.T) {
		mock := &mockAuthClient{}
		handler := RegisterHandler(mock)
		reqBody := RegisterRequest{Email: "invalid", Password: "password123", FullName: "Test"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed on the 'email' tag")
	})

	t.Run("weak password", func(t *testing.T) {
		mock := &mockAuthClient{}
		handler := RegisterHandler(mock)
		reqBody := RegisterRequest{Email: "test@example.com", Password: "123", FullName: "Test"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed on the 'min' tag")
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockAuthClient{
			registerFunc: func(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "email already exists")
			},
		}
		handler := RegisterHandler(mock)
		reqBody := RegisterRequest{Email: "test@example.com", Password: "password123", FullName: "Test"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "email already exists")
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockAuthClient{
			loginFunc: func(ctx context.Context, email, password string) (*auth.LoginResponse, error) {
				return &auth.LoginResponse{AccessToken: "token", TokenType: "Bearer"}, nil
			},
		}
		handler := LoginHandler(mock)
		reqBody := LoginRequest{Email: "test@example.com", Password: "pass"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp auth.LoginResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "token", resp.AccessToken)
	})

	t.Run("invalid email", func(t *testing.T) {
		mock := &mockAuthClient{}
		handler := LoginHandler(mock)
		reqBody := LoginRequest{Email: "invalid", Password: "pass"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed on the 'email' tag")
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockAuthClient{
			loginFunc: func(ctx context.Context, email, password string) (*auth.LoginResponse, error) {
				return nil, status.Error(codes.Unauthenticated, "invalid credentials")
			},
		}
		handler := LoginHandler(mock)
		reqBody := LoginRequest{Email: "test@example.com", Password: "wrong"}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid credentials")
	})
}
