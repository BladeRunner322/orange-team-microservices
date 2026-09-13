package usecases

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

type Logout struct {
	refreshRepo ports.RefreshTokenRepository
	logger      *logger.Logger
}

func NewLogout(refreshRepo ports.RefreshTokenRepository, log *logger.Logger) *Logout {
	return &Logout{
		refreshRepo: refreshRepo,
		logger:      log,
	}
}

func (uc *Logout) Execute(ctx context.Context, rawRefreshToken string) error {
	log := uc.logger.With("operation", "Logout")

	refreshToken, err := domain.NewRefreshToken(rawRefreshToken)
	if err != nil {
		log.Warn("invalid refresh token format")
		return domain.ErrInvalidRefreshToken
	}

	if err := uc.refreshRepo.Delete(ctx, refreshToken); err != nil {
		log.Error("failed to delete refresh token", "error", err)
		return fmt.Errorf("delete refresh token: %w", err)
	}

	log.Info("user logged out")
	return nil
}
