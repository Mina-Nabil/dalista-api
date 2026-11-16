-- +goose Up

-- A "today's list" is a one-time, real List generated from a periodic list's
-- need-to-buy items. It behaves like any normal (non-periodic) list; this
-- column links it back to its parent periodic list so it can be nested under
-- it and kept out of the main lists feed.
ALTER TABLE lists ADD COLUMN IF NOT EXISTS source_periodic_list_id UUID REFERENCES lists (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_lists_source_periodic ON lists (source_periodic_list_id, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_lists_source_periodic;
ALTER TABLE lists DROP COLUMN IF EXISTS source_periodic_list_id;
