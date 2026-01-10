-- Custom Domain Management

-- name: CreateCustomDomain :one
INSERT INTO custom_domains (gid, domain, store_id, tenant_id, verification_token)
VALUES ($1, $2, $3, $4, '')
RETURNING *;

-- name: GetCustomDomainByID :one
SELECT * FROM custom_domains
WHERE id = $1;

-- name: GetCustomDomainByGID :one
SELECT * FROM custom_domains
WHERE gid = $1;

-- name: GetCustomDomainByDomain :one
SELECT * FROM custom_domains
WHERE domain = $1;

-- name: GetCustomDomainsByStoreID :many
SELECT * FROM custom_domains
WHERE store_id = $1
ORDER BY is_primary DESC, created_at ASC;

-- name: GetCustomDomainsByTenantID :many
SELECT * FROM custom_domains
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: GetVerifiedCustomDomainByDomain :one
SELECT * FROM custom_domains
WHERE domain = $1 AND verification_status = 'verified';

-- name: UpdateCustomDomainVerificationStatus :one
UPDATE custom_domains
SET
    verification_status = $2,
    verified_at = CASE WHEN $2 = 'verified' THEN now() ELSE verified_at END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateCustomDomainSSLStatus :one
UPDATE custom_domains
SET
    ssl_status = $2,
    ssl_expires_at = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetPrimaryDomain :exec
UPDATE custom_domains
SET is_primary = (id = $2), updated_at = now()
WHERE store_id = $1;

-- name: DeleteCustomDomain :exec
DELETE FROM custom_domains
WHERE id = $1 AND store_id = $2;

-- name: GetStoreByCustomDomain :one
SELECT s.* FROM stores s
JOIN custom_domains cd ON s.id = cd.store_id
WHERE cd.domain = $1
  AND cd.verification_status = 'verified';

-- name: GetPendingDomainVerifications :many
SELECT * FROM custom_domains
WHERE verification_status = 'pending'
  AND created_at > now() - INTERVAL '7 days'
ORDER BY created_at ASC;

-- name: CountCustomDomainsByStoreID :one
SELECT COUNT(*) FROM custom_domains
WHERE store_id = $1;
