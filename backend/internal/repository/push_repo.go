package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PushRepo is the data-access layer for Web Push subscriptions.
type PushRepo struct {
	db *gorm.DB
}

func NewPushRepo(db *gorm.DB) *PushRepo { return &PushRepo{db: db} }

// Upsert stores a subscription, updating keys if the endpoint already exists.
func (r *PushRepo) Upsert(ctx context.Context, s *models.PushSubscription) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "p256dh", "auth", "user_agent", "updated_at"}),
	}).Create(s).Error
}

func (r *PushRepo) DeleteByEndpoint(ctx context.Context, userID uint64, endpoint string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND endpoint = ?", userID, endpoint).
		Delete(&models.PushSubscription{}).Error
}

func (r *PushRepo) ListByUser(ctx context.Context, userID uint64) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

// DeleteByID removes a subscription (used to prune endpoints the push service
// reports as gone/410).
func (r *PushRepo) DeleteByID(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.PushSubscription{}, id).Error
}
