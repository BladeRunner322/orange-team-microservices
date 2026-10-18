package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHabit(t *testing.T) {
	userID := uuid.New()
	name, _ := NewName("Пить воду")
	description, _ := NewDescription("2 литра в день")

	habit := NewHabit(userID, name, description)

	t.Run("генерирует id", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, habit.ID())
	})

	t.Run("поля соответствуют переданным", func(t *testing.T) {
		assert.Equal(t, userID, habit.UserID())
		assert.Equal(t, name, habit.Name())
		assert.Equal(t, description, habit.Description())
	})

	t.Run("createdAt заполнен, updatedAt nil", func(t *testing.T) {
		assert.False(t, habit.CreatedAt().IsZero())
		assert.Nil(t, habit.UpdatedAt())
	})

	t.Run("isDeleted = false", func(t *testing.T) {
		assert.False(t, habit.IsDeleted())
	})

	t.Run("id уникален между вызовами", func(t *testing.T) {
		other := NewHabit(userID, name, description)
		assert.NotEqual(t, habit.ID(), other.ID())
	})
}

func TestRestoreHabit(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	name, _ := NewName("Пить воду")
	description, _ := NewDescription("2 литра в день")
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	t.Run("восстанавливает все поля", func(t *testing.T) {
		habit := RestoreHabit(
			id, userID, name, description,
			false, createdAt, &updatedAt,
		)

		assert.Equal(t, id, habit.ID())
		assert.Equal(t, userID, habit.UserID())
		assert.Equal(t, name, habit.Name())
		assert.Equal(t, description, habit.Description())
		assert.False(t, habit.IsDeleted())
		assert.Equal(t, createdAt, habit.CreatedAt())
		require.NotNil(t, habit.UpdatedAt())
		assert.Equal(t, updatedAt, *habit.UpdatedAt())
	})

	t.Run("удалённая привычка", func(t *testing.T) {
		habit := RestoreHabit(
			id, userID, name, description,
			true, createdAt, &updatedAt,
		)

		assert.True(t, habit.IsDeleted())
	})

	t.Run("updatedAt nil", func(t *testing.T) {
		habit := RestoreHabit(
			id, userID, name, description,
			false, createdAt, nil,
		)

		assert.Nil(t, habit.UpdatedAt())
	})
}

func TestHabit_ApplyPatch(t *testing.T) {
	newTestHabit := func() Habit {
		name, _ := NewName("Пить воду")
		description, _ := NewDescription("2 литра в день")
		return NewHabit(uuid.New(), name, description)
	}

	t.Run("пустой патч — ничего не меняется", func(t *testing.T) {
		habit := newTestHabit()
		originalID := habit.ID()

		habit.ApplyPatch(NewHabitPatch(nil, nil))

		assert.Equal(t, originalID, habit.ID())
		assert.Equal(t, "Пить воду", habit.Name().String())
		assert.Equal(t, "2 литра в день", habit.Description().String())
	})

	t.Run("только name", func(t *testing.T) {
		habit := newTestHabit()
		newName, _ := NewName("Читать")

		habit.ApplyPatch(NewHabitPatch(&newName, nil))

		assert.Equal(t, "Читать", habit.Name().String())
		assert.Equal(t, "2 литра в день", habit.Description().String())
	})

	t.Run("только description", func(t *testing.T) {
		habit := newTestHabit()
		newDescription, _ := NewDescription("30 минут перед сном")

		habit.ApplyPatch(NewHabitPatch(nil, &newDescription))

		assert.Equal(t, "Пить воду", habit.Name().String())
		assert.Equal(t, "30 минут перед сном", habit.Description().String())
	})

	t.Run("все поля", func(t *testing.T) {
		habit := newTestHabit()
		newName, _ := NewName("Читать")
		newDescription, _ := NewDescription("30 минут")

		habit.ApplyPatch(NewHabitPatch(&newName, &newDescription))

		assert.Equal(t, "Читать", habit.Name().String())
		assert.Equal(t, "30 минут", habit.Description().String())
	})

	t.Run("сброс description в пустое", func(t *testing.T) {
		habit := newTestHabit()
		empty, _ := NewDescription("")

		habit.ApplyPatch(NewHabitPatch(nil, &empty))

		assert.Equal(t, "Пить воду", habit.Name().String())
		assert.Equal(t, "", habit.Description().String())
	})
}

func TestHabit_IsDeleted(t *testing.T) {
	t.Run("активная привычка", func(t *testing.T) {
		name, _ := NewName("Пить воду")
		description, _ := NewDescription("")
		habit := NewHabit(uuid.New(), name, description)
		assert.False(t, habit.IsDeleted())
	})

	t.Run("удалённая привычка", func(t *testing.T) {
		name, _ := NewName("Пить воду")
		description, _ := NewDescription("")
		habit := RestoreHabit(
			uuid.New(), uuid.New(), name, description,
			true, time.Now().UTC(), nil,
		)
		assert.True(t, habit.IsDeleted())
	})
}
