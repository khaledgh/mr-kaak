package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/mrkaak/restaurant-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Cache is a cache-aside wrapper around Redis. Every read goes through
// GetOrSet; every write goes through an explicit invalidator (see keys.go).
//
// Graceful degradation: if Redis is unavailable, reads fall through to the
// loader (DB) and writes/deletes become no-ops. The app stays fully functional
// without Redis — only the speed-up is lost. This matters because Redis may not
// be running in development yet.
type Cache struct {
	rdb *redis.Client
}

// NewCache wraps a Redis client with cache-aside helpers.
func NewCache(rdb *redis.Client) *Cache { return &Cache{rdb: rdb} }

// GetOrSet returns the cached value for key, or loads it via loader, caches it
// (best-effort), and returns it. A Redis outage degrades to calling loader
// directly. A cache *miss* is normal; a cache *error* is logged once.
func GetOrSet[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, loader func(context.Context) (T, error)) (T, error) {
	var zero T

	raw, err := c.rdb.Get(ctx, key).Bytes()
	switch {
	case err == nil:
		var v T
		if jerr := json.Unmarshal(raw, &v); jerr == nil {
			return v, nil
		}
		// Corrupt entry — drop it and reload.
		_ = c.rdb.Del(ctx, key).Err()
	case errors.Is(err, redis.Nil):
		// Normal miss; fall through to loader.
	default:
		// Redis error (e.g. down): degrade silently to the DB loader.
		logger.FromContext(ctx).Warn("cache get failed; loading from source",
			slog.String("key", key), slog.Any("err", err))
		return loader(ctx)
	}

	v, err := loader(ctx)
	if err != nil {
		return zero, err
	}

	if buf, merr := json.Marshal(v); merr == nil {
		if serr := c.rdb.Set(ctx, key, buf, ttl).Err(); serr != nil {
			logger.FromContext(ctx).Warn("cache set failed",
				slog.String("key", key), slog.Any("err", serr))
		}
	}
	return v, nil
}

// Del removes keys (best-effort). Used by surgical invalidators.
func (c *Cache) Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		logger.FromContext(ctx).Warn("cache del failed", slog.Any("err", err))
	}
}

// Incr bumps a counter and returns the new value (used for versioned
// namespaces — see Keys.MenuVersion). Returns 0 on Redis error so callers can
// still operate (the version simply won't advance until Redis is back).
func (c *Cache) Incr(ctx context.Context, key string) int64 {
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		logger.FromContext(ctx).Warn("cache incr failed", slog.String("key", key), slog.Any("err", err))
		return 0
	}
	return n
}

// Version reads a namespace version counter, treating "missing" and "Redis
// down" both as version 0.
func (c *Cache) Version(ctx context.Context, key string) int64 {
	n, err := c.rdb.Get(ctx, key).Int64()
	if err != nil {
		return 0
	}
	return n
}

// FlushAll clears the entire cache DB index. Safe because cache uses a
// dedicated DB index (see config Redis.CacheDB), so this never touches the
// Asynq queue / pub-sub data.
func (c *Cache) FlushAll(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}
