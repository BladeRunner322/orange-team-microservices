package health

import (
	"context"

	"github.com/BladeRunner322/orange-team-microservices/pkg/postgres"
	"github.com/BladeRunner322/orange-team-microservices/pkg/redis"
)

// PostgresCheck возвращает Check для пула Postgres.
func PostgresCheck(pool *postgres.PgxPool) Check {
	return func(ctx context.Context) error {
		return pool.Ping(ctx)
	}
}

// RedisCheck возвращает Check для клиента Redis.
func RedisCheck(client *redis.Client) Check {
	return func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	}
}
