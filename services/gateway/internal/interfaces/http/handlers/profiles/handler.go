package profiles

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/sync/errgroup"

	profilespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

// workoutScoreClient получает суммарный балл за тренировки текущего пользователя.
type workoutScoreClient interface {
	GetUserScore(ctx context.Context) (*workoutspb.GetUserScoreResponse, error)
}

// GetUserHandler возвращает профиль текущего пользователя и балл за тренировки.
func GetUserHandler(profilesClient ports.ProfilesClientInterface, workoutsClient workoutScoreClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		group, ctx := errgroup.WithContext(r.Context())

		var profile *profilespb.UserProfile
		var score *workoutspb.GetUserScoreResponse

		group.Go(func() error {
			resp, err := profilesClient.GetMyProfile(ctx)
			if err != nil {
				return err
			}
			profile = resp
			return nil
		})

		group.Go(func() error {
			resp, err := workoutsClient.GetUserScore(ctx)
			if err == nil {
				score = resp
			}
			// Профиль можно вернуть и при недоступности Workouts.
			return nil
		})

		if err := group.Wait(); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			if code >= http.StatusInternalServerError {
				code = http.StatusServiceUnavailable
			}
			httputil.SendError(w, code, msg)
			return
		}

		dto := GetUserResponse{
			UserProfileResponse:   protoToUserProfileResponse(profile),
			WorkoutScore:          score.GetScore(),
			WorkoutScoreAvailable: score != nil,
		}
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
