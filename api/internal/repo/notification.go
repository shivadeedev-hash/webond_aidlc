package repo

import (
	"context"
	"database/sql"

	"github.com/webond/api/internal/db"
)

// NotificationRepo is in-app + simulated email outbox persistence (FR-24, FR-26).
type NotificationRepo interface {
	Create(ctx context.Context, arg db.CreateNotificationParams) (sql.Result, error)
	ListByUser(ctx context.Context, arg db.ListNotificationsByUserParams) ([]db.NotificationOutbox, error)
	// One-time codes (Add-on 2026-11-12b). GetLatestOTP must run in a transaction.
	GetLatestOTP(ctx context.Context, arg db.GetLatestOTPParams) (db.NotificationOutbox, error)
	IncrementOTPAttempts(ctx context.Context, id uint64) error
	ConsumeOTPs(ctx context.Context, arg db.ConsumeOTPsParams) error
	MarkRead(ctx context.Context, arg db.MarkNotificationReadParams) error
	// Email delivery (transactional outbox, Task 10.8).
	ClaimPendingEmail(ctx context.Context, arg db.ClaimPendingEmailParams) (db.ClaimPendingEmailRow, error)
	MarkEmailSent(ctx context.Context, arg db.MarkEmailSentParams) error
	MarkEmailFailed(ctx context.Context, arg db.MarkEmailFailedParams) error
}

type notificationRepo struct{ q *db.Queries }

func (r *notificationRepo) ClaimPendingEmail(ctx context.Context, arg db.ClaimPendingEmailParams) (db.ClaimPendingEmailRow, error) {
	return r.q.ClaimPendingEmail(ctx, arg)
}
func (r *notificationRepo) MarkEmailSent(ctx context.Context, arg db.MarkEmailSentParams) error {
	return r.q.MarkEmailSent(ctx, arg)
}
func (r *notificationRepo) MarkEmailFailed(ctx context.Context, arg db.MarkEmailFailedParams) error {
	return r.q.MarkEmailFailed(ctx, arg)
}

func (r *notificationRepo) Create(ctx context.Context, arg db.CreateNotificationParams) (sql.Result, error) {
	return r.q.CreateNotification(ctx, arg)
}
func (r *notificationRepo) ListByUser(ctx context.Context, arg db.ListNotificationsByUserParams) ([]db.NotificationOutbox, error) {
	return r.q.ListNotificationsByUser(ctx, arg)
}
func (r *notificationRepo) GetLatestOTP(ctx context.Context, arg db.GetLatestOTPParams) (db.NotificationOutbox, error) {
	return r.q.GetLatestOTP(ctx, arg)
}
func (r *notificationRepo) IncrementOTPAttempts(ctx context.Context, id uint64) error {
	return r.q.IncrementOTPAttempts(ctx, id)
}
func (r *notificationRepo) ConsumeOTPs(ctx context.Context, arg db.ConsumeOTPsParams) error {
	return r.q.ConsumeOTPs(ctx, arg)
}
func (r *notificationRepo) MarkRead(ctx context.Context, arg db.MarkNotificationReadParams) error {
	return r.q.MarkNotificationRead(ctx, arg)
}

// AuditRepo is audit-log persistence (NFR-02, BR-08).
type AuditRepo interface {
	Create(ctx context.Context, arg db.CreateAuditLogParams) (sql.Result, error)
	ListByUser(ctx context.Context, arg db.ListAuditLogsByUserParams) ([]db.AuditLog, error)
}

type auditRepo struct{ q *db.Queries }

func (r *auditRepo) Create(ctx context.Context, arg db.CreateAuditLogParams) (sql.Result, error) {
	return r.q.CreateAuditLog(ctx, arg)
}
func (r *auditRepo) ListByUser(ctx context.Context, arg db.ListAuditLogsByUserParams) ([]db.AuditLog, error) {
	return r.q.ListAuditLogsByUser(ctx, arg)
}

// PriceGeneratorRepo is the auto price-generator singleton state (US-030 Add-on).
type PriceGeneratorRepo interface {
	Get(ctx context.Context) (db.PriceGeneratorState, error)
	Update(ctx context.Context, arg db.UpdatePriceGeneratorStateParams) error
}

type priceGeneratorRepo struct{ q *db.Queries }

func (r *priceGeneratorRepo) Get(ctx context.Context) (db.PriceGeneratorState, error) {
	return r.q.GetPriceGeneratorState(ctx)
}
func (r *priceGeneratorRepo) Update(ctx context.Context, arg db.UpdatePriceGeneratorStateParams) error {
	return r.q.UpdatePriceGeneratorState(ctx, arg)
}
