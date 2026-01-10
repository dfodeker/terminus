-- +goose Up

CREATE TABLE custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gid BIGINT UNIQUE NOT NULL,

    -- Domain information
    domain TEXT UNIQUE NOT NULL,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Verification
    verification_status TEXT NOT NULL DEFAULT 'pending' CHECK (verification_status IN ('pending', 'verified', 'failed', 'expired')),
    verification_token TEXT NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ,

    -- SSL/TLS status
    ssl_status TEXT NOT NULL DEFAULT 'pending' CHECK (ssl_status IN ('pending', 'active', 'failed', 'expired')),
    ssl_expires_at TIMESTAMPTZ,

    -- Primary domain flag (one primary per store)
    is_primary BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_custom_domains_store_id ON custom_domains(store_id);
CREATE INDEX IF NOT EXISTS idx_custom_domains_tenant_id ON custom_domains(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX IF NOT EXISTS idx_custom_domains_verification_status ON custom_domains(verification_status);
CREATE INDEX IF NOT EXISTS idx_custom_domains_gid ON custom_domains(gid);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION generate_domain_verification_token()
RETURNS TEXT AS $$
BEGIN
    RETURN 'mystoreos-verify-' || encode(gen_random_bytes(16), 'hex');
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_domain_verification_token()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.verification_token IS NULL OR NEW.verification_token = '' THEN
        NEW.verification_token := generate_domain_verification_token();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trigger_set_domain_verification_token
    BEFORE INSERT ON custom_domains
    FOR EACH ROW
    EXECUTE FUNCTION set_domain_verification_token();

-- +goose Down

DROP TRIGGER IF EXISTS trigger_set_domain_verification_token ON custom_domains;
DROP FUNCTION IF EXISTS set_domain_verification_token();
DROP FUNCTION IF EXISTS generate_domain_verification_token();
DROP INDEX IF EXISTS idx_custom_domains_gid;
DROP INDEX IF EXISTS idx_custom_domains_verification_status;
DROP INDEX IF EXISTS idx_custom_domains_domain;
DROP INDEX IF EXISTS idx_custom_domains_tenant_id;
DROP INDEX IF EXISTS idx_custom_domains_store_id;
DROP TABLE IF EXISTS custom_domains;
