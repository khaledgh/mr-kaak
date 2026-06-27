// Package payment defines a pluggable payment-provider abstraction (plan §10).
// Each provider is gated by a settings flag so an admin can enable/disable a
// method without a deploy. Adding a future gateway is a new Provider
// implementation plus one settings flag — no checkout rewrite.
package payment

import "context"

// Settings is the subset of app settings a provider needs. The concrete
// settings repo satisfies this via small adapters in the service layer.
type Settings interface {
	GetBool(ctx context.Context, key string, def bool) bool
	GetString(ctx context.Context, key, def string) string
}

// ChargeRequest is the input to create a payment for an order.
type ChargeRequest struct {
	OrderID        uint64
	OrderCode      string
	AmountCents    int64
	Currency       string
	SourceToken    string // single-use card token from the Web Payments SDK
	IdempotencyKey string
}

// ChargeResult is the outcome of a charge attempt.
type ChargeResult struct {
	ProviderRef string // provider payment id
	Succeeded   bool
	RawPayload  []byte
}

// WebhookEvent is a normalized provider webhook outcome.
type WebhookEvent struct {
	ProviderRef string
	OrderCode   string
	Paid        bool
	Refunded    bool
	RawPayload  []byte
}

// Provider is implemented by each payment method.
type Provider interface {
	// Key is the stable method identifier ("square", "cod").
	Key() string
	// Enabled reports whether the method is turned on in settings.
	Enabled(ctx context.Context, s Settings) bool
	// Configured reports whether the required credentials are present.
	Configured(ctx context.Context, s Settings) bool
	// Charge attempts a payment. COD is a no-op success.
	Charge(ctx context.Context, s Settings, req ChargeRequest) (ChargeResult, error)
	// VerifyAndParseWebhook validates the signature and normalizes the event.
	VerifyAndParseWebhook(ctx context.Context, s Settings, payload []byte, signature, requestURL string) (WebhookEvent, error)
	// Refund reverses a charge (amount in cents).
	Refund(ctx context.Context, s Settings, providerRef string, amountCents int64, currency string) (ChargeResult, error)
}
