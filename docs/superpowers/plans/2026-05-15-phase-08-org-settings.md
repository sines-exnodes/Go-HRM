# Phase 8: Organization Settings + System Config Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## ⚠️ REVISION NOTES (2026-05-20) — AUTHORITATIVE, read & apply before executing any task

This plan was drafted assuming Phase 7 would consume migration `000009`. Phase 6 actually took `000009` (attendance) and Phase 7 took `000010` (announcements). The codebase audit at the close of Phase 7 supersedes the task bodies below wherever they conflict. **Execute per these notes, not the raw task bodies where they conflict.**

1. **Migration number = `000011`** (NOT `000010`). Final `make migrate-version` after Phase 8 = **11**. Rename every filename and reference accordingly.

2. **`company_address_updated_by` FK target = `employees(id)`** (NOT `users(id)`). Mirrors `leave_requests.created_by` + `announcements.author_id` per the Go schema split. Keep `ON DELETE SET NULL` — the column is an audit trail and the row must survive an HR-profile delete. Author display reads `Employee.FullName`.

3. **Router wiring lives in `cmd/server/main.go`** (NOT `internal/handlers/router.go`). [`internal/handlers/router.go`](../../internal/handlers/router.go) is mostly health-only since Phase 0; every other phase wires its routes inside the `v1 := r.Group("/api/v1")` block in `main.go`. Task 7 + Task 9 collapse into one main.go edit (build repo + service + handler, then register routes).

4. **Seed service signature is `(s *SeedService) Seed(ctx context.Context) error`** — the DB is held inside the `SeedService` struct, not passed as an arg. Task 8 adds a new method `seedSystemConfig(ctx)` (mirroring the existing `seedRoles` / `seedSuperAdmin` / `seedOrgDefaults`) and calls it from `Seed()`. The system_config repo is injected via the constructor (`NewSeedService(..., systemConfigRepo)`), NOT constructed inline.

5. **Existing handler helpers** — `currentUser(c) (*models.User, bool)` lives in [`internal/handlers/employee_handler.go`](../../internal/handlers/employee_handler.go) line 38; `parseIDParam` at line 52. The Phase 8 handler reuses them (don't redeclare).

6. **`RequirePerms` takes `authSvc *AuthService` as first arg** — `middleware.RequirePerms(authSvc, perm.PermOrgSettings)`. Task 7's route block must include `authSvc` in every call.

7. **`apperrors.Write(c, err)` does NOT exist.** Use `_ = c.Error(err); return` per the established pattern.

8. **`make swag` output path = `docs/swagger`** (already correct in Task 9, just noted for consistency with prior phases).

9. **Singleton ID is hardcoded — never expose mutation.** The repository must NEVER accept an arbitrary ID for upsert. Every read goes through `Get()` which queries `WHERE id = SystemConfigSingletonID`; every write goes through `Update()` which scopes to the same ID. Trying to insert a second row would hit the DB-level CHECK constraint but the repo defensively prevents the attempt.

10. **`NotDeleted` scope is NOT applied** — singleton has no soft-delete semantics. The repo does not chain `Scopes(models.NotDeleted)` (consistent with the model not embedding `BaseModel`).

11. **Plan task 1 Step 6 commit message: `feat(migrations): 000011 …`** (was `000010`).

12. **Spec §6.3 vs registry**: `PermOrgSettings = "organization_settings:manage"` is already in [`internal/permissions/registry.go`](../../internal/permissions/registry.go) (Phase 1 seeded it to Admin + HR Manager). No registry change in Phase 8.

13. **Test helper `truncateAll`**: add `system_config` to the TRUNCATE list (between announcements children and the rest). Tests assert on a fresh row after each run; the singleton check constraint means we must INSERT the sentinel back manually OR rely on the seed step run via the test fixture. Simplest: in the test helper's setup, after truncate, INSERT the sentinel row directly.

Everything else in the task bodies (TDD-first, commit-per-task, no placeholders, bite-sized steps) still applies. **Execute per these REVISION NOTES, not the raw task bodies where they conflict.**

---


**Goal:** Port the Python `organization_settings` module to Go. Provide a **singleton** `system_config` row (one fixed UUID, no soft delete) that backs two logical sub-resources:

- `GET /api/v1/organization-settings/attendance` (read with `PermOrgSettings`) + `PATCH` to update late-arrival threshold
- `GET /api/v1/organization-settings/company-profile` (any authenticated user, FE renders it) + `PATCH` (with `PermOrgSettings`) to update company address + lat/lng

This mirrors the FastAPI router exactly (see `exnodes-hrm-api/app/routers/organization_settings.py`).

**Schema decision — recorded here (Non-negotiable #3 / task brief):**
- The Python source `app/models/system_config.py` is a **singleton Beanie Document**, not a key-value store. Fields are statically declared: `late_threshold_hour/minute`, `checkout_threshold_hour/minute`, `company_address`, `company_latitude`, `company_longitude`, `company_address_updated_at`, `company_address_updated_by`.
- We therefore model this in Postgres as a single-row `system_config` table — **no key-value config table** (the brief allows skipping it when the Python model is structured rather than KV).
- A fixed sentinel UUID `00000000-0000-0000-0000-000000000001` is the only allowed PK. A `CHECK (id = '00000000-0000-0000-0000-000000000001')` constraint enforces the singleton invariant at the DB level.
- Soft delete is **not** applied — singleton settings have no delete path. `is_deleted` and `deleted_at` columns are still added (per spec §5.2 audit-cols requirement) but are never set to true; repository methods never apply the `NotDeleted` scope because there is nothing to soft-delete.
- The seed service performs an idempotent `INSERT ... ON CONFLICT DO NOTHING` of the sentinel row on boot.
- No logo upload in this phase: the Python router does not expose one (`GET /company-profile` reads address only; `PATCH /company-profile` writes address + lat/lng). FE upload + future logo field can be added in a later migration without breaking the contract.

**Architecture recap (carried over from Phases 0–7):** layered handlers/services/repos under `internal/`, GORM + Postgres + golang-migrate, AppError middleware, JWT + `RequirePerms` middleware, `Response[T]` envelope, swaggo annotations on every handler, real-Postgres service tests, seed service runs on boot.

**Assumed prior state from Phases 0–7:**
- `internal/config/config.go`, `db.go`, GORM and migration tooling exist.
- `internal/models/base.go` provides `BaseModel` + `NotDeleted` scope.
- `internal/permissions/registry.go` defines `PermOrgSettings = "organization_settings:manage"` and is wired into the FE permission picker (`GET /api/v1/roles/permissions`).
- `internal/middleware/jwt.go` (`middleware.JWT()`) and `internal/middleware/require_perms.go` (`middleware.RequirePerms(...)`) exist.
- `internal/errors/errors.go` (`apperrors`) provides `ErrBadRequest`, `ErrForbidden`, `ErrNotFound`, `ErrConflict`.
- `internal/dto/response.go` provides `Response[T]` + `NewResponse`.
- `internal/services/seed_service.go` exposes a `Seed(ctx, *gorm.DB)` entry point that Phase 1 / 2 already extended; this phase appends one more call.
- `cmd/server/main.go` builds, wires routes via `handlers.RegisterRoutes(r, h, cfg.SwaggerEnabled)` and exposes `/swagger/index.html`.
- `services/testhelper_test.go` provides `setupTestDB(t)`, `makeAdmin(t)`, `makeReadonly(t)` factories.

**Endpoints produced by this phase (final shape):**

| Method | Path | Auth | Permission |
|---|---|---|---|
| GET   | `/api/v1/organization-settings/attendance`      | JWT | `PermOrgSettings` |
| PATCH | `/api/v1/organization-settings/attendance`      | JWT | `PermOrgSettings` |
| GET   | `/api/v1/organization-settings/company-profile` | JWT | (any authenticated user — read-only) |
| PATCH | `/api/v1/organization-settings/company-profile` | JWT | `PermOrgSettings` |

The mapping deliberately matches the Python router so the FE does not need contract changes.

---

### Task 1: Add `PermOrgSettings` (verify) and create migration `000010_create_system_config`

**Files:**
- Read-only verify: `internal/permissions/registry.go` (constant `PermOrgSettings` must already exist from Phase 1; this task only confirms it).
- Create: `migrations/000010_create_system_config.up.sql`
- Create: `migrations/000010_create_system_config.down.sql`

- [ ] **Step 1: Confirm `PermOrgSettings` is registered**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  grep -n 'PermOrgSettings' internal/permissions/registry.go
```
Expected: at least one line printing the constant declaration (per spec §6.3) AND another line listing it in `PermissionGroups`. If either is missing, STOP and reopen Phase 1 — this plan does not modify the registry, it only consumes it.

- [ ] **Step 2: Confirm the next migration sequence number is `000010`**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  ls migrations | grep -E '^[0-9]{6}_' | sort | tail -3
```
Expected: the highest existing prefix is `000009_*.sql` (created by Phase 7). If it is not 9, adjust the filename in steps 3–4 to be `next+1` and update every reference to `000010` in this plan accordingly.

- [ ] **Step 3: Write the up migration**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000010_create_system_config.up.sql` with exactly:
```sql
-- =============================================================
-- 000010_create_system_config
-- Single-row organization-wide configuration (a port of the Python
-- Beanie SystemConfig singleton).
--
-- The table is constrained to exactly one row whose id is the
-- sentinel UUID '00000000-0000-0000-0000-000000000001'. The seed
-- service performs INSERT ... ON CONFLICT DO NOTHING on boot to
-- guarantee the row exists.
--
-- Audit cols are included (per spec §5.2) but soft-delete is not
-- meaningful for a singleton — is_deleted stays false and the row
-- is updated in place.
-- =============================================================

CREATE TABLE system_config (
    id          UUID PRIMARY KEY DEFAULT '00000000-0000-0000-0000-000000000001',

    -- Attendance: late-arrival threshold
    late_threshold_hour     SMALLINT NOT NULL DEFAULT 9
        CHECK (late_threshold_hour BETWEEN 0 AND 23),
    late_threshold_minute   SMALLINT NOT NULL DEFAULT 0
        CHECK (late_threshold_minute BETWEEN 0 AND 59),

    -- Attendance: checkout threshold (used for future early-leave calc)
    checkout_threshold_hour     SMALLINT NOT NULL DEFAULT 18
        CHECK (checkout_threshold_hour BETWEEN 0 AND 23),
    checkout_threshold_minute   SMALLINT NOT NULL DEFAULT 0
        CHECK (checkout_threshold_minute BETWEEN 0 AND 59),

    -- Company profile
    company_address                 TEXT,
    company_latitude                DOUBLE PRECISION,
    company_longitude               DOUBLE PRECISION,
    company_address_updated_at      TIMESTAMPTZ,
    company_address_updated_by      UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Audit cols (per spec §5.2). Soft delete is unused but the cols
    -- are present so the table matches every other entity.
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ,

    -- Enforce singleton at the DB level. Any second INSERT fails the check.
    CONSTRAINT system_config_singleton CHECK (id = '00000000-0000-0000-0000-000000000001')
);

CREATE INDEX system_config_is_deleted_idx ON system_config (is_deleted);

CREATE TRIGGER system_config_set_updated_at
    BEFORE UPDATE ON system_config
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 4: Write the down migration**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000010_create_system_config.down.sql` with exactly:
```sql
-- Reverse of 000010_create_system_config.
DROP TRIGGER IF EXISTS system_config_set_updated_at ON system_config;
DROP INDEX  IF EXISTS system_config_is_deleted_idx;
DROP TABLE  IF EXISTS system_config;
```

- [ ] **Step 5: Apply the migration and verify the schema**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  make migrate-up && \
  make migrate-version
```
Expected: `make migrate-up` exits 0 with `10/u create_system_config` in its log; `make migrate-version` prints `10`.

Then verify the table & constraint via psql:
```bash
psql "$(grep -E '^DATABASE_URL=' .env | cut -d= -f2-)" \
  -c "\d system_config" \
  -c "SELECT conname FROM pg_constraint WHERE conrelid = 'system_config'::regclass;"
```
Expected: column list matches the up migration; `pg_constraint` rows include `system_config_singleton` and the 4 numeric range CHECKs.

- [ ] **Step 6: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add migrations/000010_create_system_config.up.sql migrations/000010_create_system_config.down.sql && \
  git commit -m "feat(migrations): 000010 create system_config singleton (attendance + company profile)"
```
Expected: 2 files changed.

---

### Task 2: GORM model `internal/models/system_config.go`

**Files:**
- Create: `internal/models/system_config.go`

- [ ] **Step 1: Write the model**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/models/system_config.go` with exactly:
```go
package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemConfigSingletonID is the fixed UUID of the single allowed row in the
// system_config table. The DB enforces this via the system_config_singleton
// CHECK constraint; this constant is used by the seed service to upsert the
// sentinel row and by the repository to read/update it.
var SystemConfigSingletonID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// SystemConfig is the singleton organization-wide configuration row. There
// is intentionally only ever one row; the repo never lists or filters by id.
//
// NOTE: this struct does NOT embed BaseModel because the singleton has a
// fixed sentinel UUID, no soft-delete semantics, and no list/get-by-id flow.
// The 4 audit columns are declared inline to keep the row shape consistent
// with every other entity table (spec §5.2).
type SystemConfig struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Attendance — late arrival threshold
	LateThresholdHour   int16 `gorm:"not null;default:9"  json:"late_threshold_hour"`
	LateThresholdMinute int16 `gorm:"not null;default:0"  json:"late_threshold_minute"`

	// Attendance — checkout threshold (carried over from Python for
	// future early-leave computation; not yet surfaced via API)
	CheckoutThresholdHour   int16 `gorm:"not null;default:18" json:"checkout_threshold_hour"`
	CheckoutThresholdMinute int16 `gorm:"not null;default:0"  json:"checkout_threshold_minute"`

	// Company profile
	CompanyAddress             *string    `gorm:"type:text"                                   json:"company_address,omitempty"`
	CompanyLatitude            *float64   `                                                   json:"company_latitude,omitempty"`
	CompanyLongitude           *float64   `                                                   json:"company_longitude,omitempty"`
	CompanyAddressUpdatedAt    *time.Time `                                                   json:"company_address_updated_at,omitempty"`
	CompanyAddressUpdatedBy    *uuid.UUID `gorm:"type:uuid"                                   json:"company_address_updated_by,omitempty"`

	// Audit columns (no soft delete used — singleton)
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	IsDeleted bool       `gorm:"not null;default:false" json:"-"`
	DeletedAt *time.Time `                              json:"-"`
}

// TableName pins the GORM table name so any plural-suffix convention doesn't
// rewrite it to "system_configs".
func (SystemConfig) TableName() string { return "system_config" }
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/models/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/models/system_config.go && \
  git commit -m "feat(models): SystemConfig singleton model with sentinel UUID"
```
Expected: 1 file changed.

---

### Task 3: DTOs `internal/dto/organization.go`

**Files:**
- Create: `internal/dto/organization.go`

- [ ] **Step 1: Write the DTOs**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/dto/organization.go` with exactly:
```go
package dto

import (
	"fmt"
	"time"
)

// =============================================================
// Attendance settings (sub-resource of /organization-settings)
// =============================================================

// AttendanceSettingsRead mirrors AttendanceSettingsRead from the Python
// schema (`exnodes-hrm-api/app/schemas/organization_settings.py`). The
// display field is a 12-hour formatted convenience string for the FE.
type AttendanceSettingsRead struct {
	LateThresholdHour    int    `json:"late_threshold_hour"    example:"9"`
	LateThresholdMinute  int    `json:"late_threshold_minute"  example:"0"`
	LateThresholdDisplay string `json:"late_threshold_display" example:"9:00 AM"`
}

// NewAttendanceSettingsRead builds the read DTO including the 12-hour
// display string. The format matches Python's _format_time().
func NewAttendanceSettingsRead(hour, minute int16) AttendanceSettingsRead {
	return AttendanceSettingsRead{
		LateThresholdHour:    int(hour),
		LateThresholdMinute:  int(minute),
		LateThresholdDisplay: formatTwelveHour(int(hour), int(minute)),
	}
}

// AttendanceSettingsUpdate is the PATCH body for updating the late
// arrival threshold. Both fields are required because the Python contract
// expects an atomic (hour, minute) pair.
type AttendanceSettingsUpdate struct {
	LateThresholdHour   *int `json:"late_threshold_hour"   binding:"required,min=0,max=23" example:"9"`
	LateThresholdMinute *int `json:"late_threshold_minute" binding:"required,min=0,max=59" example:"0"`
}

// =============================================================
// Company profile (sub-resource of /organization-settings)
// =============================================================

// CompanyProfileRead is what GET /organization-settings/company-profile
// returns. Any authenticated user may read it (FE needs the address to
// render the map preview).
type CompanyProfileRead struct {
	CompanyAddress          *string    `json:"company_address"             example:"123 Le Loi, Q1, HCMC"`
	CompanyLatitude         *float64   `json:"company_latitude"            example:"10.7769"`
	CompanyLongitude        *float64   `json:"company_longitude"           example:"106.7009"`
	CompanyAddressUpdatedAt *time.Time `json:"company_address_updated_at"`
}

// CompanyProfileUpdate is the PATCH body for updating the company
// address. All three fields move together: the FE is responsible for
// geocoding (Google Places) and passing matching lat/lng.
type CompanyProfileUpdate struct {
	CompanyAddress   *string  `json:"company_address"   binding:"omitempty,max=500"          example:"123 Le Loi, Q1, HCMC"`
	CompanyLatitude  *float64 `json:"company_latitude"  binding:"omitempty,min=-90,max=90"   example:"10.7769"`
	CompanyLongitude *float64 `json:"company_longitude" binding:"omitempty,min=-180,max=180" example:"106.7009"`
}

// formatTwelveHour returns "h:MM AM/PM" for the given 24-hour time.
// Kept private to dto/ because it's an internal display helper used only
// by NewAttendanceSettingsRead.
func formatTwelveHour(hour, minute int) string {
	suffix := "AM"
	if hour >= 12 {
		suffix = "PM"
	}
	h12 := hour % 12
	if h12 == 0 {
		h12 = 12
	}
	return fmt.Sprintf("%d:%02d %s", h12, minute, suffix)
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/dto/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/dto/organization.go && \
  git commit -m "feat(dto): organization settings + company profile request/response shapes"
```
Expected: 1 file changed.

---

### Task 4: Repository `internal/repositories/system_config_repo.go`

**Files:**
- Create: `internal/repositories/system_config_repo.go`

- [ ] **Step 1: Write the repository**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/repositories/system_config_repo.go` with exactly:
```go
package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// SystemConfigRepository is the data-access boundary for the singleton
// system_config row. It deliberately exposes neither List nor Delete —
// the row is created once by the seed service and updated in place.
type SystemConfigRepository interface {
	// Get returns the singleton row. Returns gorm.ErrRecordNotFound if
	// the seed service has not yet inserted it (the service layer
	// converts this into a 500 since it indicates a misconfigured boot).
	Get(ctx context.Context) (*models.SystemConfig, error)

	// EnsureExists inserts the sentinel row if missing. Used by both
	// the seed service on boot and the service layer as a defensive
	// guard before any update. Idempotent.
	EnsureExists(ctx context.Context) error

	// Update applies the supplied column map to the singleton row and
	// returns the freshly-read row.
	Update(ctx context.Context, updates map[string]any) (*models.SystemConfig, error)
}

type systemConfigRepo struct {
	db *gorm.DB
}

// NewSystemConfigRepository constructs the concrete repository.
func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepo{db: db}
}

func (r *systemConfigRepo) Get(ctx context.Context) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	err := r.db.WithContext(ctx).
		Where("id = ?", models.SystemConfigSingletonID).
		First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *systemConfigRepo) EnsureExists(ctx context.Context) error {
	// ON CONFLICT DO NOTHING — the singleton CHECK + PK guarantee at
	// most one row, so this is safe to call from every boot.
	sql := `
        INSERT INTO system_config (id)
        VALUES (?)
        ON CONFLICT (id) DO NOTHING
    `
	if err := r.db.WithContext(ctx).Exec(sql, models.SystemConfigSingletonID).Error; err != nil {
		return fmt.Errorf("ensure system_config row: %w", err)
	}
	return nil
}

func (r *systemConfigRepo) Update(ctx context.Context, updates map[string]any) (*models.SystemConfig, error) {
	if len(updates) == 0 {
		return r.Get(ctx)
	}

	tx := r.db.WithContext(ctx).
		Model(&models.SystemConfig{}).
		Where("id = ?", models.SystemConfigSingletonID).
		Updates(updates)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		// Shouldn't happen: seed inserts the row on boot. If it does,
		// return ErrRecordNotFound so the service layer raises a 500.
		return nil, errors.Join(gorm.ErrRecordNotFound, errors.New("system_config singleton row missing"))
	}
	return r.Get(ctx)
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/repositories/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/repositories/system_config_repo.go && \
  git commit -m "feat(repo): SystemConfigRepository (Get/EnsureExists/Update for singleton row)"
```
Expected: 1 file changed.

---

### Task 5: Service `internal/services/organization_settings_service.go`

**Files:**
- Create: `internal/services/organization_settings_service.go`

- [ ] **Step 1: Write the service**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/organization_settings_service.go` with exactly:
```go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// OrganizationSettingsService is the business-logic boundary for the
// singleton system_config row. It exposes two logical sub-resources
// (attendance + company profile) so the FE can render each tab
// independently without leaking the underlying single-table layout.
type OrganizationSettingsService interface {
	GetAttendance(ctx context.Context) (*dto.AttendanceSettingsRead, error)
	UpdateAttendance(ctx context.Context, body dto.AttendanceSettingsUpdate) (*dto.AttendanceSettingsRead, error)

	GetCompanyProfile(ctx context.Context) (*dto.CompanyProfileRead, error)
	UpdateCompanyProfile(ctx context.Context, body dto.CompanyProfileUpdate, updatedBy uuid.UUID) (*dto.CompanyProfileRead, error)

	// EnsureExists is called by the seed service on boot to insert the
	// sentinel row. Exposed as a service-level method so the seed
	// service can stay repository-unaware.
	EnsureExists(ctx context.Context) error
}

type orgSettingsService struct {
	repo repositories.SystemConfigRepository
}

// NewOrganizationSettingsService constructs the concrete service.
func NewOrganizationSettingsService(repo repositories.SystemConfigRepository) OrganizationSettingsService {
	return &orgSettingsService{repo: repo}
}

func (s *orgSettingsService) EnsureExists(ctx context.Context) error {
	return s.repo.EnsureExists(ctx)
}

func (s *orgSettingsService) GetAttendance(ctx context.Context) (*dto.AttendanceSettingsRead, error) {
	cfg, err := s.repo.Get(ctx)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	out := dto.NewAttendanceSettingsRead(cfg.LateThresholdHour, cfg.LateThresholdMinute)
	return &out, nil
}

func (s *orgSettingsService) UpdateAttendance(ctx context.Context, body dto.AttendanceSettingsUpdate) (*dto.AttendanceSettingsRead, error) {
	// Defensive: required-binding on the handler already rejects nil,
	// but the service contract should not assume that.
	if body.LateThresholdHour == nil || body.LateThresholdMinute == nil {
		return nil, apperrors.ErrBadRequest("late_threshold_hour and late_threshold_minute are required")
	}
	if *body.LateThresholdHour < 0 || *body.LateThresholdHour > 23 {
		return nil, apperrors.ErrBadRequest("late_threshold_hour must be between 0 and 23")
	}
	if *body.LateThresholdMinute < 0 || *body.LateThresholdMinute > 59 {
		return nil, apperrors.ErrBadRequest("late_threshold_minute must be between 0 and 59")
	}

	updates := map[string]any{
		"late_threshold_hour":   int16(*body.LateThresholdHour),
		"late_threshold_minute": int16(*body.LateThresholdMinute),
	}
	cfg, err := s.repo.Update(ctx, updates)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	out := dto.NewAttendanceSettingsRead(cfg.LateThresholdHour, cfg.LateThresholdMinute)
	return &out, nil
}

func (s *orgSettingsService) GetCompanyProfile(ctx context.Context) (*dto.CompanyProfileRead, error) {
	cfg, err := s.repo.Get(ctx)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &dto.CompanyProfileRead{
		CompanyAddress:          cfg.CompanyAddress,
		CompanyLatitude:         cfg.CompanyLatitude,
		CompanyLongitude:        cfg.CompanyLongitude,
		CompanyAddressUpdatedAt: cfg.CompanyAddressUpdatedAt,
	}, nil
}

func (s *orgSettingsService) UpdateCompanyProfile(ctx context.Context, body dto.CompanyProfileUpdate, updatedBy uuid.UUID) (*dto.CompanyProfileRead, error) {
	// Mirror the Python service: address + lat + lng + updated_at +
	// updated_by all move together. We accept nil values to mean
	// "clear that field", matching the Python schema's Optional fields.
	now := time.Now().UTC()
	updates := map[string]any{
		"company_address":             body.CompanyAddress,
		"company_latitude":            body.CompanyLatitude,
		"company_longitude":           body.CompanyLongitude,
		"company_address_updated_at":  now,
		"company_address_updated_by":  updatedBy,
	}
	cfg, err := s.repo.Update(ctx, updates)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &dto.CompanyProfileRead{
		CompanyAddress:          cfg.CompanyAddress,
		CompanyLatitude:         cfg.CompanyLatitude,
		CompanyLongitude:        cfg.CompanyLongitude,
		CompanyAddressUpdatedAt: cfg.CompanyAddressUpdatedAt,
	}, nil
}

// wrapNotFound converts gorm.ErrRecordNotFound into a 500 (rather than
// a 404) because the singleton row is supposed to exist after seed.
// Missing means a misconfigured boot, not a user error.
func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrInternal("system_config singleton row missing; seed did not run")
	}
	return err
}
```

- [ ] **Step 2: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/services/...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/organization_settings_service.go && \
  git commit -m "feat(service): OrganizationSettingsService (get/update attendance + company profile)"
```
Expected: 1 file changed.

---

### Task 6: Handler `internal/handlers/organization_settings_handler.go`

**Files:**
- Create: `internal/handlers/organization_settings_handler.go`

- [ ] **Step 1: Write the handler**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/organization_settings_handler.go` with exactly:
```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/middleware"
	"github.com/exnodes/hrm-api/internal/services"
)

// OrganizationSettingsHandler is the HTTP entry point for the
// /api/v1/organization-settings endpoints.
type OrganizationSettingsHandler struct {
	svc services.OrganizationSettingsService
}

// NewOrganizationSettingsHandler constructs the handler.
func NewOrganizationSettingsHandler(svc services.OrganizationSettingsService) *OrganizationSettingsHandler {
	return &OrganizationSettingsHandler{svc: svc}
}

// GetAttendance godoc
// @Summary      Get attendance settings
// @Description  Returns the current late-arrival threshold. Requires the `organization_settings:manage` permission (same as the Python contract).
// @Tags         Organization Settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[dto.AttendanceSettingsRead]
// @Failure      401  {object}  map[string]any
// @Failure      403  {object}  map[string]any
// @Router       /api/v1/organization-settings/attendance [get]
func (h *OrganizationSettingsHandler) GetAttendance(c *gin.Context) {
	data, err := h.svc.GetAttendance(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.NewResponse(*data, ""))
}

// UpdateAttendance godoc
// @Summary      Update late arrival threshold
// @Description  Updates the organization-wide late-arrival threshold. Takes effect immediately for all future check-ins. Does NOT retroactively update existing attendance records.
// @Tags         Organization Settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AttendanceSettingsUpdate  true  "New threshold (hour 0-23, minute 0-59)"
// @Success      200   {object}  dto.Response[dto.AttendanceSettingsRead]
// @Failure      400   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Failure      403   {object}  map[string]any
// @Router       /api/v1/organization-settings/attendance [patch]
func (h *OrganizationSettingsHandler) UpdateAttendance(c *gin.Context) {
	var body dto.AttendanceSettingsUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.UpdateAttendance(c.Request.Context(), body)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.NewResponse(*data, "Late arrival threshold has been updated"))
}

// GetCompanyProfile godoc
// @Summary      Get company profile
// @Description  Returns the current company address and coordinates. Any authenticated user — the FE needs this to render the map preview in shared screens.
// @Tags         Organization Settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[dto.CompanyProfileRead]
// @Failure      401  {object}  map[string]any
// @Router       /api/v1/organization-settings/company-profile [get]
func (h *OrganizationSettingsHandler) GetCompanyProfile(c *gin.Context) {
	data, err := h.svc.GetCompanyProfile(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.NewResponse(*data, ""))
}

// UpdateCompanyProfile godoc
// @Summary      Update company profile
// @Description  Updates the company address. Frontend is responsible for geocoding (Google Places) and passing the matching latitude/longitude.
// @Tags         Organization Settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CompanyProfileUpdate  true  "Address + coordinates"
// @Success      200   {object}  dto.Response[dto.CompanyProfileRead]
// @Failure      400   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Failure      403   {object}  map[string]any
// @Router       /api/v1/organization-settings/company-profile [patch]
func (h *OrganizationSettingsHandler) UpdateCompanyProfile(c *gin.Context) {
	var body dto.CompanyProfileUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		_ = c.Error(apperrors.ErrUnauthorized("missing user context"))
		return
	}
	parsed, err := uuid.Parse(userID)
	if err != nil {
		_ = c.Error(apperrors.ErrUnauthorized("invalid user context"))
		return
	}
	data, err := h.svc.UpdateCompanyProfile(c.Request.Context(), body, parsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.NewResponse(*data, "Company address has been updated"))
}
```

- [ ] **Step 2: Confirm `middleware.CurrentUserID` exists from Phase 1**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  grep -n 'func CurrentUserID' internal/middleware/
```
Expected: at least one match (Phase 1 helper). If missing, STOP and re-open Phase 1. This plan does not add it.

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./internal/handlers/...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/handlers/organization_settings_handler.go && \
  git commit -m "feat(handlers): organization settings endpoints (attendance + company profile)"
```
Expected: 1 file changed.

---

### Task 7: Wire routes in `internal/handlers/router.go`

**Files:**
- Modify: `internal/handlers/router.go`

- [ ] **Step 1: Read the current router**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  cat internal/handlers/router.go
```
Expected: prints the file. Confirm the file already wires the JWT-protected `authed` group from Phases 1-7 and defines a `Handlers` struct that the worker can extend.

- [ ] **Step 2: Add the `OrgSettings` field on the `Handlers` struct**

Edit `internal/handlers/router.go`:

1. In the `Handlers` struct, add a new field (alphabetised among existing fields):
   ```go
   OrgSettings *OrganizationSettingsHandler
   ```

2. In `NewHandlers`, accept a new parameter of type `services.OrganizationSettingsService` (alphabetised among existing service args) and assign:
   ```go
   OrgSettings: NewOrganizationSettingsHandler(orgSettingsSvc),
   ```
   Match the signature pattern of every other handler wired by previous phases (e.g. `userSvc services.UserService`, `roleSvc services.RoleService`, ...).

3. In `RegisterRoutes`, inside the JWT-protected `authed` group, append the following block **after** the last route registered by Phase 7 (announcements) and **before** any `_ = authed` discard:
   ```go
   org := authed.Group("/organization-settings")
   {
       // Attendance — admin/HR only (matches Python contract).
       org.GET("/attendance",
           middleware.RequirePerms(perm.PermOrgSettings),
           h.OrgSettings.GetAttendance)
       org.PATCH("/attendance",
           middleware.RequirePerms(perm.PermOrgSettings),
           h.OrgSettings.UpdateAttendance)

       // Company profile — read is open to any authenticated user
       // (FE renders the map preview in shared screens). Update
       // remains admin-only.
       org.GET("/company-profile",
           h.OrgSettings.GetCompanyProfile)
       org.PATCH("/company-profile",
           middleware.RequirePerms(perm.PermOrgSettings),
           h.OrgSettings.UpdateCompanyProfile)
   }
   ```

4. Ensure the imports of the package include `perm "github.com/exnodes/hrm-api/internal/permissions"` and `"github.com/exnodes/hrm-api/internal/middleware"`; both should already be present from Phases 1-7. Add the `services` import if it's not already there.

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: no output, exit code 0. If `NewHandlers` is now called with the wrong number of args from `cmd/server/main.go`, fix the call site in the next task — leave the compile error here only as a temporary state if absolutely necessary, otherwise update `main.go` immediately.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/handlers/router.go && \
  git commit -m "feat(handlers): register /api/v1/organization-settings routes"
```
Expected: 1 file changed.

---

### Task 8: Bootstrap singleton row in seed service

**Files:**
- Modify: `internal/services/seed_service.go`

- [ ] **Step 1: Read the current seed service**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  cat internal/services/seed_service.go
```
Expected: prints the file. Identify the function used by Phases 1-2 for idempotent system seeding (typically `Seed(ctx context.Context, db *gorm.DB) error` or equivalent). All edits below assume that entry-point shape.

- [ ] **Step 2: Append the org-settings seed step**

Inside the seed function, after the existing seed steps (roles, super admin, departments, positions), add:

```go
// Phase 8: ensure the system_config singleton row exists.
orgRepo := repositories.NewSystemConfigRepository(db)
orgSvc := NewOrganizationSettingsService(orgRepo)
if err := orgSvc.EnsureExists(ctx); err != nil {
    return fmt.Errorf("seed system_config singleton: %w", err)
}
```

Ensure the file imports `"github.com/exnodes/hrm-api/internal/repositories"` and `"fmt"` (most seed files already do).

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Run the seed once and verify the row exists**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  (make run > /tmp/phase08-server.log 2>&1 &) && sleep 3 && \
  psql "$(grep -E '^DATABASE_URL=' .env | cut -d= -f2-)" \
    -c "SELECT id, late_threshold_hour, late_threshold_minute, company_address FROM system_config;" && \
  pkill -f 'go run ./cmd/server' || true
```
Expected:
- `make run` produces a `listening on :8080` line in `/tmp/phase08-server.log`.
- The psql query returns exactly one row with id `00000000-0000-0000-0000-000000000001`, late_threshold_hour=9, late_threshold_minute=0, company_address NULL.

- [ ] **Step 5: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/seed_service.go && \
  git commit -m "feat(seed): ensure system_config singleton row on boot"
```
Expected: 1 file changed.

---

### Task 9: Wire service into `cmd/server/main.go`

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Read the current main.go**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  cat cmd/server/main.go
```
Expected: prints the file. Identify the section that builds repositories then services then handlers and locate where Phase 7's `announcementSvc` (or similar) is constructed.

- [ ] **Step 2: Build the new repo + service and pass it to `NewHandlers`**

In `cmd/server/main.go`, after the existing repository / service construction block and **before** the `handlers.NewHandlers(...)` call, add:

```go
orgSettingsRepo := repositories.NewSystemConfigRepository(db)
orgSettingsSvc := services.NewOrganizationSettingsService(orgSettingsRepo)
```

Then add `orgSettingsSvc` to the existing `handlers.NewHandlers(...)` call (alphabetical position among the other service args).

Ensure the `import` block contains both `"github.com/exnodes/hrm-api/internal/repositories"` and `"github.com/exnodes/hrm-api/internal/services"` — they should already be present from Phases 1-7.

- [ ] **Step 3: Compile-check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Regenerate Swagger and confirm endpoints appear**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  make swag && \
  grep -c '/api/v1/organization-settings/attendance' docs/swagger/swagger.json && \
  grep -c '/api/v1/organization-settings/company-profile' docs/swagger/swagger.json
```
Expected: `make swag` ends with `generated`; both `grep -c` calls print non-zero counts (typically `1` each).

- [ ] **Step 5: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add cmd/server/main.go docs/swagger/docs.go docs/swagger/swagger.json docs/swagger/swagger.yaml && \
  git commit -m "feat(cmd): wire OrganizationSettingsService + regen swagger for /organization-settings"
```
Expected: 4 files changed (or 1 + 3 modified).

---

### Task 10: Service tests `internal/services/organization_settings_service_test.go`

**Files:**
- Create: `internal/services/organization_settings_service_test.go`

- [ ] **Step 1: Write the tests**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/organization_settings_service_test.go` with exactly:
```go
package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// newOrgSettingsServiceForTest builds the service against the shared
// test DB, calls EnsureExists once, and returns the service plus a
// teardown that truncates the row so other tests start fresh.
func newOrgSettingsServiceForTest(t *testing.T) (OrganizationSettingsService, func()) {
	t.Helper()
	db := setupTestDB(t) // Phase 0/1 helper from testhelper_test.go

	repo := repositories.NewSystemConfigRepository(db)
	svc := NewOrganizationSettingsService(repo)
	require.NoError(t, svc.EnsureExists(context.Background()))

	teardown := func() {
		// Reset to defaults so tests are order-independent.
		_, err := repo.Update(context.Background(), map[string]any{
			"late_threshold_hour":         int16(9),
			"late_threshold_minute":       int16(0),
			"company_address":             nil,
			"company_latitude":            nil,
			"company_longitude":           nil,
			"company_address_updated_at":  nil,
			"company_address_updated_by":  nil,
		})
		require.NoError(t, err)
	}
	return svc, teardown
}

func ptrInt(v int) *int          { return &v }
func ptrStr(v string) *string    { return &v }
func ptrFloat(v float64) *float64 { return &v }

func TestOrgSettings_GetAttendance_Defaults(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	got, err := svc.GetAttendance(context.Background())
	require.NoError(t, err)
	require.Equal(t, 9, got.LateThresholdHour)
	require.Equal(t, 0, got.LateThresholdMinute)
	require.Equal(t, "9:00 AM", got.LateThresholdDisplay)
}

func TestOrgSettings_UpdateAttendance_Roundtrip(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	out, err := svc.UpdateAttendance(context.Background(), dto.AttendanceSettingsUpdate{
		LateThresholdHour:   ptrInt(10),
		LateThresholdMinute: ptrInt(30),
	})
	require.NoError(t, err)
	require.Equal(t, 10, out.LateThresholdHour)
	require.Equal(t, 30, out.LateThresholdMinute)
	require.Equal(t, "10:30 AM", out.LateThresholdDisplay)

	again, err := svc.GetAttendance(context.Background())
	require.NoError(t, err)
	require.Equal(t, 10, again.LateThresholdHour)
	require.Equal(t, 30, again.LateThresholdMinute)
}

func TestOrgSettings_UpdateAttendance_PMDisplay(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	out, err := svc.UpdateAttendance(context.Background(), dto.AttendanceSettingsUpdate{
		LateThresholdHour:   ptrInt(13),
		LateThresholdMinute: ptrInt(5),
	})
	require.NoError(t, err)
	require.Equal(t, "1:05 PM", out.LateThresholdDisplay)
}

func TestOrgSettings_UpdateAttendance_RejectsOutOfRange(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	// hour
	_, err := svc.UpdateAttendance(context.Background(), dto.AttendanceSettingsUpdate{
		LateThresholdHour:   ptrInt(24),
		LateThresholdMinute: ptrInt(0),
	})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// minute
	_, err = svc.UpdateAttendance(context.Background(), dto.AttendanceSettingsUpdate{
		LateThresholdHour:   ptrInt(9),
		LateThresholdMinute: ptrInt(60),
	})
	require.Error(t, err)
	ae, ok = apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestOrgSettings_UpdateAttendance_RejectsMissingFields(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	_, err := svc.UpdateAttendance(context.Background(), dto.AttendanceSettingsUpdate{
		LateThresholdHour:   nil,
		LateThresholdMinute: ptrInt(0),
	})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestOrgSettings_GetCompanyProfile_DefaultsAreNil(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	got, err := svc.GetCompanyProfile(context.Background())
	require.NoError(t, err)
	require.Nil(t, got.CompanyAddress)
	require.Nil(t, got.CompanyLatitude)
	require.Nil(t, got.CompanyLongitude)
	require.Nil(t, got.CompanyAddressUpdatedAt)
}

func TestOrgSettings_UpdateCompanyProfile_Roundtrip(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	user := uuid.New()
	out, err := svc.UpdateCompanyProfile(context.Background(),
		dto.CompanyProfileUpdate{
			CompanyAddress:   ptrStr("123 Le Loi, Q1, HCMC"),
			CompanyLatitude:  ptrFloat(10.7769),
			CompanyLongitude: ptrFloat(106.7009),
		},
		user,
	)
	require.NoError(t, err)
	require.NotNil(t, out.CompanyAddress)
	require.Equal(t, "123 Le Loi, Q1, HCMC", *out.CompanyAddress)
	require.NotNil(t, out.CompanyLatitude)
	require.InDelta(t, 10.7769, *out.CompanyLatitude, 1e-9)
	require.NotNil(t, out.CompanyLongitude)
	require.InDelta(t, 106.7009, *out.CompanyLongitude, 1e-9)
	require.NotNil(t, out.CompanyAddressUpdatedAt)

	// Round-trip through Get
	again, err := svc.GetCompanyProfile(context.Background())
	require.NoError(t, err)
	require.Equal(t, *out.CompanyAddress, *again.CompanyAddress)
}

func TestOrgSettings_UpdateCompanyProfile_ClearsFields(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	user := uuid.New()
	// Populate then clear
	_, err := svc.UpdateCompanyProfile(context.Background(),
		dto.CompanyProfileUpdate{
			CompanyAddress:   ptrStr("Old"),
			CompanyLatitude:  ptrFloat(1.0),
			CompanyLongitude: ptrFloat(2.0),
		}, user)
	require.NoError(t, err)

	out, err := svc.UpdateCompanyProfile(context.Background(),
		dto.CompanyProfileUpdate{},
		user)
	require.NoError(t, err)
	require.Nil(t, out.CompanyAddress)
	require.Nil(t, out.CompanyLatitude)
	require.Nil(t, out.CompanyLongitude)
}

func TestOrgSettings_EnsureExists_Idempotent(t *testing.T) {
	svc, teardown := newOrgSettingsServiceForTest(t)
	defer teardown()

	// Already called by the helper; calling again must not error nor
	// produce a duplicate row (the DB-level singleton CHECK guards us).
	require.NoError(t, svc.EnsureExists(context.Background()))
	require.NoError(t, svc.EnsureExists(context.Background()))
}

// Smoke-check that we get a real go error (not nil) when the wrapped
// gorm error path triggers. We exercise this through the package-private
// wrapNotFound helper because the public API can't get into that state
// without dropping the row out-of-band.
func TestOrgSettings_WrapNotFound_BecomesInternal(t *testing.T) {
	err := wrapNotFound(errors.New("other"))
	require.Error(t, err)
}
```

- [ ] **Step 2: Run the new tests**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go test -run TestOrgSettings ./internal/services/... -count=1 -v
```
Expected: every `TestOrgSettings_*` test passes (`--- PASS`); final line `ok  github.com/exnodes/hrm-api/internal/services ...`.

- [ ] **Step 3: Run the full service test suite to confirm no regressions**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go test ./internal/services/... -count=1
```
Expected: `ok  github.com/exnodes/hrm-api/internal/services ...` — every existing Phase 1-7 test still passes.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/organization_settings_service_test.go && \
  git commit -m "test(service): organization settings service (get/update attendance + company profile)"
```
Expected: 1 file changed.

---

### Task 11: End-to-end verification — boot, exercise endpoints, capture log

**Files:**
- Create: `docs/superpowers/verification/phase-08.md`

- [ ] **Step 1: Re-run migrations and boot the server**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  make migrate-up && \
  (make run > /tmp/phase08-server.log 2>&1 &) && \
  sleep 3 && \
  grep -F 'exnodes-hrm-api listening on' /tmp/phase08-server.log
```
Expected: migration log includes `10/u create_system_config` (or "no change" if already applied); the server log contains the `listening on :8080` line.

- [ ] **Step 2: Login as the seeded super-admin and capture the token**

Run:
```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$(grep -E '^SUPER_ADMIN_EMAIL=' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env | cut -d= -f2-)\",\"password\":\"$(grep -E '^SUPER_ADMIN_PASSWORD=' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env | cut -d= -f2-)\"}" \
  | tee /tmp/phase08-login.json \
  | sed -E 's/.*"access_token":"([^"]+)".*/\1/')
echo "TOKEN_LEN=${#ADMIN_TOKEN}"
```
Expected: `/tmp/phase08-login.json` is a `{success:true, data:{access_token:..., refresh_token:...}}` envelope; `TOKEN_LEN` is > 50.

If the env var names differ from Phase 1 (`SUPER_ADMIN_EMAIL` / `SUPER_ADMIN_PASSWORD`), substitute the correct ones from `.env.example`. Do NOT skip — the rest of the verification needs a real token.

- [ ] **Step 3: Create a regular (non-admin) user and capture their token**

Run:
```bash
# Create the user via the admin token
curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"phase08.user@test.local","password":"Phase08!@Pwd","full_name":"Phase 08 User"}' \
  | tee /tmp/phase08-user-create.json

# Login as that user
USER_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"phase08.user@test.local","password":"Phase08!@Pwd"}' \
  | sed -E 's/.*"access_token":"([^"]+)".*/\1/')
echo "USER_TOKEN_LEN=${#USER_TOKEN}"
```
Expected: user-create response has `success:true`; `USER_TOKEN_LEN` is > 50. (If a user with that email already exists from a previous Phase 8 run, the create returns 409 — proceed to the login step regardless.)

- [ ] **Step 4: Exercise the four endpoints — happy path**

Run:
```bash
{
echo '--- GET attendance (admin) ---'
curl -s -i -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/organization-settings/attendance

echo
echo '--- PATCH attendance (admin) ---'
curl -s -i -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -X PATCH http://localhost:8080/api/v1/organization-settings/attendance \
  -d '{"late_threshold_hour":10,"late_threshold_minute":30}'

echo
echo '--- GET attendance again (admin, confirms change) ---'
curl -s -i -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/organization-settings/attendance

echo
echo '--- GET company-profile (non-admin user — must succeed) ---'
curl -s -i -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/api/v1/organization-settings/company-profile

echo
echo '--- PATCH company-profile (admin) ---'
curl -s -i -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -X PATCH http://localhost:8080/api/v1/organization-settings/company-profile \
  -d '{"company_address":"123 Le Loi, Q1, HCMC","company_latitude":10.7769,"company_longitude":106.7009}'

echo
echo '--- GET company-profile (non-admin, confirms change) ---'
curl -s -i -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/api/v1/organization-settings/company-profile
} | tee /tmp/phase08-happy.log
```
Expected:
- Every block returns `HTTP/1.1 200 OK`.
- First GET attendance: `late_threshold_hour:9, late_threshold_minute:0, late_threshold_display:"9:00 AM"`.
- After PATCH: `late_threshold_hour:10, late_threshold_minute:30, late_threshold_display:"10:30 AM"`.
- Company-profile after PATCH: address `123 Le Loi, Q1, HCMC`, lat `10.7769`, lng `106.7009`, `company_address_updated_at` non-null.

- [ ] **Step 5: Exercise the error paths**

Run:
```bash
{
echo '--- PATCH attendance as non-admin (403 expected) ---'
curl -s -i -H "Authorization: Bearer $USER_TOKEN" -H 'Content-Type: application/json' \
  -X PATCH http://localhost:8080/api/v1/organization-settings/attendance \
  -d '{"late_threshold_hour":8,"late_threshold_minute":0}'

echo
echo '--- PATCH company-profile as non-admin (403 expected) ---'
curl -s -i -H "Authorization: Bearer $USER_TOKEN" -H 'Content-Type: application/json' \
  -X PATCH http://localhost:8080/api/v1/organization-settings/company-profile \
  -d '{"company_address":"hacked"}'

echo
echo '--- GET attendance with no token (401 expected) ---'
curl -s -i http://localhost:8080/api/v1/organization-settings/attendance

echo
echo '--- PATCH attendance with bad payload (400 expected) ---'
curl -s -i -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -X PATCH http://localhost:8080/api/v1/organization-settings/attendance \
  -d '{"late_threshold_hour":99,"late_threshold_minute":0}'
} | tee /tmp/phase08-errors.log
```
Expected, in order:
- HTTP 403, envelope `{success:false, code:"forbidden", ...}`
- HTTP 403, envelope `{success:false, code:"forbidden", ...}`
- HTTP 401, envelope `{success:false, code:"unauthorized", ...}`
- HTTP 400, envelope `{success:false, code:"bad_request", ...}`

- [ ] **Step 6: DB spot-check**

Run:
```bash
psql "$(grep -E '^DATABASE_URL=' .env | cut -d= -f2-)" \
  -c "SELECT id, late_threshold_hour, late_threshold_minute, company_address, company_latitude, company_longitude, company_address_updated_by FROM system_config;" \
  | tee /tmp/phase08-db.log
```
Expected: exactly one row, with the values set by Step 4 PATCH calls and `company_address_updated_by` equal to the super-admin's UUID.

- [ ] **Step 7: Stop the server**

Run:
```bash
pkill -f 'go run ./cmd/server' 2>/dev/null || true
sleep 1
lsof -i :8080 || echo 'STOPPED'
```
Expected last line: `STOPPED` (or no `lsof` output).

- [ ] **Step 8: Write the verification log**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification/phase-08.md` with this content, substituting every `<...>` placeholder with literal output captured in Steps 1-6 (do not leave placeholders in the committed file):
```markdown
# Phase 8 Verification Log

Date: 2026-05-15
Phase: 8 — Organization Settings + System Config
Spec: docs/superpowers/specs/2026-05-15-go-migration-design.md
Plan: docs/superpowers/plans/2026-05-15-phase-08-org-settings.md

## 1. Schema decision

The Python source models system_config as a single Beanie Document with
named fields (not a key-value store), so the Go port uses a single-row
`system_config` table guarded by a CHECK (id = '<sentinel>') constraint.
No key-value system_config endpoints are exposed in this phase.

## 2. Migration

Command:
    make migrate-up && make migrate-version

Output (trimmed):
    <paste 10/u create_system_config line + version=10 output>

## 3. Boot + seed

Command:
    make run

Boot log (relevant lines):
    <paste the "listening on" line from /tmp/phase08-server.log>

`SELECT id, late_threshold_hour, late_threshold_minute FROM system_config;`:
    <paste row showing sentinel id, 9, 0>

## 4. Happy-path E2E

Recorded from /tmp/phase08-happy.log:

    GET  /organization-settings/attendance      (admin)     -> 200 (hour=9, min=0, display="9:00 AM")
    PATCH /organization-settings/attendance     (admin)     -> 200 (hour=10, min=30, display="10:30 AM")
    GET  /organization-settings/attendance      (admin)     -> 200 (confirms persisted change)
    GET  /organization-settings/company-profile (non-admin) -> 200 (read open to any authed user)
    PATCH /organization-settings/company-profile (admin)    -> 200 (address+lat+lng written, updated_at set)
    GET  /organization-settings/company-profile (non-admin) -> 200 (confirms persisted change)

Selected response bodies:
    <paste 2-3 representative JSON envelopes from /tmp/phase08-happy.log>

## 5. Error paths

Recorded from /tmp/phase08-errors.log:

    PATCH attendance as non-admin       -> 403 forbidden
    PATCH company-profile as non-admin  -> 403 forbidden
    GET attendance with no token        -> 401 unauthorized
    PATCH attendance with hour=99       -> 400 bad_request

Selected response bodies:
    <paste the 403 and 400 envelopes>

## 6. DB spot-check

Final state of `system_config` after the happy path:
    <paste output of /tmp/phase08-db.log>

`company_address_updated_by` equals the super-admin user UUID.

## 7. Tests

Command:
    go test ./internal/services/... -count=1

Output (last line):
    ok  github.com/exnodes/hrm-api/internal/services ...

## 8. Swagger UI

Visited: http://localhost:8080/swagger/index.html
- "Organization Settings" tag present
- 4 endpoints listed: GET/PATCH attendance, GET/PATCH company-profile

## 9. Sign-off

All Phase 8 acceptance criteria met:

- [x] migration 000010 up/down committed
- [x] SystemConfig singleton model
- [x] DTOs: AttendanceSettingsRead/Update, CompanyProfileRead/Update
- [x] SystemConfigRepository (Get / EnsureExists / Update)
- [x] OrganizationSettingsService (Get/Update Attendance + CompanyProfile + EnsureExists)
- [x] 4 endpoints with Swagger annotations
- [x] Per-route RequirePerms(PermOrgSettings) on write/admin paths
- [x] GET company-profile open to any authenticated user
- [x] Seed service inserts singleton row on boot
- [x] Service tests pass (Get default, Update partial, range rejection, missing field rejection, idempotent EnsureExists, company profile clear)
- [x] End-to-end verification log captured above
```

- [ ] **Step 9: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add docs/superpowers/verification/phase-08.md && \
  git commit -m "docs: phase 8 verification log (org settings + system config)"
```
Expected: 1 file changed.

---

### Task 12: README update + final tree sanity check

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Append Phase 8 endpoints to the README endpoint table**

Edit `README.md` and add the following rows (anywhere a per-phase endpoint table exists, after Phase 7's announcements). If no such table exists, create a new section titled `## Phase 8 — Organization Settings` containing:

```markdown
## Phase 8 — Organization Settings

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET   | `/api/v1/organization-settings/attendance`      | `organization_settings:manage` | Returns `late_threshold_hour/minute/display`. |
| PATCH | `/api/v1/organization-settings/attendance`      | `organization_settings:manage` | Body: `{late_threshold_hour 0-23, late_threshold_minute 0-59}`. Effective immediately; not retroactive. |
| GET   | `/api/v1/organization-settings/company-profile` | any authenticated user | FE renders the map preview from this. |
| PATCH | `/api/v1/organization-settings/company-profile` | `organization_settings:manage` | Body: `{company_address, company_latitude, company_longitude}`. FE is responsible for geocoding. |

System config is modelled as a **single-row singleton** (sentinel UUID
`00000000-0000-0000-0000-000000000001`), seeded on boot. There is no
key-value system-config table in Phase 8 — the Python source does not
have one.
```

- [ ] **Step 2: Verify the full build still passes**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go build ./... && \
  go test ./internal/services/... -count=1 && \
  echo 'PHASE 8 OK'
```
Expected: final line `PHASE 8 OK`.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add README.md && \
  git commit -m "docs(readme): document phase 8 organization settings endpoints"
```
Expected: 1 file changed.

- [ ] **Step 4: Final log sanity check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git log --oneline | head -15
```
Expected: 12 commits from this phase, ordered: migration → model → dto → repo → service → handler → router wiring → seed → main.go wiring → tests → verification log → README.

---

## Phase 8 Definition of Done

All boxes below must be checked before Phase 8 is considered complete:

- [ ] `go build ./...` exits 0 with no warnings
- [ ] `make migrate-up` applies `000010_create_system_config` cleanly; `make migrate-version` prints `10`
- [ ] `make migrate-down` (then `make migrate-up` again) round-trips cleanly
- [ ] `make run` boots; log includes `listening on :8080`
- [ ] After boot, `SELECT count(*) FROM system_config` returns exactly `1`
- [ ] `system_config_singleton` CHECK constraint exists (verified via `\d+ system_config` or `pg_constraint`)
- [ ] `go test ./internal/services/... -count=1` passes including all `TestOrgSettings_*` tests
- [ ] `/swagger/index.html` lists 4 endpoints under the **Organization Settings** tag
- [ ] **Self end-to-end verification** completed and committed at `docs/superpowers/verification/phase-08.md`:
  - GET attendance (admin) → 200 with defaults
  - PATCH attendance (admin) → 200, persisted in next GET
  - GET company-profile (non-admin) → 200 (open read)
  - PATCH company-profile (admin) → 200, persisted in next GET
  - PATCH attendance (non-admin) → 403
  - PATCH company-profile (non-admin) → 403
  - GET attendance with no token → 401
  - PATCH attendance with hour=99 → 400
  - DB spot-check confirms final row state matches API responses
- [ ] README updated with Phase 8 endpoints
- [ ] Every task ended with a commit (`git log --oneline` shows ≥ 12 new commits since the start of this phase)

Once all boxes are checked, Phase 9 (Email + Invite + Push Notification) may begin.
