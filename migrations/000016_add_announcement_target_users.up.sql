-- =========================================================================
-- 000016_add_announcement_target_users
-- Phase 7.5: hybrid targeting — per-user (custom) audience alongside the
-- existing per-department audience. Closes the announcements parity audit
-- decision #6 (the deferred parity gap from migration 000010 REVISION
-- NOTES #4).
--
-- Adds:
--   * announcement_target_users — join row {announcement_id, employee_id}
--     for target_audience='custom'. FK targets employees(id), NOT users(id),
--     per the Go schema split (every cross-aggregate FK from Phase 2 onward
--     targets the HR profile).
--   * Updates the announcements.target_audience CHECK constraint to allow
--     the new 'custom' value. Existing rows ('all', 'department') pass
--     unchanged; the down migration reverts the constraint and refuses to
--     proceed if any 'custom' rows exist (RAISE EXCEPTION inside DO block —
--     a destructive migration would otherwise silently lose the audience
--     mapping).
-- =========================================================================

CREATE TABLE announcement_target_users (
    announcement_id UUID         NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    employee_id     UUID         NOT NULL REFERENCES employees(id)     ON DELETE CASCADE,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL,

    PRIMARY KEY (announcement_id, employee_id)
);

CREATE INDEX idx_announcement_target_users_employee_id
    ON announcement_target_users(employee_id);

CREATE TRIGGER trg_announcement_target_users_set_updated_at
    BEFORE UPDATE ON announcement_target_users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Update the audience CHECK constraint to admit 'custom'.
ALTER TABLE announcements DROP CONSTRAINT announcements_target_audience_check;
ALTER TABLE announcements
    ADD CONSTRAINT announcements_target_audience_check
    CHECK (target_audience IN ('all','department','custom'));
