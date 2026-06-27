package cache

import (
	"context"
	"fmt"
)

// Centralized cache key registry. Keeping every key here makes invalidation
// auditable — you can see exactly which keys a mutation must clear (plan §3.4).

const (
	keyMenuVersion = "menu:version" // versioned-namespace counter for the menu
	keySettings    = "settings"
)

// TTLs per resource (plan §3.4).
const (
	TTLMenu     = 10 * 60 // seconds; menu changes rarely
	TTLCategory = 30 * 60
	TTLSettings = 30 * 60
)

// MenuKey returns the cache key for the active menu in a locale, embedding the
// current menu version. Bumping the version (BumpMenuVersion) orphans all old
// menu keys at once — an O(1) "flush the whole menu" with no key scanning.
func (c *Cache) MenuKey(ctx context.Context, locale string) string {
	return fmt.Sprintf("menu:v%d:active:%s", c.Version(ctx, keyMenuVersion), locale)
}

// BumpMenuVersion invalidates every cached menu (all locales) instantly.
func (c *Cache) BumpMenuVersion(ctx context.Context) {
	c.Incr(ctx, keyMenuVersion)
}

// SettingsKey is the cache key for the settings blob.
func (c *Cache) SettingsKey() string { return keySettings }
