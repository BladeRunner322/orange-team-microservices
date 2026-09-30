package profilesgrpc

import (
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// sexFromProto конвертирует NullableString в Nullable[domain.Sex].
func sexFromProto(wrapper *profiles.NullableString) (nullable.Nullable[domain.Sex], error) {
	if wrapper == nil {
		return nullable.Nullable[domain.Sex]{}, nil
	}

	if wrapper.Value == nil {
		return nullable.Nullable[domain.Sex]{Set: true}, nil
	}

	sex, err := domain.NewSex(*wrapper.Value)
	if err != nil {
		return nullable.Nullable[domain.Sex]{}, fmt.Errorf("validate sex: %w", err)
	}

	return nullable.Nullable[domain.Sex]{Set: true, Value: &sex}, nil
}

// weightKgFromProto конвертирует NullableDouble (кг) в Nullable[domain.Weight] (граммы).
func weightKgFromProto(wrapper *profiles.NullableDouble) (nullable.Nullable[domain.Weight], error) {
	if wrapper == nil {
		return nullable.Nullable[domain.Weight]{}, nil
	}

	if wrapper.Value == nil {
		return nullable.Nullable[domain.Weight]{Set: true}, nil
	}

	weight, err := domain.NewWeightFromKilograms(*wrapper.Value)
	if err != nil {
		return nullable.Nullable[domain.Weight]{}, fmt.Errorf("validate weight: %w", err)
	}

	return nullable.Nullable[domain.Weight]{Set: true, Value: &weight}, nil
}

// birthDateFromProto конвертирует NullableString (YYYY-MM-DD) в Nullable[domain.BirthDate].
func birthDateFromProto(wrapper *profiles.NullableString) (nullable.Nullable[domain.BirthDate], error) {
	if wrapper == nil {
		return nullable.Nullable[domain.BirthDate]{}, nil
	}

	if wrapper.Value == nil {
		return nullable.Nullable[domain.BirthDate]{Set: true}, nil
	}

	birthDateParsed, err := time.Parse("2006-01-02", *wrapper.Value)
	if err != nil {
		return nullable.Nullable[domain.BirthDate]{}, fmt.Errorf("parse birth_date: %w", err)
	}

	birthDate, err := domain.NewBirthDate(birthDateParsed)
	if err != nil {
		return nullable.Nullable[domain.BirthDate]{}, fmt.Errorf("validate birth_date: %w", err)
	}

	return nullable.Nullable[domain.BirthDate]{Set: true, Value: &birthDate}, nil
}

// heightCmFromProto конвертирует NullableInt32 в Nullable[domain.Height].
func heightCmFromProto(wrapper *profiles.NullableInt32) (nullable.Nullable[domain.Height], error) {
	if wrapper == nil {
		return nullable.Nullable[domain.Height]{}, nil
	}

	if wrapper.Value == nil {
		return nullable.Nullable[domain.Height]{Set: true}, nil
	}

	height, err := domain.NewHeight(int(*wrapper.Value))
	if err != nil {
		return nullable.Nullable[domain.Height]{}, fmt.Errorf("validate height: %w", err)
	}

	return nullable.Nullable[domain.Height]{Set: true, Value: &height}, nil
}

// ===== Преобразования из protobuf в параметры use case =====

// ToDomainProfilePatch преобразует pb.PatchMyProfileRequest в параметры для use case.
func ToDomainProfilePatch(req *profiles.PatchMyProfileRequest) (domain.ProfilePatch, error) {
	sex, err := sexFromProto(req.Sex)
	if err != nil {
		return domain.ProfilePatch{}, err
	}

	weight, err := weightKgFromProto(req.WeightKg)
	if err != nil {
		return domain.ProfilePatch{}, err
	}

	birthDate, err := birthDateFromProto(req.BirthDate)
	if err != nil {
		return domain.ProfilePatch{}, err
	}

	height, err := heightCmFromProto(req.HeightCm)
	if err != nil {
		return domain.ProfilePatch{}, err
	}

	return domain.NewProfilePatch(sex, weight, birthDate, height), nil
}

// ToDomainUserID преобразует pb.GetProfileRequest в параметры для use case.
func ToDomainUserID(req *profiles.GetProfileRequest) (uuid.UUID, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse user_id: %w", err)
	}

	return userID, nil
}

// ===== Преобразования из доменных объектов в protobuf =====

// ToProtoProfile преобразует профиль в pb.UserProfile.
func ToProtoProfile(profile domain.Profile) *profiles.UserProfile {
	sex := ""
	if s := profile.Sex(); s != nil {
		sex = s.String()
	}

	weightKg := 0.0
	if w := profile.Weight(); w != nil {
		weightKg = w.Kilograms()
	}

	birthDate := ""
	if bd := profile.BirthDate(); bd != nil {
		birthDate = bd.String()
	}

	heightCm := int32(0)
	if h := profile.Height(); h != nil {
		heightCm = int32(h.Centimeters())
	}

	var updatedAt *timestamppb.Timestamp
	if t := profile.UpdatedAt(); t != nil {
		updatedAt = timestamppb.New(*t)
	}

	return &profiles.UserProfile{
		UserId:           profile.UserID().String(),
		Sex:              sex,
		WeightKg:         weightKg,
		BirthDate:        birthDate,
		HeightCm:         heightCm,
		ProfileCompleted: profile.Completed(),
		CreatedAt:        timestamppb.New(profile.CreatedAt()),
		UpdatedAt:        updatedAt,
	}
}
