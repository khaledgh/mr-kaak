package cache

import (
	"context"
	"testing"
	"time"

	"github.com/mrkaak/restaurant-api/internal/config"
)

// TestGetOrSetDegradesWithoutRedis verifies the key resilience property: when
// Redis is unreachable, GetOrSet still returns the loader's value instead of
// erroring. This is what lets the app run without Redis in development.
func TestGetOrSetDegradesWithoutRedis(t *testing.T) {
	// Point at a port nothing is listening on.
	rdb := New(config.Redis{Addr: "127.0.0.1:6390", CacheDB: 0})
	defer func() { _ = rdb.Close() }()
	c := NewCache(rdb)

	calls := 0
	loader := func(ctx context.Context) (string, error) {
		calls++
		return "from-db", nil
	}

	ctx := context.Background()
	got, err := GetOrSet(ctx, c, "k", time.Minute, loader)
	if err != nil {
		t.Fatalf("expected graceful degradation, got error: %v", err)
	}
	if got != "from-db" {
		t.Errorf("value = %q, want from-db", got)
	}

	// With Redis down there is no caching, so a second call hits the loader again.
	if _, _ = GetOrSet(ctx, c, "k", time.Minute, loader); calls != 2 {
		t.Errorf("loader calls = %d, want 2 (no caching when Redis down)", calls)
	}
}
