-- name: CreateUser :one
INSERT INTO users (
    username, email, phone, password, nickname, avatar, gender
) VALUES (
             $1, $2, $3, $4, $5, $6, $7
         )
    RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL
    LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL
    LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 AND deleted_at IS NULL
    LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = $1 AND deleted_at IS NULL
    LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET
    nickname = COALESCE(sqlc.narg('nickname'), nickname),
    avatar = COALESCE(sqlc.narg('avatar'), avatar),
    gender = COALESCE(sqlc.narg('gender'), gender),
    birthday = COALESCE(sqlc.narg('birthday'), birthday),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password = $1,
    password_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET
    last_login_at = $1,
    last_login_ip = $2,
    updated_at = NOW()
WHERE id = $3 AND deleted_at IS NULL;

-- name: VerifyUserEmail :exec
UPDATE users
SET
    is_email_verified = TRUE,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: VerifyUserPhone :exec
UPDATE users
SET
    is_phone_verified = TRUE,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserStatus :exec
UPDATE users
SET
    status = $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
    LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;