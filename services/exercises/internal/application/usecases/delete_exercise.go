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

type DeleteExercise struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewDeleteExercise создаёт usecase для soft-delete упражнения (admin-only).
func NewDeleteExercise(repo ports.Repository, log *logger.Logger) *DeleteExercise {
	return &DeleteExercise{repo: repo, logger: log}
}

// Execute помечает упражнение удалённым (deleted_at = NOW()).
// Возвращает ErrExerciseNotFound, если упражнение не найдено или уже удалено.
func (uc *DeleteExercise) Execute(ctx context.Context, id uuid.UUID) error {
	log := uc.logger.With("operation", "DeleteExercise", "exercise_id", id)

	exercise, err := uc.repo.GetExercise(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			log.Warn("exercise not found")
			return domain.ErrExerciseNotFound
		}

		log.Error("failed to get exercise", "error", err)
		return fmt.Errorf("get exercise: %w", err)
	}

	if exercise.IsDeleted() {
		log.Warn("exercise is deleted")
		return domain.ErrExerciseNotFound
	}

	if err := uc.repo.MarkDeleted(ctx, id); err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			log.Warn("exercise already deleted")
			return domain.ErrExerciseNotFound
		}

		log.Error("failed to mark deleted", "error", err)
		return fmt.Errorf("mark deleted: %w", err)
	}

	log.Info("exercise deleted")
	return nil
}
