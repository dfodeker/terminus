-- +goose Up

-- Tenant-level permissions
INSERT INTO permissions (id, key, description, created_at, updated_at) VALUES
    (gen_random_uuid(), 'tenant:owner', 'Owner of the tenant with all permissions', now(), now()),
    (gen_random_uuid(), 'tenant:manage', 'Full control over tenant settings and configuration', now(), now()),
    (gen_random_uuid(), 'tenant:view', 'View tenant information and settings', now(), now()),
    (gen_random_uuid(), 'tenant:invite_users', 'Invite new users to the tenant', now(), now()),
    (gen_random_uuid(), 'tenant:manage_users', 'Manage tenant users, roles, and permissions', now(), now()),
    (gen_random_uuid(), 'tenant:remove_users', 'Remove users from the tenant', now(), now());

-- Store-level permissions
INSERT INTO permissions (id, key, description, created_at, updated_at) VALUES
    (gen_random_uuid(), 'stores:create', 'Create new stores within the tenant', now(), now()),
    (gen_random_uuid(), 'stores:view', 'View store information', now(), now()),
    (gen_random_uuid(), 'stores:edit', 'Edit store settings and configuration', now(), now()),
    (gen_random_uuid(), 'stores:delete', 'Delete stores', now(), now());

-- Product permissions
INSERT INTO permissions (id, key, description, created_at, updated_at) VALUES
    (gen_random_uuid(), 'products:view', 'View products and product details', now(), now()),
    (gen_random_uuid(), 'products:create', 'Create new products', now(), now()),
    (gen_random_uuid(), 'products:edit', 'Edit product information', now(), now()),
    (gen_random_uuid(), 'products:delete', 'Delete products', now(), now());

-- Inventory permissions
INSERT INTO permissions (id, key, description, created_at, updated_at) VALUES
    (gen_random_uuid(), 'inventory:view', 'View inventory levels and history', now(), now()),
    (gen_random_uuid(), 'inventory:manage', 'Manage inventory, adjust stock levels', now(), now());

-- Order permissions (for future use)
INSERT INTO permissions (id, key, description, created_at, updated_at) VALUES
    (gen_random_uuid(), 'orders:view', 'View orders and order details', now(), now()),
    (gen_random_uuid(), 'orders:manage', 'Manage orders, update status, process refunds', now(), now());

-- Analytics permissions
INSERT INTO permissions (id, key, description, created_at, updated_at) VALUES
    (gen_random_uuid(), 'analytics:view', 'View reports and analytics dashboards', now(), now());

-- +goose Down

DELETE FROM permissions WHERE key IN (
    'tenant:owner', 'tenant:manage', 'tenant:view', 'tenant:invite_users', 'tenant:manage_users', 'tenant:remove_users',
    'stores:create', 'stores:view', 'stores:edit', 'stores:delete',
    'products:view', 'products:create', 'products:edit', 'products:delete',
    'inventory:view', 'inventory:manage',
    'orders:view', 'orders:manage',
    'analytics:view'
);
