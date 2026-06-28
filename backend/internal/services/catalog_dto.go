package services

// LocalizedText is a name/description pair for one locale.
type LocalizedText struct {
	Name        string `json:"name" validate:"required,max=200"`
	Description string `json:"description" validate:"omitempty,max=2000"`
}

// CategoryInput is the admin create/update payload for a category.
type CategoryInput struct {
	Slug      string `json:"slug" validate:"required,max=140,slug"`
	SortOrder int    `json:"sort_order" validate:"omitempty,min=0"`
	ImageURL  string `json:"image_url" validate:"omitempty,max=500"`
	IsActive  *bool  `json:"is_active"`
	// Translations maps locale -> {name, description}. The default locale entry
	// is required so a label is never empty.
	Translations map[string]LocalizedText `json:"translations" validate:"required,min=1,dive"`
}

// VariantInput is a size/price option.
type VariantInput struct {
	SKU        string `json:"sku" validate:"omitempty,max=60"`
	Label      string `json:"label" validate:"required,max=80"`
	PriceCents int64  `json:"price_cents" validate:"min=0"`
	IsDefault  bool   `json:"is_default"`
	SortOrder  int    `json:"sort_order" validate:"omitempty,min=0"`
}

// ModifierInput is one option within a modifier group.
type ModifierInput struct {
	Label           string `json:"label" validate:"required,max=120"`
	PriceDeltaCents int64  `json:"price_delta_cents"`
	IsDefault       bool   `json:"is_default"`
	IsAvailable     *bool  `json:"is_available"`
	SortOrder       int    `json:"sort_order" validate:"omitempty,min=0"`
}

// ModifierGroupInput is a group of options with selection bounds.
type ModifierGroupInput struct {
	Label      string          `json:"label" validate:"required,max=120"`
	MinSelect  int             `json:"min_select" validate:"min=0"`
	MaxSelect  int             `json:"max_select" validate:"min=0"`
	IsRequired bool            `json:"is_required"`
	SortOrder  int             `json:"sort_order" validate:"omitempty,min=0"`
	Modifiers  []ModifierInput `json:"modifiers" validate:"dive"`
}

// ProductInput is the admin create/update payload for a product.
type ProductInput struct {
	CategoryID        uint64                   `json:"category_id" validate:"required"`
	Slug              string                   `json:"slug" validate:"required,max=160,slug"`
	PricingMode       string                   `json:"pricing_mode" validate:"required,oneof=unit weight"`
	BasePriceCents    int64                    `json:"base_price_cents" validate:"min=0"`
	IsPreorder        bool                     `json:"is_preorder"`
	PreorderLeadHours int                      `json:"preorder_lead_hours" validate:"min=0"`
	IsAvailable       *bool                    `json:"is_available"`
	ImageURL          string                   `json:"image_url" validate:"omitempty,max=500"`
	Allergens         []string                 `json:"allergens" validate:"omitempty,dive,max=40"`
	SortOrder         int                      `json:"sort_order" validate:"omitempty,min=0"`
	Translations      map[string]LocalizedText `json:"translations" validate:"required,min=1,dive"`
	Variants          []VariantInput           `json:"variants" validate:"dive"`
	ModifierGroups    []ModifierGroupInput     `json:"modifier_groups" validate:"dive"`
}

// AvailabilityInput toggles a product's availability.
type AvailabilityInput struct {
	IsAvailable bool `json:"is_available"`
}

// AdminCategoryDTO is the full admin view of a category: all locales' translations
// are included instead of a single resolved locale.
type AdminCategoryDTO struct {
	ID           uint64                   `json:"id"`
	Slug         string                   `json:"slug"`
	SortOrder    int                      `json:"sort_order"`
	ImageURL     string                   `json:"image_url,omitempty"`
	IsActive     bool                     `json:"is_active"`
	Translations map[string]LocalizedText `json:"translations"`
}

// AdminProductDTO is the full admin view of a product with all translations.
type AdminProductDTO struct {
	ID                uint64                   `json:"id"`
	CategoryID        uint64                   `json:"category_id"`
	Slug              string                   `json:"slug"`
	PricingMode       string                   `json:"pricing_mode"`
	BasePriceCents    int64                    `json:"base_price_cents"`
	IsPreorder        bool                     `json:"is_preorder"`
	PreorderLeadHours int                      `json:"preorder_lead_hours"`
	IsAvailable       bool                     `json:"is_available"`
	ImageURL          string                   `json:"image_url,omitempty"`
	Allergens         []string                 `json:"allergens"`
	SortOrder         int                      `json:"sort_order"`
	Translations      map[string]LocalizedText `json:"translations"`
	Variants          any                      `json:"variants"`
	ModifierGroups    any                      `json:"modifier_groups"`
}
