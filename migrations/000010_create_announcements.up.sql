-- =========================================================================
-- 000010_create_announcements
-- Phase 7: Announcements + Mobile Announcements + SSE realtime push.
--
-- Five tables:
--   * announcements                          — the publishable entity
--   * announcement_labels                    — explicit join: announcement ↔ label (labels seeded in P4)
--   * announcement_target_departments        — when target_audience='department'
--   * announcement_attachments               — multipart-uploaded files (image + PDF)
--   * announcement_views                     — per-user read marker (auth-level)
--
-- FK target convention:
--   * announcements.author_id → employees(id)  (Go schema split — Phase 2+).
--     Mirrors leave_requests.created_by + attendance.employee_id. ON DELETE
--     RESTRICT so author attribution survives a hard-delete.
--   * announcement_views.user_id → users(id) — views are auth-level (the
--     read marker is keyed on the logged-in session, not the HR profile).
--   * announcement_labels.label_id → labels(id) (Phase 4).
--   * announcement_target_departments.department_id → departments(id) (Phase 3).
--   * Every child table ON DELETE CASCADE from announcements — hard-delete
--     of an announcement removes its labels/targets/attachments/views.
--
-- target_audience enum is 'all' | 'department'. The Python source also has
-- 'custom' (per-user targeting), but no `announcement_target_users` table
-- exists in this migration — Phase 7.5 adds it if BA confirms the need
-- (REVISION NOTES item #4).
-- =========================================================================

CREATE TABLE announcements (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT         NOT NULL,
    body            TEXT         NOT NULL,
    summary         TEXT         NULL,
    author_id       UUID         NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    status          TEXT         NOT NULL DEFAULT 'draft'
                                 CHECK (status IN ('draft','scheduled','published','archived')),
    scheduled_at    TIMESTAMPTZ  NULL,
    published_at    TIMESTAMPTZ  NULL,
    target_audience TEXT         NOT NULL DEFAULT 'all'
                                 CHECK (target_audience IN ('all','department')),
    pinned          BOOLEAN      NOT NULL DEFAULT FALSE,
    cover_image_url TEXT         NULL,

    -- Audit columns
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL
);

CREATE INDEX idx_announcements_status        ON announcements(status);
CREATE INDEX idx_announcements_published_at  ON announcements(published_at DESC);
CREATE INDEX idx_announcements_author_id     ON announcements(author_id);
CREATE INDEX idx_announcements_pinned        ON announcements(pinned) WHERE pinned = TRUE;
CREATE INDEX idx_announcements_is_deleted    ON announcements(is_deleted);

CREATE TRIGGER trg_announcements_set_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---- announcement_labels (explicit join — audit cols required, see
--      REVISION NOTES #10. Mirrors employee_skills from Phase 4.) ----

CREATE TABLE announcement_labels (
    announcement_id UUID         NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    label_id        UUID         NOT NULL REFERENCES labels(id)        ON DELETE CASCADE,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL,

    PRIMARY KEY (announcement_id, label_id)
);

CREATE INDEX idx_announcement_labels_label_id ON announcement_labels(label_id);

CREATE TRIGGER trg_announcement_labels_set_updated_at
    BEFORE UPDATE ON announcement_labels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---- announcement_target_departments ----

CREATE TABLE announcement_target_departments (
    announcement_id UUID         NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    department_id   UUID         NOT NULL REFERENCES departments(id)   ON DELETE CASCADE,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL,

    PRIMARY KEY (announcement_id, department_id)
);

CREATE INDEX idx_announcement_target_depts_dept_id
    ON announcement_target_departments(department_id);

CREATE TRIGGER trg_announcement_target_departments_set_updated_at
    BEFORE UPDATE ON announcement_target_departments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---- announcement_attachments (one row per uploaded file) ----

CREATE TABLE announcement_attachments (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID         NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    url             TEXT         NOT NULL,
    filename        TEXT         NOT NULL,
    content_type    TEXT         NOT NULL,
    size_bytes      BIGINT       NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL
);

CREATE INDEX idx_announcement_attachments_announcement_id
    ON announcement_attachments(announcement_id);
CREATE INDEX idx_announcement_attachments_is_deleted
    ON announcement_attachments(is_deleted);

CREATE TRIGGER trg_announcement_attachments_set_updated_at
    BEFORE UPDATE ON announcement_attachments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---- announcement_views (per-user read marker — auth-level) ----

CREATE TABLE announcement_views (
    announcement_id UUID         NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id         UUID         NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    viewed_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL,

    PRIMARY KEY (announcement_id, user_id)
);

CREATE INDEX idx_announcement_views_user_id ON announcement_views(user_id);

CREATE TRIGGER trg_announcement_views_set_updated_at
    BEFORE UPDATE ON announcement_views
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
