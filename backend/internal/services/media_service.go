package services

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/mrkaak/restaurant-api/internal/media"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/storage"
	"github.com/mrkaak/restaurant-api/pkg/pagination"
)

// mediaRepo is the minimal interface MediaService needs from the repository.
type mediaRepo interface {
	Create(ctx context.Context, m *models.Media) error
	List(ctx context.Context, p pagination.Params, q string) ([]models.Media, int64, error)
	GetByID(ctx context.Context, id uint64) (*models.Media, error)
	SoftDelete(ctx context.Context, id uint64) error
}

type MediaService struct {
	repo        mediaRepo
	store       storage.Storage
	maxBytes    int64
	allowedMIME []string
}

func NewMediaService(repo mediaRepo, store storage.Storage, maxBytes int64, allowedMIME []string) *MediaService {
	return &MediaService{
		repo:        repo,
		store:       store,
		maxBytes:    maxBytes,
		allowedMIME: allowedMIME,
	}
}

// UploadInput carries the raw upload data from the handler.
type UploadInput struct {
	Data         []byte
	OriginalName string
	UserID       *uint64
}

// MediaItem is the public DTO returned to callers.
type MediaItem struct {
	ID           uint64    `json:"id"`
	URL          string    `json:"url"`
	ThumbURL     string    `json:"thumb_url"`
	OriginalName string    `json:"original_name"`
	MIME         string    `json:"mime"`
	SizeBytes    int64     `json:"size_bytes"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Alt          *string   `json:"alt"`
	CreatedAt    time.Time `json:"created_at"`
}

var ErrMediaTooLarge    = fmt.Errorf("file exceeds the maximum allowed size")
var ErrMediaInvalidType = fmt.Errorf("file type is not allowed")

func (s *MediaService) Upload(ctx context.Context, in UploadInput) (*MediaItem, error) {
	if int64(len(in.Data)) > s.maxBytes {
		return nil, ErrMediaTooLarge
	}

	// Sniff the real MIME type (not the browser-supplied Content-Type).
	mt := mimetype.Detect(in.Data)
	mime := mt.String()
	// Normalise e.g. "image/jpeg; charset=..." → "image/jpeg"
	if i := strings.Index(mime, ";"); i != -1 {
		mime = strings.TrimSpace(mime[:i])
	}

	if !slices.Contains(s.allowedMIME, mime) {
		return nil, ErrMediaInvalidType
	}

	// Decode, resize, and re-encode as JPEG.
	processed, err := media.Process(bytes.NewReader(in.Data))
	if err != nil {
		return nil, fmt.Errorf("image processing failed: %w", err)
	}

	// Build a collision-resistant filename.
	base := sanitiseFilename(strings.TrimSuffix(in.OriginalName, path.Ext(in.OriginalName)))
	key := fmt.Sprintf("%s_%d.jpg", base, time.Now().UnixNano())
	thumbKey := "thumb/" + key

	url, err := s.store.Save(ctx, key, bytes.NewReader(processed.OriginalBytes), "image/jpeg")
	if err != nil {
		return nil, fmt.Errorf("store original: %w", err)
	}
	thumbURL, err := s.store.Save(ctx, thumbKey, bytes.NewReader(processed.ThumbBytes), "image/jpeg")
	if err != nil {
		_ = s.store.Delete(ctx, key)
		return nil, fmt.Errorf("store thumbnail: %w", err)
	}

	m := &models.Media{
		Filename:     key,
		OriginalName: in.OriginalName,
		MIME:         mime,
		Ext:          ".jpg",
		SizeBytes:    int64(len(processed.OriginalBytes)),
		Width:        processed.Width,
		Height:       processed.Height,
		URL:          url,
		ThumbURL:     thumbURL,
		CreatedBy:    in.UserID,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		_ = s.store.Delete(ctx, key)
		_ = s.store.Delete(ctx, thumbKey)
		return nil, err
	}

	return toMediaItem(m), nil
}

func (s *MediaService) List(ctx context.Context, p pagination.Params, q string) ([]MediaItem, int64, error) {
	rows, total, err := s.repo.List(ctx, p, q)
	if err != nil {
		return nil, 0, err
	}
	out := make([]MediaItem, len(rows))
	for i := range rows {
		out[i] = *toMediaItem(&rows[i])
	}
	return out, total, nil
}

func (s *MediaService) Delete(ctx context.Context, id uint64) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	// Best-effort file removal; soft-delete row is the source of truth.
	_ = s.store.Delete(ctx, m.Filename)
	_ = s.store.Delete(ctx, "thumb/"+m.Filename)
	return nil
}

func toMediaItem(m *models.Media) *MediaItem {
	return &MediaItem{
		ID:           m.ID,
		URL:          m.URL,
		ThumbURL:     m.ThumbURL,
		OriginalName: m.OriginalName,
		MIME:         m.MIME,
		SizeBytes:    m.SizeBytes,
		Width:        m.Width,
		Height:       m.Height,
		Alt:          m.Alt,
		CreatedAt:    m.CreatedAt,
	}
}

// sanitiseFilename removes characters that could break storage paths.
func sanitiseFilename(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	s := b.String()
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		s = "upload"
	}
	return s
}
