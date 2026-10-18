-- +goose Up

-- Enum types (idempotent per the Lumina convention).
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE billing_provider AS ENUM ('apple_iap', 'google_play', 'stripe');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE subscription_tier AS ENUM ('free', 'basic', 'unlimited');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE subscription_status AS ENUM ('active', 'grace_period', 'expired', 'canceled', 'revoked');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- Provider-agnostic subscription. One current row per user. 'stripe' is a
-- reserved enum value (not used in v1). Absence of a row == implicit free tier.
CREATE TABLE IF NOT EXISTS subscriptions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    billing_provider      billing_provider NOT NULL,
    tier                  subscription_tier NOT NULL DEFAULT 'free',
    status                subscription_status NOT NULL DEFAULT 'active',
    product_id            VARCHAR(255),
    provider_reference_id VARCHAR(512),
    current_period_end    TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_tier ON subscriptions (tier);
CREATE INDEX IF NOT EXISTS idx_subscriptions_provider ON subscriptions (billing_provider);

-- Apple StoreKit transaction records, keyed by the stable original transaction id.
CREATE TABLE IF NOT EXISTS apple_iap_receipts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    original_transaction_id VARCHAR(255) NOT NULL UNIQUE,
    transaction_id          VARCHAR(255),
    product_id              VARCHAR(255),
    environment             VARCHAR(32),
    expires_at              TIMESTAMPTZ,
    raw_payload             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_apple_receipts_user ON apple_iap_receipts (user_id);

-- Google Play purchase records, keyed by the purchase token.
CREATE TABLE IF NOT EXISTS google_play_purchases (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    purchase_token       VARCHAR(1024) NOT NULL UNIQUE,
    product_id           VARCHAR(255),
    order_id             VARCHAR(255),
    expiry_time          TIMESTAMPTZ,
    acknowledgement_state INT,
    raw_payload          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_google_purchases_user ON google_play_purchases (user_id);

-- Per-user storage accounting, enforced against the tier's quota.
CREATE TABLE IF NOT EXISTS storage_usage (
    user_id    UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    bytes_used BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS storage_usage;
DROP TABLE IF EXISTS google_play_purchases;
DROP TABLE IF EXISTS apple_iap_receipts;
DROP TABLE IF EXISTS subscriptions;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS subscription_tier;
DROP TYPE IF EXISTS billing_provider;
