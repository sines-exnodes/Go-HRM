# Resume Checkpoint

**Last updated:** 2026-05-20
**Stopped at:** Phases 0–5 done, live-verified, and committed on `main`. Phase 5 (Leave Requests + Quota) closed.
**Branch:** `main`
**HEAD:** ~125 commits (run `git log --oneline -1` for exact SHA)
**DB migration version:** **8** (main + test DB both at 8)

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 6 per docs/superpowers/CHECKPOINT.md"*.

Plan: [`docs/superpowers/plans/2026-05-15-phase-06-attendance.md`](plans/2026-05-15-phase-06-attendance.md). **The plan already has an authoritative `## ⚠️ REVISION NOTES (2026-05-20)` block at the top — execute per those notes, not the raw task bodies where they conflict.** Pre-audit findings encoded there:

- Migration is **`000009_create_attendance`** (NOT `000011` from the draft body). Final `make migrate-version` after Phase 6 = **9**.
- `attendance.employee_id` and `attendance_sessions.attendance_id` follow the Phase 2+ schema split — `employee_id` references **`employees(id)`** (NOT `users(id)`). Rename column + indexes + repo/service/handler params from `user_id`/`userID` → `employee_id`/`employeeID` throughout.
- All four attendance perms are **already seeded correctly** (Admin/HR/Manager have Read+Manage; Employee has Read). No seed gap predicted — but Phase 5 had a denied-but-real seed gap, so verify live anyway.
- Reuse the Phase 5 patterns: `truncateToDate()` helper for Postgres DATE interop, `resolveCurrentEmployee()` for current-user→employee resolution, `hasXxxManageAll(c)` handler helper for `asAdmin bool`, two-layer access control (route `RequirePerms` + service ownership).
- `is_late` computed once from the FIRST check-in vs threshold — do NOT recompute on subsequent sessions.
- Defend the `uq_attendance_sessions_one_open` partial unique with a service-level `FindOpenSession()` check that returns a clean 409, not a 500 DB violation.

### Resume entry points (any of these gets you oriented)

1. **`docs/superpowers/CHECKPOINT.md`** (this file) — single resume source of truth.
2. **`.serena/memories/project_overview.md`** — points back to (1) plus a code-map and a session-boot protocol. Serena MCP `list_memories()` surfaces these automatically. **Wired in this session — commit `800e16e`.**
3. **`CLAUDE.md`** — auto-loaded into every Claude session; restates the boot protocol.

## Current state

### Phase 0 — DONE ✅
Phase 0 verification log: [docs/superpowers/verification/phase-00.md](verification/phase-00.md)

### Phase 1 — DONE ✅
Auth + RBAC. Phase 1 verification log: [docs/superpowers/verification/phase-01.md](verification/phase-01.md)

### Phase 2 — DONE ✅
Users + Employees + Dependents. Phase 2 verification log: [docs/superpowers/verification/phase-02.md](verification/phase-02.md). Code review applied (see review-fixes.md).

### Phase 3 — DONE ✅
Departments + Positions. Phase 3 verification log: [docs/superpowers/verification/phase-03.md](verification/phase-03.md).

### Phase 4 — DONE ✅
Skills + employee_skills + announcement labels. Phase 4 verification log: [docs/superpowers/verification/phase-04.md](verification/phase-04.md). Seed gap closed (`PermAnnounceManage` → Admin + HR Manager).

### Phase 5 — DONE ✅ (2026-05-20)

`leave_requests` table (migration **000008**) + the full Leave Request lifecycle (Create / Update / Approve / Reject / Cancel / Delete / Get / List / Balance / Dashboard / History) with attachment upload. Live verification: [docs/superpowers/verification/phase-05.md](verification/phase-05.md) (22 steps, all green). Highlights:

- Migration **000008** — single `leave_requests` table, both `employee_id` and `created_by` FK → `employees(id) ON DELETE RESTRICT`. Round-trip verified (8 ↔ 7). Indexes on `employee_id`, `status`, `(from_date, to_date)`, `is_deleted`, and `created_at DESC`. Five CHECK constraints (three enum domains + `total_days >= 0` + `to_date >= from_date`).
- **Quota source-of-truth is `employee_leave_quotas`** (created in Phase 2 migration 000004) — NOT a duplicate `leave_quotas` table and NOT columns on `users`. The plan's draft was wrong about this; the REVISION NOTES block at the top of the Phase 5 plan corrected it before execution.
- **Warnings are non-blocking**: insufficient quota + date overlap produce `warnings: []string` on Create/Update responses but the row is still created. Proven live in verification step 5.
- **`created_by` semantics** (REVISION NOTES #6): admin acting on behalf sets `employee_id = subject`, `created_by = admin's employee_id` (not user id). Proven live in step 13.
- **Half-day rule**: requires `from_date == to_date`; rejected `400` otherwise. Total days = `(to - from).days + 1`, ×0.5 when half. Proven live in step 7.
- **Status state machine**: pending → approved/rejected/cancelled; approved → cancelled; rejected/cancelled are terminal (409 on edit). Admin patch on approved reverts to pending (Python contract). Proven live in steps 8/10/12.
- **Balance is a live SUM** of approved `total_days` over the calendar year. Cancelling an approved row naturally restores remaining — no separate restore path. Proven live in steps 9/11.
- **Two distinct access guards**: route-level `RequirePerms` (with `{required, missing}` evidence shape) and service-level ownership (`row.EmployeeID == currentEmp.ID || asAdmin`). Both exercised; both produce 403 with different message shapes by design. Verified step 14a/14b/15/16.
- **Soft-delete works**: live row stays in Postgres with `is_deleted=t, deleted_at IS NOT NULL`, hidden from every read via the `NotDeleted` scope. psql spot-check step 21.

#### Phase 5 load-bearing fix surfaced during verification

REVISION NOTES item #4 had claimed the `Employee` role seed was complete. **It wasn't** — the role only carried `PermLeaveRead + PermLeaveCreate`, so non-admin owners couldn't even reach the service body for `cancel/update/delete` on their own pending request (route-level `RequirePerms` fired first with 403). This is the Phase-5 analog of Phase 4's `PermAnnounceManage` gap.

**Fix:** added `PermLeaveUpdate, PermLeaveCancel, PermLeaveDelete` to the Employee role in `seed_service.go` (commit included in the Phase 5 verification commit). The seed-merge logic ran cleanly at next boot:

```
seed: merged permissions into role "Employee"
```

Cross-employee writes are still rejected by the service's ownership branch — granting these perms cannot leak.

#### Phase 5 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `063bb0d` | 1 | Migration 000008 — leave_requests table |
| `1818ac6` | 2 | Models: LeaveRequest + enum constants |
| `ced69dc` | 3 | DTOs: write inputs, read outputs, balance, dashboard, lists |
| `6dc0969` | 4 | LeaveRequestRepository (interface + Postgres impl) |
| `ed1bdcd` | 5–8 | LeaveService (11 methods, attachment upload pattern) |
| `43abc0f` | 9–13 | leave_handler HTTP layer (11 endpoints, multipart) |
| `214acad` | 14–15 | Wire routes in main.go + regen Swagger |
| `6bfa6de` | 16–19 | 22 integration tests + truncateAll order fix |
| _next_ | 20–21 | Verification log + Employee seed-merge fix + this checkpoint update |

## TOOLING NOTE (updated 2026-05-20 post-restart-investigation)

Subagent dispatch (`Agent` with `subagent_type`) is **structurally unavailable** in the Claude Agent SDK / VSCode-extension runtime that hosts these sessions. Investigation log:

- **Phase 4 session:** every probed `subagent_type` returned "not found" with an empty available-agents list. Worked around by inline execution.
- **Phase 5 session:** same. Worked around by inline execution.
- **Restart probe (2026-05-20):** even `code-simplifier` from the **working** `claude-plugins-official` marketplace (proven dispatchable from the standalone Claude Code CLI in terminal) returns "not found" in this SDK runtime. Plugin files, manifests, frontmatter, and `enabledPlugins` config are all OK — verified by file inspection.

**Conclusion:** The Agent tool's `subagent_type` registry is populated by the host environment, and the VSCode-extension SDK runtime intentionally exposes an **empty** registry. Plugin agents in `~/.claude/plugins/marketplaces/.../agents/*.md` are only dispatchable when running `claude` from terminal. This is not a fixable misconfiguration — it's a runtime capability gap.

**Implication for Phase 6:** Do NOT re-probe subagent dispatch in the next IDE session — you will get the same "not found" + empty list. Two viable paths:

1. **Inline by project-owner (recommended, proven across Phases 4 + 5).** Continue the same cadence: commit-per-task, batch logical tasks, end with a live verification log and CHECKPOINT update. ~18 tasks in Phase 6, comparable to Phase 5's 21.
2. **Run Phase 6 from terminal `claude` CLI** if the user prefers the proper `subagent-driven-development` pattern. Loses IDE integration; requires copy/paste of output. Use only if Phase 6 size becomes painful.

Phase 6 default is path (1) — user explicitly chose this at session close.

## Code review status

Phase 0–3: review applied, fixes committed (`docs/superpowers/verification/review-fixes.md`). Two top security fixes live-re-verified.

Phase 4 + 5: **review not yet requested.** Recommendation — bundle one review of the multipart upload pattern that's now shared across `readAvatar` (Phase 2), `readSkillIcon` (Phase 4), and `readLeaveAttachment` (Phase 5). The `http.DetectContentType` content-sniff is now tested at three sites; the duplication between the three readers is a candidate for a small shared helper. Doing both Phase 4 + Phase 5 reviews together covers the full multipart surface before Phase 7 (Announcements) potentially adds another attachment site.

## Local environment notes

- Postgres: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB `exnodes_hrm_test`. Both at migration version **8**.
- `.env` is git-ignored. To exercise the attachment-upload path live (Phase 5), reuse the MinIO recipe from `docs/superpowers/verification/phase-04.md` §3 (`phase04-minio`, `localhost:19000`, bucket `hrm-uploads`).
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** `users` (auth) ⟂ `employees` (HR) ⟂ `dependents` ⟂ `employee_skills` ⟂ `employee_leave_quotas` ⟂ `leave_requests`. Every cross-aggregate FK from Phase 2 onward targets `employees(id)`, NOT `users(id)`. Source of truth: [`migrations/000008_create_leave_requests.up.sql`](../../migrations/000008_create_leave_requests.up.sql).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`. See `[[feedback-migrations]]`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at` + `BEFORE UPDATE` trigger. See `[[feedback-audit-fields]]`.
- **DoD per phase:** must include real end-to-end verification (run server, curl flows, DB spot-check), commit verification log to `docs/superpowers/verification/phase-NN.md`. See `[[feedback-self-verify-each-phase]]`. Phase 5 enforced this — caught the Employee seed gap.
- **Attachment upload pattern:** mandatory `http.DetectContentType` sniff with allowlist; never trust the client-supplied `Content-Type`. Three sites now use this (avatar, skill icon, leave attachment). The handler-side readers (`readAvatar`/`readSkillIcon`/`readLeaveAttachment`) are intentionally duplicated for different size caps + ext allowlists but are a refactor candidate.
- **Warnings vs errors (Phase 5):** insufficient quota and date overlap are **non-blocking warnings**, not 400s. The Create/Update endpoints return the created/updated row plus a `warnings: []string` array. Do NOT promote either case to a hard error without re-auditing the FE.
- **Balance = live SUM**, not a stored counter. Cancelling an approved row naturally restores remaining. No materialized view, no background job.
- **Two-layer access control:** route-level `RequirePerms` for blanket gates; service-level ownership branch for owner-or-admin semantics. Both produce 403 but with different message shapes — this is deliberate, not a bug.

## Outstanding micro-items

- Untracked (intentional, IDE/project rules): `.claude/`, `AGENTS.md`, `CLAUDE.md`.
- Phase 5 plan file has unticked `- [ ]` checkboxes in the draft task bodies — superseded by REVISION NOTES; not worth a churn commit.
- Phase 2 carryover `EmployeeService.toRead` department/position nil-projection gap is still open — does NOT block Phase 6 but should be addressed when FE needs embedded objects on `GET /employees/me`.
- Manager role completeness: has `Read/Create/Approve/Cancel/Manage` but lacks `Update/Delete` on leave_requests. Symmetric with Admin/HR who have everything. Flagged for the next BA pass.
