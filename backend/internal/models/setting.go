package models

import (
	"encoding/json"
	"time"
)

// Setting is one row of the key/value settings store (plan §4). value_json
// keeps each value typed. Setting keys used across the app:
const (
	SettingCODEnabled    = "cod_enabled"
	SettingSquareEnabled = "square_enabled"
	SettingTaxPercent    = "tax_percent"
	SettingCurrency      = "currency"
)

type Setting struct {
	Key       string          `gorm:"column:key;primaryKey;size:128" json:"key"`
	ValueJSON json.RawMessage `gorm:"column:value_json;type:json" json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName pins the table (GORM would otherwise pluralize to "settings",
// which is correct here, but we make it explicit).
func (Setting) TableName() string { return "settings" }
