// Package ports — интерфейсы, которые domain требует от инфраструктуры.
package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, error)
	// EnsureExists создаёт пустой профиль, если его ещё нет.
	// Идемпотентен: повторный вызов с тем же userID — no-op.
	// Используется для lazy-create (ADR-001).
	EnsureExists(ctx context.Context, profile domain.Profile) error
	Update(ctx context.Context, profile domain.Profile) (domain.Profile, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}
