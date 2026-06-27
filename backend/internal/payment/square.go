package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// squareVersion pins the Square API version (plan §10.1). Bump deliberately
// after testing against a newer version.
const squareVersion = "2026-05-20"

// Square implements the Provider interface against Square's Payments API. It
// uses net/http directly (no SDK dependency). All credentials come from
// settings so they can be rotated without a deploy.
type Square struct {
	client *http.Client
}

func NewSquare() *Square {
	return &Square{client: &http.Client{Timeout: 15 * time.Second}}
}

func (sq *Square) Key() string { return "square" }

func (sq *Square) Enabled(ctx context.Context, s Settings) bool {
	return s.GetBool(ctx, "square_enabled", false)
}

// Configured requires an access token + location id.
func (sq *Square) Configured(ctx context.Context, s Settings) bool {
	return s.GetString(ctx, "square_access_token", "") != "" &&
		s.GetString(ctx, "square_location_id", "") != ""
}

func (sq *Square) baseURL(ctx context.Context, s Settings) string {
	if s.GetString(ctx, "square_environment", "sandbox") == "production" {
		return "https://connect.squareup.com"
	}
	return "https://connect.squareupsandbox.com"
}

// Charge calls POST /v2/payments to capture the tokenized card. The amount is
// sent in CAD cents with an idempotency key so a retry never double-charges.
func (sq *Square) Charge(ctx context.Context, s Settings, req ChargeRequest) (ChargeResult, error) {
	body := map[string]any{
		"source_id":       req.SourceToken,
		"idempotency_key": req.IdempotencyKey,
		"amount_money":    map[string]any{"amount": req.AmountCents, "currency": req.Currency},
		"location_id":     s.GetString(ctx, "square_location_id", ""),
		"reference_id":    req.OrderCode,
		"autocomplete":    true,
	}
	raw, err := sq.do(ctx, s, http.MethodPost, "/v2/payments", body)
	if err != nil {
		return ChargeResult{RawPayload: raw}, err
	}

	var resp struct {
		Payment struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"payment"`
	}
	_ = json.Unmarshal(raw, &resp)
	return ChargeResult{
		ProviderRef: resp.Payment.ID,
		Succeeded:   resp.Payment.Status == "COMPLETED" || resp.Payment.Status == "APPROVED",
		RawPayload:  raw,
	}, nil
}

// Refund calls POST /v2/refunds against a prior payment.
func (sq *Square) Refund(ctx context.Context, s Settings, providerRef string, amountCents int64, currency string) (ChargeResult, error) {
	body := map[string]any{
		"idempotency_key": fmt.Sprintf("refund-%s-%d", providerRef, time.Now().UnixNano()),
		"payment_id":      providerRef,
		"amount_money":    map[string]any{"amount": amountCents, "currency": currency},
	}
	raw, err := sq.do(ctx, s, http.MethodPost, "/v2/refunds", body)
	if err != nil {
		return ChargeResult{RawPayload: raw}, err
	}
	var resp struct {
		Refund struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"refund"`
	}
	_ = json.Unmarshal(raw, &resp)
	return ChargeResult{
		ProviderRef: resp.Refund.ID,
		Succeeded:   resp.Refund.Status == "COMPLETED" || resp.Refund.Status == "PENDING",
		RawPayload:  raw,
	}, nil
}

// VerifyAndParseWebhook validates Square's HMAC-SHA256 signature (over
// requestURL + body, keyed by the webhook signature key) and extracts the
// payment outcome. Never trust a webhook without verifying the signature.
func (sq *Square) VerifyAndParseWebhook(ctx context.Context, s Settings, payload []byte, signature, requestURL string) (WebhookEvent, error) {
	key := s.GetString(ctx, "square_webhook_signature_key", "")
	if key == "" {
		return WebhookEvent{}, fmt.Errorf("square webhook signature key not configured")
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(requestURL))
	mac.Write(payload)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return WebhookEvent{}, fmt.Errorf("square webhook signature mismatch")
	}

	var evt struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				Payment struct {
					ID          string `json:"id"`
					Status      string `json:"status"`
					ReferenceID string `json:"reference_id"`
				} `json:"payment"`
				Refund struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"refund"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &evt); err != nil {
		return WebhookEvent{}, fmt.Errorf("parse webhook: %w", err)
	}

	out := WebhookEvent{RawPayload: payload, OrderCode: evt.Data.Object.Payment.ReferenceID}
	switch evt.Type {
	case "payment.updated", "payment.created":
		out.ProviderRef = evt.Data.Object.Payment.ID
		out.Paid = evt.Data.Object.Payment.Status == "COMPLETED"
	case "refund.updated", "refund.created":
		out.ProviderRef = evt.Data.Object.Refund.ID
		out.Refunded = evt.Data.Object.Refund.Status == "COMPLETED"
	}
	return out, nil
}

// do performs an authenticated Square API call and returns the raw body.
func (sq *Square) do(ctx context.Context, s Settings, method, path string, body any) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, sq.baseURL(ctx, s)+path, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.GetString(ctx, "square_access_token", ""))
	req.Header.Set("Square-Version", squareVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := sq.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return raw, fmt.Errorf("square api %s %s: status %d", method, path, resp.StatusCode)
	}
	return raw, nil
}
