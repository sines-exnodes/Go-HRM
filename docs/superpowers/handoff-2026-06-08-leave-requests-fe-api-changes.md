# Leave Requests API — Changes for the FE Team

**Date:** 2026-06-08
**Commits:** `0132dd8` (Plan A — permission split) · `d632a69` (Plan B — bug fixes + export)
**Backend:** Go HRM API, base path `/api/v1`. All endpoints require `Authorization: Bearer <access_token>` unless noted.
**Source of truth:** regenerated Swagger at `/swagger/index.html` (this doc is the human summary). Drop this into the web repo `api_info_go/leave_requests.md`.

---

## ⚠️ Breaking changes — read first

| # | Endpoint | Before | After | FE action |
|---|---|---|---|---|
| **B1** | `POST /:id/cancel` | `data` was a `LeaveRequestRead` object | `data` is `LeaveRequestRead` **+** `was_approved: bool` (flattened into the same object) | Read `data.was_approved` to know whether to show a "quota restored" toast. Shape is a superset — existing field reads still work. |
| **B2** | Attachment upload (Create / Update) | Max size 10 MB | Max size **5 MB** | Update any client-side size validation and error-message copy to "5 MB". |
| **B3** | Role permission assignments | `leave_requests:approve` was the single approve permission | Split into **`leave_requests:approve_team`** and **`leave_requests:approve_all`** | Update the role-management screen to display the two new items and stop showing the legacy `approve` entry as an assignable option (it still works at runtime for existing assignments but should not appear in the picker). |

Everything else below is additive and backward-compatible.

---

## Endpoint map (full leave-requests surface)

| Method | Path | Returns | Permission |
|---|---|---|---|
| GET | `/leave-requests` | paginated `LeaveRequestRead` | `leave_requests:read` |
| POST | `/leave-requests` | `LeaveRequestWriteResult` | `leave_requests:create` |
| **GET** | **`/leave-requests/export`** | `.xlsx` file | `leave_requests:read` (**new**) |
| GET | `/leave-requests/dashboard/me` | `LeaveDashboardRead` (10 upcoming + 10 history) | authenticated |
| GET | `/leave-requests/history/me` | paginated `LeaveRequestRead` | authenticated |
| GET | `/leave-requests/balance/:employee_id` | `LeaveBalanceSummary` | `leave_requests:read` |
| GET | `/leave-requests/:id` | `LeaveRequestRead` | `leave_requests:read` |
| PATCH | `/leave-requests/:id` | `LeaveRequestWriteResult` | `leave_requests:update` |
| POST | `/leave-requests/:id/approve` | `LeaveRequestRead` | `approve_team` **or** `approve_all` (checked in handler — see §Approve/Reject) |
| POST | `/leave-requests/:id/reject` | `LeaveRequestRead` | `approve_team` **or** `approve_all` |
| POST | `/leave-requests/:id/cancel` | `LeaveRequestRead` + `was_approved` (**changed** — see B1) | `leave_requests:cancel` |
| POST | `/leave-requests/:id/delete` | `{}` soft-delete | `leave_requests:delete` |

---

## Permission split — Approve / Reject (B3 in detail)

### Old behaviour
A single `leave_requests:approve` permission gated both `/approve` and `/reject` at the router level. Any user with that permission could approve any employee's request.

### New behaviour
The permission check moved **into the handler**. Two scoped variants exist:

| Permission | What it allows |
|---|---|
| `leave_requests:approve_team` | Approve/reject requests from employees in your **transitive reporting chain** (BFS from your employee record via `line_manager_id`). Attempting to approve a non-subordinate returns `403`. |
| `leave_requests:approve_all` | Approve/reject **any** employee's request, regardless of reporting line. |

The legacy `leave_requests:approve` constant is preserved at runtime and is treated as `approve_all` — existing role assignments that still carry it keep working. However it will not appear in the role-management permission picker (`GET /roles/permissions`) anymore. When updating roles, assign one of the two new scoped permissions instead.

### Seeded defaults (after next boot)
| Role | Leave approve permission |
|---|---|
| Super Admin | `*` (wildcard) |
| Admin | `leave_requests:approve_all` |
| HR Manager | `leave_requests:approve_all` |
| Manager | `leave_requests:approve_team` |
| Employee | — |

Manager also gained `leave_requests:update` and `leave_requests:delete` (was missing these).

### Error responses
`403 Forbidden` — returned when:
- The caller has neither `approve_team` nor `approve_all` (nor `*`).
- The caller has only `approve_team` and tries to approve a request belonging to someone outside their reporting chain.

---

## Cancel — `POST /leave-requests/:id/cancel` (B1 in detail)

### Before
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "cancelled",
    /* …other LeaveRequestRead fields… */
  }
}
```

### After
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "cancelled",
    /* …other LeaveRequestRead fields… */
    "was_approved": true
  }
}
```

`was_approved` is `true` when the request was in `approved` state before it was cancelled. Use this to decide whether to show a notice that the employee's leave quota has been restored. When `was_approved: false` the request was still `pending` — no quota was consumed.

---

## Excel export — `GET /leave-requests/export` (new)

Returns a streamed `.xlsx` file.

**Headers in response:**
- `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- `Content-Disposition: attachment; filename="leave-requests.xlsx"`

**Query params** (same as the list endpoint):

| Param | Type | Notes |
|---|---|---|
| `status` | `[]string` | Repeated param — e.g. `?status=pending&status=approved`. Valid values: `pending`, `approved`, `rejected`, `cancelled`. |
| `department_id` | uuid | Filter by department. |
| `position_id` | uuid | Filter by position. |
| `search` | string | Employee name search. |

**Permission:** `leave_requests:read`. Non-managers see only their own requests; admin/manager see all.

**Columns (in order):**
`Employee | Department | Position | Leave Type | From Date | To Date | Period | Total Days | Reason | Status | Created At`

**FE:** trigger a download — e.g. open the URL with the auth token, or use a `blob` download via `fetch`. No JSON envelope — the body is the file.

---

## Attachment changes (B2 in detail)

| | Before | After |
|---|---|---|
| Max size | 10 MB | **5 MB** |
| Accepted types | PDF, PNG, JPEG, GIF, WEBP | PDF, PNG, JPEG, GIF, WEBP, **DOCX** |
| Error message | "Attachment must not exceed 10MB" | "Attachment must not exceed 5MB" |

Client-side: update your `maxSize` guard to `5 * 1024 * 1024` bytes and add `.docx` to the `accept` attribute on the file input.

DOCX detection note: the server sniffs file content, not the `Content-Type` header. A `.docx` file is a valid Word document as long as the file extension is `.docx` — the server will detect the Office Open XML container correctly.

---

## Dashboard — `GET /leave-requests/dashboard/me`

Upcoming and history lists now return **10 items each** (was 5). No shape change. If you were hardcoding a "show 5" assumption in the FE, the extra items will now appear.

### Response shape (unchanged)
```json
{
  "success": true,
  "data": {
    "upcoming": [ /* up to 10 LeaveRequestRead, status pending/approved, from_date >= today */ ],
    "history":  [ /* up to 10 LeaveRequestRead, to_date < today */ ],
    "balance": { /* LeaveBalanceSummary for current year */ }
  }
}
```

---

## LeaveRequestRead shape (reference)

No new fields added. Shown here for completeness.

```json
{
  "id": "uuid",
  "employee":   { "id": "uuid", "name": "Jane Smith" },
  "department": { "id": "uuid", "name": "Engineering" },
  "position":   { "id": "uuid", "name": "Senior Engineer" },
  "from_date":  "2026-08-01T00:00:00Z",
  "to_date":    "2026-08-03T00:00:00Z",
  "leave_period": "full_day",
  "leave_type": "annual",
  "total_days": 3.0,
  "reason": "Family vacation",
  "attachment_url": "https://… | null",
  "status": "pending",
  "created_by": "uuid",
  "created_at": "2026-07-15T10:00:00Z",
  "updated_at": "2026-07-15T10:00:00Z"
}
```

`leave_period` values: `full_day` · `morning_half` · `afternoon_half`

`leave_type` values: `annual` · `sick` · `personal` · `maternity` · `unpaid`

`status` values: `pending` · `approved` · `rejected` · `cancelled`

---

## Not in this release

- **Leave quota carry-forward** — no carry-forward endpoint. The balance endpoint (`GET /leave-requests/balance/:employee_id`) computes used days dynamically from approved requests in the current calendar year.
- **Bulk approve** — not implemented. One request per call.
- **Manager-approval notification** — no push/email trigger on approve/reject yet. Phase 9 push infrastructure is available; wiring to leave events is a follow-up.

---

## Quick FE checklist

- [ ] Update file-input `accept` to include `.docx`; update client-side size guard to **5 MB** (B2).
- [ ] Read `data.was_approved` from the cancel response to conditionally show a "quota restored" toast (B1).
- [ ] In the role-management permission picker, show `leave_requests:approve_team` and `leave_requests:approve_all` instead of the legacy `leave_requests:approve`. The legacy item will not appear in `GET /roles/permissions` (B3).
- [ ] Wire a "Download Excel" button on the leave-requests admin list to `GET /leave-requests/export` (file download, same filters as the list) (new).
- [ ] Dashboard: if you were capping the upcoming/history lists at 5 client-side, remove the cap — the server now returns up to 10 (additive).
