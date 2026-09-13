//go:build integration
// +build integration

package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
)

func TestLimiter_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Поднимаем Redis в Docker
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	client, err := redis.NewClient(ctx, redis.Config{
		Addr: endpoint,
		DB:   0,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	limiter := NewLimiter(client.Client)

	// Правило: 5 токенов за минуту, burst = 5.
	// Восстановление: 1 токен в 12 секунд.
	rule := Rule{Rate: 5, Burst: 5, Interval: time.Minute}

	t.Run("первые 5 запросов проходят, 6-й блокируется", func(t *testing.T) {
		key := "test:burst"

		for i := 1; i <= 5; i++ {
			res, err := limiter.Allow(ctx, key, rule)
			require.NoError(t, err)
			assert.True(t, res.Allowed, "request %d should be allowed", i)
		}

		res, err := limiter.Allow(ctx, key, rule)
		require.NoError(t, err)
		assert.False(t, res.Allowed, "6th request should be blocked")
		assert.Greater(t, res.RetryAfter, time.Duration(0))
	})

	t.Run("разные ключи — независимые ведра", func(t *testing.T) {
		keyA := "test:userA"
		keyB := "test:userB"

		// Исчерпаем ведро A
		for i := 0; i < 5; i++ {
			_, _ = limiter.Allow(ctx, keyA, rule)
		}
		resA, _ := limiter.Allow(ctx, keyA, rule)
		assert.False(t, resA.Allowed)

		// Ведро B всё ещё полное
		resB, _ := limiter.Allow(ctx, keyB, rule)
		assert.True(t, resB.Allowed)
	})

	t.Run("RetryAfter возвращает разумное значение", func(t *testing.T) {
		key := "test:retryafter"

		// Исчерпаем полностью
		for i := 0; i < 5; i++ {
			_, _ = limiter.Allow(ctx, key, rule)
		}

		res, err := limiter.Allow(ctx, key, rule)
		require.NoError(t, err)
		assert.False(t, res.Allowed)
		// 1 токен / (5 токенов/60 сек) = 12 сек, но с округлением вверх
		assert.GreaterOrEqual(t, res.RetryAfter, 11*time.Second)
		assert.LessOrEqual(t, res.RetryAfter, 12*time.Second)
	})

	t.Run("невалидное правило возвращает ошибку", func(t *testing.T) {
		_, err := limiter.Allow(ctx, "test:invalid", Rule{Rate: 0, Burst: 5, Interval: time.Minute})
		assert.Error(t, err)
	})
}
