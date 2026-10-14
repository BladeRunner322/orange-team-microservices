// Package usecases реализует сценарии работы Habits-сервиса.
package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
)

type CreateHabit struct {
	habitRepo ports.HabitRepository
	logger    *logger.Logger
}

func NewCreateHabit(habitRepo ports.HabitRepository, log *logger.Logger) *CreateHabit {
	return &CreateHabit{habitRepo: habitRepo, logger: log}
}

// Execute валидирует поля через VO, создаёт привычку.
// user_id берётся из context (ставит Gateway через metadata).
func (uc *CreateHabit) Execute(ctx context.Context, nameRaw, descriptionRaw string) (domain.Habit, error) {
	// TODO:
	//   1. userID := userIDFromContext(ctx)
	//   2. name, err := domain.NewName(nameRaw)
	//   3. description, err := domain.NewDescription(descriptionRaw)
	//   4. habit := domain.NewHabit(userID, name, description)
	//   5. return uc.habitRepo.Create(ctx, habit)
	return domain.Habit{}, nil
}