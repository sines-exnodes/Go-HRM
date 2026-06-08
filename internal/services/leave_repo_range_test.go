package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

func TestLeaveRepo_ApprovedForEmployeesInRange(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()
	repo := repositories.NewLeaveRequestRepository(testDB)

	_, emp := makeEmpUser(t, "leaverange@example.com", "LeaveRange")

	mk := func(from, to string, status models.LeaveStatus) {
		f, _ := time.Parse("2006-01-02", from)
		tt, _ := time.Parse("2006-01-02", to)
		require.NoError(t, testDB.Create(&models.LeaveRequest{
			EmployeeID: emp.ID, FromDate: f, ToDate: tt,
			LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual,
			TotalDays: 1, Reason: "x", Status: status, CreatedBy: emp.ID,
		}).Error)
	}
	mk("2026-05-10", "2026-05-12", models.LeaveStatusApproved) // overlaps the month
	mk("2026-05-20", "2026-05-20", models.LeaveStatusPending)  // wrong status — excluded
	mk("2026-06-01", "2026-06-02", models.LeaveStatusApproved) // outside the month — excluded

	from, _ := time.Parse("2006-01-02", "2026-05-01")
	to, _ := time.Parse("2006-01-02", "2026-05-31")
	got, err := repo.ApprovedForEmployeesInRange(ctx, []uuid.UUID{emp.ID}, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "2026-05-10", got[0].FromDate.Format("2006-01-02"))
}
