package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/internal/models"
)

// TokenType distinguishes short-lived access tokens from long-lived refresh
// tokens. They are signed with different secrets so an access token can never
// be replayed as a refresh token (or vice versa).
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrWrongType    = errors.New("wrong token type")
)

// Claims is the JWT payload. TokenVersion mirrors the user's TokenVersion at
// issue time; refresh is rejected if the user's current version is higher
// (enables logout-everywhere without a server-side token store).
type Claims struct {
	Role         models.Role `json:"role"`
	Type         TokenType   `json:"typ"`
	TokenVersion int         `json:"tv"`
	jwt.RegisteredClaims
}

// TokenPair is what login/refresh hand back to the client.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access-token lifetime in seconds
	TokenType    string `json:"token_type"` // always "Bearer"
}

// Manager issues and verifies tokens using the configured secrets and TTLs.
type Manager struct {
	cfg config.JWT
	now func() time.Time // injectable for tests
}

func NewManager(cfg config.JWT) *Manager {
	return &Manager{cfg: cfg, now: time.Now}
}

// Issue mints a fresh access+refresh pair for the user.
func (m *Manager) Issue(u *models.User) (TokenPair, error) {
	access, err := m.sign(u, AccessToken, m.cfg.AccessSecret, m.cfg.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := m.sign(u, RefreshToken, m.cfg.RefreshSecret, m.cfg.RefreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(m.cfg.AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (m *Manager) sign(u *models.User, typ TokenType, secret string, ttl time.Duration) (string, error) {
	now := m.now()
	claims := Claims{
		Role:         u.Role,
		Type:         typ,
		TokenVersion: u.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseAccess validates an access token and returns its claims.
func (m *Manager) ParseAccess(raw string) (*Claims, error) {
	return m.parse(raw, AccessToken, m.cfg.AccessSecret)
}

// ParseRefresh validates a refresh token and returns its claims.
func (m *Manager) ParseRefresh(raw string) (*Claims, error) {
	return m.parse(raw, RefreshToken, m.cfg.RefreshSecret)
}

func (m *Manager) parse(raw string, want TokenType, secret string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != want {
		return nil, ErrWrongType
	}
	return claims, nil
}

// UserID extracts the numeric subject from claims.
func (c *Claims) UserID() (uint64, error) {
	return strconv.ParseUint(c.Subject, 10, 64)
}
