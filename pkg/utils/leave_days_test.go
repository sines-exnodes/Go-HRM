package utils_test

import (
	"testing"
	"time"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func d(y, m, day int) time.Time {
	return time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)
}

func TestCalcLeaveDays_NoHolidays_FullDay(t *testing.T) {
	result := utils.CalcLeaveDays(d(2025, 4, 28), d(2025, 4, 30), models.LeavePeriodFullDay, nil)
	assert.Equal(t, 3.0, result)
}

func TestCalcLeaveDays_FullDay_ExcludesHoliday(t *testing.T) {
	holidays := []utils.DateRange{{From: d(2025, 4, 30), To: d(2025, 4, 30)}}
	result := utils.CalcLeaveDays(d(2025, 4, 28), d(2025, 4, 30), models.LeavePeriodFullDay, holidays)
	assert.Equal(t, 2.0, result)
}

func TestCalcLeaveDays_HalfDay_NoHoliday(t *testing.T) {
	result := utils.CalcLeaveDays(d(2025, 4, 29), d(2025, 4, 29), models.LeavePeriodMorningHalf, nil)
	assert.Equal(t, 0.5, result)
}

func TestCalcLeaveDays_HalfDay_HalfExcluded(t *testing.T) {
	// AC-12: half-day on a holiday date loses 0.5 → 0.0
	holidays := []utils.DateRange{{From: d(2025, 4, 30), To: d(2025, 4, 30)}}
	result := utils.CalcLeaveDays(d(2025, 4, 30), d(2025, 4, 30), models.LeavePeriodMorningHalf, holidays)
	assert.Equal(t, 0.0, result)
}

func TestCalcLeaveDays_MultiDayHolidayOverlap(t *testing.T) {
	// Leave: Jan 28–Feb 3 (7 days). Holiday: Jan 27–Feb 2 (6 overlap days). Result: 1.
	holidays := []utils.DateRange{{From: d(2025, 1, 27), To: d(2025, 2, 2)}}
	result := utils.CalcLeaveDays(d(2025, 1, 28), d(2025, 2, 3), models.LeavePeriodFullDay, holidays)
	assert.Equal(t, 1.0, result)
}

func TestCalcLeaveDays_ClampAtZero(t *testing.T) {
	// Leave entirely within holiday — must not return negative.
	holidays := []utils.DateRange{{From: d(2025, 4, 28), To: d(2025, 5, 5)}}
	result := utils.CalcLeaveDays(d(2025, 4, 29), d(2025, 4, 30), models.LeavePeriodFullDay, holidays)
	assert.Equal(t, 0.0, result)
}

func TestCalcLeaveDays_InvertedRange_ReturnsZero(t *testing.T) {
	// from > to is invalid input; both full-day and half-day must return 0, not 0.5.
	assert.Equal(t, 0.0, utils.CalcLeaveDays(d(2025, 5, 5), d(2025, 5, 3), models.LeavePeriodFullDay, nil))
	assert.Equal(t, 0.0, utils.CalcLeaveDays(d(2025, 5, 5), d(2025, 5, 3), models.LeavePeriodMorningHalf, nil))
}
