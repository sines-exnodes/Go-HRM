# Code Map — Where to Find Things

A pointer-based map to the source tree. For symbol-level lookup (functions, types, references), prefer Serena's `find_symbol` / `get_references` over `grep`.

## Source-tree layout

```
cmd/server/main.go             entry point + Swagger title/annotations + ALL route wiring
internal/
  config/                      env loader, GORM connect, boot-time migration version assert, storage config
  models/                      BaseModel + per-entity GORM models (one file per entity)
  dto/                         request/response envelopes; the validation boundary
  repositories/                GORM data access — interface + impl per entity
  services/                    business logic; returns *errors.AppError
  handlers/                    Gin handlers + router.go (RegisterRoutes is mostly health-only;
                               actual route wiring lives in cmd/server/main.go)
  middleware/                  CORS, Recovery, ErrorHandler, JWT auth (RequirePerms)
  permissions/                 Permission constant registry + groups + IsValid
  errors/                      AppError type + factory helpers (apperrors.ErrNotFound, etc.)
  sse/                         realtime event hub (Phase 7+; doesn't exist yet)
pkg/utils/                     generic shared helpers (password.go, jwt.go, BuildILIKEPattern, ...)
migrations/                    golang-migrate SQL files: NNNNNN_<name>.up/down.sql
docs/
  superpowers/                 specs, plans, verification logs, CHECKPOINT.md
  swagger/                     GENERATED — do NOT hand-edit (regen via `make swag`)
ba-requirements/               BA-produced requirement docs (Vietnamese + English technical terms)
```

## The non-negotiable layering rule

```
handler  →  service  →  repository  →  GORM
```

One-directional. Specifically:
- Handlers **never** touch the DB directly. They call services and write `dto.Response[T]` envelopes.
- Services **never** import `gin`. They take `context.Context` and return `*errors.AppError` (which the `ErrorHandler` middleware renders).
- Repositories **expose interfaces** so services are unit-testable.
- DTOs are the validation boundary. Self-service updates use a field-by-field whitelist copy — fields absent from the DTO are silently un-updatable by design.

## Where to find domain modules (by entity)

| Entity | Migration | Model | Repo | Service | Handler |
|---|---|---|---|---|---|
| Auth (login/refresh/logout) | `000001_init_extensions` (extensions only); JWT keys via env | — (claims) | — | `auth_service.go` | `auth_handler.go` |
| Roles + permissions | `000002_create_roles_users` | `role.go` | `role_repo.go` | (seed only — `seed_service.go`) | `role_handler.go` (list-only) |
| Users (auth) | `000002_create_roles_users` | `user.go` | `user_repo.go` | `user_service.go` | `user_handler.go` |
| Employees (HR profile) | `000003_create_employees_dependents` | `employee.go` | `employee_repo.go` | `employee_service.go` | `employee_handler.go` |
| Dependents | `000003_…` | `dependent.go` | `dependent_repo.go` | `dependent_service.go` | `dependent_handler.go` |
| Device tokens / notification settings / leave quotas | `000004_phase2_extras` | `device_token.go` / `user_notification_settings.go` / `employee_leave_quota.go` | `device_token_repo.go` / `notification_settings_repo.go` / `leave_quota_repo.go` | (mostly via `user_service` / `employee_service`) | (admin endpoints inside `user_handler` / `employee_handler`) |
| Departments + positions | `000005_create_departments_positions` | `department.go` / `position.go` | `department_repo.go` / `position_repo.go` | `department_service.go` / `position_service.go` | `department_handler.go` / `position_handler.go` |
| Skills + employee_skills | `000006_create_skills` | `skill.go` / `employee_skill.go` | `skill_repo.go` / `employee_skill_repo.go` | `skill_service.go` | `skill_handler.go` |
| Announcement labels | `000007_create_labels` | `label.go` | `label_repo.go` | `label_service.go` | `label_handler.go` |
| **Leave requests** (Phase 5) | `000008_create_leave_requests` | `leave_request.go` | `leave_request_repo.go` | `leave_service.go` | `leave_handler.go` |

## Conventions to match (AGENTS.md Rule 11)

- **Soft delete** uses the custom `models.NotDeleted` GORM scope. NEVER use GORM's built-in `gorm.DeletedAt`.
- **Migrations** are versioned SQL only (`make migrate-new name=<snake>`). NEVER `db.AutoMigrate()`. The server asserts the applied migration version on boot.
- **Audit columns** on every entity table: `created_at`, `updated_at`, `is_deleted BOOLEAN`, `deleted_at TIMESTAMPTZ`, plus a `BEFORE UPDATE` trigger calling `set_updated_at()`.
- **PKs** are UUIDs via `gen_random_uuid()` (pgcrypto extension; see migration 000001).
- **Errors** flow as `*errors.AppError` from services; the `ErrorHandler` middleware renders them. Don't write ad-hoc `c.JSON(400, gin.H{...})` for errors.
- **Repos** expose an interface (`type XxxRepository interface { ... }`) + a lowercase impl (`type xxxRepository struct { db *gorm.DB }`) + a constructor (`func NewXxxRepository(db *gorm.DB) XxxRepository`).
- **Index naming**: `idx_<table>_<col>` for indexes, `uq_<table>_<predicate>_live` for partial unique, `trg_<table>_set_updated_at` for triggers. The Phase 5 migration drift away from this (e.g. `leave_requests_<col>_idx` in the plan draft) was corrected before commit.
- **Multipart uploads** (avatar / skill icon / leave attachment) mandate `http.DetectContentType` sniffing + an allowlist. Client `Content-Type` is treated as a hint only (review-fix #2 pattern from Phase 2).
- **Permission gating**: route-level via `middleware.RequirePerms(authSvc, perms...)` (with `authSvc *AuthService` as the FIRST arg); service-level ownership via an `asAdmin bool` passed in by the handler (precomputed from JWT-preloaded `user.Roles`). Two distinct guards by design.

## Schema split

`users` (auth) ⟂ `employees` (HR profile) ⟂ `dependents` ⟂ `employee_skills` ⟂ `employee_leave_quotas` ⟂ `leave_requests`. **Every cross-aggregate FK from Phase 2 onward targets `employees(id)`, NOT `users(id)`** — the Go schema split puts HR-domain references on the employee row. The Python source mostly used `users(id)`; do not copy that pattern.

## Common quick-lookup queries

- "Where does X permission live?" → `internal/permissions/registry.go` (constants) + `internal/services/seed_service.go` (role assignments)
- "What does the API return for entity X?" → `internal/dto/x.go` (XCreate, XUpdate, XRead, XListQuery)
- "How is X soft-deleted?" → repo's `SoftDelete()` writes `is_deleted=true, deleted_at=NOW()` via GORM `Updates(...)`; reads chain `.Scopes(models.NotDeleted)`
- "How are routes wired?" → `cmd/server/main.go`, the big `v1 := r.Group("/api/v1")` block (NOT `internal/handlers/router.go` — that's stale)
- "Why was X migrated this way?" → start at `docs/superpowers/plans/2026-05-15-phase-NN-*.md` REVISION NOTES, then the corresponding `docs/superpowers/verification/phase-NN.md`
