-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = ?;

-- name: GetOrderByTxnRef :one
-- Access pattern: order by txn_ref (+ ownership check in service, EX-15).
SELECT * FROM orders WHERE txn_ref = ?;

-- name: GetOrderByIdempotencyKey :one
-- Idempotency guard (EX-11, NFR-05) — durable via UNIQUE(idempotency_key).
SELECT * FROM orders WHERE idempotency_key = ?;

-- name: CreateOrder :execresult
INSERT INTO orders (txn_ref, user_id, asset_type, asset_id, side, order_mode, amount, units, weight_grams, price, fee, net_amount, status, idempotency_key, dca_plan_id, dca_round_id, trade_date, settlement_date, realized_pl, risk_ack, channel)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateOrderStatus :exec
UPDATE orders SET status = ? WHERE id = ?;

-- name: SetOrderExecution :exec
UPDATE orders SET status = ?, price = ?, units = ?, weight_grams = ?, net_amount = ?, realized_pl = ? WHERE id = ?;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;
