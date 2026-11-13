package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
)

// OTPDigits is the length of the numeric one-time codes e-mailed for account
// verification and password reset (Add-on 2026-11-12b).
const OTPDigits = 6

var otpSpace = big.NewInt(1_000_000) // 10^OTPDigits

// NewOTP returns a uniformly random 6-digit code from crypto/rand.
func NewOTP() (string, error) {
	n, err := rand.Int(rand.Reader, otpSpace)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf("%0*d", OTPDigits, n.Int64()), nil
}

// IsOTPFormat reports whether s is exactly OTPDigits ASCII digits.
func IsOTPFormat(s string) bool {
	if len(s) != OTPDigits {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// OTPKey derives the HMAC key for one-time codes from the refresh secret, so
// no extra secret has to be provisioned. Rotating the refresh secret simply
// invalidates codes that are still pending (they live 15 minutes).
func (m *Manager) OTPKey() []byte {
	mac := hmac.New(sha256.New, m.refreshSecret)
	mac.Write([]byte("webond-otp-v1"))
	return mac.Sum(nil)
}

// HashOTP binds a code to its user and purpose (e.g. "verify", "reset") and
// returns the hex HMAC-SHA256 stored instead of the code itself.
func HashOTP(key []byte, userID uint64, purpose, code string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(strconv.FormatUint(userID, 10) + "|" + purpose + "|" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

// OTPMatches compares a submitted code with a stored hash in constant time.
func OTPMatches(key []byte, userID uint64, purpose, code, storedHash string) bool {
	return hmac.Equal([]byte(HashOTP(key, userID, purpose, code)), []byte(storedHash))
}
