// Package workoutsgrpc реализует gRPC-сервер Workouts-сервиса.
package workoutsgrpc

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/usecases"
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

func (s *Server) CreateWorkout(ctx context.Context, req *workouts.CreateWorkoutRequest) (*workouts.Workout, error) {
	// TODO: uc.Execute + маппинг ошибок в codes + ToProtoWorkout
	return nil, nil
}

func (s *Server) GetWorkout(ctx context.Context, req *workouts.GetWorkoutRequest) (*workouts.Workout, error) {
	return nil, nil
}

func (s *Server) GetWorkouts(ctx context.Context, req *workouts.GetWorkoutsRequest) (*workouts.GetWorkoutsResponse, error) {
	return nil, nil
}

func (s *Server) PatchWorkout(ctx context.Context, req *workouts.PatchWorkoutRequest) (*workouts.Workout, error) {
	return nil, nil
}

func (s *Server) DeleteWorkout(ctx context.Context, req *workouts.DeleteWorkoutRequest) (*workouts.DeleteWorkoutResponse, error) {
	return nil, nil
}

func (s *Server) CreateWorkoutExercise(ctx context.Context, req *workouts.CreateWorkoutExerciseRequest) (*workouts.WorkoutExercise, error) {
	return nil, nil
}

func (s *Server) GetWorkoutExercises(ctx context.Context, req *workouts.GetWorkoutExercisesRequest) (*workouts.GetWorkoutExercisesResponse, error) {
	return nil, nil
}

func (s *Server) PatchWorkoutExercise(ctx context.Context, req *workouts.PatchWorkoutExerciseRequest) (*workouts.WorkoutExercise, error) {
	return nil, nil
}

func (s *Server) DeleteWorkoutExercise(ctx context.Context, req *workouts.DeleteWorkoutExerciseRequest) (*workouts.DeleteWorkoutExerciseResponse, error) {
	return nil, nil
}

func (s *Server) GetUserScore(ctx context.Context, req *workouts.GetUserScoreRequest) (*workouts.GetUserScoreResponse, error) {
	return nil, nil
}
