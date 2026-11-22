package workouts

import (
	"context"

	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/services/gateway/internal/application/ports"
)

// mockWorkoutsClient — тестовая реализация операций с тренировками.
type mockWorkoutsClient struct {
	ports.WorkoutsClientInterface

	getWorkoutsFunc   func(ctx context.Context) (*workoutspb.GetWorkoutsResponse, error)
	getWorkoutFunc    func(ctx context.Context, id string) (*workoutspb.Workout, error)
	createWorkoutFunc func(ctx context.Context) (*workoutspb.Workout, error)
	patchWorkoutFunc  func(ctx context.Context, req *workoutspb.PatchWorkoutRequest) (*workoutspb.Workout, error)
	deleteWorkoutFunc func(ctx context.Context, id string) error
}

func (m *mockWorkoutsClient) GetWorkouts(ctx context.Context) (*workoutspb.GetWorkoutsResponse, error) {
	return m.getWorkoutsFunc(ctx)
}

func (m *mockWorkoutsClient) GetWorkout(ctx context.Context, id string) (*workoutspb.Workout, error) {
	return m.getWorkoutFunc(ctx, id)
}

func (m *mockWorkoutsClient) CreateWorkout(ctx context.Context) (*workoutspb.Workout, error) {
	return m.createWorkoutFunc(ctx)
}

func (m *mockWorkoutsClient) PatchWorkout(ctx context.Context, req *workoutspb.PatchWorkoutRequest) (*workoutspb.Workout, error) {
	return m.patchWorkoutFunc(ctx, req)
}

func (m *mockWorkoutsClient) DeleteWorkout(ctx context.Context, id string) error {
	return m.deleteWorkoutFunc(ctx, id)
}
