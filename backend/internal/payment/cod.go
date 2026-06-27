package payment

import (
	"context"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/models"
)

// COD is the cash-on-delivery provider: a no-op charge gated by cod_enabled.
// The order is settled by staff on delivery (plan §10.2).
type COD struct{}

func NewCOD() *COD { return &COD{} }

func (c *COD) Key() string { return "cod" }

func (c *COD) Enabled(ctx context.Context, s Settings) bool {
	return s.GetBool(ctx, models.SettingCODEnabled, true)
}

// Configured: COD needs no credentials.
func (c *COD) Configured(ctx context.Context, s Settings) bool { return true }

// Charge is a no-op: there's nothing to capture online.
func (c *COD) Charge(ctx context.Context, s Settings, req ChargeRequest) (ChargeResult, error) {
	return ChargeResult{ProviderRef: "cod:" + req.OrderCode, Succeeded: true}, nil
}

func (c *COD) VerifyAndParseWebhook(ctx context.Context, s Settings, payload []byte, signature, requestURL string) (WebhookEvent, error) {
	return WebhookEvent{}, errors.New("cod has no webhooks")
}

func (c *COD) Refund(ctx context.Context, s Settings, providerRef string, amountCents int64, currency string) (ChargeResult, error) {
	// A COD refund is handled out-of-band (cash returned); record-only.
	return ChargeResult{ProviderRef: providerRef, Succeeded: true}, nil
}
