-- +goose Up

-- Photos attached to an item. Bytes live in object storage (object_key); only
-- metadata is stored here. size_bytes is counted against the owner's storage
-- quota; position orders the images in the add/edit grid.
CREATE TABLE IF NOT EXISTS item_images (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id      UUID NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    object_key   TEXT NOT NULL,
    url          TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL,
    position     INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_item_images_item ON item_images (item_id, position);

-- +goose Down
DROP TABLE IF EXISTS item_images;
