# Attendance — Python ↔ Go API Parity Audit

**Date:** 2026-06-05
**Module:** Attendance (mobile check-in/out + web monthly matrix + export)
**Status:** IN PROGRESS — decisions D1–D6 LOCKED 2026-06-05. **Plan A DONE** (G1/G2/G4/G6 + D1 implemented, verified — full suite green on branch `worktree-feat+attendance-parity`). Plan B (G3/G5/D2/D5) in progress. G7 (holidays) blocked.
**Method:** module-by-module parity audit (same pipeline as announcements/employees/roles:
audit → locked decisions → PR(s) → verification log → FE doc → handoff)

**Sources read**
- Python: `app/routers/attendance.py`, `app/services/attendance.py`,
  `app/schemas/attendance.py`, `app/models/attendance.py`,
  `app/core/permissions.py` (at `E:\Work\exnodes-hrm-api`)
- Go: `internal/handlers/attendance_handler.go`, `internal/services/attendance_service.go`,
  `internal/services/attendance_matrix.go`, `internal/dto/attendance.go`,
  `internal/models/attendance.go`, `internal/permissions/registry.go`,
  `cmd/server/main.go` (route wiring), `migrations/000009_create_attendance.*.sql`
- BA intent:
  - `ba-requirements/.../WEB-APP/EP-004-attendance-management/US-001-attendance-list/details/DR-004-001-01-attendance-list.md` (v1.2)
  - `ba-requirements/.../MOBILE-APP/EP-003-attendance-management/US-001-daily-attendance/details/DR-003-001-01-check-in-out.md` (v1.0)
  - `ba-requirements/.../WEB-APP/EP-009-organization-settings/US-002-attendance-settings/` (thresholds)

---

## 0. TL;DR

Go's attendance API is a **superset on the admin/CRUD axis** and a **subset on the
two headline web features the BA actually asks for**:

- **In-scope, BA-backed** (status):
  1. ✅ **DONE (Plan A)** — **Leave-integrated matrix** — approved leave rendered in matrix cells (AC-016, SR-004).
  2. ✅ **DONE (Plan A)** — **Combined half-day cells** — ½-leave + worked half, `worked_half_status`, AM/PM
     thresholds (AC-026–031, SR-002/003/004/008/011).
  3. ⏳ **Plan B** — **Excel export** — bulk + per-employee `.xlsx`, incl. the two summary columns
     (AC-011/012/025, SR-009).
  4. ✅ **DONE (Plan A)** — **`on_leave` status filter** + combined-cell multi-match (AC-005/031, SR-008).
  5. ⏳ **Plan B** — **11 PM auto check-out job** (mobile AC-11/12, Rule 5) — Go has the
     `is_auto_checkout` column but nothing drives it.
  6. ✅ **DONE (Plan A)** — **Leave-aware summary math** — SR-011 half-day-worked boundaries now applied.

- **Built in NEITHER (needs a data source, BA open question):**
  7. **Holiday "H" cells** + streak-excludes-holidays (SR-006, mobile AC-16). No holiday
     calendar exists in either system; blocked.

- **Go-only extras with no current DR (one explicitly out-of-scope):**
  - **Admin manual CRUD** (`POST` / `PATCH /:id` / `DELETE /:id` / `GET /:id`) — the web DR
    SR-001 says the list is "read-only — no attendance data can be modified from this
    page" and §8 lists "Manual attendance entry or modification" as out-of-scope. No DR
    specs an admin edit surface (the mobile DR §8 only *hints* one exists).
  - **`work_location`** enum (office/remote/hybrid/field) — mobile §8 marks WFH out of v1.
  - **`is_half_day`** flipped by an hours threshold at checkout — not a BA rule; BA half-day
    is **leave-driven**, not hours-driven.
  - **`notes`**, **`GET /me`** — harmless conveniences; unspecced.

Two API-shape divergences aren't hard ACs but break a Python-built FE — see **D1/D2**.

Decisions are **OPEN** pending sign-off — see **§5 Decisions**.

---

## 1. Endpoint inventory

| # | Python (`/attendance`) | Go (`/api/v1/attendance`) | Gap |
|---|---|---|---|
| 1 | `POST /check-in` → **TodayStatusRead** | `POST /check-in` → **AttendanceRead** | return-shape differs (D2) |
| 2 | `POST /check-out` → **TodayStatusRead** | `POST /check-out` → **AttendanceRead** | return-shape differs (D2) |
| 3 | `GET /today` → TodayStatusRead | `GET /today` → TodayStatusRead | ✅ parity |
| 4 | `GET ""` (root) → **AttendanceMatrixRead** | `GET ""` (root) → **paginated list** | **same URL, different payload** (D1) |
| 5 | — | `GET /matrix` → AttendanceMatrixRead | Go moved matrix to a sub-path (D1) |
| 6 | `GET /export` → `.xlsx` | ❌ missing | **add** (G3) |
| 7 | `GET /export/{employee_id}` → `.xlsx` | ❌ missing | **add** (G3) |
| 8 | — | `GET /me` (own rows, paginated) | Go-only (keep) |
| 9 | — | `GET /:id` (one row) | Go-only, unspecced |
| 10 | — | `POST ""` admin create | Go-only, **DR-excluded** |
| 11 | — | `PATCH /:id` admin update | Go-only, **DR-excluded** |
| 12 | — | `DELETE /:id` admin delete | Go-only, **DR-excluded** |

Permission constants are **identical** on both sides — `attendance:read`,
`attendance:manage_data` — with the same "has-read-but-not-manage → own row only"
semantics. Go additionally gates its admin CRUD behind `manage_data`. The perm
plumbing is ready; the gap is in the matrix/export/job logic.

---

## 2. Data model comparison

| Field | Python (`AttendanceRecord`/`AttendanceSession`) | Go (`models.Attendance`/`AttendanceSession`) | Notes |
|---|---|---|---|
| id | Mongo ObjectId | UUID (`BaseModel`) | — |
| employee_id | `PydanticObjectId` → users | UUID → **employees(id)** | Go schema split (expected) |
| date | `date`, unique per (emp, date) | `date`, partial-unique per (emp, date) | parity |
| is_late | computed from first check-in | same (first check-in vs threshold) | parity |
| **is_half_day** | **absent** (half-day = leave-driven) | `bool`, flipped by hours threshold | **Go extra** (D5) |
| **work_location** | **absent** | `*string` enum | **Go extra** (WFH out of v1) |
| **notes** | **absent** | `*string` | **Go extra** |
| sessions | embedded `list[AttendanceSession]` | `attendance_sessions` rows (FK + cascade) | structural; parity in effect |
| session.is_auto_checkout | `bool` (driven by 11 PM job) | `bool` (**no job drives it**) | **G5** |
| audit cols / soft delete | `is_deleted` only | full 4 cols + `NotDeleted` + trigger | Go convention |

The matrix response shapes also diverge — Python `AttendanceCellRead` carries
`leave_type`, `leave_period`, `worked_half_status` (none present in Go's DTO).

---

## 3. Business-rule comparison (vs BA acceptance criteria)

| Rule | BA backing | Python | Go today | Verdict |
|---|---|---|---|---|
| GPS haversine + 50 m radius + accuracy gate | Mobile AC-01/03, Rule 1/2 | ✅ | ✅ (mirrors `accuracy > radius`) | parity |
| Check-out needs no GPS | Mobile AC-08 | ✅ | ✅ | parity |
| One open session guard (409) | — | ✅ | ✅ | parity |
| Multiple sessions/day; status from first check-in | Web SR-010, Mobile AC-09/10 | ✅ | ✅ | parity |
| Monthly count = distinct days with check-in | Mobile AC-13, Rule 4 | ✅ | ✅ | parity |
| Streak over workdays, skip today-if-no-checkin | Mobile AC-14/15/17 | ✅ | ✅ | parity |
| **Streak excludes holidays** | Mobile AC-16 | ❌ | ❌ | **G7 — neither** |
| **Auto check-out at 11 PM** | Mobile AC-11/12, Rule 5 | ✅ | ❌ (field only) | **G5 — Go gap** |
| Matrix: weekend / absent / no_data / on_time / late cells | Web AC-013/014/015 | ✅ | ✅ | parity |
| **Matrix: approved leave cells (AL/SL/…)** | Web AC-016, SR-004 | ✅ | ❌ | **G1 — Go gap** |
| **Matrix: combined half-day cells + worked_half_status** | Web AC-026–031, SR-002/003/004/008/011 | ✅ | ❌ | **G2 — Go gap** |
| **Status filter incl. `on_leave` + combined multi-match** | Web AC-005/031, SR-008 | ✅ | ❌ (5 literal statuses) | **G4 — Go gap** |
| Summary cols: full-day late/early minutes | Web AC-019–024, SR-011 | ✅ | ✅ | parity |
| **Summary cols: half-day-worked boundaries** | Web AC-028/029/030, SR-011 | ✅ | ❌ | **G6 — Go gap** (depends on G1/G2) |
| **Holiday "H" cells** | Web SR-006 | ❌ | ❌ | **G7 — neither** |
| **Excel export (bulk + per-employee, summary cols)** | Web AC-011/012/025, SR-009 | ✅ | ❌ | **G3 — Go gap** |
| Web list is **read-only** (no manual edit) | Web SR-001, §8 out-of-scope | ✅ (no CRUD) | ❌ (has admin CRUD) | Go built DR-excluded surface |

Schedule constants (BA confirmed, DR-004 SR-011 / EP-009): workday **09:00–18:00**;
lunch **12:00–13:15**; AM late threshold **09:00**; PM late threshold (when PM is the
worked half) **13:15**; auto check-out **23:00**. Go reads late/checkout thresholds
from config/env; the AM-end (12:00) and PM-late (13:15) half-day constants are
hard-coded in Python (`_AM_END`, `_PM_LATE`) and absent in Go.

---

## 4. Gaps to close (priority order)

- **G1 — Leave-integrated matrix.** Join approved `leave_requests` (EP-002) over the
  month; render AL/SL/PL/ML/UL/½ cell statuses; expand multi-day leave to per-date
  (first-leave-wins). Add `leave_type` / `leave_period` to `AttendanceCellRead`.
  *Largest piece; unblocks G2/G4/G6.*
- **G2 — Combined half-day cells.** When a date has an approved morning/afternoon-half
  leave AND a check-in, compute `worked_half_status` ∈ {on_time, late, absent} against
  the worked half's threshold (PM-worked → 13:15; AM-worked → 09:00). Add
  `worked_half_status` to the DTO.
- **G3 — Excel export.** Two endpoints (`GET /export`, `GET /export/{employee_id}`),
  `.xlsx`, gated `attendance:read`, non-managers limited to self. Columns: Employee,
  Department, day 1..N (glyph + times), Total Late Time, Total Early Time (`Xh Ym`).
  Needs a Go xlsx lib (e.g. `excelize`).
- **G4 — `on_leave` status filter + combined multi-match.** Extend the matrix filter so
  a combined-cell row matches both its worked-half status and `on_leave`.
- **G5 — 11 PM auto check-out job.** Port `auto_check_out_open_sessions(cutoff)`: mark
  open sessions with `check_in < cutoff` as `is_auto_checkout=true`, `check_out=cutoff`.
  Driver = scheduler/cron (decide mechanism — see Phase-7 SSE/cron deferral pattern).
- **G6 — Leave-aware summary math.** Apply SR-011 half-day boundaries (AM-end 12:00 /
  PM-late 13:15) once G1/G2 land; skip full-day-leave and worked-half-absent days.
- **G7 — Holiday handling (BLOCKED).** Needs a holiday calendar source. BA open question
  (DR-004 §8). Defer until EP-009 / system config provides one. Affects matrix "H" cells
  and streak. Track, don't build yet.

---

## 5. Decisions (OPEN — need sign-off before any code)

> Same "contested decision" treatment as the roles audit. Each is a point where
> Python, BA, and Go conventions disagree or a mechanism must be chosen.

- **D1 — Root `GET /attendance` contract.** ✅ **LOCKED → match Python: root `GET /attendance`
  returns the matrix** (`AttendanceMatrixRead`), gated `attendance:read`. Implications:
  - The current `/matrix` sub-path is **folded into the root**; keep `/matrix` as an alias
    for one release (optional) or drop it — confirm with FE, but the canonical path is the root.
  - Python has **no flat-list endpoint**, so the existing Go root list behavior moves out of
    the root. `GET /me` (self rows) stays. If admin still needs a flat paginated list, expose
    it at an explicit sub-path (e.g. `GET /attendance/records`) rather than the root — but it
    is a Go extra with no Python/BA backing, so default to **not** carrying it on the root.
- **D2 — check-in/out response shape.** ✅ **LOCKED → return `TodayStatusRead`** from both
  `POST /check-in` and `POST /check-out` (Python parity; saves the mobile widget a
  `GET /today` round-trip). The service already builds `TodayStatusRead` for `GET /today` —
  reuse it; the bare `AttendanceRead` is no longer returned by the action endpoints.
- **D3 — Excel library.** ✅ **LOCKED → `xuri/excelize/v2`** (add to `go.mod`; styled header
  + `Xh Ym` summary columns per SR-009).
- **D4 — Auto-checkout driver.** ✅ **LOCKED → idempotent service method now, scheduler later.**
  Implement `AttendanceService.AutoCheckOut(ctx, cutoff)` that marks every open session with
  `check_in < cutoff` as `is_auto_checkout=true, check_out=cutoff`. Expose it via an admin
  endpoint (gated `attendance:manage_data`) this round; wire a real 23:00-company-time
  scheduler in a follow-up. Logic ships now, trigger follows.
- **D5 — `is_half_day` semantics.** ✅ **LOCKED → follow BA.** Stop auto-flipping `is_half_day`
  by the hours threshold (the BA half-day is approved-leave-driven, not hours-driven). Half-day
  state in the matrix derives from approved half-day leave (G1/G2), not this column. Flag to BA
  whether the column should be removed entirely or repurposed; until then it is not auto-set.
- **D6 — Admin CRUD endpoints.** ✅ **LOCKED → keep.** Retain `GET /:id`, `POST`, `PATCH /:id`,
  `DELETE /:id` as-is (gated `attendance:manage_data`). They are a Go-only surface with no
  current DR (the web list DR is read-only), so **flag to BA** to back-fill an admin-edit DR;
  until then they stay but are documented as ahead-of-spec. No change required to ship them.

---

## 6. Out of scope for this audit / follow-ups

- Holiday calendar (G7) — blocked on a data source; track in CHECKPOINT outstanding items.
- `work_location` / WFH — mobile §8 marks out of v1; leave the column, don't build flows.
- FE wiring — web repo is self-managed (`api_info_go/`); produce an FE doc after PRs land.
- EP-009 attendance-settings (late/checkout thresholds, office GPS) already partly wired
  via config/env; switching to `system_config` lookup is a separate tracked follow-up
  (CHECKPOINT outstanding item — Phase 6).
