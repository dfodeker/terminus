-- +goose Up
UPDATE products
SET inventory_tracked = TRUE
WHERE inventory_tracked IS NULL;

ALTER TABLE products
  ALTER COLUMN inventory_tracked SET DEFAULT TRUE,
  ALTER COLUMN inventory_tracked SET NOT NULL;


UPDATE products
SET status = 'active'
WHERE status IS NULL;

ALTER TABLE products
ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP,
  ALTER COLUMN updated_at SET NOT NULL,
  ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN status SET DEFAULT 'active',
  ALTER COLUMN status SET NOT NULL;


ALTER TABLE products
  DROP CONSTRAINT IF EXISTS products_handle_key;

ALTER TABLE products
  ADD CONSTRAINT uq_products_store_handle UNIQUE (store_id, handle);


ALTER TABLE products
  DROP CONSTRAINT IF EXISTS products_sku_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_products_store_sku
  ON products (store_id, sku)
  WHERE sku IS NOT NULL;


CREATE INDEX IF NOT EXISTS idx_products_store_id ON products(store_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);

-- +goose Down


DROP INDEX IF EXISTS idx_products_status;
DROP INDEX IF EXISTS idx_products_store_id;
DROP INDEX IF EXISTS uq_products_store_sku;

ALTER TABLE products
  DROP CONSTRAINT IF EXISTS uq_products_store_handle;

ALTER TABLE products
  ADD CONSTRAINT products_handle_key UNIQUE (handle);

ALTER TABLE products
  ADD CONSTRAINT products_sku_key UNIQUE (sku);



ALTER TABLE products
  ALTER COLUMN status DROP NOT NULL,
  ALTER COLUMN status DROP DEFAULT;

ALTER TABLE products
  ALTER COLUMN inventory_tracked DROP NOT NULL,
  ALTER COLUMN inventory_tracked DROP DEFAULT;