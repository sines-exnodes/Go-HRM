# Phase 3 — Departments + Positions

| | |
|---|---|
| Status | Ready |
| Date | 2026-05-15 (re-audited 2026-05-18) |
| Owner | danny.tranhoang@exnodes.vn |
| Spec | `docs/superpowers/specs/2026-05-15-go-migration-design.md` |
| Depends on | Phase 0 (foundation), Phase 1 (auth/RBAC), Phase 2 (users/employees) |
| Target | `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/` |
| Module | `github.com/exnodes/hrm-api` |

---

## Goal

Add the `departments` and `positions` tables, their models / DTOs / repos / services / handlers / routes, and **add the deferred FK constraints from `employees` to those two tables** (the `employees.department_id` and `employees.position_id` columns were created NULLABLE and WITHOUT FK in migration `000003` — see the explicit note at the top of `migrations/000003_create_employees_dependents.up.sql`). Expose full CRUD with per-route permission middleware, and finish with an end-to-end self-verification log.

### Audited current state (read before starting — this is the real codebase)

- **The relation is `employees → departments` / `employees → positions`. NOT `users`.** `internal/models/user.go` is auth-only (no `department_id`). `internal/models/employee.go` already declares `DepartmentID *uuid.UUID`, `PositionID *uuid.UUID`, `ManagerID *uuid.UUID` (lines 38-40) with `gorm:"type:uuid"` and NO FK. Migration `000003` declares `employees.department_id UUID NULL`, `employees.position_id UUID NULL` with `idx_employees_department_id` / `idx_employees_position_id` indexes but no constraint.
- **Next free migration number is `000005`.** Existing: `000001_init_extensions`, `000002_create_roles_users`, `000003_create_employees_dependents`, `000004_phase2_extras`. (The earlier draft said `000004` — that is now taken.)
- **Permission constants already exist** in `internal/permissions/registry.go`: `PermDepartmentsRead/Create/Update/Delete` (`"departments:read"` …) and `PermPositionsRead/Create/Update/Delete` (`"positions:read"` …). They are already in `AllPermissions()` and `PermissionGroups`. **Do not re-add them — verify only.**
- **System role grants already include these perms** in `internal/services/seed_service.go` (`defaultRoles()`): Admin has full `departments:*`/`positions:*`; HR Manager has read/create/update; Manager & Employee have read via the Manager role only (Employee role does NOT carry `departments:read`). **Do not modify role grants.**
- **`set_updated_at()` trigger function** is created once in `000001_init_extensions.up.sql`; every entity table adds `CREATE TRIGGER trg_<table>_set_updated_at BEFORE UPDATE … EXECUTE FUNCTION set_updated_at()`.
- **BaseModel** (`internal/models/base.go`): `ID uuid.UUID` (`gen_random_uuid()`), `CreatedAt`, `UpdatedAt`, `IsDeleted bool`, `DeletedAt *time.Time`. Soft-delete is service-managed; `models.NotDeleted` scope filters `is_deleted = false`.
- **Repo convention**: big/shared entities expose an **interface + lowercase struct impl** (`RoleRepository` interface + `roleRepository`, `EmployeeRepository`, `UserRepository`); small struct-only repos use a concrete `*Repository` pointer (`*DependentRepository`, `*LeaveQuotaRepository`). Departments/positions are first-class CRUD entities → **use the interface + lowercase impl pattern**. Each repo defines its own `notDeleted` helper or uses `Scopes(models.NotDeleted)`.
- **Error package** is `apperrors` (import path `github.com/exnodes/hrm-api/internal/errors`, package name `apperrors`). Constructors: `apperrors.ErrNotFound(resource)`, `ErrBadRequest(msg)`, `ErrConflict(msg)`, `ErrForbidden(msg)`. Codes: `not_found`, `bad_request`, `conflict`. Test assertion: `apperrors.As(err)` → `*apperrors.AppError` with `.Code` / `.Message` / `.HTTP`.
- **Search util already exists**: `pkg/utils/search.go` exports `utils.EscapeILIKE(s)` and `utils.BuildILIKEPattern(s)` (returns `"%escaped%"`). **Do not create `EscapeLike` — use the existing `utils.BuildILIKEPattern`.**
- **Permission middleware signature** is `middleware.RequirePerms(authSvc *services.AuthService, required ...permissions.Permission)` — the **first arg is `authSvc`**. The route group is `authed := v1.Group(""); authed.Use(middleware.JWT(...))` then sub-groups like `authed.Group("/employees")`.
- **Response envelope** (`internal/dto/response.go`): `dto.Response[T]{Success, Message, Data}` and `dto.PaginatedData[T]{Items, Total, Page, PageSize, TotalPages}`. Handlers either build `dto.Response[...]` directly or use the package-local `ok(c, status, data, message)` / `okEmpty(c, message)` helpers in `internal/handlers/employee_handler.go`. Errors are surfaced with `_ = c.Error(err)` and rendered by `middleware.ErrorHandler()`.
- **Service wiring** lives inline in `cmd/server/main.go` (no DI container). Repos → services → handlers → route groups, with `seedSvc.Seed(ctx)` run on boot.
- **Service tests** use a real Postgres test DB via `internal/services/testhelper_test.go` (package `services_test`, `TestMain` applies migrations from `migrations/`, `TEST_DATABASE_URL` gates execution, `truncateAll(t)` resets, `skipIfNoDB(t)` skips when unset). Helpers: `makeRole`, `makeUser`, `makeEmployee`. **`truncateAll` must be extended to also TRUNCATE `departments, positions`.**

### Decision: department tree (`parent_id` self-reference) — YES

The old Go reference (`Exn-hr/backend/internal/models/user.go` Department struct) is **flat** (`Members []Employee`, no parent). The Go-migration design spec does not explicitly define a `parent_id` column. We nonetheless **keep the self-referential `parent_id *uuid.UUID`** because:

1. The v2 phase plan explicitly proposed an org tree, and an org hierarchy is a low-cost, forward-compatible addition (nullable, indexed, `ON DELETE SET NULL`).
2. It does not break parity — flat usage is just `parent_id = NULL` for every row.
3. The delete guard ("reject if it has child departments") is a clean invariant the spec's "delete only when empty" rule already implies.

This is a deliberate spec extension, recorded here. If the controller wants strict parity with the old Go ref, drop `ParentID` / `Parent` / `Children` from the model, the `parent_id` column + index, `assertParent`, the cycle check, `HasChildren`, and the `parent_id` list filter — everything else is unaffected.

### Non-negotiable constraints (enforced below)

1. Versioned SQL only (golang-migrate) — **no `AutoMigrate`**.
2. 4 audit cols + `set_updated_at` trigger + `is_deleted` index on `departments` AND `positions`.
3. Soft-delete sets **both** `is_deleted = true` and `deleted_at = NOW()`.
4. UUID PKs (`gen_random_uuid()`).
5. Each route declares its perms inline via `middleware.RequirePerms(authSvc, permissions.PermXxx)` using the EXISTING constant names.
6. Swagger annotations on every handler incl. `@Security BearerAuth`.
7. Search uses `ILIKE` via the existing `utils.BuildILIKEPattern`.
8. Validate FK existence in the service — `BadRequest` if the referenced department is missing.
9. Reject delete with **409 Conflict** if any employee references the dept/position; dept delete is also rejected if it has child departments or active positions.
10. Definition of Done requires end-to-end self-verification, the log committed to `docs/superpowers/verification/phase-03.md`.

### For agentic workers

Every task is bite-sized, ends with a build/test command + expected output, and a commit. Run commands from the repo root `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/` unless stated. **Local DB context**: Postgres in Docker, user `ennam` / password `ennam_dev_2026`, main DB `exnodes_hrm` (currently at migration v4), test DB `exnodes_hrm_test` (v4). Service tests run with:

```
TEST_DATABASE_URL='postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm_test?sslmode=disable'
```

No placeholders — paste the code verbatim.

---

## Tasks

1. Migration `000005_create_departments_positions` (up + down): `departments` + `positions` tables, then the deferred `ALTER TABLE employees ADD CONSTRAINT` FKs.
2. Verify (do NOT re-add) the dept/position permission constants, groups, and role grants.
3. `internal/models/department.go`.
4. `internal/models/position.go`.
5. `internal/dto/department.go`.
6. `internal/dto/position.go`.
7. `internal/repositories/department_repo.go` (interface + impl).
8. `internal/repositories/position_repo.go` (interface + impl).
9. `internal/services/department_service.go`.
10. `internal/services/position_service.go`.
11. `internal/handlers/department_handler.go` (full Swagger).
12. `internal/handlers/position_handler.go` (full Swagger).
13. Wire repos/services/handlers/routes into `cmd/server/main.go`.
14. Optionally seed a small idempotent default tree in `internal/services/seed_service.go`.
15. Service tests `internal/services/{department_service_test.go, position_service_test.go}` + extend `testhelper_test.go`.
16. Regenerate Swagger, full build + test.
17. End-to-end self-verification → `docs/superpowers/verification/phase-03.md`.
18. Update `README.md` Endpoints section.

---

### Task 1 — Migration `000005_create_departments_positions`

- [x] Confirm the next free migration index:

  ```bash
  ls /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations | sort
  ```

  Expected: highest prefix is `000004_phase2_extras`. So the new files are `000005_*`. If a `000005` already exists, bump in lockstep and update every filename/command below.

- [x] Create `migrations/000005_create_departments_positions.up.sql`:

  ```sql
  -- =========================================================================
  -- 000005_create_departments_positions
  -- departments (self-referential tree), positions (belong to a department).
  -- Also adds the deferred FK constraints on employees.department_id /
  -- employees.position_id (created NULLABLE + index, NO FK, in 000003).
  -- =========================================================================

  -- ---------------- departments ----------------
  CREATE TABLE departments (
      id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
      name        TEXT        NOT NULL,
      description TEXT        NOT NULL DEFAULT '',
      parent_id   UUID        NULL REFERENCES departments(id) ON DELETE SET NULL,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
      deleted_at  TIMESTAMPTZ NULL
  );
  CREATE UNIQUE INDEX uq_departments_name_active
      ON departments (LOWER(name)) WHERE is_deleted = FALSE;
  CREATE INDEX idx_departments_is_deleted ON departments (is_deleted);
  CREATE INDEX idx_departments_parent_id  ON departments (parent_id) WHERE parent_id IS NOT NULL;
  CREATE TRIGGER trg_departments_set_updated_at
      BEFORE UPDATE ON departments
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  -- ---------------- positions ----------------
  CREATE TABLE positions (
      id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
      name          TEXT        NOT NULL,
      description   TEXT        NOT NULL DEFAULT '',
      department_id UUID        NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
      created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted    BOOLEAN     NOT NULL DEFAULT FALSE,
      deleted_at    TIMESTAMPTZ NULL
  );
  CREATE UNIQUE INDEX uq_positions_name_dept_active
      ON positions (department_id, LOWER(name)) WHERE is_deleted = FALSE;
  CREATE INDEX idx_positions_is_deleted    ON positions (is_deleted);
  CREATE INDEX idx_positions_department_id ON positions (department_id);
  CREATE TRIGGER trg_positions_set_updated_at
      BEFORE UPDATE ON positions
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  -- ---------------- deferred FK constraints on employees ----------------
  -- employees.department_id / position_id were created NULLABLE + indexed but
  -- WITHOUT FK in 000003 (deferred to this phase). Add them now.
  ALTER TABLE employees
      ADD CONSTRAINT fk_employees_department
          FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

  ALTER TABLE employees
      ADD CONSTRAINT fk_employees_position
          FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL;
  ```

- [x] Create `migrations/000005_create_departments_positions.down.sql`:

  ```sql
  ALTER TABLE employees DROP CONSTRAINT IF EXISTS fk_employees_position;
  ALTER TABLE employees DROP CONSTRAINT IF EXISTS fk_employees_department;

  DROP TRIGGER IF EXISTS trg_positions_set_updated_at ON positions;
  DROP TABLE IF EXISTS positions;

  DROP TRIGGER IF EXISTS trg_departments_set_updated_at ON departments;
  DROP TABLE IF EXISTS departments;
  ```

- [x] Prove both directions on the main DB:

  ```bash
  cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2
  make migrate-up && make migrate-version
  make migrate-down && make migrate-version
  make migrate-up && make migrate-version
  ```

  Expected: after the final `migrate-up`, version `5`, no `dirty` flag. The intermediate down prints `4`.

- [x] Spot-check the FK constraints exist:

  ```bash
  psql "postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm?sslmode=disable" -c "\d+ employees" | grep -E "fk_employees_(department|position)"
  ```

  Expected: two `FOREIGN KEY` lines referencing `departments(id)` and `positions(id)`, both `ON DELETE SET NULL`.

- [x] Commit:

  ```bash
  git add migrations/000005_create_departments_positions.up.sql migrations/000005_create_departments_positions.down.sql
  git commit -m "feat(migrations): add departments, positions, and deferred employee FK constraints (phase 3)"
  ```

---

### Task 2 — Verify permission constants / groups / role grants (no edits expected)

These were already added in Phase 1/2. This task only **confirms** them so later tasks can rely on the exact symbol names.

- [x] Confirm the constants and groups exist:

  ```bash
  grep -nE 'PermDepartments(Read|Create|Update|Delete)|PermPositions(Read|Create|Update|Delete)' internal/permissions/registry.go
  ```

  Expected: 8 const declarations + their appearance in `AllPermissions()` + the `departments` / `positions` `PermissionGroup` entries. Exact names:
  `PermDepartmentsRead`, `PermDepartmentsCreate`, `PermDepartmentsUpdate`, `PermDepartmentsDelete`,
  `PermPositionsRead`, `PermPositionsCreate`, `PermPositionsUpdate`, `PermPositionsDelete`.

- [x] Confirm role grants already wired:

  ```bash
  grep -nE 'PermDepartments|PermPositions' internal/services/seed_service.go
  ```

  Expected: Admin role grants full `departments:*` + `positions:*`; HR Manager grants read/create/update; Manager grants read. **No change required** — if (and only if) any are unexpectedly missing, add them in the existing `defaultRoles()` slice style and note it in the commit.

- [x] Build (no-op confirmation):

  ```bash
  go build ./...
  ```

  Expected: clean build, no output. **No commit for this task unless a missing constant had to be added.**

---

### Task 3 — `internal/models/department.go`

- [x] Create `internal/models/department.go`:

  ```go
  package models

  import "github.com/google/uuid"

  // Department is an org unit. Self-referential tree via ParentID
  // (nullable; FK ON DELETE SET NULL). Employees reference a department
  // through employees.department_id (FK added in migration 000005).
  type Department struct {
      BaseModel
      Name        string     `gorm:"type:text;not null"            json:"name"`
      Description string     `gorm:"type:text;not null;default:''" json:"description"`
      ParentID    *uuid.UUID `gorm:"type:uuid;index"               json:"parent_id,omitempty"`

      // Relations — preloaded on demand, omitted from JSON when nil.
      Parent   *Department  `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
      Children []Department `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
  }

  func (Department) TableName() string { return "departments" }
  ```

- [x] Build:

  ```bash
  go build ./internal/models/...
  ```

  Expected: clean build.

- [x] Commit:

  ```bash
  git add internal/models/department.go
  git commit -m "feat(models): add Department model with self-referential tree"
  ```

---

### Task 4 — `internal/models/position.go`

- [x] Create `internal/models/position.go`:

  ```go
  package models

  import "github.com/google/uuid"

  // Position is a job role; belongs to exactly one department.
  // Employees reference a position through employees.position_id
  // (FK added in migration 000005).
  type Position struct {
      BaseModel
      Name         string    `gorm:"type:text;not null"            json:"name"`
      Description  string    `gorm:"type:text;not null;default:''" json:"description"`
      DepartmentID uuid.UUID `gorm:"type:uuid;not null;index"      json:"department_id"`

      Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
  }

  func (Position) TableName() string { return "positions" }
  ```

- [x] Build:

  ```bash
  go build ./internal/models/...
  ```

- [x] Commit:

  ```bash
  git add internal/models/position.go
  git commit -m "feat(models): add Position model belonging to a Department"
  ```

---

### Task 5 — `internal/dto/department.go`

- [x] Create `internal/dto/department.go`:

  ```go
  package dto

  import (
      "time"

      "github.com/google/uuid"
  )

  // DepartmentCreate is the request body for POST /api/v1/departments.
  type DepartmentCreate struct {
      Name        string     `json:"name"                  binding:"required,min=1,max=100"`
      Description string     `json:"description,omitempty" binding:"max=1000"`
      ParentID    *uuid.UUID `json:"parent_id,omitempty"`
  }

  // DepartmentUpdate is the request body for PATCH /api/v1/departments/:id.
  // PATCH semantics — only provided fields change. ClearParent makes the
  // department a root (distinguishes "no change" from "make root").
  type DepartmentUpdate struct {
      Name        *string    `json:"name,omitempty"        binding:"omitempty,min=1,max=100"`
      Description *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
      ParentID    *uuid.UUID `json:"parent_id,omitempty"`
      ClearParent bool       `json:"clear_parent,omitempty"`
  }

  // DepartmentRead is the wire shape returned by every department endpoint.
  type DepartmentRead struct {
      ID          uuid.UUID       `json:"id"`
      Name        string          `json:"name"`
      Description string          `json:"description"`
      ParentID    *uuid.UUID      `json:"parent_id,omitempty"`
      Parent      *DepartmentRead `json:"parent,omitempty"`
      CreatedAt   time.Time       `json:"created_at"`
      UpdatedAt   time.Time       `json:"updated_at"`
  }

  // DepartmentListQuery binds the querystring for GET /api/v1/departments.
  // ParentID == "root" (or "null") returns top-level departments only.
  type DepartmentListQuery struct {
      Page     int    `form:"page,default=1"       binding:"min=1"`
      PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
      Search   string `form:"search"`
      ParentID string `form:"parent_id"`
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/dto/...
  ```

- [x] Commit:

  ```bash
  git add internal/dto/department.go
  git commit -m "feat(dto): add department request/response shapes"
  ```

---

### Task 6 — `internal/dto/position.go`

- [x] Create `internal/dto/position.go`:

  ```go
  package dto

  import (
      "time"

      "github.com/google/uuid"
  )

  type PositionCreate struct {
      Name         string    `json:"name"                  binding:"required,min=1,max=100"`
      Description  string    `json:"description,omitempty" binding:"max=1000"`
      DepartmentID uuid.UUID `json:"department_id"         binding:"required"`
  }

  type PositionUpdate struct {
      Name         *string    `json:"name,omitempty"          binding:"omitempty,min=1,max=100"`
      Description  *string    `json:"description,omitempty"   binding:"omitempty,max=1000"`
      DepartmentID *uuid.UUID `json:"department_id,omitempty"`
  }

  type PositionRead struct {
      ID           uuid.UUID       `json:"id"`
      Name         string          `json:"name"`
      Description  string          `json:"description"`
      DepartmentID uuid.UUID       `json:"department_id"`
      Department   *DepartmentRead `json:"department,omitempty"`
      CreatedAt    time.Time       `json:"created_at"`
      UpdatedAt    time.Time       `json:"updated_at"`
  }

  type PositionListQuery struct {
      Page         int        `form:"page,default=1"       binding:"min=1"`
      PageSize     int        `form:"page_size,default=10" binding:"min=1,max=100"`
      Search       string     `form:"search"`
      DepartmentID *uuid.UUID `form:"department_id"`
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/dto/...
  ```

- [x] Commit:

  ```bash
  git add internal/dto/position.go
  git commit -m "feat(dto): add position request/response shapes"
  ```

---

### Task 7 — `internal/repositories/department_repo.go`

Note: `HasEmployees` queries the **`employees`** table (the FK lives on `employees`, NOT `users`). Uses the existing `utils.BuildILIKEPattern` and `models.NotDeleted`.

- [x] Create `internal/repositories/department_repo.go`:

  ```go
  package repositories

  import (
      "context"
      "errors"
      "strings"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  // DepartmentFilter mirrors dto.DepartmentListQuery in a service-agnostic shape.
  // ParentID semantics:
  //   nil           → no filter
  //   &uuid.Nil     → top-level only (parent_id IS NULL)
  //   &realUUID     → children of that parent
  type DepartmentFilter struct {
      Page     int
      PageSize int
      Search   string
      ParentID *uuid.UUID
  }

  type DepartmentRepository interface {
      Create(ctx context.Context, d *models.Department) error
      Update(ctx context.Context, d *models.Department) error
      SoftDelete(ctx context.Context, id uuid.UUID) error
      FindByID(ctx context.Context, id uuid.UUID, preloadParent bool) (*models.Department, error)
      FindByName(ctx context.Context, name string) (*models.Department, error)
      List(ctx context.Context, f DepartmentFilter) ([]models.Department, int64, error)
      HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
      // CountEmployees counts non-deleted employees whose department_id == id.
      CountEmployees(ctx context.Context, id uuid.UUID) (int64, error)
  }

  type departmentRepository struct{ db *gorm.DB }

  func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
      return &departmentRepository{db: db}
  }

  func (r *departmentRepository) base(ctx context.Context) *gorm.DB {
      return r.db.WithContext(ctx).Scopes(models.NotDeleted)
  }

  func (r *departmentRepository) Create(ctx context.Context, d *models.Department) error {
      return r.db.WithContext(ctx).Create(d).Error
  }

  func (r *departmentRepository) Update(ctx context.Context, d *models.Department) error {
      return r.db.WithContext(ctx).Save(d).Error
  }

  func (r *departmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
      return r.db.WithContext(ctx).
          Model(&models.Department{}).
          Where("id = ? AND is_deleted = ?", id, false).
          Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
  }

  func (r *departmentRepository) FindByID(ctx context.Context, id uuid.UUID, preloadParent bool) (*models.Department, error) {
      q := r.base(ctx)
      if preloadParent {
          q = q.Preload("Parent", models.NotDeleted)
      }
      var d models.Department
      if err := q.Where("id = ?", id).First(&d).Error; err != nil {
          return nil, err
      }
      return &d, nil
  }

  // FindByName returns (nil, nil) when no active row matches — callers treat
  // that as "available".
  func (r *departmentRepository) FindByName(ctx context.Context, name string) (*models.Department, error) {
      var d models.Department
      err := r.base(ctx).
          Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
          First(&d).Error
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, nil
          }
          return nil, err
      }
      return &d, nil
  }

  func (r *departmentRepository) List(ctx context.Context, f DepartmentFilter) ([]models.Department, int64, error) {
      q := r.base(ctx).Model(&models.Department{})
      if s := strings.TrimSpace(f.Search); s != "" {
          q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
      }
      if f.ParentID != nil {
          if *f.ParentID == uuid.Nil {
              q = q.Where("parent_id IS NULL")
          } else {
              q = q.Where("parent_id = ?", *f.ParentID)
          }
      }

      var total int64
      if err := q.Count(&total).Error; err != nil {
          return nil, 0, err
      }

      page := f.Page
      if page < 1 {
          page = 1
      }
      size := f.PageSize
      if size < 1 {
          size = 10
      }
      var items []models.Department
      err := q.
          Preload("Parent", models.NotDeleted).
          Order("LOWER(name) ASC").
          Offset((page - 1) * size).
          Limit(size).
          Find(&items).Error
      return items, total, err
  }

  func (r *departmentRepository) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
      var count int64
      err := r.base(ctx).
          Model(&models.Department{}).
          Where("parent_id = ?", id).
          Count(&count).Error
      return count > 0, err
  }

  func (r *departmentRepository) CountEmployees(ctx context.Context, id uuid.UUID) (int64, error) {
      var count int64
      err := r.db.WithContext(ctx).
          Model(&models.Employee{}).
          Where("department_id = ? AND is_deleted = ?", id, false).
          Count(&count).Error
      return count, err
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/repositories/...
  ```

- [x] Commit:

  ```bash
  git add internal/repositories/department_repo.go
  git commit -m "feat(repositories): add department repo with tree, soft-delete, ILIKE search, employee-count guard"
  ```

---

### Task 8 — `internal/repositories/position_repo.go`

Note: `CountEmployees` queries the **`employees`** table.

- [x] Create `internal/repositories/position_repo.go`:

  ```go
  package repositories

  import (
      "context"
      "errors"
      "strings"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  type PositionFilter struct {
      Page         int
      PageSize     int
      Search       string
      DepartmentID *uuid.UUID
  }

  type PositionRepository interface {
      Create(ctx context.Context, p *models.Position) error
      Update(ctx context.Context, p *models.Position) error
      SoftDelete(ctx context.Context, id uuid.UUID) error
      FindByID(ctx context.Context, id uuid.UUID, preloadDept bool) (*models.Position, error)
      FindByNameInDept(ctx context.Context, name string, departmentID uuid.UUID) (*models.Position, error)
      List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error)
      CountByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error)
      // CountEmployees counts non-deleted employees whose position_id == id.
      CountEmployees(ctx context.Context, id uuid.UUID) (int64, error)
  }

  type positionRepository struct{ db *gorm.DB }

  func NewPositionRepository(db *gorm.DB) PositionRepository {
      return &positionRepository{db: db}
  }

  func (r *positionRepository) base(ctx context.Context) *gorm.DB {
      return r.db.WithContext(ctx).Scopes(models.NotDeleted)
  }

  func (r *positionRepository) Create(ctx context.Context, p *models.Position) error {
      return r.db.WithContext(ctx).Create(p).Error
  }

  func (r *positionRepository) Update(ctx context.Context, p *models.Position) error {
      return r.db.WithContext(ctx).Save(p).Error
  }

  func (r *positionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
      return r.db.WithContext(ctx).
          Model(&models.Position{}).
          Where("id = ? AND is_deleted = ?", id, false).
          Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
  }

  func (r *positionRepository) FindByID(ctx context.Context, id uuid.UUID, preloadDept bool) (*models.Position, error) {
      q := r.base(ctx)
      if preloadDept {
          q = q.Preload("Department", models.NotDeleted)
      }
      var p models.Position
      if err := q.Where("id = ?", id).First(&p).Error; err != nil {
          return nil, err
      }
      return &p, nil
  }

  func (r *positionRepository) FindByNameInDept(ctx context.Context, name string, departmentID uuid.UUID) (*models.Position, error) {
      var p models.Position
      err := r.base(ctx).
          Where("LOWER(name) = LOWER(?) AND department_id = ?", strings.TrimSpace(name), departmentID).
          First(&p).Error
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, nil
          }
          return nil, err
      }
      return &p, nil
  }

  func (r *positionRepository) List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error) {
      q := r.base(ctx).Model(&models.Position{})
      if s := strings.TrimSpace(f.Search); s != "" {
          q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
      }
      if f.DepartmentID != nil {
          q = q.Where("department_id = ?", *f.DepartmentID)
      }

      var total int64
      if err := q.Count(&total).Error; err != nil {
          return nil, 0, err
      }
      page := f.Page
      if page < 1 {
          page = 1
      }
      size := f.PageSize
      if size < 1 {
          size = 10
      }
      var items []models.Position
      err := q.
          Preload("Department", models.NotDeleted).
          Order("LOWER(name) ASC").
          Offset((page - 1) * size).
          Limit(size).
          Find(&items).Error
      return items, total, err
  }

  func (r *positionRepository) CountByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error) {
      var count int64
      err := r.base(ctx).
          Model(&models.Position{}).
          Where("department_id = ?", departmentID).
          Count(&count).Error
      return count, err
  }

  func (r *positionRepository) CountEmployees(ctx context.Context, id uuid.UUID) (int64, error) {
      var count int64
      err := r.db.WithContext(ctx).
          Model(&models.Employee{}).
          Where("position_id = ? AND is_deleted = ?", id, false).
          Count(&count).Error
      return count, err
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/repositories/...
  ```

- [x] Commit:

  ```bash
  git add internal/repositories/position_repo.go
  git commit -m "feat(repositories): add position repo with department filter and employee-count guard"
  ```

---

### Task 9 — `internal/services/department_service.go`

Error package is `apperrors` (package name) at import path `internal/errors`. Delete rejects when children, positions, or employees reference the department.

- [x] Create `internal/services/department_service.go`:

  ```go
  package services

  import (
      "context"
      "errors"
      "fmt"
      "strings"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/dto"
      apperrors "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/repositories"
  )

  // DepartmentService owns department business logic. It also holds the
  // position repo so a single Delete call can enforce the cross-aggregate
  // invariant (a department with active positions cannot be deleted).
  type DepartmentService struct {
      repo    repositories.DepartmentRepository
      posRepo repositories.PositionRepository
  }

  func NewDepartmentService(repo repositories.DepartmentRepository, posRepo repositories.PositionRepository) *DepartmentService {
      return &DepartmentService{repo: repo, posRepo: posRepo}
  }

  func departmentToRead(d *models.Department) dto.DepartmentRead {
      out := dto.DepartmentRead{
          ID:          d.ID,
          Name:        d.Name,
          Description: d.Description,
          ParentID:    d.ParentID,
          CreatedAt:   d.CreatedAt,
          UpdatedAt:   d.UpdatedAt,
      }
      if d.Parent != nil {
          p := departmentToRead(d.Parent)
          out.Parent = &p
      }
      return out
  }

  func (s *DepartmentService) checkNameUnique(ctx context.Context, name string, excludeID *uuid.UUID) error {
      existing, err := s.repo.FindByName(ctx, name)
      if err != nil {
          return err
      }
      if existing == nil {
          return nil
      }
      if excludeID != nil && existing.ID == *excludeID {
          return nil
      }
      return apperrors.ErrConflict("Department name already exists")
  }

  // assertParent verifies the proposed parent exists and that setting it would
  // not create a cycle (only relevant when updating an existing node).
  func (s *DepartmentService) assertParent(ctx context.Context, parentID uuid.UUID, selfID *uuid.UUID) error {
      if selfID != nil && parentID == *selfID {
          return apperrors.ErrBadRequest("Department cannot be its own parent")
      }
      parent, err := s.repo.FindByID(ctx, parentID, false)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperrors.ErrBadRequest("Parent department not found")
          }
          return err
      }
      if selfID != nil {
          current := parent
          for current.ParentID != nil {
              if *current.ParentID == *selfID {
                  return apperrors.ErrBadRequest("Setting this parent would create a cycle")
              }
              next, err := s.repo.FindByID(ctx, *current.ParentID, false)
              if err != nil {
                  return err
              }
              current = next
          }
      }
      return nil
  }

  func (s *DepartmentService) Create(ctx context.Context, in dto.DepartmentCreate) (*dto.DepartmentRead, error) {
      name := strings.TrimSpace(in.Name)
      if name == "" {
          return nil, apperrors.ErrBadRequest("Department name cannot be blank")
      }
      if err := s.checkNameUnique(ctx, name, nil); err != nil {
          return nil, err
      }
      if in.ParentID != nil {
          if err := s.assertParent(ctx, *in.ParentID, nil); err != nil {
              return nil, err
          }
      }
      d := &models.Department{
          Name:        name,
          Description: strings.TrimSpace(in.Description),
          ParentID:    in.ParentID,
      }
      if err := s.repo.Create(ctx, d); err != nil {
          return nil, err
      }
      fresh, err := s.repo.FindByID(ctx, d.ID, true)
      if err != nil {
          return nil, err
      }
      out := departmentToRead(fresh)
      return &out, nil
  }

  func (s *DepartmentService) Update(ctx context.Context, id uuid.UUID, in dto.DepartmentUpdate) (*dto.DepartmentRead, error) {
      d, err := s.repo.FindByID(ctx, id, false)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperrors.ErrNotFound("Department")
          }
          return nil, err
      }
      if in.Name != nil {
          name := strings.TrimSpace(*in.Name)
          if name == "" {
              return nil, apperrors.ErrBadRequest("Department name cannot be blank")
          }
          if err := s.checkNameUnique(ctx, name, &d.ID); err != nil {
              return nil, err
          }
          d.Name = name
      }
      if in.Description != nil {
          d.Description = strings.TrimSpace(*in.Description)
      }
      switch {
      case in.ClearParent:
          d.ParentID = nil
      case in.ParentID != nil:
          if err := s.assertParent(ctx, *in.ParentID, &d.ID); err != nil {
              return nil, err
          }
          d.ParentID = in.ParentID
      }
      if err := s.repo.Update(ctx, d); err != nil {
          return nil, err
      }
      fresh, err := s.repo.FindByID(ctx, d.ID, true)
      if err != nil {
          return nil, err
      }
      out := departmentToRead(fresh)
      return &out, nil
  }

  // Delete soft-deletes the department after verifying it has no child
  // departments, no active positions, and no assigned employees. Any of those
  // returns a 409 Conflict.
  func (s *DepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
      if _, err := s.repo.FindByID(ctx, id, false); err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperrors.ErrNotFound("Department")
          }
          return err
      }

      hasChildren, err := s.repo.HasChildren(ctx, id)
      if err != nil {
          return err
      }
      if hasChildren {
          return apperrors.ErrConflict("Cannot delete department — it has child departments. Move or delete them first.")
      }

      posCount, err := s.posRepo.CountByDepartment(ctx, id)
      if err != nil {
          return err
      }
      if posCount > 0 {
          word := "position is"
          if posCount > 1 {
              word = "positions are"
          }
          return apperrors.ErrConflict(fmt.Sprintf(
              "Cannot delete — %d %s assigned to this department. Delete or reassign them first.", posCount, word))
      }

      empCount, err := s.repo.CountEmployees(ctx, id)
      if err != nil {
          return err
      }
      if empCount > 0 {
          word := "employee is"
          if empCount > 1 {
              word = "employees are"
          }
          return apperrors.ErrConflict(fmt.Sprintf(
              "Cannot delete — %d %s assigned to this department. Reassign all employees before deleting.", empCount, word))
      }
      return s.repo.SoftDelete(ctx, id)
  }

  func (s *DepartmentService) Get(ctx context.Context, id uuid.UUID) (*dto.DepartmentRead, error) {
      d, err := s.repo.FindByID(ctx, id, true)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperrors.ErrNotFound("Department")
          }
          return nil, err
      }
      out := departmentToRead(d)
      return &out, nil
  }

  func (s *DepartmentService) List(ctx context.Context, q dto.DepartmentListQuery) (*dto.PaginatedData[dto.DepartmentRead], error) {
      f := repositories.DepartmentFilter{Page: q.Page, PageSize: q.PageSize, Search: q.Search}
      switch strings.ToLower(strings.TrimSpace(q.ParentID)) {
      case "":
          // no parent filter
      case "root", "null":
          nilUUID := uuid.Nil
          f.ParentID = &nilUUID
      default:
          parsed, err := uuid.Parse(q.ParentID)
          if err != nil {
              return nil, apperrors.ErrBadRequest("Invalid parent_id")
          }
          f.ParentID = &parsed
      }

      items, total, err := s.repo.List(ctx, f)
      if err != nil {
          return nil, err
      }
      reads := make([]dto.DepartmentRead, 0, len(items))
      for i := range items {
          reads = append(reads, departmentToRead(&items[i]))
      }
      page := q.Page
      if page < 1 {
          page = 1
      }
      size := q.PageSize
      if size < 1 {
          size = 10
      }
      totalPages := 0
      if total > 0 {
          totalPages = int((total + int64(size) - 1) / int64(size))
      }
      return &dto.PaginatedData[dto.DepartmentRead]{
          Items:      reads,
          Total:      total,
          Page:       page,
          PageSize:   size,
          TotalPages: totalPages,
      }, nil
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/services/...
  ```

- [x] Commit:

  ```bash
  git add internal/services/department_service.go
  git commit -m "feat(services): add department service with unique-name, parent cycle, and cascade-delete guards"
  ```

---

### Task 10 — `internal/services/position_service.go`

- [x] Create `internal/services/position_service.go`:

  ```go
  package services

  import (
      "context"
      "errors"
      "fmt"
      "strings"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/dto"
      apperrors "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/repositories"
  )

  type PositionService struct {
      repo     repositories.PositionRepository
      deptRepo repositories.DepartmentRepository
  }

  func NewPositionService(repo repositories.PositionRepository, deptRepo repositories.DepartmentRepository) *PositionService {
      return &PositionService{repo: repo, deptRepo: deptRepo}
  }

  func positionToRead(p *models.Position) dto.PositionRead {
      out := dto.PositionRead{
          ID:           p.ID,
          Name:         p.Name,
          Description:  p.Description,
          DepartmentID: p.DepartmentID,
          CreatedAt:    p.CreatedAt,
          UpdatedAt:    p.UpdatedAt,
      }
      if p.Department != nil {
          d := departmentToRead(p.Department)
          out.Department = &d
      }
      return out
  }

  func (s *PositionService) assertDept(ctx context.Context, deptID uuid.UUID) error {
      if _, err := s.deptRepo.FindByID(ctx, deptID, false); err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperrors.ErrBadRequest("Department not found")
          }
          return err
      }
      return nil
  }

  func (s *PositionService) checkNameUniqueInDept(ctx context.Context, name string, deptID uuid.UUID, excludeID *uuid.UUID) error {
      existing, err := s.repo.FindByNameInDept(ctx, name, deptID)
      if err != nil {
          return err
      }
      if existing == nil {
          return nil
      }
      if excludeID != nil && existing.ID == *excludeID {
          return nil
      }
      return apperrors.ErrConflict("Position name already exists in this department")
  }

  func (s *PositionService) Create(ctx context.Context, in dto.PositionCreate) (*dto.PositionRead, error) {
      name := strings.TrimSpace(in.Name)
      if name == "" {
          return nil, apperrors.ErrBadRequest("Position name cannot be blank")
      }
      if err := s.assertDept(ctx, in.DepartmentID); err != nil {
          return nil, err
      }
      if err := s.checkNameUniqueInDept(ctx, name, in.DepartmentID, nil); err != nil {
          return nil, err
      }
      p := &models.Position{
          Name:         name,
          Description:  strings.TrimSpace(in.Description),
          DepartmentID: in.DepartmentID,
      }
      if err := s.repo.Create(ctx, p); err != nil {
          return nil, err
      }
      fresh, err := s.repo.FindByID(ctx, p.ID, true)
      if err != nil {
          return nil, err
      }
      out := positionToRead(fresh)
      return &out, nil
  }

  func (s *PositionService) Update(ctx context.Context, id uuid.UUID, in dto.PositionUpdate) (*dto.PositionRead, error) {
      p, err := s.repo.FindByID(ctx, id, false)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperrors.ErrNotFound("Position")
          }
          return nil, err
      }
      newDept := p.DepartmentID
      if in.DepartmentID != nil {
          if err := s.assertDept(ctx, *in.DepartmentID); err != nil {
              return nil, err
          }
          newDept = *in.DepartmentID
      }
      if in.Name != nil {
          name := strings.TrimSpace(*in.Name)
          if name == "" {
              return nil, apperrors.ErrBadRequest("Position name cannot be blank")
          }
          if err := s.checkNameUniqueInDept(ctx, name, newDept, &p.ID); err != nil {
              return nil, err
          }
          p.Name = name
      }
      if in.Description != nil {
          p.Description = strings.TrimSpace(*in.Description)
      }
      p.DepartmentID = newDept
      if err := s.repo.Update(ctx, p); err != nil {
          return nil, err
      }
      fresh, err := s.repo.FindByID(ctx, p.ID, true)
      if err != nil {
          return nil, err
      }
      out := positionToRead(fresh)
      return &out, nil
  }

  func (s *PositionService) Delete(ctx context.Context, id uuid.UUID) error {
      if _, err := s.repo.FindByID(ctx, id, false); err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperrors.ErrNotFound("Position")
          }
          return err
      }
      empCount, err := s.repo.CountEmployees(ctx, id)
      if err != nil {
          return err
      }
      if empCount > 0 {
          word := "employee is"
          if empCount > 1 {
              word = "employees are"
          }
          return apperrors.ErrConflict(fmt.Sprintf(
              "Cannot delete — %d %s assigned to this position. Reassign all employees before deleting.", empCount, word))
      }
      return s.repo.SoftDelete(ctx, id)
  }

  func (s *PositionService) Get(ctx context.Context, id uuid.UUID) (*dto.PositionRead, error) {
      p, err := s.repo.FindByID(ctx, id, true)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperrors.ErrNotFound("Position")
          }
          return nil, err
      }
      out := positionToRead(p)
      return &out, nil
  }

  func (s *PositionService) List(ctx context.Context, q dto.PositionListQuery) (*dto.PaginatedData[dto.PositionRead], error) {
      items, total, err := s.repo.List(ctx, repositories.PositionFilter{
          Page:         q.Page,
          PageSize:     q.PageSize,
          Search:       q.Search,
          DepartmentID: q.DepartmentID,
      })
      if err != nil {
          return nil, err
      }
      reads := make([]dto.PositionRead, 0, len(items))
      for i := range items {
          reads = append(reads, positionToRead(&items[i]))
      }
      page := q.Page
      if page < 1 {
          page = 1
      }
      size := q.PageSize
      if size < 1 {
          size = 10
      }
      totalPages := 0
      if total > 0 {
          totalPages = int((total + int64(size) - 1) / int64(size))
      }
      return &dto.PaginatedData[dto.PositionRead]{
          Items:      reads,
          Total:      total,
          Page:       page,
          PageSize:   size,
          TotalPages: totalPages,
      }, nil
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/services/...
  ```

- [x] Commit:

  ```bash
  git add internal/services/position_service.go
  git commit -m "feat(services): add position service with dept-FK validation and employee-count delete guard"
  ```

---

### Task 11 — `internal/handlers/department_handler.go`

Uses `dto.Response[...]` (matches `internal/dto/response.go`) and surfaces errors via `_ = c.Error(err)` so `middleware.ErrorHandler()` renders them. The cross-aggregate position guard is inside the service now (Task 9), so the handler is a thin pass-through.

- [x] Create `internal/handlers/department_handler.go`:

  ```go
  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/google/uuid"

      "github.com/exnodes/hrm-api/internal/dto"
      apperrors "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/services"
  )

  type DepartmentHandler struct {
      svc *services.DepartmentService
  }

  func NewDepartmentHandler(svc *services.DepartmentService) *DepartmentHandler {
      return &DepartmentHandler{svc: svc}
  }

  // List godoc
  // @Summary      List departments
  // @Description  Paginated list with optional name search and parent filter ("root" returns top-level only).
  // @Tags         departments
  // @Security     BearerAuth
  // @Produce      json
  // @Param        page       query    int     false  "Page number"  default(1)
  // @Param        page_size  query    int     false  "Page size"    default(10)
  // @Param        search     query    string  false  "Substring match on name (ILIKE)"
  // @Param        parent_id  query    string  false  "Filter by parent UUID, or \"root\" for top-level"
  // @Success      200  {object}  map[string]interface{}
  // @Router       /api/v1/departments [get]
  func (h *DepartmentHandler) List(c *gin.Context) {
      var q dto.DepartmentListQuery
      if err := c.ShouldBindQuery(&q); err != nil {
          _ = c.Error(apperrors.ErrBadRequest(err.Error()))
          return
      }
      data, err := h.svc.List(c.Request.Context(), q)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.DepartmentRead]]{Success: true, Data: data})
  }

  // Create godoc
  // @Summary      Create department
  // @Tags         departments
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        body  body      dto.DepartmentCreate  true  "Department payload"
  // @Success      201   {object}  map[string]interface{}
  // @Router       /api/v1/departments [post]
  func (h *DepartmentHandler) Create(c *gin.Context) {
      var in dto.DepartmentCreate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(apperrors.ErrBadRequest(err.Error()))
          return
      }
      out, err := h.svc.Create(c.Request.Context(), in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusCreated, dto.Response[*dto.DepartmentRead]{
          Success: true,
          Message: "Department created",
          Data:    out,
      })
  }

  // Get godoc
  // @Summary      Get department by ID
  // @Tags         departments
  // @Security     BearerAuth
  // @Produce      json
  // @Param        id   path      string  true  "Department UUID"
  // @Success      200  {object}  map[string]interface{}
  // @Router       /api/v1/departments/{id} [get]
  func (h *DepartmentHandler) Get(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(apperrors.ErrBadRequest("invalid id"))
          return
      }
      out, err := h.svc.Get(c.Request.Context(), id)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.DepartmentRead]{Success: true, Data: out})
  }

  // Update godoc
  // @Summary      Update department
  // @Tags         departments
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        id    path      string                true  "Department UUID"
  // @Param        body  body      dto.DepartmentUpdate  true  "Fields to update (PATCH semantics)"
  // @Success      200   {object}  map[string]interface{}
  // @Router       /api/v1/departments/{id} [patch]
  func (h *DepartmentHandler) Update(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(apperrors.ErrBadRequest("invalid id"))
          return
      }
      var in dto.DepartmentUpdate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(apperrors.ErrBadRequest(err.Error()))
          return
      }
      out, err := h.svc.Update(c.Request.Context(), id, in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.DepartmentRead]{
          Success: true,
          Message: "Department updated",
          Data:    out,
      })
  }

  // Delete godoc
  // @Summary      Delete department
  // @Description  Soft-deletes a department. Rejected with 409 if it has child departments, active positions, or assigned employees.
  // @Tags         departments
  // @Security     BearerAuth
  // @Produce      json
  // @Param        id   path      string  true  "Department UUID"
  // @Success      200  {object}  map[string]interface{}
  // @Router       /api/v1/departments/{id} [delete]
  func (h *DepartmentHandler) Delete(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(apperrors.ErrBadRequest("invalid id"))
          return
      }
      if err := h.svc.Delete(c.Request.Context(), id); err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Department deleted"})
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/handlers/...
  ```

- [x] Commit:

  ```bash
  git add internal/handlers/department_handler.go
  git commit -m "feat(handlers): add department CRUD handler with Swagger annotations"
  ```

---

### Task 12 — `internal/handlers/position_handler.go`

- [x] Create `internal/handlers/position_handler.go`:

  ```go
  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/google/uuid"

      "github.com/exnodes/hrm-api/internal/dto"
      apperrors "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/services"
  )

  type PositionHandler struct {
      svc *services.PositionService
  }

  func NewPositionHandler(svc *services.PositionService) *PositionHandler {
      return &PositionHandler{svc: svc}
  }

  // List godoc
  // @Summary      List positions
  // @Tags         positions
  // @Security     BearerAuth
  // @Produce      json
  // @Param        page           query    int     false  "Page"        default(1)
  // @Param        page_size      query    int     false  "Page size"   default(10)
  // @Param        search         query    string  false  "Substring match on name"
  // @Param        department_id  query    string  false  "Filter by department UUID"
  // @Success      200  {object}  map[string]interface{}
  // @Router       /api/v1/positions [get]
  func (h *PositionHandler) List(c *gin.Context) {
      var q dto.PositionListQuery
      if err := c.ShouldBindQuery(&q); err != nil {
          _ = c.Error(apperrors.ErrBadRequest(err.Error()))
          return
      }
      data, err := h.svc.List(c.Request.Context(), q)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.PositionRead]]{Success: true, Data: data})
  }

  // Create godoc
  // @Summary      Create position
  // @Tags         positions
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        body  body      dto.PositionCreate  true  "Position payload"
  // @Success      201   {object}  map[string]interface{}
  // @Router       /api/v1/positions [post]
  func (h *PositionHandler) Create(c *gin.Context) {
      var in dto.PositionCreate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(apperrors.ErrBadRequest(err.Error()))
          return
      }
      out, err := h.svc.Create(c.Request.Context(), in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusCreated, dto.Response[*dto.PositionRead]{
          Success: true,
          Message: "Position created",
          Data:    out,
      })
  }

  // Get godoc
  // @Summary      Get position by ID
  // @Tags         positions
  // @Security     BearerAuth
  // @Produce      json
  // @Param        id   path      string  true  "Position UUID"
  // @Success      200  {object}  map[string]interface{}
  // @Router       /api/v1/positions/{id} [get]
  func (h *PositionHandler) Get(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(apperrors.ErrBadRequest("invalid id"))
          return
      }
      out, err := h.svc.Get(c.Request.Context(), id)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.PositionRead]{Success: true, Data: out})
  }

  // Update godoc
  // @Summary      Update position
  // @Tags         positions
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        id    path      string              true  "Position UUID"
  // @Param        body  body      dto.PositionUpdate  true  "Fields to update"
  // @Success      200   {object}  map[string]interface{}
  // @Router       /api/v1/positions/{id} [patch]
  func (h *PositionHandler) Update(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(apperrors.ErrBadRequest("invalid id"))
          return
      }
      var in dto.PositionUpdate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(apperrors.ErrBadRequest(err.Error()))
          return
      }
      out, err := h.svc.Update(c.Request.Context(), id, in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.PositionRead]{
          Success: true,
          Message: "Position updated",
          Data:    out,
      })
  }

  // Delete godoc
  // @Summary      Delete position
  // @Description  Soft-deletes a position. Rejected with 409 if any employee is still assigned.
  // @Tags         positions
  // @Security     BearerAuth
  // @Produce      json
  // @Param        id   path      string  true  "Position UUID"
  // @Success      200  {object}  map[string]interface{}
  // @Router       /api/v1/positions/{id} [delete]
  func (h *PositionHandler) Delete(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(apperrors.ErrBadRequest("invalid id"))
          return
      }
      if err := h.svc.Delete(c.Request.Context(), id); err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Position deleted"})
  }
  ```

- [x] Build:

  ```bash
  go build ./internal/handlers/...
  ```

- [x] Commit:

  ```bash
  git add internal/handlers/position_handler.go
  git commit -m "feat(handlers): add position CRUD handler with Swagger annotations"
  ```

---

### Task 13 — Wire repos/services/handlers/routes in `cmd/server/main.go`

The actual `main.go` builds repos → services → handlers inline, then registers routes inside the `authed := v1.Group("")` block. **`RequirePerms`'s first argument is `authSvc`** (see `internal/middleware/permissions.go`). Mirror the existing `adminEmps` block exactly.

- [x] In `cmd/server/main.go`, in the `// ---- repositories ----` section, after `settingsRepo := ...`, add:

  ```go
  	departmentRepo := repositories.NewDepartmentRepository(db)
  	positionRepo := repositories.NewPositionRepository(db)
  ```

- [x] In the `// ---- services ----` section, after `userSvc := services.NewUserService(...)`, add:

  ```go
  	departmentSvc := services.NewDepartmentService(departmentRepo, positionRepo)
  	positionSvc := services.NewPositionService(positionRepo, departmentRepo)
  ```

- [x] In the `// ---- handlers ----` section, after `userH := handlers.NewUserHandler(userSvc)`, add:

  ```go
  	departmentH := handlers.NewDepartmentHandler(departmentSvc)
  	positionH := handlers.NewPositionHandler(positionSvc)
  ```

- [x] Inside the `authed` block, after the dependents routes (the `authed.DELETE("/employees/:id/dependents/:dependentID", depH.Delete)` line), add:

  ```go
  		// ---- /departments ----
  		departments := authed.Group("/departments")
  		departments.GET("", middleware.RequirePerms(authSvc, permissions.PermDepartmentsRead), departmentH.List)
  		departments.POST("", middleware.RequirePerms(authSvc, permissions.PermDepartmentsCreate), departmentH.Create)
  		departments.GET(":id", middleware.RequirePerms(authSvc, permissions.PermDepartmentsRead), departmentH.Get)
  		departments.PATCH(":id", middleware.RequirePerms(authSvc, permissions.PermDepartmentsUpdate), departmentH.Update)
  		departments.DELETE(":id", middleware.RequirePerms(authSvc, permissions.PermDepartmentsDelete), departmentH.Delete)

  		// ---- /positions ----
  		positions := authed.Group("/positions")
  		positions.GET("", middleware.RequirePerms(authSvc, permissions.PermPositionsRead), positionH.List)
  		positions.POST("", middleware.RequirePerms(authSvc, permissions.PermPositionsCreate), positionH.Create)
  		positions.GET(":id", middleware.RequirePerms(authSvc, permissions.PermPositionsRead), positionH.Get)
  		positions.PATCH(":id", middleware.RequirePerms(authSvc, permissions.PermPositionsUpdate), positionH.Update)
  		positions.DELETE(":id", middleware.RequirePerms(authSvc, permissions.PermPositionsDelete), positionH.Delete)
  ```

- [x] Build the whole module:

  ```bash
  go build ./...
  ```

  Expected: clean build, no output.

- [x] Commit:

  ```bash
  git add cmd/server/main.go
  git commit -m "feat(server): wire department + position repos, services, handlers, and per-route RequirePerms"
  ```

---

### Task 14 — Extend seeder with an idempotent default tree

`SeedService` (`internal/services/seed_service.go`) has no department/position repos; it uses `s.db` directly. Seed only when `departments` is empty (idempotent). Match the existing `log.Printf("seed: …")` style and the `s.Seed(ctx)` entrypoint.

- [x] Add a `seedOrgDefaults` method and call it from `Seed`:

  ```go
  // (in Seed, after seedSuperAdmin)
  func (s *SeedService) Seed(ctx context.Context) error {
      if err := s.seedRoles(ctx); err != nil {
          return err
      }
      if err := s.seedSuperAdmin(ctx); err != nil {
          return err
      }
      if err := s.seedOrgDefaults(ctx); err != nil {
          return err
      }
      return nil
  }

  // seedOrgDefaults inserts a small default department/position tree the first
  // time the departments table is empty. Idempotent: a non-empty table is left
  // untouched so manual edits are never clobbered.
  func (s *SeedService) seedOrgDefaults(ctx context.Context) error {
      var deptCount int64
      if err := s.db.WithContext(ctx).
          Model(&models.Department{}).
          Where("is_deleted = ?", false).
          Count(&deptCount).Error; err != nil {
          return err
      }
      if deptCount > 0 {
          return nil
      }

      eng := &models.Department{Name: "Engineering"}
      hr := &models.Department{Name: "Human Resources"}
      if err := s.db.WithContext(ctx).Create(eng).Error; err != nil {
          return err
      }
      if err := s.db.WithContext(ctx).Create(hr).Error; err != nil {
          return err
      }
      backend := &models.Department{Name: "Backend", ParentID: &eng.ID}
      mobile := &models.Department{Name: "Mobile", ParentID: &eng.ID}
      if err := s.db.WithContext(ctx).Create(backend).Error; err != nil {
          return err
      }
      if err := s.db.WithContext(ctx).Create(mobile).Error; err != nil {
          return err
      }

      positions := []*models.Position{
          {Name: "Software Engineer", DepartmentID: backend.ID},
          {Name: "Mobile Engineer", DepartmentID: mobile.ID},
          {Name: "HR Specialist", DepartmentID: hr.ID},
      }
      for _, p := range positions {
          if err := s.db.WithContext(ctx).Create(p).Error; err != nil {
              return err
          }
      }
      log.Printf("seed: created default org tree (4 departments, 3 positions)")
      return nil
  }
  ```

- [x] Build + run the existing seed test to confirm nothing broke:

  ```bash
  go build ./... && TEST_DATABASE_URL='postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm_test?sslmode=disable' go test ./internal/services/ -run TestSeed -count=1
  ```

  Expected: `ok  github.com/exnodes/hrm-api/internal/services`.

- [x] Commit:

  ```bash
  git add internal/services/seed_service.go
  git commit -m "feat(seed): idempotently seed default departments + positions tree"
  ```

---

### Task 15 — Service tests + extend `testhelper_test.go`

Tests are package `services_test` and use the real test DB. `truncateAll` must also wipe `departments, positions` (FK from `employees` → `departments`/`positions` means order/CASCADE matters).

- [x] In `internal/services/testhelper_test.go`, update `truncateAll` to include the new tables (departments/positions must be truncated together with employees; CASCADE handles the FK):

  ```go
  if err := testDB.Exec(`TRUNCATE TABLE device_tokens, user_notification_settings, employee_leave_quotas, dependents, employees, positions, departments, user_roles, users, roles RESTART IDENTITY CASCADE`).Error; err != nil {
      t.Fatalf("truncate: %v", err)
  }
  ```

- [x] Add helpers to `testhelper_test.go` (place after `makeEmployee`):

  ```go
  // makeEmployeeInDept inserts an employee assigned to the given department
  // (and optionally a position). Used to exercise the delete-conflict guards.
  func makeEmployeeInDept(t *testing.T, deptID uuid.UUID, posID *uuid.UUID) *models.Employee {
      t.Helper()
      u := makeUser(t, fmt.Sprintf("emp-%s@example.com", uuid.NewString()[:8]), "pw-Aa123456")
      e := &models.Employee{
          UserID:          u.ID,
          FullName:        "Dept Member",
          ContractType:    "official",
          ContractRenewal: 1,
          PaymentMethod:   "bank_transfer",
          DepartmentID:    &deptID,
          PositionID:      posID,
      }
      if err := testEmployeeRepo.Create(context.Background(), e); err != nil {
          t.Fatalf("create employee in dept: %v", err)
      }
      return e
  }
  ```

  (`fmt` and `uuid` are already imported in `testhelper_test.go`.)

- [x] Create `internal/services/department_service_test.go`. Cover at minimum these cases (package `services_test`, gate with `skipIfNoDB(t)`, `truncateAll(t)` at the top, build the service with real repos: `repositories.NewDepartmentRepository(testDB)` and `repositories.NewPositionRepository(testDB)`):

  - `TestDepartmentService_Create_OK`
  - `TestDepartmentService_Create_DuplicateName_Conflict`
  - `TestDepartmentService_Create_BlankName_BadRequest`
  - `TestDepartmentService_Create_InvalidParent_BadRequest`
  - `TestDepartmentService_Update_RenameAndReparent`
  - `TestDepartmentService_Update_CycleRejected`
  - `TestDepartmentService_Delete_BlockedByChildren_Conflict`
  - `TestDepartmentService_Delete_BlockedByPositions_Conflict`
  - `TestDepartmentService_Delete_BlockedByEmployees_Conflict`
  - `TestDepartmentService_Delete_SoftDeletesBothColumns`
  - `TestDepartmentService_List_SearchAndParentFilter`

  Reference skeleton (verbatim — others follow the same shape):

  ```go
  package services_test

  import (
      "context"
      "testing"

      "github.com/stretchr/testify/require"

      "github.com/exnodes/hrm-api/internal/dto"
      apperrors "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/repositories"
      "github.com/exnodes/hrm-api/internal/services"
  )

  func newDeptSvc(t *testing.T) (*services.DepartmentService, repositories.DepartmentRepository, repositories.PositionRepository) {
      t.Helper()
      dr := repositories.NewDepartmentRepository(testDB)
      pr := repositories.NewPositionRepository(testDB)
      return services.NewDepartmentService(dr, pr), dr, pr
  }

  func TestDepartmentService_Delete_BlockedByEmployees_Conflict(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      ctx := context.Background()

      svc, _, _ := newDeptSvc(t)
      dept, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Sales"})
      require.NoError(t, err)

      makeEmployeeInDept(t, dept.ID, nil)

      err = svc.Delete(ctx, dept.ID)
      require.Error(t, err)
      ae, ok := apperrors.As(err)
      require.True(t, ok)
      require.Equal(t, apperrors.CodeConflict, ae.Code)
      require.Contains(t, ae.Message, "Reassign all employees")
  }

  func TestDepartmentService_Delete_SoftDeletesBothColumns(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      ctx := context.Background()

      svc, _, _ := newDeptSvc(t)
      dept, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Ops"})
      require.NoError(t, err)
      require.NoError(t, svc.Delete(ctx, dept.ID))

      var isDeleted bool
      var hasDeletedAt bool
      row := testDB.Raw(
          "SELECT is_deleted, deleted_at IS NOT NULL FROM departments WHERE id = ?", dept.ID,
      ).Row()
      require.NoError(t, row.Scan(&isDeleted, &hasDeletedAt))
      require.True(t, isDeleted)
      require.True(t, hasDeletedAt)
  }
  ```

- [x] Create `internal/services/position_service_test.go` covering:

  - `TestPositionService_Create_OK`
  - `TestPositionService_Create_MissingDept_BadRequest`
  - `TestPositionService_Create_DuplicateNameInDept_Conflict`
  - `TestPositionService_Create_DuplicateNameInDifferentDept_OK`
  - `TestPositionService_Update_MoveToOtherDept`
  - `TestPositionService_Delete_BlockedByEmployees_Conflict`
  - `TestPositionService_Delete_SoftDeletesBothColumns`
  - `TestPositionService_List_SearchAndDeptFilter`

- [x] Run:

  ```bash
  TEST_DATABASE_URL='postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm_test?sslmode=disable' \
    go test ./internal/services/ -run 'Department|Position' -count=1 -v
  ```

  Expected tail:

  ```
  PASS
  ok  	github.com/exnodes/hrm-api/internal/services	X.XXXs
  ```

- [x] Commit:

  ```bash
  git add internal/services/department_service_test.go internal/services/position_service_test.go internal/services/testhelper_test.go
  git commit -m "test(services): cover departments + positions create/update/delete/list paths"
  ```

---

### Task 16 — Regenerate Swagger + full build/test

- [x] Regenerate Swagger:

  ```bash
  cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2
  swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal
  ```

  Expected last line includes `create docs/swagger/docs.go` (plus `swagger.json`, `swagger.yaml`).

- [x] Full build + test (DB-backed tests need the test DB URL; without it they self-skip but must still compile):

  ```bash
  go build ./...
  TEST_DATABASE_URL='postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm_test?sslmode=disable' \
    go test ./... -count=1
  ```

  Expected: all packages `ok` (or `[no test files]`), no `FAIL`.

- [x] Confirm no `AutoMigrate`:

  ```bash
  grep -R AutoMigrate internal cmd | wc -l
  ```

  Expected: `0`.

- [x] Boot + swagger sanity check:

  ```bash
  make migrate-up
  make run &
  SERVER_PID=$!
  until curl -sf http://localhost:8080/health >/dev/null; do sleep 1; done
  curl -s http://localhost:8080/swagger/doc.json | jq -r '.paths | keys[]' | grep -E 'departments|positions'
  ```

  Expected: `/health` reachable; grep prints the four paths:

  ```
  /api/v1/departments
  /api/v1/departments/{id}
  /api/v1/positions
  /api/v1/positions/{id}
  ```

  Leave the server running for Task 17 (or restart it there).

- [x] Commit:

  ```bash
  git add docs/swagger
  git commit -m "docs(swagger): regenerate API spec including departments + positions"
  ```

---

### Task 17 — End-to-end self-verification (real FK exercised)

This must prove the **deferred FK constraint works**: assign an employee to a created department via the existing employee admin endpoint, then attempt to delete that department → expect 409, plus an SQL spot-check.

- [ ] Ensure the server runs:

  ```bash
  curl -sf http://localhost:8080/health >/dev/null || { make run & SERVER_PID=$!; until curl -sf http://localhost:8080/health >/dev/null; do sleep 1; done; }
  mkdir -p /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification
  ```

- [ ] Walk these steps, capturing each command + HTTP code + abridged JSON (NO tokens/secrets) into `docs/superpowers/verification/phase-03.md`:

  1. **Login as super admin** (credentials from `.env`):

     ```bash
     TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
       -H 'Content-Type: application/json' \
       -d '{"email":"<SUPER_ADMIN_EMAIL>","password":"<SUPER_ADMIN_PASSWORD>"}' \
       | jq -r .data.access_token)
     test -n "$TOKEN" && echo "got token (${#TOKEN} chars)"
     ```

     Expected: `got token (NNN chars)`. Log only the length.

  2. **Create root department** → expect 201, save `ROOT_ID`.
  3. **Create child department** with `parent_id=$ROOT_ID` → expect 201, response `parent_id == $ROOT_ID`. Save `CHILD_ID`.
  4. **Create position** in the child → expect 201, response `department_id == $CHILD_ID`. Save `POS_ID`.
  5. **List departments** `?parent_id=root&search=Verify` → includes the root, excludes the child.
  6. **List positions** `?department_id=$CHILD_ID` → includes the created position.
  7. **PATCH child department** `{"description":"updated via e2e"}` → `data.description == "updated via e2e"`.
  8. **FK exercise — create an employee assigned to the child department** via the existing admin endpoint:

     ```bash
     EMP_ID=$(curl -s -X POST http://localhost:8080/api/v1/employees \
       -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
       -d "{\"email\":\"fkprobe@example.com\",\"password\":\"Pw-Aa123456\",\"full_name\":\"FK Probe\",\"department_id\":\"$CHILD_ID\",\"position_id\":\"$POS_ID\"}" \
       | jq -r .data.id)
     echo "EMP_ID=$EMP_ID"
     ```

     Expected: HTTP 201, a UUID (proves the FK accepts a valid `department_id`/`position_id`).

  9. **Attempt to delete the child department while an employee references it** → expect **409**:

     ```bash
     curl -s -o /tmp/del.json -w "%{http_code}\n" -X DELETE \
       "http://localhost:8080/api/v1/departments/$CHILD_ID" -H "Authorization: Bearer $TOKEN"
     jq . /tmp/del.json
     ```

     Expected: `409`, message contains `Reassign all employees`.

  10. **Attempt to delete the position while the employee references it** → expect **409** (`Reassign all employees`).
  11. **SQL spot-check the FK constraints + assignment**:

      ```bash
      psql "postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm?sslmode=disable" \
        -c "SELECT department_id IS NOT NULL AS has_dept, position_id IS NOT NULL AS has_pos FROM employees WHERE id = '$EMP_ID';" \
        -c "SELECT conname FROM pg_constraint WHERE conname IN ('fk_employees_department','fk_employees_position');"
      ```

      Expected: `has_dept = t`, `has_pos = t`; both constraint names present.

  12. **Clear the employee's dept/position** via the admin update endpoint, then re-delete:

      ```bash
      curl -s -X PATCH "http://localhost:8080/api/v1/employees/$EMP_ID" \
        -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
        -d '{"clear_dept":true,"clear_pos":true}' | jq '.data.id'
      ```

      Expected: HTTP 200.

  13. **Delete the position** → expect 200.
  14. **Delete the child department** (now empty) → expect 200.
  15. **DB spot-check soft-delete columns**:

      ```bash
      psql "postgres://ennam:ennam_dev_2026@localhost:5432/exnodes_hrm?sslmode=disable" \
        -c "SELECT is_deleted, deleted_at IS NOT NULL FROM departments WHERE id = '$CHILD_ID';"
      ```

      Expected: `is_deleted = t`, `deleted_at IS NOT NULL = t`.

  16. **Error path — invalid parent_id** on create → `400`, `Parent department not found`.
  17. **Error path — duplicate name** (re-create the root name) → `409`, `Department name already exists`.
  18. **Error path — missing permission**: log in as a user lacking `departments:create` (create one with the super admin token via `POST /api/v1/employees` with no roles, or reuse a seeded non-admin), POST a department → `403`.
  19. **Cleanup**: delete the root department and the FK-probe employee. Expect 200 each.

- [ ] Stop the server:

  ```bash
  kill "$SERVER_PID" 2>/dev/null || pkill -f 'exnodes-hrm-api' || true
  ```

- [ ] Write all sections (command + HTTP code + abridged JSON, no secrets) into `docs/superpowers/verification/phase-03.md`.

- [ ] Commit:

  ```bash
  git add docs/superpowers/verification/phase-03.md
  git commit -m "docs(verification): phase-03 end-to-end log (departments + positions + FK)"
  ```

---

### Task 18 — Update `README.md`

- [ ] In `README.md`, in the **Endpoints** section, after the Phase 2 `Employees`/`Users` block, add:

  ```md
  ### Departments

  | Method | Path                     | Permission          |
  |--------|--------------------------|---------------------|
  | GET    | /api/v1/departments      | departments:read    |
  | POST   | /api/v1/departments      | departments:create  |
  | GET    | /api/v1/departments/{id} | departments:read    |
  | PATCH  | /api/v1/departments/{id} | departments:update  |
  | DELETE | /api/v1/departments/{id} | departments:delete  |

  Self-referential `parent_id` (UUID or `"root"` filter on list). Delete is
  blocked (409) while child departments, active positions, or assigned
  employees exist.

  ### Positions

  | Method | Path                   | Permission        |
  |--------|------------------------|-------------------|
  | GET    | /api/v1/positions      | positions:read    |
  | POST   | /api/v1/positions      | positions:create  |
  | GET    | /api/v1/positions/{id} | positions:read    |
  | PATCH  | /api/v1/positions/{id} | positions:update  |
  | DELETE | /api/v1/positions/{id} | positions:delete  |

  Each position belongs to exactly one department. Delete is blocked (409)
  while employees are assigned. The `employees.department_id` /
  `employees.position_id` FK constraints (deferred from migration 000003) are
  added in migration 000005.
  ```

- [ ] Commit:

  ```bash
  git add README.md
  git commit -m "docs(readme): list phase 3 department + position endpoints"
  ```

---

## Definition of Done (checklist)

- [ ] Migration `000005` up + down applies cleanly forward and reverse (`make migrate-up`/`migrate-down`/`migrate-up`, version `5`, not dirty).
- [ ] `departments` + `positions` have UUID PK, 4 audit cols, `set_updated_at` trigger, `is_deleted` index.
- [ ] `employees.department_id` → `departments(id)` and `employees.position_id` → `positions(id)` FK constraints exist with `ON DELETE SET NULL` (verified via `\d+ employees`).
- [ ] Models embed `BaseModel`; soft-delete sets **both** `is_deleted = true` and `deleted_at = NOW()` (asserted in `*_SoftDeletesBothColumns` tests).
- [ ] All five department + all five position routes appear in Swagger with `@Security BearerAuth`.
- [ ] Every route uses `middleware.RequirePerms(authSvc, permissions.PermXxx)` with the EXISTING constant names.
- [ ] No new permission constants/groups/role grants were added (they already existed) — Task 2 verified only.
- [ ] Repos follow the interface + lowercase-impl convention; lists use `models.NotDeleted` + `utils.BuildILIKEPattern`.
- [ ] Delete guards return 409 for: child departments, active positions, assigned employees (dept); assigned employees (position) — proven by service tests AND the e2e log.
- [ ] Service tests pass: `go test ./internal/services/ -run 'Department|Position' -count=1 -v` (with `TEST_DATABASE_URL`).
- [ ] Full suite passes: `go test ./... -count=1` (with `TEST_DATABASE_URL`), no `FAIL`.
- [ ] `grep -R AutoMigrate internal cmd | wc -l` returns `0`.
- [ ] End-to-end log committed at `docs/superpowers/verification/phase-03.md` including the FK exercise (employee assigned via existing endpoint → dept delete returns 409) and the SQL spot-check.
- [ ] `README.md` updated.

## Out of scope (this phase)

- Bulk endpoints, CSV import/export.
- A dedicated `GET /departments/tree` endpoint — list with `parent_id` filtering suffices until the FE asks.
- Restore-from-soft-delete endpoints (global, later phase).
- Any changes to `users` (it is auth-only and has no department/position columns).
