package exercises

import (
	"testing"
	"time"

	exercisespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/exercises"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ===== createRequestToProto =====
func TestCreateRequestToProto(t *testing.T) {
	req := CreateExerciseRequest{
		Name:        "Жим лёжа",
		Description: "Базовое",
		Difficulty:  5,
		Type:        "weight",
	}

	out := createRequestToProto(req)

	assert.Equal(t, "Жим лёжа", out.Name)
	assert.Equal(t, "Базовое", out.Description)
	assert.Equal(t, int32(5), out.Difficulty)
	assert.Equal(t, "weight", out.Type)
}

// ===== patchRequestToProto =====
func TestPatchRequestToProto(t *testing.T) {
	t.Run("пустой патч — все поля nil", func(t *testing.T) {
		out := patchRequestToProto("abc-123", PatchExerciseRequest{})

		assert.Equal(t, "abc-123", out.Id)
		assert.Nil(t, out.Name)
		assert.Nil(t, out.Description)
		assert.Nil(t, out.Difficulty)
	})

	t.Run("все поля заданы", func(t *testing.T) {
		name := "Новое имя"
		desc := "Новое описание"
		diff := int32(9)

		out := patchRequestToProto("abc-123", PatchExerciseRequest{
			Name:        &name,
			Description: &desc,
			Difficulty:  &diff,
		})

		assert.Equal(t, "abc-123", out.Id)
		require.NotNil(t, out.Name)
		assert.Equal(t, "Новое имя", *out.Name)
		require.NotNil(t, out.Description)
		assert.Equal(t, "Новое описание", *out.Description)
		require.NotNil(t, out.Difficulty)
		assert.Equal(t, int32(9), *out.Difficulty)
	})

	t.Run("частичный патч — только difficulty", func(t *testing.T) {
		diff := int32(7)

		out := patchRequestToProto("abc-123", PatchExerciseRequest{
			Difficulty: &diff,
		})

		assert.Equal(t, "abc-123", out.Id)
		assert.Nil(t, out.Name)
		assert.Nil(t, out.Description)
		require.NotNil(t, out.Difficulty)
		assert.Equal(t, int32(7), *out.Difficulty)
	})
}

// ===== protoToExerciseResponse =====
func TestProtoToExerciseResponse(t *testing.T) {
	t.Run("полное упражнение", func(t *testing.T) {
		created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		updated := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

		out := protoToExerciseResponse(&exercisespb.Exercise{
			Id:          "abc-123",
			Name:        "Жим лёжа",
			Description: "Базовое упражнение",
			Difficulty:  5,
			Type:        "weight",
			CreatedAt:   timestamppb.New(created),
			UpdatedAt:   timestamppb.New(updated),
			IsDeleted:   false,
		})

		assert.Equal(t, "abc-123", out.ID)
		assert.Equal(t, "Жим лёжа", out.Name)
		assert.Equal(t, "Базовое упражнение", out.Description)
		assert.Equal(t, int32(5), out.Difficulty)
		assert.Equal(t, "weight", out.Type)
		assert.False(t, out.IsDeleted)
		assert.Equal(t, created, out.CreatedAt)
		require.NotNil(t, out.UpdatedAt)
		assert.Equal(t, updated, *out.UpdatedAt)
	})

	t.Run("updated_at nil", func(t *testing.T) {
		out := protoToExerciseResponse(&exercisespb.Exercise{
			Id:        "abc-123",
			CreatedAt: timestamppb.New(time.Now()),
			UpdatedAt: nil,
		})

		assert.Nil(t, out.UpdatedAt)
	})

	t.Run("удалённое упражнение", func(t *testing.T) {
		out := protoToExerciseResponse(&exercisespb.Exercise{
			Id:        "abc-123",
			IsDeleted: true,
		})

		assert.True(t, out.IsDeleted)
	})
}

// ===== protoToExerciseListResponse =====
func TestProtoToExerciseListResponse(t *testing.T) {
	t.Run("список из двух", func(t *testing.T) {
		out := protoToExerciseListResponse(&exercisespb.GetExercisesResponse{
			Exercises: []*exercisespb.Exercise{
				{Id: "id-1", Name: "Жим лёжа"},
				{Id: "id-2", Name: "Планка"},
			},
		})

		require.Len(t, out, 2)
		assert.Equal(t, "id-1", out[0].ID)
		assert.Equal(t, "id-2", out[1].ID)
	})

	t.Run("пустой список — make, не nil", func(t *testing.T) {
		out := protoToExerciseListResponse(&exercisespb.GetExercisesResponse{
			Exercises: []*exercisespb.Exercise{},
		})

		assert.NotNil(t, out, "должен быть []ExerciseResponse{}, а не nil")
		assert.Empty(t, out)
	})
}
