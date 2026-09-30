package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
)

type GetProfile struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewGetProfile создаёт usecase для получения профиля по user_id.
func NewGetProfile(repo ports.Repository, log *logger.Logger) *GetProfile {
	return &GetProfile{repo: repo, logger: log}
}

// Execute возвращает профиль указанного пользователя.
// Если профиля нет — создаёт пустой (lazy-create).
// Внутренний метод: вызывается другими сервисами (например, Workouts),
// user_id приходит аргументом, а не из metadata.
func (uc *GetProfile) Execute(ctx context.Context, userID uuid.UUID) (domain.Profile, error) {
	log := uc.logger.With("operation", "GetProfile")

	return getOrCreateProfile(ctx, uc.repo, log, userID)
}
