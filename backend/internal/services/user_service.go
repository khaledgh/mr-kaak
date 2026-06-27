package services

import (
	"context"
	"errors"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/canada"
)

// UserService handles profile, address management, and admin user operations.
type UserService struct {
	users     *repository.UserRepo
	addresses *repository.AddressRepo
}

func NewUserService(users *repository.UserRepo, addresses *repository.AddressRepo) *UserService {
	return &UserService{users: users, addresses: addresses}
}

// GetByID loads a user, translating repo not-found to the domain error.
func (s *UserService) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	u, err := s.users.FindByID(ctx, id)
	return mapUserErr(u, err)
}

// UpdateProfile updates the caller's own name/phone.
func (s *UserService) UpdateProfile(ctx context.Context, id uint64, in UpdateProfileInput) (*models.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if u, err = mapUserErr(u, err); err != nil {
		return nil, err
	}
	if in.Name != "" {
		u.Name = strings.TrimSpace(in.Name)
	}
	if in.Phone != "" {
		if e164, perr := validator.NormalizePhoneCA(in.Phone); perr == nil {
			u.PhoneE164 = e164
		}
	}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// --- Admin operations ---

// ListUsers returns a paginated, filtered user listing with the total count.
func (s *UserService) ListUsers(ctx context.Context, opts repository.UserListOptions) ([]models.User, int64, error) {
	return s.users.List(ctx, opts)
}

// ChangeRole sets a user's RBAC role.
func (s *UserService) ChangeRole(ctx context.Context, id uint64, role models.Role) (*models.User, error) {
	if !role.Valid() {
		return nil, ErrForbidden
	}
	if err := s.users.UpdateFields(ctx, id, map[string]any{"role": role}); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

// SetStatus suspends or reactivates a user. Suspending also invalidates their
// refresh tokens so existing sessions can't be silently extended.
func (s *UserService) SetStatus(ctx context.Context, id uint64, status models.UserStatus) error {
	if err := s.users.UpdateFields(ctx, id, map[string]any{"status": status}); err != nil {
		return err
	}
	if status == models.UserSuspended {
		return s.users.BumpTokenVersion(ctx, id)
	}
	return nil
}

// --- Addresses ---

// ListAddresses returns a user's addresses.
func (s *UserService) ListAddresses(ctx context.Context, userID uint64) ([]models.Address, error) {
	return s.addresses.ListByUser(ctx, userID)
}

// AddAddress creates an address for the user (postal code normalized to
// canonical form). If it's the user's first address, it becomes the default.
func (s *UserService) AddAddress(ctx context.Context, userID uint64, in AddressInput) (*models.Address, error) {
	a := buildAddress(userID, in)
	if err := s.addresses.Create(ctx, &a); err != nil {
		return nil, err
	}

	existing, err := s.addresses.ListByUser(ctx, userID)
	if err == nil && len(existing) == 1 {
		_ = s.addresses.SetDefault(ctx, userID, a.ID)
	}
	return &a, nil
}

// UpdateAddress edits an address the user owns.
func (s *UserService) UpdateAddress(ctx context.Context, userID, addressID uint64, in AddressInput) (*models.Address, error) {
	a, err := s.addresses.FindByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if a.UserID != userID {
		return nil, ErrForbidden
	}

	updated := buildAddress(userID, in)
	updated.ID = a.ID
	if err := s.addresses.Update(ctx, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteAddress removes an address the user owns.
func (s *UserService) DeleteAddress(ctx context.Context, userID, addressID uint64) error {
	if err := s.addresses.Delete(ctx, userID, addressID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// SetDefaultAddress marks one of the user's addresses as default.
func (s *UserService) SetDefaultAddress(ctx context.Context, userID, addressID uint64) error {
	if err := s.addresses.SetDefault(ctx, userID, addressID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func buildAddress(userID uint64, in AddressInput) models.Address {
	return models.Address{
		UserID:       userID,
		Label:        strings.TrimSpace(in.Label),
		Line1:        strings.TrimSpace(in.Line1),
		Line2:        strings.TrimSpace(in.Line2),
		City:         strings.TrimSpace(in.City),
		ProvinceCode: strings.ToUpper(strings.TrimSpace(in.ProvinceCode)),
		PostalCode:   canada.NormalizePostal(in.PostalCode),
		CountryCode:  "CA",
		Lat:          in.Lat,
		Lng:          in.Lng,
		PhoneE164:    normalizePhone(in.Phone),
		Notes:        strings.TrimSpace(in.Notes),
	}
}

func normalizePhone(raw string) string {
	if raw == "" {
		return ""
	}
	if e164, err := validator.NormalizePhoneCA(raw); err == nil {
		return e164
	}
	return ""
}

func mapUserErr(u *models.User, err error) (*models.User, error) {
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}
