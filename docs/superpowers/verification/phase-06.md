# Phase 6 — Attendance: End-to-End Verification Log

**Date:** 2026-05-20
**Branch:** `main`
**Migration version:** `9`
**Server:** `go run ./cmd/server` (PID 48659 in this session)
**Base URL:** `http://localhost:8080/api/v1`

---

## Summary

20 end-to-end steps exercising the 10 attendance endpoints + 1 DB
spot-check. All green after the load-bearing repo fix (see "Repo fix
surfaced during verification" below). Phase 6 patterns:

- `employee_id` is the FK on `attendance` (NOT `user_id`) — verified at
  the DB level (`SELECT employee_id FROM attendance`).
- `is_late` is computed once from the first check-in; subsequent
  sessions don't recompute it (verified by unit test
  `TestAttendance_IsLate_NotRecomputed_OnSecondSession`).
- The partial unique index `uq_attendance_sessions_one_open` is
  defended at the service layer via `FindOpenSession()` — double
  check-in / check-out returns a clean 409.
- Non-admin callers to `GET /attendance` are silently scoped to own
  `employee_id` (Python contract). 403-only paths are `POST/PATCH/DELETE
  /attendance` (require `attendance:manage_data`).
- Two-table soft delete works: `attendance.is_deleted=t,
  deleted_at IS NOT NULL`; subsequent reads via `Get` return 404.

---

## Endpoints exercised

| # | Endpoint | Auth | Status | Notes |
|---|---|---|---|---|
| 1 | `POST /auth/login` (super admin) | – | 200 | seeded |
| 2 | `POST /employees` × 2 | admin | 201 × 2 | walker + second |
| 3 | `POST /auth/login` (walker) | – | 200 | |
| 4 | `POST /attendance/check-in` | walker | 200 | `is_late=true` (15:08 HCM > 09:00 threshold) |
| 5 | `POST /attendance/check-in` (double) | walker | **409** | `"You are already checked in"` |
| 6 | `POST /attendance/check-out` | walker | 200 | `is_half_day=true` (0.01h < 4h threshold) |
| 7 | `POST /attendance/check-out` (double) | walker | **409** | `"You are not currently checked in"` |
| 8 | `GET /attendance/today` | walker | 200 | `status=checked_out, monthly_count=1, streak=1` (after repo fix) |
| 9 | `GET /attendance/me` | walker | 200 | `total=1`, own row only |
| 10 | `GET /attendance` (walker, no manage) | walker | **200** | scoped to self (1 row, own `employee_id`) |
| 11 | `GET /attendance` (admin) | admin | 200 | sees both rows |
| 12 | `POST /attendance` (manual create) | admin | 201 | second user, 2026-05-10 |
| 13 | `PATCH /attendance/:id` | admin | 200 | `is_half_day` + `notes` updated |
| 14 | `GET /attendance` (admin) after create | admin | 200 | `total=2`, dates `[2026-05-20, 2026-05-10]` |
| 15 | `GET /attendance?start_date=2026-05-15&end_date=2026-05-31` | admin | 200 | `total=1`, only `2026-05-20` |
| 16 | `GET /attendance/matrix?month=5&year=2026` | admin | 200 | `year=2026, month=5, days_in_month=31, total_rows=8` |
| 17 | `DELETE /attendance/:id` | admin | 200 | soft delete |
| 18 | `GET /attendance/:id` (deleted) | admin | **404** | `"Attendance not found"` |
| 19 | `POST /attendance/check-in` (no token) | – | **401** | `"Could not validate credentials"` |
| 20 | psql spot-check | – | – | see "DB spot-check" below |

> Step 10 deviates from the original plan body, which predicted **403**.
> The REVISION NOTES + the seeded `Employee` role (which carries
> `PermAttendanceRead`) place the route behind `RequirePerms(PermAttendanceRead)`;
> non-managers pass the route gate and the service silently filters them
> to own rows. This matches the Python contract ("managers see all,
> non-managers see only their own row").

---

## Repo fix surfaced during verification

Step 8 (`GET /attendance/today`) failed with **500** on the first pass:

```
column reference "is_deleted" is ambiguous (SQLSTATE 42702)
```

The `MonthlyCheckInCount` and `DatesWithCheckIn` repo methods join
`attendance_sessions s` (which carries its own `is_deleted`) and were
applying the unqualified `models.NotDeleted` scope (`WHERE is_deleted =
?`). Postgres flagged the ambiguity. The `List` method has the same
shape when `DepartmentID` is filtered (joins `employees e`) — same
latent bug.

**Fix** (`internal/repositories/attendance_repo.go`):
- `MonthlyCheckInCount` + `DatesWithCheckIn`: replace `r.base(ctx)`
  (which applies `NotDeleted`) with `r.db.WithContext(ctx).Where(
  "attendance.is_deleted = ?", false)`.
- `List`: same change — qualifies the soft-delete predicate to
  `attendance.is_deleted` instead of relying on `NotDeleted`.

Two regression tests added:

- `TestAttendance_Today_AfterCheckOut` — runs the full check-in / check-out
  cycle then calls `Today()`, asserting no 500.
- `TestAttendance_List_DepartmentFilter` — creates one employee with a
  department and one without, then filters by `DepartmentID` and
  asserts only the matching row surfaces.

After the fix, all 22 attendance service tests pass and the e2e walk
completes cleanly.

---

## DB spot-check

```
                  id                  |             employee_id              |    date    | is_late | is_half_day | is_deleted | has_deleted_at
--------------------------------------+--------------------------------------+------------+---------+-------------+------------+----------------
 3b90551a-f7fb-4d7d-9769-b05fa006ccec | 9480188b-258a-4695-bcdb-8f524ee4aebc | 2026-05-20 | t       | t           | f          | f
 5c070877-8f1c-4075-b4c4-c43838a86b44 | 5f746b6a-a491-4fa9-aa24-a585a51ef9b0 | 2026-05-10 | t       | t           | t          | t
```

- The live walker row (2026-05-20) preserves `is_deleted=f`.
- The admin-deleted manual entry (2026-05-10, soft-deleted in step 17)
  carries `is_deleted=t` and `deleted_at IS NOT NULL` — psql confirms
  the soft-delete semantics, and the API `Get` returns 404 (step 18).

Sessions table:

```
            attendance_id             |    date    | sessions |           first_in            |           last_out
--------------------------------------+------------+----------+-------------------------------+-------------------------------
 3b90551a-f7fb-4d7d-9769-b05fa006ccec | 2026-05-20 |        1 | 2026-05-20 08:08:02.790902+00 | 2026-05-20 08:08:22.512837+00
```

The soft-deleted attendance row's sessions are excluded from the
`is_deleted=false` filter (no row for `5c070877-…` in this list).

---

## Test summary

```
$ go test ./internal/services -run 'TestAttendance_' -count=1
ok  github.com/exnodes/hrm-api/internal/services  12.157s
```

22 attendance tests, all PASS. Full service suite (incl. Phases 1-5)
remains green.

## What's deferred

- **Excel export** (`/attendance/export`, `/attendance/export/{id}`) —
  Python has it; scope-bounded out of Phase 6.
- **Auto-checkout cron** (closes open sessions at 23:00 company TZ).
  The `OpenSessionsBefore(ctx, cutoff)` repo method is wired so the
  cron has a hook when introduced.
- **`system_config`-backed late threshold** — Phase 8 will introduce
  the row; until then env vars (`LATE_THRESHOLD_HOUR/_MINUTE`,
  `COMPANY_TIMEZONE`) apply.
- **Half-day cells from approved leave** in the matrix — overlap with
  Phase 5 leave matrix; deferred to a Phase 6.5 refinement.
