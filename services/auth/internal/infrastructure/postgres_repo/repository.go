// Package postgres_repo — реализация ports.Repository для PostgreSQL:
// работа с таблицей auth.users и маппинг UserModel ↔ domain.User.
package postgres_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository реализует интерфейс ports.Repository для PostgreSQL.
type Repository struct {
	db        *pgxpool.Pool
	getter    *trmpgx.CtxGetter
	opTimeout time.Duration
}

// NewRepository создаёт новый репозиторий с пулом соединений.
func NewRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter, opTimeout time.Duration) *Repository {
	return &Repository{
		db:        db,
		getter:    getter,
		opTimeout: opTimeout,
	}
}

// queryExecutor выбирает транзакцию или пул и подключает обработку ошибок pgx.
func (r *Repository) queryExecutor(ctx context.Context) postgres.QueryExecutor {
	return postgres.WithMappedErrors(r.getter.DefaultTrOrDB(ctx, r.db))
}

// Save сохраняет пользователя в БД.
func (r *Repository) Save(ctx context.Context, user domain.User) error {
	m := DomainToModel(user)

	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		INSERT INTO auth.users (id, email, password_hash, full_name, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.queryExecutor(ctx).Exec(ctx, query,
		m.ID,
		m.Email,
		m.PasswordHash,
		m.FullName,
		m.Role,
		m.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgerrors.ErrViolatesUnique) {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("save user: %w", err)
	}

	return nil
}

// FindByEmail ищет пользователя по email.
func (r *Repository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {

	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM auth.users 
		WHERE email = $1
	`
	var m UserModel
	row := r.queryExecutor(ctx).QueryRow(ctx, query, email.String())
	err := m.Scan(row)

	if errors.Is(err, pgerrors.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}

	user, err := ModelToDomain(m)
	if err != nil {
		return domain.User{}, fmt.Errorf("user %s: %w", m.ID, err)
	}

	return user, nil
}

// FindByID ищет пользователя по ID.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	query := `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM auth.users
		WHERE id = $1
	`
	var m UserModel
	row := r.queryExecutor(ctx).QueryRow(ctx, query, id)
	err := m.Scan(row)

	if errors.Is(err, pgerrors.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}

	user, err := ModelToDomain(m)
	if err != nil {
		return domain.User{}, fmt.Errorf("user %s: %w", m.ID, err)
	}

	return user, nil
}
