package usecases

import (
	"context"

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
	// TODO
	return nil
}

// Execute возвращает все активные привычки пользователя.
// Для каждой — стрик и флаг «выполнено сегодня».
func (uc *GetHabits) Execute(ctx context.Context) ([]HabitWithStreak, error) {
	// TODO
	return nil, nil
}