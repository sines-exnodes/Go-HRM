# Conventions — Non-Obvious, Project-Wide Rules

Match these even if you disagree (AGENTS.md Rule 11). Surface dissent; do not fork silently.

## Schema / migrations

- Versioned SQL only via `golang-migrate`. **NEVER** `db.AutoMigrate()` — boot-time assert in `internal/config` refuses to start if applied version is behind or dirty.
- Every entity table carries the four audit columns: `created_at`, `updated_at`, `is_deleted BOOLEAN NOT NULL DEFAULT false`, `deleted_at TIMESTAMPTZ`. Plus a per-table `BEFORE UPDATE` trigger `trg_<table>_set_updated_at` calling `set_updated_at()`.
- PKs: UUIDs via `gen_random_uuid()` (pgcrypto, enabled in migration `000001_init_extensions`).
- Index naming: `idx_<table>_<col>` / partial unique `uq_<table>_<predicate>_live` / trigger `trg_<table>_set_updated_at`. Phase 5 plan draft used `leave_requests_<col>_idx` — that was corrected before commit; do not reintroduce.
- Cross-aggregate FKs from Phase 2 onward target `employees(id)`, NEVER `users(id)`. (The Python source mostly used `users(id)`. Do not copy.)

## Soft delete

- Reads chain `.Scopes(models.NotDeleted)` — the custom scope.
- Deletes write `is_deleted=true, deleted_at=NOW()` via GORM `Updates(...)`.
- **NEVER** use GORM's built-in `gorm.DeletedAt` field/struct.

## Layering (one-directional)

`handler → service → repository → GORM`. Violations to watch for:

- Handlers must never call GORM or `*gorm.DB` directly. They call services and write `dto.Response[T]` envelopes via Gin.
- Services must never import `gin`. They take `context.Context` and return `*errors.AppError`.
- Repositories expose an interface so services are unit-testable. Pattern: `type XxxRepository interface { ... }` + lowercase impl `type xxxRepository struct { db *gorm.DB }` + constructor `func NewXxxRepository(db *gorm.DB) XxxRepository`.
- DTOs are the validation boundary. Self-service updates use a field-by-field whitelist copy from the DTO — fields absent from the DTO are silently un-updatable **by design**.

## Errors

- Services return `*errors.AppError` (constructors in `internal/errors/`, e.g. `apperrors.ErrNotFound`).
- `middleware.ErrorHandler` renders the JSON envelope. Do NOT write ad-hoc `c.JSON(400, gin.H{"error": ...})` for error paths.

## Auth / permissions

- Route-level gating: `middleware.RequirePerms(authSvc, perms...)` — `authSvc *AuthService` is the FIRST positional arg. AND semantics across listed perms; `*` wildcard granted by seed Admin role.
- Service-level ownership: pass `asAdmin bool` in from the handler (precomputed from JWT-preloaded `user.Roles`). Two distinct guards by design — don't collapse them.
- JWT: HS256, separate access + refresh tokens. Bcrypt cost 12 for passwords.
- Employee auto-creation assigns the seeded "Employee" role (carries `auth:login`) when no `role_ids` are given.

## Schema split (users ⟂ employees ⟂ dependents)

- A user MAY have one employee record (one-to-one optional). Auth lives on `users`; HR profile on `employees`; dependents on `dependents`.
- `employee_skills`, `employee_leave_quotas`, `leave_requests` all FK to `employees(id)`.

## DTO shape

- Per entity: `XCreate`, `XUpdate`, `XRead`, `XListQuery`. Located in `internal/dto/x.go`.
- `dto.Response[T]` is the success envelope.

## Routes wiring

- Authoritative route table lives in `cmd/server/main.go` in the big `v1 := r.Group("/api/v1")` block.
- `internal/handlers/router.go` is **stale** (health-only). Do not assume it's complete.

## Multipart uploads (avatar / skill icon / leave attachment)

- Sniff content-type via `http.DetectContentType` against a server-side allowlist.
- Treat client-provided `Content-Type` as a hint only. This is the review-fix #2 pattern from Phase 2 — preserved across all upload handlers.

## Swagger / docs

- `docs/swagger/` is GENERATED. Hand-edits will be clobbered by `make swag`.
- Regenerate whenever handler annotations change. Don't commit stale swagger.

## Requirement-doc language

- BA requirements (`ba-requirements/`) are written in Vietnamese with English technical terms (Nhân viên / Employee, Phòng ban / Department, Chấm công / Attendance, etc.). Read them before implementing a module — they are the source of feature intent.
