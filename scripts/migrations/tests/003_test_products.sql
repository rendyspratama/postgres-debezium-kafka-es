-- Test case for products table
-- Created: 2024-03-30

BEGIN;

-- Setup: Insert test categories and operators
INSERT INTO categories (name, status) 
VALUES ('Test Category', 1) 
RETURNING id AS category_id;

INSERT INTO operators (name, status) 
VALUES ('Test Operator', 1) 
RETURNING id AS operator_id;

DO $$
DECLARE
    test_category_id INTEGER;
    test_operator_id INTEGER;
    test_product_id INTEGER;
BEGIN
    -- Get dependency IDs
    SELECT id INTO test_category_id FROM categories WHERE name = 'Test Category';
    SELECT id INTO test_operator_id FROM operators WHERE name = 'Test Operator';

    -- Test 1: Insert products
    INSERT INTO products (name, category_id, operator_id, price, status) VALUES
        ('Test Product 1', test_category_id, test_operator_id, 10000.00, 1),
        ('Test Product 2', test_category_id, test_operator_id, 20000.00, 1)
    RETURNING id INTO test_product_id;

    -- Test 2: Verify foreign key constraints
    BEGIN
        INSERT INTO products (name, category_id, operator_id, price, status)
        VALUES ('Test Product 3', 999999, test_operator_id, 30000.00, 1);
        RAISE EXCEPTION 'Should not allow invalid category_id';
    EXCEPTION
        WHEN foreign_key_violation THEN
            -- Expected error
    END;

    -- Test 3: Verify price constraints
    BEGIN
        INSERT INTO products (name, category_id, operator_id, price, status)
        VALUES ('Test Product 3', test_category_id, test_operator_id, -1000, 1);
        RAISE EXCEPTION 'Should not allow negative price';
    EXCEPTION
        WHEN check_violation THEN
            -- Expected error
    END;

    -- Test 4: Verify timestamps and update
    UPDATE products 
    SET price = 15000.00 
    WHERE id = test_product_id;

    IF NOT EXISTS (
        SELECT 1 FROM products 
        WHERE id = test_product_id 
        AND price = 15000.00
        AND updated_at > created_at
    ) THEN
        RAISE EXCEPTION 'Update failed or updated_at not changed';
    END IF;

    -- Test 5: Verify cascade delete
    DELETE FROM categories WHERE id = test_category_id;
    
    IF EXISTS (
        SELECT 1 FROM products WHERE category_id = test_category_id
    ) THEN
        RAISE EXCEPTION 'Cascade delete not working for categories';
    END IF;
END $$;

ROLLBACK; 