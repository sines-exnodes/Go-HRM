# Frontend API Changes — Employee Parity Round 2

**Date:** 2026-06-01 · **Backend branch:** `feat/employees-parity-2` · **Migration version:** 19
**Backend verification:** [verification/employees-parity-2.md](verification/employees-parity-2.md) · **Spec:** [specs/2026-06-01-employee-api-parity-2-design.md](specs/2026-06-01-employee-api-parity-2-design.md)

This is the FE-facing summary of the employee/user API changes. Copy the relevant
parts into the web repo's `api_info_go/employee.md` / `me.md`.

## TL;DR

| # | Change | Type | FE action |
|---|---|---|---|
| 1 | `full_name` → **`first_name` + `last_name`** on employee/user create, update, self-update, read, and the login summary | **BREAKING** | Split name inputs; read first/last; compose for display |
| 2 | List **filters now work + are multi-select** (`department_id`/`position_id`/`role_id`/`manager_id` as repeated params → `IN`) | Fix + additive | Send repeated params; multi-select chips supported |
| 3 | **`skill_ids` accepted inline** on create/update | Additive | Send `skill_ids` in the create/update payload (no separate call needed) |
| 4 | **`experience_year` is a 4-digit calendar year** (career start), not a count | **BEHAVIOR** | Send/display a year (e.g. `2018`), validated `>1900 && ≤ current year` |
| 5 | Employee read now includes resolved **`department` / `position`** objects | Additive | Use `read.department.name` / `read.position.name` (no longer null) |
| 6 | **Direct reports** endpoint unchanged | — | No change |

---

## 1. Name split — `full_name` → `first_name` + `last_name`  (BREAKING)

The employee profile now stores the name as two fields, matching the BA forms
(First Name / Last Name columns) and the Python source.

**Requests** — `POST /api/v1/employees`, `PATCH /api/v1/employees/{id}`, `PATCH /api/v1/employees/me`:
- `full_name` is **no longer accepted**. Send `first_name` + `last_name`.
- On **create**, both are **required** (min 1, max 100 chars each).
- On **update / self-update**, both are optional (omit to leave unchanged; min 1, max 100 if sent).

```jsonc
// BEFORE
{ "email": "...", "password": "...", "full_name": "Henry Tran", ... }
// AFTER
{ "email": "...", "password": "...", "first_name": "Henry", "last_name": "Tran", ... }
```

**Responses** — `GET /api/v1/employees/{id}`, `GET /api/v1/employees/me`, and the login/refresh
`user.employee` summary now return `first_name` + `last_name` (NO `full_name`):

```jsonc
// EmployeeRead / login user.employee
{ "id": "...", "first_name": "Henry", "last_name": "Tran", ... }
```

**Compose the display name** as `` `${first_name} ${last_name}`.trim() ``.

**Still composed for you (no change):** the embedded **briefs** keep a read-only
`full_name` you can show directly — `manager`, direct-reports rows,
manager-candidates, attendance rows, announcement author/recipient, invite inviter.
Example (Work Profile line manager):

```jsonc
"manager": { "id": "...", "full_name": "Sarah Le", "position": "CTO", "department": "Engineering", "is_active": true }
```

> Mononyms: a one-word name persists as `first_name="X", last_name=""` — render the trimmed composition (`"X"`), don't show a trailing space.

---

## 2. User-list filters — now working + multi-select

`GET /api/v1/employees` previously returned **400** for any `department_id`/`position_id`/
`role_id`/`manager_id` value (a binding bug). Fixed, and they're now **multi-select**.

- Pass each filter as a **repeated query param** for OR-within-filter; different filters AND together (per DR-001-005-01):
  ```
  GET /api/v1/employees?department_id=<uuid>&department_id=<uuid>&role_id=<uuid>&search=rogers&page=1&page_size=10
  ```
- **Status** chip → single `is_active` bool: send `is_active=true` (Active) or `is_active=false` (Inactive); omit when both/neither selected (= all).
- **Search** matches `first_name`, `last_name`, `phone`, `personal_email`, and login `email` (case-insensitive, partial).
- Default sort: **first name, then last name** (A→Z).
- Invalid UUID in any filter → `400 invalid <field>`.

Response envelope (unchanged): `{ items: [...], total, page, page_size, total_pages }`.

> Note (BA divergence, intentional): the Python source only multi-selected role; per DR-001-005-01 the Go API multi-selects **department, position, and role**. Department/Position name columns are now populated on each row (see §5).

---

## 3. Skills assigned inline on create/update

`POST /api/v1/employees` and `PATCH /api/v1/employees/{id}` now accept **`skill_ids`** —
no separate call needed for the Create/Edit User forms.

- **Create:** `"skill_ids": ["<skill-uuid>", ...]` (optional; omitted/empty = none).
- **Update:** pointer-to-slice replace semantics —
  - omit `skill_ids` → leave the set unchanged,
  - `"skill_ids": []` → clear all,
  - `"skill_ids": [ids]` → replace the whole set.
- An invalid `skill_id` → `400` and **no user is created** (validated before the write).
- **Read** returns the resolved set: `"skills": [ { "id": "...", "name": "React JS" }, ... ]` (display comma-separated).

The standalone `GET` / `PUT /api/v1/employees/{id}/skills` endpoints still exist (unchanged) if you prefer separate calls.

Skill catalog (EP-008 US-003) is unchanged: `GET/POST/PATCH/DELETE /api/v1/skills` with `{ id, name, description, icon_url }`.

---

## 4. `experience_year` is a career-start YEAR

Semantics changed from a **count of years** to a **4-digit career-start year**
(BA "Experience from"). Existing data was migrated (count → `currentYear − count`).

- Send/display a year, e.g. `"experience_year": 2018`.
- Validation: must be `> 1900` and `≤ current year` (else `400`). Optional/nullable.
- Display as a 4-digit year; "Experience from" in the Work Profile.

---

## 5. Department & Position resolved on the employee read

`GET /api/v1/employees/{id}` and `/employees/me` now return resolved objects
(previously always `null`). Use these for the list **Department/Position columns**
and the user-details **Work Profile**:

```jsonc
"department": { "id": "...", "name": "Engineering" },   // null when unassigned
"position":   { "id": "...", "name": "Senior Engineer" } // null when unassigned
```

---

## 6. Direct reports — no change

`GET /api/v1/employees/{id}/direct-reports` is unchanged — a standalone endpoint
(NOT embedded in the user read, matching Python). Returns all reports (active +
inactive), alphabetical by name:

```jsonc
[ { "id": "...", "full_name": "...", "avatar_url": "...|null", "position": "...|null", "department": "...|null", "is_active": true } ]
```

---

## Payload gotcha — empty optional dates → `null` (not `""`)

For an empty optional **date** field, send `null` or **omit the key** — never `""`. Go rejects an empty-string timestamp at JSON binding → **`400 bad_request`** (`parsing time "" … cannot parse "" as "2006"`) before any logic runs (no-op). Applies to `dob`, `id_issue_date`, `contract_sign_date`, `contract_end_date`, `join_date`. Empty **strings** for *text* fields (`cv_url`, addresses, `id_number`, `bank_*`) are fine. **FE rule:** before submit, coerce `""` → `null` for date (and number) fields. (Pre-existing contract behavior; surfaced during round-2 form integration.)

## Not in this round (still as before / deferred)

- Avatar / CV / ID-card **image upload** endpoints — still accept URLs in the payload (`avatar_url`, `cv_url`, `id_front_image`, `id_back_image`); dedicated multipart upload endpoints remain deferred (avatar upload via `PATCH /employees/{id}/avatar` and `/employees/me/avatar` is unchanged).
- Salary/banking field-perm gating + bank-account masking — unchanged (`users:{salary,banking}_{view,manage}`; account masked to `•••• 1234` on reads).
- Swagger (`docs/swagger/`) regenerated — reflects all of the above.

## Reference

Full endpoint set is in the regenerated Swagger UI (`/swagger/index.html`). Source DTOs: `internal/dto/employee.go`, `internal/dto/auth.go` (login summary). Backend commits on `feat/employees-parity-2`: `5ae3b6d` (filters), `ddf0d94` (experience), `56f53ef` (skills), `b5bd158` (dept/pos), `c52d66a` (name split), `68ddb37` (experience migration), `5c986d8`/`05cf0de` (swagger).
