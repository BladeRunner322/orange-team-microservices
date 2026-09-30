package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfile(t *testing.T) {
	t.Run("все поля заполнены", func(t *testing.T) {
		userID := uuid.New()
		sex, _ := NewSex("male")
		weight, _ := NewWeightFromKilograms(80.5)
		height, _ := NewHeight(180)
		birthDate, _ := NewBirthDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))

		profile := NewProfile(userID, &sex, &weight, &birthDate, &height)

		assert.Equal(t, userID, profile.UserID())
		require.NotNil(t, profile.Sex())
		assert.Equal(t, SexMale, *profile.Sex())
		require.NotNil(t, profile.Weight())
		assert.Equal(t, 80_500, profile.Weight().Grams())
		require.NotNil(t, profile.Height())
		assert.Equal(t, 180, profile.Height().Centimeters())
		require.NotNil(t, profile.BirthDate())
		assert.Equal(t, "1990-05-15", profile.BirthDate().String())
	})

	t.Run("createdAt заполнен, updatedAt nil", func(t *testing.T) {
		profile := NewEmptyProfile(uuid.New())

		assert.False(t, profile.CreatedAt().IsZero())
		assert.Nil(t, profile.UpdatedAt())
	})
}

func TestNewEmptyProfile(t *testing.T) {
	userID := uuid.New()
	profile := NewEmptyProfile(userID)

	assert.Equal(t, userID, profile.UserID())
	assert.Nil(t, profile.Sex())
	assert.Nil(t, profile.Weight())
	assert.Nil(t, profile.BirthDate())
	assert.Nil(t, profile.Height())
	assert.False(t, profile.Completed())
}

func TestRestoreProfile(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	profile := RestoreProfile(userID, nil, nil, nil, nil, createdAt, &updatedAt)

	assert.Equal(t, userID, profile.UserID())
	assert.Equal(t, createdAt, profile.CreatedAt())
	require.NotNil(t, profile.UpdatedAt())
	assert.Equal(t, updatedAt, *profile.UpdatedAt())
}

func TestProfileCompleted(t *testing.T) {
	makeProfile := func(sex *Sex, weight *Weight, bd *BirthDate, height *Height) Profile {
		return NewProfile(uuid.New(), sex, weight, bd, height)
	}

	male, _ := NewSex("male")
	weight, _ := NewWeightFromKilograms(80.5)
	bd, _ := NewBirthDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
	height, _ := NewHeight(180)

	t.Run("все поля заданы → true", func(t *testing.T) {
		profile := makeProfile(&male, &weight, &bd, &height)
		assert.True(t, profile.Completed())
	})

	t.Run("без sex → false", func(t *testing.T) {
		profile := makeProfile(nil, &weight, &bd, &height)
		assert.False(t, profile.Completed())
	})

	t.Run("без weight → false", func(t *testing.T) {
		profile := makeProfile(&male, nil, &bd, &height)
		assert.False(t, profile.Completed())
	})

	t.Run("без birthDate → false", func(t *testing.T) {
		profile := makeProfile(&male, &weight, nil, &height)
		assert.False(t, profile.Completed())
	})

	t.Run("без height → false", func(t *testing.T) {
		profile := makeProfile(&male, &weight, &bd, nil)
		assert.False(t, profile.Completed())
	})

	t.Run("пустой профиль → false", func(t *testing.T) {
		assert.False(t, NewEmptyProfile(uuid.New()).Completed())
	})
}
