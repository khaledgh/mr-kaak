package models

// Address is a Canadian delivery/contact address. lat/lng are retained for
// delivery-zone matching (plan §5). country_code is fixed to CA for now but
// the column leaves room to expand.
type Address struct {
	Base
	UserID       uint64   `gorm:"index;not null" json:"user_id"`
	Label        string   `gorm:"size:60" json:"label,omitempty"` // "Home", "Work"
	Line1        string   `gorm:"size:200;not null" json:"line1"`
	Line2        string   `gorm:"size:200" json:"line2,omitempty"`
	City         string   `gorm:"size:120;not null" json:"city"`
	ProvinceCode string   `gorm:"size:2;not null" json:"province_code"`
	PostalCode   string   `gorm:"size:7;not null" json:"postal_code"` // "A1A 1A1"
	CountryCode  string   `gorm:"size:2;not null;default:CA" json:"country_code"`
	Lat          *float64 `gorm:"type:decimal(10,7)" json:"lat,omitempty"`
	Lng          *float64 `gorm:"type:decimal(10,7)" json:"lng,omitempty"`
	PhoneE164    string   `gorm:"size:20" json:"phone,omitempty"`
	Notes        string   `gorm:"size:255" json:"notes,omitempty"`
}
