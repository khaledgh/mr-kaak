// Package auth handles credential hashing and JWT issuance/verification.
// It is independent of HTTP and storage so it can be unit-tested in isolation.
package auth

import "golang.org/x/crypto/bcrypt"

// bcryptCost balances security and login latency. 12 is a sensible default
// for an interactive API in 2026.
const bcryptCost = 12

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
