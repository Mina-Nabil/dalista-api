-- +goose Up

-- Baseline migration: enable the Postgres extensions the schema relies on.
-- pgcrypto provides gen_random_uuid(), used as the DEFAULT for every UUID
-- primary key across Dalista's tables (feature migrations follow in later phases).
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

-- +goose Down
-- Intentionally does not DROP EXTENSION: other objects may depend on it, and
-- dropping a shared extension on rollback is unsafe.
SELECT 1;
