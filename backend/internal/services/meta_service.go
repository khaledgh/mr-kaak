package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mrkaak/restaurant-api/internal/repository"
)

// MetaService implements the server-side Conversions API (plan §14). It is
// fully settings-driven: with no pixel id / token configured it is a no-op.
// PII is hashed (SHA-256) before sending, per Meta's requirements.
type MetaService struct {
	settings *repository.SettingsRepo
	users    *repository.UserRepo
	orders   *repository.OrderRepo
	http     *http.Client
}

func NewMetaService(settings *repository.SettingsRepo, users *repository.UserRepo, orders *repository.OrderRepo) *MetaService {
	return &MetaService{settings: settings, users: users, orders: orders, http: &http.Client{Timeout: 8 * time.Second}}
}

// Config returns the public Pixel config for the frontend (empty if unset).
func (s *MetaService) Config(ctx context.Context) map[string]any {
	pixel := s.settings.GetString(ctx, "meta_pixel_id", "")
	return map[string]any{"pixel_id": pixel, "enabled": pixel != ""}
}

// SendPurchase posts a server-side Purchase event for an order. Best-effort:
// returns nil when Meta isn't configured.
func (s *MetaService) SendPurchase(ctx context.Context, orderID uint64) error {
	pixel := s.settings.GetString(ctx, "meta_pixel_id", "")
	token := s.settings.GetString(ctx, "meta_capi_token", "")
	if pixel == "" || token == "" {
		return nil // not configured -> no-op
	}

	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	user, _ := s.users.FindByID(ctx, order.UserID)

	userData := map[string]any{}
	if user != nil {
		if user.Email != "" {
			userData["em"] = []string{sha256hex(user.Email)}
		}
		if user.PhoneE164 != "" {
			userData["ph"] = []string{sha256hex(strings.TrimPrefix(user.PhoneE164, "+"))}
		}
	}

	event := map[string]any{
		"event_name":    "Purchase",
		"event_time":    time.Now().Unix(),
		"action_source": "website",
		"event_id":      order.Code, // dedupes against the browser Pixel event
		"user_data":     userData,
		"custom_data": map[string]any{
			"currency": order.Currency,
			"value":    float64(order.TotalCents) / 100.0,
		},
	}
	body := map[string]any{"data": []any{event}}
	if tc := s.settings.GetString(ctx, "meta_test_event_code", ""); tc != "" {
		body["test_event_code"] = tc
	}

	buf, _ := json.Marshal(body)
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?access_token=%s", pixel, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("meta capi: status %d", resp.StatusCode)
	}
	return nil
}

// sha256hex lowercases, trims, and SHA-256 hashes PII per Meta's spec.
func sha256hex(v string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(v))))
	return hex.EncodeToString(sum[:])
}
