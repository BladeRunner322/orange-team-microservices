package usecases

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type Login struct {
	repo            ports.Repository
	tokenManager    ports.TokenManager
	refreshRepo     ports.RefreshTokenRepository
	refreshTokenTTL time.Duration
	logger          *logger.Logger
}

func NewLogin(
	repo ports.Repository,
	tokenManager ports.TokenManager,
	refreshRepo ports.RefreshTokenRepository,
	refreshTokenTTL time.Duration,
	log *logger.Logger,
) *Login {
	return &Login{
		repo:            repo,
		tokenManager:    tokenManager,
		refreshRepo:     refreshRepo,
		refreshTokenTTL: refreshTokenTTL,
		logger:          log,
	}
}

// LoginResult — результат успешного логина.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

func (uc *Login) Execute(ctx context.Context, emailStr, password string) (LoginResult, error) {
	log := uc.logger.With("email", emailStr, "operation", "Login")

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		log.Warn("invalid email", "error", err)
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		log.Warn("user not found")
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash().String()),
		[]byte(password),
	); err != nil {
		log.Warn("invalid password")
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	// Генерируем access token
	accessToken, err := uc.tokenManager.Generate(ctx, user.ID().String())
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	// Генерируем refresh token
	refreshToken, err := domain.GenerateRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return LoginResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	// Сохраняем refresh в Redis
	if err := uc.refreshRepo.Save(ctx, refreshToken, user.ID().String(), uc.refreshTokenTTL); err != nil {
		log.Error("failed to save refresh token", "error", err)
		return LoginResult{}, fmt.Errorf("save refresh token: %w", err)
	}

	log.Info("user logged in successfully", "user_id", user.ID().String())
	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.String(),
	}, nil
}
