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

type UncompleteHabit struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewUncompleteHabit(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *UncompleteHabit {
	return &UncompleteHabit{
		habitRepo:      habitRepo,
		completionRepo: completionRepo,
		logger:         log,
	}
}

// Execute снимает отметку выполнения.
// Проверки те же, что в CompleteHabit: не будущее, до 7 дней назад.
// DELETE идемпотентен — повторный вызов не падает.
func (uc *UncompleteHabit) Execute(ctx context.Context, id uuid.UUID, date time.Time) error {
	log := uc.logger.With("operation", "UncompleteHabit", "habit_id", id)

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

	if err := uc.completionRepo.Uncomplete(ctx, id, date); err != nil {
		log.Error("failed to uncomplete habit", "error", err)
		return fmt.Errorf("uncomplete habit: %w", err)
	}

	return nil
}
