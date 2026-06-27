package models

import "time"

// CouponType is the discount mechanic.
type CouponType string

const (
	CouponPercent      CouponType = "percent"       // value = 0..100
	CouponFixed        CouponType = "fixed"         // value = cents off
	CouponFreeDelivery CouponType = "free_delivery" // waives delivery fee
)

// Coupon is a discount code with eligibility and usage limits.
type Coupon struct {
	Base
	Code             string     `gorm:"size:60;uniqueIndex;not null" json:"code"`
	Type             CouponType `gorm:"type:enum('percent','fixed','free_delivery');not null" json:"type"`
	Value            int64      `gorm:"not null;default:0" json:"value"`
	MinOrderCents    int64      `gorm:"not null;default:0" json:"min_order_cents"`
	MaxDiscountCents int64      `gorm:"not null;default:0" json:"max_discount_cents"`
	UsageLimit       int        `gorm:"not null;default:0" json:"usage_limit"`
	UsedCount        int        `gorm:"not null;default:0" json:"used_count"`
	PerUserLimit     int        `gorm:"not null;default:0" json:"per_user_limit"`
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
	IsActive         bool       `gorm:"not null;default:true" json:"is_active"`
}

// CouponRedemption records a single use of a coupon by a user.
type CouponRedemption struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CouponID  uint64    `gorm:"index:idx_redemptions_coupon_user,priority:1;not null" json:"coupon_id"`
	UserID    uint64    `gorm:"index:idx_redemptions_coupon_user,priority:2;not null" json:"user_id"`
	OrderID   *uint64   `json:"order_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Banner is a scheduled marketing slide.
type Banner struct {
	Base
	Title     string     `gorm:"size:160;not null" json:"title"`
	ImageURL  string     `gorm:"size:500;not null" json:"image_url"`
	LinkURL   string     `gorm:"size:500" json:"link_url,omitempty"`
	SortOrder int        `gorm:"not null;default:0" json:"sort_order"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	IsActive  bool       `gorm:"not null;default:true" json:"is_active"`
}
