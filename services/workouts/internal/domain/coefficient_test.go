package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func TestCalculatePersonalScoreCoefficient(t *testing.T) {
	t.Run("nil weight или height → 1", func(t *testing.T) {
		t.Run("weight nil", func(t *testing.T) {
			assert.Equal(t, 1, CalculatePersonalScoreCoefficient(UserProfileData{
				WeightGrams: nil,
				HeightCM:    intPtr(180),
			}))
		})

		t.Run("height nil", func(t *testing.T) {
			assert.Equal(t, 1, CalculatePersonalScoreCoefficient(UserProfileData{
				WeightGrams: intPtr(80000),
				HeightCM:    nil,
			}))
		})

		t.Run("оба nil", func(t *testing.T) {
			assert.Equal(t, 1, CalculatePersonalScoreCoefficient(UserProfileData{}))
		})
	})

	t.Run("расчёт по BMI (female)", func(t *testing.T) {
		tests := []struct {
			name        string
			weightGrams int
			heightCM    int
			want        int
		}{
			// BMI = 80 / 1.8² = 24.69 → coeff = 1 + 6.69 × 9/17 = 4.54 → 5
			{"80kg/180cm", 80000, 180, 5},
			// BMI = 70 / 1.75² = 22.86 → coeff = 1 + 4.86 × 9/17 = 3.57 → 4
			{"70kg/175cm", 70000, 175, 4},
			// BMI = 90 / 1.8² = 27.78 → coeff = 1 + 9.78 × 9/17 = 6.18 → 6
			{"90kg/180cm", 90000, 180, 6},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := CalculatePersonalScoreCoefficient(UserProfileData{
					Sex:         strPtr("female"),
					WeightGrams: intPtr(tt.weightGrams),
					HeightCM:    intPtr(tt.heightCM),
				})
				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("male даёт +0.5", func(t *testing.T) {
		// 85kg/180cm: BMI = 26.23, coeff(female) = 5.36 → 5, coeff(male) = 5.86 → 6
		profile := func(sex *string) UserProfileData {
			return UserProfileData{
				Sex:         sex,
				WeightGrams: intPtr(85000),
				HeightCM:    intPtr(180),
			}
		}

		assert.Equal(t, 5, CalculatePersonalScoreCoefficient(profile(strPtr("female"))))
		assert.Equal(t, 6, CalculatePersonalScoreCoefficient(profile(strPtr("male"))))
	})

	t.Run("sex nil — без бонуса", func(t *testing.T) {
		// те же данные, что female, но sex = nil → тот же результат
		got := CalculatePersonalScoreCoefficient(UserProfileData{
			Sex:         nil,
			WeightGrams: intPtr(85000),
			HeightCM:    intPtr(180),
		})
		assert.Equal(t, 5, got)
	})

	t.Run("неизвестное значение sex — без бонуса", func(t *testing.T) {
		for _, sex := range []string{"other", "MALE", "Male", ""} {
			t.Run(sex, func(t *testing.T) {
				got := CalculatePersonalScoreCoefficient(UserProfileData{
					Sex:         strPtr(sex),
					WeightGrams: intPtr(85000),
					HeightCM:    intPtr(180),
				})
				assert.Equal(t, 5, got, "sex=%q не должен давать +0.5", sex)
			})
		}
	})

	t.Run("низкий BMI — clamp на 1", func(t *testing.T) {
		// 45kg/180cm: BMI = 13.9, coeff = 1 + (-4.1) × 9/17 = -1.17 → clamp 1
		got := CalculatePersonalScoreCoefficient(UserProfileData{
			Sex:         strPtr("female"),
			WeightGrams: intPtr(45000),
			HeightCM:    intPtr(180),
		})
		assert.Equal(t, 1, got)
	})

	t.Run("очень высокий BMI — clamp на 10", func(t *testing.T) {
		// 140kg/180cm: BMI = 43.2, coeff = 1 + 25.2 × 9/17 = 14.34 → clamp 10
		got := CalculatePersonalScoreCoefficient(UserProfileData{
			Sex:         strPtr("female"),
			WeightGrams: intPtr(140000),
			HeightCM:    intPtr(180),
		})
		assert.Equal(t, 10, got)
	})

	t.Run("male при высоком BMI — cap остаётся 10", func(t *testing.T) {
		// coeff(female) = 14.34 → clamp 10, male: 10 + 0.5 = 10.5 → clamp 10
		got := CalculatePersonalScoreCoefficient(UserProfileData{
			Sex:         strPtr("male"),
			WeightGrams: intPtr(140000),
			HeightCM:    intPtr(180),
		})
		assert.Equal(t, 10, got)
	})

	t.Run("результат всегда в [1, 10]", func(t *testing.T) {
		// Перебор разных комбинаций — invariant
		for _, weight := range []int{40000, 60000, 80000, 100000, 140000} {
			for _, height := range []int{150, 170, 190, 210} {
				for _, sex := range []string{"male", "female"} {
					got := CalculatePersonalScoreCoefficient(UserProfileData{
						Sex:         strPtr(sex),
						WeightGrams: intPtr(weight),
						HeightCM:    intPtr(height),
					})
					assert.GreaterOrEqual(t, got, MinScoreCoefficient)
					assert.LessOrEqual(t, got, MaxScoreCoefficient)
				}
			}
		}
	})
}