package postgres_repo

import (
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/google/uuid"
)

type WorkoutExerciseModel struct {
	ID           uuid.UUID
	Version      int
	WorkoutID    uuid.UUID
	ExerciseID   uuid.UUID
	Difficulty   int
	Weight       *int
	Sets         *int
	Reps         *int
	Duration     *int
	Completed    bool
	ExerciseLoad int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (m *WorkoutExerciseModel) Scan(row postgres.Row) error {
	return row.Scan(
		&m.ID, &m.Version, &m.WorkoutID, &m.ExerciseID, &m.Difficulty,
		&m.Weight, &m.Sets, &m.Reps, &m.Duration,
		&m.Completed, &m.ExerciseLoad,
		&m.CreatedAt, &m.UpdatedAt,
	)
}
