-- wallets: cash balance per user (FR-07, BR-11). data-model.md § wallets
CREATE TABLE wallets (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    balance    DECIMAL(18,4) NOT NULL DEFAULT 0,
    reserved   DECIMAL(18,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_wallets_user (user_id),
    CONSTRAINT fk_wallets_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- wallet_transactions: wallet ledger. data-model.md § wallet_transactions
-- amount is signed per BR-15 (deposit/coupon/maturity/release = +, withdraw/hold = -).
CREATE TABLE wallet_transactions (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    wallet_id     BIGINT UNSIGNED NOT NULL,
    type          ENUM('deposit','withdraw','hold','release','settlement','coupon','maturity') NOT NULL,
    amount        DECIMAL(18,4) NOT NULL,
    balance_after DECIMAL(18,4) NOT NULL,
    ref_txn_id    BIGINT NULL,
    status        ENUM('done','pending') NOT NULL DEFAULT 'done',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_wallet_txn_wallet_created (wallet_id, created_at),
    CONSTRAINT fk_wallet_txn_wallet FOREIGN KEY (wallet_id) REFERENCES wallets (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
