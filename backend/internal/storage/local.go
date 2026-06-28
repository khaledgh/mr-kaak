package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage persists files under UploadDir and returns relative paths
// (e.g. "/uploads/abc.jpg"). The public base URL is prepended at the service
// layer so it is never baked into the database.
type LocalStorage struct {
	UploadDir string // absolute or relative path to the local upload root
}

// NewLocal returns a LocalStorage after ensuring UploadDir exists.
func NewLocal(uploadDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: create upload dir: %w", err)
	}
	return &LocalStorage{
		UploadDir: uploadDir,
	}, nil
}

func (s *LocalStorage) Save(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	dst := filepath.Join(s.UploadDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("storage: mkdir: %w", err)
	}
	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("storage: create file: %w", err)
	}
	defer f.Close()
	if _, err = io.Copy(f, r); err != nil {
		return "", fmt.Errorf("storage: write file: %w", err)
	}
	return "/uploads/" + key, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	dst := filepath.Join(s.UploadDir, filepath.FromSlash(key))
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete: %w", err)
	}
	return nil
}
