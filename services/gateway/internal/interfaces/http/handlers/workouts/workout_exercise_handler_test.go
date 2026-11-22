package workouts

import (
	"context"
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

// serveWorkoutExerciseRequest пропускает запрос через маршрут Chi.
func serveWorkoutExerciseRequest(method, pattern, path, body string, handler http.HandlerFunc) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	router.MethodFunc(method, pattern, handler)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func TestCreateWorkoutExerciseHandler(t *testing.T) {
	const pattern = "/workouts/{workoutId}/exercises"
	const path = "/workouts/workout-1/exercises"

	t.Run("создаёт упражнение", func(t *testing.T) {
		mock := &mockWorkoutsClient{
			createWorkoutExerciseFunc: func(_ context.Context, req *workoutspb.CreateWorkoutExerciseRequest) (*workoutspb.WorkoutExercise, error) {
				assert.Equal(t, "workout-1", req.WorkoutId)
				assert.Equal(t, "catalog-1", req.ExerciseId)
				require.NotNil(t, req.Duration)
				assert.Equal(t, int32(60), *req.Duration)
				return &workoutspb.WorkoutExercise{Id: "entry-1", ExerciseId: req.ExerciseId}, nil
			},
		}

		rr := serveWorkoutExerciseRequest(http.MethodPost, pattern, path,
			`{"exercise_id":"catalog-1","duration":60}`, CreateWorkoutExerciseHandler(mock))

		require.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Body.String(), `"id":"entry-1"`)
	})

	t.Run("отклоняет запрос без exercise_id", func(t *testing.T) {
		mock := &mockWorkoutsClient{
			createWorkoutExerciseFunc: func(_ context.Context, _ *workoutspb.CreateWorkoutExerciseRequest) (*workoutspb.WorkoutExercise, error) {
				t.Fatal("при некорректном запросе вызван Workouts")
				return nil, nil
			},
		}

		rr := serveWorkoutExerciseRequest(http.MethodPost, pattern, path,
			`{"duration":60}`, CreateWorkoutExerciseHandler(mock))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGetWorkoutExercisesHandler(t *testing.T) {
	mock := &mockWorkoutsClient{
		getWorkoutExercisesFunc: func(_ context.Context, workoutID string) (*workoutspb.GetWorkoutExercisesResponse, error) {
			assert.Equal(t, "workout-1", workoutID)
			return &workoutspb.GetWorkoutExercisesResponse{}, nil
		},
	}

	rr := serveWorkoutExerciseRequest(http.MethodGet,
		"/workouts/{workoutId}/exercises", "/workouts/workout-1/exercises", "", GetWorkoutExercisesHandler(mock))

	require.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `[]`, rr.Body.String())
}

func TestPatchWorkoutExerciseHandler(t *testing.T) {
	mock := &mockWorkoutsClient{
		patchWorkoutExerciseFunc: func(_ context.Context, req *workoutspb.PatchWorkoutExerciseRequest) (*workoutspb.WorkoutExercise, error) {
			assert.Equal(t, "workout-1", req.WorkoutId)
			assert.Equal(t, "entry-1", req.WorkoutExerciseId)
			require.NotNil(t, req.Completed)
			assert.False(t, *req.Completed)
			return &workoutspb.WorkoutExercise{Id: req.WorkoutExerciseId}, nil
		},
	}

	rr := serveWorkoutExerciseRequest(http.MethodPatch,
		"/workouts/{workoutId}/exercises/{exerciseId}", "/workouts/workout-1/exercises/entry-1",
		`{"completed":false}`, PatchWorkoutExerciseHandler(mock))

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeleteWorkoutExerciseHandler(t *testing.T) {
	const pattern = "/workouts/{workoutId}/exercises/{exerciseId}"
	const path = "/workouts/workout-1/exercises/entry-1"

	for _, tt := range []struct {
		name string
		err  error
		code int
	}{
		{name: "упражнение удалено", code: http.StatusNoContent},
		{name: "упражнение не найдено", err: status.Error(codes.NotFound, "workout exercise not found"), code: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockWorkoutsClient{
				deleteWorkoutExerciseFunc: func(_ context.Context, req *workoutspb.DeleteWorkoutExerciseRequest) error {
					assert.Equal(t, "workout-1", req.WorkoutId)
					assert.Equal(t, "entry-1", req.WorkoutExerciseId)
					return tt.err
				},
			}

			rr := serveWorkoutExerciseRequest(http.MethodDelete, pattern, path, "", DeleteWorkoutExerciseHandler(mock))

			require.Equal(t, tt.code, rr.Code)
			if tt.code == http.StatusNoContent {
				assert.Empty(t, rr.Body.String())
			}
		})
	}
}
