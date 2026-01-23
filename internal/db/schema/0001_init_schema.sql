-- db/migrations/00001_init_schema.sql

-- +goose Up

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";




-- =============================================================================
-- ORGANIZATIONS
-- =============================================================================
-- An organization is the top-level entity that owns multiple shops.
-- Billing, team members, and org-wide settings are at this level.

CREATE TABLE organizations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Identity
    name                TEXT NOT NULL,
    handle              TEXT NOT NULL UNIQUE,  -- URL-safe identifier
    
    -- Plan & billing
    plan                TEXT NOT NULL DEFAULT 'free' 
                        CHECK (plan IN ('free', 'starter', 'pro', 'enterprise')),
    billing_email       TEXT,
    stripe_customer_id  TEXT,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'suspended', 'deleted')),
    
    -- Limits based on plan
    max_shops           INT NOT NULL DEFAULT 1,
    max_members         INT NOT NULL DEFAULT 5,
    
    -- Settings
    settings            JSONB NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_organizations_handle ON organizations(handle) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_status ON organizations(status) WHERE deleted_at IS NULL;

-- =============================================================================
-- USERS
-- =============================================================================
-- Users can belong to multiple organizations with different roles.

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Identity
    email               TEXT NOT NULL UNIQUE,
    email_verified      BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Authentication
    password_hash       TEXT,  -- NULL if using OAuth only
    
    -- Profile
    first_name          TEXT,
    last_name           TEXT,
    avatar_url          TEXT,
    phone               TEXT,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'invited', 'suspended', 'deleted')),
    
    -- Preferences
    locale              TEXT NOT NULL DEFAULT 'en',
    timezone            TEXT NOT NULL DEFAULT 'UTC',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at       TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;


-- =============================================================================
-- ORGANIZATION MEMBERS
-- =============================================================================
-- Junction table for users belonging to organizations with roles.

CREATE TABLE organization_members (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Role within the organization
    role                TEXT NOT NULL DEFAULT 'member'
                        CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    
    -- Invitation tracking
    invited_by          UUID REFERENCES users(id),
    invited_at          TIMESTAMPTZ,
    accepted_at         TIMESTAMPTZ,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_org_member UNIQUE (organization_id, user_id)
);

CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_org_members_user ON organization_members(user_id);



-- +goose Down
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;