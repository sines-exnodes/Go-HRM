# Holiday Management — Design Spec

**Date:** 2026-06-15  
**Status:** Approved  
**DR:** `ba-requirements/docs/PLATFORMS/WEB-APP/EP-009-organization-settings/US-003-holiday-management/details/DR-009-003-01-holiday-management.md`  
**Migration:** 000023  

---

## Goal

Add a company holiday calendar to the Go HRM API. Holidays are year-scoped records that:
1. Can be created, edited, deleted, and imported from a Vietnamese public holiday preset.
2. Are subtracted from leave day counts — both at leave creation time and whenever the holiday calendar changes (recalculation).

This closes the blocked item G7 from the attendance parity audit (holiday "H" cells and streak exclusion).

---

## Decisions

| # | Question | Decision |
|---|---|---|
| D1 | Which roles get `holidays_manage`? | Admin + HR Manager. All 5 roles get `holidays_view`. |
| D2 | Holidays subtracted at leave creation? | Yes — full formula applied at creation AND recalculation (DR Rule 4). |
| D3 | Vietnamese preset storage | Separate `holiday_templates` table, seeded in migration 000023. |
| D4 | Recalculation architecture | Shared `CalcLeaveDays()` utility; HolidayService injects LeaveRequestRepository; LeaveService injects HolidayRepository. No circular dependency. |

---

## Schema — Migration 000023

### `holidays` table

```sql
CREATE TABLE holidays (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    year        INTEGER     NOT NULL,
    name        TEXT        NOT NULL,
    from_date   DATE        NOT NULL,
    to_date     DATE        NOT NULL,
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enforce unique name per year among non-deleted rows (AC-03)
CREATE UNIQUE INDEX uq_holidays_name_year
    ON holidays (year, LOWER(name))
    WHERE is_deleted = FALSE;

-- Fast year-scoped queries
CREATE INDEX idx_holidays_year
    ON holidays (year, from_date)
    WHERE is_deleted = FALSE;

CREATE TRIGGER trg_holidays_set_updated_at
    BEFORE UPDATE ON holidays
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

`total_days` is **not stored** — computed as `(to_date - from_date + 1)` at query time (DR Rule 3).

### `holiday_templates` table

```sql
CREATE TABLE holiday_templates (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    year        INTEGER     NOT NULL,
    name        TEXT        NOT NULL,
    from_date   DATE        NOT NULL,
    to_date     DATE        NOT NULL,
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_holiday_templates_year ON holiday_templates (year);

CREATE TRIGGER trg_holiday_templates_set_updated_at
    BEFORE UPDATE ON holiday_templates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

Seeded in the migration with Vietnamese public holiday presets for supported years (2025, 2026, 2027 minimum). Read-only at runtime — users never mutate this table.

---

## Files

### New files

| File | Responsibility |
|---|---|
| `migrations/000023_holidays.up.sql` | Create `holidays` + `holiday_templates` tables; seed templates |
| `migrations/000023_holidays.down.sql` | Drop both tables |
| `internal/models/holiday.go` | `Holiday` + `HolidayTemplate` GORM models |
| `internal/dto/holiday.go` | `HolidayCreate`, `HolidayUpdate`, `HolidayListQuery`, `HolidayRead`, `HolidayTemplateRead`, `HolidayImportRequest`, `HolidayImportResult` DTOs |
| `internal/repositories/holiday_repo.go` | `HolidayRepository` interface + GORM impl |
| `internal/repositories/holiday_template_repo.go` | `HolidayTemplateRepository` interface + GORM impl |
| `internal/services/holiday_service.go` | `HolidayService` — CRUD + import + recalculation trigger |
| `internal/services/holiday_service_test.go` | Integration tests (TDD) |
| `pkg/utils/leave_days.go` | `CalcLeaveDays()` shared helper |
| `internal/handlers/holiday_handler.go` | 7 Gin handlers |

### Modified files

| File | Change |
|---|---|
| `internal/repositories/leave_repo.go` | Add `FindApprovedOverlapping(ctx, from, to)` + `BulkUpdateTotalDays(ctx, updates)` |
| `internal/services/leave_service.go` | Inject `HolidayRepository`; modify `Create` + `Update` to call `CalcLeaveDays` with holidays |
| `internal/services/leave_service_test.go` | Add tests for holiday-aware total_days at creation |
| `internal/services/testhelper_test.go` | Add `holidays,` + `holiday_templates,` to TRUNCATE list |
| `internal/permissions/registry.go` | Add `PermOrgHolidaysView` + `PermOrgHolidaysManage`; add to `AllPermissions()` + `PermissionGroups` |
| `internal/services/seed_service.go` | Seed `holidays_view` to all 5 roles; `holidays_manage` to Admin + HR Manager |
| `cmd/server/main.go` | Wire `HolidayRepository`, `HolidayTemplateRepository`, `HolidayService`, `HolidayHandler`; register routes |

---

## Shared Helper — `CalcLeaveDays`

```go
// pkg/utils/leave_days.go

// DateRange is a closed [From, To] interval in UTC date space.
type DateRange struct {
    From time.Time
    To   time.Time
}

// CalcLeaveDays returns the number of leave days consumed by a request,
// excluding any calendar days that fall within a company holiday range.
//
// Full-day leave: each overlapping holiday calendar day subtracts 1.0.
// Half-day leave: each overlapping holiday day subtracts 0.5 (AC-12).
func CalcLeaveDays(from, to time.Time, period models.LeavePeriod, holidays []DateRange) float64
```

**Algorithm:**
1. Build a set of all calendar dates in the leave range `[from, to]`.
2. For each date in the set, check if it falls inside any holiday range.
3. `holidayDays` = count of leave dates that are holidays.
4. `calendarDays` = `to.Sub(from).Hours()/24 + 1`.
5. If half-day period: base = 0.5; each holiday day subtracts 0.5. Result = `max(0, 0.5 - holidayDays*0.5)`.
6. If full-day period: result = `max(0, calendarDays - float64(holidayDays))`.

---

## Repository Layer

### `HolidayRepository` (new)

```go
type HolidayRepository interface {
    List(ctx context.Context, q HolidayListQuery) ([]Holiday, int64, error)
    Get(ctx, id uuid.UUID) (*Holiday, error)
    Create(ctx, h *Holiday) error
    Update(ctx, h *Holiday) error
    Delete(ctx, id uuid.UUID) error
    FindInRange(ctx context.Context, from, to time.Time) ([]Holiday, error) // for CalcLeaveDays at leave creation
    YearsWithHolidays(ctx context.Context) ([]int, error)                   // for year dropdown (AC-01)
    ExistsByNameAndYear(ctx context.Context, name string, year int, excludeID *uuid.UUID) (bool, error) // duplicate check
}
```

`List` filters by `year` (required), optional `search` (case-insensitive name match), ordered by `from_date ASC`.

`FindInRange` returns all non-deleted holidays whose date range overlaps `[from, to]` — used both at leave creation time and during recalculation.

### `HolidayTemplateRepository` (new)

```go
type HolidayTemplateRepository interface {
    ListByYear(ctx context.Context, year int) ([]HolidayTemplate, error)
}
```

### New methods on `LeaveRequestRepository` (existing)

```go
// FindApprovedOverlapping returns all Approved leave requests whose [from_date, to_date]
// overlaps [from, to]. Used by HolidayService after holiday mutations.
FindApprovedOverlapping(ctx context.Context, from, to time.Time) ([]LeaveRequest, error)

// BulkUpdateTotalDays updates total_days for the given request IDs.
// Returns the count of rows updated.
BulkUpdateTotalDays(ctx context.Context, updates []TotalDaysUpdate) (int, error)

// TotalDaysUpdate pairs a request ID with its recomputed total_days.
type TotalDaysUpdate struct {
    ID        uuid.UUID
    TotalDays float64
}
```

---

## Service Layer

### `HolidayService` (new)

```go
type HolidayService struct {
    repo         HolidayRepository
    templateRepo HolidayTemplateRepository
    leaveRepo    LeaveRequestRepository  // for recalculation
}
```

**Methods:**

`Create(ctx, req HolidayCreate) (*HolidayRead, int, *AppError)`
- Validate: name not blank, `to_date >= from_date`, name unique in year.
- Insert holiday.
- Recalculate affected Approved leaves overlapping `[from_date, to_date]` → returns affected count.

`Update(ctx, id uuid.UUID, req HolidayUpdate) (*HolidayRead, int, *AppError)`
- Fetch existing (404 if not found/deleted).
- Validate updated fields.
- Save.
- Recalculate leaves overlapping **union of old and new date ranges**.
- Returns affected count.

`Delete(ctx, id uuid.UUID) (int, *AppError)`
- Soft-delete.
- Recalculate affected Approved leaves overlapping the deleted holiday's range (adds days back).
- Returns affected count (used in delete toast — AC-07).

`List(ctx, q HolidayListQuery) (PaginatedData[HolidayRead], *AppError)`
- Paginated, year-scoped, optional name search.
- Ordered by `from_date ASC`.
- `total_days` computed in service: `int(to.Sub(from).Hours()/24) + 1`.

`GetYears(ctx) ([]int, *AppError)`
- Returns sorted list of years that have at least one holiday, always including the current year.

`ListTemplates(ctx, year int) ([]HolidayTemplateRead, *AppError)`
- Returns preset rows for the year; empty slice if none (no error — FE shows "no template" state).

`Import(ctx, req HolidayImportRequest) (*HolidayImportResult, *AppError)`
- For each selected template ID: check if name already exists in target year → skip if so.
- Bulk-insert non-skipped holidays.
- Recalculate affected Approved leaves for all inserted holiday ranges.
- Returns `{imported: N, skipped: M}`.

**Private helper — `recalculateAffectedLeaves`:**
```go
func (s *HolidayService) recalculateAffectedLeaves(ctx context.Context, from, to time.Time) (int, error)
```
1. `leaveRepo.FindApprovedOverlapping(ctx, from, to)` — get affected leaves.
2. `holidayRepo.FindInRange(ctx, leave.FromDate, leave.ToDate)` — per leave, fetch its current holidays.
3. `CalcLeaveDays(...)` — recompute each.
4. `leaveRepo.BulkUpdateTotalDays(ctx, updates)` — batch update.
5. Return affected count.

### Modified `LeaveService`

Add `holidayRepo HolidayRepository` field (injected).

Modify `Create` and `Update`:
```go
holidays, _ := s.holidayRepo.FindInRange(ctx, req.FromDate, req.ToDate)
ranges := make([]utils.DateRange, len(holidays))
for i, h := range holidays { ranges[i] = utils.DateRange{From: h.FromDate, To: h.ToDate} }
totalDays := utils.CalcLeaveDays(req.FromDate, req.ToDate, req.LeavePeriod, ranges)
```

---

## DTOs

```go
// HolidayCreate — POST /holidays body
type HolidayCreate struct {
    Year     int       `json:"year"      validate:"required,min=2000,max=2100"`
    Name     string    `json:"name"      validate:"required,max=100"`
    FromDate time.Time `json:"from_date" validate:"required"`
    ToDate   time.Time `json:"to_date"   validate:"required"`
}

// HolidayUpdate — PATCH /holidays/:id body (all optional)
type HolidayUpdate struct {
    Name     *string    `json:"name"`
    FromDate *time.Time `json:"from_date"`
    ToDate   *time.Time `json:"to_date"`
}

// HolidayListQuery — GET /holidays query params
type HolidayListQuery struct {
    Year     int    `form:"year"`
    Search   string `form:"search"`
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
}

// HolidayRead — response shape
type HolidayRead struct {
    ID        uuid.UUID `json:"id"`
    Year      int       `json:"year"`
    Name      string    `json:"name"`
    FromDate  time.Time `json:"from_date"`
    ToDate    time.Time `json:"to_date"`
    TotalDays int       `json:"total_days"` // computed: (to - from) + 1
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// HolidayTemplateRead — GET /holidays/templates response item
type HolidayTemplateRead struct {
    ID        uuid.UUID `json:"id"`
    Year      int       `json:"year"`
    Name      string    `json:"name"`
    FromDate  time.Time `json:"from_date"`
    ToDate    time.Time `json:"to_date"`
    TotalDays int       `json:"total_days"`
}

// HolidayImportRequest — POST /holidays/import body
type HolidayImportRequest struct {
    Year        int         `json:"year"         validate:"required"`
    TemplateIDs []uuid.UUID `json:"template_ids" validate:"required,min=1"`
}

// HolidayImportResult — POST /holidays/import response
type HolidayImportResult struct {
    Imported int `json:"imported"`
    Skipped  int `json:"skipped"`
}
```

---

## Permissions

```go
PermOrgHolidaysView   Permission = "organization:holidays_view"
PermOrgHolidaysManage Permission = "organization:holidays_manage"
```

New `PermissionGroup` in `PermissionGroups`:
```go
{
    Resource: "organization_holidays", Label: "Holidays",
    Permissions: []PermissionItem{
        {PermOrgHolidaysView,   "View Holidays",   "Browse the company holiday calendar"},
        {PermOrgHolidaysManage, "Manage Holidays", "Create, edit, delete, and import company holidays"},
    },
},
```

**Seed defaults:**

| Role | holidays_view | holidays_manage |
|---|---|---|
| Super Admin | ✅ (wildcard `*`) | ✅ |
| Admin | ✅ | ✅ |
| HR Manager | ✅ | ✅ |
| Manager | ✅ | ❌ |
| Employee | ✅ | ❌ |

---

## Routes

```
GET    /api/v1/holidays              → RequirePerms(holidays_view)
POST   /api/v1/holidays              → RequirePerms(holidays_manage)
GET    /api/v1/holidays/years        → RequirePerms(holidays_view)
GET    /api/v1/holidays/templates    → RequirePerms(holidays_manage)
POST   /api/v1/holidays/import       → RequirePerms(holidays_manage)
PATCH  /api/v1/holidays/:id          → RequirePerms(holidays_manage)
DELETE /api/v1/holidays/:id          → RequirePerms(holidays_manage)
```

Static routes (`/years`, `/templates`, `/import`) registered before `/:id` to avoid Gin wildcard conflicts.

---

## Response Envelopes

**Create (201):**
```json
{
  "success": true,
  "message": "Holiday has been created",
  "data": { /* HolidayRead */ }
}
```

**Update (200):**
```json
{
  "success": true,
  "message": "Holiday has been updated",
  "data": { /* HolidayRead */ }
}
```

**Delete (200):**
```json
{
  "success": true,
  "message": "Holiday has been deleted."            // no recalc
  // OR
  "message": "Holiday deleted. 3 leave request(s) recalculated."  // with recalc
}
```

**Import (200):**
```json
{
  "success": true,
  "message": "12 holidays imported for 2026"
  // OR "10 holidays imported, 2 skipped (already exist)"
  "data": { "imported": 10, "skipped": 2 }
}
```

---

## Error Cases

| Code | When |
|---|---|
| 400 | `to_date` before `from_date`; blank name; `from_date` missing |
| 409 | Duplicate name in same year (AC-03) |
| 404 | Holiday not found or soft-deleted |
| 403 | Missing permission |

---

## Testing Strategy

Integration tests in `internal/services/holiday_service_test.go` (package `services_test`, same pattern as `leave_service_test.go`):

- `Create_HappyPath` — creates, verifies DB row, `total_days` computed correctly
- `Create_DuplicateName_SameYear` → 409
- `Create_ToDateBeforeFromDate` → 400
- `Create_TriggersLeaveRecalculation` — seed Approved leave overlapping holiday; verify `total_days` decreases
- `Update_ChangeDates_Recalculates` — seed Approved leave; update holiday dates; verify recalc
- `Delete_ReturnsAffectedCount` — seed 2 Approved leaves; delete holiday; verify `total_days` restored + count = 2
- `Delete_NoOverlap_ReturnsZero`
- `List_YearScoped_SortedByFromDate`
- `List_SearchFilter`
- `Import_SkipsDuplicates` — seed 2 existing; import 12 templates; verify imported=10, skipped=2
- `GetYears_AlwaysIncludesCurrentYear`
- `CalcLeaveDays_FullDay_ExcludesHoliday` (unit test in `pkg/utils`)
- `CalcLeaveDays_HalfDay_HalfExcluded` (unit test — AC-12)
- `LeaveCreate_SubtractsHolidays` — modify leave creation test to verify holiday subtraction

---

## Out of Scope

Per DR §8:
- Attendance "H" cell display (separate attendance change, unblocked by this module)
- Recurring/auto-repeating holidays across years
- Overlap detection between two holidays
- Bulk delete
- Per-department holiday exceptions
- Leave notification on recalculation
