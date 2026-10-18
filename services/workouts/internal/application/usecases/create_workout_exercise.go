package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type CreateWorkoutExercise struct {
	workoutRepo     ports.WorkoutRepository
	exerciseRepo    ports.WorkoutExerciseRepository
	exercisesClient ports.ExercisesClient
	logger          *logger.Logger
}

func NewCreateWorkoutExercise(
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	exercisesClient ports.ExercisesClient,
	log *logger.Logger,
) *CreateWorkoutExercise {
	return &CreateWorkoutExercise{
		workoutRepo:     workoutRepo,
		exerciseRepo:    exerciseRepo,
		exercisesClient: exercisesClient,
		logger:          log,
	}
}

func (uc *CreateWorkoutExercise) Execute(
	ctx context.Context,
	workoutID uuid.UUID,
	exerciseID uuid.UUID,
	weight *int,
	sets *int,
	reps *int,
	duration *int,
	completed bool,
) (domain.WorkoutExercise, error) {
	// TODO:
	//  1. userID from ctx
	//  2. workout := workoutRepo.GetByID; check ownership + CanModifyWorkoutExercise
	//  3. exerciseInfo := exercisesClient.GetExercise(ctx, exerciseID)
	//  4. construct domain.NewWorkoutExercise
	//  5. workoutRepo.WithTx -> exerciseRepo.Create + recalculateScore
	return domain.WorkoutExercise{}, nil
}
