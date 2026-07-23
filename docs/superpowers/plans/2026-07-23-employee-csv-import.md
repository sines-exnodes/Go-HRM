# Employee CSV Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let admins bulk-create employees from a `.csv` upload via `POST /api/v1/employees/import`, reusing existing `EmployeeService.Create` validation, with per-row success/error reporting.

**Architecture:** Multipart upload → parse CSV with stdlib `encoding/csv` → resolve human-readable refs (department/position/role names, manager email) → call `EmployeeService.Create` **once per row** (no bulk SQL). Partial success: one bad row never rolls back good rows. No migration. Response is a structured report, not a list of full employee reads.

**Tech Stack:** Go stdlib `encoding/csv` + `mime/multipart` · existing Gin handler/service layering · `employees:create` permission · no new deps.

**BA note:** DR-001-005-02 explicitly lists bulk/CSV import as out of scope. This plan is an **owner-requested feature** with locked defaults below. Flag deviations for BA if product later writes a DR.

---

## Locked decisions (do not re-open mid-implement)

| # | Decision | Choice | Why |
|---|----------|--------|-----|
| D1 | Failure mode | **Partial success** — continue after row errors | HR imports always have dirty rows; all-or-nothing is hostile |
| D2 | Persistence | **Reuse `EmployeeService.Create` per row** | Keeps email uniqueness, role default, manager validation, salary defaults identical to single create |
| D3 | Permission | **`employees:create` only** — no new perm | Same capability as single create |
| D4 | Required columns | `email`, `first_name`, `last_name` | Matches API `EmployeeCreate` binding, not the stricter BA form |
| D5 | FK refs in CSV | **Names / emails**, not UUIDs | Humans edit CSVs; resolve via existing `FindByName` / `FindByEmail` |
| D6 | Manager in same file | **Must already exist** | v1 single-pass; import managers first (or leave blank) |
| D7 | `send_invite` | Form field, default **false** | Bulk invite emails are foot-guns; opt-in only |
| D8 | Limits | File **≤ 2 MiB**, body **≤ 500 data rows** | Protects request time; Create is not cheap |
| D9 | Salary/banking columns | Honoured only if caller has `users:salary_manage` / `users:banking_manage`; else **row error** if those columns non-empty | Mirrors create handler guards |
| D10 | Template | `GET /api/v1/employees/import/template` returns header + 1 example row | FE/docs need a canonical shape |
| D11 | Migration | **None** | Create path already sufficient |
| D12 | Update existing | **Not in v1** — duplicate email = row error | Import = create only |

### CSV column contract

Header row **required**. Column names case-insensitive; normalize to lower snake_case after trim. Unknown columns → whole-file 400 (fail loud; typos otherwise silent).

| Column | Required | Maps to | Notes |
|--------|----------|---------|-------|
| `email` | yes | `EmployeeCreate.Email` | lowercased; unique |
| `first_name` | yes | `FirstName` | |
| `last_name` | yes | `LastName` | |
| `phone` | no | `Phone` | |
| `personal_email` | no | `PersonalEmail` | `optemail` rules |
| `gender` | no | `Gender` | `male` / `female` / `other` |
| `dob` | no | `DOB` | `YYYY-MM-DD` only |
| `nationality` | no | `Nationality` | free text |
| `id_number` | no | `IDNumber` | |
| `permanent_address` | no | `PermanentAddress` | |
| `current_address` | no | `CurrentAddress` | |
| `education` | no | `Education` | enum as create |
| `marital_status` | no | `MaritalStatus` | enum as create |
| `experience_year` | no | `ExperienceYear` | career-start year int |
| `department` | no | resolve → `DepartmentID` | case-insensitive name match; missing = row error |
| `position` | no | resolve → `PositionID` | same |
| `manager_email` | no | resolve → `ManagerID` | must be existing active user with employee row |
| `join_date` | no | `JoinDate` | `YYYY-MM-DD` |
| `role` | no | resolve → `RoleIDs` (single) | empty → default Employee role via Create |
| `is_active` | no | `IsActive` | `true`/`false`/`1`/`0`; default true |
| `contract_type` | no | `ContractType` | |
| `basic_salary` | no | `BasicSalary` | needs salary_manage |
| `insurance_salary` | no | `InsuranceSalary` | needs salary_manage |
| `bank_account` | no | `BankAccount` | needs banking_manage |
| `bank_name` | no | `BankName` | needs banking_manage |
| `bank_holder_name` | no | `BankHolderName` | needs banking_manage |
| `payment_method` | no | `PaymentMethod` | needs banking_manage |
| `social_insurance_number` | no | `SocialInsuranceNumber` | |
| `tax_identification_number` | no | `TaxIdentificationNumber` | |

**Not supported in v1 (row-level or file-level):** avatar/CV/ID images, emergency contacts multi-row, multi-role, multi-skill, contracts. Use single-create API after import.

### API shape

```
POST /api/v1/employees/import
Authorization: Bearer …
Content-Type: multipart/form-data
  file: <csv bytes>          required
  send_invite: true|false    optional, default false

GET /api/v1/employees/import/template
Authorization: Bearer …
→ text/csv attachment
```

**Success HTTP always 200** when the file parses (even if every row failed). Whole-file problems (bad multipart, empty file, bad headers, over limit) → **400**. Auth/perm → **401/403**.

```json
{
  "success": true,
  "message": "Import finished: 3 created, 1 failed",
  "data": {
    "total_rows": 4,
    "created": 3,
    "failed": 1,
    "results": [
      {"row": 2, "ok": true,  "email": "a@ex.com", "employee_id": "…", "user_id": "…"},
      {"row": 3, "ok": false, "email": "b@ex.com", "error": "A user with this email already exists"},
      {"row": 4, "ok": true,  "email": "c@ex.com", "employee_id": "…", "user_id": "…"},
      {"row": 5, "ok": false, "email": "", "error": "email is required"}
    ]
  }
}
```

`row` is **1-based file line number** (header = line 1; first data row = 2).

---

## File map

| Path | Action | Responsibility |
|------|--------|----------------|
| `internal/dto/employee.go` | Modify | `EmployeeImportResult`, `EmployeeImportRowResult` |
| `internal/services/employee_import.go` | **Create** | Parse CSV, resolve refs, loop Create, invite fan-out |
| `internal/services/employee_import_test.go` | **Create** | Integration tests (test DB) |
| `internal/handlers/employee_handler.go` | Modify | `Import`, `ImportTemplate` |
| `cmd/server/main.go` | Modify | Register routes **before** `:id` wildcard |
| `docs/swagger/` | Regen | `make swag` |
| `docs/superpowers/verification/employee-csv-import.md` | **Create** | Live smoke log |
| `docs/superpowers/CHECKPOINT.md` | Modify | Session end |

No new model, repo, migration, or permission constant.

---

### Task 1: DTOs

**Files:**
- Modify: `internal/dto/employee.go` (append after `EmployeeCreate`)

- [ ] **Step 1: Add result types**

```go
// EmployeeImportRowResult is one data-row outcome from CSV import.
// Row is the 1-based line number in the uploaded file (header = 1).
type EmployeeImportRowResult struct {
	Row        int        `json:"row"`
	OK         bool       `json:"ok"`
	Email      string     `json:"email"`
	EmployeeID *uuid.UUID `json:"employee_id,omitempty"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// EmployeeImportResult is the POST /employees/import response data.
type EmployeeImportResult struct {
	TotalRows int                       `json:"total_rows"`
	Created   int                       `json:"created"`
	Failed    int                       `json:"failed"`
	Results   []EmployeeImportRowResult `json:"results"`
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/dto/employee.go
git commit -m "feat(employees): add CSV import result DTOs"
```

---

### Task 2: Import service — parse + resolve + loop Create

**Files:**
- Create: `internal/services/employee_import.go`
- Test: `internal/services/employee_import_test.go`

**Constants (file-local):**

```go
const (
	employeeImportMaxBytes = 2 * 1024 * 1024
	employeeImportMaxRows  = 500
)

// Canonical header set — any other header is a whole-file error.
var employeeImportAllowedHeaders = map[string]struct{}{
	"email": {}, "first_name": {}, "last_name": {}, "phone": {},
	"personal_email": {}, "gender": {}, "dob": {}, "nationality": {},
	"id_number": {}, "permanent_address": {}, "current_address": {},
	"education": {}, "marital_status": {}, "experience_year": {},
	"department": {}, "position": {}, "manager_email": {}, "join_date": {},
	"role": {}, "is_active": {}, "contract_type": {},
	"basic_salary": {}, "insurance_salary": {},
	"bank_account": {}, "bank_name": {}, "bank_holder_name": {}, "payment_method": {},
	"social_insurance_number": {}, "tax_identification_number": {},
}

var employeeImportRequiredHeaders = []string{"email", "first_name", "last_name"}
```

**Public API on `EmployeeService`:**

```go
// ImportCSV parses a CSV body and creates one employee per data row.
// perms gate salary/banking columns the same way the create handler does.
// sendInvite triggers PasswordResetService.RequestReset best-effort after each success.
func (s *EmployeeService) ImportCSV(
	ctx context.Context,
	file []byte,
	perms EmployeeFieldPerms, // existing type used by GuardSalaryWrite / ApplyEmployeeFieldVisibility
	sendInvite bool,
) (*dto.EmployeeImportResult, error)
```

If `EmployeeService` does not yet hold `PasswordResetService` / reset collaborator, **do not inject it**. Mirror the create handler: return created emails to the handler and let the handler call `h.reset.RequestReset` for each success when `sendInvite` is true. Prefer that (keeps service free of reset dependency):

```go
func (s *EmployeeService) ImportCSV(
	ctx context.Context,
	file []byte,
	perms EmployeeFieldPerms,
) (*dto.EmployeeImportResult, error)
```

Handler owns invite fan-out (same pattern as `Create`).

**Algorithm:**

1. If `len(file) == 0` or `> employeeImportMaxBytes` → `ErrBadRequest`.
2. `csv.NewReader(bytes.NewReader(file))`; `FieldsPerRecord = -1` (allow ragged? **No** — set equal to header length; missing trailing cells OK via `LazyQuotes=true`, empty string for missing).
3. Read header; normalize each: `strings.ToLower(strings.TrimSpace(h))`.
4. Validate: every header ∈ allowed; every required present; no duplicates → else 400 with message listing the problem.
5. For each subsequent record until EOF:
   - If data row count would exceed 500 → 400 before processing more (or after reading all? Prefer **abort whole file** with 400 if row count > 500 after a full read of headers+rows into memory — simpler: count while iterating, fail remaining rows with same error is worse; **whole-file 400 if >500 data rows** after pre-counting is fine if we first read all records into `[][]string`).
   - **Preferred parse strategy:** read all records first (`reader.ReadAll()`), then if `len(records) < 2` → 400 empty; if `len(records)-1 > 500` → 400; then process.
6. Per data row index `i` (file line = `i+2`):
   - Build map `col → cell` (trim cells).
   - Validate required non-empty; parse optionals; resolve department/position/role/manager.
   - If salary/banking cells non-empty and perms deny → row error, continue.
   - Build `dto.EmployeeCreate` (no password; no send_invite field).
   - `s.Create(ctx, in)` — on error, append `OK:false, Error: appErr.Message` (or `err.Error()`); on success append ids from returned `EmployeeRead`.
7. Aggregate counts; return result. **Never return a non-nil error for per-row failures.**

**Ref resolution helpers (private):**

```go
func (s *EmployeeService) resolveImportDepartment(ctx context.Context, name string) (*uuid.UUID, error)
// depts.FindByName; nil result → ErrBadRequest("unknown department: …")

func (s *EmployeeService) resolveImportPosition(ctx context.Context, name string) (*uuid.UUID, error)

func (s *EmployeeService) resolveImportRole(ctx context.Context, name string) (uuid.UUID, error)
// roles.FindByName — note: returns gorm.ErrRecordNotFound, not (nil,nil)

func (s *EmployeeService) resolveImportManager(ctx context.Context, email string) (*uuid.UUID, error)
// users.FindByEmail → emps.FindByUserID → ManagerID = employee.ID
// inactive user → ErrBadRequest
```

Check whether `EmployeeService` already has `depts` / `positions` / `roles` / `users` / `emps` collaborators. Open `NewEmployeeService` — inject only what is missing (prefer existing fields). If department repo is not on the service, add it to the struct + `NewEmployeeService` + `main.go` wiring. **Impact:** update all test constructors (`newEmpSvc` / similar in test helpers).

**Date parse:** `time.Parse("2006-01-02", s)` only. Reject RFC3339 with time component for CSV simplicity.

**Bool parse:** `true`/`false`/`1`/`0` case-insensitive; empty = default true for `is_active`.

**Float parse:** `strconv.ParseFloat` for salaries.

---

### Task 3: Failing tests first (TDD)

**Files:**
- Create: `internal/services/employee_import_test.go`

Use existing test DB harness (`testhelper_test.go`). Count PASS/SKIP — require `TEST_DATABASE_URL` or `make test-db-up`.

- [ ] **Step 1: Write tests (expect FAIL before service exists)**

Minimum cases:

1. `TestImportCSV_CreatesValidRows` — 2 good rows → `Created=2`, employees exist by email.
2. `TestImportCSV_DuplicateEmail_IsRowError` — seed one user; CSV has same email + one good → `Created=1, Failed=1`.
3. `TestImportCSV_MissingRequiredColumn_WholeFile400` — no `email` header.
4. `TestImportCSV_UnknownHeader_WholeFile400`.
5. `TestImportCSV_UnknownDepartment_RowError` — continues, other rows OK.
6. `TestImportCSV_ResolvesRoleAndManager` — role name `Manager`, manager_email of seeded employee.
7. `TestImportCSV_EmptyFile_400`.
8. `TestImportCSV_OverMaxRows_400` — optional if slow; can unit-test the limit with a synthetic 501-row buffer without asserting DB.
9. `TestImportCSV_SalaryWithoutPerm_RowError` — `basic_salary` set, perms without salary_manage.

Helper:

```go
func csvBytes(header string, rows ...string) []byte {
	var b strings.Builder
	b.WriteString(header)
	b.WriteByte('\n')
	for _, r := range rows {
		b.WriteString(r)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}
```

- [ ] **Step 2: Run tests — confirm fail**

```bash
export TEST_DATABASE_URL='postgres://postgres:devpassword@localhost:5432/exnodes_hrm_test?sslmode=disable'
go test ./internal/services -run 'TestImportCSV_' -count=1 -v
```

Expected: compile fail or FAIL (method missing).

- [ ] **Step 3: Implement `employee_import.go` until all green**

- [ ] **Step 4: Run full services suite**

```bash
go test ./internal/services -count=1 2>&1 | tee /tmp/imp-test.txt
grep -cE '^--- PASS' /tmp/imp-test.txt
grep -E '^--- (SKIP|FAIL)' /tmp/imp-test.txt
```

Expected: 0 FAIL; only legitimate skip is AWS opt-in if any.

- [ ] **Step 5: Commit**

```bash
git add internal/services/employee_import.go internal/services/employee_import_test.go
# plus any NewEmployeeService signature / main.go DI if deps added
git commit -m "feat(employees): CSV import service with per-row create"
```

---

### Task 4: Handler + routes + template

**Files:**
- Modify: `internal/handlers/employee_handler.go`
- Modify: `cmd/server/main.go` (register **before** `adminEmps.GET(":id", …)`)

- [ ] **Step 1: Handler methods**

```go
// Import godoc
// @Summary      Import employees from CSV
// @Tags         employees
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file         formData  file    true   "CSV file"
// @Param        send_invite  formData  bool    false  "send set-password email per created row"
// @Success      200  {object}  dto.Response[dto.EmployeeImportResult]
// @Failure      400  {object}  map[string]interface{}
// @Router       /api/v1/employees/import [post]
func (h *EmployeeHandler) Import(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("file is required"))
		return
	}
	if fh.Size > 2*1024*1024 {
		_ = c.Error(apperrors.ErrBadRequest("file exceeds 2MB limit"))
		return
	}
	f, err := fh.Open()
	if err != nil {
		_ = c.Error(apperrors.ErrBadRequest("cannot open file"))
		return
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, 2*1024*1024+1))
	if err != nil {
		_ = c.Error(apperrors.ErrInternal("failed to read file"))
		return
	}
	if len(content) > 2*1024*1024 {
		_ = c.Error(apperrors.ErrBadRequest("file exceeds 2MB limit"))
		return
	}
	sendInvite := strings.EqualFold(c.PostForm("send_invite"), "true") || c.PostForm("send_invite") == "1"

	out, err := h.svc.ImportCSV(c.Request.Context(), content, employeeFieldPerms(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	if sendInvite && h.reset != nil {
		for _, r := range out.Results {
			if r.OK && r.Email != "" {
				_ = h.reset.RequestReset(c.Request.Context(), r.Email)
			}
		}
	}
	msg := fmt.Sprintf("Import finished: %d created, %d failed", out.Created, out.Failed)
	ok(c, http.StatusOK, out, msg)
}

// ImportTemplate godoc
// @Summary      Download employee CSV import template
// @Tags         employees
// @Security     BearerAuth
// @Produce      text/csv
// @Success      200  {string}  string
// @Router       /api/v1/employees/import/template [get]
func (h *EmployeeHandler) ImportTemplate(c *gin.Context) {
	const body = "email,first_name,last_name,phone,department,position,role,manager_email,join_date,is_active\n" +
		"jane.doe@example.com,Jane,Doe,+84901234567,Engineering,Software Engineer,Employee,,2026-07-01,true\n"
	c.Header("Content-Disposition", `attachment; filename="employee-import-template.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(body))
}
```

- [ ] **Step 2: Routes in `main.go`** — static segments before `:id`:

```go
adminEmps.POST("/import", middleware.RequirePerms(authSvc, permissions.PermEmployeesCreate), empH.Import)
adminEmps.GET("/import/template", middleware.RequirePerms(authSvc, permissions.PermEmployeesCreate), empH.ImportTemplate)
```

Place next to `manager-candidates` registration.

- [ ] **Step 3: Build + swag**

```bash
go build ./...
go vet ./...
export PATH="$(go env GOPATH)/bin:$PATH"
make swag
```

- [ ] **Step 4: Commit**

```bash
git add internal/handlers/employee_handler.go cmd/server/main.go docs/swagger/
git commit -m "feat(employees): CSV import HTTP endpoints + template"
```

---

### Task 5: Live verification

**Files:**
- Create: `docs/superpowers/verification/employee-csv-import.md`

- [ ] **Step 1: Boot server** (`make run` or `PORT=8082 make run`)

- [ ] **Step 2: Smoke script** (admin JWT required)

```bash
BASE=http://localhost:8082
TOKEN=… # login as admin

# template
curl -sS -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/employees/import/template" -o /tmp/emp-tpl.csv
head -2 /tmp/emp-tpl.csv

# import good + bad
cat > /tmp/emp-import.csv <<'EOF'
email,first_name,last_name,department
alice.import@example.com,Alice,Import,
bob.dup@example.com,Bob,Dup,
EOF
# pre-create bob.dup if needed to force duplicate

curl -sS -H "Authorization: Bearer $TOKEN" \
  -F "file=@/tmp/emp-import.csv" \
  -F "send_invite=false" \
  "$BASE/api/v1/employees/import" | jq .

# expect: alice ok, bob fail if pre-existed; 401 without token; 403 without employees:create
```

- [ ] **Step 3: DB spot-check**

```sql
SELECT e.first_name, e.last_name, u.email
FROM employees e JOIN users u ON u.id = e.user_id
WHERE u.email LIKE '%import@example.com' AND e.is_deleted = false;
```

- [ ] **Step 4: Write verification log + update CHECKPOINT**

- [ ] **Step 5: Commit docs**

```bash
git add docs/superpowers/verification/employee-csv-import.md docs/superpowers/CHECKPOINT.md
git commit -m "docs: employee CSV import verification + checkpoint"
```

---

## Out of scope (v1)

- Upsert / update-on-duplicate
- Multi-role or multi-skill columns
- Emergency contacts / file attachments
- Async job queue for huge files
- Excel (`.xlsx`) — user asked CSV only
- Cross-row manager ordering (managers must pre-exist)
- New BA DR authoring (optional follow-up)

## Follow-ups (not this plan)

- Two-pass import for managers defined later in the same file
- Dry-run mode (`dry_run=true` validates without write)
- Soft-delete email reuse edge cases (partial unique on email — already handled by Create)

---

## Self-review

| Spec item | Task |
|-----------|------|
| CSV upload | T2, T4 |
| Per-row Create reuse | T2 |
| Partial success report | T1, T2 |
| Name/email FK resolution | T2 |
| Permission = create | T4 |
| Template download | T4 |
| Tests + live verify | T3, T5 |
| No migration | (none) |

No TBD placeholders. Types consistent across tasks.

---

## Execution handoff

Plan saved to `docs/superpowers/plans/2026-07-23-employee-csv-import.md`.

**Options:**

1. **Subagent-driven** (recommended) — fresh subagent per task, review between tasks  
2. **Inline** — this session, `executing-plans`, checkpoints  

**Which approach?**
