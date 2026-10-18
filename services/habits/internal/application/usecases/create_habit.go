// Package usecases реализует сценарии работы Habits-сервиса.
package usecases

import (
	"context"
	"errors"
	"fmt"

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
	log := uc.logger.With("operation", "CreateHabit")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Habit{}, err
	}

	name, err := domain.NewName(nameRaw)
	if err != nil {
		log.Warn("invalid name", "error", err)
		return domain.Habit{}, err
	}

	description, err := domain.NewDescription(descriptionRaw)
	if err != nil {
		log.Warn("invalid description", "error", err)
		return domain.Habit{}, err
	}

	habit := domain.NewHabit(userID, name, description)

	created, err := uc.habitRepo.Create(ctx, habit)
	if err != nil {
		if errors.Is(err, domain.ErrHabitNameExists) {
			log.Warn("habit name already exists", "name", name)
			return domain.Habit{}, domain.ErrHabitNameExists
		}

		log.Error("failed to create habit", "error", err)
		return domain.Habit{}, fmt.Errorf("create habit: %w", err)
	}

	return created, nil
}
