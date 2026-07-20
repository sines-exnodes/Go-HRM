-- migrations/000028_notifications.up.sql
--
-- Mobile in-app notification feed (DR-MOB-005-001-01). Fan-out on write:
-- one row per (recipient, event), carrying a snapshot of the title and body
-- taken at creation time.
--
-- The snapshot is deliberate (DR Rule 12) — editing an announcement after
-- publish must NOT rewrite what the employee was already told. It is also
-- why source_id carries no foreign key: the row must survive its source
-- being deleted (DR Rule 13), and source_id points into two different
-- tables depending on `type`.
--
-- user_id targets users(id), not employees(id). Notifications are an
-- auth-level surface consumed by the logged-in session, the same exception
-- already made for announcement_views and device_tokens.

CREATE TABLE notifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id),
    type       TEXT        NOT NULL,
    title      TEXT        NOT NULL,
    body       TEXT        NOT NULL DEFAULT '',
    source_id  UUID        NOT NULL,
    read_at    TIMESTAMPTZ,
    is_deleted BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- List: WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX idx_notifications_user_created
    ON notifications (user_id, created_at DESC)
    WHERE is_deleted = FALSE;

-- Unread count for the dashboard header bell.
CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id)
    WHERE read_at IS NULL AND is_deleted = FALSE;

-- DR Rule 5 — one notification per event. Producers insert with
-- ON CONFLICT DO NOTHING, so re-publishing an announcement or retrying an
-- approve silently no-ops instead of duplicating. Removing this index
-- silently breaks Rule 5.
CREATE UNIQUE INDEX uq_notifications_user_source
    ON notifications (user_id, type, source_id)
    WHERE is_deleted = FALSE;

CREATE TRIGGER trg_notifications_set_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
