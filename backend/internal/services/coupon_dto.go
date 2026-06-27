package services

import "time"

// CouponInput is the admin create/update payload for a coupon.
type CouponInput struct {
	Code             string     `json:"code" validate:"required,max=60"`
	Type             string     `json:"type" validate:"required,oneof=percent fixed free_delivery"`
	Value            int64      `json:"value" validate:"min=0"`
	MinOrderCents    int64      `json:"min_order_cents" validate:"min=0"`
	MaxDiscountCents int64      `json:"max_discount_cents" validate:"min=0"`
	UsageLimit       int        `json:"usage_limit" validate:"min=0"`
	PerUserLimit     int        `json:"per_user_limit" validate:"min=0"`
	StartsAt         *time.Time `json:"starts_at"`
	EndsAt           *time.Time `json:"ends_at"`
	IsActive         *bool      `json:"is_active"`
}

// ValidateCouponInput is the checkout preview request.
type ValidateCouponInput struct {
	Code             string `json:"code" validate:"required,max=60"`
	SubtotalCents    int64  `json:"subtotal_cents" validate:"min=0"`
	DeliveryFeeCents int64  `json:"delivery_fee_cents" validate:"min=0"`
}

// BannerInput is the admin create/update payload for a banner.
type BannerInput struct {
	Title     string     `json:"title" validate:"omitempty,max=160"`
	ImageURL  string     `json:"image_url" validate:"required,url,max=500"`
	LinkURL   string     `json:"link_url" validate:"omitempty,url,max=500"`
	SortOrder int        `json:"sort_order" validate:"omitempty,min=0"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	IsActive  *bool      `json:"is_active"`
}
