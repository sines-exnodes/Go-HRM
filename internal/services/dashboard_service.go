package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
)

const dashboardEmptyMessage = "No dashboard items are available yet."

// DashboardService composes the read-only web dashboard from existing source
// modules. It owns no data and does not introduce dashboard-specific access
// control; each widget follows the source module's permissions and scope.
type DashboardService struct {
	cfg           *config.Config
	emps          repositories.EmployeeRepository
	leaves        repositories.LeaveRequestRepository
	quota         repositories.LeaveQuotaRepository
	attendance    repositories.AttendanceRepository
	announcements repositories.AnnouncementRepository
	holidays      repositories.HolidayRepository
	notifications repositories.NotificationRepository
}

func NewDashboardService(
	cfg *config.Config,
	emps repositories.EmployeeRepository,
	leaves repositories.LeaveRequestRepository,
	quota repositories.LeaveQuotaRepository,
	attendance repositories.AttendanceRepository,
	announcements repositories.AnnouncementRepository,
	holidays repositories.HolidayRepository,
	notifications repositories.NotificationRepository,
) *DashboardService {
	return &DashboardService{
		cfg:           cfg,
		emps:          emps,
		leaves:        leaves,
		quota:         quota,
		attendance:    attendance,
		announcements: announcements,
		holidays:      holidays,
		notifications: notifications,
	}
}

// Get returns the fixed-order dashboard for the authenticated user.
func (s *DashboardService) Get(ctx context.Context, user *models.User) (*dto.DashboardRead, error) {
	var currentEmp *models.Employee
	emp, err := s.emps.FindByUserID(ctx, user.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		currentEmp = emp
	}

	out := &dto.DashboardRead{
		Greeting: dto.DashboardGreetingRead{
			Name:  dashboardGreetingName(user, currentEmp),
			Email: user.Email,
		},
		Widgets: []dto.DashboardWidgetRead{},
	}

	loc := time.UTC
	if s.cfg != nil {
		loc = loadTZ(s.cfg.CompanyTimezone)
	}
	_, today := todayInTZ(loc)

	if w, ok, err := s.attendanceWidget(ctx, user, currentEmp, today); err != nil {
		return nil, err
	} else if ok {
		out.Widgets = append(out.Widgets, w)
	}
	if w, ok, err := s.pendingApprovalsWidget(ctx, user, currentEmp); err != nil {
		return nil, err
	} else if ok {
		out.Widgets = append(out.Widgets, w)
	}
	if w, ok, err := s.leaveSummaryWidget(ctx, user, currentEmp, today); err != nil {
		return nil, err
	} else if ok {
		out.Widgets = append(out.Widgets, w)
	}
	if w, ok, err := s.announcementsWidget(ctx, user, currentEmp); err != nil {
		return nil, err
	} else if ok {
		out.Widgets = append(out.Widgets, w)
	}
	if w, ok, err := s.holidaysWorkdaysWidget(ctx, user, today); err != nil {
		return nil, err
	} else if ok {
		out.Widgets = append(out.Widgets, w)
	}
	if w, ok, err := s.workforceSummaryWidget(ctx, user, today); err != nil {
		return nil, err
	} else if ok {
		out.Widgets = append(out.Widgets, w)
	}

	// The bell count is independent of the widget list — an employee with no
	// visible widgets can still have unread notifications.
	if s.notifications != nil {
		unread, err := s.notifications.CountUnread(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		out.UnreadNotificationCount = unread
	}

	out.Empty = len(out.Widgets) == 0
	if out.Empty {
		out.EmptyMessage = dashboardEmptyMessage
	}
	return out, nil
}

func (s *DashboardService) attendanceWidget(ctx context.Context, user *models.User, currentEmp *models.Employee, today time.Time) (dto.DashboardWidgetRead, bool, error) {
	if !dashboardHasAny(user, permissions.PermAttendanceRead, permissions.PermAttendanceManage) {
		return dto.DashboardWidgetRead{}, false, nil
	}
	if dashboardHas(user, permissions.PermAttendanceManage) {
		return s.attendanceOrganizationWidget(ctx, today)
	}
	if currentEmp == nil {
		return dto.DashboardWidgetRead{}, false, nil
	}
	row, err := s.attendance.FindByEmployeeAndDate(ctx, currentEmp.ID, today)
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	status := "not_checked_in"
	isLate := false
	if row != nil && len(row.Sessions) > 0 {
		isLate = row.IsLate
		if row.Sessions[len(row.Sessions)-1].CheckOut == nil {
			status = "checked_in"
		} else {
			status = "checked_out"
		}
	}
	monthly, err := s.attendance.MonthlyCheckInCount(ctx, currentEmp.ID, today.Year(), int(today.Month()))
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	items := []dto.DashboardItemRead{}
	if isLate {
		date := today
		items = append(items, dto.DashboardItemRead{
			Title:       "Late check-in",
			Description: "Your first check-in was after the configured threshold.",
			Status:      "late",
			Date:        &date,
			URL:         "/attendance",
		})
	}
	return dto.DashboardWidgetRead{
		ID:    "attendance_overview",
		Title: "Attendance Overview",
		Scope: "own",
		Metrics: []dto.DashboardMetricRead{
			{Key: "today_status", Label: "Today Status", Value: status},
			{Key: "monthly_check_ins", Label: "Monthly Check-ins", Value: float64(monthly)},
		},
		Items:        items,
		EmptyMessage: "No attendance activity for today.",
	}, true, nil
}

func (s *DashboardService) attendanceOrganizationWidget(ctx context.Context, today time.Time) (dto.DashboardWidgetRead, bool, error) {
	active := true
	emps, totalActive, err := s.emps.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 1000, IsActive: &active})
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	rows, totalPresent, err := s.attendance.List(ctx, repositories.AttendanceListFilter{
		StartDate: &today,
		EndDate:   &today,
		Page:      1,
		PageSize:  1000,
	})
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	late := 0
	items := []dto.DashboardItemRead{}
	for i := range rows {
		if !rows[i].IsLate {
			continue
		}
		late++
		if len(items) >= 5 {
			continue
		}
		name := "Employee"
		if rows[i].Employee != nil {
			name = fallbackEmployeeName(rows[i].Employee)
		}
		date := rows[i].Date
		items = append(items, dto.DashboardItemRead{
			ID:     rows[i].ID.String(),
			Title:  name,
			Status: "late",
			Date:   &date,
			URL:    "/attendance",
		})
	}
	if totalActive == 0 {
		totalActive = int64(len(emps))
	}
	absent := totalActive - totalPresent
	if absent < 0 {
		absent = 0
	}
	return dto.DashboardWidgetRead{
		ID:    "attendance_overview",
		Title: "Attendance Overview",
		Scope: "organization",
		Metrics: []dto.DashboardMetricRead{
			{Key: "active_employees", Label: "Active Employees", Value: float64(totalActive)},
			{Key: "present_today", Label: "Present Today", Value: float64(totalPresent)},
			{Key: "late_today", Label: "Late Today", Value: float64(late)},
			{Key: "absent_today", Label: "Absent Today", Value: float64(absent)},
		},
		Items:        items,
		EmptyMessage: "No attendance activity for today.",
	}, true, nil
}

func (s *DashboardService) pendingApprovalsWidget(ctx context.Context, user *models.User, currentEmp *models.Employee) (dto.DashboardWidgetRead, bool, error) {
	hasAll := dashboardHasAny(user, permissions.PermLeaveApproveAll, permissions.PermLeaveApprove, permissions.PermLeaveManage)
	hasTeam := dashboardHas(user, permissions.PermLeaveApproveTeam)
	if !hasAll && !hasTeam {
		return dto.DashboardWidgetRead{}, false, nil
	}

	scope := "organization"
	var employeeIDs []uuid.UUID
	if !hasAll {
		if currentEmp == nil {
			return dto.DashboardWidgetRead{}, false, nil
		}
		subordinateSet, err := s.emps.SubordinateIDs(ctx, currentEmp.ID)
		if err != nil {
			return dto.DashboardWidgetRead{}, false, err
		}
		employeeIDs = make([]uuid.UUID, 0, len(subordinateSet))
		for id := range subordinateSet {
			employeeIDs = append(employeeIDs, id)
		}
		scope = "team"
	}

	rows, total, err := s.leaves.List(ctx, employeeIDs, []string{string(models.LeaveStatusPending)}, 1, 5)
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	items := make([]dto.DashboardItemRead, 0, len(rows))
	for i := range rows {
		items = append(items, s.leaveItem(ctx, &rows[i]))
	}
	return dto.DashboardWidgetRead{
		ID:    "pending_approvals",
		Title: "Pending Approvals",
		Scope: scope,
		Metrics: []dto.DashboardMetricRead{
			{Key: "pending_leave", Label: "Pending Leave", Value: float64(total)},
		},
		Items:        items,
		EmptyMessage: "No approvals are waiting.",
	}, true, nil
}

func (s *DashboardService) leaveSummaryWidget(ctx context.Context, user *models.User, currentEmp *models.Employee, today time.Time) (dto.DashboardWidgetRead, bool, error) {
	if !dashboardHasAny(
		user,
		permissions.PermLeaveRead,
		permissions.PermLeaveCreate,
		permissions.PermLeaveUpdate,
		permissions.PermLeaveDelete,
		permissions.PermLeaveCancel,
		permissions.PermLeaveManage,
		permissions.PermLeaveApprove,
		permissions.PermLeaveApproveTeam,
		permissions.PermLeaveApproveAll,
	) {
		return dto.DashboardWidgetRead{}, false, nil
	}

	hasOrgScope := dashboardHasAny(user, permissions.PermLeaveApproveAll, permissions.PermLeaveApprove, permissions.PermLeaveManage)
	hasTeamScope := dashboardHas(user, permissions.PermLeaveApproveTeam) && !hasOrgScope

	scope := "own"
	var employeeIDs []uuid.UUID
	if hasOrgScope {
		scope = "organization"
	} else if hasTeamScope {
		if currentEmp == nil {
			return dto.DashboardWidgetRead{}, false, nil
		}
		subordinateSet, err := s.emps.SubordinateIDs(ctx, currentEmp.ID)
		if err != nil {
			return dto.DashboardWidgetRead{}, false, err
		}
		employeeIDs = make([]uuid.UUID, 0, len(subordinateSet))
		for id := range subordinateSet {
			employeeIDs = append(employeeIDs, id)
		}
		scope = "team"
	} else {
		if currentEmp == nil {
			return dto.DashboardWidgetRead{}, false, nil
		}
		employeeIDs = []uuid.UUID{currentEmp.ID}
	}

	_, pending, err := s.leaves.List(ctx, employeeIDs, []string{string(models.LeaveStatusPending)}, 1, 1)
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	upcoming, upcomingTotal, err := s.leaves.UpcomingScoped(ctx, employeeIDs, today, 5)
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	items := make([]dto.DashboardItemRead, 0, len(upcoming))
	for i := range upcoming {
		items = append(items, s.leaveItem(ctx, &upcoming[i]))
	}

	metrics := []dto.DashboardMetricRead{
		{Key: "pending_requests", Label: "Pending Requests", Value: float64(pending)},
		{Key: "upcoming_leave", Label: "Upcoming Leave", Value: float64(upcomingTotal)},
	}
	actions := []dto.DashboardActionRead{}
	if scope == "own" {
		balance, err := s.leaveBalance(ctx, currentEmp.ID, today.Year())
		if err != nil {
			return dto.DashboardWidgetRead{}, false, err
		}
		metrics = append([]dto.DashboardMetricRead{
			{Key: "annual_remaining", Label: "Annual Remaining", Value: balance.AnnualRemaining},
			{Key: "sick_remaining", Label: "Sick Remaining", Value: balance.SickRemaining},
		}, metrics...)
		if dashboardHas(user, permissions.PermLeaveCreate) {
			actions = append(actions, dto.DashboardActionRead{
				Key:   "create_leave_request",
				Label: "Create Leave Request",
				URL:   "/leave-requests/new",
			})
		}
	}

	return dto.DashboardWidgetRead{
		ID:           "leave_summary",
		Title:        "Leave Summary",
		Scope:        scope,
		Metrics:      metrics,
		Items:        items,
		Actions:      actions,
		EmptyMessage: "No upcoming leave.",
	}, true, nil
}

func (s *DashboardService) announcementsWidget(ctx context.Context, user *models.User, currentEmp *models.Employee) (dto.DashboardWidgetRead, bool, error) {
	if !dashboardHasAny(user, permissions.PermAnnounceRead, permissions.PermAnnounceManage) {
		return dto.DashboardWidgetRead{}, false, nil
	}
	asAdmin := dashboardHas(user, permissions.PermAnnounceManage)
	if !asAdmin && currentEmp == nil {
		return dto.DashboardWidgetRead{}, false, nil
	}
	filter := repositories.AnnouncementListFilter{
		AsAdmin:  asAdmin,
		Statuses: []models.AnnouncementStatus{models.AnnouncementStatusPublished},
		Scope:    "targeted-at-me",
		Page:     1,
		PageSize: 5,
	}
	if currentEmp != nil {
		filter.CurrentEmployeeID = currentEmp.ID
		filter.CurrentDepartmentID = currentEmp.DepartmentID
	}
	rows, total, err := s.announcements.List(ctx, filter)
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	items := make([]dto.DashboardItemRead, 0, len(rows))
	for i := range rows {
		date := rows[i].CreatedAt
		if rows[i].PublishedAt != nil {
			date = *rows[i].PublishedAt
		}
		items = append(items, dto.DashboardItemRead{
			ID:     rows[i].ID.String(),
			Title:  rows[i].Title,
			Status: string(rows[i].Status),
			Date:   &date,
			URL:    "/announcements/" + rows[i].ID.String(),
		})
	}
	return dto.DashboardWidgetRead{
		ID:    "announcements",
		Title: "Announcements",
		Scope: dashboardScope(asAdmin, "organization", "targeted"),
		Metrics: []dto.DashboardMetricRead{
			{Key: "latest_sent", Label: "Latest Sent", Value: float64(total)},
		},
		Items:        items,
		EmptyMessage: "No announcements yet.",
	}, true, nil
}

func (s *DashboardService) holidaysWorkdaysWidget(ctx context.Context, user *models.User, today time.Time) (dto.DashboardWidgetRead, bool, error) {
	canHolidays := dashboardHasAny(user, permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage)
	canWorkdays := dashboardHas(user, permissions.PermOrgWorkdaysView)
	if !canHolidays && !canWorkdays {
		return dto.DashboardWidgetRead{}, false, nil
	}
	metrics := []dto.DashboardMetricRead{}
	items := []dto.DashboardItemRead{}
	if canHolidays {
		to := today.AddDate(0, 0, 7)
		rows, err := s.holidays.FindInRange(ctx, today, to)
		if err != nil {
			return dto.DashboardWidgetRead{}, false, err
		}
		metrics = append(metrics, dto.DashboardMetricRead{Key: "upcoming_holidays", Label: "Upcoming Holidays", Value: float64(len(rows))})
		for i := range rows {
			date := rows[i].FromDate
			items = append(items, dto.DashboardItemRead{
				ID:    rows[i].ID.String(),
				Title: rows[i].Name,
				Date:  &date,
				URL:   "/holidays",
			})
		}
	}
	if canWorkdays {
		metrics = append(metrics, dto.DashboardMetricRead{
			Key:   "current_month_workdays",
			Label: "Current Month Workdays",
			Value: float64(countWeekdaysInMonth(today)),
		})
	}
	return dto.DashboardWidgetRead{
		ID:           "holidays_workdays",
		Title:        "Holidays & Workdays",
		Scope:        "organization",
		Metrics:      metrics,
		Items:        items,
		EmptyMessage: "No holidays in the next 7 days.",
	}, true, nil
}

func (s *DashboardService) workforceSummaryWidget(ctx context.Context, user *models.User, today time.Time) (dto.DashboardWidgetRead, bool, error) {
	if !dashboardHasAny(user, permissions.PermEmployeesRead, permissions.PermUsersRead) {
		return dto.DashboardWidgetRead{}, false, nil
	}
	active := true
	emps, total, err := s.emps.List(ctx, dto.EmployeeListQuery{Page: 1, PageSize: 1000, IsActive: &active})
	if err != nil {
		return dto.DashboardWidgetRead{}, false, err
	}
	joiners := 0
	items := []dto.DashboardItemRead{}
	for i := range emps {
		if emps[i].JoinDate == nil || emps[i].JoinDate.Year() != today.Year() || emps[i].JoinDate.Month() != today.Month() {
			continue
		}
		joiners++
		if len(items) >= 5 {
			continue
		}
		date := *emps[i].JoinDate
		items = append(items, dto.DashboardItemRead{
			ID:    emps[i].ID.String(),
			Title: fallbackEmployeeName(&emps[i]),
			Date:  &date,
			URL:   "/employees/" + emps[i].ID.String(),
		})
	}
	return dto.DashboardWidgetRead{
		ID:    "workforce_summary",
		Title: "Workforce Summary",
		Scope: "organization",
		Metrics: []dto.DashboardMetricRead{
			{Key: "active_employees", Label: "Active Employees", Value: float64(total)},
			{Key: "current_month_joiners", Label: "Current Month Joiners", Value: float64(joiners)},
		},
		Items:        items,
		EmptyMessage: "No new joiners this month.",
	}, true, nil
}

func (s *DashboardService) leaveBalance(ctx context.Context, employeeID uuid.UUID, year int) (dto.LeaveBalanceSummary, error) {
	grouped, err := s.leaves.SumApprovedDays(ctx, employeeID, year)
	if err != nil {
		return dto.LeaveBalanceSummary{}, err
	}
	annualQuota := 12.0
	sickQuota := 6.0
	q, err := s.quota.GetByEmployee(ctx, employeeID)
	if err != nil {
		return dto.LeaveBalanceSummary{}, err
	}
	if q != nil {
		annualQuota = q.AnnualLeaveQuota
		sickQuota = q.SickLeaveQuota
	}
	annualUsed := grouped[models.LeaveTypeAnnual].Days
	sickUsed := grouped[models.LeaveTypeSick].Days
	var totalCount int64
	for _, v := range grouped {
		totalCount += v.Count
	}
	return dto.LeaveBalanceSummary{
		Year:            year,
		AnnualQuota:     annualQuota,
		AnnualUsed:      annualUsed,
		AnnualRemaining: annualQuota - annualUsed,
		SickQuota:       sickQuota,
		SickUsed:        sickUsed,
		SickRemaining:   sickQuota - sickUsed,
		LeavesThisYear:  int(totalCount),
	}, nil
}

func (s *DashboardService) leaveItem(ctx context.Context, lr *models.LeaveRequest) dto.DashboardItemRead {
	name := "Employee"
	if emp, err := s.emps.FindByID(ctx, lr.EmployeeID); err == nil && emp != nil {
		name = fallbackEmployeeName(emp)
	}
	date := lr.FromDate
	return dto.DashboardItemRead{
		ID:          lr.ID.String(),
		Title:       fmt.Sprintf("%s - %s leave", name, string(lr.LeaveType)),
		Description: lr.Reason,
		Status:      string(lr.Status),
		Date:        &date,
		URL:         "/leave-requests/" + lr.ID.String(),
	}
}

func dashboardHasAny(user *models.User, perms ...permissions.Permission) bool {
	for _, p := range perms {
		if dashboardHas(user, p) {
			return true
		}
	}
	return false
}

func dashboardHas(user *models.User, perm permissions.Permission) bool {
	if user == nil {
		return false
	}
	for _, role := range user.Roles {
		for _, raw := range role.Permissions {
			if raw == string(permissions.PermAll) || raw == string(perm) {
				return true
			}
		}
	}
	return false
}

func dashboardGreetingName(user *models.User, emp *models.Employee) string {
	if emp != nil {
		if name := fallbackEmployeeName(emp); name != "" {
			return name
		}
	}
	if user != nil {
		return user.Email
	}
	return ""
}

func fallbackEmployeeName(emp *models.Employee) string {
	if emp == nil {
		return ""
	}
	if name := strings.TrimSpace(emp.FullName()); name != "" {
		return name
	}
	if emp.User != nil && strings.TrimSpace(emp.User.Email) != "" {
		return emp.User.Email
	}
	return emp.ID.String()
}

func dashboardScope(admin bool, adminScope, userScope string) string {
	if admin {
		return adminScope
	}
	return userScope
}

func countWeekdaysInMonth(day time.Time) int {
	from := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
	to := from.AddDate(0, 1, 0)
	count := 0
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		if isWorkday(d) {
			count++
		}
	}
	return count
}
