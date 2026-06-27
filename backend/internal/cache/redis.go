// Package cache wraps Redis. It provides the cache client used by the
// cache-aside helpers (added in the catalog phase) and a health ping.
// A dedicated DB index is used for cache so it can be flushed independently
// of the Asynq queue / pub-sub data (plan §3.4).
package cache

import (
	"context"
	"time"

	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/redis/go-redis/v9"
)

// New opens a Redis client against the cache DB index. Timeouts are bounded so
// a Redis blip degrades quickly (e.g. fast /readyz failure) instead of having
// requests hang on connection retries.
func New(cfg config.Redis) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.CacheDB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		MaxRetries:   1,
		PoolSize:     20,
	})
}

// Ping verifies Redis is reachable within the context deadline.
func Ping(ctx context.Context, rdb *redis.Client) error {
	return rdb.Ping(ctx).Err()
}
