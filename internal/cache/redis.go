package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config for Redis connection
type Config struct {
	Addr     string
	Password string
	DB       int
}

// RedisSessionStore implements service.SessionStore using Redis.
type RedisSessionStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewSessionStore returns a RedisSessionStore that blacklists tokens for ttl duration.
func NewSessionStore(client *redis.Client, ttl time.Duration) *RedisSessionStore {
	return &RedisSessionStore{client: client, ttl: ttl}
}

func (r *RedisSessionStore) BlacklistToken(ctx context.Context, token string) error {
	return r.client.Set(ctx, token, "1", r.ttl).Err()
}

func (r *RedisSessionStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	n, err := r.client.Exists(ctx, token).Result()
	return n > 0, err
}
