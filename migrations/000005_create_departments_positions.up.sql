-- =========================================================================
-- 000005_create_departments_positions
-- departments (self-referential tree), positions (belong to a department).
-- Also adds the deferred FK constraints on employees.department_id /
-- employees.position_id (created NULLABLE + index, NO FK, in 000003).
-- =========================================================================

-- ---------------- departments ----------------
CREATE TABLE departments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    parent_id   UUID        NULL REFERENCES departments(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX uq_departments_name_active
    ON departments (LOWER(name)) WHERE is_deleted = FALSE;
CREATE INDEX idx_departments_is_deleted ON departments (is_deleted);
CREATE INDEX idx_departments_parent_id  ON departments (parent_id) WHERE parent_id IS NOT NULL;
CREATE TRIGGER trg_departments_set_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- positions ----------------
CREATE TABLE positions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    department_id UUID        NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted    BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at    TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX uq_positions_name_dept_active
    ON positions (department_id, LOWER(name)) WHERE is_deleted = FALSE;
CREATE INDEX idx_positions_is_deleted    ON positions (is_deleted);
CREATE INDEX idx_positions_department_id ON positions (department_id);
CREATE TRIGGER trg_positions_set_updated_at
    BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- deferred FK constraints on employees ----------------
-- employees.department_id / position_id were created NULLABLE + indexed but
-- WITHOUT FK in 000003 (deferred to this phase). Add them now.
ALTER TABLE employees
    ADD CONSTRAINT fk_employees_department
        FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

ALTER TABLE employees
    ADD CONSTRAINT fk_employees_position
        FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL;
