-- db/queries/products.sql

-- name: CreateProduct :one
INSERT INTO products (
    shop_id, organization_id, title, description, description_html, handle,
    price_cents, compare_at_price_cents, cost_cents,
    vendor, product_type, tags,
    sku, barcode, inventory_quantity, track_inventory,
    status, seo_title, seo_description, template_suffix
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
)
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProductByHandle :one
SELECT * FROM products 
WHERE shop_id = $1 AND handle = $2 AND deleted_at IS NULL;

-- name: ListProductsByShop :many
SELECT * FROM products
WHERE shop_id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: ListProductsByOrganization :many
SELECT * FROM products
WHERE organization_id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: ListProductsByIDs :many
SELECT * FROM products
WHERE id = ANY(sqlc.arg('ids')::UUID[]) AND deleted_at IS NULL;

-- name: UpdateProduct :one
UPDATE products
SET 
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    description_html = COALESCE(sqlc.narg('description_html'), description_html),
    handle = COALESCE(sqlc.narg('handle'), handle),
    price_cents = COALESCE(sqlc.narg('price_cents'), price_cents),
    compare_at_price_cents = sqlc.narg('compare_at_price_cents'),
    cost_cents = sqlc.narg('cost_cents'),
    vendor = COALESCE(sqlc.narg('vendor'), vendor),
    product_type = COALESCE(sqlc.narg('product_type'), product_type),
    tags = COALESCE(sqlc.narg('tags'), tags),
    sku = COALESCE(sqlc.narg('sku'), sku),
    barcode = COALESCE(sqlc.narg('barcode'), barcode),
    inventory_quantity = COALESCE(sqlc.narg('inventory_quantity'), inventory_quantity),
    track_inventory = COALESCE(sqlc.narg('track_inventory'), track_inventory),
    status = COALESCE(sqlc.narg('status'), status),
    seo_title = COALESCE(sqlc.narg('seo_title'), seo_title),
    seo_description = COALESCE(sqlc.narg('seo_description'), seo_description),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateProductStatus :one
UPDATE products
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: PublishProduct :one
UPDATE products
SET status = 'active', published_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UnpublishProduct :one
UPDATE products
SET status = 'draft', published_at = NULL, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveProduct :one
UPDATE products
SET status = 'archived', updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProduct :exec
UPDATE products SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: DeductInventory :one
UPDATE products
SET inventory_quantity = inventory_quantity - sqlc.arg('quantity'), updated_at = NOW()
WHERE id = $1 
  AND deleted_at IS NULL
  AND inventory_quantity >= sqlc.arg('quantity')
RETURNING *;

-- name: AddInventory :one
UPDATE products
SET inventory_quantity = inventory_quantity + sqlc.arg('quantity'), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CountProductsByShop :one
SELECT COUNT(*) FROM products
WHERE shop_id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'));

-- name: CountProductsByOrganization :one
SELECT COUNT(*) FROM products
WHERE organization_id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'));

-- name: SearchProducts :many
SELECT * FROM products
WHERE shop_id = $1 AND deleted_at IS NULL
  AND (title ILIKE '%' || sqlc.arg('query') || '%' OR sku ILIKE '%' || sqlc.arg('query') || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');