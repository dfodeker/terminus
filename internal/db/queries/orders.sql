-- db/queries/orders.sql

-- name: CreateOrder :one
INSERT INTO orders (
    shop_id, organization_id, customer_id, order_number, email, phone,
    currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents,
    status, financial_status, fulfillment_status,
    shipping_address, billing_address, note, note_attributes, tags,
    source_name, source_identifier
) VALUES (
    $1, $2, $3, next_order_number($1), $4, $5,
    $6, $7, $8, $9, $10, $11,
    $12, $13, $14,
    $15, $16, $17, $18, $19,
    $20, $21
)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrderByNumber :one
SELECT * FROM orders WHERE shop_id = $1 AND order_number = $2;

-- name: ListOrdersByShop :many
SELECT * FROM orders
WHERE shop_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('financial_status')::TEXT IS NULL OR financial_status = sqlc.narg('financial_status'))
  AND (sqlc.narg('fulfillment_status')::TEXT IS NULL OR fulfillment_status = sqlc.narg('fulfillment_status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: ListOrdersByCustomer :many
SELECT * FROM orders
WHERE customer_id = $1
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: ListOrdersByOrganization :many
SELECT * FROM orders
WHERE organization_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('cursor')::TIMESTAMPTZ IS NULL OR created_at < sqlc.narg('cursor'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: UpdateOrderStatus :one
UPDATE orders
SET 
    status = COALESCE(sqlc.narg('status'), status),
    financial_status = COALESCE(sqlc.narg('financial_status'), financial_status),
    fulfillment_status = COALESCE(sqlc.narg('fulfillment_status'), fulfillment_status),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ConfirmOrder :one
UPDATE orders
SET status = 'confirmed', processed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelOrder :one
UPDATE orders
SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CloseOrder :one
UPDATE orders
SET status = 'archived', closed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderRefund :one
UPDATE orders
SET 
    refunded_cents = refunded_cents + sqlc.arg('refund_amount'),
    financial_status = CASE 
        WHEN refunded_cents + sqlc.arg('refund_amount') >= total_cents THEN 'refunded'
        ELSE 'partially_refunded'
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateOrderLineItem :one
INSERT INTO order_line_items (
    order_id, shop_id, product_id, variant_id,
    title, variant_title, sku, quantity, fulfillable_quantity,
    unit_price_cents, discount_cents, total_cents,
    taxable, tax_lines, requires_shipping, properties
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: GetOrderLineItems :many
SELECT * FROM order_line_items WHERE order_id = $1;

-- name: UpdateLineItemFulfillment :one
UPDATE order_line_items
SET 
    fulfilled_quantity = fulfilled_quantity + sqlc.arg('quantity'),
    fulfillable_quantity = fulfillable_quantity - sqlc.arg('quantity')
WHERE id = $1 AND fulfillable_quantity >= sqlc.arg('quantity')
RETURNING *;

-- name: GetOrderWithLineItems :many
SELECT 
    o.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', li.id,
                'title', li.title,
                'variant_title', li.variant_title,
                'sku', li.sku,
                'quantity', li.quantity,
                'fulfilled_quantity', li.fulfilled_quantity,
                'unit_price_cents', li.unit_price_cents,
                'discount_cents', li.discount_cents,
                'total_cents', li.total_cents
            )
        ) FILTER (WHERE li.id IS NOT NULL),
        '[]'
    ) as line_items
FROM orders o
LEFT JOIN order_line_items li ON li.order_id = o.id
WHERE o.id = $1
GROUP BY o.id;

-- name: CountOrdersByShop :one
SELECT COUNT(*) FROM orders
WHERE shop_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'));

-- name: CountOrdersByCustomer :one
SELECT COUNT(*) FROM orders WHERE customer_id = $1;

-- name: GetOrderTotalsByShop :one
SELECT 
    COUNT(*) as order_count,
    COALESCE(SUM(total_cents), 0) as total_revenue_cents,
    COALESCE(AVG(total_cents), 0) as average_order_cents
FROM orders
WHERE shop_id = $1 AND status != 'cancelled';