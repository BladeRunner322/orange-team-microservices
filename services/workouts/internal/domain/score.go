package domain

const (
	// MaximumWorkoutScore — верхняя граница score, к которой формула
	// стремится асимптотически (никогда не достигает).
	MaximumWorkoutScore = 10_000

	// ScoreSaturationScale — масштаб в формуле. Чем больше, тем медленнее
	// кривая насыщается.
	ScoreSaturationScale = 100_000

	// MaximumScaledLoad — предел scaledLoad = coefficient × effectiveLoad.
	// При его достижении score выходит на насыщение, дальнейший набор
	// очков невозможен.
	MaximumScaledLoad = 100_000_000
)

// CalculateScore считает score тренировки по формуле ADR-025 п.2.
//
// effectiveLoad — Σ(exercise_load × difficulty) по completed упражнениям.
// coefficient — personal_score_coefficient (снапшот, ADR-001).
//
// Целочисленная арифметика без float — детерминированный результат.
func CalculateScore(effectiveLoad, coefficient int) int {
	return saturatedScore(effectiveLoad, coefficient)
}

// saturatedScore — реализация формулы:
//
//	K = coefficient / ScoreSaturationScale
//	score = round(MaximumWorkoutScore × K × effectiveLoad / (1 + K × effectiveLoad))
//
// После умножения числителя и знаменателя на ScoreSaturationScale:
//
//	score = round(MaximumWorkoutScore × scaledLoad / (ScoreSaturationScale + scaledLoad))
//	где scaledLoad = coefficient × effectiveLoad
//
// При scaledLoad → ∞ score → MaximumWorkoutScore (асимптота).
// При scaledLoad = MaximumScaledLoad score ≈ MaximumWorkoutScore, но строго меньше.
//
// Внутри — int64: MaximumWorkoutScore × MaximumScaledLoad = 10^12,
// не помещается в int32 (защита для 32-битных платформ).
func saturatedScore(effectiveLoad, coefficient int) int {
	scaledLoad := int64(effectiveLoad) * int64(coefficient)
	denominator := int64(ScoreSaturationScale) + scaledLoad
	score := (int64(MaximumWorkoutScore)*scaledLoad + denominator/2) / denominator

	return int(score)
}

// AddLoadWithLimit прибавляет load × difficulty к current, не превышая limit.
// Защищает от переполнения при накоплении effectiveLoad в цикле по упражнениям.
//
// Возвращает min(current + load × difficulty, limit).
// Если current уже равен limit — возвращает limit без умножения.
func AddLoadWithLimit(current, load, difficulty, limit int) int {
	if current >= limit || load > (limit-current)/difficulty {
		return limit
	}

	return current + load*difficulty
}

// DivideRoundUp — целочисленное деление с округлением вверх.
// divisor должен быть > 0.
//
//	DivideRoundUp(10, 3) = 4
//	DivideRoundUp(9, 3)  = 3
func DivideRoundUp(value, divisor int) int {
	return (value + divisor - 1) / divisor
}