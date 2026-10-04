package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
)

// mockRepository — реализация ports.Repository для тестов.
// Хранит упражнения в map по id, эмулирует правила SQL:
//   - unique-имя среди активных (partial unique index);
//   - UPDATE/MarkDeleted не работают с удалёнными (WHERE deleted_at IS NULL).
type mockRepository struct {
	exercises map[uuid.UUID]domain.Exercise

	// err — если задана, все методы вернут её вместо нормальной логики.
	// Используется для проверки проброса инфраструктурных ошибок.
	err error
}

func newMockRepository() *mockRepository {
	return &mockRepository{exercises: make(map[uuid.UUID]domain.Exercise)}
}

// Create сохраняет упражнение. Проверяет уникальность name среди активных.
func (m *mockRepository) Create(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error) {
	if m.err != nil {
		return domain.Exercise{}, m.err
	}

	for _, existing := range m.exercises {
		if existing.IsDeleted() {
			continue
		}
		if existing.Name() == exercise.Name() {
			return domain.Exercise{}, domain.ErrExerciseNameExists
		}
	}

	m.exercises[exercise.ID()] = exercise

	return exercise, nil
}

// GetExercise возвращает упражнение по id, включая удалённые (ADR-008).
func (m *mockRepository) GetExercise(ctx context.Context, id uuid.UUID) (domain.Exercise, error) {
	if m.err != nil {
		return domain.Exercise{}, m.err
	}

	exercise, ok := m.exercises[id]
	if !ok {
		return domain.Exercise{}, domain.ErrExerciseNotFound
	}

	return exercise, nil
}

// GetExercises возвращает только активные упражнения (deleted_at IS NULL).
func (m *mockRepository) GetExercises(ctx context.Context) ([]domain.Exercise, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := make([]domain.Exercise, 0, len(m.exercises))
	for _, exercise := range m.exercises {
		if !exercise.IsDeleted() {
			result = append(result, exercise)
		}
	}

	return result, nil
}

// Update сохраняет изменения. Не работает с удалёнными и несуществующими.
// Проверяет уникальность name среди активных (кроме самого себя).
func (m *mockRepository) Update(ctx context.Context, exercise domain.Exercise) (domain.Exercise, error) {
	if m.err != nil {
		return domain.Exercise{}, m.err
	}

	existing, ok := m.exercises[exercise.ID()]
	if !ok || existing.IsDeleted() {
		return domain.Exercise{}, domain.ErrExerciseNotFound
	}

	for id, other := range m.exercises {
		if id == exercise.ID() || other.IsDeleted() {
			continue
		}
		if other.Name() == exercise.Name() {
			return domain.Exercise{}, domain.ErrExerciseNameExists
		}
	}

	m.exercises[exercise.ID()] = exercise

	return exercise, nil
}

// MarkDeleted помечает упражнение удалённым. Если уже удалено или не
// существует — ErrExerciseNotFound.
func (m *mockRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}

	existing, ok := m.exercises[id]
	if !ok || existing.IsDeleted() {
		return domain.ErrExerciseNotFound
	}

	deleted := domain.RestoreExercise(
		existing.ID(),
		existing.Name(),
		existing.Description(),
		existing.Difficulty(),
		existing.ExerciseType(),
		true,
		existing.CreatedAt(),
		existing.UpdatedAt(),
	)
	m.exercises[id] = deleted

	return nil
}
