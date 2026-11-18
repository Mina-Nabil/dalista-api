-- +goose Up

-- Quantity ("count") for an item on a periodic list: how many to buy when the
-- item is marked need-to-buy. Defaults to 1, shown by default in the UI's count
-- box. Carried forward into generated today's lists and variant snapshots.
ALTER TABLE items ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 1;
ALTER TABLE list_variant_items ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 1;

-- +goose Down

ALTER TABLE items DROP COLUMN IF EXISTS quantity;
ALTER TABLE list_variant_items DROP COLUMN IF EXISTS quantity;
