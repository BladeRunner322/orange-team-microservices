package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
)

// ExercisesClientInterface — контракт gRPC-клиента к Exercises Service.
type ExercisesClientInterface interface {
	CreateExercise(ctx context.Context, req *exercises.CreateExerciseRequest) (*exercises.Exercise, error)
	GetExercise(ctx context.Context, id string) (*exercises.Exercise, error)
	GetExercises(ctx context.Context) (*exercises.GetExercisesResponse, error)
	PatchExercise(ctx context.Context, req *exercises.PatchExerciseRequest) (*exercises.Exercise, error)
	DeleteExercise(ctx context.Context, id string) error
	Close()
}
