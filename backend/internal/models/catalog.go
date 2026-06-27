package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// PricingMode distinguishes the two fundamentally different ways an item is
// priced (plan §1): per-unit (with size variants & modifiers) vs per-kilogram.
type PricingMode string

const (
	PricingUnit   PricingMode = "unit"   // base_price_cents = base unit price
	PricingWeight PricingMode = "weight" // base_price_cents = price per kg
)

func (m PricingMode) Valid() bool { return m == PricingUnit || m == PricingWeight }

// Allergens is a JSON string array stored in a single column.
type Allergens []string

func (a Allergens) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *Allergens) Scan(src any) error {
	if src == nil {
		*a = Allergens{}
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return errors.New("allergens: unsupported scan type")
	}
}

// Category groups products on the menu. Display name/description come from the
// translations table; the model carries only language-neutral fields plus a
// transient Name/Description populated by the i18n resolver for responses.
type Category struct {
	Base
	Slug      string `gorm:"size:140;uniqueIndex;not null" json:"slug"`
	SortOrder int    `gorm:"not null;default:0" json:"sort_order"`
	ImageURL  string `gorm:"size:500" json:"image_url,omitempty"`
	IsActive  bool   `gorm:"not null;default:true" json:"is_active"`

	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`

	// Resolved, non-persisted display fields (filled per request locale).
	Name        string `gorm:"-" json:"name"`
	Description string `gorm:"-" json:"description,omitempty"`
}

// Product is a menu item in one of the two pricing modes.
type Product struct {
	Base
	CategoryID        uint64      `gorm:"index;not null" json:"category_id"`
	Slug              string      `gorm:"size:160;uniqueIndex;not null" json:"slug"`
	PricingMode       PricingMode `gorm:"type:enum('unit','weight');not null;default:unit" json:"pricing_mode"`
	BasePriceCents    int64       `gorm:"not null;default:0" json:"base_price_cents"`
	IsPreorder        bool        `gorm:"not null;default:false" json:"is_preorder"`
	PreorderLeadHours int         `gorm:"not null;default:0" json:"preorder_lead_hours"`
	IsAvailable       bool        `gorm:"not null;default:true" json:"is_available"`
	ImageURL          string      `gorm:"size:500" json:"image_url,omitempty"`
	Allergens         Allergens   `gorm:"column:allergens_json;type:json" json:"allergens"`
	SortOrder         int         `gorm:"not null;default:0" json:"sort_order"`

	Variants       []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	ModifierGroups []ModifierGroup  `gorm:"foreignKey:ProductID" json:"modifier_groups,omitempty"`

	Name        string `gorm:"-" json:"name"`
	Description string `gorm:"-" json:"description,omitempty"`
}

// ProductVariant is a size/price option for a per-unit product (L/S).
type ProductVariant struct {
	Base
	ProductID  uint64 `gorm:"index;not null" json:"product_id"`
	SKU        string `gorm:"size:60" json:"sku,omitempty"`
	Label      string `gorm:"size:80;not null" json:"label"`
	PriceCents int64  `gorm:"not null;default:0" json:"price_cents"`
	IsDefault  bool   `gorm:"not null;default:false" json:"is_default"`
	SortOrder  int    `gorm:"not null;default:0" json:"sort_order"`
}

// ModifierGroup is a set of options/add-ons for a product, with select bounds.
type ModifierGroup struct {
	Base
	ProductID  uint64 `gorm:"index;not null" json:"product_id"`
	Label      string `gorm:"size:120;not null" json:"label"`
	MinSelect  int    `gorm:"not null;default:0" json:"min_select"`
	MaxSelect  int    `gorm:"not null;default:1" json:"max_select"`
	IsRequired bool   `gorm:"not null;default:false" json:"is_required"`
	SortOrder  int    `gorm:"not null;default:0" json:"sort_order"`

	Modifiers []Modifier `gorm:"foreignKey:GroupID" json:"modifiers,omitempty"`
}

// Modifier is one option in a group (e.g. Extra Cheese +2.00).
type Modifier struct {
	Base
	GroupID         uint64 `gorm:"index;not null" json:"group_id"`
	Label           string `gorm:"size:120;not null" json:"label"`
	PriceDeltaCents int64  `gorm:"not null;default:0" json:"price_delta_cents"`
	IsDefault       bool   `gorm:"not null;default:false" json:"is_default"`
	IsAvailable     bool   `gorm:"not null;default:true" json:"is_available"`
	SortOrder       int    `gorm:"not null;default:0" json:"sort_order"`
}

// Translation is one localized field value for an entity (generic i18n store).
type Translation struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	EntityType string `gorm:"size:40;not null;uniqueIndex:uq_translation,priority:1" json:"entity_type"`
	EntityID   uint64 `gorm:"not null;uniqueIndex:uq_translation,priority:2" json:"entity_id"`
	Locale     string `gorm:"size:10;not null;uniqueIndex:uq_translation,priority:3" json:"locale"`
	Field      string `gorm:"size:40;not null;uniqueIndex:uq_translation,priority:4" json:"field"`
	Value      string `gorm:"type:text;not null" json:"value"`
}

// Entity type constants for the translations table.
const (
	EntityCategory = "category"
	EntityProduct  = "product"
)

// Translation field constants.
const (
	FieldName        = "name"
	FieldDescription = "description"
)
