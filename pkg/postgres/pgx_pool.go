package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxPool — реализация интерфейса Pool на основе pgxpool.
type PgxPool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

// NewPgxPool создаёт новый пул соединений с PostgreSQL.
func NewPgxPool(ctx context.Context, cfg Config) (*PgxPool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	pgxCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping pgxpool: %w", err)
	}

	return &PgxPool{
		Pool:      pool,
		opTimeout: cfg.Timeout,
	}, nil
}

// Query выполняет запрос и возвращает строки.
func (p *PgxPool) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}
	return pgxRows{rows}, nil
}

// QueryRow выполняет запрос и возвращает одну строку.
func (p *PgxPool) QueryRow(ctx context.Context, sql string, args ...any) Row {
	row := p.Pool.QueryRow(ctx, sql, args...)
	return pgxRow{row}
}

// Exec выполняет запрос и возвращает команду-тег.
func (p *PgxPool) Exec(ctx context.Context, sql string, args ...any) (CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}
	return pgxCommandTag{tag}, nil
}

// OpTimeout возвращает таймаут операций.
func (p *PgxPool) OpTimeout() time.Duration {
	return p.opTimeout
}

// Close закрывает пул.
func (p *PgxPool) Close() {
	p.Pool.Close()
}
