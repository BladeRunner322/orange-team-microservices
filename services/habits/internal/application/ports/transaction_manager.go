package ports

import "context"

// TransactionManager выполняет функцию в транзакции.
type TransactionManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}
