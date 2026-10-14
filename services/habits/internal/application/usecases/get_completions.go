package usecases

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

type GetCompletions struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewGetCompletions(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *GetCompletions {
	// TODO
	return nil
}

// Execute возвращает историю выполнений привычки за период [from, to].
// Если период не задан (zero time) — используются последние 30 дней.
// Проверяет, что привычка принадлежит текущему пользователю.
func (uc *GetCompletions) Execute(ctx context.Context, id uuid.UUID, from, to time.Time) ([]domain.Completion, error) {
	// TODO
	return nil, nil
}