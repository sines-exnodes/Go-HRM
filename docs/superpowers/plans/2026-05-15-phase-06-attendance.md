# Phase 6 — Attendance Implementation Plan

| | |
|---|---|
| Status | Ready to execute |
| Date | 2026-05-15 |
| Owner | danny.tranhoang@exnodes.vn |
| Spec | `docs/superpowers/specs/2026-05-15-go-migration-design.md` |
| Phase | 6 of 15 (depends on Phases 0–5) |
| Module path | `github.com/exnodes/hrm-api` |
| Project root | `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/` |

## Source-of-truth findings (read before tasks)

**Python attendance shape** (`exnodes-hrm-api/app/models/attendance.py`):
- One row per `(employee_id, date)` — unique compound key.
- Each row carries `sessions: list[AttendanceSession]` where each session is `{check_in, check_out?, is_auto_checkout}`.
- `is_late` is a row-level flag set from the **first** check-in vs configurable threshold (default 09:00 in `COMPANY_TIMEZONE`, default `Asia/Ho_Chi_Minh`).
- GPS validated only on check-in when `OFFICE_GPS_ENABLED=true`. Haversine vs configured office lat/lng/radius.
- Half-day flag in source attendance is **not** a column — half-day comes from `leave_requests.leave_period`. Spec §5.5 mentions `late_and_halfday` migration which refers to leave half-day, not attendance. We add `is_half_day` to attendance for the admin-manual case (HR records partial-day attendance for someone who came in late and went home early).
- Multiple sessions per day supported: user checks out, then checks back in for a second session (e.g. lunch back). All sessions stored in a child table.

**Endpoints in Python**:
- `POST /api/v1/attendance/check-in` (body: lat, lng, accuracy)
- `POST /api/v1/attendance/check-out`
- `GET  /api/v1/attendance/today` → `TodayStatusRead {status, is_late, sessions[], current_check_in, monthly_count, streak}`
- `GET  /api/v1/attendance` → monthly **matrix** (rows = employees, cells = days). Permission `attendance:read`. Managers see all; non-managers see only their own row.
- `GET  /api/v1/attendance/export` and `/api/v1/attendance/export/{employee_id}` → Excel.

**Spec §6.3 permissions already declared**:
- `PermAttendanceRead = "attendance:read"`
- `PermAttendanceManage = "attendance:manage_data"`

**Decisions for this phase** (extension over plain Python port):
1. **Relational shape**: `attendance` (row per user+date) **+** `attendance_sessions` (child rows). Unique `(user_id, date)`.
2. **Admin manual CRUD** added on top of Python (Python has no admin create endpoint, but BA EP-004 requires HR adjustments).
3. **Excel export** is **deferred** to a follow-up phase — out of scope here to keep size bounded. Add a TODO comment in the handler.
4. **Late threshold** read from `system_config` (Phase 8 will add the row). For Phase 6, fallback to env vars `LATE_THRESHOLD_HOUR`/`LATE_THRESHOLD_MINUTE` (defaults `9` / `0`) and `COMPANY_TIMEZONE` (default `Asia/Ho_Chi_Minh`). Service reads via config helper, not via system_config repo (system_config not yet built).
5. **GPS validation** is wired but disabled by default (`OFFICE_GPS_ENABLED=false`). Office coords from env.
6. **Auto-checkout cron job** is deferred (no cron infra yet); leave a stub method `AutoCheckoutOpenSessions(ctx, cutoff)` that the cron can wire into later.
7. **Matrix endpoint** ported exactly — pagination by employees, per-employee per-day cells. Leave-half-day combined-cell logic is preserved.

---

## Tasks

- [ ] T1 — Create migration `000011_create_attendance.up/.down.sql` with `attendance` + `attendance_sessions` tables and triggers
- [ ] T2 — Register attendance permissions (already in registry per spec; verify and add to `PermissionGroups`)
- [ ] T3 — Add attendance env keys to `internal/config/config.go` and `.env.example`
- [ ] T4 — Implement `internal/models/attendance.go` (`Attendance` + `AttendanceSession`)
- [ ] T5 — Implement `internal/dto/attendance.go` with all request/response shapes
- [ ] T6 — Implement `internal/repositories/attendance_repo.go` (interface + impl)
- [ ] T7 — Implement `internal/services/attendance_helpers.go` (timezone, threshold, haversine, hours-between)
- [ ] T8 — Implement `internal/services/attendance_service.go` — check-in / check-out / today / get / list
- [ ] T9 — Extend `attendance_service.go` — admin create / update / delete / matrix
- [ ] T10 — Implement `internal/handlers/attendance_handler.go` with Swagger annotations
- [ ] T11 — Wire routes into `cmd/server/main.go`
- [ ] T12 — Service tests: check-in happy path + late detection
- [ ] T13 — Service tests: check-in conflict, check-out flows, multi-session, missing perm
- [ ] T14 — Service tests: list filters, ownership rule, admin CRUD
- [ ] T15 — Service tests: matrix endpoint (managers see all, employees see own)
- [ ] T16 — Run migration and full unit-test suite; regenerate Swagger
- [ ] T17 — Self end-to-end verification against the running server; log to `docs/superpowers/verification/phase-06.md`
- [ ] T18 — Update README endpoint table; final commit

---

## T1 — Migration `000011_create_attendance`

- [ ] Create the up migration file.

```bash
mkdir -p /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations
```

Expected: command exits 0 (directory may already exist from Phase 0).

- [ ] Write `migrations/000011_create_attendance.up.sql`:

```sql
-- 000011_create_attendance.up.sql
-- Phase 6 — Attendance

CREATE TABLE IF NOT EXISTS attendance (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    date        DATE NOT NULL,
    is_late     BOOLEAN     NOT NULL DEFAULT FALSE,
    is_half_day BOOLEAN     NOT NULL DEFAULT FALSE,
    work_location TEXT      NULL,
    notes       TEXT        NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ NULL,
    CONSTRAINT attendance_user_date_unique UNIQUE (user_id, date),
    CONSTRAINT attendance_work_location_chk
        CHECK (work_location IS NULL OR work_location IN ('office','remote','hybrid','field'))
);

CREATE INDEX IF NOT EXISTS idx_attendance_user_id    ON attendance(user_id);
CREATE INDEX IF NOT EXISTS idx_attendance_date       ON attendance(date);
CREATE INDEX IF NOT EXISTS idx_attendance_user_date  ON attendance(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_attendance_is_deleted ON attendance(is_deleted);

DROP TRIGGER IF EXISTS trg_attendance_updated_at ON attendance;
CREATE TRIGGER trg_attendance_updated_at
    BEFORE UPDATE ON attendance
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS attendance_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attendance_id   UUID NOT NULL REFERENCES attendance(id) ON DELETE CASCADE,
    check_in        TIMESTAMPTZ NOT NULL,
    check_out       TIMESTAMPTZ NULL,
    is_auto_checkout BOOLEAN    NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ NULL,
    CONSTRAINT attendance_sessions_checkout_after_checkin
        CHECK (check_out IS NULL OR check_out >= check_in)
);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_att_id ON attendance_sessions(attendance_id);
CREATE INDEX IF NOT EXISTS idx_attendance_sessions_check_in ON attendance_sessions(check_in);
CREATE INDEX IF NOT EXISTS idx_attendance_sessions_is_deleted ON attendance_sessions(is_deleted);

-- Only one OPEN (check_out IS NULL) session per attendance row.
CREATE UNIQUE INDEX IF NOT EXISTS uq_attendance_sessions_one_open
    ON attendance_sessions(attendance_id)
    WHERE check_out IS NULL AND is_deleted = FALSE;

DROP TRIGGER IF EXISTS trg_attendance_sessions_updated_at ON attendance_sessions;
CREATE TRIGGER trg_attendance_sessions_updated_at
    BEFORE UPDATE ON attendance_sessions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

- [ ] Write `migrations/000011_create_attendance.down.sql`:

```sql
-- 000011_create_attendance.down.sql
DROP TRIGGER IF EXISTS trg_attendance_sessions_updated_at ON attendance_sessions;
DROP TRIGGER IF EXISTS trg_attendance_updated_at ON attendance;
DROP TABLE IF EXISTS attendance_sessions;
DROP TABLE IF EXISTS attendance;
```

- [ ] Verify migration applies cleanly.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make migrate-up
```

Expected output (last line): `migration applied: 000011_create_attendance` (or `golang-migrate` equivalent showing version `11`).

- [ ] Verify rollback works.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make migrate-down && make migrate-up
```

Expected: both succeed, final state is back at version 11.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add migrations/000011_create_attendance.up.sql migrations/000011_create_attendance.down.sql && git commit -m "feat(phase-06): add attendance + attendance_sessions migration"
```

Expected: one new commit, working tree clean for these two files.

---

## T2 — Verify attendance permissions in registry

The spec §6.3 already lists `PermAttendanceRead` and `PermAttendanceManage`. Confirm they exist and appear in `PermissionGroups`.

- [ ] Read the file.

```bash
grep -n "PermAttendance" /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/permissions/registry.go
```

Expected: two matches (`PermAttendanceRead`, `PermAttendanceManage`).

- [ ] Verify both constants are exposed in `PermissionGroups`. If absent, append:

```go
// inside PermissionGroups slice in internal/permissions/registry.go
{
    Name:        "Attendance",
    Description: "Attendance check-in/out, matrix, manual entry",
    Permissions: []Permission{
        PermAttendanceRead,
        PermAttendanceManage,
    },
},
```

- [ ] If a change was made, commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/permissions/registry.go && git commit -m "chore(phase-06): expose Attendance permission group in registry"
```

Expected: clean commit (or skip if no diff).

---

## T3 — Attendance config keys

- [ ] Edit `internal/config/config.go` to add attendance settings. Locate the `Config` struct and add the following fields (preserve existing fields):

```go
// Attendance / office-presence settings
CompanyTimezone        string  `env:"COMPANY_TIMEZONE"          envDefault:"Asia/Ho_Chi_Minh"`
LateThresholdHour      int     `env:"LATE_THRESHOLD_HOUR"       envDefault:"9"`
LateThresholdMinute    int     `env:"LATE_THRESHOLD_MINUTE"     envDefault:"0"`
CheckoutThresholdHour  int     `env:"CHECKOUT_THRESHOLD_HOUR"   envDefault:"18"`
CheckoutThresholdMinute int    `env:"CHECKOUT_THRESHOLD_MINUTE" envDefault:"0"`
OfficeGPSEnabled       bool    `env:"OFFICE_GPS_ENABLED"        envDefault:"false"`
OfficeLatitude         float64 `env:"OFFICE_LATITUDE"           envDefault:"0.0"`
OfficeLongitude        float64 `env:"OFFICE_LONGITUDE"          envDefault:"0.0"`
OfficeRadiusMeters     float64 `env:"OFFICE_RADIUS_METERS"      envDefault:"50.0"`
HalfDayHoursThreshold  float64 `env:"HALF_DAY_HOURS_THRESHOLD"  envDefault:"4.0"`
```

> NOTE: if Phase 0 uses `godotenv` + manual parsing instead of `caarlos0/env`, adapt: read each via `os.Getenv` with the same defaults inside the config loader.

- [ ] Append to `.env.example`:

```dotenv
# --- Attendance (Phase 6) ---
COMPANY_TIMEZONE=Asia/Ho_Chi_Minh
LATE_THRESHOLD_HOUR=9
LATE_THRESHOLD_MINUTE=0
CHECKOUT_THRESHOLD_HOUR=18
CHECKOUT_THRESHOLD_MINUTE=0
OFFICE_GPS_ENABLED=false
OFFICE_LATITUDE=0.0
OFFICE_LONGITUDE=0.0
OFFICE_RADIUS_METERS=50.0
HALF_DAY_HOURS_THRESHOLD=4.0
```

- [ ] Verify project still compiles.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: empty output, exit code 0.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/config/config.go .env.example && git commit -m "feat(phase-06): add attendance/office config keys"
```

---

## T4 — Models

- [ ] Write `internal/models/attendance.go`:

```go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Attendance is a per-user, per-day attendance row.
// Sessions live in attendance_sessions and are loaded via the AttendanceSessions association.
type Attendance struct {
	BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Date         time.Time `gorm:"type:date;not null;index" json:"date"`
	IsLate       bool      `gorm:"not null;default:false" json:"is_late"`
	IsHalfDay    bool      `gorm:"not null;default:false" json:"is_half_day"`
	WorkLocation *string   `gorm:"type:text" json:"work_location,omitempty"`
	Notes        *string   `gorm:"type:text" json:"notes,omitempty"`

	User     *User                `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Sessions []AttendanceSession  `gorm:"foreignKey:AttendanceID;references:ID;constraint:OnDelete:CASCADE" json:"sessions,omitempty"`
}

func (Attendance) TableName() string { return "attendance" }

type AttendanceSession struct {
	BaseModel
	AttendanceID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"attendance_id"`
	CheckIn        time.Time  `gorm:"not null" json:"check_in"`
	CheckOut       *time.Time `json:"check_out,omitempty"`
	IsAutoCheckout bool       `gorm:"not null;default:false" json:"is_auto_checkout"`
}

func (AttendanceSession) TableName() string { return "attendance_sessions" }
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/models/attendance.go && git commit -m "feat(phase-06): add Attendance and AttendanceSession models"
```

---

## T5 — DTOs

- [ ] Write `internal/dto/attendance.go`:

```go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ---------- Requests ----------

type AttendanceCheckInReq struct {
	// Optional override. Defaults to "now" in company TZ when omitted.
	CheckIn      *time.Time `json:"check_in,omitempty"`
	WorkLocation *string    `json:"work_location,omitempty" binding:"omitempty,oneof=office remote hybrid field"`
	Notes        *string    `json:"notes,omitempty"`
	// GPS — only validated when OFFICE_GPS_ENABLED=true.
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Accuracy  *float64 `json:"accuracy,omitempty"`
}

type AttendanceCheckOutReq struct {
	CheckOut *time.Time `json:"check_out,omitempty"`
	Notes    *string    `json:"notes,omitempty"`
}

type AttendanceAdminCreateReq struct {
	UserID       uuid.UUID  `json:"user_id" binding:"required"`
	Date         string     `json:"date" binding:"required" example:"2026-05-15"` // YYYY-MM-DD in company TZ
	CheckIn      *time.Time `json:"check_in,omitempty"`
	CheckOut     *time.Time `json:"check_out,omitempty"`
	IsLate       *bool      `json:"is_late,omitempty"`
	IsHalfDay    *bool      `json:"is_half_day,omitempty"`
	WorkLocation *string    `json:"work_location,omitempty" binding:"omitempty,oneof=office remote hybrid field"`
	Notes        *string    `json:"notes,omitempty"`
}

type AttendanceAdminUpdateReq struct {
	IsLate       *bool      `json:"is_late,omitempty"`
	IsHalfDay    *bool      `json:"is_half_day,omitempty"`
	WorkLocation *string    `json:"work_location,omitempty" binding:"omitempty,oneof=office remote hybrid field"`
	Notes        *string    `json:"notes,omitempty"`
	// Adjust the first session's times (admin correction).
	CheckIn  *time.Time `json:"check_in,omitempty"`
	CheckOut *time.Time `json:"check_out,omitempty"`
}

type AttendanceListQuery struct {
	Page         int    `form:"page,default=1"      binding:"min=1"`
	PageSize     int    `form:"page_size,default=20" binding:"min=1,max=100"`
	UserID       string `form:"user_id"`
	DepartmentID string `form:"department_id"`
	StartDate    string `form:"start_date"` // YYYY-MM-DD inclusive
	EndDate      string `form:"end_date"`   // YYYY-MM-DD inclusive
	Status       string `form:"status"`     // on_time|late|absent
}

type AttendanceMatrixQuery struct {
	Month        int    `form:"month"        binding:"omitempty,min=1,max=12"`
	Year         int    `form:"year"         binding:"omitempty,min=2000"`
	Page         int    `form:"page,default=1"        binding:"min=1"`
	PageSize     int    `form:"page_size,default=20"  binding:"min=1,max=100"`
	Search       string `form:"search"`
	DepartmentID string `form:"department_id"`
	Status       string `form:"status"` // CSV: on_time,late,absent,on_leave
}

// ---------- Responses ----------

type AttendanceSessionRead struct {
	ID              uuid.UUID  `json:"id"`
	CheckIn         time.Time  `json:"check_in"`
	CheckOut        *time.Time `json:"check_out,omitempty"`
	IsAutoCheckout  bool       `json:"is_auto_checkout"`
	HoursWorked     *float64   `json:"hours_worked,omitempty"`
}

type AttendanceUserBrief struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
}

type AttendanceRead struct {
	ID           uuid.UUID               `json:"id"`
	UserID       uuid.UUID               `json:"user_id"`
	User         *AttendanceUserBrief    `json:"user,omitempty"`
	Date         string                  `json:"date"` // YYYY-MM-DD in company TZ
	IsLate       bool                    `json:"is_late"`
	IsHalfDay    bool                    `json:"is_half_day"`
	WorkLocation *string                 `json:"work_location,omitempty"`
	Notes        *string                 `json:"notes,omitempty"`
	Sessions     []AttendanceSessionRead `json:"sessions"`
	CheckIn      *time.Time              `json:"check_in,omitempty"`     // first session
	CheckOut     *time.Time              `json:"check_out,omitempty"`    // last session
	HoursWorked  *float64                `json:"hours_worked,omitempty"` // total
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

type TodayStatusRead struct {
	Status         string                  `json:"status"` // not_checked_in|checked_in|checked_out
	IsLate         bool                    `json:"is_late"`
	Sessions       []AttendanceSessionRead `json:"sessions"`
	CurrentCheckIn *time.Time              `json:"current_check_in,omitempty"`
	MonthlyCount   int                     `json:"monthly_count"`
	Streak         int                     `json:"streak"`
}

// ---------- Matrix ----------

type AttendanceCellRead struct {
	Date             string                  `json:"date"`
	Day              int                     `json:"day"`
	Status           string                  `json:"status"`
	CheckIn          *time.Time              `json:"check_in,omitempty"`
	CheckOut         *time.Time              `json:"check_out,omitempty"`
	HoursWorked      *float64                `json:"hours_worked,omitempty"`
	IsLate           bool                    `json:"is_late"`
	Sessions         []AttendanceSessionRead `json:"sessions,omitempty"`
}

type AttendanceRowRead struct {
	EmployeeID       uuid.UUID                          `json:"employee_id"`
	EmployeeName     string                             `json:"employee_name"`
	AvatarURL        *string                            `json:"avatar_url,omitempty"`
	DepartmentName   *string                            `json:"department_name,omitempty"`
	Cells            map[int]AttendanceCellRead         `json:"cells"`
	TotalLateMinutes int                                `json:"total_late_minutes"`
	TotalEarlyMinutes int                               `json:"total_early_minutes"`
}

type AttendanceMatrixRead struct {
	Year        int                  `json:"year"`
	Month       int                  `json:"month"`
	DaysInMonth int                  `json:"days_in_month"`
	Items       []AttendanceRowRead  `json:"items"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
	TotalPages  int                  `json:"total_pages"`
}
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/dto/attendance.go && git commit -m "feat(phase-06): add attendance DTOs (check-in/out, admin, matrix)"
```

---

## T6 — Repository

- [ ] Write `internal/repositories/attendance_repo.go`:

```go
package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/models"
)

type AttendanceListFilter struct {
	UserID       *uuid.UUID
	DepartmentID *uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	Status       string // on_time | late
	Page         int
	PageSize     int
}

type AttendanceRepository interface {
	Create(ctx context.Context, a *models.Attendance) error
	Update(ctx context.Context, a *models.Attendance) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	FindByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.Attendance, error)
	List(ctx context.Context, f AttendanceListFilter) ([]models.Attendance, int64, error)
	MonthlyCheckInCount(ctx context.Context, userID uuid.UUID, year int, month int) (int64, error)
	DatesWithCheckIn(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]time.Time, error)
	ListForUsersInRange(ctx context.Context, userIDs []uuid.UUID, from, to time.Time) ([]models.Attendance, error)

	CreateSession(ctx context.Context, s *models.AttendanceSession) error
	UpdateSession(ctx context.Context, s *models.AttendanceSession) error
	OpenSessionsBefore(ctx context.Context, cutoff time.Time) ([]models.AttendanceSession, error)
}

type attendanceRepo struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepo{db: db}
}

// notDeleted is a scope shared with the rest of the project (re-declared here only
// if models.NotDeleted is not yet exported; otherwise use models.NotDeleted).
func notDeleted(db *gorm.DB) *gorm.DB { return db.Where("is_deleted = ?", false) }

func (r *attendanceRepo) Create(ctx context.Context, a *models.Attendance) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *attendanceRepo) Update(ctx context.Context, a *models.Attendance) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *attendanceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": &now}).Error
}

func (r *attendanceRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Sessions", "is_deleted = ?", false).
		Preload("User").
		First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepo) FindByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Preload("Sessions", "is_deleted = ?", false).
		Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepo) List(ctx context.Context, f AttendanceListFilter) ([]models.Attendance, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Attendance{}).Scopes(notDeleted)
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.DepartmentID != nil {
		q = q.Joins("JOIN users u ON u.id = attendance.user_id").
			Where("u.department_id = ?", *f.DepartmentID)
	}
	if f.StartDate != nil {
		q = q.Where("date >= ?", f.StartDate.Format("2006-01-02"))
	}
	if f.EndDate != nil {
		q = q.Where("date <= ?", f.EndDate.Format("2006-01-02"))
	}
	switch strings.ToLower(f.Status) {
	case "late":
		q = q.Where("is_late = ?", true)
	case "on_time":
		q = q.Where("is_late = ?", false)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	var rows []models.Attendance
	err := q.
		Preload("Sessions", "is_deleted = ?", false).
		Preload("User").
		Order("date DESC, created_at DESC").
		Limit(size).Offset(offset).
		Find(&rows).Error
	return rows, total, err
}

func (r *attendanceRepo) MonthlyCheckInCount(ctx context.Context, userID uuid.UUID, year, month int) (int64, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Scopes(notDeleted).
		Joins("JOIN attendance_sessions s ON s.attendance_id = attendance.id AND s.is_deleted = false").
		Where("attendance.user_id = ? AND attendance.date BETWEEN ? AND ?",
			userID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Distinct("attendance.id").
		Count(&n).Error
	return n, err
}

func (r *attendanceRepo) DatesWithCheckIn(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]time.Time, error) {
	type row struct{ D time.Time }
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Scopes(notDeleted).
		Joins("JOIN attendance_sessions s ON s.attendance_id = attendance.id AND s.is_deleted = false").
		Where("attendance.user_id = ? AND attendance.date BETWEEN ? AND ?",
			userID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Distinct("attendance.date").
		Order("attendance.date").
		Pluck("attendance.date", &rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]time.Time, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.D)
	}
	return out, nil
}

func (r *attendanceRepo) ListForUsersInRange(ctx context.Context, userIDs []uuid.UUID, from, to time.Time) ([]models.Attendance, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var rows []models.Attendance
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Scopes(notDeleted).
		Preload("Sessions", "is_deleted = ?", false).
		Where("user_id IN ? AND date BETWEEN ? AND ?",
			userIDs, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Find(&rows).Error
	return rows, err
}

func (r *attendanceRepo) CreateSession(ctx context.Context, s *models.AttendanceSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *attendanceRepo) UpdateSession(ctx context.Context, s *models.AttendanceSession) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *attendanceRepo) OpenSessionsBefore(ctx context.Context, cutoff time.Time) ([]models.AttendanceSession, error) {
	var rows []models.AttendanceSession
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Where("check_out IS NULL AND check_in < ?", cutoff).
		Find(&rows).Error
	return rows, err
}
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/repositories/attendance_repo.go && git commit -m "feat(phase-06): add attendance repository (CRUD, monthly count, range listing)"
```

---

## T7 — Service helpers

- [ ] Write `internal/services/attendance_helpers.go`:

```go
package services

import (
	"math"
	"time"
)

// haversineMeters returns great-circle distance in meters.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0
	rad := math.Pi / 180.0
	phi1, phi2 := lat1*rad, lat2*rad
	dPhi := (lat2 - lat1) * rad
	dLam := (lon2 - lon1) * rad
	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLam/2)*math.Sin(dLam/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// loadTZ returns the configured company timezone, defaulting to UTC on parse error.
func loadTZ(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// todayInTZ returns "now" and today's date (midnight) in the given timezone.
func todayInTZ(loc *time.Location) (time.Time, time.Time) {
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return now, today
}

// isWorkday returns true for Mon-Fri.
func isWorkday(d time.Time) bool {
	wd := d.Weekday()
	return wd != time.Saturday && wd != time.Sunday
}

// hoursBetween returns total hours between in and out, rounded to 2 decimals.
// Returns nil when out is nil.
func hoursBetween(in time.Time, out *time.Time) *float64 {
	if out == nil {
		return nil
	}
	h := math.Round(out.Sub(in).Hours()*100) / 100
	return &h
}

// thresholdAt returns time-of-day h:m on the same calendar day as ref, in ref's location.
func thresholdAt(ref time.Time, hour, minute int) time.Time {
	return time.Date(ref.Year(), ref.Month(), ref.Day(), hour, minute, 0, 0, ref.Location())
}

// parseDateYMD parses YYYY-MM-DD in the given location.
func parseDateYMD(s string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, loc)
}
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/attendance_helpers.go && git commit -m "feat(phase-06): add attendance service helpers (haversine, tz, hours)"
```

---

## T8 — Service core (check-in / check-out / today / get / list)

- [ ] Write `internal/services/attendance_service.go`:

```go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/config"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	perm "github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/repositories"
)

type AttendanceService struct {
	cfg      *config.Config
	repo     repositories.AttendanceRepository
	userRepo repositories.UserRepository
	roleRepo repositories.RoleRepository
}

func NewAttendanceService(
	cfg *config.Config,
	repo repositories.AttendanceRepository,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
) *AttendanceService {
	return &AttendanceService{cfg: cfg, repo: repo, userRepo: userRepo, roleRepo: roleRepo}
}

// ---------- shared helpers ----------

func (s *AttendanceService) tz() *time.Location {
	return loadTZ(s.cfg.CompanyTimezone)
}

func (s *AttendanceService) effectivePerms(ctx context.Context, userID uuid.UUID) (map[perm.Permission]struct{}, error) {
	// Phase 1 exposed a helper to resolve permissions; replicate the lookup here
	// in case it's not exported. roleRepo.FindByUserID returns the user's roles
	// with their permissions JSONB already populated.
	roles, err := s.roleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make(map[perm.Permission]struct{}, 8)
	for _, r := range roles {
		for _, p := range r.Permissions {
			out[perm.Permission(p)] = struct{}{}
		}
	}
	return out, nil
}

func (s *AttendanceService) hasAny(ps map[perm.Permission]struct{}, want ...perm.Permission) bool {
	if _, ok := ps[perm.PermAll]; ok {
		return true
	}
	for _, p := range want {
		if _, ok := ps[p]; ok {
			return true
		}
	}
	return false
}

func (s *AttendanceService) toRead(a *models.Attendance) dto.AttendanceRead {
	out := dto.AttendanceRead{
		ID:           a.ID,
		UserID:       a.UserID,
		Date:         a.Date.Format("2006-01-02"),
		IsLate:       a.IsLate,
		IsHalfDay:    a.IsHalfDay,
		WorkLocation: a.WorkLocation,
		Notes:        a.Notes,
		Sessions:     make([]dto.AttendanceSessionRead, 0, len(a.Sessions)),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
	var totalHours float64
	for _, sess := range a.Sessions {
		hw := hoursBetween(sess.CheckIn, sess.CheckOut)
		out.Sessions = append(out.Sessions, dto.AttendanceSessionRead{
			ID:             sess.ID,
			CheckIn:        sess.CheckIn,
			CheckOut:       sess.CheckOut,
			IsAutoCheckout: sess.IsAutoCheckout,
			HoursWorked:    hw,
		})
		if hw != nil {
			totalHours += *hw
		}
	}
	if len(a.Sessions) > 0 {
		first := a.Sessions[0]
		last := a.Sessions[len(a.Sessions)-1]
		ci := first.CheckIn
		out.CheckIn = &ci
		out.CheckOut = last.CheckOut
		if totalHours > 0 {
			out.HoursWorked = &totalHours
		}
	}
	if a.User != nil {
		out.User = &dto.AttendanceUserBrief{
			ID:        a.User.ID,
			FullName:  a.User.FullName,
			Email:     a.User.Email,
			AvatarURL: a.User.AvatarURL,
		}
	}
	return out
}

// ---------- Check-in ----------

func (s *AttendanceService) CheckIn(ctx context.Context, currentUser *models.User, req dto.AttendanceCheckInReq) (dto.AttendanceRead, error) {
	loc := s.tz()
	nowLocal, todayLocal := todayInTZ(loc)
	when := nowLocal
	if req.CheckIn != nil {
		when = req.CheckIn.In(loc)
	}

	// GPS validation
	if s.cfg.OfficeGPSEnabled {
		if req.Latitude == nil || req.Longitude == nil {
			return dto.AttendanceRead{}, apperr.ErrBadRequest("GPS coordinates required")
		}
		if req.Accuracy != nil && *req.Accuracy > s.cfg.OfficeRadiusMeters {
			return dto.AttendanceRead{}, apperr.ErrBadRequest(
				fmt.Sprintf("GPS accuracy (%dm) is too low", int(*req.Accuracy)))
		}
		dist := haversineMeters(*req.Latitude, *req.Longitude, s.cfg.OfficeLatitude, s.cfg.OfficeLongitude)
		if dist > s.cfg.OfficeRadiusMeters {
			return dto.AttendanceRead{}, apperr.ErrBadRequest(
				fmt.Sprintf("You are %dm from the office (limit %dm)", int(dist), int(s.cfg.OfficeRadiusMeters)))
		}
	}

	existing, err := s.repo.FindByUserAndDate(ctx, currentUser.ID, todayLocal)
	if err != nil {
		return dto.AttendanceRead{}, err
	}

	if existing != nil {
		// If there is an OPEN session, reject (user must check out first).
		for _, sess := range existing.Sessions {
			if sess.CheckOut == nil {
				return dto.AttendanceRead{}, apperr.ErrConflict("You are already checked in")
			}
		}
		// Append another session to the same day.
		sess := &models.AttendanceSession{
			AttendanceID: existing.ID,
			CheckIn:      when.UTC(),
		}
		if err := s.repo.CreateSession(ctx, sess); err != nil {
			return dto.AttendanceRead{}, err
		}
		// Reload with sessions for the response.
		reloaded, err := s.repo.FindByID(ctx, existing.ID)
		if err != nil {
			return dto.AttendanceRead{}, err
		}
		return s.toRead(reloaded), nil
	}

	// First check-in of the day → compute is_late.
	lateAt := thresholdAt(when, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
	isLate := when.After(lateAt)

	row := &models.Attendance{
		UserID:       currentUser.ID,
		Date:         todayLocal,
		IsLate:       isLate,
		WorkLocation: req.WorkLocation,
		Notes:        req.Notes,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return dto.AttendanceRead{}, err
	}
	sess := &models.AttendanceSession{
		AttendanceID: row.ID,
		CheckIn:      when.UTC(),
	}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return dto.AttendanceRead{}, err
	}
	reloaded, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(reloaded), nil
}

// ---------- Check-out ----------

func (s *AttendanceService) CheckOut(ctx context.Context, currentUser *models.User, req dto.AttendanceCheckOutReq) (dto.AttendanceRead, error) {
	loc := s.tz()
	nowLocal, todayLocal := todayInTZ(loc)
	when := nowLocal
	if req.CheckOut != nil {
		when = req.CheckOut.In(loc)
	}

	row, err := s.repo.FindByUserAndDate(ctx, currentUser.ID, todayLocal)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	if row == nil {
		return dto.AttendanceRead{}, apperr.ErrBadRequest("No check-in found for today")
	}

	var open *models.AttendanceSession
	for i := range row.Sessions {
		if row.Sessions[i].CheckOut == nil {
			open = &row.Sessions[i]
			break
		}
	}
	if open == nil {
		return dto.AttendanceRead{}, apperr.ErrConflict("You are not currently checked in")
	}

	whenUTC := when.UTC()
	if whenUTC.Before(open.CheckIn) {
		return dto.AttendanceRead{}, apperr.ErrBadRequest("Check-out time cannot be before check-in")
	}
	open.CheckOut = &whenUTC
	if err := s.repo.UpdateSession(ctx, open); err != nil {
		return dto.AttendanceRead{}, err
	}

	// Update notes / half-day flag.
	if req.Notes != nil {
		row.Notes = req.Notes
	}
	// Compute total hours; flag half-day if under threshold.
	var total float64
	reloaded, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	for _, sess := range reloaded.Sessions {
		if hw := hoursBetween(sess.CheckIn, sess.CheckOut); hw != nil {
			total += *hw
		}
	}
	if total > 0 && total < s.cfg.HalfDayHoursThreshold {
		reloaded.IsHalfDay = true
	}
	if req.Notes != nil {
		reloaded.Notes = req.Notes
	}
	if err := s.repo.Update(ctx, reloaded); err != nil {
		return dto.AttendanceRead{}, err
	}

	final, err := s.repo.FindByID(ctx, reloaded.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(final), nil
}

// ---------- Today status ----------

func (s *AttendanceService) Today(ctx context.Context, currentUser *models.User) (dto.TodayStatusRead, error) {
	loc := s.tz()
	_, todayLocal := todayInTZ(loc)

	row, err := s.repo.FindByUserAndDate(ctx, currentUser.ID, todayLocal)
	if err != nil {
		return dto.TodayStatusRead{}, err
	}

	out := dto.TodayStatusRead{Status: "not_checked_in", Sessions: []dto.AttendanceSessionRead{}}
	if row != nil && len(row.Sessions) > 0 {
		out.IsLate = row.IsLate
		for _, sess := range row.Sessions {
			out.Sessions = append(out.Sessions, dto.AttendanceSessionRead{
				ID:             sess.ID,
				CheckIn:        sess.CheckIn,
				CheckOut:       sess.CheckOut,
				IsAutoCheckout: sess.IsAutoCheckout,
				HoursWorked:    hoursBetween(sess.CheckIn, sess.CheckOut),
			})
		}
		last := row.Sessions[len(row.Sessions)-1]
		if last.CheckOut == nil {
			out.Status = "checked_in"
			ci := last.CheckIn
			out.CurrentCheckIn = &ci
		} else {
			out.Status = "checked_out"
		}
	}

	cnt, err := s.repo.MonthlyCheckInCount(ctx, currentUser.ID, todayLocal.Year(), int(todayLocal.Month()))
	if err != nil {
		return dto.TodayStatusRead{}, err
	}
	out.MonthlyCount = int(cnt)

	from := todayLocal.AddDate(-1, 0, 0)
	dates, err := s.repo.DatesWithCheckIn(ctx, currentUser.ID, from, todayLocal)
	if err != nil {
		return dto.TodayStatusRead{}, err
	}
	hit := make(map[string]struct{}, len(dates))
	for _, d := range dates {
		hit[d.Format("2006-01-02")] = struct{}{}
	}
	streak := 0
	d := todayLocal
	if isWorkday(d) {
		if _, ok := hit[d.Format("2006-01-02")]; !ok {
			d = d.AddDate(0, 0, -1)
		}
	}
	for d.After(from) || d.Equal(from) {
		if isWorkday(d) {
			if _, ok := hit[d.Format("2006-01-02")]; ok {
				streak++
			} else {
				break
			}
		}
		d = d.AddDate(0, 0, -1)
	}
	out.Streak = streak
	return out, nil
}

// ---------- Get by ID ----------

func (s *AttendanceService) Get(ctx context.Context, currentUser *models.User, id uuid.UUID) (dto.AttendanceRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AttendanceRead{}, apperr.ErrNotFound("attendance")
		}
		return dto.AttendanceRead{}, err
	}
	if row.UserID != currentUser.ID {
		ps, err := s.effectivePerms(ctx, currentUser.ID)
		if err != nil {
			return dto.AttendanceRead{}, err
		}
		if !s.hasAny(ps, perm.PermAttendanceRead, perm.PermAttendanceManage) {
			return dto.AttendanceRead{}, apperr.ErrForbidden("not allowed")
		}
	}
	return s.toRead(row), nil
}

// ---------- List (admin or filtered) ----------

func (s *AttendanceService) List(ctx context.Context, currentUser *models.User, q dto.AttendanceListQuery) (dto.PaginatedData[dto.AttendanceRead], error) {
	ps, err := s.effectivePerms(ctx, currentUser.ID)
	if err != nil {
		return dto.PaginatedData[dto.AttendanceRead]{}, err
	}
	canManage := s.hasAny(ps, perm.PermAttendanceManage)
	canRead := canManage || s.hasAny(ps, perm.PermAttendanceRead)

	loc := s.tz()
	f := repositories.AttendanceListFilter{
		Page:     q.Page,
		PageSize: q.PageSize,
		Status:   q.Status,
	}
	if q.StartDate != "" {
		t, err := parseDateYMD(q.StartDate, loc)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperr.ErrBadRequest("invalid start_date")
		}
		f.StartDate = &t
	}
	if q.EndDate != "" {
		t, err := parseDateYMD(q.EndDate, loc)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperr.ErrBadRequest("invalid end_date")
		}
		f.EndDate = &t
	}
	if q.UserID != "" {
		uid, err := uuid.Parse(q.UserID)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperr.ErrBadRequest("invalid user_id")
		}
		f.UserID = &uid
	}
	if q.DepartmentID != "" {
		did, err := uuid.Parse(q.DepartmentID)
		if err != nil {
			return dto.PaginatedData[dto.AttendanceRead]{}, apperr.ErrBadRequest("invalid department_id")
		}
		f.DepartmentID = &did
	}

	// Ownership / scope rules.
	if !canRead {
		me := currentUser.ID
		f.UserID = &me
	} else if !canManage {
		// canRead but not manage → can read any but in this codebase we mirror
		// Python "manage_data sees all"; without it, force own.
		me := currentUser.ID
		if f.UserID == nil || *f.UserID != me {
			f.UserID = &me
		}
	}

	rows, total, err := s.repo.List(ctx, f)
	if err != nil {
		return dto.PaginatedData[dto.AttendanceRead]{}, err
	}
	items := make([]dto.AttendanceRead, 0, len(rows))
	for i := range rows {
		items = append(items, s.toRead(&rows[i]))
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 20
	}
	totalPages := int((total + int64(size) - 1) / int64(size))
	return dto.PaginatedData[dto.AttendanceRead]{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: totalPages,
	}, nil
}
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0. If `roleRepo.FindByUserID` doesn't exist with that exact signature in Phase 1, swap to whatever the auth phase exposes (e.g. `userRepo.GetEffectivePermissions(ctx, userID)`). The contract here: given a user ID return their permission strings.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/attendance_service.go && git commit -m "feat(phase-06): attendance service — check-in, check-out, today, list"
```

---

## T9 — Service admin operations + matrix

- [ ] Append to `internal/services/attendance_service.go` (or write to `internal/services/attendance_matrix.go`):

```go
package services

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

// ---------- Admin: create ----------

func (s *AttendanceService) AdminCreate(ctx context.Context, req dto.AttendanceAdminCreateReq) (dto.AttendanceRead, error) {
	loc := s.tz()
	day, err := parseDateYMD(req.Date, loc)
	if err != nil {
		return dto.AttendanceRead{}, apperr.ErrBadRequest("invalid date (YYYY-MM-DD)")
	}

	// Conflict if a row already exists.
	existing, err := s.repo.FindByUserAndDate(ctx, req.UserID, day)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	if existing != nil {
		return dto.AttendanceRead{}, apperr.ErrConflict("Attendance for this user/date already exists")
	}

	row := &models.Attendance{
		UserID:       req.UserID,
		Date:         day,
		WorkLocation: req.WorkLocation,
		Notes:        req.Notes,
	}
	if req.IsLate != nil {
		row.IsLate = *req.IsLate
	}
	if req.IsHalfDay != nil {
		row.IsHalfDay = *req.IsHalfDay
	}

	// Auto-derive is_late if not given and check-in supplied.
	if req.CheckIn != nil && req.IsLate == nil {
		ci := req.CheckIn.In(loc)
		lateAt := thresholdAt(ci, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
		row.IsLate = ci.After(lateAt)
	}

	if err := s.repo.Create(ctx, row); err != nil {
		return dto.AttendanceRead{}, err
	}
	if req.CheckIn != nil {
		sess := &models.AttendanceSession{
			AttendanceID: row.ID,
			CheckIn:      req.CheckIn.UTC(),
			CheckOut:     req.CheckOut,
		}
		if req.CheckOut != nil {
			co := req.CheckOut.UTC()
			sess.CheckOut = &co
		}
		if err := s.repo.CreateSession(ctx, sess); err != nil {
			return dto.AttendanceRead{}, err
		}
	}
	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(final), nil
}

// ---------- Admin: update ----------

func (s *AttendanceService) AdminUpdate(ctx context.Context, id uuid.UUID, req dto.AttendanceAdminUpdateReq) (dto.AttendanceRead, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.AttendanceRead{}, apperr.ErrNotFound("attendance")
	}
	if req.IsLate != nil {
		row.IsLate = *req.IsLate
	}
	if req.IsHalfDay != nil {
		row.IsHalfDay = *req.IsHalfDay
	}
	if req.WorkLocation != nil {
		row.WorkLocation = req.WorkLocation
	}
	if req.Notes != nil {
		row.Notes = req.Notes
	}

	if (req.CheckIn != nil || req.CheckOut != nil) && len(row.Sessions) > 0 {
		first := &row.Sessions[0]
		if req.CheckIn != nil {
			first.CheckIn = req.CheckIn.UTC()
		}
		if req.CheckOut != nil {
			co := req.CheckOut.UTC()
			first.CheckOut = &co
		}
		if err := s.repo.UpdateSession(ctx, first); err != nil {
			return dto.AttendanceRead{}, err
		}
	} else if req.CheckIn != nil {
		sess := &models.AttendanceSession{
			AttendanceID: row.ID,
			CheckIn:      req.CheckIn.UTC(),
		}
		if req.CheckOut != nil {
			co := req.CheckOut.UTC()
			sess.CheckOut = &co
		}
		if err := s.repo.CreateSession(ctx, sess); err != nil {
			return dto.AttendanceRead{}, err
		}
	}
	if err := s.repo.Update(ctx, row); err != nil {
		return dto.AttendanceRead{}, err
	}
	final, err := s.repo.FindByID(ctx, row.ID)
	if err != nil {
		return dto.AttendanceRead{}, err
	}
	return s.toRead(final), nil
}

// ---------- Admin: delete ----------

func (s *AttendanceService) AdminDelete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return apperr.ErrNotFound("attendance")
	}
	return s.repo.SoftDelete(ctx, id)
}

// ---------- Matrix ----------

// Status values used in matrix cells.
const (
	matrixOnTime = "on_time"
	matrixLate   = "late"
	matrixAbsent = "absent"
	matrixWknd   = "weekend"
	matrixNoData = "no_data"
)

func (s *AttendanceService) Matrix(ctx context.Context, currentUser *models.User, q dto.AttendanceMatrixQuery) (dto.AttendanceMatrixRead, error) {
	loc := s.tz()
	now, _ := todayInTZ(loc)

	year := q.Year
	month := q.Month
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 20
	}

	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	last := first.AddDate(0, 1, 0).Add(-time.Second)
	daysInMonth := last.Day()

	ps, err := s.effectivePerms(ctx, currentUser.ID)
	if err != nil {
		return dto.AttendanceMatrixRead{}, err
	}
	canManage := s.hasAny(ps, perm.PermAttendanceManage)

	var employees []models.User
	if canManage {
		filter := repositories.UserListFilter{
			Page:     1,
			PageSize: 10000, // matrix loads all employees then paginates rows; matches Python
			Search:   q.Search,
		}
		if q.DepartmentID != "" {
			did, err := uuid.Parse(q.DepartmentID)
			if err != nil {
				return dto.AttendanceMatrixRead{}, apperr.ErrBadRequest("invalid department_id")
			}
			filter.DepartmentID = &did
		}
		rows, _, err := s.userRepo.List(ctx, filter)
		if err != nil {
			return dto.AttendanceMatrixRead{}, err
		}
		employees = rows
	} else {
		// Only self.
		me, err := s.userRepo.FindByID(ctx, currentUser.ID)
		if err != nil {
			return dto.AttendanceMatrixRead{}, err
		}
		employees = []models.User{*me}
	}

	ids := make([]uuid.UUID, 0, len(employees))
	for _, e := range employees {
		ids = append(ids, e.ID)
	}
	records, err := s.repo.ListForUsersInRange(ctx, ids, first, last)
	if err != nil {
		return dto.AttendanceMatrixRead{}, err
	}
	byEmp := make(map[uuid.UUID]map[string]models.Attendance, len(employees))
	for _, r := range records {
		m, ok := byEmp[r.UserID]
		if !ok {
			m = make(map[string]models.Attendance)
			byEmp[r.UserID] = m
		}
		m[r.Date.Format("2006-01-02")] = r
	}

	statusSet := parseCSVSet(q.Status)

	rows := make([]dto.AttendanceRowRead, 0, len(employees))
	lateAt := thresholdAt(first, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
	earlyAt := thresholdAt(first, s.cfg.CheckoutThresholdHour, s.cfg.CheckoutThresholdMinute)

	for _, emp := range employees {
		cells := make(map[int]dto.AttendanceCellRead, daysInMonth)
		empRecs := byEmp[emp.ID]
		var totalLate, totalEarly int
		cellStatusUnion := make(map[string]struct{}, 8)

		for d := 1; d <= daysInMonth; d++ {
			day := time.Date(year, time.Month(month), d, 0, 0, 0, 0, loc)
			cell := dto.AttendanceCellRead{
				Date: day.Format("2006-01-02"),
				Day:  d,
			}
			switch {
			case day.Weekday() == time.Saturday || day.Weekday() == time.Sunday:
				cell.Status = matrixWknd
			default:
				rec, ok := empRecs[day.Format("2006-01-02")]
				if ok && len(rec.Sessions) > 0 {
					sort.Slice(rec.Sessions, func(i, j int) bool {
						return rec.Sessions[i].CheckIn.Before(rec.Sessions[j].CheckIn)
					})
					f := rec.Sessions[0]
					l := rec.Sessions[len(rec.Sessions)-1]
					ci := f.CheckIn
					cell.CheckIn = &ci
					cell.CheckOut = l.CheckOut
					var total float64
					for _, sess := range rec.Sessions {
						if hw := hoursBetween(sess.CheckIn, sess.CheckOut); hw != nil {
							total += *hw
						}
					}
					if total > 0 {
						cell.HoursWorked = &total
					}
					cell.IsLate = rec.IsLate
					if rec.IsLate {
						cell.Status = matrixLate
					} else {
						cell.Status = matrixOnTime
					}

					// late / early minute totals
					ciLocal := f.CheckIn.In(loc)
					ref := thresholdAt(ciLocal, s.cfg.LateThresholdHour, s.cfg.LateThresholdMinute)
					if ciLocal.After(ref) {
						totalLate += int(ciLocal.Sub(ref).Minutes())
					}
					if l.CheckOut != nil {
						coLocal := l.CheckOut.In(loc)
						refE := thresholdAt(coLocal, s.cfg.CheckoutThresholdHour, s.cfg.CheckoutThresholdMinute)
						if coLocal.Before(refE) {
							totalEarly += int(refE.Sub(coLocal).Minutes())
						}
					}
				} else if day.Before(now) {
					cell.Status = matrixAbsent
				} else {
					cell.Status = matrixNoData
				}
			}
			cells[d] = cell
			cellStatusUnion[cell.Status] = struct{}{}
		}

		if statusSet != nil {
			matched := false
			for k := range statusSet {
				if _, ok := cellStatusUnion[k]; ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		row := dto.AttendanceRowRead{
			EmployeeID:        emp.ID,
			EmployeeName:      emp.FullName,
			AvatarURL:         emp.AvatarURL,
			Cells:             cells,
			TotalLateMinutes:  totalLate,
			TotalEarlyMinutes: totalEarly,
		}
		rows = append(rows, row)
	}

	total := len(rows)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	pageRows := rows[start:end]
	totalPages := (total + size - 1) / size

	_ = lateAt
	_ = earlyAt // reserved for absent-day calculations in later refinement

	return dto.AttendanceMatrixRead{
		Year:        year,
		Month:       month,
		DaysInMonth: daysInMonth,
		Items:       pageRows,
		Total:       total,
		Page:        page,
		PageSize:    size,
		TotalPages:  totalPages,
	}, nil
}

func parseCSVSet(s string) map[string]struct{} {
	if s == "" {
		return nil
	}
	out := make(map[string]struct{}, 4)
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			v := s[start:i]
			if v != "" {
				out[v] = struct{}{}
			}
			start = i + 1
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0. If `repositories.UserListFilter` differs in name (e.g. `UserFilter`), adapt the call site to whatever Phase 2 produced.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/ && git commit -m "feat(phase-06): admin CRUD + monthly matrix for attendance service"
```

---

## T10 — Handler with Swagger

- [ ] Write `internal/handlers/attendance_handler.go`:

```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/dto"
	apperr "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/middleware"
	"github.com/exnodes/hrm-api/internal/services"
)

type AttendanceHandler struct {
	svc *services.AttendanceService
}

func NewAttendanceHandler(svc *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

// CheckIn godoc
// @Summary  Record a check-in for the authenticated user
// @Tags     Attendance
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      dto.AttendanceCheckInReq  true  "check-in payload"
// @Success  200   {object}  dto.Response[dto.AttendanceRead]
// @Failure  400   {object}  dto.Response[any]
// @Failure  401   {object}  dto.Response[any]
// @Failure  409   {object}  dto.Response[any]
// @Router   /attendance/check-in [post]
func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req dto.AttendanceCheckInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	cu := middleware.MustCurrentUser(c)
	out, err := h.svc.CheckIn(c.Request.Context(), cu, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Message: "Checked in", Data: out})
}

// CheckOut godoc
// @Summary  Record a check-out for the authenticated user
// @Tags     Attendance
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      dto.AttendanceCheckOutReq  false  "check-out payload"
// @Success  200   {object}  dto.Response[dto.AttendanceRead]
// @Failure  400   {object}  dto.Response[any]
// @Failure  409   {object}  dto.Response[any]
// @Router   /attendance/check-out [post]
func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	var req dto.AttendanceCheckOutReq
	_ = c.ShouldBindJSON(&req)
	cu := middleware.MustCurrentUser(c)
	out, err := h.svc.CheckOut(c.Request.Context(), cu, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Message: "Checked out", Data: out})
}

// Today godoc
// @Summary  Get today's attendance status for the authenticated user
// @Tags     Attendance
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  dto.Response[dto.TodayStatusRead]
// @Router   /attendance/today [get]
func (h *AttendanceHandler) Today(c *gin.Context) {
	cu := middleware.MustCurrentUser(c)
	out, err := h.svc.Today(c.Request.Context(), cu)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.TodayStatusRead]{Success: true, Data: out})
}

// Me godoc
// @Summary  List my attendance rows
// @Tags     Attendance
// @Produce  json
// @Security BearerAuth
// @Param    page       query  int     false  "page"
// @Param    page_size  query  int     false  "page size"
// @Param    start_date query  string  false  "YYYY-MM-DD"
// @Param    end_date   query  string  false  "YYYY-MM-DD"
// @Success  200  {object}  dto.Response[dto.PaginatedData[dto.AttendanceRead]]
// @Router   /attendance/me [get]
func (h *AttendanceHandler) Me(c *gin.Context) {
	var q dto.AttendanceListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	cu := middleware.MustCurrentUser(c)
	q.UserID = cu.ID.String()
	out, err := h.svc.List(c.Request.Context(), cu, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.AttendanceRead]]{Success: true, Data: out})
}

// List godoc
// @Summary  List attendance rows (admin/HR)
// @Tags     Attendance
// @Produce  json
// @Security BearerAuth
// @Param    page          query  int     false  "page"
// @Param    page_size     query  int     false  "page size"
// @Param    user_id       query  string  false  "filter by user"
// @Param    department_id query  string  false  "filter by department"
// @Param    start_date    query  string  false  "YYYY-MM-DD"
// @Param    end_date      query  string  false  "YYYY-MM-DD"
// @Param    status        query  string  false  "on_time|late"
// @Success  200  {object}  dto.Response[dto.PaginatedData[dto.AttendanceRead]]
// @Router   /attendance [get]
func (h *AttendanceHandler) List(c *gin.Context) {
	var q dto.AttendanceListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	cu := middleware.MustCurrentUser(c)
	out, err := h.svc.List(c.Request.Context(), cu, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.AttendanceRead]]{Success: true, Data: out})
}

// Get godoc
// @Summary  Get an attendance row by ID
// @Tags     Attendance
// @Produce  json
// @Security BearerAuth
// @Param    id  path  string  true  "attendance id"
// @Success  200  {object}  dto.Response[dto.AttendanceRead]
// @Failure  404  {object}  dto.Response[any]
// @Router   /attendance/{id} [get]
func (h *AttendanceHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	cu := middleware.MustCurrentUser(c)
	out, err := h.svc.Get(c.Request.Context(), cu, id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Data: out})
}

// AdminCreate godoc
// @Summary  Admin manual create of an attendance row
// @Tags     Attendance
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  dto.AttendanceAdminCreateReq  true  "payload"
// @Success  201  {object}  dto.Response[dto.AttendanceRead]
// @Failure  409  {object}  dto.Response[any]
// @Router   /attendance [post]
func (h *AttendanceHandler) AdminCreate(c *gin.Context) {
	var req dto.AttendanceAdminCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.AdminCreate(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[dto.AttendanceRead]{Success: true, Message: "Created", Data: out})
}

// AdminUpdate godoc
// @Summary  Admin update of an attendance row
// @Tags     Attendance
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id    path  string                          true  "attendance id"
// @Param    body  body  dto.AttendanceAdminUpdateReq    true  "payload"
// @Success  200  {object}  dto.Response[dto.AttendanceRead]
// @Router   /attendance/{id} [patch]
func (h *AttendanceHandler) AdminUpdate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	var req dto.AttendanceAdminUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	out, err := h.svc.AdminUpdate(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceRead]{Success: true, Message: "Updated", Data: out})
}

// AdminDelete godoc
// @Summary  Admin soft-delete of an attendance row
// @Tags     Attendance
// @Produce  json
// @Security BearerAuth
// @Param    id  path  string  true  "attendance id"
// @Success  200  {object}  dto.Response[any]
// @Router   /attendance/{id} [delete]
func (h *AttendanceHandler) AdminDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperr.ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.AdminDelete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Deleted"})
}

// Matrix godoc
// @Summary  Monthly attendance matrix
// @Tags     Attendance
// @Produce  json
// @Security BearerAuth
// @Param    month         query  int     false  "1-12"
// @Param    year          query  int     false  "YYYY"
// @Param    page          query  int     false  "page"
// @Param    page_size     query  int     false  "page size"
// @Param    search        query  string  false  "name filter"
// @Param    department_id query  string  false  "department UUID"
// @Param    status        query  string  false  "CSV: on_time,late,absent"
// @Success  200  {object}  dto.Response[dto.AttendanceMatrixRead]
// @Router   /attendance/matrix [get]
func (h *AttendanceHandler) Matrix(c *gin.Context) {
	var q dto.AttendanceMatrixQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperr.ErrBadRequest(err.Error()))
		return
	}
	cu := middleware.MustCurrentUser(c)
	out, err := h.svc.Matrix(c.Request.Context(), cu, q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.AttendanceMatrixRead]{Success: true, Data: out})
}
```

- [ ] Verify compile.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```

Expected: exit 0. If `middleware.MustCurrentUser` is named differently (e.g. `middleware.CurrentUser`), adapt.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/handlers/attendance_handler.go && git commit -m "feat(phase-06): attendance handlers with Swagger annotations"
```

---

## T11 — Wire routes into `cmd/server/main.go`

- [ ] Edit `cmd/server/main.go`. In the dependency-construction section, after the user/role services exist, add:

```go
// Phase 6 — Attendance
attendanceRepo := repositories.NewAttendanceRepository(db)
attendanceSvc := services.NewAttendanceService(cfg, attendanceRepo, userRepo, roleRepo)
attendanceH := handlers.NewAttendanceHandler(attendanceSvc)
```

- [ ] In the router registration block (inside the `authed := v1.Group("")` scope), add:

```go
att := authed.Group("/attendance")
{
    // Self-service
    att.POST("/check-in",  attendanceH.CheckIn)
    att.POST("/check-out", attendanceH.CheckOut)
    att.GET("/today",      attendanceH.Today)
    att.GET("/me",         attendanceH.Me)

    // Permissioned reads
    att.GET("",        middleware.RequirePerms(perm.PermAttendanceRead),   attendanceH.List)
    att.GET("/matrix", middleware.RequirePerms(perm.PermAttendanceRead),   attendanceH.Matrix)
    att.GET(":id",     middleware.RequirePerms(perm.PermAttendanceRead),   attendanceH.Get)

    // Admin manage
    att.POST("",        middleware.RequirePerms(perm.PermAttendanceManage), attendanceH.AdminCreate)
    att.PATCH(":id",    middleware.RequirePerms(perm.PermAttendanceManage), attendanceH.AdminUpdate)
    att.DELETE(":id",   middleware.RequirePerms(perm.PermAttendanceManage), attendanceH.AdminDelete)
}
```

> Note Gin route precedence: `GET /attendance/today`, `/me`, `/matrix` must be registered before `GET /:id` so the literal segments win over the wildcard. The block above respects that ordering.

- [ ] Verify build + boot.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./... && go vet ./...
```

Expected: exit 0 on both.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add cmd/server/main.go && git commit -m "feat(phase-06): wire attendance routes into server"
```

---

## T12 — Service tests: check-in happy path + late detection

- [ ] Read `internal/services/testhelper_test.go` and confirm fixtures `makeUser`, `makeAdmin`, `makeReadonly`, `freshDB(t)` exist (Phase 0–1 set them up). If they're named differently (`createUser`, etc.), use whatever Phase 0 produced — same pattern across all tests below.

- [ ] Write `internal/services/attendance_service_test.go`:

```go
package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/repositories"
)

func newAttendanceSvc(t *testing.T) *AttendanceService {
	t.Helper()
	db := freshDB(t)
	cfg := &config.Config{
		CompanyTimezone:         "Asia/Ho_Chi_Minh",
		LateThresholdHour:       9,
		LateThresholdMinute:     0,
		CheckoutThresholdHour:   18,
		CheckoutThresholdMinute: 0,
		HalfDayHoursThreshold:   4.0,
	}
	return NewAttendanceService(
		cfg,
		repositories.NewAttendanceRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewRoleRepository(db),
	)
}

func TestCheckIn_HappyPath(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "alice@exnodes.vn")

	// Pin check-in to 08:30 local — should NOT be late.
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	ci := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	out, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci})

	require.NoError(t, err)
	assert.False(t, out.IsLate)
	require.Len(t, out.Sessions, 1)
	assert.Nil(t, out.Sessions[0].CheckOut)
}

func TestCheckIn_LateDetection(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "bob@exnodes.vn")

	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	ci := time.Date(2026, 5, 15, 9, 30, 0, 0, loc) // 30 min after threshold
	out, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci})

	require.NoError(t, err)
	assert.True(t, out.IsLate)
}
```

- [ ] Run tests.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go test ./internal/services -run 'TestCheckIn_' -v
```

Expected: `PASS` for both tests.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/attendance_service_test.go && git commit -m "test(phase-06): attendance check-in happy path + late detection"
```

---

## T13 — Service tests: conflicts, multi-session, check-out flows

- [ ] Append to `internal/services/attendance_service_test.go`:

```go
func TestCheckIn_AlreadyOpenSession_Conflicts(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "carol@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)
	t1 := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &t1})
	require.NoError(t, err)

	t2 := time.Date(2026, 5, 15, 8, 35, 0, 0, loc)
	_, err = svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &t2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already checked in")
}

func TestCheckOut_WithoutCheckIn_BadRequest(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "dave@exnodes.vn")

	_, err := svc.CheckOut(context.Background(), u, dto.AttendanceCheckOutReq{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No check-in")
}

func TestCheckOut_HappyPath_NotHalfDayWhenLongEnough(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "eve@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	ci := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := time.Date(2026, 5, 15, 17, 30, 0, 0, loc) // 9h
	out, err := svc.CheckOut(context.Background(), u, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)
	assert.False(t, out.IsHalfDay)
	require.NotNil(t, out.HoursWorked)
	assert.InDelta(t, 9.0, *out.HoursWorked, 0.05)
}

func TestCheckOut_HalfDayFlaggedForShortDay(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "frank@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	ci := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := time.Date(2026, 5, 15, 11, 0, 0, 0, loc) // 2.5h < 4h threshold
	out, err := svc.CheckOut(context.Background(), u, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)
	assert.True(t, out.IsHalfDay)
}

func TestCheckOut_AlreadyClosed_Conflicts(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "grace@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	ci := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci})
	require.NoError(t, err)

	co := time.Date(2026, 5, 15, 17, 0, 0, 0, loc)
	_, err = svc.CheckOut(context.Background(), u, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.NoError(t, err)

	_, err = svc.CheckOut(context.Background(), u, dto.AttendanceCheckOutReq{CheckOut: &co})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not currently checked in")
}

func TestCheckIn_SecondSession_AfterCheckout(t *testing.T) {
	svc := newAttendanceSvc(t)
	u := makeUser(t, svc, "henry@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	ci1 := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci1})
	require.NoError(t, err)
	co1 := time.Date(2026, 5, 15, 12, 0, 0, 0, loc)
	_, err = svc.CheckOut(context.Background(), u, dto.AttendanceCheckOutReq{CheckOut: &co1})
	require.NoError(t, err)

	ci2 := time.Date(2026, 5, 15, 13, 30, 0, 0, loc)
	out, err := svc.CheckIn(context.Background(), u, dto.AttendanceCheckInReq{CheckIn: &ci2})
	require.NoError(t, err)
	require.Len(t, out.Sessions, 2)
	assert.Nil(t, out.Sessions[1].CheckOut)
}
```

- [ ] Run.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go test ./internal/services -run 'TestCheckIn_|TestCheckOut_' -v
```

Expected: all `PASS`.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/attendance_service_test.go && git commit -m "test(phase-06): attendance conflicts, multi-session, half-day flow"
```

---

## T14 — Service tests: list filters, ownership, admin CRUD

- [ ] Append:

```go
func TestList_OwnerSeesOnlySelf(t *testing.T) {
	svc := newAttendanceSvc(t)
	alice := makeUser(t, svc, "alice2@exnodes.vn")
	bob := makeUser(t, svc, "bob2@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	t1 := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, _ = svc.CheckIn(context.Background(), alice, dto.AttendanceCheckInReq{CheckIn: &t1})
	_, _ = svc.CheckIn(context.Background(), bob, dto.AttendanceCheckInReq{CheckIn: &t1})

	out, err := svc.List(context.Background(), alice, dto.AttendanceListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, alice.ID, out.Items[0].UserID)
}

func TestList_ManagerSeesAll(t *testing.T) {
	svc := newAttendanceSvc(t)
	mgr := makeAdmin(t, svc, "manager@exnodes.vn") // has * permissions
	alice := makeUser(t, svc, "alice3@exnodes.vn")
	bob := makeUser(t, svc, "bob3@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	t1 := time.Date(2026, 5, 15, 8, 30, 0, 0, loc)
	_, _ = svc.CheckIn(context.Background(), alice, dto.AttendanceCheckInReq{CheckIn: &t1})
	_, _ = svc.CheckIn(context.Background(), bob, dto.AttendanceCheckInReq{CheckIn: &t1})

	out, err := svc.List(context.Background(), mgr, dto.AttendanceListQuery{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(out.Items), 2)
}

func TestList_DateRangeFilter(t *testing.T) {
	svc := newAttendanceSvc(t)
	mgr := makeAdmin(t, svc, "mgr2@exnodes.vn")
	alice := makeUser(t, svc, "alice4@exnodes.vn")
	loc, _ := time.LoadLocation(svc.cfg.CompanyTimezone)

	// Two attendance rows on different dates via admin create.
	_, err := svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{
		UserID:  alice.ID,
		Date:    "2026-05-10",
	})
	require.NoError(t, err)
	_, err = svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{
		UserID:  alice.ID,
		Date:    "2026-05-20",
	})
	require.NoError(t, err)

	out, err := svc.List(context.Background(), mgr, dto.AttendanceListQuery{
		Page: 1, PageSize: 50,
		StartDate: "2026-05-15", EndDate: "2026-05-31",
	})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "2026-05-20", out.Items[0].Date)
	_ = loc
}

func TestAdminCreate_DuplicateConflicts(t *testing.T) {
	svc := newAttendanceSvc(t)
	alice := makeUser(t, svc, "alice5@exnodes.vn")

	_, err := svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{UserID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)
	_, err = svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{UserID: alice.ID, Date: "2026-05-10"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAdminUpdate_ChangesNotesAndIsLate(t *testing.T) {
	svc := newAttendanceSvc(t)
	alice := makeUser(t, svc, "alice6@exnodes.vn")

	created, err := svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{UserID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)

	newNotes := "manual late entry"
	trueP := true
	updated, err := svc.AdminUpdate(context.Background(), created.ID, dto.AttendanceAdminUpdateReq{
		Notes:  &newNotes,
		IsLate: &trueP,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.Notes)
	assert.Equal(t, "manual late entry", *updated.Notes)
	assert.True(t, updated.IsLate)
}

func TestAdminDelete_RemovesRow(t *testing.T) {
	svc := newAttendanceSvc(t)
	alice := makeUser(t, svc, "alice7@exnodes.vn")
	mgr := makeAdmin(t, svc, "mgr3@exnodes.vn")

	created, err := svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{UserID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)
	require.NoError(t, svc.AdminDelete(context.Background(), created.ID))

	_, err = svc.Get(context.Background(), mgr, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGet_OwnershipEnforced(t *testing.T) {
	svc := newAttendanceSvc(t)
	alice := makeUser(t, svc, "alice8@exnodes.vn")
	bob := makeUser(t, svc, "bob4@exnodes.vn") // plain user, no perms

	created, err := svc.AdminCreate(context.Background(), dto.AttendanceAdminCreateReq{UserID: alice.ID, Date: "2026-05-10"})
	require.NoError(t, err)

	_, err = svc.Get(context.Background(), bob, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}
```

- [ ] Run.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go test ./internal/services -run 'TestList_|TestAdmin|TestGet_' -v
```

Expected: all `PASS`.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/attendance_service_test.go && git commit -m "test(phase-06): list filters, ownership, admin CRUD"
```

---

## T15 — Service tests: matrix

- [ ] Append:

```go
func TestMatrix_ManagerSeesAllEmployees(t *testing.T) {
	svc := newAttendanceSvc(t)
	mgr := makeAdmin(t, svc, "mgr4@exnodes.vn")
	makeUser(t, svc, "e1@exnodes.vn")
	makeUser(t, svc, "e2@exnodes.vn")

	out, err := svc.Matrix(context.Background(), mgr, dto.AttendanceMatrixQuery{Month: 5, Year: 2026, Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, 2026, out.Year)
	assert.Equal(t, 5, out.Month)
	assert.GreaterOrEqual(t, len(out.Items), 3) // mgr + 2 employees (at least)
	assert.Equal(t, 31, out.DaysInMonth)
}

func TestMatrix_EmployeeSeesOwnRow(t *testing.T) {
	svc := newAttendanceSvc(t)
	emp := makeUser(t, svc, "soleemp@exnodes.vn")

	out, err := svc.Matrix(context.Background(), emp, dto.AttendanceMatrixQuery{Month: 5, Year: 2026, Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, emp.ID, out.Items[0].EmployeeID)
}

func TestMatrix_WeekendsMarked(t *testing.T) {
	svc := newAttendanceSvc(t)
	emp := makeUser(t, svc, "weekendcheck@exnodes.vn")

	out, err := svc.Matrix(context.Background(), emp, dto.AttendanceMatrixQuery{Month: 5, Year: 2026, Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	row := out.Items[0]
	// May 2 2026 is a Saturday, May 3 a Sunday.
	assert.Equal(t, "weekend", row.Cells[2].Status)
	assert.Equal(t, "weekend", row.Cells[3].Status)
}
```

- [ ] Run.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go test ./internal/services -run 'TestMatrix_' -v
```

Expected: all `PASS`.

- [ ] Run full attendance suite + module-level vet.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go vet ./... && go test ./internal/services -run 'TestCheckIn_|TestCheckOut_|TestList_|TestAdmin|TestGet_|TestMatrix_' -count=1 -v
```

Expected: vet clean; all attendance tests `PASS`.

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add internal/services/attendance_service_test.go && git commit -m "test(phase-06): attendance matrix coverage"
```

---

## T16 — Migrate, regenerate Swagger, run full suite

- [ ] Apply migrations on a clean DB.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make migrate-down && make migrate-up
```

Expected: applies through `000011_create_attendance`. Final line shows version 11.

- [ ] Confirm tables exist.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && psql "$(grep -E '^DATABASE_URL=' .env | cut -d= -f2-)" -c "\dt attendance*"
```

Expected: both `attendance` and `attendance_sessions` listed.

- [ ] Regenerate Swagger.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make swag || (command -v swag >/dev/null && swag init -g cmd/server/main.go -o docs/swagger)
```

Expected: writes/updates `docs/swagger/swagger.json` and `swagger.yaml` without errors.

- [ ] Run the entire service-test suite.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go test ./internal/services/... -count=1
```

Expected: `ok  github.com/exnodes/hrm-api/internal/services …`. No failures from any phase.

- [ ] Boot the server (background) for end-to-end verification.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && (nohup go run ./cmd/server >/tmp/hrm.log 2>&1 &) && sleep 3 && curl -fsS http://localhost:8080/health
```

Expected: `{"success":true,"data":{"status":"ok"}}` (or whatever Phase 0 returns from `/health`).

- [ ] Commit Swagger regenerated artifacts (if changed).

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add docs/swagger && git commit -m "chore(phase-06): regenerate Swagger for attendance endpoints" || echo "no swagger changes"
```

---

## T17 — Self end-to-end verification + log

Create the verification log directory if missing, then walk all endpoints and capture exact requests/responses.

- [ ] Set up shell vars and log target.

```bash
mkdir -p /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification
BASE=http://localhost:8080/api/v1
LOG=/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification/phase-06.md
: > "$LOG"
```

- [ ] Login as super admin (seeded in Phase 1).

```bash
ADMIN=$(curl -fsS -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@exnodes.local","password":"ChangeMe!1"}')
echo "$ADMIN" | tee -a "$LOG"
ADMIN_TOKEN=$(echo "$ADMIN" | jq -r '.data.access_token')
test -n "$ADMIN_TOKEN" -a "$ADMIN_TOKEN" != "null"
```

Expected: `success:true`; token non-empty.

- [ ] Create two test users for the walk.

```bash
curl -fsS -X POST "$BASE/users" -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" \
  -d '{"email":"walker@exnodes.local","password":"Walk1234!","full_name":"Walker User"}' | tee -a "$LOG"
curl -fsS -X POST "$BASE/users" -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" \
  -d '{"email":"second@exnodes.local","password":"Walk1234!","full_name":"Second User"}' | tee -a "$LOG"
```

Expected: two `201`-like envelopes with `success:true`.

- [ ] Login as the walker.

```bash
WALKER=$(curl -fsS -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"walker@exnodes.local","password":"Walk1234!"}')
echo "$WALKER" | tee -a "$LOG"
WALKER_TOKEN=$(echo "$WALKER" | jq -r '.data.access_token')
```

- [ ] Check-in (happy path).

```bash
curl -fsS -X POST "$BASE/attendance/check-in" -H "Authorization: Bearer $WALKER_TOKEN" \
  -H "Content-Type: application/json" -d '{"work_location":"office"}' | tee -a "$LOG"
```

Expected: `success:true`, `data.sessions[0].check_out=null`.

- [ ] Check-in again — should conflict.

```bash
echo "## conflict on double check-in" >> "$LOG"
curl -sS -o /tmp/resp.json -w 'HTTP %{http_code}\n' \
  -X POST "$BASE/attendance/check-in" -H "Authorization: Bearer $WALKER_TOKEN" \
  -H "Content-Type: application/json" -d '{}' | tee -a "$LOG"
cat /tmp/resp.json | tee -a "$LOG"
```

Expected: `HTTP 409` with `success:false`, message contains "already checked in".

- [ ] Check-out.

```bash
curl -fsS -X POST "$BASE/attendance/check-out" -H "Authorization: Bearer $WALKER_TOKEN" \
  -H "Content-Type: application/json" -d '{"notes":"e2e walk"}' | tee -a "$LOG"
```

Expected: `success:true`, last session has `check_out` non-null.

- [ ] Check-out again — should error.

```bash
echo "## conflict on double check-out" >> "$LOG"
curl -sS -o /tmp/resp.json -w 'HTTP %{http_code}\n' \
  -X POST "$BASE/attendance/check-out" -H "Authorization: Bearer $WALKER_TOKEN" \
  -H "Content-Type: application/json" -d '{}' | tee -a "$LOG"
cat /tmp/resp.json | tee -a "$LOG"
```

Expected: `HTTP 409` or `400`; message "not currently checked in".

- [ ] GET today + GET me.

```bash
curl -fsS "$BASE/attendance/today" -H "Authorization: Bearer $WALKER_TOKEN" | tee -a "$LOG"
curl -fsS "$BASE/attendance/me?page=1&page_size=10" -H "Authorization: Bearer $WALKER_TOKEN" | tee -a "$LOG"
```

Expected: `today.status="checked_out"`, `me` returns one row.

- [ ] Walker tries to list all → forbidden (no manage perm).

```bash
echo "## forbidden list as plain user" >> "$LOG"
curl -sS -o /tmp/resp.json -w 'HTTP %{http_code}\n' \
  "$BASE/attendance?page=1&page_size=20" -H "Authorization: Bearer $WALKER_TOKEN" | tee -a "$LOG"
cat /tmp/resp.json | tee -a "$LOG"
```

Expected: `HTTP 403` (failed `RequirePerms(PermAttendanceRead)`).

- [ ] Admin lists all and creates manual entry for SECOND user.

```bash
SECOND_ID=$(curl -fsS "$BASE/users?search=second" -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.data.items[0].id')
echo "second user id: $SECOND_ID" | tee -a "$LOG"

curl -fsS "$BASE/attendance?page=1&page_size=20" -H "Authorization: Bearer $ADMIN_TOKEN" | tee -a "$LOG"

CREATED=$(curl -fsS -X POST "$BASE/attendance" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$SECOND_ID\",\"date\":\"2026-05-10\",\"is_late\":true,\"notes\":\"manual\"}")
echo "$CREATED" | tee -a "$LOG"
CREATED_ID=$(echo "$CREATED" | jq -r '.data.id')
```

Expected: `201` with `success:true`; id captured.

- [ ] Admin updates the manual entry.

```bash
curl -fsS -X PATCH "$BASE/attendance/$CREATED_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" \
  -d '{"is_half_day": true, "notes": "manual revised"}' | tee -a "$LOG"
```

Expected: `success:true`, `is_half_day:true`, notes updated.

- [ ] Date-range filter.

```bash
curl -fsS "$BASE/attendance?start_date=2026-05-01&end_date=2026-05-15&page=1&page_size=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | tee -a "$LOG"
```

Expected: items only within range.

- [ ] Matrix.

```bash
curl -fsS "$BASE/attendance/matrix?month=5&year=2026&page=1&page_size=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | tee -a "$LOG"
```

Expected: `year:2026`, `month:5`, `days_in_month:31`, at least one item.

- [ ] Admin deletes the manual entry.

```bash
curl -fsS -X DELETE "$BASE/attendance/$CREATED_ID" -H "Authorization: Bearer $ADMIN_TOKEN" | tee -a "$LOG"
curl -sS -o /tmp/resp.json -w 'HTTP %{http_code}\n' \
  "$BASE/attendance/$CREATED_ID" -H "Authorization: Bearer $ADMIN_TOKEN" | tee -a "$LOG"
cat /tmp/resp.json | tee -a "$LOG"
```

Expected: first call `success:true`; follow-up `HTTP 404`.

- [ ] Unauthenticated check-in → 401.

```bash
echo "## unauthenticated check-in" >> "$LOG"
curl -sS -o /tmp/resp.json -w 'HTTP %{http_code}\n' -X POST "$BASE/attendance/check-in" \
  -H "Content-Type: application/json" -d '{}' | tee -a "$LOG"
cat /tmp/resp.json | tee -a "$LOG"
```

Expected: `HTTP 401`.

- [ ] DB spot-check.

```bash
psql "$(grep -E '^DATABASE_URL=' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env | cut -d= -f2-)" \
  -c "SELECT COUNT(*) FROM attendance WHERE is_deleted=false;" \
  -c "SELECT COUNT(*) FROM attendance_sessions WHERE is_deleted=false;" | tee -a "$LOG"
```

Expected: counts match what the walk created (walker has 1 attendance + 1 session; second-user manual entry was soft-deleted).

- [ ] Stop the server.

```bash
pkill -f 'go run ./cmd/server' || pkill -f 'exe/server' || true
```

- [ ] Commit verification log.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add docs/superpowers/verification/phase-06.md && git commit -m "docs(phase-06): end-to-end verification log"
```

---

## T18 — README update + final summary commit

- [ ] Edit `README.md`. In the endpoint table, add an Attendance section:

```markdown
### Attendance

| Method | Path                                  | Auth                              | Description                              |
|--------|---------------------------------------|-----------------------------------|------------------------------------------|
| POST   | `/api/v1/attendance/check-in`         | authenticated                     | Record check-in (creates day row if new) |
| POST   | `/api/v1/attendance/check-out`        | authenticated                     | Close the open session                   |
| GET    | `/api/v1/attendance/today`            | authenticated                     | Today's status + monthly count + streak  |
| GET    | `/api/v1/attendance/me`               | authenticated                     | List my own attendance rows              |
| GET    | `/api/v1/attendance`                  | `attendance:read`                 | List rows (admin/HR; filters)            |
| GET    | `/api/v1/attendance/matrix`           | `attendance:read`                 | Monthly attendance matrix                |
| GET    | `/api/v1/attendance/:id`              | `attendance:read` or owner        | Get a specific row                       |
| POST   | `/api/v1/attendance`                  | `attendance:manage_data`          | Admin manual create                      |
| PATCH  | `/api/v1/attendance/:id`              | `attendance:manage_data`          | Admin update                             |
| DELETE | `/api/v1/attendance/:id`              | `attendance:manage_data`          | Admin soft-delete                        |
```

- [ ] Commit.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && git add README.md && git commit -m "docs(phase-06): document attendance endpoints in README"
```

- [ ] Final self-check.

```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go vet ./... && go test ./... -count=1
```

Expected: every package passes.

---

## Definition of Done — Phase 6

All boxes below must be checked before declaring Phase 6 complete.

- [ ] `migrations/000011_create_attendance.up.sql` + `.down.sql` committed; `make migrate-up` and `make migrate-down` both succeed.
- [ ] `internal/models/attendance.go` with `Attendance` + `AttendanceSession`, BaseModel embedded.
- [ ] `internal/dto/attendance.go` covers check-in/out, admin CRUD, list/matrix queries.
- [ ] `internal/repositories/attendance_repo.go` interface + implementation; `NotDeleted` scope applied.
- [ ] `internal/services/attendance_service.go` (+ helpers + matrix file) implements: CheckIn, CheckOut, Today, Get, List, AdminCreate, AdminUpdate, AdminDelete, Matrix.
- [ ] `internal/handlers/attendance_handler.go` with Swagger annotations on every handler.
- [ ] Routes wired in `cmd/server/main.go` with `RequirePerms(...)` per spec §6.4.
- [ ] `go test ./internal/services/... -count=1` passes (includes 14+ attendance tests covering happy + error paths).
- [ ] Swagger regenerated; `/swagger/index.html` shows all 10 new endpoints.
- [ ] End-to-end verification walk completed against the running server and logged to `docs/superpowers/verification/phase-06.md`. The log includes: login → check-in → conflict check-in → check-out → conflict check-out → today → me → 403 list as employee → admin list → admin create → admin update → date-range filter → matrix → admin delete → 401 unauthenticated → DB spot-check.
- [ ] README endpoint table updated.
- [ ] No edits to migration files of prior phases (`000001`–`000010`).
- [ ] `go vet ./...` clean.

## Out of scope (for follow-up)

- Excel export endpoints (`/attendance/export` and `/attendance/export/{employee_id}`).
- Auto-checkout cron at 23:00 company time (stub method `OpenSessionsBefore` already present in repo).
- Half-day cells in matrix derived from approved leave (overlap with Phase 5 leave matrix logic — handled in a Phase 6.5 refinement once leave matrix is finalized).
- Reading the late threshold from `system_config` (Phase 8 will introduce the table; until then env vars apply).

## Self-review notes

- Spec §5.2 audit columns: present on both tables. Triggers wired. ✓
- Spec §5.4 UUID PKs: yes. ✓
- Spec §6.4 per-route permissions: every admin route has `RequirePerms(PermAttendanceManage)`; reads use `PermAttendanceRead`. Owner endpoints (`/today`, `/me`, `/check-in`, `/check-out`) use auth only. ✓
- Spec §7.3 repo returns `gorm.ErrRecordNotFound`; service maps via `apperr.ErrNotFound`. ✓
- Constraint §7 unique(user_id, date): enforced by DB, also defended at service level by `FindByUserAndDate` before insert. ✓
- Constraint §8 timezone: every late check uses `s.tz()` from `COMPANY_TIMEZONE`. ✓
- Constraint §9 DoD includes verification log committed to `docs/superpowers/verification/phase-06.md`. ✓
