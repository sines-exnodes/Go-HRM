-- =========================================================================
-- 000009_create_attendance
-- Phase 6: Attendance (check-in / check-out / sessions / monthly matrix).
--
-- Two-table design — one row per (employee_id, date) in `attendance`, with
-- N child sessions in `attendance_sessions`. A "session" is a check-in/
-- check-out pair; multiple sessions per day are supported (e.g. lunch
-- back). `is_late` is computed once from the FIRST check-in vs the
-- configured threshold and stored on the day row — subsequent sessions
-- do NOT re-evaluate it (REVISION NOTES item #5).
--
-- FK target = employees(id) — the HR profile — NOT users(id). The Go
-- schema split puts every HR-domain FK on employees from Phase 2 onward,
-- mirroring leave_requests / dependents / employee_leave_quotas /
-- employee_skills. ON DELETE RESTRICT (mirror of leave_requests) keeps
-- audit-traceability — CASCADE would silently erase attendance history
-- when an employee row is hard-deleted.
--
-- Sessions cascade-delete from their parent attendance row, because a
-- hard-deleted attendance row's sessions are unreachable anyway. Soft
-- delete uses the custom is_deleted/deleted_at columns (NOT GORM's
-- built-in DeletedAt); reads scope through NotDeleted in the repo.
--
-- Constraints:
--   - UNIQUE (employee_id, date) on `attendance` — at most one row per
--     employee per day; service-level FindByEmployeeAndDate defends to
--     surface clean 409s instead of 500-on-insert.
--   - CHECK on attendance.work_location restricts the enum domain.
--   - CHECK on attendance_sessions enforces check_out >= check_in.
--   - Partial UNIQUE on attendance_sessions ensures at most one OPEN
--     session per attendance row at any time (REVISION NOTES item #6).
-- =========================================================================

CREATE TABLE attendance (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID         NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    date            DATE         NOT NULL,
    is_late         BOOLEAN      NOT NULL DEFAULT FALSE,
    is_half_day     BOOLEAN      NOT NULL DEFAULT FALSE,
    work_location   TEXT         NULL
                                 CHECK (work_location IS NULL OR work_location IN ('office','remote','hybrid','field')),
    notes           TEXT         NULL,

    -- Audit columns (every entity table)
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL,

    CONSTRAINT attendance_employee_date_unique UNIQUE (employee_id, date)
);

CREATE INDEX idx_attendance_employee_id    ON attendance(employee_id);
CREATE INDEX idx_attendance_date           ON attendance(date);
CREATE INDEX idx_attendance_employee_date  ON attendance(employee_id, date DESC);
CREATE INDEX idx_attendance_is_deleted     ON attendance(is_deleted);

CREATE TRIGGER trg_attendance_set_updated_at
    BEFORE UPDATE ON attendance
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE attendance_sessions (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id     UUID         NOT NULL REFERENCES attendance(id) ON DELETE CASCADE,
    check_in          TIMESTAMPTZ  NOT NULL,
    check_out         TIMESTAMPTZ  NULL,
    is_auto_checkout  BOOLEAN      NOT NULL DEFAULT FALSE,

    -- Audit columns
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted        BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at        TIMESTAMPTZ  NULL,

    CONSTRAINT attendance_sessions_checkout_after_checkin
        CHECK (check_out IS NULL OR check_out >= check_in)
);

CREATE INDEX idx_attendance_sessions_attendance_id ON attendance_sessions(attendance_id);
CREATE INDEX idx_attendance_sessions_check_in      ON attendance_sessions(check_in);
CREATE INDEX idx_attendance_sessions_is_deleted    ON attendance_sessions(is_deleted);

-- At most one OPEN (check_out IS NULL) live session per attendance row.
-- Partial unique index — soft-deleted rows are excluded so they don't
-- block a fresh check-in.
CREATE UNIQUE INDEX uq_attendance_sessions_one_open
    ON attendance_sessions(attendance_id)
    WHERE check_out IS NULL AND is_deleted = FALSE;

CREATE TRIGGER trg_attendance_sessions_set_updated_at
    BEFORE UPDATE ON attendance_sessions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
