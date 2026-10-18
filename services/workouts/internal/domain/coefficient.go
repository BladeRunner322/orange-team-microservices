package domain

const (
	MinScoreCoefficient = 1
	MaxScoreCoefficient = 10
)

// UserProfileData — минимальные данные профиля, нужные для расчёта коэффициента.
type UserProfileData struct {
	Sex         *string
	WeightGrams *int
	HeightCM    *int
}

// CalculatePersonalScoreCoefficient вычисляет коэффициент по формуле из ADR-025 п.2.
// BMI-based, +0.5 для male, cap [1, 10].
func CalculatePersonalScoreCoefficient(p UserProfileData) int {
	// TODO: если weight/height nil → 1
	// TODO: BMI = weight / (height_m²)
	// TODO: coeff = clamp(1 + (BMI-18) × 9/17, 1, 10)
	// TODO: если male → +0.5, cap 10
	return 1
}
