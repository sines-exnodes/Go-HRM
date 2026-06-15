package utils

import (
	"time"

	"github.com/exnodes/hrm-api/internal/models"
)

// DateRange is a closed [From, To] date interval (UTC, time component ignored).
type DateRange struct {
	From time.Time
	To   time.Time
}

// CalcLeaveDays returns the number of leave days consumed, excluding any
// calendar days that fall within a company holiday range.
//
// Full-day leave: each overlapping holiday calendar day subtracts 1.0.
// Half-day leave (single calendar day only): each overlapping holiday day subtracts 0.5 (AC-12).
// Result is clamped to 0 — it never returns negative.
// Precondition: from <= to. Inverted range returns 0.
func CalcLeaveDays(from, to time.Time, period models.LeavePeriod, holidays []DateRange) float64 {
	from = truncDay(from)
	to = truncDay(to)

	if to.Before(from) {
		return 0
	}

	calendarDays := int(to.Sub(from).Hours()/24) + 1

	holidayDays := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, h := range holidays {
			hFrom := truncDay(h.From)
			hTo := truncDay(h.To)
			if !d.Before(hFrom) && !d.After(hTo) {
				holidayDays++
				break
			}
		}
	}

	if models.IsHalfDayPeriod(period) {
		result := 0.5 - float64(holidayDays)*0.5
		if result < 0 {
			return 0
		}
		return result
	}
	result := float64(calendarDays - holidayDays)
	if result < 0 {
		return 0
	}
	return result
}

func truncDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
