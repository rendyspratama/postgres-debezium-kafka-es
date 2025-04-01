-- Seed data for digital products and Indonesian operators
-- Created: 2024-03-28 12:00:00

BEGIN;

-- Categories
INSERT INTO categories (id, name, status, created_at, updated_at) VALUES
    (1, 'Pulsa', 1, NOW(), NOW()),
    (2, 'Paket Data', 1, NOW(), NOW()),
    (3, 'Listrik', 1, NOW(), NOW()),
    (4, 'BPJS', 1, NOW(), NOW()),
    (5, 'Roaming', 1, NOW(), NOW());

-- Operators (removed category_id as operators can belong to multiple categories)
INSERT INTO operators (id, name, status, created_at, updated_at) VALUES
    (1, 'Axis', 1, NOW(), NOW()),
    (2, 'Indosat', 1, NOW(), NOW()),
    (3, 'Simpati', 1, NOW(), NOW()),
    (4, 'Smartfren', 1, NOW(), NOW()),
    (5, 'XL', 1, NOW(), NOW()),
    (6, 'Telkomsel', 1, NOW(), NOW()),
    (7, 'PLN', 1, NOW(), NOW()),
    (8, 'BPJS Kesehatan', 1, NOW(), NOW()),
    (9, 'BPJS Ketenagakerjaan', 1, NOW(), NOW());

-- Products (relationship between categories and operators is managed here)
INSERT INTO products (id, name, category_id, operator_id, price, status, created_at, updated_at) VALUES
    -- Pulsa Products
    (1, 'Pulsa 5.000', 1, 1, 5500, 1, NOW(), NOW()),
    (2, 'Pulsa 10.000', 1, 1, 10500, 1, NOW(), NOW()),
    (3, 'Pulsa 20.000', 1, 1, 20500, 1, NOW(), NOW()),
    (4, 'Pulsa 50.000', 1, 1, 50500, 1, NOW(), NOW()),
    (5, 'Pulsa 100.000', 1, 1, 100500, 1, NOW(), NOW()),
    
    (6, 'Pulsa 5.000', 1, 2, 5500, 1, NOW(), NOW()),
    (7, 'Pulsa 10.000', 1, 2, 10500, 1, NOW(), NOW()),
    (8, 'Pulsa 20.000', 1, 2, 20500, 1, NOW(), NOW()),
    (9, 'Pulsa 50.000', 1, 2, 50500, 1, NOW(), NOW()),
    (10, 'Pulsa 100.000', 1, 2, 100500, 1, NOW(), NOW()),
    
    -- Paket Data Products
    (11, 'Paket Data 1GB/7 Hari', 2, 1, 15000, 1, NOW(), NOW()),
    (12, 'Paket Data 3GB/7 Hari', 2, 1, 25000, 1, NOW(), NOW()),
    (13, 'Paket Data 5GB/7 Hari', 2, 1, 35000, 1, NOW(), NOW()),
    (14, 'Paket Data 10GB/30 Hari', 2, 1, 50000, 1, NOW(), NOW()),
    (15, 'Paket Data 20GB/30 Hari', 2, 1, 80000, 1, NOW(), NOW()),
    
    (16, 'Paket Data 1GB/7 Hari', 2, 2, 15000, 1, NOW(), NOW()),
    (17, 'Paket Data 3GB/7 Hari', 2, 2, 25000, 1, NOW(), NOW()),
    (18, 'Paket Data 5GB/7 Hari', 2, 2, 35000, 1, NOW(), NOW()),
    (19, 'Paket Data 10GB/30 Hari', 2, 2, 50000, 1, NOW(), NOW()),
    (20, 'Paket Data 20GB/30 Hari', 2, 2, 80000, 1, NOW(), NOW()),
    
    -- Roaming Products
    (21, 'Roaming Malaysia 1GB/7 Hari', 5, 1, 150000, 1, NOW(), NOW()),
    (22, 'Roaming Singapore 1GB/7 Hari', 5, 1, 180000, 1, NOW(), NOW()),
    (23, 'Roaming Malaysia 1GB/7 Hari', 5, 2, 150000, 1, NOW(), NOW()),
    (24, 'Roaming Singapore 1GB/7 Hari', 5, 2, 180000, 1, NOW(), NOW()),
    
    -- Listrik Products
    (25, 'Token Listrik 20.000', 3, 7, 20000, 1, NOW(), NOW()),
    (26, 'Token Listrik 50.000', 3, 7, 50000, 1, NOW(), NOW()),
    (27, 'Token Listrik 100.000', 3, 7, 100000, 1, NOW(), NOW()),
    (28, 'Token Listrik 200.000', 3, 7, 200000, 1, NOW(), NOW()),
    (29, 'Token Listrik 500.000', 3, 7, 500000, 1, NOW(), NOW()),
    
    -- BPJS Products
    (30, 'BPJS Kesehatan Kelas 1', 4, 8, 150000, 1, NOW(), NOW()),
    (31, 'BPJS Kesehatan Kelas 2', 4, 8, 100000, 1, NOW(), NOW()),
    (32, 'BPJS Kesehatan Kelas 3', 4, 8, 70000, 1, NOW(), NOW()),
    (33, 'BPJS Ketenagakerjaan JHT', 4, 9, 97000, 1, NOW(), NOW()),
    (34, 'BPJS Ketenagakerjaan JKK', 4, 9, 30000, 1, NOW(), NOW());

COMMIT;