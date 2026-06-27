package repository

import (
	"context"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// PaymentRepo is the data-access layer for payment transactions.
type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) *PaymentRepo { return &PaymentRepo{db: db} }

func (r *PaymentRepo) Create(ctx context.Context, t *models.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// ExistsByRef reports whether a provider_ref has already been recorded — the
// webhook idempotency check.
func (r *PaymentRepo) ExistsByRef(ctx context.Context, provider, ref string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.PaymentTransaction{}).
		Where("provider = ? AND provider_ref = ?", provider, ref).Count(&n).Error
	return n > 0, err
}

func (r *PaymentRepo) ListByOrder(ctx context.Context, orderID uint64) ([]models.PaymentTransaction, error) {
	var txns []models.PaymentTransaction
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("id ASC").Find(&txns).Error
	return txns, err
}

// LatestCharge returns the most recent succeeded charge for an order (used to
// find the payment to refund).
func (r *PaymentRepo) LatestCharge(ctx context.Context, orderID uint64) (*models.PaymentTransaction, error) {
	var t models.PaymentTransaction
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND kind = ? AND status = ?", orderID, models.TxnCharge, models.TxnSucceeded).
		Order("id DESC").First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}
