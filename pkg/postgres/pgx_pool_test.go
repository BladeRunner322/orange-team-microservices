//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgconfig"
	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPgxPool_Coverage(t *testing.T) {
	ctx := context.Background()

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
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	pool, err := NewPgxPool(ctx, pgconfig.Config{
		Host:     host,
		Port:     mappedPort.Port(),
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Timeout:  5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS cov_items (
			id    SERIAL PRIMARY KEY,
			name  TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	t.Run("OpTimeout возвращает заданный таймаут", func(t *testing.T) {
		assert.Equal(t, 5*time.Second, pool.OpTimeout())
	})

	t.Run("Exec — RowsAffected через CommandTag", func(t *testing.T) {
		_, err := pool.Exec(ctx, `TRUNCATE cov_items`)
		require.NoError(t, err)

		tag, err := pool.Exec(ctx, `INSERT INTO cov_items (name) VALUES ($1)`, "x")
		require.NoError(t, err)
		assert.Equal(t, int64(1), tag.RowsAffected())

		tag, err = pool.Exec(ctx, `INSERT INTO cov_items (name) VALUES ($1), ($2)`, "y", "z")
		require.NoError(t, err)
		assert.Equal(t, int64(2), tag.RowsAffected())
	})

	t.Run("Exec — ошибка маппится через pgerrors", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO nonexistent_table VALUES (1)`)
		require.Error(t, err)
		assert.ErrorIs(t, err, pgerrors.ErrUnknown)
	})

	t.Run("Query — успешное чтение строк", func(t *testing.T) {
		_, err := pool.Exec(ctx, `TRUNCATE cov_items`)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO cov_items (name) VALUES ($1), ($2)`, "a", "b")
		require.NoError(t, err)

		rows, err := pool.Query(ctx, `SELECT name FROM cov_items ORDER BY name`)
		require.NoError(t, err)
		defer rows.Close()

		var names []string
		for rows.Next() {
			var n string
			require.NoError(t, rows.Scan(&n))
			names = append(names, n)
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, []string{"a", "b"}, names)
	})

	t.Run("QueryRow — успешный scan через pgxRow", func(t *testing.T) {
		var count int
		err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM cov_items`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("QueryRow — ошибка ErrNoRows", func(t *testing.T) {
		var name string
		err := pool.QueryRow(ctx, `SELECT name FROM cov_items WHERE id = 999999`).Scan(&name)
		assert.ErrorIs(t, err, pgerrors.ErrNoRows)
	})
}
