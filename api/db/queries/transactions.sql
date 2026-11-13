-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = ?;

-- name: GetTransactionByTxnRef :one
-- Access pattern: transaction by txn_ref (ownership check in service, EX-15).
SELECT * FROM transactions WHERE txn_ref = ?;

-- name: CreateTransaction :execresult
INSERT INTO transactions (txn_ref, user_id, type, asset_type, asset_id, amount, quantity, price, fee, net_amount, status, order_id, dca_plan_id, dca_round, trade_date, settlement_date, accrued_interest, realized_pl, ref_txn_id, occurred_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTransactionStatus :exec
UPDATE transactions SET status = ? WHERE id = ?;

-- name: ListTransactionsByUser :many
-- Access pattern: history list newest-first (INDEX(user_id, occurred_at DESC)).
SELECT * FROM transactions
WHERE user_id = ?
ORDER BY occurred_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountTransactionsByUser :one
SELECT COUNT(*) FROM transactions WHERE user_id = ?;

-- name: CreateTransactionStatusHistory :execresult
INSERT INTO transaction_status_history (transaction_id, status, note, changed_by, changed_at)
VALUES (?, ?, ?, ?, ?);

-- name: ListTransactionStatusHistory :many
SELECT * FROM transaction_status_history
WHERE transaction_id = ?
ORDER BY changed_at ASC, id ASC;
