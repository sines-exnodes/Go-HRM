-- =========================================================================
-- 000003_create_employees_dependents
-- employees (1-1 with users), dependents (1-N with employees).
--
-- NOTE: department_id and position_id on employees are NULLABLE and
-- intentionally have NO FK CONSTRAINT here. The departments / positions
-- tables are introduced in Phase 3; the FK constraints are added in that
-- phase (ALTER TABLE employees ADD CONSTRAINT ...).
-- =========================================================================

-- ---------------- employees ----------------
CREATE TABLE employees (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                     UUID NOT NULL UNIQUE REFERENCES users(id),

    -- Personal info
    full_name                   TEXT NOT NULL,
    phone                       TEXT NULL,
    personal_email              CITEXT NULL,
    gender                      TEXT NULL,                 -- male / female / other
    permanent_address           TEXT NULL,
    current_address             TEXT NULL,
    dob                         DATE NULL,
    nationality                 TEXT NULL,
    id_number                   TEXT NULL,
    id_issue_date               DATE NULL,
    id_front_image              TEXT NULL,
    id_back_image               TEXT NULL,
    avatar_url                  TEXT NULL,
    education                   TEXT NULL,                 -- high_school / college / university / master
    marital_status              TEXT NULL,                 -- single / married / other

    -- Emergency contact
    emergency_contact_name      TEXT NULL,
    emergency_contact_relation  TEXT NULL,
    emergency_contact_phone     TEXT NULL,

    -- Work info
    -- department_id / position_id: NO FK yet — added in Phase 3.
    department_id               UUID NULL,
    position_id                 UUID NULL,
    manager_id                  UUID NULL REFERENCES employees(id),
    join_date                   DATE NULL,
    contract_type               TEXT NOT NULL DEFAULT 'official',  -- probation / official
    contract_sign_date          DATE NULL,
    contract_end_date           DATE NULL,
    contract_renewal            INT  NOT NULL DEFAULT 1,

    -- Salary & insurance
    basic_salary                NUMERIC(18,2) NOT NULL DEFAULT 0,
    insurance_salary            NUMERIC(18,2) NOT NULL DEFAULT 0,

    -- Banking
    bank_account                TEXT NULL,
    bank_name                   TEXT NULL,
    bank_holder_name            TEXT NULL,
    payment_method              TEXT NOT NULL DEFAULT 'bank_transfer',  -- bank_transfer / cash

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted                  BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at                  TIMESTAMPTZ NULL
);
CREATE INDEX idx_employees_is_deleted     ON employees (is_deleted);
CREATE INDEX idx_employees_department_id  ON employees (department_id);
CREATE INDEX idx_employees_position_id    ON employees (position_id);
CREATE INDEX idx_employees_manager_id     ON employees (manager_id);
-- (user_id already has a UNIQUE index by virtue of UNIQUE on the column)
CREATE TRIGGER trg_employees_set_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- dependents ----------------
CREATE TABLE dependents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    full_name       TEXT NOT NULL,
    dob             DATE NULL,
    gender          TEXT NULL,           -- male / female / other
    relationship    TEXT NOT NULL,       -- child / parent / spouse / other
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL
);
CREATE INDEX idx_dependents_is_deleted  ON dependents (is_deleted);
CREATE INDEX idx_dependents_employee_id ON dependents (employee_id);
CREATE TRIGGER trg_dependents_set_updated_at
    BEFORE UPDATE ON dependents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
