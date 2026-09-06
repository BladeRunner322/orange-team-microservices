package postgres

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
	"github.com/google/uuid"
)

type InMemoryRepository struct {
	users map[string]domain.User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{users: make(map[string]domain.User)}
}

func (r *InMemoryRepository) Save(ctx context.Context, user domain.User) error {
	r.users[user.Email.String()] = user
	return nil
}

func (r *InMemoryRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	user, ok := r.users[email.String()]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}
