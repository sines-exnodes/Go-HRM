-- Reverse of 000001_init_extensions.
-- Extensions are dropped only if no later object depends on them; in
-- practice this down step is intended for local resets, not production.
DROP FUNCTION IF EXISTS set_updated_at();

DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS "uuid-ossp";
