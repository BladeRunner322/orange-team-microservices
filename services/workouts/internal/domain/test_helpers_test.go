package domain

import (
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
)

// Хелперы для построения nullable.Nullable[Intensity] в тестах.
// Три состояния: не трогать (Set=false), установить (Set=true, Value≠nil), сбросить (Set=true, Value=nil).

func nullableIntensityNone() nullable.Nullable[Intensity] {
	return nullable.Nullable[Intensity]{}
}

func nullableIntensitySet(v *Intensity) nullable.Nullable[Intensity] {
	return nullable.Nullable[Intensity]{Set: true, Value: v}
}

func nullableIntensityClear() nullable.Nullable[Intensity] {
	return nullable.Nullable[Intensity]{Set: true, Value: nil}
}
