-- name: GetFundByID :one
SELECT * FROM funds WHERE id = ?;

-- name: GetFundByCode :one
SELECT * FROM funds WHERE code = ?;

-- name: ListFunds :many
-- Access pattern: market list by type (filtered on category / risk level).
SELECT * FROM funds
WHERE is_active = TRUE
ORDER BY code
LIMIT ? OFFSET ?;

-- name: CreateFund :execresult
INSERT INTO funds (code, name, category, risk_level, min_purchase, settlement_days, cutoff_time, fact_sheet_url, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetLatestFundNav :one
-- Access pattern: latest NAV for a fund ((fund_id, nav_date DESC)).
SELECT * FROM fund_navs
WHERE fund_id = ?
ORDER BY nav_date DESC
LIMIT 1;

-- name: ListFundNavs :many
-- Price/NAV history series for charts over a date range.
SELECT * FROM fund_navs
WHERE fund_id = ? AND nav_date >= ?
ORDER BY nav_date ASC;

-- name: UpsertFundNav :exec
INSERT INTO fund_navs (fund_id, nav_date, nav)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE nav = VALUES(nav);
