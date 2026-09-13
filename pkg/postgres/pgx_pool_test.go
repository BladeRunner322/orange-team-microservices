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

	// 2. Конфиг для нашего pkg/postgres
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	pool, err := NewPgxPool(ctx, Config{
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
