package postgres_repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
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

// Save сохраняет пользователя в БД.
func (r *Repository) Save(ctx context.Context, user domain.User) error {
	m := domainToUserModel(user)
	query := `
		INSERT INTO auth.users (id, email, password_hash, full_name, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		m.ID,
		m.Email,
		m.PasswordHash,
		m.FullName,
		m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

// FindByEmail ищет пользователя по email.
func (r *Repository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, created_at, updated_at
		FROM auth.users WHERE email = $1
	`
	var m UserModel
	row := r.pool.QueryRow(ctx, query, email.String())
	err := row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.FullName, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, postgres.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return userModelToDomain(m)
}

// FindByID ищет пользователя по ID.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, created_at, updated_at
		FROM auth.users WHERE id = $1
	`
	var m UserModel
	row := r.pool.QueryRow(ctx, query, id)
	err := row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.FullName, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, postgres.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return userModelToDomain(m)
}
