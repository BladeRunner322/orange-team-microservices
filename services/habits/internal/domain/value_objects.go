package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// habitTextPattern — допустимые символы для name и description.
// Буквы (латиница, кириллица), цифры, пробелы и пунктуация ., - () / + = " '
var habitTextPattern = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9 .,\-()/+="']+$`)

// Name — value object названия привычки.
// 1–100 символов после trim.
type Name string

const (
	MinNameLength = 1
	MaxNameLength = 100
)

// NewName создаёт Name или возвращает ErrInvalidName.
func NewName(raw string) (Name, error) {
	trimmed := strings.TrimSpace(raw)

	length := utf8.RuneCountInString(trimmed)
	if length < MinNameLength || length > MaxNameLength {
		return "", ErrInvalidName
	}

	if !habitTextPattern.MatchString(trimmed) {
		return "", ErrInvalidName
	}

	return Name(trimmed), nil
}

func (n Name) String() string { return string(n) }

// Description — value object описания привычки.
// 0–500 символов после trim. Пустое допустимо.
type Description string

const (
	MaxDescriptionLength = 500
)

// NewDescription создаёт Description или возвращает ErrInvalidDescription.
func NewDescription(raw string) (Description, error) {
	trimmed := strings.TrimSpace(raw)

	if utf8.RuneCountInString(trimmed) > MaxDescriptionLength {
		return "", ErrInvalidDescription
	}

	// Пустое — валидно, regex не проверяем.
	if trimmed == "" {
		return Description(""), nil
	}

	if !habitTextPattern.MatchString(trimmed) {
		return "", ErrInvalidDescription
	}

	return Description(trimmed), nil
}

func (d Description) String() string { return string(d) }

// Date — value object календарного дня (без времени).
// Хранится в UTC 00:00.
//
// Проверка диапазона («не в будущем», «не старше 7 дней») — в usecase.
// VO отвечает только за формат: «это календарный день, а не момент времени».
type Date time.Time

// NewDate валидирует формат и нормализует в UTC 00:00.
// Zero time отклоняется — это признак «дата не задана».
func NewDate(raw time.Time) (Date, error) {
	if raw.IsZero() {
		return Date{}, ErrInvalidDate
	}

	u := raw.UTC()
	normalized := time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)

	return Date(normalized), nil
}

// Time возвращает внутреннее time.Time (UTC 00:00).
// Используется при передаче в SQL.
func (d Date) Time() time.Time {
	return time.Time(d)
}

// String возвращает "YYYY-MM-DD".
// Используется при логировании и маппинге в proto.
func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}
