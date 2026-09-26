package domain

import "github.com/google/uuid"

type Profile struct {
	userID    uuid.UUID
	sex       *Sex
	weight    *Weight
	birthDate *BirthDate
	height    *Height
}

// NewProfile создаёт профиль с указанными полями.
func NewProfile(userID uuid.UUID, sex *Sex, weight *Weight, birthDate *BirthDate, height *Height) Profile {
	return Profile{
		userID:    userID,
		sex:       sex,
		weight:    weight,
		birthDate: birthDate,
		height:    height,
	}
}

// NewEmptyProfile создаёт пустой профиль.
func NewEmptyProfile(userID uuid.UUID) Profile {
	return NewProfile(userID, nil, nil, nil, nil)
}

// Completed возвращает true, если заполнены все 4 поля профиля.
func (p Profile) Completed() bool {
	return p.sex != nil && p.weight != nil && p.birthDate != nil && p.height != nil
}

// Геттеры (публичные)
func (p Profile) UserID() uuid.UUID     { return p.userID }
func (p Profile) Sex() *Sex             { return p.sex }
func (p Profile) Weight() *Weight       { return p.weight }
func (p Profile) BirthDate() *BirthDate { return p.birthDate }
func (p Profile) Height() *Height       { return p.height }
