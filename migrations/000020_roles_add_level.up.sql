-- =========================================================================
-- 000020_roles_add_level
-- Adds the role authority `level` (1..100). Python parity: an assigner may
-- only grant roles at or below their own max level. Backfills the known
-- system roles by name; any other existing row keeps the DEFAULT 100.
-- =========================================================================
ALTER TABLE roles
    ADD COLUMN level INT NOT NULL DEFAULT 100
        CHECK (level BETWEEN 1 AND 100);

UPDATE roles SET level = 100 WHERE name = 'Super Admin';
UPDATE roles SET level = 90  WHERE name = 'Admin';
UPDATE roles SET level = 80  WHERE name = 'HR Manager';
UPDATE roles SET level = 50  WHERE name = 'Manager';
UPDATE roles SET level = 10  WHERE name = 'Employee';
