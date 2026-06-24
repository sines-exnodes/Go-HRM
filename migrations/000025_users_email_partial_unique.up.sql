-- Replace the hard unique constraint on users.email with a partial unique index
-- so that soft-deleted rows (is_deleted = true) do not block re-use of an email.
ALTER TABLE users DROP CONSTRAINT users_email_key;
CREATE UNIQUE INDEX users_email_key ON users (email) WHERE is_deleted = FALSE;
