-- bonds: bond master. data-model.md § bonds
-- Add-on: min_lot default 100,000 (user CR, overrides BR-04 1,000).
CREATE TABLE bonds (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code            VARCHAR(32) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    issuer          VARCHAR(255) NULL,
    coupon_rate     DECIMAL(9,4) NOT NULL DEFAULT 0,
    coupon_frequency TINYINT NOT NULL DEFAULT 2,
    face_value      DECIMAL(18,4) NOT NULL DEFAULT 1000,
    maturity_date   DATE NULL,
    ytm             DECIMAL(9,4) NULL,
    credit_rating   VARCHAR(16) NULL,
    offering_start  DATE NULL,
    offering_end    DATE NULL,
    min_lot         DECIMAL(18,4) NOT NULL DEFAULT 100000,
    risk_level      TINYINT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_bonds_code (code),
    KEY idx_bonds_maturity (maturity_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- bond_prices: daily clean price + accrued interest. data-model.md § bond_prices
CREATE TABLE bond_prices (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    bond_id          BIGINT UNSIGNED NOT NULL,
    price_date       DATE NOT NULL,
    clean_price      DECIMAL(18,4) NOT NULL,
    accrued_interest DECIMAL(18,4) NOT NULL DEFAULT 0,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_bond_prices_bond_date (bond_id, price_date),
    CONSTRAINT fk_bond_prices_bond FOREIGN KEY (bond_id) REFERENCES bonds (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- bond_coupon_schedules: coupon calendar per bond. data-model.md § bond_coupon_schedules
CREATE TABLE bond_coupon_schedules (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    bond_id         BIGINT UNSIGNED NOT NULL,
    coupon_date     DATE NOT NULL,
    amount_per_unit DECIMAL(18,4) NOT NULL,
    paid            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_bond_coupon_bond_date (bond_id, coupon_date),
    KEY idx_bond_coupon_date_paid (coupon_date, paid),
    CONSTRAINT fk_bond_coupon_bond FOREIGN KEY (bond_id) REFERENCES bonds (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
