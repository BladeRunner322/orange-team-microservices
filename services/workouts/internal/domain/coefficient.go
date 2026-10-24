package domain

const (
	MinScoreCoefficient = 1
	MaxScoreCoefficient = 10

	// baseBMI — опорная точка BMI, при которой коэффициент = 1.
	baseBMI = 18.0

	// coefficientSlopeNumerator / coefficientSlopeDivisor — наклон
	// линейной части: 1 + (BMI - baseBMI) × numerator / divisor.
	coefficientSlopeNumerator = 9.0
	coefficientSlopeDivisor   = 17.0

	// maleCoefficientBonus — надбавка к коэффициенту для male.
	maleCoefficientBonus = 0.5

	// Пересчёт единиц.
	gramsPerKilogram    = 1000.0
	centimetersPerMeter = 100.0

	// Сдвиг для round-half-up.
	roundHalfUpOffset = 0.5

	// Значение sex из proto Profiles (sex ∈ {"male", "female"}).
	maleSexValue = "male"
)

// UserProfileData — вход для расчёта коэффициента.
// Примитивы, не VO Profiles — чтобы не связывать домены (ADR-004).
type UserProfileData struct {
	Sex         *string
	WeightGrams *int
	HeightCM    *int
}

// CalculatePersonalScoreCoefficient считает коэффициент по формуле ADR-025 п.2:
//
//	BMI = weight_kg / height_m²
//	coeff = clamp(1 + (BMI - 18) × 9/17, 1, 10)
//	male → +0.5, cap 10
//
// Незаполненный профиль → 1. Возвращает снапшот на момент создания
// тренировки (ADR-001): изменения профиля не влияют на старые тренировки.
func CalculatePersonalScoreCoefficient(p UserProfileData) int {
	if p.WeightGrams == nil || p.HeightCM == nil {
		return MinScoreCoefficient
	}

	weightKg := float64(*p.WeightGrams) / gramsPerKilogram
	heightM := float64(*p.HeightCM) / centimetersPerMeter

	bmi := weightKg / (heightM * heightM)

	coeff := 1.0 + (bmi-baseBMI)*coefficientSlopeNumerator/coefficientSlopeDivisor
	coeff = clampCoefficient(coeff)

	if p.Sex != nil && *p.Sex == maleSexValue {
		coeff += maleCoefficientBonus
		coeff = clampCoefficient(coeff)
	}

	return int(coeff + roundHalfUpOffset)
}

func clampCoefficient(c float64) float64 {
	if c < float64(MinScoreCoefficient) {
		return float64(MinScoreCoefficient)
	}
	if c > float64(MaxScoreCoefficient) {
		return float64(MaxScoreCoefficient)
	}
	return c
}
