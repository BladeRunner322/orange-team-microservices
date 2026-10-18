package workoutsgrpc

import (
	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

func ToDomainID(raw string) (uuid.UUID, error) {
	// TODO: uuid.Parse + wrap err
	return uuid.Nil, nil
}

func ToProtoWorkout(w domain.Workout) *workouts.Workout {
	// TODO: заполнить поля, включая *timestamppb.Timestamp для started_at/completed_at/created_at/updated_at
	return nil
}

func ToProtoWorkouts(ws []domain.Workout) []*workouts.Workout {
	// TODO
	return nil
}

func ToProtoWorkoutExercise(e domain.WorkoutExercise) *workouts.WorkoutExercise {
	// TODO
	return nil
}

func ToProtoWorkoutExercises(es []domain.WorkoutExercise) []*workouts.WorkoutExercise {
	// TODO
	return nil
}

// ToDomainWorkoutPatch конвертирует PatchWorkoutRequest в domain.WorkoutPatch.
func ToDomainWorkoutPatch(req *workouts.PatchWorkoutRequest) (domain.WorkoutPatch, error) {
	// TODO: optional status/intensity → nullable.Nullable
	return domain.WorkoutPatch{}, nil
}

// ToDomainWorkoutExercisePatch конвертирует PatchWorkoutExerciseRequest в domain.WorkoutExercisePatch.
func ToDomainWorkoutExercisePatch(req *workouts.PatchWorkoutExerciseRequest) (domain.WorkoutExercisePatch, error) {
	// TODO: optional weight/sets/reps/duration/completed → nullable.Nullable
	return domain.WorkoutExercisePatch{}, nil
}
