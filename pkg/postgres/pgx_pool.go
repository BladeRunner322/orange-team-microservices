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

// BeginTx открывает транзакцию.
func (p *PgxPool) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, mapErrors(err)
	}
	return NewTx(tx), nil
}

// WithTx выполняет функцию fn в транзакции.
//
// Если fn возвращает ошибку — транзакция откатывается и ошибка возвращается наверх.
// Если fn вернула nil — транзакция коммитится.
func (p *PgxPool) WithTx(ctx context.Context, fn func(Tx) error) error {
	tx, err := p.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// В случае паники — откатываем, чтобы не оставить транзакцию открытой.
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback tx after error '%w': %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// OpTimeout возвращает таймаут операций.
func (p *PgxPool) OpTimeout() time.Duration {
	return p.opTimeout
}

// Close закрывает пул.
func (p *PgxPool) Close() {
	p.Pool.Close()
}
