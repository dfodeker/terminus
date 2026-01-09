-- +goose Up

CREATE TABLE product_variants (
  id             uuid PRIMARY KEY NOT NULL,
  tenant_id      uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  store_id        uuid NOT NULL REFERENCES stores(id)   ON DELETE CASCADE,
  product_id     uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  sku            text,
  barcode        text,
  title          text NOT NULL DEFAULT 'Default Title',
  price_cents    integer NOT NULL CHECK (price_cents >= 0),
  compare_at_cents integer CHECK (compare_at_cents IS NULL OR compare_at_cents >= 0),
  option_values  jsonb NOT NULL DEFAULT '{}'::jsonb, -- {"Size":"M","Color":"Black"}
  status         text NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (store_id, sku) DEFERRABLE INITIALLY IMMEDIATE
);


-- +goose Down
DROP TABLE product_variants;