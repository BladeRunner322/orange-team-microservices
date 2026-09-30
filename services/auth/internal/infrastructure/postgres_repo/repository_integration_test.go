//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"os"
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
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
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

	// 4. Применяем реальные миграции из файлов в порядке номеров.
	// Это гарантирует, что тест проверяет ту же схему, что и прод.
	migrationFiles := []string{
		"../../../migrations/000001_create_users_table.up.sql",
		"../../../migrations/000002_add_role.up.sql",
	}
	for _, path := range migrationFiles {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(content))
		require.NoError(t, err)
	}

	// 5. Создаём репозиторий поверх пула
	repo := NewRepository(pool)

	// 6. Прогоняем тесты
	runRepositoryTests(t, repo)
}

func runRepositoryTests(t *testing.T, repo *Repository) {
	ctx := context.Background()

	// newUser создаёт доменного пользователя с заданным email.
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
		// Пользователь создан в памяти, но не сохранён в БД.
		u := newUser("ghost@example.com")

		_, err := repo.FindByID(ctx, u.ID())

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("Save returns ErrEmailAlreadyExists on duplicate", func(t *testing.T) {
		user1 := newUser("dup@example.com")
		require.NoError(t, repo.Save(ctx, user1))

		// Второй пользователь с тем же email — unique violation должен
		// превратиться в доменную ошибку.
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
