package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/payment"
	"github.com/mrkaak/restaurant-api/internal/repository"
)

// PaymentService coordinates payment providers with order + transaction state.
// Providers are looked up by key; availability is settings-driven (plan §10.3).
type PaymentService struct {
	providers map[string]payment.Provider
	orders    *repository.OrderRepo
	payments  *repository.PaymentRepo
	settings  *repository.SettingsRepo
	currency  string
}

func NewPaymentService(orders *repository.OrderRepo, payments *repository.PaymentRepo, settings *repository.SettingsRepo, currency string, providers ...payment.Provider) *PaymentService {
	m := make(map[string]payment.Provider, len(providers))
	for _, p := range providers {
		m[p.Key()] = p
	}
	return &PaymentService{providers: m, orders: orders, payments: payments, settings: settings, currency: currency}
}

// AvailableMethods returns the payment methods that are both enabled and
// configured — exactly what the checkout should offer.
func (s *PaymentService) AvailableMethods(ctx context.Context) []string {
	var out []string
	for _, key := range []string{"cod", "square"} { // stable display order
		p, ok := s.providers[key]
		if ok && p.Enabled(ctx, s.settings) && p.Configured(ctx, s.settings) {
			out = append(out, key)
		}
	}
	return out
}

// ConfirmCardPayment finalizes a Square (card) order: charges the tokenized
// card, records the transaction, and on success advances the order to paid →
// confirmed. This is the SDK-callback path; the webhook is the backstop.
func (s *PaymentService) ConfirmCardPayment(ctx context.Context, userID uint64, orderCode, sourceToken, idempotencyKey string) (*models.Order, error) {
	o, err := s.orders.FindByCode(ctx, orderCode)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if o.UserID != userID {
		return nil, ErrForbidden
	}
	if o.PaymentStatus == models.PayPaid {
		return o, nil // already paid (idempotent)
	}
	prov, ok := s.providers["square"]
	if !ok || !prov.Enabled(ctx, s.settings) || !prov.Configured(ctx, s.settings) {
		return nil, ErrPaymentMethodDisabled
	}

	res, chargeErr := prov.Charge(ctx, s.settings, payment.ChargeRequest{
		OrderID: o.ID, OrderCode: o.Code, AmountCents: o.TotalCents,
		Currency: o.Currency, SourceToken: sourceToken, IdempotencyKey: idempotencyKey,
	})

	txn := &models.PaymentTransaction{
		OrderID: o.ID, Provider: "square", ProviderRef: res.ProviderRef,
		Kind: models.TxnCharge, AmountCents: o.TotalCents, Currency: o.Currency,
		Status: models.TxnSucceeded, RawPayload: json.RawMessage(res.RawPayload),
	}
	if chargeErr != nil || !res.Succeeded {
		txn.Status = models.TxnFailed
		_ = s.payments.Create(ctx, txn)
		_ = s.orders.SetPaymentStatus(ctx, o.ID, models.PayFailed)
		return nil, ErrPaymentFailed
	}
	if err := s.payments.Create(ctx, txn); err != nil {
		return nil, err
	}
	_ = s.orders.SetPaymentStatus(ctx, o.ID, models.PayPaid)
	// pending_payment -> paid -> confirmed
	if o.Status == models.StatusPendingPayment {
		_ = s.orders.AdvanceStatus(ctx, o, models.StatusPaid, nil, "card payment captured")
		_ = s.orders.AdvanceStatus(ctx, o, models.StatusConfirmed, nil, "order confirmed")
	}
	return s.orders.FindByCode(ctx, orderCode)
}

// HandleSquareWebhook verifies and processes a Square webhook idempotently. It
// is the source of truth: it confirms the order even if the SDK callback was
// lost (plan §10.1).
func (s *PaymentService) HandleSquareWebhook(ctx context.Context, payload []byte, signature, requestURL string) error {
	prov, ok := s.providers["square"]
	if !ok {
		return ErrPaymentMethodDisabled
	}
	evt, err := prov.VerifyAndParseWebhook(ctx, s.settings, payload, signature, requestURL)
	if err != nil {
		return ErrWebhookInvalid
	}
	if evt.ProviderRef == "" {
		return nil // nothing actionable
	}

	// Idempotency: skip events we've already recorded.
	seen, err := s.payments.ExistsByRef(ctx, "square", evt.ProviderRef)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}
	if evt.OrderCode == "" {
		return nil
	}
	o, err := s.orders.FindByCode(ctx, evt.OrderCode)
	if err != nil {
		return nil // unknown order; ack to stop retries
	}

	switch {
	case evt.Paid:
		_ = s.payments.Create(ctx, &models.PaymentTransaction{
			OrderID: o.ID, Provider: "square", ProviderRef: evt.ProviderRef,
			Kind: models.TxnCharge, AmountCents: o.TotalCents, Currency: o.Currency,
			Status: models.TxnSucceeded, RawPayload: json.RawMessage(evt.RawPayload),
		})
		if o.PaymentStatus != models.PayPaid {
			_ = s.orders.SetPaymentStatus(ctx, o.ID, models.PayPaid)
			if o.Status == models.StatusPendingPayment {
				_ = s.orders.AdvanceStatus(ctx, o, models.StatusPaid, nil, "payment confirmed (webhook)")
				_ = s.orders.AdvanceStatus(ctx, o, models.StatusConfirmed, nil, "order confirmed")
			}
		}
	case evt.Refunded:
		_ = s.payments.Create(ctx, &models.PaymentTransaction{
			OrderID: o.ID, Provider: "square", ProviderRef: evt.ProviderRef,
			Kind: models.TxnRefund, Currency: o.Currency,
			Status: models.TxnRefunded, RawPayload: json.RawMessage(evt.RawPayload),
		})
		_ = s.orders.SetPaymentStatus(ctx, o.ID, models.PayRefunded)
	}
	return nil
}

// Refund issues a refund for an order's latest succeeded charge (admin action).
func (s *PaymentService) Refund(ctx context.Context, orderID uint64) error {
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return mapNotFound(err)
	}
	charge, err := s.payments.LatestCharge(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNothingToRefund
		}
		return err
	}
	prov, ok := s.providers[o.PaymentMethod]
	if !ok {
		return ErrPaymentMethodDisabled
	}
	res, err := prov.Refund(ctx, s.settings, charge.ProviderRef, o.TotalCents, o.Currency)
	if err != nil || !res.Succeeded {
		return ErrRefundFailed
	}
	_ = s.payments.Create(ctx, &models.PaymentTransaction{
		OrderID: o.ID, Provider: o.PaymentMethod, ProviderRef: res.ProviderRef,
		Kind: models.TxnRefund, AmountCents: o.TotalCents, Currency: o.Currency,
		Status: models.TxnRefunded, RawPayload: json.RawMessage(res.RawPayload),
	})
	_ = s.orders.SetPaymentStatus(ctx, o.ID, models.PayRefunded)
	if o.Status.CanTransitionTo(models.StatusRefunded) {
		_ = s.orders.AdvanceStatus(ctx, o, models.StatusRefunded, nil, "refunded by staff")
	}
	return nil
}
