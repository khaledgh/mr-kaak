package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// ZoneRepo is the data-access layer for delivery zones.
type ZoneRepo struct {
	db *gorm.DB
}

func NewZoneRepo(db *gorm.DB) *ZoneRepo { return &ZoneRepo{db: db} }

func (r *ZoneRepo) List(ctx context.Context, activeOnly bool) ([]models.DeliveryZone, error) {
	q := r.db.WithContext(ctx).Order("sort_order ASC, id ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var zones []models.DeliveryZone
	return zones, q.Find(&zones).Error
}

// ActiveGlobal returns active store-wide zones.
func (r *ZoneRepo) ActiveGlobal(ctx context.Context) ([]models.DeliveryZone, error) {
	var zones []models.DeliveryZone
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND scope = ?", true, models.ScopeGlobal).
		Order("sort_order ASC, id ASC").Find(&zones).Error
	return zones, err
}

// ActiveForProducts returns active product-scoped zones for the given products.
func (r *ZoneRepo) ActiveForProducts(ctx context.Context, productIDs []uint64) ([]models.DeliveryZone, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	var zones []models.DeliveryZone
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND scope = ? AND product_id IN ?", true, models.ScopeProduct, productIDs).
		Find(&zones).Error
	return zones, err
}

func (r *ZoneRepo) FindByID(ctx context.Context, id uint64) (*models.DeliveryZone, error) {
	var z models.DeliveryZone
	if err := r.db.WithContext(ctx).First(&z, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &z, nil
}

func (r *ZoneRepo) Create(ctx context.Context, z *models.DeliveryZone) error {
	return r.db.WithContext(ctx).Create(z).Error
}

func (r *ZoneRepo) Update(ctx context.Context, z *models.DeliveryZone) error {
	return r.db.WithContext(ctx).Save(z).Error
}

func (r *ZoneRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.DeliveryZone{}, id).Error
}
