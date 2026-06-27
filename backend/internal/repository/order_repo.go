package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// OrderRepo is the data-access layer for orders. The checkout write is a single
// transaction: order + items + initial status history + atomic coupon
// redemption, so a placed order is always internally consistent (plan §3.3).
type OrderRepo struct {
	db      *gorm.DB
	coupons *CouponRepo
}

func NewOrderRepo(db *gorm.DB, coupons *CouponRepo) *OrderRepo {
	return &OrderRepo{db: db, coupons: coupons}
}

// RedeemInstruction tells Create to atomically consume a coupon use in the same
// transaction as the order insert.
type RedeemInstruction struct {
	CouponID uint64
	UserID   uint64
}

// Create inserts the order (with its items), the initial status-history row, and
// optionally redeems a coupon — all atomically. A duplicate idempotency key
// surfaces as a duplicate-key error the service translates.
func (r *OrderRepo) Create(ctx context.Context, o *models.Order, redeem *RedeemInstruction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		hist := &models.OrderStatusHistory{OrderID: o.ID, ToStatus: o.Status, Note: "order created"}
		if err := tx.Create(hist).Error; err != nil {
			return err
		}
		if redeem != nil {
			if err := r.coupons.RedeemTx(tx, redeem.CouponID, redeem.UserID, &o.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByIdempotency returns a prior order for (user, key), or nil if none.
func (r *OrderRepo) FindByIdempotency(ctx context.Context, userID uint64, key string) (*models.Order, error) {
	var o models.Order
	err := r.db.WithContext(ctx).
		Preload("Items").Preload("History").
		Where("user_id = ? AND idempotency_key = ?", userID, key).First(&o).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &o, nil
}

// FindByCode loads an order (with items + history) by its public code.
func (r *OrderRepo) FindByCode(ctx context.Context, code string) (*models.Order, error) {
	var o models.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("History", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC, id ASC") }).
		Where("code = ?", code).First(&o).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &o, nil
}

func (r *OrderRepo) FindByID(ctx context.Context, id uint64) (*models.Order, error) {
	var o models.Order
	if err := r.db.WithContext(ctx).Preload("Items").First(&o, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &o, nil
}

// ListByUser returns a user's orders (newest first) plus the total count.
func (r *OrderRepo) ListByUser(ctx context.Context, userID uint64, limit, offset int) ([]models.Order, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Order{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []models.Order
	err := q.Preload("Items").Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	return orders, total, err
}

// AdminList returns orders filtered by status (optional) with the total count.
func (r *OrderRepo) AdminList(ctx context.Context, status string, limit, offset int) ([]models.Order, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Order{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []models.Order
	err := q.Preload("Items").Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error
	return orders, total, err
}

// AdvanceStatus updates an order's status and appends a history row atomically.
// The caller validates that the transition is allowed.
func (r *OrderRepo) AdvanceStatus(ctx context.Context, o *models.Order, to models.OrderStatus, actorID *uint64, note string) error {
	from := o.Status
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(o).Update("status", to).Error; err != nil {
			return err
		}
		o.Status = to
		return tx.Create(&models.OrderStatusHistory{
			OrderID: o.ID, FromStatus: from, ToStatus: to, ActorID: actorID, Note: note,
		}).Error
	})
}

// SetPaymentStatus updates the payment_status column (used by the payment layer).
func (r *OrderRepo) SetPaymentStatus(ctx context.Context, id uint64, status models.PaymentStatus) error {
	return r.db.WithContext(ctx).Model(&models.Order{Base: models.Base{ID: id}}).
		Update("payment_status", status).Error
}
