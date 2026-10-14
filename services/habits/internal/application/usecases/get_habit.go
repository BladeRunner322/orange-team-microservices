package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/google/uuid"
)

type GetHabit struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewGetHabit(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *GetHabit {
	// TODO
	return nil
}

// Execute возвращает привычку по id + стрик + флаг «выполнено сегодня».
// Проверяет, что привычка принадлежит текущему пользователю.
func (uc *GetHabit) Execute(ctx context.Context, id uuid.UUID) (HabitWithStreak, error) {
	// TODO
	return HabitWithStreak{}, nil
}
