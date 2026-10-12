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
//
// Идемпотентен: повторный вызов (или DELETE несуществующего профиля)
// возвращает nil, а не ошибку (см. ADR-007 F-3). Это защищает от
// retry после network timeout — клиент может безопасно повторить запрос.
func (uc *DeleteMyProfile) Execute(ctx context.Context) error {
	log := uc.logger.With("operation", "DeleteMyProfile")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			// Идемпотентность: профиля нет — считаем успехом.
			log.Warn("profile already deleted", "user_id", userID)
			return nil
		}

		log.Error("failed to delete profile", "error", err, "user_id", userID)
		return fmt.Errorf("delete profile: %w", err)
	}

	return nil
}
