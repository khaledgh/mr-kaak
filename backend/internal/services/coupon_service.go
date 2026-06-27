package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"gorm.io/gorm"
)

// CouponService validates and applies coupons. Validation is read-only and used
// for the checkout preview; redemption happens atomically inside the order
// transaction via RedeemTx (plan §11).
type CouponService struct {
	coupons *repository.CouponRepo
}

func NewCouponService(coupons *repository.CouponRepo) *CouponService {
	return &CouponService{coupons: coupons}
}

// Discount is the computed effect of a coupon on an order.
type Discount struct {
	Coupon            *models.Coupon `json:"-"`
	Code              string         `json:"code"`
	DiscountCents     int64          `json:"discount_cents"`
	FreeDelivery      bool           `json:"free_delivery"`
	AppliedToSubtotal bool           `json:"applied_to_subtotal"`
}

// Validate checks eligibility and computes the discount for a user's cart.
// subtotalCents is the pre-discount item subtotal; deliveryFeeCents is the
// quoted delivery fee (needed for free_delivery coupons).
func (s *CouponService) Validate(ctx context.Context, code string, userID uint64, subtotalCents, deliveryFeeCents int64) (*Discount, error) {
	c, err := s.coupons.FindByCode(ctx, normalizeCode(code))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCouponInvalid
		}
		return nil, err
	}

	now := time.Now()
	switch {
	case !c.IsActive:
		return nil, ErrCouponInvalid
	case c.StartsAt != nil && now.Before(*c.StartsAt):
		return nil, ErrCouponNotStarted
	case c.EndsAt != nil && now.After(*c.EndsAt):
		return nil, ErrCouponExpired
	case subtotalCents < c.MinOrderCents:
		return nil, ErrCouponMinOrder
	case c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit:
		return nil, ErrCouponExhausted
	}

	if c.PerUserLimit > 0 {
		used, err := s.coupons.UserRedemptionCount(ctx, c.ID, userID)
		if err != nil {
			return nil, err
		}
		if used >= int64(c.PerUserLimit) {
			return nil, ErrCouponPerUser
		}
	}

	return computeDiscount(c, subtotalCents, deliveryFeeCents), nil
}

// computeDiscount applies the coupon mechanic, capping percent discounts.
func computeDiscount(c *models.Coupon, subtotalCents, deliveryFeeCents int64) *Discount {
	d := &Discount{Coupon: c, Code: c.Code}
	switch c.Type {
	case models.CouponPercent:
		amount := subtotalCents * c.Value / 100
		if c.MaxDiscountCents > 0 && amount > c.MaxDiscountCents {
			amount = c.MaxDiscountCents
		}
		d.DiscountCents = clampDiscount(amount, subtotalCents)
		d.AppliedToSubtotal = true
	case models.CouponFixed:
		d.DiscountCents = clampDiscount(c.Value, subtotalCents)
		d.AppliedToSubtotal = true
	case models.CouponFreeDelivery:
		d.DiscountCents = deliveryFeeCents
		d.FreeDelivery = true
	}
	return d
}

// clampDiscount ensures a subtotal discount never exceeds the subtotal.
func clampDiscount(amount, subtotal int64) int64 {
	if amount > subtotal {
		return subtotal
	}
	if amount < 0 {
		return 0
	}
	return amount
}

// RedeemTx is invoked inside the order transaction to atomically consume a use.
func (s *CouponService) RedeemTx(tx *gorm.DB, couponID, userID uint64, orderID *uint64) error {
	if err := s.coupons.RedeemTx(tx, couponID, userID, orderID); err != nil {
		if errors.Is(err, repository.ErrCouponExhausted) {
			return ErrCouponExhausted
		}
		return err
	}
	return nil
}

// --- Admin CRUD ---

func (s *CouponService) List(ctx context.Context) ([]models.Coupon, error) {
	return s.coupons.List(ctx)
}

func (s *CouponService) Create(ctx context.Context, in CouponInput) (*models.Coupon, error) {
	c := in.toModel()
	if err := s.coupons.Create(ctx, c); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return c, nil
}

func (s *CouponService) Update(ctx context.Context, id uint64, in CouponInput) (*models.Coupon, error) {
	existing, err := s.coupons.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	model := in.toModel()
	model.ID = existing.ID
	model.UsedCount = existing.UsedCount // preserve consumption count across edits
	if err := s.coupons.Update(ctx, model); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return model, nil
}

func (s *CouponService) Delete(ctx context.Context, id uint64) error {
	return s.coupons.Delete(ctx, id)
}

func (in CouponInput) toModel() *models.Coupon {
	return &models.Coupon{
		Code:             normalizeCode(in.Code),
		Type:             models.CouponType(in.Type),
		Value:            in.Value,
		MinOrderCents:    in.MinOrderCents,
		MaxDiscountCents: in.MaxDiscountCents,
		UsageLimit:       in.UsageLimit,
		PerUserLimit:     in.PerUserLimit,
		StartsAt:         in.StartsAt,
		EndsAt:           in.EndsAt,
		IsActive:         derefBool(in.IsActive, true),
	}
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
