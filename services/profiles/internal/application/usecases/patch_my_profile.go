package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

type PatchMyProfile struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewPatchMyProfile создаёт usecase для обновления профиля по user_id.
func NewPatchMyProfile(repo ports.Repository, log *logger.Logger) *PatchMyProfile {
	return &PatchMyProfile{repo: repo, logger: log}
}

// Execute применяет патч к профилю текущего пользователя.
// Если профиля нет — возвращает ErrProfileNotFound (сначала GET).
func (uc *PatchMyProfile) Execute(ctx context.Context, patch domain.ProfilePatch) (domain.Profile, error) {
	log := uc.logger.With("operation", "PatchMyProfile")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Profile{}, err
	}

	profile, err := uc.repo.GetByUserID(ctx, userID)

	if errors.Is(err, domain.ErrProfileNotFound) {
		log.Warn("profile not found", "user_id", userID)
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	if err != nil {
		log.Error("failed to get profile", "error", err, "user_id", userID)
		return domain.Profile{}, fmt.Errorf("get profile: %w", err)
	}

	profile.ApplyPatch(patch)

	updated, err := uc.repo.Update(ctx, profile)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			log.Warn("profile not found during update", "user_id", userID)
			return domain.Profile{}, domain.ErrProfileNotFound
		}

		log.Error("failed to update profile", "error", err, "user_id", userID)
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}

	log.Info("profile patched", "user_id", userID)

	return updated, nil
}
