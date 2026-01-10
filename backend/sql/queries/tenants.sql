-- name: CreateTenant :one
INSERT INTO tenants (id, gid, name, status, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, 'active', now(), now())
RETURNING *;

-- name: GetTenantByGID :one
SELECT * FROM tenants
WHERE gid = $1;

-- name: GetTenantByID :one
SELECT * FROM tenants
WHERE id = $1;

-- name: GetTenantByName :one
SELECT * FROM tenants
WHERE name = $1;

-- name: GetTenantsByUserID :many
SELECT t.* FROM tenants t
JOIN tenant_users tu ON t.id = tu.tenant_id
WHERE tu.user_id = $1 AND tu.status = 'active';

-- name: UpdateTenantStatus :one
UPDATE tenants
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateTenantUser :one
INSERT INTO tenant_users (id, tenant_id, user_id, status, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, now(), now())
RETURNING *;

-- name: GetTenantUsersByTenantID :many
SELECT * FROM tenant_users
WHERE tenant_id = $1;

-- name: GetTenantUser :one
SELECT * FROM tenant_users
WHERE tenant_id = $1 AND user_id = $2;

-- name: UpdateTenantUserStatus :one
UPDATE tenant_users
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetTenantUsersWithDetails :many
SELECT tu.*, u.email FROM tenant_users tu
JOIN users u ON tu.user_id = u.id
WHERE tu.tenant_id = $1;

-- name: GetTenantsByUserIDPaginated :many
SELECT t.id, t.gid, t.name, t.status, t.created_at, t.updated_at
FROM tenants t
JOIN tenant_users tu ON t.id = tu.tenant_id
WHERE tu.user_id = $1
  AND tu.status = 'active'
  AND (
    $2::boolean = false
    OR (t.created_at, t.id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY t.created_at DESC, t.id DESC
LIMIT $5;

-- name: GetTenantUsersWithDetailsPaginated :many
SELECT tu.id, tu.tenant_id, tu.user_id, tu.status, tu.created_at, tu.updated_at, u.email
FROM tenant_users tu
JOIN users u ON tu.user_id = u.id
WHERE tu.tenant_id = $1
  AND (
    $2::boolean = false
    OR (tu.created_at, tu.id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY tu.created_at DESC, tu.id DESC
LIMIT $5;

-- name: GetTenantUserByID :one
SELECT * FROM tenant_users
WHERE id = $1;
