-- gold_products: gold master. data-model.md § gold_products
CREATE TABLE gold_products (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code       VARCHAR(32) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    purity     VARCHAR(16) NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_gold_products_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- gold_prices: BID/ASK per gram over time. data-model.md § gold_prices
-- Rules: Buy uses ASK, Sell uses BID (BR-09); block if > 5 min old (EX-10).
CREATE TABLE gold_prices (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    gold_product_id BIGINT UNSIGNED NOT NULL,
    priced_at       DATETIME NOT NULL,
    bid_per_gram    DECIMAL(18,4) NOT NULL,
    ask_per_gram    DECIMAL(18,4) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_gold_prices_product_priced (gold_product_id, priced_at DESC),
    CONSTRAINT fk_gold_prices_product FOREIGN KEY (gold_product_id) REFERENCES gold_products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
