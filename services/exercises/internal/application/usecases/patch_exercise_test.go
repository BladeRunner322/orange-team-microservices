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

// patchWithName — хелпер: собирает патч только с name.
func patchWithName(raw string) domain.ExercisePatch {
	name, _ := domain.NewName(raw)
	return domain.NewExercisePatch(&name, nil, nil)
}

func TestPatchExercise_Execute(t *testing.T) {
	t.Run("успешно патчит name", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Старое имя", "Описание", 5, "weight")
		require.NoError(t, err)

		uc := NewPatchExercise(repo, logger.NewTestLogger())

		// Act
		updated, err := uc.Execute(context.Background(), created.ID(), patchWithName("Новое имя"))

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Новое имя", updated.Name().String())
		assert.Equal(t, "Описание", updated.Description().String())
		assert.Equal(t, 5, updated.Difficulty().Int())
		assert.Equal(t, domain.ExerciseTypeWeight, updated.ExerciseType())
	})

	t.Run("успешно патчит несколько полей", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Старое имя", "Старое описание", 5, "weight")
		require.NoError(t, err)

		newName, _ := domain.NewName("Новое имя")
		newDescription, _ := domain.NewDescription("Новое описание")
		newDifficulty, _ := domain.NewDifficulty(9)
		patch := domain.NewExercisePatch(&newName, &newDescription, &newDifficulty)

		uc := NewPatchExercise(repo, logger.NewTestLogger())

		// Act
		updated, err := uc.Execute(context.Background(), created.ID(), patch)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Новое имя", updated.Name().String())
		assert.Equal(t, "Новое описание", updated.Description().String())
		assert.Equal(t, 9, updated.Difficulty().Int())
	})

	t.Run("пустой патч — ничего не меняется", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)

		patch := domain.NewExercisePatch(nil, nil, nil)
		uc := NewPatchExercise(repo, logger.NewTestLogger())

		// Act
		updated, err := uc.Execute(context.Background(), created.ID(), patch)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Жим лёжа", updated.Name().String())
		assert.Equal(t, "Описание", updated.Description().String())
		assert.Equal(t, 5, updated.Difficulty().Int())
	})

	t.Run("упражнение не найдено — ErrExerciseNotFound", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewPatchExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), patchWithName("Новое имя"))

		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("патч удалённого упражнения — ErrExerciseNotFound", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)
		require.NoError(t, repo.MarkDeleted(context.Background(), created.ID()))

		uc := NewPatchExercise(repo, logger.NewTestLogger())

		// Act
		_, err = uc.Execute(context.Background(), created.ID(), patchWithName("Новое имя"))

		// Assert — правило "нельзя патчить удалённое" живёт в Update (WHERE deleted_at IS NULL)
		assert.ErrorIs(t, err, domain.ErrExerciseNotFound)
	})

	t.Run("конфликт имён — ErrExerciseNameExists", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		first, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание 1", 5, "weight")
		require.NoError(t, err)
		_, err = createUC.Execute(context.Background(), "Приседания", "Описание 2", 7, "weight")
		require.NoError(t, err)

		uc := NewPatchExercise(repo, logger.NewTestLogger())

		// Act — патчим first, ставим имя второго
		_, err = uc.Execute(context.Background(), first.ID(), patchWithName("Приседания"))

		// Assert
		assert.ErrorIs(t, err, domain.ErrExerciseNameExists)
	})

	t.Run("патч с тем же именем — не конфликт", func(t *testing.T) {
		// Arrange
		repo := newMockRepository()
		createUC := NewCreateExercise(repo, logger.NewTestLogger())
		created, err := createUC.Execute(context.Background(), "Жим лёжа", "Описание", 5, "weight")
		require.NoError(t, err)

		uc := NewPatchExercise(repo, logger.NewTestLogger())

		// Act — патчим имя на то же самое
		updated, err := uc.Execute(context.Background(), created.ID(), patchWithName("Жим лёжа"))

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Жим лёжа", updated.Name().String())
	})

	t.Run("ошибка репозитория при Get — пробрасывается", func(t *testing.T) {
		repo := newMockRepository()
		repo.err = errors.New("db is down")
		uc := NewPatchExercise(repo, logger.NewTestLogger())

		_, err := uc.Execute(context.Background(), uuid.New(), patchWithName("Новое имя"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is down")
	})
}
