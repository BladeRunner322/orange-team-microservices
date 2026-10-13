// Package postgres_repo — реализация ports.Repository для PostgreSQL:
// работа с таблицей profiles.users и маппинг ProfileModel ↔ domain.Profile.
package postgres_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
)

// Repository реализует интерфейс ports.Repository для PostgreSQL.
type Repository struct {
	pool      postgres.Pool // nil, если репозиторий привязан к транзакции
	tx        postgres.Tx   // nil, если репозиторий не в транзакции
	opTimeout time.Duration
}

// NewRepository создаёт новый репозиторий с пулом соединений.
func NewRepository(pool postgres.Pool) *Repository {
	return &Repository{
		pool:      pool,
		opTimeout: pool.OpTimeout(),
	}
}

// querier возвращает актуальный источник запросов: транзакцию или пул.
func (r *Repository) querier() postgres.Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.pool
}

// WithTx выполняет fn внутри транзакции. Если репозиторий уже привязан
// к транзакции — вложенная не открывается, fn вызывается с текущим.
func (r *Repository) WithTx(ctx context.Context, fn func(ports.Repository) error) error {
	if r.tx != nil {
		return fn(r)
	}
	return r.pool.WithTx(ctx, func(tx postgres.Tx) error {
		txRepo := &Repository{tx: tx, opTimeout: r.opTimeout}
		return fn(txRepo)
	})
}

// GetByUserID возвращает профиль по user_id.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
        SELECT user_id, sex, weight_grams, birth_date, height_cm, created_at, updated_at
        FROM profiles.users
        WHERE user_id = $1
    `

	var m ProfileModel
	row := r.querier().QueryRow(ctx, query, userID)
	err := m.Scan(row)

	if errors.Is(err, pgerrors.ErrNoRows) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	if err != nil {
		return domain.Profile{}, fmt.Errorf("find profile by user_id: %w", err)
	}

	profile, err := ModelToDomain(m)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("profile %s: %w", m.UserID, err)
	}

	return profile, nil
}

// GetByUserIDForUpdate возвращает профиль с блокировкой строки (SELECT ... FOR UPDATE).
// Блокировка держится до COMMIT/ROLLBACK транзакции. Если репозиторий не в tx —
// поведение как у обычного SELECT, блокировка снимается сразу.
func (r *Repository) GetByUserIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
        SELECT user_id, sex, weight_grams, birth_date, height_cm, created_at, updated_at
        FROM profiles.users
        WHERE user_id = $1
        FOR UPDATE
    `

	var m ProfileModel
	row := r.querier().QueryRow(ctx, query, userID)
	err := m.Scan(row)

	if errors.Is(err, pgerrors.ErrNoRows) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	if err != nil {
		return domain.Profile{}, fmt.Errorf("find profile by user_id for update: %w", err)
	}

	profile, err := ModelToDomain(m)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("profile %s: %w", m.UserID, err)
	}

	return profile, nil
}

// EnsureExists создаёт пустой профиль, если его нет (idempotent, для lazy-create).
func (r *Repository) EnsureExists(ctx context.Context, profile domain.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	m := DomainToModel(profile)

	query := `
	    INSERT INTO profiles.users (user_id, created_at)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO NOTHING
	`
	_, err := r.querier().Exec(ctx, query, m.UserID, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("ensure profile: %w", err)
	}

	return nil
}

// Update обновляет все поля профиля и возвращает актуальный профиль из БД
func (r *Repository) Update(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	m := DomainToModel(profile)

	query := `
	   	UPDATE profiles.users
		SET
			sex = $2,
			weight_grams = $3,
			birth_date = $4,
			height_cm = $5,
			updated_at = NOW()
		WHERE user_id = $1
		RETURNING user_id, sex, weight_grams, birth_date, height_cm, created_at, updated_at
	`
	row := r.querier().QueryRow(
		ctx,
		query,
		m.UserID,
		m.Sex,
		m.WeightGrams,
		m.BirthDate,
		m.HeightCM,
	)
	err := m.Scan(row)

	if errors.Is(err, pgerrors.ErrNoRows) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	if err != nil {
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}

	updated, err := ModelToDomain(m)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("profile %s: %w", m.UserID, err)
	}

	return updated, nil
}

// Delete удаляет профиль по user_id.
func (r *Repository) Delete(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		DELETE FROM profiles.users
		WHERE user_id = $1
	`

	cmdTag, err := r.querier().Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}
