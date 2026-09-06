package usecases

import (
	"context"
	"testing"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestValidateToken_Execute(t *testing.T) {
	log := logger.NewTestLogger()
	tokenManager := mockTokenManager{}

	t.Run("валидный токен", func(t *testing.T) {
		uc := NewValidateToken(tokenManager, log)
		userID, err := uc.Execute(context.Background(), "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userID != "user-id" {
			t.Errorf("expected user-id, got %s", userID)
		}
	})

	t.Run("невалидный токен", func(t *testing.T) {
		tokenManagerErr := mockTokenManager{validateErr: domain.ErrInvalidCredentials}
		uc := NewValidateToken(tokenManagerErr, log)
		_, err := uc.Execute(context.Background(), "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
