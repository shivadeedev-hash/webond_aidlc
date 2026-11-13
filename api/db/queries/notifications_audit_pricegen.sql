-- name: CreateNotification :execresult
INSERT INTO notification_outbox (user_id, channel, category, subject, body, token, token_expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListNotificationsByUser :many
-- Access pattern: notification inbox (INDEX(user_id, created_at DESC)).
SELECT * FROM notification_outbox
WHERE user_id = ? AND channel = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetLatestOTP :one
-- Newest open one-time code of a user for a purpose (verify/reset), locked for
-- the attempt check (Add-on 2026-11-12b). Must run inside a transaction.
SELECT * FROM notification_outbox
WHERE user_id = ? AND category = ? AND channel = 'email'
  AND token IS NOT NULL AND read_at IS NULL
ORDER BY id DESC
LIMIT 1
FOR UPDATE;

-- name: IncrementOTPAttempts :exec
UPDATE notification_outbox SET token_attempts = token_attempts + 1 WHERE id = ?;

-- name: ConsumeOTPs :exec
-- Closes every open code of a user for a purpose (after success, or before a
-- new code is issued so only the newest one works).
UPDATE notification_outbox SET read_at = ?
WHERE user_id = ? AND category = ? AND token IS NOT NULL AND read_at IS NULL;

-- name: MarkNotificationRead :exec
UPDATE notification_outbox SET read_at = ? WHERE id = ? AND user_id = ?;

-- name: ClaimPendingEmail :one
-- Locks the oldest undelivered email row still under the retry limit, with the
-- recipient address (Task 10.8). Must run inside a transaction; SKIP LOCKED lets
-- several API instances deliver in parallel without sending the same row twice.
SELECT n.id, n.user_id, n.category, n.subject, n.body, n.token, n.send_attempts, u.email
FROM notification_outbox n
JOIN users u ON u.id = n.user_id
WHERE n.channel = 'email' AND n.sent_at IS NULL AND n.send_attempts < sqlc.arg(max_attempts)
  AND (n.next_attempt_at IS NULL OR n.next_attempt_at <= sqlc.arg(now))
ORDER BY n.id
LIMIT 1
FOR UPDATE OF n SKIP LOCKED;

-- name: MarkEmailSent :exec
-- body: when not NULL, replaces the stored body (used to remove one-time codes
-- once they have been e-mailed).
UPDATE notification_outbox
SET sent_at = sqlc.arg(sent_at), send_attempts = send_attempts + 1, last_error = NULL,
    body = COALESCE(sqlc.narg(body), body)
WHERE id = sqlc.arg(id);

-- name: MarkEmailFailed :exec
UPDATE notification_outbox
SET send_attempts = send_attempts + 1, next_attempt_at = ?, last_error = ?
WHERE id = ?;

-- name: CreateAuditLog :execresult
INSERT INTO audit_logs (user_id, action, entity, entity_id, detail_json, ip)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetPriceGeneratorState :one
SELECT * FROM price_generator_state WHERE id = 1;

-- name: UpdatePriceGeneratorState :exec
UPDATE price_generator_state
SET enabled = ?, interval_seconds = ?, volatility_pct = ?, last_run_at = ?
WHERE id = 1;
