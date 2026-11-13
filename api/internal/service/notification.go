package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/repo"
)

// NotificationService writes in-app + simulated email rows to notification_outbox
// including OTP/verify/reset tokens (components.md § NotificationService, FR-24,
// FR-26). It honours the per-user email on/off preference (US-029): in-app rows
// are always written; email rows are skipped when the user disabled email.
type NotificationService struct {
	store *repo.Store
	clock Clock
}

// NewNotificationService builds a NotificationService.
func NewNotificationService(store *repo.Store, clock Clock) *NotificationService {
	return &NotificationService{store: store, clock: clock}
}

// Notice is an in-app/email notification to emit.
type Notice struct {
	UserID   uint64 // 0 → system (NULL user_id)
	Category db.NotificationOutboxCategory
	Subject  string
	Body     string
	// Token/TokenExpiry are set for verify/reset email entries (US-002/004).
	Token       string
	TokenExpiry sql.NullTime
	// EmailEnabled controls whether an email row is written (US-029). Callers
	// pass the user's preference; when false only the in-app row is created.
	EmailEnabled bool
	// EmailOnly suppresses the in-app row (used for verify/reset which are email
	// inbox entries, not dashboard notifications).
	EmailOnly bool
}

// Emit writes the configured rows for a notice. It is safe to call inside a
// transaction (pass a tx-bound store via EmitTx).
func (s *NotificationService) Emit(ctx context.Context, n Notice) error {
	return s.emit(ctx, s.store, n)
}

// EmitTx writes rows using a transaction-bound store so notification writes are
// atomic with the triggering mutation (e.g. order completion).
func (s *NotificationService) EmitTx(ctx context.Context, tx *repo.Store, n Notice) error {
	return s.emit(ctx, tx, n)
}

func (s *NotificationService) emit(ctx context.Context, store *repo.Store, n Notice) error {
	var userID sql.NullInt64
	if n.UserID != 0 {
		userID = nullInt64(int64(n.UserID))
	}

	if !n.EmailOnly {
		if _, err := store.Notifications.Create(ctx, db.CreateNotificationParams{
			UserID:   userID,
			Channel:  db.NotificationOutboxChannelInapp,
			Category: n.Category,
			Subject:  nullString(n.Subject),
			Body:     nullString(n.Body),
		}); err != nil {
			return fmt.Errorf("create in-app notification: %w", err)
		}
	}

	// Email row: written when email is enabled, OR always for token-carrying
	// verify/reset entries (the reviewer inbox must contain them — FR-26).
	writeEmail := n.EmailEnabled || n.Token != ""
	if writeEmail {
		if _, err := store.Notifications.Create(ctx, db.CreateNotificationParams{
			UserID:         userID,
			Channel:        db.NotificationOutboxChannelEmail,
			Category:       n.Category,
			Subject:        nullString(n.Subject),
			Body:           nullString(n.Body),
			Token:          nullString(n.Token),
			TokenExpiresAt: n.TokenExpiry,
		}); err != nil {
			return fmt.Errorf("create email notification: %w", err)
		}
	}
	return nil
}

// ListInApp returns a page of the user's in-app notifications (US-029).
func (s *NotificationService) ListInApp(ctx context.Context, userID uint64, page, pageSize int) ([]db.NotificationOutbox, int64, error) {
	limit, offset := pageBounds(page, pageSize)
	rows, err := s.store.Notifications.ListByUser(ctx, db.ListNotificationsByUserParams{
		UserID:  nullInt64(int64(userID)),
		Channel: db.NotificationOutboxChannelInapp,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	// Count is approximate via the page; return the page length for meta since
	// there is no dedicated count query. Callers use hasMore from page size.
	return rows, int64(offset + len(rows)), nil
}

// ListInbox returns a page of the user's simulated email inbox incl. OTP/verify/
// reset entries (US-002/004, FR-26 — reviewer convenience).
func (s *NotificationService) ListInbox(ctx context.Context, userID uint64, page, pageSize int) ([]db.NotificationOutbox, int64, error) {
	limit, offset := pageBounds(page, pageSize)
	rows, err := s.store.Notifications.ListByUser(ctx, db.ListNotificationsByUserParams{
		UserID:  nullInt64(int64(userID)),
		Channel: db.NotificationOutboxChannelEmail,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list inbox: %w", err)
	}
	return rows, int64(offset + len(rows)), nil
}

// MarkRead marks an in-app notification read, enforcing ownership (US-029).
func (s *NotificationService) MarkRead(ctx context.Context, userID, id uint64) error {
	if err := s.store.Notifications.MarkRead(ctx, db.MarkNotificationReadParams{
		ReadAt: nullTime(s.clock.Now()),
		ID:     id,
		UserID: nullInt64(int64(userID)),
	}); err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

// pageBounds normalises page/pageSize to a (limit, offset) pair with sane
// defaults (page≥1, 1≤pageSize≤100).
func pageBounds(page, pageSize int) (limit, offset int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageSize, (page - 1) * pageSize
}

// notFound builds a typed TXN_NOT_FOUND / generic not-found domain error.
func notFound(code domain.ErrorCode, msg string) *domain.AppError {
	return domain.NewError(code, msg)
}
