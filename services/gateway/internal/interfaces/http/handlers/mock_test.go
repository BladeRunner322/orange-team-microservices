package handlers

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
)

type mockAuthClient struct {
	registerFunc     func(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error)
	loginFunc        func(ctx context.Context, email, password string) (*auth.LoginResponse, error)
	validateFunc     func(ctx context.Context, token string) (string, error)
	refreshTokenFunc func(ctx context.Context, refreshToken string) (*auth.RefreshTokenResponse, error)
	logoutFunc       func(ctx context.Context, refreshToken string) error
}

func (m *mockAuthClient) Register(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, email, password, fullName)
	}
	return &auth.RegisterResponse{Id: "test-id", Email: email, FullName: fullName}, nil
}

func (m *mockAuthClient) Login(ctx context.Context, email, password string) (*auth.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, email, password)
	}
	return &auth.LoginResponse{AccessToken: "test-token", TokenType: "Bearer"}, nil
}

func (m *mockAuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}
	return "user-id", nil
}

func (m *mockAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshTokenResponse, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(ctx, refreshToken)
	}
	return &auth.RefreshTokenResponse{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
		TokenType:    "Bearer",
	}, nil
}

func (m *mockAuthClient) Logout(ctx context.Context, refreshToken string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, refreshToken)
	}
	return nil
}

func (m *mockAuthClient) Close() {}
