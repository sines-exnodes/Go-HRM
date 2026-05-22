-- =========================================================================
-- 000012_create_invites
-- Phase 9 — admin-generated email invitations.
--
-- The invite is a first-class row that defers user creation until the
-- invitee posts to /api/v1/invites/accept. Diverges from the Python
-- source (which set a temporary password on a pre-created user) — the
-- BA intent is "invite an email; the user row is created on accept".
--
-- FK targets (Go schema split, Phase 2+):
--   * invited_by       → employees(id) ON DELETE RESTRICT — audit-trail,
--     mirrors leave_requests.created_by + announcements.author_id +
--     system_config.company_address_updated_by.
--   * accepted_user_id → users(id) ON DELETE SET NULL — auth-level marker
--     populated on accept, mirrors announcement_views.user_id.
--   * department_id, position_id → SET NULL — invitees can be assigned
--     a dept/position at invite time; if the dept/position is deleted
--     before accept, the invitee just lands without that linkage.
--
-- role_ids is a UUID[] column (per the Python source); on accept the
-- service iterates the array and calls UserRepository.ReplaceRoles().
--
-- Partial unique indexes:
--   * (token) WHERE is_deleted = FALSE — token must be globally unique
--     among live invites. Soft-deleted (revoked) invites release their
--     token so re-invites won't collide.
--   * (email) WHERE accepted_at IS NULL AND is_deleted = FALSE — at most
--     one pending invite per email at any time. Resending generates the
--     same row (no rotation); revoking + re-creating produces a new row
--     with a fresh token.
-- =========================================================================

CREATE TABLE invites (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email            CITEXT       NOT NULL,
    full_name        TEXT         NULL,
    token            TEXT         NOT NULL,
    role_ids         UUID[]       NOT NULL DEFAULT '{}',
    department_id    UUID         NULL REFERENCES departments(id) ON DELETE SET NULL,
    position_id      UUID         NULL REFERENCES positions(id)   ON DELETE SET NULL,
    expires_at       TIMESTAMPTZ  NOT NULL,
    accepted_at      TIMESTAMPTZ  NULL,
    accepted_user_id UUID         NULL REFERENCES users(id)     ON DELETE SET NULL,
    invited_by       UUID         NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    last_email_error TEXT         NULL,

    -- Audit columns (per spec §5.2)
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted       BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at       TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX uq_invites_token_live
    ON invites(token) WHERE is_deleted = FALSE;

CREATE UNIQUE INDEX uq_invites_pending_email
    ON invites(email) WHERE accepted_at IS NULL AND is_deleted = FALSE;

CREATE INDEX idx_invites_email      ON invites(email);
CREATE INDEX idx_invites_is_deleted ON invites(is_deleted);
CREATE INDEX idx_invites_expires_at ON invites(expires_at);
CREATE INDEX idx_invites_invited_by ON invites(invited_by);

CREATE TRIGGER trg_invites_set_updated_at
    BEFORE UPDATE ON invites
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
