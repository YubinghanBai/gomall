-- name: AddToCart :one
INSERT INTO carts (user_id, product_id, quantity, selected)
VALUES ($1, $2, $3, TRUE)
    ON CONFLICT (user_id, product_id)
  DO UPDATE SET
    quantity = carts.quantity + EXCLUDED.quantity,
             updated_at = NOW()
             RETURNING *;

-- name: GetCartByUserID :many
SELECT * FROM carts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetCartItem :one
SELECT * FROM carts
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
    LIMIT 1;

-- name: GetCartItemByProduct :one
SELECT * FROM carts
WHERE user_id = $1 AND product_id = $2 AND deleted_at IS NULL
    LIMIT 1;

-- name: UpdateCartQuantity :exec
UPDATE carts
SET
    quantity = $1,
    updated_at = NOW()
WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL;

-- name: UpdateCartSelected :exec
UPDATE carts
SET
    selected = $1,
    updated_at = NOW()
WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL;

-- name: UpdateAllCartSelected :exec
UPDATE carts
SET
    selected = $1,
    updated_at = NOW()
WHERE user_id = $2 AND deleted_at IS NULL;

-- name: DeleteCartItem :exec
UPDATE carts
SET deleted_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: ClearCart :exec
UPDATE carts
SET deleted_at = NOW()
WHERE user_id = $1;

-- name: GetSelectedCartItems :many
SELECT * FROM carts
WHERE user_id = $1 AND selected = TRUE AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CountCartItems :one
SELECT COUNT(*) FROM carts
WHERE user_id = $1 AND deleted_at IS NULL;