// Package workoutsgrpc реализует gRPC-сервер Workouts-сервиса.
package workoutsgrpc

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/usecases"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	workouts.UnimplementedWorkoutsServiceServer

	createWorkoutUC *usecases.CreateWorkout
	getWorkoutUC    *usecases.GetWorkout
	getWorkoutsUC   *usecases.GetWorkouts
	patchWorkoutUC  *usecases.PatchWorkout
	deleteWorkoutUC *usecases.DeleteWorkout

	createWorkoutExerciseUC *usecases.CreateWorkoutExercise
	getWorkoutExercisesUC   *usecases.GetWorkoutExercises
	patchWorkoutExerciseUC  *usecases.PatchWorkoutExercise
	deleteWorkoutExerciseUC *usecases.DeleteWorkoutExercise

	getUserScoreUC *usecases.GetUserScore
}

func NewServer(
	createWorkoutUC *usecases.CreateWorkout,
	getWorkoutUC *usecases.GetWorkout,
	getWorkoutsUC *usecases.GetWorkouts,
	patchWorkoutUC *usecases.PatchWorkout,
	deleteWorkoutUC *usecases.DeleteWorkout,
	createWorkoutExerciseUC *usecases.CreateWorkoutExercise,
	getWorkoutExercisesUC *usecases.GetWorkoutExercises,
	patchWorkoutExerciseUC *usecases.PatchWorkoutExercise,
	deleteWorkoutExerciseUC *usecases.DeleteWorkoutExercise,
	getUserScoreUC *usecases.GetUserScore,
) *Server {
	return &Server{
		createWorkoutUC:         createWorkoutUC,
		getWorkoutUC:            getWorkoutUC,
		getWorkoutsUC:           getWorkoutsUC,
		patchWorkoutUC:          patchWorkoutUC,
		deleteWorkoutUC:         deleteWorkoutUC,
		createWorkoutExerciseUC: createWorkoutExerciseUC,
		getWorkoutExercisesUC:   getWorkoutExercisesUC,
		patchWorkoutExerciseUC:  patchWorkoutExerciseUC,
		deleteWorkoutExerciseUC: deleteWorkoutExerciseUC,
		getUserScoreUC:          getUserScoreUC,
	}
}

// mapError конвертирует доменные ошибки в gRPC-коды.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "unauthenticated")

	case errors.Is(err, domain.ErrWorkoutNotFound):
		return status.Error(codes.NotFound, "workout not found")

	case errors.Is(err, domain.ErrWorkoutExerciseNotFound):
		return status.Error(codes.NotFound, "workout exercise not found")

	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.Aborted, "conflict: concurrent update")

	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return status.Error(codes.FailedPrecondition, "invalid status transition")

	case errors.Is(err, domain.ErrCannotCompleteEmptyWorkout):
		return status.Error(codes.FailedPrecondition, "cannot complete workout with no completed exercises")

	case errors.Is(err, domain.ErrCannotModifyWorkoutExercise):
		return status.Error(codes.FailedPrecondition, "cannot modify workout exercise in terminal status")

	case errors.Is(err, domain.ErrCannotMarkExerciseCompleted):
		return status.Error(codes.FailedPrecondition, "can only mark exercise as completed when workout is in progress")

	case errors.Is(err, domain.ErrInvalidWorkoutExerciseForType):
		return status.Error(codes.InvalidArgument, "workout exercise fields do not match exercise type")

	// VO-ошибки — все InvalidArgument
	case errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidWeight),
		errors.Is(err, domain.ErrInvalidSets),
		errors.Is(err, domain.ErrInvalidReps),
		errors.Is(err, domain.ErrInvalidDuration),
		errors.Is(err, domain.ErrInvalidDifficulty),
		errors.Is(err, domain.ErrInvalidIntensity),
		errors.Is(err, domain.ErrInvalidExerciseType):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// CreateWorkout создаёт новую тренировку.
func (s *Server) CreateWorkout(ctx context.Context, req *workouts.CreateWorkoutRequest) (*workouts.Workout, error) {
	created, err := s.createWorkoutUC.Execute(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	resp := protoWorkoutFromDomain(created)

	return resp, nil
}

// GetWorkout возвращает тренировку по id.
func (s *Server) GetWorkout(ctx context.Context, req *workouts.GetWorkoutRequest) (*workouts.Workout, error) {
	id, err := domainUUIDFromString(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	workout, err := s.getWorkoutUC.Execute(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}

	resp := protoWorkoutFromDomain(workout)

	return resp, nil
}

// GetWorkouts возвращает список тренировок пользователя.
func (s *Server) GetWorkouts(ctx context.Context, req *workouts.GetWorkoutsRequest) (*workouts.GetWorkoutsResponse, error) {
	list, err := s.getWorkoutsUC.Execute(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &workouts.GetWorkoutsResponse{
		Workouts: protoWorkoutsFromDomain(list),
	}

	return resp, nil
}

// PatchWorkout применяет патч к тренировке.
func (s *Server) PatchWorkout(ctx context.Context, req *workouts.PatchWorkoutRequest) (*workouts.Workout, error) {
	id, err := domainUUIDFromString(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	patch, err := domainWorkoutPatchFromProto(req)
	if err != nil {
		return nil, mapError(err)
	}

	updated, err := s.patchWorkoutUC.Execute(ctx, id, patch)
	if err != nil {
		return nil, mapError(err)
	}

	resp := protoWorkoutFromDomain(updated)

	return resp, nil
}

// DeleteWorkout помечает тренировку удалённой.
func (s *Server) DeleteWorkout(ctx context.Context, req *workouts.DeleteWorkoutRequest) (*workouts.DeleteWorkoutResponse, error) {
	id, err := domainUUIDFromString(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	if err := s.deleteWorkoutUC.Execute(ctx, id); err != nil {
		return nil, mapError(err)
	}

	resp := &workouts.DeleteWorkoutResponse{}

	return resp, nil
}

// CreateWorkoutExercise добавляет упражнение в тренировку.
func (s *Server) CreateWorkoutExercise(ctx context.Context, req *workouts.CreateWorkoutExerciseRequest) (*workouts.WorkoutExercise, error) {
	workoutID, err := domainUUIDFromString(req.WorkoutId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	exerciseID, err := domainUUIDFromString(req.ExerciseId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid exercise id")
	}

	weight, err := domainWeightFromProto(req.Weight)
	if err != nil {
		return nil, mapError(err)
	}

	sets, err := domainSetsFromProto(req.Sets)
	if err != nil {
		return nil, mapError(err)
	}

	reps, err := domainRepsFromProto(req.Reps)
	if err != nil {
		return nil, mapError(err)
	}

	duration, err := domainDurationFromProto(req.Duration)
	if err != nil {
		return nil, mapError(err)
	}

	created, err := s.createWorkoutExerciseUC.Execute(
		ctx,
		workoutID,
		exerciseID,
		weight,
		sets,
		reps,
		duration,
		req.Completed,
	)
	if err != nil {
		return nil, mapError(err)
	}

	resp := protoWorkoutExerciseFromDomain(created)

	return resp, nil
}

// GetWorkoutExercises возвращает список упражнений в тренировке.
func (s *Server) GetWorkoutExercises(ctx context.Context, req *workouts.GetWorkoutExercisesRequest) (*workouts.GetWorkoutExercisesResponse, error) {
	workoutID, err := domainUUIDFromString(req.WorkoutId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	list, err := s.getWorkoutExercisesUC.Execute(ctx, workoutID)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &workouts.GetWorkoutExercisesResponse{
		WorkoutExercises: protoWorkoutExercisesFromDomain(list),
	}

	return resp, nil
}

// PatchWorkoutExercise применяет патч к упражнению в тренировке.
func (s *Server) PatchWorkoutExercise(ctx context.Context, req *workouts.PatchWorkoutExerciseRequest) (*workouts.WorkoutExercise, error) {
	workoutID, err := domainUUIDFromString(req.WorkoutId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	exerciseID, err := domainUUIDFromString(req.WorkoutExerciseId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout exercise id")
	}

	patch, err := domainWorkoutExercisePatchFromProto(req)
	if err != nil {
		return nil, mapError(err)
	}

	updated, err := s.patchWorkoutExerciseUC.Execute(ctx, workoutID, exerciseID, patch)
	if err != nil {
		return nil, mapError(err)
	}

	resp := protoWorkoutExerciseFromDomain(updated)

	return resp, nil
}

// DeleteWorkoutExercise удаляет упражнение из тренировки.
func (s *Server) DeleteWorkoutExercise(ctx context.Context, req *workouts.DeleteWorkoutExerciseRequest) (*workouts.DeleteWorkoutExerciseResponse, error) {
	workoutID, err := domainUUIDFromString(req.WorkoutId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout id")
	}

	exerciseID, err := domainUUIDFromString(req.WorkoutExerciseId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workout exercise id")
	}

	if err := s.deleteWorkoutExerciseUC.Execute(ctx, workoutID, exerciseID); err != nil {
		return nil, mapError(err)
	}

	resp := &workouts.DeleteWorkoutExerciseResponse{}

	return resp, nil
}

// GetUserScore возвращает сумму очков пользователя.
func (s *Server) GetUserScore(ctx context.Context, req *workouts.GetUserScoreRequest) (*workouts.GetUserScoreResponse, error) {
	score, err := s.getUserScoreUC.Execute(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &workouts.GetUserScoreResponse{
		Score: score,
	}

	return resp, nil
}
