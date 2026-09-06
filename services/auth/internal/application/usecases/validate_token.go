package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
)

type Validate struct {
	tokenManager ports.TokenManager
	logger       *logger.Logger
}

func NewValidateToken(tokenManager ports.TokenManager, log *logger.Logger) *Validate {
	return &Validate{tokenManager: tokenManager, logger: log}
}

func (uc *Validate) Execute(ctx context.Context, token string) (string, error) {
	log := uc.logger.With("operation", "ValidateToken")
	userID, err := uc.tokenManager.Validate(ctx, token)
	if err != nil {
		log.Warn("invalid token", "error", err)
		return "", err
	}
	log.Info("token validated successfully", "user_id", userID)
	return userID, nil
}
