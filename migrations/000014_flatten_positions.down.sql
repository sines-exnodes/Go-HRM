-- =========================================================================
-- Down: restore the positions.department_id column + FK + partial index.
--
-- Data loss: the original department_id values were dropped on the up
-- migration. The column is recreated as NULLABLE here so the rollback
-- succeeds against rows that no longer carry that info. If you need the
-- old NOT NULL invariant back, populate department_id first (e.g. via a
-- join through employees.position_id → employees.department_id majority
-- vote) and then ALTER COLUMN ... SET NOT NULL by hand.
-- =========================================================================

DROP INDEX IF EXISTS uq_positions_name_active;

ALTER TABLE positions
    ADD COLUMN department_id UUID NULL;

CREATE INDEX idx_positions_department_id ON positions (department_id);

CREATE UNIQUE INDEX uq_positions_name_dept_active
    ON positions (department_id, LOWER(name)) WHERE is_deleted = FALSE;

ALTER TABLE positions
    ADD CONSTRAINT fk_positions_department
        FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE RESTRICT;
