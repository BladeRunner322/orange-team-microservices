package habits

import (
	"testing"
	"time"

	habitspb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/habits"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ===== protoToHabitResponse =====
func TestProtoToHabitResponse(t *testing.T) {
	t.Run("полная привычка", func(t *testing.T) {
		created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		updated := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

		out := protoToHabitResponse(&habitspb.Habit{
			Id:             "abc",
			UserId:         "user-1",
			Name:           "Пить воду",
			Description:    "2 литра",
			Streak:         5,
			CompletedToday: true,
			IsDeleted:      false,
			CreatedAt:      timestamppb.New(created),
			UpdatedAt:      timestamppb.New(updated),
		})

		assert.Equal(t, "abc", out.ID)
		assert.Equal(t, "user-1", out.UserID)
		assert.Equal(t, "Пить воду", out.Name)
		assert.Equal(t, "2 литра", out.Description)
		assert.Equal(t, 5, out.Streak)
		assert.True(t, out.CompletedToday)
		assert.False(t, out.IsDeleted)
		assert.Equal(t, created, out.CreatedAt)
		require.NotNil(t, out.UpdatedAt)
		assert.Equal(t, updated, *out.UpdatedAt)
	})

	t.Run("updated_at nil", func(t *testing.T) {
		out := protoToHabitResponse(&habitspb.Habit{
			Id:        "abc",
			CreatedAt: timestamppb.New(time.Now()),
		})

		assert.Nil(t, out.UpdatedAt)
	})
}

// ===== protoToHabitListResponse =====
func TestProtoToHabitListResponse(t *testing.T) {
	t.Run("список из двух", func(t *testing.T) {
		out := protoToHabitListResponse(&habitspb.GetHabitsResponse{
			Habits: []*habitspb.Habit{
				{Id: "id-1", Name: "Первая"},
				{Id: "id-2", Name: "Вторая"},
			},
		})

		require.Len(t, out, 2)
		assert.Equal(t, "id-1", out[0].ID)
	})

	t.Run("пустой — make, не nil", func(t *testing.T) {
		out := protoToHabitListResponse(&habitspb.GetHabitsResponse{
			Habits: []*habitspb.Habit{},
		})

		assert.NotNil(t, out)
		assert.Empty(t, out)
	})
}

// ===== protoToCompletionListResponse =====
func TestProtoToCompletionListResponse(t *testing.T) {
	t.Run("список из одной", func(t *testing.T) {
		created := time.Date(2026, 10, 14, 8, 30, 0, 0, time.UTC)

		out := protoToCompletionListResponse(&habitspb.GetCompletionsResponse{
			Completions: []*habitspb.Completion{
				{HabitId: "abc", Date: "2026-10-14", CreatedAt: timestamppb.New(created)},
			},
		})

		require.Len(t, out, 1)
		assert.Equal(t, "abc", out[0].HabitID)
		assert.Equal(t, "2026-10-14", out[0].Date)
		assert.Equal(t, created, out[0].CreatedAt)
	})

	t.Run("пустой — make, не nil", func(t *testing.T) {
		out := protoToCompletionListResponse(&habitspb.GetCompletionsResponse{
			Completions: []*habitspb.Completion{},
		})

		assert.NotNil(t, out)
		assert.Empty(t, out)
	})
}

// ===== createRequestToProto =====
func TestCreateRequestToProto(t *testing.T) {
	out := createRequestToProto(CreateHabitRequest{
		Name:        "Пить воду",
		Description: "2 литра",
	})

	assert.Equal(t, "Пить воду", out.Name)
	assert.Equal(t, "2 литра", out.Description)
}

// ===== patchRequestToProto =====
func TestPatchRequestToProto(t *testing.T) {
	t.Run("пустой патч — поля nil", func(t *testing.T) {
		out := patchRequestToProto("abc", PatchHabitRequest{})

		assert.Equal(t, "abc", out.Id)
		assert.Nil(t, out.Name)
		assert.Nil(t, out.Description)
	})

	t.Run("только name", func(t *testing.T) {
		name := "Новое имя"
		out := patchRequestToProto("abc", PatchHabitRequest{Name: &name})

		assert.Equal(t, "abc", out.Id)
		require.NotNil(t, out.Name)
		assert.Equal(t, "Новое имя", *out.Name)
		assert.Nil(t, out.Description)
	})

	t.Run("только description", func(t *testing.T) {
		desc := "Новое описание"
		out := patchRequestToProto("abc", PatchHabitRequest{Description: &desc})

		require.NotNil(t, out.Description)
		assert.Equal(t, "Новое описание", *out.Description)
		assert.Nil(t, out.Name)
	})
}

// ===== completeRequestToProto =====
func TestCompleteRequestToProto(t *testing.T) {
	out := completeRequestToProto("abc", CompleteHabitRequest{Date: "2026-10-15"})

	assert.Equal(t, "abc", out.Id)
	assert.Equal(t, "2026-10-15", out.Date)
}

// ===== uncompleteRequestToProto =====
func TestUncompleteRequestToProto(t *testing.T) {
	out := uncompleteRequestToProto("abc", UncompleteHabitRequest{Date: "2026-10-14"})

	assert.Equal(t, "abc", out.Id)
	assert.Equal(t, "2026-10-14", out.Date)
}