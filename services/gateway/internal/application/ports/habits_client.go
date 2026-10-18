package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
)

// HabitsClientInterface — контракт gRPC-клиента к Habits Service.
type HabitsClientInterface interface {
	CreateHabit(ctx context.Context, req *habits.CreateHabitRequest) (*habits.Habit, error)
	GetHabit(ctx context.Context, id string) (*habits.Habit, error)
	GetHabits(ctx context.Context) (*habits.GetHabitsResponse, error)
	PatchHabit(ctx context.Context, req *habits.PatchHabitRequest) (*habits.Habit, error)
	DeleteHabit(ctx context.Context, id string) error
	CompleteHabit(ctx context.Context, req *habits.CompleteHabitRequest) error
	UncompleteHabit(ctx context.Context, req *habits.UncompleteHabitRequest) error
	GetCompletions(ctx context.Context, req *habits.GetCompletionsRequest) (*habits.GetCompletionsResponse, error)
	Close()
}
