package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// ---- Helpers ----

func newLeaveSvc(t *testing.T, up services.Uploader) (*services.LeaveService, repositories.LeaveRequestRepository, repositories.LeaveQuotaRepository) {
	t.Helper()
	lr := repositories.NewLeaveRequestRepository(testDB)
	emps := repositories.NewEmployeeRepository(testDB)
	depts := repositories.NewDepartmentRepository(testDB)
	pos := repositories.NewPositionRepository(testDB)
	quotaRepo := repositories.NewLeaveQuotaRepository(testDB)
	holidayRepo := repositories.NewHolidayRepository(testDB)
	return services.NewLeaveService(lr, emps, depts, pos, quotaRepo, up, holidayRepo), lr, quotaRepo
}

// makeLeaveQuota seeds a quota row for an employee. The leave service falls
// back to (0, 0) when no row exists; tests that need a specific budget use
// this helper to set the figures explicitly.
func makeLeaveQuota(t *testing.T, employeeID uuid.UUID, annual, sick float64) {
	t.Helper()
	repo := repositories.NewLeaveQuotaRepository(testDB)
	if err := repo.Upsert(context.Background(), employeeID, annual, sick); err != nil {
		t.Fatalf("upsert quota: %v", err)
	}
}

// makeEmpUser is a convenience: create a user + their employee in one call.
// The returned user has its UserID/Employee.ID pair set up so the service
// can resolve "current user → current employee" via FindByUserID.
func makeEmpUser(t *testing.T, email, fullName string) (*models.User, *models.Employee) {
	t.Helper()
	u := makeUser(t, email, "pw-Aa123456")
	e := makeEmployee(t, u, fullName)
	return u, e
}

// dateAt builds a UTC midnight date — matches how the service stores
// from_date/to_date after truncateToDate.
func dateAt(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// pdfBytes is a minimal valid PDF prefix that http.DetectContentType
// recognises as application/pdf.
var pdfBytes = []byte("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n1 0 obj\n<<>>\nendobj\n%%EOF\n")

// ---- Create — happy path (Task 16) ----

func TestLeaveService_Create_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	makeLeaveQuota(t, emp.ID, 12, 6)

	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 3),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "vacation",
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Request.Employee)
	require.Equal(t, emp.ID.String(), res.Request.Employee.ID)
	require.Equal(t, models.LeaveStatusPending, res.Request.Status)
	require.InDelta(t, 3.0, res.Request.TotalDays, 0.001)
	// created_by defaults to the employee themselves when no admin override.
	require.Equal(t, emp.ID.String(), res.Request.CreatedBy)
	// No warnings: quota 12, used 0, requested 3 → 9 remaining; no overlap.
	require.Empty(t, res.Warnings)
}

func TestLeaveService_Create_AdminOnBehalfOf_SetsCreatedByToAdmin(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, admin := makeEmpUser(t, "admin@example.com", "Admin")
	_, subject := makeEmpUser(t, "bob@example.com", "Bob")
	makeLeaveQuota(t, subject.ID, 12, 6)

	sid := subject.ID
	res, err := svc.Create(ctx, admin.UserID, true, dto.LeaveRequestCreate{
		EmployeeID:  &sid,
		FromDate:    dateAt(2026, 7, 10),
		ToDate:      dateAt(2026, 7, 10),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "covered by admin",
	}, nil)
	require.NoError(t, err)
	require.Equal(t, subject.ID.String(), res.Request.Employee.ID)
	require.Equal(t, admin.ID.String(), res.Request.CreatedBy)
}

func TestLeaveService_Create_NonAdminOnBehalfOfOther_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	_, other := makeEmpUser(t, "bob@example.com", "Bob")

	sid := other.ID
	_, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		EmployeeID:  &sid,
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "sneaky",
	}, nil)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeForbidden, ae.Code)
}

// ---- Validation: dates + half-day rule (Task 17) ----

func TestLeaveService_Create_ToBeforeFrom_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	_, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 5),
		ToDate:      dateAt(2026, 6, 3),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "x",
	}, nil)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLeaveService_Create_HalfDayMustBeSingleDay(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	_, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 2),
		LeavePeriod: models.LeavePeriodMorningHalf,
		LeaveType:   models.LeaveTypePersonal,
		Reason:      "x",
	}, nil)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
}

func TestLeaveService_Create_HalfDay_TotalDaysHalf(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodAfternoonHalf,
		LeaveType:   models.LeaveTypePersonal,
		Reason:      "doctor",
	}, nil)
	require.NoError(t, err)
	require.InDelta(t, 0.5, res.Request.TotalDays, 0.001)
}

// ---- Warnings — insufficient quota + overlap (Task 17) ----

func TestLeaveService_Create_InsufficientQuota_WarnsButCreates(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	makeLeaveQuota(t, emp.ID, 2, 6) // only 2 days of annual

	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 5), // 5 days requested
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "long vacation",
	}, nil)
	require.NoError(t, err) // non-blocking warning
	require.NotEmpty(t, res.Warnings)
	require.Equal(t, models.LeaveStatusPending, res.Request.Status)
}

func TestLeaveService_Create_PersonalLeave_NoQuotaWarning(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	// No quota row — non-quota types must never trigger the warning.

	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 7),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypePersonal,
		Reason:      "personal",
	}, nil)
	require.NoError(t, err)
	require.Empty(t, res.Warnings)
}

func TestLeaveService_Create_DateOverlap_WarnsButCreates(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	makeLeaveQuota(t, emp.ID, 12, 6)

	// First request: 6/1..6/3 (pending). No warning.
	_, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 3),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "first",
	}, nil)
	require.NoError(t, err)

	// Second request overlapping with the first: 6/2..6/4. Should warn.
	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 2),
		ToDate:      dateAt(2026, 6, 4),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "second",
	}, nil)
	require.NoError(t, err) // non-blocking
	require.NotEmpty(t, res.Warnings)
}

// ---- Status transitions (Task 18) ----

func TestLeaveService_ApproveAndCancel_BalanceLifecycle(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	makeLeaveQuota(t, emp.ID, 12, 6)

	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 2),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		Reason:      "vacation",
	}, nil)
	require.NoError(t, err)
	leaveID := uuid.MustParse(res.Request.ID)

	// Approve → status=approved.
	read, err := svc.Approve(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)
	require.NoError(t, err)
	require.Equal(t, models.LeaveStatusApproved, read.Status)

	// Balance reflects the 2 approved days against quota 12 → remaining 10.
	bal, err := svc.GetBalance(ctx, emp.ID, 2026)
	require.NoError(t, err)
	require.InDelta(t, 2.0, bal.AnnualUsed, 0.001)
	require.InDelta(t, 10.0, bal.AnnualRemaining, 0.001)

	// Cancel approved → status=cancelled.
	read, _, err = svc.Cancel(ctx, leaveID, emp.UserID, false)
	require.NoError(t, err)
	require.Equal(t, models.LeaveStatusCancelled, read.Status)

	// Balance restored: cancelled rows are excluded from the SUM.
	bal, err = svc.GetBalance(ctx, emp.ID, 2026)
	require.NoError(t, err)
	require.InDelta(t, 0.0, bal.AnnualUsed, 0.001)
	require.InDelta(t, 12.0, bal.AnnualRemaining, 0.001)
}

func TestLeaveService_Reject_FromPending_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 6, 1),
		ToDate:      dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypePersonal,
		Reason:      "x",
	}, nil)
	require.NoError(t, err)
	read, err := svc.Reject(ctx, uuid.MustParse(res.Request.ID), uuid.Nil, services.ApproveScopeAll)
	require.NoError(t, err)
	require.Equal(t, models.LeaveStatusRejected, read.Status)
}

func TestLeaveService_ApproveAfterReject_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	res, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, nil)
	leaveID := uuid.MustParse(res.Request.ID)
	_, _ = svc.Reject(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)

	_, err := svc.Approve(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

func TestLeaveService_Update_ApprovedByAdmin_RevertsToPending(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	_, admin := makeEmpUser(t, "admin@example.com", "Admin")

	res, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, nil)
	leaveID := uuid.MustParse(res.Request.ID)
	_, _ = svc.Approve(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)

	newReason := "updated reason"
	out, err := svc.Update(ctx, leaveID, admin.UserID, true, dto.LeaveRequestUpdate{Reason: &newReason}, nil)
	require.NoError(t, err)
	require.Equal(t, models.LeaveStatusPending, out.Request.Status)
	require.Equal(t, "updated reason", out.Request.Reason)
}

func TestLeaveService_Update_RejectedTerminal_Conflict(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	_, admin := makeEmpUser(t, "admin@example.com", "Admin")

	res, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, nil)
	leaveID := uuid.MustParse(res.Request.ID)
	_, _ = svc.Reject(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)

	newReason := "no edits"
	_, err := svc.Update(ctx, leaveID, admin.UserID, true, dto.LeaveRequestUpdate{Reason: &newReason}, nil)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeConflict, ae.Code)
}

// ---- Ownership + delete rules (Task 18) ----

func TestLeaveService_Update_NotOwnerNotAdmin_Forbidden(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, owner := makeEmpUser(t, "owner@example.com", "Owner")
	_, stranger := makeEmpUser(t, "stranger@example.com", "Stranger")

	res, _ := svc.Create(ctx, owner.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, nil)
	leaveID := uuid.MustParse(res.Request.ID)

	newReason := "stranger trying to edit"
	_, err := svc.Update(ctx, leaveID, stranger.UserID, false, dto.LeaveRequestUpdate{Reason: &newReason}, nil)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeForbidden, ae.Code)
}

func TestLeaveService_Delete_NonAdminOwnerOnlyPending(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	// pending → owner CAN delete
	res, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "p",
	}, nil)
	require.NoError(t, svc.Delete(ctx, uuid.MustParse(res.Request.ID), emp.UserID, false))

	// approved → owner CANNOT delete
	res2, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 5), ToDate: dateAt(2026, 6, 5),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "q",
	}, nil)
	leaveID := uuid.MustParse(res2.Request.ID)
	_, _ = svc.Approve(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)
	err := svc.Delete(ctx, leaveID, emp.UserID, false)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeForbidden, ae.Code)
}

func TestLeaveService_Delete_AdminCanDeleteAnyStatus(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	_, admin := makeEmpUser(t, "admin@example.com", "Admin")

	res, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, nil)
	leaveID := uuid.MustParse(res.Request.ID)
	_, _ = svc.Approve(ctx, leaveID, uuid.Nil, services.ApproveScopeAll)

	// Admin deletes the approved row.
	require.NoError(t, svc.Delete(ctx, leaveID, admin.UserID, true))
}

func TestLeaveService_Get_OwnerOrAdminOnly(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, owner := makeEmpUser(t, "owner@example.com", "Owner")
	_, stranger := makeEmpUser(t, "stranger@example.com", "Stranger")
	_, admin := makeEmpUser(t, "admin@example.com", "Admin")

	res, _ := svc.Create(ctx, owner.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, nil)
	leaveID := uuid.MustParse(res.Request.ID)

	// Owner: OK.
	got, err := svc.Get(ctx, leaveID, owner.UserID, false)
	require.NoError(t, err)
	require.Equal(t, res.Request.ID, got.ID)

	// Stranger: 403.
	_, err = svc.Get(ctx, leaveID, stranger.UserID, false)
	require.Error(t, err)

	// Admin: OK.
	got, err = svc.Get(ctx, leaveID, admin.UserID, true)
	require.NoError(t, err)
	require.Equal(t, res.Request.ID, got.ID)
}

// ---- Balance corner case ----

func TestLeaveService_GetBalance_NoQuotaRow_DefaultsTo12And6(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	bal, err := svc.GetBalance(ctx, emp.ID, 2026)
	require.NoError(t, err)
	// No explicit quota row → falls back to DB column defaults (12 annual / 6 sick)
	// to match the employee profile view.
	require.Equal(t, 12.0, bal.AnnualQuota)
	require.Equal(t, 6.0, bal.SickQuota)
	require.Equal(t, 12.0, bal.AnnualRemaining)
	require.Equal(t, 0, bal.LeavesThisYear)
}

// ---- List + History + Dashboard (Task 19) ----

func TestLeaveService_List_FiltersByStatus(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	// Create 3 pending + 1 approved.
	for i := 0; i < 4; i++ {
		from := dateAt(2026, 6, 1+i)
		res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
			FromDate: from, ToDate: from,
			LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal,
			Reason: fmt.Sprintf("day %d", i),
		}, nil)
		require.NoError(t, err)
		if i == 0 {
			_, _ = svc.Approve(ctx, uuid.MustParse(res.Request.ID), uuid.Nil, services.ApproveScopeAll)
		}
	}

	// List status=pending → 3.
	page, err := svc.List(ctx, dto.LeaveListQuery{Page: 1, PageSize: 10, Status: []string{"pending"}})
	require.NoError(t, err)
	require.Equal(t, int64(3), page.Total)
	for _, it := range page.Items {
		require.Equal(t, models.LeaveStatusPending, it.Status)
	}

	// List status=approved → 1.
	page, err = svc.List(ctx, dto.LeaveListQuery{Page: 1, PageSize: 10, Status: []string{"approved"}})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
}

func TestLeaveService_ListMyHistory_ReturnsPastOrTerminal(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	// Past row (to_date in 2020 — always past).
	pastReq := models.LeaveRequest{
		EmployeeID:  emp.ID,
		FromDate:    dateAt(2020, 1, 1),
		ToDate:      dateAt(2020, 1, 1),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypePersonal,
		TotalDays:   1,
		Reason:      "ancient",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, testDB.Create(&pastReq).Error)

	// Future pending (NOT in history — to_date future and not terminal).
	res, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2099, 1, 1), ToDate: dateAt(2099, 1, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "future",
	}, nil)
	require.NotEmpty(t, res.Request.ID)

	// Future cancelled (terminal — IS in history).
	res2, _ := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2099, 2, 1), ToDate: dateAt(2099, 2, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "future-cancelled",
	}, nil)
	_, _, err := svc.Cancel(ctx, uuid.MustParse(res2.Request.ID), emp.UserID, false)
	require.NoError(t, err)

	page, err := svc.ListMyHistory(ctx, emp.UserID, dto.LeaveHistoryQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	// Expect ancient (past) + future-cancelled (terminal) = 2 rows.
	require.Equal(t, int64(2), page.Total)
	for _, it := range page.Items {
		// Future-pending must not appear.
		require.NotEqual(t, "future", it.Reason)
	}
}

func TestLeaveService_GetMyDashboard_Composes(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")
	makeLeaveQuota(t, emp.ID, 12, 6)

	// Upcoming: 1 future-pending row.
	_, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2099, 1, 1), ToDate: dateAt(2099, 1, 2),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "future",
	}, nil)
	require.NoError(t, err)
	// History: 1 past row inserted directly so to_date is < today.
	pastReq := models.LeaveRequest{
		EmployeeID:  emp.ID,
		FromDate:    dateAt(2020, 1, 1),
		ToDate:      dateAt(2020, 1, 1),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		TotalDays:   1,
		Reason:      "ancient",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, testDB.Create(&pastReq).Error)

	dash, err := svc.GetMyDashboard(ctx, emp.UserID)
	require.NoError(t, err)
	require.Equal(t, 12.0, dash.Balance.AnnualQuota)
	require.GreaterOrEqual(t, len(dash.Upcoming), 1)
	require.GreaterOrEqual(t, len(dash.History), 1)
}

// ---- Attachment content-spoof (Task 17 cross-check) ----

func TestLeaveService_Create_AttachmentSpoof_BadRequest(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	up := &stubUploader{}
	svc, _, _ := newLeaveSvc(t, up)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	// Plain text bytes with a .pdf extension and a faked PDF Content-Type.
	// The service MUST sniff the actual bytes (http.DetectContentType)
	// and reject — no upload should happen.
	spoof := &services.AttachmentUpload{
		Content:     []byte("this is just text, not a real PDF"),
		ContentType: "application/pdf",
		Ext:         ".pdf",
	}
	_, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, spoof)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	require.Equal(t, apperrors.CodeBadRequest, ae.Code)
	require.Equal(t, int32(0), up.uploaded, "no upload should be attempted for a spoofed attachment")
}

func TestLeaveService_Create_AttachmentPDF_OK(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	up := &stubUploader{}
	svc, _, _ := newLeaveSvc(t, up)
	_, emp := makeEmpUser(t, "alice@example.com", "Alice")

	att := &services.AttachmentUpload{
		Content:     pdfBytes,
		ContentType: "application/pdf",
		Ext:         ".pdf",
	}
	res, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate: dateAt(2026, 6, 1), ToDate: dateAt(2026, 6, 1),
		LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypePersonal, Reason: "x",
	}, att)
	require.NoError(t, err)
	require.NotNil(t, res.Request.AttachmentURL)
	require.Equal(t, int32(1), up.uploaded)
}

// TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus verifies that an empty
// PATCH body does not reset an Approved request back to Pending. This was a
// Go-specific regression: status-transition logic ran unconditionally even when
// no fields changed (G3 fix).
func TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
	_, emp := makeEmpUser(t, "emp-noop@x.com", "Emp NoOp")
	makeLeaveQuota(t, emp.ID, 10, 5)

	createRes, err := svc.Create(ctx, emp.UserID, false, dto.LeaveRequestCreate{
		FromDate:    dateAt(2026, 9, 1),
		ToDate:      dateAt(2026, 9, 1),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypePersonal,
		Reason:      "noop test",
	}, nil)
	require.NoError(t, err)

	id := uuid.MustParse(createRes.Request.ID)
	require.NoError(t, testDB.Model(&models.LeaveRequest{}).
		Where("id = ?", id).
		Update("status", models.LeaveStatusApproved).Error)

	updateRes, err := svc.Update(ctx, id, emp.UserID, false, dto.LeaveRequestUpdate{}, nil)
	require.NoError(t, err, "empty PATCH must not error")
	assert.Equal(t, string(models.LeaveStatusApproved), string(updateRes.Request.Status),
		"empty PATCH must not revert Approved→Pending (G3 regression)")
}

// TestCancel_WasApprovedTrue verifies that cancelling an Approved request
// returns wasApproved=true so callers know quota should be restored (G7).
func TestCancel_WasApprovedTrue(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _, _ := newLeaveSvc(t, nil)
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
	require.NoError(t, testDB.Model(&models.LeaveRequest{}).
		Where("id = ?", id).
		Update("status", models.LeaveStatusApproved).Error)

	_, wasApproved, err := svc.Cancel(ctx, id, emp.UserID, false)
	require.NoError(t, err)
	assert.True(t, wasApproved, "cancelling an Approved request must return wasApproved=true")
}

// TestCancel_WasApprovedFalse verifies that cancelling a Pending request
// returns wasApproved=false.
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
