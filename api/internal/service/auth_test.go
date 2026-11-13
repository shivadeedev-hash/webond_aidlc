package service

import (
	"context"
	"testing"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
)

func TestAuth_RegisterVerifyLogin(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	notify := NewNotificationService(store, clock)
	svc := NewAuthService(store, testTokens(), notify, clock)
	ctx := context.Background()

	email := uniqueEmail()
	reg, err := svc.Register(ctx, RegisterInput{
		Email: email, Password: "secret123", FullName: "Alice", Mobile: "0800000000", Consent: true,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, reg.UserID) })
	if len(reg.VerifyCode) != 6 {
		t.Fatalf("expected a 6-digit verify code, got %q", reg.VerifyCode)
	}

	// Login before verification is blocked (BR-01).
	if _, err := svc.Login(ctx, email, "secret123"); err == nil {
		t.Fatal("expected login to be blocked before verification")
	}

	// Verify activates the account (US-002).
	if err := svc.Verify(ctx, email, reg.VerifyCode); err != nil {
		t.Fatalf("verify: %v", err)
	}
	u, _ := store.Users.GetByEmail(ctx, email)
	if u.Status != db.UsersStatusActive {
		t.Fatalf("status after verify: got %s want active", u.Status)
	}

	// Login now succeeds and issues both tokens (US-003).
	res, err := svc.Login(ctx, email, "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("expected access + refresh tokens")
	}

	// Refresh rotates an access token (EX-03) and returns the user so the SPA
	// can restore its session after a reload.
	refreshed, err := svc.Refresh(ctx, res.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Error("refresh must return an access token")
	}
	if refreshed.User.Email != email {
		t.Errorf("refresh user email = %q, want %q", refreshed.User.Email, email)
	}
	// A garbage refresh token is rejected.
	if _, err := svc.Refresh(ctx, "not-a-token"); err == nil {
		t.Error("expected an error for an invalid refresh token")
	}
}

func TestAuth_DuplicateEmail(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	svc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)
	ctx := context.Background()

	email := uniqueEmail()
	reg, err := svc.Register(ctx, RegisterInput{Email: email, Password: "secret123", FullName: "Bob", Consent: true})
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, reg.UserID) })

	_, err = svc.Register(ctx, RegisterInput{Email: email, Password: "secret123", FullName: "Bob2", Consent: true})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeDuplicateEmail {
		t.Fatalf("expected DUPLICATE_EMAIL, got %v", err)
	}
}

func TestAuth_LockoutAfterFiveFailures(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	svc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)
	ctx := context.Background()

	email := uniqueEmail()
	reg, err := svc.Register(ctx, RegisterInput{Email: email, Password: "secret123", FullName: "Carol", Consent: true})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, reg.UserID) })
	_ = svc.Verify(ctx, email, reg.VerifyCode)

	// 5 wrong attempts → lock (EX-01).
	for i := 0; i < maxFailedLogins; i++ {
		if _, err := svc.Login(ctx, email, "wrong-password"); err == nil {
			t.Fatalf("attempt %d: expected failure", i)
		}
	}
	u, _ := store.Users.GetByEmail(ctx, email)
	if u.Status != db.UsersStatusLocked {
		t.Fatalf("expected locked status, got %s", u.Status)
	}
	// Even the correct password is rejected while locked.
	if _, err := svc.Login(ctx, email, "secret123"); err == nil {
		t.Fatal("expected locked account to reject correct password")
	}

	// After the lockout window elapses, correct password works again.
	clock.Advance(lockoutDuration + 1)
	if _, err := svc.Login(ctx, email, "secret123"); err != nil {
		t.Fatalf("login after lockout window: %v", err)
	}
}

func TestAuth_ForgotAndReset(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	svc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)
	ctx := context.Background()

	email := uniqueEmail()
	reg, err := svc.Register(ctx, RegisterInput{Email: email, Password: "secret123", FullName: "Dave", Consent: true})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, reg.UserID) })
	_ = svc.Verify(ctx, email, reg.VerifyCode)

	// Forgot for an unknown email still succeeds silently (no enumeration).
	if _, err := svc.ForgotPassword(ctx, "nobody-"+email); err != nil {
		t.Fatalf("forgot unknown: %v", err)
	}

	token, err := svc.ForgotPassword(ctx, email)
	if err != nil || len(token) != 6 {
		t.Fatalf("forgot: code=%q err=%v", token, err)
	}
	if err := svc.ResetPassword(ctx, email, token, "newsecret9"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	// Old password no longer works; new one does.
	if _, err := svc.Login(ctx, email, "secret123"); err == nil {
		t.Fatal("old password should be rejected")
	}
	if _, err := svc.Login(ctx, email, "newsecret9"); err != nil {
		t.Fatalf("new password login: %v", err)
	}
	// Reusing the consumed token is rejected.
	if err := svc.ResetPassword(ctx, email, token, "another99"); err == nil {
		t.Fatal("expected consumed reset token to be rejected")
	}
}

// asApp is a tiny errors.As wrapper kept local to avoid importing errors in
// every test file.
func asApp(err error, target **domain.AppError) bool {
	if e, ok := err.(*domain.AppError); ok {
		*target = e
		return true
	}
	return false
}
