package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email domain.Email) (domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}
