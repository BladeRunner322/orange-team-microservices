package usecases

import (
	"context"
	"errors"

	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/domain"
)

type RefreshLeaderboard struct {
	source ports.WorkoutsClientInterface
	store  ports.Repository
	limit  int32
}

func NewRefreshLeaderboard(source ports.WorkoutsClientInterface, store ports.Repository, limit int32) *RefreshLeaderboard {
	return &RefreshLeaderboard{source: source, store: store, limit: limit}
}

// Execute обновляет топ за текущий период по данным Workouts.
func (uc *RefreshLeaderboard) Execute(ctx context.Context, kind domain.PeriodKind) error {
	// TODO: Запросить очки за период, присвоить места и сохранить топ после успешного ответа.
	return errors.ErrUnsupported
}
