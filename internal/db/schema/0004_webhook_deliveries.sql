-- +goose Up
-- +goose StatementBegin

-- Webhook deliveries table - stores delivery attempts for webhooks
CREATE TABLE webhook_deliveries (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id          UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shop_id             UUID REFERENCES shops(id) ON DELETE SET NULL,
    
    -- Request details
    topic               TEXT NOT NULL,
    endpoint_url        TEXT NOT NULL,
    request_headers     JSONB,
    request_body        BYTEA,
    
    -- Response details
    response_status     INT,
    response_headers    JSONB,
    response_body       BYTEA,
    
    -- Status tracking
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'success', 'failed', 'retrying')),
    attempts            INT NOT NULL DEFAULT 0,
    next_retry_at       TIMESTAMPTZ,
    error_message       TEXT,
    duration_ms         INT,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at        TIMESTAMPTZ
);

-- Indexes for webhook deliveries
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_org ON webhook_deliveries(organization_id);
CREATE INDEX idx_webhook_deliveries_shop ON webhook_deliveries(shop_id) WHERE shop_id IS NOT NULL;
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_pending_retry ON webhook_deliveries(next_retry_at) 
    WHERE status IN ('pending', 'retrying') AND next_retry_at IS NOT NULL;
CREATE INDEX idx_webhook_deliveries_created ON webhook_deliveries(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS webhook_deliveries;

-- +goose StatementEnd
