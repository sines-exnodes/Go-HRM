-- =========================================================================
-- 000017_employee_parity
-- Employees Python-shape parity (audit decisions #4, #8):
--
--   #8  Add the two profile fields Python returns that the Go split dropped:
--         * employees.experience_year INT  (years of experience)
--         * employees.cv_url          TEXT (CV file URL; server-side upload
--                                            endpoint deferred per audit #11)
--
--   #4  Emergency contacts become a LIST (Python models them as a list of
--       sub-documents). Replace the single flat trio
--         (emergency_contact_name / _relation / _phone)
--       with a child table employee_emergency_contacts (1-N), mirroring the
--       dependents table shape. The existing single contact is migrated into
--       the new table; the flat columns are then dropped.
--
-- FK targets employees(id) per the Go schema split. Audit columns + the
-- BEFORE UPDATE trigger match every other entity table.
-- =========================================================================

-- ---- #8: new scalar profile fields ----
ALTER TABLE employees ADD COLUMN experience_year INT  NULL;
ALTER TABLE employees ADD COLUMN cv_url          TEXT NULL;

-- ---- #4: employee_emergency_contacts child table ----
CREATE TABLE employee_emergency_contacts (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   UUID        NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    full_name     TEXT        NOT NULL,
    relationship  TEXT        NOT NULL DEFAULT '',
    phone_number  TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted    BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at    TIMESTAMPTZ NULL
);
CREATE INDEX idx_employee_emergency_contacts_is_deleted  ON employee_emergency_contacts(is_deleted);
CREATE INDEX idx_employee_emergency_contacts_employee_id ON employee_emergency_contacts(employee_id);
CREATE TRIGGER trg_employee_emergency_contacts_set_updated_at
    BEFORE UPDATE ON employee_emergency_contacts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---- migrate the existing single contact into the new table ----
-- Only rows with a meaningful name carry over; a row with just a stray phone
-- and no name is treated as no contact (avoids junk empty-name rows).
INSERT INTO employee_emergency_contacts (employee_id, full_name, relationship, phone_number)
SELECT id,
       emergency_contact_name,
       COALESCE(emergency_contact_relation, ''),
       COALESCE(emergency_contact_phone, '')
FROM employees
WHERE emergency_contact_name IS NOT NULL
  AND btrim(emergency_contact_name) <> '';

-- ---- drop the flat columns now that data is migrated ----
ALTER TABLE employees DROP COLUMN emergency_contact_name;
ALTER TABLE employees DROP COLUMN emergency_contact_relation;
ALTER TABLE employees DROP COLUMN emergency_contact_phone;
