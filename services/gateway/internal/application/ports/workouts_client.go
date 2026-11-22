package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
)

// WorkoutsClientInterface — контракт gRPC-клиента к сервису Workouts.
type WorkoutsClientInterface interface {
	CreateWorkout(ctx context.Context) (*workouts.Workout, error)
	GetWorkout(ctx context.Context, id string) (*workouts.Workout, error)
	GetWorkouts(ctx context.Context) (*workouts.GetWorkoutsResponse, error)
	PatchWorkout(ctx context.Context, req *workouts.PatchWorkoutRequest) (*workouts.Workout, error)
	DeleteWorkout(ctx context.Context, id string) error

	CreateWorkoutExercise(ctx context.Context, req *workouts.CreateWorkoutExerciseRequest) (*workouts.WorkoutExercise, error)
	GetWorkoutExercises(ctx context.Context, workoutID string) (*workouts.GetWorkoutExercisesResponse, error)
	PatchWorkoutExercise(ctx context.Context, req *workouts.PatchWorkoutExerciseRequest) (*workouts.WorkoutExercise, error)
	DeleteWorkoutExercise(ctx context.Context, req *workouts.DeleteWorkoutExerciseRequest) error

	GetUserScore(ctx context.Context) (*workouts.GetUserScoreResponse, error)
	Close()
}
