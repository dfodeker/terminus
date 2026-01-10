-- +goose Up

-- Add gid columns to all entity tables (nullable initially for existing data)
ALTER TABLE tenants ADD COLUMN gid BIGINT UNIQUE;
ALTER TABLE stores ADD COLUMN gid BIGINT UNIQUE;
ALTER TABLE products ADD COLUMN gid BIGINT UNIQUE;
ALTER TABLE product_variants ADD COLUMN gid BIGINT UNIQUE;
ALTER TABLE users ADD COLUMN gid BIGINT UNIQUE;
ALTER TABLE roles ADD COLUMN gid BIGINT UNIQUE;
ALTER TABLE permissions ADD COLUMN gid BIGINT UNIQUE;

-- Create partial indexes for efficient GID lookups (only on non-null values)
CREATE INDEX IF NOT EXISTS idx_tenants_gid ON tenants(gid) WHERE gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_stores_gid ON stores(gid) WHERE gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_gid ON products(gid) WHERE gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_product_variants_gid ON product_variants(gid) WHERE gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_gid ON users(gid) WHERE gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_roles_gid ON roles(gid) WHERE gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_permissions_gid ON permissions(gid) WHERE gid IS NOT NULL;

-- +goose Down

-- Drop indexes
DROP INDEX IF EXISTS idx_permissions_gid;
DROP INDEX IF EXISTS idx_roles_gid;
DROP INDEX IF EXISTS idx_users_gid;
DROP INDEX IF EXISTS idx_product_variants_gid;
DROP INDEX IF EXISTS idx_products_gid;
DROP INDEX IF EXISTS idx_stores_gid;
DROP INDEX IF EXISTS idx_tenants_gid;

-- Remove gid columns
ALTER TABLE permissions DROP COLUMN IF EXISTS gid;
ALTER TABLE roles DROP COLUMN IF EXISTS gid;
ALTER TABLE users DROP COLUMN IF EXISTS gid;
ALTER TABLE product_variants DROP COLUMN IF EXISTS gid;
ALTER TABLE products DROP COLUMN IF EXISTS gid;
ALTER TABLE stores DROP COLUMN IF EXISTS gid;
ALTER TABLE tenants DROP COLUMN IF EXISTS gid;
