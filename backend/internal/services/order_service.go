package services

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log/slog"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/geo"
	"github.com/mrkaak/restaurant-api/internal/i18n"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

// OrderService orchestrates checkout: server-side re-pricing, delivery quoting,
// coupon application, tax, and the atomic order write (plan §9). It depends on
// several repos/services because checkout is the app's central integration point.
type OrderService struct {
	orders        *repository.OrderRepo
	catalog       *repository.CatalogRepo
	translations  *repository.TranslationRepo
	addresses     *repository.AddressRepo
	settings      *repository.SettingsRepo
	zones         *ZoneService
	coupons       *CouponService
	events        OrderEventPublisher
	notifier      OrderNotifier
	defaultLocale string
	currency      string
}

// OrderEventPublisher pushes live status events (SSE/Redis). Optional.
type OrderEventPublisher interface {
	PublishOrderEvent(ctx context.Context, orderCode, status, message string)
}

// OrderNotifier enqueues out-of-band notifications (push/email) and analytics.
// Optional. The Asynq enqueuer implements it.
type OrderNotifier interface {
	NotifyOrderStatus(ctx context.Context, orderID uint64, status string)
	TrackPurchase(ctx context.Context, orderID uint64)
}

// WithRealtime attaches the live-event publisher and notifier (wired in the
// server). Returns the service for chaining.
func (s *OrderService) WithRealtime(events OrderEventPublisher, notifier OrderNotifier) *OrderService {
	s.events, s.notifier = events, notifier
	return s
}

func NewOrderService(
	orders *repository.OrderRepo,
	catalog *repository.CatalogRepo,
	translations *repository.TranslationRepo,
	addresses *repository.AddressRepo,
	settings *repository.SettingsRepo,
	zones *ZoneService,
	coupons *CouponService,
	defaultLocale, currency string,
) *OrderService {
	return &OrderService{
		orders: orders, catalog: catalog, translations: translations,
		addresses: addresses, settings: settings, zones: zones, coupons: coupons,
		defaultLocale: defaultLocale, currency: currency,
	}
}

// Checkout validates and places an order. idempotencyKey (from the
// Idempotency-Key header) makes retries safe: a repeated request returns the
// already-created order instead of charging/creating twice.
func (s *OrderService) Checkout(ctx context.Context, userID uint64, idempotencyKey string, in CheckoutInput) (*models.Order, error) {
	// Idempotency fast-path: return the prior order if this key was seen.
	if idempotencyKey != "" {
		if existing, err := s.orders.FindByIdempotency(ctx, userID, idempotencyKey); err == nil {
			return existing, nil
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}

	// 1. Re-price each line from the DB.
	items, subtotal, productIDs, err := s.priceItems(ctx, in.Items)
	if err != nil {
		return nil, err
	}

	// 2. Fulfillment: delivery quote + address snapshot, or pickup.
	var (
		addrSnap    *models.AddressSnapshot
		deliveryFee int64
	)
	if in.FulfillmentType == string(models.FulfillmentDelivery) {
		snap, fee, qerr := s.quoteDelivery(ctx, userID, in.AddressID, productIDs, subtotal)
		if qerr != nil {
			return nil, qerr
		}
		addrSnap, deliveryFee = snap, fee
	}

	// 3. Coupon (optional): compute discount + free-delivery.
	var (
		discount     *Discount
		couponID     *uint64
		subDiscount  int64
		freeDelivery bool
	)
	if code := strings.TrimSpace(in.CouponCode); code != "" {
		discount, err = s.coupons.Validate(ctx, code, userID, subtotal, deliveryFee)
		if err != nil {
			return nil, err
		}
		couponID = &discount.Coupon.ID
		if discount.AppliedToSubtotal {
			subDiscount = discount.DiscountCents
		}
		freeDelivery = discount.FreeDelivery
	}

	// 4. Tax on the discounted subtotal (delivery is not taxed here).
	taxPercent := s.settings.GetInt(ctx, models.SettingTaxPercent, 0)
	taxable := subtotal - subDiscount
	if taxable < 0 {
		taxable = 0
	}
	tax := taxable * taxPercent / 100

	// 5. Totals. A free-delivery coupon adds the fee to discount so it nets out.
	discountCents := subDiscount
	if freeDelivery {
		discountCents += deliveryFee
	}
	total := subtotal + deliveryFee + tax - discountCents
	if total < 0 {
		total = 0
	}

	// 6. Payment method must be enabled in settings.
	method, status, payStatus, err := s.resolvePayment(ctx, in.PaymentMethod)
	if err != nil {
		return nil, err
	}

	// 7. Build and persist the order atomically (+ coupon redemption).
	order := &models.Order{
		UserID:           userID,
		Code:             generateOrderCode(),
		Status:           status,
		FulfillmentType:  models.FulfillmentType(in.FulfillmentType),
		SubtotalCents:    subtotal,
		DiscountCents:    discountCents,
		DeliveryFeeCents: deliveryFee,
		TaxCents:         tax,
		TotalCents:       total,
		Currency:         s.currency,
		PaymentMethod:    method,
		PaymentStatus:    payStatus,
		CouponID:         couponID,
		AddressSnapshot:  addrSnap,
		ScheduledFor:     in.ScheduledFor,
		Notes:            strings.TrimSpace(in.Notes),
		Items:            items,
	}
	if idempotencyKey != "" {
		order.IdempotencyKey = &idempotencyKey
	}

	var redeem *repository.RedeemInstruction
	if couponID != nil {
		redeem = &repository.RedeemInstruction{CouponID: *couponID, UserID: userID}
	}

	if err := s.orders.Create(ctx, order, redeem); err != nil {
		// Concurrent retry with the same idempotency key: return the winner.
		if repository.IsDuplicateKey(err) && idempotencyKey != "" {
			if existing, ferr := s.orders.FindByIdempotency(ctx, userID, idempotencyKey); ferr == nil {
				return existing, nil
			}
		}
		if errors.Is(err, repository.ErrCouponExhausted) {
			return nil, ErrCouponExhausted
		}
		return nil, err
	}

	logger.FromContext(ctx).Info("order placed",
		slog.String("code", order.Code), slog.Uint64("order_id", order.ID),
		slog.String("status", string(order.Status)), slog.String("payment", order.PaymentMethod),
		slog.Int64("total_cents", order.TotalCents))

	// Fire the Purchase conversion event (server-side, deduped by order code).
	if s.notifier != nil {
		s.notifier.TrackPurchase(ctx, order.ID)
	}
	return order, nil
}

// priceItems re-prices every cart line from the DB and returns the order items,
// subtotal, and the distinct product ids (for zone matching).
func (s *OrderService) priceItems(ctx context.Context, in []CartItemInput) ([]models.OrderItem, int64, []uint64, error) {
	items := make([]models.OrderItem, 0, len(in))
	var subtotal int64
	seen := map[uint64]struct{}{}
	var productIDs []uint64

	for _, ci := range in {
		p, err := s.catalog.FindProductByID(ctx, ci.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, 0, nil, ErrItemUnavailable
			}
			return nil, 0, nil, err
		}
		if !p.IsAvailable {
			return nil, 0, nil, ErrItemUnavailable
		}

		unitBase, weight, err := s.basePrice(p, ci)
		if err != nil {
			return nil, 0, nil, err
		}
		modSum, mods, err := selectModifiers(p, ci.ModifierIDs)
		if err != nil {
			return nil, 0, nil, err
		}

		unitPrice := unitBase + modSum
		lineTotal := unitPrice * int64(ci.Qty)

		items = append(items, models.OrderItem{
			ProductID:      p.ID,
			VariantID:      ci.VariantID,
			NameSnapshot:   s.productName(ctx, p),
			Qty:            ci.Qty,
			WeightGrams:    weight,
			UnitPriceCents: unitPrice,
			Modifiers:      mods,
			LineTotalCents: lineTotal,
		})
		subtotal += lineTotal

		if _, ok := seen[p.ID]; !ok {
			seen[p.ID] = struct{}{}
			productIDs = append(productIDs, p.ID)
		}
	}
	return items, subtotal, productIDs, nil
}

// basePrice returns the per-unit base price (before modifiers) and the weight
// (for per-kg items) for a cart line.
func (s *OrderService) basePrice(p *models.Product, ci CartItemInput) (int64, *int, error) {
	switch p.PricingMode {
	case models.PricingWeight:
		if ci.WeightGrams == nil || *ci.WeightGrams <= 0 {
			return 0, nil, ErrWeightRequired
		}
		// price per kg * grams / 1000, rounded to the nearest cent.
		base := (p.BasePriceCents*int64(*ci.WeightGrams) + 500) / 1000
		return base, ci.WeightGrams, nil
	default: // unit
		if ci.VariantID != nil {
			for _, v := range p.Variants {
				if v.ID == *ci.VariantID {
					return v.PriceCents, nil, nil
				}
			}
			return 0, nil, ErrVariantInvalid
		}
		return p.BasePriceCents, nil, nil
	}
}

// productName returns the default-locale snapshot name (falls back to slug).
func (s *OrderService) productName(ctx context.Context, p *models.Product) string {
	tr, err := s.translations.LoadForEntity(ctx, models.EntityProduct, p.ID)
	if err == nil {
		if name := i18n.NewResolver(s.defaultLocale, tr).Field(models.EntityProduct, p.ID, models.FieldName, s.defaultLocale); name != "" {
			return name
		}
	}
	return p.Slug
}

// quoteDelivery loads + snapshots the address, matches zones, and returns the
// fee, enforcing the zone minimum-order.
func (s *OrderService) quoteDelivery(ctx context.Context, userID uint64, addressID *uint64, productIDs []uint64, subtotal int64) (*models.AddressSnapshot, int64, error) {
	if addressID == nil {
		return nil, 0, ErrAddressRequired
	}
	addr, err := s.addresses.FindByID(ctx, *addressID)
	if err != nil {
		return nil, 0, ErrAddressRequired
	}
	if addr.UserID != userID {
		return nil, 0, ErrForbidden
	}
	if addr.Lat == nil || addr.Lng == nil {
		return nil, 0, ErrAddressNoGeo
	}

	quote, err := s.zones.Quote(ctx, geo.Point{Lat: *addr.Lat, Lng: *addr.Lng}, productIDs)
	if err != nil {
		return nil, 0, err
	}
	if !quote.Deliverable {
		return nil, 0, ErrUndeliverable
	}
	if subtotal < quote.MinOrderCents {
		return nil, 0, ErrBelowMinOrder
	}

	snap := &models.AddressSnapshot{
		Label: addr.Label, Line1: addr.Line1, Line2: addr.Line2, City: addr.City,
		ProvinceCode: addr.ProvinceCode, PostalCode: addr.PostalCode, CountryCode: addr.CountryCode,
		Lat: addr.Lat, Lng: addr.Lng, Phone: addr.PhoneE164, Notes: addr.Notes,
	}
	return snap, quote.FeeCents, nil
}

// resolvePayment validates the chosen method against settings toggles and
// returns the method plus the initial order/payment status.
func (s *OrderService) resolvePayment(ctx context.Context, method string) (string, models.OrderStatus, models.PaymentStatus, error) {
	switch method {
	case "cod":
		if !s.settings.GetBool(ctx, models.SettingCODEnabled, true) {
			return "", "", "", ErrPaymentMethodDisabled
		}
		// COD skips pending_payment: straight to confirmed, settled on delivery.
		return "cod", models.StatusConfirmed, models.PayCODPending, nil
	case "square":
		if !s.settings.GetBool(ctx, models.SettingSquareEnabled, false) {
			return "", "", "", ErrPaymentMethodDisabled
		}
		// Card: awaits capture (completed by the payment phase/webhook).
		return "square", models.StatusPendingPayment, models.PayPending, nil
	default:
		return "", "", "", ErrPaymentMethodDisabled
	}
}

// --- Queries & status transitions ---

func (s *OrderService) GetByCode(ctx context.Context, userID uint64, isStaff bool, code string) (*models.Order, error) {
	o, err := s.orders.FindByCode(ctx, code)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if !isStaff && o.UserID != userID {
		return nil, ErrNotFound // don't reveal existence of others' orders
	}
	return o, nil
}

func (s *OrderService) ListMine(ctx context.Context, userID uint64, limit, offset int) ([]models.Order, int64, error) {
	return s.orders.ListByUser(ctx, userID, limit, offset)
}

func (s *OrderService) AdminList(ctx context.Context, status string, limit, offset int) ([]models.Order, int64, error) {
	return s.orders.AdminList(ctx, status, limit, offset)
}

// AdvanceStatus moves an order to a new status if the transition is allowed,
// recording the actor in the history.
func (s *OrderService) AdvanceStatus(ctx context.Context, orderID uint64, to models.OrderStatus, actorID uint64, note string) (*models.Order, error) {
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if !o.Status.CanTransitionTo(to) {
		return nil, ErrInvalidTransition
	}
	from := o.Status
	if err := s.orders.AdvanceStatus(ctx, o, to, &actorID, note); err != nil {
		return nil, err
	}
	logger.FromContext(ctx).Info("order status changed",
		slog.String("code", o.Code), slog.String("from", string(from)),
		slog.String("to", string(to)), slog.Uint64("actor_id", actorID))
	// Fan out the change: live SSE event + queued push notification (both
	// optional and best-effort, so tracking/notifications never block the write).
	if s.events != nil {
		s.events.PublishOrderEvent(ctx, o.Code, string(to), note)
	}
	if s.notifier != nil {
		s.notifier.NotifyOrderStatus(ctx, o.ID, string(to))
	}
	return o, nil
}

// generateOrderCode returns a short, human-friendly, unique-ish order code. The
// DB unique index on code is the real guarantee; collisions are astronomically
// unlikely with 8 base32 chars.
func generateOrderCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no ambiguous chars
	var b [5]byte
	_, _ = rand.Read(b[:])
	n := binary.BigEndian.Uint64(append([]byte{0, 0, 0}, b[:]...))
	out := make([]byte, 8)
	for i := range out {
		out[i] = alphabet[n%uint64(len(alphabet))]
		n /= uint64(len(alphabet))
	}
	return "K" + string(out)
}
