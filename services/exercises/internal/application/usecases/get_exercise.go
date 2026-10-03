package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"github.com/google/uuid"
)

type GetExercise struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewGetExercise создаёт usecase для получения упражнения по id.
func NewGetExercise(repo ports.Repository, log *logger.Logger) *GetExercise {
	return &GetExercise{repo: repo, logger: log}
}

// Execute возвращает упражнение по id.
func (uc *GetExercise) Execute(ctx context.Context, id uuid.UUID) (domain.Exercise, error) {
	log := uc.logger.With("operation", "GetExercise", "exercise_id", id)

	exercise, err := uc.repo.GetExercise(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			log.Warn("exercise not found")
			return domain.Exercise{}, domain.ErrExerciseNotFound
		}

		log.Error("failed to get exercise", "error", err)
		return domain.Exercise{}, fmt.Errorf("get exercise: %w", err)
	}

	log.Info("exercise fetched")

	return exercise, nil
}
