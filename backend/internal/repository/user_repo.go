package repository

import (
	"context"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a queried row does not exist. Services translate
// it into a domain-level not-found error.
var ErrNotFound = errors.New("record not found")

// UserRepo is the data-access layer for users. All user SQL lives here.
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// Create inserts a new user.
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// FindByID loads a user by primary key.
func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &u, nil
}

// FindByEmail loads a user by their unique email.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &u, nil
}

// EmailExists reports whether an account already uses the email.
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Update persists changed columns of an existing user.
func (r *UserRepo) Update(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// UpdateFields updates a targeted set of columns (avoids overwriting unrelated
// fields and skips zero-value pitfalls of Save).
func (r *UserRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.User{Base: models.Base{ID: id}}).Updates(fields).Error
}

// BumpTokenVersion invalidates all refresh tokens for a user (logout-all).
func (r *UserRepo) BumpTokenVersion(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
}

// SoftDelete marks the user deleted (GORM sets deleted_at).
func (r *UserRepo) SoftDelete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// List returns a page of users (optionally filtered by role / search term) and
// the total count for pagination metadata.
func (r *UserRepo) List(ctx context.Context, opts UserListOptions) ([]models.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.User{})
	if opts.Role != "" {
		q = q.Where("role = ?", opts.Role)
	}
	if opts.Search != "" {
		like := "%" + opts.Search + "%"
		q = q.Where("name LIKE ? OR email LIKE ? OR phone_e164 LIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	err := q.Order("id DESC").Limit(opts.Limit).Offset(opts.Offset).Find(&users).Error
	return users, total, err
}

// UserListOptions filters and paginates a user listing.
type UserListOptions struct {
	Role   string
	Search string
	Limit  int
	Offset int
}
