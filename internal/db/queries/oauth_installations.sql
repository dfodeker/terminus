-- db/queries/oauth_installations.sql

-- name: CreateOAuthInstallation :one
INSERT INTO oauth_installations (
    shop_id, app_id, access_token_hash, scopes, status
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOAuthInstallationByID :one
SELECT * FROM oauth_installations WHERE id = $1;

-- name: GetOAuthInstallationByAccessToken :one
SELECT * FROM oauth_installations
WHERE access_token_hash = $1 AND status = 'active';

-- name: GetOAuthInstallationByShopAndApp :one
SELECT * FROM oauth_installations
WHERE shop_id = $1 AND app_id = $2;

-- name: ListOAuthInstallationsByShop :many
SELECT * FROM oauth_installations
WHERE shop_id = $1
ORDER BY created_at DESC;

-- name: UpdateOAuthInstallation :one
UPDATE oauth_installations
SET
    access_token_hash = COALESCE(sqlc.narg('access_token_hash'), access_token_hash),
    scopes = COALESCE(sqlc.narg('scopes'), scopes),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UninstallOAuthInstallation :one
UPDATE oauth_installations
SET status = 'uninstalled', uninstalled_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOAuthInstallation :exec
DELETE FROM oauth_installations WHERE id = $1;
