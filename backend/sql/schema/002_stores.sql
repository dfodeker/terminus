-- +goose Up
CREATE TABLE stores (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(255) NOT NULL,
    handle VARCHAR(100) UNIQUE NOT NULL,
    address TEXT,
    status VARCHAR(50) DEFAULT 'active',
    default_currency VARCHAR(10) DEFAULT 'USD',
    timezone VARCHAR(50),
    plan VARCHAR(50) DEFAULT 'free',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);




-- +goose Down
DROP TABLE stores;