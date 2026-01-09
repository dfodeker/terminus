-- +goose Up

-- Add tenant_id column to stores (nullable initially for existing data)
ALTER TABLE stores ADD COLUMN tenant_id UUID;

-- Add foreign key constraint
ALTER TABLE stores ADD CONSTRAINT fk_stores_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for efficient tenant-based queries
CREATE INDEX IF NOT EXISTS idx_stores_tenant_id ON stores(tenant_id);

-- Drop the global unique constraint on handle
ALTER TABLE stores DROP CONSTRAINT IF EXISTS stores_handle_key;

-- Add unique constraint for handle within a tenant (handle must be unique per tenant)
ALTER TABLE stores ADD CONSTRAINT stores_tenant_handle_unique UNIQUE (tenant_id, handle);

-- +goose Down

-- Remove the tenant-scoped unique constraint
ALTER TABLE stores DROP CONSTRAINT IF EXISTS stores_tenant_handle_unique;

-- Restore global unique constraint on handle
ALTER TABLE stores ADD CONSTRAINT stores_handle_key UNIQUE (handle);

-- Drop the index
DROP INDEX IF EXISTS idx_stores_tenant_id;

-- Remove foreign key constraint
ALTER TABLE stores DROP CONSTRAINT IF EXISTS fk_stores_tenant;

-- Remove tenant_id column
ALTER TABLE stores DROP COLUMN IF EXISTS tenant_id;
