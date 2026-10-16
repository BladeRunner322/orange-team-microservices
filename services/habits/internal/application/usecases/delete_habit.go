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

type DeleteHabit struct {
	habitRepo ports.HabitRepository
	logger    *logger.Logger
}

func NewDeleteHabit(habitRepo ports.HabitRepository, log *logger.Logger) *DeleteHabit {
	return &DeleteHabit{habitRepo: habitRepo, logger: log}
}

// Execute помечает привычку удалённой (soft delete).
//
// Идемпотентен: повторный вызов или DELETE несуществующего id
// возвращает nil, а не ошибку (см. ADR-007 F-3).
// Completions остаются в БД — если привычку восстановить, история на месте.
//
// Проверка владельца: чужая привычка → ErrHabitNotFound.
func (uc *DeleteHabit) Execute(ctx context.Context, id uuid.UUID) error {
	log := uc.logger.With("operation", "DeleteHabit", "habit_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return err
	}

	habit, err := uc.habitRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrHabitNotFound) {
			// Идемпотентность: привычки нет — считаем успехом.
			log.Warn("habit not found or already deleted")
			return nil
		}

		log.Error("failed to get habit", "error", err)
		return fmt.Errorf("get habit: %w", err)
	}

	if habit.UserID() != userID {
		log.Warn("habit belongs to another user")
		return domain.ErrHabitNotFound
	}

	if err := uc.habitRepo.MarkDeleted(ctx, id); err != nil {
		if errors.Is(err, domain.ErrHabitNotFound) {
			// Гонка: между GetByID и MarkDeleted кто-то удалил.
			log.Warn("habit already deleted")
			return nil
		}

		log.Error("failed to mark deleted", "error", err)
		return fmt.Errorf("mark deleted: %w", err)
	}

	return nil
}
