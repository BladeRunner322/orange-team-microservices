package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Client — обёртка над go-redis клиентом.
type Client struct {
	*goredis.Client
}

// NewClient создаёт и проверяет соединение с Redis.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{Client: client}, nil
}

// Close закрывает соединение.
func (c *Client) Close() error {
	return c.Client.Close()
}
