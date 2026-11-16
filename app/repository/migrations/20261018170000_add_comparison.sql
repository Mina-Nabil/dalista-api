-- +goose Up

-- A persisted, shareable comparison: an ordered set of item IDs (JSONB array of
-- UUID strings). Persisted (not stateless) so it can be shared; the items are
-- access-checked per viewer on every read, so no per-item FK is needed here
-- (deleted/inaccessible items are simply withheld at view time).
CREATE TABLE IF NOT EXISTS comparison_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    item_ids    JSONB NOT NULL DEFAULT '[]',
    share_token VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_comparison_sessions_owner ON comparison_sessions (owner_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_comparison_sessions_share_token
    ON comparison_sessions (share_token) WHERE share_token IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS comparison_sessions;
