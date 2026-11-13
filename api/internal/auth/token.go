// Package auth provides JWT access/refresh token issue+verify and password
// hashing helpers (components.md § AuthService; api-spec.md § Auth). Secrets
// are injected from config and never hardcoded (NFR-01).
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token lifetimes per api-spec.md § Auth.
const (
	AccessTokenTTL  = 15 * time.Minute   // short-lived access token
	RefreshTokenTTL = 7 * 24 * time.Hour // 7-day refresh token
)

// Role is the RBAC role embedded in the access token.
type Role string

const (
	RoleInvestor Role = "investor"
	RoleAdmin    Role = "admin"
)

// TokenType distinguishes access and refresh tokens so a refresh token can
// never be presented as a bearer access token and vice versa.
type TokenType string

const (
	tokenTypeAccess  TokenType = "access"
	tokenTypeRefresh TokenType = "refresh"
)

// ErrInvalidToken is returned when a token fails signature, expiry, or type
// validation. The caller maps this to a 401 UNAUTHORIZED.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims is the JWT payload for both token types. Subject carries the user id.
type Claims struct {
	Role    Role      `json:"role"`
	TokType TokenType `json:"typ"`
	jwt.RegisteredClaims
}

// Manager issues and verifies JWTs using separate access/refresh secrets.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	now           func() time.Time // injectable clock for tests
}

// NewManager builds a token Manager from the configured secrets.
func NewManager(accessSecret, refreshSecret string) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		now:           time.Now,
	}
}

// IssueAccessToken mints a signed access token for the user and role.
func (m *Manager) IssueAccessToken(userID int64, role Role) (string, error) {
	return m.issue(userID, role, tokenTypeAccess, AccessTokenTTL, m.accessSecret)
}

// IssueRefreshToken mints a signed refresh token for the user.
func (m *Manager) IssueRefreshToken(userID int64, role Role) (string, error) {
	return m.issue(userID, role, tokenTypeRefresh, RefreshTokenTTL, m.refreshSecret)
}

func (m *Manager) issue(userID int64, role Role, typ TokenType, ttl time.Duration, secret []byte) (string, error) {
	now := m.now()
	claims := Claims{
		Role:    role,
		TokType: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign %s token: %w", typ, err)
	}
	return signed, nil
}

// VerifyAccessToken validates an access token and returns its claims.
func (m *Manager) VerifyAccessToken(raw string) (*Claims, error) {
	return m.verify(raw, tokenTypeAccess, m.accessSecret)
}

// VerifyRefreshToken validates a refresh token and returns its claims.
func (m *Manager) VerifyRefreshToken(raw string) (*Claims, error) {
	return m.verify(raw, tokenTypeRefresh, m.refreshSecret)
}

func (m *Manager) verify(raw string, want TokenType, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		// Reject any algorithm other than the HMAC we signed with (alg confusion).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokType != want {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// UserID parses the numeric user id from the token subject.
func (c *Claims) UserID() (int64, error) {
	var id int64
	if _, err := fmt.Sscanf(c.Subject, "%d", &id); err != nil {
		return 0, fmt.Errorf("parse subject: %w", err)
	}
	return id, nil
}
