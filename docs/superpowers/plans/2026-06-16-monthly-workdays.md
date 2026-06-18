# Monthly Workdays Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `GET /api/v1/workdays?year=<year>` — a single read-only endpoint that returns workday counts for each month of a year, computed live from the holiday calendar.

**Architecture:** No migration, no new DB table. A new `WorkdayService` fetches all holidays for the year via a new `HolidayRepository.FindByYear` method, then computes the 12-month breakdown in pure Go (calendar math + holiday clamping). A thin `WorkdayHandler` wraps it with query-param binding and permission gate.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, `testify` (integration tests against `exnodes_hrm_test` DB)

---

## ⚠️ Calendar note (DR TS-01 has an error)

The DR's test scenario TS-01 states "March 2026: 4 Sat + 4 Sun = 8 weekends, Workdays = 23". This is **incorrect**: March 1, 2026 is a Sunday, so March has 4 Saturdays + 5 Sundays = **9 weekends**, giving **22 workdays**. Use Go's `time.Weekday()` as the source of truth — never hardcode weekend counts from the DR.

---

## File Map

| File | Action | What changes |
|---|---|---|
| `internal/dto/workday.go` | Create | `WorkdayQuery`, `MonthWorkdaysRead`, `WorkdayTotalRead`, `WorkdayYearRead` |
| `internal/repositories/holiday_repo.go` | Modify | Add `FindByYear(ctx, year int)` to interface + impl |
| `internal/permissions/registry.go` | Modify | Add `PermOrgWorkdaysView` constant, `AllPermissions()`, `PermissionGroups` |
| `internal/services/seed_service.go` | Modify | Seed `PermOrgWorkdaysView` to all 5 roles |
| `internal/services/workday_service.go` | Create | `WorkdayService` with `GetYear(ctx, year int)` |
| `internal/services/workday_service_test.go` | Create | 7 integration tests |
| `internal/handlers/workday_handler.go` | Create | `WorkdayHandler` with `GetYear` HTTP handler |
| `cmd/server/main.go` | Modify | Wire `WorkdayService`, `WorkdayHandler`, register route |

---

## Task 1: DTOs

**Files:**
- Create: `internal/dto/workday.go`

- [ ] **Step 1: Create the DTO file**

```go
package dto

// WorkdayQuery is the GET /workdays query string.
type WorkdayQuery struct {
	Year int `form:"year" binding:"required,min=2000,max=2100"`
}

// MonthWorkdaysRead is one row in the monthly workday table.
type MonthWorkdaysRead struct {
	Month     int    `json:"month"`
	MonthName string `json:"month_name"`
	TotalDays int    `json:"total_days"`
	Weekends  int    `json:"weekends"`
	Holidays  int    `json:"holidays"`
	Workdays  int    `json:"workdays"`
}

// WorkdayTotalRead is the summed totals row.
type WorkdayTotalRead struct {
	TotalDays int `json:"total_days"`
	Weekends  int `json:"weekends"`
	Holidays  int `json:"holidays"`
	Workdays  int `json:"workdays"`
}

// WorkdayYearRead is the top-level response payload.
type WorkdayYearRead struct {
	Year   int                 `json:"year"`
	Months []MonthWorkdaysRead `json:"months"`
	Total  WorkdayTotalRead    `json:"total"`
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/dto/...
```

Expected: no output (clean build).

- [ ] **Step 3: Commit**

```bash
git add internal/dto/workday.go
git commit -m "feat(workdays): WorkdayQuery and WorkdayYearRead DTOs"
```

---

## Task 2: HolidayRepository.FindByYear

**Files:**
- Modify: `internal/repositories/holiday_repo.go`

`FindByYear` returns all non-deleted holidays for a given year, ordered by `from_date ASC`. It is used by `WorkdayService` — semantically cleaner than calling `FindInRange` with a full-year range.

- [ ] **Step 1: Add the method to the interface**

In `internal/repositories/holiday_repo.go`, add `FindByYear` to the `HolidayRepository` interface (after the existing `ExistsByNameAndYear` line):

```go
// FindByYear returns all non-deleted holidays for the given year, ordered by from_date ASC.
// Used by WorkdayService to compute monthly workday counts.
FindByYear(ctx context.Context, year int) ([]models.Holiday, error)
```

The full interface block after the change:

```go
type HolidayRepository interface {
	List(ctx context.Context, q HolidayListQuery) ([]models.Holiday, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Holiday, error)
	Create(ctx context.Context, h *models.Holiday) error
	Update(ctx context.Context, h *models.Holiday) error
	Delete(ctx context.Context, id uuid.UUID) error

	// FindInRange returns non-deleted holidays whose range overlaps [from, to].
	// Used at leave creation time and during recalculation.
	FindInRange(ctx context.Context, from, to time.Time) ([]models.Holiday, error)

	// YearsWithHolidays returns distinct years that have at least one non-deleted holiday.
	YearsWithHolidays(ctx context.Context) ([]int, error)

	// ExistsByNameAndYear checks for a duplicate name in the same year.
	// excludeID, if non-nil, skips that row (used during Update).
	ExistsByNameAndYear(ctx context.Context, name string, year int, excludeID *uuid.UUID) (bool, error)

	// FindByYear returns all non-deleted holidays for the given year, ordered by from_date ASC.
	// Used by WorkdayService to compute monthly workday counts.
	FindByYear(ctx context.Context, year int) ([]models.Holiday, error)
}
```

- [ ] **Step 2: Add the implementation**

At the end of `internal/repositories/holiday_repo.go`, add:

```go
func (r *holidayRepo) FindByYear(ctx context.Context, year int) ([]models.Holiday, error) {
	var rows []models.Holiday
	err := r.base(ctx).Where("year = ?", year).Order("from_date ASC").Find(&rows).Error
	return rows, err
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/repositories/...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add internal/repositories/holiday_repo.go
git commit -m "feat(workdays): HolidayRepository.FindByYear"
```

---

## Task 3: Permission constant + seed

**Files:**
- Modify: `internal/permissions/registry.go`
- Modify: `internal/services/seed_service.go`

### 3a: registry.go

- [ ] **Step 1: Add the constant**

In `internal/permissions/registry.go`, after the `PermOrgHolidaysManage` line:

```go
// Workdays (Monthly Workdays summary — read-only, no manage counterpart)
PermOrgWorkdaysView Permission = "organization:workdays_view"
```

- [ ] **Step 2: Add to AllPermissions()**

In the `AllPermissions()` slice, add `PermOrgWorkdaysView` after `PermOrgHolidaysManage`:

```go
PermOrgHolidaysView, PermOrgHolidaysManage,
PermOrgWorkdaysView,
PermAnnounceManage,
```

- [ ] **Step 3: Add a PermissionGroup**

In the `PermissionGroups` var, add a new group after the `organization_holidays` group:

```go
{
	Resource: "organization_workdays", Label: "Monthly Workdays",
	Permissions: []PermissionItem{
		{PermOrgWorkdaysView, "View Monthly Workdays", "View the monthly workday summary table"},
	},
},
```

### 3b: seed_service.go

- [ ] **Step 4: Seed to all 5 roles**

In `internal/services/seed_service.go`, add `permissions.PermOrgWorkdaysView` to each of the 4 non-Super-Admin roles (Super Admin has `*` which already covers it).

**Admin** — after `permissions.PermOrgHolidaysManage`:
```go
permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage,
permissions.PermOrgWorkdaysView,
```

**HR Manager** — after `permissions.PermOrgHolidaysManage`:
```go
permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage,
permissions.PermOrgWorkdaysView,
```

**Manager** — after `permissions.PermOrgHolidaysView`:
```go
permissions.PermOrgHolidaysView,
permissions.PermOrgWorkdaysView,
```

**Employee** — after `permissions.PermOrgHolidaysView`:
```go
permissions.PermOrgHolidaysView,
permissions.PermOrgWorkdaysView,
```

- [ ] **Step 5: Verify it compiles and existing tests pass**

```bash
go build ./internal/permissions/... ./internal/services/...
go test ./internal/services/ -run TestPermission -v
```

Expected: `TestPermissionGroupsContainsAll` passes (it verifies every constant in `AllPermissions()` is in a group, and every group item is in `AllPermissions()`).

- [ ] **Step 6: Commit**

```bash
git add internal/permissions/registry.go internal/services/seed_service.go
git commit -m "feat(workdays): PermOrgWorkdaysView — registry + seed all 5 roles"
```

---

## Task 4: WorkdayService + integration tests

**Files:**
- Create: `internal/services/workday_service.go`
- Create: `internal/services/workday_service_test.go`

### Key algorithm

For each month M (1–12):
- `totalDays` = `time.Date(year, time.Month(m+1), 0, ...).Day()` — day 0 of the next month = last day of M; leap-year aware automatically.
- `weekends` = count of `time.Saturday` + `time.Sunday` in M by iterating each day.
- `holidays` = for each holiday, clamp `[h.FromDate, h.ToDate]` to `[firstOfM, lastOfM]`; if the clamped range is non-empty, add the day count.
- `workdays` = `totalDays − weekends − holidays`

Holiday-weekend overlap is **not deduplicated** — this is intentional (per AC-05).

### 4a: Write tests first

- [ ] **Step 1: Create the test file**

```go
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
// and Workdays = TotalDays − Weekends.
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
	// workdays = TotalDays − Weekends − 1
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
// Jan 30–Feb 2 contributes 2 days to January and 2 days to February.
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
```

- [ ] **Step 2: Run tests — expect compile failure (service not yet defined)**

```bash
go test ./internal/services/ -run TestWorkday -v
```

Expected: compile error — `services.WorkdayService` and `services.NewWorkdayService` undefined.

### 4b: Implement the service

- [ ] **Step 3: Create the service**

```go
package services

import (
	"context"
	"time"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/dto"
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

		wd := td - we - h

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
			Workdays:  totalDays - totalWeekends - totalHolidays,
		},
	}, nil
}
```

- [ ] **Step 4: Run tests — all must pass**

```bash
go test ./internal/services/ -run TestWorkday -v
```

Expected: all 7 tests PASS. If any fail, fix the computation before proceeding.

- [ ] **Step 5: Run full test suite to confirm no regressions**

```bash
go test ./...
```

Expected: same pass/skip/fail counts as before this task.

- [ ] **Step 6: Commit**

```bash
git add internal/services/workday_service.go internal/services/workday_service_test.go
git commit -m "feat(workdays): WorkdayService.GetYear + 7 integration tests"
```

---

## Task 5: WorkdayHandler + route + Swagger

**Files:**
- Create: `internal/handlers/workday_handler.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Create the handler**

```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// WorkdayHandler handles /api/v1/workdays endpoints.
type WorkdayHandler struct {
	svc *services.WorkdayService
}

// NewWorkdayHandler constructs a WorkdayHandler.
func NewWorkdayHandler(svc *services.WorkdayService) *WorkdayHandler {
	return &WorkdayHandler{svc: svc}
}

// GetYear godoc
// @Summary      Monthly workday summary for a year
// @Description  Returns the workday count for each month of the given year, computed live from the company holiday calendar. Workdays = Total Days − Weekends − Holidays. A holiday falling on a weekend still reduces Workdays.
// @Tags         workdays
// @Security     BearerAuth
// @Produce      json
// @Param        year  query  int  true  "calendar year (2000–2100)"
// @Success      200  {object}  dto.Response[dto.WorkdayYearRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      401  {object}  dto.Response[any]
// @Failure      403  {object}  dto.Response[any]
// @Router       /workdays [get]
func (h *WorkdayHandler) GetYear(c *gin.Context) {
	var q dto.WorkdayQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.GetYear(c.Request.Context(), q.Year)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.WorkdayYearRead]{Success: true, Data: out})
}
```

- [ ] **Step 2: Wire in main.go**

In `cmd/server/main.go`, after the `holidaySvc :=` line (around line 130), add:

```go
workdaySvc := services.NewWorkdayService(holidayRepo)
```

After the `holidayH :=` line (around line 155), add:

```go
workdayH := handlers.NewWorkdayHandler(workdaySvc)
```

In the routes section, after the `// ---- /holidays ----` block, add:

```go
// ---- /workdays ----
authed.GET("/workdays", middleware.RequirePerms(authSvc, permissions.PermOrgWorkdaysView), workdayH.GetYear)
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 4: Regenerate Swagger**

```bash
make swag
```

Expected: `docs/swagger/swagger.json` and `docs/swagger/swagger.yaml` updated. Check that `/workdays` appears as a GET endpoint.

- [ ] **Step 5: Run full test suite**

```bash
go test ./...
```

Expected: all tests pass (same result as before Task 5).

- [ ] **Step 6: Format and vet**

```bash
make fmt && make vet
```

Expected: no output (clean).

- [ ] **Step 7: Commit**

```bash
git add internal/handlers/workday_handler.go cmd/server/main.go docs/swagger/
git commit -m "feat(workdays): WorkdayHandler + GET /api/v1/workdays route + Swagger regen"
```

---

## Verification checklist (after all tasks)

Run these after all 5 tasks are complete:

```bash
# Server must start cleanly
make run
```

```bash
# Get a token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"<admin-pw>"}' | jq -r '.data.access_token')

# Happy path — current year
curl -s "http://localhost:8080/api/v1/workdays?year=2026" \
  -H "Authorization: Bearer $TOKEN" | jq .

# Must have 12 items in months, a total row, all integers
# 400 — missing year
curl -s "http://localhost:8080/api/v1/workdays" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 400 — year out of range
curl -s "http://localhost:8080/api/v1/workdays?year=1999" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 403 — using a token that lacks the permission (if you have one)
```

Expected for happy path:
```json
{
  "success": true,
  "data": {
    "year": 2026,
    "months": [
      { "month": 1, "month_name": "January", "total_days": 31, ... },
      ...
      { "month": 12, "month_name": "December", "total_days": 31, ... }
    ],
    "total": { "total_days": 365, ... }
  }
}
```
