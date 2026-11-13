-- transactions: unified immutable history feed (FR-16..FR-20, BR-14). data-model.md § transactions
-- Rules: never edited/deleted; corrections via new reversal row (ref_txn_id, BR-14).
-- amount is signed per BR-15.
CREATE TABLE transactions (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    txn_ref          VARCHAR(20) NOT NULL,
    user_id          BIGINT UNSIGNED NOT NULL,
    type             ENUM('buy','sell','dca_buy','coupon','maturity','deposit','withdraw') NOT NULL,
    asset_type       ENUM('fund','bond','gold') NULL,
    asset_id         BIGINT UNSIGNED NULL,
    amount           DECIMAL(18,4) NOT NULL,
    quantity         DECIMAL(18,4) NULL,
    price            DECIMAL(18,4) NULL,
    fee              DECIMAL(18,4) NOT NULL DEFAULT 0,
    net_amount       DECIMAL(18,4) NOT NULL,
    status           ENUM('submitted','processing','completed','failed','cancelled') NOT NULL,
    order_id         BIGINT UNSIGNED NULL,
    dca_plan_id      BIGINT UNSIGNED NULL,
    dca_round        INT NULL,
    trade_date       DATE NULL,
    settlement_date  DATE NULL,
    accrued_interest DECIMAL(18,4) NULL,
    realized_pl      DECIMAL(18,4) NULL,
    ref_txn_id       BIGINT UNSIGNED NULL,
    occurred_at      DATETIME NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_transactions_txn_ref (txn_ref),
    KEY idx_transactions_user_occurred (user_id, occurred_at DESC),
    KEY idx_transactions_user_type (user_id, type),
    KEY idx_transactions_user_asset (user_id, asset_type),
    CONSTRAINT fk_transactions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- transaction_status_history: status timeline for detail view (FR-18, BR-14).
-- data-model.md § transaction_status_history
CREATE TABLE transaction_status_history (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    transaction_id BIGINT UNSIGNED NOT NULL,
    status         VARCHAR(32) NOT NULL,
    note           VARCHAR(255) NULL,
    changed_by     VARCHAR(64) NULL,
    changed_at     DATETIME NOT NULL,
    PRIMARY KEY (id),
    KEY idx_txn_status_hist_txn_changed (transaction_id, changed_at),
    CONSTRAINT fk_txn_status_hist_txn FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
