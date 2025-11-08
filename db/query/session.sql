-- name: CreateSession :one
INSERT INTO sessions (
    id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at
) VALUES (
             $1, $2, $3, $4, $5, $6, $7
         )
    RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1
    LIMIT 1;

-- name: GetUserSessions :many
SELECT * FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: BlockSession :exec
UPDATE sessions
SET is_blocked = TRUE
WHERE id = $1;

-- name: CleanExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();


