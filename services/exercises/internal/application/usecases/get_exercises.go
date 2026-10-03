package usecases

import (
	"context"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
)

type GetExercises struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewGetExercises создаёт usecase для получения списка активных упражнений.
func NewGetExercises(repo ports.Repository, log *logger.Logger) *GetExercises {
	return &GetExercises{repo: repo, logger: log}
}

// Execute возвращает список активных упражнений (без удалённых).
func (uc *GetExercises) Execute(ctx context.Context) ([]domain.Exercise, error) {
	log := uc.logger.With("operation", "GetExercises")

	exercises, err := uc.repo.GetExercises(ctx)
	if err != nil {
		log.Error("failed to get exercises", "error", err)
		return []domain.Exercise{}, fmt.Errorf("get exercises: %w", err)
	}

	log.Info("exercises fetched", "count", len(exercises))

	return exercises, nil
}
