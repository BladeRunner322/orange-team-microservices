package domain

import (
	"math"
	"time"
)

// Sex — value object пола.
type Sex string

const (
	SexMale   Sex = "male"
	SexFemale Sex = "female"
)

func NewSex(raw string) (Sex, error) {
	switch Sex(raw) {
	case SexMale, SexFemale:
		return Sex(raw), nil
	default:
		return "", ErrInvalidSex
	}
}

func (s Sex) String() string { return string(s) }

// Weight — value object веса.
type Weight int

const (
	MinWeightGrams = 40_000
	MaxWeightGrams = 150_000
)

func NewWeightFromGrams(g int) (Weight, error) {
	if g < MinWeightGrams || g > MaxWeightGrams {
		return 0, ErrInvalidWeight
	}
	if g%100 != 0 {
		return 0, ErrInvalidWeight
	}
	return Weight(g), nil
}

func NewWeightFromKilograms(kg float64) (Weight, error) {
	return NewWeightFromGrams(int(math.Round(kg * 1000)))
}

func (w Weight) Grams() int {
	return int(w)
}

func (w Weight) Kilograms() float64 {
	return float64(w) / 1000
}

// BirthDate — value object даты рождения.
type BirthDate time.Time

const (
	MinBirthYear = 1900
	MinAge       = 12
)

func NewBirthDate(date time.Time) (BirthDate, error) {
	minDate := time.Date(MinBirthYear, 1, 1, 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()

	if date.Before(minDate) || date.After(today) {
		return BirthDate{}, ErrInvalidBirthDate
	}

	years := today.Year() - date.Year()
	if today.Month() < date.Month() ||
		(today.Month() == date.Month() && today.Day() < date.Day()) {
		years--
	}

	if years < MinAge {
		return BirthDate{}, ErrInvalidBirthDate
	}

	return BirthDate(date), nil
}

func (b BirthDate) Time() time.Time {
	return time.Time(b)
}

func (b BirthDate) String() string {
	return b.Time().UTC().Format("2006-01-02")
}

// Height — value object роста.
type Height int

const (
	MinHeightCM = 140
	MaxHeightCM = 210
)

func NewHeight(cm int) (Height, error) {
	if cm < MinHeightCM || cm > MaxHeightCM {
		return 0, ErrInvalidHeight
	}

	return Height(cm), nil
}

func (h Height) Centimeters() int {
	return int(h)
}
