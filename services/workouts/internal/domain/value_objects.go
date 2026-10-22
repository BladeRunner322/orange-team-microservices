package domain

// Weight — value object веса.
type Weight int

const (
	MinWeight = 1
	MaxWeight = 1000
)

func NewWeight(raw int) (Weight, error) {
	if raw < MinWeight || raw > MaxWeight {
		return 0, ErrInvalidWeight
	}

	return Weight(raw), nil
}

func (w Weight) Int() int { return int(w) }

// Sets — value object подходов.
type Sets int

const (
	MinSets = 1
	MaxSets = 100
)

func NewSets(raw int) (Sets, error) {
	if raw < MinSets || raw > MaxSets {
		return 0, ErrInvalidSets
	}

	return Sets(raw), nil
}

func (s Sets) Int() int { return int(s) }

// Reps — value object повторений.
type Reps int

const (
	MinReps = 1
	MaxReps = 1000
)

func NewReps(raw int) (Reps, error) {
	if raw < MinReps || raw > MaxReps {
		return 0, ErrInvalidReps
	}

	return Reps(raw), nil
}

func (r Reps) Int() int { return int(r) }

// Duration — value object времени выполнения.
type Duration int

const (
	MinDuration = 1
	MaxDuration = 86400
)

func NewDuration(raw int) (Duration, error) {
	if raw < MinDuration || raw > MaxDuration {
		return 0, ErrInvalidDuration
	}

	return Duration(raw), nil
}

func (d Duration) Int() int { return int(d) }

// Difficulty — value object сложности (снапшот из Exercises).
type Difficulty int

const (
	MinDifficulty = 1
	MaxDifficulty = 10
)

func NewDifficulty(raw int) (Difficulty, error) {
	if raw < MinDifficulty || raw > MaxDifficulty {
		return 0, ErrInvalidDifficulty
	}

	return Difficulty(raw), nil
}

func (d Difficulty) Int() int { return int(d) }

// Intensity — value object интенсивности.
type Intensity int

const (
	MinIntensity = 1
	MaxIntensity = 10
)

func NewIntensity(raw int) (Intensity, error) {
	if raw < MinIntensity || raw > MaxIntensity {
		return 0, ErrInvalidIntensity
	}

	return Intensity(raw), nil
}

func (i Intensity) Int() int { return int(i) }

// Status — value object статуса.
type Status string

const (
	StatusPlanned    Status = "planned"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

func NewStatus(raw string) (Status, error) {
	switch Status(raw) {
	case StatusPlanned, StatusInProgress, StatusCompleted, StatusCancelled:
		return Status(raw), nil
	}

	return "", ErrInvalidStatus
}

func (s Status) String() string { return string(s) }

func (s Status) IsTerminal() bool {
	if s == StatusCompleted || s == StatusCancelled {
		return true
	}

	return false
}

// ExerciseType — value object типа упражнения (снапшот из Exercises).
type ExerciseType string

const (
	ExerciseTypeWeight   ExerciseType = "weight"
	ExerciseTypeDuration ExerciseType = "duration"
)

func NewExerciseType(raw string) (ExerciseType, error) {
	switch ExerciseType(raw) {
	case ExerciseTypeWeight, ExerciseTypeDuration:
		return ExerciseType(raw), nil
	}

	return "", ErrInvalidExerciseType
}

func (t ExerciseType) String() string { return string(t) }
