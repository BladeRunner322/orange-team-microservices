package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

type PatchHabit struct {
	habitRepo ports.HabitRepository
	logger    *logger.Logger
}

func NewPatchHabit(habitRepo ports.HabitRepository, log *logger.Logger) *PatchHabit {
	return &PatchHabit{habitRepo: habitRepo, logger: log}
}

// Execute применяет патч к привычке.
// Обёрнут в habitRepo.WithTx (read-then-write, защита от lost update).
// Проверяет, что привычка принадлежит текущему пользователю.
func (uc *PatchHabit) Execute(ctx context.Context, id uuid.UUID, patch domain.HabitPatch) (domain.Habit, error) {
	log := uc.logger.With("operation", "PatchHabit", "habit_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Habit{}, err
	}

	var updated domain.Habit
	err = uc.habitRepo.WithTx(ctx, func(txRepo ports.HabitRepository) error {
		// SELECT ... FOR UPDATE — блокируем строку до COMMIT.
		// Защита от lost update (ADR-007 F-2).
		habit, err := txRepo.GetByIDForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrHabitNotFound) {
				log.Warn("habit not found")
				return domain.ErrHabitNotFound
			}
			log.Error("failed to get habit for update", "error", err)
			return fmt.Errorf("get habit for update: %w", err)
		}

		if habit.UserID() != userID {
			log.Warn("habit belongs to another user")
			return domain.ErrHabitNotFound
		}

		habit.ApplyPatch(patch)

		updated, err = txRepo.Update(ctx, habit)
		if err != nil {
			if errors.Is(err, domain.ErrHabitNotFound) {
				log.Warn("habit not found or already deleted")
				return domain.ErrHabitNotFound
			}
			if errors.Is(err, domain.ErrHabitNameExists) {
				log.Warn("habit name already exists", "name", habit.Name())
				return domain.ErrHabitNameExists
			}
			log.Error("failed to update habit", "error", err)
			return fmt.Errorf("update habit: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Habit{}, err
	}

	return updated, nil
}
