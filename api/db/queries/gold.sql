-- name: GetGoldProductByID :one
SELECT * FROM gold_products WHERE id = ?;

-- name: GetGoldProductByCode :one
SELECT * FROM gold_products WHERE code = ?;

-- name: ListGoldProducts :many
SELECT * FROM gold_products
WHERE is_active = TRUE
ORDER BY code
LIMIT ? OFFSET ?;

-- name: CreateGoldProduct :execresult
INSERT INTO gold_products (code, name, purity, is_active)
VALUES (?, ?, ?, ?);

-- name: GetLatestGoldPrice :one
-- Access pattern: latest gold price ((gold_product_id, priced_at DESC)); EX-10 staleness checked in service.
SELECT * FROM gold_prices
WHERE gold_product_id = ?
ORDER BY priced_at DESC
LIMIT 1;

-- name: ListGoldPrices :many
SELECT * FROM gold_prices
WHERE gold_product_id = ? AND priced_at >= ?
ORDER BY priced_at ASC;

-- name: CreateGoldPrice :execresult
INSERT INTO gold_prices (gold_product_id, priced_at, bid_per_gram, ask_per_gram)
VALUES (?, ?, ?, ?);
