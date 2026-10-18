package habits

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/interfaces/http/httputil"
)

var validate = validator.New()

// GetHabitsHandler возвращает список привычек пользователя со стриками.
func GetHabitsHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := habitsClient.GetHabits(r.Context())
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToHabitListResponse(resp))
	}
}

// GetHabitHandler возвращает привычку по id.
func GetHabitHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "habitId")

		resp, err := habitsClient.GetHabit(r.Context(), id)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToHabitResponse(resp))
	}
}

// CreateHabitHandler создаёт новую привычку.
func CreateHabitHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateHabitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := habitsClient.CreateHabit(r.Context(), createRequestToProto(req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusCreated, protoToHabitResponse(resp))
	}
}

// PatchHabitHandler применяет патч к привычке.
func PatchHabitHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "habitId")

		var req PatchHabitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := habitsClient.PatchHabit(r.Context(), patchRequestToProto(id, req))
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToHabitResponse(resp))
	}
}

// DeleteHabitHandler помечает привычку удалённой (soft delete, идемпотентен).
func DeleteHabitHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "habitId")

		if err := habitsClient.DeleteHabit(r.Context(), id); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// CompleteHabitHandler отмечает выполнение привычки на дату.
func CompleteHabitHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "habitId")

		var req CompleteHabitRequest
		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httputil.SendError(w, http.StatusBadRequest, "invalid request body")
				return
			}
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := habitsClient.CompleteHabit(r.Context(), completeRequestToProto(id, req)); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UncompleteHabitHandler снимает отметку выполнения привычки.
func UncompleteHabitHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "habitId")

		var req UncompleteHabitRequest
		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httputil.SendError(w, http.StatusBadRequest, "invalid request body")
				return
			}
		}

		if err := validate.Struct(req); err != nil {
			httputil.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := habitsClient.UncompleteHabit(r.Context(), uncompleteRequestToProto(id, req)); err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetCompletionsHandler возвращает историю выполнений за период.
func GetCompletionsHandler(habitsClient ports.HabitsClientInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "habitId")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		req := &habits.GetCompletionsRequest{
			Id:   id,
			From: from,
			To:   to,
		}

		resp, err := habitsClient.GetCompletions(r.Context(), req)
		if err != nil {
			code, msg := httputil.GrpcErrorToHTTP(err)
			httputil.SendError(w, code, msg)
			return
		}

		httputil.SendJSON(w, http.StatusOK, protoToCompletionListResponse(resp))
	}
}
