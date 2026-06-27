package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// AddressRepo is the data-access layer for user addresses.
type AddressRepo struct {
	db *gorm.DB
}

func NewAddressRepo(db *gorm.DB) *AddressRepo { return &AddressRepo{db: db} }

func (r *AddressRepo) Create(ctx context.Context, a *models.Address) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AddressRepo) Update(ctx context.Context, a *models.Address) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// FindByID loads an address by id. The caller must check ownership.
func (r *AddressRepo) FindByID(ctx context.Context, id uint64) (*models.Address, error) {
	var a models.Address
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &a, nil
}

// ListByUser returns all of a user's addresses, newest first.
func (r *AddressRepo) ListByUser(ctx context.Context, userID uint64) ([]models.Address, error) {
	var addrs []models.Address
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&addrs).Error
	return addrs, err
}

// Delete soft-deletes an address scoped to its owner (prevents cross-user
// deletion even if an id is guessed).
func (r *AddressRepo) Delete(ctx context.Context, userID, id uint64) error {
	res := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.Address{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SetDefault sets the user's default address inside a transaction, verifying
// the address belongs to the user.
func (r *AddressRepo) SetDefault(ctx context.Context, userID, addressID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Address{}).
			Where("id = ? AND user_id = ?", addressID, userID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Model(&models.User{}).Where("id = ?", userID).
			UpdateColumn("default_address_id", addressID).Error
	})
}
