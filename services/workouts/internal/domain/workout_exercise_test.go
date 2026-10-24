package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
//  Хелперы
// ============================================================

func ptrWeight(t *testing.T, raw int) *Weight {
	t.Helper()
	w, err := NewWeight(raw)
	require.NoError(t, err)
	return &w
}

func ptrSets(t *testing.T, raw int) *Sets {
	t.Helper()
	s, err := NewSets(raw)
	require.NoError(t, err)
	return &s
}

func ptrReps(t *testing.T, raw int) *Reps {
	t.Helper()
	r, err := NewReps(raw)
	require.NoError(t, err)
	return &r
}

func ptrDuration(t *testing.T, raw int) *Duration {
	t.Helper()
	d, err := NewDuration(raw)
	require.NoError(t, err)
	return &d
}

func ptrDifficulty(t *testing.T, raw int) *Difficulty {
	t.Helper()
	d, err := NewDifficulty(raw)
	require.NoError(t, err)
	return &d
}

func ptrBool(v bool) *bool { return &v }

// newTestWorkoutExerciseWeight — валидное weight-упражнение.
func newTestWorkoutExerciseWeight(t *testing.T) WorkoutExercise {
	t.Helper()
	e, err := NewWorkoutExercise(
		uuid.New(), uuid.New(), *ptrDifficulty(t, 5),
		ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil,
		false, ExerciseTypeWeight,
	)
	require.NoError(t, err)
	return e
}

// newTestWorkoutExerciseDuration — валидное duration-упражнение.
func newTestWorkoutExerciseDuration(t *testing.T) WorkoutExercise {
	t.Helper()
	e, err := NewWorkoutExercise(
		uuid.New(), uuid.New(), *ptrDifficulty(t, 3),
		nil, nil, nil, ptrDuration(t, 600),
		false, ExerciseTypeDuration,
	)
	require.NoError(t, err)
	return e
}

// ============================================================
//  NewWorkoutExercise — валидация и расчёт exercise_load
// ============================================================

func TestNewWorkoutExercise_Weight(t *testing.T) {
	workoutID := uuid.New()
	exerciseID := uuid.New()
	difficulty := *ptrDifficulty(t, 5)
	weight := ptrWeight(t, 100)
	sets := ptrSets(t, 4)
	reps := ptrReps(t, 10)

	e, err := NewWorkoutExercise(
		workoutID, exerciseID, difficulty,
		weight, sets, reps, nil,
		false, ExerciseTypeWeight,
	)
	require.NoError(t, err)

	t.Run("поля соответствуют переданным", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, e.ID())
		assert.Equal(t, 0, e.Version())
		assert.Equal(t, workoutID, e.WorkoutID())
		assert.Equal(t, exerciseID, e.ExerciseID())
		assert.Equal(t, difficulty, e.Difficulty())
		assert.Equal(t, weight, e.Weight())
		assert.Equal(t, sets, e.Sets())
		assert.Equal(t, reps, e.Reps())
		assert.Nil(t, e.Duration())
		assert.False(t, e.Completed())
	})

	t.Run("exercise_load = weight × reps × sets", func(t *testing.T) {
		// 100 × 10 × 4 = 4000
		assert.Equal(t, 4000, e.ExerciseLoad())
	})

	t.Run("created_at == updated_at", func(t *testing.T) {
		assert.Equal(t, e.CreatedAt(), e.UpdatedAt())
		assert.False(t, e.CreatedAt().IsZero())
	})
}

func TestNewWorkoutExercise_Duration(t *testing.T) {
	difficulty := *ptrDifficulty(t, 3)
	duration := ptrDuration(t, 600)

	e, err := NewWorkoutExercise(
		uuid.New(), uuid.New(), difficulty,
		nil, nil, nil, duration,
		false, ExerciseTypeDuration,
	)
	require.NoError(t, err)

	t.Run("поля", func(t *testing.T) {
		assert.Nil(t, e.Weight())
		assert.Nil(t, e.Sets())
		assert.Nil(t, e.Reps())
		assert.Equal(t, duration, e.Duration())
	})

	t.Run("exercise_load = duration", func(t *testing.T) {
		assert.Equal(t, 600, e.ExerciseLoad())
	})
}

func TestNewWorkoutExercise_ValidationErrors(t *testing.T) {
	workoutID := uuid.New()
	exerciseID := uuid.New()
	difficulty := *ptrDifficulty(t, 5)

	t.Run("weight-тип без weight", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			nil, ptrSets(t, 4), ptrReps(t, 10), nil,
			false, ExerciseTypeWeight,
		)
		assert.ErrorIs(t, err, ErrInvalidWorkoutExerciseForType)
	})

	t.Run("weight-тип без sets", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			ptrWeight(t, 100), nil, ptrReps(t, 10), nil,
			false, ExerciseTypeWeight,
		)
		assert.ErrorIs(t, err, ErrInvalidWorkoutExerciseForType)
	})

	t.Run("weight-тип без reps", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			ptrWeight(t, 100), ptrSets(t, 4), nil, nil,
			false, ExerciseTypeWeight,
		)
		assert.ErrorIs(t, err, ErrInvalidWorkoutExerciseForType)
	})

	t.Run("weight-тип с duration", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), ptrDuration(t, 600),
			false, ExerciseTypeWeight,
		)
		assert.ErrorIs(t, err, ErrInvalidWorkoutExerciseForType)
	})

	t.Run("duration-тип без duration", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			nil, nil, nil, nil,
			false, ExerciseTypeDuration,
		)
		assert.ErrorIs(t, err, ErrInvalidWorkoutExerciseForType)
	})

	t.Run("duration-тип с weight", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			ptrWeight(t, 100), nil, nil, ptrDuration(t, 600),
			false, ExerciseTypeDuration,
		)
		assert.ErrorIs(t, err, ErrInvalidWorkoutExerciseForType)
	})

	t.Run("неизвестный type", func(t *testing.T) {
		_, err := NewWorkoutExercise(
			workoutID, exerciseID, difficulty,
			ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), nil,
			false, ExerciseType("bogus"),
		)
		assert.ErrorIs(t, err, ErrInvalidExerciseType)
	})
}

// ============================================================
//  RestoreWorkoutExercise
// ============================================================

func TestRestoreWorkoutExercise(t *testing.T) {
	id := uuid.New()
	workoutID := uuid.New()
	exerciseID := uuid.New()
	difficulty := *ptrDifficulty(t, 5)
	weight := ptrWeight(t, 100)
	sets := ptrSets(t, 4)
	reps := ptrReps(t, 10)
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	e := RestoreWorkoutExercise(
		id, 2, workoutID, exerciseID, difficulty,
		weight, sets, reps, nil,
		true, 4000,
		createdAt, updatedAt,
	)

	assert.Equal(t, id, e.ID())
	assert.Equal(t, 2, e.Version())
	assert.Equal(t, workoutID, e.WorkoutID())
	assert.Equal(t, exerciseID, e.ExerciseID())
	assert.Equal(t, difficulty, e.Difficulty())
	assert.Equal(t, weight, e.Weight())
	assert.Equal(t, sets, e.Sets())
	assert.Equal(t, reps, e.Reps())
	assert.Nil(t, e.Duration())
	assert.True(t, e.Completed())
	assert.Equal(t, 4000, e.ExerciseLoad())
	assert.Equal(t, createdAt, e.CreatedAt())
	assert.Equal(t, updatedAt, e.UpdatedAt())
}

// ============================================================
//  ApplyPatch
// ============================================================

func TestWorkoutExercise_ApplyPatch_NoChanges(t *testing.T) {
	t.Run("пустой патч — no-op, updatedAt не трогается", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)
		beforeUpdatedAt := e.UpdatedAt()
		beforeLoad := e.ExerciseLoad()

		err := e.ApplyPatch(NewWorkoutExercisePatch(nil, nil, nil, nil, nil), ExerciseTypeWeight)
		require.NoError(t, err)

		assert.Equal(t, beforeLoad, e.ExerciseLoad())
		assert.Equal(t, beforeUpdatedAt, e.UpdatedAt())
	})
}

func TestWorkoutExercise_ApplyPatch_Completed(t *testing.T) {
	e := newTestWorkoutExerciseWeight(t)
	beforeUpdatedAt := e.UpdatedAt()
	time.Sleep(1 * time.Millisecond)

	err := e.ApplyPatch(
		NewWorkoutExercisePatch(nil, nil, nil, nil, ptrBool(true)),
		ExerciseTypeWeight,
	)
	require.NoError(t, err)

	assert.True(t, e.Completed())
	assert.True(t, e.UpdatedAt().After(beforeUpdatedAt))
}

func TestWorkoutExercise_ApplyPatch_WeightRecalculatesLoad(t *testing.T) {
	t.Run("weight меняется — load пересчитывается", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)
		assert.Equal(t, 4000, e.ExerciseLoad())

		newWeight := ptrWeight(t, 200)
		err := e.ApplyPatch(
			NewWorkoutExercisePatch(newWeight, nil, nil, nil, nil),
			ExerciseTypeWeight,
		)
		require.NoError(t, err)

		assert.Equal(t, 200, e.Weight().Int())
		// 200 × 10 × 4 = 8000
		assert.Equal(t, 8000, e.ExerciseLoad())
	})

	t.Run("sets меняется — load пересчитывается", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)

		newSets := ptrSets(t, 5)
		err := e.ApplyPatch(
			NewWorkoutExercisePatch(nil, newSets, nil, nil, nil),
			ExerciseTypeWeight,
		)
		require.NoError(t, err)

		// 100 × 10 × 5 = 5000
		assert.Equal(t, 5000, e.ExerciseLoad())
	})

	t.Run("reps меняется — load пересчитывается", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)

		newReps := ptrReps(t, 12)
		err := e.ApplyPatch(
			NewWorkoutExercisePatch(nil, nil, newReps, nil, nil),
			ExerciseTypeWeight,
		)
		require.NoError(t, err)

		// 100 × 12 × 4 = 4800
		assert.Equal(t, 4800, e.ExerciseLoad())
	})
}

func TestWorkoutExercise_ApplyPatch_DurationRecalculatesLoad(t *testing.T) {
	e := newTestWorkoutExerciseDuration(t)
	assert.Equal(t, 600, e.ExerciseLoad())

	newDuration := ptrDuration(t, 1200)
	err := e.ApplyPatch(
		NewWorkoutExercisePatch(nil, nil, nil, newDuration, nil),
		ExerciseTypeDuration,
	)
	require.NoError(t, err)

	assert.Equal(t, 1200, e.Duration().Int())
	assert.Equal(t, 1200, e.ExerciseLoad())
}

func TestWorkoutExercise_ApplyPatch_WrongTypeFieldsIgnored(t *testing.T) {
	// Патч с weight на duration-типе: weight не применится,
	// потому что switch идёт по ExerciseTypeDuration и weight не обрабатывает.
	e := newTestWorkoutExerciseDuration(t)

	err := e.ApplyPatch(
		NewWorkoutExercisePatch(ptrWeight(t, 100), nil, nil, nil, nil),
		ExerciseTypeDuration,
	)
	require.NoError(t, err)

	assert.Nil(t, e.Weight(), "weight не должен примениться на duration-типе")
	assert.Equal(t, 600, e.ExerciseLoad(), "load не должен измениться")
}

// ============================================================
//  validateForWorkoutExerciseType (прямой тест приватного метода)
// ============================================================

func TestValidateForWorkoutExerciseType(t *testing.T) {
	weight, sets, reps, duration := ptrWeight(t, 100), ptrSets(t, 4), ptrReps(t, 10), ptrDuration(t, 600)

	t.Run("weight-тип: валидные поля", func(t *testing.T) {
		e := RestoreWorkoutExercise(
			uuid.New(), 0, uuid.New(), uuid.New(), *ptrDifficulty(t, 5),
			weight, sets, reps, nil,
			false, 4000, time.Now(), time.Now(),
		)
		assert.NoError(t, e.validateForWorkoutExerciseType(ExerciseTypeWeight))
	})

	t.Run("duration-тип: валидные поля", func(t *testing.T) {
		e := RestoreWorkoutExercise(
			uuid.New(), 0, uuid.New(), uuid.New(), *ptrDifficulty(t, 5),
			nil, nil, nil, duration,
			false, 600, time.Now(), time.Now(),
		)
		assert.NoError(t, e.validateForWorkoutExerciseType(ExerciseTypeDuration))
	})

	t.Run("неизвестный type", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)
		err := e.validateForWorkoutExerciseType(ExerciseType("bogus"))
		assert.ErrorIs(t, err, ErrInvalidExerciseType)
	})
}

// ============================================================
//  recalculateExerciseLoad (прямой тест приватного метода)
// ============================================================

func TestRecalculateExerciseLoad(t *testing.T) {
	t.Run("weight: w × r × s", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)
		e.recalculateExerciseLoad(ExerciseTypeWeight)
		assert.Equal(t, 4000, e.ExerciseLoad())
	})

	t.Run("duration: duration", func(t *testing.T) {
		e := newTestWorkoutExerciseDuration(t)
		e.recalculateExerciseLoad(ExerciseTypeDuration)
		assert.Equal(t, 600, e.ExerciseLoad())
	})

	t.Run("пересчёт после изменения полей", func(t *testing.T) {
		e := newTestWorkoutExerciseWeight(t)
		*e.weight = *ptrWeight(t, 50)
		*e.reps = *ptrReps(t, 20)

		e.recalculateExerciseLoad(ExerciseTypeWeight)
		// 50 × 20 × 4 = 4000
		assert.Equal(t, 4000, e.ExerciseLoad())
	})
}