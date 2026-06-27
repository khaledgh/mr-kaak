// Package models holds the GORM models that define the database schema.
// Every table embeds Base for a consistent id + audit timestamps + soft delete.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Base is embedded in every model: bigint auto-increment PK, created/updated
// timestamps, and a soft-delete column (plan §17: soft deletes everywhere).
type Base struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
