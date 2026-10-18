-- +goose Up

-- Categories: taxonomy for organizing both items and lists. User-creatable;
-- seeded defaults have owner_id NULL + is_system = true (see the seed migration).
CREATE TABLE IF NOT EXISTS categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID REFERENCES users (id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_categories_owner ON categories (owner_id);
-- A user can't have two categories with the same name; system categories
-- (owner_id NULL) are deduplicated separately.
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_owner_name
    ON categories (owner_id, lower(name)) WHERE owner_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_system_name
    ON categories (lower(name)) WHERE owner_id IS NULL;

-- Grouped Lists: home-page containers of Lists.
CREATE TABLE IF NOT EXISTS grouped_lists (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_grouped_lists_owner ON grouped_lists (owner_id);

-- Lists: standalone or nested under a Grouped List. public/private and
-- periodic/standard flags. Optional category (e.g. a public catalog's type).
CREATE TABLE IF NOT EXISTS lists (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    grouped_list_id UUID REFERENCES grouped_lists (id) ON DELETE SET NULL,
    category_id     UUID REFERENCES categories (id) ON DELETE SET NULL,
    name            VARCHAR(255) NOT NULL,
    is_public       BOOLEAN NOT NULL DEFAULT FALSE,
    is_periodic     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_lists_owner ON lists (owner_id);
CREATE INDEX IF NOT EXISTS idx_lists_grouped_list ON lists (grouped_list_id);
CREATE INDEX IF NOT EXISTS idx_lists_public ON lists (is_public) WHERE is_public = TRUE;

-- Items: belong to a List. category_id NULL == the system MISC bucket.
-- owner_id is denormalized from the list for fast ownership checks.
-- attributes holds comparison custom fields (brand/category are first-class).
CREATE TABLE IF NOT EXISTS items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    list_id     UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories (id) ON DELETE SET NULL,
    name        VARCHAR(255) NOT NULL,
    brand       VARCHAR(255),
    price       NUMERIC(12, 2),
    notes       TEXT,
    attributes  JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_items_list ON items (list_id);
CREATE INDEX IF NOT EXISTS idx_items_owner ON items (owner_id);
CREATE INDEX IF NOT EXISTS idx_items_category ON items (category_id);

-- +goose Down
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS lists;
DROP TABLE IF EXISTS grouped_lists;
DROP TABLE IF EXISTS categories;
