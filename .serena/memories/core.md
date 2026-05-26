# Core — Project Graph Root

Go 1.25 + Gin + GORM + Postgres 14+ rewrite of Python Exnodes HRM backend. Module `github.com/exnodes/hrm-api`. Phased, spec-first; each phase = vertical slice (migration → models → DTOs → repo → service → handler → tests → live verification → commit).

**Authoritative live state is NOT in these memories.** Resume from `docs/superpowers/CHECKPOINT.md` before reading code. Memories here are durable pointers and conventions; CHECKPOINT.md wins on any disagreement about phase/status.

## Memory graph

- Session boot order, what to read first, anti-patterns when orienting: `mem:resume_protocol`
- High-level project intent, where to find specs/plans/verification, tech stack snapshot, local DB/storage env: `mem:project_overview`
- Source-tree map, per-entity (migration/model/repo/service/handler) lookup table, schema split rules, where routes are actually wired: `mem:code_map`
- Pinned dependency versions and stack components (HTTP / ORM / migrations / auth / storage / docs / tests): `mem:tech_stack`
- Day-to-day commands (make targets, migration creation, Swagger regen, dev seed, docker stack): `mem:suggested_commands`
- Non-obvious code/schema conventions enforced project-wide (audit columns, soft delete, layering, error model, naming, permission gating, multipart upload sniffing): `mem:conventions`
- Pre-commit / pre-PR verification commands and when to regenerate Swagger: `mem:task_completion`
- Style + add/update threshold + reference syntax for these memories themselves: `mem:memory_maintenance`

## Project-wide invariants (also enforced in `mem:conventions`)

- `db.AutoMigrate()` is **prohibited**. Schema = versioned SQL only in `migrations/` (`golang-migrate`). Server asserts applied version on boot and refuses to start if behind or dirty.
- Layering is one-directional: handler → service → repository → GORM. Handlers never touch DB. Services never import `gin`. Repos expose interfaces.
- Soft delete uses the custom `models.NotDeleted` GORM scope. **Never** `gorm.DeletedAt`.
- Every entity table has `created_at`, `updated_at`, `is_deleted BOOLEAN`, `deleted_at TIMESTAMPTZ` + `BEFORE UPDATE` trigger calling `set_updated_at()`. PKs are UUIDs via `gen_random_uuid()` (pgcrypto).
- Services return `*errors.AppError`. Error JSON is rendered by `ErrorHandler` middleware. No ad-hoc `c.JSON(400, gin.H{...})` for errors.
- Cross-aggregate FKs from Phase 2 onward target `employees(id)`, NOT `users(id)`. The Python source mostly used `users(id)`; do not copy.
- Routes wired in `cmd/server/main.go` (the big `v1 := r.Group("/api/v1")` block). `internal/handlers/router.go` is stale (health-only).
- Generated artifacts under `docs/swagger/` are NEVER hand-edited; regenerate via `make swag` whenever handler annotations change.
