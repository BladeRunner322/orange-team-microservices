package ports

import (
	"context"

	"github.com/google/uuid"
)

// UserProfile — данные профиля, которые Workouts получает от Profiles.
type UserProfile struct {
	UserID      uuid.UUID
	Sex         *string
	WeightGrams *int
	HeightCM    *int
}

// ExerciseInfo — данные упражнения, которые Workouts получает от Exercises.
type ExerciseInfo struct {
	ID         uuid.UUID
	Type       string
	Difficulty int
}

// ProfilesClient — контракт gRPC-клиента к Profiles.
type ProfilesClient interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (UserProfile, error)
	IsHealthy(ctx context.Context) error
	Close()
}

// ExercisesClient — контракт gRPC-клиента к Exercises.
type ExercisesClient interface {
	GetExercise(ctx context.Context, id uuid.UUID) (ExerciseInfo, error)
	IsHealthy(ctx context.Context) error
	Close()
}
