package workouts

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

// CreateWorkoutExerciseHandler добавляет упражнение в тренировку.
func CreateWorkoutExerciseHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workoutID := chi.URLParam(r, "workoutId")

		var req CreateWorkoutExerciseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := workoutsClient.CreateWorkoutExercise(r.Context(), createWorkoutExerciseRequestToProto(workoutID, req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusCreated, protoToWorkoutExerciseResponse(resp))
	}
}

// GetWorkoutExercisesHandler возвращает список упражнений тренировки.
func GetWorkoutExercisesHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workoutID := chi.URLParam(r, "workoutId")

		resp, err := workoutsClient.GetWorkoutExercises(r.Context(), workoutID)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToWorkoutExerciseListResponse(resp))
	}
}

// PatchWorkoutExerciseHandler обновляет переданные поля упражнения в тренировке.
func PatchWorkoutExerciseHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workoutID := chi.URLParam(r, "workoutId")
		workoutExerciseID := chi.URLParam(r, "exerciseId")

		var req PatchWorkoutExerciseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := workoutsClient.PatchWorkoutExercise(r.Context(), patchWorkoutExerciseRequestToProto(workoutID, workoutExerciseID, req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToWorkoutExerciseResponse(resp))
	}
}

// DeleteWorkoutExerciseHandler удаляет упражнение из тренировки.
func DeleteWorkoutExerciseHandler(workoutsClient ports.WorkoutsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &workoutspb.DeleteWorkoutExerciseRequest{
			WorkoutId:         chi.URLParam(r, "workoutId"),
			WorkoutExerciseId: chi.URLParam(r, "exerciseId"),
		}

		if err := workoutsClient.DeleteWorkoutExercise(r.Context(), req); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
