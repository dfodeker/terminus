-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (
    webhook_id, organization_id, shop_id, topic, endpoint_url,
    request_headers, request_body, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetWebhookDeliveryByID :one
SELECT * FROM webhook_deliveries WHERE id = $1;

-- name: UpdateWebhookDelivery :one
UPDATE webhook_deliveries
SET 
    response_status = COALESCE(sqlc.narg('response_status'), response_status),
    response_headers = COALESCE(sqlc.narg('response_headers'), response_headers),
    response_body = COALESCE(sqlc.narg('response_body'), response_body),
    status = COALESCE(sqlc.narg('status'), status),
    attempts = COALESCE(sqlc.narg('attempts'), attempts),
    next_retry_at = sqlc.narg('next_retry_at'),
    error_message = sqlc.narg('error_message'),
    duration_ms = COALESCE(sqlc.narg('duration_ms'), duration_ms),
    delivered_at = sqlc.narg('delivered_at')
WHERE id = $1
RETURNING *;

-- name: ListWebhookDeliveriesByWebhook :many
SELECT * FROM webhook_deliveries
WHERE webhook_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListWebhookDeliveriesByShop :many
SELECT * FROM webhook_deliveries
WHERE shop_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPendingWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE status IN ('pending', 'retrying')
  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkWebhookDeliverySuccess :one
UPDATE webhook_deliveries
SET 
    status = 'success',
    response_status = $2,
    response_body = $3,
    duration_ms = $4,
    delivered_at = NOW(),
    attempts = attempts + 1
WHERE id = $1
RETURNING *;

-- name: MarkWebhookDeliveryFailed :one
UPDATE webhook_deliveries
SET 
    status = CASE WHEN attempts + 1 >= 5 THEN 'failed' ELSE 'retrying' END,
    error_message = $2,
    next_retry_at = $3,
    attempts = attempts + 1
WHERE id = $1
RETURNING *;

-- name: CountWebhookDeliveriesByWebhook :one
SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1;

-- name: CountWebhookDeliveriesByStatus :one
SELECT COUNT(*) FROM webhook_deliveries 
WHERE webhook_id = $1 AND status = $2;

-- name: DeleteOldWebhookDeliveries :exec
DELETE FROM webhook_deliveries
WHERE created_at < $1;
