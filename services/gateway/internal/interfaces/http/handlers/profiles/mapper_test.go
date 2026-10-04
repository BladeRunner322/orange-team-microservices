package profiles

import (
	"testing"
	"time"

	profilespb "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/profiles"
	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ===== patchRequestToProto =====
func TestPatchRequestToProto(t *testing.T) {
	t.Run("все поля пустые — все wrapper'ы nil", func(t *testing.T) {
		req := PatchUserRequest{}
		out := patchRequestToProto(req)

		assert.Nil(t, out.Sex)
		assert.Nil(t, out.WeightKg)
		assert.Nil(t, out.BirthDate)
		assert.Nil(t, out.HeightCm)
	})

	t.Run("поля со значением", func(t *testing.T) {
		male := "male"
		bd := "1990-05-15"
		kg := 80.5
		cm := int32(180)

		req := PatchUserRequest{
			Sex:       nullable.Nullable[string]{Set: true, Value: &male},
			WeightKg:  nullable.Nullable[float64]{Set: true, Value: &kg},
			BirthDate: nullable.Nullable[string]{Set: true, Value: &bd},
			HeightCm:  nullable.Nullable[int32]{Set: true, Value: &cm},
		}
		out := patchRequestToProto(req)

		require.NotNil(t, out.Sex)
		require.NotNil(t, out.Sex.Value)
		assert.Equal(t, "male", *out.Sex.Value)

		require.NotNil(t, out.WeightKg)
		require.NotNil(t, out.WeightKg.Value)
		assert.Equal(t, 80.5, *out.WeightKg.Value)

		require.NotNil(t, out.BirthDate)
		require.NotNil(t, out.BirthDate.Value)
		assert.Equal(t, "1990-05-15", *out.BirthDate.Value)

		require.NotNil(t, out.HeightCm)
		require.NotNil(t, out.HeightCm.Value)
		assert.Equal(t, int32(180), *out.HeightCm.Value)
	})

	t.Run("поля на сброс (Set=true, Value=nil)", func(t *testing.T) {
		req := PatchUserRequest{
			Sex:      nullable.Nullable[string]{Set: true, Value: nil},
			WeightKg: nullable.Nullable[float64]{Set: true, Value: nil},
		}
		out := patchRequestToProto(req)

		require.NotNil(t, out.Sex, "wrapper должен быть создан")
		assert.Nil(t, out.Sex.Value, "Value должен быть nil")

		require.NotNil(t, out.WeightKg)
		assert.Nil(t, out.WeightKg.Value)

		assert.Nil(t, out.BirthDate)
		assert.Nil(t, out.HeightCm)
	})
}

// ===== nullable*ToProto =====
func TestNullableStringToProto(t *testing.T) {
	t.Run("Set=false → nil", func(t *testing.T) {
		out := nullableStringToProto(nullable.Nullable[string]{})
		assert.Nil(t, out)
	})

	t.Run("Set=true, Value=nil → пустой wrapper", func(t *testing.T) {
		out := nullableStringToProto(nullable.Nullable[string]{Set: true, Value: nil})
		require.NotNil(t, out)
		assert.Nil(t, out.Value)
	})

	t.Run("Set=true, Value=&v → wrapper с v", func(t *testing.T) {
		v := "male"
		out := nullableStringToProto(nullable.Nullable[string]{Set: true, Value: &v})
		require.NotNil(t, out)
		require.NotNil(t, out.Value)
		assert.Equal(t, "male", *out.Value)
	})
}

func TestNullableDoubleToProto(t *testing.T) {
	t.Run("Set=false → nil", func(t *testing.T) {
		out := nullableDoubleToProto(nullable.Nullable[float64]{})
		assert.Nil(t, out)
	})

	t.Run("Set=true, Value=nil → пустой wrapper", func(t *testing.T) {
		out := nullableDoubleToProto(nullable.Nullable[float64]{Set: true, Value: nil})
		require.NotNil(t, out)
		assert.Nil(t, out.Value)
	})

	t.Run("Set=true, Value=&v → wrapper с v", func(t *testing.T) {
		v := 80.5
		out := nullableDoubleToProto(nullable.Nullable[float64]{Set: true, Value: &v})
		require.NotNil(t, out)
		require.NotNil(t, out.Value)
		assert.Equal(t, 80.5, *out.Value)
	})
}

func TestNullableInt32ToProto(t *testing.T) {
	t.Run("Set=false → nil", func(t *testing.T) {
		out := nullableInt32ToProto(nullable.Nullable[int32]{})
		assert.Nil(t, out)
	})

	t.Run("Set=true, Value=nil → пустой wrapper", func(t *testing.T) {
		out := nullableInt32ToProto(nullable.Nullable[int32]{Set: true, Value: nil})
		require.NotNil(t, out)
		assert.Nil(t, out.Value)
	})

	t.Run("Set=true, Value=&v → wrapper с v", func(t *testing.T) {
		v := int32(180)
		out := nullableInt32ToProto(nullable.Nullable[int32]{Set: true, Value: &v})
		require.NotNil(t, out)
		require.NotNil(t, out.Value)
		assert.Equal(t, int32(180), *out.Value)
	})
}

// ===== protoToUserProfileResponse =====
func TestProtoToUserProfileResponse(t *testing.T) {
	t.Run("полный профиль", func(t *testing.T) {
		created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		updated := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

		out := protoToUserProfileResponse(&profilespb.UserProfile{
			UserId:           "user-123",
			Sex:              "male",
			WeightKg:         80.5,
			BirthDate:        "1990-05-15",
			HeightCm:         180,
			ProfileCompleted: true,
			CreatedAt:        timestamppb.New(created),
			UpdatedAt:        timestamppb.New(updated),
		})

		assert.Equal(t, "user-123", out.UserID)
		assert.Equal(t, "male", out.Sex)
		assert.Equal(t, 80.5, out.WeightKg)
		assert.Equal(t, "1990-05-15", out.BirthDate)
		assert.Equal(t, int32(180), out.HeightCm)
		assert.True(t, out.ProfileCompleted)
		assert.Equal(t, created, out.CreatedAt)
		require.NotNil(t, out.UpdatedAt)
		assert.Equal(t, updated, *out.UpdatedAt)
	})

	t.Run("updated_at nil — указатель nil", func(t *testing.T) {
		out := protoToUserProfileResponse(&profilespb.UserProfile{
			UserId:    "user-123",
			CreatedAt: timestamppb.New(time.Now()),
			UpdatedAt: nil,
		})

		assert.Equal(t, "user-123", out.UserID)
		assert.Nil(t, out.UpdatedAt)
	})

	t.Run("пустой профиль — zero values", func(t *testing.T) {
		out := protoToUserProfileResponse(&profilespb.UserProfile{
			UserId: "user-123",
		})

		assert.Equal(t, "user-123", out.UserID)
		assert.Empty(t, out.Sex)
		assert.Equal(t, 0.0, out.WeightKg)
		assert.False(t, out.ProfileCompleted)
	})
}
