package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage persists files under UploadDir and returns absolute URLs built
// from BaseURL. Directory structure mirrors the key (e.g. "thumb/abc.jpg"
// writes to <UploadDir>/thumb/abc.jpg).
type LocalStorage struct {
	UploadDir string // absolute or relative path to the local upload root
	BaseURL   string // public base URL without trailing slash
}

// NewLocal returns a LocalStorage after ensuring UploadDir exists.
func NewLocal(uploadDir, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: create upload dir: %w", err)
	}
	return &LocalStorage{
		UploadDir: uploadDir,
		BaseURL:   strings.TrimRight(baseURL, "/"),
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
	return s.BaseURL + "/uploads/" + key, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	dst := filepath.Join(s.UploadDir, filepath.FromSlash(key))
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete: %w", err)
	}
	return nil
}
