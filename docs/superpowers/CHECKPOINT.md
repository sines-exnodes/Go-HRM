# Resume Checkpoint

**Last updated:** 2026-05-18
**Stopped at:** Phase 1 COMPLETE (all 18 tasks). Next action: Phase 2 — Employees + Dependents Module (26 tasks).
**HEAD commit:** `46c5d26 docs(plan): tick Phase 1 Tasks 16-17 + DoD complete`
**Branch:** `main` (36 commits)

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 2 per docs/superpowers/CHECKPOINT.md"*.
Plan: `docs/superpowers/plans/2026-05-15-phase-02-users.md` (26 tasks, Employees + Dependents Module).

Claude should:
1. Read this file + `docs/superpowers/specs/2026-05-15-go-migration-design.md`
2. Skim `docs/superpowers/plans/2026-05-15-phase-01-auth-rbac.md` from `### Task 9:` onward
3. Verify state still matches "Current state" section below before dispatching the next implementer
4. Continue per the `superpowers:subagent-driven-development` pattern (one subagent per task, batch trivial tasks where the plan content is mechanical)

## Current state

### Phase 0 — DONE ✅
16 commits. README + verification log committed. `make migrate-up && make run && curl /health` all verified end-to-end against local Postgres (Docker `ennam-ecom-postgres`, user `ennam` / pass `ennam_dev_2026`, DB `exnodes_hrm`).

### Phase 1 — Tasks 1 through 8 DONE ✅

| Task | Commit | Summary |
|---|---|---|
| 1 | `5358b8d` | Migrations 000002 (roles/users/user_roles) + 000003 (employees/dependents). DB migrate-version = 3. |
| 2 | `7196399` | GORM models: Role, User (auth-only), Employee (HR fields, manager self-ref), Dependent |
| 3 | `e3301c6` | Permission registry — 35 constants + PermissionGroups + IsValid + 2 tests passing |
| 4 | `868149a` | `pkg/utils/password.go` (bcrypt) — TDD, 2 tests pass |
| 5 | `cc80dcd` | `pkg/utils/jwt.go` (HS256 access/refresh) — TDD, 4 tests pass + `.env.example` updated |
| 6 | `03325f6` | `internal/repositories/role_repo.go` interface + GORM impl |
| 7 | `048b648` | `internal/repositories/user_repo.go` with role preloads + employee preload |
| 7b | `a298f67` | `internal/repositories/employee_repo.go` + `dependent_repo.go` |
| 8 | `3a60c84` | `internal/services/testhelper_test.go` — Postgres test DB harness, factories `makeUser`/`makeRole`/`makeEmployee`. New `make test-db-up` target. Test DB: `exnodes_hrm_test` (uses `TEST_DATABASE_URL` env, derives from main DB creds if unset). |

`go build ./...` clean. `go test ./pkg/utils/... ./internal/permissions/...` all pass (8 tests).

### Phase 1 — REMAINING (Tasks 9-17)

| Task | What |
|---|---|
| 9 | Auth service — `Login`, `Refresh`, `Logout`, `ResolveUserPermissions`. TDD. Returns `LoginResult{Tokens, User}` with preloaded employee. |
| 10 | Seed service — idempotent. Creates 5 system roles + super admin user + **matching employee record** for super admin (full_name from `SUPER_ADMIN_NAME` env). TDD. |
| 11 | `internal/middleware/auth.go` — JWT parse + load user + session-invalidation check (vs `email_changed_at` / `password_reset_at`) |
| 12 | `internal/middleware/permissions.go` — `RequirePerms(...)` variadic, wildcard `*` bypass, all-required AND semantics |
| 13 | `internal/dto/auth.go` — LoginRequest, LoginResponse with embedded user/employee summary, RefreshRequest |
| 14 | `internal/handlers/auth_handler.go` — POST `/api/v1/auth/login`, `/refresh`, `/logout` + Swagger |
| 15 | `internal/handlers/role_handler.go` — GET `/api/v1/roles/permissions` (PermissionGroups catalog for FE picker) |
| 16 | Wire into `cmd/server/main.go` — repos, services, middlewares, routes, seed on boot |
| 17 | End-to-end live verification + commit `docs/superpowers/verification/phase-01.md` |

After Phase 1 → Phase 2 (Employees + Dependents Module, 26 tasks — plan file rewritten).

## Local environment notes

- Postgres: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam`, pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB `exnodes_hrm_test` (created).
- `.env` is git-ignored — credentials live there. `.env.example` has the plan-spec defaults (`postgres/postgres`).
- Migrations applied to main DB: version 3. Test DB also at version 3.
- Go 1.24 pinned in `go.mod` (toolchain may upgrade some patches on `tidy`).
- Phase 0 verification log: [docs/superpowers/verification/phase-00.md](docs/superpowers/verification/phase-00.md)

## Key design decisions (do NOT redo)

- **Schema split:** `users` (auth) ⟂ `employees` (HR) ⟂ `dependents`. Mirrors `exn-hr/Exn-hr/backend` pattern. Stored in memory: `~/.claude/.../memory/project_employee_schema.md`.
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`. See `[[feedback-migrations]]`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at`. See `[[feedback-audit-fields]]`.
- **DoD per phase:** must include real end-to-end verification (run server, curl flows, DB spot-check), commit verification log to `docs/superpowers/verification/phase-NN.md`. See `[[feedback-self-verify-each-phase]]`.

## Outstanding micro-items (no action needed unless asked)

- Untracked: `.claude/` (IDE-local tooling)
- Untracked: `pkg/utils/.gitkeep` was deleted in an earlier task but not staged — git status shows it as deleted. Will be cleaned up automatically when next task touches pkg/utils.
- Plan file `docs/superpowers/plans/2026-05-15-phase-01-auth-rbac.md` has unstaged checkbox edits (Tasks 1-8 ticked) — will get committed when current task batch starts again, or via standalone "docs(plan): tick Phase 1 progress" commit at session resume.
