package services

// Request DTOs. Handlers bind JSON into these and run them through the
// validator before calling a service. Keeping them here makes each service's
// input contract explicit and testable.

type RegisterInput struct {
	Name     string `json:"name" validate:"required,min=2,max=120"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Phone    string `json:"phone" validate:"omitempty,e164ca"`
	Password string `json:"password" validate:"required,min=8,max=72"` // bcrypt input cap is 72 bytes
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UpdateProfileInput struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=120"`
	Phone string `json:"phone" validate:"omitempty,e164ca"`
}

type ChangeRoleInput struct {
	Role string `json:"role" validate:"required,role"`
}

type AddressInput struct {
	Label        string   `json:"label" validate:"omitempty,max=60"`
	Line1        string   `json:"line1" validate:"required,max=200"`
	Line2        string   `json:"line2" validate:"omitempty,max=200"`
	City         string   `json:"city" validate:"required,max=120"`
	ProvinceCode string   `json:"province_code" validate:"required,ca_province"`
	PostalCode   string   `json:"postal_code" validate:"required,ca_postal"`
	Lat          *float64 `json:"lat" validate:"omitempty,latitude"`
	Lng          *float64 `json:"lng" validate:"omitempty,longitude"`
	Phone        string   `json:"phone" validate:"omitempty,e164ca"`
	Notes        string   `json:"notes" validate:"omitempty,max=255"`
}
