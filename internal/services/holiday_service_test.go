package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func makeHolidaySvc(t *testing.T) *services.HolidayService {
	t.Helper()
	return services.NewHolidayService(
		repositories.NewHolidayRepository(testDB),
		repositories.NewHolidayTemplateRepository(testDB),
		repositories.NewLeaveRequestRepository(testDB),
	)
}

func TestHoliday_Create_HappyPath(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	out, affected, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)
	require.NotNil(t, out)
	assert.NotEqual(t, uuid.Nil, out.ID)
	assert.Equal(t, "Liberation Day", out.Name)
	assert.Equal(t, 1, out.TotalDays)
	assert.Equal(t, 0, affected)
}

func TestHoliday_Create_DuplicateName_SameYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	req := dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	}
	_, _, aerr := svc.Create(context.Background(), req)
	require.Nil(t, aerr)

	_, _, aerr = svc.Create(context.Background(), req)
	require.NotNil(t, aerr)
	assert.Equal(t, 409, aerr.HTTP)
}

func TestHoliday_Create_ToDateBeforeFromDate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	_, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Bad Holiday",
		FromDate: time.Date(2025, 5, 2, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.NotNil(t, aerr)
	assert.Equal(t, 400, aerr.HTTP)
}

func TestHoliday_Create_TriggersLeaveRecalculation(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	role := makeRole(t, "Employee", nil, false)
	user := makeUser(t, "emp1@test.com", "pw-Aa123456", role)
	emp := makeEmployee(t, user, "Test Employee")

	leaveRepo := repositories.NewLeaveRequestRepository(testDB)
	lr := &models.LeaveRequest{
		EmployeeID:  emp.ID,
		FromDate:    time.Date(2025, 4, 28, 0, 0, 0, 0, time.UTC),
		ToDate:      time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		TotalDays:   4.0,
		Reason:      "test",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, leaveRepo.Create(context.Background(), lr))

	svc := makeHolidaySvc(t)
	_, affected, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)
	assert.Equal(t, 1, affected)

	updated, err := leaveRepo.FindByID(context.Background(), lr.ID)
	require.NoError(t, err)
	assert.Equal(t, 3.0, updated.TotalDays)
}

func TestHoliday_Delete_ReturnsAffectedCount(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	role := makeRole(t, "Employee", nil, false)
	user := makeUser(t, "emp2@test.com", "pw-Aa123456", role)
	emp := makeEmployee(t, user, "Test Employee 2")

	leaveRepo := repositories.NewLeaveRequestRepository(testDB)
	lr := &models.LeaveRequest{
		EmployeeID:  emp.ID,
		FromDate:    time.Date(2025, 4, 29, 0, 0, 0, 0, time.UTC),
		ToDate:      time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		TotalDays:   2.0,
		Reason:      "test",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, leaveRepo.Create(context.Background(), lr))

	svc := makeHolidaySvc(t)
	out, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)

	affected, aerr := svc.Delete(context.Background(), out.ID)
	require.Nil(t, aerr)
	assert.Equal(t, 1, affected)

	restored, err := leaveRepo.FindByID(context.Background(), lr.ID)
	require.NoError(t, err)
	assert.Equal(t, 3.0, restored.TotalDays)
}

func TestHoliday_Delete_NotFound(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	_, aerr := svc.Delete(context.Background(), uuid.New())
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.HTTP)
}

func TestHoliday_List_YearScoped_SortedByFromDate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	for _, name := range []string{"Tết", "Liberation Day", "National Day"} {
		var from, to time.Time
		switch name {
		case "Tết":
			from = time.Date(2025, 1, 27, 0, 0, 0, 0, time.UTC)
			to = time.Date(2025, 2, 2, 0, 0, 0, 0, time.UTC)
		case "Liberation Day":
			from = time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC)
			to = time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC)
		default:
			from = time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)
			to = time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)
		}
		_, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
			Year: 2025, Name: name, FromDate: from, ToDate: to,
		})
		require.Nil(t, aerr)
	}

	page, aerr := svc.List(context.Background(), dto.HolidayListQuery{Year: 2025, Page: 1, PageSize: 10})
	require.Nil(t, aerr)
	assert.Equal(t, int64(3), page.Total)
	assert.Equal(t, "Tết", page.Items[0].Name)
}

func TestHoliday_Import_SkipsDuplicates(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)

	_, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2026,
		Name:     "Tết Dương Lịch",
		FromDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)

	templateRepo := repositories.NewHolidayTemplateRepository(testDB)
	templates, err := templateRepo.ListByYear(context.Background(), 2026)
	require.NoError(t, err)
	require.NotEmpty(t, templates, "migration must have seeded 2026 templates")

	ids := make([]uuid.UUID, len(templates))
	for i, tmpl := range templates {
		ids[i] = tmpl.ID
	}

	result, aerr := svc.Import(context.Background(), dto.HolidayImportRequest{
		Year:        2026,
		TemplateIDs: ids,
	})
	require.Nil(t, aerr)
	assert.Equal(t, len(templates)-1, result.Imported)
	assert.Equal(t, 1, result.Skipped)
}

func TestHoliday_GetYears_AlwaysIncludesCurrentYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	years, aerr := svc.GetYears(context.Background())
	require.Nil(t, aerr)
	currentYear := time.Now().UTC().Year()
	assert.Contains(t, years, currentYear)
}

func TestHoliday_Update_ChangeDates_Recalculates(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	role := makeRole(t, "Employee", nil, false)
	user := makeUser(t, fmt.Sprintf("emp3-%s@test.com", uuid.NewString()[:6]), "pw-Aa123456", role)
	emp := makeEmployee(t, user, "Test Employee 3")

	leaveRepo := repositories.NewLeaveRequestRepository(testDB)
	lr := &models.LeaveRequest{
		EmployeeID:  emp.ID,
		FromDate:    time.Date(2025, 4, 28, 0, 0, 0, 0, time.UTC),
		ToDate:      time.Date(2025, 5, 2, 0, 0, 0, 0, time.UTC),
		LeavePeriod: models.LeavePeriodFullDay,
		LeaveType:   models.LeaveTypeAnnual,
		TotalDays:   4.0,
		Reason:      "test",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, leaveRepo.Create(context.Background(), lr))

	svc := makeHolidaySvc(t)
	h, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)

	newTo := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	_, affected, aerr := svc.Update(context.Background(), h.ID, dto.HolidayUpdate{ToDate: &newTo})
	require.Nil(t, aerr)
	assert.Equal(t, 1, affected)

	updated, err := leaveRepo.FindByID(context.Background(), lr.ID)
	require.NoError(t, err)
	assert.Equal(t, 3.0, updated.TotalDays)
}
