# Attendance API — Changes for the FE Team

**Date:** 2026-06-08
**Branch / PR:** `worktree-feat+attendance-parity` → [PR #16](https://github.com/sines-exnodes/Go-HRM/pull/16)
**Backend:** Go HRM API, base path `/api/v1`. All endpoints require `Authorization: Bearer <access_token>` unless noted.
**Source of truth:** regenerated Swagger at `/swagger/index.html` (this doc is the human summary). Drop this into the web repo `api_info_go/attendance.md`.

---

## ⚠️ Breaking changes — read first

| # | Endpoint | Before | After | FE action |
|---|---|---|---|---|
| **B1** | `GET /api/v1/attendance` (root) | flat **list** of attendance rows | **monthly matrix** (`AttendanceMatrixRead`) | If you were calling the root for a list, switch to **`GET /api/v1/attendance/records`**. The root + `GET /api/v1/attendance/matrix` now both return the matrix. |
| **B2** | `POST /api/v1/attendance/check-in` | returned the attendance **row** (`AttendanceRead`) | returns **`TodayStatusRead`** (today-widget payload) | Update the check-in handler to read `status`/`sessions`/`monthly_count`/`streak` straight from the response — no follow-up `GET /today` needed. |
| **B3** | `POST /api/v1/attendance/check-out` | returned `AttendanceRead` | returns **`TodayStatusRead`** | Same as B2. |
| **B4** | `is_half_day` on the attendance row | auto-set `true` when a worked day was shorter than the threshold | **no longer auto-set by hours** | Don't rely on `is_half_day` for "short day". Half-day is now driven exclusively by **approved half-day leave**, surfaced in the **matrix cell** (`status: "half_day_leave"` + `worked_half_status`). |

Everything else below is additive (new endpoints, new response fields) and backward-compatible.

---

## Endpoint map (full attendance surface)

| Method | Path | Returns | Permission |
|---|---|---|---|
| POST | `/attendance/check-in` | `TodayStatusRead` | authenticated |
| POST | `/attendance/check-out` | `TodayStatusRead` | authenticated |
| GET | `/attendance/today` | `TodayStatusRead` | authenticated |
| GET | `/attendance/me` | paginated `AttendanceRead` (own rows) | authenticated |
| **GET** | **`/attendance`** | **`AttendanceMatrixRead`** (matrix) | `attendance:read` |
| GET | `/attendance/matrix` | `AttendanceMatrixRead` (alias of root) | `attendance:read` |
| **GET** | **`/attendance/records`** | paginated `AttendanceRead` (flat list) | `attendance:read` |
| GET | `/attendance/{id}` | `AttendanceRead` | `attendance:read` (owner or manager) |
| **GET** | **`/attendance/export`** | `.xlsx` file | `attendance:read` |
| **GET** | **`/attendance/export/{employee_id}`** | `.xlsx` file | `attendance:read` |
| POST | `/attendance` | `AttendanceRead` (admin manual create) | `attendance:manage_data` |
| PATCH | `/attendance/{id}` | `AttendanceRead` (admin edit) | `attendance:manage_data` |
| DELETE | `/attendance/{id}` | `{}` (soft delete) | `attendance:manage_data` |
| **POST** | **`/attendance/auto-checkout`** | `{ closed: N }` | `attendance:manage_data` |

**Visibility rule** (matrix, list, get): a caller with `attendance:read` but **without** `attendance:manage_data` sees **only their own** row/records; with `attendance:manage_data` they see **all** employees. Check-in/out/today/me are always self-scoped.

---

## The matrix — `GET /api/v1/attendance` (and `/matrix`)

### Query params
| Param | Type | Default | Notes |
|---|---|---|---|
| `month` | int 1–12 | current month (company TZ) | |
| `year` | int ≥2000 | current year | |
| `page` | int ≥1 | 1 | rows are paginated after building |
| `page_size` | int 1–100 | 20 | |
| `search` | string | — | employee-name filter (managers only) |
| `department_id` | uuid | — | managers only |
| `status` | CSV string | — | filter rows by cell status — see below |

### `status` filter values (CSV, e.g. `status=late,on_leave`)
`on_time` · `late` · `absent` · **`on_leave`** (matches any leave-type cell) — a row is returned if **any** of its cells matches **any** selected value. A combined half-day cell matches **both** its `worked_half_status` (e.g. `late`) **and** `on_leave` simultaneously.

### Response — `AttendanceMatrixRead`
```json
{
  "success": true,
  "data": {
    "year": 2026,
    "month": 4,
    "days_in_month": 30,
    "items": [ /* AttendanceRowRead[] */ ],
    "total": 12,
    "page": 1,
    "page_size": 20,
    "total_pages": 1
  }
}
```

### `AttendanceRowRead`
```json
{
  "employee_id": "uuid",
  "employee_name": "Jane Smith",
  "avatar_url": "https://… | null",
  "department_name": "Engineering | null",
  "cells": { "1": { /* AttendanceCellRead */ }, "2": { … }, "30": { … } },
  "total_late_minutes": 18,
  "total_early_minutes": 0
}
```
- `cells` is a **map keyed by day-of-month** (`"1"`..`"N"`), not an array. Iterate `1..days_in_month`.
- `total_late_minutes` / `total_early_minutes` are the monthly sums. Render as `Xh Ym` (e.g. 18 → `0h 18m`). These are **leave-aware** (half-day-worked days use the worked half's boundary; full-day-leave/weekend/absent contribute 0).

### `AttendanceCellRead` — **new fields**
```json
{
  "date": "2026-04-09",
  "day": 9,
  "status": "half_day_leave",
  "check_in": "2026-04-09T06:25:00Z",       // optional, UTC — convert to company TZ for display
  "check_out": "2026-04-09T11:00:00Z",      // optional
  "hours_worked": 4.58,                      // optional, sum across sessions
  "is_late": false,
  "leave_type": "annual",                    // NEW — present on leave cells
  "leave_period": "morning_half",            // NEW — full_day | morning_half | afternoon_half
  "worked_half_status": "late",              // NEW — only on half_day_leave cells: on_time | late | absent
  "sessions": [ /* AttendanceSessionRead[] */ ]  // optional
}
```

### `status` enum (cell) — **expanded**
| Status | Meaning | Suggested glyph |
|---|---|---|
| `on_time` | present, on time | ✓ |
| `late` | first check-in after the late threshold | L |
| `absent` | past workday, no record, no leave | A |
| `weekend` | Sat/Sun | — |
| `no_data` | today/future workday, nothing yet | (blank) |
| `annual_leave` | full-day approved annual leave | AL |
| `sick_leave` | full-day sick leave | SL |
| `personal_leave` | full-day personal leave | PL |
| `maternity_leave` | full-day maternity leave | ML |
| `unpaid_leave` | full-day unpaid leave | UL |
| **`half_day_leave`** | half-day leave (see combined-cell rules) | ½ |

### Combined cells (half-day leave + a worked half)
When `status == "half_day_leave"`:
- `leave_period` tells you which half is on leave: `morning_half` (AM on leave, PM worked) or `afternoon_half` (AM worked, PM on leave).
- `worked_half_status` is the worked half's outcome: `on_time` | `late` | `absent` (`absent` = no check-in on the worked half).
- `leave_type` is the underlying leave type (annual/sick/…) for the tooltip; the glyph stays `½`.
- If a record exists for the worked half, `check_in`/`check_out`/`hours_worked`/`sessions` are populated.

Render this as the diagonal split per DR-004-001-01 §4 (AM = top-left, PM = bottom-right). `worked_half_status` drives the worked-half glyph; `leave_period` decides which corner is the `½`.

---

## Excel export — `GET /attendance/export` and `/attendance/export/{employee_id}`

- Returns a streamed **`.xlsx`** (`Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `Content-Disposition: attachment; filename="attendance[…].xlsx"`).
- Query params: `month`, `year` (same defaults as the matrix).
- `/export` → all visible employees (managers: all; non-managers: just themselves).
- `/export/{employee_id}` → one employee; a non-manager may only export **themselves** (else `403`).
- Columns: `Employee | Department | 1 | 2 | … | N | Total Late Time | Total Early Time`. Totals match the on-screen matrix exactly (same `Xh Ym` format).
- FE: trigger a download (e.g. open the URL with the auth header / use a blob download). No JSON envelope — it's a file body.

---

## Check-in / check-out / today — `TodayStatusRead`

`POST /attendance/check-in`, `POST /attendance/check-out`, and `GET /attendance/today` all return:
```json
{
  "success": true,
  "message": "Checked in",                 // or "Checked out"; absent on /today
  "data": {
    "status": "checked_in",                // not_checked_in | checked_in | checked_out
    "is_late": true,
    "sessions": [ /* AttendanceSessionRead[] */ ],
    "current_check_in": "2026-…Z | null",  // open session's check-in, for the live work timer
    "monthly_count": 7,                      // distinct days with a check-in this month
    "streak": 3                              // consecutive workdays (Mon–Fri) with a check-in
  }
}
```
`AttendanceSessionRead`:
```json
{ "id": "uuid", "check_in": "…Z", "check_out": "…Z | null", "is_auto_checkout": false, "hours_worked": 8.5 }
```

### Check-in request body
```json
{ "check_in": "optional ISO ts", "work_location": "office|remote|hybrid|field (optional)",
  "notes": "optional", "latitude": 0.0, "longitude": 0.0, "accuracy": 0.0 }
```
GPS fields are only validated when the server has `OFFICE_GPS_ENABLED=true` (haversine within the configured radius). Check-out body is optional (`{ "check_out": "…", "notes": "…" }`) and defaults to "now".

**Errors:** `409` "You are already checked in" (open session exists) / `409` "You are not currently checked in" (check-out with no open session) / `400` "No check-in found for today".

---

## Auto check-out (admin) — `POST /attendance/auto-checkout`

Closes all still-open sessions whose check-in precedes a cutoff (stamps `check_out` + `is_auto_checkout=true`). Idempotent.
- Query: `cutoff` (optional, RFC3339) — defaults to now.
- Response: `{ "success": true, "message": "Auto check-out complete", "data": { "closed": 3 } }`
- Requires `attendance:manage_data`. (A real 23:00-company-time scheduler is a backend follow-up; for now this endpoint is the trigger.)

---

## Flat list — `GET /attendance/records` (was the old root)

Paginated `AttendanceRead`. Query: `page`, `page_size`, `employee_id`, `department_id`, `start_date`, `end_date` (`YYYY-MM-DD`), `status` (`on_time|late`). Non-managers are scoped to self. This is the **Go-only** flat view (the matrix is the BA-canonical screen) — use it only if you specifically need a row list.

`AttendanceRead` shape:
```json
{
  "id": "uuid", "employee_id": "uuid",
  "employee": { "id": "uuid", "full_name": "Jane Smith", "avatar_url": "…",
                "department": { "id": "uuid", "name": "Engineering" },
                "position":   { "id": "uuid", "name": "Senior Eng" } },
  "date": "2026-04-09", "is_late": false, "is_half_day": false,
  "work_location": "office | null", "notes": "… | null",
  "sessions": [ /* AttendanceSessionRead[] */ ],
  "check_in": "…Z", "check_out": "…Z | null", "hours_worked": 8.5,
  "created_at": "…Z", "updated_at": "…Z"
}
```

---

## Not in this release (so FE doesn't wait on them)
- **Holiday cells** (`H` status) and holiday-aware streak — not implemented (no holiday-calendar source yet). Treat unknown future statuses defensively.
- The 23:00 auto check-out **scheduler** — logic exists, automatic trigger is a follow-up; the admin endpoint is available now.
- Admin manual CRUD (`POST`/`PATCH`/`DELETE /attendance`) exists but is not yet covered by a BA screen spec — coordinate with BA before building UI for it.

---

## Quick FE checklist
- [ ] Point the attendance-list screen at `GET /attendance` (matrix) — **not** expecting a flat list anymore (B1).
- [ ] Read check-in/out widget state from the **action response** (`TodayStatusRead`), drop the extra `GET /today` round-trip (B2/B3).
- [ ] Render the 11 cell statuses incl. the 5 leave types + `half_day_leave`; implement the combined-cell split using `leave_period` + `worked_half_status` (B4).
- [ ] Add the `on_leave` value to the status filter chips.
- [ ] Wire the two `/export` buttons (bulk + per-row gear) to file downloads.
- [ ] Stop treating `is_half_day` on the row as a short-day flag.
