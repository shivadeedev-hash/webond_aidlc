package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/mail"
)

// fakeSender records messages and can be told to fail.
type fakeSender struct {
	mu   sync.Mutex
	sent []mail.Message
	fail error
}

func (f *fakeSender) Name() string { return "fake" }
func (f *fakeSender) Send(_ context.Context, m mail.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail != nil {
		return f.fail
	}
	f.sent = append(f.sent, m)
	return nil
}

func TestComposeEmailText(t *testing.T) {
	v := ComposeEmailText(db.NotificationOutboxCategoryVerify, "code 123456", "http://localhost:5173")
	if v != "code 123456" || strings.Contains(v, "http") {
		t.Fatalf("OTP e-mail must be sent as-is with no link: %q", v)
	}
	o := ComposeEmailText(db.NotificationOutboxCategoryOrder, "Your order completed.", "https://app.example.com")
	if !strings.HasPrefix(o, "Your order completed.") || !strings.Contains(o, "https://app.example.com") {
		t.Fatalf("event text wrong: %q", o)
	}
}

func TestTruncate_KeepsUTF8Valid(t *testing.T) {
	s := strings.Repeat("ก", 100) // 3 bytes per rune
	got := truncate(s, 10)
	if len(got) > 10 || len(got)%3 != 0 {
		t.Fatalf("bad truncation length %d", len(got))
	}
}

// TestEmailDispatcher_RegisterDeliversOnce registers a user and checks the
// verification mail is sent exactly once with a working link (Task 10.8).
func TestEmailDispatcher_RegisterDeliversOnce(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ctx := context.Background()
	authSvc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)

	email := uniqueEmail()
	res, err := authSvc.Register(ctx, RegisterInput{Email: email, Password: "Passw0rd!", FullName: "Mail Test", Consent: true})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, res.UserID) })

	fs := &fakeSender{}
	d := NewEmailDispatcher(store, fs, clock, quietLogger(), "http://localhost:5173", 0)

	// Other tests may leave pending rows in the shared test DB; drain until our
	// message shows up.
	for i := 0; i < 50; i++ {
		n, err := d.DeliverPending(ctx)
		if err != nil {
			t.Fatalf("deliver: %v", err)
		}
		if n == 0 {
			break
		}
	}
	var mine []mail.Message
	for _, m := range fs.sent {
		if m.To == email {
			mine = append(mine, m)
		}
	}
	if len(mine) != 1 {
		t.Fatalf("expected exactly 1 mail to %s, got %d", email, len(mine))
	}
	if !strings.Contains(mine[0].Text, res.VerifyCode) || strings.Contains(mine[0].Text, "http") {
		t.Fatalf("mail must contain the OTP and no link: %q", mine[0].Text)
	}
	// After sending, the plain code is gone from the database.
	var stored string
	if err := pool.QueryRowContext(ctx,
		"SELECT body FROM notification_outbox WHERE user_id = ? AND channel = 'email'", res.UserID).Scan(&stored); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if strings.Contains(stored, res.VerifyCode) {
		t.Fatal("OTP still stored in plain text after sending")
	}

	// A second pass sends nothing more for this user.
	before := len(fs.sent)
	if _, err := d.DeliverPending(ctx); err != nil {
		t.Fatalf("second pass: %v", err)
	}
	for _, m := range fs.sent[before:] {
		if m.To == email {
			t.Fatal("verification mail was sent twice")
		}
	}

	// The token from the mail activates the account.
	if err := authSvc.Verify(ctx, email, res.VerifyCode); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

// TestEmailDispatcher_FailureIsRecordedAndRetryBounded checks a failing SMTP
// send keeps the row pending with last_error, and stops after the max attempts.
func TestEmailDispatcher_FailureIsRecordedAndRetryBounded(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ctx := context.Background()
	authSvc := NewAuthService(store, testTokens(), NewNotificationService(store, clock), clock)

	res, err := authSvc.Register(ctx, RegisterInput{Email: uniqueEmail(), Password: "Passw0rd!", FullName: "Fail Test", Consent: true})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { cleanupUser(t, pool, res.UserID) })

	fs := &fakeSender{fail: errors.New("535 authentication failed")}
	d := NewEmailDispatcher(store, fs, clock, quietLogger(), "http://localhost:5173", 0)

	var attempts int
	var sentAt sql.NullTime
	var lastErr sql.NullString
	read := func() {
		t.Helper()
		if err := pool.QueryRowContext(ctx,
			"SELECT send_attempts, sent_at, last_error FROM notification_outbox WHERE user_id = ? AND channel = 'email'",
			res.UserID).Scan(&attempts, &sentAt, &lastErr); err != nil {
			t.Fatalf("read row: %v", err)
		}
	}
	// Keep passing until our row hits the limit (other pending rows in the
	// shared test DB may be claimed first), then a few extra passes.
	if _, err := d.DeliverPending(ctx); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	read()
	if attempts == 0 {
		t.Fatal("first attempt did not happen")
	}
	// Backoff: an immediate retry must not happen.
	first := attempts
	_, _ = d.DeliverPending(ctx)
	read()
	if attempts != first {
		t.Fatalf("retried before the backoff delay: attempts %d → %d", first, attempts)
	}
	// Advance past each backoff window until the retry limit is reached.
	for i := 0; i < 200 && attempts < emailMaxAttempts; i++ {
		clock.Advance(emailRetryDelay(uint32(attempts)) + time.Second)
		if _, err := d.DeliverPending(ctx); err != nil {
			t.Fatalf("deliver: %v", err)
		}
		read()
	}
	clock.Advance(time.Hour)
	for i := 0; i < 3; i++ {
		_, _ = d.DeliverPending(ctx)
	}
	read()
	if attempts != emailMaxAttempts {
		t.Fatalf("attempts = %d, want %d (bounded retries)", attempts, emailMaxAttempts)
	}
	if sentAt.Valid {
		t.Fatal("failed mail must not be marked sent")
	}
	if !lastErr.Valid || !strings.Contains(lastErr.String, "535") {
		t.Fatalf("last_error not recorded: %v", lastErr)
	}
}
