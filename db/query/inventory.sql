-- Inventory Queries

-- name: CreateInventory :one
INSERT INTO inventory (
    product_id,
    available_stock,
    reserved_stock,
    low_stock_threshold
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetInventoryByProductID :one
SELECT * FROM inventory
WHERE product_id = $1 AND deleted_at IS NULL;

-- name: GetInventoryByID :one
SELECT * FROM inventory
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListInventories :many
SELECT * FROM inventory
WHERE deleted_at IS NULL
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: ListLowStockInventories :many
SELECT * FROM inventory
WHERE available_stock <= low_stock_threshold AND deleted_at IS NULL
ORDER BY available_stock ASC
LIMIT $1 OFFSET $2;

-- name: CountInventories :one
SELECT COUNT(*) FROM inventory
WHERE deleted_at IS NULL;

-- name: CountLowStockInventories :one
SELECT COUNT(*) FROM inventory
WHERE available_stock <= low_stock_threshold AND deleted_at IS NULL;

-- name: UpdateInventoryStock :exec
UPDATE inventory
SET
    available_stock = $1,
    reserved_stock = $2,
    version = version + 1,
    updated_at = NOW()
WHERE product_id = $3 AND version = $4 AND deleted_at IS NULL;

-- name: ReserveStock :exec
UPDATE inventory
SET
    available_stock = available_stock - $1,
    reserved_stock = reserved_stock + $1,
    version = version + 1,
    updated_at = NOW()
WHERE product_id = $2
    AND available_stock >= $1
    AND version = $3
    AND deleted_at IS NULL;

-- name: ReleaseReservedStock :exec
UPDATE inventory
SET
    available_stock = available_stock + $1,
    reserved_stock = reserved_stock - $1,
    version = version + 1,
    updated_at = NOW()
WHERE product_id = $2
    AND reserved_stock >= $1
    AND version = $3
    AND deleted_at IS NULL;

-- name: DeductReservedStock :exec
UPDATE inventory
SET
    reserved_stock = reserved_stock - $1,
    version = version + 1,
    updated_at = NOW()
WHERE product_id = $2
    AND reserved_stock >= $1
    AND version = $3
    AND deleted_at IS NULL;

-- name: AddAvailableStock :exec
UPDATE inventory
SET
    available_stock = available_stock + $1,
    version = version + 1,
    updated_at = NOW()
WHERE product_id = $2 AND deleted_at IS NULL;

-- name: UpdateLowStockThreshold :exec
UPDATE inventory
SET
    low_stock_threshold = $1,
    updated_at = NOW()
WHERE product_id = $2 AND deleted_at IS NULL;

-- name: DeleteInventory :exec
UPDATE inventory
SET
    deleted_at = NOW()
WHERE product_id = $1 AND deleted_at IS NULL;

-- Inventory Logs Queries

-- name: CreateInventoryLog :one
INSERT INTO inventory_logs (
    product_id,
    order_id,
    change_type,
    quantity_change,
    before_available,
    after_available,
    before_reserved,
    after_reserved,
    reason,
    operator_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetInventoryLogsByProductID :many
SELECT * FROM inventory_logs
WHERE product_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetInventoryLogsByOrderID :many
SELECT * FROM inventory_logs
WHERE order_id = sqlc.arg(order_id)::bigint
ORDER BY created_at DESC;

-- name: CountInventoryLogsByProductID :one
SELECT COUNT(*) FROM inventory_logs
WHERE product_id = $1;

-- Inventory Reservations Queries

-- name: CreateInventoryReservation :one
INSERT INTO inventory_reservations (
    product_id,
    order_id,
    quantity,
    status,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetInventoryReservationByID :one
SELECT * FROM inventory_reservations
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetInventoryReservationByOrderID :many
SELECT * FROM inventory_reservations
WHERE order_id = $1 AND deleted_at IS NULL;

-- name: GetActiveReservationsByProductID :many
SELECT * FROM inventory_reservations
WHERE product_id = $1
    AND status = 'active'
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateReservationStatus :exec
UPDATE inventory_reservations
SET
    status = $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: ConfirmReservation :exec
UPDATE inventory_reservations
SET
    status = 'confirmed',
    updated_at = NOW()
WHERE order_id = $1 AND status = 'active' AND deleted_at IS NULL;

-- name: CancelReservation :exec
UPDATE inventory_reservations
SET
    status = 'cancelled',
    updated_at = NOW()
WHERE order_id = $1 AND status = 'active' AND deleted_at IS NULL;

-- name: GetExpiredReservations :many
SELECT * FROM inventory_reservations
WHERE status = 'active'
    AND expires_at < NOW()
    AND deleted_at IS NULL
ORDER BY expires_at ASC
LIMIT $1;

-- name: DeleteReservation :exec
UPDATE inventory_reservations
SET
    deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
