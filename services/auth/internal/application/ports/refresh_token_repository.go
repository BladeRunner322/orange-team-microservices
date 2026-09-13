package ports

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

// RefreshTokenRepository — хранилище refresh-токенов (Redis).
type RefreshTokenRepository interface {
	// Save сохраняет refresh-токен для пользователя с указанным TTL.
	Save(ctx context.Context, token domain.RefreshToken, userID string, ttl time.Duration) error

	// GetUserID возвращает user_id по refresh-токену.
	// Если токен не найден — возвращает domain.ErrInvalidRefreshToken.
	GetUserID(ctx context.Context, token domain.RefreshToken) (string, error)

	// Delete удаляет refresh-токен (используется при rotation и logout).
	Delete(ctx context.Context, token domain.RefreshToken) error

	// DeleteAllForUser удаляет все refresh-токены пользователя (revoke all).
	DeleteAllForUser(ctx context.Context, userID string) error
}
