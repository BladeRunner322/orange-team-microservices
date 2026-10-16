package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
)

// HabitWithStreak — привычка + её текущий стрик + флаг «выполнено сегодня».
// Возвращается из GetHabit и GetHabits, мапится в proto.
type HabitWithStreak struct {
	Habit          domain.Habit
	Streak         int
	CompletedToday bool
}

type GetHabits struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewGetHabits(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *GetHabits {
	return &GetHabits{
		habitRepo:      habitRepo,
		completionRepo: completionRepo,
		logger:         log,
	}
}

// Execute возвращает все активные привычки пользователя.
// Для каждой — стрик и флаг «выполнено сегодня».
func (uc *GetHabits) Execute(ctx context.Context) ([]HabitWithStreak, error) {
	log := uc.logger.With("operation", "GetHabits")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return nil, err
	}

	habits, err := uc.habitRepo.ListByUser(ctx, userID)
	if err != nil {
		log.Error("failed to list habits", "error", err)
		return nil, fmt.Errorf("list habits: %w", err)
	}

	now := time.Now().UTC()
	today := dayOnly(now)

	result := make([]HabitWithStreak, 0, len(habits))
	for _, habit := range habits {
		streak, err := uc.completionRepo.Streak(ctx, habit.ID(), today)
		if err != nil {
			log.Error("failed to calculate streak", "error", err, "habit_id", habit.ID())
			return nil, fmt.Errorf("calculate streak for habit %s: %w", habit.ID(), err)
		}

		completedToday, err := uc.completionRepo.IsCompleted(ctx, habit.ID(), today)
		if err != nil {
			log.Error("failed to check completion", "error", err, "habit_id", habit.ID())
			return nil, fmt.Errorf("check completion for habit %s: %w", habit.ID(), err)
		}

		result = append(result, HabitWithStreak{
			Habit:          habit,
			Streak:         streak,
			CompletedToday: completedToday,
		})
	}

	return result, nil
}
