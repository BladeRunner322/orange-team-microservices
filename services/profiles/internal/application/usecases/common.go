package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
)

// getOrCreateProfile возвращает профиль пользователя или создаёт пустой,
// если его ещё нет (lazy-create, см. ADR-001).
func getOrCreateProfile(
	ctx context.Context,
	repo ports.Repository,
	log *logger.Logger,
	userID uuid.UUID,
) (domain.Profile, error) {
	profile, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, domain.ErrProfileNotFound) {
			log.Error("failed to get profile", "error", err, "user_id", userID)
			return domain.Profile{}, fmt.Errorf("get profile: %w", err)
		}

		log.Info("profile not found, creating empty", "user_id", userID)
		profile = domain.NewEmptyProfile(userID)
		if err := repo.Upsert(ctx, profile); err != nil {
			log.Error("failed to upsert empty profile", "error", err, "user_id", userID)
			return domain.Profile{}, fmt.Errorf("upsert empty profile: %w", err)
		}
		return profile, nil
	}

	return profile, nil
}
