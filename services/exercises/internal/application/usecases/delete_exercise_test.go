package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
)

func TestDeleteExercise_Execute(t *testing.T) {
	t.Run("успешно удаляет упражнение", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)

		uc := NewDeleteExercise(repo, logger.NewTestLogger())

		// Act
		err = uc.Execute(context.Background(), created.ID())

		// Assert
		require.NoError(t, err)

		// Упражнение помечено удалённым
		found, err := repo.GetExercise(context.Background(), created.ID())
		require.NoError(t, err)
		assert.True(t, found.IsDeleted())

		// В списке активных его нет
		list, err := repo.GetExercises(context.Background())
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("идемпотентность: повторный вызов — не ошибка", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)

		uc := NewDeleteExercise(repo, logger.NewTestLogger())

		// Act — два раза подряд
		require.NoError(t, uc.Execute(context.Background(), created.ID()))
		err = uc.Execute(context.Background(), created.ID())

		// Assert — второй раз тоже nil (см. ADR-007 F-3)
		assert.NoError(t, err, "повторный DELETE должен быть идемпотентен")
	})

	t.Run("идемпотентность: несуществующий id — не ошибка", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		uc := NewDeleteExercise(repo, logger.NewTestLogger())

		// Act — удаляем то, чего нет
		err := uc.Execute(context.Background(), uuid.New())

		// Assert — тоже nil
		assert.NoError(t, err, "DELETE несуществующего должен быть идемпотентен")
	})

	t.Run("ошибка репозитория — пробрасывается", func(t *testing.T) {
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewDeleteExercise(repo, logger.NewTestLogger())

		err := uc.Execute(context.Background(), uuid.New())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
