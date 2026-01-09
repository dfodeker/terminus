-- name: CreateProduct :one
INSERT INTO products (id, store_id, handle, name, description, inventory_tracked, sku, tags, status, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING *;


-- name: GetProductsByStore :many
SELECT * FROM products
WHERE store_id = $1
ORDER BY created_at ASC;




-- name: GetProductsByStorePaginated :many
SELECT id, store_id, handle, name, description, inventory_tracked, sku, tags, status, created_at, updated_at
FROM products
WHERE store_id = $1
  AND (
    $2::boolean = false
    OR (created_at, id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $5;





-- name: GetProductByHandle :one
SELECT * FROM products
WHERE store_id = $1 AND handle = $2;


-- name: UpdateProduct :one
UPDATE products
SET 
    handle = $3,
    name = $4,
    description = $5,
    inventory_tracked = $6,
    sku = $7,
    tags = $8,
    status = $9,
    updated_at = NOW()
WHERE id = $1 AND store_id = $2
RETURNING *;    


-- name: DeleteProduct :one
DELETE FROM products
WHERE id = $1 AND store_id = $2
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 AND store_id = $2;

-- name: GetProductByIDOnly :one
SELECT * FROM products
WHERE id = $1;


