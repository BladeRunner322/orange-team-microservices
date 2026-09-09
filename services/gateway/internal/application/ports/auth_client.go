package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
)

type AuthClientInterface interface {
	ValidateToken(ctx context.Context, token string) (string, error)
	Register(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error)
	Login(ctx context.Context, email, password string) (*auth.LoginResponse, error)
	Close()
}
