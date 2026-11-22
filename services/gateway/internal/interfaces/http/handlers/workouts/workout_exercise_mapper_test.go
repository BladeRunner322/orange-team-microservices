package workouts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
)

func TestProtoToWorkoutExerciseListResponse(t *testing.T) {
	createdAt := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	weight, sets, reps := int32(25), int32(3), int32(10)

	got := protoToWorkoutExerciseListResponse(&workoutspb.GetWorkoutExercisesResponse{
		WorkoutExercises: []*workoutspb.WorkoutExercise{
			{
				Id: "entry-1", Version: 2, WorkoutId: "workout-1", ExerciseId: "catalog-1",
				Difficulty: 3, Weight: &weight, Sets: &sets, Reps: &reps,
				Completed: true, ExerciseLoad: 75,
				CreatedAt: timestamppb.New(createdAt), UpdatedAt: timestamppb.New(updatedAt),
			},
		},
	})

	require.Len(t, got, 1)
	assert.Equal(t, WorkoutExerciseResponse{
		ID: "entry-1", Version: 2, WorkoutID: "workout-1", ExerciseID: "catalog-1",
		Difficulty: 3, Weight: &weight, Sets: &sets, Reps: &reps,
		Completed: true, ExerciseLoad: 75, CreatedAt: createdAt, UpdatedAt: &updatedAt,
	}, got[0])
}

func TestCreateWorkoutExerciseRequestToProto(t *testing.T) {
	weight, sets, reps := int32(25), int32(3), int32(10)
	got := createWorkoutExerciseRequestToProto("workout-1", CreateWorkoutExerciseRequest{
		ExerciseID: "catalog-1",
		Weight:     &weight,
		Sets:       &sets,
		Reps:       &reps,
		Completed:  true,
	})

	assert.Equal(t, "workout-1", got.WorkoutId)
	assert.Equal(t, "catalog-1", got.ExerciseId)
	assert.Equal(t, &weight, got.Weight)
	assert.Equal(t, &sets, got.Sets)
	assert.Equal(t, &reps, got.Reps)
	assert.Nil(t, got.Duration)
	assert.True(t, got.Completed)
}

func TestPatchWorkoutExerciseRequestToProto(t *testing.T) {
	t.Run("сохраняет пропущенные поля", func(t *testing.T) {
		got := patchWorkoutExerciseRequestToProto("workout-1", "entry-1", PatchWorkoutExerciseRequest{})

		assert.Equal(t, "workout-1", got.WorkoutId)
		assert.Equal(t, "entry-1", got.WorkoutExerciseId)
		assert.Nil(t, got.Completed)
		assert.Nil(t, got.Duration)
	})

	t.Run("передаёт заданные значения", func(t *testing.T) {
		completed, duration := false, int32(90)
		got := patchWorkoutExerciseRequestToProto("workout-1", "entry-1", PatchWorkoutExerciseRequest{
			Duration:  &duration,
			Completed: &completed,
		})

		assert.Equal(t, &duration, got.Duration)
		require.NotNil(t, got.Completed)
		assert.False(t, *got.Completed)
	})
}
