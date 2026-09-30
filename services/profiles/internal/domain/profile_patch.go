package domain

import "github.com/BladeRunner322/orange-team-microservices/pkg/nullable"

// ProfilePatch — набор изменений для профиля.
// Каждое поле может быть: не задано (Set: false),
// установлено в NULL (Set: true, Value: nil),
// или установлено в значение (Set: true, Value: &x).
type ProfilePatch struct {
	Sex       nullable.Nullable[Sex]
	Weight    nullable.Nullable[Weight]
	BirthDate nullable.Nullable[BirthDate]
	Height    nullable.Nullable[Height]
}

// NewProfilePatch создаёт патч из набора nullable-полей.
func NewProfilePatch(
	sex nullable.Nullable[Sex],
	weight nullable.Nullable[Weight],
	birthDate nullable.Nullable[BirthDate],
	height nullable.Nullable[Height],
) ProfilePatch {
	return ProfilePatch{
		Sex:       sex,
		Weight:    weight,
		BirthDate: birthDate,
		Height:    height,
	}
}
