# Leave Requests Parity — Plan B: Bug Fixes + Excel Export

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development`
> (recommended) or `superpowers:executing-plans` to implement task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.
>
> **⚠️ REVISION NOTES — read before task bodies:**
> None yet. This plan is new.

**Goal:** Fix five correctness bugs (empty-PATCH revert, DOCX MIME detection, 5 MB attachment
limit, dashboard limit, cancel wasApproved response) and add an Excel export endpoint for
leave requests (resolves G3, G4, G5, G6, G7, G10 from the 2026-06-08 parity audit).

**Architecture:** All changes are in the service and handler layers; no DB migration required.
The Excel export follows the established `attendance_export.go` pattern — a new
`internal/services/leave_export.go` file, a `writeLeaveXlsx` helper in the handler, and a
new GET route `/api/v1/leave-requests/export`. Bug fixes are surgical edits to existing
functions in `leave_service.go` and `leave_handler.go`.

**Tech Stack:** Go + Gin + GORM; `excelize/v2` (already in go.mod from attendance parity);
`internal/services/leave_service.go`, `internal/services/leave_export.go` (new),
`internal/handlers/leave_handler.go`, `cmd/server/main.go`,
`internal/services/leave_service_test.go`.

---

## File map

| File | Change |
|---|---|
| `internal/services/leave_service.go` | G3: no-op guard in Update; G6: dashboard limit 5→10; G7: Cancel returns wasApproved; G4+G5: MIME/size fixes in uploadAttachment |
| `internal/services/leave_export.go` | New: ExportLeave service method |
| `internal/handlers/leave_handler.go` | G5: maxLeaveAttachmentBytes 10→5 MB; G7: Cancel handler (out, wasApproved, err); G10: Export handler + writeLeaveXlsx |
| `cmd/server/main.go` | G10: new export route |
| `internal/services/leave_service_test.go` | Add bug-fix regression tests |

---

## Task 1: Fix G5 — Attachment size limit 10 MB → 5 MB

**Files:**
- Modify: `internal/services/leave_service.go` (lines ~24, ~153–176)
- Modify: `internal/handlers/leave_handler.go` (lines ~36, ~72–85)

- [ ] **Step 1: Update the service constant**

In `internal/services/leave_service.go`, line ~24:

```go
// old
leaveAttachmentMaxBytes = 10 * 1024 * 1024
// new
leaveAttachmentMaxBytes = 5 * 1024 * 1024
```

- [ ] **Step 2: Update the handler constant**

In `internal/handlers/leave_handler.go`, line ~36:

```go
// old
maxLeaveAttachmentBytes = 10 * 1024 * 1024
// new
maxLeaveAttachmentBytes = 5 * 1024 * 1024
```

- [ ] **Step 3: Update error messages that reference "10MB"**

In `internal/handlers/leave_handler.go`, lines ~72–85, change the error message from
`"Attachment too large (max 10MB)"` (or similar wording) to `"Attachment too large (max 5MB)"`.
Do the same in `internal/services/leave_service.go` if the service also produces a message
containing "10MB".

- [ ] **Step 4: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/... ./internal/handlers/...
```

Expected: no output (success).

---

## Task 2: Fix G4 — DOCX MIME detection (ZIP sniff + extension fallback)

**Files:**
- Modify: `internal/services/leave_service.go` (function `uploadAttachment`, lines ~153–176)

**Background:** `http.DetectContentType` sniffs the first 512 bytes of a DOCX file and
returns `"application/zip"` because DOCX is a ZIP container. Python's `python-magic` reads
the full magic bytes and correctly identifies DOCX. The Go fix: after the content-type sniff,
if the detected type is `"application/zip"` AND the original filename extension is `.docx`,
treat it as `"application/vnd.openxmlformats-officedocument.wordprocessingml.document"`.

- [ ] **Step 1: Read the current uploadAttachment function**

Read lines 153–176 of `internal/services/leave_service.go` to confirm exact variable names
and the structure of the MIME check block.

- [ ] **Step 2: Add DOCX extension-fallback after the sniff call**

Locate the line that calls `http.DetectContentType(buf)` and the subsequent MIME allowlist
check. Insert between them:

```go
const docxMIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
// http.DetectContentType returns "application/zip" for DOCX (ZIP container).
// Fall back to the extension when the filename explicitly declares .docx.
if sniffed == "application/zip" && strings.EqualFold(filepath.Ext(header.Filename), ".docx") {
    sniffed = docxMIME
}
```

Also ensure `docxMIME` (or the literal string) is in `allowedAttachmentMIME`:

```go
var allowedAttachmentMIME = map[string]string{
    "application/pdf":                                                      ".pdf",
    "image/jpeg":                                                           ".jpg",
    "image/png":                                                            ".png",
    docxMIME: ".docx",
}
```

The constant `docxMIME` can be defined as a `const` at package level (near
`leaveAttachmentMaxBytes`) or inlined as the literal string in the map.

- [ ] **Step 3: Ensure needed imports are present**

The fix needs `path/filepath` and `strings`. Add them to the import block if not already
present.

- [ ] **Step 4: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/...
```

Expected: no output.

---

## Task 3: Fix G3 — Empty PATCH no-op guard (Approved→Pending revert bug)

**Files:**
- Modify: `internal/services/leave_service.go` (function `Update`, line ~461–463)

**Background:** When a client sends an empty PATCH body (all fields nil), Go's `Update`
unconditionally resets the row status to `pending`. The fix: detect "no field was set" and
return the existing row unchanged, skipping the status transition.

- [ ] **Step 1: Read the Update function body**

Read `internal/services/leave_service.go` lines 440–545 to confirm the exact location of
the ownership check and the subsequent status-transition logic.

- [ ] **Step 2: Determine field variable names**

Locate where `in.FromDate`, `in.ToDate`, `in.LeavePeriod`, `in.LeaveType`, `in.Reason` are
applied to the row, and where `attachment` (the uploaded file) is passed. The
`hasChanges` check must be inserted AFTER the ownership check but BEFORE any of these
fields are written.

- [ ] **Step 3: Insert the no-op guard**

After the ownership-check block (approximately line 461, just before the status switch or
field-assignment block), insert:

```go
// Guard: if no fields were provided, return the existing row unchanged.
// Without this, an empty PATCH body resets an approved request back to pending
// via the status-transition logic below (G3 fix).
hasChanges := in.FromDate != nil || in.ToDate != nil || in.LeavePeriod != nil ||
    in.LeaveType != nil || in.Reason != nil || attachment != nil
if !hasChanges {
    read, err := s.populateRead(ctx, row)
    if err != nil {
        return nil, err
    }
    return &dto.LeaveRequestWriteResult{Request: read, Warnings: []string{}}, nil
}
```

**Important:** The variable names (`in`, `attachment`, `row`, `populateRead`) must match
what's already used in the function. Read the function body first and adjust accordingly.

- [ ] **Step 4: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/...
```

Expected: no output.

---

## Task 4: Fix G6 — GetMyDashboard hardcoded limit 5 → 10

**Files:**
- Modify: `internal/services/leave_service.go` (function `GetMyDashboard`, lines ~772, ~776)

- [ ] **Step 1: Read GetMyDashboard**

Read the function body to confirm the exact integer literals used for `Upcoming` and
`History` calls (both are `5`).

- [ ] **Step 2: Replace both limits**

Change both `5` to `10`:

```go
upcoming, err := s.leaves.Upcoming(ctx, emp.ID, now, 10)
// ...
history, err := s.leaves.History(ctx, emp.ID, now, 10)
```

If a constant already controls this, update the constant instead.

- [ ] **Step 3: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/...
```

Expected: no output.

---

## Task 5: Fix G7 — Cancel returns wasApproved in response

**Files:**
- Modify: `internal/services/leave_service.go` (function `Cancel`, lines ~563–579)
- Modify: `internal/handlers/leave_handler.go` (function `Cancel`, lines ~386–402)

**Background:** Python's cancel endpoint sets `was_approved` in the response body. Go's
Cancel returns only `(*dto.LeaveRequestRead, error)`. The fix: change the return signature
to `(*dto.LeaveRequestRead, bool, error)` and capture the pre-cancel status.

- [ ] **Step 1: Read the Cancel function in leave_service.go**

Read lines 563–579 of `internal/services/leave_service.go` to see the exact current
signature and body.

- [ ] **Step 2: Update the Cancel service signature**

Old signature:
```go
func (s *LeaveService) Cancel(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (*dto.LeaveRequestRead, error)
```

New signature:
```go
func (s *LeaveService) Cancel(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, asAdmin bool) (*dto.LeaveRequestRead, bool, error)
```

- [ ] **Step 3: Capture wasApproved before transitioning**

In the function body, before the `transitionStatus(...)` call (or however Cancel triggers
the status change), add:

```go
wasApproved := row.Status == models.LeaveStatusApproved
```

And update all return sites:

```go
// error returns become:
return nil, false, err
// success:
return read, wasApproved, nil
```

- [ ] **Step 4: Update the Cancel handler to use the new signature**

In `internal/handlers/leave_handler.go`, function `Cancel` (lines ~386–402), change:

```go
out, err := h.svc.Cancel(c.Request.Context(), id, u.ID, asAdmin)
if err != nil {
    _ = c.Error(err)
    return
}
c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestRead]{Success: true, Data: out})
```

to:

```go
out, wasApproved, err := h.svc.Cancel(c.Request.Context(), id, u.ID, asAdmin)
if err != nil {
    _ = c.Error(err)
    return
}
// was_approved signals to the caller that quota must be adjusted (G7).
type cancelResult struct {
    *dto.LeaveRequestRead
    WasApproved bool `json:"was_approved"`
}
c.JSON(http.StatusOK, dto.Response[cancelResult]{
    Success: true,
    Data:    cancelResult{LeaveRequestRead: out, WasApproved: wasApproved},
})
```

- [ ] **Step 5: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/... ./internal/handlers/...
```

Expected: no output.

---

## Task 6: Add G10 — Excel export (leave_export.go + Export handler)

**Files:**
- Create: `internal/services/leave_export.go`
- Modify: `internal/handlers/leave_handler.go` (add `Export` handler + `writeLeaveXlsx`)
- Modify: `cmd/server/main.go` (add export route)

### Sub-task 6a: Create leave_export.go

- [ ] **Step 1: Create the service export file**

Create `internal/services/leave_export.go`:

```go
package services

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/exnodes/hrm-api/internal/dto"
)

// ExportLeave builds an .xlsx of leave requests matching the query. When
// asAdmin is true, all employees are visible; otherwise the export is scoped
// to the current user's own requests. Mirrors Python's leave export endpoint
// (G10 from the 2026-06-08 parity audit).
func (s *LeaveService) ExportLeave(
	ctx context.Context,
	currentUserID uuid.UUID,
	asAdmin bool,
	q dto.LeaveListQuery,
) ([]byte, error) {
	// Cap at a reasonable page size for single-call export. If a future
	// requirement needs true streaming, replace with a repo-level ListAll.
	q.Page = 1
	q.PageSize = 5000

	result, err := s.GetList(ctx, currentUserID, asAdmin, q)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)

	header := []any{
		"Employee", "Department", "Position",
		"Leave Type", "From Date", "To Date", "Period", "Total Days",
		"Reason", "Status", "Created At",
	}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, err
	}

	for i, lr := range result.Items {
		empName := ""
		deptName := ""
		posName := ""
		if lr.Employee != nil {
			empName = lr.Employee.Name
		}
		if lr.Department != nil {
			deptName = lr.Department.Name
		}
		if lr.Position != nil {
			posName = lr.Position.Name
		}
		row := []any{
			empName,
			deptName,
			posName,
			string(lr.LeaveType),
			lr.FromDate.Format("2006-01-02"),
			lr.ToDate.Format("2006-01-02"),
			string(lr.LeavePeriod),
			fmt.Sprintf("%.1f", lr.TotalDays),
			lr.Reason,
			string(lr.Status),
			lr.CreatedAt.Format("2006-01-02 15:04"),
		}
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+2), &row); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

**Important: verify field names against `dto.LeaveRequestRead`.**
Known fields: `ID string`, `Employee *LeaveRefRead` (with `Name string`),
`Department *LeaveRefRead`, `Position *LeaveRefRead`, `FromDate time.Time`, `ToDate time.Time`,
`LeavePeriod`, `LeaveType`, `TotalDays float64`, `Reason string`, `AttachmentURL string`,
`Status`, `CreatedBy`, `CreatedAt time.Time`, `UpdatedAt time.Time`.

- [ ] **Step 2: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/...
```

Expected: no output.

### Sub-task 6b: Add Export handler and writeLeaveXlsx to leave_handler.go

- [ ] **Step 3: Add writeLeaveXlsx helper (near writeXlsx attendance pattern)**

At the end of `internal/handlers/leave_handler.go`, add:

```go
// writeLeaveXlsx writes buf to the response as an xlsx download. Mirrors the
// attendance handler's writeXlsx (G10).
func writeLeaveXlsx(c *gin.Context, data []byte, basename string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", basename+".xlsx"))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
```

Ensure `fmt` is in the import block (it almost certainly is already).

- [ ] **Step 4: Add Export handler**

Add the Export handler to `leave_handler.go`:

```go
// Export godoc
// @Summary      Export leave requests as Excel
// @Description  Returns an .xlsx download of leave requests matching the same filters as the list endpoint. Requires leave_requests:read.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        status    query  []string false "filter by status"
// @Param        department_id query string false "filter by department UUID"
// @Param        position_id   query string false "filter by position UUID"
// @Param        search        query string false "search employee name"
// @Success      200 {string} binary
// @Failure      403 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/export [get]
func (h *LeaveHandler) Export(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	asAdmin := hasLeaveManageAll(c)

	var q dto.LeaveListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrValidation(err.Error()))
		return
	}

	data, err := h.svc.ExportLeave(c.Request.Context(), u.ID, asAdmin, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	writeLeaveXlsx(c, data, "leave-requests")
}
```

Ensure `apperrors "github.com/exnodes/hrm-api/internal/errors"` is in the import block
(it should be already).

- [ ] **Step 5: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/handlers/...
```

Expected: no output.

### Sub-task 6c: Register the export route

- [ ] **Step 6: Add export route in cmd/server/main.go**

Locate the leave-requests route block (~lines 290–305). Add the export route BEFORE the
`:id` parameter routes to avoid Gin treating "export" as an ID:

```go
// Export must come before :id routes to avoid "export" being parsed as a UUID.
leaves.GET("export", middleware.RequirePerms(authSvc, permissions.PermLeaveRead), leaveH.Export)
// ... existing :id routes ...
```

- [ ] **Step 7: Verify full binary builds**

```bash
cd e:\Work\Go-HRM && go build ./...
```

Expected: no output.

---

## Task 7: Regression tests for all bug fixes

**Files:**
- Modify: `internal/services/leave_service_test.go`

- [ ] **Step 1: Add TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus (G3)**

Add to `internal/services/leave_service_test.go`:

```go
// TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus verifies that sending an
// empty UpdateLeaveRequest (all nil fields) does not reset an Approved request
// back to Pending. This was a Go-specific regression not present in Python
// because Python's Pydantic model excludes unset fields from the DB write.
func TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus(t *testing.T) {
    skipIfNoDB(t)
    truncateAll(t)
    ctx := context.Background()

    svc, leavesRepo, _ := newLeaveSvc(t, nil)
    _, emp := makeEmpUser(t, "emp-noop@x.com", "Emp NoOp")
    makeLeaveQuota(t, emp.ID, 10, 5)

    // Create a leave request and manually set it to Approved in DB.
    createRes, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
        FromDate:    dateAt(2026, 9, 1),
        ToDate:      dateAt(2026, 9, 1),
        LeavePeriod: models.LeavePeriodFullDay,
        LeaveType:   models.LeaveTypePersonal,
        Reason:      "noop test",
    }, nil)
    require.NoError(t, err)

    id := uuid.MustParse(createRes.Request.ID)
    require.NoError(t, leavesRepo.Update(ctx, &models.LeaveRequest{
        BaseModel: models.BaseModel{ID: id},
        Status:    models.LeaveStatusApproved,
    }))

    // Send an empty PATCH — no fields set.
    updateRes, err := svc.Update(ctx, id, emp.UserID, false, dto.UpdateLeaveRequest{}, nil)
    require.NoError(t, err, "empty PATCH must not error")
    assert.Equal(t, string(models.LeaveStatusApproved), string(updateRes.Request.Status),
        "empty PATCH must not revert Approved→Pending (G3 regression)")
}
```

- [ ] **Step 2: Add TestCancel_WasApprovedTrue (G7)**

```go
// TestCancel_WasApprovedTrue verifies that cancelling an Approved request
// returns wasApproved=true, so callers know quota must be restored (G7).
func TestCancel_WasApprovedTrue(t *testing.T) {
    skipIfNoDB(t)
    truncateAll(t)
    ctx := context.Background()

    svc, leavesRepo, _ := newLeaveSvc(t, nil)
    _, emp := makeEmpUser(t, "emp-cancel-was@x.com", "Emp CancelWas")
    makeLeaveQuota(t, emp.ID, 10, 5)

    createRes, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
        FromDate:    dateAt(2026, 10, 1),
        ToDate:      dateAt(2026, 10, 1),
        LeavePeriod: models.LeavePeriodFullDay,
        LeaveType:   models.LeaveTypePersonal,
        Reason:      "cancel was approved",
    }, nil)
    require.NoError(t, err)

    id := uuid.MustParse(createRes.Request.ID)
    require.NoError(t, leavesRepo.Update(ctx, &models.LeaveRequest{
        BaseModel: models.BaseModel{ID: id},
        Status:    models.LeaveStatusApproved,
    }))

    _, wasApproved, err := svc.Cancel(ctx, id, emp.UserID, false)
    require.NoError(t, err)
    assert.True(t, wasApproved, "cancelling an Approved request must return wasApproved=true (G7)")
}

// TestCancel_WasApprovedFalse verifies that cancelling a Pending request
// returns wasApproved=false (no quota adjustment needed).
func TestCancel_WasApprovedFalse(t *testing.T) {
    skipIfNoDB(t)
    truncateAll(t)
    ctx := context.Background()

    svc, _, _ := newLeaveSvc(t, nil)
    _, emp := makeEmpUser(t, "emp-cancel-pending@x.com", "Emp CancelPending")
    makeLeaveQuota(t, emp.ID, 10, 5)

    createRes, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
        FromDate:    dateAt(2026, 10, 5),
        ToDate:      dateAt(2026, 10, 5),
        LeavePeriod: models.LeavePeriodFullDay,
        LeaveType:   models.LeaveTypePersonal,
        Reason:      "cancel pending",
    }, nil)
    require.NoError(t, err)

    id := uuid.MustParse(createRes.Request.ID)
    _, wasApproved, err := svc.Cancel(ctx, id, emp.UserID, false)
    require.NoError(t, err)
    assert.False(t, wasApproved, "cancelling a Pending request must return wasApproved=false")
}
```

**Important:** The `leavesRepo.Update` call sets only `Status`. Check whether the GORM
`Update(ctx, &models.LeaveRequest{...})` with partial struct correctly patches only the
status field, or if you need `testDB.Model(...).Where("id = ?", id).Update("status", ...)`.
Use whichever pattern the test file already uses for similar state-mutations.

- [ ] **Step 3: Run the new tests**

```bash
cd e:\Work\Go-HRM && go test ./internal/services/... -run "TestUpdate_EmptyPatch|TestCancel_Was" -v
```

Expected: all tests PASS (or SKIP if no test DB).

---

## Task 8: Final checks, Swagger regen, and commit

- [ ] **Step 1: Format and vet**

```bash
cd e:\Work\Go-HRM && make fmt && make vet
```

Expected: no output / no errors.

- [ ] **Step 2: Full test suite**

```bash
cd e:\Work\Go-HRM && make test
```

Expected: all tests pass (or DB-dependent tests skip cleanly).

- [ ] **Step 3: Regenerate Swagger (handler annotations changed)**

```bash
cd e:\Work\Go-HRM && make swag
```

Expected: no errors; `docs/swagger/` updated.

- [ ] **Step 4: Commit**

```bash
git add internal/services/leave_service.go \
        internal/services/leave_export.go \
        internal/handlers/leave_handler.go \
        cmd/server/main.go \
        internal/services/leave_service_test.go \
        docs/swagger/
git commit -m "fix+feat(leave): G3/G4/G5/G6/G7/G10 parity fixes + Excel export

Fixes:
  G3: empty PATCH no longer reverts approved→pending (hasChanges guard)
  G4: DOCX MIME detection via extension fallback after ZIP sniff
  G5: attachment size limit 10MB→5MB (service + handler)
  G6: GetMyDashboard upcoming/history limit 5→10
  G7: Cancel returns (read, wasApproved, err) so callers know to restore quota

Feature:
  G10: GET /api/v1/leave-requests/export → xlsx via excelize (same pattern
       as attendance matrix export; leave_requests:read required)

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Self-Review

**Spec coverage check:**
- G3 (empty PATCH revert) → Task 3 (no-op guard) + Task 7 regression test ✅
- G4 (DOCX MIME detection) → Task 2 (ZIP sniff + extension fallback) ✅
- G5 (5 MB limit) → Task 1 (constant + error message) ✅
- G6 (dashboard limit 5→10) → Task 4 ✅
- G7 (cancel wasApproved) → Task 5 (signature + handler) + Task 7 regression test ✅
- G10 (Excel export) → Task 6 (leave_export.go + handler + route) ✅

**Placeholder scan:** No TBD or "similar to" references in code blocks.
All function signatures reference exact types verified in the codebase:
- `dto.LeaveListQuery` — confirmed in `internal/dto/leave.go`
- `dto.LeaveRequestRead.Employee *LeaveRefRead` / `.Department` / `.Position` — confirmed
- `models.LeaveStatusApproved` / `LeaveStatusPending` / `LeaveStatusCancelled` — confirmed
- `s.GetList(ctx, currentUserID, asAdmin, q)` — confirmed in leave_service.go
- `leavesRepo.Update(ctx, *models.LeaveRequest)` — confirmed in leave_request_repo.go

**Type consistency check:**
- `ExportLeave` in `leave_export.go` calls `s.GetList` with the same `dto.LeaveListQuery`
  signature the handler already uses. ✅
- `Cancel` new signature `(*dto.LeaveRequestRead, bool, error)` — both service and handler
  updated in the same task (Task 5). ✅
- `writeLeaveXlsx(c, data, basename)` — `data []byte` matches `ExportLeave` return type. ✅

**Route ordering note (critical):**
`leaves.GET("export", ...)` MUST be registered BEFORE any `leaves.GET(":id", ...)` route,
otherwise Gin matches the literal string "export" as a UUID parameter. Explicitly called out
in Task 6, Step 6. ✅
