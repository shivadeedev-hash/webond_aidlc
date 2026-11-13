ALTER TABLE notification_outbox
    DROP KEY idx_notif_user_category,
    DROP COLUMN token_attempts;
