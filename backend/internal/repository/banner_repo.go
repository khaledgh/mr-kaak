package repository

import (
	"context"
	"time"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// BannerRepo is the data-access layer for marketing banners.
type BannerRepo struct {
	db *gorm.DB
}

func NewBannerRepo(db *gorm.DB) *BannerRepo { return &BannerRepo{db: db} }

// ActiveNow returns banners that are active and within their schedule window.
func (r *BannerRepo) ActiveNow(ctx context.Context, now time.Time) ([]models.Banner, error) {
	var bs []models.Banner
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("starts_at IS NULL OR starts_at <= ?", now).
		Where("ends_at IS NULL OR ends_at >= ?", now).
		Order("sort_order ASC, id ASC").Find(&bs).Error
	return bs, err
}

func (r *BannerRepo) List(ctx context.Context) ([]models.Banner, error) {
	var bs []models.Banner
	return bs, r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&bs).Error
}

func (r *BannerRepo) FindByID(ctx context.Context, id uint64) (*models.Banner, error) {
	var b models.Banner
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &b, nil
}

func (r *BannerRepo) Create(ctx context.Context, b *models.Banner) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *BannerRepo) Update(ctx context.Context, b *models.Banner) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *BannerRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Banner{}, id).Error
}
