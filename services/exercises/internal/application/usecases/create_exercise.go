// Package usecases реализует сценарии работы с упражнениями.
package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
)

type CreateExercise struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewCreateExercise создаёт usecase для добавления упражнения (admin-only).
func NewCreateExercise(repo ports.Repository, log *logger.Logger) *CreateExercise {
	return &CreateExercise{repo: repo, logger: log}
}

// Execute валидирует поля, создаёт упражнение и сохраняет в БД.
// Возвращает ErrExerciseNameExists, если name уже занят.
func (uc *CreateExercise) Execute(ctx context.Context, nameRaw string, descriptionRaw string, difficultyRaw int, exerciseTypeRaw string) (domain.Exercise, error) {
	log := uc.logger.With("operation", "CreateExercise")

	name, err := domain.NewName(nameRaw)
	if err != nil {
		log.Warn("invalid name", "error", err, "raw", nameRaw)
		return domain.Exercise{}, err
	}

	description, err := domain.NewDescription(descriptionRaw)
	if err != nil {
		log.Warn("invalid description", "error", err, "raw", descriptionRaw)
		return domain.Exercise{}, err
	}

	difficulty, err := domain.NewDifficulty(difficultyRaw)
	if err != nil {
		log.Warn("invalid difficulty", "error", err, "raw", difficultyRaw)
		return domain.Exercise{}, err
	}

	exerciseType, err := domain.NewExerciseType(exerciseTypeRaw)
	if err != nil {
		log.Warn("invalid exercise_type", "error", err, "raw", exerciseTypeRaw)
		return domain.Exercise{}, err
	}

	exercise := domain.NewExercise(name, description, difficulty, exerciseType)

	created, err := uc.repo.Create(ctx, exercise)
	if err != nil {
		if errors.Is(err, domain.ErrExerciseNameExists) {
			log.Warn("exercise name already exists", "name", name)
			return domain.Exercise{}, domain.ErrExerciseNameExists
		}

		log.Error("failed to create exercise", "error", err)
		return domain.Exercise{}, fmt.Errorf("create exercise: %w", err)
	}

	return created, nil
}
