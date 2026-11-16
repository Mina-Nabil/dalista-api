-- +goose Up

-- Promote item attributes out of the single free-form items.attributes jsonb
-- blob into a dedicated 1:1 side table: well-known attributes as typed columns,
-- plus an `extra` jsonb for arbitrary custom key/values. Keeps `items` lean;
-- brand/price stay first-class on items (comparison keys on them).
CREATE TABLE IF NOT EXISTS item_attributes (
    item_id    UUID PRIMARY KEY REFERENCES items (id) ON DELETE CASCADE,
    url        TEXT NULL,
    location   TEXT NULL,
    facebook   TEXT NULL,
    instagram  TEXT NULL,
    phone      TEXT NULL,
    extra      JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Backfill from the existing items.attributes jsonb: map the well-known keys to
-- their typed columns and everything else into `extra`. Only items whose blob is
-- a non-empty object get a row.
-- +goose StatementBegin
INSERT INTO item_attributes (item_id, url, location, facebook, instagram, phone, extra)
SELECT
    i.id,
    NULLIF(i.attributes ->> 'url', ''),
    NULLIF(i.attributes ->> 'location', ''),
    NULLIF(i.attributes ->> 'facebook', ''),
    NULLIF(i.attributes ->> 'instagram', ''),
    NULLIF(i.attributes ->> 'phone', ''),
    COALESCE(
        (i.attributes::jsonb) - 'url' - 'location' - 'facebook' - 'instagram' - 'phone',
        '{}'::jsonb
    )
FROM items i
WHERE i.attributes IS NOT NULL
  AND jsonb_typeof(i.attributes::jsonb) = 'object'
  AND i.attributes::jsonb <> '{}'::jsonb
ON CONFLICT (item_id) DO NOTHING;
-- +goose StatementEnd

-- The blob is now superseded by the typed table.
ALTER TABLE items DROP COLUMN IF EXISTS attributes;

-- +goose Down

-- Restore the jsonb blob and backfill it from the typed table (well-known
-- columns merged back over `extra`), then drop the side table.
ALTER TABLE items ADD COLUMN IF NOT EXISTS attributes JSONB NOT NULL DEFAULT '{}';

-- +goose StatementBegin
UPDATE items i
SET attributes = (
    a.extra
    || CASE WHEN a.url       IS NOT NULL THEN jsonb_build_object('url', a.url)             ELSE '{}'::jsonb END
    || CASE WHEN a.location  IS NOT NULL THEN jsonb_build_object('location', a.location)   ELSE '{}'::jsonb END
    || CASE WHEN a.facebook  IS NOT NULL THEN jsonb_build_object('facebook', a.facebook)   ELSE '{}'::jsonb END
    || CASE WHEN a.instagram IS NOT NULL THEN jsonb_build_object('instagram', a.instagram) ELSE '{}'::jsonb END
    || CASE WHEN a.phone     IS NOT NULL THEN jsonb_build_object('phone', a.phone)         ELSE '{}'::jsonb END
)
FROM item_attributes a
WHERE a.item_id = i.id;
-- +goose StatementEnd

DROP TABLE IF EXISTS item_attributes;
