# Holiday Management API — FE Integration Guide

**Date:** 2026-06-16  
**Commits:** `04b4167` → `c56b1db` (migration 000023 through verification fix)  
**Backend:** Go HRM API, base path `/api/v1`. All endpoints require `Authorization: Bearer <access_token>`.  
**Source of truth:** Swagger at `/swagger/index.html`. Drop this into the web repo as `api_info_go/holidays.md`.

---

## What this module does

Company holiday calendar — admins/HR managers define public holidays; the system automatically subtracts those days from leave request `total_days` when a leave falls on a holiday date.

**Side-effect to know about:** When a holiday is created, updated, or deleted, all `approved` leave requests whose date range overlaps the changed holiday range are silently **recalculated** — their `total_days` in the DB updates automatically. The response from Create/Update/Delete includes an `affected` count (returned in the message string, not a separate field).

---

## Permissions

| Permission | String | Who has it |
|---|---|---|
| View holidays | `organization:holidays_view` | Admin, HR Manager, Manager, Employee |
| Manage holidays | `organization:holidays_manage` | Admin, HR Manager only |

---

## Endpoint map

| Method | Path | Permission | Returns |
|---|---|---|---|
| GET | `/holidays` | `holidays_view` | paginated `HolidayRead` |
| POST | `/holidays` | `holidays_manage` | `HolidayRead` (201) |
| PATCH | `/holidays/:id` | `holidays_manage` | `HolidayRead` |
| DELETE | `/holidays/:id` | `holidays_manage` | message only |
| GET | `/holidays/years` | `holidays_view` | `[]int` |
| GET | `/holidays/templates` | `holidays_manage` | `[]HolidayTemplateRead` |
| POST | `/holidays/import` | `holidays_manage` | `HolidayImportResult` |

---

## Response shapes

### `HolidayRead`

```json
{
  "id": "uuid",
  "year": 2026,
  "name": "Tết Nguyên Đán",
  "from_date": "2026-01-28T00:00:00Z",
  "to_date":   "2026-02-03T00:00:00Z",
  "total_days": 7,
  "created_at": "2026-06-16T10:00:00Z",
  "updated_at": "2026-06-16T10:00:00Z"
}
```

`total_days` is **computed** (not stored): `(to_date − from_date).days + 1`. It reflects the calendar span of the holiday, NOT how many leave days it consumes from an employee.

### `HolidayTemplateRead`

```json
{
  "id": "uuid",
  "year": 2026,
  "name": "Tết Nguyên Đán",
  "from_date": "2026-01-28T00:00:00Z",
  "to_date":   "2026-02-03T00:00:00Z",
  "total_days": 7
}
```

Templates are read-only Vietnamese public holiday presets seeded in the DB (5 per year for 2025/2026/2027). They cannot be created, edited, or deleted — only imported.

### `HolidayImportResult`

```json
{
  "imported": 4,
  "skipped":  1
}
```

---

## Endpoints in detail

### `GET /holidays` — List holidays for a year

**Query params:**

| Param | Type | Required | Notes |
|---|---|---|---|
| `year` | int | **yes** | e.g. `2026`. Returns 400 if missing or out of range 2000–2100. |
| `search` | string | no | Case-insensitive name search |
| `page` | int | no | Default: 1 |
| `page_size` | int | no | Default: 20, max 100 |

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [ /* HolidayRead[] sorted by from_date ASC */ ],
    "total": 5,
    "page": 1,
    "page_size": 20,
    "total_pages": 1
  }
}
```

---

### `POST /holidays` — Create a holiday

**Body:**
```json
{
  "year": 2026,
  "name": "Liberation Day",
  "from_date": "2026-04-30T00:00:00Z",
  "to_date":   "2026-04-30T00:00:00Z"
}
```

All four fields are **required**. `to_date` must be ≥ `from_date` (400 otherwise).

**Success (201):**
```json
{
  "success": true,
  "message": "Holiday has been created",
  "data": { /* HolidayRead */ }
}
```

**Errors:**
- `400` — missing/invalid fields, or `to_date < from_date`
- `409` — a holiday with this name already exists in this year (case-insensitive)

---

### `PATCH /holidays/:id` — Partial update

All fields optional. Only include what you're changing.

**Body:**
```json
{
  "name":      "Liberation & Reunification Day",
  "from_date": "2026-04-30T00:00:00Z",
  "to_date":   "2026-05-01T00:00:00Z"
}
```

**Success (200):**
```json
{
  "success": true,
  "message": "Holiday has been updated",
  "data": { /* HolidayRead with updated values */ }
}
```

**Errors:**
- `400` — invalid fields or resulting `to_date < from_date`
- `404` — holiday not found
- `409` — new name conflicts with an existing holiday in the same year

---

### `DELETE /holidays/:id` — Soft-delete a holiday

No body required.

**Success (200):**
```json
{
  "success": true,
  "message": "Holiday has been deleted"
}
```

When the deletion triggers leave recalculation the message is:
```json
{
  "success": true,
  "message": "Holiday deleted. 3 leave request(s) recalculated."
}
```

**Errors:**
- `404` — holiday not found

---

### `GET /holidays/years` — List years with holidays

Returns all years that have at least one holiday, **always including the current year** even if it has none yet.

**Response:**
```json
{
  "success": true,
  "data": [2025, 2026, 2027]
}
```

Use this to populate the year selector on the holiday calendar screen.

---

### `GET /holidays/templates` — List Vietnamese holiday presets

**Query params:**

| Param | Type | Required | Notes |
|---|---|---|---|
| `year` | int | **yes** | Must be 2000–2100. Returns 400 otherwise. |

**Response:**
```json
{
  "success": true,
  "data": [ /* HolidayTemplateRead[] sorted by from_date ASC */ ]
}
```

Currently seeded years: **2025, 2026, 2027** — 5 templates each (Tết Dương Lịch, Tết Nguyên Đán, Giỗ Tổ Hùng Vương, Ngày Giải Phóng & Quốc Tế Lao Động, Quốc Khánh).

Note: this endpoint is gated `holidays_manage` (not view), so it's only shown to Admin/HR Manager.

---

### `POST /holidays/import` — Import presets

**Body:**
```json
{
  "year": 2026,
  "template_ids": [
    "uuid-1",
    "uuid-2",
    "uuid-3"
  ]
}
```

Both fields **required**. `template_ids` must be non-empty.

Skips any template whose name **already exists** (case-insensitive) as a holiday in that year. Reports the skip count.

**Typical flow:**
1. Call `GET /holidays/templates?year=2026` — show checkbox list of presets.
2. User selects all or a subset.
3. POST their IDs to `/holidays/import` with the target year.

**Success (200):**
```json
{
  "success": true,
  "message": "4 holiday(s) imported for 2026, 1 skipped (already exist)",
  "data": {
    "imported": 4,
    "skipped":  1
  }
}
```

When nothing is skipped, the message omits the skipped clause:
```json
{
  "message": "5 holiday(s) imported for 2026",
  "data": { "imported": 5, "skipped": 0 }
}
```

**Errors:**
- `400` — missing `year`, empty `template_ids`, or invalid UUID in list

---

## Leave recalculation — what the FE should know

When a holiday is **created, updated, or deleted**, approved leave requests that overlap the changed date range are automatically recalculated in the background (same request, synchronous). Their `total_days` in the DB updates silently.

**Practical impact for the FE:**
- If the user creates a holiday while a leave-requests list is open, those leave `total_days` values may be stale in the FE cache. Recommend **invalidating / refetching** leave-request lists after any successful holiday mutation.
- The holiday endpoints return the count of affected leaves in the response `message` string only (not a machine-readable field). Parse it if you want to show a toast, or just refresh leaves unconditionally.

---

## Date format

All dates are **ISO 8601 UTC** strings: `"2026-04-30T00:00:00Z"`.

Holidays are date-only (no time component), but the server accepts and returns them in full RFC 3339 form with midnight UTC. When sending create/update payloads, use midnight UTC (T00:00:00Z).

---

## Suggested screen flow

```
Holiday Calendar screen (Admin / HR Manager only for mutations)
│
├── Year selector ← populated from GET /holidays/years
│
├── Holiday list ← GET /holidays?year=<selected>
│     ├── [Edit] → PATCH /holidays/:id
│     └── [Delete] → DELETE /holidays/:id  (+ refresh leave list)
│
├── [+ Add holiday] → POST /holidays  (+ refresh leave list)
│
└── [Import presets] modal
      ├── GET /holidays/templates?year=<selected>  (checkbox list)
      └── [Import selected] → POST /holidays/import  (+ refresh list)
```

Manager and Employee roles see the list (view) but not the mutation buttons.

---

## Quick FE checklist

- [ ] Wire year selector to `GET /holidays/years` (always pre-selects current year).
- [ ] Holiday list calls `GET /holidays?year=<year>` — `year` param is **required**.
- [ ] Create/Update forms validate `to_date ≥ from_date` client-side (server also returns 400).
- [ ] Show 409 error as "A holiday with this name already exists for this year" (server message is similar).
- [ ] After any holiday Create / Update / Delete, **invalidate the leave-requests cache** so `total_days` values refresh.
- [ ] Import flow: fetch templates first (with year), show checkboxes, post selected `template_ids`.
- [ ] Delete success: parse the message to conditionally show "N leave requests were recalculated" info text, OR just always show a generic success + refresh.
- [ ] `total_days` on `HolidayRead` = calendar span of the holiday (not leave days consumed) — display as "X day(s)" in the UI.
- [ ] Gate mutation buttons (`+ Add`, Edit, Delete, Import) behind `organization:holidays_manage`.
