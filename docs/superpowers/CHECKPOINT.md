# Resume Checkpoint

**Last updated:** 2026-05-20
**Stopped at:** Phases 0–6 done, live-verified, and committed on `main`. Phase 6 (Attendance) closed.
**Branch:** `main`
**HEAD:** ~140 commits (run `git log --oneline -1` for exact SHA)
**DB migration version:** **9** (main + test DB both at 9)

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 7 (Announcements + SSE) per docs/superpowers/CHECKPOINT.md"*.

Phase 7 plan: not yet drafted. The spec
([`docs/superpowers/specs/2026-05-15-go-migration-design.md`](specs/2026-05-15-go-migration-design.md))
covers the high-level shape (announcements + SSE realtime, labels are
already from Phase 4). Generate the per-phase plan via the same pattern
as Phases 4–6 before executing.

### Resume entry points (any of these gets you oriented)

1. **`docs/superpowers/CHECKPOINT.md`** (this file) — single resume source of truth.
2. **`.serena/memories/project_overview.md`** — points back to (1) plus a code-map and a session-boot protocol. Serena MCP `list_memories()` surfaces these automatically.
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

### Phase 5 — DONE ✅
Leave Requests + Quota. Phase 5 verification log: [docs/superpowers/verification/phase-05.md](verification/phase-05.md). Load-bearing fix: `Employee` role seed gap (`PermLeaveUpdate/Cancel/Delete`) closed.

### Phase 6 — DONE ✅ (2026-05-20)

`attendance` + `attendance_sessions` tables (migration **000009**) + the full
attendance lifecycle (Check-In / Check-Out / Today / Get / List / Me /
Admin Create / Admin Update / Admin Delete / Matrix) with GPS-disabled-by-
default office-presence checks. Live verification:
[docs/superpowers/verification/phase-06.md](verification/phase-06.md) (20
steps, all green after the repo fix). Highlights:

- **Migration 000009** — two tables: `attendance` (one row per
  `(employee_id, date)` with `UNIQUE (employee_id, date)`) and
  `attendance_sessions` (N child rows with `ON DELETE CASCADE`). FK
  target is **`employees(id)`** per the Go schema split (REVISION
  NOTES #2). Round-trip verified (9 ↔ 8). Indexes:
  `idx_attendance_employee_id`, `idx_attendance_date`,
  `idx_attendance_employee_date`, `idx_attendance_is_deleted`,
  `idx_attendance_sessions_attendance_id`, etc. Partial unique
  `uq_attendance_sessions_one_open` enforces at most one open
  session per attendance row.
- **`is_late` is row-level and computed once** from the first check-in
  vs the configured threshold (`LATE_THRESHOLD_HOUR:_MINUTE` in
  `COMPANY_TIMEZONE`). A second on-time session after check-out does
  NOT clear an earlier late flag (REVISION NOTES #5, unit-tested
  `TestAttendance_IsLate_NotRecomputed_OnSecondSession`).
- **Open-session conflict is service-defended.** `FindOpenSession()`
  returns a clean 409 (`apperrors.ErrConflict`) instead of letting
  Postgres surface a partial-unique-violation 500 (REVISION NOTES #6).
- **Half-day flagging** at check-out: when total hours-worked across
  all sessions falls below `HALF_DAY_HOURS_THRESHOLD` (default 4h),
  `is_half_day` flips to true. Proven live in step 6.
- **`Today()` includes streak**: walks backward from today over
  workdays (Mon-Fri) counting consecutive days with at least one
  check-in. Step 8 verifies `monthly_count=1, streak=1`.
- **Admin manual entry** (`POST /attendance`) for HR to record
  attendance on behalf of an employee — beyond the Python source which
  has no admin create endpoint. Auto-derives `is_late` from the
  supplied `CheckIn` when omitted. Step 12.
- **Non-admin scope on `GET /attendance`** is silent (not 403) — the
  service filters by own `employee_id` when `asAdmin=false`. Matches
  the Python contract. Step 10 verifies one row for walker.
- **Soft-delete works**: live row stays in Postgres with
  `is_deleted=t, deleted_at IS NOT NULL`, hidden from every read via
  the `attendance.is_deleted=false` predicate. psql spot-check step 20.

#### Phase 6 load-bearing fix surfaced during verification

Step 8 (`GET /attendance/today`) failed **500** on the first pass:

```
column reference "is_deleted" is ambiguous (SQLSTATE 42702)
```

The `MonthlyCheckInCount`, `DatesWithCheckIn`, and `List` repo methods
all join secondary tables (`attendance_sessions` and `employees`) that
each carry their own `is_deleted` column. The unqualified
`models.NotDeleted` scope (`WHERE is_deleted = ?`) collided in
Postgres. The unit tests didn't catch this because:
1. `Today()` was never directly tested.
2. The `DepartmentID` filter branch in `List()` was never exercised.

**Fix:** replaced `r.base(ctx)` (which applies `NotDeleted`) with
`r.db.WithContext(ctx).Where("attendance.is_deleted = ?", false)` in
those three methods. Added two regression tests:
`TestAttendance_Today_AfterCheckOut` and
`TestAttendance_List_DepartmentFilter`. After the fix, all 22
attendance tests pass and the e2e walk completes cleanly.

This is the Phase-6 analog of Phase 4/5's silent seed gaps — the
REVISION NOTES had said "no seed gap predicted" (which was correct for
attendance), but the live verification surfaced a different load-
bearing latent bug that unit tests had missed. The verification step
is what made it visible.

#### Phase 6 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `f2be65d` | T1 | Migration 000009 — attendance + attendance_sessions |
| `ae4a507` | T3 | attendance/office config keys + getEnvFloat helper |
| `ebb5e52` | T4 | Attendance + AttendanceSession models |
| `3b66b44` | T5 | Attendance DTOs |
| `7d65389` | T6 | Attendance repository |
| `7f0e529` | T7 | Service helpers (haversine, tz, hours) |
| `06e3309` | T8 | Service core (check-in/out/today/get/list) |
| `6a4776a` | T9 | Admin CRUD + matrix |
| `ffd421d` | T10 | Handlers with Swagger |
| `329514b` | T11 | Wire routes into server |
| `cbd2061` | T12-T15 | 20 service integration tests + truncateAll order fix |
| `3440a90` | T16 | Regen Swagger + gofmt |
| `edff0b1` | T17 | Repo fix (ambiguous `is_deleted`) + 2 regression tests + verification log |
| _this_ | T18 | README endpoint table + CHECKPOINT update |

## TOOLING NOTE (unchanged from 2026-05-20)

Subagent dispatch (`Agent` with `subagent_type`) is **structurally
unavailable** in the Claude Agent SDK / VSCode-extension runtime. See
the Phase 5 CHECKPOINT for the full investigation. Phase 6 was
executed inline by the project owner; same path is recommended for
Phase 7. Both inline cadence (commit-per-task) and terminal-CLI
subagent dispatch produce shippable code.

## Code review status

Phase 0–3: review applied, fixes committed (`docs/superpowers/verification/review-fixes.md`).

Phase 4–6: **review not yet requested.** Recommendation — bundle one
review of the multipart upload pattern (now at three sites: avatar
P2, skill icon P4, leave attachment P5) plus the two-layer access
control pattern (route-level `RequirePerms` + service-level `asAdmin
bool` ownership branch — used in Phase 5 leave + Phase 6 attendance).
Doing the multipart review before Phase 7 (which likely adds
announcement attachments) keeps the duplication review surface
bounded.

## Local environment notes

- Postgres: Docker container `ennam-ecom-postgres` at `localhost:5432`,
  user `ennam` / pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB
  `exnodes_hrm_test`. Both at migration version **9**.
- `.env` is git-ignored. Phase 6 added attendance env keys (see
  `.env.example`): `COMPANY_TIMEZONE`, `LATE_THRESHOLD_HOUR/_MINUTE`,
  `CHECKOUT_THRESHOLD_HOUR/_MINUTE`, `OFFICE_GPS_ENABLED`,
  `OFFICE_LATITUDE/_LONGITUDE/_RADIUS_METERS`,
  `HALF_DAY_HOURS_THRESHOLD`. All default to safe values (GPS disabled,
  9:00 late threshold in Asia/Ho_Chi_Minh).
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** every cross-aggregate FK from Phase 2 onward
  targets `employees(id)`, NOT `users(id)`. `attendance.employee_id`
  follows this. Source of truth: [`migrations/000009_create_attendance.up.sql`](../../migrations/000009_create_attendance.up.sql).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER
  `AutoMigrate()`. See `[[feedback-migrations]]`.
- **Audit cols:** every entity has `created_at + updated_at +
  is_deleted + deleted_at` + `BEFORE UPDATE` trigger. See
  `[[feedback-audit-fields]]`.
- **DoD per phase:** must include real end-to-end verification (run
  server, curl flows, DB spot-check), commit verification log to
  `docs/superpowers/verification/phase-NN.md`. See
  `[[feedback-self-verify-each-phase]]`. Phase 6 enforced this —
  caught the ambiguous-`is_deleted` repo bug that unit tests missed.
- **Two-layer access control:** route-level `RequirePerms` for blanket
  gates; service-level `asAdmin bool` for owner-or-admin semantics.
  Same shape across Phase 5 leave + Phase 6 attendance.
- **`is_late` computed once** from the first check-in (Phase 6
  invariant — REVISION NOTES #5). Never recompute on subsequent
  sessions.
- **Open-session conflict** is service-defended via `FindOpenSession()`
  — clean 409, not a Postgres 500 (Phase 6 — REVISION NOTES #6).
- **Repo joins MUST qualify `is_deleted`** when joining a second
  table that carries its own `is_deleted` column — `models.NotDeleted`
  is fine for single-table reads, but ambiguous after a JOIN. Phase 6
  fix.

## Outstanding micro-items

- Untracked (intentional, IDE/project rules): `.claude/`, `AGENTS.md`,
  `CLAUDE.md`.
- Phase 6 plan file has unticked `- [ ]` checkboxes in the draft task
  bodies — superseded by REVISION NOTES; not worth a churn commit.
- Phase 5 manager-role completeness gap (lacks `Update/Delete` on
  leave_requests) still open — flagged for the next BA pass.
- Excel export (attendance), auto-checkout cron, system_config-backed
  late threshold, half-day matrix cells from approved leave — all
  deferred from Phase 6 (see verification log §"What's deferred").
