package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestValidateToken_Execute(t *testing.T) {
	log := logger.NewTestLogger()

	t.Run("валидный токен", func(t *testing.T) {
		tokenManager := mockTokenManager{validateErr: nil}
		uc := NewValidateToken(tokenManager, log)

		userID, err := uc.Execute(context.Background(), "valid-token")

		assert.NoError(t, err)
		assert.Equal(t, "user-id", userID)
	})

	t.Run("невалидный токен", func(t *testing.T) {
		tokenManager := mockTokenManager{validateErr: domain.ErrInvalidCredentials}
		uc := NewValidateToken(tokenManager, log)

		_, err := uc.Execute(context.Background(), "")

		assert.Error(t, err)
	})
}
