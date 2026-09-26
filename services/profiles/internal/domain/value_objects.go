package domain

import (
	"math"
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

// // BirthDate — value object даты рождения.
// type BirthDate time.Time

// const (
// 	MinBirthYear = 1900
// )

// func NewBirthDate(date time.Time) (BirthDate, error) {
// 	// TODO: проверка "не раньше MinBirthYear"
// 	// TODO: проверка "не позже time.Now()"
// 	// TODO: return BirthDate(date), nil
// }

// func (b BirthDate) Time() time.Time {
// 	// TODO
// }

// func (b BirthDate) String() string {
// 	// TODO: формат YYYY-MM-DD
// }

// // Height — value object роста.
// type Height int

// // Границы
// const (
// 	MinHeightCM = 140
// 	MaxHeightCM = 210
// )

// func NewHeight(cm int) (Height, error) {
// 	// TODO: проверка диапазона
// 	// TODO: return Height(cm), nil

// }

// func (h Height) Centimeters() int {
// 	// TODO
// }
