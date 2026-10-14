// Package habitsgrpc реализует gRPC-сервер Habits-сервиса.
package habitsgrpc

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/usecases"
)

// Server реализует habits.HabitsServiceServer.
type Server struct {
	habits.UnimplementedHabitsServiceServer
	createHabitUC     *usecases.CreateHabit
	getHabitUC        *usecases.GetHabit
	getHabitsUC       *usecases.GetHabits
	patchHabitUC      *usecases.PatchHabit
	deleteHabitUC     *usecases.DeleteHabit
	completeHabitUC   *usecases.CompleteHabit
	uncompleteHabitUC *usecases.UncompleteHabit
	getCompletionsUC  *usecases.GetCompletions
}

func NewServer(
	createHabitUC *usecases.CreateHabit,
	getHabitUC *usecases.GetHabit,
	getHabitsUC *usecases.GetHabits,
	patchHabitUC *usecases.PatchHabit,
	deleteHabitUC *usecases.DeleteHabit,
	completeHabitUC *usecases.CompleteHabit,
	uncompleteHabitUC *usecases.UncompleteHabit,
	getCompletionsUC *usecases.GetCompletions,
) *Server {
	// TODO
	return nil
}

// CreateHabit создаёт привычку. Маппит domain.ErrInvalidName/ErrInvalidDescription
// в codes.InvalidArgument, ErrUnauthenticated в codes.Unauthenticated.
func (s *Server) CreateHabit(ctx context.Context, req *habits.CreateHabitRequest) (*habits.Habit, error) {
	// TODO
	return nil, nil
}

// GetHabit возвращает привычку по id со стриком.
// Маппит ErrHabitNotFound в codes.NotFound.
func (s *Server) GetHabit(ctx context.Context, req *habits.GetHabitRequest) (*habits.Habit, error) {
	// TODO
	return nil, nil
}

// GetHabits возвращает список привычек пользователя со стриками.
func (s *Server) GetHabits(ctx context.Context, req *habits.GetHabitsRequest) (*habits.GetHabitsResponse, error) {
	// TODO
	return nil, nil
}

// PatchHabit применяет патч. type immutable нет — тут только name/description.
func (s *Server) PatchHabit(ctx context.Context, req *habits.PatchHabitRequest) (*habits.Habit, error) {
	// TODO
	return nil, nil
}

// DeleteHabit помечает привычку удалённой. Идемпотентно.
func (s *Server) DeleteHabit(ctx context.Context, req *habits.DeleteHabitRequest) (*habits.DeleteHabitResponse, error) {
	// TODO
	return nil, nil
}

// CompleteHabit отмечает выполнение. Маппит ErrInvalidDate в codes.InvalidArgument.
func (s *Server) CompleteHabit(ctx context.Context, req *habits.CompleteHabitRequest) (*habits.CompleteHabitResponse, error) {
	// TODO
	return nil, nil
}

// UncompleteHabit снимает отметку.
func (s *Server) UncompleteHabit(ctx context.Context, req *habits.UncompleteHabitRequest) (*habits.UncompleteHabitResponse, error) {
	// TODO
	return nil, nil
}

// GetCompletions возвращает историю за период.
func (s *Server) GetCompletions(ctx context.Context, req *habits.GetCompletionsRequest) (*habits.GetCompletionsResponse, error) {
	// TODO
	return nil, nil
}