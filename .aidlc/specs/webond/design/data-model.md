# Data Model

## Overview
- **Database**: MySQL 8.0 (InnoDB). All money/unit columns are `DECIMAL(18,4)` (NFR-05) — never float.
- **Access layer**: sqlc (type-safe Go from SQL). Monetary values map to `shopspring/decimal` in Go.
- **Migrations**: golang-migrate SQL files in `api/migrations/`. Seed data in `api/seed/`.
- **Conventions**: snake_case tables/columns; `id` BIGINT UNSIGNED AUTO_INCREMENT PK (or CHAR(36) UUID where noted); `created_at`/`updated_at` TIMESTAMP; soft-delete via `is_active` where relevant.

## Entities

### users
- **Purpose**: Account & auth (Investor/Admin).
- **Fields**: `id` PK; `email` VARCHAR(255) UNIQUE NOT NULL; `password_hash` VARCHAR(255) (bcrypt); `full_name` VARCHAR(255); `mobile` VARCHAR(20); `date_of_birth` DATE; `role` ENUM('investor','admin') DEFAULT 'investor'; `status` ENUM('pending_verification','active','locked') DEFAULT 'pending_verification'; `bank_account_masked` VARCHAR(64) NULL; `pdpa_consent_at` TIMESTAMP NULL; `failed_login_count` INT DEFAULT 0; `locked_until` TIMESTAMP NULL; `email_notifications_enabled` BOOL DEFAULT TRUE; timestamps.
- **Indexes**: UNIQUE(email).
- **Business rules**: email unique (EX-02); lock after 5 fails for 15 min (EX-01); must be `active` + risk assessment valid before trading (BR-01).

### risk_assessments
- **Purpose**: Risk suitability result (FR-04, BR-01).
- **Fields**: `id` PK; `user_id` FK→users; `risk_level` TINYINT (1–5); `answers_json` JSON; `assessed_at` TIMESTAMP; `expires_at` TIMESTAMP (assessed_at + 2y); timestamps.
- **Indexes**: INDEX(user_id, assessed_at).
- **Rules**: latest non-expired row governs trading eligibility (BR-01) and risk-mismatch checks (BR-08).

### wallets
- **Purpose**: Cash balance per user (FR-07, BR-11).
- **Fields**: `id` PK; `user_id` FK UNIQUE; `balance` DECIMAL(18,4) DEFAULT 0; `reserved` DECIMAL(18,4) DEFAULT 0; timestamps. (available = balance − reserved, computed.)
- **Rules**: withdrawals only from available; reserve on pending buy, release on cancel/complete (BR-11).

### wallet_transactions
- **Purpose**: Wallet ledger (deposit/withdraw/hold/release/settlement).
- **Fields**: `id` PK; `wallet_id` FK; `type` ENUM('deposit','withdraw','hold','release','settlement','coupon','maturity'); `amount` DECIMAL(18,4) (signed per BR-15); `balance_after` DECIMAL(18,4); `ref_txn_id` BIGINT NULL; `status` ENUM('done','pending') DEFAULT 'done'; `created_at`.
- **Indexes**: INDEX(wallet_id, created_at).

### funds (Mutual Fund master)
- **Fields**: `id` PK; `code` VARCHAR(32) UNIQUE (e.g. K-GLOBAL); `name` VARCHAR(255); `category` VARCHAR(64); `risk_level` TINYINT; `min_purchase` DECIMAL(18,4) DEFAULT 500; `settlement_days` TINYINT DEFAULT 3 (T+n); `cutoff_time` TIME DEFAULT '15:30:00'; `fact_sheet_url` VARCHAR(512); `is_active` BOOL DEFAULT TRUE; timestamps.
- **Indexes**: UNIQUE(code), INDEX(category), INDEX(risk_level).

### fund_navs
- **Fields**: `id` PK; `fund_id` FK; `nav_date` DATE; `nav` DECIMAL(18,4); `created_at`.
- **Indexes**: UNIQUE(fund_id, nav_date), INDEX(fund_id, nav_date DESC).

### bonds
- **Fields**: `id` PK; `code` VARCHAR(32) UNIQUE (e.g. LB28DA); `name` VARCHAR(255); `issuer` VARCHAR(255); `coupon_rate` DECIMAL(9,4); `coupon_frequency` TINYINT (per year); `face_value` DECIMAL(18,4) DEFAULT 1000; `maturity_date` DATE; `ytm` DECIMAL(9,4); `credit_rating` VARCHAR(16); `offering_start` DATE; `offering_end` DATE; `min_lot` DECIMAL(18,4) DEFAULT 100000; `risk_level` TINYINT; `is_active` BOOL DEFAULT TRUE; timestamps.
- **Indexes**: UNIQUE(code), INDEX(maturity_date).
- **Add-on**: `min_lot` default 100,000 (user CR, overrides BR-04 1,000).

### bond_prices
- **Fields**: `id` PK; `bond_id` FK; `price_date` DATE; `clean_price` DECIMAL(18,4); `accrued_interest` DECIMAL(18,4); `created_at`.
- **Indexes**: UNIQUE(bond_id, price_date).

### bond_coupon_schedules
- **Fields**: `id` PK; `bond_id` FK; `coupon_date` DATE; `amount_per_unit` DECIMAL(18,4); `paid` BOOL DEFAULT FALSE; `created_at`.
- **Indexes**: INDEX(bond_id, coupon_date), INDEX(coupon_date, paid).

### gold_products
- **Fields**: `id` PK; `code` VARCHAR(32) UNIQUE (e.g. GOLD-965); `name` VARCHAR(255) (e.g. Gold 96.5%); `purity` VARCHAR(16); `is_active` BOOL DEFAULT TRUE; timestamps.

### gold_prices
- **Fields**: `id` PK; `gold_product_id` FK; `priced_at` DATETIME; `bid_per_gram` DECIMAL(18,4); `ask_per_gram` DECIMAL(18,4); `created_at`.
- **Indexes**: INDEX(gold_product_id, priced_at DESC).
- **Rules**: Buy uses ASK, Sell uses BID (BR-09); block trading if > 5 min old (EX-10).

### orders
- **Purpose**: Buy/Sell orders for all asset types (FR-08..FR-12).
- **Fields**: `id` PK; `txn_ref` VARCHAR(20) UNIQUE (WB+yyMMdd+7 running, BR-16); `user_id` FK; `asset_type` ENUM('fund','bond','gold'); `asset_id` BIGINT; `side` ENUM('buy','sell'); `order_mode` ENUM('amount','units','weight'); `amount` DECIMAL(18,4) NULL; `units` DECIMAL(18,4) NULL; `weight_grams` DECIMAL(18,4) NULL; `price` DECIMAL(18,4) NULL (NAV/clean price/gold price at execution); `fee` DECIMAL(18,4) DEFAULT 0 (BR-17); `net_amount` DECIMAL(18,4) NULL; `status` ENUM('pending','processing','completed','failed','cancelled'); `idempotency_key` VARCHAR(64) NULL; `dca_plan_id` BIGINT NULL; `dca_round_id` BIGINT NULL; `trade_date` DATE NULL; `settlement_date` DATE NULL; `realized_pl` DECIMAL(18,4) NULL; `risk_ack` BOOL DEFAULT FALSE; `channel` VARCHAR(16) DEFAULT 'web'; timestamps.
- **Indexes**: UNIQUE(txn_ref), UNIQUE(idempotency_key), INDEX(user_id, created_at), INDEX(status).
- **Rules**: idempotency (EX-11, NFR-05); cancellable only pending MF before cut-off (BR-13).

### holdings
- **Purpose**: Current positions per user+asset (FR-21, BR-10 avg cost).
- **Fields**: `id` PK; `user_id` FK; `asset_type` ENUM('fund','bond','gold'); `asset_id` BIGINT; `quantity` DECIMAL(18,4) (units/grams/bond units); `locked_quantity` DECIMAL(18,4) DEFAULT 0; `avg_cost` DECIMAL(18,4); `total_cost` DECIMAL(18,4); timestamps.
- **Indexes**: UNIQUE(user_id, asset_type, asset_id).

### dca_plans
- **Purpose**: DCA plans for MF & Gold (FR-13/14, BR-06/07).
- **Fields**: `id` PK; `user_id` FK; `asset_type` ENUM('fund','gold'); `asset_id` BIGINT; `amount_per_round` DECIMAL(18,4); `frequency` ENUM('weekly','monthly'); `run_day` TINYINT (1–7 weekly / 1–28 monthly); `start_date` DATE; `end_date` DATE NULL; `max_rounds` INT NULL; `rounds_executed` INT DEFAULT 0; `total_invested` DECIMAL(18,4) DEFAULT 0; `consecutive_failures` INT DEFAULT 0; `next_run_date` DATE NULL; `status` ENUM('active','paused','completed','cancelled') DEFAULT 'active'; timestamps.
- **Indexes**: INDEX(user_id), INDEX(status, next_run_date).
- **Rules**: bond excluded (BR-06); min 500 MF / 100 gold (BR-07); pause after 3 fails (EX-09).

### dca_rounds
- **Fields**: `id` PK; `dca_plan_id` FK; `run_date` DATE; `order_id` BIGINT NULL; `status` ENUM('success','failed'); `failure_reason` VARCHAR(255) NULL; `created_at`.
- **Indexes**: UNIQUE(dca_plan_id, run_date) (idempotent per FR-15).

### transactions
- **Purpose**: Unified immutable history feed (FR-16..FR-20, BR-14).
- **Fields**: `id` PK; `txn_ref` VARCHAR(20) UNIQUE; `user_id` FK; `type` ENUM('buy','sell','dca_buy','coupon','maturity','deposit','withdraw'); `asset_type` ENUM('fund','bond','gold') NULL; `asset_id` BIGINT NULL; `amount` DECIMAL(18,4) (signed per BR-15); `quantity` DECIMAL(18,4) NULL; `price` DECIMAL(18,4) NULL; `fee` DECIMAL(18,4) DEFAULT 0; `net_amount` DECIMAL(18,4); `status` ENUM('submitted','processing','completed','failed','cancelled'); `order_id` BIGINT NULL; `dca_plan_id` BIGINT NULL; `dca_round` INT NULL; `trade_date` DATE NULL; `settlement_date` DATE NULL; `accrued_interest` DECIMAL(18,4) NULL; `realized_pl` DECIMAL(18,4) NULL; `ref_txn_id` BIGINT NULL (reversal link BR-14); `occurred_at` DATETIME; `created_at`.
- **Indexes**: UNIQUE(txn_ref), INDEX(user_id, occurred_at DESC), INDEX(user_id, type), INDEX(user_id, asset_type).
- **Rules**: never edited/deleted; corrections via new reversal row (BR-14).

### transaction_status_history
- **Fields**: `id` PK; `transaction_id` FK; `status` VARCHAR(32); `note` VARCHAR(255); `changed_by` VARCHAR(64); `changed_at` DATETIME.
- **Indexes**: INDEX(transaction_id, changed_at).
- **Rules**: status timeline for detail view (FR-18); every change appended (BR-14).

### notification_outbox
- **Purpose**: In-app + simulated email inbox incl. OTP/verify/reset (FR-24, FR-26).
- **Fields**: `id` PK; `user_id` FK NULL; `channel` ENUM('inapp','email'); `category` ENUM('verify','reset','order','dca','coupon','maturity','general'); `subject` VARCHAR(255); `body` TEXT; `token` VARCHAR(128) NULL (OTP/reset, with expiry); `token_expires_at` TIMESTAMP NULL; `read_at` TIMESTAMP NULL; `created_at`.
- **Indexes**: INDEX(user_id, created_at DESC), INDEX(token).

### audit_logs
- **Purpose**: Audit for login/orders/DCA/admin + risk acks (NFR-02, BR-08).
- **Fields**: `id` PK; `user_id` FK NULL; `action` VARCHAR(64); `entity` VARCHAR(64); `entity_id` BIGINT NULL; `detail_json` JSON; `ip` VARCHAR(45); `created_at`.
- **Indexes**: INDEX(user_id, created_at), INDEX(action).

### price_generator_state
- **Purpose**: Toggle + config for the auto price-generator (US-030 Add-on).
- **Fields**: `id` TINYINT PK (singleton=1); `enabled` BOOL DEFAULT FALSE; `interval_seconds` INT DEFAULT 60; `last_run_at` DATETIME NULL; `volatility_pct` DECIMAL(6,4) DEFAULT 0.5; `updated_at`.

## ER Diagram (ASCII)
```
users 1──1 wallets 1──* wallet_transactions
users 1──* risk_assessments
users 1──* orders *──? dca_plans 1──* dca_rounds
users 1──* holdings
users 1──* transactions 1──* transaction_status_history
users 1──* notification_outbox
users 1──* audit_logs

funds 1──* fund_navs
bonds 1──* bond_prices
bonds 1──* bond_coupon_schedules
gold_products 1──* gold_prices
orders / holdings / transactions reference (asset_type, asset_id) → funds|bonds|gold_products
price_generator_state (singleton)
```

## Access Patterns

| Query | Frequency | Index used |
|-------|-----------|------------|
| Login by email | High | UNIQUE(email) |
| Market list by type + latest price/return | High | fund/bond/gold code + *_navs/*_prices (date DESC) |
| Latest NAV/price for asset | High | (asset_id, date DESC) |
| User holdings | High | UNIQUE(user_id, asset_type, asset_id) |
| History list (user, newest, filters) | High | INDEX(user_id, occurred_at DESC) + type/asset_type |
| Transaction by txn_ref (+ ownership) | Medium | UNIQUE(txn_ref) + user_id check (EX-15) |
| Due DCA plans | Medium (job) | INDEX(status, next_run_date) |
| Unpaid coupons due today | Medium (job) | INDEX(coupon_date, paid) |
| Notification inbox | Medium | INDEX(user_id, created_at DESC) |
