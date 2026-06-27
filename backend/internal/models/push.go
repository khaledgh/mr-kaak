package models

import "time"

// PushSubscription is a browser Web Push endpoint owned by a user (plan §8).
type PushSubscription struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	Endpoint  string    `gorm:"size:500;uniqueIndex;not null" json:"endpoint"`
	P256dh    string    `gorm:"size:255;not null" json:"-"`
	Auth      string    `gorm:"size:255;not null" json:"-"`
	UserAgent string    `gorm:"size:255" json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
