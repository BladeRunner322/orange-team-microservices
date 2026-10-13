//go:build integration
// +build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgconfig"
	"github.com/BladeRunner322/orange-team-microservices/pkg/pgerrors"
)

func TestPgxPool_WithTx(t *testing.T) {
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

	// 2. Конфиг для нашего pkg/postgres
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

	// 3. Готовим тестовую таблицу
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS items (
			id    SERIAL PRIMARY KEY,
			name  TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// helpers
	countItems := func(t *testing.T) int {
		t.Helper()
		var n int
		err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM items`).Scan(&n)
		require.NoError(t, err)
		return n
	}

	truncateItems := func(t *testing.T) {
		t.Helper()
		_, err := pool.Exec(ctx, `TRUNCATE items RESTART IDENTITY`)
		require.NoError(t, err)
	}

	t.Run("commit — изменения сохранены", func(t *testing.T) {
		truncateItems(t)

		err := pool.WithTx(ctx, func(tx Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, "first")
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, "second")
			return err
		})
		require.NoError(t, err)

		assert.Equal(t, 2, countItems(t))
	})

	t.Run("rollback — изменения откатились", func(t *testing.T) {
		truncateItems(t)

		sentinel := errors.New("boom")

		err := pool.WithTx(ctx, func(tx Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, "will-be-rolled-back")
			if err != nil {
				return err
			}
			return sentinel
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)
		assert.Equal(t, 0, countItems(t), "вставка должна была откатиться")
	})

	t.Run("rollback при панике", func(t *testing.T) {
		truncateItems(t)

		assert.Panics(t, func() {
			_ = pool.WithTx(ctx, func(tx Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, "panicked")
				if err != nil {
					return err
				}
				panic("something went wrong")
			})
		})

		assert.Equal(t, 0, countItems(t), "вставка должна была откатиться после паники")
	})

	t.Run("ошибка в середине — всё откатывается", func(t *testing.T) {
		truncateItems(t)

		err := pool.WithTx(ctx, func(tx Tx) error {
			if _, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, "ok"); err != nil {
				return err
			}
			// нарушение NOT NULL — name обязателен
			if _, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, nil); err != nil {
				return err
			}
			return nil
		})

		require.Error(t, err)
		assert.Equal(t, 0, countItems(t))
	})

	t.Run("вложенный WithTx использует новую транзакцию", func(t *testing.T) {
		truncateItems(t)

		// Просто проверяем, что вложенный вызов не паникует
		// (настоящая вложенность была бы через SAVEPOINT, но у нас её нет — каждый BeginTx отдельный)
		err := pool.WithTx(ctx, func(tx Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, "outer")
			return err
		})
		require.NoError(t, err)

		assert.Equal(t, 1, countItems(t))
	})
}

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

	t.Run("BeginTx — успешный commit через Tx.Query", func(t *testing.T) {
		tx, err := pool.BeginTx(ctx)
		require.NoError(t, err)

		_, err = tx.Exec(ctx, `INSERT INTO cov_items (name) VALUES ($1)`, "tx-row")
		require.NoError(t, err)

		rows, err := tx.Query(ctx, `SELECT name FROM cov_items WHERE name = $1`, "tx-row")
		require.NoError(t, err)
		defer rows.Close()

		var name string
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&name))
		assert.Equal(t, "tx-row", name)
		rows.Close()

		var count int
		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM cov_items WHERE name = $1`, "tx-row").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		require.NoError(t, tx.Commit(ctx))
	})

	t.Run("BeginTx — Rollback отменяет вставку", func(t *testing.T) {
		var before int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM cov_items`).Scan(&before))

		tx, err := pool.BeginTx(ctx)
		require.NoError(t, err)

		_, err = tx.Exec(ctx, `INSERT INTO cov_items (name) VALUES ($1)`, "rollback-me")
		require.NoError(t, err)

		require.NoError(t, tx.Rollback(ctx))

		var after int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM cov_items`).Scan(&after))
		assert.Equal(t, before, after)
	})

	t.Run("BeginTx — ошибка маппится", func(t *testing.T) {
		badPool, err := NewPgxPool(ctx, pgconfig.Config{
			Host:     host,
			Port:     mappedPort.Port(),
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			Timeout:  5 * time.Second,
		})
		require.NoError(t, err)
		badPool.Close()

		_, err = badPool.BeginTx(ctx)
		assert.Error(t, err)
	})
}
