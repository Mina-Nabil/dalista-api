-- +goose Up

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE share_role AS ENUM ('editor', 'viewer');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- A list shared with a registered user, at a role. Unique per (list, user).
CREATE TABLE IF NOT EXISTS list_shares (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id    UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role       share_role NOT NULL,
    invited_by UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_list_shares_list_user ON list_shares (list_id, user_id);
CREATE INDEX IF NOT EXISTS idx_list_shares_user ON list_shares (user_id);

-- A share addressed to a phone number with no account yet. Resolved into a
-- list_share on phone verification. Unique per (list, phone) while unconsumed.
CREATE TABLE IF NOT EXISTS pending_invites (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id     UUID NOT NULL REFERENCES lists (id) ON DELETE CASCADE,
    phone       VARCHAR(32) NOT NULL,
    role        share_role NOT NULL,
    invited_by  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pending_invites_list_phone
    ON pending_invites (list_id, phone) WHERE consumed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pending_invites_phone ON pending_invites (phone) WHERE consumed_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS pending_invites;
DROP TABLE IF EXISTS list_shares;
DROP TYPE IF EXISTS share_role;
