package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// NewRedis creates and verifies a Redis client connection.
// addr is in "host:port" format; password may be empty string if auth is disabled.
func NewRedis(addr, password string) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis: failed to ping server at %s — check address and credentials: %w", addr, err)
	}

	return client, nil
}
