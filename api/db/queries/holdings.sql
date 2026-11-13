-- name: GetHolding :one
-- Access pattern: position by (user, asset_type, asset_id) (UNIQUE key).
SELECT * FROM holdings
WHERE user_id = ? AND asset_type = ? AND asset_id = ?;

-- name: GetHoldingForUpdate :one
SELECT * FROM holdings
WHERE user_id = ? AND asset_type = ? AND asset_id = ?
FOR UPDATE;

-- name: ListHoldingsByUser :many
-- Access pattern: all holdings for a user (portfolio).
SELECT * FROM holdings
WHERE user_id = ? AND quantity > 0
ORDER BY asset_type, asset_id;

-- name: UpsertHolding :exec
INSERT INTO holdings (user_id, asset_type, asset_id, quantity, locked_quantity, avg_cost, total_cost)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    quantity = VALUES(quantity),
    locked_quantity = VALUES(locked_quantity),
    avg_cost = VALUES(avg_cost),
    total_cost = VALUES(total_cost);

-- name: ListHoldersByAsset :many
-- Access pattern: all holders of an asset (bond coupon/maturity payouts, FR-23).
SELECT * FROM holdings
WHERE asset_type = ? AND asset_id = ? AND quantity > 0
ORDER BY user_id;
