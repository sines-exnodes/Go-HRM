-- =========================================================================
-- 000002_create_roles_users
-- roles, users, user_roles + audit columns + triggers + indexes
-- users is LEAN (auth only). HR fields live on the employees table created
-- in 000003.
-- =========================================================================

-- ---------------- roles ----------------
CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL UNIQUE,
    description     TEXT NOT NULL DEFAULT '',
    is_system       BOOLEAN NOT NULL DEFAULT FALSE,
    permissions     JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL
);
CREATE INDEX idx_roles_is_deleted ON roles (is_deleted);
CREATE INDEX idx_roles_name ON roles (name);
CREATE TRIGGER trg_roles_set_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- users (AUTH ONLY) ----------------
-- No full_name, no department_id, no position_id, no role string.
-- HR / profile fields live on the employees table (000003).
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               CITEXT NOT NULL UNIQUE,
    password_hash       TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    email_changed_at    TIMESTAMPTZ NULL,
    password_reset_at   TIMESTAMPTZ NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted          BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at          TIMESTAMPTZ NULL
);
CREATE INDEX idx_users_is_deleted ON users (is_deleted);
CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- user_roles ----------------
CREATE TABLE user_roles (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ NULL,
    PRIMARY KEY (user_id, role_id)
);
CREATE INDEX idx_user_roles_is_deleted ON user_roles (is_deleted);
CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);
CREATE TRIGGER trg_user_roles_set_updated_at
    BEFORE UPDATE ON user_roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
