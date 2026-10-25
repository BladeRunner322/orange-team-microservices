package usecases

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

// ============================================================
//  VO-хелперы (упрощают создание валидных VO в тестах)
// ============================================================

func ptrWeight(t *testing.T, raw int) *domain.Weight {
	t.Helper()
	w, err := domain.NewWeight(raw)
	require.NoError(t, err)
	return &w
}

func ptrSets(t *testing.T, raw int) *domain.Sets {
	t.Helper()
	s, err := domain.NewSets(raw)
	require.NoError(t, err)
	return &s
}

func ptrReps(t *testing.T, raw int) *domain.Reps {
	t.Helper()
	r, err := domain.NewReps(raw)
	require.NoError(t, err)
	return &r
}

func ptrDuration(t *testing.T, raw int) *domain.Duration {
	t.Helper()
	d, err := domain.NewDuration(raw)
	require.NoError(t, err)
	return &d
}

func ptrDifficulty(t *testing.T, raw int) *domain.Difficulty {
	t.Helper()
	d, err := domain.NewDifficulty(raw)
	require.NoError(t, err)
	return &d
}

// ============================================================
//  Хелперы для создания Workout в нужном состоянии
// ============================================================

// newWorkoutInStatus создаёт тренировку в заданном статусе.
func newWorkoutInStatus(t *testing.T, userID uuid.UUID, status domain.Status) domain.Workout {
	t.Helper()
	now := time.Now().UTC()
	return domain.RestoreWorkout(
		uuid.New(), 0, userID, status,
		nil, nil, 0, nil, 5,
		false, now, now,
	)
}

// newDeletedWorkout создаёт удалённую тренировку.
func newDeletedWorkout(t *testing.T, userID uuid.UUID) domain.Workout {
	t.Helper()
	now := time.Now().UTC()
	return domain.RestoreWorkout(
		uuid.New(), 0, userID, domain.StatusPlanned,
		nil, nil, 0, nil, 5,
		true, now, now,
	)
}
