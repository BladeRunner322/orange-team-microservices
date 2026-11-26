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

type PatchWorkout struct {
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.WorkoutExerciseRepository
	transactions ports.TransactionManager
	logger       *logger.Logger
}

func NewPatchWorkout(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	transactions ports.TransactionManager,
	log *logger.Logger,
) *PatchWorkout {
	return &PatchWorkout{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
		transactions: transactions,
		logger:       log,
	}
}

func (uc *PatchWorkout) Execute(ctx context.Context, id uuid.UUID, patch domain.WorkoutPatch) (domain.Workout, error) {
	log := uc.logger.With("operation", "PatchWorkout", "workout_id", id)

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Workout{}, err
	}

	var updated domain.Workout
	err = uc.transactions.Do(ctx, func(txCtx context.Context) error {
		workout, err := uc.workoutRepo.GetByID(txCtx, userID, id)
		if err != nil {
			if errors.Is(err, domain.ErrWorkoutNotFound) {
				log.Warn("workout not found")
				return domain.ErrWorkoutNotFound
			}
			log.Error("failed to get workout", "error", err)
			return fmt.Errorf("get workout: %w", err)
		}

		if workout.IsDeleted() {
			log.Warn("workout is deleted")
			return domain.ErrWorkoutNotFound
		}

		if err := workout.ApplyPatch(patch); err != nil {
			log.Warn("failed to apply patch to workout", "error", err)
			return err
		}

		if workout.Status() == domain.StatusCompleted {
			exercises, err := uc.exerciseRepo.ListByWorkout(txCtx, id)
			if err != nil {
				log.Error("failed to list exercises for completion check", "error", err)
				return fmt.Errorf("list exercises: %w", err)
			}

			hasCompleted := false
			for _, ex := range exercises {
				if ex.Completed() {
					hasCompleted = true
					break
				}
			}
			if !hasCompleted {
				log.Warn("cannot complete workout with no completed exercises")
				return domain.ErrCannotCompleteEmptyWorkout
			}
		}

		updated, err = uc.workoutRepo.Update(txCtx, workout)
		if err != nil {
			log.Error("failed to update workout", "error", err)
			return fmt.Errorf("update workout: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.Workout{}, err
	}

	return updated, nil
}
