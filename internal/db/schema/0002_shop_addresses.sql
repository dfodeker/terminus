-- +goose Up
-- =============================================================================
-- SHOPS
-- =============================================================================
-- Storefronts within an organization.

CREATE TABLE shops (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    gid             BIGINT       UNIQUE,
    -- Identity
    name                TEXT NOT NULL,
    handle              TEXT NOT NULL,  -- Unique within org
    subdomain           TEXT NOT NULL UNIQUE,  -- Globally unique
    custom_domain       TEXT UNIQUE,
    
    -- Settings
    currency            TEXT NOT NULL DEFAULT 'USD',
    locale              TEXT NOT NULL DEFAULT 'en',
    timezone            TEXT NOT NULL DEFAULT 'UTC',
    
    -- Contact 
    shop_owner          TEXT,
    email               TEXT,
    phone               TEXT,
    
    -- Tracking
    source              TEXT NOT NULL DEFAULT 'organic',
    referral_code       TEXT,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active' 
                        CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    
    -- Unique handle within organization
    CONSTRAINT uq_shop_org_handle UNIQUE (organization_id, handle)
);

CREATE INDEX idx_shops_org ON shops(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_shops_subdomain ON shops(subdomain) WHERE deleted_at IS NULL;
CREATE INDEX idx_shops_custom_domain ON shops(custom_domain) WHERE custom_domain IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_shops_status ON shops(organization_id, status) WHERE deleted_at IS NULL;


-- =============================================================================
-- SHOP MEMBERS
-- =============================================================================
-- Fine-grained access control per shop within an organization.

CREATE TABLE shop_members (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Role within this specific shop
    role                TEXT NOT NULL DEFAULT 'staff'
                        CHECK (role IN ('manager', 'staff', 'viewer')),
    
    -- Permissions (fine-grained)
    permissions         TEXT[] NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_shop_member UNIQUE (shop_id, user_id)
);

CREATE INDEX idx_shop_members_shop ON shop_members(shop_id);
CREATE INDEX idx_shop_members_user ON shop_members(user_id);
CREATE INDEX idx_shop_members_org ON shop_members(organization_id);

-- =============================================================================
-- API CREDENTIALS
-- =============================================================================
-- API keys scoped to organization (all shops) or specific shops.

CREATE TABLE api_credentials (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shop_id             UUID REFERENCES shops(id) ON DELETE CASCADE,  -- NULL = org-wide
    
    -- Key identification
    name                TEXT NOT NULL,
    key_prefix          TEXT NOT NULL,  -- First 8 chars for identification
    key_hash            TEXT NOT NULL,  -- bcrypt hash of full key
    
    -- Permissions
    scopes              TEXT[] NOT NULL DEFAULT '{}',
    
    -- Environment
    environment         TEXT NOT NULL DEFAULT 'live'
                        CHECK (environment IN ('live', 'test')),
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'revoked')),
    
    -- Tracking
    last_used_at        TIMESTAMPTZ,
    created_by          UUID REFERENCES users(id),
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at          TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ
);

CREATE INDEX idx_api_credentials_org ON api_credentials(organization_id);
CREATE INDEX idx_api_credentials_shop ON api_credentials(shop_id) WHERE shop_id IS NOT NULL;
CREATE INDEX idx_api_credentials_prefix ON api_credentials(key_prefix) WHERE status = 'active';



-- =============================================================================
-- ADDRESSES
-- =============================================================================
-- Shared addresses table for reuse across entities.

CREATE TABLE addresses (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Address fields
    first_name          TEXT,
    last_name           TEXT,
    company             TEXT,
    address1            TEXT NOT NULL,
    address2            TEXT,
    city                TEXT NOT NULL,
    province            TEXT,
    province_code       TEXT,
    country             TEXT NOT NULL,
    country_code        TEXT NOT NULL CHECK (length(country_code) = 2),
    zip                 TEXT,
    phone               TEXT,
    
    -- Metadata
    label               TEXT,  -- "Home", "Work", "Warehouse 1", etc.
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_org ON addresses(organization_id);

-- =============================================================================
-- SHOP ADDRESSES
-- =============================================================================
-- Junction table linking shops to addresses.

CREATE TABLE shop_addresses (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    address_id          UUID NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
    
    address_type        TEXT NOT NULL DEFAULT 'primary' 
                        CHECK (address_type IN ('primary', 'billing', 'shipping', 'warehouse', 'return')),
    is_default          BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_shop_address UNIQUE (shop_id, address_id)
);

CREATE INDEX idx_shop_addresses_shop ON shop_addresses(shop_id);
CREATE INDEX idx_shop_addresses_address ON shop_addresses(address_id);

-- Only one default per address type per shop
CREATE UNIQUE INDEX idx_shop_default_address_type 
    ON shop_addresses(shop_id, address_type) 
    WHERE is_default = TRUE;

-- +goose Down

DROP TABLE IF EXISTS shop_addresses;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS api_credentials;
DROP TABLE IF EXISTS shop_members;
DROP TABLE IF EXISTS shops;
