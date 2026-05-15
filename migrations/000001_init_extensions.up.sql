-- =============================================================
-- 000001_init_extensions
-- Foundational Postgres extensions + shared updated_at trigger.
-- All later migrations may assume these objects exist.
-- =============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- Shared trigger function used by every entity table's BEFORE UPDATE
-- trigger. Keeps updated_at in sync without per-row application code.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
