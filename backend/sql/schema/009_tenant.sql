-- +goose Up

CREATE TABLE tenants (
    id UUID PRIMARY KEY NOT NULL,
    name TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','suspended', 'deleted')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenant_users (
    id UUID PRIMARY KEY NOT NULL,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','invited','removed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_id ON tenant_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_users_user_id ON tenant_users(user_id);

CREATE TABLE roles (
    id UUID PRIMARY KEY NOT NULL,
    tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         text NOT NULL,
    description TEXT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    UNIQUE (tenant_id, name)
);


CREATE TABLE permissions (
    id UUID PRIMARY KEY NOT NULL,
    key TEXT UNIQUE NOT NULL, --eg 'products:read', 'products:write'
    description TEXT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE role_permissions (
  role_id       uuid NOT NULL REFERENCES roles(id)        ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions(id)  ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE tenant_user_roles (
  tenant_user_id uuid NOT NULL REFERENCES tenant_users(id) ON DELETE CASCADE,
  role_id        uuid NOT NULL REFERENCES roles(id)        ON DELETE CASCADE,
  PRIMARY KEY (tenant_user_id, role_id)
);


-- +goose Down
DROP TABLE IF EXISTS tenant_user_roles;
DROP TABLE IF EXISTS role_permissions;


DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;


DROP INDEX IF EXISTS idx_tenant_users_user_id;
DROP INDEX IF EXISTS idx_tenant_users_tenant_id;


DROP TABLE IF EXISTS tenant_users;
DROP TABLE IF EXISTS tenants;