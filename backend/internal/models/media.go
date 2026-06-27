package models

import "time"

type Media struct {
	ID           uint64     `gorm:"primarykey"         json:"id"`
	Filename     string     `gorm:"not null"           json:"filename"`
	OriginalName string     `gorm:"not null"           json:"original_name"`
	MIME         string     `gorm:"column:mime;not null" json:"mime"`
	Ext          string     `gorm:"not null"           json:"ext"`
	SizeBytes    int64      `gorm:"not null"           json:"size_bytes"`
	Width        int        `gorm:"not null"           json:"width"`
	Height       int        `gorm:"not null"           json:"height"`
	URL          string     `gorm:"not null"           json:"url"`
	ThumbURL     string     `gorm:"not null"           json:"thumb_url"`
	Alt          *string    `json:"alt"`
	CreatedBy    *uint64    `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `gorm:"index"              json:"-"`
}
