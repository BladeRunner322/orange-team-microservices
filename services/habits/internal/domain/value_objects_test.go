package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewName(t *testing.T) {
	t.Run("валидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"1 символ — минимум", "А"},
			{"кириллица с пробелом", "Пить воду"},
			{"латиница с дефисом", "Push-ups"},
			{"скобки", "Читать (нонфикшн)"},
			{"знаки = и +", "A=B+C"},
			{"апостроф", "It's a habit"},
			{"точка и запятая", "Читать 20 минут."},
			{"слэш", "Бег/ходьба"},
			{"цифры", "Пить 2 литра"},
			{"буква ё", "Ёжик заводить"},
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
		t.Run("пустая строка — ошибка", func(t *testing.T) {
			_, err := NewName("")
			assert.ErrorIs(t, err, ErrInvalidName)
		})

		t.Run("1 символ — ровно минимум", func(t *testing.T) {
			name, err := NewName("А")
			require.NoError(t, err)
			assert.Equal(t, "А", name.String())
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
		// "абв" — 3 руны, 6 байт.
		name, err := NewName("абв")

		require.NoError(t, err)
		assert.Equal(t, "абв", name.String())
	})

	t.Run("trim пробелов", func(t *testing.T) {
		name, err := NewName("  Пить воду  ")

		require.NoError(t, err)
		assert.Equal(t, "Пить воду", name.String())
	})

	t.Run("невалидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"только пробелы", "   "},
			{"символ @", "Пить@воду"},
			{"точка с запятой", "Пить;воду"},
			{"перевод строки", "Пить\nводу"},
			{"табуляция", "Пить\tводу"},
			{"восклицательный знак", "Пить воду!"},
			{"решётка", "Пить #1"},
			{"двоеточие", "Пить: вода"},
			{"эмодзи", "Пить 😀"},
			{"знак процента", "Пить 100%"},
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
			{"пустая строка — валидно", ""},
			{"только пробелы — trim в пустое", "   "},
			{"обычное описание", "2 литра в день"},
			{"с цифрами и знаками", "Читать 20-30 минут ежедневно."},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				description, err := NewDescription(tt.input)

				require.NoError(t, err)
				// После trim пробелы исчезают
				expected := strings.TrimSpace(tt.input)
				assert.Equal(t, expected, description.String())
			})
		}
	})

	t.Run("границы длины", func(t *testing.T) {
		t.Run("500 рун — ровно максимум", func(t *testing.T) {
			input := strings.Repeat("a", 500)
			description, err := NewDescription(input)
			require.NoError(t, err)
			assert.Equal(t, input, description.String())
		})

		t.Run("501 руна — больше максимума", func(t *testing.T) {
			_, err := NewDescription(strings.Repeat("a", 501))
			assert.ErrorIs(t, err, ErrInvalidDescription)
		})
	})

	t.Run("trim пробелов", func(t *testing.T) {
		description, err := NewDescription("  2 литра в день  ")

		require.NoError(t, err)
		assert.Equal(t, "2 литра в день", description.String())
	})

	t.Run("невалидные значения", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"символ @", "Описание@текст"},
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

func TestNewDate(t *testing.T) {
	t.Run("валидная дата нормализуется в UTC 00:00", func(t *testing.T) {
		// Дата с временем и другим часовым поясом
		input := time.Date(2026, 10, 15, 15, 42, 33, 0, time.FixedZone("MSK", 3*60*60))

		d, err := NewDate(input)
		require.NoError(t, err)

		expected := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, d.Time())
	})

	t.Run("zero time — ошибка", func(t *testing.T) {
		_, err := NewDate(time.Time{})
		assert.ErrorIs(t, err, ErrInvalidDate)
	})

	t.Run("время внутри дня обнуляется", func(t *testing.T) {
		input := time.Date(2026, 10, 15, 23, 59, 59, 0, time.UTC)

		d, err := NewDate(input)
		require.NoError(t, err)

		expected := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, d.Time())
	})

	t.Run("Time возвращает time.Time", func(t *testing.T) {
		input := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		d, err := NewDate(input)
		require.NoError(t, err)

		assert.Equal(t, input, d.Time())
	})

	t.Run("String возвращает YYYY-MM-DD", func(t *testing.T) {
		input := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		d, err := NewDate(input)
		require.NoError(t, err)

		assert.Equal(t, "2026-10-15", d.String())
	})
}
