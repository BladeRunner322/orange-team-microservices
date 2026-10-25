package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
)

// ============================================================
//  mockWorkoutRepository
// ============================================================

type mockWorkoutRepository struct {
	workouts     map[uuid.UUID]domain.Workout
	exerciseRepo *mockWorkoutExerciseRepository // нужен для WithTx

	// err — если задана, все методы возвращают её.
	err error
}

func newMockWorkoutRepository() *mockWorkoutRepository {
	return &mockWorkoutRepository{
		workouts:     make(map[uuid.UUID]domain.Workout),
		exerciseRepo: newMockWorkoutExerciseRepository(),
	}
}

func (m *mockWorkoutRepository) WithTx(
	ctx context.Context,
	fn func(ports.WorkoutRepository, ports.WorkoutExerciseRepository) error,
) error {
	if m.err != nil {
		return m.err
	}
	// В памяти транзакции нет — прокидываем те же моки.
	return fn(m, m.exerciseRepo)
}

func (m *mockWorkoutRepository) Create(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	if m.err != nil {
		return domain.Workout{}, m.err
	}
	m.workouts[workout.ID()] = workout
	return workout, nil
}

func (m *mockWorkoutRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (domain.Workout, error) {
	if m.err != nil {
		return domain.Workout{}, m.err
	}

	workout, ok := m.workouts[id]
	if !ok {
		return domain.Workout{}, domain.ErrWorkoutNotFound
	}

	// имитируем WHERE user_id = $1 AND id = $2
	if workout.UserID() != userID {
		return domain.Workout{}, domain.ErrWorkoutNotFound
	}

	return workout, nil
}

func (m *mockWorkoutRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Workout, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := make([]domain.Workout, 0)
	for _, w := range m.workouts {
		if w.UserID() == userID && !w.IsDeleted() {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *mockWorkoutRepository) Update(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	if m.err != nil {
		return domain.Workout{}, m.err
	}

	existing, ok := m.workouts[workout.ID()]
	if !ok || existing.IsDeleted() {
		return domain.Workout{}, domain.ErrWorkoutNotFound
	}

	m.workouts[workout.ID()] = workout
	return workout, nil
}

func (m *mockWorkoutRepository) MarkDeleted(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}

	existing, ok := m.workouts[id]
	if !ok || existing.IsDeleted() || existing.UserID() != userID {
		return domain.ErrWorkoutNotFound
	}

	// имитируем soft delete: пересоздаём сущность с isDeleted = true
	// (в реальной БД это делается через UPDATE ... SET deleted_at = NOW())
	deleted := domain.RestoreWorkout(
		existing.ID(),
		existing.Version(),
		existing.UserID(),
		existing.Status(),
		existing.StartedAt(),
		existing.CompletedAt(),
		existing.WorkoutScore(),
		existing.Intensity(),
		existing.PersonalScoreCoefficient(),
		true,
		existing.CreatedAt(),
		existing.UpdatedAt(),
	)
	m.workouts[id] = deleted
	return nil
}

func (m *mockWorkoutRepository) GetUserScore(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}

	var sum int64
	for _, w := range m.workouts {
		if w.UserID() == userID && !w.IsDeleted() && w.Status() == domain.StatusCompleted {
			sum += int64(w.WorkoutScore())
		}
	}
	return sum, nil
}

// ============================================================
//  mockWorkoutExerciseRepository
// ============================================================

type mockWorkoutExerciseRepository struct {
	exercises map[uuid.UUID]domain.WorkoutExercise
	err       error
}

func newMockWorkoutExerciseRepository() *mockWorkoutExerciseRepository {
	return &mockWorkoutExerciseRepository{
		exercises: make(map[uuid.UUID]domain.WorkoutExercise),
	}
}

func (m *mockWorkoutExerciseRepository) Create(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error) {
	if m.err != nil {
		return domain.WorkoutExercise{}, m.err
	}
	m.exercises[exercise.ID()] = exercise
	return exercise, nil
}

func (m *mockWorkoutExerciseRepository) GetByID(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) (domain.WorkoutExercise, error) {
	if m.err != nil {
		return domain.WorkoutExercise{}, m.err
	}

	exercise, ok := m.exercises[id]
	if !ok || exercise.WorkoutID() != workoutID {
		return domain.WorkoutExercise{}, domain.ErrWorkoutExerciseNotFound
	}
	return exercise, nil
}

func (m *mockWorkoutExerciseRepository) ListByWorkout(ctx context.Context, workoutID uuid.UUID) ([]domain.WorkoutExercise, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := make([]domain.WorkoutExercise, 0)
	for _, e := range m.exercises {
		if e.WorkoutID() == workoutID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockWorkoutExerciseRepository) Update(ctx context.Context, exercise domain.WorkoutExercise) (domain.WorkoutExercise, error) {
	if m.err != nil {
		return domain.WorkoutExercise{}, m.err
	}

	existing, ok := m.exercises[exercise.ID()]
	if !ok || existing.WorkoutID() != exercise.WorkoutID() {
		return domain.WorkoutExercise{}, domain.ErrWorkoutExerciseNotFound
	}

	m.exercises[exercise.ID()] = exercise
	return exercise, nil
}

func (m *mockWorkoutExerciseRepository) Delete(ctx context.Context, workoutID uuid.UUID, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}

	existing, ok := m.exercises[id]
	if !ok || existing.WorkoutID() != workoutID {
		return domain.ErrWorkoutExerciseNotFound
	}

	delete(m.exercises, id)
	return nil
}

// ============================================================
//  mockProfilesClient
// ============================================================

type mockProfilesClient struct {
	profile ports.UserProfile
	err     error
}

func (m *mockProfilesClient) GetProfile(ctx context.Context, userID uuid.UUID) (ports.UserProfile, error) {
	if m.err != nil {
		return ports.UserProfile{}, m.err
	}
	return m.profile, nil
}

func (m *mockProfilesClient) IsHealthy(ctx context.Context) error { return nil }
func (m *mockProfilesClient) Close()                              {}

// ============================================================
//  mockExercisesClient
// ============================================================

type mockExercisesClient struct {
	info ports.ExerciseInfo
	err  error
}

func (m *mockExercisesClient) GetExercise(ctx context.Context, id uuid.UUID) (ports.ExerciseInfo, error) {
	if m.err != nil {
		return ports.ExerciseInfo{}, m.err
	}
	return m.info, nil
}

func (m *mockExercisesClient) IsHealthy(ctx context.Context) error { return nil }
func (m *mockExercisesClient) Close()                              {}
