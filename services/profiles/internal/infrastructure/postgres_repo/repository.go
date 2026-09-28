// Package postgres_repo — реализация ports.Repository для PostgreSQL.
package postgres_repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	"github.com/google/uuid"
)

// Repository реализует интерфейс ports.Repository для PostgreSQL.
type Repository struct {
	pool postgres.Pool
}

// NewRepository создаёт новый репозиторий с пулом соединений.
func NewRepository(pool postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetByUserID возвращает профиль по user_id.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT user_id, sex, weight_grams, birth_date, height_cm, created_at, updated_at
        FROM profiles.users
        WHERE user_id = $1
    `

	var m ProfileModel
	row := r.pool.QueryRow(ctx, query, userID)
	err := row.Scan(
		&m.UserID,
		&m.Sex,
		&m.WeightGrams,
		&m.BirthDate,
		&m.HeightCM,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if errors.Is(err, postgres.ErrNoRows) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}

	if err != nil {
		return domain.Profile{}, fmt.Errorf("find profile by user_id: %w", err)
	}

	return ModelToDomain(m)
}

// Upsert создаёт пустой профиль, если его нет (idempotent, для lazy-create).
func (r *Repository) Upsert(ctx context.Context, profile domain.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	m := DomainToModel(profile)

	query := `
	    INSERT INTO profiles.users (user_id, created_at)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, m.UserID, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}

	return nil
}

// Update обновляет все поля профиля и ставит updated_at.
func (r *Repository) Update(ctx context.Context, profile domain.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
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
	`
	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		m.UserID,
		m.Sex,
		m.WeightGrams,
		m.BirthDate,
		m.HeightCM,
	)

	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrProfileNotFound
	}

	return nil

}

// Delete удаляет профиль по user_id.
func (r *Repository) Delete(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		DELETE FROM profiles.users
		WHERE user_id = $1
	`

	cmdTag, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}
