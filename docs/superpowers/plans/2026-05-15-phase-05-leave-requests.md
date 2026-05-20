# Phase 5 — Leave Requests + Leave Quota

| | |
|---|---|
| Status | Draft |
| Date | 2026-05-15 |
| Owner | danny.tranhoang@exnodes.vn |
| Spec | `docs/superpowers/specs/2026-05-15-go-migration-design.md` |
| Source (Python) | `exnodes-hrm-api/app/routers/leave_requests.py`, `services/leave_request.py`, `schemas/leave_request.py`, `models/leave_request.py` |
| Reference (Go) | `exn-hr/Exn-hr/backend/internal/services/leave_service.go` |
| BA | `ba-requirements/docs/PLATFORMS/WEB-APP/EP-002-leave-management/` |
| Depends on | Phases 0–4 (foundation, auth/RBAC, users incl. quota fields, departments, positions, skills) |

---

## ⚠️ REVISION NOTES (2026-05-20) — AUTHORITATIVE, read & apply before executing any task

This plan was drafted pre-schema-split AND pre-Phase-2-extras (when leave quotas got their own table). The codebase audit performed at the start of Phase 5 supersedes the task bodies below wherever they conflict. Apply these corrections:

1. **Migration number.** `000001`–`000007` are taken (latest = `000007_create_labels` from Phase 4). The Phase-5 migration is **`000008_create_leave_requests`** (NOT `000011` as written below). Final `make migrate-version` after this phase = **8**. Renumber every filename, `make migrate-*` reference, and `migrate-version` expectation accordingly.

2. **FK target = `employees(id)`, NOT `users(id)`.** The Python source's `LeaveRequest.employee_id` actually references the User document, but in the Go schema split (Phase 1) the HR profile lives on `employees`, and every cross-aggregate FK introduced from Phase 2 onward targets `employees(id)`. The leave-request `employee_id` and `created_by` columns must therefore be:
   - `employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT`
   - `created_by  UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT`
   Anywhere the task bodies write `REFERENCES users(id)` is wrong — replace with `employees(id)`. The service-layer ownership check (`request.employee_id == currentEmployee.ID` or admin perm) is unchanged in spirit, but it compares **employee IDs**, not user IDs.

3. **Leave quota table ALREADY EXISTS.** Migration `000004_phase2_extras` created `employee_leave_quotas` (one row per employee, columns `annual_leave_quota NUMERIC(6,2)` + `sick_leave_quota NUMERIC(6,2)`) with the standard 4 audit cols + trigger. The Go model is `models.EmployeeLeaveQuota` and the repo `repositories.LeaveQuotaRepository` (already wired into `cmd/server/main.go` as `quotaRepo`). The plan's Goal-section sentence "Per-user quotas live on the `users` table … there is no separate `leave_quotas` table" is **wrong** — ignore it. The balance endpoint must call `LeaveQuotaRepository.GetByEmployee` for the quota figures and SUM live (`is_deleted=false`) approved `leave_requests.total_days` per (employee_id, leave_type, year) to compute usage. Do NOT create a new quota table or add quota columns to any other table.

4. **Permission constants + seeds are complete.** `PermLeaveRead/Create/Update/Delete/Approve/Cancel/Manage` + `PermLeaveQuotaManage` already exist in `internal/permissions/registry.go` and are granted to the appropriate system roles in `seed_service.go` (Admin, HR Manager, Manager, Employee per their domain). **No seed gap to close** in this phase (contrast Phase 4's `PermAnnounceManage` gap).

5. **Endpoint inventory (matches Python — keep the POST-for-actions form).** Final routes under `/api/v1/leave-requests`:
   - `GET    /`                    — list (paginated, status[] filter, dept/position filter, search, sort)   `PermLeaveRead`
   - `POST   /`                    — create (multipart for optional `attachment` file + JSON `data` part) `PermLeaveCreate`
   - `GET    /balance/:employee_id` — `PermLeaveRead` (uses **employee** UUID, not user UUID)
   - `GET    /dashboard/me`        — auth-only (current employee derived from token)
   - `GET    /history/me`          — auth-only (paginated)
   - `GET    /:id`                  — `PermLeaveRead` (with ownership fallback for owner-without-perm: not strictly the Python behavior — verify in source before implementing; if Python requires perm, do the same)
   - `PATCH  /:id`                  — `PermLeaveUpdate` (admin) or owner-of-pending (ownership branch enforced in service)
   - `POST   /:id/approve`         — `PermLeaveApprove`
   - `POST   /:id/reject`          — `PermLeaveApprove` (Python uses the same permission for both — keep it)
   - `POST   /:id/cancel`          — `PermLeaveCancel` (admin) or owner (service branch)
   - `POST   /:id/delete`          — **POST, not DELETE** (Python: line 246 of `routers/leave_requests.py`). `PermLeaveDelete`. Admin (`PermLeaveApprove` OR wildcard) can delete any status; non-admin only `pending` and only their own.

6. **`created_by` semantics.** Python tracks who submitted the request — defaults to the current user, but when an admin (`PermLeaveManage`) creates on behalf of another employee, `employee_id` is the subject and `created_by` is the admin. The Go model must mirror this: nullable IS WRONG; both columns are NOT NULL with `created_by` defaulting to the same value as `employee_id` when no admin override.

7. **Half-day arithmetic.** `total_days = (to_date - from_date).days + 1` (inclusive, all calendar days — NO business-day skip), multiplied by `0.5` when `leave_period ∈ {morning_half, afternoon_half}`. A half-day range must satisfy `from_date == to_date` — reject otherwise (`apperrors.ErrBadRequest`).

8. **Warnings vs errors.** Insufficient quota and date overlap with the same employee's existing live (non-rejected, non-cancelled) requests are **non-blocking warnings**, NOT errors. The Create/Update responses include a `warnings: []string` array; the request is still created/updated with the warnings attached. This is the load-bearing behavior shift from a "validate strictly" mindset — preserve it.

9. **Status state machine.**
   - `pending → approved | rejected | cancelled`
   - `approved → cancelled` (balance "restored" automatically because the balance query only sums status='approved' AND is_deleted=false)
   - `rejected | cancelled` are terminal — Update/Approve/Reject/Cancel on these states all return 409.
   - On `PATCH` of an `approved` request by admin: revert status to `pending` (Python contract — preserve).

10. **Conventions (from Phases 1-4, verified):** repo = interface + lowercase impl + `New…() Interface` constructor; `models.NotDeleted` scope; soft-delete sets `is_deleted=true` + `deleted_at=NOW()`; services return `*apperrors.AppError`; per-route `middleware.RequirePerms(authSvc, permissions.PermXxx)` (**first arg `authSvc`**); attachment upload reuses Phase-2 `UploadService` (`Uploader` interface) + **mandatory `http.DetectContentType` content-sniff** (review-fix #2 pattern, also applied in Phase 4 skill icons); search via `utils.BuildILIKEPattern`; Swagger `@Security BearerAuth` on every endpoint; `make swag` regenerated + committed; tests are `package services_test`, real Postgres test DB, extend `truncateAll` in FK-safe order for `leave_requests` BEFORE `employee_leave_quotas, dependents, employees` etc. (leave_requests references employees; truncate leave_requests first).

11. **DoD = real live verification** committed to `docs/superpowers/verification/phase-05.md`. Suggested minimum flow:
    - Boot server, migrate-version=8.
    - Login as admin → create leave request for self (full_day, annual) → 201 with `total_days=1`, no warnings.
    - Create overlapping request → 201 with warnings (overlap message).
    - Create with `to_date < from_date` → 400.
    - Create morning_half with `from_date != to_date` → 400.
    - Approve → 200, status=approved.
    - GET balance → annual_used reflects the approved days; sick_used=0.
    - Cancel approved → 200, status=cancelled; balance restored on next GET balance.
    - Edit cancelled → 409.
    - Admin creates request on behalf of another employee (`employee_id != current`) → 201, `created_by` = admin's employee_id.
    - Non-admin tries to PATCH someone else's pending request → 403.
    - Non-admin tries `POST /:id/delete` on someone else's request → 403.
    - Non-admin tries `POST /:id/delete` on their own approved → 403 (only pending allowed for non-admin).
    - Attachment upload (valid PDF/PNG) → stored, URL returned. Content-spoof (text bytes with `.pdf` ext + `Content-Type: application/pdf`) → 400.
    - Soft-delete row spot-check via psql.
    - 401 (no token), 403 (Employee role lacks Approve on POST /approve).

Everything else in the task bodies (layering, commit-per-task, no placeholders, bite-sized steps) still applies. **Execute per these REVISION NOTES, not the raw task bodies where they conflict.**

---

## 0. Goals

Port the Python leave-request module 1:1 to Go + Postgres. Single `leave_requests` table with **enum string columns** for `leave_type`, `leave_period`, and `status` (Python uses `StrEnum`, not a separate types table). **Per-user quotas live on the `users` table** (`annual_leave_quota`, `sick_leave_quota`) added in Phase 2 — there is no separate `leave_quotas` table. The "balance" endpoint is a derived aggregate (sum of approved `total_days` per type within a year, subtracted from the user's stored quota).

Faithfulness rules ported from Python:
- Quota types = `{annual, sick}` only. Other types (`personal`, `maternity`, `unpaid`) have no quota and never trigger insufficient-quota warnings.
- Insufficient quota and overlapping dates produce **non-blocking warnings**, never hard errors (Python: `warnings: list[str]`). The request is still created/updated.
- `total_days = (to_date - from_date).days + 1`, multiplied by `0.5` if `leave_period in {morning_half, afternoon_half}`. Inclusive count, **all days included** (no business-day skip in Python).
- Status transitions: only `pending → approved/rejected/cancelled`; `approved → cancelled` (balance "restored" naturally because balance is a live query over approved+not-deleted rows).
- On Update, if current status is `approved` → revert to `pending`. Editing `rejected/cancelled` is forbidden.
- Ownership rules: employees can only edit/delete own requests. Delete by non-admin requires status `pending`. Admin (`leave_requests:manage`) can act on anyone.

## 1. Deliverables

1. `migrations/000011_create_leave_requests.up.sql` + `.down.sql` — single `leave_requests` table with CHECK constraints for the three enums; indexes on `employee_id`, `status`, `(from_date, to_date)`, `is_deleted`; `set_updated_at` trigger.
2. `internal/models/leave_request.go` — `LeaveRequest` with `BaseModel` embed; type aliases `LeaveType`, `LeavePeriod`, `LeaveStatus` (string constants).
3. `internal/dto/leave.go` — Create, Update, Read, RefRead, ListQuery, BalanceSummary, DashboardRead.
4. `internal/repositories/leave_request_repo.go` — interface + GORM impl with `NotDeleted` scope; filter by status list, department, position, search; sort `-created_at` / `-to_date` / `+from_date`; aggregate `SumApprovedDays(employeeID, leaveType, year)`.
5. `internal/services/leave_service.go` — Create, Update, Approve, Reject, Cancel, Delete, Get, List, GetBalance, GetMyDashboard, ListMyHistory, helper `populateRead`.
6. `internal/handlers/leave_handler.go` — 10 endpoints below, all with full Swagger annotations.
7. Route wiring in `cmd/server/main.go` + permission constants already in `internal/permissions/registry.go` (added in Phase 1).
8. Service tests: happy path + warnings + status transitions + ownership + delete rules + balance aggregation + dashboard composition.
9. `docs/superpowers/verification/phase-05.md` — captured curl session with expected DB-state spot checks.

## 2. Endpoint inventory (matches Python router exactly, with REST nouns kept on `POST` for actions as in Python)

| Method | Path | Permission |
|---|---|---|
| GET    | `/api/v1/leave-requests`                       | `PermLeaveRead` |
| GET    | `/api/v1/leave-requests/balance/:employee_id`  | `PermLeaveRead` |
| GET    | `/api/v1/leave-requests/dashboard/me`          | JWT only (self-scoped) |
| GET    | `/api/v1/leave-requests/history/me`            | JWT only (self-scoped) |
| POST   | `/api/v1/leave-requests`                       | `PermLeaveCreate` |
| GET    | `/api/v1/leave-requests/:id`                   | `PermLeaveRead` |
| PATCH  | `/api/v1/leave-requests/:id`                   | `PermLeaveUpdate` + ownership-or-manage |
| POST   | `/api/v1/leave-requests/:id/approve`           | `PermLeaveApprove` |
| POST   | `/api/v1/leave-requests/:id/reject`            | `PermLeaveApprove` |
| POST   | `/api/v1/leave-requests/:id/cancel`            | `PermLeaveCancel` |
| POST   | `/api/v1/leave-requests/:id/delete`            | `PermLeaveDelete` |

Notes:
- Attachment upload (Python uses `multipart/form-data` with `data` JSON string + `attachment` file). Go endpoint accepts the same shape via `c.PostForm("data")` + `c.FormFile("attachment")` so the FE contract is unchanged. Storage path is via Supabase S3 (Phase 0 supplied the client).
- `PermLeaveQuotaManage` exists but quota mutation happens **on the user** via the existing `PATCH /users/:id` (Phase 2) — there is no `/leave-quotas/:userID` endpoint, contrary to the prompt template, because no such endpoint exists in Python. The plan documents this deviation explicitly in the verification log.

---

## 3. Tasks

Each task ends with a commit. Self-check after every task: `go build ./...` and (where relevant) `go test ./internal/services/... -run <Test>`.

### Task 1 — Migration: leave_requests table

- [ ] Create `migrations/000011_create_leave_requests.up.sql` with the SQL below.

```sql
-- migrations/000011_create_leave_requests.up.sql
BEGIN;

CREATE TABLE leave_requests (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id       UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    from_date         DATE NOT NULL,
    to_date           DATE NOT NULL,
    leave_period      TEXT NOT NULL DEFAULT 'full_day'
                       CHECK (leave_period IN ('full_day','morning_half','afternoon_half')),
    leave_type        TEXT NOT NULL
                       CHECK (leave_type IN ('annual','sick','personal','maternity','unpaid')),
    total_days        NUMERIC(5,1) NOT NULL CHECK (total_days >= 0),
    reason            TEXT NOT NULL,
    attachment_url    TEXT,
    status            TEXT NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending','approved','rejected','cancelled')),
    created_by        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Audit columns (every entity table)
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_deleted        BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT leave_requests_date_range_chk CHECK (to_date >= from_date)
);

CREATE INDEX leave_requests_employee_id_idx     ON leave_requests(employee_id);
CREATE INDEX leave_requests_status_idx          ON leave_requests(status);
CREATE INDEX leave_requests_from_to_idx         ON leave_requests(from_date, to_date);
CREATE INDEX leave_requests_is_deleted_idx      ON leave_requests(is_deleted);
CREATE INDEX leave_requests_created_at_desc_idx ON leave_requests(created_at DESC);

CREATE TRIGGER leave_requests_set_updated_at
    BEFORE UPDATE ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
```

- [ ] Create `migrations/000011_create_leave_requests.down.sql`:

```sql
-- migrations/000011_create_leave_requests.down.sql
BEGIN;
DROP TRIGGER IF EXISTS leave_requests_set_updated_at ON leave_requests;
DROP TABLE IF EXISTS leave_requests;
COMMIT;
```

- [ ] Run: `make migrate-up`
  - Expected: `migration applied 000011/u create_leave_requests`
- [ ] Verify table exists: `psql "$DATABASE_URL" -c "\d leave_requests"`
  - Expected: shows all columns, three CHECK constraints, five indexes, one trigger.
- [ ] Run: `make migrate-down && make migrate-up`
  - Expected: down succeeds, up reapplies cleanly. Version returns to `000011`.
- [ ] **Commit:** `feat(phase-05): add leave_requests migration`

### Task 2 — Model: LeaveRequest + enum constants

- [ ] Create `internal/models/leave_request.go`:

```go
package models

import (
    "time"

    "github.com/google/uuid"
)

type LeaveType string
type LeavePeriod string
type LeaveStatus string

const (
    LeaveTypeAnnual    LeaveType = "annual"
    LeaveTypeSick      LeaveType = "sick"
    LeaveTypePersonal  LeaveType = "personal"
    LeaveTypeMaternity LeaveType = "maternity"
    LeaveTypeUnpaid    LeaveType = "unpaid"
)

const (
    LeavePeriodFullDay      LeavePeriod = "full_day"
    LeavePeriodMorningHalf  LeavePeriod = "morning_half"
    LeavePeriodAfternoonHalf LeavePeriod = "afternoon_half"
)

const (
    LeaveStatusPending   LeaveStatus = "pending"
    LeaveStatusApproved  LeaveStatus = "approved"
    LeaveStatusRejected  LeaveStatus = "rejected"
    LeaveStatusCancelled LeaveStatus = "cancelled"
)

// QuotaLeaveTypes mirrors Python's QUOTA_TYPES — only these types deduct from quota.
func QuotaLeaveTypes() map[LeaveType]struct{} {
    return map[LeaveType]struct{}{LeaveTypeAnnual: {}, LeaveTypeSick: {}}
}

type LeaveRequest struct {
    BaseModel
    EmployeeID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"employee_id"`
    FromDate      time.Time   `gorm:"type:date;not null"        json:"from_date"`
    ToDate        time.Time   `gorm:"type:date;not null"        json:"to_date"`
    LeavePeriod   LeavePeriod `gorm:"type:text;not null;default:'full_day'" json:"leave_period"`
    LeaveType     LeaveType   `gorm:"type:text;not null"        json:"leave_type"`
    TotalDays     float64     `gorm:"type:numeric(5,1);not null" json:"total_days"`
    Reason        string      `gorm:"type:text;not null"        json:"reason"`
    AttachmentURL *string     `gorm:"type:text"                 json:"attachment_url,omitempty"`
    Status        LeaveStatus `gorm:"type:text;not null;default:'pending';index" json:"status"`
    CreatedBy     uuid.UUID   `gorm:"type:uuid;not null"        json:"created_by"`
}

func (LeaveRequest) TableName() string { return "leave_requests" }
```

- [ ] Run: `go build ./...`
  - Expected: no errors.
- [ ] **Commit:** `feat(phase-05): add LeaveRequest model + enum constants`

### Task 3 — DTOs

- [ ] Create `internal/dto/leave.go`:

```go
package dto

import (
    "time"

    "github.com/exnodes/hrm-api/internal/models"
)

type LeaveRequestCreate struct {
    EmployeeID  *string             `json:"employee_id,omitempty"` // admin-only when set
    FromDate    time.Time           `json:"from_date" binding:"required"`
    ToDate      time.Time           `json:"to_date" binding:"required"`
    LeavePeriod models.LeavePeriod  `json:"leave_period" binding:"required"`
    LeaveType   models.LeaveType    `json:"leave_type" binding:"required"`
    Reason      string              `json:"reason" binding:"required,min=1"`
}

type LeaveRequestUpdate struct {
    FromDate    *time.Time           `json:"from_date,omitempty"`
    ToDate      *time.Time           `json:"to_date,omitempty"`
    LeavePeriod *models.LeavePeriod  `json:"leave_period,omitempty"`
    LeaveType   *models.LeaveType    `json:"leave_type,omitempty"`
    Reason      *string              `json:"reason,omitempty"`
}

type LeaveRefRead struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type LeaveRequestRead struct {
    ID            string              `json:"id"`
    Employee      *LeaveRefRead       `json:"employee,omitempty"`
    Department    *LeaveRefRead       `json:"department,omitempty"`
    Position      *LeaveRefRead       `json:"position,omitempty"`
    FromDate      time.Time           `json:"from_date"`
    ToDate        time.Time           `json:"to_date"`
    LeavePeriod   models.LeavePeriod  `json:"leave_period"`
    LeaveType     models.LeaveType    `json:"leave_type"`
    TotalDays     float64             `json:"total_days"`
    Reason        string              `json:"reason"`
    AttachmentURL *string             `json:"attachment_url,omitempty"`
    Status        models.LeaveStatus  `json:"status"`
    CreatedBy     string              `json:"created_by"`
    CreatedAt     time.Time           `json:"created_at"`
    UpdatedAt     time.Time           `json:"updated_at"`
}

type LeaveBalanceSummary struct {
    Year             int     `json:"year"`
    AnnualQuota      float64 `json:"annual_quota"`
    AnnualUsed       float64 `json:"annual_used"`
    AnnualRemaining  float64 `json:"annual_remaining"`
    SickQuota        float64 `json:"sick_quota"`
    SickUsed         float64 `json:"sick_used"`
    SickRemaining    float64 `json:"sick_remaining"`
    LeavesThisYear   int     `json:"leaves_this_year"`
}

type LeaveDashboardRead struct {
    Balance  LeaveBalanceSummary `json:"balance"`
    Upcoming []LeaveRequestRead  `json:"upcoming"`
    History  []LeaveRequestRead  `json:"history"`
}

type LeaveListQuery struct {
    Page         int      `form:"page,default=1" binding:"min=1"`
    PageSize     int      `form:"page_size,default=10" binding:"min=1,max=100"`
    Search       string   `form:"search"`
    Status       []string `form:"status"` // repeat-param: ?status=pending&status=approved
    DepartmentID string   `form:"department_id"`
    PositionID   string   `form:"position_id"`
}

type LeaveHistoryQuery struct {
    Page      int        `form:"page,default=1" binding:"min=1"`
    PageSize  int        `form:"page_size,default=10" binding:"min=1,max=100"`
    Status    []string   `form:"status"`
    StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
    EndDate   *time.Time `form:"end_date"   time_format:"2006-01-02"`
}
```

- [ ] Run: `go build ./...`
  - Expected: no errors.
- [ ] **Commit:** `feat(phase-05): add leave DTOs`

### Task 4 — Repository

- [ ] Create `internal/repositories/leave_request_repo.go`:

```go
package repositories

import (
    "context"
    "errors"
    "time"

    "github.com/exnodes/hrm-api/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type LeaveRequestRepository interface {
    Create(ctx context.Context, lr *models.LeaveRequest) error
    Update(ctx context.Context, lr *models.LeaveRequest) error
    Save(ctx context.Context, lr *models.LeaveRequest) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error)
    List(ctx context.Context, employeeIDs []uuid.UUID, statuses []string, page, pageSize int) ([]models.LeaveRequest, int64, error)
    ListByUser(ctx context.Context, userID uuid.UUID, filter ListByUserFilter) ([]models.LeaveRequest, int64, error)
    Upcoming(ctx context.Context, userID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error)
    History(ctx context.Context, userID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error)
    SumApprovedDays(ctx context.Context, userID uuid.UUID, year int) (map[models.LeaveType]struct {
        Days  float64
        Count int64
    }, error)
    Overlapping(ctx context.Context, userID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]models.LeaveRequest, error)
}

type ListByUserFilter struct {
    Statuses  []string
    StartDate *time.Time
    EndDate   *time.Time
    Page      int
    PageSize  int
}

type leaveRequestRepo struct{ db *gorm.DB }

func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository { return &leaveRequestRepo{db: db} }

func (r *leaveRequestRepo) base(ctx context.Context) *gorm.DB {
    return r.db.WithContext(ctx).Model(&models.LeaveRequest{}).Scopes(models.NotDeleted)
}

func (r *leaveRequestRepo) Create(ctx context.Context, lr *models.LeaveRequest) error {
    return r.db.WithContext(ctx).Create(lr).Error
}

func (r *leaveRequestRepo) Update(ctx context.Context, lr *models.LeaveRequest) error {
    return r.db.WithContext(ctx).Save(lr).Error
}

func (r *leaveRequestRepo) Save(ctx context.Context, lr *models.LeaveRequest) error {
    return r.db.WithContext(ctx).Save(lr).Error
}

func (r *leaveRequestRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
    now := time.Now().UTC()
    return r.db.WithContext(ctx).Model(&models.LeaveRequest{}).
        Where("id = ? AND is_deleted = false", id).
        Updates(map[string]any{"is_deleted": true, "deleted_at": now, "updated_at": now}).Error
}

func (r *leaveRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error) {
    var lr models.LeaveRequest
    err := r.base(ctx).Where("id = ?", id).First(&lr).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
    if err != nil {
        return nil, err
    }
    return &lr, nil
}

func (r *leaveRequestRepo) List(ctx context.Context, employeeIDs []uuid.UUID, statuses []string, page, pageSize int) ([]models.LeaveRequest, int64, error) {
    q := r.base(ctx)
    if employeeIDs != nil { // nil = no filter; empty = no matches
        if len(employeeIDs) == 0 {
            return []models.LeaveRequest{}, 0, nil
        }
        q = q.Where("employee_id IN ?", employeeIDs)
    }
    if len(statuses) > 0 {
        q = q.Where("status IN ?", statuses)
    }
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var items []models.LeaveRequest
    err := q.Order("created_at DESC").
        Offset((page - 1) * pageSize).Limit(pageSize).
        Find(&items).Error
    return items, total, err
}

func (r *leaveRequestRepo) ListByUser(ctx context.Context, userID uuid.UUID, filter ListByUserFilter) ([]models.LeaveRequest, int64, error) {
    today := time.Now().UTC().Truncate(24 * time.Hour)
    q := r.base(ctx).
        Where("employee_id = ?", userID).
        Where("(to_date < ? OR status IN ?)", today, []string{string(models.LeaveStatusRejected), string(models.LeaveStatusCancelled)})
    if len(filter.Statuses) > 0 {
        q = q.Where("status IN ?", filter.Statuses)
    }
    if filter.StartDate != nil {
        q = q.Where("from_date >= ?", *filter.StartDate)
    }
    if filter.EndDate != nil {
        q = q.Where("to_date <= ?", *filter.EndDate)
    }
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var items []models.LeaveRequest
    err := q.Order("to_date DESC").
        Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).
        Find(&items).Error
    return items, total, err
}

func (r *leaveRequestRepo) Upcoming(ctx context.Context, userID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error) {
    var items []models.LeaveRequest
    err := r.base(ctx).
        Where("employee_id = ?", userID).
        Where("status IN ?", []string{string(models.LeaveStatusPending), string(models.LeaveStatusApproved)}).
        Where("from_date >= ?", today).
        Order("from_date ASC").Limit(limit).
        Find(&items).Error
    return items, err
}

func (r *leaveRequestRepo) History(ctx context.Context, userID uuid.UUID, today time.Time, limit int) ([]models.LeaveRequest, error) {
    var items []models.LeaveRequest
    err := r.base(ctx).
        Where("employee_id = ?", userID).
        Where("to_date < ?", today).
        Order("to_date DESC").Limit(limit).
        Find(&items).Error
    return items, err
}

type aggRow struct {
    LeaveType string
    Days      float64
    Count     int64
}

func (r *leaveRequestRepo) SumApprovedDays(ctx context.Context, userID uuid.UUID, year int) (map[models.LeaveType]struct {
    Days  float64
    Count int64
}, error) {
    start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
    var rows []aggRow
    err := r.base(ctx).
        Select("leave_type, COALESCE(SUM(total_days),0) as days, COUNT(*) as count").
        Where("employee_id = ? AND status = ? AND from_date >= ? AND from_date <= ?",
            userID, models.LeaveStatusApproved, start, end).
        Group("leave_type").
        Scan(&rows).Error
    if err != nil {
        return nil, err
    }
    out := map[models.LeaveType]struct {
        Days  float64
        Count int64
    }{}
    for _, r := range rows {
        out[models.LeaveType(r.LeaveType)] = struct {
            Days  float64
            Count int64
        }{Days: r.Days, Count: r.Count}
    }
    return out, nil
}

func (r *leaveRequestRepo) Overlapping(ctx context.Context, userID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]models.LeaveRequest, error) {
    q := r.base(ctx).
        Where("employee_id = ?", userID).
        Where("status IN ?", []string{string(models.LeaveStatusPending), string(models.LeaveStatusApproved)}).
        Where("from_date <= ? AND to_date >= ?", to, from)
    if excludeID != nil {
        q = q.Where("id <> ?", *excludeID)
    }
    var items []models.LeaveRequest
    err := q.Find(&items).Error
    return items, err
}
```

- [ ] Run: `go build ./...`
  - Expected: no errors.
- [ ] **Commit:** `feat(phase-05): add LeaveRequestRepository`

### Task 5 — Service: skeleton + helpers

- [ ] Create `internal/services/leave_service.go` skeleton (Create + helpers; remaining methods added in subsequent tasks):

```go
package services

import (
    "context"
    "fmt"
    "time"

    "github.com/exnodes/hrm-api/internal/dto"
    apperr "github.com/exnodes/hrm-api/internal/errors"
    "github.com/exnodes/hrm-api/internal/models"
    "github.com/exnodes/hrm-api/internal/permissions"
    "github.com/exnodes/hrm-api/internal/repositories"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type LeaveService struct {
    db       *gorm.DB
    leaveRepo repositories.LeaveRequestRepository
    userRepo  repositories.UserRepository
    deptRepo  repositories.DepartmentRepository
    posRepo   repositories.PositionRepository
    roleSvc   *RoleService // for resolveUserPermissions
}

func NewLeaveService(
    db *gorm.DB,
    leaveRepo repositories.LeaveRequestRepository,
    userRepo repositories.UserRepository,
    deptRepo repositories.DepartmentRepository,
    posRepo repositories.PositionRepository,
    roleSvc *RoleService,
) *LeaveService {
    return &LeaveService{db: db, leaveRepo: leaveRepo, userRepo: userRepo, deptRepo: deptRepo, posRepo: posRepo, roleSvc: roleSvc}
}

func calculateTotalDays(from, to time.Time, period models.LeavePeriod) float64 {
    days := float64(int(to.Sub(from).Hours()/24)) + 1
    if period == models.LeavePeriodMorningHalf || period == models.LeavePeriodAfternoonHalf {
        return days * 0.5
    }
    return days
}

func (s *LeaveService) canManage(ctx context.Context, user *models.User) (bool, error) {
    perms, err := s.roleSvc.ResolveUserPermissions(ctx, user.ID)
    if err != nil {
        return false, err
    }
    if _, ok := perms[permissions.PermAll]; ok {
        return true, nil
    }
    _, ok := perms[permissions.PermLeaveManage]
    return ok, nil
}

func (s *LeaveService) computeBalance(ctx context.Context, user *models.User, year int) (dto.LeaveBalanceSummary, error) {
    grouped, err := s.leaveRepo.SumApprovedDays(ctx, user.ID, year)
    if err != nil {
        return dto.LeaveBalanceSummary{}, err
    }
    annual := grouped[models.LeaveTypeAnnual].Days
    sick := grouped[models.LeaveTypeSick].Days
    var totalCount int64
    for _, v := range grouped { totalCount += v.Count }
    return dto.LeaveBalanceSummary{
        Year:            year,
        AnnualQuota:     user.AnnualLeaveQuota,
        AnnualUsed:      annual,
        AnnualRemaining: user.AnnualLeaveQuota - annual,
        SickQuota:       user.SickLeaveQuota,
        SickUsed:        sick,
        SickRemaining:   user.SickLeaveQuota - sick,
        LeavesThisYear:  int(totalCount),
    }, nil
}

func remainingForType(b dto.LeaveBalanceSummary, t models.LeaveType) (float64, bool) {
    switch t {
    case models.LeaveTypeAnnual:
        return b.AnnualRemaining, true
    case models.LeaveTypeSick:
        return b.SickRemaining, true
    default:
        return 0, false
    }
}

func (s *LeaveService) populateRead(ctx context.Context, lr *models.LeaveRequest) (dto.LeaveRequestRead, error) {
    out := dto.LeaveRequestRead{
        ID: lr.ID.String(),
        FromDate: lr.FromDate, ToDate: lr.ToDate,
        LeavePeriod: lr.LeavePeriod, LeaveType: lr.LeaveType,
        TotalDays: lr.TotalDays, Reason: lr.Reason,
        AttachmentURL: lr.AttachmentURL, Status: lr.Status,
        CreatedBy: lr.CreatedBy.String(),
        CreatedAt: lr.CreatedAt, UpdatedAt: lr.UpdatedAt,
    }
    emp, err := s.userRepo.GetByID(ctx, lr.EmployeeID)
    if err == nil && emp != nil {
        out.Employee = &dto.LeaveRefRead{ID: emp.ID.String(), Name: fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)}
        if emp.DepartmentID != nil {
            if d, _ := s.deptRepo.GetByID(ctx, *emp.DepartmentID); d != nil {
                out.Department = &dto.LeaveRefRead{ID: d.ID.String(), Name: d.Name}
            }
        }
        if emp.PositionID != nil {
            if p, _ := s.posRepo.GetByID(ctx, *emp.PositionID); p != nil {
                out.Position = &dto.LeaveRefRead{ID: p.ID.String(), Name: p.Name}
            }
        }
    }
    return out, nil
}

func mustParseUUID(s string) (uuid.UUID, error) {
    id, err := uuid.Parse(s)
    if err != nil {
        return uuid.Nil, apperr.ErrBadRequest("invalid uuid: " + s)
    }
    return id, nil
}
```

- [ ] Run: `go build ./...`
  - Expected: no errors.
- [ ] **Commit:** `feat(phase-05): scaffold LeaveService + helpers`

### Task 6 — Service: Create

- [ ] Add to `internal/services/leave_service.go`:

```go
type CreateResult struct {
    Request  *models.LeaveRequest
    Warnings []string
}

func (s *LeaveService) Create(ctx context.Context, currentUser *models.User, in dto.LeaveRequestCreate) (*CreateResult, error) {
    if in.ToDate.Before(in.FromDate) {
        return nil, apperr.ErrBadRequest("To Date must be on or after From Date")
    }

    var employeeID uuid.UUID
    if in.EmployeeID != nil && *in.EmployeeID != "" {
        canMgr, err := s.canManage(ctx, currentUser)
        if err != nil { return nil, err }
        if !canMgr {
            return nil, apperr.ErrForbidden("You do not have permission to create leave requests on behalf of others")
        }
        eid, err := mustParseUUID(*in.EmployeeID); if err != nil { return nil, err }
        emp, err := s.userRepo.GetByID(ctx, eid)
        if err != nil || emp == nil { return nil, apperr.ErrBadRequest("Employee not found") }
        employeeID = emp.ID
    } else {
        employeeID = currentUser.ID
    }

    totalDays := calculateTotalDays(in.FromDate, in.ToDate, in.LeavePeriod)
    warnings := []string{}

    if _, isQuota := models.QuotaLeaveTypes()[in.LeaveType]; isQuota {
        emp, _ := s.userRepo.GetByID(ctx, employeeID)
        bal, err := s.computeBalance(ctx, emp, time.Now().UTC().Year())
        if err != nil { return nil, err }
        if rem, ok := remainingForType(bal, in.LeaveType); ok && totalDays > rem {
            warnings = append(warnings, fmt.Sprintf("Insufficient %s leave balance. %.1f days remaining, %.1f days requested.", in.LeaveType, rem, totalDays))
        }
    }

    overlaps, err := s.leaveRepo.Overlapping(ctx, employeeID, in.FromDate, in.ToDate, nil)
    if err != nil { return nil, err }
    if len(overlaps) > 0 {
        warnings = append(warnings, "This employee already has a leave request for overlapping dates.")
    }

    lr := &models.LeaveRequest{
        EmployeeID: employeeID,
        FromDate: in.FromDate, ToDate: in.ToDate,
        LeavePeriod: in.LeavePeriod, LeaveType: in.LeaveType,
        TotalDays: totalDays, Reason: in.Reason,
        Status: models.LeaveStatusPending, CreatedBy: currentUser.ID,
    }
    if err := s.leaveRepo.Create(ctx, lr); err != nil { return nil, err }
    return &CreateResult{Request: lr, Warnings: warnings}, nil
}
```

- [ ] **Commit:** `feat(phase-05): leave service Create with quota+overlap warnings`

### Task 7 — Service: Update / status transitions / delete

- [ ] Add to `internal/services/leave_service.go`:

```go
type UpdateResult = CreateResult

func (s *LeaveService) Update(ctx context.Context, currentUser *models.User, id uuid.UUID, in dto.LeaveRequestUpdate) (*UpdateResult, error) {
    lr, err := s.leaveRepo.GetByID(ctx, id)
    if err != nil { return nil, apperr.ErrNotFound("leave request") }

    canMgr, err := s.canManage(ctx, currentUser); if err != nil { return nil, err }
    if !canMgr && lr.EmployeeID != currentUser.ID {
        return nil, apperr.ErrForbidden("You can only edit your own leave requests")
    }
    if lr.Status != models.LeaveStatusPending && lr.Status != models.LeaveStatusApproved {
        return nil, apperr.ErrBadRequest(fmt.Sprintf("Cannot edit a %s leave request. Only Pending or Approved requests can be edited.", lr.Status))
    }

    from := lr.FromDate; to := lr.ToDate
    period := lr.LeavePeriod; ltype := lr.LeaveType
    if in.FromDate != nil { from = *in.FromDate }
    if in.ToDate != nil { to = *in.ToDate }
    if in.LeavePeriod != nil { period = *in.LeavePeriod }
    if in.LeaveType != nil { ltype = *in.LeaveType }
    if to.Before(from) { return nil, apperr.ErrBadRequest("To Date must be on or after From Date") }

    totalDays := calculateTotalDays(from, to, period)
    warnings := []string{}
    if _, isQuota := models.QuotaLeaveTypes()[ltype]; isQuota {
        emp, _ := s.userRepo.GetByID(ctx, lr.EmployeeID)
        bal, err := s.computeBalance(ctx, emp, time.Now().UTC().Year()); if err != nil { return nil, err }
        rem, _ := remainingForType(bal, ltype)
        if lr.Status == models.LeaveStatusApproved && lr.LeaveType == ltype {
            rem += lr.TotalDays // give back current request's days
        }
        if totalDays > rem {
            warnings = append(warnings, fmt.Sprintf("Insufficient %s leave balance. %.1f days remaining, %.1f days requested.", ltype, rem, totalDays))
        }
    }
    overlaps, err := s.leaveRepo.Overlapping(ctx, lr.EmployeeID, from, to, &lr.ID); if err != nil { return nil, err }
    if len(overlaps) > 0 {
        warnings = append(warnings, "This employee already has a leave request for overlapping dates.")
    }

    lr.FromDate = from; lr.ToDate = to
    lr.LeavePeriod = period; lr.LeaveType = ltype
    lr.TotalDays = totalDays
    if in.Reason != nil { lr.Reason = *in.Reason }
    if lr.Status == models.LeaveStatusApproved {
        lr.Status = models.LeaveStatusPending
    }
    if err := s.leaveRepo.Save(ctx, lr); err != nil { return nil, err }
    return &UpdateResult{Request: lr, Warnings: warnings}, nil
}

func (s *LeaveService) Approve(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error) {
    lr, err := s.leaveRepo.GetByID(ctx, id); if err != nil { return nil, apperr.ErrNotFound("leave request") }
    if lr.Status != models.LeaveStatusPending {
        return nil, apperr.ErrBadRequest("Only pending requests can be approved")
    }
    lr.Status = models.LeaveStatusApproved
    return lr, s.leaveRepo.Save(ctx, lr)
}

func (s *LeaveService) Reject(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error) {
    lr, err := s.leaveRepo.GetByID(ctx, id); if err != nil { return nil, apperr.ErrNotFound("leave request") }
    if lr.Status != models.LeaveStatusPending {
        return nil, apperr.ErrBadRequest("Only pending requests can be rejected")
    }
    lr.Status = models.LeaveStatusRejected
    return lr, s.leaveRepo.Save(ctx, lr)
}

func (s *LeaveService) Cancel(ctx context.Context, currentUser *models.User, id uuid.UUID) (*models.LeaveRequest, bool, error) {
    lr, err := s.leaveRepo.GetByID(ctx, id); if err != nil { return nil, false, apperr.ErrNotFound("leave request") }
    if lr.Status != models.LeaveStatusPending && lr.Status != models.LeaveStatusApproved {
        return nil, false, apperr.ErrBadRequest("Only pending or approved requests can be cancelled")
    }
    wasApproved := lr.Status == models.LeaveStatusApproved
    lr.Status = models.LeaveStatusCancelled
    return lr, wasApproved, s.leaveRepo.Save(ctx, lr)
}

func (s *LeaveService) Delete(ctx context.Context, currentUser *models.User, id uuid.UUID, isAdmin bool) error {
    lr, err := s.leaveRepo.GetByID(ctx, id); if err != nil { return apperr.ErrNotFound("leave request") }
    if !isAdmin {
        if lr.EmployeeID != currentUser.ID {
            return apperr.ErrForbidden("You can only delete your own leave requests")
        }
        if lr.Status != models.LeaveStatusPending {
            return apperr.ErrForbidden("You can only delete your own pending leave requests")
        }
    }
    return s.leaveRepo.SoftDelete(ctx, lr.ID)
}
```

- [ ] **Commit:** `feat(phase-05): leave service Update/Approve/Reject/Cancel/Delete`

### Task 8 — Service: Get / List / Balance / Dashboard / History

- [ ] Add to `internal/services/leave_service.go`:

```go
func (s *LeaveService) Get(ctx context.Context, id uuid.UUID) (*models.LeaveRequest, error) {
    lr, err := s.leaveRepo.GetByID(ctx, id)
    if err != nil { return nil, apperr.ErrNotFound("leave request") }
    return lr, nil
}

func (s *LeaveService) List(ctx context.Context, q dto.LeaveListQuery) ([]models.LeaveRequest, int64, error) {
    var employeeIDs []uuid.UUID // nil = no filter

    needsUserFilter := q.DepartmentID != "" || q.PositionID != "" || q.Search != ""
    if needsUserFilter {
        users, err := s.userRepo.SearchForLeaveFilter(ctx, q.Search, q.DepartmentID, q.PositionID)
        if err != nil { return nil, 0, err }
        ids := make([]uuid.UUID, 0, len(users))
        for _, u := range users { ids = append(ids, u.ID) }
        employeeIDs = ids // empty slice short-circuits in repo to (nil, 0)
    }

    if q.Page < 1 { q.Page = 1 }
    if q.PageSize < 1 { q.PageSize = 10 }
    return s.leaveRepo.List(ctx, employeeIDs, q.Status, q.Page, q.PageSize)
}

func (s *LeaveService) GetBalance(ctx context.Context, employeeID uuid.UUID, year *int) (dto.LeaveBalanceSummary, error) {
    user, err := s.userRepo.GetByID(ctx, employeeID)
    if err != nil || user == nil { return dto.LeaveBalanceSummary{}, apperr.ErrNotFound("employee") }
    y := time.Now().UTC().Year()
    if year != nil { y = *year }
    return s.computeBalance(ctx, user, y)
}

func (s *LeaveService) GetMyDashboard(ctx context.Context, user *models.User, limit int) (*dto.LeaveDashboardRead, error) {
    today := time.Now().UTC().Truncate(24 * time.Hour)
    bal, err := s.computeBalance(ctx, user, today.Year()); if err != nil { return nil, err }
    upDocs, err := s.leaveRepo.Upcoming(ctx, user.ID, today, limit); if err != nil { return nil, err }
    histDocs, err := s.leaveRepo.History(ctx, user.ID, today, limit); if err != nil { return nil, err }
    out := &dto.LeaveDashboardRead{Balance: bal, Upcoming: []dto.LeaveRequestRead{}, History: []dto.LeaveRequestRead{}}
    for i := range upDocs   { r, _ := s.populateRead(ctx, &upDocs[i]);   out.Upcoming = append(out.Upcoming, r) }
    for i := range histDocs { r, _ := s.populateRead(ctx, &histDocs[i]); out.History  = append(out.History,  r) }
    return out, nil
}

func (s *LeaveService) ListMyHistory(ctx context.Context, user *models.User, q dto.LeaveHistoryQuery) ([]models.LeaveRequest, int64, error) {
    if q.Page < 1 { q.Page = 1 }
    if q.PageSize < 1 { q.PageSize = 10 }
    return s.leaveRepo.ListByUser(ctx, user.ID, repositories.ListByUserFilter{
        Statuses: q.Status, StartDate: q.StartDate, EndDate: q.EndDate,
        Page: q.Page, PageSize: q.PageSize,
    })
}
```

- [ ] Add `SearchForLeaveFilter(ctx, search, deptID, posID)` to `UserRepository` (already created in Phase 2; extend its interface and GORM impl):

```go
// internal/repositories/user_repo.go (extend)
func (r *userRepo) SearchForLeaveFilter(ctx context.Context, search, deptID, posID string) ([]models.User, error) {
    q := r.db.WithContext(ctx).Model(&models.User{}).Scopes(models.NotDeleted)
    if deptID != "" { q = q.Where("department_id = ?", deptID) }
    if posID  != "" { q = q.Where("position_id = ?", posID) }
    if search != "" {
        esc := utils.EscapeLike(search)
        like := "%" + esc + "%"
        q = q.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone_number ILIKE ?", like, like, like, like)
    }
    var users []models.User
    err := q.Find(&users).Error
    return users, err
}
```

- [ ] Run: `go build ./...`
  - Expected: no errors.
- [ ] **Commit:** `feat(phase-05): leave service List/Balance/Dashboard/History + user filter helper`

### Task 9 — Handler skeleton + routes

- [ ] Create `internal/handlers/leave_handler.go`:

```go
package handlers

import (
    "encoding/json"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/exnodes/hrm-api/internal/dto"
    apperr "github.com/exnodes/hrm-api/internal/errors"
    "github.com/exnodes/hrm-api/internal/middleware"
    "github.com/exnodes/hrm-api/internal/permissions"
    "github.com/exnodes/hrm-api/internal/services"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type LeaveHandler struct {
    svc     *services.LeaveService
    roleSvc *services.RoleService
    upload  *services.UploadService // saveLeaveAttachment / deleteFile (Phase 0)
}

func NewLeaveHandler(svc *services.LeaveService, roleSvc *services.RoleService, upload *services.UploadService) *LeaveHandler {
    return &LeaveHandler{svc: svc, roleSvc: roleSvc, upload: upload}
}

func (h *LeaveHandler) Register(rg *gin.RouterGroup) {
    g := rg.Group("/leave-requests")
    g.Use(middleware.JWT())

    g.GET("",                                  middleware.RequirePerms(permissions.PermLeaveRead),    h.List)
    g.GET("/balance/:employee_id",             middleware.RequirePerms(permissions.PermLeaveRead),    h.GetBalance)
    g.GET("/dashboard/me",                                                                            h.GetMyDashboard)
    g.GET("/history/me",                                                                              h.GetMyHistory)
    g.POST("",                                 middleware.RequirePerms(permissions.PermLeaveCreate),  h.Create)
    g.GET("/:id",                              middleware.RequirePerms(permissions.PermLeaveRead),    h.Get)
    g.PATCH("/:id",                            middleware.RequirePerms(permissions.PermLeaveUpdate),  h.Update)
    g.POST("/:id/approve",                     middleware.RequirePerms(permissions.PermLeaveApprove), h.Approve)
    g.POST("/:id/reject",                      middleware.RequirePerms(permissions.PermLeaveApprove), h.Reject)
    g.POST("/:id/cancel",                      middleware.RequirePerms(permissions.PermLeaveCancel),  h.Cancel)
    g.POST("/:id/delete",                      middleware.RequirePerms(permissions.PermLeaveDelete),  h.Delete)
}
```

- [ ] **Commit:** `feat(phase-05): leave handler skeleton + routes`

### Task 10 — Handler: List + Get + Balance

- [ ] Add to `internal/handlers/leave_handler.go`:

```go
// List godoc
// @Summary  List leave requests
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    page query int false "Page" default(1)
// @Param    page_size query int false "Page size" default(10)
// @Param    search query string false "Search by employee name/email/phone"
// @Param    status query []string false "Status filter (repeatable)" collectionFormat(multi)
// @Param    department_id query string false "Filter by department"
// @Param    position_id query string false "Filter by position"
// @Success  200 {object} dto.Response[dto.PaginatedData[dto.LeaveRequestRead]]
// @Router   /api/v1/leave-requests [get]
func (h *LeaveHandler) List(c *gin.Context) {
    var q dto.LeaveListQuery
    if err := c.ShouldBindQuery(&q); err != nil { _ = c.Error(apperr.ErrBadRequest(err.Error())); return }
    items, total, err := h.svc.List(c.Request.Context(), q)
    if err != nil { _ = c.Error(err); return }
    out := make([]dto.LeaveRequestRead, 0, len(items))
    for i := range items {
        r, _ := h.svc.PopulateRead(c.Request.Context(), &items[i]) // expose populateRead via wrapper
        out = append(out, r)
    }
    pages := 0
    if total > 0 { pages = int((total + int64(q.PageSize) - 1) / int64(q.PageSize)) }
    c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.LeaveRequestRead]]{
        Success: true,
        Data: dto.PaginatedData[dto.LeaveRequestRead]{Items: out, Total: total, Page: q.Page, PageSize: q.PageSize, TotalPages: pages},
    })
}

// Get godoc
// @Summary  Get leave request
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    id path string true "Leave request id"
// @Success  200 {object} dto.Response[dto.LeaveRequestRead]
// @Router   /api/v1/leave-requests/{id} [get]
func (h *LeaveHandler) Get(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid id")); return }
    lr, err := h.svc.Get(c.Request.Context(), id); if err != nil { _ = c.Error(err); return }
    r, _ := h.svc.PopulateRead(c.Request.Context(), lr)
    c.JSON(http.StatusOK, dto.Response[dto.LeaveRequestRead]{Success: true, Data: r})
}

// GetBalance godoc
// @Summary  Per-type leave balance
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    employee_id path string true "Employee id"
// @Param    year query int false "Calendar year (defaults to current)"
// @Success  200 {object} dto.Response[dto.LeaveBalanceSummary]
// @Router   /api/v1/leave-requests/balance/{employee_id} [get]
func (h *LeaveHandler) GetBalance(c *gin.Context) {
    eid, err := uuid.Parse(c.Param("employee_id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid employee_id")); return }
    var yp *int
    if y := c.Query("year"); y != "" {
        yi, err := strconv.Atoi(y); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid year")); return }
        yp = &yi
    }
    bal, err := h.svc.GetBalance(c.Request.Context(), eid, yp); if err != nil { _ = c.Error(err); return }
    c.JSON(http.StatusOK, dto.Response[dto.LeaveBalanceSummary]{Success: true, Data: bal})
}
```

- [ ] Expose `PopulateRead` on `LeaveService` (rename `populateRead` to `PopulateRead`):

```go
func (s *LeaveService) PopulateRead(ctx context.Context, lr *models.LeaveRequest) (dto.LeaveRequestRead, error) { ... }
```

- [ ] **Commit:** `feat(phase-05): leave handler List/Get/Balance`

### Task 11 — Handler: Dashboard/me + History/me

- [ ] Add:

```go
// GetMyDashboard godoc
// @Summary  Mobile leave dashboard for the current user
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    limit query int false "Max per tab" default(10)
// @Success  200 {object} dto.Response[dto.LeaveDashboardRead]
// @Router   /api/v1/leave-requests/dashboard/me [get]
func (h *LeaveHandler) GetMyDashboard(c *gin.Context) {
    user := middleware.MustCurrentUser(c)
    limit := 10
    if v := c.Query("limit"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 50 { limit = n }
    }
    d, err := h.svc.GetMyDashboard(c.Request.Context(), user, limit); if err != nil { _ = c.Error(err); return }
    c.JSON(http.StatusOK, dto.Response[*dto.LeaveDashboardRead]{Success: true, Data: d})
}

// GetMyHistory godoc
// @Summary  Paginated leave history for the current user
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    page query int false "Page" default(1)
// @Param    page_size query int false "Page size" default(10)
// @Param    status query []string false "Status filter (repeatable)" collectionFormat(multi)
// @Param    start_date query string false "From-date lower bound YYYY-MM-DD"
// @Param    end_date   query string false "To-date upper bound YYYY-MM-DD"
// @Success  200 {object} dto.Response[dto.PaginatedData[dto.LeaveRequestRead]]
// @Router   /api/v1/leave-requests/history/me [get]
func (h *LeaveHandler) GetMyHistory(c *gin.Context) {
    user := middleware.MustCurrentUser(c)
    var q dto.LeaveHistoryQuery
    if err := c.ShouldBindQuery(&q); err != nil { _ = c.Error(apperr.ErrBadRequest(err.Error())); return }
    items, total, err := h.svc.ListMyHistory(c.Request.Context(), user, q); if err != nil { _ = c.Error(err); return }
    out := make([]dto.LeaveRequestRead, 0, len(items))
    for i := range items { r, _ := h.svc.PopulateRead(c.Request.Context(), &items[i]); out = append(out, r) }
    pages := 0
    if total > 0 { pages = int((total + int64(q.PageSize) - 1) / int64(q.PageSize)) }
    c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.LeaveRequestRead]]{
        Success: true,
        Data: dto.PaginatedData[dto.LeaveRequestRead]{Items: out, Total: total, Page: q.Page, PageSize: q.PageSize, TotalPages: pages},
    })
}
```

- [ ] **Commit:** `feat(phase-05): leave handler Dashboard/History`

### Task 12 — Handler: Create + Update (multipart form)

- [ ] Add helper `parseMultipartLeave[T any](c *gin.Context) (*T, *multipart.FileHeader, error)` and Create/Update handlers:

```go
// Create godoc
// @Summary  Create a leave request (multipart: data JSON + optional attachment)
// @Tags     Leave Requests
// @Security BearerAuth
// @Accept   mpfd
// @Param    data formData string true "JSON string of LeaveRequestCreate"
// @Param    attachment formData file false "Optional attachment"
// @Success  201 {object} dto.Response[dto.LeaveRequestRead]
// @Router   /api/v1/leave-requests [post]
func (h *LeaveHandler) Create(c *gin.Context) {
    raw := c.PostForm("data")
    if raw == "" { _ = c.Error(apperr.ErrBadRequest("data field is required")); return }
    var in dto.LeaveRequestCreate
    if err := json.Unmarshal([]byte(raw), &in); err != nil {
        _ = c.Error(apperr.ErrBadRequest("invalid data JSON: " + err.Error())); return
    }
    currentUser := middleware.MustCurrentUser(c)
    result, err := h.svc.Create(c.Request.Context(), currentUser, in); if err != nil { _ = c.Error(err); return }

    if fh, _ := c.FormFile("attachment"); fh != nil {
        url, uerr := h.upload.SaveLeaveAttachment(c.Request.Context(), fh)
        if uerr != nil { _ = c.Error(uerr); return }
        result.Request.AttachmentURL = &url
        if err := h.svc.SetAttachment(c.Request.Context(), result.Request.ID, url); err != nil { _ = c.Error(err); return }
    }

    r, _ := h.svc.PopulateRead(c.Request.Context(), result.Request)
    msg := "Leave request has been submitted"
    if len(result.Warnings) > 0 { msg += ". Warning: " + strings.Join(result.Warnings, " | ") }
    c.JSON(http.StatusCreated, dto.Response[dto.LeaveRequestRead]{Success: true, Message: msg, Data: r})
}

// Update godoc
// @Summary  Update a leave request (multipart: data JSON + optional attachment)
// @Tags     Leave Requests
// @Security BearerAuth
// @Accept   mpfd
// @Param    id path string true "Leave request id"
// @Param    data formData string true "JSON string of LeaveRequestUpdate"
// @Param    attachment formData file false "Optional attachment"
// @Success  200 {object} dto.Response[dto.LeaveRequestRead]
// @Router   /api/v1/leave-requests/{id} [patch]
func (h *LeaveHandler) Update(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid id")); return }
    raw := c.PostForm("data"); if raw == "" { _ = c.Error(apperr.ErrBadRequest("data field is required")); return }
    var in dto.LeaveRequestUpdate
    if err := json.Unmarshal([]byte(raw), &in); err != nil {
        _ = c.Error(apperr.ErrBadRequest("invalid data JSON: " + err.Error())); return
    }
    currentUser := middleware.MustCurrentUser(c)
    result, err := h.svc.Update(c.Request.Context(), currentUser, id, in); if err != nil { _ = c.Error(err); return }

    if fh, _ := c.FormFile("attachment"); fh != nil {
        if result.Request.AttachmentURL != nil { _ = h.upload.DeleteFile(c.Request.Context(), *result.Request.AttachmentURL) }
        url, uerr := h.upload.SaveLeaveAttachment(c.Request.Context(), fh); if uerr != nil { _ = c.Error(uerr); return }
        if err := h.svc.SetAttachment(c.Request.Context(), result.Request.ID, url); err != nil { _ = c.Error(err); return }
        result.Request.AttachmentURL = &url
    }
    r, _ := h.svc.PopulateRead(c.Request.Context(), result.Request)
    msg := "Leave request has been updated"
    if result.Request.Status == "pending" && len(result.Warnings) > 0 { msg += ". Warning: " + strings.Join(result.Warnings, " | ") }
    c.JSON(http.StatusOK, dto.Response[dto.LeaveRequestRead]{Success: true, Message: msg, Data: r})
}

// guard unused imports in case io/time aren't referenced elsewhere
var _ io.Reader
var _ time.Time
```

- [ ] Add `SetAttachment(ctx, id, url)` to service to persist the attachment URL after the file upload:

```go
func (s *LeaveService) SetAttachment(ctx context.Context, id uuid.UUID, url string) error {
    lr, err := s.leaveRepo.GetByID(ctx, id); if err != nil { return err }
    lr.AttachmentURL = &url
    return s.leaveRepo.Save(ctx, lr)
}
```

- [ ] **Commit:** `feat(phase-05): leave handler Create+Update multipart`

### Task 13 — Handler: Approve / Reject / Cancel / Delete

- [ ] Add:

```go
// Approve godoc
// @Summary  Approve a pending leave request
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    id path string true "Leave request id"
// @Success  200 {object} dto.Response[dto.LeaveRequestRead]
// @Router   /api/v1/leave-requests/{id}/approve [post]
func (h *LeaveHandler) Approve(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid id")); return }
    lr, err := h.svc.Approve(c.Request.Context(), id); if err != nil { _ = c.Error(err); return }
    r, _ := h.svc.PopulateRead(c.Request.Context(), lr)
    c.JSON(http.StatusOK, dto.Response[dto.LeaveRequestRead]{Success: true, Message: "Leave request has been approved", Data: r})
}

// Reject godoc
// @Summary  Reject a pending leave request
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    id path string true "Leave request id"
// @Success  200 {object} dto.Response[dto.LeaveRequestRead]
// @Router   /api/v1/leave-requests/{id}/reject [post]
func (h *LeaveHandler) Reject(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid id")); return }
    lr, err := h.svc.Reject(c.Request.Context(), id); if err != nil { _ = c.Error(err); return }
    r, _ := h.svc.PopulateRead(c.Request.Context(), lr)
    c.JSON(http.StatusOK, dto.Response[dto.LeaveRequestRead]{Success: true, Message: "Leave request has been rejected", Data: r})
}

// Cancel godoc
// @Summary  Cancel a pending or approved leave request
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    id path string true "Leave request id"
// @Success  200 {object} dto.Response[dto.LeaveRequestRead]
// @Router   /api/v1/leave-requests/{id}/cancel [post]
func (h *LeaveHandler) Cancel(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid id")); return }
    currentUser := middleware.MustCurrentUser(c)
    lr, wasApproved, err := h.svc.Cancel(c.Request.Context(), currentUser, id); if err != nil { _ = c.Error(err); return }
    r, _ := h.svc.PopulateRead(c.Request.Context(), lr)
    msg := "Leave request has been cancelled"
    if wasApproved { msg += ". Leave balance restored." }
    c.JSON(http.StatusOK, dto.Response[dto.LeaveRequestRead]{Success: true, Message: msg, Data: r})
}

// Delete godoc
// @Summary  Soft-delete a leave request
// @Tags     Leave Requests
// @Security BearerAuth
// @Param    id path string true "Leave request id"
// @Success  200 {object} dto.Response[any]
// @Router   /api/v1/leave-requests/{id}/delete [post]
func (h *LeaveHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id")); if err != nil { _ = c.Error(apperr.ErrBadRequest("invalid id")); return }
    currentUser := middleware.MustCurrentUser(c)
    perms, _ := h.roleSvc.ResolveUserPermissions(c.Request.Context(), currentUser.ID)
    _, allWildcard := perms[permissions.PermAll]
    _, hasApprove := perms[permissions.PermLeaveApprove]
    isAdmin := allWildcard || hasApprove
    if err := h.svc.Delete(c.Request.Context(), currentUser, id, isAdmin); err != nil { _ = c.Error(err); return }
    c.JSON(http.StatusOK, dto.Response[any]{Success: true, Message: "Leave request has been deleted"})
}
```

- [ ] Run: `go build ./...`
  - Expected: no errors.
- [ ] **Commit:** `feat(phase-05): leave handler Approve/Reject/Cancel/Delete`

### Task 14 — Wire into main.go

- [ ] Edit `cmd/server/main.go` to instantiate the repo, service, handler and register the routes group inside the existing `v1 := api.Group("/api/v1")` block:

```go
leaveRepo  := repositories.NewLeaveRequestRepository(db)
leaveSvc   := services.NewLeaveService(db, leaveRepo, userRepo, deptRepo, posRepo, roleSvc)
leaveH     := handlers.NewLeaveHandler(leaveSvc, roleSvc, uploadSvc)
leaveH.Register(v1)
```

- [ ] Run: `make run` (background) then `curl -s localhost:8080/health`
  - Expected: `{"success":true,"data":{"status":"ok"}}` and Swagger lists 10 new leave routes.
- [ ] Run: `curl -s localhost:8080/swagger/doc.json | jq '.paths | keys[] | select(contains("leave-requests"))'`
  - Expected: 10 paths printed.
- [ ] **Commit:** `feat(phase-05): wire leave module into server`

### Task 15 — Run swag init

- [ ] Run: `make swag`
  - Expected: regenerates `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`. No swag warnings about missing fields.
- [ ] Run: `git status --short docs/`
  - Expected: only `docs/swagger.{json,yaml}` and `docs/docs.go` modified.
- [ ] **Commit:** `chore(phase-05): regenerate swagger docs`

### Task 16 — Service tests: scaffolding + create happy path

- [ ] Create `internal/services/leave_service_test.go`:

```go
package services_test

import (
    "context"
    "testing"
    "time"

    "github.com/exnodes/hrm-api/internal/dto"
    "github.com/exnodes/hrm-api/internal/models"
    "github.com/exnodes/hrm-api/internal/services"
    "github.com/stretchr/testify/require"
)

func TestLeave_Create_Annual_SufficientBalance(t *testing.T) {
    h := newTestHarness(t) // sets up DB, repos, services from testhelper_test.go (Phase 1)
    defer h.Cleanup(t)

    emp := h.MakeUser(t, "alice@example.com", withAnnualQuota(12), withSickQuota(6))
    ctx := context.Background()

    in := dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 7),
        ToDate:   today().AddDate(0, 0, 9),
        LeavePeriod: models.LeavePeriodFullDay,
        LeaveType:   models.LeaveTypeAnnual,
        Reason:      "vacation",
    }
    res, err := h.LeaveSvc.Create(ctx, emp, in)
    require.NoError(t, err)
    require.Equal(t, 3.0, res.Request.TotalDays)
    require.Empty(t, res.Warnings)
    require.Equal(t, models.LeaveStatusPending, res.Request.Status)
}

func today() time.Time { return time.Now().UTC().Truncate(24 * time.Hour) }
```

- [ ] Run: `go test ./internal/services/... -run TestLeave_Create_Annual_SufficientBalance`
  - Expected: `PASS`.
- [ ] **Commit:** `test(phase-05): leave Create happy path`

### Task 17 — Service tests: warnings (insufficient quota, overlap, half-day)

- [ ] Append to `leave_service_test.go`:

```go
func TestLeave_Create_InsufficientQuota_Warns(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "bob@example.com", withAnnualQuota(2))
    ctx := context.Background()
    in := dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 7), ToDate: today().AddDate(0, 0, 12),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual,
        Reason: "trip",
    }
    res, err := h.LeaveSvc.Create(ctx, emp, in)
    require.NoError(t, err)
    require.NotEmpty(t, res.Warnings)
    require.Contains(t, res.Warnings[0], "Insufficient annual leave balance")
    require.Equal(t, models.LeaveStatusPending, res.Request.Status) // still created
}

func TestLeave_Create_HalfDay(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "carol@example.com", withAnnualQuota(12))
    ctx := context.Background()
    d := today().AddDate(0, 0, 7)
    in := dto.LeaveRequestCreate{
        FromDate: d, ToDate: d, LeavePeriod: models.LeavePeriodMorningHalf,
        LeaveType: models.LeaveTypeAnnual, Reason: "doctor",
    }
    res, err := h.LeaveSvc.Create(ctx, emp, in)
    require.NoError(t, err)
    require.Equal(t, 0.5, res.Request.TotalDays)
}

func TestLeave_Create_OverlapWarns(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "dan@example.com", withAnnualQuota(12))
    ctx := context.Background()
    base := today().AddDate(0, 0, 10)
    _, err := h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: base, ToDate: base.AddDate(0, 0, 2),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "a",
    })
    require.NoError(t, err)
    res, err := h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: base.AddDate(0, 0, 1), ToDate: base.AddDate(0, 0, 3),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "b",
    })
    require.NoError(t, err)
    require.NotEmpty(t, res.Warnings)
    require.Contains(t, res.Warnings[len(res.Warnings)-1], "overlapping")
}
```

- [ ] Run: `go test ./internal/services/... -run TestLeave_Create_`
  - Expected: all `PASS`.
- [ ] **Commit:** `test(phase-05): leave create warnings`

### Task 18 — Service tests: status transitions + ownership + delete rules

- [ ] Append:

```go
func TestLeave_ApproveThenCancel_RestoresBalance(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "eve@example.com", withAnnualQuota(12))
    admin := h.MakeUser(t, "admin@example.com", withRole("Admin"))
    ctx := context.Background()
    res, _ := h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 5), ToDate: today().AddDate(0, 0, 7),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "x",
    })
    lr, err := h.LeaveSvc.Approve(ctx, res.Request.ID); require.NoError(t, err)
    require.Equal(t, models.LeaveStatusApproved, lr.Status)

    bal, _ := h.LeaveSvc.GetBalance(ctx, emp.ID, nil)
    require.Equal(t, 3.0, bal.AnnualUsed)

    lr2, was, err := h.LeaveSvc.Cancel(ctx, admin, lr.ID); require.NoError(t, err)
    require.True(t, was)
    require.Equal(t, models.LeaveStatusCancelled, lr2.Status)

    bal2, _ := h.LeaveSvc.GetBalance(ctx, emp.ID, nil)
    require.Equal(t, 0.0, bal2.AnnualUsed) // restored — derived from approved+not-cancelled
}

func TestLeave_RejectThenApprove_Forbidden(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "f@example.com", withAnnualQuota(12))
    ctx := context.Background()
    res, _ := h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 5), ToDate: today().AddDate(0, 0, 5),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "y",
    })
    _, err := h.LeaveSvc.Reject(ctx, res.Request.ID); require.NoError(t, err)
    _, err = h.LeaveSvc.Approve(ctx, res.Request.ID)
    require.Error(t, err)
    require.Contains(t, err.Error(), "Only pending requests can be approved")
}

func TestLeave_Update_NonOwner_Forbidden(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    owner := h.MakeUser(t, "owner@example.com", withAnnualQuota(12))
    intruder := h.MakeUser(t, "intruder@example.com", withAnnualQuota(12)) // no manage perm
    ctx := context.Background()
    res, _ := h.LeaveSvc.Create(ctx, owner, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 5), ToDate: today().AddDate(0, 0, 5),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "z",
    })
    nr := "edited"
    _, err := h.LeaveSvc.Update(ctx, intruder, res.Request.ID, dto.LeaveRequestUpdate{Reason: &nr})
    require.Error(t, err)
    require.Contains(t, err.Error(), "You can only edit your own leave requests")
}

func TestLeave_Delete_NonOwner_Forbidden_NonPending(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    owner := h.MakeUser(t, "g@example.com", withAnnualQuota(12))
    ctx := context.Background()
    res, _ := h.LeaveSvc.Create(ctx, owner, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 5), ToDate: today().AddDate(0, 0, 5),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "k",
    })
    _, _ = h.LeaveSvc.Approve(context.Background(), res.Request.ID)
    err := h.LeaveSvc.Delete(ctx, owner, res.Request.ID, false /* not admin */)
    require.Error(t, err)
    require.Contains(t, err.Error(), "pending leave requests")
}
```

- [ ] Run: `go test ./internal/services/... -run TestLeave_`
  - Expected: all `PASS`.
- [ ] **Commit:** `test(phase-05): leave transitions + ownership + delete`

### Task 19 — Service tests: List + History + Dashboard

- [ ] Append:

```go
func TestLeave_List_FilterByStatus(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "h@example.com", withAnnualQuota(12))
    ctx := context.Background()
    a, _ := h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 5), ToDate: today().AddDate(0, 0, 5),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "a",
    })
    _, _ = h.LeaveSvc.Approve(ctx, a.Request.ID)
    _, _ = h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 10), ToDate: today().AddDate(0, 0, 10),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "b",
    })
    items, total, err := h.LeaveSvc.List(ctx, dto.LeaveListQuery{Page: 1, PageSize: 10, Status: []string{"pending"}})
    require.NoError(t, err)
    require.Equal(t, int64(1), total)
    require.Len(t, items, 1)
    require.Equal(t, models.LeaveStatusPending, items[0].Status)
}

func TestLeave_GetMyDashboard_SeparatesUpcomingHistory(t *testing.T) {
    h := newTestHarness(t); defer h.Cleanup(t)
    emp := h.MakeUser(t, "i@example.com", withAnnualQuota(12))
    ctx := context.Background()
    // Past
    _, _ = h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, -10), ToDate: today().AddDate(0, 0, -9),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "past",
    })
    // Future
    _, _ = h.LeaveSvc.Create(ctx, emp, dto.LeaveRequestCreate{
        FromDate: today().AddDate(0, 0, 5), ToDate: today().AddDate(0, 0, 6),
        LeavePeriod: models.LeavePeriodFullDay, LeaveType: models.LeaveTypeAnnual, Reason: "future",
    })
    d, err := h.LeaveSvc.GetMyDashboard(ctx, emp, 10)
    require.NoError(t, err)
    require.Len(t, d.Upcoming, 1)
    require.Len(t, d.History, 1)
    require.Equal(t, "future", d.Upcoming[0].Reason)
    require.Equal(t, "past", d.History[0].Reason)
}
```

- [ ] Run: `go test ./internal/services/... -run TestLeave_ -count=1 -v`
  - Expected: all `PASS`. No data leaks between tests (testhelper truncates).
- [ ] **Commit:** `test(phase-05): leave list + dashboard`

### Task 20 — End-to-end verification log

- [ ] Boot the server in another shell: `make run` (background).
- [ ] Capture curl session into `docs/superpowers/verification/phase-05.md`. Replace `$TOKEN` with the value from a fresh `/auth/login`. Replace `$LID` with the id returned by Create.

```bash
# 1. Login as a regular employee (created in Phase 2 seed)
curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"employee@example.com","password":"Passw0rd!"}' | tee /tmp/login.json
TOKEN=$(jq -r '.data.access_token' /tmp/login.json)

# 2. Create a leave request (multipart)
curl -s -X POST localhost:8080/api/v1/leave-requests \
  -H "Authorization: Bearer $TOKEN" \
  -F 'data={"from_date":"2026-06-01","to_date":"2026-06-03","leave_period":"full_day","leave_type":"annual","reason":"Trip"}' \
  | tee /tmp/lr.json
LID=$(jq -r '.data.id' /tmp/lr.json)

# 3. Confirm DB row (quota NOT deducted — annual_leave_quota is on users; balance is a live aggregate)
psql "$DATABASE_URL" -c "SELECT id,status,total_days,leave_type FROM leave_requests WHERE id='$LID';"

# 4. Dashboard /me — request should appear in upcoming
curl -s localhost:8080/api/v1/leave-requests/dashboard/me \
  -H "Authorization: Bearer $TOKEN" | jq '.data.upcoming[].id, .data.balance'

# 5. Owner update reason
curl -s -X PATCH localhost:8080/api/v1/leave-requests/$LID \
  -H "Authorization: Bearer $TOKEN" \
  -F 'data={"reason":"Trip with family"}' | jq '.data.reason'

# 6. Approve as admin
curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"Adm1n!"}' > /tmp/admin.json
ATOK=$(jq -r '.data.access_token' /tmp/admin.json)
curl -s -X POST localhost:8080/api/v1/leave-requests/$LID/approve \
  -H "Authorization: Bearer $ATOK" | jq '.data.status'   # expect "approved"

# 7. Verify balance reflects deduction (annual_used = 3)
curl -s "localhost:8080/api/v1/leave-requests/balance/$EMP?year=2026" \
  -H "Authorization: Bearer $ATOK" | jq '.data'

# 8. List with status filter
curl -s "localhost:8080/api/v1/leave-requests?status=approved&page=1&page_size=10" \
  -H "Authorization: Bearer $ATOK" | jq '.data.total, .data.items[].status'

# 9. Cancel approved as owner (Cancel perm not required for own request? — Python requires Permission.LEAVE_REQUESTS_CANCEL.
#    Self-cancel needs that perm on the employee role; if not granted, expect 403. Use admin to demonstrate cancel restoring balance.)
curl -s -X POST localhost:8080/api/v1/leave-requests/$LID/cancel \
  -H "Authorization: Bearer $ATOK" | jq '.message'   # expect "...balance restored."
curl -s "localhost:8080/api/v1/leave-requests/balance/$EMP?year=2026" \
  -H "Authorization: Bearer $ATOK" | jq '.data.annual_used' # expect 0

# 10. Reject case
curl -s -X POST localhost:8080/api/v1/leave-requests \
  -H "Authorization: Bearer $TOKEN" \
  -F 'data={"from_date":"2026-07-01","to_date":"2026-07-01","leave_period":"full_day","leave_type":"sick","reason":"flu"}' > /tmp/lr2.json
LID2=$(jq -r '.data.id' /tmp/lr2.json)
curl -s -X POST localhost:8080/api/v1/leave-requests/$LID2/reject \
  -H "Authorization: Bearer $ATOK" | jq '.data.status'   # expect "rejected"

# --- Error cases ---

# 11. Approve already-approved → 400
curl -s -X POST localhost:8080/api/v1/leave-requests/$LID2/approve \
  -H "Authorization: Bearer $ATOK" | jq '.success, .message'  # false, "Only pending..."

# 12. Edit by non-owner → 403
curl -s -X POST localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"other@example.com","password":"Passw0rd!"}' > /tmp/other.json
OTOK=$(jq -r '.data.access_token' /tmp/other.json)
curl -s -X PATCH localhost:8080/api/v1/leave-requests/$LID2 \
  -H "Authorization: Bearer $OTOK" \
  -F 'data={"reason":"hack"}' | jq '.success, .message'  # false, "You can only edit your own..."

# 13. Create with employee_id by non-admin → 403
curl -s -X POST localhost:8080/api/v1/leave-requests \
  -H "Authorization: Bearer $TOKEN" \
  -F "data={\"employee_id\":\"$OTHER_ID\",\"from_date\":\"2026-08-01\",\"to_date\":\"2026-08-01\",\"leave_period\":\"full_day\",\"leave_type\":\"annual\",\"reason\":\"x\"}" \
  | jq '.success, .message'  # false, "You do not have permission..."

# 14. Final DB state spot-check
psql "$DATABASE_URL" -c "SELECT status, COUNT(*) FROM leave_requests WHERE is_deleted = false GROUP BY status;"
```

- [ ] Each curl response field listed under `# expect` must match. Paste the actual JSON response under each step into the verification file.
- [ ] **Commit:** `docs(phase-05): end-to-end verification log`

### Task 21 — README + final self-check

- [ ] Update `README.md` "Endpoints" section to list the 10 new leave routes.
- [ ] Run all checks in one go:
  - `make migrate-down && make migrate-up`
  - `go vet ./...`
  - `go test ./internal/services/... -count=1`
  - `make run` then `curl -s localhost:8080/swagger/doc.json | jq '.paths | keys | length'` — confirm count increased by exactly 10 vs. Phase 4 baseline.
- [ ] Confirm no `*.up.sql` or `*.down.sql` file from earlier phases was modified: `git log --diff-filter=M --name-only --pretty=format: -- migrations/ | sort -u`
- [ ] **Commit:** `docs(phase-05): README endpoint list`

---

## 4. Decisions log (for reviewers)

| Decision | Rationale |
|---|---|
| No `leave_quotas` table — quota fields stay on `users` | Python source stores `annual_leave_quota` + `sick_leave_quota` on the user; "balance" is a derived aggregate over approved, non-deleted requests. Adding a separate table would diverge from the Python contract and require dual-write for cancel/reject "restore" semantics. Cancel/reject "restore" is implicit because the aggregate excludes those statuses. |
| No `leave_types` lookup table | Python uses a fixed `StrEnum` of 5 values. A lookup table buys nothing for a fixed enum and would require a join on every read. We use `TEXT + CHECK` for the same safety with no join cost. |
| Quota mutation endpoint stays on `PATCH /users/:id` (Phase 2) | No `/leave-quotas/:userID` exists in Python. `PermLeaveQuotaManage` constant is reserved for future use. |
| Action verbs use `POST` not `PATCH` | Matches Python router exactly (`POST /{id}/approve`, etc.). FE contract preserved. |
| `total_days` numeric(5,1) | Half-days are 0.5 increments; one decimal is sufficient and avoids float drift. |
| Warnings non-blocking | Python returns warnings inside the success response. Behavioral parity is more important than enforcing harder constraints; future tightening can happen in a follow-up phase. |

## 5. Definition of Done

- [ ] All 21 tasks committed.
- [ ] `make migrate-up` clean. `make migrate-down` returns to Phase 4 schema (verified by `make migrate-version`).
- [ ] `go vet ./...` clean. `go test ./internal/services/...` green.
- [ ] Swagger UI shows the 10 new routes with examples.
- [ ] `docs/superpowers/verification/phase-05.md` exists and includes captured request/response pairs for every numbered step in Task 20.
- [ ] No prior migration file modified.
- [ ] README endpoint list updated.
