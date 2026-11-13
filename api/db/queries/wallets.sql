-- name: GetWalletByUser :one
SELECT * FROM wallets WHERE user_id = ?;

-- name: GetWalletByUserForUpdate :one
-- Locks the wallet row inside a transaction for reserve/release/settle (BR-11).
SELECT * FROM wallets WHERE user_id = ? FOR UPDATE;

-- name: CreateWallet :execresult
INSERT INTO wallets (user_id, balance, reserved) VALUES (?, 0, 0);

-- name: UpdateWalletBalances :exec
UPDATE wallets SET balance = ?, reserved = ? WHERE id = ?;

-- name: CreateWalletTransaction :execresult
INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, ref_txn_id, status)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListWalletTransactions :many
SELECT * FROM wallet_transactions
WHERE wallet_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountWalletTransactions :one
SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = ?;
