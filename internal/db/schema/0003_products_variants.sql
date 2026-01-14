-- +goose Up

-- =============================================================================
-- PRODUCTS
-- =============================================================================

CREATE TABLE products (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Core fields
    title               TEXT NOT NULL,
    description         TEXT,
    description_html    TEXT,
    handle                TEXT NOT NULL,
    
    -- Pricing (in cents)
    price_cents         BIGINT NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    compare_at_price_cents BIGINT CHECK (compare_at_price_cents IS NULL OR compare_at_price_cents >= 0),
    cost_cents          BIGINT CHECK (cost_cents IS NULL OR cost_cents >= 0),
    
    -- Organization
    vendor              TEXT,
    product_type        TEXT,
    tags                TEXT[] NOT NULL DEFAULT '{}',
    
    -- Inventory
    sku                 TEXT,
    barcode             TEXT,
    inventory_quantity  INT NOT NULL DEFAULT 0,
    track_inventory     BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'draft' 
                        CHECK (status IN ('draft', 'active', 'archived')),
    
    -- SEO
    seo_title           TEXT,
    seo_description     TEXT,
    
    -- Template
    template_suffix     TEXT,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at        TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    
    CONSTRAINT uq_product_shop_handle UNIQUE (shop_id, handle)
);

CREATE INDEX idx_products_shop ON products(shop_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_org ON products(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_shop_status ON products(shop_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_shop_created ON products(shop_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_shop_sku ON products(shop_id, sku) WHERE sku IS NOT NULL AND deleted_at IS NULL;



-- =============================================================================
-- PRODUCT VARIANTS
-- =============================================================================

CREATE TABLE product_variants (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id          UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    
    -- Core fields
    title               TEXT NOT NULL,  -- e.g., "Large / Red"
    
    -- Pricing
    price_cents         BIGINT NOT NULL CHECK (price_cents >= 0),
    compare_at_price_cents BIGINT CHECK (compare_at_price_cents IS NULL OR compare_at_price_cents >= 0),
    cost_cents          BIGINT CHECK (cost_cents IS NULL OR cost_cents >= 0),
    
    -- Inventory
    sku                 TEXT,
    barcode             TEXT,
    inventory_quantity  INT NOT NULL DEFAULT 0,
    inventory_policy    TEXT NOT NULL DEFAULT 'deny' 
                        CHECK (inventory_policy IN ('deny', 'continue')),
    
    -- Options (up to 3)
    option1_name        TEXT,
    option1_value       TEXT,
    option2_name        TEXT,
    option2_value       TEXT,
    option3_name        TEXT,
    option3_value       TEXT,
    
    -- Shipping
    weight_grams        INT,
    requires_shipping   BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Position
    position            INT NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_variants_product ON product_variants(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_variants_shop ON product_variants(shop_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_variants_shop_sku ON product_variants(shop_id, sku) WHERE sku IS NOT NULL AND deleted_at IS NULL;

-- Unique SKU per shop (if SKU is set)
CREATE UNIQUE INDEX idx_variants_unique_sku 
    ON product_variants(shop_id, sku) 
    WHERE sku IS NOT NULL AND sku != '' AND deleted_at IS NULL;

-- =============================================================================
-- PRODUCT IMAGES
-- =============================================================================

CREATE TABLE product_images (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id          UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    
    -- Image data
    src                 TEXT NOT NULL,
    alt                 TEXT,
    width               INT,
    height              INT,
    
    -- Position
    position            INT NOT NULL DEFAULT 0,
    
    -- Variant association
    variant_ids         UUID[] NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_images_product ON product_images(product_id);
CREATE INDEX idx_product_images_shop ON product_images(shop_id);

-- =============================================================================
-- COLLECTIONS
-- =============================================================================

CREATE TABLE collections (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Core fields
    title               TEXT NOT NULL,
    slug                TEXT NOT NULL,
    description         TEXT,
    description_html    TEXT,
    
    -- Type
    collection_type     TEXT NOT NULL DEFAULT 'manual'
                        CHECK (collection_type IN ('manual', 'smart')),
    
    -- Smart collection rules (JSON)
    rules               JSONB,
    disjunctive         BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Image
    image_url           TEXT,
    
    -- SEO
    seo_title           TEXT,
    seo_description     TEXT,
    
    -- Sorting
    sort_order          TEXT NOT NULL DEFAULT 'manual'
                        CHECK (sort_order IN (
                            'manual', 'best-selling', 'alpha-asc', 'alpha-desc',
                            'price-asc', 'price-desc', 'created-desc', 'created-asc'
                        )),
    
    -- Status
    published           BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at        TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    
    CONSTRAINT uq_collection_shop_slug UNIQUE (shop_id, slug)
);

CREATE INDEX idx_collections_shop ON collections(shop_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_org ON collections(organization_id) WHERE deleted_at IS NULL;

-- =============================================================================
-- COLLECTION PRODUCTS
-- =============================================================================

CREATE TABLE collection_products (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id       UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    product_id          UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Position for manual ordering
    position            INT NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_collection_product UNIQUE (collection_id, product_id)
);

CREATE INDEX idx_collection_products_collection ON collection_products(collection_id);
CREATE INDEX idx_collection_products_product ON collection_products(product_id);


-- =============================================================================
-- CUSTOMERS
-- =============================================================================

CREATE TABLE customers (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Identity
    email               TEXT NOT NULL,
    phone               TEXT,
    first_name          TEXT,
    last_name           TEXT,
    
    -- Auth (optional, for customer accounts)
    password_hash       TEXT,
    
    -- Marketing
    accepts_marketing   BOOLEAN NOT NULL DEFAULT FALSE,
    marketing_opt_in_at TIMESTAMPTZ,
    
    -- Stats (denormalized)
    orders_count        INT NOT NULL DEFAULT 0,
    total_spent_cents   BIGINT NOT NULL DEFAULT 0,
    
    -- Tags & notes
    tags                TEXT[] NOT NULL DEFAULT '{}',
    note                TEXT,
    
    -- Tax
    tax_exempt          BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active' 
                        CHECK (status IN ('active', 'disabled', 'invited')),
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    
    CONSTRAINT uq_customer_shop_email UNIQUE (shop_id, email)
);

CREATE INDEX idx_customers_shop ON customers(shop_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_org ON customers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_shop_email ON customers(shop_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_shop_created ON customers(shop_id, created_at DESC) WHERE deleted_at IS NULL;

-- =============================================================================
-- CUSTOMER ADDRESSES
-- =============================================================================

CREATE TABLE customer_addresses (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id         UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    address_id          UUID NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    
    -- Default flags
    is_default_billing  BOOLEAN NOT NULL DEFAULT FALSE,
    is_default_shipping BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_customer_address UNIQUE (customer_id, address_id)
);

CREATE INDEX idx_customer_addresses_customer ON customer_addresses(customer_id);
CREATE INDEX idx_customer_addresses_shop ON customer_addresses(shop_id);

-- Only one default billing/shipping per customer
CREATE UNIQUE INDEX idx_customer_default_billing 
    ON customer_addresses(customer_id) 
    WHERE is_default_billing = TRUE;

CREATE UNIQUE INDEX idx_customer_default_shipping 
    ON customer_addresses(customer_id) 
    WHERE is_default_shipping = TRUE;

-- =============================================================================
-- ORDERS
-- =============================================================================

CREATE TABLE orders (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id         UUID REFERENCES customers(id) ON DELETE SET NULL,
    
    -- Human-readable order number (per shop)
    order_number        BIGINT NOT NULL,
    
    -- Contact
    email               TEXT NOT NULL,
    phone               TEXT,
    
    -- Currency
    currency            TEXT NOT NULL DEFAULT 'USD',
    
    -- Pricing (all in cents)
    subtotal_cents      BIGINT NOT NULL DEFAULT 0,
    discount_cents      BIGINT NOT NULL DEFAULT 0,
    shipping_cents      BIGINT NOT NULL DEFAULT 0,
    tax_cents           BIGINT NOT NULL DEFAULT 0,
    total_cents         BIGINT NOT NULL DEFAULT 0,
    
    -- Refunds
    refunded_cents      BIGINT NOT NULL DEFAULT 0,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'confirmed', 'cancelled', 'archived')),
    financial_status    TEXT NOT NULL DEFAULT 'pending' 
                        CHECK (financial_status IN (
                            'pending', 'authorized', 'paid', 'partially_paid', 
                            'refunded', 'partially_refunded', 'voided'
                        )),
    fulfillment_status  TEXT 
                        CHECK (fulfillment_status IS NULL OR fulfillment_status IN (
                            'unfulfilled', 'partial', 'fulfilled', 'restocked'
                        )),
    
    -- Addresses (JSONB snapshots - immutable after order)
    shipping_address    JSONB,
    billing_address     JSONB,
    
    -- Notes
    note                TEXT,
    note_attributes     JSONB,
    
    -- Tags
    tags                TEXT[] NOT NULL DEFAULT '{}',
    
    -- Source
    source_name         TEXT NOT NULL DEFAULT 'web',
    source_identifier   TEXT,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    closed_at           TIMESTAMPTZ,
    
    CONSTRAINT uq_order_shop_number UNIQUE (shop_id, order_number)
);

CREATE INDEX idx_orders_shop ON orders(shop_id);
CREATE INDEX idx_orders_org ON orders(organization_id);
CREATE INDEX idx_orders_customer ON orders(customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_orders_shop_created ON orders(shop_id, created_at DESC);
CREATE INDEX idx_orders_shop_status ON orders(shop_id, status, financial_status);

-- =============================================================================
-- ORDER LINE ITEMS
-- =============================================================================

CREATE TABLE order_line_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    
    -- Product reference (can be NULL if deleted)
    product_id          UUID REFERENCES products(id) ON DELETE SET NULL,
    variant_id          UUID REFERENCES product_variants(id) ON DELETE SET NULL,
    
    -- Snapshot at time of order
    title               TEXT NOT NULL,
    variant_title       TEXT,
    sku                 TEXT,
    
    -- Quantity
    quantity            INT NOT NULL CHECK (quantity > 0),
    fulfillable_quantity INT NOT NULL DEFAULT 0,
    fulfilled_quantity  INT NOT NULL DEFAULT 0,
    
    -- Pricing (per unit, in cents)
    unit_price_cents    BIGINT NOT NULL,
    discount_cents      BIGINT NOT NULL DEFAULT 0,
    total_cents         BIGINT NOT NULL,
    
    -- Tax
    taxable             BOOLEAN NOT NULL DEFAULT TRUE,
    tax_lines           JSONB,
    
    -- Fulfillment
    requires_shipping   BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Custom properties
    properties          JSONB,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_line_items_order ON order_line_items(order_id);
CREATE INDEX idx_line_items_shop ON order_line_items(shop_id);
CREATE INDEX idx_line_items_product ON order_line_items(product_id) WHERE product_id IS NOT NULL;

-- =============================================================================
-- TRANSACTIONS
-- =============================================================================

CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    
    -- Type
    kind                TEXT NOT NULL
                        CHECK (kind IN ('authorization', 'capture', 'sale', 'void', 'refund')),
    
    -- Amount
    amount_cents        BIGINT NOT NULL,
    currency            TEXT NOT NULL DEFAULT 'USD',
    
    -- Status
    status              TEXT NOT NULL
                        CHECK (status IN ('pending', 'success', 'failure', 'error')),
    
    -- Gateway
    gateway             TEXT NOT NULL,
    gateway_transaction_id TEXT,
    
    -- Error info
    error_code          TEXT,
    error_message       TEXT,
    
    -- Parent (for refunds)
    parent_id           UUID REFERENCES transactions(id),
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at        TIMESTAMPTZ
);

CREATE INDEX idx_transactions_order ON transactions(order_id);
CREATE INDEX idx_transactions_shop ON transactions(shop_id);

-- =============================================================================
-- FULFILLMENTS
-- =============================================================================

CREATE TABLE fulfillments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'open', 'success', 'cancelled', 'error', 'failure')),
    
    -- Tracking
    tracking_company    TEXT,
    tracking_number     TEXT,
    tracking_url        TEXT,
    tracking_numbers    TEXT[] NOT NULL DEFAULT '{}',
    tracking_urls       TEXT[] NOT NULL DEFAULT '{}',
    
    -- Shipment status
    shipment_status     TEXT,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fulfillments_order ON fulfillments(order_id);
CREATE INDEX idx_fulfillments_shop ON fulfillments(shop_id);

-- =============================================================================
-- FULFILLMENT LINE ITEMS
-- =============================================================================

CREATE TABLE fulfillment_line_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fulfillment_id      UUID NOT NULL REFERENCES fulfillments(id) ON DELETE CASCADE,
    order_line_item_id  UUID NOT NULL REFERENCES order_line_items(id) ON DELETE CASCADE,
    
    quantity            INT NOT NULL CHECK (quantity > 0),
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fulfillment_items_fulfillment ON fulfillment_line_items(fulfillment_id);
CREATE INDEX idx_fulfillment_items_line_item ON fulfillment_line_items(order_line_item_id);

-- =============================================================================
-- THEMES
-- =============================================================================

CREATE TABLE themes (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Identity
    name                TEXT NOT NULL,
    handle              TEXT NOT NULL,
    
    -- Source
    source_type         TEXT NOT NULL DEFAULT 'custom'
                        CHECK (source_type IN ('marketplace', 'custom', 'default')),
    source_theme_id     UUID,
    version             TEXT,
    
    -- Role (only one 'main' per shop)
    role                TEXT NOT NULL DEFAULT 'unpublished'
                        CHECK (role IN ('main', 'unpublished', 'demo')),
    
    -- Processing
    processing_status   TEXT NOT NULL DEFAULT 'ready'
                        CHECK (processing_status IN ('ready', 'processing', 'error')),
    processing_error    TEXT,
    
    -- Storage
    storage_bucket      TEXT NOT NULL,
    storage_prefix      TEXT NOT NULL,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at        TIMESTAMPTZ
);

CREATE INDEX idx_themes_shop ON themes(shop_id);
CREATE INDEX idx_themes_org ON themes(organization_id);

-- Only one main theme per shop
CREATE UNIQUE INDEX idx_themes_shop_main 
    ON themes(shop_id) 
    WHERE role = 'main';

-- =============================================================================
-- THEME SETTINGS
-- =============================================================================

CREATE TABLE theme_settings (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    theme_id            UUID NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    
    -- Settings data
    settings_data       JSONB NOT NULL DEFAULT '{}',
    templates_data      JSONB NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_theme_settings_theme ON theme_settings(theme_id);

-- =============================================================================
-- WEBHOOKS
-- =============================================================================

CREATE TABLE webhooks (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shop_id             UUID REFERENCES shops(id) ON DELETE CASCADE,  -- NULL = all shops
    
    -- Config
    topic               TEXT NOT NULL,
    endpoint_url        TEXT NOT NULL,
    
    -- Security
    secret_hash         TEXT NOT NULL,
    
    -- Filtering
    fields              TEXT[],
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'paused', 'disabled')),
    
    -- Failure tracking
    failure_count       INT NOT NULL DEFAULT 0,
    last_failure_at     TIMESTAMPTZ,
    last_failure_reason TEXT,
    
    -- API version
    api_version         TEXT NOT NULL DEFAULT '2025-01',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_org ON webhooks(organization_id);
CREATE INDEX idx_webhooks_shop ON webhooks(shop_id) WHERE shop_id IS NOT NULL;
CREATE INDEX idx_webhooks_topic ON webhooks(organization_id, topic);

-- =============================================================================
-- AUDIT LOG
-- =============================================================================

CREATE TABLE audit_logs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shop_id             UUID REFERENCES shops(id) ON DELETE SET NULL,
    user_id             UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- What happened
    action              TEXT NOT NULL,
    resource_type       TEXT NOT NULL,
    resource_id         UUID,
    
    -- Changes
    changes             JSONB,
    
    -- Context
    ip_address          INET,
    user_agent          TEXT,
    api_credential_id   UUID,
    
    -- Timestamp
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_shop ON audit_logs(shop_id, created_at DESC) WHERE shop_id IS NOT NULL;
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- =============================================================================
-- SHOP SEQUENCES
-- =============================================================================

CREATE TABLE shop_sequences (
    shop_id             UUID PRIMARY KEY REFERENCES shops(id) ON DELETE CASCADE,
    order_number        BIGINT NOT NULL DEFAULT 1000,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

-- Get next order number for a shop
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION next_order_number(p_shop_id UUID)
RETURNS BIGINT AS $$
DECLARE
    v_number BIGINT;
BEGIN
    UPDATE shop_sequences 
    SET order_number = order_number + 1,
        updated_at = NOW()
    WHERE shop_id = p_shop_id
    RETURNING order_number INTO v_number;
    
    IF v_number IS NULL THEN
        INSERT INTO shop_sequences (shop_id, order_number) 
        VALUES (p_shop_id, 1001)
        RETURNING order_number INTO v_number;
    END IF;
    
    RETURN v_number;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Auto-update updated_at timestamp
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =============================================================================
-- TRIGGERS
-- =============================================================================

CREATE TRIGGER trg_organizations_updated_at 
    BEFORE UPDATE ON organizations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_organization_members_updated_at 
    BEFORE UPDATE ON organization_members 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_shops_updated_at 
    BEFORE UPDATE ON shops 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_shop_members_updated_at 
    BEFORE UPDATE ON shop_members 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_api_credentials_updated_at 
    BEFORE UPDATE ON api_credentials 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_addresses_updated_at 
    BEFORE UPDATE ON addresses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_product_variants_updated_at 
    BEFORE UPDATE ON product_variants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_product_images_updated_at 
    BEFORE UPDATE ON product_images 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_collections_updated_at 
    BEFORE UPDATE ON collections 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_customers_updated_at 
    BEFORE UPDATE ON customers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_fulfillments_updated_at 
    BEFORE UPDATE ON fulfillments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_themes_updated_at 
    BEFORE UPDATE ON themes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_theme_settings_updated_at 
    BEFORE UPDATE ON theme_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_webhooks_updated_at 
    BEFORE UPDATE ON webhooks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- +goose Down

-- Drop triggers
DROP TRIGGER IF EXISTS trg_webhooks_updated_at ON webhooks;
DROP TRIGGER IF EXISTS trg_theme_settings_updated_at ON theme_settings;
DROP TRIGGER IF EXISTS trg_themes_updated_at ON themes;
DROP TRIGGER IF EXISTS trg_fulfillments_updated_at ON fulfillments;
DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS trg_customers_updated_at ON customers;
DROP TRIGGER IF EXISTS trg_collections_updated_at ON collections;
DROP TRIGGER IF EXISTS trg_product_images_updated_at ON product_images;
DROP TRIGGER IF EXISTS trg_product_variants_updated_at ON product_variants;
DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
DROP TRIGGER IF EXISTS trg_addresses_updated_at ON addresses;
DROP TRIGGER IF EXISTS trg_api_credentials_updated_at ON api_credentials;
DROP TRIGGER IF EXISTS trg_shop_members_updated_at ON shop_members;
DROP TRIGGER IF EXISTS trg_shops_updated_at ON shops;
DROP TRIGGER IF EXISTS trg_organization_members_updated_at ON organization_members;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TRIGGER IF EXISTS trg_organizations_updated_at ON organizations;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at();
DROP FUNCTION IF EXISTS next_order_number(UUID);

-- Drop tables (reverse order of creation)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS theme_settings;
DROP TABLE IF EXISTS themes;
DROP TABLE IF EXISTS shop_sequences;
DROP TABLE IF EXISTS fulfillment_line_items;
DROP TABLE IF EXISTS fulfillments;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS order_line_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customer_addresses;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS collection_products;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;


-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
