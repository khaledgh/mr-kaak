// Package storage provides a pluggable abstraction for persisting uploaded files.
// LocalStorage (this package) writes to disk; an S3Storage can be dropped in
// later without touching handler or service code.
package storage

import (
	"context"
	"io"
)

// Storage saves and removes binary assets keyed by a caller-chosen path (e.g.
// "abc123.jpg" or "thumb/abc123.jpg"). Implementations must return an absolute,
// publicly reachable URL so callers can store it directly in the DB.
type Storage interface {
	Save(ctx context.Context, key string, r io.Reader, mime string) (publicURL string, err error)
	Delete(ctx context.Context, key string) error
}
