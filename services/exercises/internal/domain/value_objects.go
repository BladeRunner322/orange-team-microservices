package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// exerciseTextPattern — допустимые символы для name и description.
// Буквы (латиница, кириллица), цифры, пробелы и пунктуация ., - () / + = " '
var exerciseTextPattern = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9 .,\-()/+="']+$`)

// Name — value object названия.
type Name string

const (
	MinNameLength = 3
	MaxNameLength = 100
)

func NewName(raw string) (Name, error) {
	trimmed := strings.TrimSpace(raw)

	length := utf8.RuneCountInString(trimmed)
	if length < MinNameLength || length > MaxNameLength {
		return "", ErrInvalidName
	}

	if !exerciseTextPattern.MatchString(trimmed) {
		return "", ErrInvalidName
	}

	return Name(trimmed), nil
}

func (n Name) String() string { return string(n) }

// Description — value object описания.
type Description string

const (
	MinDescriptionLength = 1
	MaxDescriptionLength = 1000
)

func NewDescription(raw string) (Description, error) {
	trimmed := strings.TrimSpace(raw)

	length := utf8.RuneCountInString(trimmed)
	if length < MinDescriptionLength || length > MaxDescriptionLength {
		return "", ErrInvalidDescription
	}

	if !exerciseTextPattern.MatchString(trimmed) {
		return "", ErrInvalidDescription
	}

	return Description(trimmed), nil
}

func (d Description) String() string { return string(d) }

// Difficulty — value object сложности.
type Difficulty int

const (
	MinDifficulty = 1
	MaxDifficulty = 10
)

func NewDifficulty(raw int) (Difficulty, error) {
	if raw < MinDifficulty || raw > MaxDifficulty {
		return 0, ErrInvalidDifficulty
	}

	return Difficulty(raw), nil
}

func (d Difficulty) Int() int { return int(d) }

// ExerciseType — value object типа.
type ExerciseType string

const (
	ExerciseTypeWeight   ExerciseType = "weight"
	ExerciseTypeDuration ExerciseType = "duration"
)

func NewExerciseType(raw string) (ExerciseType, error) {
	switch ExerciseType(raw) {
	case ExerciseTypeWeight, ExerciseTypeDuration:
		return ExerciseType(raw), nil
	default:
		return "", ErrInvalidExerciseType
	}
}

func (t ExerciseType) String() string { return string(t) }
