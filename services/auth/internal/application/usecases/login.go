package usecases

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type Login struct {
	repo         ports.Repository
	tokenManager ports.TokenManager
	logger       *logger.Logger
}

func NewLogin(repo ports.Repository, tokenManager ports.TokenManager, log *logger.Logger) *Login {
	return &Login{repo: repo, tokenManager: tokenManager, logger: log}
}

func (uc *Login) Execute(ctx context.Context, emailStr, password string) (string, error) {
	log := uc.logger.With("email", emailStr, "operation", "Login")

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		log.Warn("invalid email", "error", err)
		return "", domain.ErrInvalidCredentials
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		log.Warn("user not found")
		return "", domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String()), []byte(password)); err != nil {
		log.Warn("invalid password")
		return "", domain.ErrInvalidCredentials
	}

	token, err := uc.tokenManager.Generate(ctx, user.ID.String())
	if err != nil {
		log.Error("failed to generate token", "error", err)
		return "", err
	}

	log.Info("user logged in successfully", "user_id", user.ID.String())
	return token, nil
}
