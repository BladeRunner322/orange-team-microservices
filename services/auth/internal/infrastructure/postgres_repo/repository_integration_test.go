//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
)

func TestRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Поднимаем PostgreSQL в Docker
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:18.6-bookworm",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
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

	// 2. Достаём host и mapped port для сборки конфига pkg/postgres
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	// 3. Создаём пул через наш pkg/postgres (как в bootstrap)
	pool, err := postgres.NewPgxPool(ctx, postgres.Config{
		Host:     host,
		Port:     mappedPort.Port(),
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	// 4. Создаём схему и таблицу (аналог миграции 000001 + 000002)
	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS auth;
		CREATE TABLE IF NOT EXISTS auth.users (
			id UUID PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			role VARCHAR(16) NOT NULL DEFAULT 'user',
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ
		);
	`)
	require.NoError(t, err)

	// 5. Создаём репозиторий поверх пула
	repo := NewRepository(pool)

	// 6. Прогоняем тесты
	runRepositoryTests(t, repo)
}

func runRepositoryTests(t *testing.T, repo *Repository) {
	ctx := context.Background()

	// Хелпер: создаёт доменного пользователя с заданным email
	newUser := func(email string) domain.User {
		e, _ := domain.NewEmail(email)
		passHash, _ := domain.NewPasswordHash("$2a$10$dummyhash")
		fullName, _ := domain.NewFullName("Test User")
		return domain.NewUser(e, passHash, fullName)
	}

	t.Run("Save and FindByEmail", func(t *testing.T) {
		user := newUser("find-by-email@example.com")

		require.NoError(t, repo.Save(ctx, user))

		found, err := repo.FindByEmail(ctx, user.Email())
		require.NoError(t, err)

		assert.Equal(t, user.ID(), found.ID())
		assert.Equal(t, user.Email().String(), found.Email().String())
		assert.Equal(t, user.FullName().String(), found.FullName().String())
		assert.Equal(t, user.Role(), found.Role())
	})

	t.Run("Save and FindByID", func(t *testing.T) {
		user := newUser("find-by-id@example.com")

		require.NoError(t, repo.Save(ctx, user))

		found, err := repo.FindByID(ctx, user.ID())
		require.NoError(t, err)

		assert.Equal(t, user.ID(), found.ID())
		assert.Equal(t, user.Email().String(), found.Email().String())
	})

	t.Run("FindByEmail not found returns ErrUserNotFound", func(t *testing.T) {
		email, _ := domain.NewEmail("nonexistent@example.com")

		_, err := repo.FindByEmail(ctx, email)

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("FindByID not found returns ErrUserNotFound", func(t *testing.T) {
		// случайный UUID, которого точно нет
		id := domain.User{}.ID() // нулевой UUID
		_ = id

		// используем uuid.New() из google/uuid — но чтобы не тащить импорт,
		// сделаем проще: создадим пользователя, получим id, но не сохраним.
		// Нет, лучше честный UUID, которого нет в БД:
		u := newUser("ghost@example.com")

		_, err := repo.FindByID(ctx, u.ID())

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	// ГЛАВНЫЙ ТЕСТ для race-фикса:
	// Repository.Save должен возвращать domain.ErrEmailAlreadyExists,
	// а не сырую ошибку postgres.ErrViolatesUnique.
	t.Run("Save returns ErrEmailAlreadyExists on duplicate", func(t *testing.T) {
		user1 := newUser("dup@example.com")
		require.NoError(t, repo.Save(ctx, user1))

		// Второй пользователь с тем же email
		user2 := newUser("dup@example.com")
		err := repo.Save(ctx, user2)

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists,
			"Repository.Save должен маппить unique violation в доменную ошибку")
	})

	t.Run("New user has role user by default", func(t *testing.T) {
		user := newUser("role-default@example.com")
		require.NoError(t, repo.Save(ctx, user))

		found, err := repo.FindByEmail(ctx, user.Email())
		require.NoError(t, err)

		assert.Equal(t, domain.RoleUser, found.Role())
	})
}
