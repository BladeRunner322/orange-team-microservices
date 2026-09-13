package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
)

// mockAuthClient — реализация ports.AuthClientInterface для тестов middleware.
type mockAuthClient struct {
	validateFunc func(ctx context.Context, token string) (ports.UserInfo, error)
}

func (m *mockAuthClient) ValidateToken(ctx context.Context, token string) (ports.UserInfo, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}
	return ports.UserInfo{UserID: "user-id", Role: "user"}, nil
}

func (m *mockAuthClient) Register(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthClient) Login(ctx context.Context, email, password string) (*auth.LoginResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshTokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthClient) Logout(ctx context.Context, refreshToken string) error {
	return errors.New("not implemented")
}

func (m *mockAuthClient) Close() {}

// nextHandler возвращает простой handler, который отвечает 200.
func nextHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
