-- name: CreateProduct :one
INSERT INTO products (
    name, description, brand, price, origin_price, cost_price,
    stock, low_stock_threshold, category_id, status, is_featured, specifications
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
         )
    RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 AND deleted_at IS NULL
    LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL
  AND (sqlc.narg('category_id')::bigint IS NULL OR category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
    LIMIT $1 OFFSET $2;

-- name: ListProductsByCategory :many
SELECT * FROM products
WHERE category_id = $1
  AND status = 'published'
  AND deleted_at IS NULL
ORDER BY sales_count DESC
    LIMIT $2 OFFSET $3;

-- name: ListFeaturedProducts :many
SELECT * FROM products
WHERE is_featured = TRUE
  AND status = 'published'
  AND deleted_at IS NULL
ORDER BY sales_count DESC
    LIMIT $1 OFFSET $2;

-- name: UpdateProduct :exec
UPDATE products
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    brand = COALESCE(sqlc.narg('brand'), brand),
    price = COALESCE(sqlc.narg('price'), price),
    origin_price = COALESCE(sqlc.narg('origin_price'), origin_price),
    stock = COALESCE(sqlc.narg('stock'), stock),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    status = COALESCE(sqlc.narg('status'), status),
    is_featured = COALESCE(sqlc.narg('is_featured'), is_featured),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdateProductStock :exec
UPDATE products
SET
    stock = stock + $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: DecrementProductStock :exec
UPDATE products
SET
    stock = stock - $1,
    updated_at = NOW()
WHERE id = $2
  AND stock >= $1
  AND deleted_at IS NULL;

-- name: IncrementProductViews :exec
UPDATE products
SET view_count = view_count + 1
WHERE id = $1 AND deleted_at IS NULL;

-- name: IncrementProductSales :exec
UPDATE products
SET sales_count = sales_count + $1
WHERE id = $2 AND deleted_at IS NULL;

-- name: DeleteProduct :exec
UPDATE products
SET deleted_at = NOW()
WHERE id = $1;

-- name: SearchProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL
  AND status = 'published'
  AND (name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
ORDER BY sales_count DESC
    LIMIT $2 OFFSET $3;

-- name: CountProducts :one
SELECT COUNT(*) FROM products
WHERE deleted_at IS NULL;

-- Product Images

-- name: CreateProductImage :one
INSERT INTO product_images (
    product_id, image_url, sort, is_main
) VALUES (
             $1, $2, $3, $4
         )
    RETURNING *;

-- name: GetProductImages :many
SELECT * FROM product_images
WHERE product_id = $1 AND deleted_at IS NULL
ORDER BY sort ASC, id ASC;

-- name: GetProductMainImage :one
SELECT * FROM product_images
WHERE product_id = $1 AND is_main = TRUE AND deleted_at IS NULL
    LIMIT 1;

-- name: UpdateProductImage :exec
UPDATE product_images
SET
    image_url = COALESCE(sqlc.narg('image_url'), image_url),
    sort = COALESCE(sqlc.narg('sort'), sort),
    is_main = COALESCE(sqlc.narg('is_main'), is_main),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: DeleteProductImage :exec
UPDATE product_images
SET deleted_at = NOW()
WHERE id = $1;

-- name: DeleteProductImages :exec
UPDATE product_images
SET deleted_at = NOW()
WHERE product_id = $1;

-- Batch Operations

-- name: GetProductsByIDs :many
SELECT * FROM products
WHERE id = ANY($1::bigint[])
  AND deleted_at IS NULL
ORDER BY sales_count DESC;

-- name: GetImagesByProductIDs :many
SELECT * FROM product_images
WHERE product_id = ANY($1::bigint[])
  AND deleted_at IS NULL
ORDER BY product_id ASC, sort ASC, id ASC;

-- Stock Management

-- name: GetLowStockProducts :many
SELECT * FROM products
WHERE stock <= low_stock_threshold
  AND status = 'published'
  AND deleted_at IS NULL
ORDER BY stock ASC
LIMIT $1 OFFSET $2;

-- name: UpdateProductStockWithVersion :one
UPDATE products
SET
    stock = stock + $1,
    updated_at = NOW()
WHERE id = $2
  AND updated_at = $3
  AND deleted_at IS NULL
RETURNING *;

-- Advanced Filtering

-- name: ListProductsByPriceRange :many
SELECT * FROM products
WHERE deleted_at IS NULL
  AND status = 'published'
  AND price BETWEEN $1 AND $2
ORDER BY sales_count DESC
LIMIT $3 OFFSET $4;

-- name: UpdateProductsStatus :exec
UPDATE products
SET status = $1, updated_at = NOW()
WHERE id = ANY($2::bigint[])
  AND deleted_at IS NULL;