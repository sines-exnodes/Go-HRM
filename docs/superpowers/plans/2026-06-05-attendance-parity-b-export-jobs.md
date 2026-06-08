# Attendance Parity B — Export, Auto-Checkout & Response Shape Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining attendance parity gaps — Excel export of the matrix (bulk + per-employee), an 11 PM auto check-out service + admin trigger, the check-in/out response shape, and the `is_half_day` semantics fix.

**Architecture:** Reuse Plan A's leave-aware matrix row builder to produce export rows, then stream an `.xlsx` via `xuri/excelize/v2`. Auto check-out lands as an idempotent service method over the existing `OpenSessionsBefore` repo query, exposed via an admin endpoint now (scheduler later). The check-in/out handlers return `TodayStatusRead` by calling the existing `Today` builder after the action. The `is_half_day` auto-flip is removed.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, testify, **`github.com/xuri/excelize/v2`** (new dep).

**Spec:** `docs/superpowers/specs/2026-06-05-attendance-parity-audit.md` (gaps G3, G5; decisions D2, D5). BA: DR-004-001-01 (AC-011/012/025, SR-009); DR-003-001-01 (AC-11/12, Rule 5).

**Dependency:** **Execute Plan A first** (`2026-06-05-attendance-parity-a-matrix-leave.md`). Export reuses Plan A's leave-aware row builder so the exported Total Late/Early columns are correct; this plan assumes the `buildAllRows` extraction from Plan A Task 2 below exists. If Plan A is not yet merged, do its Task 2 extraction here first.

**Conventions:** same as Plan A — `make fmt && make vet && make test` + `make swag` on handler changes; services return `*errors.AppError`, never import `gin`; run `gitnexus_impact` before editing a symbol.

---

## File map

| File | Responsibility | Change |
|---|---|---|
| `go.mod` / `go.sum` | deps | **Modify** — add `xuri/excelize/v2` |
| `internal/services/attendance_matrix.go` | row builder | **Modify** — extract `buildAllRows` (if not done in Plan A) |
| `internal/services/attendance_export.go` | xlsx export | **Create** — `ExportMatrix` + label/format helpers |
| `internal/services/attendance_service.go` | auto-checkout + half-day | **Modify** — `AutoCheckOut`; drop `is_half_day` flip (D5) |
| `internal/handlers/attendance_handler.go` | HTTP | **Modify** — `Export`, `ExportEmployee`, `AutoCheckOut` handlers; check-in/out return `TodayStatusRead` (D2) |
| `cmd/server/main.go` | routes | **Modify** — `/export`, `/export/:employee_id`, `/auto-checkout` |
| `internal/services/attendance_export_test.go` | tests | **Create** — export bytes + single-employee scoping |
| `internal/services/attendance_service_test.go` | tests | **Modify** — D5 half-day test; auto-checkout test |
| `docs/swagger/*` | OpenAPI | **Regenerate** via `make swag` |

---

### Task 1: Add the excelize dependency

**Files:** `go.mod`, `go.sum`

- [ ] **Step 1: Add the module**

Run: `go get github.com/xuri/excelize/v2@latest`
Expected: `go.mod` gains `github.com/xuri/excelize/v2 vX.Y.Z`; `go.sum` updated.

- [ ] **Step 2: Tidy + verify it compiles**

Run: `go mod tidy && make vet`
Expected: clean; the dep resolves.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "build(attendance): add xuri/excelize/v2 for xlsx export"
```

---

### Task 2: Extract `buildAllRows` from `Matrix` (if Plan A didn't)

**Files:** Modify `internal/services/attendance_matrix.go`

Export and Matrix must share the same leave-aware row construction. Extract the per-employee loop into a reusable method.

- [ ] **Step 1: Add the builder signature**

Add to `attendance_matrix.go`:

```go
// buildAllRows constructs every employee's matrix row (unpaginated) for the
// given month. statusSet filters rows (nil = no filter). Shared by Matrix
// (which paginates the result) and ExportMatrix (which writes all rows).
func (s *AttendanceService) buildAllRows(
	ctx context.Context,
	employees []models.Employee,
	year, month int,
	loc *time.Location,
	statusSet map[string]struct{},
) ([]dto.AttendanceRowRead, error) {
	// ... body = the per-employee loop currently inline in Matrix:
	//   - first/last/daysInMonth derived from year/month/loc
	//   - fetch records via s.repo.ListForEmployeesInRange
	//   - fetch leaves via s.leaves.ApprovedForEmployeesInRange (Plan A)
	//   - per-day cellForDay + leave overlay + accumulateSummary
	//   - status filter via statusSet + workedHalfUnion
	//   - department-name inflation
	// Return the full []dto.AttendanceRowRead (no pagination here).
}
```

- [ ] **Step 2: Move the body**

Cut the employee-fetch-onward logic out of `Matrix` (everything from the `records, err := s.repo.ListForEmployeesInRange(...)` line through the `rows = append(rows, row)` loop) into `buildAllRows`. `Matrix` keeps: input defaulting (year/month/page/size), employee resolution (admin vs self), the call `rows, err := s.buildAllRows(ctx, employees, year, month, loc, parseCSVSet(q.Status))`, then pagination + the `AttendanceMatrixRead` envelope.

- [ ] **Step 3: Run the full matrix suite — must stay green**

Run: `go test ./internal/services/ -run TestAttendance_Matrix -v`
Expected: all Plan A + original matrix tests PASS unchanged (pure refactor).

- [ ] **Step 4: Commit**

```bash
git add internal/services/attendance_matrix.go
git commit -m "refactor(attendance): extract buildAllRows for matrix + export reuse"
```

---

### Task 3: Export service — `ExportMatrix` + label/format helpers (G3, SR-009)

**Files:**
- Create: `internal/services/attendance_export.go`
- Test: `internal/services/attendance_export_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/services/attendance_export_test.go`:

```go
package services_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/exnodes/hrm-api/internal/dto"
)

func TestAttendance_Export_BulkProducesXlsx(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	mgr, _ := makeEmpUser(t, "exp-mgr@example.com", "Mgr")
	makeEmpUser(t, "exp-e1@example.com", "E1")

	data, err := svc.ExportMatrix(ctx, mgr.ID, true, dto.AttendanceMatrixQuery{Month: 5, Year: 2026}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// It must be a valid xlsx with a header row containing the summary cols.
	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close()
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	header := rows[0]
	assert.Equal(t, "Employee", header[0])
	assert.Equal(t, "Department", header[1])
	assert.Contains(t, header, "Total Late Time")
	assert.Contains(t, header, "Total Early Time")
}

func TestAttendance_Export_NonManagerSingle_DeniesOther(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	alice, _ := makeEmpUser(t, "exp-alice@example.com", "Alice")
	_, bobEmp := makeEmpUser(t, "exp-bob@example.com", "Bob")

	// Alice (non-admin) tries to export Bob → forbidden.
	_, err := svc.ExportMatrix(ctx, alice.ID, false, dto.AttendanceMatrixQuery{Month: 5, Year: 2026}, &bobEmp.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}
```

- [ ] **Step 2: Run — expect FAIL (method undefined)**

Run: `go test ./internal/services/ -run TestAttendance_Export -v`
Expected: compile error `svc.ExportMatrix undefined`.

- [ ] **Step 3: Implement the export service**

Create `internal/services/attendance_export.go`:

```go
package services

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
)

// statusLabel maps a cell status to its short Excel glyph (mirrors Python
// _STATUS_LABEL). Combined cells render the worked-half glyph alongside ½
// handled in the cell writer below.
var statusLabel = map[string]string{
	matrixOnTime:         "✓",
	matrixLate:           "L",
	matrixAbsent:         "A",
	matrixWeekend:        "—",
	matrixAnnualLeave:    "AL",
	matrixSickLeave:      "SL",
	matrixPersonalLeave:  "PL",
	matrixMaternityLeave: "ML",
	matrixUnpaidLeave:    "UL",
	matrixHalfDayLeave:   "½",
	matrixNoData:         "",
}

// formatHM renders integer minutes as "Xh Ym" (zero-padded minutes; negatives
// clipped to 0). Mirrors Python _format_hm. SR-011 display format.
func formatHM(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	return fmt.Sprintf("%dh %02dm", minutes/60, minutes%60)
}

// ExportMatrix builds an .xlsx of the monthly attendance matrix. When
// singleEmployeeID is non-nil, only that employee is exported (non-managers
// may export only themselves). Otherwise all visible employees are exported
// (admin → all; non-admin → self). Reuses buildAllRows so the summary columns
// match the on-screen matrix (SR-009 / AC-025).
func (s *AttendanceService) ExportMatrix(
	ctx context.Context,
	currentUserID uuid.UUID,
	asAdmin bool,
	q dto.AttendanceMatrixQuery,
	singleEmployeeID *uuid.UUID,
) ([]byte, error) {
	loc := s.tz()
	now, _ := todayInTZ(loc)
	year, month := q.Year, q.Month
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}

	var employees []models.Employee
	switch {
	case singleEmployeeID != nil:
		target, err := s.emps.FindByID(ctx, *singleEmployeeID)
		if err != nil {
			return nil, apperrors.ErrNotFound("Employee")
		}
		if !asAdmin {
			me, err := s.resolveCurrentEmployee(ctx, currentUserID)
			if err != nil {
				return nil, err
			}
			if me.ID != target.ID {
				return nil, apperrors.ErrForbidden("You do not have permission to export this employee's data")
			}
		}
		employees = []models.Employee{*target}
	case asAdmin:
		empQuery := dto.EmployeeListQuery{Page: 1, PageSize: 1000}
		rows, _, err := s.emps.List(ctx, empQuery)
		if err != nil {
			return nil, err
		}
		employees = rows
	default:
		me, err := s.resolveCurrentEmployee(ctx, currentUserID)
		if err != nil {
			return nil, err
		}
		employees = []models.Employee{*me}
	}

	rows, err := s.buildAllRows(ctx, employees, year, month, loc, nil)
	if err != nil {
		return nil, err
	}

	daysInMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc).AddDate(0, 1, -1).Day()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)

	// Header.
	header := []any{"Employee", "Department"}
	for d := 1; d <= daysInMonth; d++ {
		header = append(header, fmt.Sprintf("%d", d))
	}
	header = append(header, "Total Late Time", "Total Early Time")
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, err
	}

	// Data rows.
	for i, row := range rows {
		rec := []any{row.EmployeeName, deref(row.DepartmentName)}
		for d := 1; d <= daysInMonth; d++ {
			rec = append(rec, exportCellLabel(row.Cells[d], loc))
		}
		rec = append(rec, formatHM(row.TotalLateMinutes), formatHM(row.TotalEarlyMinutes))
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+2), &rec); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// exportCellLabel renders one cell as "GLYPH HH:MM-HH:MM" (times in company TZ
// when present), mirroring Python's data-row label build.
func exportCellLabel(cell dto.AttendanceCellRead, loc *time.Location) string {
	label := statusLabel[cell.Status]
	if cell.CheckIn != nil {
		ci := cell.CheckIn.In(loc).Format("15:04")
		if cell.CheckOut != nil {
			co := cell.CheckOut.In(loc).Format("15:04")
			return fmt.Sprintf("%s %s-%s", label, ci, co)
		}
		return fmt.Sprintf("%s %s", label, ci)
	}
	return label
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
```

- [ ] **Step 4: Run — expect PASS**

Run: `go test ./internal/services/ -run TestAttendance_Export -v`
Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/services/attendance_export.go internal/services/attendance_export_test.go
git commit -m "feat(attendance): xlsx ExportMatrix (bulk + single, summary cols) (G3, SR-009)"
```

---

### Task 4: Export HTTP handlers + routes (G3, AC-011/012)

**Files:**
- Modify: `internal/handlers/attendance_handler.go` (add `Export`, `ExportEmployee`)
- Modify: `cmd/server/main.go` (routes, gated `attendance:read`)

- [ ] **Step 1: Add the handlers**

Append to `attendance_handler.go`:

```go
// Export godoc
// @Summary      Export the monthly attendance matrix to Excel (all visible employees)
// @Tags         attendance
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        month  query  int  false  "1-12"
// @Param        year   query  int  false  "YYYY"
// @Success      200  {file}  binary
// @Router       /api/v1/attendance/export [get]
func (h *AttendanceHandler) Export(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	var q dto.AttendanceMatrixQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.ExportMatrix(c.Request.Context(), u.ID, hasAttendanceManageAll(c), q, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}
	writeXlsx(c, data, "attendance")
}

// ExportEmployee godoc
// @Summary      Export a single employee's monthly attendance to Excel
// @Tags         attendance
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        employee_id  path   string  true  "employee uuid"
// @Param        month        query  int     false  "1-12"
// @Param        year         query  int     false  "YYYY"
// @Success      200  {file}  binary
// @Router       /api/v1/attendance/export/{employee_id} [get]
func (h *AttendanceHandler) ExportEmployee(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	empID, err := parseIDParam(c, "employee_id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var q dto.AttendanceMatrixQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	data, err := h.svc.ExportMatrix(c.Request.Context(), u.ID, hasAttendanceManageAll(c), q, &empID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	writeXlsx(c, data, "attendance_"+empID.String())
}

// writeXlsx streams an xlsx byte slice as a download.
func writeXlsx(c *gin.Context, data []byte, basename string) {
	c.Header("Content-Disposition", `attachment; filename="`+basename+`.xlsx"`)
	c.Data(http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data)
}
```

> `parseIDParam` (used by `Get`/`AdminUpdate`) and `currentUser`/`hasAttendanceManageAll` already exist in this package. `http` is already imported.

- [ ] **Step 2: Wire the routes**

In `main.go`, in the attendance group (after the matrix/records routes from Plan A Task 8), add — gated `attendance:read`:

```go
		attendance.GET("/export", middleware.RequirePerms(authSvc, permissions.PermAttendanceRead), attendanceH.Export)
		attendance.GET("/export/:employee_id", middleware.RequirePerms(authSvc, permissions.PermAttendanceRead), attendanceH.ExportEmployee)
```

> Gin route conflict check: `/export` and `/export/:employee_id` are distinct path depths — no conflict with `GET ""`, `GET :id`, or `GET /matrix`. Confirm `:id` and `:employee_id` don't collide: `GET /:id` is a one-segment param and `/export/:employee_id` is two segments under the literal `export` — Gin handles these. If Gin panics on a wildcard conflict at boot, rename the `:id` param routes to share the same param name (`:id`) — they're already separate depths so this should not trigger.

- [ ] **Step 3: Build + boot smoke**

Run: `make vet`
Expected: clean. (A full boot test happens in Task 8.)

- [ ] **Step 4: Commit**

```bash
git add internal/handlers/attendance_handler.go cmd/server/main.go
git commit -m "feat(attendance): /export + /export/{employee_id} xlsx endpoints (G3, AC-011/012)"
```

---

### Task 5: D2 — check-in/out return `TodayStatusRead`

**Files:**
- Modify: `internal/handlers/attendance_handler.go` (`CheckIn`, `CheckOut`)

The mobile widget wants today-status back. The service already builds it (`Today`). Keep the service `CheckIn`/`CheckOut` signatures (they still perform the write + are used directly by service tests); the **handler** calls `Today` after the action and returns that — exactly mirroring Python's router (`check_in` then `get_today_status`).

- [ ] **Step 1: Rewrite the `CheckIn` handler tail**

Replace the body after binding in `CheckIn`:

```go
	if _, err := h.svc.CheckIn(c.Request.Context(), u.ID, req); err != nil {
		_ = c.Error(err)
		return
	}
	status, err := h.svc.Today(c.Request.Context(), u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.TodayStatusRead]{Success: true, Message: "Checked in", Data: status})
```

- [ ] **Step 2: Rewrite the `CheckOut` handler tail**

```go
	if _, err := h.svc.CheckOut(c.Request.Context(), u.ID, req); err != nil {
		_ = c.Error(err)
		return
	}
	status, err := h.svc.Today(c.Request.Context(), u.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.TodayStatusRead]{Success: true, Message: "Checked out", Data: status})
```

- [ ] **Step 3: Update the godoc `@Success` types**

Change both godoc blocks' `@Success` line to reference the today-status shape (cosmetic for Swagger):

```go
// @Success      200   {object}  dto.Response[dto.TodayStatusRead]
```

- [ ] **Step 4: Build**

Run: `make vet`
Expected: clean. Service tests are unaffected (they call `svc.CheckIn` directly and assert on the returned `AttendanceRead`, which the service still returns).

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/attendance_handler.go
git commit -m "feat(attendance): check-in/out return TodayStatusRead (D2, Python parity)"
```

---

### Task 6: D5 — stop auto-flipping `is_half_day`

**Files:**
- Modify: `internal/services/attendance_service.go` (`CheckOut`, ~line 274-276)
- Modify: `internal/services/attendance_service_test.go` (repurpose the short-day test)

Per the BA, half-day is approved-leave-driven, not hours-driven. Remove the hours-threshold flip in `CheckOut`.

- [ ] **Step 1: Remove the flip in `CheckOut`**

In `attendance_service.go` `CheckOut`, delete this block (~line 274-276):

```go
	if total > 0 && total < s.cfg.HalfDayHoursThreshold {
		reloaded.IsHalfDay = true
	}
```

Keep the `total` computation if it is still used for `HoursWorked`; if `total` becomes unused after deletion, remove the now-dead loop that computed it as well (the reload + `toRead` recompute hours independently). Verify `make vet` reports no unused variable.

- [ ] **Step 2: Repurpose the short-day test**

In `attendance_service_test.go`, change `TestAttendance_CheckOut_ShortDay_FlagsHalfDay` (lines ~140-156) to assert the NEW behavior — a short day no longer flags half-day:

```go
// D5: is_half_day is leave-driven, not hours-driven. A short worked day must
// NOT auto-flag half-day (the column is only set by approved half-day leave,
// surfaced in the matrix, not on the attendance row).
func TestAttendance_CheckOut_ShortDay_DoesNotFlagHalfDay(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "frank@example.com", "Frank")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := hcmTime(t, 2026, 5, 15, 11, 0) // 2.5h — short, but NOT half-day
	out, err := svc.CheckOut(ctx, u.ID, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)
	assert.False(t, out.IsHalfDay, "short worked day must not auto-flag half-day (D5)")
}
```

- [ ] **Step 3: Run the attendance service suite**

Run: `go test ./internal/services/ -run TestAttendance_CheckOut -v`
Expected: all check-out tests PASS, including the repurposed one. `TestAttendance_CheckOut_FullDay_NotHalfDay` still passes (already asserts false).

- [ ] **Step 4: Commit**

```bash
git add internal/services/attendance_service.go internal/services/attendance_service_test.go
git commit -m "fix(attendance): stop hours-based is_half_day auto-flip; half-day is leave-driven (D5)"
```

---

### Task 7: G5 — `AutoCheckOut` service + admin endpoint

**Files:**
- Modify: `internal/services/attendance_service.go` (add `AutoCheckOut`)
- Modify: `internal/handlers/attendance_handler.go` (add `AutoCheckOut` handler)
- Modify: `cmd/server/main.go` (route, gated `attendance:manage_data`)
- Test: `internal/services/attendance_service_test.go`

- [ ] **Step 1: Write the failing test**

Append to `attendance_service_test.go`:

```go
// G5: AutoCheckOut closes every open session whose check-in precedes the
// cutoff, marking it auto-checkout, and leaves already-closed sessions alone.
func TestAttendance_AutoCheckOut_ClosesOpenSessions(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc := newAttendanceSvc(t)
	u, _ := makeEmpUser(t, "autoout@example.com", "AutoOut")

	ci := hcmTime(t, 2026, 5, 15, 8, 30)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	cutoff := hcmTime(t, 2026, 5, 15, 23, 0) // 11 PM company time
	n, err := svc.AutoCheckOut(ctx, cutoff)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	// Re-running is idempotent — nothing open remains.
	n2, err := svc.AutoCheckOut(ctx, cutoff)
	require.NoError(t, err)
	assert.Equal(t, 0, n2)

	// The session is now closed + flagged.
	today, err := svc.Today(ctx, u.ID)
	require.NoError(t, err)
	// Today may not see this row (TZ-of-now), so assert via List instead.
	out, err := svc.List(ctx, u.ID, false, dto.AttendanceListQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Len(t, out.Items[0].Sessions, 1)
	require.NotNil(t, out.Items[0].Sessions[0].CheckOut)
	assert.True(t, out.Items[0].Sessions[0].IsAutoCheckout)
	_ = today
}
```

- [ ] **Step 2: Run — expect FAIL (method undefined)**

Run: `go test ./internal/services/ -run TestAttendance_AutoCheckOut -v`
Expected: compile error `svc.AutoCheckOut undefined`.

- [ ] **Step 3: Implement `AutoCheckOut`**

Add to `attendance_service.go`:

```go
// AutoCheckOut closes every open session whose check_in precedes cutoff,
// stamping check_out=cutoff and is_auto_checkout=true. Idempotent — a second
// run finds nothing open. Returns the number of sessions closed. Intended to
// run at 23:00 company time (mobile DR Rule 5); exposed via an admin endpoint
// now, a scheduler later.
func (s *AttendanceService) AutoCheckOut(ctx context.Context, cutoff time.Time) (int, error) {
	open, err := s.repo.OpenSessionsBefore(ctx, cutoff.UTC())
	if err != nil {
		return 0, err
	}
	closed := 0
	for i := range open {
		sess := open[i]
		co := cutoff.UTC()
		sess.CheckOut = &co
		sess.IsAutoCheckout = true
		if err := s.repo.UpdateSession(ctx, &sess); err != nil {
			return closed, err
		}
		closed++
	}
	return closed, nil
}
```

- [ ] **Step 4: Run — expect PASS**

Run: `go test ./internal/services/ -run TestAttendance_AutoCheckOut -v`
Expected: PASS.

- [ ] **Step 5: Add the admin handler**

Append to `attendance_handler.go`:

```go
// AutoCheckOut godoc
// @Summary      Admin: close all open sessions before a cutoff (auto check-out)
// @Description  Idempotent. Defaults the cutoff to now (company TZ) when omitted.
// @Tags         attendance
// @Security     BearerAuth
// @Produce      json
// @Param        cutoff  query  string  false  "RFC3339 cutoff; defaults to now"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/attendance/auto-checkout [post]
func (h *AttendanceHandler) AutoCheckOut(c *gin.Context) {
	cutoff := time.Now()
	if raw := c.Query("cutoff"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			_ = c.Error(apperrors.ErrBadRequest("invalid cutoff (expected RFC3339)"))
			return
		}
		cutoff = parsed
	}
	n, err := h.svc.AutoCheckOut(c.Request.Context(), cutoff)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[map[string]int]{Success: true, Message: "Auto check-out complete", Data: map[string]int{"closed": n}})
}
```

> Add `"time"` to the handler imports if not present.

- [ ] **Step 6: Wire the route**

In `main.go`, in the admin section of the attendance group (next to `POST ""`), add:

```go
		attendance.POST("/auto-checkout", middleware.RequirePerms(authSvc, permissions.PermAttendanceManage), attendanceH.AutoCheckOut)
```

- [ ] **Step 7: Build**

Run: `make vet`
Expected: clean.

- [ ] **Step 8: Commit**

```bash
git add internal/services/attendance_service.go internal/handlers/attendance_handler.go cmd/server/main.go internal/services/attendance_service_test.go
git commit -m "feat(attendance): AutoCheckOut service + admin trigger (G5, mobile AC-11/12)"
```

---

### Task 8: Swagger + full verification + checkpoint

**Files:**
- Modify: `docs/swagger/*` (regen)
- Modify: `docs/superpowers/CHECKPOINT.md`, `docs/superpowers/specs/2026-06-05-attendance-parity-audit.md`

- [ ] **Step 1: Regenerate Swagger**

Run: `make swag`
Expected: `docs/swagger/` includes `/attendance/export`, `/attendance/export/{employee_id}`, `/attendance/auto-checkout`, and the check-in/out `TodayStatusRead` responses.

- [ ] **Step 2: Format + vet + full suite**

Run: `make fmt && make vet && make test`
Expected: 0 fail, 0 skip (DB up via `make test-db-up`). New export + auto-checkout + D5 tests green; all prior attendance + Plan A tests green.

- [ ] **Step 3: No silent skips**

Run: `make test 2>&1 | grep -i skip`
Expected: none with DB up (AGENTS.md Rule 12).

- [ ] **Step 4: Boot smoke (route conflict guard)**

Run: `make build && PORT=8082 ./bin/server` (then `curl -s http://localhost:8082/health`), or rely on an existing boot test. Confirm the server boots without a Gin wildcard-conflict panic from the new `/export/:employee_id` route. Stop the server after.

- [ ] **Step 5: Update the audit spec + checkpoint**

In `2026-06-05-attendance-parity-audit.md`, mark G3, G5, D2, D5 as ✅ DONE (Plan B). In `CHECKPOINT.md`, add the "Attendance parity — Plan B (export/auto-checkout/response-shape)" entry: xlsx export (excelize), AutoCheckOut admin endpoint (scheduler still pending as a follow-up), check-in/out return TodayStatusRead, is_half_day auto-flip removed. Note remaining: G7 holidays (blocked), the real 23:00 scheduler trigger (follow-up), and the BA back-fill DR for admin CRUD (D6).

- [ ] **Step 6: Commit**

```bash
git add docs/swagger docs/superpowers/CHECKPOINT.md docs/superpowers/specs/2026-06-05-attendance-parity-audit.md
git commit -m "docs(attendance): close Plan B (export/jobs/response) in checkpoint + audit"
```

---

## Self-review checklist (run before handing off)

- **Spec coverage:** G3 (Tasks 3–4), D2 (Task 5), D5 (Task 6), G5 (Task 7). G7 out of scope (blocked); the real scheduler trigger is a tracked follow-up, not this plan. ✅
- **Plan A dependency:** `buildAllRows` (Task 2) must exist; export reuses it so summary columns match the matrix (AC-025). If Plan A merged, Task 2 is a no-op verify. ✅
- **Type consistency:** `ExportMatrix(ctx, currentUserID, asAdmin, q, *uuid.UUID) ([]byte, error)`; `AutoCheckOut(ctx, cutoff) (int, error)`; `statusLabel`/`matrix*` consts come from Plan A's `attendance_matrix.go`. The handler uses existing `currentUser`, `hasAttendanceManageAll`, `parseIDParam`. ✅
- **No schema change:** Plan B adds no migration (uses existing `is_auto_checkout` column + `OpenSessionsBefore`). Migration version stays 21. ✅
- **Dead-code check:** after removing the `is_half_day` flip (Task 6), confirm `total`/its loop is removed if unused, and `HalfDayHoursThreshold` may become unreferenced in the service — leave the config field (other code/env may read it) but confirm `make vet` is clean.
- **Placeholder scan:** the only intentionally-narrative block is `buildAllRows`'s body in Task 2 Step 1 (a relocation of existing code, described as a move in Step 2) — not new logic. All new logic (export, auto-checkout, handlers) is shown in full.
