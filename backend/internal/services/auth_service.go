package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/auth"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

// AuthService handles registration, login, and token refresh.
type AuthService struct {
	users *repository.UserRepo
	jwt   *auth.Manager
}

func NewAuthService(users *repository.UserRepo, jwt *auth.Manager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

// AuthResult bundles the authenticated user with a fresh token pair.
type AuthResult struct {
	User   *models.User
	Tokens auth.TokenPair
}

// Register creates a customer account and returns tokens. Email uniqueness is
// enforced both here (friendly error) and by the DB unique index (race-safe).
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	email := normalizeEmail(in.Email)

	exists, err := s.users.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Name:         strings.TrimSpace(in.Name),
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleCustomer,
		Status:       models.UserActive,
	}
	if in.Phone != "" {
		// DTO validation already confirmed it parses; normalize to E.164.
		if e164, perr := validator.NormalizePhoneCA(in.Phone); perr == nil {
			u.PhoneE164 = e164
		}
	}

	if err := s.users.Create(ctx, u); err != nil {
		// Translate a unique-violation that slipped past the existence check.
		if repository.IsDuplicateKey(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	logger.FromContext(ctx).Info("user registered", slog.Uint64("user_id", u.ID))
	return s.issue(u)
}

// Login authenticates by email + password.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	u, err := s.users.FindByEmail(ctx, normalizeEmail(in.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Same error as a bad password so we don't leak which emails exist.
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !auth.CheckPassword(u.PasswordHash, in.Password) {
		return nil, ErrInvalidCredentials
	}
	if !u.IsActive() {
		return nil, ErrAccountSuspended
	}
	return s.issue(u)
}

// Refresh validates a refresh token and issues a new token pair. A token is
// rejected if the user's TokenVersion has advanced past the token's (i.e. the
// user logged out everywhere or was force-logged-out).
func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (*AuthResult, error) {
	claims, err := s.jwt.ParseRefresh(in.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	id, err := claims.UserID()
	if err != nil {
		return nil, ErrInvalidToken
	}
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if !u.IsActive() {
		return nil, ErrAccountSuspended
	}
	if claims.TokenVersion != u.TokenVersion {
		return nil, ErrInvalidToken
	}
	return s.issue(u)
}

// Logout invalidates all refresh tokens for the user (bumps TokenVersion).
func (s *AuthService) Logout(ctx context.Context, userID uint64) error {
	return s.users.BumpTokenVersion(ctx, userID)
}

func (s *AuthService) issue(u *models.User) (*AuthResult, error) {
	tokens, err := s.jwt.Issue(u)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: u, Tokens: tokens}, nil
}

func normalizeEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}
