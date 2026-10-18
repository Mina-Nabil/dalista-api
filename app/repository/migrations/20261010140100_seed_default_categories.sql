-- +goose Up

-- Seed the default system categories (owner_id NULL, is_system = true). These
-- are visible to every user alongside their own categories. Idempotent via the
-- partial unique index idx_categories_system_name.
INSERT INTO categories (name, is_system)
VALUES
    ('Groceries', TRUE),
    ('Furniture', TRUE),
    ('Electronics', TRUE),
    ('Clothing', TRUE),
    ('Home & Kitchen', TRUE),
    ('Health & Beauty', TRUE),
    ('Toys & Games', TRUE),
    ('Books & Stationery', TRUE),
    ('Sports & Outdoors', TRUE),
    ('Pets', TRUE)
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM categories WHERE owner_id IS NULL AND is_system = TRUE;
