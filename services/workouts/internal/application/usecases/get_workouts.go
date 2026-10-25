package usecases

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

type GetWorkouts struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewGetWorkouts(repo ports.WorkoutRepository, log *logger.Logger) *GetWorkouts {
	return &GetWorkouts{workoutRepo: repo, logger: log}
}

// Execute возвращает список тренировок пользователя (только активные,
// без soft-deleted — фильтр deleted_at IS NULL в репозитории).
func (uc *GetWorkouts) Execute(ctx context.Context) ([]domain.Workout, error) {
	log := uc.logger.With("operation", "GetWorkouts")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return nil, err
	}

	workouts, err := uc.workoutRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Error("failed to list workouts", "error", err)
		return nil, fmt.Errorf("list workouts: %w", err)
	}

	return workouts, nil
}
