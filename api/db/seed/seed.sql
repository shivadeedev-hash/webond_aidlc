-- WeBond demo seed (Task 6.1 / US-032). data-model.md shapes.
--
-- Idempotent + safe to re-run: every INSERT uses fixed primary keys with
-- `ON DUPLICATE KEY UPDATE` (or a guarded INSERT ... SELECT ... WHERE NOT EXISTS
-- for tables without a natural unique key), so applying it twice converges to
-- the same state instead of duplicating rows.
--
-- Synthetic data only (no real PII). Password hashes are real bcrypt (cost 12):
--   investor  demo@webond.test  / Demo@1234
--   admin     admin@webond.test / Admin@1234
-- These credentials are sample/dev-only and are documented in the root README.
--
-- Dates are expressed relative to CURDATE()/NOW() so the catalog, gold-quote
-- freshness (EX-10, < 5 min), cut-off and history stay realistic on every run.

SET @now := NOW();

-- =============================================================================
-- Users (investor + admin) — fixed ids 1 and 2.
-- =============================================================================
INSERT INTO users (id, email, password_hash, full_name, mobile, date_of_birth,
                   role, status, bank_account_masked, pdpa_consent_at,
                   failed_login_count, email_notifications_enabled)
VALUES
  (1, 'demo@webond.test',
      '$2a$12$HmHPhRX8vLWKFIINPXlMkeFFWQx8xIFpc6VV3sHQMwBVo05SmsKh.',
      'สมชาย นักลงทุน', '0812345678', '1990-05-14',
      'investor', 'active', 'xxx-x-x1234-5', @now, 0, TRUE),
  (2, 'admin@webond.test',
      '$2a$12$z68W4NfoEviPgXq2ptpz6uqty1HTSW.WxYYMNiaz4YDMbqWqS4R32',
      'WeBond Admin', '0820000000', '1985-01-01',
      'admin', 'active', NULL, @now, 0, TRUE)
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash), full_name = VALUES(full_name),
  mobile = VALUES(mobile), date_of_birth = VALUES(date_of_birth),
  role = VALUES(role), status = VALUES(status),
  bank_account_masked = VALUES(bank_account_masked),
  failed_login_count = 0, locked_until = NULL;

-- Risk assessment for the demo investor (level 4, valid 2 years). id fixed = 1.
INSERT INTO risk_assessments (id, user_id, risk_level, answers_json, assessed_at, expires_at)
VALUES (1, 1, 4, JSON_OBJECT('q1',3,'q2',4,'q3',4,'q4',3,'q5',5),
        @now - INTERVAL 90 DAY, @now + INTERVAL 2 YEAR)
ON DUPLICATE KEY UPDATE
  risk_level = VALUES(risk_level), answers_json = VALUES(answers_json),
  assessed_at = VALUES(assessed_at), expires_at = VALUES(expires_at);

-- Wallet for the demo investor (id fixed = 1). balance + reserved.
INSERT INTO wallets (id, user_id, balance, reserved)
VALUES (1, 1, 285000.0000, 15000.0000)
ON DUPLICATE KEY UPDATE balance = VALUES(balance), reserved = VALUES(reserved);

-- =============================================================================
-- Catalog — Mutual Funds (3) + NAV history (~90 daily points each).
-- =============================================================================
INSERT INTO funds (id, code, name, category, risk_level, min_purchase,
                   settlement_days, cutoff_time, fact_sheet_url, is_active)
VALUES
  (1, 'K-GLOBAL', 'K Global Equity Fund',      'Equity',       5, 500.0000, 3, '15:30:00', '#', TRUE),
  (2, 'SCB-FIXED','SCB Fixed Income Fund',      'Fixed Income', 2, 500.0000, 2, '15:30:00', '#', TRUE),
  (3, 'TMB-BAL',  'TMB Balanced Allocation',    'Balanced',     3, 500.0000, 3, '15:30:00', '#', TRUE)
ON DUPLICATE KEY UPDATE
  name = VALUES(name), category = VALUES(category), risk_level = VALUES(risk_level),
  min_purchase = VALUES(min_purchase), settlement_days = VALUES(settlement_days),
  cutoff_time = VALUES(cutoff_time), fact_sheet_url = VALUES(fact_sheet_url),
  is_active = VALUES(is_active);

-- NAV history: 90 daily points per fund via a numbers helper, with a gentle
-- deterministic sine-ish walk so charts/returns look real. UNIQUE(fund_id,nav_date).
DROP TEMPORARY TABLE IF EXISTS seq90;
CREATE TEMPORARY TABLE seq90 (n INT PRIMARY KEY);
INSERT INTO seq90 (n)
SELECT a.N + b.N*10 FROM
  (SELECT 0 N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
   UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a,
  (SELECT 0 N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
   UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b
WHERE a.N + b.N*10 < 90;

-- K-GLOBAL: base 15.00, amplitude ~1.2
INSERT INTO fund_navs (fund_id, nav_date, nav)
SELECT 1, CURDATE() - INTERVAL n DAY,
       ROUND(15.0000 + 1.2 * SIN(n/9.0) + (n%7-3)*0.03, 4)
FROM seq90
ON DUPLICATE KEY UPDATE nav = VALUES(nav);

-- SCB-FIXED: base 12.30, small amplitude (fixed income)
INSERT INTO fund_navs (fund_id, nav_date, nav)
SELECT 2, CURDATE() - INTERVAL n DAY,
       ROUND(12.3000 + 0.15 * SIN(n/14.0) + (n%5-2)*0.004, 4)
FROM seq90
ON DUPLICATE KEY UPDATE nav = VALUES(nav);

-- TMB-BAL: base 10.80, medium amplitude
INSERT INTO fund_navs (fund_id, nav_date, nav)
SELECT 3, CURDATE() - INTERVAL n DAY,
       ROUND(10.8000 + 0.5 * SIN(n/11.0) + (n%6-2)*0.015, 4)
FROM seq90
ON DUPLICATE KEY UPDATE nav = VALUES(nav);

-- =============================================================================
-- Catalog — Bonds (2) + price history + coupon schedules.
-- =============================================================================
INSERT INTO bonds (id, code, name, issuer, coupon_rate, coupon_frequency, face_value,
                   maturity_date, ytm, credit_rating, offering_start, offering_end,
                   min_lot, risk_level, is_active)
VALUES
  (1, 'LB28DA', 'Thai Govt Bond 2028', 'Ministry of Finance',
      3.2500, 2, 1000.0000, CURDATE() + INTERVAL 3 YEAR, 3.1000, 'AAA',
      CURDATE() - INTERVAL 60 DAY, CURDATE() + INTERVAL 120 DAY, 100000.0000, 2, TRUE),
  (2, 'CP234A', 'CP All Corporate Bond 2027', 'CP All PCL',
      4.1000, 2, 1000.0000, CURDATE() + INTERVAL 2 YEAR, 4.0500, 'A+',
      CURDATE() - INTERVAL 45 DAY, CURDATE() + INTERVAL 90 DAY, 100000.0000, 3, TRUE)
ON DUPLICATE KEY UPDATE
  name = VALUES(name), issuer = VALUES(issuer), coupon_rate = VALUES(coupon_rate),
  coupon_frequency = VALUES(coupon_frequency), face_value = VALUES(face_value),
  maturity_date = VALUES(maturity_date), ytm = VALUES(ytm),
  credit_rating = VALUES(credit_rating), offering_start = VALUES(offering_start),
  offering_end = VALUES(offering_end), min_lot = VALUES(min_lot),
  risk_level = VALUES(risk_level), is_active = VALUES(is_active);

-- Bond price history (90 daily clean prices, par ~100). UNIQUE(bond_id, price_date).
INSERT INTO bond_prices (bond_id, price_date, clean_price, accrued_interest)
SELECT 1, CURDATE() - INTERVAL n DAY,
       ROUND(100.5000 + 0.8 * SIN(n/15.0), 4),
       ROUND(((90-n)/180.0) * (3.2500/2) * 10, 4)
FROM seq90
ON DUPLICATE KEY UPDATE clean_price = VALUES(clean_price),
       accrued_interest = VALUES(accrued_interest);

INSERT INTO bond_prices (bond_id, price_date, clean_price, accrued_interest)
SELECT 2, CURDATE() - INTERVAL n DAY,
       ROUND(99.2000 + 0.6 * SIN(n/12.0), 4),
       ROUND(((90-n)/180.0) * (4.1000/2) * 10, 4)
FROM seq90
ON DUPLICATE KEY UPDATE clean_price = VALUES(clean_price),
       accrued_interest = VALUES(accrued_interest);

-- Coupon schedules (semi-annual). No natural unique key → guarded inserts.
INSERT INTO bond_coupon_schedules (bond_id, coupon_date, amount_per_unit, paid)
SELECT * FROM (SELECT 1 AS b, (CURDATE() - INTERVAL 30 DAY) AS d, 16.2500 AS a, TRUE AS p) t
WHERE NOT EXISTS (SELECT 1 FROM bond_coupon_schedules
                  WHERE bond_id = t.b AND coupon_date = t.d);
INSERT INTO bond_coupon_schedules (bond_id, coupon_date, amount_per_unit, paid)
SELECT * FROM (SELECT 1 AS b, (CURDATE() + INTERVAL 150 DAY) AS d, 16.2500 AS a, FALSE AS p) t
WHERE NOT EXISTS (SELECT 1 FROM bond_coupon_schedules
                  WHERE bond_id = t.b AND coupon_date = t.d);
INSERT INTO bond_coupon_schedules (bond_id, coupon_date, amount_per_unit, paid)
SELECT * FROM (SELECT 2 AS b, (CURDATE() - INTERVAL 20 DAY) AS d, 20.5000 AS a, TRUE AS p) t
WHERE NOT EXISTS (SELECT 1 FROM bond_coupon_schedules
                  WHERE bond_id = t.b AND coupon_date = t.d);
INSERT INTO bond_coupon_schedules (bond_id, coupon_date, amount_per_unit, paid)
SELECT * FROM (SELECT 2 AS b, (CURDATE() + INTERVAL 160 DAY) AS d, 20.5000 AS a, FALSE AS p) t
WHERE NOT EXISTS (SELECT 1 FROM bond_coupon_schedules
                  WHERE bond_id = t.b AND coupon_date = t.d);

-- =============================================================================
-- Catalog — Gold (1 product) + price history incl. a FRESH latest tick so
-- quotes are tradeable (EX-10 blocks trading on prices > 5 min old).
-- =============================================================================
INSERT INTO gold_products (id, code, name, purity, is_active)
VALUES (1, 'GOLD-965', 'Gold 96.5%', '96.5%', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name), purity = VALUES(purity),
       is_active = VALUES(is_active);

-- 90 daily historical ticks (one per day at noon) for the chart.
INSERT INTO gold_prices (gold_product_id, priced_at, bid_per_gram, ask_per_gram)
SELECT 1, TIMESTAMP(CURDATE() - INTERVAL n DAY, '12:00:00'),
       ROUND(41000 + 600 * SIN(n/13.0) - 40, 4),
       ROUND(41000 + 600 * SIN(n/13.0) + 40, 4)
FROM seq90
WHERE NOT EXISTS (
  SELECT 1 FROM gold_prices g
  WHERE g.gold_product_id = 1
    AND g.priced_at = TIMESTAMP(CURDATE() - INTERVAL seq90.n DAY, '12:00:00'));

-- A fresh tick at NOW() so a gold quote can be created/confirmed immediately.
INSERT INTO gold_prices (gold_product_id, priced_at, bid_per_gram, ask_per_gram)
SELECT 1, @now, 41180.0000, 41260.0000
WHERE NOT EXISTS (
  SELECT 1 FROM gold_prices g
  WHERE g.gold_product_id = 1 AND g.priced_at = @now);

DROP TEMPORARY TABLE IF EXISTS seq90;

-- =============================================================================
-- Holdings for the demo investor — one per asset class (MF/Bond/Gold).
-- =============================================================================
INSERT INTO holdings (id, user_id, asset_type, asset_id, quantity, locked_quantity, avg_cost, total_cost)
VALUES
  (1, 1, 'fund', 1, 3200.5000, 0.0000, 14.2000, 45447.1000),   -- K-GLOBAL
  (2, 1, 'bond', 1,     100.0000, 0.0000, 1000.0000, 100000.0000), -- LB28DA (100 units)
  (3, 1, 'gold', 1,      12.0000, 0.0000, 40800.0000, 489600.0000) -- 12 grams gold
ON DUPLICATE KEY UPDATE
  quantity = VALUES(quantity), locked_quantity = VALUES(locked_quantity),
  avg_cost = VALUES(avg_cost), total_cost = VALUES(total_cost);

-- =============================================================================
-- DCA plans (2) — one Mutual Fund, one Gold (bond excluded per BR-06).
-- =============================================================================
INSERT INTO dca_plans (id, user_id, asset_type, asset_id, amount_per_round, frequency,
                       run_day, start_date, end_date, max_rounds, rounds_executed,
                       total_invested, consecutive_failures, next_run_date, status)
VALUES
  (1, 1, 'fund', 1, 5000.0000, 'monthly', 1, CURDATE() - INTERVAL 90 DAY, NULL, NULL,
      3, 15000.0000, 0, CURDATE() + INTERVAL 10 DAY, 'active'),
  (2, 1, 'gold', 1, 2000.0000, 'weekly', 3, CURDATE() - INTERVAL 60 DAY, NULL, NULL,
      8, 16000.0000, 0, CURDATE() + INTERVAL 4 DAY, 'active')
ON DUPLICATE KEY UPDATE
  amount_per_round = VALUES(amount_per_round), frequency = VALUES(frequency),
  run_day = VALUES(run_day), start_date = VALUES(start_date),
  max_rounds = VALUES(max_rounds), rounds_executed = VALUES(rounds_executed),
  total_invested = VALUES(total_invested), next_run_date = VALUES(next_run_date),
  status = VALUES(status);

-- DCA rounds (UNIQUE(dca_plan_id, run_date)).
INSERT INTO dca_rounds (dca_plan_id, run_date, order_id, status, failure_reason)
VALUES
  (1, CURDATE() - INTERVAL 90 DAY, NULL, 'success', NULL),
  (1, CURDATE() - INTERVAL 60 DAY, NULL, 'success', NULL),
  (1, CURDATE() - INTERVAL 30 DAY, NULL, 'success', NULL),
  (2, CURDATE() - INTERVAL 21 DAY, NULL, 'success', NULL),
  (2, CURDATE() - INTERVAL 14 DAY, NULL, 'success', NULL),
  (2, CURDATE() - INTERVAL  7 DAY, NULL, 'success', NULL)
ON DUPLICATE KEY UPDATE status = VALUES(status);

-- =============================================================================
-- Transactions (~3 months of history) + status history + wallet ledger.
-- Fixed ids 1..8 so re-runs converge. amount signed per BR-15.
-- =============================================================================
INSERT INTO transactions (id, txn_ref, user_id, type, asset_type, asset_id, amount,
                          quantity, price, fee, net_amount, status, order_id,
                          dca_plan_id, dca_round, trade_date, settlement_date,
                          accrued_interest, realized_pl, occurred_at)
VALUES
  (1, 'WBSEED00000001', 1, 'deposit', NULL, NULL,  300000.0000, NULL, NULL, 0.0000, 300000.0000,
      'completed', NULL, NULL, NULL, CURDATE() - INTERVAL 92 DAY, CURDATE() - INTERVAL 92 DAY, NULL, NULL, @now - INTERVAL 92 DAY),
  (2, 'WBSEED00000002', 1, 'buy', 'fund', 1,       -45447.1000, 3200.5000, 14.2000, 0.0000, -45447.1000,
      'completed', NULL, 1, 1, CURDATE() - INTERVAL 90 DAY, CURDATE() - INTERVAL 87 DAY, NULL, NULL, @now - INTERVAL 90 DAY),
  (3, 'WBSEED00000003', 1, 'dca_buy', 'gold', 1,    -2000.0000, 0.0489, 40900.0000, 0.0000, -2000.0000,
      'completed', NULL, 2, 1, CURDATE() - INTERVAL 21 DAY, CURDATE() - INTERVAL 19 DAY, NULL, NULL, @now - INTERVAL 21 DAY),
  (4, 'WBSEED00000004', 1, 'buy', 'bond', 1,      -100000.0000, 100.0000, 1000.0000, 0.0000, -100000.0000,
      'completed', NULL, NULL, NULL, CURDATE() - INTERVAL 50 DAY, CURDATE() - INTERVAL 48 DAY, 0.0000, NULL, @now - INTERVAL 50 DAY),
  (5, 'WBSEED00000005', 1, 'coupon', 'bond', 1,     1625.0000, NULL, NULL, 0.0000, 1625.0000,
      'completed', NULL, NULL, NULL, CURDATE() - INTERVAL 30 DAY, CURDATE() - INTERVAL 30 DAY, NULL, NULL, @now - INTERVAL 30 DAY),
  (6, 'WBSEED00000006', 1, 'buy', 'gold', 1,     -489600.0000, 12.0000, 40800.0000, 0.0000, -489600.0000,
      'completed', NULL, NULL, NULL, CURDATE() - INTERVAL 40 DAY, CURDATE() - INTERVAL 40 DAY, NULL, NULL, @now - INTERVAL 40 DAY),
  (7, 'WBSEED00000007', 1, 'sell', 'fund', 1,       15000.0000, 1050.0000, 14.5000, 0.0000, 15000.0000,
      'completed', NULL, NULL, NULL, CURDATE() - INTERVAL 10 DAY, CURDATE() - INTERVAL 7 DAY, NULL, 315.0000, @now - INTERVAL 10 DAY),
  (8, 'WBSEED00000008', 1, 'withdraw', NULL, NULL,  -20000.0000, NULL, NULL, 0.0000, -20000.0000,
      'completed', NULL, NULL, NULL, CURDATE() - INTERVAL 5 DAY, CURDATE() - INTERVAL 5 DAY, NULL, NULL, @now - INTERVAL 5 DAY)
ON DUPLICATE KEY UPDATE
  status = VALUES(status), amount = VALUES(amount), net_amount = VALUES(net_amount),
  occurred_at = VALUES(occurred_at);

-- One pending MF buy that can be cancelled before cut-off (US-027 demo).
INSERT INTO transactions (id, txn_ref, user_id, type, asset_type, asset_id, amount,
                          quantity, price, fee, net_amount, status, occurred_at,
                          trade_date, settlement_date)
VALUES
  (9, 'WBSEED00000009', 1, 'buy', 'fund', 3, -5000.0000, NULL, NULL, 0.0000, -5000.0000,
      'submitted', @now - INTERVAL 1 HOUR, CURDATE(), CURDATE() + INTERVAL 3 DAY)
ON DUPLICATE KEY UPDATE status = VALUES(status), occurred_at = VALUES(occurred_at);

-- Status history timelines (guarded; no natural unique key).
INSERT INTO transaction_status_history (transaction_id, status, note, changed_by, changed_at)
SELECT * FROM (SELECT 2 AS tid, 'submitted' AS s, 'Order received' AS note, 'system' AS cb, @now - INTERVAL 90 DAY AS ca) t
WHERE NOT EXISTS (SELECT 1 FROM transaction_status_history h WHERE h.transaction_id = t.tid AND h.status = t.s);
INSERT INTO transaction_status_history (transaction_id, status, note, changed_by, changed_at)
SELECT * FROM (SELECT 2 AS tid, 'processing' AS s, 'NAV applied at cut-off' AS note, 'system' AS cb, @now - INTERVAL 90 DAY + INTERVAL 1 HOUR AS ca) t
WHERE NOT EXISTS (SELECT 1 FROM transaction_status_history h WHERE h.transaction_id = t.tid AND h.status = t.s);
INSERT INTO transaction_status_history (transaction_id, status, note, changed_by, changed_at)
SELECT * FROM (SELECT 2 AS tid, 'completed' AS s, 'Units allotted' AS note, 'system' AS cb, @now - INTERVAL 87 DAY AS ca) t
WHERE NOT EXISTS (SELECT 1 FROM transaction_status_history h WHERE h.transaction_id = t.tid AND h.status = t.s);
INSERT INTO transaction_status_history (transaction_id, status, note, changed_by, changed_at)
SELECT * FROM (SELECT 9 AS tid, 'submitted' AS s, 'Pending cut-off' AS note, 'system' AS cb, @now - INTERVAL 1 HOUR AS ca) t
WHERE NOT EXISTS (SELECT 1 FROM transaction_status_history h WHERE h.transaction_id = t.tid AND h.status = t.s);

-- Wallet ledger. Each ledger type appears once for the demo wallet, so we dedup
-- on (wallet_id, type) — a stable business key independent of @now drift so the
-- guard still matches prior rows on re-run.
INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, status, created_at)
SELECT * FROM (SELECT 1 AS w, 'deposit' AS ty, 300000.0000 AS a, 300000.0000 AS ba, 'done' AS st, @now - INTERVAL 92 DAY AS c) t
WHERE NOT EXISTS (SELECT 1 FROM wallet_transactions x WHERE x.wallet_id=t.w AND x.type=t.ty);
INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, status, created_at)
SELECT * FROM (SELECT 1 AS w, 'settlement' AS ty, -45447.1000 AS a, 254552.9000 AS ba, 'done' AS st, @now - INTERVAL 87 DAY AS c) t
WHERE NOT EXISTS (SELECT 1 FROM wallet_transactions x WHERE x.wallet_id=t.w AND x.type=t.ty);
INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, status, created_at)
SELECT * FROM (SELECT 1 AS w, 'coupon' AS ty, 1625.0000 AS a, 256177.9000 AS ba, 'done' AS st, @now - INTERVAL 30 DAY AS c) t
WHERE NOT EXISTS (SELECT 1 FROM wallet_transactions x WHERE x.wallet_id=t.w AND x.type=t.ty);
INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, status, created_at)
SELECT * FROM (SELECT 1 AS w, 'withdraw' AS ty, -20000.0000 AS a, 236177.9000 AS ba, 'done' AS st, @now - INTERVAL 5 DAY AS c) t
WHERE NOT EXISTS (SELECT 1 FROM wallet_transactions x WHERE x.wallet_id=t.w AND x.type=t.ty);
INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, status, created_at)
SELECT * FROM (SELECT 1 AS w, 'hold' AS ty, -15000.0000 AS a, 221177.9000 AS ba, 'pending' AS st, @now - INTERVAL 1 HOUR AS c) t
WHERE NOT EXISTS (SELECT 1 FROM wallet_transactions x WHERE x.wallet_id=t.w AND x.type=t.ty);

-- =============================================================================
-- A couple of in-app notifications for the demo investor (US-029).
-- =============================================================================
-- Dedup on (user_id, subject) — stable keys, independent of @now drift.
INSERT INTO notification_outbox (user_id, channel, category, subject, body, created_at)
SELECT * FROM (SELECT 1 AS u, 'inapp' AS ch, 'order' AS cat,
       'Order completed' AS sub, 'Your K-GLOBAL buy order has been completed.' AS bo,
       @now - INTERVAL 87 DAY AS c) t
WHERE NOT EXISTS (SELECT 1 FROM notification_outbox n
                  WHERE n.user_id=t.u AND n.subject=t.sub);
INSERT INTO notification_outbox (user_id, channel, category, subject, body, created_at)
SELECT * FROM (SELECT 1 AS u, 'inapp' AS ch, 'coupon' AS cat,
       'Coupon received' AS sub, 'You received a THB 1,625.00 coupon from LB28DA.' AS bo,
       @now - INTERVAL 30 DAY AS c) t
WHERE NOT EXISTS (SELECT 1 FROM notification_outbox n
                  WHERE n.user_id=t.u AND n.subject=t.sub);

-- Keep AUTO_INCREMENT counters ahead of the fixed seed ids so app-created rows
-- never collide with seeded primary keys.
ALTER TABLE users          AUTO_INCREMENT = 1000;
ALTER TABLE funds          AUTO_INCREMENT = 1000;
ALTER TABLE bonds          AUTO_INCREMENT = 1000;
ALTER TABLE gold_products  AUTO_INCREMENT = 1000;
ALTER TABLE holdings       AUTO_INCREMENT = 1000;
ALTER TABLE dca_plans      AUTO_INCREMENT = 1000;
ALTER TABLE risk_assessments AUTO_INCREMENT = 1000;
ALTER TABLE wallets        AUTO_INCREMENT = 1000;
ALTER TABLE transactions   AUTO_INCREMENT = 1000;
