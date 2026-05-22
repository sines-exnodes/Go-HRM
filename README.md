# Exnodes HRM API v2 (Go)

Go + Gin + Postgres rewrite of the Exnodes HRM API. See
[`docs/superpowers/specs/2026-05-15-go-migration-design.md`](docs/superpowers/specs/2026-05-15-go-migration-design.md)
for the full migration design and phase plan.

This README documents **Phase 0 only**: a boot-able skeleton with `/health` and Swagger UI.
Subsequent phases (auth, users, departments, ...) plug into the structures defined here.

## Quickstart

### 1. Prerequisites

- Go **1.24** (`go version`)
- Postgres **14+** running locally (or a remote DSN you control)
- The `swag` and `migrate` CLIs (installed below)

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
export PATH="$(go env GOPATH)/bin:$PATH"
```

### 2. Clone & install deps

```bash
git clone <repo-url> exnodes-hrm-api-go-v2
cd exnodes-hrm-api-go-v2
go mod download
```

### 3. Configure env

```bash
cp .env.example .env
$EDITOR .env   # set DB_* (or DATABASE_URL) to a Postgres you can write to
```

### 4. Create the database

```bash
createdb exnodes_hrm   # or use psql: CREATE DATABASE exnodes_hrm;
```

### 5. Apply migrations

```bash
make migrate-up
```
Expected output ends with a version line; nothing fails.

### 6. Run the server

```bash
make run
```
Expected log lines:
```
exnodes-hrm-api listening on :8080 (env=development, swagger=true)
```

### 7. Smoke-test

```bash
curl -s http://localhost:8080/health | jq
```
Expected:
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "service": "exnodes-hrm-api"
  }
}
```

Then open `http://localhost:8080/swagger/index.html` and confirm `GET /health`
appears under the `system` tag.

## Make targets

| Target | What it does |
|---|---|
| `make run` | Run the API server (`go run ./cmd/server`) |
| `make build` | Build `./bin/server` |
| `make test` | Run all Go tests |
| `make tidy` | `go mod tidy` |
| `make fmt` | `gofmt -s -w .` |
| `make vet` | `go vet ./...` |
| `make swag` | Regenerate Swagger docs into `docs/swagger/` |
| `make migrate-new name=<snake>` | Create a new empty up/down migration pair |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back one migration step |
| `make migrate-version` | Print the currently applied migration version |
| `make migrate-force version=N` | Force the version (only to fix a dirty state) |

## Project layout

```
cmd/server/         Entry point (main.go)
internal/
  config/           Env loader, GORM connect, migration version check
  models/           BaseModel + per-entity models (added in later phases)
  dto/              Request/response envelopes
  repositories/     GORM data access (added in later phases)
  services/         Business logic (added in later phases)
  handlers/         Gin handlers + RegisterRoutes
  middleware/       CORS, Recovery, ErrorHandler (+ JWT in Phase 1)
  permissions/      Permission registry (added in Phase 1)
  errors/           AppError type + factory helpers
  sse/              Realtime event hub (added in Phase 7)
pkg/utils/          Generic helpers shared across modules
migrations/         golang-migrate SQL files (NNNNNN_<name>.up/down.sql)
scripts/            Shell helpers (seed, deploy, etc.)
docs/
  superpowers/      Project specs, plans, verification logs
  swagger/          Generated OpenAPI artefacts (do not hand-edit)
```

## Schema conventions (enforced from Phase 1 onward)

- Every entity table has 4 audit columns: `created_at`, `updated_at`,
  `is_deleted BOOLEAN`, `deleted_at TIMESTAMPTZ` plus a per-table
  `BEFORE UPDATE` trigger calling `set_updated_at()`.
- Primary keys are UUIDs via `gen_random_uuid()` (pgcrypto).
- Soft delete is implemented via the custom `NotDeleted` scope — NOT GORM's
  built-in `gorm.DeletedAt`.
- Schema changes are versioned SQL migration files only. `db.AutoMigrate()`
  is **prohibited**. The app verifies migration version on boot and refuses
  to start if behind or dirty.

## Phase 2 — Employees + Dependents module endpoints

Self-service (JWT only):

| Method | Path                                       | Purpose                                       |
|--------|--------------------------------------------|-----------------------------------------------|
| GET    | `/api/v1/users/me`                         | auth profile + embedded employee summary      |
| POST   | `/api/v1/users/me/change-password`         | change own password                           |
| POST   | `/api/v1/users/me/change-email`            | change own email (reauth)                     |
| POST   | `/api/v1/users/me/device-tokens`           | register FCM/APNs token                       |
| DELETE | `/api/v1/users/me/device-tokens/:token`    | remove a device token                         |
| PATCH  | `/api/v1/users/me/notification-settings`   | toggle push notifications                     |
| GET    | `/api/v1/employees/me`                     | own HR profile (with dependents)              |
| PATCH  | `/api/v1/employees/me`                     | update own HR profile (restricted whitelist)  |
| PATCH  | `/api/v1/employees/me/avatar`              | upload own avatar (multipart, 5MB image only) |

Admin (per-route perms):

| Method | Path                                              | Required perm              |
|--------|---------------------------------------------------|----------------------------|
| GET    | `/api/v1/employees`                               | `employees:read`           |
| POST   | `/api/v1/employees`                               | `employees:create`         |
| GET    | `/api/v1/employees/:id`                           | `employees:read`           |
| PATCH  | `/api/v1/employees/:id`                           | `employees:update`         |
| DELETE | `/api/v1/employees/:id`                           | `employees:delete`         |
| PATCH  | `/api/v1/employees/:id/avatar`                    | `employees:update`         |
| PATCH  | `/api/v1/employees/:id/leave-quota`               | `leave_quota:manage`       |
| PATCH  | `/api/v1/users/:id`                               | `users:update`             |
| DELETE | `/api/v1/users/:id`                               | `users:delete`             |
| PATCH  | `/api/v1/users/:id/change-password`               | `users:change_password`    |
| PUT    | `/api/v1/users/:id/roles`                         | `users:manage_roles`       |

Dependents — owner OR `dependents:manage` (enforced in handler; the employee
segment uses `:id` and the nested dependent uses `:dependentID`):

| Method | Path                                                       |
|--------|------------------------------------------------------------|
| GET    | `/api/v1/employees/:id/dependents`                         |
| POST   | `/api/v1/employees/:id/dependents`                         |
| PATCH  | `/api/v1/employees/:id/dependents/:dependentID`            |
| DELETE | `/api/v1/employees/:id/dependents/:dependentID`            |

Self-service `PATCH /employees/me` whitelist:
`phone, personal_email, permanent_address, current_address, marital_status,
emergency_contact_name, emergency_contact_relation, emergency_contact_phone`.
Any other field is **silently rejected at the DTO boundary** —
`EmployeeSelfUpdate` has no field for it, and the service applies a manual
field-by-field copy from the DTO only. Verified by direct SQL in
`docs/superpowers/verification/phase-02.md` (a `basic_salary`/`department_id`
sent to `PATCH /employees/me` does not change the stored row).

Employee creation auto-assigns the seeded **"Employee"** role (carries
`auth:login`) when the admin supplies no `role_ids`, so every created
employee is a usable self-service account.

## Phase 3 — Departments + Positions module endpoints

### Departments

| Method | Path                     | Permission          |
|--------|--------------------------|---------------------|
| GET    | /api/v1/departments      | departments:read    |
| POST   | /api/v1/departments      | departments:create  |
| GET    | /api/v1/departments/{id} | departments:read    |
| PATCH  | /api/v1/departments/{id} | departments:update  |
| DELETE | /api/v1/departments/{id} | departments:delete  |

Self-referential `parent_id` (UUID or `"root"` filter on list). Delete is
blocked (409) while child departments, active positions, or assigned
employees exist.

### Positions

| Method | Path                   | Permission        |
|--------|------------------------|-------------------|
| GET    | /api/v1/positions      | positions:read    |
| POST   | /api/v1/positions      | positions:create  |
| GET    | /api/v1/positions/{id} | positions:read    |
| PATCH  | /api/v1/positions/{id} | positions:update  |
| DELETE | /api/v1/positions/{id} | positions:delete  |

Each position belongs to exactly one department. Delete is blocked (409)
while employees are assigned. The `employees.department_id` /
`employees.position_id` FK constraints (deferred from migration 000003) are
added in migration 000005.

## Phase 6 — Attendance module endpoints

### Attendance

| Method | Path                          | Permission                                | Description                                       |
|--------|-------------------------------|-------------------------------------------|---------------------------------------------------|
| POST   | /api/v1/attendance/check-in   | authenticated                             | Record a check-in (creates the day row if absent) |
| POST   | /api/v1/attendance/check-out  | authenticated                             | Close the open session                            |
| GET    | /api/v1/attendance/today      | authenticated                             | Today's status + monthly count + streak           |
| GET    | /api/v1/attendance/me         | authenticated                             | List my own attendance rows                       |
| GET    | /api/v1/attendance            | attendance:read (manage_data sees all)    | List rows (admin: all, employee: own only)        |
| GET    | /api/v1/attendance/matrix     | attendance:read (manage_data sees all)    | Monthly attendance matrix                         |
| GET    | /api/v1/attendance/{id}       | attendance:read (owner or admin)          | Get a specific row                                |
| POST   | /api/v1/attendance            | attendance:manage_data                    | Admin manual create                               |
| PATCH  | /api/v1/attendance/{id}       | attendance:manage_data                    | Admin update                                      |
| DELETE | /api/v1/attendance/{id}       | attendance:manage_data                    | Admin soft-delete                                 |

Two-table design: `attendance` (one row per `(employee_id, date)`) + N
child `attendance_sessions`. `is_late` is computed once from the FIRST
check-in vs `LATE_THRESHOLD_HOUR:LATE_THRESHOLD_MINUTE` in
`COMPANY_TIMEZONE`; subsequent sessions don't recompute it. The partial
unique index `uq_attendance_sessions_one_open` and a service-level
`FindOpenSession()` guard prevent overlapping open sessions. Non-admin
callers of `GET /attendance` are silently scoped to own rows (Python
contract — managers see all, non-managers see only their own row).

## Phase 7 — Announcements + Mobile + SSE realtime

### Announcements (web)

| Method | Path                                  | Permission              | Description                                     |
|--------|---------------------------------------|-------------------------|-------------------------------------------------|
| GET    | /api/v1/announcements                 | authenticated           | List rows (admin: all; non-admin: visible only) |
| GET    | /api/v1/announcements/{id}            | authenticated           | Get one (403 if not visible)                    |
| POST   | /api/v1/announcements/{id}/view       | authenticated           | Mark as viewed (idempotent — preserves 1st time)|
| POST   | /api/v1/announcements                 | announcements:manage    | Create (optional `status=published` publishes)  |
| PATCH  | /api/v1/announcements/{id}            | announcements:manage    | Update (owner or admin)                         |
| DELETE | /api/v1/announcements/{id}            | announcements:manage    | Soft-delete                                     |
| POST   | /api/v1/announcements/{id}/publish    | announcements:manage    | Publish (no-op if already; broadcasts via SSE)  |

### Mobile

| Method | Path                                       | Permission    | Description                                  |
|--------|--------------------------------------------|---------------|----------------------------------------------|
| GET    | /api/v1/mobile/announcements               | authenticated | Visibility-filtered list (Body omitted)      |
| GET    | /api/v1/mobile/announcements/{id}          | authenticated | Detail (with Body + attachments)             |

### SSE

| Method | Path                                  | Permission    | Description                                            |
|--------|---------------------------------------|---------------|--------------------------------------------------------|
| GET    | /api/v1/sse/announcements             | authenticated | Long-lived event stream; emits `announcement_published`|

The SSE endpoint accepts the JWT via `Authorization: Bearer …` OR
`?token=…` because EventSource cannot set headers. Token-in-query may
appear in proxy logs — scrub at the reverse proxy and use short-lived
access tokens. Single-process in-memory hub: scaling beyond 1 replica
requires a Redis pub/sub backplane (see `internal/sse/hub.go`).

Visibility predicate (non-admin):

- `status='published'` AND `is_deleted=false` AND
- One of: `author_id == current_employee.id` OR `target_audience='all'`
  OR (`target_audience='department'` AND the announcement targets the
  user's department).

Admins (`announcements:manage`) see everything regardless of status /
audience. Author display reads from `employees.full_name` because
`author_id` references `employees(id)` per the schema split (REVISION
NOTES item #2).

## Phase 8 — Organization Settings (system_config singleton)

### Organization settings

| Method | Path                                                | Permission                  | Description                              |
|--------|-----------------------------------------------------|-----------------------------|------------------------------------------|
| GET    | /api/v1/organization-settings/attendance            | organization_settings:manage | Read attendance thresholds              |
| PATCH  | /api/v1/organization-settings/attendance            | organization_settings:manage | Partial update (pointer fields)         |
| GET    | /api/v1/organization-settings/company-profile       | authenticated               | Read company address + lat/lng          |
| PATCH  | /api/v1/organization-settings/company-profile       | organization_settings:manage | Update address + stamps updated_by/at   |

Backed by a single-row `system_config` table whose PK is the sentinel
UUID `00000000-0000-0000-0000-000000000001`. A `CHECK (id = '…0001')`
constraint at the DB level enforces the singleton invariant; the seed
service runs `INSERT … ON CONFLICT DO NOTHING` on boot. Soft-delete
columns exist for schema parity but are never written.

`company_address_updated_by` references `employees(id) ON DELETE SET
NULL` per the Go schema split. The GET projection resolves
`updated_by_name` from `Employee.FullName` (best-effort — falls back
to nil when the employee row is gone).
