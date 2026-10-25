package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================
//  DivideRoundUp
// ============================================================

func TestDivideRoundUp(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		divisor int
		want    int
	}{
		{"точное деление", 9, 3, 3},
		{"с остатком", 10, 3, 4},
		{"с остатком 1", 7, 3, 3},
		{"делитель 1", 42, 1, 42},
		{"равные", 100, 100, 1},
		{"больше на 1", 101, 100, 2},
		{"0 / 100", 0, 100, 0},
		{"MaximumScaledLoad / 1", MaximumScaledLoad, 1, MaximumScaledLoad},
		{"MaximumScaledLoad / 10", MaximumScaledLoad, 10, 10_000_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DivideRoundUp(tt.value, tt.divisor))
		})
	}
}

// ============================================================
//  AddLoadWithLimit
// ============================================================

func TestAddLoadWithLimit(t *testing.T) {
	t.Run("обычное сложение", func(t *testing.T) {
		// current=0, load=10, difficulty=5 → 0 + 50 = 50
		assert.Equal(t, 50, AddLoadWithLimit(0, 10, 5, 1000))
	})

	t.Run("накопление", func(t *testing.T) {
		// 50 + 10×5 = 100
		assert.Equal(t, 100, AddLoadWithLimit(50, 10, 5, 1000))
	})

	t.Run("cap по limit", func(t *testing.T) {
		// 950 + 50 = 1000, не превышает
		assert.Equal(t, 1000, AddLoadWithLimit(950, 10, 5, 1000))
	})

	t.Run("cap по limit при превышении", func(t *testing.T) {
		// 990 + 50 = 1040 → cap до 1000
		assert.Equal(t, 1000, AddLoadWithLimit(990, 10, 5, 1000))
	})

	t.Run("current уже на limit — no-op", func(t *testing.T) {
		assert.Equal(t, 1000, AddLoadWithLimit(1000, 10, 5, 1000))
	})

	t.Run("current больше limit — возвращает limit", func(t *testing.T) {
		// теоретически не должно происходить, но защита
		assert.Equal(t, 100, AddLoadWithLimit(200, 10, 5, 100))
	})

	t.Run("load=0 → ничего не добавляет", func(t *testing.T) {
		assert.Equal(t, 42, AddLoadWithLimit(42, 0, 5, 1000))
	})

	t.Run("difficulty=1 — просто load", func(t *testing.T) {
		assert.Equal(t, 50, AddLoadWithLimit(40, 10, 1, 1000))
	})
}

// ============================================================
//  CalculateScore
// ============================================================

func TestCalculateScore(t *testing.T) {
	t.Run("нулевая нагрузка → 0", func(t *testing.T) {
		assert.Equal(t, 0, CalculateScore(0, 1))
		assert.Equal(t, 0, CalculateScore(0, 10))
	})

	t.Run("маленькая нагрузка → маленький score", func(t *testing.T) {
		// load=1, coef=1: score = (10^4 × 1 + 50_000) / 100_001 = 0
		assert.Equal(t, 0, CalculateScore(1, 1))

		// load=100, coef=1: score = (10^6 + 50_050) / 100_100 ≈ 10
		assert.Equal(t, 10, CalculateScore(100, 1))
	})

	t.Run("score всегда < MaximumWorkoutScore", func(t *testing.T) {
		// Даже при максимальной scaledLoad score асимптотически стремится к 10000,
		// но никогда не достигает
		effectiveLoad := MaximumScaledLoad / MaxScoreCoefficient // coef=10 → load=10^7
		score := CalculateScore(effectiveLoad, MaxScoreCoefficient)

		assert.Less(t, score, MaximumWorkoutScore)
		assert.Greater(t, score, 9000, "при насыщении score близко к 10000")
	})

	t.Run("монотонность по load", func(t *testing.T) {
		// больше load → больше (или равный) score
		prev := 0
		for _, load := range []int{0, 10, 100, 1000, 10000, 100000} {
			score := CalculateScore(load, 5)
			assert.GreaterOrEqual(t, score, prev, "load=%d", load)
			prev = score
		}
	})

	t.Run("монотонность по coefficient", func(t *testing.T) {
		// при одинаковой load — больше coefficient → больше score
		load := 1000
		prev := 0
		for _, coef := range []int{1, 2, 5, 10} {
			score := CalculateScore(load, coef)
			assert.GreaterOrEqual(t, score, prev, "coef=%d", coef)
			prev = score
		}
	})

	t.Run("score не паникует на больших значениях", func(t *testing.T) {
		// защита от overflow: int64 внутри
		assert.NotPanics(t, func() {
			CalculateScore(MaximumScaledLoad, MaxScoreCoefficient)
		})
	})
}

// ============================================================
//  saturatedScore (прямой тест приватной функции)
// ============================================================

func TestSaturatedScore(t *testing.T) {
	t.Run("нулевая нагрузка → 0", func(t *testing.T) {
		assert.Equal(t, 0, saturatedScore(0, 1))
		assert.Equal(t, 0, saturatedScore(0, 10))
	})

	t.Run("совпадает с CalculateScore", func(t *testing.T) {
		tests := []struct {
			load int
			coef int
		}{
			{100, 1},
			{1000, 5},
			{10000, 10},
			{100000, 3},
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				assert.Equal(t, CalculateScore(tt.load, tt.coef), saturatedScore(tt.load, tt.coef))
			})
		}
	})
}
