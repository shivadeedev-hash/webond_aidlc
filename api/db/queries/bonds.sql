-- name: GetBondByID :one
SELECT * FROM bonds WHERE id = ?;

-- name: GetBondByCode :one
SELECT * FROM bonds WHERE code = ?;

-- name: ListBonds :many
SELECT * FROM bonds
WHERE is_active = TRUE
ORDER BY code
LIMIT ? OFFSET ?;

-- name: CreateBond :execresult
INSERT INTO bonds (code, name, issuer, coupon_rate, coupon_frequency, face_value, maturity_date, ytm, credit_rating, offering_start, offering_end, min_lot, risk_level, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetLatestBondPrice :one
SELECT * FROM bond_prices
WHERE bond_id = ?
ORDER BY price_date DESC
LIMIT 1;

-- name: ListBondPrices :many
SELECT * FROM bond_prices
WHERE bond_id = ? AND price_date >= ?
ORDER BY price_date ASC;

-- name: UpsertBondPrice :exec
INSERT INTO bond_prices (bond_id, price_date, clean_price, accrued_interest)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE clean_price = VALUES(clean_price), accrued_interest = VALUES(accrued_interest);

-- name: CreateBondCouponSchedule :execresult
INSERT INTO bond_coupon_schedules (bond_id, coupon_date, amount_per_unit, paid)
VALUES (?, ?, ?, ?);

-- name: ListUnpaidCouponsDue :many
-- Access pattern: unpaid coupons due on/before a date (INDEX(coupon_date, paid)).
SELECT * FROM bond_coupon_schedules
WHERE paid = FALSE AND coupon_date <= ?
ORDER BY coupon_date ASC;

-- name: MarkCouponPaid :exec
UPDATE bond_coupon_schedules SET paid = TRUE WHERE id = ?;
