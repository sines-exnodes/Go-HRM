# Employee CSV Import — FE API Guide

**Date:** 2026-07-23  
**Backend:** Go HRM API, base path `/api/v1`  
**Auth:** `Authorization: Bearer <access_token>` on both endpoints  
**Permission:** `employees:create` (same as single-employee create)  
**Merged:** PR #40 → `main`  
**Swagger:** `/swagger/index.html` (tags: `employees`)  
**Source of truth:** this doc + live handlers. Drop into the web repo as `api_info_go/employee-import.md` if desired.

---

## What this feature does

Bulk-create employees from a `.csv` upload. Each data row runs the same create path as `POST /api/v1/employees` (user + employee + default role). Import is **create-only** — duplicate emails fail that row; existing employees are not updated.

Two endpoints:

| Method | Path | Permission | Content-Type | Returns |
|--------|------|------------|--------------|---------|
| `GET` | `/api/v1/employees/import/template` | `employees:create` | — | `text/csv` file download |
| `POST` | `/api/v1/employees/import` | `employees:create` | `multipart/form-data` | JSON import report |

---

## 1. Download template

### Request

```http
GET /api/v1/employees/import/template
Authorization: Bearer <access_token>
```

No query params. No body.

### Response

| Status | Meaning |
|--------|---------|
| `200` | CSV body |
| `401` | missing/invalid JWT |
| `403` | missing `employees:create` |

Headers of interest:

| Header | Value |
|--------|--------|
| `Content-Type` | `text/csv; charset=utf-8` |
| `Content-Disposition` | `attachment; filename="employee-import-template.csv"` |

Body (exact):

```csv
email,first_name,last_name,phone,department,position,role,manager_email,join_date,is_active
jane.doe@example.com,Jane,Doe,+84901234567,Engineering,Software Engineer,Employee,,2026-07-01,true
```

### FE notes

- Trigger via button “Download template”. Use `fetch`/`axios` with `responseType: 'blob'` (or equivalent), then save as `employee-import-template.csv`.
- Do **not** expect a JSON envelope — this endpoint returns raw CSV, not `{ success, data }`.
- The template is a **minimal starter** (subset of allowed columns). Full column list is below; extra allowed columns can be added by the user in Excel/Sheets.

---

## 2. Import CSV

### Request

```http
POST /api/v1/employees/import
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

| Form field | Type | Required | Default | Description |
|------------|------|----------|---------|-------------|
| `file` | file | **yes** | — | `.csv` body, max **2 MiB** |
| `send_invite` | string | no | `false` | `"true"` or `"1"` (case-insensitive) → send set-password email per **successfully created** row (best-effort; import still succeeds if SMTP fails) |

Example (`curl`):

```bash
curl -X POST "$BASE/api/v1/employees/import" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@employees.csv;type=text/csv" \
  -F "send_invite=false"
```

### Response envelope (always JSON)

**HTTP 200** when the file **parses** — even if every row fails. Treat `data.failed` / `data.results[].ok` as the real outcome.

```json
{
  "success": true,
  "message": "Import finished: 2 created, 1 failed",
  "data": {
    "total_rows": 3,
    "created": 2,
    "failed": 1,
    "results": [
      {
        "row": 2,
        "ok": true,
        "email": "alice@example.com",
        "employee_id": "0179c4f6-33e0-4957-99c6-2f4b35dd3090",
        "user_id": "7342d63d-a725-4b50-8504-3e4b0f32adc5"
      },
      {
        "row": 3,
        "ok": false,
        "email": "not-an-email",
        "error": "invalid email"
      },
      {
        "row": 4,
        "ok": true,
        "email": "bob@example.com",
        "employee_id": "8979fd85-8cd7-406d-8229-9d8d93af8e17",
        "user_id": "b3f16c37-fcb0-40db-9494-0816978d5050"
      }
    ]
  }
}
```

### `EmployeeImportResult` (`data`)

| Field | Type | Description |
|-------|------|-------------|
| `total_rows` | number | Data rows processed (excludes header) |
| `created` | number | Rows that created a user+employee |
| `failed` | number | Rows that did not create |
| `results` | `EmployeeImportRowResult[]` | One entry per data row, same order as file |

### `EmployeeImportRowResult`

| Field | Type | When present | Description |
|-------|------|--------------|-------------|
| `row` | number | always | **1-based file line number** (header = line 1 → first data row = **2**) |
| `ok` | boolean | always | `true` = created |
| `email` | string | always | Email from the row (may be empty if missing) |
| `employee_id` | uuid string | success only | New employee id |
| `user_id` | uuid string | success only | New user id |
| `error` | string | failure only | Client-safe message (never raw SQL/driver text) |

### Whole-file errors (HTTP 400)

These abort the entire import — **no** `results` array. Standard error envelope:

```json
{
  "success": false,
  "message": "…",
  "code": "bad_request"
}
```

| Typical `message` | Cause |
|-------------------|--------|
| `file is required` | missing multipart field `file` |
| `file exceeds 2MB limit` | over size |
| `CSV file is empty` | empty body |
| `CSV must include a header row and at least one data row` | header only / empty |
| `CSV exceeds maximum of 500 data rows` | too many rows |
| `missing required column: email` (or `first_name` / `last_name`) | header missing required col |
| `unknown column: …` | header not in allow-list (typos fail loud) |
| `duplicate column: …` | repeated header |
| `invalid CSV: …` | parse failure |

Other statuses: `401` unauthenticated, `403` missing `employees:create`.

### Partial success (important for UI)

- One bad row **does not** roll back good rows.
- Prefer UX:

  1. Show toast/summary from `message` or `created` / `failed`.
  2. Table of `results` filtered by `ok === false` with columns: Line (`row`), Email, Error.
  3. On full success (`failed === 0`), toast success and optionally refresh user/employee list.
  4. On full failure (`created === 0 && failed > 0`) still HTTP 200 — do **not** treat as transport error.

---

## 3. CSV contract

### Rules

| Rule | Detail |
|------|--------|
| Encoding | UTF-8; leading BOM from Excel is accepted (stripped server-side) |
| Delimiter | **`,` or `;`** auto-detected from the header line (Excel EU/VN often uses `;`) |
| Header | **Required**, first row |
| Header names | Case-insensitive; trimmed; must be lower `snake_case` after normalize |
| Required columns | `email`, `first_name`, `last_name` |
| Unknown columns | Whole-file **400** (not skipped) |
| Max size | 2 MiB |
| Max data rows | 500 |
| Dates | `YYYY-MM-DD` only (`dob`, `join_date`) |
| Booleans | `true` / `false` / `1` / `0` (case-insensitive) for `is_active` |
| Empty optional cell | field left unset (defaults apply) |
| Password | **Not** set via CSV — user is passwordless until invite / reset |
| Update / upsert | **Not supported** — duplicate email → row error |

### Columns

| Column | Required | Maps to | Notes for FE / admins |
|--------|----------|---------|------------------------|
| `email` | **yes** | login email | unique; format validated |
| `first_name` | **yes** | | |
| `last_name` | **yes** | | |
| `phone` | | | |
| `personal_email` | | | must be valid email if set |
| `gender` | | | `male` \| `female` \| `other` |
| `dob` | | | `YYYY-MM-DD` |
| `nationality` | | | free text |
| `id_number` | | | |
| `permanent_address` | | | |
| `current_address` | | | |
| `education` | | | `high_school` \| `college` \| `bachelor` \| `master` \| `doctorate` |
| `marital_status` | | | `single` \| `married` \| `other` |
| `experience_year` | | | career-start year (integer) |
| `department` | | department **name** | case-insensitive; must already exist |
| `position` | | position **name** | case-insensitive; must already exist |
| `manager_email` | | line manager | must be an **existing active** user that already has an employee profile. Managers in the **same file** are not auto-ordered — import managers first |
| `join_date` | | | `YYYY-MM-DD` |
| `role` | | role **name** | single role; empty → system default `Employee` |
| `is_active` | | | default `true` if empty |
| `contract_type` | | | free text; Create default is `official` if empty |
| `basic_salary` | | | needs caller `users:salary_manage`; must be ≥ 0 |
| `insurance_salary` | | | same |
| `bank_account` | | | needs caller `users:banking_manage` |
| `bank_name` | | | same |
| `bank_holder_name` | | | same |
| `payment_method` | | | same |
| `social_insurance_number` | | | |
| `tax_identification_number` | | | |

### Not supported in v1 (use single create / later edit)

Avatar, CV, ID images, emergency contacts, multi-role, multi-skill, contracts, update-on-duplicate.

### Common row-level `error` strings

| `error` (examples) | Meaning |
|--------------------|---------|
| `email is required` | empty required cell |
| `invalid email` | bad format |
| `A user with this email already exists` | duplicate |
| `unknown department: …` | name not found |
| `unknown position: …` | name not found |
| `unknown role: …` | name not found |
| manager-related messages | email not found / inactive / no employee profile |
| `invalid gender: …` / `invalid education: …` / … | enum mismatch |
| `invalid dob: use YYYY-MM-DD` | date format |
| salary/banking forbidden text | column set without manage perm |
| `internal error processing row` | server-side failure (safe message) |

---

## 4. Suggested FE flow

```
User list (has employees:create)
  ├─ [Download template] → GET …/import/template → save blob
  └─ [Import employees]
        → file picker (.csv)
        → optional toggle “Send set-password email” → send_invite
        → POST multipart
        → if HTTP 4xx: show envelope message
        → if HTTP 200:
             summary: created / failed
             if failed > 0: error table from results where !ok
             if created > 0: refresh list
```

### TypeScript shapes (illustrative)

```ts
type EmployeeImportRowResult = {
  row: number;
  ok: boolean;
  email: string;
  employee_id?: string;
  user_id?: string;
  error?: string;
};

type EmployeeImportResult = {
  total_rows: number;
  created: number;
  failed: number;
  results: EmployeeImportRowResult[];
};

type ApiOk<T> = {
  success: true;
  message: string;
  data: T;
};
```

### Client snippets

**Template download**

```ts
const res = await fetch(`${base}/api/v1/employees/import/template`, {
  headers: { Authorization: `Bearer ${token}` },
});
if (!res.ok) throw new Error(await res.text());
const blob = await res.blob();
// saveAs(blob, 'employee-import-template.csv')
```

**Import**

```ts
const form = new FormData();
form.append('file', file); // File from <input type="file" accept=".csv,text/csv">
if (sendInvite) form.append('send_invite', 'true');

const res = await fetch(`${base}/api/v1/employees/import`, {
  method: 'POST',
  headers: { Authorization: `Bearer ${token}` },
  // do NOT set Content-Type manually — browser sets multipart boundary
  body: form,
});

const json = await res.json();
if (!res.ok || !json.success) {
  // whole-file failure
  showError(json.message ?? 'Import failed');
  return;
}

const { created, failed, results } = json.data as EmployeeImportResult;
// render summary + failed rows
```

---

## 5. Permissions & gating

| Need | Permission |
|------|------------|
| See import buttons / call either endpoint | `employees:create` |
| CSV includes `basic_salary` / `insurance_salary` | caller also needs `users:salary_manage` (else those rows fail) |
| CSV includes bank columns | caller also needs `users:banking_manage` |

Hide import UI when the session user lacks `employees:create` (same as “Add user”).

---

## 6. Out of scope / follow-ups

- No dry-run mode  
- No `.xlsx` (CSV only)  
- No async job / progress websocket for large files (cap is 500 rows)  
- Manager rows must pre-exist (no same-file dependency sort)  

---

## Related backend docs

- Plan: `docs/superpowers/plans/2026-07-23-employee-csv-import.md`  
- Verification: `docs/superpowers/verification/employee-csv-import.md`  
- Single create (field semantics): existing employee create API / DR-001-005-02  
