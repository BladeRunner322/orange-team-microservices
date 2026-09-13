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

func TestLogout_Execute(t *testing.T) {
	log := logger.NewTestLogger()

	t.Run("успешный logout", func(t *testing.T) {
		refreshRepo := newMockRefreshTokenRepository()

		token, _ := domain.GenerateRefreshToken()
		_ = refreshRepo.Save(context.Background(), token, "user-123", time.Hour)

		uc := NewLogout(refreshRepo, log)
		err := uc.Execute(context.Background(), token.String())

		require.NoError(t, err)

		// Токен должен быть удалён
		_, err = refreshRepo.GetUserID(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("невалидный формат токена", func(t *testing.T) {
		refreshRepo := newMockRefreshTokenRepository()
		uc := NewLogout(refreshRepo, log)

		err := uc.Execute(context.Background(), "short")

		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("идемпотентность — повторный logout", func(t *testing.T) {
		refreshRepo := newMockRefreshTokenRepository()
		uc := NewLogout(refreshRepo, log)

		token, _ := domain.GenerateRefreshToken()

		// Первый logout (токена нет — не ошибка)
		err := uc.Execute(context.Background(), token.String())
		require.NoError(t, err)

		// Второй logout — тоже не ошибка
		err = uc.Execute(context.Background(), token.String())
		require.NoError(t, err)
	})
}
