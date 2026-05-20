-- =========================================================================
-- 000007_create_labels
-- Announcement labels — minimal entity (name only), case-insensitive
-- unique. Endpoints are list + get-or-create only (no update/delete),
-- gated by PermAnnounceManage. Mirrors the Python collection
-- `announcement_labels`.
-- =========================================================================

CREATE TABLE labels (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX uq_labels_name_lower_live
    ON labels (LOWER(name)) WHERE is_deleted = FALSE;
CREATE INDEX idx_labels_is_deleted ON labels (is_deleted);
CREATE TRIGGER trg_labels_set_updated_at
    BEFORE UPDATE ON labels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
