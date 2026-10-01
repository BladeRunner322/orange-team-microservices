package handlers

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
)

// ============================================================
//
//	mockAuthClient
//
// ============================================================
type mockAuthClient struct {
	registerFunc     func(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error)
	loginFunc        func(ctx context.Context, email, password string) (*auth.LoginResponse, error)
	validateFunc     func(ctx context.Context, token string) (ports.UserInfo, error)
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

func (m *mockAuthClient) ValidateToken(ctx context.Context, token string) (ports.UserInfo, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}
	return ports.UserInfo{UserID: "user-id", Role: "user"}, nil
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

// ============================================================
//
//	mockProfilesClient
//
// ============================================================
type mockProfilesClient struct {
	getMyProfileFunc    func(ctx context.Context) (*profiles.UserProfile, error)
	patchMyProfileFunc  func(ctx context.Context, patch *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error)
	deleteMyProfileFunc func(ctx context.Context) error
}

func (m *mockProfilesClient) GetMyProfile(ctx context.Context) (*profiles.UserProfile, error) {
	if m.getMyProfileFunc != nil {
		return m.getMyProfileFunc(ctx)
	}
	return &profiles.UserProfile{
		UserId:           "test-user-id",
		ProfileCompleted: false,
	}, nil
}

func (m *mockProfilesClient) PatchMyProfile(ctx context.Context, patch *profiles.PatchMyProfileRequest) (*profiles.UserProfile, error) {
	if m.patchMyProfileFunc != nil {
		return m.patchMyProfileFunc(ctx, patch)
	}
	return &profiles.UserProfile{
		UserId:           "test-user-id",
		ProfileCompleted: false,
	}, nil
}

func (m *mockProfilesClient) DeleteMyProfile(ctx context.Context) error {
	if m.deleteMyProfileFunc != nil {
		return m.deleteMyProfileFunc(ctx)
	}
	return nil
}

func (m *mockProfilesClient) Close() {}
