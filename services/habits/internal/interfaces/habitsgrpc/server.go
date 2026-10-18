// Package habitsgrpc реализует gRPC-сервер Habits-сервиса.
package habitsgrpc

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/habits/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	return &Server{
		createHabitUC:     createHabitUC,
		getHabitUC:        getHabitUC,
		getHabitsUC:       getHabitsUC,
		patchHabitUC:      patchHabitUC,
		deleteHabitUC:     deleteHabitUC,
		completeHabitUC:   completeHabitUC,
		uncompleteHabitUC: uncompleteHabitUC,
		getCompletionsUC:  getCompletionsUC,
	}
}

// CreateHabit создаёт привычку.
func (s *Server) CreateHabit(
	ctx context.Context,
	req *habits.CreateHabitRequest,
) (*habits.Habit, error) {
	created, err := s.createHabitUC.Execute(ctx, req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrInvalidName):
			return nil, status.Error(codes.InvalidArgument, "invalid name")
		case errors.Is(err, domain.ErrInvalidDescription):
			return nil, status.Error(codes.InvalidArgument, "invalid description")
		case errors.Is(err, domain.ErrHabitNameExists):
			return nil, status.Error(codes.AlreadyExists, "habit with this name already exists")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := ToProtoHabit(created, 0, false)

	return resp, nil
}

// GetHabit возвращает привычку по id со стриком.
func (s *Server) GetHabit(
	ctx context.Context,
	req *habits.GetHabitRequest,
) (*habits.Habit, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	hws, err := s.getHabitUC.Execute(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrHabitNotFound):
			return nil, status.Error(codes.NotFound, "habit not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := ToProtoHabitWithStreak(hws)

	return resp, nil
}

// GetHabits возвращает список привычек пользователя со стриками.
func (s *Server) GetHabits(
	ctx context.Context,
	req *habits.GetHabitsRequest,
) (*habits.GetHabitsResponse, error) {
	list, err := s.getHabitsUC.Execute(ctx)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	protoHabits := make([]*habits.Habit, 0, len(list))
	for _, hws := range list {
		protoHabits = append(protoHabits, ToProtoHabitWithStreak(hws))
	}

	resp := &habits.GetHabitsResponse{
		Habits: protoHabits,
	}

	return resp, nil
}

// PatchHabit применяет патч к привычке.
func (s *Server) PatchHabit(
	ctx context.Context,
	req *habits.PatchHabitRequest,
) (*habits.Habit, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	patch, err := ToDomainHabitPatch(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidName):
			return nil, status.Error(codes.InvalidArgument, "invalid name")
		case errors.Is(err, domain.ErrInvalidDescription):
			return nil, status.Error(codes.InvalidArgument, "invalid description")
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid patch data")
		}
	}

	updated, err := s.patchHabitUC.Execute(ctx, id, patch)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrHabitNotFound):
			return nil, status.Error(codes.NotFound, "habit not found")
		case errors.Is(err, domain.ErrHabitNameExists):
			return nil, status.Error(codes.AlreadyExists, "habit with this name already exists")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := ToProtoHabit(updated, 0, false)

	return resp, nil
}

// DeleteHabit помечает привычку удалённой (soft delete). Идемпотентно.
func (s *Server) DeleteHabit(
	ctx context.Context,
	req *habits.DeleteHabitRequest,
) (*habits.DeleteHabitResponse, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.deleteHabitUC.Execute(ctx, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrHabitNotFound):
			// Чужая привычка — не раскрываем существование.
			return nil, status.Error(codes.NotFound, "habit not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := &habits.DeleteHabitResponse{}

	return resp, nil
}

// CompleteHabit отмечает выполнение привычки на дату.
func (s *Server) CompleteHabit(
	ctx context.Context,
	req *habits.CompleteHabitRequest,
) (*habits.CompleteHabitResponse, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	date, err := ToDomainDate(req.Date)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	if err := s.completeHabitUC.Execute(ctx, id, date); err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrHabitNotFound):
			return nil, status.Error(codes.NotFound, "habit not found")
		case errors.Is(err, domain.ErrInvalidDate):
			return nil, status.Error(codes.InvalidArgument, "invalid date: must be today or up to 7 days ago")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := &habits.CompleteHabitResponse{}

	return resp, nil
}

// UncompleteHabit снимает отметку выполнения.
func (s *Server) UncompleteHabit(
	ctx context.Context,
	req *habits.UncompleteHabitRequest,
) (*habits.UncompleteHabitResponse, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	date, err := ToDomainDate(req.Date)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	if err := s.uncompleteHabitUC.Execute(ctx, id, date); err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrHabitNotFound):
			return nil, status.Error(codes.NotFound, "habit not found")
		case errors.Is(err, domain.ErrInvalidDate):
			return nil, status.Error(codes.InvalidArgument, "invalid date: must be today or up to 7 days ago")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := &habits.UncompleteHabitResponse{}

	return resp, nil
}

// GetCompletions возвращает историю выполнений привычки за период.
func (s *Server) GetCompletions(
	ctx context.Context,
	req *habits.GetCompletionsRequest,
) (*habits.GetCompletionsResponse, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	from, err := ToDomainDateOrZero(req.From)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid from date")
	}

	to, err := ToDomainDateOrZero(req.To)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid to date")
	}

	completions, err := s.getCompletionsUC.Execute(ctx, id, from, to)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		case errors.Is(err, domain.ErrHabitNotFound):
			return nil, status.Error(codes.NotFound, "habit not found")
		case errors.Is(err, domain.ErrInvalidDate):
			return nil, status.Error(codes.InvalidArgument, "invalid period: from must be <= to")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	protoCompletions := make([]*habits.Completion, 0, len(completions))
	for _, c := range completions {
		protoCompletions = append(protoCompletions, ToProtoCompletion(c))
	}

	resp := &habits.GetCompletionsResponse{
		Completions: protoCompletions,
	}

	return resp, nil
}
