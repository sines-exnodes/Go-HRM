package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

func makeWorkdaySvc(t *testing.T) *services.WorkdayService {
	t.Helper()
	return services.NewWorkdayService(repositories.NewHolidayRepository(testDB))
}

// makeHoliday is a test helper that inserts a holiday and fails the test on error.
func makeHoliday(t *testing.T, year int, name string, from, to time.Time) {
	t.Helper()
	err := repositories.NewHolidayRepository(testDB).Create(context.Background(), &models.Holiday{
		Year:     year,
		Name:     name,
		FromDate: from,
		ToDate:   to,
	})
	require.NoError(t, err)
}

// TestWorkday_NoHolidays verifies that with no holidays all months show Holidays=0
// and Workdays = TotalDays - Weekends.
func TestWorkday_NoHolidays(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2026)
	require.Nil(t, aerr)
	require.NotNil(t, out)
	assert.Equal(t, 2026, out.Year)
	require.Len(t, out.Months, 12)

	for _, m := range out.Months {
		assert.Equal(t, 0, m.Holidays, "month %d should have 0 holidays", m.Month)
		assert.Equal(t, m.TotalDays-m.Weekends, m.Workdays, "workdays mismatch for month %d", m.Month)
	}

	// Verify February (non-leap year 2026)
	feb := out.Months[1]
	assert.Equal(t, 2, feb.Month)
	assert.Equal(t, "February", feb.MonthName)
	assert.Equal(t, 28, feb.TotalDays)

	// Total row sums correctly
	var sumDays, sumWE, sumHol, sumWD int
	for _, m := range out.Months {
		sumDays += m.TotalDays
		sumWE += m.Weekends
		sumHol += m.Holidays
		sumWD += m.Workdays
	}
	assert.Equal(t, 365, out.Total.TotalDays)
	assert.Equal(t, sumWE, out.Total.Weekends)
	assert.Equal(t, 0, out.Total.Holidays)
	assert.Equal(t, sumWD, out.Total.Workdays)
}

// TestWorkday_HolidayOnWeekday verifies a holiday on a weekday subtracts from Workdays.
// Jan 1 2026 is Thursday.
func TestWorkday_HolidayOnWeekday(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	makeHoliday(t, 2026, "New Year", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2026)
	require.Nil(t, aerr)

	jan := out.Months[0]
	assert.Equal(t, 31, jan.TotalDays)
	assert.Equal(t, 1, jan.Holidays)
	// workdays = TotalDays - Weekends - 1
	assert.Equal(t, jan.TotalDays-jan.Weekends-1, jan.Workdays)
}

// TestWorkday_HolidayOnWeekend verifies that a holiday on a weekend still counts
// as a holiday and reduces Workdays (AC-05: no deduplication).
// Jan 3 2026 is Saturday.
func TestWorkday_HolidayOnWeekend(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	makeHoliday(t, 2026, "Weekend Holiday", time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC))

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2026)
	require.Nil(t, aerr)

	jan := out.Months[0]
	assert.Equal(t, 1, jan.Holidays, "holiday on weekend must still count in Holidays")
	// Workdays is further reduced even though the holiday falls on a weekend
	assert.Equal(t, jan.TotalDays-jan.Weekends-1, jan.Workdays, "workdays must subtract both weekend and holiday (AC-05)")
}

// TestWorkday_CrossMonthHoliday verifies that a holiday spanning two months is split:
// Jan 30-Feb 2 contributes 2 days to January and 2 days to February.
func TestWorkday_CrossMonthHoliday(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	makeHoliday(t, 2026, "Cross Month",
		time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	)

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2026)
	require.Nil(t, aerr)

	jan := out.Months[0]
	assert.Equal(t, 2, jan.Holidays, "Jan 30 and Jan 31 = 2 days in January")

	feb := out.Months[1]
	assert.Equal(t, 2, feb.Holidays, "Feb 1 and Feb 2 = 2 days in February")
}

// TestWorkday_LeapYear verifies February shows 29 days in a leap year.
func TestWorkday_LeapYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2024)
	require.Nil(t, aerr)

	feb := out.Months[1]
	assert.Equal(t, 29, feb.TotalDays)
	assert.Equal(t, 365+1, out.Total.TotalDays) // 366 days
}

// TestWorkday_NonLeapYear verifies February shows 28 days in a non-leap year.
func TestWorkday_NonLeapYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2025)
	require.Nil(t, aerr)

	feb := out.Months[1]
	assert.Equal(t, 28, feb.TotalDays)
	assert.Equal(t, 365, out.Total.TotalDays)
}

// TestWorkday_TotalRow verifies the Total row is the sum of all monthly values.
func TestWorkday_TotalRow(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	// Two holidays in different months
	makeHoliday(t, 2026, "Holiday A",
		time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	)
	makeHoliday(t, 2026, "Holiday B",
		time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	)

	svc := makeWorkdaySvc(t)
	out, aerr := svc.GetYear(context.Background(), 2026)
	require.Nil(t, aerr)

	var sumDays, sumWE, sumHol, sumWD int
	for _, m := range out.Months {
		sumDays += m.TotalDays
		sumWE += m.Weekends
		sumHol += m.Holidays
		sumWD += m.Workdays
	}
	assert.Equal(t, sumDays, out.Total.TotalDays)
	assert.Equal(t, sumWE, out.Total.Weekends)
	assert.Equal(t, 2, out.Total.Holidays)
	assert.Equal(t, sumWD, out.Total.Workdays)
}
