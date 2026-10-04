package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
)

func TestGetExercise_Execute(t *testing.T) {
	t.Run("успешно возвращает упражнение", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)

		uc := NewGetExercise(repo, logger.NewTestLogger())

		// Act
		found, err := uc.Execute(context.Background(), created.ID())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, created.ID(), found.ID())
		assert.Equal(t, "Жим лёжа", found.Name().String())
	})

	t.Run("упражнение не найдено — ErrExerciseNotFound", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewGetExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New())

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("возвращает удалённое упражнение с флагом isDeleted", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(context.Background(), created.ID()))

		uc := NewGetExercise(repo, logger.NewTestLogger())

		// Act
		found, err := uc.Execute(context.Background(), created.ID())

		// Assert — по ADR-008 GetExercise возвращает и удалённые
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewGetExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
