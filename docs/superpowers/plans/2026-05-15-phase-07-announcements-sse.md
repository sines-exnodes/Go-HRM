# Phase 7 — Announcements + Mobile Announcements + SSE Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## ⚠️ REVISION NOTES (2026-05-20) — AUTHORITATIVE, read & apply before executing any task

This plan was drafted pre-Phase-5/6 (before the Go schema split and before
migrations 000008/000009 landed). The codebase audit at the close of
Phase 6 supersedes the task bodies below wherever they conflict.
**Execute per these notes, not the raw task bodies where they conflict.**

1. **Migration number.** `000001`–`000009` are taken (latest = `000009_create_attendance` from Phase 6). Phase 7 migration is **`000010_create_announcements`** (NOT `000009`). Final `make migrate-version` after this phase = **10**. Rename every filename, `make migrate-*` reference, and DoD "version 9" / "migration 000009" expectation accordingly.

2. **`announcements.author_id` FK target = `employees(id)`**, NOT `users(id)`. Mirrors `leave_requests.employee_id`/`created_by` (Phase 5) and `attendance.employee_id` (Phase 6). Use `ON DELETE RESTRICT` so audit trails survive a hard-delete of the employee. Author display reads `Employee.FullName` (the column the FE actually renders), not `User.Email`. The seeded super admin already has an employee row (Phase 1 seed), so this never blocks admin authoring.

3. **`announcement_views.user_id` FK target = `users(id)`** (correct as drafted). Views are auth-level — "this logged-in session marked this announcement as read." A user without an employee profile shouldn't even be able to log in, but the read tracker is keyed on the auth identity per the Python source. `ON DELETE CASCADE` is fine (delete a user → its view rows go too).

4. **`target_audience` enum scope reduction.** The draft CHECK says `('all','department','custom')`. There is **no `announcement_target_users` table** in the migration — `custom` would have no backing. For Phase 7, **drop `'custom'`** from the enum (`CHECK (target_audience IN ('all','department'))`). If/when BA confirms custom targeting is needed, Phase 7.5 adds the table.

5. **Handler helpers already exist — reuse, don't redeclare.** [`internal/handlers/employee_handler.go`](../../internal/handlers/employee_handler.go) line 38 defines `currentUser(c *gin.Context) (*models.User, bool)` and line 52 defines `parseIDParam(c *gin.Context, key string) (uuid.UUID, error)`. The plan's Task 13 redeclares `currentUser(c) *models.User` (different signature) and adds a new `parseUUID(c, name)` — both would collide at compile. **Use the existing helpers verbatim.**

6. **`apperrs.Write(c, err)` does NOT exist.** Our error model uses `_ = c.Error(err); return` and the `ErrorHandler` middleware renders the envelope (see `internal/middleware/error.go` + `internal/errors/errors.go`). Replace every `apperrs.Write(c, err)` in Task 13 with the `_ = c.Error(err); return` pattern (see `leave_handler.go` / `attendance_handler.go` for examples).

7. **`middleware.RequirePerms` takes `authSvc *AuthService` as first arg.** Existing signature: `RequirePerms(authSvc *services.AuthService, required ...permissions.Permission)`. Task 14 omits `authSvc` — fix every call site.

8. **JWT middleware variant — extract `parseAndLoadUser`.** Existing `middleware.JWT(users repositories.UserRepository, jwtSecret string)` carries the parse+load logic inline. The plan's `ParseAndLoadUser` does not exist yet. **Refactor:** extract the existing JWT's parse+load body into a package-private `parseAndLoadUser(ctx, token, jwtSecret string, users repositories.UserRepository) (*models.User, error)` and call it from BOTH `JWT()` and the new `JWTFromQueryOrHeader(users, jwtSecret string)`. Use `string` for the secret (matches existing), not `[]byte`. Don't break existing callers.

9. **Context keys are `auth_user` / `auth_user_id` / `auth_claims`** (constants `middleware.ContextKeyUser/UserID/Claims` in [`internal/middleware/auth.go`](../../internal/middleware/auth.go)). The plan's `c.Get("current_user")` / `c.Set("current_user", user)` is **wrong** — use the existing constants. Both middlewares set the same keys so `currentUser(c)` helper works for both.

10. **Many-to-many with audit columns ⇒ explicit join model.** GORM's `gorm:"many2many:..."` tag implicitly creates a join row WITHOUT the four audit columns. The migration's `announcement_labels` table has `created_at/updated_at/is_deleted/deleted_at` — we must declare an explicit `AnnouncementLabel` Go model and write composition manually (mirrors Phase 4's `EmployeeSkill` pattern). Don't use the `gorm:"many2many:..."` tag on `Announcement.Labels`.

11. **Repo joins MUST qualify `is_deleted`.** Phase 6 fix surfaced this: `models.NotDeleted` adds an unqualified `WHERE is_deleted = ?` which becomes ambiguous after a JOIN to any table that also carries `is_deleted` (announcement_labels, labels, announcement_target_departments, etc.). In every announcement repo method that joins, replace `r.base(ctx)` / `Scopes(models.NotDeleted)` with `r.db.WithContext(ctx).Where("announcements.is_deleted = ?", false)`.

12. **`make swag` output path = `docs/swagger`** (NOT `docs/`). Phase 0 configured it. Task 15 paths must use `docs/swagger/swagger.json` / `docs/swagger/docs.go` / `docs/swagger/swagger.yaml`.

13. **`truncateAll` test helper order**: announcement child tables must precede employees + labels + departments. Insertion point in [`internal/services/testhelper_test.go`](../../internal/services/testhelper_test.go):

    ```text
    announcement_views, announcement_attachments, announcement_target_departments,
    announcement_labels, announcements,
    ... (existing entries)
    employees, ... labels ..., departments, ...
    ```

14. **Visibility predicate (List + Get).** A non-admin caller can see an announcement when:
    - `status = 'published'` AND `is_deleted = false`, AND
    - One of: `author_id = current_employee.id` OR `target_audience = 'all'` OR (`target_audience = 'department'` AND `announcement_target_departments.department_id = current_employee.department_id`)
    OR the caller holds `PermAnnounceManage` (Admin / HR Manager) → sees everything regardless of status/target.

15. **Service signature** follows Phase 5/6: `(ctx context.Context, currentUserID uuid.UUID, asAdmin bool, ...)`. The handler precomputes `asAdmin = hasAnnounceManageAll(c)` from `user.Roles` (same shape as `hasLeaveManageAll`/`hasAttendanceManageAll`). The service internally calls `resolveCurrentEmployee()` to get the employee row when visibility filtering needs department.

16. **`sseHubAdapter`** lives in `cmd/server/sse_adapter.go` (small, separate file). The service depends on a thin interface (`HubBroadcaster`) so it stays mock-able in tests.

17. **`PermAnnounceManage` seed coverage**: already complete (Super Admin via `*`, Admin + HR Manager hold the permission directly per Phase 4 seed fix). Manager + Employee don't have it — that's correct (they can only read). REVISION NOTES from Phase 4 documented this; no seed change needed in Phase 7. Verify live anyway (Phase 5 + 6 verification surfaced load-bearing gaps the REVISION NOTES had denied).

18. **`announcement_attachments` table** kept as separate table per BA expectation (multiple files per announcement). Multipart upload uses the same `http.DetectContentType` + MIME allowlist pattern as Phase 5 leave_attachment / Phase 4 skill icon / Phase 2 avatar. Allowlist: image/* + application/pdf. Each attachment row is independently soft-deletable.

19. **DoD = real live verification** committed to `docs/superpowers/verification/phase-07.md`. Minimum flow:
    - Boot server, `migrate-version=10`.
    - Login admin → create label → create draft announcement (200, status="draft") → publish → SSE consumer receives `announcement_published` event within ~1s.
    - Login non-admin (Employee role) → list announcements → only published, target-applicable rows visible.
    - Mark-viewed endpoint → `announcement_views` row exists; second call is idempotent.
    - Update / Delete / unauthorized non-admin write → 403.
    - SSE without token → 401.
    - psql spot-check: row counts + soft-delete state.

Everything else in the task bodies (TDD-first, commit-per-task, no placeholders) still applies. **Execute per these REVISION NOTES, not the raw task bodies where they conflict.**

---

**Goal:** Port Python `announcements.py` + `mobile_announcements.py` to Go (Postgres-backed) and add an in-memory SSE hub so admins can publish announcements that fan out to connected web/mobile clients in real-time.

**Architecture:** Postgres tables under `announcements` (plus join tables for labels / target departments, plus a read-tracking + attachments table). GORM models embed `BaseModel` for soft-delete. Three-layer flow: handler -> service -> repo. Service handles visibility logic (published + targeted/all OR author OR `PermAnnounceManage`) and broadcasts to the in-memory `sse.Hub` on publish. SSE handler is a single Gin route streaming `text/event-stream` with a 30s keep-alive ticker and per-client buffered channel. JWT is accepted via `?token=` query param (EventSource cannot set headers) — documented limitation.

**Tech Stack:** Go 1.24, Gin, GORM, Postgres, `golang-migrate`, `golang-jwt/jwt/v5`, `swaggo/swag`, `testify`. In-memory single-instance hub (single replica scaling limit — documented in code).

---

## File Structure

| File | Responsibility |
|---|---|
| `migrations/000009_create_announcements.up.sql` / `.down.sql` | Tables: `announcements`, `announcement_labels`, `announcement_target_departments`, `announcement_attachments`, `announcement_views` |
| `internal/models/announcement.go` | `Announcement`, `AnnouncementAttachment`, `AnnouncementView`, `AnnouncementTargetDepartment` GORM structs |
| `internal/dto/announcement.go` | Request/response DTOs for web + mobile |
| `internal/repositories/announcement_repo.go` | `AnnouncementRepository` interface + GORM implementation |
| `internal/sse/hub.go` | `Hub`, `Event`, `Client`, `Subscribe`, `Broadcast` (in-memory, goroutine-safe) |
| `internal/sse/hub_test.go` | Unit tests for hub register/unregister/broadcast/filter |
| `internal/services/announcement_service.go` | Business logic + visibility checks + hub broadcast |
| `internal/services/announcement_service_test.go` | Service tests with mock hub + test Postgres |
| `internal/handlers/announcement_handler.go` | Web + mobile announcement routes |
| `internal/handlers/sse_handler.go` | `GET /api/v1/sse/announcements` SSE stream |
| `internal/permissions/registry.go` (existing) | Add nothing — `PermAnnounceManage` exists from Phase 0 spec |
| `internal/middleware/jwt.go` (existing) | Add `JWTFromQueryOrHeader` variant for SSE |
| `cmd/server/main.go` (existing) | Wire singleton `sse.Hub`, inject into services + handlers |
| `docs/superpowers/verification/phase-07.md` | E2E verification log (created at the end) |

---

## Task 1: Create migration files for announcements schema

**Files:**
- Create: `migrations/000009_create_announcements.up.sql`
- Create: `migrations/000009_create_announcements.down.sql`

- [ ] **Step 1: Write `000009_create_announcements.up.sql`**

```sql
-- Phase 7: announcements + mobile announcements + SSE
-- Depends on: 000002 (users), 000003 (departments), 000008 (labels) — adjust if Phase 4 used a different file number.

CREATE TABLE announcements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT NOT NULL,
    body            TEXT NOT NULL,
    summary         TEXT NULL,
    author_id       UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','scheduled','published','archived')),
    scheduled_at    TIMESTAMPTZ NULL,
    published_at    TIMESTAMPTZ NULL,
    target_audience TEXT NOT NULL DEFAULT 'all'
                    CHECK (target_audience IN ('all','department','custom')),
    pinned          BOOLEAN NOT NULL DEFAULT FALSE,
    cover_image_url TEXT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL
);
CREATE INDEX idx_announcements_status        ON announcements (status);
CREATE INDEX idx_announcements_published_at  ON announcements (published_at DESC);
CREATE INDEX idx_announcements_author_id     ON announcements (author_id);
CREATE INDEX idx_announcements_is_deleted    ON announcements (is_deleted);
CREATE TRIGGER trg_announcements_updated_at
    BEFORE UPDATE ON announcements
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE announcement_labels (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    label_id        UUID NOT NULL REFERENCES labels(id)        ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL,
    PRIMARY KEY (announcement_id, label_id)
);
CREATE INDEX idx_announcement_labels_label_id ON announcement_labels (label_id);
CREATE TRIGGER trg_announcement_labels_updated_at
    BEFORE UPDATE ON announcement_labels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE announcement_target_departments (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    department_id   UUID NOT NULL REFERENCES departments(id)   ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL,
    PRIMARY KEY (announcement_id, department_id)
);
CREATE INDEX idx_announcement_target_depts_dept_id ON announcement_target_departments (department_id);
CREATE TRIGGER trg_announcement_target_departments_updated_at
    BEFORE UPDATE ON announcement_target_departments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE announcement_attachments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    url             TEXT NOT NULL,
    filename        TEXT NOT NULL,
    content_type    TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL
);
CREATE INDEX idx_announcement_attachments_ann_id ON announcement_attachments (announcement_id);
CREATE TRIGGER trg_announcement_attachments_updated_at
    BEFORE UPDATE ON announcement_attachments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE announcement_views (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    viewed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL,
    PRIMARY KEY (announcement_id, user_id)
);
CREATE INDEX idx_announcement_views_user_id ON announcement_views (user_id);
CREATE TRIGGER trg_announcement_views_updated_at
    BEFORE UPDATE ON announcement_views
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 2: Write `000009_create_announcements.down.sql`**

```sql
DROP TABLE IF EXISTS announcement_views;
DROP TABLE IF EXISTS announcement_attachments;
DROP TABLE IF EXISTS announcement_target_departments;
DROP TABLE IF EXISTS announcement_labels;
DROP TABLE IF EXISTS announcements;
```

- [ ] **Step 3: Apply the migration on a clean DB**

Run: `make migrate-up`
Expected output (last line): a non-error exit; running `make migrate-version` should print `9` (or whatever the current latest is, must include `9`).

- [ ] **Step 4: Roll back and re-apply to confirm `.down.sql` works**

Run: `make migrate-down && make migrate-up`
Expected: both succeed; `make migrate-version` returns `9`.

- [ ] **Step 5: Commit**

```bash
git add migrations/000009_create_announcements.up.sql migrations/000009_create_announcements.down.sql
git commit -m "feat(phase-07): migration 000009 announcements + labels + targets + views"
```

---

## Task 2: GORM models for announcement entities

**Files:**
- Create: `internal/models/announcement.go`

- [ ] **Step 1: Write the file**

```go
package models

import (
	"time"

	"github.com/google/uuid"
)

// AnnouncementStatus is the lifecycle status of an announcement.
type AnnouncementStatus string

const (
	AnnouncementStatusDraft     AnnouncementStatus = "draft"
	AnnouncementStatusScheduled AnnouncementStatus = "scheduled"
	AnnouncementStatusPublished AnnouncementStatus = "published"
	AnnouncementStatusArchived  AnnouncementStatus = "archived"
)

// AnnouncementTargetAudience controls how recipients are resolved.
type AnnouncementTargetAudience string

const (
	AnnouncementTargetAll        AnnouncementTargetAudience = "all"
	AnnouncementTargetDepartment AnnouncementTargetAudience = "department"
	AnnouncementTargetCustom     AnnouncementTargetAudience = "custom"
)

// Announcement is the core entity.
type Announcement struct {
	BaseModel
	Title           string                     `gorm:"not null"               json:"title"`
	Body            string                     `gorm:"not null"               json:"body"`
	Summary         *string                    `                              json:"summary,omitempty"`
	AuthorID        uuid.UUID                  `gorm:"type:uuid;not null"     json:"author_id"`
	Author          *User                      `gorm:"foreignKey:AuthorID"    json:"author,omitempty"`
	Status          AnnouncementStatus         `gorm:"not null;default:draft" json:"status"`
	ScheduledAt     *time.Time                 `                              json:"scheduled_at,omitempty"`
	PublishedAt     *time.Time                 `                              json:"published_at,omitempty"`
	TargetAudience  AnnouncementTargetAudience `gorm:"not null;default:all"   json:"target_audience"`
	Pinned          bool                       `gorm:"not null;default:false" json:"pinned"`
	CoverImageURL   *string                    `                              json:"cover_image_url,omitempty"`

	Labels             []Label                       `gorm:"many2many:announcement_labels;joinForeignKey:announcement_id;joinReferences:label_id" json:"labels,omitempty"`
	TargetDepartments  []AnnouncementTargetDepartment `gorm:"foreignKey:AnnouncementID"                                                            json:"target_departments,omitempty"`
	Attachments        []AnnouncementAttachment       `gorm:"foreignKey:AnnouncementID"                                                            json:"attachments,omitempty"`
}

func (Announcement) TableName() string { return "announcements" }

// AnnouncementTargetDepartment is the join row when TargetAudience == "department".
type AnnouncementTargetDepartment struct {
	AnnouncementID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"announcement_id"`
	DepartmentID   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"department_id"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	IsDeleted      bool       `gorm:"not null;default:false" json:"-"`
	DeletedAt      *time.Time `                                json:"-"`
}

func (AnnouncementTargetDepartment) TableName() string { return "announcement_target_departments" }

// AnnouncementAttachment is a file linked to an announcement.
type AnnouncementAttachment struct {
	BaseModel
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null" json:"announcement_id"`
	URL            string    `gorm:"not null"           json:"url"`
	Filename       string    `gorm:"not null"           json:"filename"`
	ContentType    string    `gorm:"not null"           json:"content_type"`
	SizeBytes      int64     `gorm:"not null;default:0" json:"size_bytes"`
}

func (AnnouncementAttachment) TableName() string { return "announcement_attachments" }

// AnnouncementView records that a user viewed/read an announcement.
type AnnouncementView struct {
	AnnouncementID uuid.UUID  `gorm:"type:uuid;primaryKey"   json:"announcement_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey"   json:"user_id"`
	ViewedAt       time.Time  `gorm:"not null;default:now()" json:"viewed_at"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	IsDeleted      bool       `gorm:"not null;default:false" json:"-"`
	DeletedAt      *time.Time `                                json:"-"`
}

func (AnnouncementView) TableName() string { return "announcement_views" }
```

- [ ] **Step 2: Build to confirm models compile**

Run: `go build ./internal/models/...`
Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/models/announcement.go
git commit -m "feat(phase-07): announcement, attachment, view, target-dept models"
```

---

## Task 3: DTOs for announcements + mobile + SSE events

**Files:**
- Create: `internal/dto/announcement.go`

- [ ] **Step 1: Write DTOs**

```go
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/models"
)

// ----- Common embedded refs -----

type AuthorRef struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
}

type LabelRef struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color *string   `json:"color,omitempty"`
}

type DepartmentRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AttachmentRead struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
}

type AttachmentCreate struct {
	URL         string `json:"url"          binding:"required,url"`
	Filename    string `json:"filename"     binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	SizeBytes   int64  `json:"size_bytes"`
}

// ----- Write DTOs -----

type AnnouncementCreate struct {
	Title               string                            `json:"title"                 binding:"required,min=1,max=200"`
	Body                string                            `json:"body"                  binding:"required,min=1"`
	Summary             *string                           `json:"summary,omitempty"`
	Status              *models.AnnouncementStatus        `json:"status,omitempty"      binding:"omitempty,oneof=draft scheduled published"`
	ScheduledAt         *time.Time                        `json:"scheduled_at,omitempty"`
	TargetAudience      models.AnnouncementTargetAudience `json:"target_audience"       binding:"required,oneof=all department custom"`
	TargetDepartmentIDs []uuid.UUID                       `json:"target_department_ids,omitempty"`
	LabelIDs            []uuid.UUID                       `json:"label_ids,omitempty"`
	Pinned              bool                              `json:"pinned,omitempty"`
	CoverImageURL       *string                           `json:"cover_image_url,omitempty"`
	Attachments         []AttachmentCreate                `json:"attachments,omitempty"`
}

type AnnouncementUpdate struct {
	Title               *string                            `json:"title,omitempty"       binding:"omitempty,min=1,max=200"`
	Body                *string                            `json:"body,omitempty"        binding:"omitempty,min=1"`
	Summary             *string                            `json:"summary,omitempty"`
	Status              *models.AnnouncementStatus         `json:"status,omitempty"      binding:"omitempty,oneof=draft scheduled published archived"`
	ScheduledAt         *time.Time                         `json:"scheduled_at,omitempty"`
	TargetAudience      *models.AnnouncementTargetAudience `json:"target_audience,omitempty" binding:"omitempty,oneof=all department custom"`
	TargetDepartmentIDs *[]uuid.UUID                       `json:"target_department_ids,omitempty"`
	LabelIDs            *[]uuid.UUID                       `json:"label_ids,omitempty"`
	Pinned              *bool                              `json:"pinned,omitempty"`
	CoverImageURL       *string                            `json:"cover_image_url,omitempty"`
	Attachments         *[]AttachmentCreate                `json:"attachments,omitempty"`
}

type AnnouncementPublish struct {
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// ----- Read DTOs -----

type AnnouncementRead struct {
	ID                uuid.UUID                         `json:"id"`
	Title             string                            `json:"title"`
	Body              string                            `json:"body"`
	Summary           *string                           `json:"summary,omitempty"`
	Author            AuthorRef                         `json:"author"`
	Status            models.AnnouncementStatus         `json:"status"`
	ScheduledAt       *time.Time                        `json:"scheduled_at,omitempty"`
	PublishedAt       *time.Time                        `json:"published_at,omitempty"`
	TargetAudience    models.AnnouncementTargetAudience `json:"target_audience"`
	TargetDepartments []DepartmentRef                   `json:"target_departments,omitempty"`
	Labels            []LabelRef                        `json:"labels"`
	Pinned            bool                              `json:"pinned"`
	CoverImageURL     *string                           `json:"cover_image_url,omitempty"`
	Attachments       []AttachmentRead                  `json:"attachments,omitempty"`
	ViewCount         int64                             `json:"view_count"`
	ViewedByMe        bool                              `json:"viewed_by_me"`
	CreatedAt         time.Time                         `json:"created_at"`
	UpdatedAt         time.Time                         `json:"updated_at"`
}

// MobileAnnouncementRead is the trimmed shape returned to mobile clients.
// Body is kept full (per BA), but expensive fields (target_departments, attachments)
// are omitted from the list endpoint.
type MobileAnnouncementRead struct {
	ID          uuid.UUID                 `json:"id"`
	Title       string                    `json:"title"`
	Summary     *string                   `json:"summary,omitempty"`
	Body        string                    `json:"body,omitempty"`
	PublishedAt *time.Time                `json:"published_at,omitempty"`
	Pinned      bool                      `json:"pinned"`
	IsRead      bool                      `json:"is_read"`
	Labels      []LabelRef                `json:"labels"`
	CoverImage  *string                   `json:"cover_image,omitempty"`
	Status      models.AnnouncementStatus `json:"status"`
}

// ----- Queries -----

type AnnouncementScope string

const (
	AnnouncementScopeAll        AnnouncementScope = "all"
	AnnouncementScopeMine       AnnouncementScope = "mine"
	AnnouncementScopeTargetedMe AnnouncementScope = "targeted-at-me"
)

type AnnouncementListQuery struct {
	Page     int                        `form:"page,default=1"      binding:"min=1"`
	PageSize int                        `form:"page_size,default=20" binding:"min=1,max=100"`
	Search   string                     `form:"search"`
	Status   *models.AnnouncementStatus `form:"status"   binding:"omitempty,oneof=draft scheduled published archived"`
	LabelID  *uuid.UUID                 `form:"label_id"`
	Pinned   *bool                      `form:"pinned"`
	Scope    AnnouncementScope          `form:"scope,default=all" binding:"omitempty,oneof=all mine targeted-at-me"`
}

type MobileAnnouncementListQuery struct {
	Page     int    `form:"page,default=1"        binding:"min=1"`
	PageSize int    `form:"page_size,default=20"  binding:"min=1,max=50"`
	Search   string `form:"search"`
}

// ----- SSE event payload -----

type AnnouncementEvent struct {
	ID             uuid.UUID                         `json:"id"`
	Title          string                            `json:"title"`
	Summary        *string                           `json:"summary,omitempty"`
	PublishedAt    *time.Time                        `json:"published_at,omitempty"`
	TargetAudience models.AnnouncementTargetAudience `json:"target_audience"`
	AuthorID       uuid.UUID                         `json:"author_id"`
}
```

- [ ] **Step 2: Build**

Run: `go build ./internal/dto/...`
Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/dto/announcement.go
git commit -m "feat(phase-07): announcement DTOs for web, mobile, SSE event"
```

---

## Task 4: SSE Hub primitives (test first)

**Files:**
- Create: `internal/sse/hub_test.go`
- Create: `internal/sse/hub.go`

- [ ] **Step 1: Write the failing hub test**

```go
package sse

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHub_SubscribeAndBroadcastAll(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	uid := uuid.New()
	ch, unsubscribe := hub.Subscribe(uid)
	defer unsubscribe()

	hub.Broadcast(Event{Type: "announcement_published", Data: map[string]any{"x": 1}}, nil)

	select {
	case b := <-ch:
		var ev Event
		if err := json.Unmarshal(b, &ev); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}
		if ev.Type != "announcement_published" {
			t.Fatalf("want type announcement_published, got %s", ev.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("did not receive broadcast in 500ms")
	}
}

func TestHub_BroadcastWithFilter(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	u1 := uuid.New()
	u2 := uuid.New()
	ch1, un1 := hub.Subscribe(u1)
	defer un1()
	ch2, un2 := hub.Subscribe(u2)
	defer un2()

	// Only u1 should receive.
	filter := func(userID uuid.UUID) bool { return userID == u1 }
	hub.Broadcast(Event{Type: "announcement_published", Data: map[string]any{}}, filter)

	gotU1 := false
	gotU2 := false
	deadline := time.After(300 * time.Millisecond)

	for {
		select {
		case <-ch1:
			gotU1 = true
		case <-ch2:
			gotU2 = true
		case <-deadline:
			if !gotU1 {
				t.Fatal("u1 did not receive filtered broadcast")
			}
			if gotU2 {
				t.Fatal("u2 unexpectedly received filtered broadcast")
			}
			return
		}
	}
}

func TestHub_UnsubscribeStopsDelivery(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	uid := uuid.New()
	ch, unsubscribe := hub.Subscribe(uid)
	unsubscribe()

	hub.Broadcast(Event{Type: "x"}, nil)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("received event after unsubscribe")
		}
		// closed channel — fine.
	case <-time.After(150 * time.Millisecond):
		// also fine — channel never delivers.
	}
}

func TestHub_ConcurrentSubscribers(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	chs := make([]<-chan []byte, n)
	uns := make([]func(), n)
	for i := 0; i < n; i++ {
		ch, un := hub.Subscribe(uuid.New())
		chs[i] = ch
		uns[i] = un
		go func(idx int) {
			defer wg.Done()
			select {
			case <-chs[idx]:
			case <-time.After(500 * time.Millisecond):
				t.Errorf("client %d timed out", idx)
			}
		}(i)
	}
	hub.Broadcast(Event{Type: "ping"}, nil)
	wg.Wait()
	for _, un := range uns {
		un()
	}
}
```

- [ ] **Step 2: Run the test to confirm it fails**

Run: `go test ./internal/sse/... -run TestHub -v`
Expected: FAIL with `undefined: NewHub` (or similar — package not yet implemented).

- [ ] **Step 3: Implement `internal/sse/hub.go`**

```go
// Package sse provides an in-memory, single-instance pub/sub hub used to
// stream announcement (and future notification) events to connected clients
// over Server-Sent Events.
//
// SCALING LIMIT: This hub holds all subscriber channels in memory inside a
// single process. Horizontal scaling beyond one replica requires a shared
// backplane (Redis pub/sub, NATS, etc.) — out of scope for Phase 7.
package sse

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Event is the payload broadcast to subscribers.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// FilterFunc decides whether a given userID should receive an Event.
// nil filter means "deliver to all".
type FilterFunc func(userID uuid.UUID) bool

// client is one connected SSE listener.
type client struct {
	id     string // unique per connection (handles same-user multi-tab)
	userID uuid.UUID
	send   chan []byte
}

// Hub is goroutine-safe.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*client
	closed  bool
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[string]*client)}
}

// Subscribe registers a new listener for userID. The returned channel
// receives marshaled `data: ...\n\n` SSE payloads. The unsubscribe
// function MUST be called to free the client; it is safe to call multiple
// times.
func (h *Hub) Subscribe(userID uuid.UUID) (<-chan []byte, func()) {
	c := &client{
		id:     uuid.NewString(),
		userID: userID,
		send:   make(chan []byte, 16),
	}
	h.mu.Lock()
	h.clients[c.id] = c
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if _, ok := h.clients[c.id]; ok {
				delete(h.clients, c.id)
				close(c.send)
			}
		})
	}
	return c.send, unsubscribe
}

// Broadcast sends an event to every subscriber for whom filter(userID)
// returns true. A nil filter delivers to all subscribers. Slow consumers
// whose buffers are full are skipped (events are dropped per client, not
// per hub) so one stuck client cannot block the publisher.
func (h *Hub) Broadcast(event Event, filter FilterFunc) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	for _, c := range h.clients {
		if filter != nil && !filter(c.userID) {
			continue
		}
		select {
		case c.send <- payload:
		default:
			// buffer full — drop for this client
		}
	}
}

// ClientCount returns the current number of subscribers.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Stop closes all client channels and marks the hub as closed. Intended for
// tests and graceful shutdown.
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, c := range h.clients {
		close(c.send)
		delete(h.clients, id)
	}
}
```

- [ ] **Step 4: Re-run the test**

Run: `go test ./internal/sse/... -run TestHub -v`
Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/sse/hub.go internal/sse/hub_test.go
git commit -m "feat(phase-07): in-memory SSE hub with filtered broadcast + tests"
```

---

## Task 5: Announcement repository — interface + GORM implementation

**Files:**
- Create: `internal/repositories/announcement_repo.go`

- [ ] **Step 1: Write the file**

```go
package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

// AnnouncementRepository is the data-access contract used by the service
// layer. Defined as an interface so tests can substitute fakes for the
// service unit tests; the production GORM implementation lives below.
type AnnouncementRepository interface {
	Create(ctx context.Context, a *models.Announcement) error
	Update(ctx context.Context, a *models.Announcement) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, opts FindOptions) (*models.Announcement, error)
	List(ctx context.Context, q ListAnnouncementQuery) ([]models.Announcement, int64, error)
	Publish(ctx context.Context, id uuid.UUID, at time.Time) error
	MarkViewed(ctx context.Context, announcementID, userID uuid.UUID) error
	CountViews(ctx context.Context, announcementID uuid.UUID) (int64, error)
	ViewedByMe(ctx context.Context, announcementID, userID uuid.UUID) (bool, error)
	SetLabels(ctx context.Context, announcementID uuid.UUID, labelIDs []uuid.UUID) error
	SetTargetDepartments(ctx context.Context, announcementID uuid.UUID, departmentIDs []uuid.UUID) error
	ReplaceAttachments(ctx context.Context, announcementID uuid.UUID, atts []models.AnnouncementAttachment) error
	UserDepartmentIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// FindOptions toggles eager-loading.
type FindOptions struct {
	WithLabels            bool
	WithAttachments       bool
	WithAuthor            bool
	WithTargetDepartments bool
}

// ListAnnouncementQuery is the repo-level shape of a list request after the
// service has computed visibility scoping.
type ListAnnouncementQuery struct {
	Page                 int
	PageSize             int
	Search               string
	Status               *models.AnnouncementStatus
	LabelID              *uuid.UUID
	Pinned               *bool
	// Visibility (set by service):
	OnlyAuthorID         *uuid.UUID // when set, restrict to author = OnlyAuthorID
	OnlyPublished        bool       // restrict to status='published'
	OnlyVisibleToUserID  *uuid.UUID // visibility filter: all OR custom-member OR dept-member
}

type announcementRepo struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return &announcementRepo{db: db}
}

func notDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

func (r *announcementRepo) Create(ctx context.Context, a *models.Announcement) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *announcementRepo) Update(ctx context.Context, a *models.Announcement) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *announcementRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&models.Announcement{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *announcementRepo) FindByID(ctx context.Context, id uuid.UUID, opts FindOptions) (*models.Announcement, error) {
	q := r.db.WithContext(ctx).Scopes(notDeleted)
	if opts.WithAuthor {
		q = q.Preload("Author")
	}
	if opts.WithLabels {
		q = q.Preload("Labels", notDeleted)
	}
	if opts.WithAttachments {
		q = q.Preload("Attachments", notDeleted)
	}
	if opts.WithTargetDepartments {
		q = q.Preload("TargetDepartments", notDeleted)
	}

	var a models.Announcement
	if err := q.First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *announcementRepo) List(ctx context.Context, q ListAnnouncementQuery) ([]models.Announcement, int64, error) {
	tx := r.db.WithContext(ctx).
		Model(&models.Announcement{}).
		Scopes(notDeleted)

	if q.OnlyPublished {
		tx = tx.Where("status = ?", models.AnnouncementStatusPublished)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Search != "" {
		needle := "%" + escapeLike(q.Search) + "%"
		tx = tx.Where("title ILIKE ? ESCAPE '\\'", needle)
	}
	if q.Pinned != nil {
		tx = tx.Where("pinned = ?", *q.Pinned)
	}
	if q.OnlyAuthorID != nil {
		tx = tx.Where("author_id = ?", *q.OnlyAuthorID)
	}
	if q.LabelID != nil {
		tx = tx.Where(`id IN (
			SELECT announcement_id FROM announcement_labels
			WHERE label_id = ? AND is_deleted = false
		)`, *q.LabelID)
	}
	if q.OnlyVisibleToUserID != nil {
		uid := *q.OnlyVisibleToUserID
		tx = tx.Where(`
			target_audience = 'all'
			OR author_id = ?
			OR (target_audience = 'department' AND id IN (
				SELECT atd.announcement_id
				FROM announcement_target_departments atd
				JOIN users u ON u.department_id = atd.department_id
				WHERE u.id = ? AND atd.is_deleted = false
			))
		`, uid, uid)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}

	var items []models.Announcement
	err := tx.
		Preload("Author").
		Preload("Labels", notDeleted).
		Preload("TargetDepartments", notDeleted).
		Order("pinned DESC, COALESCE(published_at, created_at) DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *announcementRepo) Publish(ctx context.Context, id uuid.UUID, at time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&models.Announcement{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]interface{}{
			"status":       models.AnnouncementStatusPublished,
			"published_at": at,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *announcementRepo) MarkViewed(ctx context.Context, announcementID, userID uuid.UUID) error {
	view := &models.AnnouncementView{
		AnnouncementID: announcementID,
		UserID:         userID,
		ViewedAt:       time.Now().UTC(),
	}
	// ON CONFLICT DO NOTHING (composite PK).
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO announcement_views (announcement_id, user_id, viewed_at, created_at, updated_at, is_deleted)
		      VALUES (?, ?, ?, NOW(), NOW(), false)
		      ON CONFLICT (announcement_id, user_id) DO NOTHING`,
			view.AnnouncementID, view.UserID, view.ViewedAt).Error
}

func (r *announcementRepo) CountViews(ctx context.Context, announcementID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.AnnouncementView{}).
		Where("announcement_id = ? AND is_deleted = ?", announcementID, false).
		Count(&n).Error
	return n, err
}

func (r *announcementRepo) ViewedByMe(ctx context.Context, announcementID, userID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.AnnouncementView{}).
		Where("announcement_id = ? AND user_id = ? AND is_deleted = ?", announcementID, userID, false).
		Count(&n).Error
	return n > 0, err
}

func (r *announcementRepo) SetLabels(ctx context.Context, announcementID uuid.UUID, labelIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`DELETE FROM announcement_labels WHERE announcement_id = ?`,
			announcementID,
		).Error; err != nil {
			return err
		}
		for _, lid := range labelIDs {
			if err := tx.Exec(
				`INSERT INTO announcement_labels (announcement_id, label_id, created_at, updated_at, is_deleted)
				 VALUES (?, ?, NOW(), NOW(), false)`,
				announcementID, lid,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *announcementRepo) SetTargetDepartments(ctx context.Context, announcementID uuid.UUID, departmentIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`DELETE FROM announcement_target_departments WHERE announcement_id = ?`,
			announcementID,
		).Error; err != nil {
			return err
		}
		for _, did := range departmentIDs {
			if err := tx.Exec(
				`INSERT INTO announcement_target_departments (announcement_id, department_id, created_at, updated_at, is_deleted)
				 VALUES (?, ?, NOW(), NOW(), false)`,
				announcementID, did,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *announcementRepo) ReplaceAttachments(ctx context.Context, announcementID uuid.UUID, atts []models.AnnouncementAttachment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&models.AnnouncementAttachment{}).
			Where("announcement_id = ? AND is_deleted = ?", announcementID, false).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"deleted_at": &now,
			}).Error; err != nil {
			return err
		}
		for i := range atts {
			atts[i].AnnouncementID = announcementID
			if err := tx.Create(&atts[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *announcementRepo) UserDepartmentIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Users table has a single department_id; return as slice so we can grow
	// later if multi-department membership is added.
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("users").
		Where("id = ? AND is_deleted = ? AND department_id IS NOT NULL", userID, false).
		Pluck("department_id", &ids).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return ids, nil
}

// escapeLike escapes LIKE wildcards. Backslash is the escape char (matches
// the `ESCAPE '\\'` clause used in queries).
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
```

- [ ] **Step 2: Build to confirm compile**

Run: `go build ./internal/repositories/...`
Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/repositories/announcement_repo.go
git commit -m "feat(phase-07): announcement repository + visibility-aware list"
```

---

## Task 6: Announcement service skeleton + Create (TDD)

**Files:**
- Modify: `internal/services/announcement_service.go` (create)
- Modify: `internal/services/announcement_service_test.go` (create)
- Modify: `internal/services/testhelper_test.go` (extend factory helpers if needed)

- [ ] **Step 1: Write the failing service test**

```go
package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

// fakeHub captures broadcasts for assertions.
type fakeHub struct {
	events []capturedEvent
}

type capturedEvent struct {
	Type string
	Data interface{}
	// nil filter recorded as nil (i.e. broadcast-to-all)
	HasFilter bool
}

func (f *fakeHub) Broadcast(eventType string, data interface{}, filter func(uuid.UUID) bool) {
	f.events = append(f.events, capturedEvent{
		Type: eventType, Data: data, HasFilter: filter != nil,
	})
}

func TestAnnouncement_CreateDraft_DoesNotBroadcast(t *testing.T) {
	tb := newTestBox(t)
	hub := &fakeHub{}
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, hub)

	admin := tb.makeAdmin(t)

	req := dto.AnnouncementCreate{
		Title:          "Hello",
		Body:           "<p>Body</p>",
		TargetAudience: models.AnnouncementTargetAll,
	}
	got, err := svc.Create(context.Background(), admin, req)
	require.NoError(t, err)
	assert.Equal(t, models.AnnouncementStatusDraft, got.Status)
	assert.Nil(t, got.PublishedAt)
	assert.Len(t, hub.events, 0, "draft create must NOT broadcast")
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/services/ -run TestAnnouncement_CreateDraft_DoesNotBroadcast -v`
Expected: FAIL — `NewAnnouncementService` undefined / `tb.repos.Announcement` undefined.

- [ ] **Step 3: Add the `Announcement` field to the existing test repos factory**

In `internal/services/testhelper_test.go`, locate the `repoBox` (or equivalent struct that already aggregates repositories from earlier phases — created in Phase 0/1). Add:

```go
// (inside the repoBox struct definition that already groups repos)
Announcement repositories.AnnouncementRepository
```

…and in the helper that constructs it:

```go
Announcement: repositories.NewAnnouncementRepository(db),
```

Run: `grep -n "repoBox" internal/services/testhelper_test.go || true`
If `repoBox` does not exist with that exact name, find the equivalent (e.g. `testRepos`, `Repos`) by inspecting the file once and adapt. The point: the `Announcement` repo must be available on `tb.repos`.

- [ ] **Step 4: Implement the minimum service to make the test pass**

Write `internal/services/announcement_service.go`:

```go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrs "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// HubBroadcaster is the minimum surface the service needs from sse.Hub.
// Defined here so service tests can inject a fake without depending on the
// concrete hub.
type HubBroadcaster interface {
	Broadcast(eventType string, data interface{}, filter func(uuid.UUID) bool)
}

type AnnouncementService struct {
	repo     repositories.AnnouncementRepository
	userRepo repositories.UserRepository
	hub      HubBroadcaster
}

func NewAnnouncementService(
	repo repositories.AnnouncementRepository,
	userRepo repositories.UserRepository,
	hub HubBroadcaster,
) *AnnouncementService {
	return &AnnouncementService{repo: repo, userRepo: userRepo, hub: hub}
}

// Create inserts a new announcement, associating labels / target departments
// / attachments, and broadcasts an SSE event if the new announcement is
// already in 'published' status.
func (s *AnnouncementService) Create(
	ctx context.Context, author *models.User, req dto.AnnouncementCreate,
) (*models.Announcement, error) {
	if author == nil {
		return nil, apperrs.ErrUnauthorized("authenticated user required")
	}
	if err := validateTargetCombination(req.TargetAudience, req.TargetDepartmentIDs); err != nil {
		return nil, err
	}

	status := models.AnnouncementStatusDraft
	if req.Status != nil {
		status = *req.Status
	}
	var publishedAt *time.Time
	if status == models.AnnouncementStatusPublished {
		now := time.Now().UTC()
		publishedAt = &now
	}

	a := &models.Announcement{
		Title:          req.Title,
		Body:           req.Body,
		Summary:        req.Summary,
		AuthorID:       author.ID,
		Status:         status,
		ScheduledAt:    req.ScheduledAt,
		PublishedAt:    publishedAt,
		TargetAudience: req.TargetAudience,
		Pinned:         req.Pinned,
		CoverImageURL:  req.CoverImageURL,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, apperrs.ErrInternal("create announcement: " + err.Error())
	}
	if err := s.repo.SetLabels(ctx, a.ID, req.LabelIDs); err != nil {
		return nil, apperrs.ErrInternal("set labels: " + err.Error())
	}
	if req.TargetAudience == models.AnnouncementTargetDepartment {
		if err := s.repo.SetTargetDepartments(ctx, a.ID, req.TargetDepartmentIDs); err != nil {
			return nil, apperrs.ErrInternal("set target departments: " + err.Error())
		}
	}
	if len(req.Attachments) > 0 {
		atts := toAttachmentModels(req.Attachments)
		if err := s.repo.ReplaceAttachments(ctx, a.ID, atts); err != nil {
			return nil, apperrs.ErrInternal("set attachments: " + err.Error())
		}
	}

	if a.Status == models.AnnouncementStatusPublished {
		s.broadcastPublished(a)
	}

	return s.repo.FindByID(ctx, a.ID, repositories.FindOptions{
		WithAuthor: true, WithLabels: true, WithAttachments: true, WithTargetDepartments: true,
	})
}

func (s *AnnouncementService) broadcastPublished(a *models.Announcement) {
	if s.hub == nil {
		return
	}
	payload := dto.AnnouncementEvent{
		ID:             a.ID,
		Title:          a.Title,
		Summary:        a.Summary,
		PublishedAt:    a.PublishedAt,
		TargetAudience: a.TargetAudience,
		AuthorID:       a.AuthorID,
	}
	// All targeting filter: 'all' broadcasts to everyone. 'department' and
	// 'custom' would ideally filter, but recipient resolution requires DB
	// reads in the hot path — for Phase 7 we broadcast to everyone and let
	// the receiver re-check visibility on the next list call.
	s.hub.Broadcast("announcement_published", payload, nil)
}

func validateTargetCombination(t models.AnnouncementTargetAudience, deptIDs []uuid.UUID) error {
	if t == models.AnnouncementTargetDepartment && len(deptIDs) == 0 {
		return apperrs.ErrBadRequest("target_department_ids required when target_audience='department'")
	}
	return nil
}

func toAttachmentModels(in []dto.AttachmentCreate) []models.AnnouncementAttachment {
	out := make([]models.AnnouncementAttachment, 0, len(in))
	for _, a := range in {
		out = append(out, models.AnnouncementAttachment{
			URL:         a.URL,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			SizeBytes:   a.SizeBytes,
		})
	}
	return out
}

// translateRepoErr maps gorm errors to AppError.
func translateRepoErr(err error, resource string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrs.ErrNotFound(resource)
	}
	return apperrs.ErrInternal(err.Error())
}
```

- [ ] **Step 5: Run the test, expect PASS**

Run: `go test ./internal/services/ -run TestAnnouncement_CreateDraft_DoesNotBroadcast -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go internal/services/testhelper_test.go
git commit -m "feat(phase-07): announcement service Create (draft) + fakeHub + first test"
```

---

## Task 7: Publish flow (broadcast must fire)

**Files:**
- Modify: `internal/services/announcement_service.go`
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 1: Write the failing test**

Append to `announcement_service_test.go`:

```go
func TestAnnouncement_Publish_Broadcasts(t *testing.T) {
	tb := newTestBox(t)
	hub := &fakeHub{}
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, hub)

	admin := tb.makeAdmin(t)
	draft, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title:          "Will publish",
		Body:           "<p>x</p>",
		TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)
	require.Equal(t, models.AnnouncementStatusDraft, draft.Status)
	require.Len(t, hub.events, 0)

	got, err := svc.Publish(context.Background(), admin, draft.ID, dto.AnnouncementPublish{})
	require.NoError(t, err)
	assert.Equal(t, models.AnnouncementStatusPublished, got.Status)
	require.NotNil(t, got.PublishedAt)

	require.Len(t, hub.events, 1)
	assert.Equal(t, "announcement_published", hub.events[0].Type)
}

func TestAnnouncement_Publish_AlreadyPublished_Conflict(t *testing.T) {
	tb := newTestBox(t)
	hub := &fakeHub{}
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, hub)

	admin := tb.makeAdmin(t)
	pub := models.AnnouncementStatusPublished
	got, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title:          "Already pub",
		Body:           "<p>x</p>",
		Status:         &pub,
		TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)
	require.Equal(t, models.AnnouncementStatusPublished, got.Status)

	_, err = svc.Publish(context.Background(), admin, got.ID, dto.AnnouncementPublish{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already published")
}
```

- [ ] **Step 2: Run, verify failure**

Run: `go test ./internal/services/ -run TestAnnouncement_Publish -v`
Expected: FAIL — `svc.Publish undefined`.

- [ ] **Step 3: Implement `Publish` in the service**

Append to `internal/services/announcement_service.go`:

```go
// Publish transitions a draft/scheduled announcement to published, sets
// PublishedAt, and broadcasts an SSE event.
func (s *AnnouncementService) Publish(
	ctx context.Context, currentUser *models.User, id uuid.UUID, req dto.AnnouncementPublish,
) (*models.Announcement, error) {
	a, err := s.repo.FindByID(ctx, id, repositories.FindOptions{})
	if err != nil {
		return nil, translateRepoErr(err, "announcement")
	}
	if a.Status == models.AnnouncementStatusPublished {
		return nil, apperrs.ErrConflict("announcement is already published")
	}
	if a.Status == models.AnnouncementStatusArchived {
		return nil, apperrs.ErrConflict("archived announcements cannot be published")
	}

	at := time.Now().UTC()
	if req.PublishedAt != nil {
		at = req.PublishedAt.UTC()
	}
	if err := s.repo.Publish(ctx, id, at); err != nil {
		return nil, translateRepoErr(err, "announcement")
	}

	fresh, err := s.repo.FindByID(ctx, id, repositories.FindOptions{
		WithAuthor: true, WithLabels: true, WithAttachments: true, WithTargetDepartments: true,
	})
	if err != nil {
		return nil, translateRepoErr(err, "announcement")
	}
	s.broadcastPublished(fresh)
	return fresh, nil
}
```

- [ ] **Step 4: Run, verify pass**

Run: `go test ./internal/services/ -run TestAnnouncement_Publish -v`
Expected: both subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go
git commit -m "feat(phase-07): publish flow broadcasts SSE event, blocks double-publish"
```

---

## Task 8: Update + Delete

**Files:**
- Modify: `internal/services/announcement_service.go`
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 1: Write the failing tests**

Append:

```go
func TestAnnouncement_Update_AppliesPartial(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	a, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title:          "Old title",
		Body:           "<p>old</p>",
		TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	newTitle := "New title"
	got, err := svc.Update(context.Background(), admin, a.ID, dto.AnnouncementUpdate{
		Title: &newTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, "New title", got.Title)
	assert.Equal(t, "<p>old</p>", got.Body, "body must be untouched when not in patch")
}

func TestAnnouncement_Delete_SoftDeletes(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})
	admin := tb.makeAdmin(t)
	a, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Delete me", Body: "<p>x</p>", TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(context.Background(), admin, a.ID))

	_, err = tb.repos.Announcement.FindByID(context.Background(), a.ID, repositories.FindOptions{})
	require.Error(t, err, "soft-deleted row should not be returned by NotDeleted-scoped FindByID")
}
```

- [ ] **Step 2: Run, verify failure**

Run: `go test ./internal/services/ -run TestAnnouncement_Update_AppliesPartial -v`
Expected: FAIL — `svc.Update` undefined.

- [ ] **Step 3: Implement Update + Delete**

Append to `internal/services/announcement_service.go`:

```go
func (s *AnnouncementService) Update(
	ctx context.Context, currentUser *models.User, id uuid.UUID, req dto.AnnouncementUpdate,
) (*models.Announcement, error) {
	a, err := s.repo.FindByID(ctx, id, repositories.FindOptions{})
	if err != nil {
		return nil, translateRepoErr(err, "announcement")
	}

	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Body != nil {
		a.Body = *req.Body
	}
	if req.Summary != nil {
		a.Summary = req.Summary
	}
	if req.Status != nil {
		a.Status = *req.Status
		if a.Status == models.AnnouncementStatusPublished && a.PublishedAt == nil {
			now := time.Now().UTC()
			a.PublishedAt = &now
		}
	}
	if req.ScheduledAt != nil {
		a.ScheduledAt = req.ScheduledAt
	}
	if req.TargetAudience != nil {
		a.TargetAudience = *req.TargetAudience
	}
	if req.Pinned != nil {
		a.Pinned = *req.Pinned
	}
	if req.CoverImageURL != nil {
		a.CoverImageURL = req.CoverImageURL
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, apperrs.ErrInternal("update announcement: " + err.Error())
	}

	if req.LabelIDs != nil {
		if err := s.repo.SetLabels(ctx, a.ID, *req.LabelIDs); err != nil {
			return nil, apperrs.ErrInternal("set labels: " + err.Error())
		}
	}
	if req.TargetDepartmentIDs != nil {
		if err := s.repo.SetTargetDepartments(ctx, a.ID, *req.TargetDepartmentIDs); err != nil {
			return nil, apperrs.ErrInternal("set target departments: " + err.Error())
		}
	}
	if req.Attachments != nil {
		if err := s.repo.ReplaceAttachments(ctx, a.ID, toAttachmentModels(*req.Attachments)); err != nil {
			return nil, apperrs.ErrInternal("set attachments: " + err.Error())
		}
	}

	return s.repo.FindByID(ctx, a.ID, repositories.FindOptions{
		WithAuthor: true, WithLabels: true, WithAttachments: true, WithTargetDepartments: true,
	})
}

func (s *AnnouncementService) Delete(
	ctx context.Context, currentUser *models.User, id uuid.UUID,
) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return translateRepoErr(err, "announcement")
	}
	return nil
}
```

- [ ] **Step 4: Run, verify pass**

Run: `go test ./internal/services/ -run "TestAnnouncement_(Update_AppliesPartial|Delete_SoftDeletes)" -v`
Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go
git commit -m "feat(phase-07): announcement update + soft-delete"
```

---

## Task 9: Get with visibility check + view tracking

**Files:**
- Modify: `internal/services/announcement_service.go`
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 1: Write the failing tests**

Append:

```go
func TestAnnouncement_Get_Published_AllVisible(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	emp := tb.makeEmployee(t) // non-admin, no manage perm

	pub := models.AnnouncementStatusPublished
	a, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "All", Body: "<p>x</p>", Status: &pub, TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	got, err := svc.Get(context.Background(), emp, a.ID)
	require.NoError(t, err)
	assert.Equal(t, a.ID, got.ID)
	assert.True(t, got.ViewedByMe, "Get must mark view")
	assert.GreaterOrEqual(t, got.ViewCount, int64(1))
}

func TestAnnouncement_Get_Draft_HiddenFromNonAuthor(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	emp := tb.makeEmployee(t)

	draft, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Draft", Body: "<p>x</p>", TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	_, err = svc.Get(context.Background(), emp, draft.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAnnouncement_Get_Department_HiddenFromOutsideDept(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	dept := tb.makeDepartment(t, "Eng")
	otherDept := tb.makeDepartment(t, "Sales")
	insider := tb.makeEmployeeInDept(t, dept.ID)
	outsider := tb.makeEmployeeInDept(t, otherDept.ID)

	pub := models.AnnouncementStatusPublished
	a, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title:               "Eng only",
		Body:                "<p>x</p>",
		Status:              &pub,
		TargetAudience:      models.AnnouncementTargetDepartment,
		TargetDepartmentIDs: []uuid.UUID{dept.ID},
	})
	require.NoError(t, err)

	_, err = svc.Get(context.Background(), insider, a.ID)
	require.NoError(t, err, "insider must see dept-targeted announcement")

	_, err = svc.Get(context.Background(), outsider, a.ID)
	require.Error(t, err, "outsider must NOT see dept-targeted announcement")
}
```

- [ ] **Step 2: Run, verify failure**

Run: `go test ./internal/services/ -run "TestAnnouncement_Get_" -v`
Expected: FAIL — `svc.Get` undefined / `makeEmployee` / `makeEmployeeInDept` undefined.

- [ ] **Step 3: Add the helpers**

In `internal/services/testhelper_test.go` add or extend (only if absent — Phase 2 likely added `makeEmployee`):

```go
// makeEmployee creates a user with only basic 'auth' permissions (no
// PermAnnounceManage). The exact permission set depends on the Phase 2
// helper; if `makeEmployee` already exists, do not re-declare.
func (b *testBox) makeEmployeeInDept(t *testing.T, deptID uuid.UUID) *models.User {
	t.Helper()
	u := b.makeEmployee(t)
	require.NoError(t, b.db.Model(&models.User{}).
		Where("id = ?", u.ID).
		Update("department_id", deptID).Error)
	// reload
	require.NoError(t, b.db.First(u, "id = ?", u.ID).Error)
	return u
}
```

- [ ] **Step 4: Implement `Get` + visibility**

Append to `internal/services/announcement_service.go`:

```go
// Get returns one announcement enforcing visibility, and marks it viewed
// by the current user.
func (s *AnnouncementService) Get(
	ctx context.Context, currentUser *models.User, id uuid.UUID,
) (*dto.AnnouncementRead, error) {
	a, err := s.repo.FindByID(ctx, id, repositories.FindOptions{
		WithAuthor: true, WithLabels: true, WithAttachments: true, WithTargetDepartments: true,
	})
	if err != nil {
		return nil, translateRepoErr(err, "announcement")
	}

	allowed, err := s.canView(ctx, currentUser, a)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, apperrs.ErrNotFound("announcement") // hide existence
	}

	// Mark viewed (best-effort).
	_ = s.repo.MarkViewed(ctx, a.ID, currentUser.ID)

	return s.toRead(ctx, a, currentUser), nil
}

func (s *AnnouncementService) canView(
	ctx context.Context, u *models.User, a *models.Announcement,
) (bool, error) {
	if u == nil {
		return false, nil
	}
	if u.HasPermission(models.PermAnnounceManage) || a.AuthorID == u.ID {
		return true, nil
	}
	if a.Status != models.AnnouncementStatusPublished {
		return false, nil
	}
	switch a.TargetAudience {
	case models.AnnouncementTargetAll:
		return true, nil
	case models.AnnouncementTargetDepartment:
		userDeptIDs, err := s.repo.UserDepartmentIDs(ctx, u.ID)
		if err != nil {
			return false, apperrs.ErrInternal(err.Error())
		}
		set := map[uuid.UUID]struct{}{}
		for _, d := range userDeptIDs {
			set[d] = struct{}{}
		}
		for _, td := range a.TargetDepartments {
			if _, ok := set[td.DepartmentID]; ok {
				return true, nil
			}
		}
		return false, nil
	case models.AnnouncementTargetCustom:
		// 'custom' here intentionally collapses to author/manage-only
		// (Python's recipient_ids per-user list is not modeled in v2; see
		// design doc — only department and all are first-class).
		return false, nil
	}
	return false, nil
}

func (s *AnnouncementService) toRead(
	ctx context.Context, a *models.Announcement, viewer *models.User,
) *dto.AnnouncementRead {
	labels := make([]dto.LabelRef, 0, len(a.Labels))
	for _, l := range a.Labels {
		var color *string
		if l.Color != "" {
			c := l.Color
			color = &c
		}
		labels = append(labels, dto.LabelRef{ID: l.ID, Name: l.Name, Color: color})
	}

	var depts []dto.DepartmentRef
	for _, td := range a.TargetDepartments {
		depts = append(depts, dto.DepartmentRef{ID: td.DepartmentID})
	}

	atts := make([]dto.AttachmentRead, 0, len(a.Attachments))
	for _, at := range a.Attachments {
		atts = append(atts, dto.AttachmentRead{
			ID:          at.ID,
			URL:         at.URL,
			Filename:    at.Filename,
			ContentType: at.ContentType,
			SizeBytes:   at.SizeBytes,
		})
	}

	views, _ := s.repo.CountViews(ctx, a.ID)
	viewed := false
	if viewer != nil {
		viewed, _ = s.repo.ViewedByMe(ctx, a.ID, viewer.ID)
	}

	authorRef := dto.AuthorRef{ID: a.AuthorID}
	if a.Author != nil {
		authorRef.FullName = a.Author.FullName
		authorRef.Email = a.Author.Email
	}

	return &dto.AnnouncementRead{
		ID:                a.ID,
		Title:             a.Title,
		Body:              a.Body,
		Summary:           a.Summary,
		Author:            authorRef,
		Status:            a.Status,
		ScheduledAt:       a.ScheduledAt,
		PublishedAt:       a.PublishedAt,
		TargetAudience:    a.TargetAudience,
		TargetDepartments: depts,
		Labels:            labels,
		Pinned:            a.Pinned,
		CoverImageURL:     a.CoverImageURL,
		Attachments:       atts,
		ViewCount:         views,
		ViewedByMe:        viewed,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
}

// MarkViewed is the public service method behind POST /:id/view.
func (s *AnnouncementService) MarkViewed(
	ctx context.Context, currentUser *models.User, id uuid.UUID,
) error {
	a, err := s.repo.FindByID(ctx, id, repositories.FindOptions{WithTargetDepartments: true})
	if err != nil {
		return translateRepoErr(err, "announcement")
	}
	allowed, err := s.canView(ctx, currentUser, a)
	if err != nil {
		return err
	}
	if !allowed {
		return apperrs.ErrNotFound("announcement")
	}
	return s.repo.MarkViewed(ctx, id, currentUser.ID)
}
```

The `User.HasPermission(perm string)` method is assumed to exist from Phase 1 (`PermAll == "*"` bypass already wired). The constant `models.PermAnnounceManage` is the string literal `"announcements:manage"` from the Phase 0 permission registry.

> If `models.PermAnnounceManage` does not exist as a `models.*` re-export, replace with `permissions.PermAnnounceManage` and add the import accordingly. The brief lists this constant in the spec as `PermAnnounceManage` under `internal/permissions/registry.go` — its package path is `github.com/exnodes/hrm-api/internal/permissions`. Adjust the call sites in the service to `currentUser.HasPermission(string(permissions.PermAnnounceManage))` if the registry is the source of truth.

- [ ] **Step 5: Run, verify pass**

Run: `go test ./internal/services/ -run "TestAnnouncement_Get_" -v`
Expected: all three subtests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go internal/services/testhelper_test.go
git commit -m "feat(phase-07): Get with visibility filter + view tracking"
```

---

## Task 10: List with scope filter and visibility

**Files:**
- Modify: `internal/services/announcement_service.go`
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 1: Write the failing test**

Append:

```go
func TestAnnouncement_List_HidesDraftsFromNonAuthor(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	emp := tb.makeEmployee(t)

	pub := models.AnnouncementStatusPublished
	_, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Published", Body: "<p>x</p>", Status: &pub, TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Draft", Body: "<p>x</p>", TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	page, total, err := svc.List(context.Background(), emp, dto.AnnouncementListQuery{
		Page: 1, PageSize: 20, Scope: dto.AnnouncementScopeAll,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "employee must see only the published one")
	require.Len(t, page, 1)
	assert.Equal(t, "Published", page[0].Title)
}

func TestAnnouncement_List_ScopeMine(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	other := tb.makeAdmin(t) // both have manage perm

	pub := models.AnnouncementStatusPublished
	_, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Mine", Body: "<p>x</p>", Status: &pub, TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), other, dto.AnnouncementCreate{
		Title: "Other's", Body: "<p>x</p>", Status: &pub, TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	page, total, err := svc.List(context.Background(), admin, dto.AnnouncementListQuery{
		Page: 1, PageSize: 20, Scope: dto.AnnouncementScopeMine,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, page, 1)
	assert.Equal(t, "Mine", page[0].Title)
}
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./internal/services/ -run "TestAnnouncement_List_" -v`
Expected: FAIL — `svc.List` undefined.

- [ ] **Step 3: Implement `List`**

Append:

```go
// List returns announcements visible to `currentUser`. Admins / managers
// see drafts; everyone else only sees published items that target them.
func (s *AnnouncementService) List(
	ctx context.Context, currentUser *models.User, q dto.AnnouncementListQuery,
) ([]dto.AnnouncementRead, int64, error) {
	if currentUser == nil {
		return nil, 0, apperrs.ErrUnauthorized("authentication required")
	}

	rq := repositories.ListAnnouncementQuery{
		Page:     q.Page,
		PageSize: q.PageSize,
		Search:   q.Search,
		Status:   q.Status,
		LabelID:  q.LabelID,
		Pinned:   q.Pinned,
	}

	canManage := currentUser.HasPermission(string(modelsPermAnnounceManage))
	switch q.Scope {
	case dto.AnnouncementScopeMine:
		uid := currentUser.ID
		rq.OnlyAuthorID = &uid
	case dto.AnnouncementScopeTargetedMe:
		uid := currentUser.ID
		rq.OnlyPublished = true
		rq.OnlyVisibleToUserID = &uid
	default: // "all"
		if !canManage {
			uid := currentUser.ID
			rq.OnlyPublished = true
			rq.OnlyVisibleToUserID = &uid
		}
	}

	items, total, err := s.repo.List(ctx, rq)
	if err != nil {
		return nil, 0, apperrs.ErrInternal("list announcements: " + err.Error())
	}

	out := make([]dto.AnnouncementRead, 0, len(items))
	for i := range items {
		out = append(out, *s.toRead(ctx, &items[i], currentUser))
	}
	return out, total, nil
}

// modelsPermAnnounceManage holds the permission string. Defined as a local
// const to avoid the import path drift mentioned above; if your project
// re-exports `models.PermAnnounceManage`, alias to that instead.
const modelsPermAnnounceManage = "announcements:manage"
```

- [ ] **Step 4: Run, verify pass**

Run: `go test ./internal/services/ -run "TestAnnouncement_List_" -v`
Expected: both subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go
git commit -m "feat(phase-07): list with scope + visibility filters"
```

---

## Task 11: Mobile-specific service methods

**Files:**
- Modify: `internal/services/announcement_service.go`
- Modify: `internal/services/announcement_service_test.go`

- [ ] **Step 1: Write the failing test**

Append:

```go
func TestAnnouncement_ListMobile_OnlyPublishedAndPinnedFirst(t *testing.T) {
	tb := newTestBox(t)
	svc := NewAnnouncementService(tb.repos.Announcement, tb.repos.User, &fakeHub{})

	admin := tb.makeAdmin(t)
	emp := tb.makeEmployee(t)

	pub := models.AnnouncementStatusPublished
	_, err := svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Normal", Body: "<p>x</p>", Status: &pub, TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Pinned", Body: "<p>x</p>", Status: &pub, Pinned: true, TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), admin, dto.AnnouncementCreate{
		Title: "Draft", Body: "<p>x</p>", TargetAudience: models.AnnouncementTargetAll,
	})
	require.NoError(t, err)

	items, total, err := svc.ListMobile(context.Background(), emp, dto.MobileAnnouncementListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)
	assert.Equal(t, "Pinned", items[0].Title, "pinned must come first")
}
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./internal/services/ -run TestAnnouncement_ListMobile -v`
Expected: FAIL — `svc.ListMobile` undefined.

- [ ] **Step 3: Implement mobile methods**

Append:

```go
// ListMobile returns the trimmed mobile representation, paginated, only
// for the current user (published + targeted).
func (s *AnnouncementService) ListMobile(
	ctx context.Context, currentUser *models.User, q dto.MobileAnnouncementListQuery,
) ([]dto.MobileAnnouncementRead, int64, error) {
	if currentUser == nil {
		return nil, 0, apperrs.ErrUnauthorized("authentication required")
	}
	uid := currentUser.ID
	items, total, err := s.repo.List(ctx, repositories.ListAnnouncementQuery{
		Page:                q.Page,
		PageSize:            q.PageSize,
		Search:              q.Search,
		OnlyPublished:       true,
		OnlyVisibleToUserID: &uid,
	})
	if err != nil {
		return nil, 0, apperrs.ErrInternal("list announcements: " + err.Error())
	}
	out := make([]dto.MobileAnnouncementRead, 0, len(items))
	for i := range items {
		out = append(out, *s.toMobile(ctx, &items[i], currentUser))
	}
	return out, total, nil
}

// GetMobile returns the full mobile view of one announcement.
func (s *AnnouncementService) GetMobile(
	ctx context.Context, currentUser *models.User, id uuid.UUID,
) (*dto.MobileAnnouncementRead, error) {
	a, err := s.repo.FindByID(ctx, id, repositories.FindOptions{
		WithLabels: true, WithTargetDepartments: true,
	})
	if err != nil {
		return nil, translateRepoErr(err, "announcement")
	}
	allowed, err := s.canView(ctx, currentUser, a)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, apperrs.ErrNotFound("announcement")
	}
	_ = s.repo.MarkViewed(ctx, a.ID, currentUser.ID)
	return s.toMobile(ctx, a, currentUser), nil
}

func (s *AnnouncementService) toMobile(
	ctx context.Context, a *models.Announcement, viewer *models.User,
) *dto.MobileAnnouncementRead {
	labels := make([]dto.LabelRef, 0, len(a.Labels))
	for _, l := range a.Labels {
		var color *string
		if l.Color != "" {
			c := l.Color
			color = &c
		}
		labels = append(labels, dto.LabelRef{ID: l.ID, Name: l.Name, Color: color})
	}
	viewed := false
	if viewer != nil {
		viewed, _ = s.repo.ViewedByMe(ctx, a.ID, viewer.ID)
	}
	return &dto.MobileAnnouncementRead{
		ID:          a.ID,
		Title:       a.Title,
		Summary:     a.Summary,
		Body:        a.Body,
		PublishedAt: a.PublishedAt,
		Pinned:      a.Pinned,
		IsRead:      viewed,
		Labels:      labels,
		CoverImage:  a.CoverImageURL,
		Status:      a.Status,
	}
}
```

- [ ] **Step 4: Run, verify pass**

Run: `go test ./internal/services/ -run TestAnnouncement_ListMobile -v`
Expected: PASS.

- [ ] **Step 5: Run the entire service test suite to catch regressions**

Run: `go test ./internal/services/ -run TestAnnouncement -v`
Expected: all 8+ subtests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/services/announcement_service.go internal/services/announcement_service_test.go
git commit -m "feat(phase-07): mobile list + mobile detail service methods"
```

---

## Task 12: SSE handler with JWT-from-query support

**Files:**
- Modify: `internal/middleware/jwt.go` (add helper `JWTFromQueryOrHeader`)
- Create: `internal/handlers/sse_handler.go`

- [ ] **Step 1: Add the query-OR-header JWT middleware**

Open `internal/middleware/jwt.go` (existing from Phase 1). Append:

```go
// JWTFromQueryOrHeader is a JWT middleware variant for endpoints where the
// browser cannot set Authorization headers (notably EventSource for SSE).
// It accepts the token in either:
//   - `Authorization: Bearer <jwt>` header, OR
//   - `?token=<jwt>` query parameter.
//
// LIMITATION: query-param tokens may end up in proxy/server access logs.
// Operators should configure log scrubbing for the `token` parameter on
// `/api/v1/sse/*` routes. Mitigation: use short-lived access tokens.
func JWTFromQueryOrHeader(secret []byte, userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		token := ""
		if strings.HasPrefix(raw, "Bearer ") {
			token = strings.TrimPrefix(raw, "Bearer ")
		} else if q := c.Query("token"); q != "" {
			token = q
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response[any]{Success: false, Message: "missing token"})
			return
		}
		// Re-use the same parse + user-load logic used by JWT(). The exact
		// helper name depends on Phase 1 implementation. Common choices:
		//   - parseAndLoadUser(token, secret, userRepo) (*models.User, error)
		// If Phase 1 only exposed a private helper, extract it into a small
		// exported function `ParseAndLoadUser` and call it from both
		// `JWT()` and here.
		user, err := ParseAndLoadUser(c.Request.Context(), token, secret, userRepo)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response[any]{Success: false, Message: "invalid token"})
			return
		}
		c.Set("current_user", user)
		c.Set("user_id", user.ID.String())
		c.Next()
	}
}
```

If `ParseAndLoadUser` does not already exist, extract the body of the existing `JWT()` middleware's parse+load section into a new exported `ParseAndLoadUser(ctx context.Context, token string, secret []byte, userRepo repositories.UserRepository) (*models.User, error)` function. Both middlewares now delegate to it. Required imports: `net/http`, `strings`, `context`, `github.com/exnodes/hrm-api/internal/dto`, `github.com/exnodes/hrm-api/internal/models`, `github.com/exnodes/hrm-api/internal/repositories`.

- [ ] **Step 2: Build to confirm middleware still compiles**

Run: `go build ./internal/middleware/...`
Expected: exit 0.

- [ ] **Step 3: Write the SSE handler**

```go
// internal/handlers/sse_handler.go
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/sse"
)

type SSEHandler struct {
	hub *sse.Hub
}

func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// Stream godoc
// @Summary      Subscribe to announcement events via SSE
// @Description  Long-lived Server-Sent Events stream. Emits `announcement_published` events when a new announcement is published. Clients should reconnect on disconnect. Auth token may be passed as `Authorization: Bearer ...` OR `?token=` query param (EventSource limitation). Sends keep-alive comments every 30s.
// @Tags         SSE
// @Produce      text/event-stream
// @Param        token  query     string  false  "JWT access token (alternative to Authorization header)"
// @Success      200    {string}  string  "event stream"
// @Failure      401    {object}  dto.Response[any]
// @Router       /sse/announcements [get]
// @Security     BearerAuth
func (h *SSEHandler) Stream(c *gin.Context) {
	userVal, ok := c.Get("current_user")
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user, ok := userVal.(*models.User)
	if !ok || user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ch, unsubscribe := h.hub.Subscribe(user.ID)
	defer unsubscribe()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Initial hello frame so curl/EventSource confirms the stream is alive.
	connID := uuid.NewString()
	_, _ = fmt.Fprintf(c.Writer, "event: connected\ndata: {\"connection_id\":\"%s\"}\n\n", connID)
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	gone := c.Request.Context().Done()
	for {
		select {
		case <-gone:
			return
		case payload, open := <-ch:
			if !open {
				return
			}
			_, err := io.WriteString(c.Writer, fmt.Sprintf("event: announcement_published\ndata: %s\n\n", string(payload)))
			if err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := io.WriteString(c.Writer, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
```

- [ ] **Step 4: Build to confirm compile**

Run: `go build ./internal/handlers/...`
Expected: exit 0.

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/jwt.go internal/handlers/sse_handler.go
git commit -m "feat(phase-07): SSE handler + JWT-from-query-or-header middleware variant"
```

---

## Task 13: Announcement HTTP handler — full route surface

**Files:**
- Create: `internal/handlers/announcement_handler.go`

- [ ] **Step 1: Write the handler**

```go
package handlers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrs "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/services"
)

type AnnouncementHandler struct {
	svc *services.AnnouncementService
}

func NewAnnouncementHandler(svc *services.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{svc: svc}
}

// helper: read current_user set by JWT middleware
func currentUser(c *gin.Context) *models.User {
	v, ok := c.Get("current_user")
	if !ok {
		return nil
	}
	u, _ := v.(*models.User)
	return u
}

// helper: parse :id path param as UUID
func parseUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	raw := c.Param(name)
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response[any]{Success: false, Message: "invalid " + name})
		return uuid.Nil, false
	}
	return id, true
}

// List godoc
// @Summary      List announcements (web)
// @Tags         Announcements
// @Security     BearerAuth
// @Produce      json
// @Param        page         query    int     false  "page"          default(1)
// @Param        page_size    query    int     false  "page size"     default(20)
// @Param        search       query    string  false  "title search"
// @Param        status       query    string  false  "draft|scheduled|published|archived"
// @Param        label_id     query    string  false  "label UUID"
// @Param        pinned       query    bool    false  "filter by pinned"
// @Param        scope        query    string  false  "all|mine|targeted-at-me"  default(all)
// @Success      200          {object} dto.Response[dto.PaginatedData[dto.AnnouncementRead]]
// @Router       /announcements [get]
func (h *AnnouncementHandler) List(c *gin.Context) {
	user := currentUser(c)
	var q dto.AnnouncementListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response[any]{Success: false, Message: err.Error()})
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), user, q)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.AnnouncementRead]]{
		Success: true,
		Data: dto.PaginatedData[dto.AnnouncementRead]{
			Items:      items,
			Total:      total,
			Page:       q.Page,
			PageSize:   q.PageSize,
			TotalPages: totalPages(total, q.PageSize),
		},
	})
}

// Create godoc
// @Summary      Create announcement
// @Tags         Announcements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body   body      dto.AnnouncementCreate  true  "create payload"
// @Success      201    {object}  dto.Response[dto.AnnouncementRead]
// @Failure      400    {object}  dto.Response[any]
// @Failure      403    {object}  dto.Response[any]
// @Router       /announcements [post]
func (h *AnnouncementHandler) Create(c *gin.Context) {
	user := currentUser(c)
	var req dto.AnnouncementCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response[any]{Success: false, Message: err.Error()})
		return
	}
	a, err := h.svc.Create(c.Request.Context(), user, req)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	read := h.svc.ToRead(c.Request.Context(), a, user)
	msg := "Announcement saved as draft"
	if a.Status == models.AnnouncementStatusPublished {
		msg = "Announcement published"
	}
	c.JSON(http.StatusCreated, dto.Response[dto.AnnouncementRead]{Success: true, Message: msg, Data: *read})
}

// Get godoc
// @Summary      Get announcement
// @Tags         Announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Announcement UUID"
// @Success      200  {object}  dto.Response[dto.AnnouncementRead]
// @Failure      404  {object}  dto.Response[any]
// @Router       /announcements/{id} [get]
func (h *AnnouncementHandler) Get(c *gin.Context) {
	user := currentUser(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	read, err := h.svc.Get(c.Request.Context(), user, id)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AnnouncementRead]{Success: true, Data: *read})
}

// Update godoc
// @Summary      Update announcement
// @Tags         Announcements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string                  true  "Announcement UUID"
// @Param        body body      dto.AnnouncementUpdate  true  "patch payload"
// @Success      200  {object}  dto.Response[dto.AnnouncementRead]
// @Router       /announcements/{id} [patch]
func (h *AnnouncementHandler) Update(c *gin.Context) {
	user := currentUser(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AnnouncementUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response[any]{Success: false, Message: err.Error()})
		return
	}
	a, err := h.svc.Update(c.Request.Context(), user, id, req)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	read := h.svc.ToRead(c.Request.Context(), a, user)
	c.JSON(http.StatusOK, dto.Response[dto.AnnouncementRead]{Success: true, Message: "Announcement updated", Data: *read})
}

// Delete godoc
// @Summary      Delete announcement
// @Tags         Announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Announcement UUID"
// @Success      200  {object}  dto.Response[any]
// @Router       /announcements/{id} [delete]
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	user := currentUser(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), user, id); err != nil {
		apperrs.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Announcement deleted"})
}

// Publish godoc
// @Summary      Publish announcement
// @Tags         Announcements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string                   true  "Announcement UUID"
// @Param        body body      dto.AnnouncementPublish  false "optional explicit published_at"
// @Success      200  {object}  dto.Response[dto.AnnouncementRead]
// @Failure      409  {object}  dto.Response[any]  "already published or archived"
// @Router       /announcements/{id}/publish [post]
func (h *AnnouncementHandler) Publish(c *gin.Context) {
	user := currentUser(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AnnouncementPublish
	_ = c.ShouldBindJSON(&req) // body optional
	a, err := h.svc.Publish(c.Request.Context(), user, id, req)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	read := h.svc.ToRead(c.Request.Context(), a, user)
	c.JSON(http.StatusOK, dto.Response[dto.AnnouncementRead]{Success: true, Message: "Announcement published", Data: *read})
}

// MarkViewed godoc
// @Summary      Mark announcement as viewed
// @Tags         Announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Announcement UUID"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /announcements/{id}/view [post]
func (h *AnnouncementHandler) MarkViewed(c *gin.Context) {
	user := currentUser(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.MarkViewed(c.Request.Context(), user, id); err != nil {
		apperrs.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Marked as viewed"})
}

// MobileList godoc
// @Summary      Mobile: list announcements
// @Tags         Mobile - Announcements
// @Security     BearerAuth
// @Produce      json
// @Param        page       query    int     false  "page"       default(1)
// @Param        page_size  query    int     false  "page_size"  default(20)
// @Param        search     query    string  false  "title search"
// @Success      200        {object} dto.Response[dto.PaginatedData[dto.MobileAnnouncementRead]]
// @Router       /mobile/announcements [get]
func (h *AnnouncementHandler) MobileList(c *gin.Context) {
	user := currentUser(c)
	var q dto.MobileAnnouncementListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response[any]{Success: false, Message: err.Error()})
		return
	}
	items, total, err := h.svc.ListMobile(c.Request.Context(), user, q)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.MobileAnnouncementRead]]{
		Success: true,
		Data: dto.PaginatedData[dto.MobileAnnouncementRead]{
			Items:      items,
			Total:      total,
			Page:       q.Page,
			PageSize:   q.PageSize,
			TotalPages: totalPages(total, q.PageSize),
		},
	})
}

// MobileGet godoc
// @Summary      Mobile: get announcement detail
// @Tags         Mobile - Announcements
// @Security     BearerAuth
// @Produce      json
// @Param        id  path      string  true  "Announcement UUID"
// @Success      200 {object}  dto.Response[dto.MobileAnnouncementRead]
// @Failure      404 {object}  dto.Response[any]
// @Router       /mobile/announcements/{id} [get]
func (h *AnnouncementHandler) MobileGet(c *gin.Context) {
	user := currentUser(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	read, err := h.svc.GetMobile(c.Request.Context(), user, id)
	if err != nil {
		apperrs.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.MobileAnnouncementRead]{Success: true, Data: *read})
}

func totalPages(total int64, pageSize int) int {
	if total == 0 || pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}
```

This handler depends on the service exposing a public `ToRead` method (currently `toRead` private). Rename `toRead` → `ToRead` and `toMobile` → `ToMobile` in the service so the handler can reuse them:

In `internal/services/announcement_service.go` rename both methods:
- `func (s *AnnouncementService) toRead(...)` → `func (s *AnnouncementService) ToRead(...)`
- `func (s *AnnouncementService) toMobile(...)` → `func (s *AnnouncementService) ToMobile(...)`

…and update all internal callers within the same file.

- [ ] **Step 2: Build to confirm compile**

Run: `go build ./...`
Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/announcement_handler.go internal/services/announcement_service.go
git commit -m "feat(phase-07): announcement HTTP handlers with full swagger annotations"
```

---

## Task 14: Wire routes + SSE hub singleton in main

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Add the hub + service + handler wiring**

Open `cmd/server/main.go`. Locate the section where existing services and handlers are constructed (after DB connect and repo construction). Add:

```go
// --- Phase 7: announcements + SSE ---

// Single in-memory hub for the process lifetime. In-process pub/sub —
// must NOT be replicated; scaling beyond 1 replica requires Redis pub/sub
// backplane (see internal/sse/hub.go for SCALING LIMIT note).
sseHub := sse.NewHub()
defer sseHub.Stop()

announcementRepo := repositories.NewAnnouncementRepository(db)
announcementSvc := services.NewAnnouncementService(announcementRepo, userRepo, sseHubAdapter{hub: sseHub})
announcementH := handlers.NewAnnouncementHandler(announcementSvc)
sseH := handlers.NewSSEHandler(sseHub)
```

Where `userRepo` is the existing UserRepository instance from Phase 1.

Add the small adapter at the bottom of `main.go` (or in a tiny file `cmd/server/sse_adapter.go`):

```go
// sseHubAdapter bridges sse.Hub's concrete API to the service's HubBroadcaster
// interface (which uses `string` event type for testability).
type sseHubAdapter struct{ hub *sse.Hub }

func (a sseHubAdapter) Broadcast(eventType string, data interface{}, filter func(uuid.UUID) bool) {
	a.hub.Broadcast(sse.Event{Type: eventType, Data: data}, sse.FilterFunc(filter))
}
```

Required additional imports in `main.go`:

```go
"github.com/google/uuid"
"github.com/exnodes/hrm-api/internal/sse"
```

Then register routes. Locate the `authed := v1.Group("")` (or the equivalent name used in earlier phases) and append:

```go
// --- Announcements (web) ---
annGroup := authed.Group("/announcements")
{
	annGroup.GET("",         announcementH.List)
	annGroup.GET("/:id",     announcementH.Get)
	annGroup.POST("/:id/view", announcementH.MarkViewed)

	annGroup.POST("",        middleware.RequirePerms(permissions.PermAnnounceManage), announcementH.Create)
	annGroup.PATCH("/:id",   middleware.RequirePerms(permissions.PermAnnounceManage), announcementH.Update)
	annGroup.DELETE("/:id",  middleware.RequirePerms(permissions.PermAnnounceManage), announcementH.Delete)
	annGroup.POST("/:id/publish", middleware.RequirePerms(permissions.PermAnnounceManage), announcementH.Publish)
}

// --- Mobile announcements ---
mobileGroup := authed.Group("/mobile/announcements")
{
	mobileGroup.GET("",     announcementH.MobileList)
	mobileGroup.GET("/:id", announcementH.MobileGet)
}

// --- SSE ---
// SSE uses a separate JWT middleware that accepts ?token= because EventSource
// cannot set Authorization headers.
sseGroup := v1.Group("/sse")
sseGroup.Use(middleware.JWTFromQueryOrHeader(cfg.JWTSecret, userRepo))
sseGroup.GET("/announcements", sseH.Stream)
```

Adjust constant / variable names (`cfg.JWTSecret`, `permissions`, etc.) to match the actual symbols used by Phases 0–6. If `permissions.PermAnnounceManage` lives under a different package alias, use the matching one.

- [ ] **Step 2: Build the whole binary**

Run: `go build ./cmd/server/...`
Expected: exit 0.

- [ ] **Step 3: Run all service tests to catch regressions across phases**

Run: `go test ./internal/services/...`
Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go cmd/server/sse_adapter.go
git commit -m "feat(phase-07): wire sse.Hub singleton, announcement routes, SSE endpoint"
```

---

## Task 15: Generate Swagger and confirm routes are visible

**Files:**
- Modify: generated swagger files under `docs/` (auto-generated)

- [ ] **Step 1: Re-generate Swagger**

Run: `make swag` (this should invoke `swag init -g cmd/server/main.go -o docs/`; if Makefile uses a different target, run that one — the project added it in Phase 0).
Expected: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml` regenerated without errors.

- [ ] **Step 2: Start the server**

Run (in another shell or as background): `make run`
Expected: stdout contains `Listening on :8080` (or whatever port `.env` sets) and no `dirty migration` / `not up to date` errors.

- [ ] **Step 3: Verify swagger UI shows the new endpoints**

Run: `curl -s http://localhost:8080/swagger/doc.json | jq -r '.paths | keys[]' | grep -E '(announcements|sse)' | sort`
Expected output (order may vary):

```
/announcements
/announcements/{id}
/announcements/{id}/publish
/announcements/{id}/view
/mobile/announcements
/mobile/announcements/{id}
/sse/announcements
```

- [ ] **Step 4: Commit regenerated docs**

```bash
git add docs/
git commit -m "docs(phase-07): regenerate swagger with announcement + SSE routes"
```

---

## Task 16: End-to-end verification — happy path (admin lifecycle)

**Files:**
- Create: `docs/superpowers/verification/phase-07.md`

Start each step with the server already running (`make run` in another shell). All commands assume `BASE=http://localhost:8080/api/v1` and `JQ` is installed.

- [ ] **Step 1: Login as super-admin and capture access token**

Run:

```bash
export BASE=http://localhost:8080/api/v1
ADMIN_TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@example.com","password":"ChangeMe!123"}' \
  | jq -r '.data.access_token')
echo "len=${#ADMIN_TOKEN}"
```

Expected: `len=` followed by a number > 100. Capture this token; copy the raw login response into `phase-07.md` under section `## 1. Login`.

- [ ] **Step 2: Create a label (assumes Phase 4 endpoint exists)**

```bash
LABEL_ID=$(curl -s -X POST "$BASE/labels" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"General","color":"#3B82F6"}' \
  | jq -r '.data.id')
echo "label=$LABEL_ID"
```

Expected: `label=<uuid>`. Record output in verification log.

- [ ] **Step 3: Create a draft announcement with that label**

```bash
DRAFT_RESP=$(curl -s -X POST "$BASE/announcements" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"title\":\"Draft for verification\",
    \"body\":\"<p>Hello from Phase 7</p>\",
    \"target_audience\":\"all\",
    \"label_ids\":[\"$LABEL_ID\"]
  }")
DRAFT_ID=$(echo "$DRAFT_RESP" | jq -r '.data.id')
echo "$DRAFT_RESP" | jq '.data | {id,status,title,labels}'
```

Expected: status `"draft"`, labels array contains one entry, message `"Announcement saved as draft"`.

- [ ] **Step 4: Open an SSE stream in the background**

```bash
( curl -sN "$BASE/sse/announcements?token=$ADMIN_TOKEN" \
    > /tmp/sse-phase07.log 2>&1 & )
SSE_PID=$!
sleep 1
head -n 5 /tmp/sse-phase07.log
```

Expected: first two non-empty lines look like:

```
event: connected
data: {"connection_id":"<uuid>"}
```

(plus possibly a `: keepalive` later).

- [ ] **Step 5: Publish the draft → expect SSE event on the stream**

```bash
curl -s -X POST "$BASE/announcements/$DRAFT_ID/publish" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq '.data | {id, status, published_at}'
sleep 1
grep -A1 "announcement_published" /tmp/sse-phase07.log | head -n 20
```

Expected (publish response): `"status":"published"`, `"published_at"` not null.
Expected (SSE log):

```
event: announcement_published
data: {"id":"<uuid>","title":"Draft for verification",...
```

- [ ] **Step 6: Create + publish a second announcement to confirm continuous streaming**

```bash
SECOND_ID=$(curl -s -X POST "$BASE/announcements" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Second one","body":"<p>x</p>","target_audience":"all","status":"published"}' \
  | jq -r '.data.id')
sleep 1
grep -c "announcement_published" /tmp/sse-phase07.log
```

Expected: count >= 2.

- [ ] **Step 7: Stop the SSE stream**

```bash
kill $SSE_PID 2>/dev/null || true
```

- [ ] **Step 8: Log everything into `docs/superpowers/verification/phase-07.md`**

Create the file with the literal commands run and their captured outputs from Steps 1–7. Heading shape:

```markdown
# Phase 7 verification — Announcements + Mobile + SSE

Date: <UTC date>
Server: localhost:8080
DB: exnodes_hrm (migration version 9)

## 1. Login (admin)
<command + truncated response>

## 2. Create label
...

## 3. Create draft announcement
...

## 4. Subscribe to SSE
...

## 5. Publish announcement
...

## 6. Verify event on SSE
...
```

- [ ] **Step 9: Commit the partial log (more sections added in next tasks)**

```bash
git add docs/superpowers/verification/phase-07.md
git commit -m "docs(phase-07): verification log — admin happy path + SSE broadcast"
```

---

## Task 17: E2E verification — non-admin visibility and view tracking

- [ ] **Step 1: Create a non-admin user via admin API**

```bash
EMP_RESP=$(curl -s -X POST "$BASE/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"emp1@example.com","password":"Emp!12345","full_name":"Emp One"}')
EMP_ID=$(echo "$EMP_RESP" | jq -r '.data.id')
echo "emp_id=$EMP_ID"
```

Expected: a UUID. (Assumes Phase 2 users endpoint creates an Employee-role user by default.)

- [ ] **Step 2: Login as the employee**

```bash
EMP_TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"emp1@example.com","password":"Emp!12345"}' \
  | jq -r '.data.access_token')
echo "len=${#EMP_TOKEN}"
```

Expected: numeric length > 100.

- [ ] **Step 3: Employee lists announcements — must NOT see drafts**

First, create a fresh draft as admin so the data shape is mixed:

```bash
DRAFT2=$(curl -s -X POST "$BASE/announcements" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Visible-only-to-admin","body":"<p>x</p>","target_audience":"all"}' \
  | jq -r '.data.id')

curl -s "$BASE/announcements?page=1&page_size=50" \
  -H "Authorization: Bearer $EMP_TOKEN" \
  | jq '.data.items[] | {title,status}'
```

Expected output: zero rows with `status: "draft"`. Rows include `"Draft for verification"` (now published), `"Second one"` (published). Does NOT include `"Visible-only-to-admin"`.

- [ ] **Step 4: Employee marks viewed**

```bash
curl -s -X POST "$BASE/announcements/$DRAFT_ID/view" \
  -H "Authorization: Bearer $EMP_TOKEN" | jq
```

Expected: `{"success":true,"message":"Marked as viewed"}`. Re-fetching `/announcements/$DRAFT_ID` as employee should now return `viewed_by_me: true` and `view_count >= 1`.

- [ ] **Step 5: Mobile list**

```bash
curl -s "$BASE/mobile/announcements?page=1&page_size=10" \
  -H "Authorization: Bearer $EMP_TOKEN" \
  | jq '.data | {total, items: [.items[] | {title, pinned, is_read}]}'
```

Expected: at least one item with `is_read: true` (the one just viewed).

- [ ] **Step 6: Append to verification log + commit**

Append sections `## 7. Non-admin list` / `## 8. Mark viewed` / `## 9. Mobile list` to `docs/superpowers/verification/phase-07.md`.

```bash
git add docs/superpowers/verification/phase-07.md
git commit -m "docs(phase-07): verification log — non-admin visibility + mobile"
```

---

## Task 18: E2E verification — error paths

- [ ] **Step 1: Non-admin tries to create — must 403**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/announcements" \
  -H "Authorization: Bearer $EMP_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"nope","body":"<p>x</p>","target_audience":"all"}'
```

Expected: `403`.

- [ ] **Step 2: Get a non-existent announcement — 404**

```bash
curl -s -o /dev/null -w "%{http_code}\n" \
  "$BASE/announcements/00000000-0000-0000-0000-000000000000" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Expected: `404`.

- [ ] **Step 3: Double-publish — must 409**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/announcements/$DRAFT_ID/publish" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Expected: `409`. Confirm response body: `{"success":false,"message":"announcement is already published",...}` via:

```bash
curl -s -X POST "$BASE/announcements/$DRAFT_ID/publish" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

- [ ] **Step 4: Department-targeted visibility denied for outsider**

```bash
# 1. Pick any existing department UUID
DEPT_ID=$(curl -s "$BASE/departments?page=1&page_size=1" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq -r '.data.items[0].id')

# 2. Create published dept-targeted announcement (emp1 is NOT in that dept by default — verify)
DEPT_ANN=$(curl -s -X POST "$BASE/announcements" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"title\":\"Dept only\",
    \"body\":\"<p>x</p>\",
    \"status\":\"published\",
    \"target_audience\":\"department\",
    \"target_department_ids\":[\"$DEPT_ID\"]
  }" | jq -r '.data.id')

# 3. Employee (no department) tries to fetch -> 404 (hide existence)
curl -s -o /dev/null -w "%{http_code}\n" \
  "$BASE/announcements/$DEPT_ANN" \
  -H "Authorization: Bearer $EMP_TOKEN"
```

Expected: `404` for the third command (visibility-denied collapses to not-found).

- [ ] **Step 5: SSE with bad token — must 401**

```bash
curl -sN -o /dev/null -w "%{http_code}\n" "$BASE/sse/announcements?token=garbage"
```

Expected: `401`.

- [ ] **Step 6: Append to log + commit**

Append section `## 10. Error paths` with all five subcases and captured exit codes to `docs/superpowers/verification/phase-07.md`.

```bash
git add docs/superpowers/verification/phase-07.md
git commit -m "docs(phase-07): verification log — error paths"
```

---

## Task 19: DB spot-check

- [ ] **Step 1: Connect to Postgres and verify rows**

Run (replace `psql` connect string with the one used by `make migrate-up`):

```bash
psql "$DATABASE_URL" -c "SELECT id, title, status, target_audience, pinned FROM announcements WHERE is_deleted=false ORDER BY created_at;"
```

Expected: at least 4 rows (the draft now published, the second one, "Visible-only-to-admin" draft, "Dept only" published).

- [ ] **Step 2: Verify view tracking and join tables**

```bash
psql "$DATABASE_URL" -c "SELECT COUNT(*) AS views FROM announcement_views WHERE is_deleted=false;"
psql "$DATABASE_URL" -c "SELECT COUNT(*) AS labels FROM announcement_labels WHERE is_deleted=false;"
psql "$DATABASE_URL" -c "SELECT COUNT(*) AS dept_targets FROM announcement_target_departments WHERE is_deleted=false;"
```

Expected: views >= 1, labels >= 1, dept_targets >= 1.

- [ ] **Step 3: Verify soft-delete works end-to-end**

```bash
TO_DELETE=$(curl -s -X POST "$BASE/announcements" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Will be deleted","body":"<p>x</p>","target_audience":"all"}' \
  | jq -r '.data.id')

curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$BASE/announcements/$TO_DELETE" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Expect: row still in DB, but is_deleted = true
psql "$DATABASE_URL" -c "SELECT id, is_deleted, deleted_at FROM announcements WHERE id='$TO_DELETE';"

# And NOT returned by list
curl -s "$BASE/announcements?page=1&page_size=50" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq -r '.data.items[].id' | grep -q "$TO_DELETE" && echo "LEAK" || echo "OK"
```

Expected: DELETE returns `200`, psql shows `is_deleted | t`, final grep prints `OK`.

- [ ] **Step 4: Append DB section + commit**

Append `## 11. DB spot-check` to verification log.

```bash
git add docs/superpowers/verification/phase-07.md
git commit -m "docs(phase-07): verification log — DB spot-check + soft-delete"
```

---

## Task 20: Update README and finalize

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add Phase 7 endpoint summary**

Open `README.md`. Find the existing "API endpoints" section (created in earlier phases). Append:

```markdown
### Announcements (Phase 7)

| Method | Path                                       | Permission                    |
|--------|--------------------------------------------|-------------------------------|
| GET    | /api/v1/announcements                      | authenticated                 |
| POST   | /api/v1/announcements                      | `announcements:manage`        |
| GET    | /api/v1/announcements/{id}                 | authenticated + visibility    |
| PATCH  | /api/v1/announcements/{id}                 | `announcements:manage`        |
| DELETE | /api/v1/announcements/{id}                 | `announcements:manage`        |
| POST   | /api/v1/announcements/{id}/publish         | `announcements:manage`        |
| POST   | /api/v1/announcements/{id}/view            | authenticated                 |
| GET    | /api/v1/mobile/announcements               | authenticated                 |
| GET    | /api/v1/mobile/announcements/{id}          | authenticated                 |
| GET    | /api/v1/sse/announcements                  | authenticated (`?token=` ok)  |

The SSE endpoint streams `announcement_published` events with a 30s keep-alive comment. It is an in-process, single-replica hub — see `internal/sse/hub.go` for the scaling limit note. Tokens passed via `?token=` may end up in proxy logs; configure log scrubbing.
```

- [ ] **Step 2: Sanity build + full test suite one more time**

Run:
```bash
go build ./...
go test ./...
```
Expected: both commands exit 0; test summary shows all packages PASS.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs(phase-07): README — new endpoints + SSE scaling note"
```

---

## Spec coverage check (self-review)

- [x] Migration `000009_create_announcements.up.sql` + `.down.sql` with all five tables and audit cols — **Task 1**
- [x] `internal/models/announcement.go` with `Announcement`, `AnnouncementAttachment`, `AnnouncementView`, `AnnouncementTargetDepartment` — **Task 2**
- [x] DTOs: `AnnouncementCreate`, `AnnouncementUpdate`, `AnnouncementRead`, `AnnouncementListQuery`, `MobileAnnouncementRead`, `AnnouncementPublish` — **Task 3**
- [x] Repo interface + impl: `Create`, `Update`, `SoftDelete`, `FindByID(opts)`, `List(filter, currentUser)`, `Publish`, `MarkViewed`, `SetLabels`, `SetTargetDepartments` — **Task 5**
- [x] Service: `Create` (broadcasts when published), `Update`, `Publish`, `Delete`, `Get` (visibility + mark viewed), `List` (scope), `ListMobile`, `GetMobile`, `MarkViewed` — **Tasks 6-11**
- [x] SSE hub with `Subscribe`, `Broadcast(filter)`, buffered chan (16), sync.RWMutex, single-instance, scaling-limit doc — **Task 4**
- [x] SSE handler: `text/event-stream`, JWT via query OR header, register/unregister, `data:` frames, 30s keep-alive — **Task 12**
- [x] All 9 routes (web + mobile + SSE) with `RequirePerms(PermAnnounceManage)` on management ones, full Swagger — **Tasks 13, 14**
- [x] Hub singleton wired in `main.go` — **Task 14**
- [x] Service tests: create draft (no broadcast), publish (broadcast via mock hub), update, delete, visibility, view tracking, list filters, mobile — **Tasks 6-11**
- [x] Definition-of-Done verification log at `docs/superpowers/verification/phase-07.md`: login → create draft with labels → publish → SSE `curl -N` → publish another → list as employee → mark viewed → mobile list → non-admin create (403) → visibility denied (404) → double-publish (409) — **Tasks 16-19**

All deliverables traced to a task. No placeholders. Method names consistent (`ToRead`/`ToMobile` renamed at Task 13 with all callers updated). `PermAnnounceManage` source-of-truth flagged at Task 9 to handle either `models.*` re-export or `permissions.*` package.

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-15-phase-07-announcements-sse.md`. Two execution options:

1. **Subagent-Driven (recommended)** — dispatch one fresh subagent per task with review between tasks.
2. **Inline Execution** — execute tasks in this session with checkpoints.

Which approach?
