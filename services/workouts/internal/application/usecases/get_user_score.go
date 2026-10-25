package usecases

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
)

type GetUserScore struct {
	workoutRepo ports.WorkoutRepository
	logger      *logger.Logger
}

func NewGetUserScore(repo ports.WorkoutRepository, log *logger.Logger) *GetUserScore {
	return &GetUserScore{workoutRepo: repo, logger: log}
}

// Execute возвращает сумму workout_score по всем completed тренировкам
// пользователя (ADR-002). user_id берётся из context (ставит Gateway).
func (uc *GetUserScore) Execute(ctx context.Context) (int64, error) {
	log := uc.logger.With("operation", "GetUserScore")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return 0, err
	}

	score, err := uc.workoutRepo.GetUserScore(ctx, userID)
	if err != nil {
		log.Error("failed to get user score", "error", err)
		return 0, fmt.Errorf("get user score: %w", err)
	}

	return score, nil
}
