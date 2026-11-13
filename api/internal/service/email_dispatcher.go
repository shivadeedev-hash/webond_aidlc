package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/mail"
	"github.com/webond/api/internal/repo"
)

// Email delivery tuning (Task 10.8).
const (
	// emailMaxAttempts is how often a row is tried before it is left for an
	// operator (last_error explains why).
	emailMaxAttempts = 5
	// emailBatchPerTick caps the messages sent per poll so one tick stays short.
	emailBatchPerTick = 20
	// emailSendTimeout bounds one SMTP conversation.
	emailSendTimeout = 30 * time.Second
	// maxErrorLen fits notification_outbox.last_error VARCHAR(255).
	maxErrorLen = 255
	// emailRetryBase is the first retry delay; it doubles per attempt
	// (30s, 1m, 2m, 4m → 5 attempts within about 8 minutes, then it gives up).
	emailRetryBase = 30 * time.Second
)

// emailRetryDelay is the wait before the next attempt after `attempts` failures.
func emailRetryDelay(attempts uint32) time.Duration {
	if attempts == 0 {
		attempts = 1
	}
	return emailRetryBase << (attempts - 1)
}

// EmailDispatcher delivers notification_outbox email rows after their business
// transaction committed (transactional outbox). Rows are claimed one at a time
// with SELECT … FOR UPDATE SKIP LOCKED, so several API instances can run it at
// once without sending a row twice; delivery is at-least-once.
type EmailDispatcher struct {
	store     *repo.Store
	sender    mail.Sender
	clock     Clock
	logger    *slog.Logger
	publicURL string
	interval  time.Duration

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

// NewEmailDispatcher builds a dispatcher. publicURL is the web origin used in
// links (e.g. http://localhost:5173); interval is the poll period.
func NewEmailDispatcher(store *repo.Store, sender mail.Sender, clock Clock, logger *slog.Logger, publicURL string, interval time.Duration) *EmailDispatcher {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &EmailDispatcher{
		store: store, sender: sender, clock: clock, logger: logger,
		publicURL: strings.TrimRight(publicURL, "/"), interval: interval,
	}
}

// Start runs the poll loop in the background until Stop is called.
func (d *EmailDispatcher) Start(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel != nil {
		return
	}
	loopCtx, cancel := context.WithCancel(ctx)
	d.cancel = cancel
	d.done = make(chan struct{})
	go d._loop(loopCtx, d.done)
	d.logger.Info("email dispatcher started", "sender", d.sender.Name(), "interval", d.interval.String())
}

// Stop ends the loop and waits for an in-flight send to finish.
func (d *EmailDispatcher) Stop() {
	d.mu.Lock()
	cancel, done := d.cancel, d.done
	d.cancel, d.done = nil, nil
	d.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
}

// DeliverPending sends up to emailBatchPerTick pending rows and returns how
// many were processed (sent or failed). Exposed for tests and the poll loop.
func (d *EmailDispatcher) DeliverPending(ctx context.Context) (int, error) {
	n := 0
	for n < emailBatchPerTick {
		if ctx.Err() != nil {
			return n, nil
		}
		processed, err := d._deliverOne(ctx)
		if err != nil {
			return n, err
		}
		if !processed {
			return n, nil
		}
		n++
	}
	return n, nil
}

func (d *EmailDispatcher) _loop(ctx context.Context, done chan struct{}) {
	defer close(done)
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		if _, err := d.DeliverPending(ctx); err != nil && ctx.Err() == nil {
			d.logger.Error("email delivery pass failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// _deliverOne claims, sends and records one row inside a transaction. It
// returns processed=false when nothing is pending.
func (d *EmailDispatcher) _deliverOne(ctx context.Context) (bool, error) {
	processed := false
	err := d.store.WithinTx(ctx, func(tx *repo.Store) error {
		now := d.clock.Now()
		row, err := tx.Notifications.ClaimPendingEmail(ctx, db.ClaimPendingEmailParams{
			MaxAttempts: emailMaxAttempts,
			Now:         nullTime(now),
		})
		if err != nil {
			if isNoRows(err) {
				return nil
			}
			return err
		}
		processed = true

		msg := mail.Message{
			To:      row.Email,
			Subject: row.Subject.String,
			Text:    ComposeEmailText(row.Category, row.Body.String, d.publicURL),
		}
		sendCtx, cancel := context.WithTimeout(ctx, emailSendTimeout)
		sendErr := d.sender.Send(sendCtx, msg)
		cancel()

		if sendErr != nil {
			d.logger.Warn("email delivery failed",
				"notificationId", row.ID, "to", mail.MaskAddress(row.Email),
				"attempt", row.SendAttempts+1, "error", sendErr)
			return tx.Notifications.MarkEmailFailed(ctx, db.MarkEmailFailedParams{
				NextAttemptAt: nullTime(now.Add(emailRetryDelay(row.SendAttempts + 1))),
				LastError:     nullString(truncate(sendErr.Error(), maxErrorLen)),
				ID:            row.ID,
			})
		}
		d.logger.Info("email delivered", "notificationId", row.ID,
			"to", mail.MaskAddress(row.Email), "category", string(row.Category), "sender", d.sender.Name())
		// One-time codes leave the database once e-mailed: only the HMAC in
		// `token` stays, which is enough to check the code.
		var scrub sql.NullString
		if row.Token.Valid {
			scrub = nullString(otpSentBody)
		}
		return tx.Notifications.MarkEmailSent(ctx, db.MarkEmailSentParams{
			SentAt: nullTime(now),
			Body:   scrub,
			ID:     row.ID,
		})
	})
	if errors.Is(err, context.Canceled) {
		return processed, nil
	}
	return processed, err
}

// otpSentBody replaces an OTP e-mail body after delivery.
const otpSentBody = "The one-time code was sent by e-mail."

// ComposeEmailText turns an outbox row into the e-mail body. Verification and
// reset e-mails already contain the one-time code and are sent as-is; event
// notifications get a short footer with the site address.
func ComposeEmailText(category db.NotificationOutboxCategory, body, publicURL string) string {
	if category == db.NotificationOutboxCategoryVerify || category == db.NotificationOutboxCategoryReset {
		return body
	}
	return body + "\n\n— WeBond\n" + strings.TrimRight(publicURL, "/") + "\n"
}

// truncate shortens s to at most n bytes without splitting a UTF-8 rune.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && (s[cut]&0xC0) == 0x80 {
		cut--
	}
	return s[:cut]
}
