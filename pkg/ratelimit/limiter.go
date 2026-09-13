package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Rule — параметры Token Bucket.
type Rule struct {
	// Rate — сколько токенов восстанавливается за Interval.
	Rate int

	// Burst — максимальное количество токенов в ведре (capacity).
	Burst int

	// Interval — за какой период восстанавливается Rate токенов.
	Interval time.Duration
}

// Limiter — Token Bucket на Redis.
type Limiter struct {
	client *redis.Client
	script *redis.Script
}

// NewLimiter создаёт новый limiter.
func NewLimiter(client *redis.Client) *Limiter {
	return &Limiter{
		client: client,
		script: redis.NewScript(tokenBucketLua),
	}
}

// Allow проверяет, можно ли пропустить запрос с данным ключом.
//
// Если токенов достаточно — возвращает Result{Allowed: true} и списывает 1 токен.
// Если токенов нет — возвращает Result{Allowed: false, RetryAfter: N}.
func (l *Limiter) Allow(ctx context.Context, key string, rule Rule) (Result, error) {
	if rule.Rate <= 0 || rule.Burst <= 0 || rule.Interval <= 0 {
		return Result{}, fmt.Errorf("invalid rule: rate=%d burst=%d interval=%s",
			rule.Rate, rule.Burst, rule.Interval)
	}

	// rate в токенах/секунду.
	refillRate := float64(rule.Rate) / rule.Interval.Seconds()

	// now в секундах с миллисекундной точностью.
	now := float64(time.Now().UnixNano()) / 1e9

	// TTL — время полного восстановления ведра + 1 секунда запаса.
	ttl := int(rule.Interval.Seconds()) + 1

	res, err := l.script.Run(ctx, l.client, []string{key},
		refillRate,
		rule.Burst,
		now,
		ttl,
	).Slice()
	if err != nil {
		return Result{}, fmt.Errorf("run rate limit script: %w", err)
	}

	allowed := res[0].(int64) == 1
	retryAfterSec := res[1].(int64)

	return Result{
		Allowed:    allowed,
		RetryAfter: time.Duration(retryAfterSec) * time.Second,
	}, nil
}

// tokenBucketLua — Lua-скрипт для атомарного Token Bucket.
//
// KEYS[1] — ключ (например, "rl:login:<ip>:<email>")
// ARGV[1] — refill rate (токенов/секунду, float)
// ARGV[2] — burst (capacity, int)
// ARGV[3] — now (unix seconds, float)
// ARGV[4] — TTL ключа в секундах (int)
//
// Возвращает: { allowed (0/1), retry_after_seconds (int) }
const tokenBucketLua = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens = burst
    last_refill = now
end

local elapsed = now - last_refill
if elapsed < 0 then elapsed = 0 end

-- Пополняем ведро.
tokens = math.min(burst, tokens + elapsed * rate)

local allowed = 0
local retry_after = 0

if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
else
    -- Сколько секунд до восстановления 1 токена.
    retry_after = math.ceil((1 - tokens) / rate)
end

redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
redis.call('EXPIRE', key, ttl)

return { allowed, retry_after }
`
