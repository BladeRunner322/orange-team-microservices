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
		tokenManager := mockTokenManager{validateErr: nil, role: "admin"}
		uc := NewValidateToken(tokenManager, log)

		info, err := uc.Execute(context.Background(), "valid-token")

		assert.NoError(t, err)
		assert.Equal(t, "user-id", info.UserID)
		assert.Equal(t, "admin", info.Role)
	})

	t.Run("невалидный токен", func(t *testing.T) {
		tokenManager := mockTokenManager{validateErr: domain.ErrInvalidCredentials}
		uc := NewValidateToken(tokenManager, log)

		_, err := uc.Execute(context.Background(), "")

		assert.Error(t, err)
	})
}
