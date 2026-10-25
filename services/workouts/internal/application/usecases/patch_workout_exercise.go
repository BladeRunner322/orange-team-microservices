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

type PatchWorkoutExercise struct {
	workoutRepo     ports.WorkoutRepository
	exerciseRepo    ports.WorkoutExerciseRepository
	exercisesClient ports.ExercisesClient
	logger          *logger.Logger
}

func NewPatchWorkoutExercise(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	exercisesClient ports.ExercisesClient,
	log *logger.Logger,
) *PatchWorkoutExercise {
	return &PatchWorkoutExercise{
		workoutRepo:     workoutRepo,
		exerciseRepo:    exerciseRepo,
		exercisesClient: exercisesClient,
		logger:          log,
	}
}

func (uc *PatchWorkoutExercise) Execute(
	ctx context.Context,
	workoutID uuid.UUID,
	workoutExerciseID uuid.UUID,
	patch domain.WorkoutExercisePatch,
) (domain.WorkoutExercise, error) {
	log := uc.logger.With("operation", "PatchWorkoutExercise", "workout_id", workoutID, "workout_exercise_id", workoutExerciseID)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.WorkoutExercise{}, err
	}

	workout, err := uc.workoutRepo.GetByID(ctx, userID, workoutID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkoutNotFound) {
			log.Warn("workout not found")
			return domain.WorkoutExercise{}, domain.ErrWorkoutNotFound
		}
		log.Error("failed to get workout", "error", err)
		return domain.WorkoutExercise{}, fmt.Errorf("get workout: %w", err)
	}

	if workout.IsDeleted() {
		log.Warn("workout is deleted")
		return domain.WorkoutExercise{}, domain.ErrWorkoutNotFound
	}

	if !workout.CanModifyWorkoutExercise() {
		log.Warn("cannot modify workout exercise in terminal status", "status", workout.Status())
		return domain.WorkoutExercise{}, domain.ErrCannotModifyWorkoutExercise
	}

	if patch.Completed != nil && *patch.Completed && workout.Status() != domain.StatusInProgress {
		log.Warn("cannot mark exercise as completed when workout is not in progress", "status", workout.Status())
		return domain.WorkoutExercise{}, domain.ErrCannotMarkExerciseCompleted
	}

	exercise, err := uc.exerciseRepo.GetByID(ctx, workoutID, workoutExerciseID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkoutExerciseNotFound) {
			log.Warn("workout exercise not found")
			return domain.WorkoutExercise{}, domain.ErrWorkoutExerciseNotFound
		}
		log.Error("failed to get workout exercise", "error", err)
		return domain.WorkoutExercise{}, fmt.Errorf("get workout exercise: %w", err)
	}

	exerciseInfo, err := uc.exercisesClient.GetExercise(ctx, exercise.ExerciseID())
	if err != nil {
		log.Error("failed to get exercise info", "error", err)
		return domain.WorkoutExercise{}, fmt.Errorf("get exercise info: %w", err)
	}

	exerciseType, err := domain.NewExerciseType(exerciseInfo.Type)
	if err != nil {
		log.Error("invalid exercise type from exercises service", "error", err, "type", exerciseInfo.Type)
		return domain.WorkoutExercise{}, fmt.Errorf("invalid exercise type: %w", err)
	}

	if err := exercise.ApplyPatch(patch, exerciseType); err != nil {
		log.Warn("failed to apply patch to workout exercise", "error", err)
		return domain.WorkoutExercise{}, err
	}

	var updated domain.WorkoutExercise
	err = uc.workoutRepo.WithTx(ctx, func(wRepo ports.WorkoutRepository, eRepo ports.WorkoutExerciseRepository) error {
		var err error
		updated, err = eRepo.Update(ctx, exercise)
		if err != nil {
			return fmt.Errorf("update workout exercise: %w", err)
		}

		return recalculateScore(ctx, wRepo, eRepo, userID, workoutID)
	})
	if err != nil {
		log.Error("failed to patch workout exercise in transaction", "error", err)
		return domain.WorkoutExercise{}, err
	}

	return updated, nil
}
