package usecases

import (
	"context"

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
	// TODO:
	//  1. workout := workoutRepo.GetByID(ctx, userID, workoutID)
	//  2. exercises := exerciseRepo.ListByWorkout(ctx, workoutID)
	//  3. effectiveLoad = Σ AddLoadWithLimit(...) по completed
	//  4. score := CalculateScore(effectiveLoad, workout.PersonalScoreCoefficient())
	//  5. workout.SetWorkoutScore(score)
	//  6. workoutRepo.Update(ctx, workout)
	return nil
}
