//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
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

	// 2. Собираем конфиг для подключения
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	// 3. Создаём пул через pkg/postgres
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

	// 4. Применяем реальную миграцию из файла
	migration, err := os.ReadFile("../../../migrations/000001_create_profiles_table.up.sql")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)

	// 5. Создаём репозиторий
	repo := NewRepository(pool)

	// 6. Прогоняем подтесты
	runRepositoryTests(t, repo)
}

func runRepositoryTests(t *testing.T, repo *Repository) {
	ctx := context.Background()

	newEmptyProfile := func() domain.Profile {
		return domain.NewEmptyProfile(uuid.New())
	}

	t.Run("Upsert создаёт пустой профиль", func(t *testing.T) {
		profile := newEmptyProfile()

		require.NoError(t, repo.Upsert(ctx, profile))

		got, err := repo.GetByUserID(ctx, profile.UserID())
		require.NoError(t, err)

		assert.Equal(t, profile.UserID(), got.UserID())
		assert.False(t, got.Completed())
		assert.Nil(t, got.Sex())
		assert.Nil(t, got.Weight())
		assert.Nil(t, got.BirthDate())
		assert.Nil(t, got.Height())
	})

	t.Run("Upsert дважды не падает (idempotent)", func(t *testing.T) {
		profile := newEmptyProfile()

		require.NoError(t, repo.Upsert(ctx, profile))
		require.NoError(t, repo.Upsert(ctx, profile))

		got, err := repo.GetByUserID(ctx, profile.UserID())
		require.NoError(t, err)
		assert.Equal(t, profile.UserID(), got.UserID())
	})

	t.Run("GetByUserID — not found", func(t *testing.T) {
		_, err := repo.GetByUserID(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("Update обновляет поля и возвращает свежий updated_at", func(t *testing.T) {
		profile := newEmptyProfile()
		require.NoError(t, repo.Upsert(ctx, profile))

		sex, err := domain.NewSex("male")
		require.NoError(t, err)

		weight, err := domain.NewWeightFromKilograms(80.5)
		require.NoError(t, err)

		height, err := domain.NewHeight(180)
		require.NoError(t, err)

		birthDate, err := domain.NewBirthDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		updated := domain.RestoreProfile(
			profile.UserID(),
			&sex, &weight, &birthDate, &height,
			profile.CreatedAt(),
			profile.UpdatedAt(),
		)

		result, err := repo.Update(ctx, updated)
		require.NoError(t, err)

		// То, что вернул Update, содержит заполненный updated_at
		require.NotNil(t, result.UpdatedAt(), "Update должен вернуть заполненный updated_at")

		got, err := repo.GetByUserID(ctx, profile.UserID())
		require.NoError(t, err)

		assert.True(t, got.Completed())
		require.NotNil(t, got.Sex())
		assert.Equal(t, domain.SexMale, *got.Sex())
		require.NotNil(t, got.Weight())
		assert.Equal(t, 80500, got.Weight().Grams())
		require.NotNil(t, got.Height())
		assert.Equal(t, 180, got.Height().Centimeters())
		require.NotNil(t, got.BirthDate())
		assert.Equal(t, "1990-05-15", got.BirthDate().String())

		// Дополнительно: updated_at из возврата совпадает с тем, что в БД
		require.NotNil(t, got.UpdatedAt())
		assert.Equal(t, *result.UpdatedAt(), *got.UpdatedAt())
	})

	t.Run("Update несуществующего — ErrProfileNotFound", func(t *testing.T) {
		profile := newEmptyProfile()

		_, err := repo.Update(ctx, profile)
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("Delete удаляет профиль", func(t *testing.T) {
		profile := newEmptyProfile()
		require.NoError(t, repo.Upsert(ctx, profile))

		require.NoError(t, repo.Delete(ctx, profile.UserID()))

		_, err := repo.GetByUserID(ctx, profile.UserID())
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("Delete несуществующего — ErrProfileNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})
}
