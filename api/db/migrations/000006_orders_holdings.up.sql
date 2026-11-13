-- orders: buy/sell orders for all asset types (FR-08..FR-12). data-model.md § orders
-- asset_id / dca_plan_id / dca_round_id are logical references (no hard FK: asset
-- polymorphic across funds|bonds|gold_products; dca columns avoid a circular FK).
CREATE TABLE orders (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    txn_ref         VARCHAR(20) NOT NULL,
    user_id         BIGINT UNSIGNED NOT NULL,
    asset_type      ENUM('fund','bond','gold') NOT NULL,
    asset_id        BIGINT UNSIGNED NOT NULL,
    side            ENUM('buy','sell') NOT NULL,
    order_mode      ENUM('amount','units','weight') NOT NULL,
    amount          DECIMAL(18,4) NULL,
    units           DECIMAL(18,4) NULL,
    weight_grams    DECIMAL(18,4) NULL,
    price           DECIMAL(18,4) NULL,
    fee             DECIMAL(18,4) NOT NULL DEFAULT 0,
    net_amount      DECIMAL(18,4) NULL,
    status          ENUM('pending','processing','completed','failed','cancelled') NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(64) NULL,
    dca_plan_id     BIGINT UNSIGNED NULL,
    dca_round_id    BIGINT UNSIGNED NULL,
    trade_date      DATE NULL,
    settlement_date DATE NULL,
    realized_pl     DECIMAL(18,4) NULL,
    risk_ack        BOOLEAN NOT NULL DEFAULT FALSE,
    channel         VARCHAR(16) NOT NULL DEFAULT 'web',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_orders_txn_ref (txn_ref),
    UNIQUE KEY uq_orders_idempotency_key (idempotency_key),
    KEY idx_orders_user_created (user_id, created_at),
    KEY idx_orders_status (status),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- holdings: current positions per user+asset (FR-21, BR-10 avg cost). data-model.md § holdings
CREATE TABLE holdings (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id         BIGINT UNSIGNED NOT NULL,
    asset_type      ENUM('fund','bond','gold') NOT NULL,
    asset_id        BIGINT UNSIGNED NOT NULL,
    quantity        DECIMAL(18,4) NOT NULL DEFAULT 0,
    locked_quantity DECIMAL(18,4) NOT NULL DEFAULT 0,
    avg_cost        DECIMAL(18,4) NOT NULL DEFAULT 0,
    total_cost      DECIMAL(18,4) NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_holdings_user_asset (user_id, asset_type, asset_id),
    CONSTRAINT fk_holdings_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
