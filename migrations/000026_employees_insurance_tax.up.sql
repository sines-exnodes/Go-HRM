-- Add social_insurance_number and tax_identification_number to employees.
-- Both are optional free-text fields (max 50 chars, no format validation per BA DR-001-005-02 v1.4).

ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS social_insurance_number TEXT,
    ADD COLUMN IF NOT EXISTS tax_identification_number TEXT;
