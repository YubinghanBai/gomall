-- Orders Queries

-- name: CreateOrder :one
INSERT INTO orders (
    order_no,
    user_id,
    total_amount,
    discount_amount,
    shipping_fee,
    pay_amount,
    status,
    payment_status,
    ship_status,
    receiver_name,
    receiver_phone,
    receiver_address,
    receiver_zip_code,
    remark
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetOrderByOrderNo :one
SELECT * FROM orders
WHERE order_no = $1 AND deleted_at IS NULL;

-- name: ListUserOrders :many
SELECT * FROM orders
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserOrders :one
SELECT COUNT(*) FROM orders
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET
    status = $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: UpdateOrderPaymentStatus :exec
UPDATE orders
SET
    payment_status = $1,
    paid_at = CASE WHEN $1 = 'paid' THEN NOW() ELSE paid_at END,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: UpdateOrderShipStatus :exec
UPDATE orders
SET
    ship_status = $1,
    shipped_at = CASE WHEN $1 = 'shipped' THEN NOW() ELSE shipped_at END,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: CancelOrder :exec
UPDATE orders
SET
    status = 'cancelled',
    cancelled_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- Order Items Queries

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id,
    product_id,
    product_name,
    product_image,
    quantity,
    unit_price,
    total_price
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: GetOrderItemsByIDs :many
SELECT * FROM order_items
WHERE order_id = ANY($1::bigint[]) AND deleted_at IS NULL
ORDER BY order_id, id;
