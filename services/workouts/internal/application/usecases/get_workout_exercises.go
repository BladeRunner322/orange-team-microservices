package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type GetWorkoutExercises struct {
	workoutRepo  ports.WorkoutRepository
	exerciseRepo ports.WorkoutExerciseRepository
	logger       *logger.Logger
}

func NewGetWorkoutExercises(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	log *logger.Logger,
) *GetWorkoutExercises {
	return &GetWorkoutExercises{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
		logger:       log,
	}
}

func (uc *GetWorkoutExercises) Execute(ctx context.Context, workoutID uuid.UUID) ([]domain.WorkoutExercise, error) {
	// TODO:
	//  1. userID from ctx
	//  2. workout := workoutRepo.GetByID(ctx, userID, workoutID)
	//     - если ErrWorkoutNotFound → domain.ErrWorkoutNotFound
	//     - если workout.UserID() != userID → domain.ErrWorkoutNotFound
	//  3. if workout.IsDeleted() { return domain.ErrWorkoutNotFound }  // soft delete
	//  4. exercises := exerciseRepo.ListByWorkout(ctx, workoutID)
	//     - возвращает make([]domain.WorkoutExercise, 0) при пустом списке (не nil)
	return []domain.WorkoutExercise{}, nil
}
