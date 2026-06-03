-- 000018_employee_name_split
-- Python parity: names are stored as separate first_name + last_name columns
-- (app/models/user.py:88-89). Drop the single full_name column and backfill
-- the split. Legacy single-token names land as first_name=token, last_name=''
-- (the DEFAULT '' tolerates them; the min=1 rule applies only to new writes).
ALTER TABLE employees ADD COLUMN first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE employees ADD COLUMN last_name  TEXT NOT NULL DEFAULT '';

UPDATE employees
SET first_name = split_part(full_name, ' ', 1),
    last_name  = btrim(substr(full_name, length(split_part(full_name, ' ', 1)) + 1));

ALTER TABLE employees DROP COLUMN full_name;
