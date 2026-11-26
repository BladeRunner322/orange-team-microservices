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
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.WorkoutExerciseRepository
	transactions ports.TransactionManager
	logger       *logger.Logger
}

func NewDeleteWorkoutExercise(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	transactions ports.TransactionManager,
	log *logger.Logger,
) *DeleteWorkoutExercise {
	return &DeleteWorkoutExercise{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
		transactions: transactions,
		logger:       log,
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

	err = uc.transactions.Do(ctx, func(txCtx context.Context) error {
		workout, err := uc.workoutRepo.GetByID(txCtx, userID, workoutID)
		if err != nil {
			return fmt.Errorf("get workout: %w", err)
		}
		if workout.IsDeleted() {
			return domain.ErrWorkoutNotFound
		}
		if !workout.CanModifyWorkoutExercise() {
			return domain.ErrCannotModifyWorkoutExercise
		}

		if err := uc.exerciseRepo.Delete(txCtx, workoutID, workoutExerciseID); err != nil {
			if errors.Is(err, domain.ErrWorkoutExerciseNotFound) {
				return nil
			}
			return fmt.Errorf("delete workout exercise: %w", err)
		}

		err = recalculateWorkoutScore(txCtx, uc.workoutRepo, uc.exerciseRepo, workout)
		return err
	})
	if err != nil {
		log.Error("failed to delete workout exercise", "error", err)
		return err
	}

	return nil
}
