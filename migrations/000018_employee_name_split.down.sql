-- Down for 000018_employee_name_split. Lossless: full_name = first || ' ' || last.
ALTER TABLE employees ADD COLUMN full_name TEXT NOT NULL DEFAULT '';

UPDATE employees SET full_name = btrim(first_name || ' ' || last_name);

ALTER TABLE employees DROP COLUMN first_name;
ALTER TABLE employees DROP COLUMN last_name;
