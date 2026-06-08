# Leave Requests Parity — Plan A: Permission Split + Manager Seed

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development`
> (recommended) or `superpowers:executing-plans` to implement task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.
>
> **⚠️ REVISION NOTES — read before task bodies:**
> None yet. This plan is new.

**Goal:** Add `leave_requests:approve_team` and `leave_requests:approve_all` to the
permission registry; wire team-scoped BFS authority check into the Approve/Reject service
methods; update seed so Admin/HR Manager → approve_all, Manager → approve_team + update +
delete (resolves G1, G2, G8 from the 2026-06-08 parity audit).

**Architecture:** New `ApproveScope` type + `checkCanApproveOrReject` in the service layer
(mirrors Python's `check_can_approve_or_reject`). Handler pre-computes scope from JWT claims
via `resolveApproveScope`; service enforces it. Reuses existing `emps.SubordinateIDs`
(employee-parity BFS — already in `repositories.EmployeeRepository`). No DB migration
required — only permission strings change in the seed.

**Tech Stack:** Go + Gin + GORM; `internal/permissions/registry.go`,
`internal/services/leave_service.go`, `internal/handlers/leave_handler.go`,
`cmd/server/main.go`, `internal/services/seed_service.go`;
test target: `internal/services/leave_approve_test.go`

---

## File map

| File | Change |
|---|---|
| `internal/permissions/registry.go` | Add 2 new perm constants; add to `AllPermissions()`; update `PermissionGroups` catalog |
| `internal/services/leave_service.go` | Add `ApproveScope` type; add `checkCanApproveOrReject`; update `Approve`/`Reject` signatures |
| `internal/handlers/leave_handler.go` | Add `resolveApproveScope` helper; update `Approve`/`Reject` handlers |
| `cmd/server/main.go` | Remove `RequirePerms` gate from approve/reject routes (perm check moves to handler) |
| `internal/services/seed_service.go` | Update Admin/HR Manager → approve_all; Manager → approve_team + update + delete |
| `internal/services/leave_approve_test.go` | New: approve scope integration tests |

---

## Task 1: Add new permission constants to registry

**Files:**
- Modify: `internal/permissions/registry.go`

- [ ] **Step 1: Read the current constant block**

Read lines 63–87 of `internal/permissions/registry.go` to locate the Leave Requests section.

- [ ] **Step 2: Add PermLeaveApproveTeam and PermLeaveApproveAll**

In `internal/permissions/registry.go`, after `PermLeaveApprove` (line ~68), insert:

```go
// Leave Requests
PermLeaveRead    Permission = "leave_requests:read"
PermLeaveCreate  Permission = "leave_requests:create"
PermLeaveUpdate  Permission = "leave_requests:update"
PermLeaveDelete  Permission = "leave_requests:delete"
PermLeaveApprove     Permission = "leave_requests:approve"      // legacy; treat as approve_all at runtime
PermLeaveApproveTeam Permission = "leave_requests:approve_team" // approve own subordinate chain only (BFS)
PermLeaveApproveAll  Permission = "leave_requests:approve_all"  // approve any employee's request
PermLeaveCancel  Permission = "leave_requests:cancel"
PermLeaveManage  Permission = "leave_requests:manage"
```

- [ ] **Step 3: Add new perms to AllPermissions()**

In `AllPermissions()` (line ~102), extend the Leave line:

```go
PermLeaveRead, PermLeaveCreate, PermLeaveUpdate, PermLeaveDelete,
PermLeaveApprove, PermLeaveApproveTeam, PermLeaveApproveAll, PermLeaveCancel, PermLeaveManage,
```

- [ ] **Step 4: Update PermissionGroups catalog**

Locate the `leave_requests` block in `PermissionGroups` (~line 215). Replace the single
`PermLeaveApprove` catalog entry with two scoped entries. Remove `PermLeaveApprove` from
the catalog (it stays as a constant for runtime backward compat, but new UI assignments
should only pick team or all):

```go
{
    Resource: "leave_requests", Label: "Leave Requests",
    Permissions: []PermissionItem{
        {PermLeaveRead, "View Leave Requests", "List and view leave requests"},
        {PermLeaveCreate, "Create Leave Requests", "Submit leave requests"},
        {PermLeaveUpdate, "Edit Leave Requests", "Update leave request details"},
        {PermLeaveDelete, "Delete Leave Requests", "Soft-delete leave requests"},
        {PermLeaveApproveTeam, "Approve Own Team's Requests", "Approve or reject leave requests from employees in your direct reporting chain"},
        {PermLeaveApproveAll, "Approve All Leave Requests", "Approve or reject any employee's leave request regardless of reporting line"},
        {PermLeaveCancel, "Cancel Leave Requests", "Cancel pending or approved leave requests"},
        {PermLeaveManage, "Manage Others' Leave Requests", "Create, edit, and view leave requests on behalf of other employees"},
    },
},
```

- [ ] **Step 5: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/permissions/...
```

Expected: no output (success).

---

## Task 2: Add ApproveScope type and checkCanApproveOrReject to leave_service.go

**Files:**
- Modify: `internal/services/leave_service.go`

- [ ] **Step 1: Add ApproveScope type after the import block**

Add the following after the `import (...)` block and before the first `const` block in
`internal/services/leave_service.go`:

```go
// ApproveScope is the approval authority pre-computed by the handler from JWT
// claims and passed into Approve/Reject so the service enforces scope without
// importing gin.
type ApproveScope int

const (
	ApproveScopeTeam ApproveScope = 1 // approve_team: BFS subordinate chain only
	ApproveScopeAll  ApproveScope = 2 // approve_all, legacy approve, or wildcard *
)
```

- [ ] **Step 2: Replace Approve method (line ~549)**

Current code:
```go
func (s *LeaveService) Approve(ctx context.Context, id uuid.UUID) (*dto.LeaveRequestRead, error) {
	return s.transitionStatus(ctx, id, models.LeaveStatusApproved, []models.LeaveStatus{models.LeaveStatusPending})
}
```

Replace with:
```go
// Approve transitions a pending request to approved. The handler resolves the
// approve scope from JWT claims and passes it here; we enforce the BFS
// subordinate-chain restriction when scope == ApproveScopeTeam.
func (s *LeaveService) Approve(ctx context.Context, id uuid.UUID, approverUserID uuid.UUID, scope ApproveScope) (*dto.LeaveRequestRead, error) {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Leave request")
		}
		return nil, err
	}
	if err := s.checkCanApproveOrReject(ctx, approverUserID, scope, row); err != nil {
		return nil, err
	}
	return s.transitionStatus(ctx, id, models.LeaveStatusApproved, []models.LeaveStatus{models.LeaveStatusPending})
}
```

- [ ] **Step 3: Replace Reject method (line ~556)**

Current code:
```go
func (s *LeaveService) Reject(ctx context.Context, id uuid.UUID) (*dto.LeaveRequestRead, error) {
	return s.transitionStatus(ctx, id, models.LeaveStatusRejected, []models.LeaveStatus{models.LeaveStatusPending})
}
```

Replace with:
```go
// Reject transitions a pending request to rejected. Same scope check as Approve.
func (s *LeaveService) Reject(ctx context.Context, id uuid.UUID, approverUserID uuid.UUID, scope ApproveScope) (*dto.LeaveRequestRead, error) {
	row, err := s.leaves.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("Leave request")
		}
		return nil, err
	}
	if err := s.checkCanApproveOrReject(ctx, approverUserID, scope, row); err != nil {
		return nil, err
	}
	return s.transitionStatus(ctx, id, models.LeaveStatusRejected, []models.LeaveStatus{models.LeaveStatusPending})
}
```

- [ ] **Step 4: Add checkCanApproveOrReject after Reject (before Cancel)**

Insert this function between Reject and Cancel in `leave_service.go`:

```go
// checkCanApproveOrReject enforces approval authority.
//   - ApproveScopeAll: any request is allowed (no further check).
//   - ApproveScopeTeam: lr.EmployeeID must be in the approver's transitive
//     subordinate chain (BFS via line_manager_id). Reuses SubordinateIDs from
//     the employee-parity line-manager work.
func (s *LeaveService) checkCanApproveOrReject(ctx context.Context, approverUserID uuid.UUID, scope ApproveScope, lr *models.LeaveRequest) error {
	if scope == ApproveScopeAll {
		return nil
	}
	approverEmp, err := s.emps.FindByUserID(ctx, approverUserID)
	if err != nil {
		return apperrors.ErrForbidden("Approver has no employee record")
	}
	subordinates, err := s.emps.SubordinateIDs(ctx, approverEmp.ID)
	if err != nil {
		return err
	}
	if !subordinates[lr.EmployeeID] {
		return apperrors.ErrForbidden("You can only approve leave requests for employees in your reporting chain")
	}
	return nil
}
```

- [ ] **Step 5: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/...
```

Expected: compilation error about the handler calling `svc.Approve(ctx, id)` with the old
signature. Fix that in Task 3. Here we just verify no other compile errors.

---

## Task 3: Add resolveApproveScope helper and update Approve/Reject handlers

**Files:**
- Modify: `internal/handlers/leave_handler.go`

- [ ] **Step 1: Add resolveApproveScope helper alongside hasLeaveManageAll**

After the closing `}` of `hasLeaveManageAll` (line ~53), insert:

```go
// resolveApproveScope inspects JWT-preloaded user roles for an approve perm.
// Returns (scope, true) when any approve variant is found; (0, false) otherwise.
// Priority: wildcard / approve_all / legacy approve → ApproveScopeAll;
//           approve_team → ApproveScopeTeam (weaker; keep scanning for all).
func resolveApproveScope(c *gin.Context) (services.ApproveScope, bool) {
	u, okC := currentUser(c)
	if !okC {
		return 0, false
	}
	var found services.ApproveScope
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			switch permissions.Permission(p) {
			case permissions.PermAll, permissions.PermLeaveApproveAll, permissions.PermLeaveApprove:
				return services.ApproveScopeAll, true // strongest; short-circuit
			case permissions.PermLeaveApproveTeam:
				found = services.ApproveScopeTeam // weaker; keep scanning for all
			}
		}
	}
	if found != 0 {
		return found, true
	}
	return 0, false
}
```

- [ ] **Step 2: Replace Approve handler (lines ~341–353)**

Current handler code:
```go
func (h *LeaveHandler) Approve(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Approve(c.Request.Context(), id)
	...
```

Replace with:
```go
// Approve godoc
// @Summary      Approve a pending leave request
// @Description  Requires approve_team (own subordinate chain only) or approve_all (any employee). Permission is checked in the handler; the service enforces BFS scope.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "leave request uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/{id}/approve [post]
func (h *LeaveHandler) Approve(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	scope, ok := resolveApproveScope(c)
	if !ok {
		_ = c.Error(apperrors.ErrForbidden("Insufficient approve permission"))
		return
	}
	out, err := h.svc.Approve(c.Request.Context(), id, u.ID, scope)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestRead]{Success: true, Data: out})
}
```

- [ ] **Step 3: Replace Reject handler (lines ~363–375)**

Current code:
```go
func (h *LeaveHandler) Reject(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	...
	out, err := h.svc.Reject(c.Request.Context(), id)
```

Replace with:
```go
// Reject godoc
// @Summary      Reject a pending leave request
// @Description  Same permission semantics as Approve — requires approve_team or approve_all.
// @Tags         leave-requests
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "leave request uuid"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/leave-requests/{id}/reject [post]
func (h *LeaveHandler) Reject(c *gin.Context) {
	u, okC := currentUser(c)
	if !okC {
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	scope, ok := resolveApproveScope(c)
	if !ok {
		_ = c.Error(apperrors.ErrForbidden("Insufficient approve permission"))
		return
	}
	out, err := h.svc.Reject(c.Request.Context(), id, u.ID, scope)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.LeaveRequestRead]{Success: true, Data: out})
}
```

- [ ] **Step 4: Verify the handler compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/handlers/...
```

Expected: no output (success).

---

## Task 4: Remove RequirePerms gate from approve/reject routes

**Files:**
- Modify: `cmd/server/main.go` (lines ~297–298)

- [ ] **Step 1: Remove the middleware gate**

Current code (lines ~297–298):
```go
leaves.POST(":id/approve", middleware.RequirePerms(authSvc, permissions.PermLeaveApprove), leaveH.Approve)
leaves.POST(":id/reject",  middleware.RequirePerms(authSvc, permissions.PermLeaveApprove), leaveH.Reject)
```

Replace with (the `authed` group already provides JWT auth; perm check is now in the handler):
```go
leaves.POST(":id/approve", leaveH.Approve)
leaves.POST(":id/reject",  leaveH.Reject)
```

- [ ] **Step 2: Verify binary builds end-to-end**

```bash
cd e:\Work\Go-HRM && go build ./...
```

Expected: no output (success).

---

## Task 5: Update seed — Admin/HR Manager → approve_all, Manager → approve_team + update + delete

**Files:**
- Modify: `internal/services/seed_service.go`

- [ ] **Step 1: Update Admin seed (line ~88–89)**

Current:
```go
permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
```

Replace with:
```go
permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
permissions.PermLeaveApproveAll, permissions.PermLeaveCancel, permissions.PermLeaveManage,
```

- [ ] **Step 2: Update HR Manager seed (line ~121–122)**

Current:
```go
permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
```

Replace with:
```go
permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
permissions.PermLeaveApproveAll, permissions.PermLeaveCancel, permissions.PermLeaveManage,
```

- [ ] **Step 3: Update Manager seed (lines ~142–144)**

Current:
```go
permissions.PermLeaveRead, permissions.PermLeaveCreate,
permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
```

Replace with (adds update + delete; swaps approve → approve_team):
```go
permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
permissions.PermLeaveApproveTeam, permissions.PermLeaveCancel, permissions.PermLeaveManage,
```

- [ ] **Step 4: Verify it compiles**

```bash
cd e:\Work\Go-HRM && go build ./internal/services/...
```

Expected: no output.

---

## Task 6: Integration tests for approve scope enforcement

**Files:**
- Create: `internal/services/leave_approve_test.go`

- [ ] **Step 1: Write the test file**

Create `internal/services/leave_approve_test.go`:

```go
package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/services"
)

// setupApproveChain creates a manager and a subordinate employee, returning both.
// The subordinate's line_manager_id is set to the manager via a direct DB write
// (avoids spinning up EmployeeService just for relationship wiring).
func setupApproveChain(t *testing.T) (mgr, sub *models.Employee) {
	t.Helper()
	_, mgr = makeEmpUser(t, "mgr-approve@x.com", "Manager A")
	_, sub = makeEmpUser(t, "sub-approve@x.com", "Subordinate B")
	require.NoError(t,
		testDB.Exec("UPDATE employees SET line_manager_id = ? WHERE id = ?", mgr.ID, sub.ID).Error,
	)
	return
}

// makeLeave creates a pending leave request for the given employee.
func makeLeave(t *testing.T, svc *services.LeaveService, ownerUserID uuid.UUID) dto.LeaveRequestRead {
	t.Helper()
	res, err := svc.Create(context.Background(), ownerUserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 8, 1),
		ToDate:      dateAt(2026, 8, 1),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypePersonal,
		Reason:      "test",
	}, nil)
	require.NoError(t, err)
	return res.Request
}

// TestApprove_AllScope_CanApproveAny verifies that ApproveScopeAll passes
// regardless of the employee-manager relationship. Without this, only users
// with an explicit manage or wildcard perm could approve across the company.
func TestApprove_AllScope_CanApproveAny(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	approver, _ := makeEmpUser(t, "approver-all@x.com", "Approver All")
	_, subject := makeEmpUser(t, "subject-all@x.com", "Subject All")
	// No manager relationship between approver and subject.

	lr := makeLeave(t, svc, subject.UserID)

	_, err := svc.Approve(ctx, uuid.MustParse(lr.ID), approver.UserID, services.ApproveScopeAll)
	require.NoError(t, err, "ApproveScopeAll must approve any request")
}

// TestApprove_TeamScope_CanApproveSubordinate verifies that ApproveScopeTeam
// succeeds when the leave owner reports (directly) to the approver.
func TestApprove_TeamScope_CanApproveSubordinate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	mgr, sub := setupApproveChain(t)

	lr := makeLeave(t, svc, sub.UserID)

	_, err := svc.Approve(ctx, uuid.MustParse(lr.ID), mgr.UserID, services.ApproveScopeTeam)
	require.NoError(t, err, "ApproveScopeTeam must approve a direct subordinate's request")
}

// TestApprove_TeamScope_RejectsNonSubordinate verifies that ApproveScopeTeam
// returns 403 when the leave owner is NOT in the approver's reporting chain.
// This is the core safety property: Managers cannot approve across the company.
func TestApprove_TeamScope_RejectsNonSubordinate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	mgr, _ := setupApproveChain(t)
	_, unrelated := makeEmpUser(t, "unrelated@x.com", "Unrelated")

	lr := makeLeave(t, svc, unrelated.UserID)

	_, err := svc.Approve(ctx, uuid.MustParse(lr.ID), mgr.UserID, services.ApproveScopeTeam)
	require.Error(t, err, "ApproveScopeTeam must reject non-subordinate")
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeForbidden, ae.Code)
}

// TestReject_TeamScope_CanRejectSubordinate mirrors the approve test for reject.
func TestReject_TeamScope_CanRejectSubordinate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	mgr, sub := setupApproveChain(t)

	lr := makeLeave(t, svc, sub.UserID)

	_, err := svc.Reject(ctx, uuid.MustParse(lr.ID), mgr.UserID, services.ApproveScopeTeam)
	require.NoError(t, err, "ApproveScopeTeam must reject a direct subordinate's request")
}

// TestApprove_LegacyApproveScope_TreatedAsAll verifies backward compat:
// ApproveScopeAll covers legacy `leave_requests:approve` assignments. This is
// a unit-level check — the handler maps legacy perm → ApproveScopeAll.
// Covered by TestApprove_AllScope_CanApproveAny (same service path); no extra test needed.
// Left as documentation.
var _ = time.Now // suppress unused import
```

- [ ] **Step 2: Run the new tests**

```bash
cd e:\Work\Go-HRM && go test ./internal/services/... -run TestApprove -v
```

Expected: all 4 tests PASS (or SKIP with "no test DB" if the test DB is not available).
If they fail, diagnose before proceeding.

---

## Task 7: Final checks and commit

- [ ] **Step 1: Format and vet**

```bash
cd e:\Work\Go-HRM && make fmt && make vet
```

Expected: no output for `fmt`; no errors for `vet`.

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
git add internal/permissions/registry.go \
        internal/services/leave_service.go \
        internal/services/leave_approve_test.go \
        internal/handlers/leave_handler.go \
        cmd/server/main.go \
        internal/services/seed_service.go \
        docs/swagger/
git commit -m "feat(leave): add approve_team/approve_all permission split with BFS scope enforcement (G1+G2+G8)

- Add PermLeaveApproveTeam / PermLeaveApproveAll to registry and UI catalog;
  keep legacy PermLeaveApprove as backward-compat constant (not in catalog)
- Service: ApproveScope type + checkCanApproveOrReject (mirrors Python's
  check_can_approve_or_reject); reuses emps.SubordinateIDs for BFS chain check
- Handler: resolveApproveScope pre-computes scope from JWT claims; Approve/Reject
  handlers pass (approverUserID, scope) to service; perm gate moves out of router
  middleware and into handler (matches Python architecture)
- Seed: Admin/HR Manager -> approve_all; Manager -> approve_team + adds update + delete

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Self-Review

**Spec coverage check:**
- G1 (permission registry gap) → Task 1 ✅
- G2 (approve scope check missing) → Task 2 (`checkCanApproveOrReject`) ✅
- G8 (Manager seed missing update/delete) → Task 5 ✅
- D1 = A (full split with legacy backward compat) → Tasks 1+2+3+4+5 ✅

**Placeholder scan:** No TBD or "similar to" references. All code blocks are complete.

**Type consistency check:**
- `ApproveScope` defined in `leave_service.go`, referenced as `services.ApproveScope` in
  `leave_handler.go`. ✅
- `checkCanApproveOrReject` signature uses `*models.LeaveRequest` (same type as `FindByID`
  return). ✅
- `emps.SubordinateIDs(ctx, empID uuid.UUID) (map[uuid.UUID]bool, error)` — matches the
  interface definition at `employee_repo.go:42`. ✅
- `emps.FindByUserID(ctx, userID uuid.UUID) (*models.Employee, error)` — matches
  `employee_repo.go:22`. ✅
