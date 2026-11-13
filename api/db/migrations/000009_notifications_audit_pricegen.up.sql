-- notification_outbox: in-app + simulated email incl. OTP/verify/reset (FR-24, FR-26).
-- data-model.md § notification_outbox
CREATE TABLE notification_outbox (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id          BIGINT UNSIGNED NULL,
    channel          ENUM('inapp','email') NOT NULL,
    category         ENUM('verify','reset','order','dca','coupon','maturity','general') NOT NULL,
    subject          VARCHAR(255) NULL,
    body             TEXT NULL,
    token            VARCHAR(128) NULL,
    token_expires_at TIMESTAMP NULL,
    read_at          TIMESTAMP NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_notif_user_created (user_id, created_at DESC),
    KEY idx_notif_token (token),
    CONSTRAINT fk_notif_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- audit_logs: audit for login/orders/DCA/admin + risk acks (NFR-02, BR-08).
-- data-model.md § audit_logs
CREATE TABLE audit_logs (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id     BIGINT UNSIGNED NULL,
    action      VARCHAR(64) NOT NULL,
    entity      VARCHAR(64) NULL,
    entity_id   BIGINT UNSIGNED NULL,
    detail_json JSON NULL,
    ip          VARCHAR(45) NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_audit_user_created (user_id, created_at),
    KEY idx_audit_action (action),
    CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- price_generator_state: toggle + config for the auto price-generator (US-030 Add-on).
-- Singleton row (id = 1). data-model.md § price_generator_state
CREATE TABLE price_generator_state (
    id               TINYINT UNSIGNED NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    interval_seconds INT NOT NULL DEFAULT 60,
    last_run_at      DATETIME NULL,
    volatility_pct   DECIMAL(6,4) NOT NULL DEFAULT 0.5,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed the singleton row so the generator state always exists (disabled by default).
INSERT INTO price_generator_state (id, enabled, interval_seconds, volatility_pct)
VALUES (1, FALSE, 60, 0.5);
