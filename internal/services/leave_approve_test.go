package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/services"
)

// setupApproveChain creates a manager + subordinate linked by manager_id.
//
// The column is manager_id (migration 000003), NOT line_manager_id — the
// latter is the name of the FEATURE (the line-manager suite, PR #10), not of
// the column. Writing to line_manager_id makes Postgres reject the UPDATE
// with SQLSTATE 42703 and every caller of this helper fails before it reaches
// any approval logic.
func setupApproveChain(t *testing.T) (mgr, sub *models.Employee) {
	t.Helper()
	_, mgr = makeEmpUser(t, "mgr-approve@x.com", "Manager A")
	_, sub = makeEmpUser(t, "sub-approve@x.com", "Subordinate B")
	require.NoError(t,
		testDB.Exec("UPDATE employees SET manager_id = ? WHERE id = ?", mgr.ID, sub.ID).Error,
	)
	return
}

// makeTestLeave creates a pending leave request owned by ownerUserID.
func makeTestLeave(t *testing.T, svc *services.LeaveService, ownerUserID uuid.UUID) dto.LeaveRequestRead {
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

// TestApprove_AllScope_CanApproveAny verifies ApproveScopeAll bypasses the
// reporting-chain check and can approve any employee's request.
func TestApprove_AllScope_CanApproveAny(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	approver, _ := makeEmpUser(t, "approver-all@x.com", "Approver All")
	_, subject := makeEmpUser(t, "subject-all@x.com", "Subject All")
	makeLeaveQuota(t, subject.ID, 10, 5)

	lr := makeTestLeave(t, svc, subject.UserID)

	_, err := svc.Approve(ctx, uuid.MustParse(lr.ID), approver.ID, services.ApproveScopeAll)
	require.NoError(t, err, "ApproveScopeAll must approve any request")
}

// TestApprove_TeamScope_CanApproveSubordinate verifies ApproveScopeTeam allows
// approving a direct subordinate's request.
func TestApprove_TeamScope_CanApproveSubordinate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	mgr, sub := setupApproveChain(t)
	makeLeaveQuota(t, sub.ID, 10, 5)

	lr := makeTestLeave(t, svc, sub.UserID)

	_, err := svc.Approve(ctx, uuid.MustParse(lr.ID), mgr.UserID, services.ApproveScopeTeam)
	require.NoError(t, err, "ApproveScopeTeam must allow approving a subordinate's request")
}

// TestApprove_TeamScope_RejectsNonSubordinate verifies the core safety property:
// ApproveScopeTeam must return 403 when the leave owner is NOT in the approver's chain.
func TestApprove_TeamScope_RejectsNonSubordinate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	mgr, _ := setupApproveChain(t)
	_, unrelated := makeEmpUser(t, "unrelated@x.com", "Unrelated")
	makeLeaveQuota(t, unrelated.ID, 10, 5)

	lr := makeTestLeave(t, svc, unrelated.UserID)

	_, err := svc.Approve(ctx, uuid.MustParse(lr.ID), mgr.UserID, services.ApproveScopeTeam)
	require.Error(t, err, "ApproveScopeTeam must deny non-subordinate request approval")
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.CodeForbidden, ae.Code)
}

// TestReject_TeamScope_CanRejectSubordinate mirrors approve test for reject.
func TestReject_TeamScope_CanRejectSubordinate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	mgr, sub := setupApproveChain(t)
	makeLeaveQuota(t, sub.ID, 10, 5)

	lr := makeTestLeave(t, svc, sub.UserID)

	_, err := svc.Reject(ctx, uuid.MustParse(lr.ID), mgr.UserID, services.ApproveScopeTeam)
	require.NoError(t, err, "ApproveScopeTeam must allow rejecting a subordinate's request")
}
