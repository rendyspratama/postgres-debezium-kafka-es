-- Test case for categories table
-- Created: 2024-03-30

BEGIN;

-- Test 1: Insert test categories
INSERT INTO categories (name, status) VALUES
    ('Test Category 1', 1),
    ('Test Category 2', 1)
RETURNING id;

-- Test 2: Verify category count
DO $$
DECLARE
    category_count INTEGER;
    test_category_id INTEGER;
BEGIN
    -- Basic count check
    SELECT COUNT(*) INTO category_count FROM categories WHERE name LIKE 'Test Category%';
    IF category_count < 2 THEN
        RAISE EXCEPTION 'Category count is less than expected';
    END IF;

    -- Get test category ID for further tests
    SELECT id INTO test_category_id FROM categories WHERE name = 'Test Category 1';

    -- Test 3: Verify timestamps
    IF NOT EXISTS (
        SELECT 1 FROM categories 
        WHERE id = test_category_id 
        AND created_at IS NOT NULL 
        AND updated_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Category timestamps are missing';
    END IF;

    -- Test 4: Verify name constraints
    BEGIN
        INSERT INTO categories (name, status) VALUES (NULL, 1);
        RAISE EXCEPTION 'Should not allow NULL name';
    EXCEPTION
        WHEN not_null_violation THEN
            -- Expected error
    END;

    -- Test 5: Verify status constraints
    BEGIN
        INSERT INTO categories (name, status) VALUES ('Test Category 3', NULL);
        RAISE EXCEPTION 'Should not allow NULL status';
    EXCEPTION
        WHEN not_null_violation THEN
            -- Expected error
    END;

    -- Test 6: Test update functionality
    UPDATE categories SET name = 'Updated Category 1' WHERE id = test_category_id;
    
    IF NOT EXISTS (
        SELECT 1 FROM categories 
        WHERE id = test_category_id 
        AND name = 'Updated Category 1'
        AND updated_at > created_at
    ) THEN
        RAISE EXCEPTION 'Update failed or updated_at not changed';
    END IF;
END $$;

-- Cleanup
ROLLBACK; 