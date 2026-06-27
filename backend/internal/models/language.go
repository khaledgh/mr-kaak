package models

import "time"

// Language is an admin-configurable locale. Adding/activating a row makes the
// locale available without a redeploy (plan §7).
type Language struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code       string    `gorm:"size:10;uniqueIndex;not null" json:"code"`
	Name       string    `gorm:"size:80;not null" json:"name"`
	NativeName string    `gorm:"size:80;not null" json:"native_name"`
	IsDefault  bool      `gorm:"not null;default:false" json:"is_default"`
	IsRTL      bool      `gorm:"column:is_rtl;not null;default:false" json:"is_rtl"`
	IsActive   bool      `gorm:"not null;default:true" json:"is_active"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UIString is one localized static UI string (frontend bundle entry).
type UIString struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Locale    string    `gorm:"size:10;not null;uniqueIndex:uq_ui_string,priority:1" json:"locale"`
	Key       string    `gorm:"column:key;size:160;not null;uniqueIndex:uq_ui_string,priority:2" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
