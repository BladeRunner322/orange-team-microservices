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

func TestCreateExercise_Execute(t *testing.T) {
	t.Run("успешно создаёт упражнение", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		// Act
		created, err := uc.Execute(context.Background(), "Жим лёжа", "Базовое упражнение", 5, "weight")

		// Assert
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, created.ID())
		assert.Equal(t, "Жим лёжа", created.Name().String())
		assert.Equal(t, "Базовое упражнение", created.Description().String())
		assert.Equal(t, 5, created.Difficulty().Int())
		assert.Equal(t, domain.ExerciseTypeWeight, created.ExerciseType())
		assert.False(t, created.IsDeleted())

		// Действительно сохранено в репозитории
		saved, err := repo.GetExercise(context.Background(), created.ID())
		require.NoError(t, err)
		assert.Equal(t, created.ID(), saved.ID())
	})

	t.Run("невалидное name — ErrInvalidName", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Жм", "Описание", 5, "weight")

		assert.ErrorIs(t, err, domain.ErrInvalidName)
	})

	t.Run("невалидное description — ErrInvalidDescription", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Жим лёжа", "", 5, "weight")

		assert.ErrorIs(t, err, domain.ErrInvalidDescription)
	})

	t.Run("невалидный difficulty — ErrInvalidDifficulty", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Жим лёжа", "Описание", 0, "weight")

		assert.ErrorIs(t, err, domain.ErrInvalidDifficulty)
	})

	t.Run("невалидный type — ErrInvalidExerciseType", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Жим лёжа", "Описание", 5, "cardio")

		assert.ErrorIs(t, err, domain.ErrInvalidExerciseType)
	})

	t.Run("имя уже занято — ErrExerciseNameExists", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Жим лёжа", "Первое описание", 5, "weight")
		require.NoError(t, err)

		_, err = uc.Execute(context.Background(), "Жим лёжа", "Второе описание", 7, "duration")

		assert.ErrorIs(t, err, domain.ErrExerciseNameExists)
	})

	t.Run("ошибка репозитория при Create — пробрасывается", func(t *testing.T) {
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewCreateExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
