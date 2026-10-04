package profiles

import (
	"encoding/json"
	"net/http"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

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
