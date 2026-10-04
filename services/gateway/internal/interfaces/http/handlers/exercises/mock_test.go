package exercises

import (
	"context"

	exercisespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
)

// mockExercisesClient — реализация ports.ExercisesClientInterface для тестов.
type mockExercisesClient struct {
	getExercisesFunc   func(ctx context.Context) (*exercisespb.GetExercisesResponse, error)
	getExerciseFunc    func(ctx context.Context, id string) (*exercisespb.Exercise, error)
	createExerciseFunc func(ctx context.Context, req *exercisespb.CreateExerciseRequest) (*exercisespb.Exercise, error)
	patchExerciseFunc  func(ctx context.Context, req *exercisespb.PatchExerciseRequest) (*exercisespb.Exercise, error)
	deleteExerciseFunc func(ctx context.Context, id string) error
}

func (m *mockExercisesClient) GetExercises(ctx context.Context) (*exercisespb.GetExercisesResponse, error) {
	if m.getExercisesFunc != nil {
		return m.getExercisesFunc(ctx)
	}
	return &exercisespb.GetExercisesResponse{Exercises: []*exercisespb.Exercise{}}, nil
}

func (m *mockExercisesClient) GetExercise(ctx context.Context, id string) (*exercisespb.Exercise, error) {
	if m.getExerciseFunc != nil {
		return m.getExerciseFunc(ctx, id)
	}
	return &exercisespb.Exercise{Id: id, Name: "test", Description: "test", Difficulty: 5, Type: "weight"}, nil
}

func (m *mockExercisesClient) CreateExercise(ctx context.Context, req *exercisespb.CreateExerciseRequest) (*exercisespb.Exercise, error) {
	if m.createExerciseFunc != nil {
		return m.createExerciseFunc(ctx, req)
	}
	return &exercisespb.Exercise{Id: "test-id", Name: req.Name, Description: req.Description, Difficulty: req.Difficulty, Type: req.Type}, nil
}

func (m *mockExercisesClient) PatchExercise(ctx context.Context, req *exercisespb.PatchExerciseRequest) (*exercisespb.Exercise, error) {
	if m.patchExerciseFunc != nil {
		return m.patchExerciseFunc(ctx, req)
	}
	return &exercisespb.Exercise{Id: req.Id}, nil
}

func (m *mockExercisesClient) DeleteExercise(ctx context.Context, id string) error {
	if m.deleteExerciseFunc != nil {
		return m.deleteExerciseFunc(ctx, id)
	}
	return nil
}

func (m *mockExercisesClient) Close() {}
