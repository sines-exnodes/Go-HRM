# Resume Checkpoint

**Last updated:** 2026-05-20
**Stopped at:** Phases 0–7 done, live-verified, and committed on `main`. Phase 7 (Announcements + Mobile + SSE) closed.
**Branch:** `main`
**HEAD:** ~160 commits (run `git log --oneline -1` for exact SHA)
**DB migration version:** **10** (main + test DB both at 10)

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 8 (Organization Settings) per docs/superpowers/CHECKPOINT.md"*.

Phase 8 plan exists: [`docs/superpowers/plans/2026-05-15-phase-08-org-settings.md`](plans/2026-05-15-phase-08-org-settings.md). Same drill as Phases 5–7: read CHECKPOINT → spec → plan REVISION NOTES → execute task-by-task. Likely Phase-8 REVISION NOTES needed:

- Migration **`000011_create_organization_settings`** (migrations 000001–000010 are taken).
- Single-row `organization_settings` table OR key-value `system_config` — pick per spec §10. The Phase 6 attendance plan already references `system_config` as the eventual home for `LATE_THRESHOLD_*` overrides; Phase 8 should land that table.
- Verify schema split rule (likely no cross-aggregate FK — settings is a singleton).
- `PermOrgSettings` already exists in registry + seeded to Admin + HR Manager (Phase 1/4 seed).

### Resume entry points (any of these gets you oriented)

1. **`docs/superpowers/CHECKPOINT.md`** (this file) — single resume source of truth.
2. **`.serena/memories/project_overview.md`** — points back to (1) plus a code-map.
3. **`CLAUDE.md`** — auto-loaded into every Claude session; restates the boot protocol.

## Current state

### Phase 0 — DONE ✅
Foundation. Phase 0 verification log: [docs/superpowers/verification/phase-00.md](verification/phase-00.md)

### Phase 1 — DONE ✅
Auth + RBAC. Phase 1 verification log: [docs/superpowers/verification/phase-01.md](verification/phase-01.md)

### Phase 2 — DONE ✅
Users + Employees + Dependents. Phase 2 verification log: [docs/superpowers/verification/phase-02.md](verification/phase-02.md). Code review applied.

### Phase 3 — DONE ✅
Departments + Positions. Phase 3 verification log: [docs/superpowers/verification/phase-03.md](verification/phase-03.md).

### Phase 4 — DONE ✅
Skills + announcement labels. Phase 4 verification log: [docs/superpowers/verification/phase-04.md](verification/phase-04.md).

### Phase 5 — DONE ✅
Leave Requests + Quota. Phase 5 verification log: [docs/superpowers/verification/phase-05.md](verification/phase-05.md). Employee seed gap closed.

### Phase 6 — DONE ✅
Attendance + matrix. Phase 6 verification log: [docs/superpowers/verification/phase-06.md](verification/phase-06.md). Repo `is_deleted` join-ambiguity fix surfaced + closed during live verification.

### Phase 7 — DONE ✅ (2026-05-20)

Announcements (web + mobile) + SSE realtime push. Five tables (migration **000010**) + 10 web routes + 2 mobile routes + 1 SSE route. Live verification: [docs/superpowers/verification/phase-07.md](verification/phase-07.md) (17 e2e steps + DB spot-check, all green). Highlights:

- **Migration 000010** — five tables: `announcements`, `announcement_labels` (explicit join), `announcement_target_departments`, `announcement_attachments`, `announcement_views`. `author_id → employees(id) RESTRICT` (REVISION NOTES #2). `target_audience` enum `('all','department')` — 'custom' dropped (no backing table; deferred to Phase 7.5 if BA asks). `announcement_views.user_id → users(id) CASCADE` (auth-level read marker).
- **Explicit join models** (REVISION NOTES #10) for `AnnouncementLabel` + `AnnouncementTargetDepartment` because the join rows carry audit columns. `gorm:"many2many:..."` tag would strip those. Mirrors Phase 4 `EmployeeSkill`.
- **In-memory SSE hub** (`internal/sse/hub.go`) — 7 unit tests passing under `-race`. Buffer size 16 per client; full buffer drops events for that client only (slow consumer cannot block publisher). Single-process: scaling beyond 1 replica requires Redis pub/sub.
- **`JWTFromQueryOrHeader` middleware** for SSE — EventSource cannot set Authorization headers (REVISION NOTES #8). Refactored existing `JWT` to share an internal `applyAuth()` helper.
- **Two-layer access control** (same as Phase 5 / 6): route-level `RequirePerms(authSvc, PermAnnounceManage)` upstream; service-level ownership branch via the `asAdmin bool` precomputed by the handler from JWT-preloaded `user.Roles`.
- **Visibility predicate** (REVISION NOTES #14) for non-admin reads. Applied at the SQL layer in `repo.List` via `applyAudienceFilter()` and in-process for `Get` via `canSee()`. Verified live: drafts hidden, department-targeted rows visible only to dept members.
- **Idempotent `MarkViewed`** — `ON CONFLICT (announcement_id, user_id) DO NOTHING` preserves the FIRST view time per Python contract. Second call is a no-op. Verified live (step 10).
- **SSE broadcast is "refresh-trigger" not "data-push"** — the FE refetches the visible list on receipt, so the audience filter doesn't need to be duplicated in the hub layer. Broadcast filter is nil (everyone), visibility is applied on the GET path.

#### Phase 7 load-bearing fixes surfaced during verification + service tests

1. **Composite-PK reactivation pattern** (Phase 7 analog of the Phase 4 `EmployeeSkill.ReplaceForEmployee` design): `announcement_labels` and `announcement_target_departments` have a composite PK `(announcement_id, FK_id)` — no separate `id` column. The original "soft-delete existing + insert new" approach failed because the second insert tripped the PK constraint. Fix: snapshot the join rows (including soft-deleted), soft-delete the ones no longer wanted, and **reactivate** existing rows by setting `is_deleted=false` instead of inserting parallel rows. Service test `TestAnnouncement_Update_LabelsReplaceSet` would have failed at runtime without this; it caught it before live verification.

2. **`UpsertView` ON CONFLICT clause** — GORM v2 doesn't honor `Set("gorm:insert_option", "ON CONFLICT ...")` (that was a v1 API). Replaced with `Clauses(clause.OnConflict{...})` from `gorm.io/gorm/clause`. Surfaced by `TestAnnouncement_MarkViewed_Idempotent`.

3. **Label model has no Color field** — initial DTO `AnnouncementLabelBrief` had `color` based on the plan draft; the actual `models.Label` only has `Name`. Dropped from DTO + service projection before commit (caught by compile error).

#### Phase 7 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `1a0d15c` | – | Plan REVISION NOTES at the top |
| `6c45c72` | T1 | Migration 000010 — announcements + 4 child tables |
| `c72e8a3` | T2 | Announcement, label-join, target-dept, attachment, view models |
| `cdb1139` | T3 | Web + mobile + SSE event DTOs |
| `5970510` | T4 | SSE hub + 7 unit tests under -race |
| `a87b579` | T5 | Announcement repository |
| `4fb7520` | T6-T11 | Announcement service (CRUD + Publish + visibility + Mobile + broadcast) |
| `344ce68` | – | 21 service integration tests + 2 repo bug fixes (composite-PK + ON CONFLICT) |
| `4a426e1` | T12 | SSE handler + JWTFromQueryOrHeader |
| `9532e65` | T13 | Announcement HTTP handler (web + mobile + view + publish) |
| `4f53151` | T14 | Wire routes + SSE hub singleton + adapter |
| `317aa4a` | T15 | Regen Swagger + gofmt (accidentally tracked .claude/, AGENTS.md, CLAUDE.md) |
| _(next)_ | – | Untrack .claude/ + AGENTS.md + CLAUDE.md |
| `fd8b678` | T16-T19 | E2E verification log (17 steps + DB spot-check) |
| _this_ | T20 | README + CHECKPOINT update |

## TOOLING NOTE (unchanged)

Subagent dispatch (`Agent` with `subagent_type`) is **structurally unavailable** in the VSCode-extension SDK runtime. Investigation from Phase 5; same path expected for Phase 8+. Inline by project-owner (commit-per-task) continues to be the proven cadence.

## Code review status

Phase 0–3: review applied, fixes committed.

Phase 4–7: **review not yet requested.** Recommendation — bundle one review covering:
- Multipart upload pattern (now at 3 sites: avatar P2, skill icon P4, leave attachment P5; will be 4 if Phase 7 attachment-upload endpoint is wired).
- Two-layer access control (`RequirePerms` + `asAdmin bool` ownership branch — Phases 5/6/7).
- Composite-PK reactivation pattern (`AnnouncementLabel` + `AnnouncementTargetDepartment` in Phase 7 + `EmployeeSkill` from Phase 4) — three sites now share this.
- SSE hub design (single-process, drop-on-full-buffer) before Phase 9 likely adds push-notification fan-out.

## Local environment notes

- Postgres: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB `exnodes_hrm_test`. Both at migration version **10**.
- **Port 8080 conflict**: `ennam-kg-server` container holds host port 8080 in this dev environment. Phase 7 live verification ran on `PORT=8082`. CI default stays 8080; only local-dev needs the swap.
- `.env` is git-ignored.
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** every cross-aggregate FK from Phase 2 onward targets `employees(id)`, NOT `users(id)`. Includes `announcements.author_id`. Exceptions: `users.id` is the FK target for auth-level surfaces (`device_tokens`, `user_notification_settings`, `announcement_views`).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at` + `BEFORE UPDATE` trigger.
- **Composite-PK join models:** when the join carries audit columns AND a composite PK, do NOT use `gorm:"many2many:..."` tag — declare an explicit Go model (mirrors `EmployeeSkill`). Replace-set logic uses the snapshot-diff-reactivate pattern (see Phase 7 `ReplaceLabels`).
- **Repo joins MUST qualify `is_deleted`** — `models.NotDeleted` becomes ambiguous after a JOIN to another table that carries `is_deleted`. Phase 6 + Phase 7 both encountered this. Pattern: `db.Where("<table>.is_deleted = ?", false)` instead of `Scopes(models.NotDeleted)`.
- **SSE broadcast is a "refresh hint"** — the FE refetches via GET on receipt; visibility is enforced on the read path, not at the hub. Keeps the audience logic in one place.
- **Idempotent view tracking** — `Clauses(clause.OnConflict{DoNothing:true})`. Preserves first view time. Use this pattern for any per-user marker table.

## Outstanding micro-items

- Untracked (intentional): `.claude/`, `AGENTS.md`, `CLAUDE.md`.
- Phase 5/6/7 plan files have unticked `- [ ]` checkboxes in draft task bodies — superseded by REVISION NOTES blocks at the top; not worth churn commits.
- Phase 5 manager-role completeness gap (lacks `Update/Delete` on leave_requests) still open — flagged for next BA pass.
- Phase 7 attachment-upload HTTP handler is deferred (schema + repo are in place; BA confirms whether ship-blocker).
- Phase 7 `target_audience='custom'` deferred (no backing `announcement_target_users` table; Phase 7.5 if needed).
- Phase 7 scheduled-publish cron deferred (column wired, no background job).
- Phase 6 deferred items still open (Excel export, auto-checkout cron, system_config-backed late threshold, half-day matrix cells from leave).
