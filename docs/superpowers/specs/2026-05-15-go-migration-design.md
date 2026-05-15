# Exnodes HRM API v2 — Migration Design (Python → Go)

| | |
|---|---|
| Status | Draft |
| Date | 2026-05-15 |
| Owner | danny.tranhoang@exnodes.vn |
| Source project | `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api` (Python FastAPI + MongoDB) |
| Target project | `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2` (Go + Gin + Postgres) |
| Reference stack | `/Users/sines/Documents/Work/exn-hr/Exn-hr/backend` |
| BA requirements | `exnodes-hrm-api-go-v2/ba-requirements/` |

---

## 1. Goal

Migrate the existing Python FastAPI + MongoDB HRM API to Go while expanding scope to cover the full Business Analysis (BA) requirement set for Exnodes HRM (8 epics: foundation, employee management, attendance & leave, payroll, recruitment, performance, training, organization data).

This is a **rewrite, not a refactor.** Source code and BA documents are inputs; output is a new Go codebase under `exnodes-hrm-api-go-v2/`.

## 2. Strategy

**Hybrid migration:**

- **Modules already in Python (12 modules)** → port 1:1, preserve behavior, change DB layer Mongo → Postgres.
- **Modules described only in BA docs (payroll, recruitment, performance, training, request-ticket, overtime, …)** → build new from BA spec.

API contract: same envelope (`{success, message, data}` + paginated `{items, total, page, page_size, total_pages}`), same path prefix (`/api/v1`), same auth model. Field-level adjustments allowed (e.g., `_id` ObjectID → `id` UUID). FE updates limited to small adjustments.

Data: greenfield. No MongoDB → Postgres data export needed.

## 3. Stack

| Concern | Choice |
|---|---|
| Language | Go 1.24 |
| Web framework | Gin (`github.com/gin-gonic/gin`) |
| ORM | GORM (`gorm.io/gorm` + `gorm.io/driver/postgres`) |
| Database | Postgres |
| Migrations | `golang-migrate/migrate/v4` (CLI + library) — versioned SQL files only |
| Auth | `golang-jwt/jwt/v5`, bcrypt via `golang.org/x/crypto/bcrypt` |
| File storage | Supabase Storage (S3-compatible) via `aws-sdk-go-v2` |
| API docs | `swaggo/swag` + `swaggo/gin-swagger` (Swagger UI at `/swagger/index.html`) |
| Config | `github.com/joho/godotenv` |
| Test | Standard `testing` + `testify` + real Postgres test DB |
| Module path | `github.com/exnodes/hrm-api` |

## 4. Layout

```
exnodes-hrm-api-go-v2/
├── ba-requirements/           # existing BA docs (untouched)
├── cmd/server/main.go         # entry point
├── internal/
│   ├── config/                # env, db, settings
│   ├── models/                # GORM models (one per entity)
│   ├── dto/                   # request/response shapes
│   ├── repositories/          # data access (interface + impl)
│   ├── services/              # business logic + tests
│   ├── handlers/              # Gin handlers + route registration
│   ├── middleware/            # JWT, RequirePerms, CORS, error, recovery
│   ├── permissions/           # permission constants + groups
│   ├── errors/                # AppError hierarchy
│   └── sse/                   # realtime hub for announcements/notifications
├── pkg/utils/                 # jwt, password, pagination, search escape, html
├── migrations/                # NNNNNN_<name>.up.sql / .down.sql
├── scripts/                   # seed.sh, migrate.sh
├── docs/                      # superpowers/specs/ + generated swagger docs
├── .env.example
├── go.mod / go.sum
├── Makefile                   # run, test, migrate-*, swag
└── README.md
```

## 5. Database & migrations

### 5.1 Migration tooling

- All schema changes are versioned SQL files via `golang-migrate`.
- Filename convention: `NNNNNN_<snake_name>.up.sql` + `.down.sql`. Sequential 6-digit prefix.
- Existing migration files are **immutable** — schema corrections come as new migrations.
- `db.AutoMigrate()` is **prohibited** in app code.
- Migrations execute via Makefile targets, never silently on app boot.
- On boot, app checks current DB version and **fails loud** if not up to date.

Makefile targets:
```
migrate-new name=...    # create empty up/down pair with next sequence
migrate-up              # apply all pending migrations
migrate-down            # rollback one step
migrate-version         # print current applied version
migrate-force version=  # only for fixing dirty state
```

### 5.2 Audit columns (required on every entity table)

```sql
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
deleted_at  TIMESTAMPTZ NULL
```

Plus a per-table `BEFORE UPDATE` trigger calling shared `set_updated_at()` function, plus index on `is_deleted`.

### 5.3 BaseModel

```go
// internal/models/base.go
type BaseModel struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
    UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
    IsDeleted bool       `gorm:"not null;default:false;index" json:"-"`
    DeletedAt *time.Time `json:"-"`
}
```

Embed in every entity. Soft-delete via custom scope:

```go
func NotDeleted(db *gorm.DB) *gorm.DB { return db.Where("is_deleted = ?", false) }
```

Delete operation sets both `is_deleted = TRUE, deleted_at = NOW()`. Restore reverses both.

### 5.4 ID strategy

- All primary keys = `UUID` (`gen_random_uuid()` from `pgcrypto`).
- No auto-increment sequences. Makes it easier to map from Mongo ObjectId semantics, and avoids ID enumeration in URLs.

### 5.5 Initial migration plan (Phase 1)

| # | Name | Contents |
|---|---|---|
| 000001 | init_extensions | `CREATE EXTENSION uuid-ossp, pgcrypto, citext` + `set_updated_at()` function |
| 000002 | create_roles_users | `roles` (permissions JSONB), `users`, `user_roles` join |
| 000003 | create_departments_positions | `departments`, `positions` |
| 000004 | create_attendance | `attendance` with late/halfday flags |
| 000005 | create_leave_requests | `leave_requests` + status enum + quota table |
| 000006 | create_announcements_labels | `announcements`, `labels`, `announcement_labels` |
| 000007 | create_skills | `skills`, `user_skills` |
| 000008 | create_system_config | `system_config` key-value |
| 000009 | create_audit_log | generic audit table |
| 000010 | seed_system_data | idempotent seed: 5 system roles + default depts/positions |

Phase 2 (BA-only modules) adds further migrations grouped per epic.

## 6. Auth & RBAC

### 6.1 JWT

- HS256, `JWT_SECRET_KEY` from env.
- Payload: `{sub: user_id, type: "access"|"refresh", exp, iat, roles: [...]}`.
- TTLs configurable per env.
- Sign/verify in `pkg/utils/jwt.go`.

### 6.2 Models

```
users        (id, email CITEXT UNIQUE, password_hash, full_name, department_id, position_id,
              is_active, email_changed_at, password_reset_at, ...audit)
roles        (id, name UNIQUE, description, is_system, permissions JSONB, ...audit)
user_roles   (user_id, role_id, composite PK, ...audit)
```

### 6.3 Permission registry

`internal/permissions/registry.go` exposes:

```go
type Permission string

const (
    PermAll              Permission = "*"
    PermAuthLogin        Permission = "auth:login"
    PermUsersRead        Permission = "users:read"
    PermUsersCreate      Permission = "users:create"
    PermUsersUpdate      Permission = "users:update"
    PermUsersDelete      Permission = "users:delete"
    PermUsersManageRoles Permission = "users:manage_roles"
    PermUsersChangePwd   Permission = "users:change_password"
    PermRolesRead        Permission = "roles:read"
    PermRolesCreate      Permission = "roles:create"
    PermRolesUpdate      Permission = "roles:update"
    PermRolesDelete      Permission = "roles:delete"
    PermDepartmentsRead   Permission = "departments:read"
    PermDepartmentsCreate Permission = "departments:create"
    PermDepartmentsUpdate Permission = "departments:update"
    PermDepartmentsDelete Permission = "departments:delete"
    PermPositionsRead    Permission = "positions:read"
    PermPositionsCreate  Permission = "positions:create"
    PermPositionsUpdate  Permission = "positions:update"
    PermPositionsDelete  Permission = "positions:delete"
    PermSkillsRead       Permission = "skills:read"
    PermSkillsCreate     Permission = "skills:create"
    PermSkillsUpdate     Permission = "skills:update"
    PermSkillsDelete     Permission = "skills:delete"
    PermLeaveRead        Permission = "leave_requests:read"
    PermLeaveCreate      Permission = "leave_requests:create"
    PermLeaveUpdate      Permission = "leave_requests:update"
    PermLeaveDelete      Permission = "leave_requests:delete"
    PermLeaveApprove     Permission = "leave_requests:approve"
    PermLeaveCancel      Permission = "leave_requests:cancel"
    PermLeaveManage      Permission = "leave_requests:manage"
    PermLeaveQuotaManage Permission = "leave_quota:manage"
    PermAttendanceRead   Permission = "attendance:read"
    PermAttendanceManage Permission = "attendance:manage_data"
    PermOrgSettings      Permission = "organization_settings:manage"
    PermAnnounceManage   Permission = "announcements:manage"
)

// PermissionGroups feeds GET /api/v1/roles/permissions for FE picker
var PermissionGroups = []PermissionGroup{ ... }
```

Phase 2 epics extend the registry with module-specific permissions.

### 6.4 Per-route permission binding

Each endpoint declares its required permissions inline at the route definition — mirroring FastAPI's `dependencies=[Depends(require_permissions(...))]` pattern.

```go
authed := v1.Group("")
authed.Use(middleware.JWT())

// Self-service (any authenticated user)
authed.GET("/users/me", h.User.GetMe)
authed.PATCH("/users/me", h.User.UpdateMe)

// Permission-gated
users := authed.Group("/users")
users.GET("",        middleware.RequirePerms(perm.PermUsersRead),         h.User.List)
users.POST("",       middleware.RequirePerms(perm.PermUsersCreate),       h.User.Create)
users.GET("/:id",    middleware.RequirePerms(perm.PermUsersRead),         h.User.Get)
users.PATCH("/:id",  middleware.RequirePerms(perm.PermUsersUpdate),       h.User.Update)
users.DELETE("/:id", middleware.RequirePerms(perm.PermUsersDelete),       h.User.Delete)
users.PUT("/:id/roles", middleware.RequirePerms(perm.PermUsersManageRoles), h.User.AssignRoles)
```

`RequirePerms` semantics: all listed permissions must be in the user's effective set (union of role.permissions across user's roles). Wildcard `*` bypasses.

### 6.5 Seed system data

`services/seed_service.go` runs idempotently on app boot:

- 5 system roles: Super Admin (`["*"]`), Admin, HR Manager, Manager, Employee — `is_system=true`.
- 1 super admin user from env credentials.
- Default departments and positions.

System roles cannot be deleted or renamed.

### 6.6 Errors

```go
// internal/errors/errors.go
type AppError struct {
    Code    string         // "not_found", "bad_request", "conflict", "forbidden", "unauthorized"
    Message string
    HTTP    int
    Details map[string]any
}

func ErrNotFound(resource string) *AppError    { ... }
func ErrBadRequest(msg string) *AppError       { ... }
func ErrConflict(msg string) *AppError         { ... }
func ErrForbidden(msg string) *AppError        { ... }
func ErrUnauthorized(msg string) *AppError     { ... }
```

`middleware.ErrorHandler()` converts `*AppError` to JSON `{success: false, message, code, details}`.

## 7. Layered pattern

### 7.1 Handler
Parse/validate request, call service, wrap response. No DB access, no business logic.

### 7.2 Service
Business logic, cross-entity rules. Raises `*AppError`. Receives `context.Context`. Can call multiple repositories.

### 7.3 Repository
Pure data access (GORM). Interface + implementation for mockability. Applies `NotDeleted` scope by default. Returns `gorm.ErrRecordNotFound`; service converts to `AppError`.

### 7.4 Response envelopes

```go
// internal/dto/response.go
type Response[T any] struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    Data    T      `json:"data,omitempty"`
}

type PaginatedData[T any] struct {
    Items      []T   `json:"items"`
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
}
```

### 7.5 Search

Case-insensitive search uses `ILIKE` with escaped `%` and `_`. Helper in `pkg/utils/search.go`.

## 8. Testing

- Per-service test files: `services/*_service_test.go`.
- Shared `services/testhelper_test.go` sets up a real Postgres test database (`exnodes_hrm_test`), runs migrations, provides factory helpers (`makeAdmin`, `makeReadonly`, `makeDepartment`, etc.) and per-test table cleanup.
- Handler tests optional — coverage achieved through service tests.

## 9. Phasing & per-phase verification

### 9.1 Phase list

| # | Name | Scope |
|---|---|---|
| 0 | Foundation infrastructure | go.mod, layout, config, DB connect, migrate CLI, Gin skeleton, error middleware, BaseModel, Swagger setup, Makefile, .env.example, `/health` |
| 1 | Auth + RBAC core | migrations for roles/users/user_roles, JWT, password, login/refresh/logout, JWT middleware, RequirePerms middleware, permission registry, seed |
| 2 | Users module | admin CRUD, `/users/me`, change password/email, role assignment, avatar upload (S3), device tokens, leave quota, notification settings |
| 3 | Departments + Positions | full CRUD + filters |
| 4 | Skills + Labels | full CRUD |
| 5 | Leave Requests + Quota | request flow, approve/cancel, quota tracking |
| 6 | Attendance | check-in/out, manage_data, list with filters |
| 7 | Announcements + Mobile Announcements | CRUD + SSE for realtime push |
| 8 | Organization Settings + System Config | key-value config CRUD |
| 9 | Email + Invite + Push Notification | SMTP, invite flow, FCM/web push |
| — | end of Python-port phase | |
| 10 | EP-003 Request Ticket | BA-spec build |
| 11 | EP-004 Overtime | BA-spec build |
| 12 | Payroll | BA-spec build |
| 13 | Recruitment | BA-spec build |
| 14 | Performance Management | BA-spec build |
| 15 | Training & Development | BA-spec build |

### 9.2 Definition of Done (every phase)

1. Migration files committed (up + down).
2. Models, repository, service, handler, routes wired.
3. Swagger annotations on every new handler — endpoint visible on `/swagger/index.html`.
4. Service-level tests pass (`go test ./internal/services/...`).
5. `make migrate-up` succeeds on a clean DB.
6. `make run` boots the server with no error.
7. **Self end-to-end verification**: server is running and the agent must call the new endpoints to walk:
   - The phase's main happy-path flow (e.g., Phase 2: login → create user → list → get → update → delete → list).
   - At least one error path (auth, missing permission, not-found, conflict).
   - DB state spot-check.
8. Verification log (commands + key responses) committed alongside phase.
9. README updated with new endpoints.

### 9.3 Out of scope (initial release, deferred)

- i18n / localization (per BA scope decision).
- AI-based recommendations or analytics.
- Mobile native app (web-responsive only).
- Tax filing or benefits administration.
- Time tracking for project/billing purposes.
- MongoDB → Postgres data migration script (greenfield).

## 10. Risks & mitigations

| Risk | Mitigation |
|---|---|
| Document → relational schema mismatch (embedded docs, dynamic fields) | Per-module schema design step before migration is authored. Use JSONB sparingly for genuinely dynamic fields (e.g., `roles.permissions`). |
| FE breakage from envelope/field-name changes | Stay within the agreed "same shape" envelope. Document any non-trivial field change in the phase deliverable. |
| Per-phase scope creep into BA-only modules during Phase 1 | Hard cut: Phase 1 only ports existing Python modules. BA-only modules start at Phase 10. |
| Migration files mistakenly edited after release | Reviewer checklist: only **append** new files; never modify existing `*.up.sql`/`*.down.sql`. |
| Forgetting `is_deleted` filter in queries → leaks soft-deleted rows | Default `NotDeleted` scope on every list query in repositories. Explicit `Unscoped()` only where intentional. |

## 11. Open items (to be addressed during planning)

- Specific permission constants for Phase 2 BA-only modules — pulled from each epic's REQUIREMENTS.md when reached.
- Exact migration files for each BA-only module — designed at the start of that phase, not now.
- SMTP provider selection (carryover from Python `.env`).
- Push notification provider — FCM key location, web-push keypair handling.

## 12. Change log

| Version | Date | Changes | Author |
|---|---|---|---|
| 0.1 | 2026-05-15 | Initial draft from brainstorming session | Claude + danny.tranhoang |
