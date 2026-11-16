-- +goose Up

-- Generic "bought / included" checkbox state for items — used on today's lists
-- (and any non-periodic list) as a todo-style toggle. Distinct from the
-- periodic-only need_to_buy state.
ALTER TABLE items ADD COLUMN IF NOT EXISTS is_bought BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down

ALTER TABLE items DROP COLUMN IF EXISTS is_bought;
