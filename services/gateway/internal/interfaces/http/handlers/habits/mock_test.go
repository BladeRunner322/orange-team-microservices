package habits

import (
	"context"

	habitspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
)

// mockHabitsClient — реализация ports.HabitsClientInterface для тестов.
type mockHabitsClient struct {
	getHabitsFunc       func(ctx context.Context) (*habitspb.GetHabitsResponse, error)
	getHabitFunc        func(ctx context.Context, id string) (*habitspb.Habit, error)
	createHabitFunc     func(ctx context.Context, req *habitspb.CreateHabitRequest) (*habitspb.Habit, error)
	patchHabitFunc      func(ctx context.Context, req *habitspb.PatchHabitRequest) (*habitspb.Habit, error)
	deleteHabitFunc     func(ctx context.Context, id string) error
	completeHabitFunc   func(ctx context.Context, req *habitspb.CompleteHabitRequest) error
	uncompleteHabitFunc func(ctx context.Context, req *habitspb.UncompleteHabitRequest) error
	getCompletionsFunc  func(ctx context.Context, req *habitspb.GetCompletionsRequest) (*habitspb.GetCompletionsResponse, error)
}

func (m *mockHabitsClient) GetHabits(ctx context.Context) (*habitspb.GetHabitsResponse, error) {
	if m.getHabitsFunc != nil {
		return m.getHabitsFunc(ctx)
	}
	return &habitspb.GetHabitsResponse{Habits: []*habitspb.Habit{}}, nil
}

func (m *mockHabitsClient) GetHabit(ctx context.Context, id string) (*habitspb.Habit, error) {
	if m.getHabitFunc != nil {
		return m.getHabitFunc(ctx, id)
	}
	return &habitspb.Habit{Id: id, Name: "test"}, nil
}

func (m *mockHabitsClient) CreateHabit(ctx context.Context, req *habitspb.CreateHabitRequest) (*habitspb.Habit, error) {
	if m.createHabitFunc != nil {
		return m.createHabitFunc(ctx, req)
	}
	return &habitspb.Habit{Id: "new-id", Name: req.Name, Description: req.Description}, nil
}

func (m *mockHabitsClient) PatchHabit(ctx context.Context, req *habitspb.PatchHabitRequest) (*habitspb.Habit, error) {
	if m.patchHabitFunc != nil {
		return m.patchHabitFunc(ctx, req)
	}
	return &habitspb.Habit{Id: req.Id}, nil
}

func (m *mockHabitsClient) DeleteHabit(ctx context.Context, id string) error {
	if m.deleteHabitFunc != nil {
		return m.deleteHabitFunc(ctx, id)
	}
	return nil
}

func (m *mockHabitsClient) CompleteHabit(ctx context.Context, req *habitspb.CompleteHabitRequest) error {
	if m.completeHabitFunc != nil {
		return m.completeHabitFunc(ctx, req)
	}
	return nil
}

func (m *mockHabitsClient) UncompleteHabit(ctx context.Context, req *habitspb.UncompleteHabitRequest) error {
	if m.uncompleteHabitFunc != nil {
		return m.uncompleteHabitFunc(ctx, req)
	}
	return nil
}

func (m *mockHabitsClient) GetCompletions(ctx context.Context, req *habitspb.GetCompletionsRequest) (*habitspb.GetCompletionsResponse, error) {
	if m.getCompletionsFunc != nil {
		return m.getCompletionsFunc(ctx, req)
	}
	return &habitspb.GetCompletionsResponse{Completions: []*habitspb.Completion{}}, nil
}

func (m *mockHabitsClient) Close() {}