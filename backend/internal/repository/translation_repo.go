package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TranslationRepo is the data-access layer for the generic translations store.
type TranslationRepo struct {
	db *gorm.DB
}

func NewTranslationRepo(db *gorm.DB) *TranslationRepo { return &TranslationRepo{db: db} }

// LoadForEntities returns all translation rows for a set of entity ids of one
// type — one query to translate an entire menu.
func (r *TranslationRepo) LoadForEntities(ctx context.Context, entityType string, ids []uint64) ([]models.Translation, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []models.Translation
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id IN ?", entityType, ids).
		Find(&rows).Error
	return rows, err
}

// LoadForEntity returns all translations for a single entity.
func (r *TranslationRepo) LoadForEntity(ctx context.Context, entityType string, id uint64) ([]models.Translation, error) {
	var rows []models.Translation
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, id).
		Find(&rows).Error
	return rows, err
}

// Upsert inserts or updates one localized field value (idempotent on the
// unique key entity_type+entity_id+locale+field).
func (r *TranslationRepo) Upsert(ctx context.Context, t *models.Translation) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "entity_type"}, {Name: "entity_id"}, {Name: "locale"}, {Name: "field"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(t).Error
}

// UpsertMany applies several field translations in one transaction. Used by
// admin create/update of categories/products.
func (r *TranslationRepo) UpsertMany(ctx context.Context, rows []models.Translation) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "entity_type"}, {Name: "entity_id"}, {Name: "locale"}, {Name: "field"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&rows).Error
}

// DeleteForEntity removes all translations for an entity (on hard delete).
func (r *TranslationRepo) DeleteForEntity(ctx context.Context, entityType string, id uint64) error {
	return r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, id).
		Delete(&models.Translation{}).Error
}
