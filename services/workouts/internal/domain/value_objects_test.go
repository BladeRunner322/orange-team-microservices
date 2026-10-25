package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
//  Weight
// ============================================================

func TestNewWeight(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, raw := range []int{MinWeight, 500, MaxWeight} {
			w, err := NewWeight(raw)
			require.NoError(t, err)
			assert.Equal(t, raw, w.Int())
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 — минимум", func(t *testing.T) {
			w, err := NewWeight(1)
			require.NoError(t, err)
			assert.Equal(t, 1, w.Int())
		})

		t.Run("1000 — максимум", func(t *testing.T) {
			w, err := NewWeight(1000)
			require.NoError(t, err)
			assert.Equal(t, 1000, w.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewWeight(0)
			assert.ErrorIs(t, err, ErrInvalidWeight)
		})

		t.Run("1001 — выше максимума", func(t *testing.T) {
			_, err := NewWeight(1001)
			assert.ErrorIs(t, err, ErrInvalidWeight)
		})

		t.Run("отрицательное", func(t *testing.T) {
			_, err := NewWeight(-1)
			assert.ErrorIs(t, err, ErrInvalidWeight)
		})
	})
}

// ============================================================
//  Sets
// ============================================================

func TestNewSets(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, raw := range []int{MinSets, 50, MaxSets} {
			s, err := NewSets(raw)
			require.NoError(t, err)
			assert.Equal(t, raw, s.Int())
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 — минимум", func(t *testing.T) {
			s, err := NewSets(1)
			require.NoError(t, err)
			assert.Equal(t, 1, s.Int())
		})

		t.Run("100 — максимум", func(t *testing.T) {
			s, err := NewSets(100)
			require.NoError(t, err)
			assert.Equal(t, 100, s.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewSets(0)
			assert.ErrorIs(t, err, ErrInvalidSets)
		})

		t.Run("101 — выше максимума", func(t *testing.T) {
			_, err := NewSets(101)
			assert.ErrorIs(t, err, ErrInvalidSets)
		})
	})
}

// ============================================================
//  Reps
// ============================================================

func TestNewReps(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, raw := range []int{MinReps, 500, MaxReps} {
			r, err := NewReps(raw)
			require.NoError(t, err)
			assert.Equal(t, raw, r.Int())
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 — минимум", func(t *testing.T) {
			r, err := NewReps(1)
			require.NoError(t, err)
			assert.Equal(t, 1, r.Int())
		})

		t.Run("1000 — максимум", func(t *testing.T) {
			r, err := NewReps(1000)
			require.NoError(t, err)
			assert.Equal(t, 1000, r.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewReps(0)
			assert.ErrorIs(t, err, ErrInvalidReps)
		})

		t.Run("1001 — выше максимума", func(t *testing.T) {
			_, err := NewReps(1001)
			assert.ErrorIs(t, err, ErrInvalidReps)
		})
	})
}

// ============================================================
//  Duration
// ============================================================

func TestNewDuration(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, raw := range []int{MinDuration, 3600, MaxDuration} {
			d, err := NewDuration(raw)
			require.NoError(t, err)
			assert.Equal(t, raw, d.Int())
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 секунда — минимум", func(t *testing.T) {
			d, err := NewDuration(1)
			require.NoError(t, err)
			assert.Equal(t, 1, d.Int())
		})

		t.Run("86400 секунд (24 часа) — максимум", func(t *testing.T) {
			d, err := NewDuration(86400)
			require.NoError(t, err)
			assert.Equal(t, 86400, d.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewDuration(0)
			assert.ErrorIs(t, err, ErrInvalidDuration)
		})

		t.Run("86401 — выше максимума", func(t *testing.T) {
			_, err := NewDuration(86401)
			assert.ErrorIs(t, err, ErrInvalidDuration)
		})
	})
}

// ============================================================
//  Difficulty
// ============================================================

func TestNewDifficulty(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, raw := range []int{MinDifficulty, 5, MaxDifficulty} {
			d, err := NewDifficulty(raw)
			require.NoError(t, err)
			assert.Equal(t, raw, d.Int())
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 — минимум", func(t *testing.T) {
			d, err := NewDifficulty(1)
			require.NoError(t, err)
			assert.Equal(t, 1, d.Int())
		})

		t.Run("10 — максимум", func(t *testing.T) {
			d, err := NewDifficulty(10)
			require.NoError(t, err)
			assert.Equal(t, 10, d.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewDifficulty(0)
			assert.ErrorIs(t, err, ErrInvalidDifficulty)
		})

		t.Run("11 — выше максимума", func(t *testing.T) {
			_, err := NewDifficulty(11)
			assert.ErrorIs(t, err, ErrInvalidDifficulty)
		})
	})
}

// ============================================================
//  Intensity
// ============================================================

func TestNewIntensity(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, raw := range []int{MinIntensity, 5, MaxIntensity} {
			i, err := NewIntensity(raw)
			require.NoError(t, err)
			assert.Equal(t, raw, i.Int())
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 — минимум", func(t *testing.T) {
			i, err := NewIntensity(1)
			require.NoError(t, err)
			assert.Equal(t, 1, i.Int())
		})

		t.Run("10 — максимум", func(t *testing.T) {
			i, err := NewIntensity(10)
			require.NoError(t, err)
			assert.Equal(t, 10, i.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewIntensity(0)
			assert.ErrorIs(t, err, ErrInvalidIntensity)
		})

		t.Run("11 — выше максимума", func(t *testing.T) {
			_, err := NewIntensity(11)
			assert.ErrorIs(t, err, ErrInvalidIntensity)
		})
	})
}

// ============================================================
//  Status
// ============================================================

func TestNewStatus(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		tests := []struct {
			raw  string
			want Status
		}{
			{"planned", StatusPlanned},
			{"in_progress", StatusInProgress},
			{"completed", StatusCompleted},
			{"cancelled", StatusCancelled},
		}

		for _, tt := range tests {
			t.Run(tt.raw, func(t *testing.T) {
				s, err := NewStatus(tt.raw)
				require.NoError(t, err)
				assert.Equal(t, tt.want, s)
			})
		}
	})

	t.Run("невалидные значения", func(t *testing.T) {
		for _, raw := range []string{"", "PLANNED", "in-progress", "done", "active"} {
			t.Run(raw, func(t *testing.T) {
				_, err := NewStatus(raw)
				assert.ErrorIs(t, err, ErrInvalidStatus)
			})
		}
	})
}

func TestStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusPlanned, false},
		{StatusInProgress, false},
		{StatusCompleted, true},
		{StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsTerminal())
		})
	}
}

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "planned", StatusPlanned.String())
	assert.Equal(t, "in_progress", StatusInProgress.String())
	assert.Equal(t, "completed", StatusCompleted.String())
	assert.Equal(t, "cancelled", StatusCancelled.String())
}

// ============================================================
//  ExerciseType
// ============================================================

func TestNewExerciseType(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		t.Run("weight", func(t *testing.T) {
			et, err := NewExerciseType("weight")
			require.NoError(t, err)
			assert.Equal(t, ExerciseTypeWeight, et)
		})

		t.Run("duration", func(t *testing.T) {
			et, err := NewExerciseType("duration")
			require.NoError(t, err)
			assert.Equal(t, ExerciseTypeDuration, et)
		})
	})

	t.Run("невалидные значения", func(t *testing.T) {
		tests := []string{"", "WEIGHT", "Weight", "cardio", "strength", " weight"}
		for _, raw := range tests {
			t.Run(raw, func(t *testing.T) {
				_, err := NewExerciseType(raw)
				assert.ErrorIs(t, err, ErrInvalidExerciseType)
			})
		}
	})
}

func TestExerciseType_String(t *testing.T) {
	assert.Equal(t, "weight", ExerciseTypeWeight.String())
	assert.Equal(t, "duration", ExerciseTypeDuration.String())
}

// ============================================================
//  Защита от опечаток в константах
// ============================================================

// TestValueObjectConstantsGuard проверяет, что публичные константы
// не изменились случайно. Это контракт с БД (CHECK-констрейнты в миграциях)
// и с proto (enum-like строки в ExerciseType, Status).
func TestValueObjectConstantsGuard(t *testing.T) {
	t.Run("Status значения", func(t *testing.T) {
		assert.Equal(t, "planned", string(StatusPlanned))
		assert.Equal(t, "in_progress", string(StatusInProgress))
		assert.Equal(t, "completed", string(StatusCompleted))
		assert.Equal(t, "cancelled", string(StatusCancelled))
	})

	t.Run("ExerciseType значения", func(t *testing.T) {
		assert.Equal(t, "weight", string(ExerciseTypeWeight))
		assert.Equal(t, "duration", string(ExerciseTypeDuration))
	})

	t.Run("диапазоны VO", func(t *testing.T) {
		assert.Equal(t, 1, MinWeight)
		assert.Equal(t, 1000, MaxWeight)
		assert.Equal(t, 1, MinSets)
		assert.Equal(t, 100, MaxSets)
		assert.Equal(t, 1, MinReps)
		assert.Equal(t, 1000, MaxReps)
		assert.Equal(t, 1, MinDuration)
		assert.Equal(t, 86400, MaxDuration)
		assert.Equal(t, 1, MinDifficulty)
		assert.Equal(t, 10, MaxDifficulty)
		assert.Equal(t, 1, MinIntensity)
		assert.Equal(t, 10, MaxIntensity)
	})
}
