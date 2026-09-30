package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSex(t *testing.T) {
	t.Run("male", func(t *testing.T) {
		sex, err := NewSex("male")
		require.NoError(t, err)
		assert.Equal(t, SexMale, sex)
	})

	t.Run("female", func(t *testing.T) {
		sex, err := NewSex("female")
		require.NoError(t, err)
		assert.Equal(t, SexFemale, sex)
	})

	t.Run("пустая строка", func(t *testing.T) {
		_, err := NewSex("")
		assert.ErrorIs(t, err, ErrInvalidSex)
	})

	t.Run("неизвестное значение", func(t *testing.T) {
		_, err := NewSex("blabla")
		assert.ErrorIs(t, err, ErrInvalidSex)
	})

	t.Run("неверный регистр", func(t *testing.T) {
		_, err := NewSex("Male")
		assert.ErrorIs(t, err, ErrInvalidSex)
	})
}

func TestNewWeightFromGrams(t *testing.T) {
	t.Run("минимальный вес", func(t *testing.T) {
		w, err := NewWeightFromGrams(40_000)
		require.NoError(t, err)
		assert.Equal(t, 40_000, w.Grams())
	})

	t.Run("максимальный вес", func(t *testing.T) {
		w, err := NewWeightFromGrams(150_000)
		require.NoError(t, err)
		assert.Equal(t, 150_000, w.Grams())
	})

	t.Run("типичный вес", func(t *testing.T) {
		w, err := NewWeightFromGrams(80_500)
		require.NoError(t, err)
		assert.Equal(t, 80_500, w.Grams())
	})

	t.Run("ниже минимума", func(t *testing.T) {
		_, err := NewWeightFromGrams(39_900)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})

	t.Run("выше максимума", func(t *testing.T) {
		_, err := NewWeightFromGrams(150_100)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})

	t.Run("не кратно 100 граммам", func(t *testing.T) {
		_, err := NewWeightFromGrams(80_550)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})

	t.Run("ноль", func(t *testing.T) {
		_, err := NewWeightFromGrams(0)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})
}

func TestNewWeightFromKilograms(t *testing.T) {
	t.Run("80.5 кг → 80500 грамм", func(t *testing.T) {
		w, err := NewWeightFromKilograms(80.5)
		require.NoError(t, err)
		assert.Equal(t, 80_500, w.Grams())
	})

	t.Run("80.3 кг → округление до 80300 (Round)", func(t *testing.T) {
		w, err := NewWeightFromKilograms(80.3)
		require.NoError(t, err)
		// math.Round защищает от float-погрешности 80.3 * 1000
		assert.Equal(t, 80_300, w.Grams())
	})

	t.Run("80.0 кг → 80000 грамм", func(t *testing.T) {
		w, err := NewWeightFromKilograms(80.0)
		require.NoError(t, err)
		assert.Equal(t, 80_000, w.Grams())
	})

	t.Run("границы 40.0 и 150.0", func(t *testing.T) {
		for _, kg := range []float64{40.0, 150.0} {
			_, err := NewWeightFromKilograms(kg)
			assert.NoError(t, err, "kg=%v", kg)
		}
	})

	t.Run("ниже 40 кг", func(t *testing.T) {
		_, err := NewWeightFromKilograms(39.9)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})

	t.Run("выше 150 кг", func(t *testing.T) {
		_, err := NewWeightFromKilograms(150.1)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})

	t.Run("80.55 кг — не кратно 100 г", func(t *testing.T) {
		_, err := NewWeightFromKilograms(80.55)
		assert.ErrorIs(t, err, ErrInvalidWeight)
	})
}

func TestWeightRoundTrip(t *testing.T) {
	t.Run("Grams → Kilograms", func(t *testing.T) {
		w, err := NewWeightFromGrams(80_500)
		require.NoError(t, err)
		assert.Equal(t, 80.5, w.Kilograms())
	})

	t.Run("Kilograms → Grams", func(t *testing.T) {
		w, err := NewWeightFromKilograms(80.5)
		require.NoError(t, err)
		assert.Equal(t, 80_500, w.Grams())
	})
}

func TestNewHeight(t *testing.T) {
	t.Run("минимальный рост", func(t *testing.T) {
		h, err := NewHeight(140)
		require.NoError(t, err)
		assert.Equal(t, 140, h.Centimeters())
	})

	t.Run("максимальный рост", func(t *testing.T) {
		h, err := NewHeight(210)
		require.NoError(t, err)
		assert.Equal(t, 210, h.Centimeters())
	})

	t.Run("типичный рост", func(t *testing.T) {
		h, err := NewHeight(180)
		require.NoError(t, err)
		assert.Equal(t, 180, h.Centimeters())
	})

	t.Run("ниже минимума", func(t *testing.T) {
		_, err := NewHeight(139)
		assert.ErrorIs(t, err, ErrInvalidHeight)
	})

	t.Run("выше максимума", func(t *testing.T) {
		_, err := NewHeight(211)
		assert.ErrorIs(t, err, ErrInvalidHeight)
	})
}

func TestNewBirthDate(t *testing.T) {
	t.Run("типичная дата 1990-05-15", func(t *testing.T) {
		bd, err := NewBirthDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.Equal(t, "1990-05-15", bd.String())
	})

	t.Run("раньше 1900 года", func(t *testing.T) {
		_, err := NewBirthDate(time.Date(1850, 1, 1, 0, 0, 0, 0, time.UTC))
		assert.ErrorIs(t, err, ErrInvalidBirthDate)
	})

	t.Run("в будущем", func(t *testing.T) {
		future := time.Now().UTC().AddDate(0, 0, 1)
		_, err := NewBirthDate(future)
		assert.ErrorIs(t, err, ErrInvalidBirthDate)
	})

	t.Run("5 лет — меньше MinUserAge", func(t *testing.T) {
		fiveYearsAgo := time.Now().UTC().AddDate(-5, 0, 0)
		_, err := NewBirthDate(fiveYearsAgo)
		assert.ErrorIs(t, err, ErrInvalidBirthDate)
	})

	t.Run("20 лет — ок", func(t *testing.T) {
		twentyYearsAgo := time.Now().UTC().AddDate(-20, 0, 0)
		_, err := NewBirthDate(twentyYearsAgo)
		assert.NoError(t, err)
	})

	t.Run("ровно сегодня — не валидно (0 лет)", func(t *testing.T) {
		_, err := NewBirthDate(time.Now().UTC())
		assert.ErrorIs(t, err, ErrInvalidBirthDate)
	})
}
