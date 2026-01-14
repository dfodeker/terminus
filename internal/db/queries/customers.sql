-- db/queries/customers.sql

-- name: CreateCustomer :one
INSERT INTO customers (
    shop_id, organization_id, email, phone, first_name, last_name,
    accepts_marketing, marketing_opt_in_at, tags, note, tax_exempt, status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetCustomerByID :one
SELECT * FROM customers WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCustomerByEmail :one
SELECT * FROM customers 
WHERE shop_id = $1 AND email = $2 AND deleted_at IS NULL;

-- name: ListCustomersByShop :many
SELECT * FROM customers
WHERE shop_id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: ListCustomersByOrganization :many
SELECT * FROM customers
WHERE organization_id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: UpdateCustomer :one
UPDATE customers
SET
    email = COALESCE(sqlc.narg('email'), email),
    phone = COALESCE(sqlc.narg('phone'), phone),
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    accepts_marketing = COALESCE(sqlc.narg('accepts_marketing'), accepts_marketing),
    tags = COALESCE(sqlc.narg('tags'), tags),
    note = COALESCE(sqlc.narg('note'), note),
    tax_exempt = COALESCE(sqlc.narg('tax_exempt'), tax_exempt),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCustomerStats :exec
UPDATE customers
SET 
    orders_count = orders_count + 1,
    total_spent_cents = total_spent_cents + sqlc.arg('amount_cents'),
    updated_at = NOW()
WHERE id = $1;

-- name: UpsertCustomer :one
INSERT INTO customers (shop_id, organization_id, email, phone, first_name, last_name)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (shop_id, email) 
DO UPDATE SET
    phone = COALESCE(EXCLUDED.phone, customers.phone),
    first_name = COALESCE(EXCLUDED.first_name, customers.first_name),
    last_name = COALESCE(EXCLUDED.last_name, customers.last_name),
    updated_at = NOW()
RETURNING *;

-- name: SoftDeleteCustomer :exec
UPDATE customers SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: CountCustomersByShop :one
SELECT COUNT(*) FROM customers
WHERE shop_id = $1 AND deleted_at IS NULL;

-- name: CountCustomersByOrganization :one
SELECT COUNT(*) FROM customers
WHERE organization_id = $1 AND deleted_at IS NULL;

-- name: SearchCustomers :many
SELECT * FROM customers
WHERE shop_id = $1 AND deleted_at IS NULL
  AND (
    email ILIKE '%' || sqlc.arg('query') || '%' 
    OR first_name ILIKE '%' || sqlc.arg('query') || '%'
    OR last_name ILIKE '%' || sqlc.arg('query') || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: SetCustomerPassword :exec
UPDATE customers
SET password_hash = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateMarketingOptIn :one
UPDATE customers
SET 
    accepts_marketing = $2,
    marketing_opt_in_at = CASE WHEN $2 = TRUE THEN NOW() ELSE marketing_opt_in_at END,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;