package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewName(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"ровно на границе 3 руны", "Жим"},
			{"кириллица с пробелом", "Жим лёжа"},
			{"латиница с дефисом", "Push-up"},
			{"скобки", "Bench Press (dumbbell)"},
			{"знаки = и +", "A=B+C"},
			{"апостроф", "Exercise's name"},
			{"точка и запятая", "Упражнение 1. Базовое"},
			{"слэш", "Бег/ходьба"},
			{"цифры", "Упражнение 42"},
			{"буква ё", "Ёжик"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				name, err := NewName(tt.input)

				require.NoError(t, err)
				assert.Equal(t, tt.input, name.String())
			})
		}
	})

	t.Run("границы длины", func(t *testing.T) {
		t.Run("2 руны — меньше минимума", func(t *testing.T) {
			_, err := NewName("Жм")
			assert.ErrorIs(t, err, ErrInvalidName)
		})

		t.Run("3 руны — ровно минимум", func(t *testing.T) {
			name, err := NewName("Жим")
			require.NoError(t, err)
			assert.Equal(t, "Жим", name.String())
		})

		t.Run("100 рун — ровно максимум", func(t *testing.T) {
			input := strings.Repeat("a", 100)
			name, err := NewName(input)
			require.NoError(t, err)
			assert.Equal(t, input, name.String())
		})

		t.Run("101 руна — больше максимума", func(t *testing.T) {
			_, err := NewName(strings.Repeat("a", 101))
			assert.ErrorIs(t, err, ErrInvalidName)
		})
	})

	t.Run("длина считается в рунах, а не в байтах", func(t *testing.T) {
		// "абв" — 3 руны, 6 байт. Если бы использовался len(), тест бы
		// упал на нижней границе. Проверяет utf8.RuneCountInString.
		name, err := NewName("абв")

		require.NoError(t, err)
		assert.Equal(t, "абв", name.String())
	})

	t.Run("trim пробелов", func(t *testing.T) {
		name, err := NewName("  Жим лёжа  ")

		require.NoError(t, err)
		assert.Equal(t, "Жим лёжа", name.String())
	})

	t.Run("невалидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"пустая строка", ""},
			{"только пробелы", "   "},
			{"символ @", "Жим@лёжа"},
			{"точка с запятой", "Жим;лёжа"},
			{"перевод строки", "Жим\nлёжа"},
			{"табуляция", "Жим\tлёжа"},
			{"восклицательный знак", "Жим!"},
			{"решётка", "Жим #1"},
			{"двоеточие", "Жим: базовый"},
			{"эмодзи", "Жим 😀"},
			{"знак процента", "100% жим"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewName(tt.input)
				assert.ErrorIs(t, err, ErrInvalidName)
			})
		}
	})
}

func TestNewDescription(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"1 символ", "a"},
			{"обычное описание", "Базовое упражнение для груди."},
			{"с цифрами и знаками", "Выполняется 3-4 подхода, 8-12 повторений."},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				description, err := NewDescription(tt.input)

				require.NoError(t, err)
				assert.Equal(t, tt.input, description.String())
			})
		}
	})

	t.Run("границы длины", func(t *testing.T) {
		t.Run("пустая строка — меньше минимума", func(t *testing.T) {
			_, err := NewDescription("")
			assert.ErrorIs(t, err, ErrInvalidDescription)
		})

		t.Run("только пробелы — после trim пусто", func(t *testing.T) {
			_, err := NewDescription("   ")
			assert.ErrorIs(t, err, ErrInvalidDescription)
		})

		t.Run("1 символ — ровно минимум", func(t *testing.T) {
			description, err := NewDescription("a")
			require.NoError(t, err)
			assert.Equal(t, "a", description.String())
		})

		t.Run("1000 рун — ровно максимум", func(t *testing.T) {
			input := strings.Repeat("a", 1000)
			description, err := NewDescription(input)
			require.NoError(t, err)
			assert.Equal(t, input, description.String())
		})

		t.Run("1001 руна — больше максимума", func(t *testing.T) {
			_, err := NewDescription(strings.Repeat("a", 1001))
			assert.ErrorIs(t, err, ErrInvalidDescription)
		})
	})

	t.Run("trim пробелов", func(t *testing.T) {
		description, err := NewDescription("  Базовое упражнение  ")

		require.NoError(t, err)
		assert.Equal(t, "Базовое упражнение", description.String())
	})

	t.Run("невалидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"символ @", "Описание@текст"},
			{"точка с запятой", "Описание; текст"},
			{"перевод строки", "Первая строка\nВторая"},
			{"эмодзи", "Описание 😀"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewDescription(tt.input)
				assert.ErrorIs(t, err, ErrInvalidDescription)
			})
		}
	})
}

func TestNewDifficulty(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		for _, input := range []int{1, 5, 10} {
			t.Run("", func(t *testing.T) {
				difficulty, err := NewDifficulty(input)

				require.NoError(t, err)
				assert.Equal(t, input, difficulty.Int())
			})
		}
	})

	t.Run("границы", func(t *testing.T) {
		t.Run("1 — минимум", func(t *testing.T) {
			difficulty, err := NewDifficulty(1)
			require.NoError(t, err)
			assert.Equal(t, 1, difficulty.Int())
		})

		t.Run("10 — максимум", func(t *testing.T) {
			difficulty, err := NewDifficulty(10)
			require.NoError(t, err)
			assert.Equal(t, 10, difficulty.Int())
		})

		t.Run("0 — ниже минимума", func(t *testing.T) {
			_, err := NewDifficulty(0)
			assert.ErrorIs(t, err, ErrInvalidDifficulty)
		})

		t.Run("11 — выше максимума", func(t *testing.T) {
			_, err := NewDifficulty(11)
			assert.ErrorIs(t, err, ErrInvalidDifficulty)
		})

		t.Run("отрицательное", func(t *testing.T) {
			_, err := NewDifficulty(-1)
			assert.ErrorIs(t, err, ErrInvalidDifficulty)
		})

		t.Run("большое число", func(t *testing.T) {
			_, err := NewDifficulty(100)
			assert.ErrorIs(t, err, ErrInvalidDifficulty)
		})
	})
}

func TestNewExerciseType(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		t.Run("weight", func(t *testing.T) {
			exerciseType, err := NewExerciseType("weight")
			require.NoError(t, err)
			assert.Equal(t, ExerciseTypeWeight, exerciseType)
			assert.Equal(t, "weight", exerciseType.String())
		})

		t.Run("duration", func(t *testing.T) {
			exerciseType, err := NewExerciseType("duration")
			require.NoError(t, err)
			assert.Equal(t, ExerciseTypeDuration, exerciseType)
			assert.Equal(t, "duration", exerciseType.String())
		})
	})

	t.Run("невалидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"пустая строка", ""},
			{"weight с заглавной", "Weight"},
			{"WEIGHT полностью капсом", "WEIGHT"},
			{"duration с заглавной", "Duration"},
			{"неизвестный тип cardio", "cardio"},
			{"неизвестный тип strength", "strength"},
			{"пробел вокруг", " weight"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewExerciseType(tt.input)
				assert.ErrorIs(t, err, ErrInvalidExerciseType)
			})
		}
	})
}
