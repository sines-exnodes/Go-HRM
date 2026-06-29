ALTER TABLE employees
    DROP COLUMN IF EXISTS social_insurance_number,
    DROP COLUMN IF EXISTS tax_identification_number;
