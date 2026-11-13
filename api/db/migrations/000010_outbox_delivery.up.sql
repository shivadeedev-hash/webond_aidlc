-- Email delivery state for the transactional outbox (Task 10.8, FR-01/FR-24/FR-26).
-- Email rows are written inside the business transaction and delivered over SMTP
-- by a background dispatcher after commit; these columns record the outcome.
ALTER TABLE notification_outbox
    ADD COLUMN sent_at         TIMESTAMP    NULL AFTER read_at,
    ADD COLUMN send_attempts   INT UNSIGNED NOT NULL DEFAULT 0 AFTER sent_at,
    ADD COLUMN next_attempt_at TIMESTAMP    NULL AFTER send_attempts,
    ADD COLUMN last_error      VARCHAR(255) NULL AFTER next_attempt_at,
    ADD KEY idx_notif_pending (channel, sent_at, send_attempts);

-- Rows created before delivery existed were only ever "simulated"; mark them as
-- handled so enabling delivery never sends a backlog of old messages.
UPDATE notification_outbox SET sent_at = created_at WHERE channel = 'email';
