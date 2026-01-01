-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY NOT NULL,
    store_id UUID NOT NULL,
    handle VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    inventory_tracked BOOLEAN DEFAULT TRUE,
    sku VARCHAR(100) UNIQUE,
    tags TEXT,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE products;