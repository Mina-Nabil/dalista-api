-- +goose Up

-- Comments on items. list_id is denormalized from the item for access scoping.
-- author_id is NULL for system-generated comments (is_system = true).
CREATE TABLE IF NOT EXISTS comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id    UUID NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    list_id    UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    author_id  UUID REFERENCES users (id) ON DELETE SET NULL,
    body       TEXT NOT NULL,
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_comments_item ON comments (item_id, created_at);

-- Activity log powering the Timeline feed. Scoped by list_id; actor_id NULL
-- for system events.
CREATE TABLE IF NOT EXISTS activity_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id     UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    actor_id    UUID REFERENCES users (id) ON DELETE SET NULL,
    action      VARCHAR(64) NOT NULL,
    target_type VARCHAR(32),
    target_id   UUID,
    payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_list_created ON activity_events (list_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS comments;
