package usecases

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/google/uuid"
)

// userIDFromContext извлекает user_id из context и парсит в uuid.UUID.
// Возвращает domain.ErrUnauthenticated, если user_id отсутствует или невалиден.
func userIDFromContext(ctx context.Context, log *logger.Logger) (uuid.UUID, error) {
	// TODO: по образцу Profiles/usecases/common.go
	return uuid.Nil, nil
}
