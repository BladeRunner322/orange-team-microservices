package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type DeleteWorkoutExercise struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewDeleteWorkoutExercise(
	workoutRepo ports.WorkoutRepository,
	log *logger.Logger,
) *DeleteWorkoutExercise {
	return &DeleteWorkoutExercise{
		workoutRepo: workoutRepo,
		logger:      log,
	}
}

// Execute удаляет упражнение из тренировки и пересчитывает её балл.
func (uc *DeleteWorkoutExercise) Execute(ctx context.Context, workoutID, workoutExerciseID uuid.UUID) error {
	log := uc.logger.With(
		"operation", "DeleteWorkoutExercise",
		"workout_id", workoutID,
		"workout_exercise_id", workoutExerciseID,
	)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return err
	}

	err = uc.workoutRepo.WithTx(ctx, func(wRepo ports.WorkoutRepository, eRepo ports.WorkoutExerciseRepository) error {
		workout, err := wRepo.GetByID(ctx, userID, workoutID)
		if err != nil {
			return fmt.Errorf("get workout: %w", err)
		}
		if workout.IsDeleted() {
			return domain.ErrWorkoutNotFound
		}
		if !workout.CanModifyWorkoutExercise() {
			return domain.ErrCannotModifyWorkoutExercise
		}

		if err := eRepo.Delete(ctx, workoutID, workoutExerciseID); err != nil {
			if errors.Is(err, domain.ErrWorkoutExerciseNotFound) {
				return nil
			}
			return fmt.Errorf("delete workout exercise: %w", err)
		}

		return recalculateWorkoutScore(ctx, wRepo, eRepo, workout)
	})
	if err != nil {
		log.Error("failed to delete workout exercise", "error", err)
		return err
	}

	return nil
}
