# Phase 3 — Departments + Positions

| | |
|---|---|
| Status | Ready |
| Date | 2026-05-15 |
| Owner | danny.tranhoang@exnodes.vn |
| Spec | `docs/superpowers/specs/2026-05-15-go-migration-design.md` |
| Depends on | Phase 0 (foundation), Phase 1 (auth/RBAC), Phase 2 (users) |
| Target | `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/` |
| Module | `github.com/exnodes/hrm-api` |

---

## Goal

Port the Python FastAPI `departments` and `positions` modules to Go + Gin + GORM + Postgres, add the missing FK constraints from `users` to those two tables (added nullable in Phase 1), expose full CRUD with per-route permission middleware, and finish with an end-to-end self-verification log.

Source modules being ported (1:1 with field-level adjustments):

- `app/routers/departments.py` + `app/services/department.py` + `app/schemas/department.py` + `app/models/department.py`
- `app/routers/positions.py` + `app/services/position.py` + `app/schemas/position.py` + `app/models/position.py`

Go reference for layering: `/Users/sines/Documents/Work/exn-hr/Exn-hr/backend/internal/{services,repositories,handlers}/department_*.go`.

Behavioral deltas vs Python source:

| Change | Reason |
|---|---|
| `parent_id` on departments (self-FK, nullable) | Spec change — Python had flat departments; v2 supports a tree |
| `department_id` on positions (NOT NULL FK to departments) | Spec change — positions in v2 belong to a department |
| `description` on both | Spec change — keep parity with reference Go stack |
| UUID PKs, `is_deleted` soft-delete, audit cols, `set_updated_at` trigger | Project-wide convention |
| Search via `ILIKE` instead of `$regex` | Postgres-native |
| Delete blocked also when **children exist** (departments only) | New invariant from tree structure |
| FK from `users.department_id` / `users.position_id` set `ON DELETE SET NULL` | Defence in depth — service layer still rejects with 409 before that fires |

Non-negotiable constraints (from spec §5, §6, §7):

1. Versioned SQL only — no `AutoMigrate`.
2. Audit cols + `set_updated_at` trigger + `is_deleted` index on every new table.
3. Soft-delete sets **both** `is_deleted=true` and `deleted_at=NOW()`.
4. UUID PKs only.
5. Each route declares its perms inline via `middleware.RequirePerms(...)`.
6. Swagger annotations on every handler.
7. Search uses `ILIKE` with escaped `%`/`_` via `pkg/utils/search.go`.
8. Validate FK existence in service — return `BadRequest` if missing.
9. Reject delete if children/users exist — return `Conflict` (409).
10. Definition of Done requires end-to-end self-verification with the log committed to `docs/superpowers/verification/phase-03.md`.

---

## Tasks

1. Create migration `000004_create_departments_positions` (up + down) with both tables, triggers, indexes, and the two `ALTER TABLE users` FK constraints.
2. Add the four new permission constants to `internal/permissions/registry.go` (`PermDepartmentsRead/Create/Update/Delete`, `PermPositionsRead/Create/Update/Delete`) plus a Permission Group entry. Grant them to the appropriate system roles in the seeder.
3. Add `internal/models/department.go` (BaseModel embed + `ParentID *uuid.UUID` + `Parent *Department` + `Children []Department`).
4. Add `internal/models/position.go` (BaseModel embed + `DepartmentID uuid.UUID` + `Department *Department`).
5. Add `internal/dto/department.go` (Create / Update / Read / ListQuery).
6. Add `internal/dto/position.go` (Create / Update / Read / ListQuery).
7. Add `internal/repositories/department_repo.go` (interface + GORM impl).
8. Add `internal/repositories/position_repo.go` (interface + GORM impl).
9. Add `internal/services/department_service.go` with the same validations as Python (`_check_name_unique`, no-children, no-users-assigned-on-delete) plus parent existence validation.
10. Add `internal/services/position_service.go` (validate dept exists; no-users-assigned-on-delete).
11. Add `internal/handlers/department_handler.go` with full Swagger.
12. Add `internal/handlers/position_handler.go` with full Swagger.
13. Wire services + handlers in `cmd/server/main.go` (route registration with per-route `RequirePerms`).
14. Optionally extend `internal/services/seed_service.go` to seed a small default tree (parity with Python seeder if any), idempotent.
15. Add service tests under `internal/services/{department_service_test.go, position_service_test.go}`.
16. Run `make migrate-up`, `go test ./...`, `make run`, and `swag init`. Fix any failures inline.
17. End-to-end self-verification with curl against the running server. Capture every command + key response into `docs/superpowers/verification/phase-03.md`.
18. Update `README.md` "Endpoints" section with the new routes.

---

## Steps

### Task 1 — Migration `000004_create_departments_positions`

- [ ] Pre-check: confirm next migration index.

  ```bash
  ls /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations | sort
  ```

  Expected: the highest existing prefix is `000003`. If not, increment accordingly and update the filenames below in lockstep before creating files.

- [ ] Create `migrations/000004_create_departments_positions.up.sql`:

  ```sql
  -- Phase 3: departments + positions, plus FK constraints on users.

  CREATE TABLE departments (
      id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name        TEXT NOT NULL,
      description TEXT NOT NULL DEFAULT '',
      parent_id   UUID NULL REFERENCES departments(id) ON DELETE SET NULL,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
      deleted_at  TIMESTAMPTZ NULL
  );

  CREATE UNIQUE INDEX uq_departments_name_active
      ON departments (LOWER(name))
      WHERE is_deleted = FALSE;

  CREATE INDEX idx_departments_is_deleted ON departments (is_deleted);
  CREATE INDEX idx_departments_parent_id  ON departments (parent_id) WHERE parent_id IS NOT NULL;

  CREATE TRIGGER trg_departments_set_updated_at
      BEFORE UPDATE ON departments
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  CREATE TABLE positions (
      id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name          TEXT NOT NULL,
      description   TEXT NOT NULL DEFAULT '',
      department_id UUID NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
      created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted    BOOLEAN     NOT NULL DEFAULT FALSE,
      deleted_at    TIMESTAMPTZ NULL
  );

  CREATE UNIQUE INDEX uq_positions_name_dept_active
      ON positions (department_id, LOWER(name))
      WHERE is_deleted = FALSE;

  CREATE INDEX idx_positions_is_deleted    ON positions (is_deleted);
  CREATE INDEX idx_positions_department_id ON positions (department_id);

  CREATE TRIGGER trg_positions_set_updated_at
      BEFORE UPDATE ON positions
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  -- Backfill the FK constraints on users (added as nullable cols without FKs in Phase 1).
  ALTER TABLE users
      ADD CONSTRAINT fk_users_department
          FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

  ALTER TABLE users
      ADD CONSTRAINT fk_users_position
          FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL;
  ```

- [ ] Create `migrations/000004_create_departments_positions.down.sql`:

  ```sql
  ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_position;
  ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_department;

  DROP TRIGGER IF EXISTS trg_positions_set_updated_at ON positions;
  DROP TABLE IF EXISTS positions;

  DROP TRIGGER IF EXISTS trg_departments_set_updated_at ON departments;
  DROP TABLE IF EXISTS departments;
  ```

- [ ] Apply forward, roll back, and re-apply to prove both directions:

  ```bash
  cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2
  make migrate-up
  make migrate-version
  make migrate-down
  make migrate-version
  make migrate-up
  make migrate-version
  ```

  Expected last line: `4` (or whatever the new latest version is). No `dirty` flag.

- [ ] Commit:

  ```bash
  git add migrations/000004_create_departments_positions.up.sql migrations/000004_create_departments_positions.down.sql
  git commit -m "feat(migrations): add departments, positions, and user FK constraints (phase 3)"
  ```

---

### Task 2 — Permission constants + grants

- [ ] Open `internal/permissions/registry.go` and confirm the four permissions already declared in the spec exist (`PermDepartmentsRead/Create/Update/Delete`, `PermPositionsRead/Create/Update/Delete`). If any are missing, add them in the same block style:

  ```go
  const (
      // ... existing ...

      PermDepartmentsRead   Permission = "departments:read"
      PermDepartmentsCreate Permission = "departments:create"
      PermDepartmentsUpdate Permission = "departments:update"
      PermDepartmentsDelete Permission = "departments:delete"

      PermPositionsRead   Permission = "positions:read"
      PermPositionsCreate Permission = "positions:create"
      PermPositionsUpdate Permission = "positions:update"
      PermPositionsDelete Permission = "positions:delete"
  )
  ```

- [ ] Ensure both groups appear in `PermissionGroups` (the slice consumed by `GET /api/v1/roles/permissions`). Append if missing:

  ```go
  {
      Key:   "departments",
      Label: "Departments",
      Permissions: []Permission{
          PermDepartmentsRead, PermDepartmentsCreate, PermDepartmentsUpdate, PermDepartmentsDelete,
      },
  },
  {
      Key:   "positions",
      Label: "Positions",
      Permissions: []Permission{
          PermPositionsRead, PermPositionsCreate, PermPositionsUpdate, PermPositionsDelete,
      },
  },
  ```

- [ ] In `internal/services/seed_service.go`, ensure the system role grants include the new permissions:

  - **Super Admin** — already `["*"]`, nothing to change.
  - **Admin** — full `departments:*` and `positions:*`.
  - **HR Manager** — full `departments:*` and `positions:*`.
  - **Manager** — `departments:read`, `positions:read`.
  - **Employee** — `departments:read`, `positions:read`.

  Idempotent seed must update the permission JSON when run on an existing DB (the existing seeder pattern from Phase 1 already does so for system roles — extend the slice literals only).

- [ ] Build:

  ```bash
  go build ./...
  ```

  Expected: clean build, no output.

- [ ] Commit:

  ```bash
  git add internal/permissions/registry.go internal/services/seed_service.go
  git commit -m "feat(permissions): register departments + positions perms and seed role grants"
  ```

---

### Task 3 — `internal/models/department.go`

- [ ] Create `internal/models/department.go`:

  ```go
  package models

  import (
      "github.com/google/uuid"
  )

  // Department represents an org unit. Self-referential tree via ParentID.
  type Department struct {
      BaseModel
      Name        string     `gorm:"type:text;not null"               json:"name"`
      Description string     `gorm:"type:text;not null;default:''"    json:"description"`
      ParentID    *uuid.UUID `gorm:"type:uuid;index"                  json:"parent_id,omitempty"`

      // Relations (preloaded on demand, never serialized when nil).
      Parent   *Department  `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
      Children []Department `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
  }

  // TableName pins to the migration-created table.
  func (Department) TableName() string { return "departments" }
  ```

- [ ] Build:

  ```bash
  go build ./internal/models/...
  ```

  Expected: clean build.

- [ ] Commit:

  ```bash
  git add internal/models/department.go
  git commit -m "feat(models): add Department model with self-referential tree"
  ```

---

### Task 4 — `internal/models/position.go`

- [ ] Create `internal/models/position.go`:

  ```go
  package models

  import (
      "github.com/google/uuid"
  )

  // Position represents a job role; belongs to exactly one department.
  type Position struct {
      BaseModel
      Name         string    `gorm:"type:text;not null"            json:"name"`
      Description  string    `gorm:"type:text;not null;default:''" json:"description"`
      DepartmentID uuid.UUID `gorm:"type:uuid;not null;index"      json:"department_id"`

      Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
  }

  func (Position) TableName() string { return "positions" }
  ```

- [ ] Build:

  ```bash
  go build ./internal/models/...
  ```

- [ ] Commit:

  ```bash
  git add internal/models/position.go
  git commit -m "feat(models): add Position model belonging to a Department"
  ```

---

### Task 5 — `internal/dto/department.go`

- [ ] Create `internal/dto/department.go`:

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
  // All fields optional — service applies only those provided (pointer-based).
  type DepartmentUpdate struct {
      Name        *string     `json:"name,omitempty"        binding:"omitempty,min=1,max=100"`
      Description *string     `json:"description,omitempty" binding:"omitempty,max=1000"`
      ParentID    *uuid.UUID  `json:"parent_id,omitempty"`
      // ClearParent sets ParentID to NULL when true. Allows distinguishing "no change" vs "make root".
      ClearParent bool        `json:"clear_parent,omitempty"`
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
  type DepartmentListQuery struct {
      Page     int    `form:"page,default=1"       binding:"min=1"`
      PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
      Search   string `form:"search"`
      // ParentID filters by parent. Special value "root" returns top-level departments only.
      ParentID string `form:"parent_id"`
  }
  ```

- [ ] Build:

  ```bash
  go build ./internal/dto/...
  ```

- [ ] Commit:

  ```bash
  git add internal/dto/department.go
  git commit -m "feat(dto): add department request/response shapes"
  ```

---

### Task 6 — `internal/dto/position.go`

- [ ] Create `internal/dto/position.go`:

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

- [ ] Build:

  ```bash
  go build ./internal/dto/...
  ```

- [ ] Commit:

  ```bash
  git add internal/dto/position.go
  git commit -m "feat(dto): add position request/response shapes"
  ```

---

### Task 7 — `internal/repositories/department_repo.go`

- [ ] Create `internal/repositories/department_repo.go`:

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
  type DepartmentFilter struct {
      Page     int
      PageSize int
      Search   string
      // ParentID semantics:
      //   nil           → no filter
      //   uuid.Nil ptr  → top-level only (parent_id IS NULL)
      //   real uuid ptr → children of that parent
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
      HasUsers(ctx context.Context, id uuid.UUID) (int64, error)
  }

  type departmentRepo struct {
      db *gorm.DB
  }

  func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
      return &departmentRepo{db: db}
  }

  func (r *departmentRepo) notDeleted(ctx context.Context) *gorm.DB {
      return r.db.WithContext(ctx).Scopes(models.NotDeleted)
  }

  func (r *departmentRepo) Create(ctx context.Context, d *models.Department) error {
      return r.db.WithContext(ctx).Create(d).Error
  }

  func (r *departmentRepo) Update(ctx context.Context, d *models.Department) error {
      return r.db.WithContext(ctx).Save(d).Error
  }

  func (r *departmentRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
      return r.db.WithContext(ctx).
          Model(&models.Department{}).
          Where("id = ? AND is_deleted = FALSE", id).
          Updates(map[string]any{
              "is_deleted": true,
              "deleted_at": gorm.Expr("NOW()"),
          }).Error
  }

  func (r *departmentRepo) FindByID(ctx context.Context, id uuid.UUID, preloadParent bool) (*models.Department, error) {
      q := r.notDeleted(ctx)
      if preloadParent {
          q = q.Preload("Parent", models.NotDeleted)
      }
      var d models.Department
      if err := q.Where("id = ?", id).First(&d).Error; err != nil {
          return nil, err
      }
      return &d, nil
  }

  func (r *departmentRepo) FindByName(ctx context.Context, name string) (*models.Department, error) {
      var d models.Department
      err := r.notDeleted(ctx).
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

  func (r *departmentRepo) List(ctx context.Context, f DepartmentFilter) ([]models.Department, int64, error) {
      base := r.notDeleted(ctx).Model(&models.Department{})

      if s := strings.TrimSpace(f.Search); s != "" {
          base = base.Where("name ILIKE ?", "%"+utils.EscapeLike(s)+"%")
      }
      if f.ParentID != nil {
          if *f.ParentID == uuid.Nil {
              base = base.Where("parent_id IS NULL")
          } else {
              base = base.Where("parent_id = ?", *f.ParentID)
          }
      }

      var total int64
      if err := base.Count(&total).Error; err != nil {
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
      err := base.
          Preload("Parent", models.NotDeleted).
          Order("LOWER(name) ASC").
          Offset((page - 1) * size).
          Limit(size).
          Find(&items).Error
      return items, total, err
  }

  func (r *departmentRepo) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
      var count int64
      err := r.notDeleted(ctx).
          Model(&models.Department{}).
          Where("parent_id = ?", id).
          Count(&count).Error
      return count > 0, err
  }

  func (r *departmentRepo) HasUsers(ctx context.Context, id uuid.UUID) (int64, error) {
      var count int64
      err := r.db.WithContext(ctx).
          Table("users").
          Where("department_id = ? AND is_deleted = FALSE", id).
          Count(&count).Error
      return count, err
  }
  ```

- [ ] If `pkg/utils/search.go` does not already export `EscapeLike(string) string`, add it now:

  ```go
  package utils

  import "strings"

  // EscapeLike escapes %, _, and \ so a user-supplied search fragment is safe
  // to interpolate into an ILIKE pattern (the caller still adds the wrapping %).
  func EscapeLike(s string) string {
      r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
      return r.Replace(s)
  }
  ```

- [ ] Build:

  ```bash
  go build ./internal/repositories/... ./pkg/utils/...
  ```

- [ ] Commit:

  ```bash
  git add internal/repositories/department_repo.go pkg/utils/search.go
  git commit -m "feat(repositories): add department repo with tree, soft-delete, ILIKE search"
  ```

---

### Task 8 — `internal/repositories/position_repo.go`

- [ ] Create `internal/repositories/position_repo.go`:

  ```go
  package repositories

  import (
      "context"
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
      HasUsers(ctx context.Context, id uuid.UUID) (int64, error)
  }

  type positionRepo struct {
      db *gorm.DB
  }

  func NewPositionRepository(db *gorm.DB) PositionRepository {
      return &positionRepo{db: db}
  }

  func (r *positionRepo) notDeleted(ctx context.Context) *gorm.DB {
      return r.db.WithContext(ctx).Scopes(models.NotDeleted)
  }

  func (r *positionRepo) Create(ctx context.Context, p *models.Position) error {
      return r.db.WithContext(ctx).Create(p).Error
  }

  func (r *positionRepo) Update(ctx context.Context, p *models.Position) error {
      return r.db.WithContext(ctx).Save(p).Error
  }

  func (r *positionRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
      return r.db.WithContext(ctx).
          Model(&models.Position{}).
          Where("id = ? AND is_deleted = FALSE", id).
          Updates(map[string]any{
              "is_deleted": true,
              "deleted_at": gorm.Expr("NOW()"),
          }).Error
  }

  func (r *positionRepo) FindByID(ctx context.Context, id uuid.UUID, preloadDept bool) (*models.Position, error) {
      q := r.notDeleted(ctx)
      if preloadDept {
          q = q.Preload("Department", models.NotDeleted)
      }
      var p models.Position
      if err := q.Where("id = ?", id).First(&p).Error; err != nil {
          return nil, err
      }
      return &p, nil
  }

  func (r *positionRepo) FindByNameInDept(ctx context.Context, name string, departmentID uuid.UUID) (*models.Position, error) {
      var p models.Position
      err := r.notDeleted(ctx).
          Where("LOWER(name) = LOWER(?) AND department_id = ?", strings.TrimSpace(name), departmentID).
          First(&p).Error
      if err != nil {
          if err == gorm.ErrRecordNotFound {
              return nil, nil
          }
          return nil, err
      }
      return &p, nil
  }

  func (r *positionRepo) List(ctx context.Context, f PositionFilter) ([]models.Position, int64, error) {
      base := r.notDeleted(ctx).Model(&models.Position{})

      if s := strings.TrimSpace(f.Search); s != "" {
          base = base.Where("name ILIKE ?", "%"+utils.EscapeLike(s)+"%")
      }
      if f.DepartmentID != nil {
          base = base.Where("department_id = ?", *f.DepartmentID)
      }

      var total int64
      if err := base.Count(&total).Error; err != nil {
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
      err := base.
          Preload("Department", models.NotDeleted).
          Order("LOWER(name) ASC").
          Offset((page - 1) * size).
          Limit(size).
          Find(&items).Error
      return items, total, err
  }

  func (r *positionRepo) CountByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error) {
      var count int64
      err := r.notDeleted(ctx).
          Model(&models.Position{}).
          Where("department_id = ?", departmentID).
          Count(&count).Error
      return count, err
  }

  func (r *positionRepo) HasUsers(ctx context.Context, id uuid.UUID) (int64, error) {
      var count int64
      err := r.db.WithContext(ctx).
          Table("users").
          Where("position_id = ? AND is_deleted = FALSE", id).
          Count(&count).Error
      return count, err
  }
  ```

- [ ] Build:

  ```bash
  go build ./internal/repositories/...
  ```

- [ ] Commit:

  ```bash
  git add internal/repositories/position_repo.go
  git commit -m "feat(repositories): add position repo with department filter and user-count check"
  ```

---

### Task 9 — `internal/services/department_service.go`

- [ ] Create `internal/services/department_service.go`:

  ```go
  package services

  import (
      "context"
      "errors"
      "fmt"
      "strings"

      "github.com/google/uuid"
      "gorm.io/gorm"

      apperr "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/dto"
      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/repositories"
  )

  type DepartmentService struct {
      repo repositories.DepartmentRepository
  }

  func NewDepartmentService(repo repositories.DepartmentRepository) *DepartmentService {
      return &DepartmentService{repo: repo}
  }

  // toRead converts a model into the wire shape. Always uses the preloaded parent when present.
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
      return apperr.ErrConflict("Department name already exists")
  }

  // assertParent verifies the proposed parent exists and would not create a cycle.
  func (s *DepartmentService) assertParent(ctx context.Context, parentID uuid.UUID, selfID *uuid.UUID) error {
      if selfID != nil && parentID == *selfID {
          return apperr.ErrBadRequest("Department cannot be its own parent")
      }
      parent, err := s.repo.FindByID(ctx, parentID, false)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperr.ErrBadRequest("Parent department not found")
          }
          return err
      }
      // Walk up the chain to detect cycles when updating an existing node.
      if selfID != nil {
          current := parent
          for current.ParentID != nil {
              if *current.ParentID == *selfID {
                  return apperr.ErrBadRequest("Setting this parent would create a cycle")
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
          return nil, apperr.ErrBadRequest("Department name cannot be blank")
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

      // Re-fetch with Parent preloaded for the response.
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
              return nil, apperr.ErrNotFound("Department")
          }
          return nil, err
      }

      if in.Name != nil {
          name := strings.TrimSpace(*in.Name)
          if name == "" {
              return nil, apperr.ErrBadRequest("Department name cannot be blank")
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

  func (s *DepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
      if _, err := s.repo.FindByID(ctx, id, false); err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperr.ErrNotFound("Department")
          }
          return err
      }

      hasChildren, err := s.repo.HasChildren(ctx, id)
      if err != nil {
          return err
      }
      if hasChildren {
          return apperr.ErrConflict("Cannot delete department — it has child departments. Move or delete them first.")
      }

      userCount, err := s.repo.HasUsers(ctx, id)
      if err != nil {
          return err
      }
      if userCount > 0 {
          plural := "employee is"
          if userCount > 1 {
              plural = "employees are"
          }
          return apperr.ErrConflict(fmt.Sprintf(
              "Cannot delete — %d %s assigned to this department. Reassign all employees before deleting.",
              userCount, plural,
          ))
      }
      // Positions in this dept also block delete (RESTRICT FK), surface a clean message.
      // We keep this check in the position service via CountByDepartment indirectly; here we
      // mirror the Python contract: dept must be empty of both users and positions.
      // The handler/service composition wires both checks via the team-level orchestrator.
      return s.repo.SoftDelete(ctx, id)
  }

  func (s *DepartmentService) Get(ctx context.Context, id uuid.UUID) (*dto.DepartmentRead, error) {
      d, err := s.repo.FindByID(ctx, id, true)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperr.ErrNotFound("Department")
          }
          return nil, err
      }
      out := departmentToRead(d)
      return &out, nil
  }

  func (s *DepartmentService) List(ctx context.Context, q dto.DepartmentListQuery) (*dto.PaginatedData[dto.DepartmentRead], error) {
      f := repositories.DepartmentFilter{
          Page:     q.Page,
          PageSize: q.PageSize,
          Search:   q.Search,
      }
      switch strings.ToLower(strings.TrimSpace(q.ParentID)) {
      case "":
          // no filter
      case "root", "null":
          nilUUID := uuid.Nil
          f.ParentID = &nilUUID
      default:
          parsed, err := uuid.Parse(q.ParentID)
          if err != nil {
              return nil, apperr.ErrBadRequest("Invalid parent_id")
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

  // DeleteWithPositionCheck composes the position repo check so the handler can call
  // a single method. We expose this here to avoid leaking the cross-repo check into
  // the handler layer.
  func (s *DepartmentService) DeleteWithPositionCheck(ctx context.Context, id uuid.UUID, posRepo repositories.PositionRepository) error {
      if _, err := s.repo.FindByID(ctx, id, false); err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return apperr.ErrNotFound("Department")
          }
          return err
      }
      posCount, err := posRepo.CountByDepartment(ctx, id)
      if err != nil {
          return err
      }
      if posCount > 0 {
          plural := "position is"
          if posCount > 1 {
              plural = "positions are"
          }
          return apperr.ErrConflict(fmt.Sprintf(
              "Cannot delete — %d %s assigned to this department. Delete or reassign them first.",
              posCount, plural,
          ))
      }
      return s.Delete(ctx, id)
  }
  ```

- [ ] Build:

  ```bash
  go build ./internal/services/...
  ```

- [ ] Commit:

  ```bash
  git add internal/services/department_service.go
  git commit -m "feat(services): add department service with unique-name, parent, and cascade-delete guards"
  ```

---

### Task 10 — `internal/services/position_service.go`

- [ ] Create `internal/services/position_service.go`:

  ```go
  package services

  import (
      "context"
      "errors"
      "fmt"
      "strings"

      "github.com/google/uuid"
      "gorm.io/gorm"

      apperr "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/dto"
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
              return apperr.ErrBadRequest("Department not found")
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
      return apperr.ErrConflict("Position name already exists in this department")
  }

  func (s *PositionService) Create(ctx context.Context, in dto.PositionCreate) (*dto.PositionRead, error) {
      name := strings.TrimSpace(in.Name)
      if name == "" {
          return nil, apperr.ErrBadRequest("Position name cannot be blank")
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
              return nil, apperr.ErrNotFound("Position")
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
              return nil, apperr.ErrBadRequest("Position name cannot be blank")
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
              return apperr.ErrNotFound("Position")
          }
          return err
      }
      userCount, err := s.repo.HasUsers(ctx, id)
      if err != nil {
          return err
      }
      if userCount > 0 {
          plural := "employee is"
          if userCount > 1 {
              plural = "employees are"
          }
          return apperr.ErrConflict(fmt.Sprintf(
              "Cannot delete — %d %s assigned to this position. Reassign all employees before deleting.",
              userCount, plural,
          ))
      }
      return s.repo.SoftDelete(ctx, id)
  }

  func (s *PositionService) Get(ctx context.Context, id uuid.UUID) (*dto.PositionRead, error) {
      p, err := s.repo.FindByID(ctx, id, true)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperr.ErrNotFound("Position")
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

- [ ] Build:

  ```bash
  go build ./internal/services/...
  ```

- [ ] Commit:

  ```bash
  git add internal/services/position_service.go
  git commit -m "feat(services): add position service with dept-FK validation and user-count delete guard"
  ```

---

### Task 11 — `internal/handlers/department_handler.go`

- [ ] Create `internal/handlers/department_handler.go`:

  ```go
  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/google/uuid"

      "github.com/exnodes/hrm-api/internal/dto"
      "github.com/exnodes/hrm-api/internal/repositories"
      "github.com/exnodes/hrm-api/internal/services"
  )

  type DepartmentHandler struct {
      svc     *services.DepartmentService
      posRepo repositories.PositionRepository // used for the cross-aggregate delete guard
  }

  func NewDepartmentHandler(svc *services.DepartmentService, posRepo repositories.PositionRepository) *DepartmentHandler {
      return &DepartmentHandler{svc: svc, posRepo: posRepo}
  }

  // List godoc
  // @Summary      List departments
  // @Description  Paginated list with optional name search and parent filter ("root" returns top-level only).
  // @Tags         Departments
  // @Security     BearerAuth
  // @Param        page       query    int     false  "Page number"  default(1)
  // @Param        page_size  query    int     false  "Page size"    default(10)
  // @Param        search     query    string  false  "Substring match on name (ILIKE)"
  // @Param        parent_id  query    string  false  "Filter by parent UUID, or \"root\" for top-level"
  // @Success      200  {object}  dto.Response[dto.PaginatedData[dto.DepartmentRead]]
  // @Failure      400  {object}  dto.Response[any]
  // @Failure      401  {object}  dto.Response[any]
  // @Failure      403  {object}  dto.Response[any]
  // @Router       /api/v1/departments [get]
  func (h *DepartmentHandler) List(c *gin.Context) {
      var q dto.DepartmentListQuery
      if err := c.ShouldBindQuery(&q); err != nil {
          _ = c.Error(err)
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
  // @Tags         Departments
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        body  body      dto.DepartmentCreate  true  "Department payload"
  // @Success      201   {object}  dto.Response[dto.DepartmentRead]
  // @Failure      400   {object}  dto.Response[any]
  // @Failure      401   {object}  dto.Response[any]
  // @Failure      403   {object}  dto.Response[any]
  // @Failure      409   {object}  dto.Response[any]
  // @Router       /api/v1/departments [post]
  func (h *DepartmentHandler) Create(c *gin.Context) {
      var in dto.DepartmentCreate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(err)
          return
      }
      out, err := h.svc.Create(c.Request.Context(), in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusCreated, dto.Response[*dto.DepartmentRead]{
          Success: true,
          Message: "Department created successfully",
          Data:    out,
      })
  }

  // Get godoc
  // @Summary      Get department by ID
  // @Tags         Departments
  // @Security     BearerAuth
  // @Param        id   path      string  true  "Department UUID"
  // @Success      200  {object}  dto.Response[dto.DepartmentRead]
  // @Failure      400  {object}  dto.Response[any]
  // @Failure      404  {object}  dto.Response[any]
  // @Router       /api/v1/departments/{id} [get]
  func (h *DepartmentHandler) Get(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(err)
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
  // @Tags         Departments
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        id    path      string                true  "Department UUID"
  // @Param        body  body      dto.DepartmentUpdate  true  "Fields to update (PATCH semantics)"
  // @Success      200   {object}  dto.Response[dto.DepartmentRead]
  // @Failure      400   {object}  dto.Response[any]
  // @Failure      404   {object}  dto.Response[any]
  // @Failure      409   {object}  dto.Response[any]
  // @Router       /api/v1/departments/{id} [patch]
  func (h *DepartmentHandler) Update(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(err)
          return
      }
      var in dto.DepartmentUpdate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(err)
          return
      }
      out, err := h.svc.Update(c.Request.Context(), id, in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.DepartmentRead]{
          Success: true,
          Message: "Department updated successfully",
          Data:    out,
      })
  }

  // Delete godoc
  // @Summary      Delete department
  // @Description  Soft-deletes a department. Rejected with 409 if it has child departments, active positions, or assigned users.
  // @Tags         Departments
  // @Security     BearerAuth
  // @Param        id   path      string  true  "Department UUID"
  // @Success      200  {object}  dto.Response[any]
  // @Failure      404  {object}  dto.Response[any]
  // @Failure      409  {object}  dto.Response[any]
  // @Router       /api/v1/departments/{id} [delete]
  func (h *DepartmentHandler) Delete(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(err)
          return
      }
      if err := h.svc.DeleteWithPositionCheck(c.Request.Context(), id, h.posRepo); err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Department deleted successfully"})
  }
  ```

- [ ] Build:

  ```bash
  go build ./internal/handlers/...
  ```

- [ ] Commit:

  ```bash
  git add internal/handlers/department_handler.go
  git commit -m "feat(handlers): add department CRUD handler with Swagger annotations"
  ```

---

### Task 12 — `internal/handlers/position_handler.go`

- [ ] Create `internal/handlers/position_handler.go`:

  ```go
  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/google/uuid"

      "github.com/exnodes/hrm-api/internal/dto"
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
  // @Tags         Positions
  // @Security     BearerAuth
  // @Param        page           query    int     false  "Page"        default(1)
  // @Param        page_size      query    int     false  "Page size"   default(10)
  // @Param        search         query    string  false  "Substring match on name"
  // @Param        department_id  query    string  false  "Filter by department UUID"
  // @Success      200  {object}  dto.Response[dto.PaginatedData[dto.PositionRead]]
  // @Failure      400  {object}  dto.Response[any]
  // @Router       /api/v1/positions [get]
  func (h *PositionHandler) List(c *gin.Context) {
      var q dto.PositionListQuery
      if err := c.ShouldBindQuery(&q); err != nil {
          _ = c.Error(err)
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
  // @Tags         Positions
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        body  body      dto.PositionCreate  true  "Position payload"
  // @Success      201   {object}  dto.Response[dto.PositionRead]
  // @Failure      400   {object}  dto.Response[any]
  // @Failure      409   {object}  dto.Response[any]
  // @Router       /api/v1/positions [post]
  func (h *PositionHandler) Create(c *gin.Context) {
      var in dto.PositionCreate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(err)
          return
      }
      out, err := h.svc.Create(c.Request.Context(), in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusCreated, dto.Response[*dto.PositionRead]{
          Success: true,
          Message: "Position created successfully",
          Data:    out,
      })
  }

  // Get godoc
  // @Summary      Get position by ID
  // @Tags         Positions
  // @Security     BearerAuth
  // @Param        id   path      string  true  "Position UUID"
  // @Success      200  {object}  dto.Response[dto.PositionRead]
  // @Failure      400  {object}  dto.Response[any]
  // @Failure      404  {object}  dto.Response[any]
  // @Router       /api/v1/positions/{id} [get]
  func (h *PositionHandler) Get(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(err)
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
  // @Tags         Positions
  // @Security     BearerAuth
  // @Accept       json
  // @Produce      json
  // @Param        id    path      string              true  "Position UUID"
  // @Param        body  body      dto.PositionUpdate  true  "Fields to update"
  // @Success      200   {object}  dto.Response[dto.PositionRead]
  // @Failure      400   {object}  dto.Response[any]
  // @Failure      404   {object}  dto.Response[any]
  // @Failure      409   {object}  dto.Response[any]
  // @Router       /api/v1/positions/{id} [patch]
  func (h *PositionHandler) Update(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(err)
          return
      }
      var in dto.PositionUpdate
      if err := c.ShouldBindJSON(&in); err != nil {
          _ = c.Error(err)
          return
      }
      out, err := h.svc.Update(c.Request.Context(), id, in)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[*dto.PositionRead]{
          Success: true,
          Message: "Position updated successfully",
          Data:    out,
      })
  }

  // Delete godoc
  // @Summary      Delete position
  // @Description  Soft-deletes a position. Rejected with 409 if any user is still assigned.
  // @Tags         Positions
  // @Security     BearerAuth
  // @Param        id   path      string  true  "Position UUID"
  // @Success      200  {object}  dto.Response[any]
  // @Failure      404  {object}  dto.Response[any]
  // @Failure      409  {object}  dto.Response[any]
  // @Router       /api/v1/positions/{id} [delete]
  func (h *PositionHandler) Delete(c *gin.Context) {
      id, err := uuid.Parse(c.Param("id"))
      if err != nil {
          _ = c.Error(err)
          return
      }
      if err := h.svc.Delete(c.Request.Context(), id); err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Position deleted successfully"})
  }
  ```

- [ ] Build:

  ```bash
  go build ./internal/handlers/...
  ```

- [ ] Commit:

  ```bash
  git add internal/handlers/position_handler.go
  git commit -m "feat(handlers): add position CRUD handler with Swagger annotations"
  ```

---

### Task 13 — Wire routes in `cmd/server/main.go`

- [ ] Read `cmd/server/main.go` and add the wiring inside the authenticated `v1` group, alongside the existing users/roles wiring. Use the same per-route `RequirePerms` pattern from Phase 1/2.

  Required additions (insert near the other handler instantiations):

  ```go
  // Phase 3: departments + positions
  deptRepo := repositories.NewDepartmentRepository(db)
  posRepo  := repositories.NewPositionRepository(db)

  deptSvc := services.NewDepartmentService(deptRepo)
  posSvc  := services.NewPositionService(posRepo, deptRepo)

  deptH := handlers.NewDepartmentHandler(deptSvc, posRepo)
  posH  := handlers.NewPositionHandler(posSvc)
  ```

  And the route block (mirror the users block already present):

  ```go
  perm := permissions.Permission
  _ = perm // package alias

  departments := authed.Group("/departments")
  departments.GET("",        middleware.RequirePerms(permissions.PermDepartmentsRead),   deptH.List)
  departments.POST("",       middleware.RequirePerms(permissions.PermDepartmentsCreate), deptH.Create)
  departments.GET("/:id",    middleware.RequirePerms(permissions.PermDepartmentsRead),   deptH.Get)
  departments.PATCH("/:id",  middleware.RequirePerms(permissions.PermDepartmentsUpdate), deptH.Update)
  departments.DELETE("/:id", middleware.RequirePerms(permissions.PermDepartmentsDelete), deptH.Delete)

  positions := authed.Group("/positions")
  positions.GET("",        middleware.RequirePerms(permissions.PermPositionsRead),   posH.List)
  positions.POST("",       middleware.RequirePerms(permissions.PermPositionsCreate), posH.Create)
  positions.GET("/:id",    middleware.RequirePerms(permissions.PermPositionsRead),   posH.Get)
  positions.PATCH("/:id",  middleware.RequirePerms(permissions.PermPositionsUpdate), posH.Update)
  positions.DELETE("/:id", middleware.RequirePerms(permissions.PermPositionsDelete), posH.Delete)
  ```

  Drop the `perm := permissions.Permission; _ = perm` line if not needed in the existing style — kept here only as a reminder that the symbol path is `permissions.PermXxx`. Match the existing file's import aliases exactly.

- [ ] Build the whole module:

  ```bash
  go build ./...
  ```

  Expected: clean build, no output.

- [ ] Commit:

  ```bash
  git add cmd/server/main.go
  git commit -m "feat(server): wire department + position routes with per-route RequirePerms"
  ```

---

### Task 14 — Extend seeder (idempotent default tree)

- [ ] In `internal/services/seed_service.go`, after the system-roles seed and super-admin seed, add a `seedOrgDefaults(ctx)` call. The function inserts a small idempotent default tree only when the `departments` and `positions` tables are empty (use `COUNT(*) WHERE is_deleted = FALSE`):

  Defaults:
  - Departments: `Engineering` (root), `Human Resources` (root), `Mobile` (child of Engineering), `Backend` (child of Engineering).
  - Positions: `Software Engineer` in `Backend`, `Mobile Engineer` in `Mobile`, `HR Specialist` in `Human Resources`.

  Use `repositories.DepartmentRepository` and `repositories.PositionRepository` directly. Pseudocode:

  ```go
  func (s *SeedService) seedOrgDefaults(ctx context.Context) error {
      var deptCount int64
      if err := s.db.WithContext(ctx).Table("departments").Where("is_deleted = FALSE").Count(&deptCount).Error; err != nil {
          return err
      }
      if deptCount > 0 {
          return nil // idempotent: nothing to do
      }
      // ... create the four departments + three positions via the repos ...
      return nil
  }
  ```

  Wire `seedOrgDefaults` into the existing `Run`/`Seed` entrypoint in the same file.

- [ ] Run the server once to seed:

  ```bash
  make migrate-up
  make run &      # or run in another terminal
  sleep 3
  curl -s http://localhost:8080/health
  pkill -f "exnodes-hrm-api" || true
  ```

  Expected health response: `{"status":"ok",...}`. Then verify directly:

  ```bash
  psql "$DATABASE_URL" -c "SELECT name, parent_id FROM departments WHERE is_deleted = FALSE ORDER BY name;"
  psql "$DATABASE_URL" -c "SELECT name, department_id FROM positions WHERE is_deleted = FALSE ORDER BY name;"
  ```

  Expected: four department rows, three position rows.

- [ ] Commit:

  ```bash
  git add internal/services/seed_service.go
  git commit -m "feat(seed): idempotently seed default departments + positions tree"
  ```

---

### Task 15 — Service tests

- [ ] Create `internal/services/department_service_test.go` covering at minimum:

  - `TestDepartmentService_Create_OK`
  - `TestDepartmentService_Create_DuplicateName_Conflict`
  - `TestDepartmentService_Create_BlankName_BadRequest`
  - `TestDepartmentService_Create_InvalidParent_BadRequest`
  - `TestDepartmentService_Update_RenameAndReparent`
  - `TestDepartmentService_Update_CycleRejected`
  - `TestDepartmentService_Delete_BlockedByChildren_Conflict`
  - `TestDepartmentService_Delete_BlockedByUsers_Conflict`
  - `TestDepartmentService_Delete_BlockedByPositions_Conflict`
  - `TestDepartmentService_Delete_SoftDeletesBothColumns`
  - `TestDepartmentService_List_SearchAndParentFilter`

  Use the existing `testhelper_test.go` (real Postgres test DB + table cleanup) — do not introduce mocks. Helpers `makeDepartment(t, name, parent)`, `makePosition(t, name, deptID)`, `makeUserInDept(t, deptID)` may need to be added to `testhelper_test.go` if not already present.

  Skeleton for one test (verbatim, others follow same pattern):

  ```go
  func TestDepartmentService_Delete_BlockedByUsers_Conflict(t *testing.T) {
      ctx := context.Background()
      tx := newTestTx(t)
      defer tx.Rollback()

      deptRepo := repositories.NewDepartmentRepository(tx)
      posRepo  := repositories.NewPositionRepository(tx)
      svc      := services.NewDepartmentService(deptRepo)

      dept, err := svc.Create(ctx, dto.DepartmentCreate{Name: "Sales"})
      require.NoError(t, err)
      makeUserInDept(t, tx, dept.ID)

      err = svc.DeleteWithPositionCheck(ctx, dept.ID, posRepo)
      require.Error(t, err)

      var appErr *apperr.AppError
      require.ErrorAs(t, err, &appErr)
      require.Equal(t, "conflict", appErr.Code)
      require.Contains(t, appErr.Message, "Reassign all employees")
  }
  ```

- [ ] Create `internal/services/position_service_test.go` covering:

  - `TestPositionService_Create_OK`
  - `TestPositionService_Create_MissingDept_BadRequest`
  - `TestPositionService_Create_DuplicateNameInDept_Conflict`
  - `TestPositionService_Create_DuplicateNameInDifferentDept_OK`
  - `TestPositionService_Update_MoveToOtherDept`
  - `TestPositionService_Delete_BlockedByUsers_Conflict`
  - `TestPositionService_Delete_SoftDeletesBothColumns`
  - `TestPositionService_List_SearchAndDeptFilter`

- [ ] Run:

  ```bash
  go test ./internal/services/... -run 'Department|Position' -count=1 -v
  ```

  Expected tail:

  ```
  PASS
  ok  	github.com/exnodes/hrm-api/internal/services	X.XXXs
  ```

- [ ] Commit:

  ```bash
  git add internal/services/department_service_test.go internal/services/position_service_test.go internal/services/testhelper_test.go
  git commit -m "test(services): cover departments + positions create/update/delete/list paths"
  ```

---

### Task 16 — Regenerate Swagger and run full build + test

- [ ] Regenerate Swagger:

  ```bash
  cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2
  swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal
  ```

  Expected last line: `create docs/swagger/docs.go` (and `swagger.json`, `swagger.yaml`).

- [ ] Full test pass:

  ```bash
  go test ./... -count=1
  ```

  Expected: all packages `ok`, no `FAIL`.

- [ ] Boot the server:

  ```bash
  make migrate-up
  make run &
  SERVER_PID=$!
  sleep 3
  curl -s http://localhost:8080/health | jq .
  curl -s http://localhost:8080/swagger/doc.json | jq '.paths | keys[]' | grep -E 'departments|positions'
  ```

  Expected: `/health` returns `{"status":"ok",...}`. The grep prints 10 lines:

  ```
  "/api/v1/departments"
  "/api/v1/departments/{id}"
  "/api/v1/positions"
  "/api/v1/positions/{id}"
  ```

  (4 paths in 10 grep lines because some paths host multiple methods — actual count is "all four paths visible".)

- [ ] Commit:

  ```bash
  git add docs/swagger
  git commit -m "docs(swagger): regenerate API spec including departments + positions"
  ```

---

### Task 17 — End-to-end self-verification

- [ ] Ensure the server is still running (from Task 16). If not:

  ```bash
  make run &
  SERVER_PID=$!
  sleep 3
  ```

- [ ] Open a new file `docs/superpowers/verification/phase-03.md` (create the directory if it does not exist):

  ```bash
  mkdir -p /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification
  ```

- [ ] Walk the happy path and the negative paths with curl. Capture each command **and** the relevant fields of the response into `phase-03.md` under sub-headings. The required walk:

  1. **Login as super admin** (credentials come from `.env`):

     ```bash
     TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
       -H 'Content-Type: application/json' \
       -d '{"email":"super.admin@exnodes.vn","password":"<from .env>"}' \
       | jq -r .data.access_token)
     test -n "$TOKEN" && echo "got token (${#TOKEN} chars)"
     ```

     Expected: prints `got token (NNN chars)`. Record only the length, not the token itself.

  2. **Create root department**:

     ```bash
     ROOT_ID=$(curl -s -X POST http://localhost:8080/api/v1/departments \
       -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
       -d '{"name":"VerifyRoot","description":"e2e root"}' \
       | jq -r .data.id)
     echo "ROOT_ID=$ROOT_ID"
     ```

     Expected: a UUID. HTTP 201. Save the response body to the log.

  3. **Create child department** (parent_id = ROOT_ID):

     ```bash
     CHILD_ID=$(curl -s -X POST http://localhost:8080/api/v1/departments \
       -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
       -d "{\"name\":\"VerifyChild\",\"parent_id\":\"$ROOT_ID\"}" \
       | jq -r .data.id)
     echo "CHILD_ID=$CHILD_ID"
     ```

     Expected: HTTP 201, response contains `"parent_id":"$ROOT_ID"`.

  4. **Create position in child**:

     ```bash
     POS_ID=$(curl -s -X POST http://localhost:8080/api/v1/positions \
       -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
       -d "{\"name\":\"VerifyEngineer\",\"department_id\":\"$CHILD_ID\"}" \
       | jq -r .data.id)
     echo "POS_ID=$POS_ID"
     ```

     Expected: HTTP 201, response contains `"department_id":"$CHILD_ID"`.

  5. **List departments filtered to root only**:

     ```bash
     curl -s "http://localhost:8080/api/v1/departments?parent_id=root&search=Verify" \
       -H "Authorization: Bearer $TOKEN" | jq '.data.items[].name'
     ```

     Expected: includes `"VerifyRoot"`, does NOT include `"VerifyChild"`.

  6. **List positions filtered by department**:

     ```bash
     curl -s "http://localhost:8080/api/v1/positions?department_id=$CHILD_ID" \
       -H "Authorization: Bearer $TOKEN" | jq '.data.items[].name'
     ```

     Expected: contains `"VerifyEngineer"`.

  7. **PATCH department description**:

     ```bash
     curl -s -X PATCH "http://localhost:8080/api/v1/departments/$CHILD_ID" \
       -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
       -d '{"description":"updated via e2e"}' | jq '.data.description'
     ```

     Expected: `"updated via e2e"`.

  8. **Try DELETE the child department while it still has a position (expect 409)**:

     ```bash
     curl -s -o /tmp/del-resp.json -w "%{http_code}\n" -X DELETE \
       "http://localhost:8080/api/v1/departments/$CHILD_ID" \
       -H "Authorization: Bearer $TOKEN"
     cat /tmp/del-resp.json | jq .
     ```

     Expected: status `409`, body contains message about positions assigned.

  9. **DELETE the position**:

     ```bash
     curl -s -o /tmp/posdel.json -w "%{http_code}\n" -X DELETE \
       "http://localhost:8080/api/v1/positions/$POS_ID" \
       -H "Authorization: Bearer $TOKEN"
     cat /tmp/posdel.json | jq .
     ```

     Expected: `200`, `success: true`.

  10. **DELETE the child department (now succeeds)**:

      ```bash
      curl -s -o /tmp/childdel.json -w "%{http_code}\n" -X DELETE \
        "http://localhost:8080/api/v1/departments/$CHILD_ID" \
        -H "Authorization: Bearer $TOKEN"
      cat /tmp/childdel.json | jq .
      ```

      Expected: `200`.

  11. **List again, confirm child is gone**:

      ```bash
      curl -s "http://localhost:8080/api/v1/departments?search=VerifyChild" \
        -H "Authorization: Bearer $TOKEN" | jq '.data.total'
      ```

      Expected: `0`.

  12. **Error path — invalid parent_id on create**:

      ```bash
      curl -s -o /tmp/badparent.json -w "%{http_code}\n" -X POST \
        http://localhost:8080/api/v1/departments \
        -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
        -d '{"name":"Orphan","parent_id":"00000000-0000-0000-0000-000000000000"}'
      cat /tmp/badparent.json | jq .
      ```

      Expected: `400`, message `"Parent department not found"`.

  13. **Error path — duplicate name**:

      ```bash
      curl -s -o /tmp/dup.json -w "%{http_code}\n" -X POST \
        http://localhost:8080/api/v1/departments \
        -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
        -d '{"name":"VerifyRoot"}'
      cat /tmp/dup.json | jq .
      ```

      Expected: `409`, message `"Department name already exists"`.

  14. **Error path — missing permission** (log in as a non-admin seeded user, or create a test user without the perm):

      ```bash
      # assumes an Employee-role user already seeded for testing; otherwise create one
      EMP_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
        -H 'Content-Type: application/json' \
        -d '{"email":"employee.test@exnodes.vn","password":"<from seed>"}' \
        | jq -r .data.access_token)

      curl -s -o /tmp/forb.json -w "%{http_code}\n" -X POST \
        http://localhost:8080/api/v1/departments \
        -H "Authorization: Bearer $EMP_TOKEN" -H 'Content-Type: application/json' \
        -d '{"name":"NotAllowed"}'
      cat /tmp/forb.json | jq .
      ```

      Expected: `403`, message indicating missing permission. If no employee user is seeded yet, create one via the users endpoint with the super admin token first and include that step in the log.

  15. **DB spot-check soft-delete columns**:

      ```bash
      psql "$DATABASE_URL" -c "SELECT name, is_deleted, deleted_at IS NOT NULL AS has_deleted_at FROM departments WHERE name = 'VerifyChild';"
      ```

      Expected: one row, `is_deleted = t`, `has_deleted_at = t`.

  16. **Cleanup root**:

      ```bash
      curl -s -X DELETE "http://localhost:8080/api/v1/departments/$ROOT_ID" \
        -H "Authorization: Bearer $TOKEN" | jq .
      ```

      Expected: `200`. (Children already soft-deleted, so the cascade guard passes.)

- [ ] Stop the server:

  ```bash
  pkill -f "exnodes-hrm-api" || kill $SERVER_PID || true
  ```

- [ ] Write all 16 sub-sections into `docs/superpowers/verification/phase-03.md`. Use the section title from each numbered step and paste **command + HTTP code + abridged response JSON** (no secrets, no tokens).

- [ ] Commit:

  ```bash
  git add docs/superpowers/verification/phase-03.md
  git commit -m "docs(verification): phase-03 end-to-end log (departments + positions)"
  ```

---

### Task 18 — Update `README.md`

- [ ] Open `README.md`. In the **Endpoints** section, add a new sub-section after Phase 2's `Users`:

  ```md
  ### Departments

  | Method | Path                          | Permission             |
  |--------|-------------------------------|------------------------|
  | GET    | /api/v1/departments           | departments:read       |
  | POST   | /api/v1/departments           | departments:create     |
  | GET    | /api/v1/departments/{id}      | departments:read       |
  | PATCH  | /api/v1/departments/{id}      | departments:update     |
  | DELETE | /api/v1/departments/{id}      | departments:delete     |

  Supports self-referential `parent_id` (UUID or `"root"` filter on list).

  ### Positions

  | Method | Path                          | Permission             |
  |--------|-------------------------------|------------------------|
  | GET    | /api/v1/positions             | positions:read         |
  | POST   | /api/v1/positions             | positions:create       |
  | GET    | /api/v1/positions/{id}        | positions:read         |
  | PATCH  | /api/v1/positions/{id}        | positions:update       |
  | DELETE | /api/v1/positions/{id}        | positions:delete       |

  Each position belongs to exactly one department; delete is blocked while users are assigned.
  ```

- [ ] Commit:

  ```bash
  git add README.md
  git commit -m "docs(readme): list phase 3 department + position endpoints"
  ```

---

## Definition of Done (checklist)

- [ ] Migration `000004` up + down applies cleanly forward and reverse on a fresh DB.
- [ ] `departments` and `positions` tables have audit cols, `set_updated_at` trigger, `is_deleted` index.
- [ ] `users.department_id` and `users.position_id` have FK constraints with `ON DELETE SET NULL`.
- [ ] Models embed `BaseModel`; soft-delete sets **both** `is_deleted` and `deleted_at`.
- [ ] All five department routes + all five position routes return Swagger entries.
- [ ] Every route declares its required permission inline via `middleware.RequirePerms(...)`.
- [ ] Service tests pass: `go test ./internal/services/... -run 'Department|Position' -count=1 -v`.
- [ ] Full test suite passes: `go test ./... -count=1`.
- [ ] `make migrate-up && make run` boots clean; `/health` returns ok; `/swagger/index.html` shows both new tag groups.
- [ ] End-to-end self-verification log committed at `docs/superpowers/verification/phase-03.md` containing all 16 walk steps with commands + responses.
- [ ] `README.md` updated.
- [ ] No `db.AutoMigrate` calls anywhere in the codebase (`grep -R AutoMigrate internal cmd | wc -l` returns 0).
- [ ] No `FAIL` lines in `go test` output.

## Out of scope (this phase)

- Bulk endpoints, CSV import/export.
- Department tree return endpoint (`GET /departments/tree`) — defer until FE requests it; current list supports parent filtering.
- Restore-from-soft-delete endpoint (covered globally in a later phase).
- Permission-group endpoint changes beyond appending the two new groups.
