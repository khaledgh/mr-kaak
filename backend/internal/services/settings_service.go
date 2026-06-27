package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/repository"
)

// SettingsService exposes the admin settings editor. Secret values are masked
// on read so credentials are never echoed back to the dashboard.
type SettingsService struct {
	settings *repository.SettingsRepo
}

func NewSettingsService(settings *repository.SettingsRepo) *SettingsService {
	return &SettingsService{settings: settings}
}

// secretKey reports whether a setting holds a credential that must be masked.
func secretKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "token") || strings.Contains(k, "secret") ||
		strings.HasSuffix(k, "signature_key") || strings.Contains(k, "password")
}

// All returns every setting, masking secrets to a boolean "is set" indicator.
func (s *SettingsService) All(ctx context.Context) (map[string]any, error) {
	raw, err := s.settings.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(raw))
	for k, v := range raw {
		if secretKey(k) {
			var str string
			_ = json.Unmarshal(v, &str)
			out[k] = map[string]bool{"is_set": str != ""}
			continue
		}
		out[k] = json.RawMessage(v)
	}
	return out, nil
}

// Update upserts the provided settings. Each value is stored as-is (JSON typed).
func (s *SettingsService) Update(ctx context.Context, values map[string]json.RawMessage) error {
	for k, v := range values {
		var decoded any
		if err := json.Unmarshal(v, &decoded); err != nil {
			return err
		}
		if err := s.settings.Set(ctx, k, decoded); err != nil {
			return err
		}
	}
	return nil
}
