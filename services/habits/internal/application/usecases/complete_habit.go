package usecases

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/google/uuid"
)

type CompleteHabit struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewCompleteHabit(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *CompleteHabit {
	// TODO
	return nil
}

// Execute отмечает выполнение привычки на дату.
// Проверяет:
//   - привычка существует, активна, принадлежит пользователю;
//   - date не в будущем;
//   - date не старше 7 дней назад.
//
// Повторная отметка на ту же дату — no-op (ON CONFLICT DO NOTHING).
func (uc *CompleteHabit) Execute(ctx context.Context, id uuid.UUID, date time.Time) error {
	// TODO
	return nil
}