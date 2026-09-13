//go:build integration
// +build integration

package redis_repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestRefreshTokenRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Поднимаем Redis в Docker
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	// 2. Подключаемся
	client, err := redis.NewClient(ctx, redis.Config{
		Addr: endpoint,
		DB:   0,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	repo := NewRepository(client)

	// 3. Тесты
	t.Run("Save and GetUserID", func(t *testing.T) {
		token, _ := domain.GenerateRefreshToken()
		err := repo.Save(ctx, token, "user-123", time.Hour)
		require.NoError(t, err)

		userID, err := repo.GetUserID(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", userID)
	})

	t.Run("GetUserID not found", func(t *testing.T) {
		token, _ := domain.GenerateRefreshToken()
		_, err := repo.GetUserID(ctx, token)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("Delete removes token", func(t *testing.T) {
		token, _ := domain.GenerateRefreshToken()
		_ = repo.Save(ctx, token, "user-123", time.Hour)

		err := repo.Delete(ctx, token)
		require.NoError(t, err)

		_, err = repo.GetUserID(ctx, token)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})

	t.Run("Delete is idempotent", func(t *testing.T) {
		token, _ := domain.GenerateRefreshToken()

		// Удаление несуществующего токена — не ошибка
		err := repo.Delete(ctx, token)
		require.NoError(t, err)
	})

	t.Run("DeleteAllForUser", func(t *testing.T) {
		userID := "user-456"

		// Создаём 3 токена для одного пользователя
		tokens := make([]domain.RefreshToken, 3)
		for i := range tokens {
			tok, _ := domain.GenerateRefreshToken()
			tokens[i] = tok
			_ = repo.Save(ctx, tok, userID, time.Hour)
		}

		// + 1 токен другого пользователя
		otherToken, _ := domain.GenerateRefreshToken()
		_ = repo.Save(ctx, otherToken, "user-789", time.Hour)

		// Удаляем все для user-456
		err := repo.DeleteAllForUser(ctx, userID)
		require.NoError(t, err)

		// Все 3 удалены
		for _, tok := range tokens {
			_, err := repo.GetUserID(ctx, tok)
			assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
		}

		// Токен другого пользователя остался
		uid, err := repo.GetUserID(ctx, otherToken)
		require.NoError(t, err)
		assert.Equal(t, "user-789", uid)
	})

	t.Run("TTL applied", func(t *testing.T) {
		token, _ := domain.GenerateRefreshToken()
		err := repo.Save(ctx, token, "user-ttl", 10*time.Second)
		require.NoError(t, err)

		// Проверяем, что TTL установлен (примерно 10 секунд)
		ttl, err := client.TTL(ctx, refreshTokenKey(token)).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
		assert.LessOrEqual(t, ttl, 10*time.Second)
	})
}
