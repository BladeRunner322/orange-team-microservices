package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

type DeleteMyProfile struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewDeleteMyProfile создаёт usecase для удаления профиля по user_id.
func NewDeleteMyProfile(repo ports.Repository, log *logger.Logger) *DeleteMyProfile {
	return &DeleteMyProfile{repo: repo, logger: log}
}

// Execute удаляет профиль текущего пользователя.
// Если профиля нет — возвращает ErrProfileNotFound.
func (uc *DeleteMyProfile) Execute(ctx context.Context) error {
	log := uc.logger.With("operation", "DeleteMyProfile")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			log.Warn("profile not found", "user_id", userID)
			return domain.ErrProfileNotFound
		}

		log.Error("failed to delete profile", "error", err, "user_id", userID)
		return fmt.Errorf("delete profile: %w", err)
	}

	log.Info("profile deleted", "user_id", userID)

	return nil
}
