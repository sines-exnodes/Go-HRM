-- =========================================================================
-- 000004_phase2_extras
-- device_tokens (push), user_notification_settings (toggle), employee_leave_quotas
-- All entity tables carry the standard 4 audit cols and use UUID PKs.
-- =========================================================================

-- ---------------- device_tokens ----------------
CREATE TABLE device_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       TEXT        NOT NULL,
    device_id   TEXT        NOT NULL,
    platform    TEXT        NOT NULL DEFAULT 'unknown',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ NULL,
    UNIQUE (user_id, device_id)
);
CREATE INDEX idx_device_tokens_user_id    ON device_tokens(user_id);
CREATE INDEX idx_device_tokens_is_deleted ON device_tokens(is_deleted);
CREATE TRIGGER trg_device_tokens_set_updated_at
    BEFORE UPDATE ON device_tokens
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- user_notification_settings ----------------
CREATE TABLE user_notification_settings (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    notifications_enabled BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted            BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at            TIMESTAMPTZ NULL
);
CREATE INDEX idx_user_notification_settings_is_deleted ON user_notification_settings(is_deleted);
CREATE TRIGGER trg_user_notification_settings_set_updated_at
    BEFORE UPDATE ON user_notification_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- employee_leave_quotas ----------------
CREATE TABLE employee_leave_quotas (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id        UUID         NOT NULL UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
    annual_leave_quota NUMERIC(6,2) NOT NULL DEFAULT 12.00,
    sick_leave_quota   NUMERIC(6,2) NOT NULL DEFAULT 6.00,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted         BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at         TIMESTAMPTZ  NULL
);
CREATE INDEX idx_employee_leave_quotas_is_deleted ON employee_leave_quotas(is_deleted);
CREATE TRIGGER trg_employee_leave_quotas_set_updated_at
    BEFORE UPDATE ON employee_leave_quotas
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
