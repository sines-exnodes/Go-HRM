# User Contracts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a per-user employment-contract register (list, create, update, delete, attachment upload) to the HRM API under `/api/v1/users/:id/contracts`.

**Architecture:** Contracts are a sub-resource of users stored in `user_contracts` with FK to `employees(id)`. The service resolves `user_id → employee_id` via `EmployeeRepository.FindByUserID`. Two fine-grained permissions (`users:contracts_view` / `users:contracts_manage`) mirror the salary/banking pattern.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, golang-migrate SQL migrations, AWS S3 via `Uploader` interface, stretchr/testify integration tests.

---

## ⚠️ REVISION NOTES
_None yet — this is v1.0._

---

## File Map

| Action | Path |
|---|---|
| Create | `migrations/000022_user_contracts.up.sql` |
| Create | `migrations/000022_user_contracts.down.sql` |
| Create | `internal/models/user_contract.go` |
| Create | `internal/dto/user_contract.go` |
| Create | `internal/repositories/user_contract_repo.go` |
| Create | `internal/services/user_contract_service.go` |
| Create | `internal/services/user_contract_service_test.go` |
| Create | `internal/handlers/user_contract_handler.go` |
| Modify | `internal/permissions/registry.go` |
| Modify | `internal/services/seed_service.go` |
| Modify | `internal/services/testhelper_test.go` |
| Modify | `cmd/server/main.go` |

---

## Task 1: Migration 000022

**Files:**
- Create: `migrations/000022_user_contracts.up.sql`
- Create: `migrations/000022_user_contracts.down.sql`

Context: Every entity gets the 4 audit columns + a `BEFORE UPDATE` trigger calling `set_updated_at()` (already exists in DB from earlier migrations). PK is `gen_random_uuid()`. Soft delete uses `is_deleted` boolean — **not** GORM's built-in.

- [ ] **Step 1: Create the up migration**

```sql
-- migrations/000022_user_contracts.up.sql
CREATE TABLE user_contracts (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    UUID        NOT NULL REFERENCES employees(id),
    contract_type  TEXT        NOT NULL,
    signed_date    DATE        NOT NULL,
    expiry_date    DATE,
    is_endless     BOOLEAN     NOT NULL DEFAULT false,
    attachment_url TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_deleted     BOOLEAN     NOT NULL DEFAULT false,
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_user_contracts_employee_id
    ON user_contracts(employee_id)
    WHERE is_deleted = false;

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON user_contracts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 2: Create the down migration**

```sql
-- migrations/000022_user_contracts.down.sql
DROP TABLE IF EXISTS user_contracts;
```

- [ ] **Step 3: Verify migration files exist**

```bash
ls migrations/ | grep 000022
```

Expected output:
```
000022_user_contracts.down.sql
000022_user_contracts.up.sql
```

- [ ] **Step 4: Commit**

```bash
git add migrations/000022_user_contracts.up.sql migrations/000022_user_contracts.down.sql
git commit -m "feat(contracts): migration 000022 — user_contracts table"
```

---

## Task 2: Model + DTO

**Files:**
- Create: `internal/models/user_contract.go`
- Create: `internal/dto/user_contract.go`

Context: `BaseModel` in `internal/models/base.go` provides the 4 audit columns (ID uuid, CreatedAt, UpdatedAt, IsDeleted, DeletedAt). `ContractType` is stored as text — only `"labour_contract"` in v1 but kept as a typed constant so more types can be added without a data migration. DTO `UserContractUpdate` uses pointer fields for partial-PATCH semantics (nil = no change). `AttachmentURL *string` in Update: nil means "leave as-is", empty-string pointer means "remove the attachment".

- [ ] **Step 1: Create the GORM model**

```go
// internal/models/user_contract.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// ContractType is the employment contract category. Only labour_contract in v1.
type ContractType string

const ContractTypeLabour ContractType = "labour_contract"

// UserContract maps to the user_contracts table. Belongs to one employee.
// ExpiryDate is nil when IsEndless is true — never store a sentinel far-future date.
type UserContract struct {
	BaseModel
	EmployeeID    uuid.UUID    `gorm:"type:uuid;not null;index"`
	ContractType  ContractType `gorm:"type:text;not null"`
	SignedDate    time.Time    `gorm:"type:date;not null"`
	ExpiryDate    *time.Time   `gorm:"type:date"`
	IsEndless     bool         `gorm:"not null;default:false"`
	AttachmentURL *string      `gorm:"type:text"`
}

func (UserContract) TableName() string { return "user_contracts" }
```

- [ ] **Step 2: Create the DTOs**

```go
// internal/dto/user_contract.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserContractCreate is the create request body.
type UserContractCreate struct {
	ContractType  string     `json:"contract_type"  binding:"required,oneof=labour_contract"`
	SignedDate    time.Time  `json:"signed_date"    binding:"required"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	IsEndless     bool       `json:"is_endless"`
	AttachmentURL *string    `json:"attachment_url"`
}

// UserContractUpdate is the partial-PATCH request body.
// nil pointer = leave field unchanged.
// AttachmentURL non-nil empty string ("") = remove the attachment.
type UserContractUpdate struct {
	ContractType  *string    `json:"contract_type"  binding:"omitempty,oneof=labour_contract"`
	SignedDate    *time.Time `json:"signed_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	IsEndless     *bool      `json:"is_endless"`
	AttachmentURL *string    `json:"attachment_url"`
}

// UserContractListQuery holds the list filter and pagination params.
type UserContractListQuery struct {
	Page       int        `form:"page"`
	PageSize   int        `form:"page_size"`
	SignedFrom *time.Time `form:"signed_from"`
	SignedTo   *time.Time `form:"signed_to"`
	ExpiryFrom *time.Time `form:"expiry_from"`
	ExpiryTo   *time.Time `form:"expiry_to"`
}

// UserContractRead is the API response shape for a single contract.
type UserContractRead struct {
	ID            uuid.UUID  `json:"id"`
	ContractType  string     `json:"contract_type"`
	SignedDate    time.Time  `json:"signed_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	IsEndless     bool       `json:"is_endless"`
	AttachmentURL *string    `json:"attachment_url"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UserContractAttachmentResponse is returned after a successful attachment upload.
type UserContractAttachmentResponse struct {
	AttachmentURL string `json:"attachment_url"`
}
```

- [ ] **Step 3: Compile check**

```bash
go build ./internal/models/... ./internal/dto/...
```

Expected: no output (clean build).

- [ ] **Step 4: Commit**

```bash
git add internal/models/user_contract.go internal/dto/user_contract.go
git commit -m "feat(contracts): UserContract model + DTOs"
```

---

## Task 3: Repository + truncateAll update

**Files:**
- Create: `internal/repositories/user_contract_repo.go`
- Modify: `internal/services/testhelper_test.go` (add `user_contracts` to TRUNCATE)

Context: The `List` filter for expiry date naturally excludes endless contracts — SQL `WHERE expiry_date >= ?` returns NULL as false, so endless rows (NULL expiry_date) are excluded when an expiry filter is active. Page/pageSize clamping is done here so the repo always gets valid values. `Get` uses `First` (returns `gorm.ErrRecordNotFound` if missing) — the service converts that to a 404.

- [ ] **Step 1: Write a failing compile-time test**

Create the file `internal/repositories/user_contract_repo.go` with only the interface (no impl yet) so the package compiles:

```go
// internal/repositories/user_contract_repo.go
package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
)

// UserContractRepository is the data-access contract for user contracts.
type UserContractRepository interface {
	List(ctx context.Context, employeeID uuid.UUID, q dto.UserContractListQuery) ([]models.UserContract, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*models.UserContract, error)
	Create(ctx context.Context, c *models.UserContract) error
	Update(ctx context.Context, c *models.UserContract) error
	Delete(ctx context.Context, id uuid.UUID) error
}
```

- [ ] **Step 2: Build to verify interface compiles**

```bash
go build ./internal/repositories/...
```

Expected: no output (clean build).

- [ ] **Step 3: Add the GORM implementation**

Append to `internal/repositories/user_contract_repo.go`:

```go
type userContractRepository struct{ db *gorm.DB }

// NewUserContractRepository constructs the GORM-backed implementation.
func NewUserContractRepository(db *gorm.DB) UserContractRepository {
	return &userContractRepository{db: db}
}

func (r *userContractRepository) Create(ctx context.Context, c *models.UserContract) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *userContractRepository) Get(ctx context.Context, id uuid.UUID) (*models.UserContract, error) {
	var c models.UserContract
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = false", id).
		First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *userContractRepository) Update(ctx context.Context, c *models.UserContract) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *userContractRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := timeNow()
	return r.db.WithContext(ctx).Model(&models.UserContract{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *userContractRepository) List(ctx context.Context, employeeID uuid.UUID, q dto.UserContractListQuery) ([]models.UserContract, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.UserContract{}).
		Where("employee_id = ? AND is_deleted = false", employeeID)

	if q.SignedFrom != nil {
		db = db.Where("signed_date >= ?", q.SignedFrom)
	}
	if q.SignedTo != nil {
		db = db.Where("signed_date <= ?", q.SignedTo)
	}
	if q.ExpiryFrom != nil {
		db = db.Where("expiry_date >= ?", q.ExpiryFrom)
	}
	if q.ExpiryTo != nil {
		db = db.Where("expiry_date <= ?", q.ExpiryTo)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var contracts []models.UserContract
	err := db.Order("signed_date DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&contracts).Error
	if err != nil {
		return nil, 0, err
	}
	return contracts, total, nil
}
```

> **Note on `timeNow()`:** This helper already exists in the repositories package (used by other repos for soft-delete). If it doesn't, replace `timeNow()` with `time.Now()` and add `"time"` to the import.

- [ ] **Step 4: Build to verify implementation compiles**

```bash
go build ./internal/repositories/...
```

Expected: no output (clean build). If `timeNow` is undefined, change to `time.Now()` and add `"time"` to the import block.

- [ ] **Step 5: Add `user_contracts` to `truncateAll` in the test helper**

Open `internal/services/testhelper_test.go`. Find the long `TRUNCATE TABLE ...` statement (around line 134). Add `user_contracts` at the beginning of the list (before `invites`) because user_contracts has a FK to employees:

```go
// Before (existing line, truncated for brevity):
if err := testDB.Exec(`TRUNCATE TABLE invites, system_config, ...`).Error; err != nil {

// After:
if err := testDB.Exec(`TRUNCATE TABLE user_contracts, invites, system_config, announcement_views, announcement_attachments, announcement_target_users, announcement_target_departments, announcement_labels, announcements, labels, employee_skills, skills, device_tokens, user_notification_settings, attendance_sessions, attendance, leave_requests, employee_leave_quotas, employee_emergency_contacts, dependents, employees, positions, departments, user_roles, users, roles RESTART IDENTITY CASCADE`).Error; err != nil {
```

- [ ] **Step 6: Build again**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add internal/repositories/user_contract_repo.go internal/services/testhelper_test.go
git commit -m "feat(contracts): UserContractRepository + truncateAll update"
```

---

## Task 4: Service — Create, Get, Delete + tests (TDD)

**Files:**
- Create: `internal/services/user_contract_service.go` (skeleton + Create/Get/Delete)
- Create: `internal/services/user_contract_service_test.go` (Create/Get/Delete tests)

Context: The service resolves `user_id → employee_id` via `employeeRepo.FindByUserID`. All three methods share two private helpers: `resolveEmployee` and `fetchAndCheckOwnership`. The `validateDates` helper enforces the endless/expiry invariant. Integration tests skip when `TEST_DATABASE_URL` is not set. Each test calls `truncateAll(t)` at the top for a clean slate.

- [ ] **Step 1: Write the failing tests**

```go
// internal/services/user_contract_service_test.go
package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
)

// newContractSvc builds a UserContractService pointing at the test DB.
func newContractSvc(t *testing.T) *services.UserContractService {
	t.Helper()
	return services.NewUserContractService(
		repositories.NewUserContractRepository(testDB),
		repositories.NewEmployeeRepository(testDB),
		nil, // uploads — nil disables attachment upload; tested separately
	)
}

// dateUTC returns a UTC midnight time for a given date.
func dateUTC(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// ---- Create ----

func TestUserContract_Create_FixedTerm(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c1@example.com", "Alice Smith")
	expiry := dateUTC(2027, 6, 30)
	out, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 7, 1),
		ExpiryDate:   &expiry,
		IsEndless:    false,
	})
	require.Nil(t, err)
	assert.Equal(t, "labour_contract", out.ContractType)
	assert.False(t, out.IsEndless)
	require.NotNil(t, out.ExpiryDate)
	assert.True(t, out.ExpiryDate.Equal(expiry))
}

func TestUserContract_Create_Endless(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c2@example.com", "Bob Jones")
	expiry := dateUTC(2027, 1, 1) // should be cleared
	out, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
		IsEndless:    true,
	})
	require.Nil(t, err)
	assert.True(t, out.IsEndless)
	assert.Nil(t, out.ExpiryDate, "endless contract must have nil expiry in DB")
}

func TestUserContract_Create_ExpiryBeforeSigned(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c3@example.com", "Carol White")
	expiry := dateUTC(2026, 5, 31) // before signed date
	_, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 6, 1),
		ExpiryDate:   &expiry,
	})
	require.NotNil(t, err)
	assert.Equal(t, 400, err.StatusCode)
	assert.Contains(t, err.Message, "after signed date")
}

func TestUserContract_Create_ExpiryEqualsSigned(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c4@example.com", "Dan Brown")
	same := dateUTC(2026, 6, 1)
	_, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   same,
		ExpiryDate:   &same,
	})
	require.NotNil(t, err)
	assert.Equal(t, 400, err.StatusCode)
}

func TestUserContract_Create_MissingExpiry_NotEndless(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "c5@example.com", "Eva Green")
	_, err := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   nil,
		IsEndless:    false,
	})
	require.NotNil(t, err)
	assert.Equal(t, 400, err.StatusCode)
	assert.Contains(t, err.Message, "expiry date is required")
}

// ---- Get ----

func TestUserContract_Get_WrongEmployee(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	owner, _ := makeEmpUser(t, "owner@example.com", "Owner User")
	other, _ := makeEmpUser(t, "other@example.com", "Other User")
	expiry := dateUTC(2027, 1, 1)
	created, aerr := svc.Create(context.Background(), owner.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	// Fetch the contract using the OTHER user's ID — must 404 (no info leak)
	_, aerr = svc.Get(context.Background(), other.ID, created.ID)
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.StatusCode)
}

// ---- Delete ----

func TestUserContract_Delete_SoftDeletes(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "del@example.com", "Delete Me")
	expiry := dateUTC(2027, 6, 30)
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	aerr = svc.Delete(context.Background(), u.ID, created.ID)
	require.Nil(t, aerr)

	// Subsequent Get must return 404
	_, aerr = svc.Get(context.Background(), u.ID, created.ID)
	require.NotNil(t, aerr)
	assert.Equal(t, 404, aerr.StatusCode)

	// Verify the row is still in DB with is_deleted = true
	var count int64
	testDB.Raw("SELECT COUNT(*) FROM user_contracts WHERE id = ? AND is_deleted = true", created.ID).Scan(&count)
	assert.Equal(t, int64(1), count)
}
```

- [ ] **Step 2: Run to verify they fail (service doesn't exist yet)**

```bash
go test ./internal/services/... -run "TestUserContract_Create|TestUserContract_Get|TestUserContract_Delete" -v 2>&1 | head -20
```

Expected: compilation error — `services.NewUserContractService undefined` or `services.UserContractService undefined`.

- [ ] **Step 3: Create the service skeleton with Create, Get, Delete**

```go
// internal/services/user_contract_service.go
package services

import (
	"context"
	"errors"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)

const (
	contractAttachmentSubdir   = "contracts"
	contractAttachmentMaxBytes = 5 * 1024 * 1024
	contractDocxMIME           = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

var allowedContractMIME = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	contractDocxMIME:  true,
}

// UserContractService manages employment contracts per user.
type UserContractService struct {
	repo         repositories.UserContractRepository
	employeeRepo repositories.EmployeeRepository
	uploads      Uploader // may be nil when storage is unconfigured
}

// NewUserContractService constructs a UserContractService.
// Pass nil for uploads to disable attachment upload (graceful degradation).
func NewUserContractService(
	repo repositories.UserContractRepository,
	employeeRepo repositories.EmployeeRepository,
	uploads Uploader,
) *UserContractService {
	return &UserContractService{repo: repo, employeeRepo: employeeRepo, uploads: uploads}
}

// ---- private helpers ----

func (s *UserContractService) resolveEmployee(ctx context.Context, userID uuid.UUID) (*models.Employee, *apperrors.AppError) {
	emp, err := s.employeeRepo.FindByUserID(ctx, userID)
	if err != nil || emp == nil {
		return nil, apperrors.ErrNotFound("employee profile not found")
	}
	return emp, nil
}

func (s *UserContractService) fetchAndCheckOwnership(ctx context.Context, contractID, employeeID uuid.UUID) (*models.UserContract, *apperrors.AppError) {
	c, err := s.repo.Get(ctx, contractID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("contract not found")
		}
		return nil, apperrors.ErrInternal(err.Error())
	}
	if c.EmployeeID != employeeID {
		return nil, apperrors.ErrNotFound("contract not found")
	}
	return c, nil
}

func validateContractDates(isEndless bool, signedDate time.Time, expiryDate *time.Time) *apperrors.AppError {
	if isEndless {
		return nil
	}
	if expiryDate == nil {
		return apperrors.ErrBadRequest("expiry date is required")
	}
	if !expiryDate.After(signedDate) {
		return apperrors.ErrBadRequest("expiry date must be after signed date")
	}
	return nil
}

func toUserContractRead(c models.UserContract) dto.UserContractRead {
	return dto.UserContractRead{
		ID:            c.ID,
		ContractType:  string(c.ContractType),
		SignedDate:    c.SignedDate,
		ExpiryDate:    c.ExpiryDate,
		IsEndless:     c.IsEndless,
		AttachmentURL: c.AttachmentURL,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ---- public methods ----

// Create adds a new contract for the user identified by userID.
func (s *UserContractService) Create(ctx context.Context, userID uuid.UUID, req dto.UserContractCreate) (*dto.UserContractRead, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	expiry := req.ExpiryDate
	if req.IsEndless {
		expiry = nil
	}
	if aerr := validateContractDates(req.IsEndless, req.SignedDate, expiry); aerr != nil {
		return nil, aerr
	}
	c := &models.UserContract{
		EmployeeID:    emp.ID,
		ContractType:  models.ContractType(req.ContractType),
		SignedDate:    req.SignedDate,
		ExpiryDate:    expiry,
		IsEndless:     req.IsEndless,
		AttachmentURL: req.AttachmentURL,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := toUserContractRead(*c)
	return &out, nil
}

// Get returns a single contract. Returns 404 if the contract doesn't belong to the user.
func (s *UserContractService) Get(ctx context.Context, userID, contractID uuid.UUID) (*dto.UserContractRead, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	c, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID)
	if aerr != nil {
		return nil, aerr
	}
	out := toUserContractRead(*c)
	return &out, nil
}

// Delete soft-deletes a contract.
func (s *UserContractService) Delete(ctx context.Context, userID, contractID uuid.UUID) *apperrors.AppError {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return aerr
	}
	if _, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID); aerr != nil {
		return aerr
	}
	if err := s.repo.Delete(ctx, contractID); err != nil {
		return apperrors.ErrInternal(err.Error())
	}
	return nil
}

// List — stubbed, implemented in Task 5.
func (s *UserContractService) List(_ context.Context, _ uuid.UUID, _ dto.UserContractListQuery) (dto.PaginatedData[dto.UserContractRead], *apperrors.AppError) {
	return dto.PaginatedData[dto.UserContractRead]{}, nil
}

// Update — stubbed, implemented in Task 5.
func (s *UserContractService) Update(_ context.Context, _, _ uuid.UUID, _ dto.UserContractUpdate) (*dto.UserContractRead, *apperrors.AppError) {
	return nil, nil
}

// UploadAttachment — stubbed, implemented in Task 5.
func (s *UserContractService) UploadAttachment(_ context.Context, _, _ uuid.UUID, _ []byte, _ string) (*dto.UserContractAttachmentResponse, *apperrors.AppError) {
	return nil, nil
}
```

- [ ] **Step 4: Run the Create/Get/Delete tests**

```bash
go test ./internal/services/... -run "TestUserContract_Create|TestUserContract_Get|TestUserContract_Delete" -v
```

Expected: all PASS (or SKIP if TEST_DATABASE_URL is not set).

- [ ] **Step 5: Commit**

```bash
git add internal/services/user_contract_service.go internal/services/user_contract_service_test.go
git commit -m "feat(contracts): service Create/Get/Delete + integration tests (TDD)"
```

---

## Task 5: Service — List, Update, UploadAttachment + tests (TDD)

**Files:**
- Modify: `internal/services/user_contract_service.go` (replace stubs with real implementations)
- Modify: `internal/services/user_contract_service_test.go` (add List/Update/Upload tests)

Context: `List` normalizes page/pageSize before calling the repo. `Update` applies only non-nil DTO fields onto the fetched model ("partial PATCH"). `AttachmentURL` with empty-string value means "remove". The endless toggle re-runs `validateContractDates` against the resulting state after applying changes. `UploadAttachment` sniffs bytes (not file name) for content-type; DOCX is ZIP-sniffed but recognized by `.docx` extension (same pattern as leave attachments).

- [ ] **Step 1: Add List/Update/UploadAttachment tests to the test file**

Append to `internal/services/user_contract_service_test.go`:

```go
// ---- List + Filters ----

func TestUserContract_List_FilterSignedDate(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "list1@example.com", "List User")
	expiry := dateUTC(2027, 12, 31)

	// Inside range
	_, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 6, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	// Outside range
	_, aerr = svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2024, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	from := dateUTC(2026, 1, 1)
	to := dateUTC(2026, 12, 31)
	result, aerr := svc.List(context.Background(), u.ID, dto.UserContractListQuery{
		Page: 1, PageSize: 10,
		SignedFrom: &from,
		SignedTo:   &to,
	})
	require.Nil(t, aerr)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Items, 1)
}

func TestUserContract_List_ExpiryFilter_ExcludesEndless(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "list2@example.com", "List User 2")

	// Endless contract (no expiry)
	_, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		IsEndless:    true,
	})
	require.Nil(t, aerr)

	// Fixed-term contract expiring in 2027
	expiry := dateUTC(2027, 6, 30)
	_, aerr = svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)

	from := dateUTC(2027, 1, 1)
	to := dateUTC(2027, 12, 31)
	result, aerr := svc.List(context.Background(), u.ID, dto.UserContractListQuery{
		Page: 1, PageSize: 10,
		ExpiryFrom: &from,
		ExpiryTo:   &to,
	})
	require.Nil(t, aerr)
	assert.Equal(t, int64(1), result.Total, "endless contract must not appear in expiry filter results")
}

func TestUserContract_List_Pagination(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "list3@example.com", "Paginated User")
	for i := 0; i < 3; i++ {
		expiry := dateUTC(2027+i, 12, 31)
		_, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
			ContractType: "labour_contract",
			SignedDate:   dateUTC(2026+i, 1, 1),
			ExpiryDate:   &expiry,
		})
		require.Nil(t, aerr)
	}

	result, aerr := svc.List(context.Background(), u.ID, dto.UserContractListQuery{Page: 1, PageSize: 2})
	require.Nil(t, aerr)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.TotalPages)
}

// ---- Update ----

func TestUserContract_Update_RemoveAttachment(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "upd1@example.com", "Update User")
	url := "https://example.com/file.pdf"
	expiry := dateUTC(2027, 12, 31)
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType:  "labour_contract",
		SignedDate:    dateUTC(2026, 1, 1),
		ExpiryDate:    &expiry,
		AttachmentURL: &url,
	})
	require.Nil(t, aerr)

	// Send empty string to remove attachment
	empty := ""
	updated, aerr := svc.Update(context.Background(), u.ID, created.ID, dto.UserContractUpdate{
		AttachmentURL: &empty,
	})
	require.Nil(t, aerr)
	assert.Nil(t, updated.AttachmentURL)
}

func TestUserContract_Update_EndlessToggleOn(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "upd2@example.com", "Endless Toggle")
	expiry := dateUTC(2027, 6, 30)
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		ExpiryDate:   &expiry,
	})
	require.Nil(t, aerr)
	assert.False(t, created.IsEndless)

	endless := true
	updated, aerr := svc.Update(context.Background(), u.ID, created.ID, dto.UserContractUpdate{
		IsEndless: &endless,
	})
	require.Nil(t, aerr)
	assert.True(t, updated.IsEndless)
	assert.Nil(t, updated.ExpiryDate, "turning endless ON must clear the expiry date")
}

func TestUserContract_Update_EndlessToggleOff_RequiresExpiry(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	svc := newContractSvc(t)

	u, _ := makeEmpUser(t, "upd3@example.com", "Endless Off")
	created, aerr := svc.Create(context.Background(), u.ID, dto.UserContractCreate{
		ContractType: "labour_contract",
		SignedDate:   dateUTC(2026, 1, 1),
		IsEndless:    true,
	})
	require.Nil(t, aerr)

	notEndless := false
	_, aerr = svc.Update(context.Background(), u.ID, created.ID, dto.UserContractUpdate{
		IsEndless:  &notEndless,
		ExpiryDate: nil, // no expiry provided — must 400
	})
	require.NotNil(t, aerr)
	assert.Equal(t, 400, aerr.StatusCode)
	assert.Contains(t, aerr.Message, "expiry date is required")
}
```

- [ ] **Step 2: Run to verify the new tests fail (stubs return nil)**

```bash
go test ./internal/services/... -run "TestUserContract_List|TestUserContract_Update" -v 2>&1 | head -30
```

Expected: tests skip (no DB) or fail on assertions (stubs return zero values).

- [ ] **Step 3: Replace the List stub with the real implementation**

In `internal/services/user_contract_service.go`, replace the `List` stub:

```go
// List returns a paginated, optionally filtered list of contracts for the user.
func (s *UserContractService) List(ctx context.Context, userID uuid.UUID, q dto.UserContractListQuery) (dto.PaginatedData[dto.UserContractRead], *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return dto.PaginatedData[dto.UserContractRead]{}, aerr
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 50 {
		q.PageSize = 10
	}
	contracts, total, err := s.repo.List(ctx, emp.ID, q)
	if err != nil {
		return dto.PaginatedData[dto.UserContractRead]{}, apperrors.ErrInternal(err.Error())
	}
	items := make([]dto.UserContractRead, len(contracts))
	for i, c := range contracts {
		items[i] = toUserContractRead(c)
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(q.PageSize)))
	}
	return dto.PaginatedData[dto.UserContractRead]{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}
```

- [ ] **Step 4: Replace the Update stub with the real implementation**

In `internal/services/user_contract_service.go`, replace the `Update` stub:

```go
// Update applies a partial PATCH to a contract. Only non-nil fields are changed.
func (s *UserContractService) Update(ctx context.Context, userID, contractID uuid.UUID, req dto.UserContractUpdate) (*dto.UserContractRead, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	c, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID)
	if aerr != nil {
		return nil, aerr
	}

	// Apply non-nil fields onto the fetched model.
	if req.ContractType != nil {
		c.ContractType = models.ContractType(*req.ContractType)
	}
	if req.SignedDate != nil {
		c.SignedDate = *req.SignedDate
	}
	if req.IsEndless != nil {
		c.IsEndless = *req.IsEndless
	}
	if req.ExpiryDate != nil {
		c.ExpiryDate = req.ExpiryDate
	}
	if req.AttachmentURL != nil {
		if *req.AttachmentURL == "" {
			c.AttachmentURL = nil
		} else {
			c.AttachmentURL = req.AttachmentURL
		}
	}

	// If toggled to endless, always clear expiry.
	if c.IsEndless {
		c.ExpiryDate = nil
	}

	// Re-validate the resulting state.
	if aerr := validateContractDates(c.IsEndless, c.SignedDate, c.ExpiryDate); aerr != nil {
		return nil, aerr
	}

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	out := toUserContractRead(*c)
	return &out, nil
}
```

- [ ] **Step 5: Replace the UploadAttachment stub with the real implementation**

In `internal/services/user_contract_service.go`, replace the `UploadAttachment` stub:

```go
// UploadAttachment validates, uploads, and stores the attachment URL on the contract.
// content is the raw file bytes; ext is the lowercase file extension (e.g. ".pdf").
func (s *UserContractService) UploadAttachment(ctx context.Context, userID, contractID uuid.UUID, content []byte, ext string) (*dto.UserContractAttachmentResponse, *apperrors.AppError) {
	emp, aerr := s.resolveEmployee(ctx, userID)
	if aerr != nil {
		return nil, aerr
	}
	c, aerr := s.fetchAndCheckOwnership(ctx, contractID, emp.ID)
	if aerr != nil {
		return nil, aerr
	}
	if s.uploads == nil {
		return nil, apperrors.ErrInternal("storage is not configured; cannot upload attachment")
	}
	if len(content) == 0 {
		return nil, apperrors.ErrBadRequest("attachment file is empty")
	}
	if len(content) > contractAttachmentMaxBytes {
		return nil, apperrors.ErrBadRequest("attachment must not exceed 5 MB")
	}
	sniffLen := len(content)
	if sniffLen > 512 {
		sniffLen = 512
	}
	sniffed := http.DetectContentType(content[:sniffLen])
	if sniffed == "application/zip" && ext == ".docx" {
		sniffed = contractDocxMIME
	}
	if !allowedContractMIME[sniffed] {
		return nil, apperrors.ErrBadRequest("attachment must be PDF, PNG, JPG, or DOCX")
	}
	url, err := s.uploads.Upload(ctx, contractAttachmentSubdir, ext, content, sniffed)
	if err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	c.AttachmentURL = &url
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, apperrors.ErrInternal(err.Error())
	}
	return &dto.UserContractAttachmentResponse{AttachmentURL: url}, nil
}
```

- [ ] **Step 6: Check unused imports — add/remove as needed**

The `UploadAttachment` method uses `net/http`, `path/filepath`, and `strings`. Verify they are all referenced; remove unused ones. The `math` import is used by `List`. The `strings` import is only used if you need `strings.ToLower` — this is handled in the handler, not the service. Remove `path/filepath` and `strings` from the service if not used elsewhere.

Final import block for the service file:

```go
import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
)
```

- [ ] **Step 7: Run full test suite for this service**

```bash
go test ./internal/services/... -run "TestUserContract" -v
```

Expected: all PASS (or SKIP when TEST_DATABASE_URL not set).

- [ ] **Step 8: Run full suite (regression check)**

```bash
go test ./internal/services/... -count=1
```

Expected: all PASS, 0 FAIL.

- [ ] **Step 9: Commit**

```bash
git add internal/services/user_contract_service.go internal/services/user_contract_service_test.go
git commit -m "feat(contracts): service List/Update/UploadAttachment + full integration tests"
```

---

## Task 6: Permissions + Seed

**Files:**
- Modify: `internal/permissions/registry.go`
- Modify: `internal/services/seed_service.go`

Context: Two new constants go in the `users` resource group (matching salary/banking). Seed adds `contracts_view` to all roles (read-only access for everyone) and `contracts_manage` to Admin + HR Manager (same two roles that hold salary/banking manage perms). The seed is idempotent — it appends missing perms to existing system roles on every boot.

- [ ] **Step 1: Add permission constants to registry.go**

Open `internal/permissions/registry.go`. After `PermUsersBankingManage` (around line 28), add:

```go
	// Contracts fine-grained perms (DR-001-005-10) — two-tier mirror of salary/banking.
	PermUsersContractsView   Permission = "users:contracts_view"
	PermUsersContractsManage Permission = "users:contracts_manage"
```

- [ ] **Step 2: Add to AllPermissions()**

In `AllPermissions()`, after `PermUsersBankingManage,` add:

```go
		PermUsersContractsView, PermUsersContractsManage,
```

- [ ] **Step 3: Add to PermissionGroups**

In the `users` `PermissionGroup` `Permissions` slice, after the banking entry, add:

```go
			{PermUsersContractsView, "View Contracts", "See the contracts list and details for a user profile"},
			{PermUsersContractsManage, "Manage Contracts", "Create, edit, and delete employment contracts on a user profile"},
```

- [ ] **Step 4: Run registry test to verify AllPermissions + PermissionGroups stay in sync**

```bash
go test ./internal/permissions/... -v
```

Expected: PASS.

- [ ] **Step 5: Add permissions to seed roles in seed_service.go**

Open `internal/services/seed_service.go`. In the `systemRoles` slice:

**Super Admin** — already gets `PermAll` (wildcard), no change needed.

**Admin role** (around line 69–99): after `PermUsersBankingManage,` on line 76, add:
```go
				permissions.PermUsersContractsView, permissions.PermUsersContractsManage,
```

**HR Manager role** (around line 101–131): after `PermUsersBankingManage,` on line 111, add:
```go
				permissions.PermUsersContractsView, permissions.PermUsersContractsManage,
```

**Manager role** (around line 132–145): after `permissions.PermUsersRead,` add:
```go
				permissions.PermUsersContractsView,
```

**Employee role** (around line 147–165): after `permissions.PermAuthLogin,` add:
```go
				permissions.PermUsersContractsView,
```

- [ ] **Step 6: Build and run seed test**

```bash
go build ./... && go test ./internal/services/... -run "TestSeed" -v
```

Expected: PASS (or SKIP without DB).

- [ ] **Step 7: Commit**

```bash
git add internal/permissions/registry.go internal/services/seed_service.go
git commit -m "feat(contracts): permissions registry + seed (view all roles, manage Admin+HR Manager)"
```

---

## Task 7: Handler + Wire + Routes

**Files:**
- Create: `internal/handlers/user_contract_handler.go`
- Modify: `cmd/server/main.go`

Context: The handler follows the exact same two-layer permission pattern as `AnnouncementHandler` (`hasAnnounceManageAll`): router-level `RequirePerms(PermUsersContractsView)` + a `hasContractsManage` helper in the handler for write operations. Routes use `:id` for the user param (NOT `:user_id`) because Gin requires a single wildcard name per path position — the `adminUsers` group already uses `:id` for `/users/:id/*`. The contract param uses `:contractID`. `UploadAttachment` reads the file via `c.FormFile("file")`, slurps bytes, lowercases the extension, then delegates to the service.

- [ ] **Step 1: Create the handler**

```go
// internal/handlers/user_contract_handler.go
package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/permissions"
	"github.com/exnodes/hrm-api/internal/services"
)

// UserContractHandler owns /api/v1/users/:id/contracts.
type UserContractHandler struct {
	svc *services.UserContractService
}

func NewUserContractHandler(svc *services.UserContractService) *UserContractHandler {
	return &UserContractHandler{svc: svc}
}

// hasContractsManage returns true when the JWT-loaded caller holds
// users:contracts_manage or the wildcard *.
func hasContractsManage(c *gin.Context) bool {
	u, ok := currentUser(c)
	if !ok {
		return false
	}
	for _, r := range u.Roles {
		for _, p := range []string(r.Permissions) {
			if p == string(permissions.PermUsersContractsManage) || p == string(permissions.PermAll) {
				return true
			}
		}
	}
	return false
}

// List godoc
// @Summary      List contracts for a user
// @Tags         contracts
// @Security     BearerAuth
// @Produce      json
// @Param        id          path   string  true   "user uuid"
// @Param        page        query  int     false  "page (default 1)"
// @Param        page_size   query  int     false  "page size (default 10, max 50)"
// @Param        signed_from query  string  false  "signed date from (RFC3339)"
// @Param        signed_to   query  string  false  "signed date to (RFC3339)"
// @Param        expiry_from query  string  false  "expiry date from (RFC3339)"
// @Param        expiry_to   query  string  false  "expiry date to (RFC3339)"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/users/{id}/contracts [get]
func (h *UserContractHandler) List(c *gin.Context) {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var q dto.UserContractListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.List(c.Request.Context(), userID, q)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.UserContractRead]]{Success: true, Data: out})
}

// Get godoc
// @Summary      Get a single contract
// @Tags         contracts
// @Security     BearerAuth
// @Produce      json
// @Param        id          path  string  true  "user uuid"
// @Param        contractID  path  string  true  "contract uuid"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/users/{id}/contracts/{contractID} [get]
func (h *UserContractHandler) Get(c *gin.Context) {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, aerr := h.svc.Get(c.Request.Context(), userID, contractID)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.UserContractRead]{Success: true, Data: out})
}

// Create godoc
// @Summary      Create a contract for a user
// @Tags         contracts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                  true  "user uuid"
// @Param        body  body  dto.UserContractCreate  true  "create payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /api/v1/users/{id}/contracts [post]
func (h *UserContractHandler) Create(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.UserContractCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.Create(c.Request.Context(), userID, req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[*dto.UserContractRead]{Success: true, Message: "Contract has been created", Data: out})
}

// Update godoc
// @Summary      Update a contract (partial PATCH)
// @Tags         contracts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id          path  string                  true  "user uuid"
// @Param        contractID  path  string                  true  "contract uuid"
// @Param        body        body  dto.UserContractUpdate  true  "patch payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/users/{id}/contracts/{contractID} [patch]
func (h *UserContractHandler) Update(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req dto.UserContractUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.ErrBadRequest(err.Error()))
		return
	}
	out, aerr := h.svc.Update(c.Request.Context(), userID, contractID, req)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.UserContractRead]{Success: true, Message: "Contract has been updated", Data: out})
}

// Delete godoc
// @Summary      Soft-delete a contract
// @Tags         contracts
// @Security     BearerAuth
// @Produce      json
// @Param        id          path  string  true  "user uuid"
// @Param        contractID  path  string  true  "contract uuid"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/users/{id}/contracts/{contractID} [delete]
func (h *UserContractHandler) Delete(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if aerr := h.svc.Delete(c.Request.Context(), userID, contractID); aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[struct{}]{Success: true, Message: "Contract has been deleted"})
}

// UploadAttachment godoc
// @Summary      Upload or replace a contract attachment
// @Tags         contracts
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path      string  true  "user uuid"
// @Param        contractID  path      string  true  "contract uuid"
// @Param        file        formData  file    true  "attachment (PDF/PNG/JPG/DOCX, max 5MB)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /api/v1/users/{id}/contracts/{contractID}/attachment [post]
func (h *UserContractHandler) UploadAttachment(c *gin.Context) {
	if !hasContractsManage(c) {
		_ = c.Error(apperrors.ErrForbidden("contracts management permission required"))
		return
	}
	userID, err := parseIDParam(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}
	contractID, err := parseIDParam(c, "contractID")
	if err != nil {
		_ = c.Error(err)
		return
	}
	fileHeader, ferr := c.FormFile("file")
	if ferr != nil {
		_ = c.Error(apperrors.ErrBadRequest("file is required"))
		return
	}
	f, ferr := fileHeader.Open()
	if ferr != nil {
		_ = c.Error(apperrors.ErrBadRequest("cannot open uploaded file"))
		return
	}
	defer f.Close()
	content, ferr := io.ReadAll(f)
	if ferr != nil {
		_ = c.Error(apperrors.ErrBadRequest("cannot read uploaded file"))
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	out, aerr := h.svc.UploadAttachment(c.Request.Context(), userID, contractID, content, ext)
	if aerr != nil {
		_ = c.Error(aerr)
		return
	}
	c.JSON(http.StatusOK, dto.Response[*dto.UserContractAttachmentResponse]{Success: true, Data: out})
}
```

- [ ] **Step 2: Wire the repo, service, and handler in main.go**

Open `cmd/server/main.go`. After the `inviteRepo` line (around line 67), add:

```go
	userContractRepo := repositories.NewUserContractRepository(db)
```

After the `inviteSvc` line (around line 124), add:

```go
	userContractSvc := services.NewUserContractService(userContractRepo, employeeRepo, uploadSvc)
```

After the `notifH` line (around line 147), add:

```go
	userContractH := handlers.NewUserContractHandler(userContractSvc)
```

- [ ] **Step 3: Register the routes in main.go**

After the `adminUsers` block (the block ending around `adminUsers.PUT(":id/roles", ...)`), add:

```go
		// ---- /users/:id/contracts ----
		// NOTE: :id must match the wildcard used by adminUsers (:id) — Gin
		// requires a single wildcard name per path position.
		// Router-level gate: PermUsersContractsView. Write ops additionally
		// checked inside the handler via hasContractsManage.
		userContracts := authed.Group("/users/:id/contracts")
		userContracts.Use(middleware.RequirePerms(authSvc, permissions.PermUsersContractsView))
		userContracts.GET("", userContractH.List)
		userContracts.POST("", userContractH.Create)
		userContracts.GET(":contractID", userContractH.Get)
		userContracts.PATCH(":contractID", userContractH.Update)
		userContracts.DELETE(":contractID", userContractH.Delete)
		userContracts.POST(":contractID/attachment", userContractH.UploadAttachment)
```

- [ ] **Step 4: Build the full project**

```bash
go build ./...
```

Expected: no output (clean build). If there are import errors in the handler (unused `path/filepath` or `strings`), remove them. Both are used: `filepath.Ext` and `strings.ToLower`.

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/user_contract_handler.go cmd/server/main.go
git commit -m "feat(contracts): handler, routes, wire main.go"
```

---

## Task 8: Verification — fmt, vet, test, swag, CHECKPOINT

**Files:**
- Modify: `docs/superpowers/CHECKPOINT.md`
- Modify: `docs/swagger/` (generated — do not hand-edit)

- [ ] **Step 1: Format**

```bash
make fmt
```

Expected: no output (already formatted) or a list of files touched.

- [ ] **Step 2: Vet**

```bash
make vet
```

Expected: no output (clean).

- [ ] **Step 3: Full test suite**

```bash
make test
```

Expected: all PASS, 0 FAIL. Tests that need a DB will SKIP if `TEST_DATABASE_URL` is not set — that is acceptable.

- [ ] **Step 4: Regenerate Swagger**

```bash
make swag
```

Expected: `docs/swagger/docs.go`, `docs/swagger/swagger.json`, `docs/swagger/swagger.yaml` updated.

- [ ] **Step 5: Commit fmt/swag changes**

```bash
git add docs/swagger/ && git diff --name-only HEAD
git commit -m "docs(contracts): regenerate Swagger after user contracts handler"
```

If `make fmt` changed any files, include them in the commit too:
```bash
git add -u && git commit -m "docs(contracts): regenerate Swagger + fmt"
```

- [ ] **Step 6: Smoke test (if server is reachable)**

If the server is running locally on port 8080 (or 8082 per the dev env note in CHECKPOINT):

```bash
# 1. Get a token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@exnodes.com","password":"<your-password>"}' \
  | jq -r '.data.access_token')

# 2. List contracts for a known user UUID (replace <USER_ID>)
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/users/<USER_ID>/contracts | jq

# Expected: {"success":true,"data":{"items":[],"total":0,"page":1,"page_size":10,"total_pages":0}}

# 3. Create a contract
curl -s -X POST http://localhost:8080/api/v1/users/<USER_ID>/contracts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"contract_type":"labour_contract","signed_date":"2026-07-01T00:00:00Z","expiry_date":"2027-06-30T00:00:00Z","is_endless":false}' | jq

# Expected: {"success":true,"message":"Contract has been created","data":{...}}
```

- [ ] **Step 7: Update CHECKPOINT.md**

Open `docs/superpowers/CHECKPOINT.md` and replace the content to reflect:
- User Contracts module is implemented and verified
- Migration 000022 applied
- Next priorities updated

Key fields to update:
```
**Last updated:** 2026-06-12
**Stopped at:** User Contracts module implemented (migration 000022). Next: deploy to dev, then Request Tickets module (EP-003, migration 000023).
**DB migration version:** 22
```

Also update the "Subsequent priorities" list: bump Request Tickets from migration 000022 to 000023, and add "User Contracts deployed to dev" as immediate next step.

- [ ] **Step 8: Commit CHECKPOINT**

```bash
git add docs/superpowers/CHECKPOINT.md
git commit -m "docs(checkpoint): User Contracts module complete — migration 000022, 6 endpoints"
```

---

## Completion Criteria

All of the following must be true before this work is "done":

- [ ] `go build ./...` — clean
- [ ] `make fmt && make vet` — clean
- [ ] `make test` — 0 FAIL (skips acceptable without DB)
- [ ] `make swag` — Swagger regenerated; contracts endpoints visible in the generated spec
- [ ] All 14 integration tests in `user_contract_service_test.go` PASS against the test DB
- [ ] 6 routes registered and reachable: List / Get / Create / Update / Delete / UploadAttachment
- [ ] `PermUsersContractsView` and `PermUsersContractsManage` in `AllPermissions()` and `PermissionGroups`
- [ ] Seed: view perm on all 5 roles; manage perm on Admin + HR Manager
- [ ] CHECKPOINT.md updated with migration 000022 and next steps
