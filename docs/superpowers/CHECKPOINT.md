# Resume Checkpoint

**Last updated:** 2026-05-18
**Stopped at:** Phase 0-3 done+reviewed+fixed+verified. Phase 4 plan re-audited & corrected (REVISION NOTES in plan). NOT yet executed. Next: execute Phase 4 — Skills + Labels.
**HEAD commit:** ~105 commits on `main` (run `git log --oneline -1`)
**Branch:** `main`

## Phase 4 readiness

Plan `docs/superpowers/plans/2026-05-15-phase-04-skills-labels.md` has an authoritative `## ⚠️ REVISION NOTES (2026-05-18)` block right after the header — it supersedes the older task bodies. Key corrections: migrations are **000006_create_skills** + **000007_create_labels** (final migrate-version after phase = 7); employee↔skill via **`employee_skills`** join (NOT user_skills); skill icon upload reuses Phase-2 `UploadService` + review-fix-#2 content-sniff; **labels = Python scope only** (GET list + POST get-or-create under `/api/v1/announcement-labels`, gated by `PermAnnounceManage`, NO update/delete); **seed gap: `PermAnnounceManage` is granted to NO role in `seed_service.go`** — Phase 4 must add it to Admin (idempotent). No new permission constants needed. Execute per REVISION NOTES, not the raw task bodies where they conflict.

## TOOLING NOTE (2026-05-18)

During this session the subagent (`Agent`), `TodoWrite`, and `ToolSearch` tools became unavailable (serena MCP disconnected, tool context degraded). Phase 0-3 were executed via the efficient subagent-driven batch pattern. If those tools are still unavailable on resume, Phase 4+ must be executed inline (slower, more context-heavy) OR the session restarted to restore tooling. Prefer restoring subagent capability before executing further phases.

## Code review status (Phase 0-3)

Full review done (verdict was REQUEST_CHANGES). Fixed: all 5 Critical (refresh session-invalidation, avatar content-sniff, atomic admin employee update, GET /users + /users/:id, user_roles soft-delete) + Important #6 (Dependent/LeaveQuota repo interfaces), #8 (bcrypt cost 12), #9 (CORS env-gate). Both top security fixes live-verified (`docs/superpowers/verification/review-fixes.md`). 68 service tests pass. Deferred (NOT fixed): review #7 (unreachable guard) + all 4 Minor + the EmployeeService.toRead nil department/position projection gap.

**Storage env note:** the app reads `STORAGE_*` env keys (`internal/config/storage.go`), NOT `SUPABASE_*`. The dead duplicate `SUPABASE_*` block was removed from `.env.example` + `config.go`. Any avatar/file upload (Phase 2 avatar, Phase 9 uploads) needs a real Supabase S3 `STORAGE_*` config; local dev can point at MinIO. Storage is optional at boot (non-fatal) — upload endpoints 500 if unconfigured.

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 4 per docs/superpowers/CHECKPOINT.md"*.
Plan: `docs/superpowers/plans/2026-05-15-phase-04-skills-labels.md` (16 tasks).

NOTE: Phase 4 plan was written pre-schema-split — re-audit it first (skills/labels likely reference employee_id not user_id; check existing permission constants + migration number — next free is 000006).

## KNOWN FOLLOW-UP (deferred, not yet actioned)

`EmployeeService.toRead` / `toSummary` leave `department` / `position` nested refs nil (TODO carried from Phase 2 — DB FKs are correct & SQL-proven, only the JSON projection omits the embedded objects). Wiring the employee read view to preload+embed department/position was NOT in any phase plan task list. Address as a small dedicated task before/within a later phase or when FE needs the embedded objects.

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
