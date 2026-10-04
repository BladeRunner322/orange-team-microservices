package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
)

func TestGetExercises_Execute(t *testing.T) {
	t.Run("возвращает список активных упражнений", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание 1", 5, "weight")
		require.NoError(t, err)
		_, err = createUC.Execute(context.Background(), "Приседания", "Описание 2", 7, "weight")
		require.NoError(t, err)
		_, err = createUC.Execute(context.Background(), "Планка", "Описание 3", 3, "duration")
		require.NoError(t, err)

		uc := NewGetExercises(repo, logger.NewTestLogger())

		// Act
		list, err := uc.Execute(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Len(t, list, 3)
	})

	t.Run("не включает удалённые", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())

		active, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание 1", 5, "weight")
		require.NoError(t, err)
		toDelete, err := createUC.Execute(context.Background(), "Приседания", "Описание 2", 7, "weight")
		require.NoError(t, err)

		require.NoError(t, repo.MarkDeleted(context.Background(), toDelete.ID()))

		uc := NewGetExercises(repo, logger.NewTestLogger())

		// Act
		list, err := uc.Execute(context.Background())

		// Assert
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, active.ID(), list[0].ID())
	})

	t.Run("пустой список — возвращает не-nil пустой slice", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewGetExercises(repo, logger.NewTestLogger())

		list, err := uc.Execute(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, list, "должен быть []domain.Exercise{}, а не nil")
		assert.Empty(t, list)
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewGetExercises(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
