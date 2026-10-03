// Package ports — интерфейсы, которые domain требует от инфраструктуры.
package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	// Create сохраняет новое упражнение и возвращает его из БД
	// (с заполненными created_at / updated_at).
	Create(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error)

	// GetExercise возвращает упражнение по id, включая удалённые (см. ADR-008).
	// Если не найдено — domain.ErrExerciseNotFound.
	GetExercise(ctx context.Context, id uuid.UUID) (domain.Exercise, error)

	// GetExercises возвращает список активных упражнений (deleted_at IS NULL).
	GetExercises(ctx context.Context) ([]domain.Exercise, error)

	// Update сохраняет все бизнес-поля упражнения и возвращает актуальное
	// состояние из БД. Не трогает deleted_at.
	Update(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error)

	// MarkDeleted помечает упражнение удалённым (soft delete).
	// Если уже удалено — domain.ErrExerciseNotFound.
	MarkDeleted(ctx context.Context, id uuid.UUID) error
}
