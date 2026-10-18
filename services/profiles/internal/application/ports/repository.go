// Package ports — интерфейсы, которые domain требует от инфраструктуры.
package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	// WithTx выполняет fn внутри транзакции. В fn передаётся репозиторий,
	// привязанный к tx. Если fn возвращает ошибку — откат.
	//
	// Используется usecase'ами для read-then-write, где между чтением
	// и записью нельзя допустить гонку (ADR-007 F-2).
	WithTx(ctx context.Context, fn func(Repository) error) error

	// EnsureExists создаёт пустой профиль, если его ещё нет.
	// Идемпотентен: повторный вызов с тем же userID — no-op.
	// Используется для lazy-create (ADR-001).
	EnsureExists(ctx context.Context, profile domain.Profile) error

	// GetByUserID возвращает профиль без блокировки.
	// Используется в read-only usecase'ах (GetMyProfile, GetProfile).
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, error)

	// GetByUserIDForUpdate возвращает профиль с блокировкой строки
	// (SELECT ... FOR UPDATE). Блокировка держится до COMMIT/ROLLBACK
	// транзакции, в которой вызван метод. Должен вызываться только
	// внутри WithTx.
	GetByUserIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.Profile, error)

	Update(ctx context.Context, profile domain.Profile) (domain.Profile, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}
