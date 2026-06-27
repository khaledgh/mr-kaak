package repository

import (
	"context"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// ErrCouponExhausted means the coupon's global usage limit was reached.
var ErrCouponExhausted = errors.New("coupon usage limit reached")

// CouponRepo is the data-access layer for coupons + redemptions.
type CouponRepo struct {
	db *gorm.DB
}

func NewCouponRepo(db *gorm.DB) *CouponRepo { return &CouponRepo{db: db} }

func (r *CouponRepo) FindByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var c models.Coupon
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&c).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &c, nil
}

func (r *CouponRepo) FindByID(ctx context.Context, id uint64) (*models.Coupon, error) {
	var c models.Coupon
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &c, nil
}

func (r *CouponRepo) List(ctx context.Context) ([]models.Coupon, error) {
	var cs []models.Coupon
	return cs, r.db.WithContext(ctx).Order("id DESC").Find(&cs).Error
}

func (r *CouponRepo) Create(ctx context.Context, c *models.Coupon) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CouponRepo) Update(ctx context.Context, c *models.Coupon) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CouponRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Coupon{}, id).Error
}

// UserRedemptionCount returns how many times a user has redeemed a coupon.
func (r *CouponRepo) UserRedemptionCount(ctx context.Context, couponID, userID uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.CouponRedemption{}).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).Count(&n).Error
	return n, err
}

// RedeemTx atomically increments used_count (respecting usage_limit) and records
// a redemption row, all inside the caller's order transaction. Returns
// ErrCouponExhausted if the global usage limit is already reached.
//
// The conditional UPDATE is the concurrency guard: only one transaction can take
// the last remaining use because the row is locked and the WHERE check is atomic.
func (r *CouponRepo) RedeemTx(tx *gorm.DB, couponID, userID uint64, orderID *uint64) error {
	res := tx.Model(&models.Coupon{}).
		Where("id = ? AND (usage_limit = 0 OR used_count < usage_limit)", couponID).
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCouponExhausted
	}
	return tx.Create(&models.CouponRedemption{CouponID: couponID, UserID: userID, OrderID: orderID}).Error
}
