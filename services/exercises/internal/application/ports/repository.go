// Package ports — интерфейсы, которые domain требует от инфраструктуры.
package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	// WithTx выполняет fn внутри транзакции. В fn передаётся репозиторий,
	// привязанный к tx. Если fn возвращает ошибку — откат.
	//
	// Используется usecase'ами для read-then-write, где между чтением
	// и записью нельзя допустить гонку (ADR-007 F-2).
	WithTx(ctx context.Context, fn func(Repository) error) error

	// Create сохраняет новое упражнение и возвращает его из БД
	// (с заполненными created_at / updated_at).
	Create(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error)

	// GetExercise возвращает упражнение по id, включая удалённые (см. ADR-008).
	// Если не найдено — domain.ErrExerciseNotFound.
	GetExercise(ctx context.Context, id uuid.UUID) (domain.Exercise, error)

	// GetExerciseForUpdate возвращает упражнение с блокировкой строки
	// (SELECT ... FOR UPDATE). Блокировка держится до COMMIT/ROLLBACK
	// транзакции, в которой вызван метод. Должен вызываться только
	// внутри WithTx.
	GetExerciseForUpdate(ctx context.Context, id uuid.UUID) (domain.Exercise, error)

	// GetExercises возвращает список активных упражнений (deleted_at IS NULL).
	GetExercises(ctx context.Context) ([]domain.Exercise, error)

	// Update сохраняет все бизнес-поля упражнения и возвращает актуальное
	// состояние из БД. Не трогает deleted_at.
	Update(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error)

	// MarkDeleted помечает упражнение удалённым (soft delete).
	// Если уже удалено — domain.ErrExerciseNotFound.
	MarkDeleted(ctx context.Context, id uuid.UUID) error
}