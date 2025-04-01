-- Drop triggers first
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_operators_updated_at ON operators;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in correct order
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS operators;
DROP TABLE IF EXISTS categories;

-- Revoke permissions
REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM debezium;
REVOKE USAGE ON SCHEMA public FROM debezium;

-- Drop publication
DROP PUBLICATION IF EXISTS dbz_publication;

-- Drop role
DROP ROLE IF EXISTS debezium;
