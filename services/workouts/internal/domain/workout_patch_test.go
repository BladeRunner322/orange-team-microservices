package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkoutPatch(t *testing.T) {
	t.Run("пустой патч", func(t *testing.T) {
		p := NewWorkoutPatch(nil, nullableIntensityNone())

		assert.Nil(t, p.Status)
		assert.False(t, p.Intensity.Set)
	})

	t.Run("status задан", func(t *testing.T) {
		s := StatusInProgress
		p := NewWorkoutPatch(&s, nullableIntensityNone())

		assert.NotNil(t, p.Status)
		assert.Equal(t, StatusInProgress, *p.Status)
	})

	t.Run("intensity задан", func(t *testing.T) {
		i, _ := NewIntensity(7)
		p := NewWorkoutPatch(nil, nullableIntensitySet(&i))

		assert.Nil(t, p.Status)
		assert.True(t, p.Intensity.Set)
		assert.NotNil(t, p.Intensity.Value)
		assert.Equal(t, 7, p.Intensity.Value.Int())
	})

	t.Run("intensity сброс", func(t *testing.T) {
		p := NewWorkoutPatch(nil, nullableIntensityClear())

		assert.True(t, p.Intensity.Set)
		assert.Nil(t, p.Intensity.Value)
	})
}