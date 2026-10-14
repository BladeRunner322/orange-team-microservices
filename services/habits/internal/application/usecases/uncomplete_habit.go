package usecases

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/google/uuid"
)

type UncompleteHabit struct {
	habitRepo      ports.HabitRepository
	completionRepo ports.CompletionRepository
	logger         *logger.Logger
}

func NewUncompleteHabit(habitRepo ports.HabitRepository, completionRepo ports.CompletionRepository, log *logger.Logger) *UncompleteHabit {
	// TODO
	return nil
}

// Execute снимает отметку выполнения.
// Проверки те же, что в CompleteHabit: не будущее, до 7 дней назад.
// DELETE идемпотентен — повторный вызов не падает.
func (uc *UncompleteHabit) Execute(ctx context.Context, id uuid.UUID, date time.Time) error {
	// TODO
	return nil
}