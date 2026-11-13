-- One-time codes (Add-on 2026-11-12b): count wrong attempts per code so a
-- 6-digit OTP cannot be brute-forced (5 attempts, then a new code is needed).
-- `token` now stores an HMAC of the code, never the code itself.
ALTER TABLE notification_outbox
    ADD COLUMN token_attempts INT UNSIGNED NOT NULL DEFAULT 0 AFTER token_expires_at,
    ADD KEY idx_notif_user_category (user_id, category, id);

-- Old link tokens can no longer be used: consume any that are still open.
UPDATE notification_outbox
SET read_at = CURRENT_TIMESTAMP
WHERE token IS NOT NULL AND read_at IS NULL;
