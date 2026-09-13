package ratelimit

import (
	"errors"
	"time"
)

// ErrRateLimited возвращается, когда лимит превышен.
var ErrRateLimited = errors.New("rate limit exceeded")

// Result — результат проверки лимита.
type Result struct {
	// Allowed — можно ли пропустить запрос.
	Allowed bool

	// RetryAfter — сколько ждать до следующей попытки.
	// Имеет смысл только если Allowed == false.
	RetryAfter time.Duration
}
