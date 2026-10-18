package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewCompletion(t *testing.T) {
	habitID := uuid.New()
	date, _ := NewDate(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))

	completion := NewCompletion(habitID, date)

	t.Run("поля соответствуют переданным", func(t *testing.T) {
		assert.Equal(t, habitID, completion.HabitID())
		assert.Equal(t, date, completion.Date())
	})

	t.Run("createdAt заполнен", func(t *testing.T) {
		assert.False(t, completion.CreatedAt().IsZero())
	})
}

func TestRestoreCompletion(t *testing.T) {
	habitID := uuid.New()
	date, _ := NewDate(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))
	createdAt := time.Date(2026, 10, 15, 8, 30, 0, 0, time.UTC)

	completion := RestoreCompletion(habitID, date, createdAt)

	assert.Equal(t, habitID, completion.HabitID())
	assert.Equal(t, date, completion.Date())
	assert.Equal(t, createdAt, completion.CreatedAt())
}
