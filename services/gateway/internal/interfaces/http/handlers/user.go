package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

// PatchUserRequest — тело запроса PATCH /users/me.
// Nullable-поля позволяют отличить «не трогать» от «сбросить в NULL».
type PatchUserRequest struct {
	Sex       nullable.Nullable[string]  `json:"sex"`
	WeightKg  nullable.Nullable[float64] `json:"weight_kg"`
	BirthDate nullable.Nullable[string]  `json:"birth_date"`
	HeightCm  nullable.Nullable[int32]   `json:"height_cm"`
}

// UserProfileResponse — ответ с профилем пользователя.
type UserProfileResponse struct {
	UserID           string     `json:"user_id"`
	Sex              string     `json:"sex,omitempty"`
	WeightKg         float64    `json:"weight_kg,omitempty"`
	BirthDate        string     `json:"birth_date,omitempty"`
	HeightCm         int32      `json:"height_cm,omitempty"`
	ProfileCompleted bool       `json:"profile_completed"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

// GetUserHandler возвращает профиль текущего пользователя.
func GetUserHandler(profilesClient ports.ProfilesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := profilesClient.GetMyProfile(r.Context())
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		dto := protoToUserProfileResponse(resp)
		httputil.SendJSON(w, http.StatusOK, dto)
	}
}

// PatchUserHandler применяет патч к профилю текущего пользователя.
func PatchUserHandler(profilesClient ports.ProfilesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PatchUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		protoReq := patchRequestToProto(req)

		resp, err := profilesClient.PatchMyProfile(r.Context(), protoReq)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		dto := protoToUserProfileResponse(resp)
		httputil.SendJSON(w, http.StatusOK, dto)
	}
}

// DeleteUserHandler удаляет профиль текущего пользователя.
func DeleteUserHandler(profilesClient ports.ProfilesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := profilesClient.DeleteMyProfile(r.Context()); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
