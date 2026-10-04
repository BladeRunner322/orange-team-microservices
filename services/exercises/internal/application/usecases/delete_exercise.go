package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"github.com/google/uuid"
)

type DeleteExercise struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewDeleteExercise создаёт usecase для soft-delete упражнения (admin-only).
func NewDeleteExercise(repo ports.Repository, log *logger.Logger) *DeleteExercise {
	return &DeleteExercise{repo: repo, logger: log}
}

// / Execute помечает упражнение удалённым (deleted_at = NOW()).
//
// Идемпотентен: повторный вызов (или DELETE несуществующего id)
// возвращает nil, а не ошибку (см. ADR-007 F-3). Это защищает
// от retry после network timeout — прокси/клиент может безопасно
// повторить запрос.
func (uc *DeleteExercise) Execute(ctx context.Context, id uuid.UUID) error {
	log := uc.logger.With("operation", "DeleteExercise", "exercise_id", id)

	if err := uc.repo.MarkDeleted(ctx, id); err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			log.Warn("exercise not found or already deleted")
			return nil
		}

		log.Error("failed to mark deleted", "error", err)
		return fmt.Errorf("mark deleted: %w", err)
	}

	return nil
}
