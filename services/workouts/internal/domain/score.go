package domain

const (
	MaximumWorkoutScore  = 10_000
	ScoreSaturationScale = 100_000
	MaximumScaledLoad    = 100_000_000
)

// CalculateScore вычисляет итоговый score тренировки по формуле из ADR-025 п.2.
// Целочисленная реализация, чтобы избежать floating-point погрешностей.
func CalculateScore(effectiveLoad, coefficient int) int {
	// TODO: реализовать saturatedScore
	// score = min(10000, round(10000 * K * load / (1 + K * load))), где K = coef/100_000
	return 0
}

// AddLoadWithLimit добавляет load×difficulty к текущему effectiveLoad с защитой от overflow.
func AddLoadWithLimit(current, load, difficulty, limit int) int {
	// TODO: current + load×difficulty, cap по limit
	return 0
}

// SaturatedScore — внутренняя часть формулы.
func SaturatedScore(effectiveLoad, coefficient int) int {
	// TODO
	return 0
}

// DivideRoundUp — округление вверх.
func DivideRoundUp(value, divisor int) int {
	// TODO
	return 0
}
