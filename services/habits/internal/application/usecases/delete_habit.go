package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
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
func (uc *DeleteHabit) Execute(ctx context.Context, id uuid.UUID) error {
	// TODO: по образцу DeleteExercise
	return nil
}