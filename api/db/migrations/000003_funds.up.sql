-- funds: Mutual Fund master. data-model.md § funds
CREATE TABLE funds (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code           VARCHAR(32) NOT NULL,
    name           VARCHAR(255) NOT NULL,
    category       VARCHAR(64) NULL,
    risk_level     TINYINT NULL,
    min_purchase   DECIMAL(18,4) NOT NULL DEFAULT 500,
    settlement_days TINYINT NOT NULL DEFAULT 3,
    cutoff_time    TIME NOT NULL DEFAULT '15:30:00',
    fact_sheet_url VARCHAR(512) NULL,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_funds_code (code),
    KEY idx_funds_category (category),
    KEY idx_funds_risk_level (risk_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- fund_navs: daily NAV per fund. data-model.md § fund_navs
CREATE TABLE fund_navs (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    fund_id    BIGINT UNSIGNED NOT NULL,
    nav_date   DATE NOT NULL,
    nav        DECIMAL(18,4) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_fund_navs_fund_date (fund_id, nav_date),
    KEY idx_fund_navs_fund_date (fund_id, nav_date DESC),
    CONSTRAINT fk_fund_navs_fund FOREIGN KEY (fund_id) REFERENCES funds (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
