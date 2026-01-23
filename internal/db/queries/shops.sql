

-- name: GetShopByID :one
SELECT * FROM shops WHERE id = $1 AND deleted_at IS NULL;

-- name: GetShopByGID :one
SELECT * FROM shops WHERE gid = $1 AND deleted_at IS NULL;

-- name: GetShopBySubdomain :one
SELECT * FROM shops WHERE subdomain = $1 AND deleted_at IS NULL;

-- name: GetShopByCustomDomain :one
SELECT * FROM shops WHERE custom_domain = $1 AND deleted_at IS NULL;

-- name: GetShopByHandle :one
SELECT * FROM shops 
WHERE organization_id = $1 AND handle = $2 AND deleted_at IS NULL;

-- name: CreateShop :one
INSERT INTO shops (
    organization_id, 
    name, 
    handle, 
    subdomain, 
    custom_domain,
    currency, 
    locale,
    timezone, 
    shop_owner,
    email,
    phone,
    source,
    referral_code,
    status,
    gid
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: UpdateShop :one
UPDATE shops
SET 
    name = COALESCE(sqlc.narg('name'), name),
    handle = COALESCE(sqlc.narg('handle'), handle),
    custom_domain = COALESCE(sqlc.narg('custom_domain'), custom_domain),
    currency = COALESCE(sqlc.narg('currency'), currency),
    locale = COALESCE(sqlc.narg('locale'), locale),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    shop_owner = COALESCE(sqlc.narg('shop_owner'), shop_owner),
    email = COALESCE(sqlc.narg('email'), email),
    phone = COALESCE(sqlc.narg('phone'), phone),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteShop :exec
UPDATE shops 
SET deleted_at = NOW(), updated_at = NOW() 
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListShopsByOrganization :many
SELECT * FROM shops 
WHERE organization_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListShopsByOrganizationWithStatus :many
SELECT * FROM shops 
WHERE organization_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountShopsByOrganization :one
SELECT COUNT(*) FROM shops 
WHERE organization_id = $1 AND deleted_at IS NULL;

-- name: ShopExistsBySubdomain :one
SELECT EXISTS(SELECT 1 FROM shops WHERE subdomain = $1 AND deleted_at IS NULL);

-- name: ShopExistsByCustomDomain :one
SELECT EXISTS(SELECT 1 FROM shops WHERE custom_domain = $1 AND deleted_at IS NULL);

-- name: UpdateShopStatus :one
UPDATE shops 
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SetShopCustomDomain :one
UPDATE shops 
SET custom_domain = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ClearShopCustomDomain :one
UPDATE shops 
SET custom_domain = NULL, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;