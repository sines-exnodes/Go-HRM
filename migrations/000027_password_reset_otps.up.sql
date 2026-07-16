-- migrations/000027_password_reset_otps.up.sql
--
-- Mobile forgot-password flow (DR-001-001-02). The web flow mails a one-click
-- link backed by password_reset_tokens (000024); mobile mails a 6-digit code
-- the user types into the app. The code is bcrypt-hashed — rows are looked up
-- by user_id, never by code.
--
-- Rows are NOT hard-deleted on supersede: the rate-limit window counts every
-- row created for a user regardless of is_deleted, so soft-deleting a
-- superseded code must not reset the counter.

CREATE TABLE password_reset_otps (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id),
    code_hash        TEXT        NOT NULL,
    expires_at       TIMESTAMPTZ NOT NULL,
    consumed_at      TIMESTAMPTZ,
    attempt_count    INTEGER     NOT NULL DEFAULT 0,
    last_email_error TEXT,
    is_deleted       BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Active-code lookup: latest unconsumed row for a user.
CREATE INDEX idx_password_reset_otps_user_active
    ON password_reset_otps (user_id, created_at DESC)
    WHERE is_deleted = FALSE;

-- Rate-limit window scan: counts every row in the window, deleted or not.
CREATE INDEX idx_password_reset_otps_user_created
    ON password_reset_otps (user_id, created_at DESC);

CREATE TRIGGER trg_password_reset_otps_set_updated_at
    BEFORE UPDATE ON password_reset_otps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
