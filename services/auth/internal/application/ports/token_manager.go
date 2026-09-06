package ports

import "context"

type TokenManager interface {
	Generate(ctx context.Context, userID string) (string, error)
	Validate(ctx context.Context, token string) (string, error)
}
