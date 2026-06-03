# Employee API Parity Round 2 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the Go employee/user API to parity with the Python source + current BA docs — working multi-select list filters, `experience_year` as a career-start year, inline skill assignment on create/update, resolved department/position names on reads, and a `full_name` → `first_name`/`last_name` split.

**Architecture:** Layered Go/Gin/GORM (`handler → service → repository → GORM`). Six changes land as sequenced commits ordered so `make build && make vet && make test` stays green at every commit: the four schema-independent changes first (filters, experience validation, skills, dept/position), then the atomic name split (drops the `full_name` column), then the experience data migration, then docs/verification.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, golang-migrate (versioned SQL), testify + Postgres test DB.

---

## Pre-flight notes (read once)

- **Branch:** work happens on `feat/employees-parity-2` (already checked out; the spec is committed there).
- **Migration version assert is dynamic.** `internal/config/db.go:AssertMigrationsUpToDate` compares the applied `schema_migrations.version` against the highest `NNNNNN` file on disk — there is **no hardcoded version constant to bump**. Creating a migration pair and running `make migrate-up` is sufficient.
- **`make test` needs the Postgres test DB.** Run `make test-db-up` once. Per CHECKPOINT, the local DB is the docker container `ennam-ecom-postgres` (user `ennam`/`ennam_dev_2026`); from the host, migrations/tests may need `DB_HOST=localhost` or a `DATABASE_URL` override. Port 8080 is often busy → use `PORT=8082` for live smoke.
- **`make swag` regenerates Swagger** — never hand-edit `docs/swagger/`.
- **Spec:** [docs/superpowers/specs/2026-06-01-employee-api-parity-2-design.md](docs/superpowers/specs/2026-06-01-employee-api-parity-2-design.md).
- **`models.Employee.FullName` is removed in Task 5 and replaced by a `FullName()` *method*** that composes `FirstName + " " + LastName`. Tasks 1–4 do **not** touch names, so they compile/test against the current `FullName` string field. Only Task 5 flips it.
- **Other models keep their own `FullName` columns** (`Dependent`, `Invite`, `EmployeeEmergencyContact`) — those are separate tables and are NOT changed.

---

## File Structure

**Created:**
- `migrations/000018_employee_name_split.up.sql` / `.down.sql` — drop `full_name`, add `first_name`/`last_name`.
- `migrations/000019_employee_experience_year_to_start_year.up.sql` / `.down.sql` — count→year normalization.

**Modified:**
- `internal/dto/employee.go` — `EmployeeListQuery` (multi-select), `EmployeeCreate`/`EmployeeUpdate`/`EmployeeSelfUpdate`/`EmployeeRead` (first/last + `skill_ids`).
- `internal/dto/auth.go` — `EmployeeSummary` (first/last).
- `internal/models/employee.go` — `FirstName`/`LastName` fields + `FullName()` method.
- `internal/repositories/employee_repo.go` — `List` (IN filters, preload dept/pos, order/search), `FindByIDWithFull`/`FindByUserIDWithFull` (preload dept/pos), `ListDirectReports`/`ListManagerCandidates` (order/search).
- `internal/services/employee_service.go` — `toRead` (dept/pos refs, first/last), experience validation, skill wiring, `Create`/`Update`/`SelfUpdate`, `toSummary`, `toManagerCandidate`/`toDirectReport`, `NewEmployeeService`.
- `internal/services/skill_service.go` — extract `ValidateSkillIDs`.
- `internal/handlers/employee_handler.go` — `List` (parse filters), Swagger params.
- `internal/handlers/auth_handler.go` — `toUserSummary` (first/last).
- `internal/services/{attendance_service,attendance_matrix,leave_service,announcement_service,organization_settings_service,invite_service}.go` — `e.FullName` → `e.FullName()` (Employee reads only).
- `internal/services/seed_service.go` — admin first/last.
- `cmd/server/main.go` — reorder skillSvc before empSvc; inject into `NewEmployeeService`.
- Test fixtures (Task 5): `employee_service_test.go`, `employee_parity_test.go`, `employee_line_manager_test.go`, `auth_service_test.go`, `seed_service_test.go`, `user_service_test.go`, `testhelper_test.go`, `attendance_service_test.go`, `push_notification_service_test.go`, `dependent_service_test.go`.
- `scripts/smoke-employees-parity.sh` — first/last + experience-as-year (Task 7).

---

## Task 1: Multi-select list filters

Make `department_id` / `position_id` / `manager_id` / `role_id` accept repeated query params (`?department_id=A&department_id=B`) → SQL `IN`. Fixes the current `400`-on-any-value bug (gin can't bind `*uuid.UUID` from a query).

**Files:**
- Modify: `internal/dto/employee.go:283-292` (EmployeeListQuery)
- Modify: `internal/repositories/employee_repo.go:175-236` (List)
- Modify: `internal/handlers/employee_handler.go:130-168` (List + Swagger)
- Test: `internal/services/employee_service_test.go`

- [ ] **Step 1: Write the failing test** — append to `internal/services/employee_service_test.go`:

```go
func TestEmployeeService_List_MultiSelectDepartmentFilter(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	// Two real departments to satisfy the employees.department_id FK.
	dA, dB := uuid.New(), uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dA, "Alpha").Error)
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dB, "Beta").Error)

	mk := func(email, name string, dept uuid.UUID) {
		v, err := svc.Create(ctx, dto.EmployeeCreate{Email: email, Password: "Pass12345", FullName: name})
		require.NoError(t, err)
		require.NoError(t, testDB.Exec("UPDATE employees SET department_id = ? WHERE id = ?", dept, v.ID).Error)
	}
	mk("a@x.com", "Ann", dA)
	mk("b@x.com", "Bob", dB)
	mk("c@x.com", "Cara", uuid.Nil) // unassigned department -> excluded by the filter

	// Filter by BOTH departments -> OR within the filter -> 2 rows.
	items, total, err := svc.List(ctx, dto.EmployeeListQuery{
		Page: 1, PageSize: 20, DepartmentIDs: []uuid.UUID{dA, dB},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
}
```

- [ ] **Step 2: Run it to confirm it fails to compile** — `DepartmentIDs` does not exist yet.

Run: `make test 2>&1 | head -30` (or `go test ./internal/services/ -run TestEmployeeService_List_MultiSelectDepartmentFilter`)
Expected: FAIL — `unknown field 'DepartmentIDs' in struct literal of type dto.EmployeeListQuery`.

- [ ] **Step 3: Rewrite `EmployeeListQuery`** in `internal/dto/employee.go` (replace lines 283-292):

```go
type EmployeeListQuery struct {
	Page     int    `form:"page,default=1"       binding:"gte=1"`
	PageSize int    `form:"page_size,default=20" binding:"gte=1,lte=100"`
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`

	// Raw repeated query params (HTTP boundary). gin binds repeated keys
	// (?department_id=a&department_id=b) into these string slices; uuid.UUID
	// cannot be bound directly from a query param (that was the 400 bug).
	DepartmentIDsRaw []string `form:"department_id"`
	PositionIDsRaw   []string `form:"position_id"`
	ManagerIDsRaw    []string `form:"manager_id"`
	RoleIDsRaw       []string `form:"role_id"`

	// Parsed by the handler (ParseFilters); consumed by the repo. Not bound.
	DepartmentIDs []uuid.UUID `form:"-"`
	PositionIDs   []uuid.UUID `form:"-"`
	ManagerIDs    []uuid.UUID `form:"-"`
	RoleIDs       []uuid.UUID `form:"-"`
}

// ParseFilters converts the raw repeated-param strings into parsed UUID
// slices, returning a 400 AppError on the first invalid value. Empty/absent
// params yield empty slices (= no filter).
func (q *EmployeeListQuery) ParseFilters() error {
	parse := func(name string, raw []string, dst *[]uuid.UUID) error {
		for _, s := range raw {
			if s == "" {
				continue
			}
			id, err := uuid.Parse(s)
			if err != nil {
				return apperrors.ErrBadRequest("invalid " + name)
			}
			*dst = append(*dst, id)
		}
		return nil
	}
	if err := parse("department_id", q.DepartmentIDsRaw, &q.DepartmentIDs); err != nil {
		return err
	}
	if err := parse("position_id", q.PositionIDsRaw, &q.PositionIDs); err != nil {
		return err
	}
	if err := parse("manager_id", q.ManagerIDsRaw, &q.ManagerIDs); err != nil {
		return err
	}
	if err := parse("role_id", q.RoleIDsRaw, &q.RoleIDs); err != nil {
		return err
	}
	return nil
}
```

Add the apperrors import to `internal/dto/employee.go` (it currently imports only `time` + `uuid`). Update the import block:

```go
import (
	"time"

	"github.com/google/uuid"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
)
```

- [ ] **Step 4: Update the repo filter logic** in `internal/repositories/employee_repo.go` — replace the four single-value `if q.XID != nil` blocks (lines 197-214) with `IN`-based slice filters. Replace:

```go
	if q.DepartmentID != nil {
		tx = tx.Where("employees.department_id = ?", *q.DepartmentID)
	}
	if q.PositionID != nil {
		tx = tx.Where("employees.position_id = ?", *q.PositionID)
	}
	if q.ManagerID != nil {
		tx = tx.Where("employees.manager_id = ?", *q.ManagerID)
	}
	if q.IsActive != nil {
		tx = tx.Where("users.is_active = ?", *q.IsActive)
	}
	if q.RoleID != nil {
		tx = tx.Where(
			"EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = employees.user_id AND ur.role_id = ? AND ur.is_deleted = false)",
			*q.RoleID,
		)
	}
```

with:

```go
	if len(q.DepartmentIDs) > 0 {
		tx = tx.Where("employees.department_id IN ?", q.DepartmentIDs)
	}
	if len(q.PositionIDs) > 0 {
		tx = tx.Where("employees.position_id IN ?", q.PositionIDs)
	}
	if len(q.ManagerIDs) > 0 {
		tx = tx.Where("employees.manager_id IN ?", q.ManagerIDs)
	}
	if q.IsActive != nil {
		tx = tx.Where("users.is_active = ?", *q.IsActive)
	}
	if len(q.RoleIDs) > 0 {
		tx = tx.Where(
			"EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = employees.user_id AND ur.role_id IN ? AND ur.is_deleted = false)",
			q.RoleIDs,
		)
	}
```

- [ ] **Step 5: Parse filters in the handler** — in `internal/handlers/employee_handler.go`, inside `List`, after the `ShouldBindQuery` block (line 150) add the parse call. Replace lines 145-151:

```go
func (h *EmployeeHandler) List(c *gin.Context) {
	var q dto.EmployeeListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
```

with:

```go
func (h *EmployeeHandler) List(c *gin.Context) {
	var q dto.EmployeeListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	if err := q.ParseFilters(); err != nil {
		_ = c.Error(err)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
```

- [ ] **Step 6: Update the Swagger annotations** for `List` (lines 138-142) to reflect multi-value params:

```go
// @Param        department_id query []string false "department uuid(s) — repeatable, OR within filter"
// @Param        position_id   query []string false "position uuid(s) — repeatable"
// @Param        manager_id    query []string false "manager uuid(s) — repeatable"
// @Param        role_id       query []string false "role uuid(s) — repeatable"
```

- [ ] **Step 7: Run the test to verify it passes**

Run: `go test ./internal/services/ -run TestEmployeeService_List_MultiSelectDepartmentFilter -v`
Expected: PASS

- [ ] **Step 8: Full build + vet + test**

Run: `make build && make vet && make test`
Expected: all green.

- [ ] **Step 9: Commit**

```bash
git add internal/dto/employee.go internal/repositories/employee_repo.go internal/handlers/employee_handler.go internal/services/employee_service_test.go
git commit -m "feat(employees): multi-select list filters (dept/position/role/manager via IN)

Bind repeated query params as []string -> uuid.Parse -> SQL IN, fixing
the 400-on-any-value bug (gin cannot bind uuid.UUID from a query). Status
stays the single is_active bool. Follows BA DR-001-005-01.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Experience-year validation (career-start year)

Reject `experience_year` values that are not a sane 4-digit calendar year `≤` the current year.

**Files:**
- Modify: `internal/services/employee_service.go` (imports + new helper + Create + Update)
- Test: `internal/services/employee_service_test.go`

- [ ] **Step 1: Write the failing test** — append to `internal/services/employee_service_test.go`:

```go
func TestEmployeeService_Create_RejectsBadExperienceYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	future := time.Now().UTC().Year() + 1
	_, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "future@x.com", Password: "Pass12345", FullName: "Future Year",
		ExperienceYear: &future,
	})
	require.Error(t, err, "a future experience_year must be rejected")

	old := 1800
	_, err = svc.Create(ctx, dto.EmployeeCreate{
		Email: "old@x.com", Password: "Pass12345", FullName: "Too Old",
		ExperienceYear: &old,
	})
	require.Error(t, err, "experience_year <= 1900 must be rejected")

	good := 2018
	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "good@x.com", Password: "Pass12345", FullName: "Good Year",
		ExperienceYear: &good,
	})
	require.NoError(t, err)
	require.NotNil(t, v.ExperienceYear)
	assert.Equal(t, 2018, *v.ExperienceYear)
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/services/ -run TestEmployeeService_Create_RejectsBadExperienceYear -v`
Expected: FAIL (the future/old creates currently succeed).

- [ ] **Step 3: Add the validator + imports** in `internal/services/employee_service.go`. Add `"fmt"` and `"time"` to the import block (currently has `context, errors, log, net/http, sort, strings`). Then add the helper near the other package helpers (e.g. after `boolToRenewal`):

```go
// validateExperienceYear enforces the BA contract (DR-001-005-02/03/04):
// experience_year is a career-start (4-digit) year, must be > 1900 and not in
// the future. nil = not provided = valid.
func validateExperienceYear(y *int) error {
	if y == nil {
		return nil
	}
	cur := time.Now().UTC().Year()
	if *y <= 1900 || *y > cur {
		return apperrors.ErrBadRequest(fmt.Sprintf("experience_year must be a year between 1901 and %d", cur))
	}
	return nil
}
```

- [ ] **Step 4: Call it in `Create`** — at the top of `Create` (after the email-exists check, before hashing the password, around line 353), add:

```go
	if err := validateExperienceYear(in.ExperienceYear); err != nil {
		return nil, err
	}
```

- [ ] **Step 5: Call it in `Update`** — at the top of `Update` (right after the `FindByIDWithFull` + not-found handling, before the self-deactivate guard, around line 505), add:

```go
	if err := validateExperienceYear(in.ExperienceYear); err != nil {
		return nil, err
	}
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `go test ./internal/services/ -run TestEmployeeService_Create_RejectsBadExperienceYear -v`
Expected: PASS

- [ ] **Step 7: Build + vet + test, then commit**

```bash
make build && make vet && make test
git add internal/services/employee_service.go internal/services/employee_service_test.go
git commit -m "feat(employees): validate experience_year as a career-start year (>1900, <= current year)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Inline skill assignment on create/update

Accept `skill_ids` in the employee create/update payload (Python parity), reusing the existing replace-set logic. Keep the standalone `PUT /employees/:id/skills`.

**Files:**
- Modify: `internal/services/skill_service.go` (extract `ValidateSkillIDs`)
- Modify: `internal/dto/employee.go` (EmployeeCreate + EmployeeUpdate)
- Modify: `internal/services/employee_service.go` (interface, struct, constructor, Create, Update)
- Modify: `cmd/server/main.go` (reorder skillSvc; inject)
- Test: `internal/services/employee_parity_test.go`

- [ ] **Step 1: Write the failing test** — append to `internal/services/employee_parity_test.go`:

```go
// Inline skill_ids on create + update (Python parity).
func TestEmployeeParity_InlineSkillAssignment(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, deps := newEmpSvc(testDB)
	ctx := context.Background()

	// Two skills in the catalog.
	s1, s2 := uuid.New(), uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO skills (id, name, description) VALUES (?, ?, '')", s1, "Go").Error)
	require.NoError(t, testDB.Exec("INSERT INTO skills (id, name, description) VALUES (?, ?, '')", s2, "SQL").Error)

	// Create with skills inline.
	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "skilled@x.com", Password: "Pass12345", FullName: "Skilled One",
		SkillIDs: []uuid.UUID{s1, s2},
	})
	require.NoError(t, err)
	require.Len(t, v.Skills, 2, "both skills must be assigned on create")

	// Update -> replace down to one.
	one := []uuid.UUID{s1}
	v2, err := svc.Update(ctx, v.ID, dto.EmployeeUpdate{SkillIDs: &one}, deps.adminUserID)
	require.NoError(t, err)
	require.Len(t, v2.Skills, 1)
	assert.Equal(t, "Go", v2.Skills[0].Name)

	// Update with empty slice -> clear all.
	empty := []uuid.UUID{}
	v3, err := svc.Update(ctx, v.ID, dto.EmployeeUpdate{SkillIDs: &empty}, deps.adminUserID)
	require.NoError(t, err)
	require.Len(t, v3.Skills, 0)

	// Invalid skill on create -> 400, no user created.
	_, err = svc.Create(ctx, dto.EmployeeCreate{
		Email: "bad@x.com", Password: "Pass12345", FullName: "Bad Skill",
		SkillIDs: []uuid.UUID{uuid.New()},
	})
	require.Error(t, err)
}
```

> NOTE: `newEmpSvc` must now return a SkillService-backed assigner and an `adminUserID`. Step 2 adjusts the test helper.

- [ ] **Step 2: Adjust the test helper `newEmpSvc`** in `internal/services/testhelper_test.go` so the EmployeeService is built with the skill assigner. Find the current `newEmpSvc` constructor and update it to construct a real `SkillService` and pass it into `NewEmployeeService` (matching the new 8-arg signature from Step 5), and to expose an `adminUserID` field on whatever struct it returns. If `newEmpSvc` returns `(*EmployeeService, <deps>)`, add a `SkillService` to the deps and an `adminUserID` (any non-nil uuid is fine for the `callerUserID` arg in Update — it only matters for the self-deactivate guard, which these tests don't trigger). Concretely, wherever it calls `NewEmployeeService(testDB, empRepo, depRepo, userRepo, roleRepo, quotaRepo, uploadSvc)`, change to:

```go
	skillSvc := NewSkillService(NewSkillRepository(testDB), NewEmployeeSkillRepository(testDB), empRepo, uploadSvc)
	svc := NewEmployeeService(testDB, empRepo, depRepo, userRepo, roleRepo, quotaRepo, uploadSvc, skillSvc)
```

and ensure the returned deps struct carries `adminUserID uuid.UUID` (set it to `uuid.New()`).

> If the existing `newEmpSvc` signature/return differs, adapt minimally — the only hard requirements are (a) pass a real `*SkillService` as the 8th arg and (b) expose a `uuid.UUID` the test can pass as `callerUserID`.

- [ ] **Step 3: Extract `ValidateSkillIDs`** in `internal/services/skill_service.go`. Add a public method and refactor `ReplaceForEmployee` to use it. Insert before `ReplaceForEmployee`:

```go
// ValidateSkillIDs de-duplicates the requested set and verifies every id
// references a live skill. Returns the cleaned slice or a 400 with the
// offending id — used both by ReplaceForEmployee and by inline assignment on
// employee create/update.
func (s *SkillService) ValidateSkillIDs(ctx context.Context, skillIDs []uuid.UUID) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{}, len(skillIDs))
	cleaned := make([]uuid.UUID, 0, len(skillIDs))
	for _, sid := range skillIDs {
		if sid == uuid.Nil {
			return nil, apperrors.ErrBadRequest("skill_ids contains an empty UUID")
		}
		if _, dup := seen[sid]; dup {
			continue
		}
		seen[sid] = struct{}{}
		if _, err := s.repo.FindByID(ctx, sid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.ErrBadRequest(fmt.Sprintf("Skill %s not found", sid))
			}
			return nil, err
		}
		cleaned = append(cleaned, sid)
	}
	return cleaned, nil
}
```

Then replace the inline validation loop inside `ReplaceForEmployee` (the `seen`/`cleaned` block) with:

```go
	cleaned, err := s.ValidateSkillIDs(ctx, skillIDs)
	if err != nil {
		return nil, err
	}
```

(keep the subsequent `s.empSkills.ReplaceForEmployee(ctx, employeeID, cleaned)` + `s.ListForEmployee(...)` calls).

- [ ] **Step 4: Add `skill_ids` to the DTOs** in `internal/dto/employee.go`. In `EmployeeCreate` (after `RoleIDs`, line 201):

```go
	// Skills assigned at creation (inline — Python parity). Empty/absent = none.
	SkillIDs []uuid.UUID `json:"skill_ids,omitempty"`
```

In `EmployeeUpdate` (after `IsActive`, line 247) — pointer-to-slice PATCH semantics:

```go
	// Skills replace-set: nil/absent = leave unchanged, [] = clear all,
	// non-empty = replace the whole set (inline — Python parity).
	SkillIDs *[]uuid.UUID `json:"skill_ids,omitempty"`
```

- [ ] **Step 5: Wire the assigner into EmployeeService** in `internal/services/employee_service.go`. Add the interface + struct field + constructor param. Replace the struct + constructor (lines 147-167):

```go
// skillAssigner is the slice of SkillService that EmployeeService needs to
// apply inline skill_ids on create/update. Kept narrow for testability;
// satisfied by *SkillService.
type skillAssigner interface {
	ValidateSkillIDs(ctx context.Context, skillIDs []uuid.UUID) ([]uuid.UUID, error)
	ReplaceForEmployee(ctx context.Context, employeeID uuid.UUID, skillIDs []uuid.UUID) ([]dto.SkillRead, error)
}

type EmployeeService struct {
	db      *gorm.DB
	emps    repositories.EmployeeRepository
	deps    repositories.DependentRepository
	users   repositories.UserRepository
	roles   repositories.RoleRepository
	quota   repositories.LeaveQuotaRepository
	uploads Uploader
	skills  skillAssigner
}

func NewEmployeeService(
	db *gorm.DB,
	emps repositories.EmployeeRepository,
	deps repositories.DependentRepository,
	users repositories.UserRepository,
	roles repositories.RoleRepository,
	quota repositories.LeaveQuotaRepository,
	uploads Uploader,
	skills skillAssigner,
) *EmployeeService {
	return &EmployeeService{db: db, emps: emps, deps: deps, users: users, roles: roles, quota: quota, uploads: uploads, skills: skills}
}
```

- [ ] **Step 6: Apply skills in `Create`** — pre-validate before the tx and apply after commit. In `Create`, after the `validateExperienceYear` call from Task 2 (and before hashing), add the pre-validation:

```go
	if len(in.SkillIDs) > 0 {
		if _, err := s.skills.ValidateSkillIDs(ctx, in.SkillIDs); err != nil {
			return nil, err
		}
	}
```

Then after the emergency-contacts block near the end of `Create` (before `return s.Get(ctx, createdEmp.ID)`), add:

```go
	if len(in.SkillIDs) > 0 {
		if _, err := s.skills.ReplaceForEmployee(ctx, createdEmp.ID, in.SkillIDs); err != nil {
			return nil, err
		}
	}
```

- [ ] **Step 7: Apply skills in `Update`** — after the emergency-contacts replace block (before `return s.Get(ctx, id)`), add:

```go
	if in.SkillIDs != nil {
		if _, err := s.skills.ReplaceForEmployee(ctx, e.ID, *in.SkillIDs); err != nil {
			return nil, err
		}
	}
```

- [ ] **Step 8: Reorder + inject in `cmd/server/main.go`.** Move the `skillSvc :=` line (currently 93) to *before* `empSvc :=` (currently 88), and add `skillSvc` as the 8th arg. The block (lines 84-93) becomes:

```go
	uploadSvc, err := services.NewUploadService(context.Background(), cfg.Storage)
	if err != nil {
		log.Fatalf("upload service: %v", err)
	}
	skillSvc := services.NewSkillService(skillRepo, employeeSkillRepo, employeeRepo, uploadSvc)
	empSvc := services.NewEmployeeService(db, employeeRepo, dependentRepo, userRepo, roleRepo, quotaRepo, uploadSvc, skillSvc)
	depSvc := services.NewDependentService(dependentRepo, employeeRepo)
	userSvc := services.NewUserService(userRepo, employeeRepo, tokenRepo, settingsRepo, empSvc)
	departmentSvc := services.NewDepartmentService(departmentRepo)
	positionSvc := services.NewPositionService(positionRepo)
```

(Delete the old `skillSvc := ...` line further down so it is not declared twice.)

- [ ] **Step 9: Run the test to verify it passes**

Run: `go test ./internal/services/ -run TestEmployeeParity_InlineSkillAssignment -v`
Expected: PASS

- [ ] **Step 10: Build + vet + full test, then commit**

```bash
make build && make vet && make test
git add internal/services/skill_service.go internal/dto/employee.go internal/services/employee_service.go internal/services/testhelper_test.go internal/services/employee_parity_test.go cmd/server/main.go
git commit -m "feat(employees): inline skill_ids on create/update (Python parity)

Extract SkillService.ValidateSkillIDs; inject SkillService into
EmployeeService as a narrow skillAssigner; apply replace-set on create
(pre-validated before the user tx) and update (nil=leave, []=clear).
Keeps the standalone PUT /employees/:id/skills.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Resolve department/position names on the employee read

Populate `EmployeeRead.department` / `EmployeeRead.position` (currently always `null`).

**Files:**
- Modify: `internal/repositories/employee_repo.go` (List + FindByIDWithFull + FindByUserIDWithFull preloads)
- Modify: `internal/services/employee_service.go` (toRead + helpers)
- Test: `internal/services/employee_service_test.go`

- [ ] **Step 1: Write the failing test** — append to `internal/services/employee_service_test.go`:

```go
func TestEmployeeService_Read_ResolvesDepartmentAndPosition(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	dID, pID := uuid.New(), uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", dID, "Engineering").Error)
	require.NoError(t, testDB.Exec("INSERT INTO positions (id, name) VALUES (?, ?)", pID, "Senior Engineer").Error)

	v, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "wp@x.com", Password: "Pass12345", FullName: "Work Profile",
		DepartmentID: &dID, PositionID: &pID,
	})
	require.NoError(t, err)

	got, err := svc.Get(ctx, v.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Department, "department ref must be resolved on read")
	assert.Equal(t, "Engineering", got.Department.Name)
	require.NotNil(t, got.Position, "position ref must be resolved on read")
	assert.Equal(t, "Senior Engineer", got.Position.Name)
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/services/ -run TestEmployeeService_Read_ResolvesDepartmentAndPosition -v`
Expected: FAIL — `got.Department` is nil.

- [ ] **Step 3: Add preloads in the repo.** In `internal/repositories/employee_repo.go`:
  - In `FindByIDWithFull` (after the `Preload("Manager.Position", notDeleted)` line ~139) add:

```go
		Preload("Department", notDeleted).
		Preload("Position", notDeleted).
```

  - In `FindByUserIDWithFull` (same spot, ~161) add the identical two preloads.
  - In `List` (after `Preload("Manager.Position", notDeleted)` ~182) add the identical two preloads.

- [ ] **Step 4: Add ref helpers + populate in `toRead`.** In `internal/services/employee_service.go`, add helpers next to `positionName`/`departmentName`:

```go
func departmentRef(e *models.Employee) *dto.RefRead {
	if e != nil && e.Department != nil && e.Department.Name != "" {
		return &dto.RefRead{ID: e.Department.ID, Name: e.Department.Name}
	}
	return nil
}

func positionRef(e *models.Employee) *dto.RefRead {
	if e != nil && e.Position != nil && e.Position.Name != "" {
		return &dto.RefRead{ID: e.Position.ID, Name: e.Position.Name}
	}
	return nil
}
```

In `toRead`, set the two fields on the `out` struct (replace the trailing `// Department/Position refs preloaded in Phase 3; intentionally nil until then.` comment / `return out`):

```go
	out.Department = departmentRef(e)
	out.Position = positionRef(e)
	return out
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/services/ -run TestEmployeeService_Read_ResolvesDepartmentAndPosition -v`
Expected: PASS

- [ ] **Step 6: Build + vet + test, then commit**

```bash
make build && make vet && make test
git add internal/repositories/employee_repo.go internal/services/employee_service.go internal/services/employee_service_test.go
git commit -m "fix(employees): resolve own department/position names on the employee read

Preload Department+Position in List/FindByIDWithFull/FindByUserIDWithFull
and emit them as RefRead on EmployeeRead (closes the Phase-3 null gap for
the user-list columns and user-details Work Profile).

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Name split — `full_name` → `first_name` + `last_name`

The atomic schema change. Migration 000018 drops `full_name` and adds the two columns; the model, all read sites, the create/update/self-update/read/summary DTOs, the repo ordering/search, the seed, and every test fixture flip in one commit so build+tests stay green.

**Files:** migration pair + `internal/models/employee.go`, `internal/dto/employee.go`, `internal/dto/auth.go`, `internal/services/employee_service.go`, `internal/handlers/auth_handler.go`, `internal/services/{attendance_service,attendance_matrix,leave_service,announcement_service,organization_settings_service,invite_service,seed_service}.go`, `internal/repositories/employee_repo.go`, and the 10 test files listed in File Structure.

- [ ] **Step 1: Create the migration pair**

Run: `make migrate-new name=employee_name_split`
Expected: creates `migrations/000018_employee_name_split.up.sql` and `.down.sql` (empty).

- [ ] **Step 2: Write `migrations/000018_employee_name_split.up.sql`**

```sql
-- 000018_employee_name_split
-- Python parity: names are stored as separate first_name + last_name columns
-- (app/models/user.py:88-89). Drop the single full_name column and backfill
-- the split. Legacy single-token names land as first_name=token, last_name=''
-- (the DEFAULT '' tolerates them; the min=1 rule applies only to new writes).
ALTER TABLE employees ADD COLUMN first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE employees ADD COLUMN last_name  TEXT NOT NULL DEFAULT '';

UPDATE employees
SET first_name = split_part(full_name, ' ', 1),
    last_name  = btrim(substr(full_name, length(split_part(full_name, ' ', 1)) + 1));

ALTER TABLE employees DROP COLUMN full_name;
```

- [ ] **Step 3: Write `migrations/000018_employee_name_split.down.sql`**

```sql
-- Down for 000018_employee_name_split. Lossless: full_name = first || ' ' || last.
ALTER TABLE employees ADD COLUMN full_name TEXT NOT NULL DEFAULT '';

UPDATE employees SET full_name = btrim(first_name || ' ' || last_name);

ALTER TABLE employees DROP COLUMN first_name;
ALTER TABLE employees DROP COLUMN last_name;
```

- [ ] **Step 4: Apply the migration**

Run: `make migrate-up && make migrate-version`
Expected: version `18`. (If migrating from the host fails on `DB_HOST`, prefix with `DATABASE_URL=postgres://…@localhost:5432/exnodes_hrm?sslmode=disable`.)

- [ ] **Step 5: Swap the model field + add the `FullName()` method** in `internal/models/employee.go`. Replace the `FullName` field line (line 15):

```go
	FullName         string     `gorm:"type:text;not null" json:"full_name"`
```

with:

```go
	FirstName        string     `gorm:"type:text;not null;default:''" json:"first_name"`
	LastName         string     `gorm:"type:text;not null;default:''" json:"last_name"`
```

Add the import `"strings"` to the model file and a method after the `TableName` func:

```go
// FullName composes the display name from the split columns. Used by the
// briefs/summaries that expose a single full_name (line manager, direct
// report, candidate, attendance, announcement author, invite inviter).
func (e *Employee) FullName() string {
	return strings.TrimSpace(e.FirstName + " " + e.LastName)
}
```

- [ ] **Step 6: Update `EmployeeRead`** in `internal/dto/employee.go` — replace `FullName string` (line 89) with:

```go
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
```

- [ ] **Step 7: Update `EmployeeCreate`/`EmployeeUpdate`/`EmployeeSelfUpdate`** in `internal/dto/employee.go`:
  - `EmployeeCreate`: replace `FullName string ... binding:"required,min=1,max=200"` (line 162) with:

```go
	FirstName        string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName         string     `json:"last_name"  binding:"required,min=1,max=100"`
```

  - `EmployeeUpdate`: replace `FullName *string` (line 207) with:

```go
	FirstName        *string    `json:"first_name,omitempty" binding:"omitempty,min=1,max=100"`
	LastName         *string    `json:"last_name,omitempty"  binding:"omitempty,min=1,max=100"`
```

  - `EmployeeSelfUpdate`: replace `FullName *string ... binding:"omitempty,min=1,max=200"` (line 260) with:

```go
	FirstName        *string    `json:"first_name,omitempty" binding:"omitempty,min=1,max=100"`
	LastName         *string    `json:"last_name,omitempty"  binding:"omitempty,min=1,max=100"`
```

- [ ] **Step 8: Update `EmployeeSummary`** in `internal/dto/auth.go` (lines 23-24) — replace `FullName string` with:

```go
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
```

- [ ] **Step 9: Update `employee_service.go`**:
  - `toRead` (line 285): replace `FullName: e.FullName,` with `FirstName: e.FirstName,` and `LastName: e.LastName,`.
  - `toRead` ManagerBrief (line 222): `FullName: e.Manager.FullName,` → `FullName: e.Manager.FullName(),`.
  - `toSummary` (line 334): replace `FullName: e.FullName,` with `FirstName: e.FirstName,` and `LastName: e.LastName,`.
  - `Create` (line 398): replace `FullName: strings.TrimSpace(in.FullName),` with `FirstName: strings.TrimSpace(in.FirstName),` and `LastName: strings.TrimSpace(in.LastName),`.
  - `Update` (line 523): replace `setIfNotNilStr("full_name", in.FullName)` with `setIfNotNilStr("first_name", in.FirstName)` and `setIfNotNilStr("last_name", in.LastName)`.
  - `SelfUpdate` (lines 646-649): replace the `if in.FullName != nil { allowed["full_name"] = strings.TrimSpace(*in.FullName) }` block with:

```go
	if in.FirstName != nil {
		allowed["first_name"] = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		allowed["last_name"] = strings.TrimSpace(*in.LastName)
	}
```

  - `toManagerCandidate` (line 920): `FullName: e.FullName,` → `FullName: e.FullName(),`.
  - `toDirectReport` (line 934): `FullName: e.FullName,` → `FullName: e.FullName(),`.

- [ ] **Step 10: Update the other Employee-model `FullName` read sites** (`x.FullName` → `x.FullName()`; these are all reads of a `*models.Employee`):
  - `internal/handlers/auth_handler.go:31` — `FullName: u.Employee.FullName,` → split: replace with `FirstName: u.Employee.FirstName,` and `LastName: u.Employee.LastName,` (this builds `dto.EmployeeSummary`).
  - `internal/services/attendance_service.go:106` — `FullName: a.Employee.FullName,` → `FullName: a.Employee.FullName(),`.
  - `internal/services/attendance_matrix.go:341` — `EmployeeName: emp.FullName,` → `EmployeeName: emp.FullName(),`.
  - `internal/services/leave_service.go:212` — `Name: emp.FullName}` → `Name: emp.FullName()}`.
  - `internal/services/announcement_service.go:156` — `FullName: a.Author.FullName,` → `FullName: a.Author.FullName(),`.
  - `internal/services/announcement_service.go:184` — `FullName: tu.Employee.FullName,` → `FullName: tu.Employee.FullName(),`.
  - `internal/services/organization_settings_service.go:128` — `name := emp.FullName` → `name := emp.FullName()`.
  - `internal/services/invite_service.go:122` — `FullName: inv.Inviter.FullName,` → `FullName: inv.Inviter.FullName(),` (Inviter is a `*models.Employee`). **Do NOT touch** `inv.FullName` on lines 106/138/215 — that is the `Invite` model's own column.

> The `Dependent`, `Invite`, and `EmployeeEmergencyContact` `FullName` fields are separate columns and remain plain string fields — leave them.

- [ ] **Step 11: Update repo ordering + search** in `internal/repositories/employee_repo.go`:
  - `List` search (lines 192-195): replace the `full_name` term:

```go
		tx = tx.Where(
			"employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR employees.phone ILIKE ? OR employees.personal_email ILIKE ? OR users.email ILIKE ?",
			p, p, p, p, p,
		)
```

  - `List` order (line 229): `Order("employees.full_name ASC")` → `Order("employees.first_name ASC, employees.last_name ASC")`.
  - `ListManagerCandidates` search (line 319): replace `employees.full_name ILIKE ?` with `(employees.first_name ILIKE ? OR employees.last_name ILIKE ?)` and add the extra `p` arg (so the args become `p, p, p, p` for first, last, position, department).
  - `ListManagerCandidates` order (line 325): `Order("LOWER(employees.full_name) ASC")` → `Order("LOWER(employees.first_name) ASC, LOWER(employees.last_name) ASC")`.
  - `ListDirectReports` order (line 336): `Order("LOWER(full_name) ASC")` → `Order("LOWER(first_name) ASC, LOWER(last_name) ASC")`.

- [ ] **Step 12: Update the seed** in `internal/services/seed_service.go`. Replace the `adminName` logic (lines 298-301) and the employee construction (line 360). Change the `adminName` default block to derive first/last from `SuperAdminName`:

```go
	adminFirst, adminLast := "Super", "Admin"
	if s.cfg.SuperAdminName != "" {
		parts := strings.SplitN(strings.TrimSpace(s.cfg.SuperAdminName), " ", 2)
		adminFirst = parts[0]
		if len(parts) > 1 {
			adminLast = strings.TrimSpace(parts[1])
		} else {
			adminLast = ""
		}
	}
```

and in the `emp := &models.Employee{...}` literal replace `FullName: adminName,` with `FirstName: adminFirst,` and `LastName: adminLast,`. Ensure `"strings"` is imported in `seed_service.go` (add if missing).

- [ ] **Step 13: Update test fixtures.** Across the test files, apply this mechanical transform:
  - Every `dto.EmployeeCreate{... FullName: "X Y" ...}` → `FirstName: "X", LastName: "Y"`. For single-token names (e.g. `"Anne"`, `"A"`, `"On"`, `"Off"`, `"The Boss"`→two tokens already), use `FirstName: "<token>", LastName: "Test"` so the `min=1` last-name rule is satisfied.
  - Every read of `view.FullName` / `got.FullName` / `.User.Employee.FullName` / `out.FullName` on an `EmployeeRead` or `EmployeeSummary` → use `FirstName`/`LastName`.
  - Manager brief / direct report / candidate `.FullName` assertions stay (those DTOs keep `full_name`, now composed) — assert the composed value (e.g. `"The Boss"`).

  Specific must-fix assertions:
  - `employee_service_test.go:67,78` — `assert.Equal(t, "Alice Smith", view.FullName)` / `got.FullName` → create with `FirstName: "Alice", LastName: "Smith"` and assert `view.FirstName=="Alice"` && `view.LastName=="Smith"`.
  - `employee_service_test.go:236-271` (List search/active) — give each name a last name with no `"an"` substring, e.g. `FirstName:"Anne", LastName:"Test"`; the `Search:"an"` assertion (`total==3`) still holds via first-name match; the active-filter assertion `items[0].FullName` → `items[0].FirstName` (`"On"`).
  - `employee_parity_test.go:30,57,116,142,162` etc. — split each `FullName`; the experience test (line 142) keep as-is otherwise; the self-update identity test (line 162-174) assert `out.FirstName=="After"` && `out.LastName=="Name"` after sending `FirstName:&first, LastName:&last`.
  - `employee_line_manager_test.go` — split all `FullName` create fields; keep `got.Manager.FullName == "The Boss"` (create the boss with `FirstName:"The", LastName:"Boss"`).
  - `auth_service_test.go:49-50` — create the test user via the path that sets the name to `"Alice Tester"` as `FirstName:"Alice", LastName:"Tester"`; assert `result.User.Employee.FirstName=="Alice"` && `LastName=="Tester"`.
  - `seed_service_test.go:76-77` — assert `u.Employee.FirstName=="Super"` && `u.Employee.LastName=="Admin"` (default when `SUPER_ADMIN_NAME` unset).
  - `user_service_test.go:192` — `assert.*employee.FullName` → `FirstName`/`LastName`; split create fixtures.
  - `testhelper_test.go:181,199`, `attendance_service_test.go:444`, `push_notification_service_test.go:137`, `dependent_service_test.go` (33,38,41,50,52,81,85,113,116) — split each `dto.EmployeeCreate{FullName:...}` into first/last (these are setup-only, no name assertions). Keep `Dependent.FullName` usages untouched.

- [ ] **Step 14: Build + vet + full test**

Run: `make build && make vet && make test`
Expected: all green. If a test references `FullName` on an `EmployeeRead`/`EmployeeSummary`, it will fail to compile — fix per Step 13. If a create fails with `Key: 'EmployeeCreate.LastName' Error:Field validation for 'LastName' failed on the 'required' tag`, a single-token fixture is missing its last name — add `LastName: "Test"`.

- [ ] **Step 15: Migration up/down round-trip check**

Run:
```bash
make migrate-down && make migrate-version   # back to 17, full_name restored
make migrate-up   && make migrate-version   # forward to 18, split restored
```
Expected: down → `17`, up → `18`, no errors.

- [ ] **Step 16: Commit**

```bash
git add migrations/000018_employee_name_split.up.sql migrations/000018_employee_name_split.down.sql \
        internal/models/employee.go internal/dto/employee.go internal/dto/auth.go \
        internal/services/employee_service.go internal/handlers/auth_handler.go \
        internal/services/attendance_service.go internal/services/attendance_matrix.go \
        internal/services/leave_service.go internal/services/announcement_service.go \
        internal/services/organization_settings_service.go internal/services/invite_service.go \
        internal/services/seed_service.go internal/repositories/employee_repo.go \
        internal/services/employee_service_test.go internal/services/employee_parity_test.go \
        internal/services/employee_line_manager_test.go internal/services/auth_service_test.go \
        internal/services/seed_service_test.go internal/services/user_service_test.go \
        internal/services/testhelper_test.go internal/services/attendance_service_test.go \
        internal/services/push_notification_service_test.go internal/services/dependent_service_test.go
git commit -m "feat(employees)!: split full_name into first_name + last_name (Python parity)

Migration 000018 drops employees.full_name and adds first_name/last_name
(backfilled by split). Model exposes a FullName() method for the briefs
that still surface a composed name (line manager, direct report, candidate,
attendance, announcement author, invite inviter). EmployeeRead/Create/
Update/SelfUpdate and EmployeeSummary now carry first_name+last_name; list
search/order use the split columns.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Experience data migration (count → start year)

Normalize any existing `experience_year` values that are still stored as a count (e.g. `7`) into a calendar year.

**Files:** migration pair only.
- Test: round-trip via the Makefile (Step 4).

- [ ] **Step 1: Create the migration pair**

Run: `make migrate-new name=employee_experience_year_to_start_year`
Expected: creates `migrations/000019_employee_experience_year_to_start_year.up.sql` and `.down.sql`.

- [ ] **Step 2: Write `migrations/000019_employee_experience_year_to_start_year.up.sql`**

```sql
-- 000019_employee_experience_year_to_start_year
-- experience_year was stored as a COUNT of years (migration 000017 comment +
-- smoke fixture used 7). The BA + Python treat it as a 4-digit career-start
-- YEAR. Convert plausible counts (< 1900) to currentYear - count; leave any
-- value already year-shaped (>= 1900) untouched.
UPDATE employees
SET experience_year = EXTRACT(YEAR FROM CURRENT_DATE)::int - experience_year
WHERE experience_year IS NOT NULL
  AND experience_year < 1900;
```

- [ ] **Step 3: Write `migrations/000019_employee_experience_year_to_start_year.down.sql`**

```sql
-- Down for 000019. The reverse (year -> count) is arithmetically approximate:
-- it assumes the apply-date year equals the up-migration's year. Documented
-- per the data-loss-guard convention.
DO $$
BEGIN
    RAISE NOTICE 'Reverting experience_year (year -> count) is approximate: it uses the current year, which may differ from the up-migration apply year.';
END$$;

UPDATE employees
SET experience_year = EXTRACT(YEAR FROM CURRENT_DATE)::int - experience_year
WHERE experience_year IS NOT NULL
  AND experience_year >= 1900;
```

- [ ] **Step 4: Apply + round-trip**

Run:
```bash
make migrate-up   && make migrate-version   # -> 19
make migrate-down && make migrate-version   # -> 18
make migrate-up   && make migrate-version   # -> 19
```
Expected: `19`, `18`, `19`, no errors.

- [ ] **Step 5: Build + test (no Go changes, but confirm the suite still passes at version 19)**

Run: `make build && make test`
Expected: green.

- [ ] **Step 6: Commit**

```bash
git add migrations/000019_employee_experience_year_to_start_year.up.sql migrations/000019_employee_experience_year_to_start_year.down.sql
git commit -m "feat(employees): migrate experience_year counts to career-start years

Migration 000019 converts existing count values (< 1900) to currentYear -
count; values already year-shaped are untouched. Down is approximate
(RAISE NOTICE) per the data-loss-guard convention.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Swagger, smoke script, verification log, checkpoint

**Files:** `docs/swagger/*` (generated), `scripts/smoke-employees-parity.sh`, `docs/superpowers/verification/employees-parity-2.md` (new), `docs/superpowers/CHECKPOINT.md`, `.serena/memories/code_map.md`, `.serena/memories/conventions.md`.

- [ ] **Step 1: Regenerate Swagger**

Run: `make swag`
Expected: `docs/swagger/` updated (multi-value filter params + first_name/last_name + skill_ids bodies). Do not hand-edit.

- [ ] **Step 2: Update the smoke script** `scripts/smoke-employees-parity.sh`:
  - Line 63: change the create payload `\"full_name\":\"Smoke $SFX\"` to `\"first_name\":\"Smoke\",\"last_name\":\"$SFX\"`; change `\"experience_year\":7` to `\"experience_year\":2018`.
  - Line 67: change the assertion to `eq 2018 "$(jqv '.data.experience_year')" "#8 experience_year echoed as a year"`.
  - Lines 100-101: change the self-edit to `{\"first_name\":\"Renamed\",\"last_name\":\"Admin\"}` and assert `eq "Renamed" "$(jqv '.data.first_name')" "#7 self can edit first_name"`.
  - Add a new block exercising a multi-select filter and inline `skill_ids` (create a skill via `POST /skills`, then `POST /employees` with `"skill_ids":["<id>"]`, assert `.data.skills | length == 1`; then `GET /employees?department_id=<id>&department_id=<id2>` returns 200).

- [ ] **Step 3: Run the server + smoke script (live e2e)**

Run (in one shell):
```bash
PORT=8082 make run    # or: PORT=8082 go run ./cmd/server
```
In another shell:
```bash
BASE=http://localhost:8082/api/v1 bash scripts/smoke-employees-parity.sh
```
Expected: all smoke assertions pass. Capture the output for the verification log.

- [ ] **Step 4: DB spot-check** — confirm the split + experience year landed:

```bash
psql "$DATABASE_URL" -c "SELECT first_name, last_name, experience_year FROM employees ORDER BY first_name LIMIT 5;"
psql "$DATABASE_URL" -c "\d employees" | grep -E "first_name|last_name|full_name|experience_year"
```
Expected: `first_name`/`last_name` present, no `full_name`, `experience_year` holding 4-digit years.

- [ ] **Step 5: Write the verification log** `docs/superpowers/verification/employees-parity-2.md` documenting: build/vet/test output, migration up/down round-trip (18 & 19), the smoke run (with assertions), and the DB spot-check. Mirror the structure of `docs/superpowers/verification/employees-line-manager.md`.

- [ ] **Step 6: Update CHECKPOINT + serena memories**
  - `docs/superpowers/CHECKPOINT.md`: bump "DB migration version" to **19**, add an "Employees parity round 2 — DONE" subsection under post-migration parity work, and move the resolved items (list-filter uuid bug, dept/position resolution) out of "Outstanding micro-items".
  - `.serena/memories/code_map.md`: note migrations 000018 (name split) + 000019 (experience year), and that `models.Employee` exposes `FullName()` (method, not column).
  - `.serena/memories/conventions.md`: update the "gin can NOT bind uuid.UUID from a query" note to record the resolved multi-select filter pattern (repeated `[]string` param + `ParseFilters` + `IN`), and add a line that names are split (first/last columns; `full_name` is a derived method/JSON only on briefs).

- [ ] **Step 7: Commit**

```bash
git add docs/swagger scripts/smoke-employees-parity.sh docs/superpowers/verification/employees-parity-2.md docs/superpowers/CHECKPOINT.md .serena/memories/code_map.md .serena/memories/conventions.md
git commit -m "docs(employees): swagger regen, smoke script, verification log + checkpoint for parity round 2

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Self-Review

**1. Spec coverage** (each spec section → task):
- §3 Migrations 000018/000019 → Tasks 5 & 6. ✔
- §4 Models (first/last + `FullName()`) → Task 5 Step 5. ✔
- §5 DTOs (Create/Update/SelfUpdate/Read/Summary/ListQuery + `skill_ids`) → Tasks 1 (ListQuery), 3 (skill_ids), 5 (names). ✔
- §6 Repository (IN filters, dept/pos preloads, order/search) → Tasks 1, 4, 5. ✔
- §7 Service (toRead dept/pos, experience validation, skill wiring) → Tasks 2, 3, 4, 5. ✔
- §8 Handlers/routes/seed/swagger → Tasks 1 (List parse + swagger), 5 (seed), 7 (swagger). Routes unchanged (direct-reports endpoint stays) — correct, no task needed. ✔
- §9 Tests & verification → Tasks 1–6 (unit) + Task 7 (smoke/verification/checkpoint). ✔
- §11 Out of scope (uploads, FE, embed) → no tasks, correct. ✔

**2. Placeholder scan:** No "TBD"/"implement later". Test-fixture transform (Task 5 Step 13) lists each file + exact rule + every must-fix assertion — mechanical, not a placeholder. ✔

**3. Type consistency:**
- `skillAssigner` interface (Task 3 Step 5) declares `ValidateSkillIDs(ctx, []uuid.UUID) ([]uuid.UUID, error)` + `ReplaceForEmployee(ctx, uuid.UUID, []uuid.UUID) ([]dto.SkillRead, error)` — matches the `*SkillService` methods (existing `ReplaceForEmployee` + the `ValidateSkillIDs` added in Step 3). ✔
- `NewEmployeeService` gains an 8th param `skills skillAssigner`; `main.go` (Step 8) and `testhelper_test.go` (Step 2) both pass `skillSvc` 8th. ✔
- `EmployeeListQuery` parsed fields `DepartmentIDs []uuid.UUID` etc. are read by the repo (Task 1 Step 4) and set by `ParseFilters` (Step 3) / tests. ✔
- `models.Employee.FullName()` method is called as `x.FullName()` everywhere a `*models.Employee` display name is needed (Task 5 Steps 9–10); `EmployeeRead`/`EmployeeSummary`/Create/Update DTOs use `FirstName`/`LastName` fields, never `FullName`. ✔
- `departmentRef`/`positionRef` return `*dto.RefRead`, matching `EmployeeRead.Department`/`Position` (`*RefRead`). ✔

No issues found.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-06-01-employee-api-parity-2.md`. Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
