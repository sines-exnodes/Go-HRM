# Attendance Parity A — Leave-Integrated Matrix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the Go monthly attendance matrix to BA parity — render approved leave in cells, support combined half-day-leave cells, make the status filter leave-aware, compute leave-aware Total Late/Early summaries, and make root `GET /attendance` return the matrix.

**Architecture:** The matrix already builds per-employee/per-day cells in `attendance_matrix.go`. This plan injects a `LeaveRequestRepository` into `AttendanceService`, batch-fetches approved leaves overlapping the month, expands them to a per-employee per-date map (first-leave-wins), and threads leave into cell-status derivation, combined-cell `worked_half_status`, the SR-011 summary math, and the status filter. Route wiring changes so the matrix is served at the root path (D1).

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, testify. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-06-05-attendance-parity-audit.md` (gaps G1, G2, G4, G6; decision D1). BA: `ba-requirements/.../EP-004-attendance-management/US-001-attendance-list/details/DR-004-001-01-attendance-list.md` v1.2 (AC-016, AC-026–031, SR-002/003/004/008/011).

**Scope note:** This plan does NOT cover Excel export (G3), auto-checkout (G5), or the check-in/out response shape (D2/D5) — those are in Plan B (`2026-06-05-attendance-parity-b-export-jobs.md`). Holiday "H" cells (G7) are out of scope (blocked on a holiday calendar source).

**Conventions to honor (AGENTS.md Rule 11):**
- `make fmt && make vet && make test` before claiming done; `make swag` when handler annotations change.
- Services return `*errors.AppError`; never import `gin`.
- Repo joins must qualify `is_deleted` (a JOIN to a table that also has `is_deleted` makes the bare `NotDeleted` scope ambiguous — see `attendance_repo.go` `List`/`MonthlyCheckInCount`).
- All date filters compare via `.Format("2006-01-02")` against the `date` columns.
- Run impact analysis before editing a symbol (`gitnexus_impact`) per CLAUDE.md.

---

## File map

| File | Responsibility | Change |
|---|---|---|
| `internal/repositories/leave_request_repo.go` | leave data access | **Modify** — add `ApprovedForEmployeesInRange` to interface + impl |
| `internal/services/attendance_service.go` | service struct + ctor | **Modify** — add `leaves` dependency |
| `internal/services/attendance_matrix.go` | matrix build | **Modify** — leave integration, combined cells, SR-011, filter |
| `internal/dto/attendance.go` | wire shapes | **Modify** — extend `AttendanceCellRead`, matrix status consts |
| `cmd/server/main.go` | DI + routes | **Modify** — pass leave repo to ctor; root path serves matrix (D1) |
| `internal/services/attendance_service_test.go` | test helper | **Modify** — `newAttendanceSvc` passes the leave repo |
| `internal/services/attendance_matrix_leave_test.go` | new tests | **Create** — leave/combined/summary/filter coverage |
| `internal/repositories/leave_request_repo_test.go` | repo test | **Modify or Create** — `ApprovedForEmployeesInRange` coverage |
| `docs/swagger/*` | OpenAPI | **Regenerate** via `make swag` |

---

## Reference: domain facts the implementer needs

- **Leave enums** (`internal/models/leave_request.go`): `LeaveType` ∈ {annual, sick, personal, maternity, unpaid}; `LeavePeriod` ∈ {full_day, morning_half, afternoon_half}; `LeaveStatus` ∈ {pending, approved, rejected, cancelled}. Helper `models.IsHalfDayPeriod(p)` exists.
- **`models.LeaveRequest`** fields: `EmployeeID uuid.UUID`, `FromDate time.Time`, `ToDate time.Time` (both `date` cols), `LeavePeriod`, `LeaveType`, `Status`.
- **Thresholds** live on `config.Config`: `LateThresholdHour/Minute` (default 9:00), `CheckoutThresholdHour/Minute` (default 18:00), `CompanyTimezone`. The AM-half end (12:00) and PM-half late (13:15) are BA-fixed constants (Python hardcodes `_AM_END`/`_PM_LATE`) — add them as Go consts.
- **`thresholdAt(ref, hour, minute)`** (attendance_helpers.go) builds an hh:mm timestamp on `ref`'s calendar day in `ref`'s location.
- **Matrix today-vs-cell**: a past workday with no record → `absent`; today/future → `no_data`; weekend → `weekend`. Existing logic in `attendance_matrix.go`.
- **Existing matrix status consts** (`attendance_matrix.go`): `matrixOnTime`, `matrixLate`, `matrixAbsent`, `matrixWeekend`, `matrixNoData`.

---

### Task 1: Leave repo — `ApprovedForEmployeesInRange`

**Files:**
- Modify: `internal/repositories/leave_request_repo.go` (interface near line 60; impl after `Overlapping` ~line 256)
- Test: `internal/repositories/leave_request_repo_test.go` (create if absent)

- [ ] **Step 1: Add the interface method**

In the `LeaveRequestRepository` interface (after the `Overlapping` declaration, ~line 60), add:

```go
	// ApprovedForEmployeesInRange returns live, status='approved' rows for the
	// given employees whose [from_date, to_date] intersects [from, to].
	// Used by the attendance matrix to overlay leave on cells. employeeIDs
	// empty → no rows.
	ApprovedForEmployeesInRange(ctx context.Context, employeeIDs []uuid.UUID, from, to time.Time) ([]models.LeaveRequest, error)
```

- [ ] **Step 2: Add the implementation**

Append to `internal/repositories/leave_request_repo.go`:

```go
func (r *leaveRequestRepo) ApprovedForEmployeesInRange(ctx context.Context, employeeIDs []uuid.UUID, from, to time.Time) ([]models.LeaveRequest, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	var items []models.LeaveRequest
	err := r.base(ctx).
		Model(&models.LeaveRequest{}).
		Where("employee_id IN ?", employeeIDs).
		Where("status = ?", string(models.LeaveStatusApproved)).
		Where("from_date <= ? AND to_date >= ?", to.Format("2006-01-02"), from.Format("2006-01-02")).
		Find(&items).Error
	return items, err
}
```

- [ ] **Step 3: Write the failing test**

Create `internal/repositories/leave_request_repo_test.go` (or add to it if present). If creating, match the package + DB-guard pattern used by other repo tests in this package (`skipIfNoDB`, `truncateAll`, `testDB` are shared in the package test main). Use `makeEmpUser`/`makeUser` helpers if they exist in the repositories test package; otherwise create the employee row directly with `testDB.Create`.

```go
func TestLeaveRepo_ApprovedForEmployeesInRange(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	repo := NewLeaveRequestRepository(testDB)

	emp := &models.Employee{
		FirstName: "Leave", LastName: "Subject",
		ContractType: "official", ContractRenewal: 1, PaymentMethod: "bank_transfer",
	}
	require.NoError(t, testDB.Create(emp).Error)

	mk := func(from, to string, status models.LeaveStatus) {
		f, _ := time.Parse("2006-01-02", from)
		tt, _ := time.Parse("2006-01-02", to)
		require.NoError(t, testDB.Create(&models.LeaveRequest{
			EmployeeID: emp.ID, FromDate: f, ToDate: tt,
			LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual,
			TotalDays: 1, Reason: "x", Status: status, CreatedBy: emp.ID,
		}).Error)
	}
	mk("2026-05-10", "2026-05-12", models.LeaveStatusApproved) // overlaps
	mk("2026-05-20", "2026-05-20", models.LeaveStatusPending)  // wrong status
	mk("2026-06-01", "2026-06-02", models.LeaveStatusApproved) // outside month

	from, _ := time.Parse("2006-01-02", "2026-05-01")
	to, _ := time.Parse("2006-01-02", "2026-05-31")
	got, err := repo.ApprovedForEmployeesInRange(ctx, []uuid.UUID{emp.ID}, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "2026-05-10", got[0].FromDate.Format("2006-01-02"))
}
```

- [ ] **Step 4: Run it — expect FAIL then PASS**

Run: `go test ./internal/repositories/ -run TestLeaveRepo_ApprovedForEmployeesInRange -v`
Expected: compiles, PASS (the method exists from Steps 1–2). If the repo test package has no shared `testDB`/`skipIfNoDB`, mirror the helper bootstrap from `attendance` service tests' `TestMain` — check `internal/repositories/*_test.go` for the existing pattern before adding a new `TestMain`.

- [ ] **Step 5: Commit**

```bash
git add internal/repositories/leave_request_repo.go internal/repositories/leave_request_repo_test.go
git commit -m "feat(attendance): add ApprovedForEmployeesInRange leave query for matrix overlay"
```

---

### Task 2: Inject `LeaveRequestRepository` into `AttendanceService`

**Files:**
- Modify: `internal/services/attendance_service.go:24-41` (struct + ctor)
- Modify: `cmd/server/main.go:97` (ctor call)
- Modify: `internal/services/attendance_service_test.go:35-42` (test helper)

- [ ] **Step 1: Add the field + ctor param**

In `attendance_service.go`, extend the struct (after `pos`):

```go
type AttendanceService struct {
	cfg    *config.Config
	repo   repositories.AttendanceRepository
	emps   repositories.EmployeeRepository
	depts  repositories.DepartmentRepository
	pos    repositories.PositionRepository
	leaves repositories.LeaveRequestRepository
}
```

And the constructor:

```go
func NewAttendanceService(
	cfg *config.Config,
	repo repositories.AttendanceRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	pos repositories.PositionRepository,
	leaves repositories.LeaveRequestRepository,
) *AttendanceService {
	return &AttendanceService{cfg: cfg, repo: repo, emps: emps, depts: depts, pos: pos, leaves: leaves}
}
```

- [ ] **Step 2: Update the DI wiring in main.go**

`cmd/server/main.go` already constructs `leaveRequestRepo` (it backs the leave service). Find that variable (grep `NewLeaveRequestRepository` in `main.go`) and pass it into the attendance ctor at line ~97:

```go
	attendanceSvc := services.NewAttendanceService(cfg, attendanceRepo, employeeRepo, departmentRepo, positionRepo, leaveRequestRepo)
```

If `leaveRequestRepo` is declared *after* the attendance ctor call, move its declaration up so it precedes line 97.

- [ ] **Step 3: Update the test helper**

`attendance_service_test.go` `newAttendanceSvc` — add the leave repo to the call:

```go
	return services.NewAttendanceService(
		cfg,
		repositories.NewAttendanceRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		repositories.NewDepartmentRepository(testDB),
		repositories.NewPositionRepository(testDB),
		repositories.NewLeaveRequestRepository(testDB),
	)
```

- [ ] **Step 4: Build to verify wiring**

Run: `make vet`
Expected: compiles clean (no unused/ missing-arg errors). Existing attendance tests still build.

- [ ] **Step 5: Commit**

```bash
git add internal/services/attendance_service.go cmd/server/main.go internal/services/attendance_service_test.go
git commit -m "refactor(attendance): inject LeaveRequestRepository into AttendanceService"
```

---

### Task 3: Extend the cell DTO + matrix status constants

**Files:**
- Modify: `internal/dto/attendance.go:149-160` (`AttendanceCellRead`)
- Modify: `internal/services/attendance_matrix.go:172-179` (status consts)

- [ ] **Step 1: Extend `AttendanceCellRead`**

Replace the struct (`attendance.go:151-160`) with:

```go
// AttendanceCellRead is one day in the matrix. Status enum:
// on_time | late | absent | weekend | no_data | annual_leave | sick_leave |
// personal_leave | maternity_leave | unpaid_leave | half_day_leave.
// For combined cells (half_day_leave + a worked half), WorkedHalfStatus is
// one of on_time | late | absent; LeaveType/LeavePeriod describe the leave half.
type AttendanceCellRead struct {
	Date             string                  `json:"date"`
	Day              int                     `json:"day"`
	Status           string                  `json:"status"`
	CheckIn          *time.Time              `json:"check_in,omitempty"`
	CheckOut         *time.Time              `json:"check_out,omitempty"`
	HoursWorked      *float64                `json:"hours_worked,omitempty"`
	IsLate           bool                    `json:"is_late"`
	LeaveType        *string                 `json:"leave_type,omitempty"`
	LeavePeriod      *string                 `json:"leave_period,omitempty"`
	WorkedHalfStatus *string                 `json:"worked_half_status,omitempty"`
	Sessions         []AttendanceSessionRead `json:"sessions,omitempty"`
}
```

- [ ] **Step 2: Extend the matrix status constants**

Replace the const block (`attendance_matrix.go:172-179`) with:

```go
// Matrix cell status enum.
const (
	matrixOnTime         = "on_time"
	matrixLate           = "late"
	matrixAbsent         = "absent"
	matrixWeekend        = "weekend"
	matrixNoData         = "no_data"
	matrixAnnualLeave    = "annual_leave"
	matrixSickLeave      = "sick_leave"
	matrixPersonalLeave  = "personal_leave"
	matrixMaternityLeave = "maternity_leave"
	matrixUnpaidLeave    = "unpaid_leave"
	matrixHalfDayLeave   = "half_day_leave"
)

// Half-day boundary constants — BA-fixed (DR-004-001-01 SR-002/SR-011 v1.2),
// mirror Python's _AM_END / _PM_LATE. AM late + workday-end thresholds come
// from config (LateThreshold / CheckoutThreshold).
const (
	amHalfEndHour  = 12 // end of the AM half (early-leave boundary when AM worked + PM on leave)
	amHalfEndMin   = 0
	pmLateHour     = 13 // PM-half late threshold (when AM on leave + PM worked)
	pmLateMin      = 15
)

// leaveTypeToStatus maps an approved full-day leave type to its cell status.
var leaveTypeToStatus = map[models.LeaveType]string{
	models.LeaveTypeAnnual:    matrixAnnualLeave,
	models.LeaveTypeSick:      matrixSickLeave,
	models.LeaveTypePersonal:  matrixPersonalLeave,
	models.LeaveTypeMaternity: matrixMaternityLeave,
	models.LeaveTypeUnpaid:    matrixUnpaidLeave,
}

// onLeaveStatuses is the set used by the on_leave status filter (G4).
var onLeaveStatuses = map[string]struct{}{
	matrixAnnualLeave: {}, matrixSickLeave: {}, matrixPersonalLeave: {},
	matrixMaternityLeave: {}, matrixUnpaidLeave: {}, matrixHalfDayLeave: {},
}
```

- [ ] **Step 3: Build**

Run: `make vet`
Expected: compiles (new symbols may be unused until Task 4 — if vet flags an unused `var`, proceed; the next task consumes them. If the build *fails* on unused package-level vars, that's fine — Go allows unused package-level declarations).

- [ ] **Step 4: Commit**

```bash
git add internal/dto/attendance.go internal/services/attendance_matrix.go
git commit -m "feat(attendance): extend matrix cell DTO + status consts for leave overlay"
```

---

### Task 4: Render approved leave in matrix cells (G1, AC-016)

**Files:**
- Modify: `internal/services/attendance_matrix.go` (`Matrix` — fetch leaves; per-day cell derivation)
- Test: `internal/services/attendance_matrix_leave_test.go` (create)

This task introduces a per-employee `leaveByDate` map and a `cellForDay` helper that mirrors Python's `_cell_for_day` precedence: **weekend → approved leave → attendance record → absent/no_data**.

- [ ] **Step 1: Write the failing test**

Create `internal/services/attendance_matrix_leave_test.go`:

```go
package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

// makeApprovedLeave inserts an approved leave row for emp over [from,to].
func makeApprovedLeave(t *testing.T, empID interface{ String() string }, from, to string, lt models.LeaveType, lp models.LeavePeriod) {
	t.Helper()
	f, err := time.Parse("2006-01-02", from)
	require.NoError(t, err)
	tt, err := time.Parse("2006-01-02", to)
	require.NoError(t, err)
	id, _ := parseUUIDForTest(empID.String())
	require.NoError(t, testDB.Create(&models.LeaveRequest{
		EmployeeID: id, FromDate: f, ToDate: tt,
		LeavePeriod: lp, LeaveType: lt, TotalDays: 1, Reason: "test",
		Status: models.LeaveStatusApproved, CreatedBy: id,
	}).Error)
}

func TestAttendance_Matrix_FullDayLeaveCell(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "leave1@example.com", "LeaveOne")

	// April 8 2026 is a Wednesday (workday). Approve full-day sick leave.
	makeApprovedLeave(t, emp.ID, "2026-04-08", "2026-04-08", models.LeaveTypeSick, models.LeavePeriodFullDay)

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	cell := out.Items[0].Cells[8]
	assert.Equal(t, "sick_leave", cell.Status)
	require.NotNil(t, cell.LeaveType)
	assert.Equal(t, "sick", *cell.LeaveType)
}
```

> The helper uses `parseUUIDForTest` and the existing `makeEmpUser`. `makeEmpUser` already returns `(user, employee)` where `employee.ID` is a `uuid.UUID` (has `.String()`). Add a tiny `parseUUIDForTest` if not present:
> ```go
> func parseUUIDForTest(s string) (uuid.UUID, error) { return uuid.Parse(s) }
> ```
> (import `github.com/google/uuid`). Simpler: change `makeApprovedLeave` to take `uuid.UUID` directly and pass `emp.ID`. Prefer that — drop the interface gymnastics:
> ```go
> func makeApprovedLeave(t *testing.T, empID uuid.UUID, from, to string, lt models.LeaveType, lp models.LeavePeriod) { ... EmployeeID: empID ... }
> ```

- [ ] **Step 2: Run it — expect FAIL**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_FullDayLeaveCell -v`
Expected: FAIL — cell status is `no_data`/`absent`, not `sick_leave` (no leave integration yet).

- [ ] **Step 3: Fetch approved leaves in `Matrix`**

In `attendance_matrix.go` `Matrix`, right after the attendance `records` fetch + `byEmp` map build (~line 247-259), add the leave fetch + per-employee per-date expansion:

```go
	// Approved leave overlay (G1). First-leave-wins on overlapping dates,
	// matching Python's _build_all_rows expansion.
	leaves, err := s.leaves.ApprovedForEmployeesInRange(ctx, ids, first, last)
	if err != nil {
		return dto.AttendanceMatrixRead{}, err
	}
	leaveByEmp := make(map[uuid.UUID]map[string]models.LeaveRequest, len(employees))
	for _, lv := range leaves {
		m, ok := leaveByEmp[lv.EmployeeID]
		if !ok {
			m = make(map[string]models.LeaveRequest)
			leaveByEmp[lv.EmployeeID] = m
		}
		for d := lv.FromDate; !d.After(lv.ToDate); d = d.AddDate(0, 0, 1) {
			key := d.Format("2006-01-02")
			if d.Before(first) || d.After(last) {
				continue
			}
			if _, exists := m[key]; !exists {
				m[key] = lv
			}
		}
	}
```

- [ ] **Step 4: Apply leave in the per-day cell loop**

In the per-day `switch` (the `default:` branch, ~line 279-319), insert a leave check **before** the attendance-record check. Replace the `default:` body so the precedence is weekend (already handled by the `case`) → leave → record → absent/no_data:

```go
			default:
				key := day.Format("2006-01-02")
				if lv, onLeave := leaveByEmp[emp.ID][key]; onLeave {
					applyLeaveCell(&cell, lv)
					// Combined-cell + summary handled in Task 5/6; for a
					// full-day leave the cell is complete here.
					if !models.IsHalfDayPeriod(lv.LeavePeriod) {
						break
					}
					// half-day leave: fall through to also read any record
					// (combined cell). Task 5 fills WorkedHalfStatus.
					if rec, ok := empRecs[key]; ok && len(rec.Sessions) > 0 {
						attachRecordToCell(&cell, rec, loc)
					}
					break
				}
				rec, ok := empRecs[key]
				if ok && len(rec.Sessions) > 0 {
					attachRecordToCell(&cell, rec, loc)
					if rec.IsLate {
						cell.Status = matrixLate
					} else {
						cell.Status = matrixOnTime
					}
					// per-day late/early totals — moved to Task 6.
				} else if day.Before(now) {
					cell.Status = matrixAbsent
				} else {
					cell.Status = matrixNoData
				}
			}
```

> Note: this temporarily REMOVES the inline late/early accumulation that currently lives in this branch (lines ~302-314). Task 6 reintroduces it as a single leave-aware block. Between Task 4 and Task 6 the summary totals will be zero — that's expected and covered by Task 6's tests.

- [ ] **Step 5: Add the cell helpers**

Add to `attendance_matrix.go` (package-level funcs):

```go
// applyLeaveCell sets a cell's leave fields + status from an approved leave.
// Full-day leave → the type-specific status; half-day leave → half_day_leave.
func applyLeaveCell(cell *dto.AttendanceCellRead, lv models.LeaveRequest) {
	lt := string(lv.LeaveType)
	lp := string(lv.LeavePeriod)
	cell.LeaveType = &lt
	cell.LeavePeriod = &lp
	if models.IsHalfDayPeriod(lv.LeavePeriod) {
		cell.Status = matrixHalfDayLeave
		return
	}
	if st, ok := leaveTypeToStatus[lv.LeaveType]; ok {
		cell.Status = st
	} else {
		cell.Status = matrixAnnualLeave // defensive default; all 5 types are mapped
	}
}

// attachRecordToCell fills check-in/out, hours and sessions from a record.
func attachRecordToCell(cell *dto.AttendanceCellRead, rec models.Attendance, loc *time.Location) {
	f := rec.Sessions[0]
	l := rec.Sessions[len(rec.Sessions)-1]
	ci := f.CheckIn
	cell.CheckIn = &ci
	cell.CheckOut = l.CheckOut
	cell.IsLate = rec.IsLate
	var total float64
	sessions := make([]dto.AttendanceSessionRead, 0, len(rec.Sessions))
	for _, sess := range rec.Sessions {
		hw := hoursBetween(sess.CheckIn, sess.CheckOut)
		sessions = append(sessions, dto.AttendanceSessionRead{
			ID: sess.ID, CheckIn: sess.CheckIn, CheckOut: sess.CheckOut,
			IsAutoCheckout: sess.IsAutoCheckout, HoursWorked: hw,
		})
		if hw != nil {
			total += *hw
		}
	}
	if total > 0 {
		cell.HoursWorked = &total
	}
	cell.Sessions = sessions
}
```

> These helpers replace the inline record-projection that currently lives in the `default` branch. Remove the now-duplicated inline projection so each concern is computed once.

- [ ] **Step 6: Run the test — expect PASS**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_FullDayLeaveCell -v`
Expected: PASS.

- [ ] **Step 7: Run the full matrix regression set**

Run: `go test ./internal/services/ -run TestAttendance_Matrix -v`
Expected: all existing matrix tests (`ManagerSeesAllEmployees`, `EmployeeSeesOwnRow`, `WeekendsMarked`) still PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/services/attendance_matrix.go internal/services/attendance_matrix_leave_test.go
git commit -m "feat(attendance): render approved leave in matrix cells (G1, AC-016)"
```

---

### Task 5: Combined half-day cells — `worked_half_status` (G2, AC-026..030)

**Files:**
- Modify: `internal/services/attendance_matrix.go` (combined-cell branch + helper)
- Test: `internal/services/attendance_matrix_leave_test.go`

- [ ] **Step 1: Write failing tests**

Append to `attendance_matrix_leave_test.go`:

```go
// AM on leave, PM worked late (in 13:25 > 13:15 PM threshold) → ½ + worked_half late.
func TestAttendance_Matrix_CombinedCell_PMWorkedLate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "combo1@example.com", "Combo")

	// April 9 2026 = Thursday. Morning-half annual leave + PM check-in 13:25.
	makeApprovedLeave(t, emp.ID, "2026-04-09", "2026-04-09", models.LeaveTypeAnnual, models.LeavePeriodMorningHalf)
	ci := hcmTime(t, 2026, 4, 9, 13, 25)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	cell := out.Items[0].Cells[9]
	assert.Equal(t, "half_day_leave", cell.Status)
	require.NotNil(t, cell.WorkedHalfStatus)
	assert.Equal(t, "late", *cell.WorkedHalfStatus)
}

// Morning-half leave + no PM check-in → worked half Absent (AC-030).
func TestAttendance_Matrix_CombinedCell_NoCheckIn_Absent(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "combo2@example.com", "Combo2")

	makeApprovedLeave(t, emp.ID, "2026-04-09", "2026-04-09", models.LeaveTypeAnnual, models.LeavePeriodMorningHalf)

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	cell := out.Items[0].Cells[9]
	assert.Equal(t, "half_day_leave", cell.Status)
	require.NotNil(t, cell.WorkedHalfStatus)
	assert.Equal(t, "absent", *cell.WorkedHalfStatus)
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_CombinedCell -v`
Expected: FAIL — `WorkedHalfStatus` is nil.

- [ ] **Step 3: Add `computeWorkedHalfStatus` + call it**

Add the helper to `attendance_matrix.go`:

```go
// computeWorkedHalfStatus derives the worked half's status for a combined
// half-day-leave cell (mirror Python _compute_worked_half_status).
//   - no check-in → "absent" (AC-030)
//   - PM worked (leave = morning_half) → late if first check-in > 13:15
//   - AM worked (leave = afternoon_half) → late if first check-in > 09:00
func (s *AttendanceService) computeWorkedHalfStatus(cell *dto.AttendanceCellRead, lv models.LeaveRequest, loc *time.Location) string {
	if cell.CheckIn == nil {
		return matrixAbsent
	}
	firstLocal := cell.CheckIn.In(loc)
	var threshold time.Time
	if lv.LeavePeriod == models.LeavePeriodMorningHalf {
		threshold = thresholdAt(firstLocal, pmLateHour, pmLateMin)
	} else {
		threshold = thresholdAt(firstLocal, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
	}
	if firstLocal.After(threshold) {
		return matrixLate
	}
	return matrixOnTime
}
```

In the `default:` branch (Task 4 Step 4), after `attachRecordToCell` is called for a half-day leave, set the worked-half status:

```go
				if _, onLeave := leaveByEmp[emp.ID][key]; onLeave {
					lv := leaveByEmp[emp.ID][key]
					applyLeaveCell(&cell, lv)
					if !models.IsHalfDayPeriod(lv.LeavePeriod) {
						break
					}
					if rec, ok := empRecs[key]; ok && len(rec.Sessions) > 0 {
						attachRecordToCell(&cell, rec, loc)
					}
					whs := s.computeWorkedHalfStatus(&cell, lv, loc)
					cell.WorkedHalfStatus = &whs
					break
				}
```

> Replace the Task-4 leave block with this version (it now always sets `WorkedHalfStatus` for half-day cells, including the no-check-in → absent case).

- [ ] **Step 4: Run — expect PASS**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_CombinedCell -v`
Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/services/attendance_matrix.go internal/services/attendance_matrix_leave_test.go
git commit -m "feat(attendance): combined half-day cells with worked_half_status (G2, AC-026..030)"
```

---

### Task 6: Leave-aware Total Late / Total Early summaries (G6, SR-011)

**Files:**
- Modify: `internal/services/attendance_matrix.go` (per-day summary accumulation)
- Test: `internal/services/attendance_matrix_leave_test.go`

SR-011 per-day contribution by day type:

| Day type | Late threshold | Early threshold |
|---|---|---|
| Full-day worked | 09:00 (config late) | 18:00 (config checkout) |
| AM worked + PM half leave (`afternoon_half`) | 09:00 | 12:00 (`amHalfEnd`) |
| AM half leave + PM worked (`morning_half`) | 13:15 (`pmLate`) | 18:00 |
| Full-day leave / weekend / absent / worked-half-absent | 0 | 0 |

- [ ] **Step 1: Write the failing tests**

Append to `attendance_matrix_leave_test.go`:

```go
// AC-028: morning-half leave + PM in 13:25 → 10 late minutes.
func TestAttendance_Matrix_Summary_PMWorkedLate10Min(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "sum1@example.com", "Sum1")
	makeApprovedLeave(t, emp.ID, "2026-04-09", "2026-04-09", models.LeaveTypeSick, models.LeavePeriodMorningHalf)
	ci := hcmTime(t, 2026, 4, 9, 13, 25)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	co := hcmTime(t, 2026, 4, 9, 18, 0)
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 10, out.Items[0].TotalLateMinutes)
	assert.Equal(t, 0, out.Items[0].TotalEarlyMinutes)
}

// AC-029: afternoon-half leave + AM out 11:50 → 10 early minutes (vs 12:00).
func TestAttendance_Matrix_Summary_AMWorkedEarly10Min(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "sum2@example.com", "Sum2")
	makeApprovedLeave(t, emp.ID, "2026-04-10", "2026-04-10", models.LeaveTypeAnnual, models.LeavePeriodAfternoonHalf)
	ci := hcmTime(t, 2026, 4, 10, 9, 0)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	co := hcmTime(t, 2026, 4, 10, 11, 50)
	_, err = svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 0, out.Items[0].TotalLateMinutes)
	assert.Equal(t, 10, out.Items[0].TotalEarlyMinutes)
}

// AC-022: full-day leave contributes 0 to both.
func TestAttendance_Matrix_Summary_FullDayLeaveZero(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "sum3@example.com", "Sum3")
	makeApprovedLeave(t, emp.ID, "2026-04-08", "2026-04-08", models.LeaveTypeAnnual, models.LeavePeriodFullDay)

	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 0, out.Items[0].TotalLateMinutes)
	assert.Equal(t, 0, out.Items[0].TotalEarlyMinutes)
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_Summary -v`
Expected: FAIL — totals are 0 because Task 4 removed the inline accumulation.

- [ ] **Step 3: Add a single leave-aware accumulation block**

Add the helper to `attendance_matrix.go`:

```go
// accumulateSummary adds this cell's per-day late/early minutes to the running
// totals per SR-011. Returns the (lateAdd, earlyAdd) in minutes. Non-working
// day types and the worked-half-absent case contribute 0.
func (s *AttendanceService) accumulateSummary(cell dto.AttendanceCellRead, loc *time.Location) (int, int) {
	// Skip weekend / absent / no_data / full-day leave.
	switch cell.Status {
	case matrixWeekend, matrixAbsent, matrixNoData,
		matrixAnnualLeave, matrixSickLeave, matrixPersonalLeave,
		matrixMaternityLeave, matrixUnpaidLeave:
		return 0, 0
	}
	// Half-day leave with no worked-half check-in (worked half Absent) → 0.
	if cell.Status == matrixHalfDayLeave && cell.CheckIn == nil {
		return 0, 0
	}

	// Pick the threshold pair by worked half.
	lateHour, lateMin := s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute
	earlyHour, earlyMin := s.cfg.CheckoutThresholdHour, s.cfg.CheckoutThresholdMinute
	if cell.Status == matrixHalfDayLeave && cell.LeavePeriod != nil {
		switch *cell.LeavePeriod {
		case string(models.LeavePeriodMorningHalf):
			// AM on leave, PM worked → late vs 13:15, early vs 18:00.
			lateHour, lateMin = pmLateHour, pmLateMin
		case string(models.LeavePeriodAfternoonHalf):
			// AM worked, PM on leave → late vs 09:00, early vs 12:00.
			earlyHour, earlyMin = amHalfEndHour, amHalfEndMin
		}
	}

	var lateAdd, earlyAdd int
	if cell.CheckIn != nil {
		ci := cell.CheckIn.In(loc)
		ref := thresholdAt(ci, lateHour, lateMin)
		if ci.After(ref) {
			lateAdd = int(ci.Sub(ref).Minutes())
		}
	}
	if cell.CheckOut != nil {
		co := cell.CheckOut.In(loc)
		ref := thresholdAt(co, earlyHour, earlyMin)
		if co.Before(ref) {
			earlyAdd = int(ref.Sub(co).Minutes())
		}
	}
	return lateAdd, earlyAdd
}
```

In the per-day loop, after `cells[d] = cell` (and after `cellStatusUnion` is updated), call it:

```go
			cells[d] = cell
			cellStatusUnion[cell.Status] = struct{}{}
			la, ea := s.accumulateSummary(cell, loc)
			totalLate += la
			totalEarly += ea
```

> Ensure the `cell` passed to `accumulateSummary` is the fully-populated value (after Task 5 set `WorkedHalfStatus`/`LeavePeriod`). The accumulation reads `cell.LeavePeriod` + `cell.CheckIn/Out`, which are set by `applyLeaveCell` + `attachRecordToCell`.

- [ ] **Step 4: Run — expect PASS**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_Summary -v`
Expected: all three PASS.

- [ ] **Step 5: Regression — full matrix + the full-day-worked summary still works**

Run: `go test ./internal/services/ -run TestAttendance_Matrix -v`
Expected: all PASS. (The original full-day late/early behavior is preserved by the default threshold pair.)

- [ ] **Step 6: Commit**

```bash
git add internal/services/attendance_matrix.go internal/services/attendance_matrix_leave_test.go
git commit -m "feat(attendance): leave-aware Total Late/Early summaries (G6, SR-011)"
```

---

### Task 7: `on_leave` status filter + combined multi-match (G4, AC-031)

**Files:**
- Modify: `internal/services/attendance_matrix.go` (filter block ~line 325-337)
- Test: `internal/services/attendance_matrix_leave_test.go`

Currently the filter matches the literal `cellStatusUnion`. It must also: (a) treat any of the `onLeaveStatuses` cells as matching `on_leave`; (b) match a combined cell's `worked_half_status` for `on_time`/`late`/`absent`.

- [ ] **Step 1: Write the failing test**

Append to `attendance_matrix_leave_test.go`:

```go
// AC-031: a combined cell (½ + PM late) matches BOTH "late" and "on_leave".
func TestAttendance_Matrix_StatusFilter_CombinedMultiMatch(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "filt1@example.com", "Filt")
	makeApprovedLeave(t, emp.ID, "2026-04-09", "2026-04-09", models.LeaveTypeAnnual, models.LeavePeriodMorningHalf)
	ci := hcmTime(t, 2026, 4, 9, 13, 25) // PM late
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	// Filter by "late" → row present.
	outLate, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10, Status: "late"})
	require.NoError(t, err)
	require.Len(t, outLate.Items, 1)

	// Filter by "on_leave" → same row present.
	outLeave, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10, Status: "on_leave"})
	require.NoError(t, err)
	require.Len(t, outLeave.Items, 1)

	// Filter by "absent" → row excluded (worked half was late, not absent; no absent day).
	outAbsent, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10, Status: "absent"})
	require.NoError(t, err)
	assert.Len(t, outAbsent.Items, 0)
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_StatusFilter_CombinedMultiMatch -v`
Expected: FAIL — `on_leave` is not a known cell status, and/or the combined-cell late match misses.

- [ ] **Step 3: Track a worked-half-status union + rewrite the filter**

In the per-employee loop, alongside `cellStatusUnion`, accumulate worked-half statuses. Add near the `cellStatusUnion` declaration:

```go
		workedHalfUnion := make(map[string]struct{}, 4)
```

After setting the cell (Task 6 Step 3 area):

```go
			if cell.WorkedHalfStatus != nil {
				workedHalfUnion[*cell.WorkedHalfStatus] = struct{}{}
			}
```

Replace the status-filter block (~line 325-337) with:

```go
		// Status CSV filter (G4): a row matches if ANY selected value matches
		// any of its cells. on_leave matches any leave-type cell; on_time/
		// late/absent also match a combined cell's worked_half_status (AC-031).
		if statusSet != nil {
			matched := false
			for sf := range statusSet {
				switch sf {
				case "on_leave":
					for k := range cellStatusUnion {
						if _, ok := onLeaveStatuses[k]; ok {
							matched = true
						}
					}
				default:
					if _, ok := cellStatusUnion[sf]; ok {
						matched = true
					}
					if _, ok := workedHalfUnion[sf]; ok {
						matched = true
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}
```

- [ ] **Step 4: Run — expect PASS**

Run: `go test ./internal/services/ -run TestAttendance_Matrix_StatusFilter -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/services/attendance_matrix.go internal/services/attendance_matrix_leave_test.go
git commit -m "feat(attendance): on_leave status filter + combined-cell multi-match (G4, AC-031)"
```

---

### Task 8: D1 — root `GET /attendance` returns the matrix

**Files:**
- Modify: `cmd/server/main.go:331-345` (attendance route group)
- Modify: `internal/handlers/attendance_handler.go` (Matrix godoc `@Router`; List godoc)
- Test: manual route assertion (no new service test — routing only)

Decision D1: the canonical matrix path is the **root**. Bind `GET ""` → `Matrix`. The old flat-list `List` handler moves to an explicit sub-path `GET /records` (admin convenience, no BA backing — keep but de-canonicalize). `GET /me` and `GET /:id` unchanged.

- [ ] **Step 1: Rewire the routes**

In `main.go`, replace the attendance read routes (currently lines ~339-341):

```go
		attendance.GET("/check-in", ...) // unchanged lines above
		// D1: root path serves the monthly matrix (Python/FE parity).
		attendance.GET("", middleware.RequirePerms(authSvc, permissions.PermAttendanceRead), attendanceH.Matrix)
		// /matrix kept as an explicit alias for one release.
		attendance.GET("/matrix", middleware.RequirePerms(authSvc, permissions.PermAttendanceRead), attendanceH.Matrix)
		// Flat paginated rows (Go-only convenience; not the BA matrix).
		attendance.GET("/records", middleware.RequirePerms(authSvc, permissions.PermAttendanceRead), attendanceH.List)
		attendance.GET(":id", middleware.RequirePerms(authSvc, permissions.PermAttendanceRead), attendanceH.Get)
```

> Keep the admin write routes (`POST ""`, `PATCH :id`, `DELETE :id`) exactly as they are — `POST ""` and `GET ""` coexist (different verbs). D6 keeps the admin CRUD.

- [ ] **Step 2: Update the handler godoc**

In `attendance_handler.go`, update the `Matrix` godoc `@Router` to the root and adjust `List`:

```go
// Matrix godoc
// @Summary      Monthly attendance matrix (managers: all employees; others: own row)
// @Router       /api/v1/attendance [get]
```

```go
// List godoc
// @Summary      Flat list of attendance rows (Go convenience; not the BA matrix)
// @Router       /api/v1/attendance/records [get]
```

- [ ] **Step 3: Regenerate Swagger**

Run: `make swag`
Expected: `docs/swagger/` regenerates with `GET /api/v1/attendance` → matrix and `/attendance/records` → list. No hand-edits.

- [ ] **Step 4: Build + vet**

Run: `make vet`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add cmd/server/main.go internal/handlers/attendance_handler.go docs/swagger
git commit -m "feat(attendance): root GET /attendance returns the matrix (D1); flat list moves to /records"
```

---

### Task 9: Full verification + checkpoint

**Files:**
- Modify: `docs/superpowers/CHECKPOINT.md` (note the parity work)
- Modify: `docs/superpowers/specs/2026-06-05-attendance-parity-audit.md` (mark G1/G2/G4/G6 + D1 done)

- [ ] **Step 1: Format + vet + full test suite**

Run: `make fmt && make vet && make test`
Expected: 0 fail, 0 skip (assuming the test DB is up: `make test-db-up`). The attendance package shows the new leave/combined/summary/filter tests green plus all pre-existing tests.

- [ ] **Step 2: Confirm no skips**

Run: `make test 2>&1 | grep -i skip`
Expected: only `skipIfNoDB`-guarded skips when DB is down. With the DB up, no skips. (AGENTS.md Rule 12 — "tests pass" is false if any were skipped silently.)

- [ ] **Step 3: Update the audit spec status**

In `2026-06-05-attendance-parity-audit.md`, change the §0 gap list to mark G1, G2, G4, G6 and D1 as ✅ DONE (Plan A), leaving G3/G5/D2/D5 for Plan B and G7 blocked.

- [ ] **Step 4: Update CHECKPOINT**

Add a "Attendance parity — Plan A (matrix/leave)" entry under post-migration parity work noting: leave-integrated matrix + combined half-day cells + on_leave filter + SR-011 summaries + root-path matrix (D1) landed; migration version unchanged (no schema change); Plan B (export/auto-checkout/response-shape) still pending.

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/CHECKPOINT.md docs/superpowers/specs/2026-06-05-attendance-parity-audit.md
git commit -m "docs(attendance): close Plan A (matrix/leave parity) in checkpoint + audit"
```

---

## Self-review checklist (run before handing off)

- **Spec coverage:** G1 (Task 4), G2 (Task 5), G6 (Task 6), G4 (Task 7), D1 (Task 8). G3/G5/D2/D5 explicitly deferred to Plan B; G7 out of scope. ✅
- **No schema change:** Plan A adds no migration — it only reads `leave_requests`. Migration version stays at 21. ✅
- **Type consistency:** `AttendanceCellRead.WorkedHalfStatus/LeaveType/LeavePeriod` are `*string` everywhere; matrix status consts are the single source; `accumulateSummary`/`computeWorkedHalfStatus` are methods on `*AttendanceService` (need `s.cfg`). ✅
- **Ambiguous-is_deleted rule:** the new leave query uses the leave repo's own `base` scope (single table, no join) — no qualification needed. ✅
- **Removed inline code:** Task 4 removes the old inline record-projection + late/early accumulation; Task 6 reintroduces accumulation leave-aware. Verify no double-counting remains in the `default` branch after Task 6.
