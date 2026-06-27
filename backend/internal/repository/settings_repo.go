package repository

import (
	"context"
	"encoding/json"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// SettingsRepo reads/writes the key/value settings store. Values are stored as
// JSON so they stay typed (bool, number, string, object).
type SettingsRepo struct {
	db *gorm.DB
}

func NewSettingsRepo(db *gorm.DB) *SettingsRepo { return &SettingsRepo{db: db} }

// GetRaw returns the raw JSON bytes for a key, or (nil,false) if absent.
func (r *SettingsRepo) GetRaw(ctx context.Context, key string) (json.RawMessage, bool, error) {
	var s models.Setting
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return s.ValueJSON, true, nil
}

// GetBool returns a boolean setting with a default fallback.
func (r *SettingsRepo) GetBool(ctx context.Context, key string, def bool) bool {
	raw, ok, err := r.GetRaw(ctx, key)
	if err != nil || !ok {
		return def
	}
	var v bool
	if json.Unmarshal(raw, &v) != nil {
		return def
	}
	return v
}

// GetInt returns an integer setting with a default fallback.
func (r *SettingsRepo) GetInt(ctx context.Context, key string, def int64) int64 {
	raw, ok, err := r.GetRaw(ctx, key)
	if err != nil || !ok {
		return def
	}
	var v int64
	if json.Unmarshal(raw, &v) != nil {
		return def
	}
	return v
}

// GetString returns a string setting with a default fallback.
func (r *SettingsRepo) GetString(ctx context.Context, key, def string) string {
	raw, ok, err := r.GetRaw(ctx, key)
	if err != nil || !ok {
		return def
	}
	var v string
	if json.Unmarshal(raw, &v) != nil {
		return def
	}
	return v
}

// Set upserts a setting value (JSON-encoded).
func (r *SettingsRepo) Set(ctx context.Context, key string, value any) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO settings (`key`, value_json) VALUES (?, ?) ON DUPLICATE KEY UPDATE value_json = VALUES(value_json)",
		key, string(buf),
	).Error
}

// All returns every setting as a key -> raw JSON map (admin settings screen).
func (r *SettingsRepo) All(ctx context.Context) (map[string]json.RawMessage, error) {
	var rows []models.Setting
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]json.RawMessage, len(rows))
	for _, s := range rows {
		out[s.Key] = s.ValueJSON
	}
	return out, nil
}
