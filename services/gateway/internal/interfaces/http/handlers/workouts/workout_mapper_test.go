package workouts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	workoutspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/workouts"
)

func TestProtoToWorkoutResponse(t *testing.T) {
	t.Run("заполненные поля", func(t *testing.T) {
		createdAt := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
		startedAt := createdAt.Add(time.Minute)
		completedAt := startedAt.Add(30 * time.Minute)
		updatedAt := completedAt.Add(time.Minute)
		intensity := int32(7)

		got := protoToWorkoutResponse(&workoutspb.Workout{
			Id:                       "workout-1",
			Version:                  3,
			UserId:                   "user-1",
			Status:                   "completed",
			StartedAt:                timestamppb.New(startedAt),
			CompletedAt:              timestamppb.New(completedAt),
			WorkoutScore:             42,
			Intensity:                &intensity,
			PersonalScoreCoefficient: 2,
			CreatedAt:                timestamppb.New(createdAt),
			UpdatedAt:                timestamppb.New(updatedAt),
			IsDeleted:                true,
		})

		assert.Equal(t, WorkoutResponse{
			ID:                       "workout-1",
			Version:                  3,
			UserID:                   "user-1",
			Status:                   "completed",
			StartedAt:                &startedAt,
			CompletedAt:              &completedAt,
			WorkoutScore:             42,
			Intensity:                &intensity,
			PersonalScoreCoefficient: 2,
			CreatedAt:                createdAt,
			UpdatedAt:                &updatedAt,
			IsDeleted:                true,
		}, got)
	})

	t.Run("необязательные поля отсутствуют", func(t *testing.T) {
		got := protoToWorkoutResponse(&workoutspb.Workout{
			Id:     "workout-1",
			Status: "planned",
		})

		assert.Nil(t, got.StartedAt)
		assert.Nil(t, got.CompletedAt)
		assert.Nil(t, got.UpdatedAt)
		assert.Nil(t, got.Intensity)
	})
}

func TestPatchRequestToProto(t *testing.T) {
	t.Run("передаёт идентификатор и статус", func(t *testing.T) {
		status := "in_progress"
		got := patchRequestToProto("workout-1", PatchWorkoutRequest{
			Status: &status,
		})

		assert.Equal(t, "workout-1", got.Id)
		require.NotNil(t, got.Status)
		assert.Equal(t, status, *got.Status)
		assert.Nil(t, got.Intensity)
	})

	t.Run("сохраняет отсутствие статуса", func(t *testing.T) {
		got := patchRequestToProto("workout-1", PatchWorkoutRequest{})

		assert.Equal(t, "workout-1", got.Id)
		assert.Nil(t, got.Status)
	})
}
