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
	repo         ports.Repository
	transactions ports.TransactionManager
	logger       *logger.Logger
}

// NewPatchMyProfile создаёт сценарий обновления профиля текущего пользователя.
func NewPatchMyProfile(repo ports.Repository, transactions ports.TransactionManager, log *logger.Logger) *PatchMyProfile {
	return &PatchMyProfile{
		repo:         repo,
		transactions: transactions,
		logger:       log,
	}
}

// Execute применяет патч к профилю текущего пользователя.
// Если профиля нет — возвращает ErrProfileNotFound (сначала GET).
func (uc *PatchMyProfile) Execute(ctx context.Context, patch domain.ProfilePatch) (domain.Profile, error) {
	log := uc.logger.With("operation", "PatchMyProfile")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Profile{}, err
	}

	var updated domain.Profile
	err = uc.transactions.Do(ctx, func(txCtx context.Context) error {
		// SELECT ... FOR UPDATE блокирует строку до завершения транзакции.
		// Защита от потерянного обновления (ADR-007 F-2).
		profile, err := uc.repo.GetByUserIDForUpdate(txCtx, userID)
		if err != nil {
			if errors.Is(err, domain.ErrProfileNotFound) {
				log.Warn("profile not found", "user_id", userID)
				return domain.ErrProfileNotFound
			}
			log.Error("failed to get profile for update", "error", err, "user_id", userID)
			return fmt.Errorf("get profile for update: %w", err)
		}

		profile.ApplyPatch(patch)

		updated, err = uc.repo.Update(txCtx, profile)
		if err != nil {
			if errors.Is(err, domain.ErrProfileNotFound) {
				log.Warn("profile not found during update", "user_id", userID)
				return domain.ErrProfileNotFound
			}
			log.Error("failed to update profile", "error", err, "user_id", userID)
			return fmt.Errorf("update profile: %w", err)
		}

		return nil
	})

	if err != nil {
		return domain.Profile{}, err
	}

	return updated, nil
}
