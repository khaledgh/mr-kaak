package services

import "time"

// CartItemInput is one line submitted at checkout. The server re-prices every
// line from the DB — client-sent prices are never trusted (plan §9).
type CartItemInput struct {
	ProductID   uint64   `json:"product_id" validate:"required"`
	VariantID   *uint64  `json:"variant_id"`
	Qty         int      `json:"qty" validate:"required,min=1,max=99"`
	WeightGrams *int     `json:"weight_grams" validate:"omitempty,min=1"`
	ModifierIDs []uint64 `json:"modifier_ids"`
}

// CheckoutInput is the order placement payload.
type CheckoutInput struct {
	FulfillmentType string          `json:"fulfillment_type" validate:"required,oneof=delivery pickup"`
	AddressID       *uint64         `json:"address_id"`
	PaymentMethod   string          `json:"payment_method" validate:"required,oneof=cod square"`
	CouponCode      string          `json:"coupon_code" validate:"omitempty,max=60"`
	Notes           string          `json:"notes" validate:"omitempty,max=500"`
	ScheduledFor    *time.Time      `json:"scheduled_for"`
	Items           []CartItemInput `json:"items" validate:"required,min=1,dive"`
}

// AdvanceStatusInput is the staff action to move an order forward.
type AdvanceStatusInput struct {
	Status string `json:"status" validate:"required"`
	Note   string `json:"note" validate:"omitempty,max=255"`
}
