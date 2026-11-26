package worker

import (
	"context"
	"errors"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/services/leaderboard/internal/application/usecases"
)

// RefreshWorker запускает обновление топов по расписанию.
type RefreshWorker struct {
	refresh  *usecases.RefreshLeaderboard
	interval time.Duration
}

func NewRefreshWorker(refresh *usecases.RefreshLeaderboard, interval time.Duration) *RefreshWorker {
	return &RefreshWorker{refresh: refresh, interval: interval}
}

// Run управляет циклом обновления до отмены контекста.
func (w *RefreshWorker) Run(ctx context.Context) error {
	// TODO: Обновить три топа при старте и по таймеру, остановить цикл при отмене контекста.
	return errors.ErrUnsupported
}
