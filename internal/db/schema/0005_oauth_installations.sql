-- +goose Up
-- =============================================================================
-- OAUTH INSTALLATIONS
-- =============================================================================
-- Tracks OAuth app installations for shops.

CREATE TABLE oauth_installations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    app_id              UUID NOT NULL,

    -- Token (hashed for security)
    access_token_hash   TEXT NOT NULL,

    -- Permissions granted
    scopes              TEXT[] NOT NULL DEFAULT '{}',

    -- Status
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'uninstalled', 'suspended')),

    -- Timestamps
    installed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uninstalled_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique installation per app per shop
    CONSTRAINT uq_oauth_installation UNIQUE (shop_id, app_id)
);

CREATE INDEX idx_oauth_installations_shop ON oauth_installations(shop_id);
CREATE INDEX idx_oauth_installations_app ON oauth_installations(app_id);
CREATE INDEX idx_oauth_installations_token ON oauth_installations(access_token_hash) WHERE status = 'active';

-- =============================================================================
-- REFRESH TOKENS
-- =============================================================================
-- Stores refresh tokens for user sessions.

CREATE TABLE refresh_tokens (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Token (hashed for security)
    token_hash          TEXT NOT NULL UNIQUE,

    -- Session info
    user_agent          TEXT,
    ip_address          INET,

    -- Status
    revoked             BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at          TIMESTAMPTZ,

    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ NOT NULL,
    last_used_at        TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id) WHERE revoked = FALSE;
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash) WHERE revoked = FALSE;

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS oauth_installations;
