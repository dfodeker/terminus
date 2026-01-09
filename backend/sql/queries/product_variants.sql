-- name: CreateProductVariant :one
INSERT INTO product_variants (
    id, tenant_id, store_id, product_id, sku, barcode, title,
    price_cents, compare_at_cents, option_values, status, created_at, updated_at
)
VALUES (
    gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now()
)
RETURNING *;

-- name: GetProductVariantByID :one
SELECT * FROM product_variants
WHERE id = $1;

-- name: GetProductVariantsByProductID :many
SELECT * FROM product_variants
WHERE product_id = $1
ORDER BY created_at ASC;

-- name: GetProductVariantsByStoreID :many
SELECT * FROM product_variants
WHERE store_id = $1
ORDER BY created_at DESC;

-- name: GetProductVariantsByProductIDPaginated :many
SELECT id, tenant_id, store_id, product_id, sku, barcode, title,
       price_cents, compare_at_cents, option_values, status, created_at, updated_at
FROM product_variants
WHERE product_id = $1
  AND (
    $2::boolean = false
    OR (created_at, id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $5;

-- name: GetProductVariantBySKU :one
SELECT * FROM product_variants
WHERE store_id = $1 AND sku = $2;

-- name: GetProductVariantByBarcode :one
SELECT * FROM product_variants
WHERE store_id = $1 AND barcode = $2;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET
    sku = $3,
    barcode = $4,
    title = $5,
    price_cents = $6,
    compare_at_cents = $7,
    option_values = $8,
    status = $9,
    updated_at = now()
WHERE id = $1 AND product_id = $2
RETURNING *;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants
WHERE id = $1 AND product_id = $2;

-- name: DeleteProductVariantsByProductID :exec
DELETE FROM product_variants
WHERE product_id = $1;

-- name: GetProductWithVariants :many
SELECT
    p.id as product_id,
    p.store_id,
    p.handle,
    p.name as product_name,
    p.description,
    p.inventory_tracked,
    p.sku as product_sku,
    p.tags,
    p.status as product_status,
    p.created_at as product_created_at,
    p.updated_at as product_updated_at,
    pv.id as variant_id,
    pv.tenant_id,
    pv.sku as variant_sku,
    pv.barcode,
    pv.title as variant_title,
    pv.price_cents,
    pv.compare_at_cents,
    pv.option_values,
    pv.status as variant_status,
    pv.created_at as variant_created_at,
    pv.updated_at as variant_updated_at
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id
WHERE p.id = $1 AND p.store_id = $2;

-- name: CountProductVariantsByProductID :one
SELECT COUNT(*) FROM product_variants
WHERE product_id = $1;
