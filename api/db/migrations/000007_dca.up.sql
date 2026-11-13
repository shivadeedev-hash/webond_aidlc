-- dca_plans: DCA plans for MF & Gold (FR-13/14, BR-06/07). data-model.md § dca_plans
-- Rules: bond excluded (BR-06); min 500 MF / 100 gold (BR-07); pause after 3 fails (EX-09).
CREATE TABLE dca_plans (
    id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id              BIGINT UNSIGNED NOT NULL,
    asset_type           ENUM('fund','gold') NOT NULL,
    asset_id             BIGINT UNSIGNED NOT NULL,
    amount_per_round     DECIMAL(18,4) NOT NULL,
    frequency            ENUM('weekly','monthly') NOT NULL,
    run_day              TINYINT NOT NULL,
    start_date           DATE NOT NULL,
    end_date             DATE NULL,
    max_rounds           INT NULL,
    rounds_executed      INT NOT NULL DEFAULT 0,
    total_invested       DECIMAL(18,4) NOT NULL DEFAULT 0,
    consecutive_failures INT NOT NULL DEFAULT 0,
    next_run_date        DATE NULL,
    status               ENUM('active','paused','completed','cancelled') NOT NULL DEFAULT 'active',
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_dca_plans_user (user_id),
    KEY idx_dca_plans_status_next_run (status, next_run_date),
    CONSTRAINT fk_dca_plans_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- dca_rounds: one row per scheduled run (idempotent per FR-15). data-model.md § dca_rounds
CREATE TABLE dca_rounds (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    dca_plan_id    BIGINT UNSIGNED NOT NULL,
    run_date       DATE NOT NULL,
    order_id       BIGINT UNSIGNED NULL,
    status         ENUM('success','failed') NOT NULL,
    failure_reason VARCHAR(255) NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_dca_rounds_plan_date (dca_plan_id, run_date),
    CONSTRAINT fk_dca_rounds_plan FOREIGN KEY (dca_plan_id) REFERENCES dca_plans (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
