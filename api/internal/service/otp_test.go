package service

import (
	"context"
	"strings"
	"testing"

	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/db"
)

// wrongCode returns a valid-format code different from c.
func wrongCode(c string) string {
	if c == "000000" {
		return "111111"
	}
	return "000000"
}

func newOTPFixture(t *testing.T) (*AuthService, *FixedClock, string, RegisterResult) {
	t.Helper()
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	svc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)
	email := uniqueEmail()
	reg, err := svc.Register(context.Background(), RegisterInput{Email: email, Password: "secret123", FullName: "OTP", Consent: true})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, reg.UserID) })
	return svc, clock, email, reg
}

func TestOTP_FormatAndStorage(t *testing.T) {
	for i := 0; i < 50; i++ {
		c, err := auth.NewOTP()
		if err != nil || !auth.IsOTPFormat(c) {
			t.Fatalf("bad otp %q (%v)", c, err)
		}
	}
	for _, bad := range []string{"", "12345", "1234567", "12a456", " 12345"} {
		if auth.IsOTPFormat(bad) {
			t.Errorf("%q must not be a valid OTP", bad)
		}
	}

	svc, _, _, reg := newOTPFixture(t)
	pool := rawPool(t)
	var token string
	if err := pool.QueryRow("SELECT token FROM notification_outbox WHERE user_id = ? AND category = 'verify'", reg.UserID).Scan(&token); err != nil {
		t.Fatalf("read token: %v", err)
	}
	if token == reg.VerifyCode || len(token) != 64 {
		t.Fatalf("token column must hold an HMAC, got %q", token)
	}
	_ = svc
}

func TestOTP_WrongCodeCountsAndLocksAfterFive(t *testing.T) {
	svc, _, email, reg := newOTPFixture(t)
	ctx := context.Background()
	bad := wrongCode(reg.VerifyCode)

	for i := 1; i <= otpMaxAttempts; i++ {
		err := svc.Verify(ctx, email, bad)
		if err == nil {
			t.Fatalf("attempt %d: wrong code accepted", i)
		}
		if i < otpMaxAttempts && !strings.Contains(err.Error(), "attempt(s) left") {
			t.Fatalf("attempt %d: want remaining-attempts message, got %v", i, err)
		}
	}
	// The right code is now refused too: the code is dead after 5 wrong tries.
	if err := svc.Verify(ctx, email, reg.VerifyCode); err == nil || !strings.Contains(err.Error(), "Too many") {
		t.Fatalf("code must be dead after %d wrong attempts, got %v", otpMaxAttempts, err)
	}
}

func TestOTP_ExpiresAfter15Minutes(t *testing.T) {
	svc, clock, email, reg := newOTPFixture(t)
	clock.Advance(otpTTL + 1)
	if err := svc.Verify(context.Background(), email, reg.VerifyCode); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expiry error, got %v", err)
	}
}

func TestOTP_ResendCooldownAndInvalidatesOldCode(t *testing.T) {
	svc, clock, email, reg := newOTPFixture(t)
	ctx := context.Background()

	// Within the cooldown nothing new is sent.
	if code, err := svc.ResendVerification(ctx, email); err != nil || code != "" {
		t.Fatalf("resend inside cooldown must do nothing, got %q %v", code, err)
	}
	clock.Advance(otpResendCooldown + 1)
	newCode, err := svc.ResendVerification(ctx, email)
	if err != nil || len(newCode) != 6 {
		t.Fatalf("resend after cooldown: %q %v", newCode, err)
	}
	// The first code no longer works (unless both happen to be equal).
	if newCode != reg.VerifyCode {
		if err := svc.Verify(ctx, email, reg.VerifyCode); err == nil {
			t.Fatal("old code must be invalid after a resend")
		}
	}
	if err := svc.Verify(ctx, email, newCode); err != nil {
		t.Fatalf("new code: %v", err)
	}
	// Already verified: repeat is a no-op, and resend sends nothing.
	if err := svc.Verify(ctx, email, newCode); err != nil {
		t.Fatalf("repeat verify should be idempotent: %v", err)
	}
	if code, _ := svc.ResendVerification(ctx, email); code != "" {
		t.Fatal("no code may be sent to an active account")
	}
}

func TestOTP_UnknownEmailAndCrossPurpose(t *testing.T) {
	svc, clock, email, reg := newOTPFixture(t)
	ctx := context.Background()

	if err := svc.Verify(ctx, "nobody-"+email, reg.VerifyCode); err == nil {
		t.Fatal("unknown e-mail must fail")
	}
	if code, err := svc.ResendVerification(ctx, "nobody-"+email); err != nil || code != "" {
		t.Fatalf("unknown e-mail resend must silently do nothing: %q %v", code, err)
	}
	if err := svc.Verify(ctx, email, reg.VerifyCode); err != nil {
		t.Fatalf("verify: %v", err)
	}
	// A verification code can never reset a password.
	clock.Advance(otpResendCooldown + 1)
	reset, err := svc.ForgotPassword(ctx, email)
	if err != nil || reset == "" {
		t.Fatalf("forgot: %q %v", reset, err)
	}
	if err := svc.ResetPassword(ctx, email, reg.VerifyCode, "newsecret9"); err == nil && reg.VerifyCode != reset {
		t.Fatal("a verify code must not work for password reset")
	}
	// Forgot again inside the cooldown sends nothing (anti mail-bombing).
	if again, _ := svc.ForgotPassword(ctx, email); again != "" {
		t.Fatal("forgot-password inside the cooldown must not send another code")
	}
	_ = db.NotificationOutboxCategoryReset
}

// TestOTP_ConcurrentRegistrationsAndResends: many people registering and asking
// for codes at the same time must not deadlock or fail (notification_outbox
// locking in _issueOTP / _issueOTPWithCooldown).
func TestOTP_ConcurrentRegistrationsAndResends(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	svc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)
	ctx := context.Background()

	const n = 20
	type out struct {
		id    uint64
		email string
		err   error
	}
	results := make(chan out, n)
	for i := 0; i < n; i++ {
		go func() {
			email := uniqueEmail()
			reg, err := svc.Register(ctx, RegisterInput{Email: email, Password: "secret123", FullName: "C", Consent: true})
			results <- out{reg.UserID, email, err}
		}()
	}
	var emails []string
	for i := 0; i < n; i++ {
		r := <-results
		if r.err != nil {
			t.Errorf("concurrent register failed: %v", r.err)
			continue
		}
		emails = append(emails, r.email)
		id := r.id
		t.Cleanup(func() { cleanupUser(t, pool, id) })
	}

	clock.Advance(otpResendCooldown + 1)
	errs := make(chan error, 2*len(emails))
	for _, e := range emails {
		for k := 0; k < 2; k++ { // two resends per user at once: one sends, one is a no-op
			go func(email string) {
				_, err := svc.ResendVerification(ctx, email)
				errs <- err
			}(e)
		}
	}
	for i := 0; i < 2*len(emails); i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent resend failed: %v", err)
		}
	}
	// Each user ends with exactly one open code.
	for _, e := range emails {
		var open int
		if err := pool.QueryRow(`SELECT COUNT(*) FROM notification_outbox n JOIN users u ON u.id = n.user_id
			WHERE u.email = ? AND n.category = 'verify' AND n.token IS NOT NULL AND n.read_at IS NULL`, e).Scan(&open); err != nil {
			t.Fatal(err)
		}
		if open != 1 {
			t.Errorf("%s has %d open codes, want 1", e, open)
		}
	}
}
