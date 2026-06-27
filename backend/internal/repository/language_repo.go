package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LanguageRepo is the data-access layer for languages and UI string bundles.
type LanguageRepo struct {
	db *gorm.DB
}

func NewLanguageRepo(db *gorm.DB) *LanguageRepo { return &LanguageRepo{db: db} }

func (r *LanguageRepo) List(ctx context.Context, activeOnly bool) ([]models.Language, error) {
	q := r.db.WithContext(ctx).Order("sort_order ASC, id ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var langs []models.Language
	return langs, q.Find(&langs).Error
}

func (r *LanguageRepo) FindByID(ctx context.Context, id uint64) (*models.Language, error) {
	var l models.Language
	if err := r.db.WithContext(ctx).First(&l, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &l, nil
}

func (r *LanguageRepo) Default(ctx context.Context) (*models.Language, error) {
	var l models.Language
	if err := r.db.WithContext(ctx).Where("is_default = ?", true).First(&l).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &l, nil
}

func (r *LanguageRepo) Create(ctx context.Context, l *models.Language) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *LanguageRepo) Update(ctx context.Context, l *models.Language) error {
	return r.db.WithContext(ctx).Save(l).Error
}

// SetDefault makes exactly one language the default, in a transaction.
func (r *LanguageRepo) SetDefault(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Language{}).Where("is_default = ?", true).
			Update("is_default", false).Error; err != nil {
			return err
		}
		res := tx.Model(&models.Language{}).Where("id = ?", id).
			Updates(map[string]any{"is_default": true, "is_active": true})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// CountActive returns the number of active languages (guards against
// deactivating the last one).
func (r *LanguageRepo) CountActive(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Language{}).Where("is_active = ?", true).Count(&n).Error
	return n, err
}

// --- UI strings ---

func (r *LanguageRepo) Bundle(ctx context.Context, locale string) (map[string]string, error) {
	var rows []models.UIString
	if err := r.db.WithContext(ctx).Where("locale = ?", locale).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, s := range rows {
		out[s.Key] = s.Value
	}
	return out, nil
}

func (r *LanguageRepo) UpsertStrings(ctx context.Context, rows []models.UIString) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "locale"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&rows).Error
}
