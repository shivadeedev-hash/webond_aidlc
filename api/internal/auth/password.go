package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the bcrypt work factor. Security policy requires cost ≥ 12
// (strong-authentication-and-credential-management); 12 balances strength and
// login latency.
const BcryptCost = 12

// HashPassword returns a bcrypt hash of the plaintext password. The plaintext
// is never logged or stored (operations.md § Sensitive Data Rules).
func HashPassword(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("password must not be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword reports whether the plaintext matches the stored bcrypt hash.
// It returns false (not an error) on mismatch so callers emit a generic
// invalid-credentials response (EX-01, no user enumeration).
func VerifyPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
