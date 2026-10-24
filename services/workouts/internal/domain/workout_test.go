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

// newTestWorkout создаёт тренировку с coefficient=5, status=planned.
func newTestWorkout(t *testing.T) Workout {
	t.Helper()
	return NewWorkout(uuid.New(), 5)
}

// newWorkoutInStatus создаёт тренировку в заданном статусе через RestoreWorkout.
// Используется, чтобы тестировать переходы из нетривиальных состояний.
func newWorkoutInStatus(
	t *testing.T,
	status Status,
	startedAt, completedAt *time.Time,
) Workout {
	t.Helper()
	return RestoreWorkout(
		uuid.New(),
		0,
		uuid.New(),
		status,
		startedAt,
		completedAt,
		0,
		nil,
		5,
		false,
		time.Now().UTC(),
		time.Now().UTC(),
	)
}

// ============================================================
//  NewWorkout / RestoreWorkout
// ============================================================

func TestNewWorkout(t *testing.T) {
	userID := uuid.New()
	w := NewWorkout(userID, 5)

	t.Run("генерирует id", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, w.ID())
	})

	t.Run("поля по умолчанию", func(t *testing.T) {
		assert.Equal(t, 0, w.Version())
		assert.Equal(t, userID, w.UserID())
		assert.Equal(t, StatusPlanned, w.Status())
		assert.Nil(t, w.StartedAt())
		assert.Nil(t, w.CompletedAt())
		assert.Equal(t, 0, w.WorkoutScore())
		assert.Nil(t, w.Intensity())
		assert.Equal(t, 5, w.PersonalScoreCoefficient())
		assert.False(t, w.IsDeleted())
	})

	t.Run("createdAt и updatedAt заполнены", func(t *testing.T) {
		assert.False(t, w.CreatedAt().IsZero())
		assert.False(t, w.UpdatedAt().IsZero())
		assert.Equal(t, w.CreatedAt(), w.UpdatedAt())
	})

	t.Run("id уникален между вызовами", func(t *testing.T) {
		other := NewWorkout(userID, 5)
		assert.NotEqual(t, w.ID(), other.ID())
	})
}

func TestRestoreWorkout(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	startedAt := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 10, 24, 11, 0, 0, 0, time.UTC)
	intensity, _ := NewIntensity(7)
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	t.Run("восстанавливает все поля", func(t *testing.T) {
		w := RestoreWorkout(
			id, 3, userID, StatusCompleted,
			&startedAt, &completedAt,
			100, &intensity, 8,
			false, createdAt, updatedAt,
		)

		assert.Equal(t, id, w.ID())
		assert.Equal(t, 3, w.Version())
		assert.Equal(t, userID, w.UserID())
		assert.Equal(t, StatusCompleted, w.Status())
		assert.Equal(t, &startedAt, w.StartedAt())
		assert.Equal(t, &completedAt, w.CompletedAt())
		assert.Equal(t, 100, w.WorkoutScore())
		require.NotNil(t, w.Intensity())
		assert.Equal(t, intensity, *w.Intensity())
		assert.Equal(t, 8, w.PersonalScoreCoefficient())
		assert.False(t, w.IsDeleted())
		assert.Equal(t, createdAt, w.CreatedAt())
		assert.Equal(t, updatedAt, w.UpdatedAt())
	})

	t.Run("удалённая тренировка", func(t *testing.T) {
		w := RestoreWorkout(
			id, 0, userID, StatusPlanned,
			nil, nil,
			0, nil, 5,
			true, createdAt, updatedAt,
		)

		assert.True(t, w.IsDeleted())
	})
}

// ============================================================
//  CanModifyWorkoutExercise
// ============================================================

func TestWorkout_CanModifyWorkoutExercise(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusPlanned, true},
		{StatusInProgress, true},
		{StatusCompleted, false},
		{StatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			w := newWorkoutInStatus(t, tt.status, nil, nil)
			assert.Equal(t, tt.want, w.CanModifyWorkoutExercise())
		})
	}
}

// ============================================================
//  ApplyPatch — intensity
// ============================================================

func TestWorkout_ApplyPatch_Intensity(t *testing.T) {
	t.Run("пустой патч — nothing changed", func(t *testing.T) {
		w := newTestWorkout(t)
		beforeUpdatedAt := w.UpdatedAt()

		err := w.ApplyPatch(NewWorkoutPatch(nil, nullableIntensityNone()))
		require.NoError(t, err)

		assert.Nil(t, w.Intensity())
		assert.Equal(t, beforeUpdatedAt, w.UpdatedAt(), "updatedAt не должен обновляться")
	})

	t.Run("установить intensity", func(t *testing.T) {
		w := newTestWorkout(t)
		beforeUpdatedAt := w.UpdatedAt()
		i, _ := NewIntensity(7)

		err := w.ApplyPatch(NewWorkoutPatch(nil, nullableIntensitySet(&i)))
		require.NoError(t, err)

		require.NotNil(t, w.Intensity())
		assert.Equal(t, 7, w.Intensity().Int())
		assert.True(t, w.UpdatedAt().After(beforeUpdatedAt))
	})

	t.Run("сбросить intensity в nil", func(t *testing.T) {
		i, _ := NewIntensity(7)
		w := newTestWorkout(t)
		require.NoError(t, w.ApplyPatch(NewWorkoutPatch(nil, nullableIntensitySet(&i))))

		err := w.ApplyPatch(NewWorkoutPatch(nil, nullableIntensityClear()))
		require.NoError(t, err)

		assert.Nil(t, w.Intensity())
	})
}

// ============================================================
//  ApplyPatch — status (state machine)
// ============================================================

func TestWorkout_ApplyPatch_StatusTransitions(t *testing.T) {
	tests := []struct {
		name       string
		from       Status
		to         Status
		wantErr    bool
		wantStatus Status
	}{
		{"planned → in_progress", StatusPlanned, StatusInProgress, false, StatusInProgress},
		{"planned → cancelled", StatusPlanned, StatusCancelled, false, StatusCancelled},
		{"planned → completed запрещён", StatusPlanned, StatusCompleted, true, StatusPlanned},
		{"planned → planned (self)", StatusPlanned, StatusPlanned, false, StatusPlanned},

		{"in_progress → completed", StatusInProgress, StatusCompleted, false, StatusCompleted},
		{"in_progress → cancelled", StatusInProgress, StatusCancelled, false, StatusCancelled},
		{"in_progress → planned запрещён", StatusInProgress, StatusPlanned, true, StatusInProgress},
		{"in_progress → in_progress (self)", StatusInProgress, StatusInProgress, false, StatusInProgress},

		{"completed → anything запрещён", StatusCompleted, StatusCancelled, true, StatusCompleted},
		{"completed → completed (self)", StatusCompleted, StatusCompleted, false, StatusCompleted},

		{"cancelled → anything запрещён", StatusCancelled, StatusPlanned, true, StatusCancelled},
		{"cancelled → cancelled (self)", StatusCancelled, StatusCancelled, false, StatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorkoutInStatus(t, tt.from, nil, nil)
			beforeUpdatedAt := w.UpdatedAt()

			next := tt.to
			err := w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensityNone()))

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidStatusTransition)
				assert.Equal(t, tt.from, w.Status(), "при ошибке статус не должен меняться")
				assert.Equal(t, beforeUpdatedAt, w.UpdatedAt(), "при ошибке updatedAt не должен трогаться")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, w.Status())

			if tt.from != tt.to {
				assert.True(t, w.UpdatedAt().After(beforeUpdatedAt))
			} else {
				assert.Equal(t, beforeUpdatedAt, w.UpdatedAt(), "self-transition — no-op, updatedAt не трогается")
			}
		})
	}
}

func TestWorkout_ApplyPatch_StartedAtInvariant(t *testing.T) {
	t.Run("planned → in_progress ставит started_at", func(t *testing.T) {
		w := newTestWorkout(t)
		require.Nil(t, w.StartedAt())

		next := StatusInProgress
		require.NoError(t, w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensityNone())))

		require.NotNil(t, w.StartedAt())
		assert.WithinDuration(t, time.Now().UTC(), *w.StartedAt(), 2*time.Second)
	})

	t.Run("started_at не перезаписывается при повторном переходе", func(t *testing.T) {
		w := newTestWorkout(t)
		next := StatusInProgress
		require.NoError(t, w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensityNone())))
		firstStartedAt := *w.StartedAt()

		// self-transition — started_at не должен измениться
		require.NoError(t, w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensityNone())))
		assert.Equal(t, firstStartedAt, *w.StartedAt())
	})
}

func TestWorkout_ApplyPatch_CompletedAtInvariant(t *testing.T) {
	t.Run("in_progress → completed ставит completed_at", func(t *testing.T) {
		w := newWorkoutInStatus(t, StatusInProgress, nil, nil)
		require.Nil(t, w.CompletedAt())

		next := StatusCompleted
		require.NoError(t, w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensityNone())))

		require.NotNil(t, w.CompletedAt())
		assert.WithinDuration(t, time.Now().UTC(), *w.CompletedAt(), 2*time.Second)
	})

	t.Run("completed_at не перезаписывается", func(t *testing.T) {
		existing := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		started := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

		w := RestoreWorkout(
			uuid.New(), 0, uuid.New(), StatusCompleted,
			&started, &existing,
			0, nil, 5, false,
			time.Now().UTC(), time.Now().UTC(),
		)

		next := StatusCompleted
		require.NoError(t, w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensityNone())))

		assert.Equal(t, existing, *w.CompletedAt())
	})
}

func TestWorkout_ApplyPatch_Atomicity(t *testing.T) {
	t.Run("невалидный status не применяет intensity (валидация до мутации)", func(t *testing.T) {
		w := newWorkoutInStatus(t, StatusCompleted, nil, nil)
		beforeUpdatedAt := w.UpdatedAt()

		i, _ := NewIntensity(9)
		next := StatusPlanned // из completed запрещён

		err := w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensitySet(&i)))

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidStatusTransition)
		assert.Nil(t, w.Intensity(), "intensity не должен примениться при ошибке status")
		assert.Equal(t, StatusCompleted, w.Status())
		assert.Equal(t, beforeUpdatedAt, w.UpdatedAt())
	})

	t.Run("валидный status + intensity — оба применяются", func(t *testing.T) {
		w := newTestWorkout(t)
		i, _ := NewIntensity(8)
		next := StatusInProgress

		err := w.ApplyPatch(NewWorkoutPatch(&next, nullableIntensitySet(&i)))
		require.NoError(t, err)

		assert.Equal(t, StatusInProgress, w.Status())
		require.NotNil(t, w.Intensity())
		assert.Equal(t, 8, w.Intensity().Int())
	})
}

// ============================================================
//  SetWorkoutScore
// ============================================================

func TestWorkout_SetWorkoutScore(t *testing.T) {
	t.Run("устанавливает score", func(t *testing.T) {
		w := newTestWorkout(t)
		beforeUpdatedAt := w.UpdatedAt()
		time.Sleep(1 * time.Millisecond) // чтобы updatedAt гарантированно отличался

		w.SetWorkoutScore(1500)

		assert.Equal(t, 1500, w.WorkoutScore())
		assert.True(t, w.UpdatedAt().After(beforeUpdatedAt))
	})

	t.Run("перезаписывает предыдущее значение", func(t *testing.T) {
		w := newTestWorkout(t)
		w.SetWorkoutScore(100)
		w.SetWorkoutScore(500)

		assert.Equal(t, 500, w.WorkoutScore())
	})
}