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

type PatchExercise struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewPatchExercise создаёт usecase для обновления упражнения (admin-only).
func NewPatchExercise(repo ports.Repository, log *logger.Logger) *PatchExercise {
	return &PatchExercise{repo: repo, logger: log}
}

// Execute применяет патч к упражнению по id.
// Поля патча: name, description, difficulty (type immutable — ADR-008).
// Возвращает ErrExerciseNotFound, если упражнение не найдено или удалено.
func (uc *PatchExercise) Execute(ctx context.Context, id uuid.UUID, patch domain.ExercisePatch) (domain.Exercise, error) {
	log := uc.logger.With("operation", "PatchExercise", "exercise_id", id)

	exercise, err := uc.repo.GetExercise(ctx, id)

	if err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			log.Warn("exercise not found")
			return domain.Exercise{}, domain.ErrExerciseNotFound
		}

		log.Error("failed to get exercise", "error", err)
		return domain.Exercise{}, fmt.Errorf("get exercise: %w", err)
	}

	exercise.ApplyPatch(patch)

	updated, err := uc.repo.Update(ctx, exercise)
	if err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			log.Warn("exercise not found or already deleted")
			return domain.Exercise{}, domain.ErrExerciseNotFound
		}

		if errors.Is(err, domain.ErrExerciseNameExists) {
			log.Warn("exercise name already exists", "name", exercise.Name())
			return domain.Exercise{}, domain.ErrExerciseNameExists
		}

		log.Error("failed to update exercise", "error", err)
		return domain.Exercise{}, fmt.Errorf("update exercise: %w", err)
	}

	return updated, nil
}
