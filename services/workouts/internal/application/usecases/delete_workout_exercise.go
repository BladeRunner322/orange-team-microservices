package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/google/uuid"
)

type DeleteWorkoutExercise struct {
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.WorkoutExerciseRepository
	logger       *logger.Logger
}

func NewDeleteWorkoutExercise(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	log *logger.Logger,
) *DeleteWorkoutExercise {
	return &DeleteWorkoutExercise{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
		logger:       log,
	}
}

func (uc *DeleteWorkoutExercise) Execute(ctx context.Context, workoutID, workoutExerciseID uuid.UUID) error {
	// TODO:
	//  1. userID from ctx
	//  2. workout := workoutRepo.GetByID(ctx, userID, workoutID)
	//     - если ErrWorkoutNotFound → domain.ErrWorkoutNotFound
	//     - если workout.UserID() != userID → domain.ErrWorkoutNotFound
	//  3. if workout.IsDeleted() { return domain.ErrWorkoutNotFound }  // soft delete
	//  4. if !workout.CanModifyWorkoutExercise() { return domain.ErrCannotModifyWorkoutExercise }
	//  5. workoutRepo.WithTx -> exerciseRepo.Delete + recalculateScore
	return nil
}
