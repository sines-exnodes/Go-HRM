# Resume Checkpoint

**Last updated:** 2026-05-22
**Stopped at:** Phases 0–8 done, live-verified, and committed on `main`. Phase 8 (Organization Settings) closed.
**Branch:** `main`
**HEAD:** ~165 commits (run `git log --oneline -1` for exact SHA)
**DB migration version:** **11** (main + test DB both at 11)

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 9 (Email invite + Push notifications) per docs/superpowers/CHECKPOINT.md"*.

Phase 9 plan exists: [`docs/superpowers/plans/2026-05-15-phase-09-email-invite-push.md`](plans/2026-05-15-phase-09-email-invite-push.md). Same drill as Phases 5–8: read CHECKPOINT → spec → plan REVISION NOTES → execute task-by-task. Likely Phase-9 REVISION NOTES:

- Migration **`000012_...`** (000001–000011 taken).
- `device_tokens` + `user_notification_settings` tables already exist (migration 000004 from Phase 2). Phase 9 just adds endpoints + sender plumbing. Verify before drafting any new migration.
- Email sender is likely an SMTP shim or AWS SES — pick per spec §11; pin secrets in env.
- Push notifications: FCM topic-based fanout; needs FCM_SERVICE_ACCOUNT_KEY env. Plan for graceful degradation when not configured (don't crash on boot).

### Resume entry points (any of these gets you oriented)

1. **`docs/superpowers/CHECKPOINT.md`** (this file) — single resume source of truth.
2. **`.serena/memories/project_overview.md`** — code-map + boot protocol.
3. **`CLAUDE.md`** — auto-loaded into every Claude session.

## Current state

### Phases 0–7 — DONE ✅

Foundation → Auth+RBAC → Users+Employees+Dependents → Departments+Positions → Skills+Labels → Leave Requests → Attendance → Announcements+SSE. Verification logs: [`phase-00.md`](verification/phase-00.md) through [`phase-07.md`](verification/phase-07.md).

### Phase 8 — DONE ✅ (2026-05-22)

Organization settings (singleton `system_config` table) + 4 routes (admin-gated attendance subtree + open-read company-profile + admin-gated company-profile update). Live verification: [docs/superpowers/verification/phase-08.md](verification/phase-08.md) (16 e2e steps + DB singleton-constraint check, all green). Highlights:

- **Migration 000011** — single-row `system_config` table. Sentinel PK `00000000-0000-0000-0000-000000000001`. DB-level `CHECK (id = sentinel)` constraint blocks any second-row INSERT (verified live in verification step 16). Audit columns present for schema parity but unused.
- **`company_address_updated_by → employees(id) ON DELETE SET NULL`** per the Go schema split (REVISION NOTES #2). Mirrors `leave_requests.created_by` + `announcements.author_id`. GET projection resolves `updated_by_name` from `Employee.FullName`.
- **Singleton repo API**: `Get()` / `EnsureExists()` / `UpdateFields(map[string]any)` — no list, no by-id lookup, no `NotDeleted` scope (REVISION NOTES #10).
- **Seed integration**: `SeedService` constructor now takes `systemConfigRepo`; new `seedSystemConfig(ctx)` runs `INSERT … ON CONFLICT DO NOTHING` on boot.
- **`updated_at` + `updated_by` stamping**: `UpdateCompanyProfile` only stamps when at least one of the three address fields is supplied. Empty-patch case (no address fields) does NOT churn the audit trail — verified by `TestOrgSettings_UpdateCompanyProfile_NoAddressFields_NoStamp`.
- **Binding-tag bounds** mirror the DB CHECK constraints — `late_threshold_hour ≤ 23`, `company_latitude ∈ [-90, 90]`, `company_longitude ∈ [-180, 180]`. Verified live: `hour=25` → 400, `lat=200` → 400.
- **Permission gating**: `/attendance` (GET + PATCH) and `PATCH /company-profile` require `PermOrgSettings`. `GET /company-profile` is JWT-only (open to all signed-in users so the mobile map preview renders).

#### Phase 8 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `b440fbd` | – | Plan REVISION NOTES at the top |
| `457e162` | T1 | Migration 000011 — system_config singleton |
| `21fc08b` | T2-T6 | Model + DTOs + repo + service + handler |
| `3e25441` | T7-T9 | Wire routes + service in main.go + seed integration |
| `9dbba30` | T10 | 8 service tests + truncateAll + Swagger regen (also accidentally tracked .claude/ etc.) |
| _(next)_ | – | Untrack .claude/ + AGENTS.md + CLAUDE.md |
| `ab84445` | T11 | E2E verification log |
| _this_ | T12 | README + CHECKPOINT update |

## TOOLING NOTE (unchanged)

Subagent dispatch (`Agent` with `subagent_type`) is **structurally unavailable** in the VSCode-extension SDK runtime. Inline by project-owner (commit-per-task) continues to be the proven cadence.

## Code review status

Phase 0–3: review applied, fixes committed.

Phase 4–8: **review not yet requested.** Recommendation — bundle one review covering:

- Multipart upload pattern (now at 3 sites: avatar P2, skill icon P4, leave attachment P5; will be 4 if Phase 7 attachment-upload endpoint is wired).
- Two-layer access control (`RequirePerms` + `asAdmin bool` ownership branch — Phases 5/6/7).
- Composite-PK reactivation pattern (`AnnouncementLabel` + `AnnouncementTargetDepartment` + `EmployeeSkill`).
- Singleton repo pattern (Phase 8 `SystemConfigRepository`) — only 4 methods, defensive against accidental list/insert.

## Local environment notes

- Postgres: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB `exnodes_hrm_test`. Both at migration version **11**.
- **Port 8080 conflict**: `ennam-kg-server` container holds host port 8080 in this dev environment. Phase 7 + Phase 8 live verifications ran on `PORT=8082`. CI default stays 8080.
- `.env` is git-ignored.
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** every cross-aggregate FK from Phase 2 onward targets `employees(id)`, NOT `users(id)`. Includes `system_config.company_address_updated_by`. Exceptions: `users.id` is the FK target for auth-level surfaces (`device_tokens`, `user_notification_settings`, `announcement_views`).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at` + `BEFORE UPDATE` trigger. Singletons keep the columns for schema parity even when soft-delete is meaningless.
- **Singleton tables:** PK is a fixed sentinel UUID + `CHECK (id = '…')` constraint at the DB level. Repo exposes only `Get / EnsureExists / UpdateFields(map)` — no list, no by-id lookup, no `NotDeleted` scope.
- **Composite-PK join models:** when the join carries audit columns AND a composite PK, do NOT use `gorm:"many2many:..."` tag — declare an explicit Go model. Replace-set logic uses snapshot-diff-reactivate (see Phase 7 `ReplaceLabels`).
- **Repo joins MUST qualify `is_deleted`** — `models.NotDeleted` becomes ambiguous after a JOIN to a table that carries `is_deleted`. Phase 6 + Phase 7 both encountered this.
- **SSE broadcast is a "refresh hint"** — FE refetches via GET on receipt; visibility is enforced on the read path, not at the hub.
- **Idempotent view tracking** — `Clauses(clause.OnConflict{DoNothing:true})`. Preserves first view time.
- **Partial PATCH semantics:** pointer-typed DTO fields → only write when non-nil. Empty PATCH is a no-op success. Stamp `updated_by/at` only when "real" content fields change (Phase 8 contract: address-touching fields trigger the stamp).

## Outstanding micro-items

- Untracked (intentional): `.claude/`, `AGENTS.md`, `CLAUDE.md`.
- Phase 5/6/7/8 plan files have unticked `- [ ]` checkboxes in draft task bodies — superseded by REVISION NOTES blocks; not worth churn commits.
- Phase 5 manager-role completeness gap (lacks `Update/Delete` on leave_requests) still open.
- Phase 6 attendance service still reads thresholds from env vars — should switch to `system_config` lookup now that the row exists. Small follow-up commit.
- Phase 7 attachment-upload HTTP handler is deferred (schema + repo in place).
- Phase 7 `target_audience='custom'` deferred (no backing table).
- Phase 7 scheduled-publish cron deferred (column wired, no background job).
- Phase 8 logo upload deferred (not in BA brief).
- Phase 6 + earlier deferred items still open.
