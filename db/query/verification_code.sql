-- name: CreateVerificationCode :one
INSERT INTO verification_codes (
  user_id, email, code, type, expires_at
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetVerificationCode :one
SELECT * FROM verification_codes
WHERE email = $1 AND code = $2 AND type = $3 AND is_used = FALSE
LIMIT 1;

-- name: MarkCodeAsUsed :exec
UPDATE verification_codes
SET is_used = TRUE
WHERE id = $1;

-- name: DeleteExpiredCodes :exec
DELETE FROM verification_codes
WHERE expires_at < NOW();

-- name: GetLatestVerificationCode :one
SELECT * FROM verification_codes
WHERE email = $1 AND type = $2 AND is_used = FALSE
ORDER BY created_at DESC
LIMIT 1;