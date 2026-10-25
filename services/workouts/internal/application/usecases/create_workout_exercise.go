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
	weight *domain.Weight,
	sets *domain.Sets,
	reps *domain.Reps,
	duration *domain.Duration,
	completed bool,
) (domain.WorkoutExercise, error) {
	log := uc.logger.With("operation", "CreateWorkoutExercise", "workout_id", workoutID, "exercise_id", exerciseID)

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

	if completed && workout.Status() != domain.StatusInProgress {
		log.Warn("cannot create completed exercise when workout is not in progress", "status", workout.Status())
		return domain.WorkoutExercise{}, domain.ErrCannotMarkExerciseCompleted
	}

	exerciseInfo, err := uc.exercisesClient.GetExercise(ctx, exerciseID)
	if err != nil {
		log.Error("failed to get exercise info", "error", err)
		return domain.WorkoutExercise{}, fmt.Errorf("get exercise info: %w", err)
	}

	exerciseType, err := domain.NewExerciseType(exerciseInfo.Type)
	if err != nil {
		log.Error("invalid exercise type from exercises service", "error", err, "type", exerciseInfo.Type)
		return domain.WorkoutExercise{}, fmt.Errorf("invalid exercise type: %w", err)
	}

	difficulty, err := domain.NewDifficulty(exerciseInfo.Difficulty)
	if err != nil {
		log.Error("invalid difficulty from exercises service", "error", err, "difficulty", exerciseInfo.Difficulty)
		return domain.WorkoutExercise{}, fmt.Errorf("invalid difficulty: %w", err)
	}

	workoutExercise, err := domain.NewWorkoutExercise(
		workoutID,
		exerciseID,
		difficulty,
		weight,
		sets,
		reps,
		duration,
		completed,
		exerciseType,
	)
	if err != nil {
		log.Warn("failed to construct workout exercise", "error", err)
		return domain.WorkoutExercise{}, err
	}

	var created domain.WorkoutExercise
	err = uc.workoutRepo.WithTx(ctx, func(wRepo ports.WorkoutRepository, eRepo ports.WorkoutExerciseRepository) error {
		var err error
		created, err = eRepo.Create(ctx, workoutExercise)
		if err != nil {
			return fmt.Errorf("create workout exercise: %w", err)
		}

		return recalculateScore(ctx, wRepo, eRepo, userID, workoutID)
	})
	if err != nil {
		log.Error("failed to create workout exercise in transaction", "error", err)
		return domain.WorkoutExercise{}, err
	}

	return created, nil
}
