-- =========================================================================
-- 000011_create_system_config
-- Phase 8 — singleton organization-wide configuration row.
--
-- Mirrors the Python Beanie `SystemConfig` singleton (Document, not KV).
-- Fields are statically declared:
--   * Attendance late/checkout thresholds (consumed by the Phase 6
--     attendance service via `system_config` lookup — but until Phase 6
--     is refactored, env vars LATE_THRESHOLD_* still take precedence).
--   * Company address + lat/lng for the mobile map preview screen.
--
-- The table is constrained to EXACTLY one row whose id is the sentinel
-- UUID `00000000-0000-0000-0000-000000000001`. The seed service performs
-- an idempotent INSERT ... ON CONFLICT DO NOTHING on boot. Any second
-- row would fail the singleton CHECK constraint at the DB level.
--
-- Audit columns are present per spec §5.2 but soft-delete is unused —
-- the row is updated in place and is_deleted stays false.
--
-- FK: company_address_updated_by → employees(id) ON DELETE SET NULL,
-- per the Go schema split (REVISION NOTES #2). Audit-trail column;
-- survives an employee delete.
-- =========================================================================

CREATE TABLE system_config (
    id  UUID PRIMARY KEY DEFAULT '00000000-0000-0000-0000-000000000001',

    -- Attendance: late-arrival threshold
    late_threshold_hour       SMALLINT NOT NULL DEFAULT 9
        CHECK (late_threshold_hour BETWEEN 0 AND 23),
    late_threshold_minute     SMALLINT NOT NULL DEFAULT 0
        CHECK (late_threshold_minute BETWEEN 0 AND 59),

    -- Attendance: checkout threshold (early-leave calc, reserved)
    checkout_threshold_hour   SMALLINT NOT NULL DEFAULT 18
        CHECK (checkout_threshold_hour BETWEEN 0 AND 23),
    checkout_threshold_minute SMALLINT NOT NULL DEFAULT 0
        CHECK (checkout_threshold_minute BETWEEN 0 AND 59),

    -- Company profile
    company_address                 TEXT,
    company_latitude                DOUBLE PRECISION,
    company_longitude               DOUBLE PRECISION,
    company_address_updated_at      TIMESTAMPTZ,
    company_address_updated_by      UUID REFERENCES employees(id) ON DELETE SET NULL,

    -- Audit cols (per spec §5.2). Soft delete is unused for the singleton.
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ,

    -- Enforce singleton at the DB level. Any second INSERT fails this CHECK.
    CONSTRAINT system_config_singleton CHECK (id = '00000000-0000-0000-0000-000000000001')
);

CREATE INDEX idx_system_config_is_deleted ON system_config(is_deleted);

CREATE TRIGGER trg_system_config_set_updated_at
    BEFORE UPDATE ON system_config
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
