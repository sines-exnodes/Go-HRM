# Monthly Workdays API — Design Spec

**Date:** 2026-06-16  
**Status:** Approved  
**Feature:** EP-009 US-004 — Monthly Workdays  
**DR:** `ba-requirements/docs/PLATFORMS/WEB-APP/EP-009-organization-settings/US-004-monthly-workdays/details/DR-009-004-01-monthly-workdays.md`

---

## Overview

A single read-only endpoint that returns the workday count for each month of a given year, computed live from the company holiday calendar (US-003) and weekend exclusion. No new DB tables. No stored values — all calculations performed on every request.

**Business rule:** Workdays = Total Calendar Days − Weekends − Company Holidays. Holiday-weekend overlap is **not** deduplicated — a holiday falling on a Saturday still subtracts from Workdays (this is intentional per the DR, AC-05).

---

## Endpoint

```
GET /api/v1/workdays?year=<year>
```

**Permission:** `organization:workdays_view`  
**Auth:** `Authorization: Bearer <token>` required  
**Seeded to:** All 5 roles (Super Admin via `*`, Admin, HR Manager, Manager, Employee)

### Query params

| Param | Type | Required | Validation |
|---|---|---|---|
| `year` | int | yes | 2000–2100; 400 if missing or out of range |

### Response — `200 OK`

```json
{
  "success": true,
  "data": {
    "year": 2026,
    "months": [
      { "month": 1, "month_name": "January",  "total_days": 31, "weekends": 8, "holidays": 1, "workdays": 22 },
      { "month": 2, "month_name": "February", "total_days": 28, "weekends": 8, "holidays": 0, "workdays": 20 },
      { "month": 3, "month_name": "March",    "total_days": 31, "weekends": 8, "holidays": 0, "workdays": 23 },
      { "month": 4, "month_name": "April",    "total_days": 30, "weekends": 8, "holidays": 1, "workdays": 21 },
      { "month": 5, "month_name": "May",      "total_days": 31, "weekends": 9, "holidays": 2, "workdays": 20 },
      { "month": 6, "month_name": "June",     "total_days": 30, "weekends": 8, "holidays": 0, "workdays": 22 },
      { "month": 7, "month_name": "July",     "total_days": 31, "weekends": 8, "holidays": 0, "workdays": 23 },
      { "month": 8, "month_name": "August",   "total_days": 31, "weekends": 9, "holidays": 0, "workdays": 22 },
      { "month": 9, "month_name": "September","total_days": 30, "weekends": 8, "holidays": 2, "workdays": 20 },
      { "month": 10,"month_name": "October",  "total_days": 31, "weekends": 8, "holidays": 0, "workdays": 23 },
      { "month": 11,"month_name": "November", "total_days": 30, "weekends": 8, "holidays": 0, "workdays": 22 },
      { "month": 12,"month_name": "December", "total_days": 31, "weekends": 9, "holidays": 1, "workdays": 21 }
    ],
    "total": {
      "total_days": 365,
      "weekends": 101,
      "holidays": 7,
      "workdays": 257
    }
  }
}
```

Always exactly 12 items in `months`, January → December. `total` sums all 12 rows. No pagination.

### Error responses

| Status | When |
|---|---|
| `400` | Missing `year` or outside 2000–2100 |
| `401` | Missing or expired token |
| `403` | Lacks `organization:workdays_view` |
| `500` | Unexpected DB error |

---

## Architecture

Layering: `WorkdayHandler → WorkdayService → HolidayRepository`

No new migration. No new repository. `HolidayRepository` gains one new method; all other dependencies reuse existing wired-up instances from `main.go`.

### Files

| File | Action | Responsibility |
|---|---|---|
| `internal/dto/workday.go` | Create | `WorkdayQuery`, `MonthWorkdaysRead`, `WorkdayYearRead` DTOs |
| `internal/services/workday_service.go` | Create | `WorkdayService` — `GetYear(ctx, year int)` |
| `internal/services/workday_service_test.go` | Create | Integration tests |
| `internal/handlers/workday_handler.go` | Create | `WorkdayHandler` — `GetYear` HTTP method |
| `internal/repositories/holiday_repo.go` | Modify | Add `FindByYear(ctx, year int)` to interface + impl |
| `internal/permissions/registry.go` | Modify | Add `PermOrgWorkdaysView` constant + `AllPermissions()` + `PermissionGroups()` |
| `internal/services/seed_service.go` | Modify | Seed `PermOrgWorkdaysView` to all 5 roles |
| `cmd/server/main.go` | Modify | Wire `WorkdayService`, `WorkdayHandler`, register route |

---

## Computation Rules

### `FindByYear(ctx, year int) ([]Holiday, error)`

New method on `HolidayRepository`. Returns all non-deleted holidays whose `year` column equals the requested year. No pagination. Used exclusively by `WorkdayService`.

```sql
SELECT * FROM holidays
WHERE year = ? AND is_deleted = false
ORDER BY from_date ASC
```

### Per-month computation (pure Go)

For each month M in `[January … December]`:

```
total_days  = days-in-month(year, M)           // time.Date(year, M+1, 0, ...) gives last day; leap-year aware
weekends    = count of Saturdays + Sundays in month M of year
holidays    = Σ clamp(h.from_date, h.to_date, firstDay(M), lastDay(M)) for each h in FindByYear(year)
workdays    = total_days − weekends − holidays
```

**Holiday clamping:** for a holiday with `from_date` and `to_date`, count only the days that fall within month M:

```
effective_from = max(h.from_date, firstDayOfMonth)
effective_to   = min(h.to_date, lastDayOfMonth)
if effective_from <= effective_to:
    holidays += (effective_to − effective_from).days + 1
```

This correctly handles:
- Single-day holidays (from_date == to_date)
- Multi-day holidays within one month
- Cross-month holidays (a holiday spanning Jan 30–Feb 2 contributes 2 days to January, 2 days to February)

**Holiday-weekend overlap rule (AC-05):** holidays and weekends are subtracted independently. A holiday falling on a Saturday is counted in `holidays` AND the Saturday is counted in `weekends` — both subtract from `workdays`. This is the intended business rule.

### Total row

Simple column sums across all 12 months. `total.workdays` = `total.total_days − total.weekends − total.holidays` (which equals sum of monthly `workdays` values, since the formula is linear).

---

## Permission

```go
PermOrgWorkdaysView Permission = "organization:workdays_view"
```

**Permission group:** `organization_workdays` — Label: "Monthly Workdays"  
**Seeded defaults:**

| Role | workdays_view |
|---|---|
| Super Admin | ✅ (wildcard `*`) |
| Admin | ✅ |
| HR Manager | ✅ |
| Manager | ✅ |
| Employee | ✅ |

No manage permission exists — this is a read-only computed page.

---

## DTOs

```go
// WorkdayQuery — parsed from query string
type WorkdayQuery struct {
    Year int `form:"year" binding:"required,min=2000,max=2100"`
}

// MonthWorkdaysRead — one month row in the response
type MonthWorkdaysRead struct {
    Month     int    `json:"month"`
    MonthName string `json:"month_name"`
    TotalDays int    `json:"total_days"`
    Weekends  int    `json:"weekends"`
    Holidays  int    `json:"holidays"`
    Workdays  int    `json:"workdays"`
}

// WorkdayTotalRead — summed totals row
type WorkdayTotalRead struct {
    TotalDays int `json:"total_days"`
    Weekends  int `json:"weekends"`
    Holidays  int `json:"holidays"`
    Workdays  int `json:"workdays"`
}

// WorkdayYearRead — top-level response data
type WorkdayYearRead struct {
    Year   int                 `json:"year"`
    Months []MonthWorkdaysRead `json:"months"`
    Total  WorkdayTotalRead    `json:"total"`
}
```

---

## Tests

Integration tests in `internal/services/workday_service_test.go` against the test DB. Holiday records are created via `holidayRepo.Create` in each test; `truncateAll` clears the `holidays` table between tests.

| Test | Setup | Expected |
|---|---|---|
| `TestGetYear_NoHolidays` | No holidays for 2026 | All months: Holidays=0, Workdays=TotalDays−Weekends |
| `TestGetYear_HolidayOnWeekday` | 1 holiday on Mon Jan 1 2026 | January: Holidays=1, Workdays=TotalDays−Weekends−1 |
| `TestGetYear_HolidayOnWeekend` | 1 holiday on Sat Jan 3 2026 | January: Weekends=8, Holidays=1, Workdays=31−8−1=22 (double-subtract) |
| `TestGetYear_CrossMonthHoliday` | Holiday Jan 30–Feb 2 (4 days) | January Holidays=2, February Holidays=2 |
| `TestGetYear_LeapYear` | Year 2024, no holidays | February: TotalDays=29 |
| `TestGetYear_NonLeapYear` | Year 2025, no holidays | February: TotalDays=28 |
| `TestGetYear_TotalRow` | 2 holidays in different months | Total.Holidays=2, Total.Workdays=sum of monthly Workdays |
| `TestGetYear_InvalidYear` | year=1999 or year=2101 | 400 error (validated at DTO binding layer) |

---

## Wiring (`cmd/server/main.go`)

```go
workdaySvc := services.NewWorkdayService(holidayRepo)
workdayH   := handlers.NewWorkdayHandler(workdaySvc)

// under authed group, with RequirePerms(PermOrgWorkdaysView):
authed.GET("/workdays", middleware.RequirePerms(permissions.PermOrgWorkdaysView), workdayH.GetYear)
```

`holidayRepo` is already constructed earlier in `main.go` for the holidays module — no new DB dependency.

---

## Out of Scope

- Workday overrides (no manual customization of computed values)
- Half-day holiday handling (holidays are whole calendar days only)
- Per-department or per-employee workday calendars (org-wide only)
- Export / download of the workday table
- Caching (live computation on every request is fast enough at 12 months)
- Custom non-working days beyond the company holiday calendar
