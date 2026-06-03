package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// ---------------------------------------------------------------------------
// Line-manager suite (deferred audit decision #10): assignment validation
// (self / cycle / inactive / missing), rich manager brief, candidate picker,
// direct reports. Each test states WHY the behavior matters (AGENTS Rule 9).
// ---------------------------------------------------------------------------

func uptr(u uuid.UUID) *uuid.UUID { return &u }

// helper: make a department + position, return their ids
func makeOrg(t *testing.T, deptName, posName string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	d := uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO departments (id, name) VALUES (?, ?)", d, deptName).Error)
	p := uuid.New()
	require.NoError(t, testDB.Exec("INSERT INTO positions (id, name) VALUES (?, ?)", p, posName).Error)
	return d, p
}

func TestLineManager_RejectsSelfAssignment(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	e, err := svc.Create(ctx, dto.EmployeeCreate{Email: "self-mgr@x.com", Password: "Pass12345", FirstName: "Self", LastName: "Mgr"})
	require.NoError(t, err)

	_, err = svc.Update(ctx, e.ID, dto.EmployeeUpdate{ManagerID: uptr(e.ID)}, uuid.New())
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLineManager_RejectsCycle(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	// Chain A <- B <- C  (B reports to A, C reports to B).
	a, err := svc.Create(ctx, dto.EmployeeCreate{Email: "a@x.com", Password: "Pass12345", FirstName: "A", LastName: "Test"})
	require.NoError(t, err)
	b, err := svc.Create(ctx, dto.EmployeeCreate{Email: "b@x.com", Password: "Pass12345", FirstName: "B", LastName: "Test", ManagerID: uptr(a.ID)})
	require.NoError(t, err)
	c, err := svc.Create(ctx, dto.EmployeeCreate{Email: "c@x.com", Password: "Pass12345", FirstName: "C", LastName: "Test", ManagerID: uptr(b.ID)})
	require.NoError(t, err)

	// Setting A's manager to C would create a cycle (C is in A's chain).
	_, err = svc.Update(ctx, a.ID, dto.EmployeeUpdate{ManagerID: uptr(c.ID)}, uuid.New())
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLineManager_RejectsMissingManager(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	e, err := svc.Create(ctx, dto.EmployeeCreate{Email: "miss@x.com", Password: "Pass12345", FirstName: "Miss", LastName: "Test"})
	require.NoError(t, err)

	_, err = svc.Update(ctx, e.ID, dto.EmployeeUpdate{ManagerID: uptr(uuid.New())}, uuid.New())
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLineManager_RejectsInactiveManager(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	mgr, err := svc.Create(ctx, dto.EmployeeCreate{Email: "mgr@x.com", Password: "Pass12345", FirstName: "Mgr", LastName: "Test"})
	require.NoError(t, err)
	x, err := svc.Create(ctx, dto.EmployeeCreate{Email: "x@x.com", Password: "Pass12345", FirstName: "X", LastName: "Test"})
	require.NoError(t, err)

	// Deactivate the manager (a different caller, so the self-guard doesn't trip).
	no := false
	_, err = svc.Update(ctx, mgr.ID, dto.EmployeeUpdate{IsActive: &no}, uuid.New())
	require.NoError(t, err)

	// Assigning the now-inactive manager must be rejected.
	_, err = svc.Update(ctx, x.ID, dto.EmployeeUpdate{ManagerID: uptr(mgr.ID)}, uuid.New())
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeBadRequest, ae.Code)

	// On create, too.
	_, err = svc.Create(ctx, dto.EmployeeCreate{Email: "y@x.com", Password: "Pass12345", FirstName: "Y", LastName: "Test", ManagerID: uptr(mgr.ID)})
	require.Error(t, err)
}

func TestLineManager_RichManagerBriefOnRead(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	dept, pos := makeOrg(t, "Engineering", "Engineering Manager")
	mgr, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "boss@x.com", Password: "Pass12345", FirstName: "The", LastName: "Boss",
		DepartmentID: &dept, PositionID: &pos,
	})
	require.NoError(t, err)
	rep, err := svc.Create(ctx, dto.EmployeeCreate{Email: "rep@x.com", Password: "Pass12345", FirstName: "Report", LastName: "Test", ManagerID: uptr(mgr.ID)})
	require.NoError(t, err)

	got, err := svc.Get(ctx, rep.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Manager)
	assert.Equal(t, mgr.ID, got.Manager.ID)
	assert.Equal(t, "The Boss", got.Manager.FullName)
	require.NotNil(t, got.Manager.Position)
	assert.Equal(t, "Engineering Manager", *got.Manager.Position)
	require.NotNil(t, got.Manager.Department)
	assert.Equal(t, "Engineering", *got.Manager.Department)
	assert.True(t, got.Manager.IsActive)
}

func TestLineManager_Candidates_ExcludesSelfAndSubordinates(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	a, err := svc.Create(ctx, dto.EmployeeCreate{Email: "ca@x.com", Password: "Pass12345", FirstName: "Cand", LastName: "A"})
	require.NoError(t, err)
	b, err := svc.Create(ctx, dto.EmployeeCreate{Email: "cb@x.com", Password: "Pass12345", FirstName: "Cand", LastName: "B", ManagerID: uptr(a.ID)})
	require.NoError(t, err)
	c, err := svc.Create(ctx, dto.EmployeeCreate{Email: "cc@x.com", Password: "Pass12345", FirstName: "Cand", LastName: "C", ManagerID: uptr(b.ID)})
	require.NoError(t, err)
	d, err := svc.Create(ctx, dto.EmployeeCreate{Email: "cd@x.com", Password: "Pass12345", FirstName: "Cand", LastName: "D"}) // unrelated
	require.NoError(t, err)

	rows, err := svc.ManagerCandidates(ctx, uptr(a.ID), "", 50)
	require.NoError(t, err)
	ids := map[uuid.UUID]bool{}
	for _, r := range rows {
		ids[r.ID] = true
	}
	assert.False(t, ids[a.ID], "self excluded")
	assert.False(t, ids[b.ID], "direct subordinate excluded")
	assert.False(t, ids[c.ID], "transitive subordinate excluded")
	assert.True(t, ids[d.ID], "unrelated active employee included")
}

func TestLineManager_Candidates_KeepsInactiveCurrentManager(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	mgr, err := svc.Create(ctx, dto.EmployeeCreate{Email: "km@x.com", Password: "Pass12345", FirstName: "Kept", LastName: "Mgr"})
	require.NoError(t, err)
	emp, err := svc.Create(ctx, dto.EmployeeCreate{Email: "ke@x.com", Password: "Pass12345", FirstName: "Kept", LastName: "Emp", ManagerID: uptr(mgr.ID)})
	require.NoError(t, err)

	// Deactivate the assigned manager — normally excluded, but kept for this target.
	no := false
	_, err = svc.Update(ctx, mgr.ID, dto.EmployeeUpdate{IsActive: &no}, uuid.New())
	require.NoError(t, err)

	rows, err := svc.ManagerCandidates(ctx, uptr(emp.ID), "", 50)
	require.NoError(t, err)
	var kept *dto.ManagerCandidateRead
	for i := range rows {
		if rows[i].ID == mgr.ID {
			kept = &rows[i]
		}
	}
	require.NotNil(t, kept, "the currently-assigned but deactivated manager must remain selectable")
	assert.False(t, kept.IsActive)
}

func TestLineManager_DirectReports_IncludesInactive(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	mgr, err := svc.Create(ctx, dto.EmployeeCreate{Email: "dm@x.com", Password: "Pass12345", FirstName: "DR", LastName: "Mgr"})
	require.NoError(t, err)
	r1, err := svc.Create(ctx, dto.EmployeeCreate{Email: "dr1@x.com", Password: "Pass12345", FirstName: "Report", LastName: "One", ManagerID: uptr(mgr.ID)})
	require.NoError(t, err)
	r2, err := svc.Create(ctx, dto.EmployeeCreate{Email: "dr2@x.com", Password: "Pass12345", FirstName: "Report", LastName: "Two", ManagerID: uptr(mgr.ID)})
	require.NoError(t, err)

	// Deactivate one report — it must still show up in the direct-reports list.
	no := false
	_, err = svc.Update(ctx, r2.ID, dto.EmployeeUpdate{IsActive: &no}, uuid.New())
	require.NoError(t, err)

	rows, err := svc.DirectReports(ctx, mgr.ID)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	active := map[uuid.UUID]bool{}
	for _, r := range rows {
		active[r.ID] = r.IsActive
	}
	assert.True(t, active[r1.ID], "active report present")
	_, ok := active[r2.ID]
	assert.True(t, ok, "inactive report still listed")
	assert.False(t, active[r2.ID], "inactive report flagged is_active=false")
}

// Regression for review finding #4: soft-deleted org rows must not leak into
// the manager brief (the Manager.Department/Position preloads are NotDeleted-scoped).
func TestLineManager_SoftDeletedManagerOrgNotLeaked(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc, _ := newEmpSvc(testDB)
	ctx := context.Background()

	dept, pos := makeOrg(t, "Doomed Dept", "Doomed Pos")
	mgr, err := svc.Create(ctx, dto.EmployeeCreate{
		Email: "leakboss@x.com", Password: "Pass12345", FirstName: "Leak", LastName: "Boss",
		DepartmentID: &dept, PositionID: &pos,
	})
	require.NoError(t, err)
	rep, err := svc.Create(ctx, dto.EmployeeCreate{Email: "leakrep@x.com", Password: "Pass12345", FirstName: "Leak", LastName: "Rep", ManagerID: uptr(mgr.ID)})
	require.NoError(t, err)

	// Soft-delete the manager's department + position out from under the brief.
	require.NoError(t, testDB.Exec("UPDATE departments SET is_deleted = true WHERE id = ?", dept).Error)
	require.NoError(t, testDB.Exec("UPDATE positions SET is_deleted = true WHERE id = ?", pos).Error)

	got, err := svc.Get(ctx, rep.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Manager)
	assert.Nil(t, got.Manager.Department, "soft-deleted department must not leak into the manager brief")
	assert.Nil(t, got.Manager.Position, "soft-deleted position must not leak into the manager brief")
}
