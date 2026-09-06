package usecases

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type LoginUseCase struct {
	repo         ports.Repository
	tokenManager ports.TokenManager
}

func NewLoginUseCase(repo ports.Repository, tokenManager ports.TokenManager) *LoginUseCase {
	return &LoginUseCase{repo: repo, tokenManager: tokenManager}
}

func (uc *LoginUseCase) Execute(ctx context.Context, emailStr, password string) (string, error) {
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String()), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}
	return uc.tokenManager.Generate(ctx, user.ID.String())
}
