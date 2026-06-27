package models

// Role is the RBAC role of a user. Ordered from least to most privileged for
// readability; comparisons use explicit checks, not ordering.
type Role string

const (
	RoleCustomer   Role = "customer"
	RoleKitchen    Role = "kitchen"
	RoleStaff      Role = "staff"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "super_admin"
)

// AllRoles is the set of valid roles (used for validation).
var AllRoles = []Role{RoleCustomer, RoleKitchen, RoleStaff, RoleAdmin, RoleSuperAdmin}

func (r Role) Valid() bool {
	for _, x := range AllRoles {
		if x == r {
			return true
		}
	}
	return false
}

// IsStaff reports whether the role is any non-customer (back-office) role.
func (r Role) IsStaff() bool {
	return r == RoleKitchen || r == RoleStaff || r == RoleAdmin || r == RoleSuperAdmin
}

// UserStatus controls whether a user can authenticate.
type UserStatus string

const (
	UserActive    UserStatus = "active"
	UserSuspended UserStatus = "suspended"
)

// User is an account. Authentication is email + password; phone is stored in
// E.164 for notifications and delivery contact.
type User struct {
	Base
	Name         string     `gorm:"size:120;not null" json:"name"`
	Email        string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PhoneE164    string     `gorm:"size:20;index" json:"phone,omitempty"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         Role       `gorm:"type:varchar(20);not null;default:customer;index" json:"role"`
	Status       UserStatus `gorm:"type:varchar(20);not null;default:active" json:"status"`

	// TokenVersion is bumped to invalidate all previously issued refresh tokens
	// (logout-everywhere / forced re-auth) without any server-side token store.
	TokenVersion int `gorm:"not null;default:0" json:"-"`

	DefaultAddressID *uint64   `json:"default_address_id,omitempty"`
	Addresses        []Address `gorm:"foreignKey:UserID" json:"addresses,omitempty"`
}

// IsActive reports whether the account may authenticate.
func (u *User) IsActive() bool { return u.Status == UserActive }
