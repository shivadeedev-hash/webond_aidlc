package auth

import (
	"testing"
	"time"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("S3cret!pass")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "S3cret!pass" {
		t.Fatal("hash must not equal plaintext")
	}
	if !VerifyPassword(hash, "S3cret!pass") {
		t.Error("VerifyPassword returned false for correct password")
	}
	if VerifyPassword(hash, "wrong") {
		t.Error("VerifyPassword returned true for wrong password")
	}
}

func TestHashPassword_Empty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestBcryptCost(t *testing.T) {
	if BcryptCost < 12 {
		t.Fatalf("BcryptCost = %d, policy requires ≥ 12", BcryptCost)
	}
}

func TestAccessTokenIssueVerify(t *testing.T) {
	m := NewManager("access", "refresh")
	tok, err := m.IssueAccessToken(42, RoleInvestor)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := m.VerifyAccessToken(tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Role != RoleInvestor {
		t.Errorf("role = %q, want investor", claims.Role)
	}
	uid, err := claims.UserID()
	if err != nil || uid != 42 {
		t.Errorf("UserID = %d (err %v), want 42", uid, err)
	}
}

func TestAccessTokenTTL(t *testing.T) {
	if AccessTokenTTL != 15*time.Minute {
		t.Errorf("AccessTokenTTL = %v, want 15m", AccessTokenTTL)
	}
	if RefreshTokenTTL != 7*24*time.Hour {
		t.Errorf("RefreshTokenTTL = %v, want 7d", RefreshTokenTTL)
	}
}

func TestTokenTypeIsolation(t *testing.T) {
	m := NewManager("access", "refresh")
	// A refresh token must not verify as an access token.
	refresh, err := m.IssueRefreshToken(7, RoleAdmin)
	if err != nil {
		t.Fatalf("issue refresh: %v", err)
	}
	if _, err := m.VerifyAccessToken(refresh); err == nil {
		t.Fatal("refresh token wrongly accepted as access token")
	}
	// And the refresh token verifies on the refresh path.
	if _, err := m.VerifyRefreshToken(refresh); err != nil {
		t.Fatalf("verify refresh: %v", err)
	}
}

func TestWrongSecretRejected(t *testing.T) {
	issuer := NewManager("access", "refresh")
	other := NewManager("different", "different")
	tok, _ := issuer.IssueAccessToken(1, RoleInvestor)
	if _, err := other.VerifyAccessToken(tok); err == nil {
		t.Fatal("token signed with a different secret was accepted")
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	m := NewManager("access", "refresh")
	// Force the clock into the past so the issued token is already expired.
	m.now = func() time.Time { return time.Now().Add(-1 * time.Hour) }
	tok, err := m.IssueAccessToken(1, RoleInvestor)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	m.now = time.Now
	if _, err := m.VerifyAccessToken(tok); err == nil {
		t.Fatal("expired token was accepted")
	}
}
