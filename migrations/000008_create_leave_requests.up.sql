-- =========================================================================
-- 000008_create_leave_requests
-- Phase 5: Leave requests + quota lookup.
--
-- One row per leave request. Enum-string columns for leave_type,
-- leave_period and status (Python source uses StrEnum, not a typed table).
-- Per-employee quotas already live in employee_leave_quotas (migration
-- 000004); this table does NOT duplicate them. The "balance" endpoint
-- sums approved total_days per (employee_id, leave_type, year) and
-- subtracts from the row in employee_leave_quotas.
--
-- FKs target employees(id) — the HR profile — NOT users(id). The Go
-- schema split (Phase 1) puts every HR-domain FK on employees from
-- Phase 2 onward, mirroring the existing cross-aggregate pattern in
-- dependents, employee_leave_quotas and employee_skills.
--
-- Warnings (insufficient quota, overlapping dates) are NON-blocking;
-- the service layer attaches them to the response envelope rather than
-- preventing the row from being created. CHECK constraints here only
-- guard physical invariants (enum domain, total_days >= 0,
-- to_date >= from_date).
-- =========================================================================

CREATE TABLE leave_requests (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID         NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    from_date       DATE         NOT NULL,
    to_date         DATE         NOT NULL,
    leave_period    TEXT         NOT NULL DEFAULT 'full_day'
                                 CHECK (leave_period IN ('full_day','morning_half','afternoon_half')),
    leave_type      TEXT         NOT NULL
                                 CHECK (leave_type IN ('annual','sick','personal','maternity','unpaid')),
    total_days      NUMERIC(5,1) NOT NULL CHECK (total_days >= 0),
    reason          TEXT         NOT NULL,
    attachment_url  TEXT         NULL,
    status          TEXT         NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending','approved','rejected','cancelled')),
    created_by      UUID         NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,

    -- Audit columns (every entity table)
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL,

    CONSTRAINT leave_requests_date_range_chk CHECK (to_date >= from_date)
);

CREATE INDEX idx_leave_requests_employee_id     ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status          ON leave_requests(status);
CREATE INDEX idx_leave_requests_from_to         ON leave_requests(from_date, to_date);
CREATE INDEX idx_leave_requests_is_deleted      ON leave_requests(is_deleted);
CREATE INDEX idx_leave_requests_created_at_desc ON leave_requests(created_at DESC);

CREATE TRIGGER trg_leave_requests_set_updated_at
    BEFORE UPDATE ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
