-- Test case for operators table
-- Created: 2024-03-30

BEGIN;

-- Test 1: Insert test operators
INSERT INTO operators (name, status) VALUES
    ('Test Operator 1', 1),
    ('Test Operator 2', 1)
RETURNING id;

DO $$
DECLARE
    operator_count INTEGER;
    test_operator_id INTEGER;
BEGIN
    -- Get test operator ID
    SELECT id INTO test_operator_id FROM operators WHERE name = 'Test Operator 1';

    -- Test 2: Verify operator count
    SELECT COUNT(*) INTO operator_count FROM operators WHERE name LIKE 'Test Operator%';
    IF operator_count < 2 THEN
        RAISE EXCEPTION 'Operator count is less than expected';
    END IF;

    -- Test 3: Verify timestamps
    IF NOT EXISTS (
        SELECT 1 FROM operators 
        WHERE id = test_operator_id 
        AND created_at IS NOT NULL 
        AND updated_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Operator timestamps are missing';
    END IF;

    -- Test 4: Verify name constraints
    BEGIN
        INSERT INTO operators (name, status) VALUES (NULL, 1);
        RAISE EXCEPTION 'Should not allow NULL name';
    EXCEPTION
        WHEN not_null_violation THEN
            -- Expected error
    END;

    -- Test 5: Verify empty name constraint
    BEGIN
        INSERT INTO operators (name, status) VALUES ('', 1);
        RAISE EXCEPTION 'Should not allow empty name';
    EXCEPTION
        WHEN check_violation THEN
            -- Expected error
    END;

    -- Test 6: Test update with timestamp check
    UPDATE operators SET name = 'Updated Operator 1' WHERE id = test_operator_id;
    
    IF NOT EXISTS (
        SELECT 1 FROM operators 
        WHERE id = test_operator_id 
        AND name = 'Updated Operator 1'
        AND updated_at > created_at
    ) THEN
        RAISE EXCEPTION 'Update failed or updated_at not changed';
    END IF;

    -- Test 7: Verify status values
    BEGIN
        INSERT INTO operators (name, status) VALUES ('Test Operator 3', 999);
        RAISE EXCEPTION 'Should not allow invalid status value';
    EXCEPTION
        WHEN check_violation THEN
            -- Expected error
    END;
END $$;

ROLLBACK; 