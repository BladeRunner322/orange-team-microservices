package exercises

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

var validate = validator.New()

// GetExercisesHandler возвращает список активных упражнений.
func GetExercisesHandler(exercisesClient ports.ExercisesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := exercisesClient.GetExercises(r.Context())
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToExerciseListResponse(resp))
	}
}

// GetExerciseHandler возвращает упражнение по id.
func GetExerciseHandler(exercisesClient ports.ExercisesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "exerciseId")

		resp, err := exercisesClient.GetExercise(r.Context(), id)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToExerciseResponse(resp))
	}
}

// CreateExerciseHandler создаёт новое упражнение (admin-only).
func CreateExerciseHandler(exercisesClient ports.ExercisesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateExerciseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := exercisesClient.CreateExercise(r.Context(), createRequestToProto(req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusCreated, protoToExerciseResponse(resp))
	}
}

// PatchExerciseHandler применяет патч к упражнению (admin-only).
func PatchExerciseHandler(exercisesClient ports.ExercisesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "exerciseId")

		var req PatchExerciseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := exercisesClient.PatchExercise(r.Context(), patchRequestToProto(id, req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToExerciseResponse(resp))
	}
}

// DeleteExerciseHandler помечает упражнение удалённым (admin-only, идемпотентен).
func DeleteExerciseHandler(exercisesClient ports.ExercisesClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "exerciseId")

		if err := exercisesClient.DeleteExercise(r.Context(), id); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
