-- db/queries/api_credentials.sql

-- name: CreateAPICredential :one
INSERT INTO api_credentials (
    organization_id, shop_id, name, key_prefix, key_hash,
    scopes, environment, status, created_by, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetAPICredentialByID :one
SELECT * FROM api_credentials WHERE id = $1;

-- name: GetAPICredentialByPrefix :one
SELECT * FROM api_credentials
WHERE key_prefix = $1 AND status = 'active';

-- name: ListAPICredentialsByOrganization :many
SELECT * FROM api_credentials
WHERE organization_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: ListAPICredentialsByShop :many
SELECT * FROM api_credentials
WHERE shop_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: UpdateAPICredentialLastUsed :exec
UPDATE api_credentials
SET last_used_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: RevokeAPICredential :one
UPDATE api_credentials
SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAPICredential :exec
DELETE FROM api_credentials WHERE id = $1;

-- name: UpdateAPICredential :one
UPDATE api_credentials
SET
    name = COALESCE(sqlc.narg('name'), name),
    scopes = COALESCE(sqlc.narg('scopes'), scopes),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    updated_at = NOW()
WHERE id = $1
RETURNING *;
