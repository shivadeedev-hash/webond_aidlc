-- name: CreateRiskAssessment :execresult
INSERT INTO risk_assessments (user_id, risk_level, answers_json, assessed_at, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetLatestRiskAssessment :one
-- Access pattern: latest row governs trading eligibility (BR-01).
SELECT * FROM risk_assessments
WHERE user_id = ?
ORDER BY assessed_at DESC
LIMIT 1;
