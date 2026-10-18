package habits

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	habitspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// requestWithURLParam создаёт HTTP-запрос с chi URL-параметром.
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

// ===== GetHabitsHandler =====
func TestGetHabitsHandler(t *testing.T) {
	t.Run("success — список из двух", func(t *testing.T) {
		mock := &mockHabitsClient{
			getHabitsFunc: func(ctx context.Context) (*habitspb.GetHabitsResponse, error) {
				return &habitspb.GetHabitsResponse{
					Habits: []*habitspb.Habit{
						{Id: "id-1", Name: "Пить воду", Streak: 5, CompletedToday: true},
						{Id: "id-2", Name: "Читать", Streak: 2, CompletedToday: false},
					},
				}, nil
			},
		}
		handler := GetHabitsHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/habits", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp []HabitResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		require.Len(t, resp, 2)
		assert.Equal(t, "Пить воду", resp[0].Name)
		assert.Equal(t, 5, resp[0].Streak)
		assert.True(t, resp[0].CompletedToday)
	})

	t.Run("пустой список — [] не null", func(t *testing.T) {
		mock := &mockHabitsClient{
			getHabitsFunc: func(ctx context.Context) (*habitspb.GetHabitsResponse, error) {
				return &habitspb.GetHabitsResponse{Habits: []*habitspb.Habit{}}, nil
			},
		}
		handler := GetHabitsHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/habits", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "[]\n", rr.Body.String())
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockHabitsClient{
			getHabitsFunc: func(ctx context.Context) (*habitspb.GetHabitsResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}
		handler := GetHabitsHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/habits", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// ===== GetHabitHandler =====
func TestGetHabitHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockHabitsClient{
			getHabitFunc: func(ctx context.Context, id string) (*habitspb.Habit, error) {
				return &habitspb.Habit{Id: id, Name: "Пить воду", Streak: 3}, nil
			},
		}
		handler := GetHabitHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/habits/abc", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp HabitResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "abc", resp.ID)
		assert.Equal(t, 3, resp.Streak)
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockHabitsClient{
			getHabitFunc: func(ctx context.Context, id string) (*habitspb.Habit, error) {
				return nil, status.Error(codes.NotFound, "habit not found")
			},
		}
		handler := GetHabitHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/habits/nope", "habitId", "nope", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// ===== CreateHabitHandler =====
func TestCreateHabitHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockHabitsClient{
			createHabitFunc: func(ctx context.Context, req *habitspb.CreateHabitRequest) (*habitspb.Habit, error) {
				return &habitspb.Habit{Id: "new-id", Name: req.Name, Description: req.Description}, nil
			},
		}
		handler := CreateHabitHandler(mock)

		body, _ := json.Marshal(CreateHabitRequest{Name: "Пить воду", Description: "2 литра"})
		req := httptest.NewRequest(http.MethodPost, "/habits", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp HabitResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "new-id", resp.ID)
		assert.Equal(t, "Пить воду", resp.Name)
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockHabitsClient{}
		handler := CreateHabitHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/habits", bytes.NewReader([]byte("not json")))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("validation — пустое имя", func(t *testing.T) {
		mock := &mockHabitsClient{}
		handler := CreateHabitHandler(mock)

		body, _ := json.Marshal(CreateHabitRequest{Name: ""})
		req := httptest.NewRequest(http.MethodPost, "/habits", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed on the 'required' tag")
	})

	t.Run("gRPC error — name exists", func(t *testing.T) {
		mock := &mockHabitsClient{
			createHabitFunc: func(ctx context.Context, req *habitspb.CreateHabitRequest) (*habitspb.Habit, error) {
				return nil, status.Error(codes.AlreadyExists, "habit with this name already exists")
			},
		}
		handler := CreateHabitHandler(mock)

		body, _ := json.Marshal(CreateHabitRequest{Name: "Дубликат"})
		req := httptest.NewRequest(http.MethodPost, "/habits", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})
}

// ===== PatchHabitHandler =====
func TestPatchHabitHandler(t *testing.T) {
	t.Run("success — только name", func(t *testing.T) {
		var captured *habitspb.PatchHabitRequest

		mock := &mockHabitsClient{
			patchHabitFunc: func(ctx context.Context, req *habitspb.PatchHabitRequest) (*habitspb.Habit, error) {
				captured = req
				return &habitspb.Habit{Id: req.Id, Name: *req.Name}, nil
			},
		}
		handler := PatchHabitHandler(mock)

		body, _ := json.Marshal(map[string]any{"name": "Новое имя"})
		req := requestWithURLParam(http.MethodPatch, "/habits/abc", "habitId", "abc", body)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "abc", captured.Id)
		require.NotNil(t, captured.Name)
		assert.Equal(t, "Новое имя", *captured.Name)
		assert.Nil(t, captured.Description)
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockHabitsClient{}
		handler := PatchHabitHandler(mock)

		req := requestWithURLParam(http.MethodPatch, "/habits/abc", "habitId", "abc", []byte("not json"))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockHabitsClient{
			patchHabitFunc: func(ctx context.Context, req *habitspb.PatchHabitRequest) (*habitspb.Habit, error) {
				return nil, status.Error(codes.NotFound, "habit not found")
			},
		}
		handler := PatchHabitHandler(mock)

		body, _ := json.Marshal(map[string]any{"name": "Новое"})
		req := requestWithURLParam(http.MethodPatch, "/habits/nope", "habitId", "nope", body)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// ===== DeleteHabitHandler =====
func TestDeleteHabitHandler(t *testing.T) {
	t.Run("success — 204", func(t *testing.T) {
		mock := &mockHabitsClient{}
		handler := DeleteHabitHandler(mock)

		req := requestWithURLParam(http.MethodDelete, "/habits/abc", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String())
	})

	t.Run("gRPC error", func(t *testing.T) {
		mock := &mockHabitsClient{
			deleteHabitFunc: func(ctx context.Context, id string) error {
				return status.Error(codes.Internal, "internal error")
			},
		}
		handler := DeleteHabitHandler(mock)

		req := requestWithURLParam(http.MethodDelete, "/habits/abc", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// ===== CompleteHabitHandler =====
func TestCompleteHabitHandler(t *testing.T) {
	t.Run("success — 204 без тела", func(t *testing.T) {
		var captured *habitspb.CompleteHabitRequest

		mock := &mockHabitsClient{
			completeHabitFunc: func(ctx context.Context, req *habitspb.CompleteHabitRequest) error {
				captured = req
				return nil
			},
		}
		handler := CompleteHabitHandler(mock)

		req := requestWithURLParam(http.MethodPost, "/habits/abc/complete", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "abc", captured.Id)
		assert.Equal(t, "", captured.Date, "date пустой → usecase применит today")
	})

	t.Run("success — с date", func(t *testing.T) {
		var captured *habitspb.CompleteHabitRequest

		mock := &mockHabitsClient{
			completeHabitFunc: func(ctx context.Context, req *habitspb.CompleteHabitRequest) error {
				captured = req
				return nil
			},
		}
		handler := CompleteHabitHandler(mock)

		body, _ := json.Marshal(CompleteHabitRequest{Date: "2026-10-15"})
		req := requestWithURLParam(http.MethodPost, "/habits/abc/complete", "habitId", "abc", body)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "2026-10-15", captured.Date)
	})

	t.Run("invalid date format", func(t *testing.T) {
		mock := &mockHabitsClient{}
		handler := CompleteHabitHandler(mock)

		body, _ := json.Marshal(CompleteHabitRequest{Date: "not-a-date"})
		req := requestWithURLParam(http.MethodPost, "/habits/abc/complete", "habitId", "abc", body)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("gRPC error — invalid date", func(t *testing.T) {
		mock := &mockHabitsClient{
			completeHabitFunc: func(ctx context.Context, req *habitspb.CompleteHabitRequest) error {
				return status.Error(codes.InvalidArgument, "invalid date")
			},
		}
		handler := CompleteHabitHandler(mock)

		req := requestWithURLParam(http.MethodPost, "/habits/abc/complete", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// ===== UncompleteHabitHandler =====
func TestUncompleteHabitHandler(t *testing.T) {
	t.Run("success — 204 без тела", func(t *testing.T) {
		var captured *habitspb.UncompleteHabitRequest

		mock := &mockHabitsClient{
			uncompleteHabitFunc: func(ctx context.Context, req *habitspb.UncompleteHabitRequest) error {
				captured = req
				return nil
			},
		}
		handler := UncompleteHabitHandler(mock)

		req := requestWithURLParam(http.MethodDelete, "/habits/abc/complete", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "abc", captured.Id)
	})

	t.Run("success — с date", func(t *testing.T) {
		var captured *habitspb.UncompleteHabitRequest

		mock := &mockHabitsClient{
			uncompleteHabitFunc: func(ctx context.Context, req *habitspb.UncompleteHabitRequest) error {
				captured = req
				return nil
			},
		}
		handler := UncompleteHabitHandler(mock)

		body, _ := json.Marshal(UncompleteHabitRequest{Date: "2026-10-14"})
		req := requestWithURLParam(http.MethodDelete, "/habits/abc/complete", "habitId", "abc", body)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "2026-10-14", captured.Date)
	})

	t.Run("gRPC error — not found", func(t *testing.T) {
		mock := &mockHabitsClient{
			uncompleteHabitFunc: func(ctx context.Context, req *habitspb.UncompleteHabitRequest) error {
				return status.Error(codes.NotFound, "habit not found")
			},
		}
		handler := UncompleteHabitHandler(mock)

		req := requestWithURLParam(http.MethodDelete, "/habits/nope/complete", "habitId", "nope", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// ===== GetCompletionsHandler =====
func TestGetCompletionsHandler(t *testing.T) {
	t.Run("success — с периодом", func(t *testing.T) {
		var captured *habitspb.GetCompletionsRequest

		mock := &mockHabitsClient{
			getCompletionsFunc: func(ctx context.Context, req *habitspb.GetCompletionsRequest) (*habitspb.GetCompletionsResponse, error) {
				captured = req
				return &habitspb.GetCompletionsResponse{
					Completions: []*habitspb.Completion{
						{HabitId: "abc", Date: "2026-10-14"},
					},
				}, nil
			},
		}
		handler := GetCompletionsHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/habits/abc/completions?from=2026-10-01&to=2026-10-15", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "abc", captured.Id)
		assert.Equal(t, "2026-10-01", captured.From)
		assert.Equal(t, "2026-10-15", captured.To)

		var resp []CompletionResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		require.Len(t, resp, 1)
		assert.Equal(t, "2026-10-14", resp[0].Date)
	})

	t.Run("success — без периода", func(t *testing.T) {
		var captured *habitspb.GetCompletionsRequest

		mock := &mockHabitsClient{
			getCompletionsFunc: func(ctx context.Context, req *habitspb.GetCompletionsRequest) (*habitspb.GetCompletionsResponse, error) {
				captured = req
				return &habitspb.GetCompletionsResponse{Completions: []*habitspb.Completion{}}, nil
			},
		}
		handler := GetCompletionsHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/habits/abc/completions", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, captured)
		assert.Equal(t, "", captured.From)
		assert.Equal(t, "", captured.To)
		assert.Equal(t, "[]\n", rr.Body.String())
	})

	t.Run("gRPC error — invalid period", func(t *testing.T) {
		mock := &mockHabitsClient{
			getCompletionsFunc: func(ctx context.Context, req *habitspb.GetCompletionsRequest) (*habitspb.GetCompletionsResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid period")
			},
		}
		handler := GetCompletionsHandler(mock)

		req := requestWithURLParam(http.MethodGet, "/habits/abc/completions?from=2026-10-15&to=2026-10-01", "habitId", "abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
