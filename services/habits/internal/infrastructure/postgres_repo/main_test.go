//go:build integration
// +build integration

package postgres_repo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/pkg/pgconfig"
	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testPool — общий пул на весь integration-набор.
// Поднимается один раз в TestMain: и HabitRepository, и CompletionRepository
// ходят в одну схему (completion имеет FK на habits), поэтому отдельный
// контейнер на каждый тест — избыточно.
var testPool postgres.Pool

func TestMain(m *testing.M) {
	code, err := runTests(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration setup failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runTests(m *testing.M) (int, error) {
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
	if err != nil {
		return 0, fmt.Errorf("start container: %w", err)
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return 0, fmt.Errorf("container host: %w", err)
	}

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return 0, fmt.Errorf("container port: %w", err)
	}

	pool, err := postgres.NewPgxPool(ctx, pgconfig.Config{
		Host:     host,
		Port:     mappedPort.Port(),
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Timeout:  5 * time.Second,
	})
	if err != nil {
		return 0, fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	// Реальные миграции — гарантия, что тесты гоняются на прод-схеме.
	migrations := []string{
		"../../../migrations/000001_create_habits_table.up.sql",
		"../../../migrations/000002_create_habit_completions_table.up.sql",
	}
	for _, path := range migrations {
		sql, err := os.ReadFile(path)
		if err != nil {
			return 0, fmt.Errorf("read %s: %w", path, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return 0, fmt.Errorf("apply %s: %w", path, err)
		}
	}

	testPool = pool
	return m.Run(), nil
}