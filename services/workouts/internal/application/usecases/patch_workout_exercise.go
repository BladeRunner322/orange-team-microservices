package usecases

import (
	"context"

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
	// TODO:
	//  1. userID from ctx
	//  2. workout := workoutRepo.GetByID; check ownership + CanModifyWorkoutExercise
	//  3. exercise := exerciseRepo.GetByID
	//  4. exerciseInfo := exercisesClient.GetExercise(ctx, exercise.ExerciseID())
	//  5. workoutRepo.WithTx -> exercise.ApplyPatch + exerciseRepo.Update + recalculateScore
	return domain.WorkoutExercise{}, nil
}
