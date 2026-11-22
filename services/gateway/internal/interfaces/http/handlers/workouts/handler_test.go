package workouts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
)

// workoutRequest создаёт HTTP-запрос с параметром маршрута workoutId.
func workoutRequest(method, id string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, "/workouts/"+id, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("workoutId", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestGetWorkoutsHandler(t *testing.T) {
	t.Run("список тренировок", func(t *testing.T) {
		mock := &mockWorkoutsClient{
			getWorkoutsFunc: func(_ context.Context) (*workoutspb.GetWorkoutsResponse, error) {
				return &workoutspb.GetWorkoutsResponse{
					Workouts: []*workoutspb.Workout{
						{Id: "workout-1", Status: "planned"},
						{Id: "workout-2", Status: "completed"},
					},
				}, nil
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/workouts", nil)
		rr := httptest.NewRecorder()
		GetWorkoutsHandler(mock).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []WorkoutResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		require.Len(t, resp, 2)
		assert.Equal(t, "workout-1", resp[0].ID)
		assert.Equal(t, "completed", resp[1].Status)
	})

	t.Run("пустой список возвращает массив", func(t *testing.T) {
		mock := &mockWorkoutsClient{
			getWorkoutsFunc: func(_ context.Context) (*workoutspb.GetWorkoutsResponse, error) {
				return &workoutspb.GetWorkoutsResponse{}, nil
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/workouts", nil)
		rr := httptest.NewRecorder()
		GetWorkoutsHandler(mock).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `[]`, rr.Body.String())
	})
}

func TestCreateWorkoutHandler(t *testing.T) {
	mock := &mockWorkoutsClient{
		createWorkoutFunc: func(_ context.Context) (*workoutspb.Workout, error) {
			return &workoutspb.Workout{Id: "new-workout", Status: "planned"}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/workouts", nil)
	rr := httptest.NewRecorder()
	CreateWorkoutHandler(mock).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp WorkoutResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "new-workout", resp.ID)
	assert.Equal(t, "planned", resp.Status)
}

func TestGetWorkoutHandler(t *testing.T) {
	t.Run("тренировка найдена", func(t *testing.T) {
		mock := &mockWorkoutsClient{
			getWorkoutFunc: func(_ context.Context, id string) (*workoutspb.Workout, error) {
				assert.Equal(t, "workout-1", id)
				return &workoutspb.Workout{Id: id, Status: "planned"}, nil
			},
		}

		req := workoutRequest(http.MethodGet, "workout-1", nil)
		rr := httptest.NewRecorder()
		GetWorkoutHandler(mock).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp WorkoutResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "workout-1", resp.ID)
	})

	t.Run("тренировка не найдена", func(t *testing.T) {
		mock := &mockWorkoutsClient{
			getWorkoutFunc: func(_ context.Context, _ string) (*workoutspb.Workout, error) {
				return nil, status.Error(codes.NotFound, "workout not found")
			},
		}

		req := workoutRequest(http.MethodGet, "missing", nil)
		rr := httptest.NewRecorder()
		GetWorkoutHandler(mock).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestPatchWorkoutHandler(t *testing.T) {
	intensity := int32(7)
	tests := []struct {
		name          string
		body          string
		wantIntensity *workoutspb.NullableInt32
	}{
		{
			name: "интенсивность отсутствует",
			body: `{}`,
		},
		{
			name:          "интенсивность очищается",
			body:          `{"intensity":null}`,
			wantIntensity: &workoutspb.NullableInt32{},
		},
		{
			name:          "интенсивность обновляется",
			body:          `{"intensity":7}`,
			wantIntensity: &workoutspb.NullableInt32{Value: &intensity},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockWorkoutsClient{
				patchWorkoutFunc: func(_ context.Context, req *workoutspb.PatchWorkoutRequest) (*workoutspb.Workout, error) {
					assert.Equal(t, "workout-1", req.Id)
					if tt.wantIntensity == nil {
						assert.Nil(t, req.Intensity)
					} else {
						require.NotNil(t, req.Intensity)
						assert.Equal(t, tt.wantIntensity.Value, req.Intensity.Value)
					}

					return &workoutspb.Workout{Id: req.Id}, nil
				},
			}

			req := workoutRequest(http.MethodPatch, "workout-1", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			PatchWorkoutHandler(mock).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			var resp WorkoutResponse
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			assert.Equal(t, "workout-1", resp.ID)
		})
	}

	for _, tt := range []struct {
		name string
		body string
	}{
		{name: "некорректный JSON", body: `{"intensity":`},
		{name: "недопустимый статус", body: `{"status":"invalid"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockWorkoutsClient{
				patchWorkoutFunc: func(_ context.Context, _ *workoutspb.PatchWorkoutRequest) (*workoutspb.Workout, error) {
					t.Fatal("при некорректном запросе вызван Workouts")
					return nil, nil
				},
			}

			req := workoutRequest(http.MethodPatch, "workout-1", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			PatchWorkoutHandler(mock).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestDeleteWorkoutHandler(t *testing.T) {
	mock := &mockWorkoutsClient{
		deleteWorkoutFunc: func(_ context.Context, id string) error {
			assert.Equal(t, "workout-1", id)
			return nil
		},
	}

	req := workoutRequest(http.MethodDelete, "workout-1", nil)
	rr := httptest.NewRecorder()
	DeleteWorkoutHandler(mock).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Body.String())
}
