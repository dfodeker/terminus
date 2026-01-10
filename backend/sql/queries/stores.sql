-- name: CreateStore :one
INSERT INTO stores (id, gid, name, handle, address, status, default_currency, timezone, plan, tenant_id, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
RETURNING *;

-- name: GetStoreByGID :one
SELECT * FROM stores
WHERE gid = $1;

-- name: DeleteAllStores :exec
DELETE FROM stores;

-- name: GetStores :many
SELECT * FROM stores ORDER BY created_at ASC;

-- name: GetStoreByHandle :one
SELECT * FROM stores WHERE handle = $1;

-- name: UpdateStore :one
UPDATE stores
SET
    name = $2,
    handle = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- Tenant-scoped store queries

-- name: CreateStoreForTenant :one
INSERT INTO stores (id, gid, name, handle, address, status, default_currency, timezone, plan, tenant_id, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, '', 'active', 'USD', 'UTC', $4, $5, now(), now())
RETURNING *;

-- name: GetStoresByTenantID :many
SELECT * FROM stores
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: GetStoresByTenantIDPaginated :many
SELECT id, gid, name, handle, address, status, default_currency, timezone, plan, tenant_id, created_at, updated_at
FROM stores
WHERE tenant_id = $1
  AND (
    $2::boolean = false
    OR (created_at, id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $5;

-- name: GetStoreByTenantAndHandle :one
SELECT * FROM stores
WHERE tenant_id = $1 AND handle = $2;

-- name: GetStoreByTenantAndID :one
SELECT * FROM stores
WHERE tenant_id = $1 AND id = $2;

-- name: UpdateStoreForTenant :one
UPDATE stores
SET
    name = $3,
    handle = $4,
    address = $5,
    status = $6,
    default_currency = $7,
    timezone = $8,
    plan = $9,
    updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteStoreForTenant :exec
DELETE FROM stores
WHERE id = $1 AND tenant_id = $2;
