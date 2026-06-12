# User Contracts Module — Design Spec

**Date:** 2026-06-12
**Author:** Brainstorming session
**BA Source:** `ba-requirements/docs/PLATFORMS/WEB-APP/EP-001-foundation/US-005-user-management/details/DR-001-005-10-contracts.md` (v1.2)
**Migration:** 000022
**Status:** Approved for implementation

---

## Goal

Add a per-user employment-contract register to the HRM API. HR managers can list, create, update, and delete contracts attached to a user's employee profile. Contracts support a fixed-term (expiry date required) or endless (no expiry date) model. An optional file attachment stores the signed document in S3.

---

## Architecture

The module follows the standard vertical slice: migration → model → repository → service → handler → routes. Contracts are a sub-resource of users, exposed at `/api/v1/users/:user_id/contracts`. Internally the FK targets `employees(id)` (codebase convention); the service resolves `user_id → employee_id` via `EmployeeRepository.FindByUserID` at the start of every operation.

Two permission levels mirror the salary/banking fine-grained permission pattern:
- `users:contracts_view` — list, get (seeded to all roles)
- `users:contracts_manage` — create, update, delete, upload (seeded to Admin + HR Manager)

Attachment upload is a separate endpoint (`POST .../attachment`) that uploads to S3 and writes the URL back to the contract row. Create/update accept `attachment_url` as a plain string field.

---

## 1. Data Model

### Migration 000022 — `up`

```sql
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

### Migration 000022 — `down`

```sql
DROP TABLE IF EXISTS user_contracts;
```

### GORM Model — `internal/models/user_contract.go`

```go
package models

import (
    "time"
    "github.com/google/uuid"
)

type ContractType string

const ContractTypeLabour ContractType = "labour_contract"

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

**Invariants (enforced in service, not DB):**
- When `IsEndless = true`, `ExpiryDate` is always `nil` before save
- When `IsEndless = false`, `ExpiryDate` is required and must be strictly after `SignedDate`
- `ContractType` is `"labour_contract"` (only value in v1; field is stored as text for forward-compatibility)

---

## 2. Permissions & Routes

### New permission constants — `internal/permissions/registry.go`

```go
PermUsersContractsView   Permission = "users:contracts_view"
PermUsersContractsManage Permission = "users:contracts_manage"
```

Both added to `AllPermissions()` and to the `"users"` `PermissionGroup`.

### Seeding (merge-seed, idempotent)

| Role | `contracts_view` | `contracts_manage` |
|---|---|---|
| Super Admin | ✅ | ✅ |
| Admin | ✅ | ✅ |
| HR Manager | ✅ | ✅ |
| Manager | ✅ | — |
| Employee | ✅ | — |

### Routes

All routes are registered inside the existing JWT-protected group. Router-level gate: `RequirePerms(PermUsersContractsView)`. Write operations additionally check `PermUsersContractsManage` inside the handler.

```
GET    /api/v1/users/:user_id/contracts                    List (paginated + filtered)
POST   /api/v1/users/:user_id/contracts                    Create
GET    /api/v1/users/:user_id/contracts/:id                Get single
PATCH  /api/v1/users/:user_id/contracts/:id                Update (partial)
DELETE /api/v1/users/:user_id/contracts/:id                Soft delete
POST   /api/v1/users/:user_id/contracts/:id/attachment     Upload attachment → S3
```

---

## 3. DTOs — `internal/dto/user_contract.go`

```go
// ---- Request ----

type UserContractCreate struct {
    ContractType  string     `json:"contract_type"  binding:"required,oneof=labour_contract"`
    SignedDate    time.Time  `json:"signed_date"    binding:"required"`
    ExpiryDate    *time.Time `json:"expiry_date"`
    IsEndless     bool       `json:"is_endless"`
    AttachmentURL *string    `json:"attachment_url"`
}

type UserContractUpdate struct {
    ContractType  *string    `json:"contract_type"  binding:"omitempty,oneof=labour_contract"`
    SignedDate    *time.Time `json:"signed_date"`
    ExpiryDate    *time.Time `json:"expiry_date"`
    IsEndless     *bool      `json:"is_endless"`
    AttachmentURL *string    `json:"attachment_url"` // nil = no change; "" = remove
}

type UserContractListQuery struct {
    Page        int        `form:"page"`
    PageSize    int        `form:"page_size"`
    SignedFrom  *time.Time `form:"signed_from"`
    SignedTo    *time.Time `form:"signed_to"`
    ExpiryFrom  *time.Time `form:"expiry_from"`
    ExpiryTo    *time.Time `form:"expiry_to"`
}

// ---- Response ----

type UserContractRead struct {
    ID            uuid.UUID  `json:"id"`
    ContractType  string     `json:"contract_type"`
    SignedDate    time.Time  `json:"signed_date"`
    ExpiryDate    *time.Time `json:"expiry_date"`    // nil when endless
    IsEndless     bool       `json:"is_endless"`
    AttachmentURL *string    `json:"attachment_url"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}

type UserContractAttachmentResponse struct {
    AttachmentURL string `json:"attachment_url"`
}
```

**Partial PATCH semantics:** `UserContractUpdate` pointer fields — `nil` = leave unchanged. `AttachmentURL = ""` (empty string, non-nil pointer) = remove the attachment.

---

## 4. Repository — `internal/repositories/user_contract_repo.go`

### Interface

```go
type UserContractRepository interface {
    List(ctx context.Context, employeeID uuid.UUID, q dto.UserContractListQuery) ([]models.UserContract, int64, error)
    Get(ctx context.Context, id uuid.UUID) (*models.UserContract, error)
    Create(ctx context.Context, c *models.UserContract) error
    Update(ctx context.Context, c *models.UserContract) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Implementation notes

- **`List`**: base query `WHERE employee_id = ? AND is_deleted = false`. Appends `AND signed_date >= ?` / `AND signed_date <= ?` when `SignedFrom`/`SignedTo` non-nil. Appends `AND expiry_date >= ?` / `AND expiry_date <= ?` when `ExpiryFrom`/`ExpiryTo` non-nil — SQL `NULL` rows (endless contracts) are excluded by these conditions naturally. Order: `signed_date DESC`. Returns page slice + `COUNT(*)`.
- **`Get`**: fetch by `id WHERE is_deleted = false`. Does not filter by `employee_id` — service enforces ownership.
- **`Delete`**: sets `is_deleted = true`, `deleted_at = now()`.
- **`Update`**: saves full model struct (service copies DTO fields onto fetched model before calling).

---

## 5. Service — `internal/services/user_contract_service.go`

### Struct

```go
type UserContractService struct {
    repo         repositories.UserContractRepository
    employeeRepo repositories.EmployeeRepository
    storage      *config.StorageConfig
}

func NewUserContractService(
    repo repositories.UserContractRepository,
    employeeRepo repositories.EmployeeRepository,
    storage *config.StorageConfig,
) *UserContractService
```

### Common preamble (all methods)

```
emp, err := s.employeeRepo.FindByUserID(ctx, userID)
if err / emp == nil → ErrNotFound("employee profile not found")
```

For Get / Update / Delete: additionally fetch contract, check `is_deleted`, check `contract.EmployeeID == emp.ID` → `ErrNotFound("contract not found")` on any mismatch (no information leak).

### Business logic

**Create:**
1. Common preamble (user → employee)
2. If `req.IsEndless = true` → set `ExpiryDate = nil`
3. If `req.IsEndless = false` → `ExpiryDate` required → `ErrBadRequest("expiry date is required")`; must be strictly after `SignedDate` → `ErrBadRequest("expiry date must be after signed date")`
4. Build model, call `repo.Create`

**Update (partial PATCH):**
1. Common preamble + fetch + ownership check
2. Apply non-nil DTO fields onto the fetched model
3. Re-run the IsEndless / ExpiryDate invariant on the resulting model state
4. `AttachmentURL == ""` → set model `AttachmentURL = nil`; `AttachmentURL == nil` → leave unchanged
5. Call `repo.Update`

**Delete:**
1. Common preamble + fetch + ownership check
2. `repo.Delete`

**UploadAttachment:**
1. Common preamble + fetch + ownership check
2. Validate MIME type (PDF/PNG/JPG/DOCX) + size ≤ 5 MB → `ErrBadRequest` on violation
3. Upload to S3 key `contracts/{contractID}/{filename}` via storage config
4. Set `contract.AttachmentURL = &url`, call `repo.Update`
5. Return `UserContractAttachmentResponse{AttachmentURL: url}`

### Error mapping

| Condition | Error factory |
|---|---|
| User has no employee profile | `ErrNotFound("employee profile not found")` |
| Contract not found / deleted / wrong employee | `ErrNotFound("contract not found")` |
| Expiry date missing (is_endless=false) | `ErrBadRequest("expiry date is required")` |
| Expiry date not strictly after signed date | `ErrBadRequest("expiry date must be after signed date")` |
| Attachment wrong MIME type | `ErrBadRequest("attachment must be PDF, PNG, JPG, or DOCX")` |
| Attachment exceeds 5 MB | `ErrBadRequest("attachment must not exceed 5 MB")` |

---

## 6. Handler — `internal/handlers/user_contract_handler.go`

```go
type UserContractHandler struct {
    svc *services.UserContractService
}
```

Five Gin handlers: `List`, `Get`, `Create`, `Update`, `Delete`, `UploadAttachment`.

Pattern:
- Parse `user_id` from path (existing `parseIDParam` helper)
- Parse `id` from path for single-resource operations
- For write operations: check `hasContractsManage(c)` (walks JWT roles for `PermUsersContractsManage` or `PermAll`) → `ErrForbidden` if absent
- Bind request body / query
- Delegate to service
- Return `dto.Response[T]` envelope

`UploadAttachment` uses `c.FormFile("file")` — single multipart field.

---

## 7. Testing — `internal/services/user_contract_service_test.go`

Integration tests (skip without `TEST_DATABASE_URL`). Follow the pattern of `announcement_service_test.go`.

| Test | Coverage |
|---|---|
| `TestUserContract_Create_FixedTerm` | Happy path — row stored, expiry persisted |
| `TestUserContract_Create_Endless` | `is_endless=true` → `expiry_date=nil` in DB |
| `TestUserContract_Create_ExpiryBeforeSigned` | Returns 400 |
| `TestUserContract_Create_ExpiryEqualsSigned` | Returns 400 (strictly after) |
| `TestUserContract_Create_MissingExpiry_NotEndless` | Returns 400 |
| `TestUserContract_Update_RemoveAttachment` | `attachment_url=""` → nil in DB |
| `TestUserContract_Update_EndlessToggleOn` | Clears expiry on partial update |
| `TestUserContract_Update_EndlessToggleOff` | Re-enables expiry requirement |
| `TestUserContract_Delete_SoftDeletes` | Row has `is_deleted=true`; subsequent Get returns 404 |
| `TestUserContract_Get_WrongEmployee` | Returns 404 (no information leak) |
| `TestUserContract_List_FilterSignedDate` | Range inclusive; out-of-range rows excluded |
| `TestUserContract_List_ExpiryFilter_ExcludesEndless` | Endless rows absent from expiry-filtered results |
| `TestUserContract_List_CombinedFilters` | AND semantics across both filter types |
| `TestUserContract_List_Pagination` | Correct page slice + total count |

---

## 8. File Map

| Action | Path |
|---|---|
| Create | `migrations/000022_user_contracts.up.sql` |
| Create | `migrations/000022_user_contracts.down.sql` |
| Create | `internal/models/user_contract.go` |
| Create | `internal/dto/user_contract.go` |
| Create | `internal/repositories/user_contract_repo.go` |
| Create | `internal/services/user_contract_service.go` |
| Create | `internal/handlers/user_contract_handler.go` |
| Create | `internal/services/user_contract_service_test.go` |
| Modify | `internal/permissions/registry.go` (2 consts + AllPermissions + PermissionGroups) |
| Modify | `internal/handlers/router.go` (register 6 routes) |
| Modify | `cmd/server/main.go` (wire service + handler) |
| Modify | `internal/services/user_service.go` (merge-seed new perms) |

---

## Out of Scope (v1)

- Global cross-user contracts list
- Contract types other than Labour Contract
- Derived Active/Expired status badge
- Contract Number and Note fields
- Contract type filter
- Free-text search
- Overlap prevention / validation
- Approval workflow
- Contract renewal reminders
- Bulk operations
- Mobile / tablet layout
