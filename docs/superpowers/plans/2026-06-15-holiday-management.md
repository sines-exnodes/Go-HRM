# Holiday Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add company holiday calendar — CRUD, Vietnamese preset import, and leave-day recalculation when holidays change (migration 000023 through handler registration).

**Architecture:** New `holidays` + `holiday_templates` tables. Shared `CalcLeaveDays` utility in `pkg/utils/`. Cross-repo injection: `HolidayService` injects `LeaveRequestRepository` for recalculation; `LeaveService` injects `HolidayRepository` for holiday-aware `total_days` at leave creation/update. No circular dependency.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL 14+, golang-migrate versioned SQL, testify integration tests.

---

## File map

| Action | File |
|---|---|
| Create | `migrations/000023_holidays.up.sql` |
| Create | `migrations/000023_holidays.down.sql` |
| Create | `internal/models/holiday.go` |
| Create | `internal/dto/holiday.go` |
| Create | `pkg/utils/leave_days.go` |
| Create | `pkg/utils/leave_days_test.go` |
| Modify | `internal/repositories/leave_request_repo.go` |
| Create | `internal/repositories/holiday_repo.go` |
| Create | `internal/repositories/holiday_template_repo.go` |
| Modify | `internal/services/testhelper_test.go` |
| Create | `internal/services/holiday_service.go` |
| Create | `internal/services/holiday_service_test.go` |
| Modify | `internal/services/leave_service.go` |
| Modify | `internal/permissions/registry.go` |
| Modify | `internal/services/seed_service.go` |
| Create | `internal/handlers/holiday_handler.go` |
| Modify | `cmd/server/main.go` |

---

### Task 1: Migration 000023

**Files:**
- Create: `migrations/000023_holidays.up.sql`
- Create: `migrations/000023_holidays.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- migrations/000023_holidays.up.sql

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

-- Unique name per year among non-deleted rows (AC-03)
CREATE UNIQUE INDEX uq_holidays_name_year
    ON holidays (year, LOWER(name))
    WHERE is_deleted = FALSE;

CREATE INDEX idx_holidays_year
    ON holidays (year, from_date)
    WHERE is_deleted = FALSE;

CREATE TRIGGER trg_holidays_set_updated_at
    BEFORE UPDATE ON holidays
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

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

-- Vietnamese public holiday presets (approximate statutory dates; users adjust via the UI)
INSERT INTO holiday_templates (year, name, from_date, to_date) VALUES
    -- 2025 (Ất Tỵ — Snake Year, Lunar New Year Jan 29)
    (2025, 'Tết Dương Lịch',               '2025-01-01', '2025-01-01'),
    (2025, 'Tết Nguyên Đán',               '2025-01-27', '2025-02-02'),
    (2025, 'Ngày Giỗ Tổ Hùng Vương',       '2025-04-07', '2025-04-07'),
    (2025, 'Ngày Giải Phóng & Lao Động',   '2025-04-30', '2025-05-01'),
    (2025, 'Ngày Quốc Khánh',              '2025-09-01', '2025-09-02'),
    -- 2026 (Bính Ngọ — Horse Year, Lunar New Year Feb 17)
    (2026, 'Tết Dương Lịch',               '2026-01-01', '2026-01-01'),
    (2026, 'Tết Nguyên Đán',               '2026-02-15', '2026-02-21'),
    (2026, 'Ngày Giỗ Tổ Hùng Vương',       '2026-04-26', '2026-04-26'),
    (2026, 'Ngày Giải Phóng & Lao Động',   '2026-04-30', '2026-05-01'),
    (2026, 'Ngày Quốc Khánh',              '2026-09-01', '2026-09-02'),
    -- 2027 (Đinh Mùi — Goat Year, Lunar New Year Feb 6)
    (2027, 'Tết Dương Lịch',               '2027-01-01', '2027-01-01'),
    (2027, 'Tết Nguyên Đán',               '2027-02-04', '2027-02-10'),
    (2027, 'Ngày Giỗ Tổ Hùng Vương',       '2027-04-15', '2027-04-15'),
    (2027, 'Ngày Giải Phóng & Lao Động',   '2027-04-30', '2027-05-01'),
    (2027, 'Ngày Quốc Khánh',              '2027-09-01', '2027-09-02');
```

- [ ] **Step 2: Write down migration**

```sql
-- migrations/000023_holidays.down.sql
DROP TABLE IF EXISTS holidays;
DROP TABLE IF EXISTS holiday_templates;
```

- [ ] **Step 3: Apply migration**

```bash
make migrate-up
```

Expected: `migrate: no error` (or shows version 23 applied).

- [ ] **Step 4: Verify**

```bash
make migrate-version
```

Expected output contains `23`.

- [ ] **Step 5: Commit**

```bash
git add migrations/000023_holidays.up.sql migrations/000023_holidays.down.sql
git commit -m "feat(holidays): migration 000023 — holidays + holiday_templates tables + Vietnamese preset seed"
```

---

### Task 2: Models + DTOs

**Files:**
- Create: `internal/models/holiday.go`
- Create: `internal/dto/holiday.go`

- [ ] **Step 1: Write models**

```go
// internal/models/holiday.go
package models

import "time"

// Holiday is a company-specific public holiday record.
type Holiday struct {
	BaseModel
	Year     int       `gorm:"not null;index"      json:"year"`
	Name     string    `gorm:"not null"            json:"name"`
	FromDate time.Time `gorm:"type:date;not null"  json:"from_date"`
	ToDate   time.Time `gorm:"type:date;not null"  json:"to_date"`
}

// HolidayTemplate is a read-only preset used for importing holidays.
// Seeded in migration 000023; never mutated at runtime.
type HolidayTemplate struct {
	BaseModel
	Year     int       `gorm:"not null;index"      json:"year"`
	Name     string    `gorm:"not null"            json:"name"`
	FromDate time.Time `gorm:"type:date;not null"  json:"from_date"`
	ToDate   time.Time `gorm:"type:date;not null"  json:"to_date"`
}
```

- [ ] **Step 2: Write DTOs**

```go
// internal/dto/holiday.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// HolidayCreate is the POST /holidays request body.
type HolidayCreate struct {
	Year     int       `json:"year"      binding:"required,min=2000,max=2100"`
	Name     string    `json:"name"      binding:"required,max=100"`
	FromDate time.Time `json:"from_date" binding:"required"`
	ToDate   time.Time `json:"to_date"   binding:"required"`
}

// HolidayUpdate is the PATCH /holidays/:id request body (all optional).
type HolidayUpdate struct {
	Name     *string    `json:"name"`
	FromDate *time.Time `json:"from_date"`
	ToDate   *time.Time `json:"to_date"`
}

// HolidayListQuery is the GET /holidays query string.
type HolidayListQuery struct {
	Year     int    `form:"year"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// HolidayRead is the canonical response shape for a single holiday.
type HolidayRead struct {
	ID        uuid.UUID `json:"id"`
	Year      int       `json:"year"`
	Name      string    `json:"name"`
	FromDate  time.Time `json:"from_date"`
	ToDate    time.Time `json:"to_date"`
	TotalDays int       `json:"total_days"` // computed: (to_date - from_date).Days + 1
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HolidayTemplateRead is the response shape for a preset template entry.
type HolidayTemplateRead struct {
	ID        uuid.UUID `json:"id"`
	Year      int       `json:"year"`
	Name      string    `json:"name"`
	FromDate  time.Time `json:"from_date"`
	ToDate    time.Time `json:"to_date"`
	TotalDays int       `json:"total_days"`
}

// HolidayImportRequest is the POST /holidays/import request body.
type HolidayImportRequest struct {
	Year        int         `json:"year"         binding:"required,min=2000,max=2100"`
	TemplateIDs []uuid.UUID `json:"template_ids" binding:"required,min=1"`
}

// HolidayImportResult is the POST /holidays/import response data.
type HolidayImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}
```

- [ ] **Step 3: Build check**

```bash
make build
```

Expected: compiles cleanly.

- [ ] **Step 4: Commit**

```bash
git add internal/models/holiday.go internal/dto/holiday.go
git commit -m "feat(holidays): Holiday + HolidayTemplate models and DTOs"
```

---

### Task 3: CalcLeaveDays utility + unit tests

**Files:**
- Create: `pkg/utils/leave_days.go`
- Create: `pkg/utils/leave_days_test.go`

- [ ] **Step 1: Write failing tests first**

```go
// pkg/utils/leave_days_test.go
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
```

- [ ] **Step 2: Run tests — expect compile failure**

```bash
go test ./pkg/utils/... -run TestCalcLeaveDays -v
```

Expected: compile error — `utils.CalcLeaveDays` and `utils.DateRange` undefined.

- [ ] **Step 3: Write implementation**

```go
// pkg/utils/leave_days.go
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
// Half-day leave: each overlapping holiday day subtracts 0.5 (AC-12).
// Result is clamped to 0 — it never returns negative.
func CalcLeaveDays(from, to time.Time, period models.LeavePeriod, holidays []DateRange) float64 {
	from = truncDay(from)
	to = truncDay(to)

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
```

- [ ] **Step 4: Run tests — expect all pass**

```bash
go test ./pkg/utils/... -run TestCalcLeaveDays -v
```

Expected: 6 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/utils/leave_days.go pkg/utils/leave_days_test.go
git commit -m "feat(holidays): CalcLeaveDays shared utility with unit tests"
```

---

### Task 4: Leave repo extensions

**Files:**
- Modify: `internal/repositories/leave_request_repo.go`

Add `TotalDaysUpdate` type + 2 interface methods + their implementations.

- [ ] **Step 1: Add type + interface methods**

In `internal/repositories/leave_request_repo.go`, after the `LeaveDaysCount` struct (around line 72), add:

```go
// TotalDaysUpdate pairs a leave request ID with its recomputed total_days.
// Used by BulkUpdateTotalDays.
type TotalDaysUpdate struct {
	ID        uuid.UUID
	TotalDays float64
}
```

In the `LeaveRequestRepository` interface (after `ApprovedForEmployeesInRange` at line 66), add:

```go
	// FindApprovedOverlapping returns all Approved (non-deleted) leave requests
	// whose [from_date, to_date] overlaps [from, to]. Used by HolidayService
	// when recalculating leaves after a holiday mutation.
	FindApprovedOverlapping(ctx context.Context, from, to time.Time) ([]models.LeaveRequest, error)

	// BulkUpdateTotalDays updates total_days for the given request IDs.
	// Returns the count of rows actually updated (skips already-deleted rows).
	BulkUpdateTotalDays(ctx context.Context, updates []TotalDaysUpdate) (int, error)
```

- [ ] **Step 2: Add implementations**

At the bottom of `internal/repositories/leave_request_repo.go`, add:

```go
func (r *leaveRequestRepo) FindApprovedOverlapping(ctx context.Context, from, to time.Time) ([]models.LeaveRequest, error) {
	var rows []models.LeaveRequest
	err := r.base(ctx).
		Where("status = ? AND from_date <= ? AND to_date >= ?", models.LeaveStatusApproved, to, from).
		Find(&rows).Error
	return rows, err
}

func (r *leaveRequestRepo) BulkUpdateTotalDays(ctx context.Context, updates []TotalDaysUpdate) (int, error) {
	if len(updates) == 0 {
		return 0, nil
	}
	total := 0
	for _, u := range updates {
		res := r.db.WithContext(ctx).
			Model(&models.LeaveRequest{}).
			Where("id = ? AND is_deleted = false", u.ID).
			Update("total_days", u.TotalDays)
		if res.Error != nil {
			return total, res.Error
		}
		total += int(res.RowsAffected)
	}
	return total, nil
}
```

- [ ] **Step 3: Build check**

```bash
make build
```

Expected: compiles cleanly.

- [ ] **Step 4: Commit**

```bash
git add internal/repositories/leave_request_repo.go
git commit -m "feat(holidays): add FindApprovedOverlapping + BulkUpdateTotalDays to leave repo"
```

---

### Task 5: Holiday repositories

**Files:**
- Create: `internal/repositories/holiday_repo.go`
- Create: `internal/repositories/holiday_template_repo.go`

- [ ] **Step 1: Write HolidayRepository**

```go
// internal/repositories/holiday_repo.go
package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// HolidayListQuery is the filter/pagination spec for List.
type HolidayListQuery struct {
	Year     int
	Search   string
	Page     int
	PageSize int
}

// HolidayRepository defines data access for the holidays table.
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
}

type holidayRepo struct{ db *gorm.DB }

// NewHolidayRepository constructs a Postgres-backed HolidayRepository.
func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepo{db: db}
}

func (r *holidayRepo) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Scopes(models.NotDeleted)
}

func (r *holidayRepo) List(ctx context.Context, q HolidayListQuery) ([]models.Holiday, int64, error) {
	qb := r.base(ctx).Model(&models.Holiday{}).Where("year = ?", q.Year)
	if q.Search != "" {
		qb = qb.Where("name ILIKE ?", "%"+q.Search+"%")
	}
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	var rows []models.Holiday
	err := qb.Order("from_date ASC").Limit(q.PageSize).Offset((q.Page - 1) * q.PageSize).Find(&rows).Error
	return rows, total, err
}

func (r *holidayRepo) Get(ctx context.Context, id uuid.UUID) (*models.Holiday, error) {
	var h models.Holiday
	err := r.base(ctx).First(&h, "id = ?", id).Error
	return &h, err
}

func (r *holidayRepo) Create(ctx context.Context, h *models.Holiday) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *holidayRepo) Update(ctx context.Context, h *models.Holiday) error {
	return r.db.WithContext(ctx).
		Model(h).
		Where("id = ? AND is_deleted = false", h.ID).
		Updates(map[string]any{
			"name":      h.Name,
			"from_date": h.FromDate,
			"to_date":   h.ToDate,
		}).Error
}

func (r *holidayRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&models.Holiday{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]any{"is_deleted": true, "deleted_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *holidayRepo) FindInRange(ctx context.Context, from, to time.Time) ([]models.Holiday, error) {
	var rows []models.Holiday
	err := r.base(ctx).
		Where("from_date <= ? AND to_date >= ?", to, from).
		Find(&rows).Error
	return rows, err
}

func (r *holidayRepo) YearsWithHolidays(ctx context.Context) ([]int, error) {
	var years []int
	err := r.base(ctx).
		Model(&models.Holiday{}).
		Distinct("year").
		Order("year ASC").
		Pluck("year", &years).Error
	return years, err
}

func (r *holidayRepo) ExistsByNameAndYear(ctx context.Context, name string, year int, excludeID *uuid.UUID) (bool, error) {
	qb := r.base(ctx).Model(&models.Holiday{}).
		Where("year = ? AND LOWER(name) = LOWER(?)", year, name)
	if excludeID != nil {
		qb = qb.Where("id != ?", *excludeID)
	}
	var count int64
	err := qb.Count(&count).Error
	return count > 0, err
}
```

- [ ] **Step 2: Write HolidayTemplateRepository**

```go
// internal/repositories/holiday_template_repo.go
package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

// HolidayTemplateRepository provides read-only access to the preset templates.
type HolidayTemplateRepository interface {
	ListByYear(ctx context.Context, year int) ([]models.HolidayTemplate, error)
}

type holidayTemplateRepo struct{ db *gorm.DB }

// NewHolidayTemplateRepository constructs a Postgres-backed HolidayTemplateRepository.
func NewHolidayTemplateRepository(db *gorm.DB) HolidayTemplateRepository {
	return &holidayTemplateRepo{db: db}
}

func (r *holidayTemplateRepo) ListByYear(ctx context.Context, year int) ([]models.HolidayTemplate, error) {
	var rows []models.HolidayTemplate
	err := r.db.WithContext(ctx).
		Scopes(models.NotDeleted).
		Where("year = ?", year).
		Order("from_date ASC").
		Find(&rows).Error
	return rows, err
}
```

- [ ] **Step 3: Build check**

```bash
make build
```

Expected: compiles cleanly.

- [ ] **Step 4: Commit**

```bash
git add internal/repositories/holiday_repo.go internal/repositories/holiday_template_repo.go
git commit -m "feat(holidays): HolidayRepository + HolidayTemplateRepository"
```

---

### Task 6: Test helper update

**Files:**
- Modify: `internal/services/testhelper_test.go`

- [ ] **Step 1: Prepend holidays tables to truncateAll**

In `internal/services/testhelper_test.go` line 134, change the TRUNCATE statement from:

```go
	if err := testDB.Exec(`TRUNCATE TABLE user_contracts, invites, system_config, ...`).Error; err != nil {
```

To (prepend `holidays, holiday_templates,`):

```go
	if err := testDB.Exec(`TRUNCATE TABLE holidays, holiday_templates, user_contracts, invites, system_config, announcement_views, announcement_attachments, announcement_target_users, announcement_target_departments, announcement_labels, announcements, labels, employee_skills, skills, device_tokens, user_notification_settings, attendance_sessions, attendance, leave_requests, employee_leave_quotas, employee_emergency_contacts, dependents, employees, positions, departments, user_roles, users, roles RESTART IDENTITY CASCADE`).Error; err != nil {
```

- [ ] **Step 2: Run existing tests — expect all still pass**

```bash
make test
```

Expected: all tests pass (or skip if no DB).

- [ ] **Step 3: Commit**

```bash
git add internal/services/testhelper_test.go
git commit -m "test(holidays): add holidays + holiday_templates to truncateAll"
```

---

### Task 7: HolidayService + integration tests (TDD)

**Files:**
- Create: `internal/services/holiday_service_test.go` (write first)
- Create: `internal/services/holiday_service.go`

- [ ] **Step 1: Write failing integration tests**

```go
// internal/services/holiday_service_test.go
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

	// Second create with same name + year → 409
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

	// Seed an approved leave that overlaps April 30.
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
		TotalDays:   4.0, // pre-holiday: 4 calendar days
		Reason:      "test",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, leaveRepo.Create(context.Background(), lr))

	// Create a 1-day holiday on Apr 30.
	svc := makeHolidaySvc(t)
	_, affected, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)
	assert.Equal(t, 1, affected)

	// Verify leave total_days was reduced from 4 to 3.
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
		TotalDays:   2.0, // Apr 30 was excluded as holiday
		Reason:      "test",
		Status:      models.LeaveStatusApproved,
		CreatedBy:   emp.ID,
	}
	require.NoError(t, leaveRepo.Create(context.Background(), lr))

	svc := makeHolidaySvc(t)
	// Create the holiday first.
	out, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2025,
		Name:     "Liberation Day",
		FromDate: time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)

	// Delete the holiday — leave should be recalculated (days restored).
	affected, aerr := svc.Delete(context.Background(), out.ID)
	require.Nil(t, aerr)
	assert.Equal(t, 1, affected)

	restored, err := leaveRepo.FindByID(context.Background(), lr.ID)
	require.NoError(t, err)
	assert.Equal(t, 3.0, restored.TotalDays) // Apr 29, 30, May 1 = 3 days
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
	assert.Equal(t, "Tết", page.Items[0].Name) // sorted by from_date ASC
}

func TestHoliday_Import_SkipsDuplicates(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)

	// Pre-create one holiday with same name as a template.
	_, _, aerr := svc.Create(context.Background(), dto.HolidayCreate{
		Year:     2026,
		Name:     "Tết Dương Lịch",
		FromDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	require.Nil(t, aerr)

	// Fetch 2026 templates to get their IDs.
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
	assert.Equal(t, len(templates)-1, result.Imported) // 1 skipped
	assert.Equal(t, 1, result.Skipped)
}

func TestHoliday_GetYears_AlwaysIncludesCurrentYear(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)

	svc := makeHolidaySvc(t)
	// No holidays in DB — should still return current year.
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
		TotalDays:   4.0, // Apr 30 is holiday: 5 - 1 = 4
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

	// Extend holiday to Apr 30 + May 1 → leave should lose another day.
	newTo := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	_, affected, aerr := svc.Update(context.Background(), h.ID, dto.HolidayUpdate{ToDate: &newTo})
	require.Nil(t, aerr)
	assert.Equal(t, 1, affected)

	updated, err := leaveRepo.FindByID(context.Background(), lr.ID)
	require.NoError(t, err)
	assert.Equal(t, 3.0, updated.TotalDays) // 5 days - 2 holiday days = 3
}
```

- [ ] **Step 2: Run tests — expect compile failure**

```bash
go test ./internal/services/... -run TestHoliday -v 2>&1 | head -20
```

Expected: compile error — `services.HolidayService` and `services.NewHolidayService` undefined.

- [ ] **Step 3: Write the service**

```go
// internal/services/holiday_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/pkg/utils"
)

// HolidayService manages company holiday records and triggers leave recalculation.
type HolidayService struct {
	repo         repositories.HolidayRepository
	templateRepo repositories.HolidayTemplateRepository
	leaveRepo    repositories.LeaveRequestRepository
}

// NewHolidayService constructs a HolidayService.
func NewHolidayService(
	repo repositories.HolidayRepository,
	templateRepo repositories.HolidayTemplateRepository,
	leaveRepo repositories.LeaveRequestRepository,
) *HolidayService {
	return &HolidayService{repo: repo, templateRepo: templateRepo, leaveRepo: leaveRepo}
}

// ---- Helpers ----

func toHolidayRead(h models.Holiday) dto.HolidayRead {
	return dto.HolidayRead{
		ID:        h.ID,
		Year:      h.Year,
		Name:      h.Name,
		FromDate:  h.FromDate,
		ToDate:    h.ToDate,
		TotalDays: int(h.ToDate.Sub(h.FromDate).Hours()/24) + 1,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

func toHolidayTemplateRead(t models.HolidayTemplate) dto.HolidayTemplateRead {
	return dto.HolidayTemplateRead{
		ID:        t.ID,
		Year:      t.Year,
		Name:      t.Name,
		FromDate:  t.FromDate,
		ToDate:    t.ToDate,
		TotalDays: int(t.ToDate.Sub(t.FromDate).Hours()/24) + 1,
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// recalculateAffectedLeaves recomputes total_days for all Approved leave requests
// that overlap [from, to]. Returns the count of rows updated.
func (s *HolidayService) recalculateAffectedLeaves(ctx context.Context, from, to time.Time) (int, error) {
	leaves, err := s.leaveRepo.FindApprovedOverlapping(ctx, from, to)
	if err != nil {
		return 0, err
	}
	if len(leaves) == 0 {
		return 0, nil
	}
	updates := make([]repositories.TotalDaysUpdate, 0, len(leaves))
	for _, lr := range leaves {
		holidays, err := s.repo.FindInRange(ctx, lr.FromDate, lr.ToDate)
		if err != nil {
			return 0, err
		}
		ranges := make([]utils.DateRange, len(holidays))
		for i, h := range holidays {
			ranges[i] = utils.DateRange{From: h.FromDate, To: h.ToDate}
		}
		td := utils.CalcLeaveDays(lr.FromDate, lr.ToDate, lr.LeavePeriod, ranges)
		updates = append(updates, repositories.TotalDaysUpdate{ID: lr.ID, TotalDays: td})
	}
	return s.leaveRepo.BulkUpdateTotalDays(ctx, updates)
}

// ---- Public methods ----

// Create inserts a new holiday and recalculates affected approved leaves.
// Returns the created record and the count of leave requests recalculated.
func (s *HolidayService) Create(ctx context.Context, req dto.HolidayCreate) (*dto.HolidayRead, int, *apperrors.AppError) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, 0, apperrors.ErrBadRequest("name is required")
	}
	from := truncateToDate(req.FromDate)
	to := truncateToDate(req.ToDate)
	if to.Before(from) {
		return nil, 0, apperrors.ErrBadRequest("to_date must be on or after from_date")
	}
	exists, err := s.repo.ExistsByNameAndYear(ctx, req.Name, req.Year, nil)
	if err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	if exists {
		return nil, 0, apperrors.ErrConflict(fmt.Sprintf("a holiday named %q already exists in %d", req.Name, req.Year))
	}
	h := &models.Holiday{
		Year:     req.Year,
		Name:     strings.TrimSpace(req.Name),
		FromDate: from,
		ToDate:   to,
	}
	if err := s.repo.Create(ctx, h); err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	affected, err := s.recalculateAffectedLeaves(ctx, from, to)
	if err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	out := toHolidayRead(*h)
	return &out, affected, nil
}

// Update patches an existing holiday and recalculates leaves for the union
// of old and new date ranges.
func (s *HolidayService) Update(ctx context.Context, id uuid.UUID, req dto.HolidayUpdate) (*dto.HolidayRead, int, *apperrors.AppError) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperrors.ErrNotFound("holiday")
		}
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	oldFrom := h.FromDate
	oldTo := h.ToDate

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, 0, apperrors.ErrBadRequest("name cannot be blank")
		}
		exists, err := s.repo.ExistsByNameAndYear(ctx, name, h.Year, &h.ID)
		if err != nil {
			return nil, 0, apperrors.ErrInternal(err.Error())
		}
		if exists {
			return nil, 0, apperrors.ErrConflict(fmt.Sprintf("a holiday named %q already exists in %d", name, h.Year))
		}
		h.Name = name
	}
	if req.FromDate != nil {
		h.FromDate = truncateToDate(*req.FromDate)
	}
	if req.ToDate != nil {
		h.ToDate = truncateToDate(*req.ToDate)
	}
	if h.ToDate.Before(h.FromDate) {
		return nil, 0, apperrors.ErrBadRequest("to_date must be on or after from_date")
	}
	if err := s.repo.Update(ctx, h); err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}

	// Recalculate leaves overlapping the union of old and new ranges.
	recalcFrom := minTime(oldFrom, h.FromDate)
	recalcTo := maxTime(oldTo, h.ToDate)
	affected, err := s.recalculateAffectedLeaves(ctx, recalcFrom, recalcTo)
	if err != nil {
		return nil, 0, apperrors.ErrInternal(err.Error())
	}
	out := toHolidayRead(*h)
	return &out, affected, nil
}

// Delete soft-deletes a holiday and recalculates affected approved leaves
// (restoring days that were previously subtracted).
// Returns the count of leave requests recalculated (used in the delete toast).
func (s *HolidayService) Delete(ctx context.Context, id uuid.UUID) (int, *apperrors.AppError) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.ErrNotFound("holiday")
		}
		return 0, apperrors.ErrInternal(err.Error())
	}
	from := h.FromDate
	to := h.ToDate

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.ErrNotFound("holiday")
		}
		return 0, apperrors.ErrInternal(err.Error())
	}
	// Recalculate AFTER soft-delete so FindInRange excludes the deleted holiday.
	affected, err := s.recalculateAffectedLeaves(ctx, from, to)
	if err != nil {
		return 0, apperrors.ErrInternal(err.Error())
	}
	return affected, nil
}

// List returns a paginated, year-scoped list of holidays ordered by from_date.
func (s *HolidayService) List(ctx context.Context, q dto.HolidayListQuery) (dto.PaginatedData[dto.HolidayRead], *apperrors.AppError) {
	if q.Year < 2000 || q.Year > 2100 {
		return dto.PaginatedData[dto.HolidayRead]{}, apperrors.ErrBadRequest("year must be between 2000 and 2100")
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	rows, total, err := s.repo.List(ctx, repositories.HolidayListQuery{
		Year:     q.Year,
		Search:   q.Search,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		return dto.PaginatedData[dto.HolidayRead]{}, apperrors.ErrInternal(err.Error())
	}
	items := make([]dto.HolidayRead, len(rows))
	for i, h := range rows {
		items[i] = toHolidayRead(h)
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(q.PageSize)))
	}
	return dto.PaginatedData[dto.HolidayRead]{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetYears returns distinct years that have at least one holiday, always
// including the current year in sorted order.
func (s *HolidayService) GetYears(ctx context.Context) ([]int, *apperrors.AppError) {
	years, err := s.repo.YearsWithHolidays(ctx)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	currentYear := time.Now().UTC().Year()
	for _, y := range years {
		if y == currentYear {
			return years, nil
		}
	}
	// Insert current year in sorted position.
	result := make([]int, 0, len(years)+1)
	inserted := false
	for _, y := range years {
		if !inserted && currentYear < y {
			result = append(result, currentYear)
			inserted = true
		}
		result = append(result, y)
	}
	if !inserted {
		result = append(result, currentYear)
	}
	return result, nil
}

// ListTemplates returns preset templates for the given year. Empty slice (no
// error) when no templates exist for that year.
func (s *HolidayService) ListTemplates(ctx context.Context, year int) ([]dto.HolidayTemplateRead, *apperrors.AppError) {
	rows, err := s.templateRepo.ListByYear(ctx, year)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := make([]dto.HolidayTemplateRead, len(rows))
	for i, t := range rows {
		out[i] = toHolidayTemplateRead(t)
	}
	return out, nil
}

// Import bulk-inserts selected templates into the target year's holiday list.
// Duplicates (same name already exists) are skipped.
// Returns imported + skipped counts.
func (s *HolidayService) Import(ctx context.Context, req dto.HolidayImportRequest) (*dto.HolidayImportResult, *apperrors.AppError) {
	if len(req.TemplateIDs) == 0 {
		return nil, apperrors.ErrBadRequest("template_ids must not be empty")
	}
	templates, err := s.templateRepo.ListByYear(ctx, req.Year)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	wantSet := make(map[uuid.UUID]bool, len(req.TemplateIDs))
	for _, id := range req.TemplateIDs {
		wantSet[id] = true
	}

	imported := 0
	skipped := 0
	var importedRanges []utils.DateRange

	for _, tmpl := range templates {
		if !wantSet[tmpl.ID] {
			continue
		}
		exists, err := s.repo.ExistsByNameAndYear(ctx, tmpl.Name, req.Year, nil)
		if err != nil {
			return nil, apperrors.ErrInternal(err.Error())
		}
		if exists {
			skipped++
			continue
		}
		h := &models.Holiday{
			Year:     req.Year,
			Name:     tmpl.Name,
			FromDate: tmpl.FromDate,
			ToDate:   tmpl.ToDate,
		}
		if err := s.repo.Create(ctx, h); err != nil {
			return nil, apperrors.ErrInternal(err.Error())
		}
		imported++
		importedRanges = append(importedRanges, utils.DateRange{From: h.FromDate, To: h.ToDate})
	}

	for _, r := range importedRanges {
		if _, err := s.recalculateAffectedLeaves(ctx, r.From, r.To); err != nil {
			return nil, apperrors.ErrInternal(err.Error())
		}
	}

	return &dto.HolidayImportResult{Imported: imported, Skipped: skipped}, nil
}
```

Note: `holiday_service.go` imports `"math"` for `math.Ceil` in `List`. Add it to the import block.

- [ ] **Step 4: Run tests — expect all pass**

```bash
go test ./internal/services/... -run TestHoliday -v
```

Expected: 8 tests PASS (or SKIP if no DB).

- [ ] **Step 5: Build check**

```bash
make build
```

Expected: compiles cleanly.

- [ ] **Step 6: Commit**

```bash
git add internal/services/holiday_service.go internal/services/holiday_service_test.go
git commit -m "feat(holidays): HolidayService CRUD + import + recalculation with integration tests"
```

---

### Task 8: Modify LeaveService — holiday-aware total_days

**Files:**
- Modify: `internal/services/leave_service.go`

- [ ] **Step 1: Add `holidays` field to LeaveService struct**

In `internal/services/leave_service.go`, change the struct (around line 71):

```go
// BEFORE
type LeaveService struct {
	leaves  repositories.LeaveRequestRepository
	emps    repositories.EmployeeRepository
	depts   repositories.DepartmentRepository
	pos     repositories.PositionRepository
	quota   repositories.LeaveQuotaRepository
	uploads Uploader
}

// AFTER
type LeaveService struct {
	leaves   repositories.LeaveRequestRepository
	emps     repositories.EmployeeRepository
	depts    repositories.DepartmentRepository
	pos      repositories.PositionRepository
	quota    repositories.LeaveQuotaRepository
	uploads  Uploader
	holidays repositories.HolidayRepository
}
```

- [ ] **Step 2: Update NewLeaveService constructor**

Change the `NewLeaveService` function (around line 83):

```go
// BEFORE
func NewLeaveService(
	leaves repositories.LeaveRequestRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	pos repositories.PositionRepository,
	quota repositories.LeaveQuotaRepository,
	uploads Uploader,
) *LeaveService {
	return &LeaveService{
		leaves:  leaves,
		emps:    emps,
		depts:   depts,
		pos:     pos,
		quota:   quota,
		uploads: uploads,
	}
}

// AFTER
func NewLeaveService(
	leaves repositories.LeaveRequestRepository,
	emps repositories.EmployeeRepository,
	depts repositories.DepartmentRepository,
	pos repositories.PositionRepository,
	quota repositories.LeaveQuotaRepository,
	uploads Uploader,
	holidays repositories.HolidayRepository,
) *LeaveService {
	return &LeaveService{
		leaves:   leaves,
		emps:     emps,
		depts:    depts,
		pos:      pos,
		quota:    quota,
		uploads:  uploads,
		holidays: holidays,
	}
}
```

- [ ] **Step 3: Update Create — replace calculateTotalDays with CalcLeaveDays**

In `LeaveService.Create` (around line 387), replace:

```go
// BEFORE (lines 387-390)
totalDays, err := validateDateInputs(in.FromDate, in.ToDate, in.LeavePeriod)
if err != nil {
    return nil, err
}
```

With:

```go
// AFTER
if _, err := validateDateInputs(in.FromDate, in.ToDate, in.LeavePeriod); err != nil {
    return nil, err
}
holidayRows, err := s.holidays.FindInRange(ctx, in.FromDate, in.ToDate)
if err != nil {
    return nil, apperrors.ErrInternal(err.Error())
}
hRanges := make([]utils.DateRange, len(holidayRows))
for i, h := range holidayRows {
    hRanges[i] = utils.DateRange{From: h.FromDate, To: h.ToDate}
}
totalDays := utils.CalcLeaveDays(in.FromDate, in.ToDate, in.LeavePeriod, hRanges)
```

- [ ] **Step 4: Update Update — replace calculateTotalDays with CalcLeaveDays**

In `LeaveService.Update` (around line 534), replace:

```go
// BEFORE (lines 534-540)
totalDays, err := validateDateInputs(row.FromDate, row.ToDate, row.LeavePeriod)
if err != nil {
    return nil, err
}
row.TotalDays = totalDays
row.FromDate = truncateToDate(row.FromDate)
row.ToDate = truncateToDate(row.ToDate)
```

With:

```go
// AFTER
if _, err := validateDateInputs(row.FromDate, row.ToDate, row.LeavePeriod); err != nil {
    return nil, err
}
row.FromDate = truncateToDate(row.FromDate)
row.ToDate = truncateToDate(row.ToDate)
holidayRows, err := s.holidays.FindInRange(ctx, row.FromDate, row.ToDate)
if err != nil {
    return nil, apperrors.ErrInternal(err.Error())
}
hRanges := make([]utils.DateRange, len(holidayRows))
for i, h := range holidayRows {
    hRanges[i] = utils.DateRange{From: h.FromDate, To: h.ToDate}
}
row.TotalDays = utils.CalcLeaveDays(row.FromDate, row.ToDate, row.LeavePeriod, hRanges)
```

- [ ] **Step 5: Add import for utils package**

Add `"github.com/exnodes/hrm-api/pkg/utils"` to `leave_service.go`'s import block if not already present.

- [ ] **Step 6: Build check**

```bash
make build
```

Expected: compile error about `NewLeaveService` in `main.go` — fix in next step.

- [ ] **Step 7: Fix main.go wiring**

In `cmd/server/main.go`, after the existing repo declarations (after `inviteRepo` line ~68), add:

```go
holidayRepo := repositories.NewHolidayRepository(db)
holidayTemplateRepo := repositories.NewHolidayTemplateRepository(db)
```

Then change the `leaveSvc` construction (around line 96):

```go
// BEFORE
leaveSvc := services.NewLeaveService(leaveRepo, employeeRepo, departmentRepo, positionRepo, quotaRepo, uploadSvc)

// AFTER
leaveSvc := services.NewLeaveService(leaveRepo, employeeRepo, departmentRepo, positionRepo, quotaRepo, uploadSvc, holidayRepo)
```

- [ ] **Step 8: Build check again**

```bash
make build
```

Expected: compiles cleanly.

- [ ] **Step 9: Run full test suite**

```bash
make test
```

Expected: all tests pass.

- [ ] **Step 10: Commit**

```bash
git add internal/services/leave_service.go cmd/server/main.go
git commit -m "feat(holidays): inject HolidayRepository into LeaveService — holiday-aware total_days at leave creation"
```

---

### Task 9: Permissions + Seed

**Files:**
- Modify: `internal/permissions/registry.go`
- Modify: `internal/services/seed_service.go`

- [ ] **Step 1: Add permission constants**

In `internal/permissions/registry.go`, in the `const` block after `PermOrgSettings`, add:

```go
// Holidays
PermOrgHolidaysView   Permission = "organization:holidays_view"
PermOrgHolidaysManage Permission = "organization:holidays_manage"
```

- [ ] **Step 2: Add to AllPermissions()**

In `AllPermissions()`, after `PermOrgSettings,`, add:

```go
PermOrgHolidaysView, PermOrgHolidaysManage,
```

- [ ] **Step 3: Add PermissionGroup**

In `PermissionGroups`, after the `organization_settings` group, add:

```go
{
    Resource: "organization_holidays", Label: "Holidays",
    Permissions: []PermissionItem{
        {PermOrgHolidaysView, "View Holidays", "Browse the company holiday calendar"},
        {PermOrgHolidaysManage, "Manage Holidays", "Create, edit, delete, and import company holidays"},
    },
},
```

- [ ] **Step 4: Add to seed — Admin role**

In `internal/services/seed_service.go`, in the **Admin** `Permissions` slice, after `permissions.PermInviteManage,`, add:

```go
permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage,
```

- [ ] **Step 5: Add to seed — HR Manager role**

In the **HR Manager** `Permissions` slice, after `permissions.PermInviteManage,`, add:

```go
permissions.PermOrgHolidaysView, permissions.PermOrgHolidaysManage,
```

- [ ] **Step 6: Add to seed — Manager role**

In the **Manager** `Permissions` slice (after `permissions.PermAttendanceManage,`), add:

```go
permissions.PermOrgHolidaysView,
```

- [ ] **Step 7: Add to seed — Employee role**

In the **Employee** `Permissions` slice (after `permissions.PermLeaveDelete,`), add:

```go
permissions.PermOrgHolidaysView,
```

- [ ] **Step 8: Build + test**

```bash
make build && make test
```

Expected: all pass.

- [ ] **Step 9: Commit**

```bash
git add internal/permissions/registry.go internal/services/seed_service.go
git commit -m "feat(holidays): PermOrgHolidaysView + PermOrgHolidaysManage — registry + seed all 5 roles"
```

---

### Task 10: Handler + Wire + Routes

**Files:**
- Create: `internal/handlers/holiday_handler.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Write the handler**

```go
// internal/handlers/holiday_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/services"
)

// HolidayHandler handles /api/v1/holidays endpoints.
type HolidayHandler struct {
	svc *services.HolidayService
}

// NewHolidayHandler constructs a HolidayHandler.
func NewHolidayHandler(svc *services.HolidayService) *HolidayHandler {
	return &HolidayHandler{svc: svc}
}

// List godoc
// @Summary      List holidays for a year
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Param        year       query  int     true  "calendar year (e.g. 2025)"
// @Param        search     query  string  false "name search"
// @Param        page       query  int     false "page number (default 1)"
// @Param        page_size  query  int     false "page size (default 20, max 100)"
// @Success      200  {object}  dto.Response[dto.PaginatedData[dto.HolidayRead]]
// @Failure      400  {object}  dto.Response[any]
// @Router       /holidays [get]
func (h *HolidayHandler) List(c *gin.Context) {
	var q dto.HolidayListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.List(c.Request.Context(), q)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.HolidayRead]]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create a holiday
// @Tags         holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.HolidayCreate  true  "create payload"
// @Success      201  {object}  dto.Response[dto.HolidayRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /holidays [post]
func (h *HolidayHandler) Create(c *gin.Context) {
	var req dto.HolidayCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, _, aerr := h.svc.Create(c.Request.Context(), req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.HolidayRead]{
		Success: true,
		Message: "Holiday has been created",
		Data:    out,
	})
}

// Update godoc
// @Summary      Partial-update a holiday
// @Tags         holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string              true  "holiday uuid"
// @Param        body  body  dto.HolidayUpdate   true  "patch payload"
// @Success      200  {object}  dto.Response[dto.HolidayRead]
// @Failure      400  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Failure      409  {object}  dto.Response[any]
// @Router       /holidays/{id} [patch]
func (h *HolidayHandler) Update(c *gin.Context) {
	id, aerr := parseIDParam(c, "id")
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	var req dto.HolidayUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, _, aerr := h.svc.Update(c.Request.Context(), id, req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.HolidayRead]{
		Success: true,
		Message: "Holiday has been updated",
		Data:    out,
	})
}

// Delete godoc
// @Summary      Soft-delete a holiday
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "holiday uuid"
// @Success      200  {object}  dto.Response[any]
// @Failure      404  {object}  dto.Response[any]
// @Router       /holidays/{id} [delete]
func (h *HolidayHandler) Delete(c *gin.Context) {
	id, aerr := parseIDParam(c, "id")
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	affected, aerr := h.svc.Delete(c.Request.Context(), id)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	msg := "Holiday has been deleted"
	if affected > 0 {
		msg = fmt.Sprintf("Holiday deleted. %d leave request(s) recalculated.", affected)
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: msg})
}

// GetYears godoc
// @Summary      List years that have holidays
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.Response[[]int]
// @Router       /holidays/years [get]
func (h *HolidayHandler) GetYears(c *gin.Context) {
	years, aerr := h.svc.GetYears(c.Request.Context())
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]int]{Success: true, Data: years})
}

// ListTemplates godoc
// @Summary      List Vietnamese holiday presets for a year
// @Tags         holidays
// @Security     BearerAuth
// @Produce      json
// @Param        year  query  int  true  "calendar year"
// @Success      200  {object}  dto.Response[[]dto.HolidayTemplateRead]
// @Router       /holidays/templates [get]
func (h *HolidayHandler) ListTemplates(c *gin.Context) {
	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		_ = c.Error(apperrors.ErrBadRequest("year query param must be a valid year (2000–2100)"))
		return
	}
	out, aerr := h.svc.ListTemplates(c.Request.Context(), year)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[[]dto.HolidayTemplateRead]{Success: true, Data: out})
}

// Import godoc
// @Summary      Import selected holiday presets into a year
// @Tags         holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.HolidayImportRequest  true  "import payload"
// @Success      200  {object}  dto.Response[dto.HolidayImportResult]
// @Failure      400  {object}  dto.Response[any]
// @Router       /holidays/import [post]
func (h *HolidayHandler) Import(c *gin.Context) {
	var req dto.HolidayImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.Import(c.Request.Context(), req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	msg := fmt.Sprintf("%d holiday(s) imported for %d", out.Imported, req.Year)
	if out.Skipped > 0 {
		msg = fmt.Sprintf("%d holiday(s) imported for %d, %d skipped (already exist)", out.Imported, req.Year, out.Skipped)
	}
	c.JSON(http.StatusOK, dto.Response[*dto.HolidayImportResult]{Success: true, Message: msg, Data: out})
}

// parseHolidayID is a local alias kept for clarity — delegates to the shared
// parseIDParam defined in employee_handler.go (same package).
// (No separate function needed; call parseIDParam(c, "id") directly.)

// toHolidayUUID is a helper used inline — unused; parseIDParam returns (uuid.UUID, *apperrors.AppError).
var _ = uuid.Nil // suppress unused import if uuid is only used in parseIDParam call
```

> **Note:** `parseIDParam` is defined in `internal/handlers/employee_handler.go` and is accessible from `holiday_handler.go` since both are in `package handlers`.

- [ ] **Step 2: Wire in main.go**

In `cmd/server/main.go`:

**After the `userContractSvc` line (~line 127), add:**

```go
holidaySvc := services.NewHolidayService(holidayRepo, holidayTemplateRepo, leaveRepo)
```

**After the `userContractH` handler line (~line 151), add:**

```go
holidayH := handlers.NewHolidayHandler(holidaySvc)
```

**In the routes section, after the `userContracts` group block (~line 234), add:**

```go
// ---- /holidays (Holiday Management) ----
// Static routes (years, templates, import) registered BEFORE /:id to avoid
// Gin wildcard conflicts.
holidays := authed.Group("/holidays")
holidays.GET("/years", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysView), holidayH.GetYears)
holidays.GET("/templates", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysManage), holidayH.ListTemplates)
holidays.POST("/import", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysManage), holidayH.Import)
holidays.GET("", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysView), holidayH.List)
holidays.POST("", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysManage), holidayH.Create)
holidays.PATCH(":id", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysManage), holidayH.Update)
holidays.DELETE(":id", middleware.RequirePerms(authSvc, permissions.PermOrgHolidaysManage), holidayH.Delete)
```

- [ ] **Step 3: Build check**

```bash
make build
```

Expected: compiles cleanly (fix any import issues — `uuid` may be unused if the blank var trick is not needed; remove the `var _ = uuid.Nil` line from handler if `uuid` is not directly referenced).

- [ ] **Step 4: Commit**

```bash
git add internal/handlers/holiday_handler.go cmd/server/main.go
git commit -m "feat(holidays): HolidayHandler (7 endpoints) + wire main.go + routes"
```

---

### Task 11: Verification

- [ ] **Step 1: Format + vet + test**

```bash
make fmt && make vet && make test
```

Expected: all pass, no vet warnings.

- [ ] **Step 2: Regenerate Swagger**

```bash
make swag
```

Expected: `docs/swagger/` regenerated cleanly.

- [ ] **Step 3: Start server**

```bash
make run
```

Expected: `exnodes-hrm-api listening on :8080` — no crash (migration version check passes).

- [ ] **Step 4: Login and get token**

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@local.dev","password":"admin123"}' \
  | jq -r '.data.access_token')
echo $TOKEN
```

Expected: non-empty JWT string.

- [ ] **Step 5: Create a holiday**

```bash
curl -s -X POST http://localhost:8080/api/v1/holidays \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"year":2026,"name":"Liberation Day","from_date":"2026-04-30T00:00:00Z","to_date":"2026-04-30T00:00:00Z"}' \
  | jq
```

Expected: `201` with `{"success":true,"message":"Holiday has been created","data":{"id":"...","year":2026,"name":"Liberation Day","total_days":1,...}}`.

- [ ] **Step 6: List holidays**

```bash
curl -s "http://localhost:8080/api/v1/holidays?year=2026" \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected: `{"success":true,"data":{"items":[...],"total":1,...}}`.

- [ ] **Step 7: Get years**

```bash
curl -s http://localhost:8080/api/v1/holidays/years \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected: array containing `2026` and the current year.

- [ ] **Step 8: List templates**

```bash
curl -s "http://localhost:8080/api/v1/holidays/templates?year=2026" \
  -H "Authorization: Bearer $TOKEN" | jq '.data | length'
```

Expected: `5` (five 2026 templates seeded by migration).

- [ ] **Step 9: Import templates**

```bash
# Get template IDs for 2026
TEMPLATE_IDS=$(curl -s "http://localhost:8080/api/v1/holidays/templates?year=2026" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '[.data[].id]')

curl -s -X POST http://localhost:8080/api/v1/holidays/import \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"year\":2026,\"template_ids\":$TEMPLATE_IDS}" | jq
```

Expected: `{"success":true,"message":"4 holiday(s) imported for 2026, 1 skipped (already exist)","data":{"imported":4,"skipped":1}}` (Liberation Day was created in step 5).

- [ ] **Step 10: Delete a holiday**

```bash
HOLIDAY_ID=$(curl -s "http://localhost:8080/api/v1/holidays?year=2026" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data.items[0].id')

curl -s -X DELETE "http://localhost:8080/api/v1/holidays/$HOLIDAY_ID" \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected: `{"success":true,"message":"Holiday has been deleted"}` (or with recalc count if overlapping approved leaves exist).

- [ ] **Step 11: Duplicate name → 409**

```bash
curl -s -X POST http://localhost:8080/api/v1/holidays \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"year":2026,"name":"Tết Nguyên Đán","from_date":"2026-02-15T00:00:00Z","to_date":"2026-02-21T00:00:00Z"}' \
  | jq
# Create again same name — expect 409
curl -s -X POST http://localhost:8080/api/v1/holidays \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"year":2026,"name":"Tết Nguyên Đán","from_date":"2026-02-15T00:00:00Z","to_date":"2026-02-21T00:00:00Z"}' \
  | jq '.success,.message'
```

Expected: first call 201; second call `false` + `"a holiday named \"Tết Nguyên Đán\" already exists in 2026"`.

- [ ] **Step 12: Update CHECKPOINT.md**

Replace `docs/superpowers/CHECKPOINT.md` with:

```markdown
# CHECKPOINT

## Current Phase
Phase (Holidays) — Implementation plan written, ready for execution.

## Last Verified
Phase 3 (Departments + Positions) + Phase User Contracts — verified end-to-end.

## What Was Done This Session
- Fixed double "not found" suffix in user_contract_service.go
- Diagnosed Docker postgres network isolation (postgres container had no networks)
- Diagnosed FE bug: FE was passing employee ID instead of user ID in contract URL
- Designed Holiday Management module (spec: docs/superpowers/specs/2026-06-15-holiday-management-design.md)
- Wrote implementation plan: docs/superpowers/plans/2026-06-15-holiday-management.md

## What Is Next
Execute the holiday management plan:
1. Task 1: Migration 000023
2. Tasks 2–3: Models, DTOs, CalcLeaveDays utility
3. Tasks 4–6: Repositories + test helper
4. Tasks 7–8: HolidayService (TDD) + LeaveService modification
5. Tasks 9–10: Permissions, seed, handler, wire
6. Task 11: Verification + committed log

## Known Gaps / Follow-Ups
- FE must use user_id (not employee id) when constructing /users/:id/contracts URL
- Migration 000022 (user_contracts) must be applied on dev DB: make migrate-up
- migration 000023 will be applied in Task 1 of holiday plan
```

- [ ] **Step 13: Commit verification artifacts**

```bash
make swag
git add docs/superpowers/CHECKPOINT.md docs/swagger/
git commit -m "docs(holidays): plan complete — CHECKPOINT + Swagger regenerated"
```

---

## Self-review checklist

- [x] **Spec coverage**: All 7 routes covered. CalcLeaveDays for both create + recalc. Import skip logic. GetYears always includes current year. Delete message varies by affected count.
- [x] **No placeholders**: All code blocks are complete. No "TBD" or "implement later".
- [x] **Type consistency**: `TotalDaysUpdate` defined in `leave_request_repo.go` and consumed in `holiday_service.go`. `DateRange` defined in `pkg/utils/leave_days.go` and consumed in `holiday_service.go` + `leave_service.go`. `HolidayListQuery` defined in `repositories/holiday_repo.go` (not dto) — service maps `dto.HolidayListQuery` to `repositories.HolidayListQuery`.
- [x] **Soft delete**: Uses `is_deleted=true/deleted_at=now()` explicit Updates map (not GORM's built-in DeletedAt).
- [x] **Route order**: Static routes (`/years`, `/templates`, `/import`) registered before `/:id`.
- [x] **Truncate order**: `holidays, holiday_templates` prepended before `user_contracts` (no FK deps from holidays to other tables).
- [x] **math import**: `holiday_service.go` needs `"math"` for `math.Ceil` — included in the import list described in the service code.
