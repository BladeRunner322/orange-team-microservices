package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestRefreshToken_Execute(t *testing.T) {
	log := logger.NewTestLogger()
	tokenManager := mockTokenManager{}

	t.Run("успешная ротация", func(t *testing.T) {
		refreshRepo := newMockRefreshTokenRepository()

		// Кладём валидный refresh-токен, чтобы его можно было обменять
		oldToken, _ := domain.GenerateRefreshToken()
		_ = refreshRepo.Save(context.Background(), oldToken, "user-123", time.Hour)

		uc := NewRefreshToken(refreshRepo, tokenManager, time.Hour, log)
		result, err := uc.Execute(context.Background(), oldToken.String())

		require.NoError(t, err)
		assert.Equal(t, "test-token", result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.NotEqual(t, oldToken.String(), result.RefreshToken, "старый токен должен быть заменён")

		// Старый токен должен быть удалён из репозитория
		_, err = refreshRepo.GetUserID(context.Background(), oldToken)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("невалидный формат токена", func(t *testing.T) {
		refreshRepo := newMockRefreshTokenRepository()
		uc := NewRefreshToken(refreshRepo, tokenManager, time.Hour, log)

		_, err := uc.Execute(context.Background(), "short")

		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("токен не найден в Redis", func(t *testing.T) {
		refreshRepo := newMockRefreshTokenRepository()
		uc := NewRefreshToken(refreshRepo, tokenManager, time.Hour, log)

		// Формат валидный, но в репозитории его нет
		token, _ := domain.GenerateRefreshToken()
		_, err := uc.Execute(context.Background(), token.String())

		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})
}
