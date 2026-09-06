package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
)

type ValidateTokenUseCase struct {
	tokenManager ports.TokenManager
}

func NewValidateTokenUseCase(tokenManager ports.TokenManager) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{tokenManager: tokenManager}
}

func (uc *ValidateTokenUseCase) Execute(ctx context.Context, token string) (string, error) {
	return uc.tokenManager.Validate(ctx, token)
}
