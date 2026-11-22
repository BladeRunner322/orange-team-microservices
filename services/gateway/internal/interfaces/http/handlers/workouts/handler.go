package workouts

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

var validate = validator.New()

// GetWorkoutsHandler возвращает список тренировок текущего пользователя.
func GetWorkoutsHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := workoutsClient.GetWorkouts(r.Context())
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToWorkoutListResponse(resp))
	}
}

// CreateWorkoutHandler создаёт тренировку для текущего пользователя.
func CreateWorkoutHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := workoutsClient.CreateWorkout(r.Context())
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusCreated, protoToWorkoutResponse(resp))
	}
}

// GetWorkoutHandler возвращает тренировку по идентификатору.
func GetWorkoutHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "workoutId")

		resp, err := workoutsClient.GetWorkout(r.Context(), id)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToWorkoutResponse(resp))
	}
}

// PatchWorkoutHandler обновляет переданные поля тренировки.
func PatchWorkoutHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "workoutId")

		var req PatchWorkoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := workoutsClient.PatchWorkout(r.Context(), patchRequestToProto(id, req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToWorkoutResponse(resp))
	}
}

// DeleteWorkoutHandler помечает тренировку удалённой.
func DeleteWorkoutHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "workoutId")

		if err := workoutsClient.DeleteWorkout(r.Context(), id); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
