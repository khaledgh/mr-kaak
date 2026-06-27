// Package services holds business logic. It orchestrates repositories, auth,
// (later) cache, search, and jobs. Handlers call services; services never
// write SQL directly. Services return the domain errors below, which handlers
// map to HTTP status + response codes.
package services

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountSuspended   = errors.New("account suspended")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid or expired token")

	// Catalog
	ErrSlugTaken             = errors.New("slug already in use")
	ErrCategoryNotEmpty      = errors.New("category still has products")
	ErrDefaultLocaleRequired = errors.New("a translation for the default locale is required")

	// Geo
	ErrInvalidZone = errors.New("invalid delivery zone definition")

	// Coupons
	ErrCouponInvalid    = errors.New("invalid coupon code")
	ErrCouponExpired    = errors.New("coupon has expired")
	ErrCouponNotStarted = errors.New("coupon is not active yet")
	ErrCouponMinOrder   = errors.New("order does not meet the coupon minimum")
	ErrCouponExhausted  = errors.New("coupon usage limit reached")
	ErrCouponPerUser    = errors.New("coupon already used the maximum number of times")

	// Orders
	ErrItemUnavailable       = errors.New("an item is unavailable")
	ErrVariantInvalid        = errors.New("invalid product variant")
	ErrWeightRequired        = errors.New("weight is required for this item")
	ErrModifierInvalid       = errors.New("invalid modifier selection")
	ErrModifierBounds        = errors.New("modifier selection violates group rules")
	ErrAddressRequired       = errors.New("a delivery address is required")
	ErrAddressNoGeo          = errors.New("address has no coordinates for delivery matching")
	ErrUndeliverable         = errors.New("address is outside the delivery area")
	ErrBelowMinOrder         = errors.New("order is below the delivery minimum")
	ErrPaymentMethodDisabled = errors.New("selected payment method is not available")
	ErrInvalidTransition     = errors.New("invalid order status transition")

	// Payments
	ErrPaymentFailed   = errors.New("payment failed")
	ErrWebhookInvalid  = errors.New("invalid webhook signature")
	ErrNothingToRefund = errors.New("no captured payment to refund")
	ErrRefundFailed    = errors.New("refund failed")

	// Search
	ErrSearchUnavailable = errors.New("search index is not available")
)
