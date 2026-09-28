package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/grpc/authctx"
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
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

	userIDStr, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		log.Warn("user_id is missing in context")
		return domain.Profile{}, domain.ErrUnauthenticated
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn("invalid user_id format", "user_id", userIDStr, "error", err)
		return domain.Profile{}, domain.ErrUnauthenticated
	}

	return getOrCreateProfile(ctx, uc.repo, log, userID)
}
