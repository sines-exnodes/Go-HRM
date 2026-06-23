-- migrations/000024_password_reset_tokens.up.sql

CREATE TABLE password_reset_tokens (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id),
    token            TEXT        NOT NULL,
    expires_at       TIMESTAMPTZ NOT NULL,
    used_at          TIMESTAMPTZ,
    last_email_error TEXT,
    is_deleted       BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_password_reset_tokens_token
    ON password_reset_tokens (token)
    WHERE is_deleted = FALSE;

CREATE INDEX idx_password_reset_tokens_user_id
    ON password_reset_tokens (user_id)
    WHERE is_deleted = FALSE;

CREATE TRIGGER trg_password_reset_tokens_set_updated_at
    BEFORE UPDATE ON password_reset_tokens
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
