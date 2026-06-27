package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/pkg/pagination"
	"gorm.io/gorm"
)

type MediaRepo struct{ db *gorm.DB }

func NewMediaRepo(db *gorm.DB) *MediaRepo { return &MediaRepo{db: db} }

func (r *MediaRepo) Create(ctx context.Context, m *models.Media) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *MediaRepo) List(ctx context.Context, p pagination.Params, q string) ([]models.Media, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Media{}).Where("deleted_at IS NULL")
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("original_name LIKE ? OR alt LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.Media
	err := query.Order("created_at DESC").
		Offset(p.Offset()).Limit(p.Limit()).
		Find(&items).Error
	return items, total, err
}

func (r *MediaRepo) GetByID(ctx context.Context, id uint64) (*models.Media, error) {
	var m models.Media
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &m, nil
}

func (r *MediaRepo) SoftDelete(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).
		Model(&models.Media{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", gorm.Expr("NOW()"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
