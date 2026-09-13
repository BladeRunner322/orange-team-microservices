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

	// setupUser создаёт и сохраняет пользователя в моковом репозитории,
	// возвращая его для дальнейшего использования (например, чтобы взять ID).
	setupUser := func(repo *mockRepository) domain.User {
		email, _ := domain.NewEmail("test@example.com")
		fullName, _ := domain.NewFullName("Test User")
		passHash, _ := domain.NewPasswordHash("$2a$10$dummyhash")
		user := domain.NewUser(email, passHash, fullName)
		_ = repo.Save(context.Background(), user)
		return user
	}

	t.Run("успешная ротация", func(t *testing.T) {
		userRepo := newMockRepository()
		user := setupUser(userRepo)
		refreshRepo := newMockRefreshTokenRepository()

		// Кладём валидный refresh-токен, привязанный к реальному user_id
		oldToken, _ := domain.GenerateRefreshToken()
		_ = refreshRepo.Save(context.Background(), oldToken, user.ID().String(), time.Hour)

		uc := NewRefreshToken(refreshRepo, userRepo, tokenManager, time.Hour, log)
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
		userRepo := newMockRepository()
		refreshRepo := newMockRefreshTokenRepository()
		uc := NewRefreshToken(refreshRepo, userRepo, tokenManager, time.Hour, log)

		_, err := uc.Execute(context.Background(), "short")

		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("токен не найден в Redis", func(t *testing.T) {
		userRepo := newMockRepository()
		refreshRepo := newMockRefreshTokenRepository()
		uc := NewRefreshToken(refreshRepo, userRepo, tokenManager, time.Hour, log)

		// Формат валидный, но в репозитории его нет
		token, _ := domain.GenerateRefreshToken()
		_, err := uc.Execute(context.Background(), token.String())

		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("пользователь не найден в БД", func(t *testing.T) {
		userRepo := newMockRepository() // пустой — юзера нет
		refreshRepo := newMockRefreshTokenRepository()

		// Токен валидный и привязан к какому-то user_id, которого нет в БД
		token, _ := domain.GenerateRefreshToken()
		_ = refreshRepo.Save(context.Background(), token, "00000000-0000-0000-0000-000000000000", time.Hour)

		uc := NewRefreshToken(refreshRepo, userRepo, tokenManager, time.Hour, log)
		_, err := uc.Execute(context.Background(), token.String())

		assert.Error(t, err)
	})
}
