package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func newDashboardSvc(t *testing.T) *services.DashboardService {
	t.Helper()
	return services.NewDashboardService(
		&config.Config{CompanyTimezone: "UTC"},
		repositories.NewEmployeeRepository(testDB),
		repositories.NewLeaveRequestRepository(testDB),
		repositories.NewLeaveQuotaRepository(testDB),
		repositories.NewAttendanceRepository(testDB),
		repositories.NewAnnouncementRepository(testDB),
		repositories.NewHolidayRepository(testDB),
	)
}

func dashboardWidgetIDs(d *dto.DashboardRead) []string {
	ids := make([]string, 0, len(d.Widgets))
	for _, w := range d.Widgets {
		ids = append(ids, w.ID)
	}
	return ids
}

func dashboardWidget(t *testing.T, d *dto.DashboardRead, id string) dto.DashboardWidgetRead {
	t.Helper()
	for _, w := range d.Widgets {
		if w.ID == id {
			return w
		}
	}
	t.Fatalf("widget %q not found in %#v", id, dashboardWidgetIDs(d))
	return dto.DashboardWidgetRead{}
}

func dashboardMetric(t *testing.T, w dto.DashboardWidgetRead, key string) dto.DashboardMetricRead {
	t.Helper()
	for _, m := range w.Metrics {
		if m.Key == key {
			return m
		}
	}
	t.Fatalf("metric %q not found in widget %#v", key, w.ID)
	return dto.DashboardMetricRead{}
}

func dashboardHasMetric(w dto.DashboardWidgetRead, key string) bool {
	for _, m := range w.Metrics {
		if m.Key == key {
			return true
		}
	}
	return false
}

func TestDashboardService_Get_FiltersWidgetsAndKeepsStableOrder(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	role := makeRole(t, "Dashboard Employee", []permissions.Permission{
		permissions.PermAuthLogin,
		permissions.PermAttendanceRead,
		permissions.PermLeaveRead,
		permissions.PermLeaveCreate,
		permissions.PermAnnounceRead,
		permissions.PermOrgHolidaysView,
		permissions.PermOrgWorkdaysView,
	}, false)
	user := makeUser(t, "dashboard-employee@example.com", "pw-Aa123456", role)
	user.Roles = []models.Role{*role}
	emp := makeEmployee(t, user, "Dashboard Employee")

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	attendance := &models.Attendance{EmployeeID: emp.ID, Date: today, IsLate: true}
	require.NoError(t, testDB.Create(attendance).Error)
	require.NoError(t, testDB.Create(&models.AttendanceSession{
		AttendanceID: attendance.ID,
		CheckIn:      today.Add(10 * time.Hour),
	}).Error)

	tomorrow := today.AddDate(0, 0, 1)
	require.NoError(t, testDB.Create(&models.LeaveRequest{
		EmployeeID:  emp.ID,
		CreatedBy:   emp.ID,
		FromDate:    tomorrow,
		ToDate:      tomorrow,
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		TotalDays:   1,
		Reason:      "planned leave",
		Status:      models.LeaveStatusPending,
	}).Error)

	publishedAt := now
	require.NoError(t, testDB.Create(&models.Announcement{
		AuthorID:       emp.ID,
		Title:          "Published update",
		Description:    "Visible on dashboard",
		Status:         models.AnnouncementStatusPublished,
		PublishedAt:    &publishedAt,
		TargetAudience: models.AnnouncementAudienceAll,
	}).Error)
	require.NoError(t, testDB.Create(&models.Holiday{
		Year:     tomorrow.Year(),
		Name:     "Company Holiday",
		FromDate: tomorrow,
		ToDate:   tomorrow,
	}).Error)

	dash, err := newDashboardSvc(t).Get(ctx, user)
	require.NoError(t, err)
	require.False(t, dash.Empty)
	require.Equal(t, []string{
		"attendance_overview",
		"leave_summary",
		"announcements",
		"holidays_workdays",
	}, dashboardWidgetIDs(dash))

	attendanceWidget := dashboardWidget(t, dash, "attendance_overview")
	require.Equal(t, "own", attendanceWidget.Scope)
	require.Equal(t, "checked_in", dashboardMetric(t, attendanceWidget, "today_status").Value)

	leaveWidget := dashboardWidget(t, dash, "leave_summary")
	require.NotEmpty(t, leaveWidget.Actions, "leave create quick action should be visible")

	holidayWidget := dashboardWidget(t, dash, "holidays_workdays")
	require.Equal(t, float64(1), dashboardMetric(t, holidayWidget, "upcoming_holidays").Value)
}

func TestDashboardService_Get_PendingApprovalsAreTeamScoped(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	role := makeRole(t, "Dashboard Manager", []permissions.Permission{
		permissions.PermAuthLogin,
		permissions.PermLeaveApproveTeam,
	}, false)
	managerUser := makeUser(t, "dashboard-manager@example.com", "pw-Aa123456", role)
	managerUser.Roles = []models.Role{*role}
	manager := makeEmployee(t, managerUser, "Dashboard Manager")
	subUser := makeUser(t, "dashboard-report@example.com", "pw-Aa123456")
	subordinate := makeEmployee(t, subUser, "Dashboard Report")
	subordinate.ManagerID = &manager.ID
	require.NoError(t, testDB.Save(subordinate).Error)
	outsiderUser := makeUser(t, "dashboard-outsider@example.com", "pw-Aa123456")
	outsider := makeEmployee(t, outsiderUser, "Dashboard Outsider")

	start := time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC)
	for _, empID := range []uuid.UUID{subordinate.ID, outsider.ID} {
		require.NoError(t, testDB.Create(&models.LeaveRequest{
			EmployeeID:  empID,
			CreatedBy:   empID,
			FromDate:    start,
			ToDate:      start,
			LeavePeriod: models.LeavePeriodFullDay,
			LeaveType:   models.LeaveTypePersonal,
			TotalDays:   1,
			Reason:      "pending approval",
			Status:      models.LeaveStatusPending,
		}).Error)
	}

	dash, err := newDashboardSvc(t).Get(ctx, managerUser)
	require.NoError(t, err)

	widget := dashboardWidget(t, dash, "pending_approvals")
	require.Equal(t, "team", widget.Scope)
	require.Equal(t, float64(1), dashboardMetric(t, widget, "pending_leave").Value)
	require.Len(t, widget.Items, 1)
	require.Contains(t, widget.Items[0].Title, "Dashboard Report")
}

func TestDashboardService_Get_LeaveSummaryUsesOrganizationScopeForManagers(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	role := makeRole(t, "Dashboard HR", []permissions.Permission{
		permissions.PermAuthLogin,
		permissions.PermLeaveRead,
		permissions.PermLeaveManage,
	}, false)
	hrUser := makeUser(t, "dashboard-hr@example.com", "pw-Aa123456", role)
	hrUser.Roles = []models.Role{*role}
	makeEmployee(t, hrUser, "Dashboard HR")
	targetUser := makeUser(t, "dashboard-leave-target@example.com", "pw-Aa123456")
	target := makeEmployee(t, targetUser, "Dashboard Leave Target")

	start := time.Date(2099, 1, 3, 0, 0, 0, 0, time.UTC)
	require.NoError(t, testDB.Create(&models.LeaveRequest{
		EmployeeID:  target.ID,
		CreatedBy:   target.ID,
		FromDate:    start,
		ToDate:      start,
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		TotalDays:   1,
		Reason:      "approved away day",
		Status:      models.LeaveStatusApproved,
	}).Error)

	dash, err := newDashboardSvc(t).Get(ctx, hrUser)
	require.NoError(t, err)

	widget := dashboardWidget(t, dash, "leave_summary")
	require.Equal(t, "organization", widget.Scope)
	require.Equal(t, float64(1), dashboardMetric(t, widget, "upcoming_leave").Value)
	require.False(t, dashboardHasMetric(widget, "annual_remaining"), "organization leave summary should not show the viewer's personal balance")
	require.Empty(t, widget.Actions, "organization leave summary should not show personal create-leave action")
	require.Len(t, widget.Items, 1)
	require.Contains(t, widget.Items[0].Title, "Dashboard Leave Target")
}

func TestDashboardService_Get_AuthOnlyUserGetsNeutralEmptyDashboard(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	role := makeRole(t, "Dashboard Auth Only", []permissions.Permission{permissions.PermAuthLogin}, false)
	user := makeUser(t, "dashboard-empty@example.com", "pw-Aa123456", role)
	user.Roles = []models.Role{*role}

	dash, err := newDashboardSvc(t).Get(ctx, user)
	require.NoError(t, err)
	require.True(t, dash.Empty)
	require.Empty(t, dash.Widgets)
	require.NotContains(t, dash.EmptyMessage, "attendance")
	require.NotContains(t, dash.EmptyMessage, "leave")
	require.NotContains(t, dash.EmptyMessage, "announcement")
}
