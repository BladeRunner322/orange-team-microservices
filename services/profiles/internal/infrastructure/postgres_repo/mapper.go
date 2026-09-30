// Package postgres_repo — преобразование между domain.Profile и ProfileModel.
package postgres_repo

import "github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"

// ModelToDomain преобразует модель БД в доменную сущность.
func ModelToDomain(m ProfileModel) (domain.Profile, error) {
	sex, err := nullableToDomain(m.Sex, domain.NewSex)
	if err != nil {
		return domain.Profile{}, err
	}

	weightGrams, err := nullableToDomain(m.WeightGrams, domain.NewWeightFromGrams)
	if err != nil {
		return domain.Profile{}, err
	}

	birthDate, err := nullableToDomain(m.BirthDate, domain.NewBirthDate)
	if err != nil {
		return domain.Profile{}, err
	}

	heightCM, err := nullableToDomain(m.HeightCM, domain.NewHeight)
	if err != nil {
		return domain.Profile{}, err
	}

	return domain.RestoreProfile(
		m.UserID,
		sex,
		weightGrams,
		birthDate,
		heightCM,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

// DomainToModel преобразует доменную сущность в модель БД.
func DomainToModel(profile domain.Profile) ProfileModel {
	return ProfileModel{
		UserID:      profile.UserID(),
		Sex:         nullableToModel(profile.Sex(), domain.Sex.String),
		WeightGrams: nullableToModel(profile.Weight(), domain.Weight.Grams),
		BirthDate:   nullableToModel(profile.BirthDate(), domain.BirthDate.Time),
		HeightCM:    nullableToModel(profile.Height(), domain.Height.Centimeters),
		CreatedAt:   profile.CreatedAt(),
		UpdatedAt:   profile.UpdatedAt(),
	}
}

// nullableToDomain конвертирует nullable-поле из модели БД в nullable-поле домена.
//
// Если src == nil — в БД NULL, значит в домене тоже nil.
// Иначе вызывает convert (VO-конструктор) и возвращает указатель на результат.
func nullableToDomain[T any, R any](src *T, convert func(T) (R, error)) (*R, error) {
	if src == nil {
		return nil, nil
	}
	r, err := convert(*src)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// nullableToModel конвертирует nullable-поле из домена в модель БД.
// Если src == nil — возвращает nil. Иначе вызывает convert и возвращает указатель.
func nullableToModel[T any, R any](src *T, convert func(T) R) *R {
	if src == nil {
		return nil
	}
	r := convert(*src)
	return &r
}
