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

type GetHabit struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewGetHabit(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *GetHabit {
	return &GetHabit{
		habitRepo:      habitRepo,
		completionRepo: completionRepo,
		logger:         log,
	}
}

// Execute возвращает привычку по id + стрик + флаг «выполнено сегодня».
// Проверяет, что привычка принадлежит текущему пользователю:
// чужая привычка → ErrHabitNotFound (существование не раскрываем).
func (uc *GetHabit) Execute(ctx context.Context, id uuid.UUID) (HabitWithStreak, error) {
	log := uc.logger.With("operation", "GetHabit", "habit_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return HabitWithStreak{}, err
	}

	habit, err := uc.habitRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrHabitNotFound) {
			log.Warn("habit not found")
			return HabitWithStreak{}, domain.ErrHabitNotFound
		}

		log.Error("failed to get habit", "error", err)
		return HabitWithStreak{}, fmt.Errorf("get habit: %w", err)
	}

	if habit.UserID() != userID {
		log.Warn("habit belongs to another user")
		return HabitWithStreak{}, domain.ErrHabitNotFound
	}

	now := time.Now().UTC()
	today := dayOnly(now)

	streak, err := uc.completionRepo.Streak(ctx, habit.ID(), today)
	if err != nil {
		log.Error("failed to calculate streak", "error", err)
		return HabitWithStreak{}, fmt.Errorf("calculate streak: %w", err)
	}

	completedToday, err := uc.completionRepo.IsCompleted(ctx, habit.ID(), today)
	if err != nil {
		log.Error("failed to check completion", "error", err)
		return HabitWithStreak{}, fmt.Errorf("check completion: %w", err)
	}

	result := HabitWithStreak{
		Habit:          habit,
		Streak:         streak,
		CompletedToday: completedToday,
	}

	return result, nil
}
