package redis_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

const (
	refreshTokenKeyPrefix = "refresh:"
	userTokensKeyPrefix   = "user:"
	userTokensKeySuffix   = ":tokens"
)

// refreshTokenKey возвращает Redis-ключ для refresh-токена.
func refreshTokenKey(token domain.RefreshToken) string {
	return refreshTokenKeyPrefix + token.String()
}

// userTokensKey возвращает Redis-ключ для set токенов пользователя.
func userTokensKey(userID string) string {
	return userTokensKeyPrefix + userID + userTokensKeySuffix
}

// Repository реализует ports.RefreshTokenRepository через Redis.
type Repository struct {
	client *redis.Client
}

// NewRepository создаёт новый репозиторий refresh-токенов.
func NewRepository(client *redis.Client) *Repository {
	return &Repository{client: client}
}

// Save сохраняет refresh-токен для пользователя.
func (r *Repository) Save(
	ctx context.Context,
	token domain.RefreshToken,
	userID string,
	ttl time.Duration,
) error {
	tokenKey := refreshTokenKey(token)
	userKey := userTokensKey(userID)

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, tokenKey, userID, ttl)
	pipe.SAdd(ctx, userKey, token.String())
	// TTL у user-set продлеваем на тот же срок (чтобы он не висел вечно)
	pipe.Expire(ctx, userKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

// GetUserID возвращает user_id по refresh-токену.
func (r *Repository) GetUserID(ctx context.Context, token domain.RefreshToken) (string, error) {
	tokenKey := refreshTokenKey(token)
	userID, err := r.client.Get(ctx, tokenKey).Result()
	if errors.Is(err, goredis.Nil) {
		return "", domain.ErrInvalidRefreshToken
	}
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return userID, nil
}

// Delete удаляет refresh-токен.
func (r *Repository) Delete(ctx context.Context, token domain.RefreshToken) error {
	tokenKey := refreshTokenKey(token)

	// Сначала достаём userID, чтобы убрать токен из set пользователя
	userID, err := r.client.Get(ctx, tokenKey).Result()
	if errors.Is(err, goredis.Nil) {
		// Токен уже удалён — ничего не делаем
		return nil
	}
	if err != nil {
		return fmt.Errorf("get refresh token before delete: %w", err)
	}

	userKey := userTokensKey(userID)

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, tokenKey)
	pipe.SRem(ctx, userKey, token.String())

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteAllForUser удаляет все refresh-токены пользователя.
func (r *Repository) DeleteAllForUser(ctx context.Context, userID string) error {
	userKey := userTokensKey(userID)

	tokens, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("get user tokens: %w", err)
	}

	pipe := r.client.TxPipeline()
	for _, t := range tokens {
		pipe.Del(ctx, refreshTokenKeyPrefix+t)
	}
	pipe.Del(ctx, userKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("delete all user tokens: %w", err)
	}
	return nil
}
