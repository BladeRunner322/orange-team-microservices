//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/nullable"
	"github.com/BladeRunner322/orange-team-microservices/pkg/pgconfig"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/services/profiles/internal/domain"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trmmanager "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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
	pool, err := postgres.NewPgxPool(ctx, pgconfig.Config{
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

	// Создаём менеджер и репозиторий на одном пуле, как в bootstrap.
	transactions, err := trmmanager.New(trmpgx.NewDefaultFactory(pool.Pool))
	require.NoError(t, err)

	repo := NewRepository(pool.Pool, trmpgx.DefaultCtxGetter, pool.OpTimeout())
	runRepositoryTests(t, repo, transactions)
}

func runRepositoryTests(t *testing.T, repo *Repository, transactions *trmmanager.Manager) {
	ctx := context.Background()

	newEmptyProfile := func() domain.Profile {
		return domain.NewEmptyProfile(uuid.New())
	}

	t.Run("EnsureExists создаёт пустой профиль", func(t *testing.T) {
		profile := newEmptyProfile()

		require.NoError(t, repo.EnsureExists(ctx, profile))

		got, err := repo.GetByUserID(ctx, profile.UserID())
		require.NoError(t, err)

		assert.Equal(t, profile.UserID(), got.UserID())
		assert.False(t, got.Completed())
		assert.Nil(t, got.Sex())
		assert.Nil(t, got.Weight())
		assert.Nil(t, got.BirthDate())
		assert.Nil(t, got.Height())
	})

	t.Run("EnsureExists дважды не падает (idempotent)", func(t *testing.T) {
		profile := newEmptyProfile()

		require.NoError(t, repo.EnsureExists(ctx, profile))
		require.NoError(t, repo.EnsureExists(ctx, profile))

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
		require.NoError(t, repo.EnsureExists(ctx, profile))

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
		require.NoError(t, repo.EnsureExists(ctx, profile))

		require.NoError(t, repo.Delete(ctx, profile.UserID()))

		_, err := repo.GetByUserID(ctx, profile.UserID())
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("Delete несуществующего — ErrProfileNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})

	t.Run("Do — блокировка строки, rollback и commit", func(t *testing.T) {
		profile := newEmptyProfile()
		require.NoError(t, repo.EnsureExists(ctx, profile))

		height, err := domain.NewHeight(180)
		require.NoError(t, err)
		patch := domain.ProfilePatch{
			Height: nullable.Nullable[domain.Height]{Set: true, Value: &height},
		}

		rollbackErr := errors.New("force rollback")
		var lockErr error

		err = transactions.Do(ctx, func(txCtx context.Context) error {
			current, err := repo.GetByUserIDForUpdate(txCtx, profile.UserID())
			if err != nil {
				return err
			}

			// Через пул берём другое соединение. NOWAIT проверяет,
			// что первая транзакция уже держит блокировку строки.
			lockCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			var userID uuid.UUID
			lockErr = repo.db.QueryRow(lockCtx, `
				SELECT user_id FROM profiles.users
				WHERE user_id = $1
				FOR UPDATE NOWAIT
			`, profile.UserID()).Scan(&userID)

			current.ApplyPatch(patch)
			if _, err := repo.Update(txCtx, current); err != nil {
				return err
			}

			return rollbackErr
		})
		require.ErrorIs(t, err, rollbackErr)

		var pgErr *pgconn.PgError
		require.ErrorAs(t, lockErr, &pgErr)
		require.Equal(t, "55P03", pgErr.Code, "строка должна быть заблокирована")

		afterRollback, err := repo.GetByUserID(ctx, profile.UserID())
		require.NoError(t, err)
		require.Nil(t, afterRollback.Height())

		err = transactions.Do(ctx, func(txCtx context.Context) error {
			current, err := repo.GetByUserIDForUpdate(txCtx, profile.UserID())
			if err != nil {
				return err
			}

			current.ApplyPatch(patch)
			_, err = repo.Update(txCtx, current)
			return err
		})
		require.NoError(t, err)

		afterCommit, err := repo.GetByUserID(ctx, profile.UserID())
		require.NoError(t, err)
		require.NotNil(t, afterCommit.Height())
		require.Equal(t, height, *afterCommit.Height())
	})
}
