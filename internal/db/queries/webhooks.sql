-- name: CreateWebhook :one
INSERT INTO webhooks (
    organization_id, shop_id, topic, endpoint_url, secret_hash, 
    fields, status, api_version
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetWebhookByID :one
SELECT * FROM webhooks WHERE id = $1;

-- name: GetWebhookByShopAndTopic :one
SELECT * FROM webhooks 
WHERE shop_id = $1 AND topic = $2 AND endpoint_url = $3;

-- name: UpdateWebhook :one
UPDATE webhooks
SET 
    topic = COALESCE(sqlc.narg('topic'), topic),
    endpoint_url = COALESCE(sqlc.narg('endpoint_url'), endpoint_url),
    secret_hash = COALESCE(sqlc.narg('secret_hash'), secret_hash),
    fields = COALESCE(sqlc.narg('fields'), fields),
    status = COALESCE(sqlc.narg('status'), status),
    api_version = COALESCE(sqlc.narg('api_version'), api_version),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteWebhook :exec
DELETE FROM webhooks WHERE id = $1;

-- name: ListWebhooksByShop :many
SELECT * FROM webhooks
WHERE shop_id = $1
  AND ($2::TEXT IS NULL OR topic = $2)
  AND ($3::TEXT IS NULL OR status = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListWebhooksByShopAndTopic :many
SELECT * FROM webhooks
WHERE shop_id = $1 AND topic = $2 AND status = 'active'
ORDER BY created_at DESC;

-- name: ListWebhooksByOrganization :many
SELECT * FROM webhooks
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListWebhooksByOrganizationAndTopic :many
SELECT * FROM webhooks
WHERE organization_id = $1 AND topic = $2 AND status = 'active'
ORDER BY created_at DESC;

-- name: CountWebhooksByShop :one
SELECT COUNT(*) FROM webhooks
WHERE shop_id = $1;

-- name: CountWebhooksByOrganization :one
SELECT COUNT(*) FROM webhooks
WHERE organization_id = $1;

-- name: IncrementWebhookFailureCount :one
UPDATE webhooks
SET 
    failure_count = failure_count + 1,
    last_failure_at = NOW(),
    last_failure_reason = $2,
    status = CASE WHEN failure_count + 1 >= 10 THEN 'disabled' ELSE status END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ResetWebhookFailureCount :one
UPDATE webhooks
SET 
    failure_count = 0,
    last_failure_at = NULL,
    last_failure_reason = NULL,
    status = CASE WHEN status = 'disabled' THEN 'active' ELSE status END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateWebhookStatus :one
UPDATE webhooks
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
