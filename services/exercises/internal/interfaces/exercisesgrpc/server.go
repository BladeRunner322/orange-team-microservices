// Package exercisesgrpc реализует gRPC-сервер Exercises-сервиса: обработчики RPC и мапперы proto ↔ domain.
package exercisesgrpc

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/exercises/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует exercises.ExercisesServiceServer.
// Зависимости — usecase для пяти RPC.
type Server struct {
	exercises.UnimplementedExercisesServiceServer
	createExerciseUC *usecases.CreateExercise
	getExerciseUC    *usecases.GetExercise
	getExercisesUC   *usecases.GetExercises
	patchExerciseUC  *usecases.PatchExercise
	deleteExerciseUC *usecases.DeleteExercise
}

// NewServer создаёт gRPC-сервер Exercises.
func NewServer(
	createExerciseUC *usecases.CreateExercise,
	getExerciseUC *usecases.GetExercise,
	getExercisesUC *usecases.GetExercises,
	patchExerciseUC *usecases.PatchExercise,
	deleteExerciseUC *usecases.DeleteExercise,
) *Server {
	return &Server{
		createExerciseUC: createExerciseUC,
		getExerciseUC:    getExerciseUC,
		getExercisesUC:   getExercisesUC,
		patchExerciseUC:  patchExerciseUC,
		deleteExerciseUC: deleteExerciseUC,
	}
}

// CreateExercise создаёт новое упражнение (admin-only).
// Возвращает ErrExerciseNameExists, если name уже занят.
func (s *Server) CreateExercise(
	ctx context.Context,
	req *exercises.CreateExerciseRequest,
) (*exercises.Exercise, error) {
	created, err := s.createExerciseUC.Execute(ctx, req.Name, req.Description, int(req.Difficulty), req.Type)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidName):
			return nil, status.Error(codes.InvalidArgument, "invalid name")
		case errors.Is(err, domain.ErrInvalidDescription):
			return nil, status.Error(codes.InvalidArgument, "invalid description")
		case errors.Is(err, domain.ErrInvalidDifficulty):
			return nil, status.Error(codes.InvalidArgument, "invalid difficulty")
		case errors.Is(err, domain.ErrInvalidExerciseType):
			return nil, status.Error(codes.InvalidArgument, "invalid exercise type")
		case errors.Is(err, domain.ErrExerciseNameExists):
			return nil, status.Error(codes.AlreadyExists, "exercise name already exists")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := ToProtoExercise(created)

	return resp, nil
}

// GetExercise возвращает упражнение по id (включая удалённые, см. ADR-008).
// Возвращает ErrExerciseNotFound, если не найдено.
func (s *Server) GetExercise(
	ctx context.Context,
	req *exercises.GetExerciseRequest,
) (*exercises.Exercise, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	exercise, err := s.getExerciseUC.Execute(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			return nil, status.Error(codes.NotFound, "exercise not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	resp := ToProtoExercise(exercise)

	return resp, nil
}

// GetExercises возвращает список активных упражнений (deleted_at IS NULL).
func (s *Server) GetExercises(
	ctx context.Context,
	req *exercises.GetExercisesRequest,
) (*exercises.GetExercisesResponse, error) {
	list, err := s.getExercisesUC.Execute(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	protoExercises := make([]*exercises.Exercise, 0, len(list))
	for _, ex := range list {
		protoExercises = append(protoExercises, ToProtoExercise(ex))
	}

	resp := &exercises.GetExercisesResponse{
		Exercises: protoExercises,
	}

	return resp, nil
}

// PatchExercise применяет патч к упражнению по id (admin-only).
// Поля патча: name, description, difficulty (type immutable — ADR-008).
// Возвращает ErrExerciseNotFound, если не найдено или удалено.
func (s *Server) PatchExercise(
	ctx context.Context,
	req *exercises.PatchExerciseRequest,
) (*exercises.Exercise, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	patch, err := ToDomainPatch(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidName):
			return nil, status.Error(codes.InvalidArgument, "invalid name")
		case errors.Is(err, domain.ErrInvalidDescription):
			return nil, status.Error(codes.InvalidArgument, "invalid description")
		case errors.Is(err, domain.ErrInvalidDifficulty):
			return nil, status.Error(codes.InvalidArgument, "invalid difficulty")
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid patch data")
		}
	}

	updated, err := s.patchExerciseUC.Execute(ctx, id, patch)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrExerciseNotFound):
			return nil, status.Error(codes.NotFound, "exercise not found")
		case errors.Is(err, domain.ErrExerciseNameExists):
			return nil, status.Error(codes.AlreadyExists, "exercise name already exists")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	resp := ToProtoExercise(updated)

	return resp, nil
}

// DeleteExercise помечает упражнение удалённым (soft delete, admin-only).
// Возвращает ErrExerciseNotFound, если не найдено или уже удалено.
func (s *Server) DeleteExercise(
	ctx context.Context,
	req *exercises.DeleteExerciseRequest,
) (*exercises.DeleteExerciseResponse, error) {
	id, err := ToDomainID(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.deleteExerciseUC.Execute(ctx, id); err != nil {
		if errors.Is(err, domain.ErrExerciseNotFound) {
			return nil, status.Error(codes.NotFound, "exercise not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	resp := &exercises.DeleteExerciseResponse{}

	return resp, nil
}
