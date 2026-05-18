ALTER TABLE employees DROP CONSTRAINT IF EXISTS fk_employees_position;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS fk_employees_department;

DROP TRIGGER IF EXISTS trg_positions_set_updated_at ON positions;
DROP TABLE IF EXISTS positions;

DROP TRIGGER IF EXISTS trg_departments_set_updated_at ON departments;
DROP TABLE IF EXISTS departments;
