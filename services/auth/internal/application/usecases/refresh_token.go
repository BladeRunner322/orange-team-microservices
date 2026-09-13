package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type RefreshToken struct {
	refreshRepo     ports.RefreshTokenRepository
	userRepo        ports.Repository
	tokenManager    ports.TokenManager
	refreshTokenTTL time.Duration
	logger          *logger.Logger
}

func NewRefreshToken(
	refreshRepo ports.RefreshTokenRepository,
	userRepo ports.Repository,
	tokenManager ports.TokenManager,
	refreshTokenTTL time.Duration,
	log *logger.Logger,
) *RefreshToken {
	return &RefreshToken{
		refreshRepo:     refreshRepo,
		userRepo:        userRepo,
		tokenManager:    tokenManager,
		refreshTokenTTL: refreshTokenTTL,
		logger:          log,
	}
}

// Result — результат обновления токенов.
type Result struct {
	AccessToken  string
	RefreshToken string
}

func (uc *RefreshToken) Execute(ctx context.Context, rawRefreshToken string) (Result, error) {
	log := uc.logger.With("operation", "RefreshToken")

	refreshToken, err := domain.NewRefreshToken(rawRefreshToken)
	if err != nil {
		log.Warn("invalid refresh token format")
		return Result{}, domain.ErrInvalidRefreshToken
	}

	// 1. Проверяем, что токен существует и получаем userID
	userID, err := uc.refreshRepo.GetUserID(ctx, refreshToken)
	if err != nil {
		log.Warn("refresh token not found")
		return Result{}, domain.ErrInvalidRefreshToken
	}

	// 2. Читаем пользователя, чтобы взять актуальную роль
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Warn("invalid user id in refresh token", "user_id", userID)
		return Result{}, domain.ErrInvalidRefreshToken
	}
	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		log.Error("failed to find user for refresh", "error", err, "user_id", userID)
		return Result{}, fmt.Errorf("find user for refresh: %w", err)
	}

	// 3. Удаляем старый (rotation)
	if err := uc.refreshRepo.Delete(ctx, refreshToken); err != nil {
		log.Error("failed to delete old refresh token", "error", err)
		return Result{}, fmt.Errorf("delete old refresh token: %w", err)
	}

	// 4. Генерируем новую пару
	accessToken, err := uc.tokenManager.Generate(ctx, user.ID().String(), user.Role().String())
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return Result{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := domain.GenerateRefreshToken()
	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return Result{}, fmt.Errorf("generate refresh token: %w", err)
	}

	// 5. Сохраняем новый refresh
	if err := uc.refreshRepo.Save(ctx, newRefreshToken, userID, uc.refreshTokenTTL); err != nil {
		log.Error("failed to save new refresh token", "error", err)
		return Result{}, fmt.Errorf("save refresh token: %w", err)
	}

	log.Info("refresh token rotated", "user_id", userID)
	return Result{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.String(),
	}, nil
}
