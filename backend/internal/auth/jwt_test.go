package auth

import (
	"testing"
	"time"

	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/internal/models"
)

func testManager() *Manager {
	return NewManager(config.JWT{
		AccessSecret:  "access-secret-for-tests-only-123456",
		RefreshSecret: "refresh-secret-for-tests-only-123456",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
}

func TestIssueAndParseRoundTrip(t *testing.T) {
	m := testManager()
	u := &models.User{Base: models.Base{ID: 42}, Role: models.RoleAdmin, TokenVersion: 3}

	pair, err := m.Issue(u)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	ac, err := m.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if id, _ := ac.UserID(); id != 42 {
		t.Errorf("access subject = %d, want 42", id)
	}
	if ac.Role != models.RoleAdmin {
		t.Errorf("role = %q, want admin", ac.Role)
	}
	if ac.TokenVersion != 3 {
		t.Errorf("token version = %d, want 3", ac.TokenVersion)
	}

	rc, err := m.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefresh: %v", err)
	}
	if rc.Type != RefreshToken {
		t.Errorf("refresh type = %q", rc.Type)
	}
}

func TestAccessTokenRejectedAsRefresh(t *testing.T) {
	m := testManager()
	u := &models.User{Base: models.Base{ID: 1}, Role: models.RoleCustomer}
	pair, _ := m.Issue(u)

	if _, err := m.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("expected access token to be rejected by ParseRefresh")
	}
	if _, err := m.ParseAccess(pair.RefreshToken); err == nil {
		t.Fatal("expected refresh token to be rejected by ParseAccess")
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	m := testManager()
	m.now = func() time.Time { return time.Now().Add(-1 * time.Hour) } // issue in the past
	u := &models.User{Base: models.Base{ID: 1}, Role: models.RoleCustomer}
	pair, _ := m.Issue(u)

	m.now = time.Now // verify at the present
	if _, err := m.ParseAccess(pair.AccessToken); err == nil {
		t.Fatal("expected expired access token to be rejected")
	}
}

func TestTamperedTokenRejected(t *testing.T) {
	m := testManager()
	u := &models.User{Base: models.Base{ID: 1}, Role: models.RoleCustomer}
	pair, _ := m.Issue(u)

	if _, err := m.ParseAccess(pair.AccessToken + "x"); err == nil {
		t.Fatal("expected tampered token to be rejected")
	}
}

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("supersecret1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !CheckPassword(hash, "supersecret1") {
		t.Error("correct password should verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("wrong password should not verify")
	}
}
