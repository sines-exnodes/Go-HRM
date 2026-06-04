# Roles & Permissions API Parity — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring Go's role-management API to parity with the Python source — add full role CRUD (list/get/create/update/delete), a role `level` authority hierarchy, and the supporting validation/guards — matching BA US-004 and the locked decisions in `specs/2026-06-04-roles-permissions-parity-audit.md`.

**Architecture:** Standard one-directional layering (`handler → service → repository → GORM`). A new `RoleService` owns all business logic; the existing `roleRepository` gains `List`, `SoftDelete`, and `CountUsersWithRole`; the existing `RoleHandler` (catalog-only today) gains five CRUD handlers; `UserService.AssignRoles` gains a level-authority check. One migration (`000020`) adds the `level` column.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, golang-migrate (versioned SQL), testify + a real Postgres test DB.

**Locked decisions (from the audit):**
- **D1** ADD `level` (1–100) + assignment-authority (assigner can only grant roles with `level ≤ their own max level`).
- **D2** soft delete (Go `NotDeleted` convention); name freed for reuse.
- **D3** list sorts by `level` ASC, then `name` ASC.
- Gate `GET /roles/permissions` behind `roles:read`. One list endpoint returns full `permissions[]` + `permission_count`. `is_system` rename/level-change/delete guards. Name regex + permission-string validation.
- Out of scope: the `approve_team`/`approve_all` registry delta (leave-requests pass).

**Seed levels (system roles):** Super Admin 100, Admin 90, HR Manager 80, Manager 50, Employee 10.

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `migrations/000020_roles_add_level.up.sql` / `.down.sql` | add `level` column + CHECK + backfill seed roles | Create |
| `internal/models/role.go` | add `Level int` field | Modify |
| `internal/dto/role.go` | RoleCreate / RoleUpdate / RoleRead / RoleListQuery + validators | Create |
| `internal/repositories/role_repo.go` | add `List`, `SoftDelete`, `CountUsersWithRole` | Modify |
| `internal/services/role_service.go` | CRUD + uniqueness + is_system guards + delete-blocked | Create |
| `internal/services/user_service.go` | add `roles` repo dep; level-authority in `AssignRoles` | Modify |
| `internal/handlers/role_handler.go` | List/Get/Create/Update/Delete handlers | Modify |
| `internal/services/seed_service.go` | set `Level` on seeded roles | Modify |
| `cmd/server/main.go` | construct `RoleService`, wire routes, update `NewUserService` call | Modify |
| `internal/services/role_service_test.go` | role-service integration tests | Create |
| `internal/services/user_service_test.go` | level-authority test(s) | Modify |
| `docs/swagger/` | regenerated | Generated |
| `docs/superpowers/verification/roles-permissions-parity.md` | verification log | Create |

**Note on the test DB:** the suite (`internal/services/testhelper_test.go`) drops & re-applies all migrations in `TestMain`, so the new migration is picked up automatically once the files exist. Run tests with `TEST_DATABASE_URL` set (see CHECKPOINT "Local environment notes": `postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable`).

---

## Task 1: Migration 000020 + model `Level` field

**Files:**
- Create: `migrations/000020_roles_add_level.up.sql`
- Create: `migrations/000020_roles_add_level.down.sql`
- Modify: `internal/models/role.go`

- [ ] **Step 1: Write the up migration**

`migrations/000020_roles_add_level.up.sql`:
```sql
-- =========================================================================
-- 000020_roles_add_level
-- Adds the role authority `level` (1..100). Python parity: an assigner may
-- only grant roles at or below their own max level. Backfills the known
-- system roles by name; any other existing row keeps the DEFAULT 100.
-- =========================================================================
ALTER TABLE roles
    ADD COLUMN level INT NOT NULL DEFAULT 100
        CHECK (level BETWEEN 1 AND 100);

UPDATE roles SET level = 100 WHERE name = 'Super Admin';
UPDATE roles SET level = 90  WHERE name = 'Admin';
UPDATE roles SET level = 80  WHERE name = 'HR Manager';
UPDATE roles SET level = 50  WHERE name = 'Manager';
UPDATE roles SET level = 10  WHERE name = 'Employee';
```

- [ ] **Step 2: Write the down migration**

`migrations/000020_roles_add_level.down.sql`:
```sql
ALTER TABLE roles DROP COLUMN level;
```

- [ ] **Step 3: Add the `Level` field to the model**

In `internal/models/role.go`, add `Level` to the `Role` struct (after `IsSystem`):
```go
// Role maps to the roles table.
type Role struct {
	BaseModel
	Name        string      `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Description string      `gorm:"type:text;not null;default:''" json:"description"`
	Level       int         `gorm:"not null;default:100" json:"level"`
	IsSystem    bool        `gorm:"not null;default:false" json:"is_system"`
	Permissions StringSlice `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"permissions"`
}
```

- [ ] **Step 4: Verify build + migration round-trip**

Run: `make build`
Expected: compiles clean.

Run (with test DB reachable): `make test-db-up` then `go test ./internal/services/ -run TestSeed -v`
Expected: PASS (the suite re-applies migrations including 000020; seed tests still green). If `TEST_DATABASE_URL` is unset, tests skip — that's acceptable here but you MUST run them with the DB set before claiming done (AGENTS Rule 12).

- [ ] **Step 5: Commit**

```bash
git add migrations/000020_roles_add_level.up.sql migrations/000020_roles_add_level.down.sql internal/models/role.go
git commit -m "feat(roles): migration 000020 add role level column + model field"
```

---

## Task 2: Role DTOs + validators

**Files:**
- Create: `internal/dto/role.go`

- [ ] **Step 1: Write the DTO file**

`internal/dto/role.go`:
```go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// RoleCreate is the request body for POST /api/v1/roles.
type RoleCreate struct {
	Name        string   `json:"name"                  binding:"required,min=1,max=100"`
	Description string   `json:"description,omitempty" binding:"max=1000"`
	Level       int      `json:"level"                 binding:"required,min=1,max=100"`
	Permissions []string `json:"permissions,omitempty"`
}

// RoleUpdate is the PATCH body for /api/v1/roles/:id — pointer fields, only
// provided fields change. Permissions is a pointer so "omitted" (no change)
// is distinguishable from "[]" (revoke all).
type RoleUpdate struct {
	Name        *string   `json:"name,omitempty"        binding:"omitempty,min=1,max=100"`
	Description *string   `json:"description,omitempty" binding:"omitempty,max=1000"`
	Level       *int      `json:"level,omitempty"       binding:"omitempty,min=1,max=100"`
	Permissions *[]string `json:"permissions,omitempty"`
}

// RoleRead is the wire shape returned by every role endpoint. Superset of the
// two Python list shapes: full permissions[] AND permission_count.
type RoleRead struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Level           int       `json:"level"`
	Permissions     []string  `json:"permissions"`
	PermissionCount int       `json:"permission_count"`
	IsSystem        bool      `json:"is_system"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RoleListQuery binds the querystring for GET /api/v1/roles.
type RoleListQuery struct {
	Page     int    `form:"page,default=1"       binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
}
```

- [ ] **Step 2: Verify build**

Run: `make build`
Expected: compiles clean (DTO is unused until Task 4 — that's fine, it's a type decl).

- [ ] **Step 3: Commit**

```bash
git add internal/dto/role.go
git commit -m "feat(roles): add role request/response DTOs"
```

---

## Task 3: Repository additions (List, SoftDelete, CountUsersWithRole)

**Files:**
- Modify: `internal/repositories/role_repo.go`

- [ ] **Step 1: Write failing tests for the new repo methods**

Append to a new `internal/repositories/role_repo_test.go` is NOT the project pattern (repo logic is tested via the service layer). Instead, the behaviours are covered by the service tests in Task 4. Skip a standalone repo test; proceed to implement, then Task 4 exercises these methods end-to-end.

- [ ] **Step 2: Add `RoleFilter` + extend the interface**

In `internal/repositories/role_repo.go`, add the filter type and the three methods to the `RoleRepository` interface:
```go
// RoleFilter mirrors dto.RoleListQuery in a service-agnostic shape.
type RoleFilter struct {
	Page     int
	PageSize int
	Search   string
}

// RoleRepository defines data access for roles.
type RoleRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	Update(ctx context.Context, role *models.Role) error
	List(ctx context.Context, f RoleFilter) ([]models.Role, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// CountUsersWithRole counts non-deleted user_roles rows referencing the role.
	CountUsersWithRole(ctx context.Context, id uuid.UUID) (int64, error)
}
```

- [ ] **Step 3: Implement the three methods**

Add to `internal/repositories/role_repo.go` (and add `"strings"` + `"github.com/exnodes/hrm-api/pkg/utils"` to imports):
```go
func (r *roleRepository) List(ctx context.Context, f RoleFilter) ([]models.Role, int64, error) {
	q := r.db.WithContext(ctx).Scopes(notDeleted).Model(&models.Role{})
	if s := strings.TrimSpace(f.Search); s != "" {
		q = q.Where("name ILIKE ?", utils.BuildILIKEPattern(s))
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
	var items []models.Role
	err := q.
		Order("level ASC").
		Order("LOWER(name) ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error
	return items, total, err
}

func (r *roleRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": gorm.Expr("NOW()")}).Error
}

// CountUsersWithRole counts live user_roles rows referencing the role. The
// join table carries its own is_deleted column (migration 000002).
func (r *roleRepository) CountUsersWithRole(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Where("role_id = ? AND is_deleted = ?", id, false).
		Count(&count).Error
	return count, err
}
```

> Note: `FindByName` currently returns the raw `gorm.ErrRecordNotFound`. The service (Task 4) handles that. Keep its signature unchanged to avoid touching `seed_service.go`/`auth_service.go` callers.

- [ ] **Step 4: Verify build**

Run: `make build`
Expected: compiles clean.

- [ ] **Step 5: Commit**

```bash
git add internal/repositories/role_repo.go
git commit -m "feat(roles): repo List + SoftDelete + CountUsersWithRole"
```

---

## Task 4: RoleService (CRUD + guards) — TDD

**Files:**
- Create: `internal/services/role_service.go`
- Create: `internal/services/role_service_test.go`

- [ ] **Step 1: Write the failing tests**

`internal/services/role_service_test.go`:
```go
package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newRoleSvc(t *testing.T) *services.RoleService {
	t.Helper()
	return services.NewRoleService(repositories.NewRoleRepository(testDB))
}

func TestRoleService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	r, err := svc.Create(ctx, dto.RoleCreate{
		Name:        "  Auditor  ",
		Description: "  reads things  ",
		Level:       40,
		Permissions: []string{string(permissions.PermUsersRead), string(permissions.PermRolesRead)},
	})
	require.NoError(t, err)
	require.Equal(t, "Auditor", r.Name)            // trimmed
	require.Equal(t, "reads things", r.Description) // trimmed
	require.Equal(t, 40, r.Level)
	require.False(t, r.IsSystem)
	require.Equal(t, 2, r.PermissionCount)
	require.Len(t, r.Permissions, 2)
}

func TestRoleService_Create_DuplicateName_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "Reviewer", Level: 30})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "reviewer", Level: 30}) // case-insensitive
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestRoleService_Create_InvalidName_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "bad@name!", Level: 30})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestRoleService_Create_UnknownPermission_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "Weird", Level: 30, Permissions: []string{"made:up"}})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestRoleService_Update_Partial_And_SystemGuards(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	// A normal role: rename + change level + revoke all perms.
	r, err := svc.Create(ctx, dto.RoleCreate{Name: "Temp", Level: 30, Permissions: []string{string(permissions.PermUsersRead)}})
	require.NoError(t, err)
	newName := "Temp Renamed"
	newLevel := 35
	empty := []string{}
	upd, err := svc.Update(ctx, r.ID, dto.RoleUpdate{Name: &newName, Level: &newLevel, Permissions: &empty})
	require.NoError(t, err)
	require.Equal(t, "Temp Renamed", upd.Name)
	require.Equal(t, 35, upd.Level)
	require.Equal(t, 0, upd.PermissionCount)

	// A system role: rename and level-change are rejected.
	sys := makeRole(t, "System One", []permissions.Permission{permissions.PermAuthLogin}, true)
	renamed := "Nope"
	_, err = svc.Update(ctx, sys.ID, dto.RoleUpdate{Name: &renamed})
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)

	lvl := 5
	_, err = svc.Update(ctx, sys.ID, dto.RoleUpdate{Level: &lvl})
	require.Error(t, err)

	// But a system role's permissions CAN be updated (mirrors merge-seed intent).
	perms := []string{string(permissions.PermAuthLogin), string(permissions.PermUsersRead)}
	okUpd, err := svc.Update(ctx, sys.ID, dto.RoleUpdate{Permissions: &perms})
	require.NoError(t, err)
	require.Equal(t, 2, okUpd.PermissionCount)
}

func TestRoleService_Delete_SystemRole_Rejected(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	sys := makeRole(t, "Protected", []permissions.Permission{permissions.PermAuthLogin}, true)
	err := svc.Delete(ctx, sys.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestRoleService_Delete_BlockedByAssignedUsers_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	role := makeRole(t, "Assigned", []permissions.Permission{permissions.PermAuthLogin}, false)
	makeUser(t, "holder@example.com", "pw-Aa123456", role)

	err := svc.Delete(ctx, role.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeConflict, ae.Code)
	require.Contains(t, ae.Message, "reassign")
}

func TestRoleService_Delete_SoftDeletes_And_NameReusable(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	r, err := svc.Create(ctx, dto.RoleCreate{Name: "Recyclable", Level: 30})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(ctx, r.ID))

	// Soft-deleted: invisible to Get.
	_, err = svc.Get(ctx, r.ID)
	require.Error(t, err)
	ae, ok := apperrors.As(err)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeNotFound, ae.Code)

	// Name is reusable after delete.
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "Recyclable", Level: 30})
	require.NoError(t, err)
}

func TestRoleService_List_SortedByLevelThenName_WithSearch(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newRoleSvc(t)

	_, err := svc.Create(ctx, dto.RoleCreate{Name: "Zeta", Level: 10})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "Alpha", Level: 90})
	require.NoError(t, err)
	_, err = svc.Create(ctx, dto.RoleCreate{Name: "Beta", Level: 10})
	require.NoError(t, err)

	res, err := svc.List(ctx, dto.RoleListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(3), res.Total)
	// level ASC, then name ASC: Beta(10), Zeta(10), Alpha(90)
	require.Equal(t, "Beta", res.Items[0].Name)
	require.Equal(t, "Zeta", res.Items[1].Name)
	require.Equal(t, "Alpha", res.Items[2].Name)

	// Search by name (case-insensitive, partial).
	res, err = svc.List(ctx, dto.RoleListQuery{Page: 1, PageSize: 10, Search: "et"})
	require.NoError(t, err)
	require.Equal(t, int64(2), res.Total) // Zeta + Beta
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/services/ -run TestRoleService -v`
Expected: FAIL — `undefined: services.NewRoleService` (compile error). Good.

- [ ] **Step 3: Implement the service**

`internal/services/role_service.go`:
```go
package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// roleNamePattern mirrors the Python source: letters, digits, spaces, hyphens,
// ampersands. The trailing '-' in the class is a literal hyphen.
var roleNamePattern = regexp.MustCompile(`^[a-zA-Z0-9 &-]+$`)

// RoleService owns role-management business logic.
type RoleService struct {
	repo repositories.RoleRepository
}

func NewRoleService(repo repositories.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func roleToRead(r *models.Role) dto.RoleRead {
	perms := make([]string, 0, len(r.Permissions))
	perms = append(perms, r.Permissions...)
	return dto.RoleRead{
		ID:              r.ID,
		Name:            r.Name,
		Description:     r.Description,
		Level:           r.Level,
		Permissions:     perms,
		PermissionCount: len(perms),
		IsSystem:        r.IsSystem,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func validateRoleName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", apperrors.ErrBadRequest("Role name is required")
	}
	if len(name) > 100 {
		return "", apperrors.ErrBadRequest("Role name must not exceed 100 characters")
	}
	if !roleNamePattern.MatchString(name) {
		return "", apperrors.ErrBadRequest("Role name can only contain letters, numbers, spaces, hyphens, and ampersands")
	}
	return name, nil
}

// validatePermissions rejects unknown permission strings. The wildcard '*' is
// allowed (permissions.IsValid returns true for it).
func validatePermissions(perms []string) (models.StringSlice, error) {
	out := make(models.StringSlice, 0, len(perms))
	var invalid []string
	for _, p := range perms {
		if !permissions.IsValid(permissions.Permission(p)) {
			invalid = append(invalid, p)
			continue
		}
		out = append(out, p)
	}
	if len(invalid) > 0 {
		return nil, apperrors.ErrBadRequest("Unknown permissions: " + strings.Join(invalid, ", "))
	}
	return out, nil
}

func (s *RoleService) checkNameUnique(ctx context.Context, name string, excludeID *uuid.UUID) error {
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if excludeID != nil && existing.ID == *excludeID {
		return nil
	}
	return apperrors.ErrConflict("Role name already exists")
}

func (s *RoleService) Create(ctx context.Context, in dto.RoleCreate) (*dto.RoleRead, error) {
	name, err := validateRoleName(in.Name)
	if err != nil {
		return nil, err
	}
	if err := s.checkNameUnique(ctx, name, nil); err != nil {
		return nil, err
	}
	perms, err := validatePermissions(in.Permissions)
	if err != nil {
		return nil, err
	}
	r := &models.Role{
		Name:        name,
		Description: strings.TrimSpace(in.Description),
		Level:       in.Level,
		IsSystem:    false,
		Permissions: perms,
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}
	out := roleToRead(r)
	return &out, nil
}

func (s *RoleService) Get(ctx context.Context, id uuid.UUID) (*dto.RoleRead, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Role")
		}
		return nil, err
	}
	out := roleToRead(r)
	return &out, nil
}

func (s *RoleService) Update(ctx context.Context, id uuid.UUID, in dto.RoleUpdate) (*dto.RoleRead, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Role")
		}
		return nil, err
	}

	if in.Name != nil {
		name, err := validateRoleName(*in.Name)
		if err != nil {
			return nil, err
		}
		if r.IsSystem && name != r.Name {
			return nil, apperrors.ErrBadRequest("Cannot rename a system role")
		}
		if err := s.checkNameUnique(ctx, name, &r.ID); err != nil {
			return nil, err
		}
		r.Name = name
	}
	if in.Level != nil {
		if r.IsSystem && *in.Level != r.Level {
			return nil, apperrors.ErrBadRequest("Cannot change the level of a system role")
		}
		r.Level = *in.Level
	}
	if in.Description != nil {
		r.Description = strings.TrimSpace(*in.Description)
	}
	if in.Permissions != nil {
		perms, err := validatePermissions(*in.Permissions)
		if err != nil {
			return nil, err
		}
		r.Permissions = perms
	}

	if err := s.repo.Update(ctx, r); err != nil {
		return nil, err
	}
	out := roleToRead(r)
	return &out, nil
}

func (s *RoleService) Delete(ctx context.Context, id uuid.UUID) error {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("Role")
		}
		return err
	}
	if r.IsSystem {
		return apperrors.ErrBadRequest("Cannot delete a system role")
	}
	count, err := s.repo.CountUsersWithRole(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		word := "user is"
		if count > 1 {
			word = "users are"
		}
		return apperrors.ErrConflict(fmt.Sprintf(
			"Cannot delete role '%s' — %d %s currently assigned. Please reassign them before deleting.", r.Name, count, word))
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *RoleService) List(ctx context.Context, q dto.RoleListQuery) (*dto.PaginatedData[dto.RoleRead], error) {
	items, total, err := s.repo.List(ctx, repositories.RoleFilter{Page: q.Page, PageSize: q.PageSize, Search: q.Search})
	if err != nil {
		return nil, err
	}
	reads := make([]dto.RoleRead, 0, len(items))
	for i := range items {
		reads = append(reads, roleToRead(&items[i]))
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
	return &dto.PaginatedData[dto.RoleRead]{
		Items:      reads,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}
```

> The delete-blocked message contains "reassign" (lowercase) so the test's `require.Contains(..., "reassign")` matches — keep the casing consistent if you reword.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/services/ -run TestRoleService -v`
Expected: all PASS (with `TEST_DATABASE_URL` set).

- [ ] **Step 5: Commit**

```bash
git add internal/services/role_service.go internal/services/role_service_test.go
git commit -m "feat(roles): RoleService CRUD with uniqueness, system guards, delete-blocked"
```

---

## Task 5: Level-authority in UserService.AssignRoles — TDD

**Files:**
- Modify: `internal/services/user_service.go`
- Modify: `cmd/server/main.go` (constructor call only)
- Modify: `internal/services/user_service_test.go`
- Modify: `internal/services/invite_service_test.go` (constructor arity)

> **⚠️ Behavioral change — read first.** The existing `TestUserService_AssignRoles`
> (user_service_test.go:105) uses a **role-less admin** (max level 0) to assign a
> role whose level is 100 (the DB default `makeRole` produces). Once the authority
> check lands, that assignment is correctly *forbidden* (Python would reject it
> too — `get_user_max_role_level` returns 0). So this task **must update that
> existing test** to give the admin a sufficient-level role. This is expected, not
> a regression. Two `NewUserService` call sites also change arity:
> `user_service_test.go:19` (`newUserSvc` helper) and `invite_service_test.go:54`.

- [ ] **Step 1: Add the role repo dependency + authority logic**

In `internal/services/user_service.go`:

(a) add the `roles` field + constructor param (insert as the SECOND param, after `users`):
```go
type UserService struct {
	users    repositories.UserRepository
	roles    repositories.RoleRepository
	emps     repositories.EmployeeRepository
	tokens   *repositories.DeviceTokenRepository
	settings *repositories.NotificationSettingsRepository
	empSvc   *EmployeeService
}

func NewUserService(
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	emps repositories.EmployeeRepository,
	tokens *repositories.DeviceTokenRepository,
	settings *repositories.NotificationSettingsRepository,
	empSvc *EmployeeService,
) *UserService {
	return &UserService{users: users, roles: roles, emps: emps, tokens: tokens, settings: settings, empSvc: empSvc}
}
```

(b) replace `AssignRoles` with the authority-checked version (add `"fmt"` to imports if absent):
```go
func (s *UserService) AssignRoles(ctx context.Context, id uuid.UUID, roleIDs []uuid.UUID, admin *models.User) error {
	if admin.ID == id {
		return apperrors.ErrBadRequest("You cannot change your own role")
	}
	if err := s.checkRoleAssignmentAuthority(ctx, admin, roleIDs); err != nil {
		return err
	}
	return s.users.AssignRoles(ctx, id, roleIDs)
}

// checkRoleAssignmentAuthority ports the Python rule (app/services/role.py):
// an assigner may only grant roles whose level is <= the assigner's own max
// role level. admin.Roles is preloaded by the JWT middleware on the request
// path; tests must pass a *models.User whose Roles slice is populated.
func (s *UserService) checkRoleAssignmentAuthority(ctx context.Context, admin *models.User, roleIDs []uuid.UUID) error {
	if len(roleIDs) == 0 {
		return nil
	}
	assignerMax := 0
	for _, r := range admin.Roles {
		if r.Level > assignerMax {
			assignerMax = r.Level
		}
	}
	targetRoles, err := s.roles.FindByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	for _, r := range targetRoles {
		if r.Level > assignerMax {
			return apperrors.ErrForbidden(fmt.Sprintf(
				"Cannot assign role '%s' (level %d): exceeds your authority level (%d)", r.Name, r.Level, assignerMax))
		}
	}
	return nil
}
```
> Verified present: `apperrors.ErrForbidden(msg string)` and `apperrors.CodeForbidden` (internal/errors/errors.go:90,14).

- [ ] **Step 2: Update the production constructor call**

In `cmd/server/main.go`, find `services.NewUserService(...)` and insert `roleRepo` as the second argument. `roleRepo` already exists (line ~52). Result (match the real repo var names already in `main.go`):
```go
userSvc := services.NewUserService(userRepo, roleRepo, employeeRepo, deviceTokenRepo, notifSettingsRepo, empSvc)
```

- [ ] **Step 3: Update both test call sites for the new arity**

(a) `internal/services/user_service_test.go:19` — the `newUserSvc` helper. Add the role repo as the second arg:
```go
func newUserSvc(db *gorm.DB, empSvc *services.EmployeeService) *services.UserService {
	return services.NewUserService(
		repositories.NewUserRepository(db),
		repositories.NewRoleRepository(db),
		repositories.NewEmployeeRepository(db),
		repositories.NewDeviceTokenRepository(db),
		repositories.NewNotificationSettingsRepository(db),
		empSvc,
	)
}
```
(b) `internal/services/invite_service_test.go:54` — add `repositories.NewRoleRepository(...)` (or the existing `empRepo`/`testRoleRepo` in scope) as the second arg to match the new signature.

- [ ] **Step 4: Fix the existing `TestUserService_AssignRoles` (now needs an authorized admin)**

Replace the admin fixture + the assignment call in `TestUserService_AssignRoles` (user_service_test.go ~105-133) so the admin holds a role with enough authority. `makeRole` produces level-100 roles (DB default), so give the admin a level-100 role and reload it with roles populated:
```go
	role := makeRole(t, "manager", []permissions.Permission{permissions.PermEmployeesRead}, false) // level 100 (DB default)
	target, err := empSvc.Create(ctx, dto.EmployeeCreate{
		Email: "target@example.com", Password: "Pass12345", FirstName: "Target", LastName: "Test",
	})
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(testDB)
	adminRole := makeRole(t, "admin-authority", []permissions.Permission{permissions.PermUsersManageRoles}, false) // level 100
	adminBare := makeUser(t, "admin2@example.com", "Pass12345", adminRole)

	// Self-change guard fires before authority — a bare struct is fine here.
	err = userSvc.AssignRoles(ctx, adminBare.ID, []uuid.UUID{role.ID}, adminBare)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// Authority check needs admin.Roles populated → reload with roles.
	admin, err := userRepo.FindByIDWithRoles(ctx, adminBare.ID)
	require.NoError(t, err)

	// Assign to a different user OK (admin level 100 >= role level 100).
	require.NoError(t, userSvc.AssignRoles(ctx, target.UserID, []uuid.UUID{role.ID}, admin))
	u, err := userRepo.FindByIDWithRoles(ctx, target.UserID)
	require.NoError(t, err)
	require.Len(t, u.Roles, 1)
	assert.Equal(t, "manager", u.Roles[0].Name)
```

- [ ] **Step 5: Add the new level-authority test**

Append to `internal/services/user_service_test.go`:
```go
func TestUserService_AssignRoles_LevelAuthority(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	empSvc, _ := newEmpSvc(testDB)
	userSvc := newUserSvc(testDB, empSvc)
	ctx := context.Background()
	userRepo := repositories.NewUserRepository(testDB)

	// Assigner holds a level-50 role.
	mid := makeRole(t, "Mid", []permissions.Permission{permissions.PermUsersManageRoles}, false)
	mid.Level = 50
	require.NoError(t, testRoleRepo.Update(ctx, mid))
	high := makeRole(t, "High", []permissions.Permission{permissions.PermAuthLogin}, false)
	high.Level = 90
	require.NoError(t, testRoleRepo.Update(ctx, high))
	low := makeRole(t, "Low", []permissions.Permission{permissions.PermAuthLogin}, false)
	low.Level = 10
	require.NoError(t, testRoleRepo.Update(ctx, low))

	adminBare := makeUser(t, "assigner@example.com", "Pass12345", mid)
	admin, err := userRepo.FindByIDWithRoles(ctx, adminBare.ID)
	require.NoError(t, err)
	target := makeUser(t, "lvl-target@example.com", "Pass12345", low)

	// Granting a level-90 role exceeds the assigner's level-50 authority.
	err = userSvc.AssignRoles(ctx, target.ID, []uuid.UUID{high.ID}, admin)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeForbidden, ae.Code)

	// Granting a level-10 role is within authority.
	require.NoError(t, userSvc.AssignRoles(ctx, target.ID, []uuid.UUID{low.ID}, admin))
}
```

- [ ] **Step 6: Run the full services suite**

Run: `go test ./internal/services/ -v`
Expected: all PASS — the rewritten `TestUserService_AssignRoles`, the new
`TestUserService_AssignRoles_LevelAuthority`, and every other suite (0 skip with
`TEST_DATABASE_URL` set).

- [ ] **Step 7: Commit**

```bash
git add internal/services/user_service.go internal/services/user_service_test.go internal/services/invite_service_test.go cmd/server/main.go
git commit -m "feat(roles): level-based role-assignment authority in AssignRoles"
```

---

## Task 6: Role handlers + route wiring + seed levels + gate catalog

**Files:**
- Modify: `internal/handlers/role_handler.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/services/seed_service.go`

- [ ] **Step 1: Rewrite the handler with CRUD + an injected service**

Replace `internal/handlers/role_handler.go` with:
```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// RoleHandler handles role-management endpoints.
type RoleHandler struct {
	svc *services.RoleService
}

// NewRoleHandler constructs a RoleHandler.
func NewRoleHandler(svc *services.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }

// ListPermissions godoc
// @Summary      List grouped permissions (for the role-creation picker)
// @Description  Returns the structured permission catalog. Requires roles:read.
// @Tags         Roles
// @Produce      json
// @Success      200  {object}  dto.Response[[]permissions.PermissionGroup]
// @Security     BearerAuth
// @Router       /api/v1/roles/permissions [get]
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Response[[]permissions.PermissionGroup]{
		Success: true,
		Data:    permissions.PermissionGroups,
	})
}

// List godoc
// @Summary      List roles (paginated, name search)
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(10)
// @Param        search     query  string  false  "Substring match on name (ILIKE)"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	var q dto.RoleListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.PaginatedData[dto.RoleRead]]{Success: true, Data: data})
}

// Get godoc
// @Summary      Get role by ID
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Role UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles/{id} [get]
func (h *RoleHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.Response[*dto.RoleRead]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.RoleCreate  true  "Role payload"
// @Success      201  {object}  map[string]interface{}
// @Router       /api/v1/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var in dto.RoleCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.RoleRead]{Success: true, Message: "Role created successfully", Data: out})
}

// Update godoc
// @Summary      Update role (PATCH semantics)
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string          true  "Role UUID"
// @Param        body  body  dto.RoleUpdate  true  "Fields to update"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles/{id} [patch]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	var in dto.RoleUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.RoleRead]{Success: true, Message: "Role updated successfully", Data: out})
}

// Delete godoc
// @Summary      Delete a non-system role
// @Description  Soft-deletes the role (name becomes reusable). Rejected with 400
// @Description  for system roles and 409 if any user is still assigned.
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Role UUID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Role deleted"})
}
```

- [ ] **Step 2: Wire the service + routes in main.go**

(a) Construct the service near the other service constructors (after `roleRepo` exists):
```go
roleSvc := services.NewRoleService(roleRepo)
```
(b) Update the handler constructor:
```go
roleH := handlers.NewRoleHandler(roleSvc)
```
(c) Replace the single catalog route. Find:
```go
authed.GET("/roles/permissions", roleH.ListPermissions)
```
and replace with a full role group (place near the departments group, ordering the static `/permissions` BEFORE `/:id` so it isn't captured as an id):
```go
roles := authed.Group("/roles")
roles.GET("/permissions", middleware.RequirePerms(authSvc, permissions.PermRolesRead), roleH.ListPermissions)
roles.GET("", middleware.RequirePerms(authSvc, permissions.PermRolesRead), roleH.List)
roles.POST("", middleware.RequirePerms(authSvc, permissions.PermRolesCreate), roleH.Create)
roles.GET(":id", middleware.RequirePerms(authSvc, permissions.PermRolesRead), roleH.Get)
roles.PATCH(":id", middleware.RequirePerms(authSvc, permissions.PermRolesUpdate), roleH.Update)
roles.DELETE(":id", middleware.RequirePerms(authSvc, permissions.PermRolesDelete), roleH.Delete)
```
> Gin route conflict caution: a static segment (`/permissions`) and a wildcard (`/:id`) at the same level are fine in Gin as long as both are registered; if Gin panics about conflicting wildcards, register `/permissions` and `""` first (as above). Verify at boot.

- [ ] **Step 3: Set `Level` on seeded roles**

In `internal/services/seed_service.go`:

(a) add `Level` to the `roleSeed` struct:
```go
type roleSeed struct {
	Name        string
	Description string
	Level       int
	Permissions []permissions.Permission
}
```
(b) add `Level:` to each entry in `defaultRoles()`: Super Admin `100`, Admin `90`, HR Manager `80`, Manager `50`, Employee `10`.
(c) in `seedRoles`, set the level on the create path:
```go
r := &models.Role{
	Name:        rs.Name,
	Description: rs.Description,
	Level:       rs.Level,
	IsSystem:    true,
	Permissions: desired,
}
```
> Leave the merge path's level handling alone — migration 000020 backfills levels for already-existing system roles, and the merge path must not clobber a manually-edited level (matches the existing "merge only adds missing perms" contract).

- [ ] **Step 4: Build + boot smoke**

Run: `make build`
Expected: compiles clean.

Run: `make fmt && make vet`
Expected: no diffs / no vet errors.

Run the boot smoke (DB reachable, port free): `PORT=8082 ./bin/server` in one shell, then in another:
```bash
TOKEN=... # super-admin access token from /auth/login
curl -s localhost:8082/api/v1/roles -H "Authorization: Bearer $TOKEN" | jq '.data.items[].name'
```
Expected: the 5 seeded role names, sorted by level ascending (Employee, Manager, HR Manager, Admin, Super Admin).

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/role_handler.go cmd/server/main.go internal/services/seed_service.go
git commit -m "feat(roles): wire role CRUD routes, gate catalog behind roles:read, seed levels"
```

---

## Task 7: Swagger regen + full verification + docs

**Files:**
- Generated: `docs/swagger/`
- Create: `docs/superpowers/verification/roles-permissions-parity.md`
- Modify: `docs/superpowers/CHECKPOINT.md`

- [ ] **Step 1: Regenerate Swagger**

Run: `make swag`
Expected: `docs/swagger/` updates with the new role endpoints. Do NOT hand-edit.

- [ ] **Step 2: Run the full gate**

Run: `make fmt && make vet && make test`
Expected: build clean, vet clean, **all tests pass with 0 skipped** (set `TEST_DATABASE_URL`). If any test skips, the DB isn't wired — fix and rerun (AGENTS Rule 12). Capture the test summary line.

- [ ] **Step 3: Live HTTP smoke (end-to-end)**

With the server on `PORT=8082` and a super-admin token, exercise the full flow and record actual responses in the verification log:
1. `GET /api/v1/roles/permissions` → 200, grouped catalog.
2. `GET /api/v1/roles/permissions` with a token lacking `roles:read` → 403 (proves the gate).
3. `POST /api/v1/roles` `{"name":"QA Lead","level":40,"permissions":["users:read"]}` → 201.
4. Duplicate `POST` same name (any case) → 409.
5. `POST` with `"name":"bad@name"` → 400; `POST` with `"permissions":["made:up"]` → 400.
6. `GET /api/v1/roles` → 200, level-ASC ordering, `permission_count` present.
7. `PATCH /api/v1/roles/{id}` rename + permission change → 200.
8. `PATCH` a system role's name → 400; its permissions → 200.
9. `DELETE` the QA Lead role (no users) → 200; `GET` it → 404; recreate same name → 201.
10. Assign the role to a user, then `DELETE` → 409 with the assigned-count message.
11. Level-authority: as a level-50 actor, `PUT /users/{id}/roles` granting a level-90 role → 403.
12. DB spot-check: `SELECT name, level, is_deleted FROM roles ORDER BY level;`

- [ ] **Step 4: Write the verification log**

Create `docs/superpowers/verification/roles-permissions-parity.md` with: build/vet output, test summary (N tests / 0 skip / 0 fail), the 12 smoke steps with actual status codes + bodies, the DB spot-check, and the migration up/down round-trip result.

- [ ] **Step 5: Update CHECKPOINT**

In `docs/superpowers/CHECKPOINT.md`, change the Roles & Permissions section status from "AUDIT DONE" to "DONE (merged)" once merged, note migration **000020** applied, link the verification log, and update "Latest taken migration" to **000020** / next **000021**. Mark deferred **#15** as RESOLVED.

- [ ] **Step 6: Commit**

```bash
git add docs/swagger docs/superpowers/verification/roles-permissions-parity.md docs/superpowers/CHECKPOINT.md
git commit -m "docs(roles): swagger regen + verification log + checkpoint close"
```

---

## Self-Review checklist (done while writing)

- **Spec coverage:** list (D4/D5) → Task 6 + 4; get → 4/6; create → 4/6; update (partial + system guards D9) → 4/6; delete (soft D2, blocked-if-assigned, system guard) → 4/6; level + authority (D1) → 1/5; sort by level (D3) → 3/4; catalog gate (D6) → 6; name regex + perm validation (D8) → 4. All locked decisions mapped. ✓
- **Out-of-scope respected:** registry `approve_team/all` delta untouched. ✓
- **Type consistency:** `RoleRead`/`RoleCreate`/`RoleUpdate`/`RoleListQuery`, `RoleFilter`, `NewRoleService`, `NewRoleHandler(svc)`, `NewUserService(users, roles, ...)` used consistently across tasks. ✓
- **Known risks (verified during planning):** (a) `apperrors.ErrForbidden(msg)` + `apperrors.CodeForbidden` — CONFIRMED present (errors.go:90,14); (b) `NewUserService` arity change ripples to TWO test call sites — CONFIRMED (`user_service_test.go:19`, `invite_service_test.go:54`) — both handled in Task 5; (c) the existing `TestUserService_AssignRoles` role-less-admin breaks under the authority check — CONFIRMED and rewritten in Task 5 Step 4; (d) Gin static-vs-wildcard route ordering — register `/permissions` + `""` before `/:id`; (e) `GORM default:100` means `makeRole` yields level-100 roles — relied on intentionally in tests.

---

## Execution Handoff

This plan touches shared code (`UserService` constructor, route table, seed) so the medium/complex path applies: implement on an isolated worktree branch, verify end-to-end, then request review before merge.
