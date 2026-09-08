//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Поднимаем PostgreSQL в Docker
	pgContainer, err := postgres.Run(ctx,
		"postgres:18.6-bookworm",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// 2. Миграции (создаём таблицу)
	_, err = db.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS auth;
		CREATE TABLE IF NOT EXISTS auth.users (
			id UUID PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_auth_users_email ON auth.users (email);
	`)
	require.NoError(t, err)

	// 3. Тесты репозитория (используем db напрямую)
	runRepositoryTests(t, db)
}

func runRepositoryTests(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	t.Run("Save and FindByEmail", func(t *testing.T) {
		// Создаём пользователя
		email, _ := domain.NewEmail("test@example.com")
		passHash, _ := domain.NewPasswordHash("$2a$10$dummyhash")
		fullName, _ := domain.NewFullName("Test User")
		user := domain.NewUser(email, passHash, fullName)

		// Вставляем в БД
		_, err := db.ExecContext(ctx,
			`INSERT INTO auth.users (id, email, password_hash, full_name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			user.ID(),
			user.Email().String(),
			user.PasswordHash().String(),
			user.FullName().String(),
			user.CreatedAt(),
			nil,
		)
		require.NoError(t, err)

		// Ищем по email
		var userModel UserModel
		row := db.QueryRowContext(ctx,
			`SELECT id, email, password_hash, full_name, created_at, updated_at
			FROM auth.users WHERE email = $1`,
			email.String(),
		)
		err = row.Scan(
			&userModel.ID,
			&userModel.Email,
			&userModel.PasswordHash,
			&userModel.FullName,
			&userModel.CreatedAt,
			&userModel.UpdatedAt,
		)
		require.NoError(t, err)

		// Преобразуем в домен (можно использовать маппер)
		found, err := UserModelToDomain(userModel)
		require.NoError(t, err)

		assert.Equal(t, user.ID(), found.ID())
		assert.Equal(t, user.Email().String(), found.Email().String())
		assert.Equal(t, user.FullName().String(), found.FullName().String())
	})

	t.Run("FindByEmail not found", func(t *testing.T) {
		var userModel UserModel
		row := db.QueryRowContext(ctx,
			`SELECT id, email, password_hash, full_name, created_at, updated_at
			FROM auth.users WHERE email = $1`,
			"nonexistent@example.com",
		)
		err := row.Scan(
			&userModel.ID,
			&userModel.Email,
			&userModel.PasswordHash,
			&userModel.FullName,
			&userModel.CreatedAt,
			&userModel.UpdatedAt,
		)
		assert.Error(t, err) // sql.ErrNoRows
	})

	t.Run("Duplicate email", func(t *testing.T) {
		email, _ := domain.NewEmail("duplicate@example.com")
		passHash, _ := domain.NewPasswordHash("$2a$10$dummyhash")
		fullName, _ := domain.NewFullName("User One")
		user1 := domain.NewUser(email, passHash, fullName)

		_, err := db.ExecContext(ctx,
			`INSERT INTO auth.users (id, email, password_hash, full_name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			user1.ID(),
			user1.Email().String(),
			user1.PasswordHash().String(),
			user1.FullName().String(),
			user1.CreatedAt(),
			nil,
		)
		require.NoError(t, err)

		// Вставляем второго с тем же email
		user2 := domain.NewUser(email, passHash, fullName)
		_, err = db.ExecContext(ctx,
			`INSERT INTO auth.users (id, email, password_hash, full_name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			user2.ID(),
			user2.Email().String(),
			user2.PasswordHash().String(),
			user2.FullName().String(),
			user2.CreatedAt(),
			nil,
		)
		assert.Error(t, err) // уникальность нарушена
	})
}
