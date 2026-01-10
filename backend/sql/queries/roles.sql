-- Role Management

-- name: CreateRole :one
INSERT INTO roles (id, gid, tenant_id, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, now(), now())
RETURNING *;

-- name: GetRoleByGID :one
SELECT * FROM roles
WHERE gid = $1;

-- name: GetRoleByID :one
SELECT * FROM roles
WHERE id = $1;

-- name: GetRolesByTenantID :many
SELECT * FROM roles
WHERE tenant_id = $1
ORDER BY name;

-- name: GetRoleByTenantAndName :one
SELECT * FROM roles
WHERE tenant_id = $1 AND name = $2;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;

-- Permission Management

-- name: CreatePermission :one
INSERT INTO permissions (id, gid, key, description, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, now(), now())
RETURNING *;

-- name: GetPermissionByGID :one
SELECT * FROM permissions
WHERE gid = $1;

-- name: GetPermissionByID :one
SELECT * FROM permissions
WHERE id = $1;

-- name: GetPermissionByKey :one
SELECT * FROM permissions
WHERE key = $1;

-- name: GetAllPermissions :many
SELECT * FROM permissions
ORDER BY key;

-- Role-Permission Assignment

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: GetPermissionsByRoleID :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.key;

-- Tenant User Role Assignment

-- name: AssignRoleToTenantUser :exec
INSERT INTO tenant_user_roles (tenant_user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (tenant_user_id, role_id) DO NOTHING;

-- name: RemoveRoleFromTenantUser :exec
DELETE FROM tenant_user_roles
WHERE tenant_user_id = $1 AND role_id = $2;

-- name: GetRolesByTenantUserID :many
SELECT r.* FROM roles r
JOIN tenant_user_roles tur ON r.id = tur.role_id
WHERE tur.tenant_user_id = $1
ORDER BY r.name;

-- Permission Checking Queries

-- name: CheckUserHasPermission :one
SELECT EXISTS (
    SELECT 1 FROM tenant_users tu
    JOIN tenant_user_roles tur ON tu.id = tur.tenant_user_id
    JOIN role_permissions rp ON tur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE tu.tenant_id = $1
    AND tu.user_id = $2
    AND p.key = $3
    AND tu.status = 'active'
) AS has_permission;

-- name: GetUserPermissionsInTenant :many
SELECT DISTINCT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN tenant_user_roles tur ON rp.role_id = tur.role_id
JOIN tenant_users tu ON tur.tenant_user_id = tu.id
WHERE tu.tenant_id = $1 AND tu.user_id = $2 AND tu.status = 'active'
ORDER BY p.key;

-- name: GetUserRolesInTenant :many
SELECT r.* FROM roles r
JOIN tenant_user_roles tur ON r.id = tur.role_id
JOIN tenant_users tu ON tur.tenant_user_id = tu.id
WHERE tu.tenant_id = $1 AND tu.user_id = $2 AND tu.status = 'active'
ORDER BY r.name;

-- name: GetRoleByTenantAndID :one
SELECT * FROM roles
WHERE tenant_id = $1 AND id = $2;

-- name: GetRolesByTenantIDPaginated :many
SELECT id, gid, tenant_id, name, description, created_at, updated_at
FROM roles
WHERE tenant_id = $1
  AND (
    $2::boolean = false
    OR (created_at, id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $5;

-- name: GetPermissionsByKeys :many
SELECT * FROM permissions
WHERE key = ANY($1::text[])
ORDER BY key;


