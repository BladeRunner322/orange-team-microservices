package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
)

func TestApplyPatch_DontTouch(t *testing.T) {
	// заполненный профиль
	male, _ := NewSex("male")
	weight, _ := NewWeightFromKilograms(80.5)
	bd, _ := NewBirthDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
	height, _ := NewHeight(180)
	profile := NewProfile(uuid.New(), &male, &weight, &bd, &height)

	// пустой патч — все Set: false
	patch := NewProfilePatch(
		nullable.Nullable[Sex]{},
		nullable.Nullable[Weight]{},
		nullable.Nullable[BirthDate]{},
		nullable.Nullable[Height]{},
	)

	profile.ApplyPatch(patch)

	require.NotNil(t, profile.Sex())
	assert.Equal(t, SexMale, *profile.Sex())
	require.NotNil(t, profile.Weight())
	assert.Equal(t, 80_500, profile.Weight().Grams())
	require.NotNil(t, profile.Height())
	assert.Equal(t, 180, profile.Height().Centimeters())
	require.NotNil(t, profile.BirthDate())
	assert.True(t, profile.Completed())
}

func TestApplyPatch_Set(t *testing.T) {
	profile := NewEmptyProfile(uuid.New())

	male, _ := NewSex("male")
	weight, _ := NewWeightFromKilograms(80.5)
	bd, _ := NewBirthDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
	height, _ := NewHeight(180)

	patch := NewProfilePatch(
		nullable.Nullable[Sex]{Set: true, Value: &male},
		nullable.Nullable[Weight]{Set: true, Value: &weight},
		nullable.Nullable[BirthDate]{Set: true, Value: &bd},
		nullable.Nullable[Height]{Set: true, Value: &height},
	)

	profile.ApplyPatch(patch)

	require.NotNil(t, profile.Sex())
	assert.Equal(t, SexMale, *profile.Sex())
	require.NotNil(t, profile.Weight())
	assert.Equal(t, 80_500, profile.Weight().Grams())
	require.NotNil(t, profile.BirthDate())
	assert.Equal(t, "1990-05-15", profile.BirthDate().String())
	require.NotNil(t, profile.Height())
	assert.Equal(t, 180, profile.Height().Centimeters())
	assert.True(t, profile.Completed())
}

func TestApplyPatch_Clear(t *testing.T) {
	// заполненный профиль
	male, _ := NewSex("male")
	weight, _ := NewWeightFromKilograms(80.5)
	profile := NewProfile(uuid.New(), &male, &weight, nil, nil)

	require.NotNil(t, profile.Sex())
	require.NotNil(t, profile.Weight())

	// patch: убрать sex и weight (Set: true, Value: nil)
	patch := NewProfilePatch(
		nullable.Nullable[Sex]{Set: true, Value: nil},
		nullable.Nullable[Weight]{Set: true, Value: nil},
		nullable.Nullable[BirthDate]{},
		nullable.Nullable[Height]{},
	)

	profile.ApplyPatch(patch)

	assert.Nil(t, profile.Sex())
	assert.Nil(t, profile.Weight())
	assert.False(t, profile.Completed())
}

func TestApplyPatch_Mixed(t *testing.T) {
	// профиль с sex и weight
	male, _ := NewSex("male")
	weight, _ := NewWeightFromKilograms(80.5)
	profile := NewProfile(uuid.New(), &male, &weight, nil, nil)

	// patch: убрать sex, установить height, weight не трогать
	height, _ := NewHeight(180)
	patch := NewProfilePatch(
		nullable.Nullable[Sex]{Set: true, Value: nil},
		nullable.Nullable[Weight]{}, // не трогать
		nullable.Nullable[BirthDate]{},
		nullable.Nullable[Height]{Set: true, Value: &height},
	)

	profile.ApplyPatch(patch)

	assert.Nil(t, profile.Sex())
	require.NotNil(t, profile.Weight())
	assert.Equal(t, 80_500, profile.Weight().Grams())
	require.NotNil(t, profile.Height())
	assert.Equal(t, 180, profile.Height().Centimeters())
}
