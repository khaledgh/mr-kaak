package services

// LanguageInput is the admin create/update payload for a locale.
type LanguageInput struct {
	Code       string `json:"code" validate:"required,min=2,max=10"`
	Name       string `json:"name" validate:"required,max=80"`
	NativeName string `json:"native_name" validate:"required,max=80"`
	IsRTL      bool   `json:"is_rtl"`
	IsActive   *bool  `json:"is_active"`
	SortOrder  int    `json:"sort_order" validate:"omitempty,min=0"`
}

// BundleInput upserts UI strings for one locale.
type BundleInput struct {
	Strings map[string]string `json:"strings" validate:"required,min=1"`
}
