-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
-- Access pattern: login by email (UNIQUE(email)).
SELECT * FROM users WHERE email = ?;

-- name: CreateUser :execresult
INSERT INTO users (email, password_hash, full_name, mobile, date_of_birth, role, status, pdpa_consent_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: SetUserStatus :exec
UPDATE users SET status = ? WHERE id = ?;

-- name: UpdateUserProfile :exec
UPDATE users SET full_name = ?, mobile = ?, bank_account_masked = ? WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = ? WHERE id = ?;

-- name: SetEmailNotifications :exec
UPDATE users SET email_notifications_enabled = ? WHERE id = ?;

-- name: IncrementFailedLogin :exec
UPDATE users SET failed_login_count = failed_login_count + 1 WHERE id = ?;

-- name: ResetFailedLogin :exec
UPDATE users SET failed_login_count = 0, locked_until = NULL WHERE id = ?;

-- name: LockUser :exec
UPDATE users SET status = 'locked', locked_until = ? WHERE id = ?;

-- name: LockUserRow :one
-- Row lock on the primary key (no gap lock) that serialises per-user flows
-- such as issuing and checking one-time codes (Add-on 2026-11-12b).
SELECT id FROM users WHERE id = ? FOR UPDATE;
