-- =========================================================================
-- 000013_add_users_lockout
-- Brute-force protection for /auth/login. Mirrors the Python repo's lockout
-- flow: increment failed_login_attempts on each bad password, set
-- locked_until once the threshold is hit, reset both on a successful login.
-- =========================================================================

ALTER TABLE users
    ADD COLUMN failed_login_attempts INT         NOT NULL DEFAULT 0,
    ADD COLUMN locked_until          TIMESTAMPTZ NULL;
