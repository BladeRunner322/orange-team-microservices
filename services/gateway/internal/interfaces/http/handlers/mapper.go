package handlers

import (
	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
)

// protoToUserProfileResponse преобразует proto-ответ в HTTP-DTO.
func protoToUserProfileResponse(p *profiles.UserProfile) UserProfileResponse {
	resp := UserProfileResponse{
		UserID:           p.UserId,
		Sex:              p.Sex,
		WeightKg:         p.WeightKg,
		BirthDate:        p.BirthDate,
		HeightCm:         p.HeightCm,
		ProfileCompleted: p.ProfileCompleted,
	}

	if p.CreatedAt != nil {
		resp.CreatedAt = p.CreatedAt.AsTime()
	}

	if p.UpdatedAt != nil {
		t := p.UpdatedAt.AsTime()
		resp.UpdatedAt = &t
	}

	return resp
}

// patchRequestToProto преобразует HTTP-DTO патча в proto-запрос.
func patchRequestToProto(req PatchUserRequest) *profiles.PatchMyProfileRequest {
	return &profiles.PatchMyProfileRequest{
		Sex:       nullableStringToProto(req.Sex),
		WeightKg:  nullableDoubleToProto(req.WeightKg),
		BirthDate: nullableStringToProto(req.BirthDate),
		HeightCm:  nullableInt32ToProto(req.HeightCm),
	}
}

// nullableStringToProto конвертирует nullable-поле string в proto-обёртку,
// сохраняя три состояния: не задано / сброс в NULL / значение.
func nullableStringToProto(n nullable.Nullable[string]) *profiles.NullableString {
	if !n.Set {
		return nil
	}

	if n.Value == nil {
		return &profiles.NullableString{}
	}

	return &profiles.NullableString{Value: n.Value}
}

// nullableDoubleToProto конвертирует nullable-поле float64 в proto-обёртку,
// сохраняя три состояния: не задано / сброс в NULL / значение.
func nullableDoubleToProto(n nullable.Nullable[float64]) *profiles.NullableDouble {
	if !n.Set {
		return nil
	}

	if n.Value == nil {
		return &profiles.NullableDouble{}
	}

	return &profiles.NullableDouble{Value: n.Value}
}

// nullableInt32ToProto конвертирует nullable-поле int32 в proto-обёртку,
// сохраняя три состояния: не задано / сброс в NULL / значение.
func nullableInt32ToProto(n nullable.Nullable[int32]) *profiles.NullableInt32 {
	if !n.Set {
		return nil
	}

	if n.Value == nil {
		return &profiles.NullableInt32{}
	}

	return &profiles.NullableInt32{Value: n.Value}
}
