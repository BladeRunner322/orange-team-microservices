package domain

import "time"

// ============================================================
//  Weight (1..1000 кг)
// ============================================================

type Weight int

const (
	MinWeight = 1
	MaxWeight = 1000
)

func NewWeight(raw int) (Weight, error) {
	// TODO: validate range
	return Weight(raw), nil
}

func (w Weight) Int() int { return int(w) }

// ============================================================
//  Sets (1..100)
// ============================================================

type Sets int

const (
	MinSets = 1
	MaxSets = 100
)

func NewSets(raw int) (Sets, error) {
	// TODO: validate range
	return Sets(raw), nil
}

func (s Sets) Int() int { return int(s) }

// ============================================================
//  Reps (1..1000)
// ============================================================

type Reps int

const (
	MinReps = 1
	MaxReps = 1000
)

func NewReps(raw int) (Reps, error) {
	// TODO: validate range
	return Reps(raw), nil
}

func (r Reps) Int() int { return int(r) }

// ============================================================
//  Duration (1..86400 секунд)
// ============================================================

type Duration int

const (
	MinDuration = 1
	MaxDuration = 86400
)

func NewDuration(raw int) (Duration, error) {
	// TODO: validate range
	return Duration(raw), nil
}

func (d Duration) Int() int { return int(d) }

// ============================================================
//  Difficulty (1..10), снапшот из Exercises
// ============================================================

type Difficulty int

const (
	MinDifficulty = 1
	MaxDifficulty = 10
)

func NewDifficulty(raw int) (Difficulty, error) {
	// TODO: validate range
	return Difficulty(raw), nil
}

func (d Difficulty) Int() int { return int(d) }

// ============================================================
//  Intensity (1..10)
// ============================================================

type Intensity int

const (
	MinIntensity = 1
	MaxIntensity = 10
)

func NewIntensity(raw int) (Intensity, error) {
	// TODO: validate range
	return Intensity(raw), nil
}

func (i Intensity) Int() int { return int(i) }

// ============================================================
//  Status
// ============================================================

type Status string

const (
	StatusPlanned    Status = "planned"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

func NewStatus(raw string) (Status, error) {
	// TODO: validate status
	return Status(raw), nil
}

func (s Status) String() string { return string(s) }

func (s Status) IsTerminal() bool {
	// TODO: completed / cancelled → true
	return false
}

// ============================================================
//  ExerciseType (weight / duration), снапшот из Exercises
// ============================================================

type ExerciseType string

const (
	ExerciseTypeWeight   ExerciseType = "weight"
	ExerciseTypeDuration ExerciseType = "duration"
)

func NewExerciseType(raw string) (ExerciseType, error) {
	// TODO: validate type
	return ExerciseType(raw), nil
}

func (t ExerciseType) String() string { return string(t) }

// Вспомогательная функция, используется в coefficient.go.
func dayOnly(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}
