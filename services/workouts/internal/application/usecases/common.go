package usecases

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/ctxkeys"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

// userIDFromContext достаёт user_id из context.
func userIDFromContext(ctx context.Context, log *logger.Logger) (uuid.UUID, error) {
	userIDStr, ok := ctxkeys.UserIDFromContext(ctx)
	if !ok {
		log.Warn("user_id is missing in context")
		return uuid.Nil, domain.ErrUnauthenticated
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn("invalid user_id format", "user_id", userIDStr, "error", err)
		return uuid.Nil, domain.ErrUnauthenticated
	}
	return userID, nil
}

// recalculateScore пересчитывает workout.workout_score по всем completed упражнениям.
// Должен вызываться внутри WithTx (ADR-025 п.3.1).
func recalculateScore(
	ctx context.Context,
	workoutRepo ports.WorkoutRepository,
	exerciseRepo ports.WorkoutExerciseRepository,
	userID uuid.UUID,
	workoutID uuid.UUID,
) error {
	workout, err := workoutRepo.GetByID(ctx, userID, workoutID)
	if err != nil {
		return fmt.Errorf("get workout for score: %w", err)
	}

	exercises, err := exerciseRepo.ListByWorkout(ctx, workoutID)
	if err != nil {
		return fmt.Errorf("list workout exercises for score: %w", err)
	}

	maximumEffectiveLoad := domain.DivideRoundUp(
		domain.MaximumScaledLoad,
		workout.PersonalScoreCoefficient(),
	)
	effectiveLoad := 0

	for _, ex := range exercises {
		if !ex.Completed() {
			continue
		}
		effectiveLoad = domain.AddLoadWithLimit(
			effectiveLoad,
			ex.ExerciseLoad(),
			ex.Difficulty().Int(),
			maximumEffectiveLoad,
		)
		if effectiveLoad == maximumEffectiveLoad {
			break
		}
	}

	score := domain.CalculateScore(effectiveLoad, workout.PersonalScoreCoefficient())
	workout.SetWorkoutScore(score)

	if _, err := workoutRepo.Update(ctx, workout); err != nil {
		return fmt.Errorf("update workout score: %w", err)
	}

	return nil
}
