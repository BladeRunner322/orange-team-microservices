package exercises

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	exercisespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// requestWithURLParam создаёт HTTP-запрос с установленным chi URL-параметром.
// Нужен для методов, которые читают chi.URLParam(r, "exerciseId").
func requestWithURLParam(method, target, paramName, paramValue string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(paramName, paramValue)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ===== GetExercisesHandler =====
func TestGetExercisesHandler(t *testing.T) {
	t.Run("success — список из двух упражнений", func(t *testing.T) {
		mock := &mockExercisesClient{
			getExercisesFunc: func(ctx context.Context) (*exercisespb.GetExercisesResponse, error) {
				return &exercisespb.GetExercisesResponse{
					Exercises: []*exercisespb.Exercise{
						{Id: "id-1", Name: "Жим лёжа", Description: "desc 1", Difficulty: 5, Type: "weight"},
						{Id: "id-2", Name: "Планка", Description: "desc 2", Difficulty: 3, Type: "duration"},
					},
				}, nil
			},
		}
		handler := GetExercisesHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp []ExerciseResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Len(t, resp, 2)
		assert.Equal(t, "Жим лёжа", resp[0].Name)
		assert.Equal(t, "Планка", resp[1].Name)
	})

	t.Run("пустой список — возвращает [] не null", func(t *testing.T) {
		mock := &mockExercisesClient{
			getExercisesFunc: func(ctx context.Context) (*exercisespb.GetExercisesResponse, error) {
				return &exercisespb.GetExercisesResponse{Exercises: []*exercisespb.Exercise{}}, nil
			},
		}
		handler := GetExercisesHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "[]\n", rr.Body.String())
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockExercisesClient{
			getExercisesFunc: func(ctx context.Context) (*exercisespb.GetExercisesResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}
		handler := GetExercisesHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// ===== GetExerciseHandler =====
func TestGetExerciseHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockExercisesClient{
			getExerciseFunc: func(ctx context.Context, id string) (*exercisespb.Exercise, error) {
				return &exercisespb.Exercise{
					Id:          id,
					Name:        "Жим лёжа",
					Description: "Базовое",
					Difficulty:  5,
					Type:        "weight",
				}, nil
			},
		}
		handler := GetExerciseHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/exercises/abc-123", "exerciseId", "abc-123", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp ExerciseResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "abc-123", resp.ID)
		assert.Equal(t, "Жим лёжа", resp.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockExercisesClient{
			getExerciseFunc: func(ctx context.Context, id string) (*exercisespb.Exercise, error) {
				return nil, status.Error(codes.NotFound, "exercise not found")
			},
		}
		handler := GetExerciseHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/exercises/nope", "exerciseId", "nope", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "exercise not found")
	})
}

// ===== CreateExerciseHandler =====
func TestCreateExerciseHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockExercisesClient{
			createExerciseFunc: func(ctx context.Context, req *exercisespb.CreateExerciseRequest) (*exercisespb.Exercise, error) {
				return &exercisespb.Exercise{
					Id:          "new-id",
					Name:        req.Name,
					Description: req.Description,
					Difficulty:  req.Difficulty,
					Type:        req.Type,
				}, nil
			},
		}
		handler := CreateExerciseHandler(mock)

		body, _ := json.Marshal(CreateExerciseRequest{
			Name:        "Жим лёжа",
			Description: "Базовое упражнение",
			Difficulty:  5,
			Type:        "weight",
		})
		req := httptest.NewRequest(http.MethodPost, "/exercises", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp ExerciseResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "new-id", resp.ID)
		assert.Equal(t, "Жим лёжа", resp.Name)
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockExercisesClient{}
		handler := CreateExerciseHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/exercises", bytes.NewReader([]byte("not json")))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("validation error — короткое имя", func(t *testing.T) {
		mock := &mockExercisesClient{}
		handler := CreateExerciseHandler(mock)

		body, _ := json.Marshal(CreateExerciseRequest{
			Name:        "Жи", // 2 руны < 3
			Description: "desc",
			Difficulty:  5,
			Type:        "weight",
		})
		req := httptest.NewRequest(http.MethodPost, "/exercises", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed on the 'min' tag")
	})

	t.Run("validation error — неверный type", func(t *testing.T) {
		mock := &mockExercisesClient{}
		handler := CreateExerciseHandler(mock)

		body, _ := json.Marshal(CreateExerciseRequest{
			Name:        "Жим лёжа",
			Description: "desc",
			Difficulty:  5,
			Type:        "cardio",
		})
		req := httptest.NewRequest(http.MethodPost, "/exercises", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed on the 'oneof' tag")
	})

	t.Run("gRPC error — конфликт имён", func(t *testing.T) {
		mock := &mockExercisesClient{
			createExerciseFunc: func(ctx context.Context, req *exercisespb.CreateExerciseRequest) (*exercisespb.Exercise, error) {
				return nil, status.Error(codes.AlreadyExists, "exercise name already exists")
			},
		}
		handler := CreateExerciseHandler(mock)

		body, _ := json.Marshal(CreateExerciseRequest{
			Name:        "Дубликат",
			Description: "desc",
			Difficulty:  5,
			Type:        "weight",
		})
		req := httptest.NewRequest(http.MethodPost, "/exercises", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "already exists")
	})
}

// ===== PatchExerciseHandler =====
func TestPatchExerciseHandler(t *testing.T) {
	t.Run("success — патч только имени", func(t *testing.T) {
		var captured *exercisespb.PatchExerciseRequest

		mock := &mockExercisesClient{
			patchExerciseFunc: func(ctx context.Context, req *exercisespb.PatchExerciseRequest) (*exercisespb.Exercise, error) {
				captured = req
				return &exercisespb.Exercise{Id: req.Id, Name: *req.Name}, nil
			},
		}
		handler := PatchExerciseHandler(mock)

		body, _ := json.Marshal(map[string]any{"name": "Новое имя"})
		req := requestWithURLParam(http.MethodPatch, "/exercises/abc-123", "exerciseId", "abc-123", body)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "abc-123", captured.Id)
		require.NotNil(t, captured.Name)
		assert.Equal(t, "Новое имя", *captured.Name)
		assert.Nil(t, captured.Description, "description не пришёл — nil")
		assert.Nil(t, captured.Difficulty, "difficulty не пришёл — nil")
	})

	t.Run("success — патч difficulty", func(t *testing.T) {
		var captured *exercisespb.PatchExerciseRequest

		mock := &mockExercisesClient{
			patchExerciseFunc: func(ctx context.Context, req *exercisespb.PatchExerciseRequest) (*exercisespb.Exercise, error) {
				captured = req
				return &exercisespb.Exercise{Id: req.Id}, nil
			},
		}
		handler := PatchExerciseHandler(mock)

		body, _ := json.Marshal(map[string]any{"difficulty": 9})
		req := requestWithURLParam(http.MethodPatch, "/exercises/abc-123", "exerciseId", "abc-123", body)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		require.NotNil(t, captured.Difficulty)
		assert.Equal(t, int32(9), *captured.Difficulty)
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockExercisesClient{}
		handler := PatchExerciseHandler(mock)

		req := requestWithURLParam(http.MethodPatch, "/exercises/abc", "exerciseId", "abc", []byte("not json"))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockExercisesClient{
			patchExerciseFunc: func(ctx context.Context, req *exercisespb.PatchExerciseRequest) (*exercisespb.Exercise, error) {
				return nil, status.Error(codes.NotFound, "exercise not found")
			},
		}
		handler := PatchExerciseHandler(mock)

		body, _ := json.Marshal(map[string]any{"name": "Новое"})
		req := requestWithURLParam(http.MethodPatch, "/exercises/nope", "exerciseId", "nope", body)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// ===== DeleteExerciseHandler =====
func TestDeleteExerciseHandler(t *testing.T) {
	t.Run("success — 204", func(t *testing.T) {
		mock := &mockExercisesClient{
			deleteExerciseFunc: func(ctx context.Context, id string) error {
				return nil
			},
		}
		handler := DeleteExerciseHandler(mock)

		req := requestWithURLParam(http.MethodDelete, "/exercises/abc-123", "exerciseId", "abc-123", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String())
	})

	t.Run("идемпотентность — повторный DELETE тоже успех", func(t *testing.T) {
		callCount := 0
		mock := &mockExercisesClient{
			deleteExerciseFunc: func(ctx context.Context, id string) error {
				callCount++
				return nil // Exercises идемпотентен, всегда nil
			},
		}
		handler := DeleteExerciseHandler(mock)

		for i := 0; i < 2; i++ {
			req := requestWithURLParam(http.MethodDelete, "/exercises/abc-123", "exerciseId", "abc-123", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNoContent, rr.Code)
		}

		assert.Equal(t, 2, callCount)
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockExercisesClient{
			deleteExerciseFunc: func(ctx context.Context, id string) error {
				return status.Error(codes.Internal, "internal error")
			},
		}
		handler := DeleteExerciseHandler(mock)

		req := requestWithURLParam(http.MethodDelete, "/exercises/abc", "exerciseId", "abc", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
