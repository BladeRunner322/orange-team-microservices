package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

type CompleteHabit struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewCompleteHabit(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *CompleteHabit {
	return &CompleteHabit{
		habitRepo:      habitRepo,
		completionRepo: completionRepo,
		logger:         log,
	}
}

// Execute отмечает выполнение привычки на дату.
// Проверяет:
//   - привычка существует, активна, принадлежит пользователю;
//   - date не в будущем;
//   - date не старше 7 дней назад.
//
// Повторная отметка на ту же дату — no-op (ON CONFLICT DO NOTHING).
func (uc *CompleteHabit) Execute(ctx context.Context, id uuid.UUID, date time.Time) error {
	log := uc.logger.With("operation", "CompleteHabit", "habit_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return err
	}

	habit, err := uc.habitRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrHabitNotFound) {
			log.Warn("habit not found")
			return domain.ErrHabitNotFound
		}

		log.Error("failed to get habit", "error", err)
		return fmt.Errorf("get habit: %w", err)
	}

	if habit.UserID() != userID {
		log.Warn("habit belongs to another user")
		return domain.ErrHabitNotFound
	}

	if habit.IsDeleted() {
		log.Warn("habit is deleted")
		return domain.ErrHabitNotFound
	}

	if err := validateCompletionDate(date, time.Now().UTC()); err != nil {
		log.Warn("invalid completion date", "date", date.Format("2006-01-02"))
		return domain.ErrInvalidDate
	}

	completion := domain.NewCompletion(id, mustDate(date))

	if err := uc.completionRepo.Complete(ctx, completion); err != nil {
		log.Error("failed to complete habit", "error", err)
		return fmt.Errorf("complete habit: %w", err)
	}

	return nil
}
