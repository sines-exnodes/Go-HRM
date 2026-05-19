# Phase 4 — Skills + Labels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the Python `skills` and `announcement-labels` modules to Go, exposing full CRUD for skills (with icon URL), full CRUD for labels (admin-managed), and a user↔skill assignment bridge backed by Postgres.

**Architecture:** Two new versioned SQL migrations (`000005_create_skills`, `000006_create_labels`) add `skills`, `user_skills`, and `labels` tables — all with the standard audit columns + `set_updated_at()` trigger + UUID PKs. Inside the layered Go codebase (handler → service → repository → GORM model), each entity gets a model, DTOs, repo, service, handler, and route wiring. The user↔skill bridge mirrors the Python `User.skill_ids` array as a proper join table so we can list a user's skills and reject deletion of a skill that still has employees attached. Soft-delete uses the existing `NotDeleted` GORM scope; case-insensitive name uniqueness is enforced both with a Postgres `UNIQUE` index on `LOWER(name) WHERE NOT is_deleted` and a service-layer pre-check that returns `ErrConflict`. Search uses `ILIKE` via the existing `pkg/utils/search.go` helper.

**Tech Stack:** Go 1.24, Gin, GORM, Postgres 15, golang-migrate, swaggo/swag, testify, uuid.

---

## ⚠️ REVISION NOTES (2026-05-18) — AUTHORITATIVE, read & apply before executing any task

This plan was drafted pre-schema-split and pre-Phase-0-3. The codebase audit (Python source + current Go state) supersedes the task bodies below wherever they conflict. Apply these corrections:

1. **Migration numbers.** `000001`–`000005` are taken (latest = `000005_create_departments_positions`). The two Phase-4 migrations are **`000006_create_skills`** and **`000007_create_labels`** (NOT 000005/000006 as written below). Renumber every filename, `make migrate-*` reference, and `migrate-version` expectation (final version after this phase = **7**) accordingly.

2. **Skill = catalog entity** (Python `app/models/skill.py`): fields `name TEXT NOT NULL`, `description TEXT NOT NULL DEFAULT ''`, `icon_url TEXT NULL`, + 4 audit cols + trigger + `is_deleted` index + `UNIQUE (LOWER(name)) WHERE is_deleted = FALSE`. Name validated (1–100 chars, regex `^[a-zA-Z0-9 &.+#/()-]+$`), description ≤500. Create/Update accept an **optional multipart `icon` upload** → push through the existing Phase-2 `UploadService` (Supabase S3, `Uploader` interface) and store the returned URL in `icon_url`; on update with a new icon, delete the old object. **Reuse the review-fix #2 content-sniff** (`http.DetectContentType`) — do NOT trust the client MIME header. Skill CRUD endpoints + perms exactly as Python: `GET /api/v1/skills` (`PermSkillsRead`, paginated, ILIKE search by name, sort by name ASC), `POST` (`PermSkillsCreate`, multipart), `GET /:id` (`PermSkillsRead`), `PATCH /:id` (`PermSkillsUpdate`, multipart), `DELETE /:id` (`PermSkillsDelete`).

3. **employee↔skill, NOT user↔skill.** Python's delete-guard counts employees via `User.skill_ids`. In the split schema this is an **`employee_skills`** join table: `employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE`, `skill_id UUID NOT NULL REFERENCES skills(id)`, 4 audit cols, `UNIQUE (employee_id, skill_id) WHERE is_deleted = FALSE`. Rename every `user_skill*` artifact (model `EmployeeSkill`, repo `EmployeeSkillRepository`, etc.). Assignment endpoints are nested under employees: `GET /api/v1/employees/:id/skills` (`PermEmployeesRead`), `PUT /api/v1/employees/:id/skills` (replace set — `PermEmployeesUpdate`), or `POST`/`DELETE /api/v1/employees/:id/skills/:skillID` (`PermEmployeesUpdate`) — pick PUT-replace for parity with the Python array-set semantics; document the choice. **Skill delete guard**: 409 conflict if any non-deleted `employee_skills` row references it (mirror the Phase-3 department/position guard).

4. **Label is minimal & announcement-scoped** (Python `app/models/label.py`, collection `announcement_labels`; `app/routers/labels.py`): table `labels` — `name TEXT NOT NULL` + 4 audit cols + trigger + `is_deleted` index + `UNIQUE (LOWER(name)) WHERE is_deleted = FALSE`. Name not blank, ≤50 chars. **ONLY two endpoints** (Python has no update/delete): `GET /api/v1/announcement-labels` (list ALL, sorted by name ASC, **no pagination**) and `POST /api/v1/announcement-labels` (**get-or-create** by case-insensitive name — return existing if found, else create). Delete any label Update/Delete tasks/handlers/repo-methods/tests the draft below adds — out of scope, not in Python. No cross-entity delete-guard (announcements don't exist until Phase 7).

5. **Label permission + seed gap.** Both label endpoints are gated by **`permissions.PermAnnounceManage`** (`"announcements:manage"` — already exists in `registry.go`). **AUDIT FOUND: `PermAnnounceManage` is NOT granted to any role in `seed_service.go`** (only Super Admin's `*` wildcard covers it). Add a task/step to grant `PermAnnounceManage` to the **Admin** role (and **HR Manager** if appropriate) in `seed_service.go`, idempotently, so non-superadmin admins can manage labels. No new permission constants are needed (skills + announce perms all exist).

6. **Conventions (from Phases 1-3, verified):** repos = interface + lowercase impl + `New…() Interface` constructor (see `role_repo.go`); `models.NotDeleted` scope; soft-delete sets BOTH `is_deleted=true` + `deleted_at=NOW()`; services raise `*apperrors.AppError`; per-route `middleware.RequirePerms(authSvc, permissions.PermXxx)` (**first arg `authSvc`**); route group `authed := v1.Group(""); authed.Use(middleware.JWT(...))`; search via `utils.BuildILIKEPattern` (NOT `EscapeLike`); Swagger annotations incl `@Security BearerAuth` on every endpoint; `make swag` regenerated + committed; seeder uses `s.db` directly idempotently. Tests are `package services_test`, real Postgres test DB (`TEST_DATABASE_URL='postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm_test?sslmode=disable'`), extend `truncateAll` for `employee_skills, skills, labels` in FK-safe order.

7. **DoD = real live verification** committed to `docs/superpowers/verification/phase-04.md`: server up; skill CRUD incl icon upload (valid image 200, content-spoofed rejected); assign skill to an employee; skill delete blocked (409 + SQL spot-check) while assigned then succeeds after unassign; label list + get-or-create idempotency (POST same name twice → same id); 401 (no token) / 403 (non-admin). Local DB: Postgres Docker user `ennam`/`ennam_dev_2026`, main `exnodes_hrm` (v5→v7 after this phase), test `exnodes_hrm_test`.

Everything else in the task bodies (layering, commit-per-task, no placeholders, bite-sized steps) still applies.

---

**Source references (Python — port from):**
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/routers/skills.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/services/skill.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/schemas/skill.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/models/skill.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/routers/labels.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/services/label.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/schemas/label.py`
- `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api/app/models/label.py`

**Permission verification (read from Python):**
- Skills: `SKILLS_READ` / `SKILLS_CREATE` / `SKILLS_UPDATE` / `SKILLS_DELETE` (1:1 with `PermSkillsRead/Create/Update/Delete` already in registry from Phase 1).
- Labels: Python `labels.py` gates **both** list and create with `Permission.ANNOUNCEMENTS_MANAGE`. We map this to `PermAnnounceManage` for all label endpoints (list, get, create, update, delete). No "auth-only" exception — admin permission is required for every label endpoint.
- User-skill assignment: Python attaches skills via `User.skill_ids`, so the natural permission is `PermUsersUpdate` for write and `PermUsersRead` for read.

**Behavior preserved from Python:**
- Skill name validated against regex `^[a-zA-Z0-9 &.+#/()-]+$`, trimmed, 1–100 chars.
- Skill description trimmed, ≤500 chars, default `""`.
- Skill name uniqueness is **case-insensitive**.
- Skill list is sorted alphabetically by `name`.
- Skill delete is **blocked** while at least one user is still assigned; the response includes `employee_count` details.
- Label `name` is 1–50 chars; POST is **get-or-create** semantics (case-insensitive lookup, create only if absent).
- Label list is sorted alphabetically by `name`.

**Behavior added (greenfield, not in Python):**
- Skill icon upload is **out of scope** for Phase 4 — the field exists as a free-form `icon_url TEXT` on the model so callers can set it post-hoc, but no multipart endpoint is added. Phase 2 already established the avatar upload pattern; a future task can extend it.
- Labels gain `GET /:id`, `PATCH /:id`, `DELETE /:id` to reach full CRUD as required by the phase scope. Color column added (TEXT NULL) for FE convenience.
- Explicit `POST/DELETE /api/v1/users/:userID/skills` endpoints replace Python's "edit `User.skill_ids` via user update" implicit pattern.

---

## File Map

**New migrations:**
- `migrations/000005_create_skills.up.sql`
- `migrations/000005_create_skills.down.sql`
- `migrations/000006_create_labels.up.sql`
- `migrations/000006_create_labels.down.sql`

**New models:**
- `internal/models/skill.go`
- `internal/models/user_skill.go`
- `internal/models/label.go`

**New DTOs:**
- `internal/dto/skill.go`
- `internal/dto/label.go`

**New repositories:**
- `internal/repositories/skill_repo.go`
- `internal/repositories/user_skill_repo.go`
- `internal/repositories/label_repo.go`

**New services:**
- `internal/services/skill_service.go`
- `internal/services/skill_service_test.go`
- `internal/services/label_service.go`
- `internal/services/label_service_test.go`

**New handlers:**
- `internal/handlers/skill_handler.go`
- `internal/handlers/label_handler.go`

**Modified:**
- `cmd/server/main.go` — wire skill + label repos, services, handlers, routes.

**New docs:**
- `docs/superpowers/verification/phase-04.md` — verification log.

---

## Task 1: Migration `000005_create_skills`

**Files:**
- Create: `migrations/000005_create_skills.up.sql`
- Create: `migrations/000005_create_skills.down.sql`

- [ ] **Step 1: Write `000005_create_skills.up.sql`**

```sql
-- ============================================================
-- Phase 4: skills + user_skills
-- ============================================================

CREATE TABLE IF NOT EXISTS skills (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT         NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    icon_url    TEXT         NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ  NULL
);

-- Case-insensitive unique name among live rows.
CREATE UNIQUE INDEX IF NOT EXISTS uq_skills_name_lower_live
    ON skills (LOWER(name))
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_skills_is_deleted ON skills (is_deleted);
CREATE INDEX IF NOT EXISTS idx_skills_name       ON skills (name);

CREATE TRIGGER trg_skills_set_updated_at
    BEFORE UPDATE ON skills
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ------------------------------------------------------------
-- user_skills join (replaces Python User.skill_ids array)
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_skills (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    skill_id    UUID         NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_skills_pair_live
    ON user_skills (user_id, skill_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_user_skills_user_id    ON user_skills (user_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_skill_id   ON user_skills (skill_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_is_deleted ON user_skills (is_deleted);

CREATE TRIGGER trg_user_skills_set_updated_at
    BEFORE UPDATE ON user_skills
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 2: Write `000005_create_skills.down.sql`**

```sql
DROP TRIGGER IF EXISTS trg_user_skills_set_updated_at ON user_skills;
DROP TABLE IF EXISTS user_skills;

DROP TRIGGER IF EXISTS trg_skills_set_updated_at ON skills;
DROP TABLE IF EXISTS skills;
```

- [ ] **Step 3: Apply migration on local Postgres**

Run: `make migrate-up`
Expected output (last line): `migration done: ...000005_create_skills`

- [ ] **Step 4: Spot-check schema**

Run:
```bash
psql "$DATABASE_URL" -c "\d skills" -c "\d user_skills"
```
Expected: both tables listed; columns `id, name, description, icon_url, created_at, updated_at, is_deleted, deleted_at` on `skills`; trigger `trg_skills_set_updated_at` present; unique index `uq_skills_name_lower_live` present; FK from `user_skills.user_id` → `users.id` and `user_skills.skill_id` → `skills.id`.

- [ ] **Step 5: Rollback round-trip check**

Run:
```bash
make migrate-down
psql "$DATABASE_URL" -c "\dt skills" -c "\dt user_skills"
make migrate-up
```
Expected: after `migrate-down`, `\dt` reports `Did not find any relation named "skills"` and `"user_skills"`. After re-applying `migrate-up`, both tables exist again.

- [ ] **Step 6: Commit**

```bash
git add migrations/000005_create_skills.up.sql migrations/000005_create_skills.down.sql
git commit -m "feat(phase-4): add skills + user_skills migration"
```

---

## Task 2: Migration `000006_create_labels`

**Files:**
- Create: `migrations/000006_create_labels.up.sql`
- Create: `migrations/000006_create_labels.down.sql`

- [ ] **Step 1: Write `000006_create_labels.up.sql`**

```sql
-- ============================================================
-- Phase 4: labels (announcement labels — usage joins arrive in Phase 7)
-- ============================================================

CREATE TABLE IF NOT EXISTS labels (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT         NOT NULL,
    color       TEXT         NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_labels_name_lower_live
    ON labels (LOWER(name))
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_labels_is_deleted ON labels (is_deleted);
CREATE INDEX IF NOT EXISTS idx_labels_name       ON labels (name);

CREATE TRIGGER trg_labels_set_updated_at
    BEFORE UPDATE ON labels
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 2: Write `000006_create_labels.down.sql`**

```sql
DROP TRIGGER IF EXISTS trg_labels_set_updated_at ON labels;
DROP TABLE IF EXISTS labels;
```

- [ ] **Step 3: Apply and spot-check**

Run:
```bash
make migrate-up
psql "$DATABASE_URL" -c "\d labels"
```
Expected: `labels` table with `id, name, color, created_at, updated_at, is_deleted, deleted_at`; trigger present; unique partial index `uq_labels_name_lower_live` present.

- [ ] **Step 4: Rollback round-trip check**

Run:
```bash
make migrate-down
psql "$DATABASE_URL" -c "\dt labels"
make migrate-up
```
Expected: drop, then re-create succeeds.

- [ ] **Step 5: Commit**

```bash
git add migrations/000006_create_labels.up.sql migrations/000006_create_labels.down.sql
git commit -m "feat(phase-4): add labels migration"
```

---

## Task 3: Models — `skill.go`, `user_skill.go`, `label.go`

**Files:**
- Create: `internal/models/skill.go`
- Create: `internal/models/user_skill.go`
- Create: `internal/models/label.go`

- [ ] **Step 1: Write `internal/models/skill.go`**

```go
package models

// Skill is a named competency that can be attached to users.
// Mirrors Python app/models/skill.py with audit columns + UUID PK.
type Skill struct {
	BaseModel
	Name        string  `gorm:"type:text;not null"          json:"name"`
	Description string  `gorm:"type:text;not null;default:''" json:"description"`
	IconURL     *string `gorm:"type:text"                   json:"icon_url,omitempty"`
}

// TableName overrides the default plural derived by GORM
// (we keep the explicit name to stay aligned with the migration file).
func (Skill) TableName() string { return "skills" }
```

- [ ] **Step 2: Write `internal/models/user_skill.go`**

```go
package models

import "github.com/google/uuid"

// UserSkill is the join row that attaches a Skill to a User.
// Replaces the Python User.skill_ids array. Own UUID PK + partial-unique
// index on (user_id, skill_id) WHERE is_deleted = FALSE keeps soft-delete safe.
type UserSkill struct {
	BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	SkillID uuid.UUID `gorm:"type:uuid;not null;index" json:"skill_id"`
}

func (UserSkill) TableName() string { return "user_skills" }
```

- [ ] **Step 3: Write `internal/models/label.go`**

```go
package models

// Label is a freeform tag attached to announcements (Phase 7) and other
// future resources. Sourced from Python app/models/label.py with an added
// `color` field for FE picker convenience.
type Label struct {
	BaseModel
	Name  string  `gorm:"type:text;not null" json:"name"`
	Color *string `gorm:"type:text"          json:"color,omitempty"`
}

func (Label) TableName() string { return "labels" }
```

- [ ] **Step 4: Verify build**

Run: `go build ./internal/models/...`
Expected: exit code 0, no output.

- [ ] **Step 5: Commit**

```bash
git add internal/models/skill.go internal/models/user_skill.go internal/models/label.go
git commit -m "feat(phase-4): add skill, user_skill, label models"
```

---

## Task 4: DTOs — `skill.go` and `label.go`

**Files:**
- Create: `internal/dto/skill.go`
- Create: `internal/dto/label.go`

- [ ] **Step 1: Write `internal/dto/skill.go`**

```go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// SkillCreateRequest mirrors Python SkillCreate (schemas/skill.py).
// Name validation regex is enforced in the service layer to keep DTOs
// declarative; binding-level validation handles only length + required.
type SkillCreateRequest struct {
	Name        string  `json:"name"        binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"max=500"`
	IconURL     *string `json:"icon_url"    binding:"omitempty,max=2048"`
}

// SkillUpdateRequest — every field optional (PATCH semantics).
// Pointers distinguish "not provided" from "explicitly empty".
type SkillUpdateRequest struct {
	Name        *string `json:"name"        binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IconURL     *string `json:"icon_url"    binding:"omitempty,max=2048"`
}

// SkillRead is the JSON response shape.
type SkillRead struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     *string   `json:"icon_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SkillListQuery binds the `?page=&page_size=&search=` query string.
type SkillListQuery struct {
	Page     int    `form:"page,default=1"        binding:"omitempty,min=1"`
	PageSize int    `form:"page_size,default=10"  binding:"omitempty,min=1,max=100"`
	Search   string `form:"search"                binding:"omitempty,max=100"`
}

// UserSkillAssignRequest accepts a single skill_id to attach to a user.
type UserSkillAssignRequest struct {
	SkillID uuid.UUID `json:"skill_id" binding:"required"`
}
```

- [ ] **Step 2: Write `internal/dto/label.go`**

```go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// LabelCreateRequest mirrors Python LabelCreate (schemas/label.py).
type LabelCreateRequest struct {
	Name  string  `json:"name"  binding:"required,min=1,max=50"`
	Color *string `json:"color" binding:"omitempty,max=20"`
}

type LabelUpdateRequest struct {
	Name  *string `json:"name"  binding:"omitempty,min=1,max=50"`
	Color *string `json:"color" binding:"omitempty,max=20"`
}

type LabelRead struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LabelListQuery — labels are short lists, default to 100/page.
type LabelListQuery struct {
	Page     int    `form:"page,default=1"        binding:"omitempty,min=1"`
	PageSize int    `form:"page_size,default=100" binding:"omitempty,min=1,max=200"`
	Search   string `form:"search"                binding:"omitempty,max=100"`
}
```

- [ ] **Step 3: Verify build**

Run: `go build ./internal/dto/...`
Expected: exit code 0.

- [ ] **Step 4: Commit**

```bash
git add internal/dto/skill.go internal/dto/label.go
git commit -m "feat(phase-4): add skill + label DTOs"
```

---

## Task 5: Repository — `skill_repo.go`

**Files:**
- Create: `internal/repositories/skill_repo.go`

- [ ] **Step 1: Write `internal/repositories/skill_repo.go`**

```go
package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// SkillRepository abstracts skill persistence for the service layer.
type SkillRepository interface {
	Create(ctx context.Context, skill *models.Skill) error
	Update(ctx context.Context, skill *models.Skill) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Skill, error)
	FindByNameCI(ctx context.Context, name string) (*models.Skill, error)
	List(ctx context.Context, search string, page, pageSize int) ([]models.Skill, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type skillRepo struct{ db *gorm.DB }

// NewSkillRepository wires the GORM implementation.
func NewSkillRepository(db *gorm.DB) SkillRepository { return &skillRepo{db: db} }

func (r *skillRepo) Create(ctx context.Context, skill *models.Skill) error {
	return r.db.WithContext(ctx).Create(skill).Error
}

func (r *skillRepo) Update(ctx context.Context, skill *models.Skill) error {
	return r.db.WithContext(ctx).Save(skill).Error
}

func (r *skillRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Skill, error) {
	var s models.Skill
	if err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skillRepo) FindByNameCI(ctx context.Context, name string) (*models.Skill, error) {
	var s models.Skill
	if err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skillRepo) List(ctx context.Context, search string, page, pageSize int) ([]models.Skill, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Skill{}).Scopes(models.NotDeleted)

	if s := strings.TrimSpace(search); s != "" {
		q = q.Where("name ILIKE ?", "%"+utils.EscapeLike(s)+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.Skill
	offset := (page - 1) * pageSize
	if err := q.Order("name ASC").Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *skillRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.Skill{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/repositories/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/repositories/skill_repo.go
git commit -m "feat(phase-4): add skill repository"
```

---

## Task 6: Repository — `user_skill_repo.go`

**Files:**
- Create: `internal/repositories/user_skill_repo.go`

- [ ] **Step 1: Write `internal/repositories/user_skill_repo.go`**

```go
package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// UserSkillRepository owns the user↔skill join table.
type UserSkillRepository interface {
	Assign(ctx context.Context, userID, skillID uuid.UUID) (*models.UserSkill, error)
	Unassign(ctx context.Context, userID, skillID uuid.UUID) error
	ListSkillIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ListSkillsByUser(ctx context.Context, userID uuid.UUID) ([]models.Skill, error)
	CountUsersBySkill(ctx context.Context, skillID uuid.UUID) (int64, error)
	Exists(ctx context.Context, userID, skillID uuid.UUID) (bool, error)
}

type userSkillRepo struct{ db *gorm.DB }

func NewUserSkillRepository(db *gorm.DB) UserSkillRepository { return &userSkillRepo{db: db} }

func (r *userSkillRepo) Assign(ctx context.Context, userID, skillID uuid.UUID) (*models.UserSkill, error) {
	// Try to revive any soft-deleted row first so we don't violate the partial unique index.
	var existing models.UserSkill
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND skill_id = ?", userID, skillID).
		First(&existing).Error
	switch err {
	case nil:
		if !existing.IsDeleted {
			return &existing, nil
		}
		existing.IsDeleted = false
		existing.DeletedAt = nil
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	case gorm.ErrRecordNotFound:
		row := &models.UserSkill{UserID: userID, SkillID: skillID}
		if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
			return nil, err
		}
		return row, nil
	default:
		return nil, err
	}
}

func (r *userSkillRepo) Unassign(ctx context.Context, userID, skillID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.UserSkill{}).
		Where("user_id = ? AND skill_id = ? AND is_deleted = ?", userID, skillID, false).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *userSkillRepo) ListSkillIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&models.UserSkill{}).
		Scopes(models.NotDeleted).
		Where("user_id = ?", userID).
		Pluck("skill_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *userSkillRepo) ListSkillsByUser(ctx context.Context, userID uuid.UUID) ([]models.Skill, error) {
	var skills []models.Skill
	err := r.db.WithContext(ctx).
		Table("skills AS s").
		Joins("INNER JOIN user_skills AS us ON us.skill_id = s.id AND us.is_deleted = FALSE").
		Where("us.user_id = ? AND s.is_deleted = FALSE", userID).
		Order("s.name ASC").
		Find(&skills).Error
	return skills, err
}

func (r *userSkillRepo) CountUsersBySkill(ctx context.Context, skillID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.UserSkill{}).
		Scopes(models.NotDeleted).
		Where("skill_id = ?", skillID).
		Count(&n).Error
	return n, err
}

func (r *userSkillRepo) Exists(ctx context.Context, userID, skillID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.UserSkill{}).
		Scopes(models.NotDeleted).
		Where("user_id = ? AND skill_id = ?", userID, skillID).
		Count(&n).Error
	return n > 0, err
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/repositories/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/repositories/user_skill_repo.go
git commit -m "feat(phase-4): add user_skill repository"
```

---

## Task 7: Repository — `label_repo.go`

**Files:**
- Create: `internal/repositories/label_repo.go`

- [ ] **Step 1: Write `internal/repositories/label_repo.go`**

```go
package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// LabelRepository abstracts label persistence.
type LabelRepository interface {
	Create(ctx context.Context, label *models.Label) error
	Update(ctx context.Context, label *models.Label) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error)
	FindByNameCI(ctx context.Context, name string) (*models.Label, error)
	List(ctx context.Context, search string, page, pageSize int) ([]models.Label, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type labelRepo struct{ db *gorm.DB }

func NewLabelRepository(db *gorm.DB) LabelRepository { return &labelRepo{db: db} }

func (r *labelRepo) Create(ctx context.Context, label *models.Label) error {
	return r.db.WithContext(ctx).Create(label).Error
}

func (r *labelRepo) Update(ctx context.Context, label *models.Label) error {
	return r.db.WithContext(ctx).Save(label).Error
}

func (r *labelRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error) {
	var l models.Label
	if err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *labelRepo) FindByNameCI(ctx context.Context, name string) (*models.Label, error) {
	var l models.Label
	if err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *labelRepo) List(ctx context.Context, search string, page, pageSize int) ([]models.Label, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Label{}).Scopes(models.NotDeleted)

	if s := strings.TrimSpace(search); s != "" {
		q = q.Where("name ILIKE ?", "%"+utils.EscapeLike(s)+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.Label
	offset := (page - 1) * pageSize
	if err := q.Order("name ASC").Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *labelRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.Label{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/repositories/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/repositories/label_repo.go
git commit -m "feat(phase-4): add label repository"
```

---

## Task 8: Service — `skill_service.go`

**Files:**
- Create: `internal/services/skill_service.go`

- [ ] **Step 1: Write `internal/services/skill_service.go`**

```go
package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// skillNamePattern ports the regex from Python schemas/skill.py:
// letters, digits, spaces, & . + # / ( ) - .
var skillNamePattern = regexp.MustCompile(`^[a-zA-Z0-9 &.+#/()\-]+$`)

// SkillService is the business façade used by handlers.
type SkillService interface {
	Create(ctx context.Context, req dto.SkillCreateRequest) (*dto.SkillRead, error)
	Update(ctx context.Context, id uuid.UUID, req dto.SkillUpdateRequest) (*dto.SkillRead, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.SkillRead, error)
	List(ctx context.Context, q dto.SkillListQuery) ([]dto.SkillRead, int64, error)
	Delete(ctx context.Context, id uuid.UUID) (*dto.SkillRead, error)

	AssignToUser(ctx context.Context, userID, skillID uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID, skillID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]dto.SkillRead, error)
}

type skillService struct {
	skills repositories.SkillRepository
	links  repositories.UserSkillRepository
}

// NewSkillService wires the service to its repositories.
func NewSkillService(skills repositories.SkillRepository, links repositories.UserSkillRepository) SkillService {
	return &skillService{skills: skills, links: links}
}

// --- helpers ---------------------------------------------------------------

func validateSkillName(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", apperr.ErrBadRequest("Skill name is required")
	}
	if len(v) > 100 {
		return "", apperr.ErrBadRequest("Skill name must not exceed 100 characters")
	}
	if !skillNamePattern.MatchString(v) {
		return "", apperr.ErrBadRequest("Skill name can only contain letters, numbers, spaces, hyphens, ampersands, dots, plus signs, hash signs, slashes, and parentheses")
	}
	return v, nil
}

func toSkillRead(s *models.Skill) *dto.SkillRead {
	return &dto.SkillRead{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		IconURL:     s.IconURL,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// --- CRUD ------------------------------------------------------------------

func (s *skillService) Create(ctx context.Context, req dto.SkillCreateRequest) (*dto.SkillRead, error) {
	name, err := validateSkillName(req.Name)
	if err != nil {
		return nil, err
	}
	desc := strings.TrimSpace(req.Description)

	if existing, err := s.skills.FindByNameCI(ctx, name); err == nil && existing != nil {
		return nil, apperr.ErrConflict("Skill name already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	row := &models.Skill{
		Name:        name,
		Description: desc,
		IconURL:     req.IconURL,
	}
	if err := s.skills.Create(ctx, row); err != nil {
		return nil, err
	}
	return toSkillRead(row), nil
}

func (s *skillService) Update(ctx context.Context, id uuid.UUID, req dto.SkillUpdateRequest) (*dto.SkillRead, error) {
	row, err := s.skills.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound("Skill")
		}
		return nil, err
	}

	if req.Name != nil {
		name, vErr := validateSkillName(*req.Name)
		if vErr != nil {
			return nil, vErr
		}
		if !strings.EqualFold(name, row.Name) {
			if existing, qErr := s.skills.FindByNameCI(ctx, name); qErr == nil && existing != nil && existing.ID != row.ID {
				return nil, apperr.ErrConflict("Skill name already exists")
			} else if qErr != nil && !errors.Is(qErr, gorm.ErrRecordNotFound) {
				return nil, qErr
			}
		}
		row.Name = name
	}
	if req.Description != nil {
		row.Description = strings.TrimSpace(*req.Description)
		if len(row.Description) > 500 {
			return nil, apperr.ErrBadRequest("Description must not exceed 500 characters")
		}
	}
	if req.IconURL != nil {
		row.IconURL = req.IconURL
	}

	if err := s.skills.Update(ctx, row); err != nil {
		return nil, err
	}
	return toSkillRead(row), nil
}

func (s *skillService) Get(ctx context.Context, id uuid.UUID) (*dto.SkillRead, error) {
	row, err := s.skills.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound("Skill")
		}
		return nil, err
	}
	return toSkillRead(row), nil
}

func (s *skillService) List(ctx context.Context, q dto.SkillListQuery) ([]dto.SkillRead, int64, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	rows, total, err := s.skills.List(ctx, q.Search, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.SkillRead, 0, len(rows))
	for i := range rows {
		out = append(out, *toSkillRead(&rows[i]))
	}
	return out, total, nil
}

func (s *skillService) Delete(ctx context.Context, id uuid.UUID) (*dto.SkillRead, error) {
	row, err := s.skills.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound("Skill")
		}
		return nil, err
	}

	count, err := s.links.CountUsersBySkill(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		e := apperr.ErrBadRequest("Cannot delete skill '" + row.Name + "'. " +
			"Employees are assigned to this skill. Reassign all employees before deleting this skill.")
		e.Details = map[string]any{"employee_count": count}
		return nil, e
	}

	if err := s.skills.SoftDelete(ctx, row.ID); err != nil {
		return nil, err
	}
	return toSkillRead(row), nil
}

// --- user-skill bridge -----------------------------------------------------

func (s *skillService) AssignToUser(ctx context.Context, userID, skillID uuid.UUID) error {
	// Ensure target skill exists (and isn't soft-deleted).
	if _, err := s.skills.GetByID(ctx, skillID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound("Skill")
		}
		return err
	}
	_, err := s.links.Assign(ctx, userID, skillID)
	return err
}

func (s *skillService) RemoveFromUser(ctx context.Context, userID, skillID uuid.UUID) error {
	exists, err := s.links.Exists(ctx, userID, skillID)
	if err != nil {
		return err
	}
	if !exists {
		return apperr.ErrNotFound("User skill")
	}
	return s.links.Unassign(ctx, userID, skillID)
}

func (s *skillService) ListByUser(ctx context.Context, userID uuid.UUID) ([]dto.SkillRead, error) {
	rows, err := s.links.ListSkillsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SkillRead, 0, len(rows))
	for i := range rows {
		out = append(out, *toSkillRead(&rows[i]))
	}
	return out, nil
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/services/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/services/skill_service.go
git commit -m "feat(phase-4): add skill service with user-skill bridge"
```

---

## Task 9: Service — `label_service.go`

**Files:**
- Create: `internal/services/label_service.go`

- [ ] **Step 1: Write `internal/services/label_service.go`**

```go
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// LabelService exposes label CRUD with get-or-create on POST.
type LabelService interface {
	GetOrCreate(ctx context.Context, req dto.LabelCreateRequest) (*dto.LabelRead, error)
	Update(ctx context.Context, id uuid.UUID, req dto.LabelUpdateRequest) (*dto.LabelRead, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.LabelRead, error)
	List(ctx context.Context, q dto.LabelListQuery) ([]dto.LabelRead, int64, error)
	Delete(ctx context.Context, id uuid.UUID) (*dto.LabelRead, error)
}

type labelService struct {
	labels repositories.LabelRepository
}

func NewLabelService(labels repositories.LabelRepository) LabelService {
	return &labelService{labels: labels}
}

func toLabelRead(l *models.Label) *dto.LabelRead {
	return &dto.LabelRead{
		ID:        l.ID,
		Name:      l.Name,
		Color:     l.Color,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func validateLabelName(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", apperr.ErrBadRequest("Label name cannot be blank")
	}
	if len(v) > 50 {
		return "", apperr.ErrBadRequest("Label name must not exceed 50 characters")
	}
	return v, nil
}

func (s *labelService) GetOrCreate(ctx context.Context, req dto.LabelCreateRequest) (*dto.LabelRead, error) {
	name, err := validateLabelName(req.Name)
	if err != nil {
		return nil, err
	}

	// Case-insensitive lookup → return existing (Python parity).
	existing, err := s.labels.FindByNameCI(ctx, name)
	if err == nil && existing != nil {
		// If caller sent a color and existing has none, leave it as-is to preserve get-or-create idempotency.
		return toLabelRead(existing), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	row := &models.Label{Name: name, Color: req.Color}
	if err := s.labels.Create(ctx, row); err != nil {
		return nil, err
	}
	return toLabelRead(row), nil
}

func (s *labelService) Update(ctx context.Context, id uuid.UUID, req dto.LabelUpdateRequest) (*dto.LabelRead, error) {
	row, err := s.labels.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound("Label")
		}
		return nil, err
	}

	if req.Name != nil {
		name, vErr := validateLabelName(*req.Name)
		if vErr != nil {
			return nil, vErr
		}
		if !strings.EqualFold(name, row.Name) {
			if other, qErr := s.labels.FindByNameCI(ctx, name); qErr == nil && other != nil && other.ID != row.ID {
				return nil, apperr.ErrConflict("Label name already exists")
			} else if qErr != nil && !errors.Is(qErr, gorm.ErrRecordNotFound) {
				return nil, qErr
			}
		}
		row.Name = name
	}
	if req.Color != nil {
		row.Color = req.Color
	}

	if err := s.labels.Update(ctx, row); err != nil {
		return nil, err
	}
	return toLabelRead(row), nil
}

func (s *labelService) Get(ctx context.Context, id uuid.UUID) (*dto.LabelRead, error) {
	row, err := s.labels.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound("Label")
		}
		return nil, err
	}
	return toLabelRead(row), nil
}

func (s *labelService) List(ctx context.Context, q dto.LabelListQuery) ([]dto.LabelRead, int64, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 100
	}

	rows, total, err := s.labels.List(ctx, q.Search, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.LabelRead, 0, len(rows))
	for i := range rows {
		out = append(out, *toLabelRead(&rows[i]))
	}
	return out, total, nil
}

func (s *labelService) Delete(ctx context.Context, id uuid.UUID) (*dto.LabelRead, error) {
	row, err := s.labels.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound("Label")
		}
		return nil, err
	}
	if err := s.labels.SoftDelete(ctx, row.ID); err != nil {
		return nil, err
	}
	return toLabelRead(row), nil
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/services/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/services/label_service.go
git commit -m "feat(phase-4): add label service"
```

---

## Task 10: Service tests — `skill_service_test.go`

**Files:**
- Create: `internal/services/skill_service_test.go`

- [ ] **Step 1: Write `internal/services/skill_service_test.go`**

```go
package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// newSkillService builds the service against the shared test DB.
// testDB() and resetTables() are provided by services/testhelper_test.go (Phase 1).
func newSkillService(t *testing.T) (SkillService, repositories.SkillRepository, repositories.UserSkillRepository) {
	t.Helper()
	db := testDB(t)
	resetTables(t, db, "user_skills", "skills")
	skills := repositories.NewSkillRepository(db)
	links := repositories.NewUserSkillRepository(db)
	return NewSkillService(skills, links), skills, links
}

func TestSkillService_Create_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	out, err := svc.Create(ctx, dto.SkillCreateRequest{
		Name:        "  Go Programming  ",
		Description: "  Server-side dev  ",
	})
	require.NoError(t, err)
	assert.Equal(t, "Go Programming", out.Name)
	assert.Equal(t, "Server-side dev", out.Description)
	assert.NotEqual(t, uuid.Nil, out.ID)
}

func TestSkillService_Create_DuplicateName_CaseInsensitive(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	_, err := svc.Create(ctx, dto.SkillCreateRequest{Name: "React"})
	require.NoError(t, err)

	_, err = svc.Create(ctx, dto.SkillCreateRequest{Name: "react"})
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "conflict", appErr.Code)
}

func TestSkillService_Create_InvalidCharacters(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	_, err := svc.Create(ctx, dto.SkillCreateRequest{Name: "Bad@Skill!"})
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "bad_request", appErr.Code)
}

func TestSkillService_Update_RenameAndDescription(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	created, err := svc.Create(ctx, dto.SkillCreateRequest{Name: "TypeScript", Description: "old"})
	require.NoError(t, err)

	newName := "TypeScript 5"
	newDesc := "new"
	out, err := svc.Update(ctx, created.ID, dto.SkillUpdateRequest{Name: &newName, Description: &newDesc})
	require.NoError(t, err)
	assert.Equal(t, "TypeScript 5", out.Name)
	assert.Equal(t, "new", out.Description)
}

func TestSkillService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	name := "Anything"
	_, err := svc.Update(ctx, uuid.New(), dto.SkillUpdateRequest{Name: &name})
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "not_found", appErr.Code)
}

func TestSkillService_List_SearchAndSort(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	for _, n := range []string{"React", "Redux", "Go"} {
		_, err := svc.Create(ctx, dto.SkillCreateRequest{Name: n})
		require.NoError(t, err)
	}

	all, total, err := svc.List(ctx, dto.SkillListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, all, 3)
	// Sorted alphabetically: Go, React, Redux
	assert.Equal(t, "Go", all[0].Name)
	assert.Equal(t, "React", all[1].Name)
	assert.Equal(t, "Redux", all[2].Name)

	// ILIKE search
	hits, total, err := svc.List(ctx, dto.SkillListQuery{Page: 1, PageSize: 10, Search: "re"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, hits, 2)
	assert.Equal(t, "React", hits[0].Name)
	assert.Equal(t, "Redux", hits[1].Name)
}

func TestSkillService_Delete_SoftDelete(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	created, err := svc.Create(ctx, dto.SkillCreateRequest{Name: "Kotlin"})
	require.NoError(t, err)

	_, err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.Get(ctx, created.ID)
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "not_found", appErr.Code)
}

func TestSkillService_Delete_BlockedWhenAssigned(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	created, err := svc.Create(ctx, dto.SkillCreateRequest{Name: "Rust"})
	require.NoError(t, err)

	user := makeUser(t) // helper from testhelper_test.go (Phase 1/2)
	require.NoError(t, svc.AssignToUser(ctx, user.ID, created.ID))

	_, err = svc.Delete(ctx, created.ID)
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "bad_request", appErr.Code)
	assert.EqualValues(t, 1, appErr.Details["employee_count"])
}

func TestSkillService_AssignUnassignList(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)

	skill, err := svc.Create(ctx, dto.SkillCreateRequest{Name: "Python"})
	require.NoError(t, err)
	user := makeUser(t)

	// Assign + idempotent re-assign.
	require.NoError(t, svc.AssignToUser(ctx, user.ID, skill.ID))
	require.NoError(t, svc.AssignToUser(ctx, user.ID, skill.ID))

	list, err := svc.ListByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Python", list[0].Name)

	// Unassign.
	require.NoError(t, svc.RemoveFromUser(ctx, user.ID, skill.ID))
	list, err = svc.ListByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Empty(t, list)

	// Re-assign after unassign (soft-delete revive path).
	require.NoError(t, svc.AssignToUser(ctx, user.ID, skill.ID))
	list, err = svc.ListByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestSkillService_AssignToUser_SkillNotFound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newSkillService(t)
	user := makeUser(t)
	err := svc.AssignToUser(ctx, user.ID, uuid.New())
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "not_found", appErr.Code)
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/services/ -run TestSkillService -v`
Expected: all 9 sub-tests pass (`--- PASS:` per test), final `ok github.com/exnodes/hrm-api/internal/services`.

- [ ] **Step 3: Commit**

```bash
git add internal/services/skill_service_test.go
git commit -m "test(phase-4): add skill service tests"
```

---

## Task 11: Service tests — `label_service_test.go`

**Files:**
- Create: `internal/services/label_service_test.go`

- [ ] **Step 1: Write `internal/services/label_service_test.go`**

```go
package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

func newLabelService(t *testing.T) LabelService {
	t.Helper()
	db := testDB(t)
	resetTables(t, db, "labels")
	return NewLabelService(repositories.NewLabelRepository(db))
}

func TestLabelService_GetOrCreate_NewLabel(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	red := "#ff0000"
	out, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "  Urgent  ", Color: &red})
	require.NoError(t, err)
	assert.Equal(t, "Urgent", out.Name)
	require.NotNil(t, out.Color)
	assert.Equal(t, "#ff0000", *out.Color)
}

func TestLabelService_GetOrCreate_ReturnsExistingCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	first, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "Important"})
	require.NoError(t, err)
	second, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "important"})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
}

func TestLabelService_GetOrCreate_BlankName(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	_, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "   "})
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "bad_request", appErr.Code)
}

func TestLabelService_Update_Rename(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	created, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "Beta"})
	require.NoError(t, err)

	newName := "Gamma"
	out, err := svc.Update(ctx, created.ID, dto.LabelUpdateRequest{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Gamma", out.Name)
}

func TestLabelService_Update_DuplicateName(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	_, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "Alpha"})
	require.NoError(t, err)
	bravo, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "Bravo"})
	require.NoError(t, err)

	clash := "Alpha"
	_, err = svc.Update(ctx, bravo.ID, dto.LabelUpdateRequest{Name: &clash})
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "conflict", appErr.Code)
}

func TestLabelService_List_SortedAndSearch(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	for _, n := range []string{"Zeta", "Alpha", "Beta"} {
		_, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: n})
		require.NoError(t, err)
	}

	all, total, err := svc.List(ctx, dto.LabelListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, all, 3)
	assert.Equal(t, "Alpha", all[0].Name)
	assert.Equal(t, "Beta", all[1].Name)
	assert.Equal(t, "Zeta", all[2].Name)

	hits, total, err := svc.List(ctx, dto.LabelListQuery{Page: 1, PageSize: 50, Search: "et"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, hits, 2)
	assert.Equal(t, "Beta", hits[0].Name)
	assert.Equal(t, "Zeta", hits[1].Name)
}

func TestLabelService_Delete_SoftDelete(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	created, err := svc.GetOrCreate(ctx, dto.LabelCreateRequest{Name: "Trash"})
	require.NoError(t, err)

	_, err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.Get(ctx, created.ID)
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "not_found", appErr.Code)
}

func TestLabelService_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := newLabelService(t)

	_, err := svc.Get(ctx, uuid.New())
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "not_found", appErr.Code)
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/services/ -run TestLabelService -v`
Expected: all 8 sub-tests pass, final `ok github.com/exnodes/hrm-api/internal/services`.

- [ ] **Step 3: Run the full service test suite**

Run: `go test ./internal/services/...`
Expected: `ok github.com/exnodes/hrm-api/internal/services <time>` — no regressions in Phase 1–3 suites.

- [ ] **Step 4: Commit**

```bash
git add internal/services/label_service_test.go
git commit -m "test(phase-4): add label service tests"
```

---

## Task 12: Handler — `skill_handler.go`

**Files:**
- Create: `internal/handlers/skill_handler.go`

- [ ] **Step 1: Write `internal/handlers/skill_handler.go`**

```go
package handlers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// SkillHandler exposes /api/v1/skills and the user-skill bridge.
type SkillHandler struct {
	svc services.SkillService
}

func NewSkillHandler(svc services.SkillService) *SkillHandler { return &SkillHandler{svc: svc} }

// ListSkills godoc
// @Summary      List skills
// @Description  Paginated list of skills sorted alphabetically; optional ILIKE search on name.
// @Tags         Skills
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int    false  "Page (>=1)"           default(1)
// @Param        page_size  query  int    false  "Page size (1..100)"   default(10)
// @Param        search     query  string false  "Search by skill name"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.SkillRead]]
// @Failure      401  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Router       /skills [get]
func (h *SkillHandler) List(c *gin.Context) {
	var q dto.SkillListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.SkillRead]]{
		Success: true,
		Data: dto.PaginatedData[dto.SkillRead]{
			Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
		},
	})
}

// CreateSkill godoc
// @Summary      Create a skill
// @Tags         Skills
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.SkillCreateRequest  true  "Skill payload"
// @Success      201  {object}  dto.Response[dto.SkillRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /skills [post]
func (h *SkillHandler) Create(c *gin.Context) {
	var req dto.SkillCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[dto.SkillRead]{
		Success: true,
		Message: "Skill '" + out.Name + "' has been created",
		Data:    *out,
	})
}

// GetSkill godoc
// @Summary      Get a skill by ID
// @Tags         Skills
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Skill ID (UUID)"
// @Success      200  {object}  dto.Response[dto.SkillRead]
// @Failure      404  {object}  dto.Response[any]
// @Router       /skills/{id} [get]
func (h *SkillHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.SkillRead]{Success: true, Data: *out})
}

// UpdateSkill godoc
// @Summary      Update a skill
// @Tags         Skills
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                  true   "Skill ID (UUID)"
// @Param        body  body  dto.SkillUpdateRequest  true   "Patch payload"
// @Success      200  {object}  dto.Response[dto.SkillRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /skills/{id} [patch]
func (h *SkillHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	var req dto.SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.SkillRead]{
		Success: true,
		Message: "Skill '" + out.Name + "' has been updated",
		Data:    *out,
	})
}

// DeleteSkill godoc
// @Summary      Delete a skill (soft delete)
// @Tags         Skills
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Skill ID (UUID)"
// @Success      200  {object}  dto.Response[any]
// @Failure      400  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /skills/{id} [delete]
func (h *SkillHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	out, err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{
		Success: true,
		Message: "Skill '" + out.Name + "' has been deleted",
	})
}

// ListUserSkills godoc
// @Summary      List a user's skills
// @Tags         Users,Skills
// @Security     BearerAuth
// @Produce      json
// @Param        userID  path  string  true  "User ID (UUID)"
// @Success      200  {object}  dto.Response[[]dto.SkillRead]
// @Router       /users/{userID}/skills [get]
func (h *SkillHandler) ListUserSkills(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid user id"))
		return
	}
	out, err := h.svc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.SkillRead]{Success: true, Data: out})
}

// AssignUserSkill godoc
// @Summary      Attach a skill to a user
// @Tags         Users,Skills
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        userID  path  string                       true  "User ID (UUID)"
// @Param        body    body  dto.UserSkillAssignRequest   true  "Skill to assign"
// @Success      204  "No Content"
// @Failure      400  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /users/{userID}/skills [post]
func (h *SkillHandler) AssignUserSkill(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid user id"))
		return
	}
	var req dto.UserSkillAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.AssignToUser(c.Request.Context(), userID, req.SkillID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveUserSkill godoc
// @Summary      Remove a skill from a user
// @Tags         Users,Skills
// @Security     BearerAuth
// @Produce      json
// @Param        userID   path  string  true  "User ID (UUID)"
// @Param        skillID  path  string  true  "Skill ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  dto.Response[any]
// @Router       /users/{userID}/skills/{skillID} [delete]
func (h *SkillHandler) RemoveUserSkill(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid user id"))
		return
	}
	skillID, err := uuid.Parse(c.Param("skillID"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid skill id"))
		return
	}
	if err := h.svc.RemoveFromUser(c.Request.Context(), userID, skillID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/handlers/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/skill_handler.go
git commit -m "feat(phase-4): add skill handler with user-skill endpoints"
```

---

## Task 13: Handler — `label_handler.go`

**Files:**
- Create: `internal/handlers/label_handler.go`

- [ ] **Step 1: Write `internal/handlers/label_handler.go`**

```go
package handlers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// LabelHandler exposes /api/v1/labels.
//
// NOTE: Python labels.py only had GET (list) and POST (get-or-create),
// both behind ANNOUNCEMENTS_MANAGE. Phase 4 scope adds GET-by-id, PATCH,
// DELETE for full CRUD parity. All endpoints stay under the same admin
// permission (PermAnnounceManage) per the spec's audit note.
type LabelHandler struct {
	svc services.LabelService
}

func NewLabelHandler(svc services.LabelService) *LabelHandler { return &LabelHandler{svc: svc} }

// ListLabels godoc
// @Summary      List labels
// @Tags         Labels
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int    false  "Page"             default(1)
// @Param        page_size  query  int    false  "Page size"        default(100)
// @Param        search     query  string false  "Search by name"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.LabelRead]]
// @Router       /labels [get]
func (h *LabelHandler) List(c *gin.Context) {
	var q dto.LabelListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 100
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.LabelRead]]{
		Success: true,
		Data: dto.PaginatedData[dto.LabelRead]{
			Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
		},
	})
}

// CreateLabel godoc
// @Summary      Create or fetch a label (get-or-create)
// @Description  Returns the existing label (case-insensitive name match) or creates a new one.
// @Tags         Labels
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.LabelCreateRequest  true  "Label payload"
// @Success      201  {object}  dto.Response[dto.LabelRead]
// @Failure      400  {object}  dto.Response[any]
// @Router       /labels [post]
func (h *LabelHandler) Create(c *gin.Context) {
	var req dto.LabelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.GetOrCreate(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[dto.LabelRead]{Success: true, Data: *out})
}

// GetLabel godoc
// @Summary      Get a label by ID
// @Tags         Labels
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Label ID (UUID)"
// @Success      200  {object}  dto.Response[dto.LabelRead]
// @Failure      404  {object}  dto.Response[any]
// @Router       /labels/{id} [get]
func (h *LabelHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.LabelRead]{Success: true, Data: *out})
}

// UpdateLabel godoc
// @Summary      Update a label
// @Tags         Labels
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                  true  "Label ID (UUID)"
// @Param        body  body  dto.LabelUpdateRequest  true  "Patch payload"
// @Success      200  {object}  dto.Response[dto.LabelRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /labels/{id} [patch]
func (h *LabelHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	var req dto.LabelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.LabelRead]{Success: true, Data: *out})
}

// DeleteLabel godoc
// @Summary      Delete a label (soft delete)
// @Tags         Labels
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Label ID (UUID)"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /labels/{id} [delete]
func (h *LabelHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	out, err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{
		Success: true,
		Message: "Label '" + out.Name + "' has been deleted",
	})
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./internal/handlers/...`
Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/label_handler.go
git commit -m "feat(phase-4): add label handler"
```

---

## Task 14: Wire routes in `cmd/server/main.go`

**Files:**
- Modify: `cmd/server/main.go`

The exact line numbers depend on how Phase 0–3 wrote `main.go`. Find the section where Phase 3 wired `departments` and `positions` repos/services/handlers, and add the skill + label wiring **next to it**. The pattern below is the canonical shape — adapt names if Phase 0 chose different identifiers (e.g. `db` vs `gormDB`).

- [ ] **Step 1: Add imports (top of file)**

Ensure the import block includes (add any that are missing):

```go
import (
    // ... existing imports ...
    "github.com/exnodes/hrm-api/internal/handlers"
    "github.com/exnodes/hrm-api/internal/middleware"
    perm "github.com/exnodes/hrm-api/internal/permissions"
    "github.com/exnodes/hrm-api/internal/repositories"
    "github.com/exnodes/hrm-api/internal/services"
)
```

- [ ] **Step 2: Wire repositories + services + handlers**

Add the following near the other repo/service/handler construction calls. Replace `db` with the GORM `*gorm.DB` identifier already used in `main.go`:

```go
// --- Skills + Labels (Phase 4) -------------------------------------------
skillRepo     := repositories.NewSkillRepository(db)
userSkillRepo := repositories.NewUserSkillRepository(db)
labelRepo     := repositories.NewLabelRepository(db)

skillSvc := services.NewSkillService(skillRepo, userSkillRepo)
labelSvc := services.NewLabelService(labelRepo)

skillHandler := handlers.NewSkillHandler(skillSvc)
labelHandler := handlers.NewLabelHandler(labelSvc)
```

- [ ] **Step 3: Register routes**

Inside the existing authenticated route group (the one that already has `authed.Use(middleware.JWT())`), append:

```go
// --- Skills ---------------------------------------------------------------
skills := authed.Group("/skills")
skills.GET("",        middleware.RequirePerms(perm.PermSkillsRead),   skillHandler.List)
skills.POST("",       middleware.RequirePerms(perm.PermSkillsCreate), skillHandler.Create)
skills.GET("/:id",    middleware.RequirePerms(perm.PermSkillsRead),   skillHandler.Get)
skills.PATCH("/:id",  middleware.RequirePerms(perm.PermSkillsUpdate), skillHandler.Update)
skills.DELETE("/:id", middleware.RequirePerms(perm.PermSkillsDelete), skillHandler.Delete)

// --- User <-> Skill bridge -----------------------------------------------
users := authed.Group("/users/:userID/skills")
users.GET("",            middleware.RequirePerms(perm.PermUsersRead),   skillHandler.ListUserSkills)
users.POST("",           middleware.RequirePerms(perm.PermUsersUpdate), skillHandler.AssignUserSkill)
users.DELETE("/:skillID", middleware.RequirePerms(perm.PermUsersUpdate), skillHandler.RemoveUserSkill)

// --- Labels ---------------------------------------------------------------
labels := authed.Group("/labels")
labels.GET("",        middleware.RequirePerms(perm.PermAnnounceManage), labelHandler.List)
labels.POST("",       middleware.RequirePerms(perm.PermAnnounceManage), labelHandler.Create)
labels.GET("/:id",    middleware.RequirePerms(perm.PermAnnounceManage), labelHandler.Get)
labels.PATCH("/:id",  middleware.RequirePerms(perm.PermAnnounceManage), labelHandler.Update)
labels.DELETE("/:id", middleware.RequirePerms(perm.PermAnnounceManage), labelHandler.Delete)
```

- [ ] **Step 4: Regenerate Swagger**

Run: `make swag` (alias for `swag init -g cmd/server/main.go -o docs/swagger`)
Expected: `generated docs/swagger/swagger.{json,yaml}`; no errors.

- [ ] **Step 5: Build the binary**

Run: `go build ./...`
Expected: exit code 0.

- [ ] **Step 6: Boot the server and smoke `/health`**

Run (in one shell):
```bash
make run
```
Then in another shell:
```bash
curl -s http://localhost:8080/health
```
Expected: `{"status":"ok"}` (or whatever Phase 0 returned for `/health`).
Stop the server (`Ctrl+C`) before committing.

- [ ] **Step 7: Commit**

```bash
git add cmd/server/main.go docs/swagger/
git commit -m "feat(phase-4): wire skills + labels routes"
```

---

## Task 15: End-to-end verification + verification log

**Files:**
- Create: `docs/superpowers/verification/phase-04.md`

Boot the server (`make run`) in a separate shell. Run every command below from the project root in a fresh shell with `BASE=http://localhost:8080/api/v1` exported. Capture each request + response into the verification log.

- [ ] **Step 1: Log in as super-admin and capture the access token**

Run:
```bash
BASE=http://localhost:8080/api/v1
TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@exnodes.local","password":"ChangeMe123!"}' | \
  python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["access_token"])')
echo "$TOKEN" | head -c 30
```
Expected: first 30 chars of a JWT (e.g. `eyJhbGciOiJIUzI1NiIs...`). If empty, fail loudly — seeded super-admin credentials live in `.env`.

- [ ] **Step 2: Create a skill (happy path)**

Run:
```bash
curl -s -X POST "$BASE/skills" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Go","description":"Server-side language"}'
```
Expected: HTTP 201, body has `success: true`, `data.name == "Go"`, `data.id` is a UUID. Save `SKILL_ID=$(...)`.

- [ ] **Step 3: Create the same skill again (conflict)**

Run:
```bash
curl -s -o /tmp/dup.json -w "%{http_code}\n" -X POST "$BASE/skills" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"go"}'
cat /tmp/dup.json
```
Expected: HTTP `409`; body `success: false, code: "conflict", message: "Skill name already exists"`.

- [ ] **Step 4: List + search skills**

Run:
```bash
curl -s "$BASE/skills?search=go" -H "Authorization: Bearer $TOKEN"
```
Expected: HTTP 200, `data.total == 1`, `data.items[0].name == "Go"`.

- [ ] **Step 5: Get + update skill**

Run:
```bash
curl -s "$BASE/skills/$SKILL_ID" -H "Authorization: Bearer $TOKEN"
curl -s -X PATCH "$BASE/skills/$SKILL_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"description":"Updated"}'
```
Expected: GET 200 with name "Go"; PATCH 200 with `data.description == "Updated"`.

- [ ] **Step 6: Assign skill to the super-admin user, then verify**

Run:
```bash
ME_ID=$(curl -s "$BASE/users/me" -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')

curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/users/$ME_ID/skills" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"skill_id\":\"$SKILL_ID\"}"

curl -s "$BASE/users/$ME_ID/skills" -H "Authorization: Bearer $TOKEN"
```
Expected: POST returns `204`; GET returns `{"success":true,"data":[{"id":"<uuid>","name":"Go",...}]}`.

- [ ] **Step 7: Try to delete the assigned skill (blocked)**

Run:
```bash
curl -s -o /tmp/del.json -w "%{http_code}\n" -X DELETE "$BASE/skills/$SKILL_ID" \
  -H "Authorization: Bearer $TOKEN"
cat /tmp/del.json
```
Expected: HTTP `400`; body `code: "bad_request"`; `details.employee_count == 1`.

- [ ] **Step 8: Unassign and then delete**

Run:
```bash
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$BASE/users/$ME_ID/skills/$SKILL_ID" -H "Authorization: Bearer $TOKEN"
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$BASE/skills/$SKILL_ID" -H "Authorization: Bearer $TOKEN"
curl -s -o /tmp/gone.json -w "%{http_code}\n" "$BASE/skills/$SKILL_ID" -H "Authorization: Bearer $TOKEN"
cat /tmp/gone.json
```
Expected: 204, 200, 404 (in that order); final GET body `code: "not_found"`.

- [ ] **Step 9: DB spot-check — soft delete is honored**

Run:
```bash
psql "$DATABASE_URL" -c "SELECT id, name, is_deleted, deleted_at FROM skills ORDER BY created_at DESC LIMIT 5;"
```
Expected: the skill row still exists with `is_deleted = t` and a populated `deleted_at`.

- [ ] **Step 10: Label CRUD happy path**

Run:
```bash
LABEL=$(curl -s -X POST "$BASE/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Important","color":"#ff0000"}')
echo "$LABEL"
LABEL_ID=$(echo "$LABEL" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')

curl -s -X POST "$BASE/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"important"}'

curl -s "$BASE/labels?search=imp" -H "Authorization: Bearer $TOKEN"

curl -s -X PATCH "$BASE/labels/$LABEL_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"color":"#00ff00"}'

curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$BASE/labels/$LABEL_ID" -H "Authorization: Bearer $TOKEN"
curl -s -o /tmp/labelgone.json -w "%{http_code}\n" "$BASE/labels/$LABEL_ID" -H "Authorization: Bearer $TOKEN"
cat /tmp/labelgone.json
```
Expected:
- First POST: 201, returns new label.
- Second POST (same name lowercase): 201, returns the **same** `id` (get-or-create).
- List with search: `data.total == 1`.
- PATCH: 200, `data.color == "#00ff00"`.
- DELETE: 200; subsequent GET: 404 `code: "not_found"`.

- [ ] **Step 11: Permission denial — log in as a read-only user**

Run:
```bash
RTOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"employee@exnodes.local","password":"ChangeMe123!"}' | \
  python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["access_token"])')

curl -s -o /tmp/forbidden.json -w "%{http_code}\n" -X POST "$BASE/skills" \
  -H "Authorization: Bearer $RTOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Nope"}'
cat /tmp/forbidden.json

curl -s -o /tmp/labelforbidden.json -w "%{http_code}\n" "$BASE/labels" \
  -H "Authorization: Bearer $RTOKEN"
cat /tmp/labelforbidden.json
```
Expected: both calls return `403` with `code: "forbidden"`.

- [ ] **Step 12: Unauthenticated calls are rejected**

Run:
```bash
curl -s -o /tmp/anon.json -w "%{http_code}\n" "$BASE/skills"
cat /tmp/anon.json
```
Expected: `401`, `code: "unauthorized"`.

- [ ] **Step 13: Persist the verification log**

Create `docs/superpowers/verification/phase-04.md` with this exact structure (fill the `<…>` placeholders with the real captured output from each step). The intent is that the file is a copy-pasteable transcript a reviewer can re-run.

```markdown
# Phase 4 — Skills + Labels Verification Log

**Date:** 2026-05-15
**Tester:** danny.tranhoang@exnodes.vn
**Branch/commit:** `<git rev-parse --short HEAD>`
**Env:** local Docker Postgres @ `$DATABASE_URL`, server on :8080.

## 0. Preflight
- `make migrate-up` → applied through `000006_create_labels`.
- `make run` boots cleanly. `/health` returns `{"status":"ok"}`.

## 1. Super-admin login
Request:
\`\`\`http
POST /api/v1/auth/login
{"email":"admin@exnodes.local","password":"ChangeMe123!"}
\`\`\`
Response: 200, access_token captured (first 30 chars: `<paste>`).

## 2. Create skill (201)
<paste request + response>

## 3. Duplicate skill name (409 conflict)
<paste request + response>

## 4. List + search skills (200, total=1, items[0].name=Go)
<paste>

## 5. Get + PATCH skill (200/200)
<paste>

## 6. Assign skill to super-admin (204) + list (200, length 1)
<paste>

## 7. Delete skill while assigned (400 bad_request, employee_count=1)
<paste>

## 8. Unassign then delete (204 / 200), GET 404 not_found
<paste>

## 9. DB spot-check — soft delete row
\`\`\`text
<paste psql output showing is_deleted=t, deleted_at populated>
\`\`\`

## 10. Label CRUD (create 201, get-or-create idempotent, search 1, patch 200, delete 200, get 404)
<paste>

## 11. Permission denial (403 on POST /skills + GET /labels as employee)
<paste>

## 12. Unauthenticated (401 on GET /skills with no Authorization)
<paste>

## Result
All steps green. Phase 4 ready for merge.
```

- [ ] **Step 14: Commit**

```bash
git add docs/superpowers/verification/phase-04.md
git commit -m "docs(phase-4): add end-to-end verification log"
```

---

## Task 16: Phase wrap-up — README, final tests, sanity

**Files:**
- Modify: `README.md` (Endpoints section)

- [ ] **Step 1: Append the new endpoints to the README**

Add this block under the existing "API endpoints" / "Routes" section in `README.md` (create the section if it doesn't already exist):

```markdown
### Phase 4 — Skills + Labels

| Method | Path                                         | Permission          |
|--------|----------------------------------------------|---------------------|
| GET    | /api/v1/skills                               | skills:read         |
| POST   | /api/v1/skills                               | skills:create       |
| GET    | /api/v1/skills/{id}                          | skills:read         |
| PATCH  | /api/v1/skills/{id}                          | skills:update       |
| DELETE | /api/v1/skills/{id}                          | skills:delete       |
| GET    | /api/v1/users/{userID}/skills                | users:read          |
| POST   | /api/v1/users/{userID}/skills                | users:update        |
| DELETE | /api/v1/users/{userID}/skills/{skillID}      | users:update        |
| GET    | /api/v1/labels                               | announcements:manage |
| POST   | /api/v1/labels                               | announcements:manage |
| GET    | /api/v1/labels/{id}                          | announcements:manage |
| PATCH  | /api/v1/labels/{id}                          | announcements:manage |
| DELETE | /api/v1/labels/{id}                          | announcements:manage |
```

- [ ] **Step 2: Run the full test suite once more**

Run: `go test ./...`
Expected: every package reports `ok`; no `FAIL` lines.

- [ ] **Step 3: Final build sanity**

Run: `go build ./... && make migrate-up && make migrate-version`
Expected: build clean; migrate-version prints `6` (or whatever Phase 0–3 expected sequence ends on, +2 for Phase 4's two new migrations).

- [ ] **Step 4: Verify Swagger UI**

Boot `make run`, open `http://localhost:8080/swagger/index.html`, confirm:
- "Skills" tag lists all 5 CRUD routes.
- "Labels" tag lists all 5 CRUD routes.
- "Users,Skills" tag lists the 3 bridge routes.
- Each route shows the `BearerAuth` lock icon.

- [ ] **Step 5: Commit and tag the phase**

```bash
git add README.md
git commit -m "docs(phase-4): document new endpoints in README"
git tag phase-4-done
```

---

## Definition of Done

This phase is complete when **all** of the following hold:

1. `migrations/000005_create_skills.up.sql` + `.down.sql` and `migrations/000006_create_labels.up.sql` + `.down.sql` are committed and apply/rollback cleanly on a fresh DB.
2. Every new table has the four audit columns + `set_updated_at()` trigger + `is_deleted` index.
3. Every new repository query passes through the `NotDeleted` scope unless explicitly stated otherwise.
4. Every new handler carries `middleware.RequirePerms(...)` matching the permission table above.
5. Every new handler has Swagger annotations including `@Security BearerAuth` and the routes are visible on `/swagger/index.html`.
6. `go test ./...` is fully green, including the new `skill_service_test.go` (9 tests) and `label_service_test.go` (8 tests).
7. `make migrate-up` and `make run` succeed from a freshly-cloned working tree.
8. **End-to-end transcript committed** to `docs/superpowers/verification/phase-04.md` covering: login → skill CRUD → label CRUD → user-skill assign/list/unassign → soft-delete spot-check → 409 conflict → 400 employee-count block → 403 forbidden → 401 unauthenticated.
9. README updated with the new endpoint table.
10. `phase-4-done` git tag created on the final commit.
