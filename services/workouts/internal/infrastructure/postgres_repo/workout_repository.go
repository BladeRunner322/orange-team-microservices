package postgres_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/workouts/internal/domain"
	"github.com/google/uuid"
)

type WorkoutRepository struct {
	pool      postgres.Pool
	tx        postgres.Tx
	opTimeout time.Duration
}

func NewWorkoutRepository(pool postgres.Pool) *WorkoutRepository {
	return &WorkoutRepository{
		pool:      pool,
		opTimeout: pool.OpTimeout(),
	}
}

func (r *WorkoutRepository) querier() postgres.Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

// WithTx открывает транзакцию и передаёт оба репозитория (ADR-025 п.3.1).
func (r *WorkoutRepository) WithTx(
	ctx context.Context,
	fn func(ports.WorkoutRepository, ports.WorkoutExerciseRepository) error,
) error {
	// Уже в транзакции (значит, fn вызвали изнутри другой WithTx).
	// Создаём exercise-репозиторий на той же tx и передаём оба.
	if r.tx != nil {
		exRepo := &WorkoutExerciseRepository{tx: r.tx, opTimeout: r.opTimeout}
		return fn(r, exRepo)
	}

	// Не в транзакции. Открываем новую.
	return r.pool.WithTx(ctx, func(tx postgres.Tx) error {
		txWorkoutRepo := &WorkoutRepository{tx: tx, opTimeout: r.opTimeout}
		txExerciseRepo := &WorkoutExerciseRepository{tx: tx, opTimeout: r.opTimeout}
		return fn(txWorkoutRepo, txExerciseRepo)
	})
}

// Create вставляет тренировку в БД и возвращает её из RETURNING.
func (r *WorkoutRepository) Create(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		INSERT INTO workouts.workouts (
			id, version, user_id, status, started_at, completed_at, workout_score, intensity, personal_score_coefficient, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, version, user_id, status, started_at, completed_at, workout_score, intensity, personal_score_coefficient, deleted_at, created_at, updated_at
	`
	m := WorkoutDomainToModel(workout)

	row := r.querier().QueryRow(ctx, query,
		m.ID,
		m.Version,
		m.UserID,
		m.Status,
		m.StartedAt,
		m.CompletedAt,
		m.WorkoutScore,
		m.Intensity,
		m.PersonalScoreCoefficient,
		m.CreatedAt,
	)

	if err := m.Scan(row); err != nil {
		return domain.Workout{}, fmt.Errorf("create workout: %w", err)
	}

	created, err := WorkoutModelToDomain(m)
	if err != nil {
		return domain.Workout{}, fmt.Errorf("workout %s: %w", m.ID, err)
	}

	return created, nil
}

// GetByID возвращает тренировку по id для указанного пользователя
// (включая удалённые — с флагом is_deleted, см. ADR-025 п.3.2).
// Чужая тренировка → ErrWorkoutNotFound (существование не раскрываем).
func (r *WorkoutRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (domain.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, version, user_id, status, started_at, completed_at, workout_score, intensity, personal_score_coefficient, deleted_at, created_at, updated_at
		FROM workouts.workouts
		WHERE id = $1 AND user_id = $2
	`

	var m WorkoutModel
	row := r.querier().QueryRow(ctx, query, id, userID)
	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.Workout{}, domain.ErrWorkoutNotFound
		}

		return domain.Workout{}, fmt.Errorf("find workout by id: %w", err)
	}

	workout, err := WorkoutModelToDomain(m)
	if err != nil {
		return domain.Workout{}, fmt.Errorf("workout %s: %w", m.ID, err)
	}

	return workout, nil
}

// GetByUserID возвращает список активных тренировок пользователя
// (только deleted_at IS NULL), новые сверху.
func (r *WorkoutRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, version, user_id, status, started_at, completed_at, workout_score, intensity, personal_score_coefficient, deleted_at, created_at, updated_at
		FROM workouts.workouts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.querier().Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query workouts: %w", err)
	}
	defer rows.Close()

	workouts := make([]domain.Workout, 0)
	for rows.Next() {
		var m WorkoutModel
		if err := m.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}

		workout, err := WorkoutModelToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("workout %s: %w", m.ID, err)
		}

		workouts = append(workouts, workout)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workouts: %w", err)
	}

	return workouts, nil
}

// Update обновляет тренировку и возвращает её из RETURNING.
//
// Optimistic locking: WHERE ... AND version = $N. Если строка не найдена
// (RowsAffected == 0) — ErrConflict. Usecase вызывает Update только после
// успешного GetByID, поэтому «не найдено» здесь означает конкурентное
// изменение или параллельное удаление, а не отсутствие записи.
//
// deleted_at не трогает — управляется только через MarkDeleted.
func (r *WorkoutRepository) Update(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		UPDATE workouts.workouts
		SET
			version = version + 1,
			status = $4,
			started_at = $5,
			completed_at = $6,
			workout_score = $7,
			intensity = $8,
			personal_score_coefficient = $9,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL
		RETURNING id, version, user_id, status, started_at, completed_at, workout_score, intensity, personal_score_coefficient, deleted_at, created_at, updated_at
	`

	m := WorkoutDomainToModel(workout)

	row := r.querier().QueryRow(ctx, query,
		m.ID,
		m.UserID,
		m.Version,
		m.Status,
		m.StartedAt,
		m.CompletedAt,
		m.WorkoutScore,
		m.Intensity,
		m.PersonalScoreCoefficient,
	)

	if err := m.Scan(row); err != nil {
		if errors.Is(err, pgerrors.ErrNoRows) {
			return domain.Workout{}, domain.ErrConflict
		}

		return domain.Workout{}, fmt.Errorf("update workout: %w", err)
	}

	updated, err := WorkoutModelToDomain(m)
	if err != nil {
		return domain.Workout{}, fmt.Errorf("workout %s: %w", m.ID, err)
	}

	return updated, nil
}

// MarkDeleted ставит deleted_at = NOW() и updated_at = NOW().
// Чужая или уже удалённая тренировка → ErrWorkoutNotFound.
// Идемпотентность DELETE обеспечивается в usecase (ADR-007 F-3):
// он перехватывает ErrWorkoutNotFound и возвращает nil.
func (r *WorkoutRepository) MarkDeleted(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		UPDATE workouts.workouts
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	cmdTag, err := r.querier().Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("mark workout deleted: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrWorkoutNotFound
	}

	return nil
}

// GetUserScore возвращает сумму workout_score по всем completed
// тренировкам пользователя (ADR-002). Удалённые не учитываются.
func (r *WorkoutRepository) GetUserScore(ctx context.Context, userID uuid.UUID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT COALESCE(SUM(workout_score), 0)
		FROM workouts.workouts
		WHERE user_id = $1 AND status = 'completed' AND deleted_at IS NULL
	`

	var score int64
	if err := r.querier().QueryRow(ctx, query, userID).Scan(&score); err != nil {
		return 0, fmt.Errorf("get user score: %w", err)
	}

	return score, nil
}
