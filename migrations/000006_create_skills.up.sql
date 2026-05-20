-- =========================================================================
-- 000006_create_skills
-- skills catalog + employee_skills join (employee_id ⟂ skill_id).
-- The join lives on employees (not users) because Phase 1 split auth
-- (users) from HR profile (employees); the Python source's User.skill_ids
-- maps to the HR profile, not the auth identity.
-- =========================================================================

-- ---------------- skills ----------------
CREATE TABLE skills (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    icon_url    TEXT        NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX uq_skills_name_lower_live
    ON skills (LOWER(name)) WHERE is_deleted = FALSE;
CREATE INDEX idx_skills_is_deleted ON skills (is_deleted);
CREATE INDEX idx_skills_name       ON skills (name);
CREATE TRIGGER trg_skills_set_updated_at
    BEFORE UPDATE ON skills
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------- employee_skills (join) ----------------
-- employee_id cascades: deleting an employee removes their skill links.
-- skill_id does NOT cascade: skill deletion is blocked at the service
-- layer with HTTP 409 when any live link exists (mirrors the Phase 3
-- department/position delete guard).
CREATE TABLE employee_skills (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID        NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    skill_id    UUID        NOT NULL REFERENCES skills(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX uq_employee_skills_pair_live
    ON employee_skills (employee_id, skill_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_employee_skills_employee_id ON employee_skills (employee_id);
CREATE INDEX idx_employee_skills_skill_id    ON employee_skills (skill_id);
CREATE INDEX idx_employee_skills_is_deleted  ON employee_skills (is_deleted);
CREATE TRIGGER trg_employee_skills_set_updated_at
    BEFORE UPDATE ON employee_skills
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
