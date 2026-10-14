package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"github.com/google/uuid"
)

type PatchHabit struct {
	habitRepo ports.HabitRepository
	logger    *logger.Logger
}

func NewPatchHabit(habitRepo ports.HabitRepository, log *logger.Logger) *PatchHabit {
	return &PatchHabit{habitRepo: habitRepo, logger: log}
}

// Execute применяет патч к привычке.
// Обёрнут в habitRepo.WithTx (read-then-write, защита от lost update).
// Проверяет, что привычка принадлежит текущему пользователю.
func (uc *PatchHabit) Execute(ctx context.Context, id uuid.UUID, patch domain.HabitPatch) (domain.Habit, error) {
	// TODO: по образцу PatchExercise в Exercises
	return domain.Habit{}, nil
}