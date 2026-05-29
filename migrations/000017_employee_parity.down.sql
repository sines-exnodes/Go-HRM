-- Down migration for 000017_employee_parity.
--
-- The single flat columns can hold at most ONE emergency contact per
-- employee. Refuse to proceed if any employee now has more than one (the
-- revert would silently drop the extras) — operator must reduce to one each
-- first. Same data-loss guard pattern as 000016.

DO $$
DECLARE
    n INT;
BEGIN
    SELECT COUNT(*) INTO n FROM (
        SELECT employee_id
        FROM employee_emergency_contacts
        WHERE is_deleted = FALSE
        GROUP BY employee_id
        HAVING COUNT(*) > 1
    ) x;
    IF n > 0 THEN
        RAISE EXCEPTION
            'Refusing to revert: % employee(s) have multiple emergency contacts; the single-column schema cannot hold them. Reduce to one each before downgrading.', n;
    END IF;
END$$;

-- Recreate the flat columns.
ALTER TABLE employees ADD COLUMN emergency_contact_name     TEXT NULL;
ALTER TABLE employees ADD COLUMN emergency_contact_relation TEXT NULL;
ALTER TABLE employees ADD COLUMN emergency_contact_phone    TEXT NULL;

-- Copy back the single (guaranteed-unique) contact per employee.
UPDATE employees e SET
    emergency_contact_name     = c.full_name,
    emergency_contact_relation = NULLIF(c.relationship, ''),
    emergency_contact_phone    = NULLIF(c.phone_number, '')
FROM employee_emergency_contacts c
WHERE c.employee_id = e.id AND c.is_deleted = FALSE;

DROP TRIGGER IF EXISTS trg_employee_emergency_contacts_set_updated_at ON employee_emergency_contacts;
DROP TABLE IF EXISTS employee_emergency_contacts;

ALTER TABLE employees DROP COLUMN IF EXISTS cv_url;
ALTER TABLE employees DROP COLUMN IF EXISTS experience_year;
