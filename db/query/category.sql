-- name: CreateCategory :one
INSERT INTO categories (
    name,
    slug,
    parent_id,
    icon,
    sort,
    level,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCategoryBySlug :one
SELECT * FROM categories
WHERE slug = $1 AND deleted_at IS NULL;

-- name: UpdateCategory :exec
UPDATE categories
SET
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
    icon = COALESCE(sqlc.narg('icon'), icon),
    sort = COALESCE(sqlc.narg('sort'), sort),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: DeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListCategories :many
SELECT * FROM categories
WHERE deleted_at IS NULL
  AND ($1::boolean IS NULL OR is_active = $1)
ORDER BY sort, id;

-- name: GetRootCategories :many
SELECT * FROM categories
WHERE parent_id IS NULL
  AND deleted_at IS NULL
  AND is_active = true
ORDER BY sort, id;

-- name: GetCategoryChildren :many
SELECT * FROM categories
WHERE parent_id = $1
  AND deleted_at IS NULL
ORDER BY sort, id;

-- name: CountCategoryChildren :one
SELECT COUNT(*) FROM categories
WHERE parent_id = $1 AND deleted_at IS NULL;

-- name: CountProductsByCategory :one
SELECT COUNT(*) FROM products
WHERE category_id = $1 AND deleted_at IS NULL;
