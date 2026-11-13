ALTER TABLE notification_outbox
    DROP KEY idx_notif_pending,
    DROP COLUMN last_error,
    DROP COLUMN next_attempt_at,
    DROP COLUMN send_attempts,
    DROP COLUMN sent_at;
