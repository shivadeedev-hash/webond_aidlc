-- name: GetDcaPlanByID :one
SELECT * FROM dca_plans WHERE id = ?;

-- name: ListDcaPlansByUser :many
SELECT * FROM dca_plans
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: ListDuePlans :many
-- Access pattern: due DCA plans (INDEX(status, next_run_date)).
SELECT * FROM dca_plans
WHERE status = 'active' AND next_run_date <= ?
ORDER BY next_run_date ASC;

-- name: CreateDcaPlan :execresult
INSERT INTO dca_plans (user_id, asset_type, asset_id, amount_per_round, frequency, run_day, start_date, end_date, max_rounds, next_run_date, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active');

-- name: UpdateDcaPlan :exec
UPDATE dca_plans SET amount_per_round = ?, run_day = ?, next_run_date = ? WHERE id = ?;

-- name: SetDcaPlanStatus :exec
UPDATE dca_plans SET status = ? WHERE id = ?;

-- name: RecordDcaRun :exec
UPDATE dca_plans
SET rounds_executed = ?, total_invested = ?, consecutive_failures = ?, next_run_date = ?, status = ?
WHERE id = ?;

-- name: CreateDcaRound :execresult
INSERT INTO dca_rounds (dca_plan_id, run_date, order_id, status, failure_reason)
VALUES (?, ?, ?, ?, ?);

-- name: ListDcaRounds :many
SELECT * FROM dca_rounds
WHERE dca_plan_id = ?
ORDER BY run_date DESC;
