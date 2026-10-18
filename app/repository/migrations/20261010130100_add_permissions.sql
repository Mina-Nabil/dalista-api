-- +goose Up

-- Permission registry tables (DAL-33). Rows are populated from the Go-defined
-- registry by the `syncpermissions` command, not by migrations. These tables
-- just establish the schema.

-- Canonical list of every permission string known to the app.
CREATE TABLE IF NOT EXISTS permissions (
    permission VARCHAR(128) PRIMARY KEY
);

-- Default permissions granted to each role.
CREATE TABLE IF NOT EXISTS role_permissions (
    role       VARCHAR(32) NOT NULL,
    permission VARCHAR(128) NOT NULL,
    PRIMARY KEY (role, permission)
);

-- Per-user permission grants on top of role defaults.
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    permission VARCHAR(128) NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, permission)
);

-- +goose Down
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
