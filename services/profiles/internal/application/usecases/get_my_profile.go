package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
)

type GetMyProfile struct {
	repo   ports.Repository
	logger *logger.Logger
}

// NewGetMyProfile создаёт usecase для получения профиля текущего пользователя.
func NewGetMyProfile(repo ports.Repository, log *logger.Logger) *GetMyProfile {
	return &GetMyProfile{repo: repo, logger: log}
}

// Execute возвращает профиль текущего пользователя (user_id из context).
// Если профиля нет — создаёт пустой.
func (uc *GetMyProfile) Execute(ctx context.Context) (domain.Profile, error) {
	log := uc.logger.With("operation", "GetMyProfile")

	userID, err := userIDFromContext(ctx, log)
	if err != nil {
		return domain.Profile{}, err
	}

	return getOrCreateProfile(ctx, uc.repo, log, userID)
}
