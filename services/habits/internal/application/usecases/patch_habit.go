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
	habitRepo    ports.HabitRepository
	transactions ports.TransactionManager
	logger       *logger.Logger
}

func NewPatchHabit(habitRepo ports.HabitRepository, transactions ports.TransactionManager, log *logger.Logger) *PatchHabit {
	return &PatchHabit{
		habitRepo:    habitRepo,
		transactions: transactions,
		logger:       log,
	}
}

// Execute обновляет привычку текущего пользователя в одной транзакции.
func (uc *PatchHabit) Execute(ctx context.Context, id uuid.UUID, patch domain.HabitPatch) (domain.Habit, error) {
	log := uc.logger.With("operation", "PatchHabit", "habit_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Habit{}, err
	}

	var updated domain.Habit
	err = uc.transactions.Do(ctx, func(txCtx context.Context) error {
		// SELECT ... FOR UPDATE блокирует строку до завершения транзакции.
		// Защита от потерянного обновления (ADR-007 F-2).
		habit, err := uc.habitRepo.GetByIDForUpdate(txCtx, id)
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

		updated, err = uc.habitRepo.Update(txCtx, habit)
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
