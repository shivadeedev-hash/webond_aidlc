-- users: account & auth (Investor/Admin). data-model.md § users
CREATE TABLE users (
    id                          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email                       VARCHAR(255) NOT NULL,
    password_hash               VARCHAR(255) NOT NULL,
    full_name                   VARCHAR(255) NOT NULL,
    mobile                      VARCHAR(20) NULL,
    date_of_birth               DATE NULL,
    role                        ENUM('investor','admin') NOT NULL DEFAULT 'investor',
    status                      ENUM('pending_verification','active','locked') NOT NULL DEFAULT 'pending_verification',
    bank_account_masked         VARCHAR(64) NULL,
    pdpa_consent_at             TIMESTAMP NULL,
    failed_login_count          INT NOT NULL DEFAULT 0,
    locked_until                TIMESTAMP NULL,
    email_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at                  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- risk_assessments: risk suitability result (FR-04, BR-01). data-model.md § risk_assessments
CREATE TABLE risk_assessments (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id      BIGINT UNSIGNED NOT NULL,
    risk_level   TINYINT NOT NULL,
    answers_json JSON NULL,
    assessed_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at   TIMESTAMP NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_risk_user_assessed (user_id, assessed_at),
    CONSTRAINT fk_risk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
