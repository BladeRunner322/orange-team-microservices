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
	//  2. workout := workoutRepo.GetByID(ctx, userID, workoutID)
	//     - если ErrWorkoutNotFound → domain.ErrWorkoutNotFound
	//     - если workout.UserID() != userID → domain.ErrWorkoutNotFound (не раскрываем существование)
	//  3. if workout.IsDeleted() { return domain.ErrWorkoutNotFound }  // soft delete
	//  4. if !workout.CanModifyWorkoutExercise() { return domain.ErrCannotModifyWorkoutExercise }
	//  5. exercise := exerciseRepo.GetByID(ctx, workoutID, workoutExerciseID)
	//     - если ErrWorkoutExerciseNotFound → domain.ErrWorkoutExerciseNotFound
	//  6. exerciseInfo := exercisesClient.GetExercise(ctx, exercise.ExerciseID())
	//     - если ошибка → ошибка клиента (Internal / NotFound по контексту)
	//  7. exercise.ApplyPatch(patch, exerciseInfo.Type)
	//  8. workoutRepo.WithTx -> exerciseRepo.Update + recalculateScore
	return domain.WorkoutExercise{}, nil
}
