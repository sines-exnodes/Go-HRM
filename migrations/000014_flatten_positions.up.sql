-- =========================================================================
-- 000014_flatten_positions
-- Drop the positions ↔ departments link to match Python's flat model.
--
-- Before: positions(department_id UUID NOT NULL REFERENCES departments(id)
--         ON DELETE RESTRICT), with uq_positions_name_dept_active making
--         (department_id, LOWER(name)) unique per active row.
--
-- After: positions stand alone as a global catalog. Position names are
-- unique across the whole table (case-insensitive) among non-deleted rows.
-- Employees keep their separate department_id and position_id columns;
-- the two are no longer linked through positions.
--
-- BREAKING CHANGE: any API client that read `position.department_id` or
-- `position.department` will get those fields removed from responses.
-- Existing rows lose their department_id value (column dropped).
-- The down migration restores the column shape but cannot reconstruct the
-- data — manual repair required if rolling back.
-- =========================================================================

-- The FK and partial unique index reference the column we're about to
-- drop. DROP COLUMN ... CASCADE would also drop them, but we spell them
-- out so the intent is reviewable and the down migration can recreate
-- them in the right order.
ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS fk_positions_department;

DROP INDEX IF EXISTS uq_positions_name_dept_active;
DROP INDEX IF EXISTS idx_positions_department_id;

ALTER TABLE positions
    DROP COLUMN IF EXISTS department_id;

-- Flat replacement: position names are globally unique among non-deleted
-- rows (same case-insensitive semantics as departments.uq_departments_name_active).
CREATE UNIQUE INDEX uq_positions_name_active
    ON positions (LOWER(name)) WHERE is_deleted = FALSE;
