-- +goose Up

-- Enum types (idempotent per the Lumina convention).
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'admin');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE auth_provider AS ENUM ('email', 'phone', 'google', 'facebook');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE phone_verification_purpose AS ENUM ('signin', 'verify');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- Users. Email is optional; phone is unique when present. Phone is required
-- eventually (post-signup verification) but may be NULL at initial creation for
-- the social/email signup paths, so uniqueness is enforced via partial indexes.
CREATE TABLE IF NOT EXISTS users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          VARCHAR(320),
    phone          VARCHAR(32),
    role           user_role NOT NULL DEFAULT 'user',
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE phone IS NOT NULL;

-- Auth identities: one row per linked provider for a user. password_hash is set
-- only for the 'email' provider. (provider, provider_subject) is globally unique.
CREATE TABLE IF NOT EXISTS auth_identities (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider         auth_provider NOT NULL,
    provider_subject VARCHAR(320) NOT NULL,
    password_hash    VARCHAR(255),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_identities_provider_subject
    ON auth_identities (provider, provider_subject);
CREATE INDEX IF NOT EXISTS idx_auth_identities_user ON auth_identities (user_id);

-- Phone verifications: OTP challenges for both standalone phone sign-in and
-- post-signup phone verification. code_hash stores a SHA-256 of the OTP, never
-- the plaintext. user_id is NULL for a standalone sign-in challenge (no account
-- yet) and set for a post-signup verify challenge.
CREATE TABLE IF NOT EXISTS phone_verifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone       VARCHAR(32) NOT NULL,
    code_hash   VARCHAR(64) NOT NULL,
    purpose     phone_verification_purpose NOT NULL,
    user_id     UUID REFERENCES users (id) ON DELETE CASCADE,
    attempts    INT NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_phone_verifications_phone ON phone_verifications (phone);

-- User profiles (Module 2): display data + settings. One row per user.
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id            UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    display_name       VARCHAR(255),
    picture_url        VARCHAR(1024),
    notification_prefs JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS phone_verifications;
DROP TABLE IF EXISTS auth_identities;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS phone_verification_purpose;
DROP TYPE IF EXISTS auth_provider;
DROP TYPE IF EXISTS user_role;
