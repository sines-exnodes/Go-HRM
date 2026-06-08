package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

func makeApprovedLeave(t *testing.T, empID uuid.UUID, from, to string, lt models.LeaveType, lp models.LeavePeriod) {
	t.Helper()
	f, err := time.Parse("2006-01-02", from)
	require.NoError(t, err)
	tt, err := time.Parse("2006-01-02", to)
	require.NoError(t, err)
	require.NoError(t, testDB.Create(&models.LeaveRequest{
		EmployeeID: empID, FromDate: f, ToDate: tt,
		LeavePeriod: lp, LeaveType: lt, TotalDays: 1, Reason: "test",
		Status: models.LeaveStatusApproved, CreatedBy: empID,
	}).Error)
}

func TestAttendance_Matrix_FullDayLeaveCell(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "leave1@example.com", "LeaveOne")
	makeApprovedLeave(t, emp.ID, "2026-04-08", "2026-04-08", models.LeaveTypeSick, models.LeavePeriodFullDay)
	out, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	cell := out.Items[0].Cells[8]
	assert.Equal(t, "sick_leave", cell.Status)
	require.NotNil(t, cell.LeaveType)
	assert.Equal(t, "sick", *cell.LeaveType)
}

func TestAttendance_Matrix_CombinedCell_PMWorkedLate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "combo1@example.com", "Combo")
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

func TestAttendance_Matrix_StatusFilter_CombinedMultiMatch(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	svc := newAttendanceSvc(t)
	u, emp := makeEmpUser(t, "filt1@example.com", "Filt")
	makeApprovedLeave(t, emp.ID, "2026-04-09", "2026-04-09", models.LeaveTypeAnnual, models.LeavePeriodMorningHalf)
	ci := hcmTime(t, 2026, 4, 9, 13, 25)
	_, err := svc.CheckIn(ctx, u.ID, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)
	outLate, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10, Status: "late"})
	require.NoError(t, err)
	require.Len(t, outLate.Items, 1)
	outLeave, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10, Status: "on_leave"})
	require.NoError(t, err)
	require.Len(t, outLeave.Items, 1)
	// A status genuinely absent from this row (April 2026 is wholly in the
	// past, so no cell is no_data) must NOT match — proving the filter is not
	// a blanket pass-through.
	outNoData, err := svc.Matrix(ctx, u.ID, false, dto.AttendanceMatrixQuery{Month: 4, Year: 2026, Page: 1, PageSize: 10, Status: "no_data"})
	require.NoError(t, err)
	assert.Len(t, outNoData.Items, 0)
}
