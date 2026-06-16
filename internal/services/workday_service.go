package services

import (
	"context"
	"time"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/repositories"
)

// WorkdayService computes monthly workday summaries from the holiday calendar.
type WorkdayService struct {
	holidayRepo repositories.HolidayRepository
}

// NewWorkdayService constructs a WorkdayService.
func NewWorkdayService(holidayRepo repositories.HolidayRepository) *WorkdayService {
	return &WorkdayService{holidayRepo: holidayRepo}
}

// GetYear returns the 12-month workday breakdown for the given year.
// All values are computed live — no DB writes occur.
func (s *WorkdayService) GetYear(ctx context.Context, year int) (*dto.WorkdayYearRead, *apperrors.AppError) {
	holidays, err := s.holidayRepo.FindByYear(ctx, year)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}

	months := make([]dto.MonthWorkdaysRead, 12)
	var totalDays, totalWeekends, totalHolidays int

	for m := 1; m <= 12; m++ {
		// day 0 of the next month = last day of this month (leap-year aware)
		firstDay := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		lastDay := time.Date(year, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC)

		td := lastDay.Day()

		we := 0
		for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
			if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
				we++
			}
		}

		h := 0
		for _, hol := range holidays {
			effFrom := hol.FromDate
			if effFrom.Before(firstDay) {
				effFrom = firstDay
			}
			effTo := hol.ToDate
			if effTo.After(lastDay) {
				effTo = lastDay
			}
			if !effFrom.After(effTo) {
				h += int(effTo.Sub(effFrom)/(24*time.Hour)) + 1
			}
		}

		wd := td - we // holidays are informational only (DR-009-004-01 v1.1 AC-04)

		months[m-1] = dto.MonthWorkdaysRead{
			Month:     m,
			MonthName: time.Month(m).String(),
			TotalDays: td,
			Weekends:  we,
			Holidays:  h,
			Workdays:  wd,
		}

		totalDays += td
		totalWeekends += we
		totalHolidays += h
	}

	return &dto.WorkdayYearRead{
		Year:   year,
		Months: months,
		Total: dto.WorkdayTotalRead{
			TotalDays: totalDays,
			Weekends:  totalWeekends,
			Holidays:  totalHolidays,
			Workdays:  totalDays - totalWeekends,
		},
	}, nil
}
