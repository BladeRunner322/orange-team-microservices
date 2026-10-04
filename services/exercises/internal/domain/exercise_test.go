package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExercise(t *testing.T) {
	name, _ := NewName("Жим лёжа")
	description, _ := NewDescription("Базовое упражнение")
	difficulty, _ := NewDifficulty(5)
	exerciseType, _ := NewExerciseType("weight")

	exercise := NewExercise(name, description, difficulty, exerciseType)

	t.Run("генерирует id", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, exercise.ID())
	})

	t.Run("поля соответствуют переданным", func(t *testing.T) {
		assert.Equal(t, name, exercise.Name())
		assert.Equal(t, description, exercise.Description())
		assert.Equal(t, difficulty, exercise.Difficulty())
		assert.Equal(t, exerciseType, exercise.ExerciseType())
	})

	t.Run("createdAt заполнен, updatedAt nil", func(t *testing.T) {
		assert.False(t, exercise.CreatedAt().IsZero())
		assert.Nil(t, exercise.UpdatedAt())
	})

	t.Run("isDeleted = false", func(t *testing.T) {
		assert.False(t, exercise.IsDeleted())
	})

	t.Run("id уникален между вызовами", func(t *testing.T) {
		other := NewExercise(name, description, difficulty, exerciseType)
		assert.NotEqual(t, exercise.ID(), other.ID())
	})
}

func TestRestoreExercise(t *testing.T) {
	id := uuid.New()
	name, _ := NewName("Приседания")
	description, _ := NewDescription("Со штангой")
	difficulty, _ := NewDifficulty(7)
	exerciseType, _ := NewExerciseType("duration")
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	t.Run("восстанавливает все поля", func(t *testing.T) {
		exercise := RestoreExercise(
			id, name, description, difficulty, exerciseType,
			false, createdAt, &updatedAt,
		)

		assert.Equal(t, id, exercise.ID())
		assert.Equal(t, name, exercise.Name())
		assert.Equal(t, description, exercise.Description())
		assert.Equal(t, difficulty, exercise.Difficulty())
		assert.Equal(t, exerciseType, exercise.ExerciseType())
		assert.False(t, exercise.IsDeleted())
		assert.Equal(t, createdAt, exercise.CreatedAt())
		require.NotNil(t, exercise.UpdatedAt())
		assert.Equal(t, updatedAt, *exercise.UpdatedAt())
	})

	t.Run("удалённое упражнение", func(t *testing.T) {
		exercise := RestoreExercise(
			id, name, description, difficulty, exerciseType,
			true, createdAt, &updatedAt,
		)

		assert.True(t, exercise.IsDeleted())
	})

	t.Run("updatedAt nil", func(t *testing.T) {
		exercise := RestoreExercise(
			id, name, description, difficulty, exerciseType,
			false, createdAt, nil,
		)

		assert.Nil(t, exercise.UpdatedAt())
	})
}

func TestExerciseGetGetters(t *testing.T) {
	id := uuid.New()
	name, _ := NewName("Жим")
	description, _ := NewDescription("Описание")
	difficulty, _ := NewDifficulty(3)
	exerciseType, _ := NewExerciseType("weight")
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	exercise := RestoreExercise(
		id, name, description, difficulty, exerciseType,
		false, createdAt, nil,
	)

	t.Run("возвращает те же значения", func(t *testing.T) {
		assert.Equal(t, id, exercise.ID())
		assert.Equal(t, name, exercise.Name())
		assert.Equal(t, description, exercise.Description())
		assert.Equal(t, difficulty, exercise.Difficulty())
		assert.Equal(t, exerciseType, exercise.ExerciseType())
		assert.Equal(t, createdAt, exercise.CreatedAt())
		assert.Nil(t, exercise.UpdatedAt())
		assert.False(t, exercise.IsDeleted())
	})
}
