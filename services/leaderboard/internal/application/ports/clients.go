package ports

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/domain"
	"github.com/google/uuid"
)

// UserScore — сумма очков пользователя за период, полученная от Workouts.
type UserScore struct {
	UserID uuid.UUID
	Score  int64
}

// WorkoutsClientInterface возвращает очки участников в порядке рейтинга.
type WorkoutsClientInterface interface {
	GetTop(ctx context.Context, period domain.Period, limit int32) ([]UserScore, error)
}
