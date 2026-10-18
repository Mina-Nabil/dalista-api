-- +goose Up

-- "need to buy" state on items (periodic lists).
ALTER TABLE items ADD COLUMN IF NOT EXISTS need_to_buy BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE item_state AS ENUM ('need_to_buy', 'enough');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- Transition log for the need-to-buy/enough state (drives the auto system comments).
CREATE TABLE IF NOT EXISTS item_state_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id    UUID NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    actor_id   UUID REFERENCES users (id) ON DELETE SET NULL,
    state      item_state NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_item_state_events_item ON item_state_events (item_id, created_at);

-- Dated "today's list" snapshots of a periodic list.
CREATE TABLE IF NOT EXISTS list_variants (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    periodic_list_id UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    title            VARCHAR(255) NOT NULL,
    generated_by     UUID REFERENCES users (id) ON DELETE SET NULL,
    generated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_list_variants_parent ON list_variants (periodic_list_id, generated_at DESC);

-- Frozen item snapshots belonging to a variant.
CREATE TABLE IF NOT EXISTS list_variant_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id     UUID NOT NULL REFERENCES list_variants (id) ON DELETE CASCADE,
    source_item_id UUID REFERENCES items (id) ON DELETE SET NULL,
    name           VARCHAR(255) NOT NULL,
    brand          VARCHAR(255),
    price          NUMERIC(12, 2),
    category_id    UUID REFERENCES categories (id) ON DELETE SET NULL,
    notes          TEXT,
    attributes     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_list_variant_items_variant ON list_variant_items (variant_id);

-- +goose Down
DROP TABLE IF EXISTS list_variant_items;
DROP TABLE IF EXISTS list_variants;
DROP TABLE IF EXISTS item_state_events;
DROP TYPE IF EXISTS item_state;
ALTER TABLE items DROP COLUMN IF EXISTS need_to_buy;
