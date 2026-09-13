package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
)

// UserInfo — данные пользователя, полученные от Auth по токену.
type UserInfo struct {
	UserID string
	Role   string
}

type AuthClientInterface interface {
	ValidateToken(ctx context.Context, token string) (UserInfo, error)
	Register(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error)
	Login(ctx context.Context, email, password string) (*auth.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*auth.RefreshTokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Close()
}
