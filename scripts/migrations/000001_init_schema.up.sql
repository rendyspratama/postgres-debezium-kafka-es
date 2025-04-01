-- scripts/migrations/000001_init_schema.up.sql

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";


-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    sync_status VARCHAR(50),
    last_sync TIMESTAMP,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Enable replication
ALTER TABLE categories REPLICA IDENTITY FULL;

-- Create replication role and permissions
DO $$
BEGIN
    -- Create role if not exists
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'debezium') THEN
        CREATE ROLE debezium WITH REPLICATION LOGIN PASSWORD 'debezium' SUPERUSER;
    END IF;
END
$$;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE digital_discovery TO debezium;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO debezium;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO debezium;
GRANT USAGE ON SCHEMA public TO debezium;

-- Create publication for debezium
SELECT pg_create_physical_replication_slot('debezium_slot') WHERE NOT EXISTS (
    SELECT 1 FROM pg_replication_slots WHERE slot_name = 'debezium_slot'
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'dbz_publication') THEN
        CREATE PUBLICATION dbz_publication FOR TABLE categories;
    END IF;
END
$$;

-- Create operators table
CREATE TABLE IF NOT EXISTS operators (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on status for faster filtering
CREATE INDEX IF NOT EXISTS idx_operators_status ON operators(status);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category_id INTEGER NOT NULL,
    operator_id INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_operator ON products(operator_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);

-- Create trigger function for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';