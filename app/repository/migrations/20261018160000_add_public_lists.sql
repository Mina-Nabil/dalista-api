-- +goose Up

-- Denormalized star counter on lists, maintained on star/unstar, so the
-- discovery feed can sort public lists by popularity without a join+group.
ALTER TABLE lists ADD COLUMN IF NOT EXISTS star_count INTEGER NOT NULL DEFAULT 0;

-- A user's star on a public list. Unique per (user, list).
CREATE TABLE IF NOT EXISTS public_list_stars (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    list_id    UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_public_list_stars_user_list
    ON public_list_stars (user_id, list_id);
CREATE INDEX IF NOT EXISTS idx_public_list_stars_list ON public_list_stars (list_id);

-- Discovery ordering: public lists by star count (partial index on is_public).
CREATE INDEX IF NOT EXISTS idx_lists_public_star_count
    ON lists (star_count DESC) WHERE is_public = TRUE;

-- +goose Down
DROP INDEX IF EXISTS idx_lists_public_star_count;
DROP TABLE IF EXISTS public_list_stars;
ALTER TABLE lists DROP COLUMN IF EXISTS star_count;
