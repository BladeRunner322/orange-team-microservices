package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkoutExercisePatch(t *testing.T) {
	t.Run("пустой патч", func(t *testing.T) {
		p := NewWorkoutExercisePatch(nil, nil, nil, nil, nil)

		assert.Nil(t, p.Weight)
		assert.Nil(t, p.Sets)
		assert.Nil(t, p.Reps)
		assert.Nil(t, p.Duration)
		assert.Nil(t, p.Completed)
	})

	t.Run("все поля заданы", func(t *testing.T) {
		w, _ := NewWeight(100)
		s, _ := NewSets(4)
		r, _ := NewReps(10)
		c := true

		p := NewWorkoutExercisePatch(&w, &s, &r, nil, &c)

		assert.NotNil(t, p.Weight)
		assert.Equal(t, 100, p.Weight.Int())
		assert.NotNil(t, p.Sets)
		assert.NotNil(t, p.Reps)
		assert.Nil(t, p.Duration)
		assert.NotNil(t, p.Completed)
		assert.True(t, *p.Completed)
	})
}
