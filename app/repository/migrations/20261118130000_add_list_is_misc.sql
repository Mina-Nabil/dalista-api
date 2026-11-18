-- +goose Up

-- is_misc marks a group's auto-managed "MISC" list — the catch-all bucket that
-- holds items added straight from a grouped list's Add button before they're
-- filed into a real list. At most one MISC list may exist per grouped list.
ALTER TABLE lists ADD COLUMN IF NOT EXISTS is_misc BOOLEAN NOT NULL DEFAULT FALSE;

-- Enforce one MISC list per grouped list at the DB level (partial unique index).
-- Standalone MISC lists (null grouped_list_id) are not deduped — Postgres treats
-- nulls as distinct — but the app only ever creates group-scoped MISC lists.
CREATE UNIQUE INDEX IF NOT EXISTS idx_lists_group_misc
  ON lists (grouped_list_id) WHERE is_misc;

-- +goose Down

DROP INDEX IF EXISTS idx_lists_group_misc;
ALTER TABLE lists DROP COLUMN IF EXISTS is_misc;
