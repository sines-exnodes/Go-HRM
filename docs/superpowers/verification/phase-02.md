# Phase 2 — Employees + Dependents Module — Verification Log

Date: 2026-05-18
Operator: Claude Code (agent)
Server: `make run` (background), `:8080`, env=development, migration v4 applied.

> All output below is REAL captured curl/psql output, not placeholders.

## Bug found + fixed during verification

**Bug:** A freshly created employee (`POST /api/v1/employees` without `role_ids`)
got **zero roles**, so login was rejected with HTTP 403
`"You do not have permission to access this system."` (auth gate requires
`*` or `auth:login`). This broke the entire self-service half of Phase 2 and
contradicted the Definition of Done ("Self-service can: GET /users/me ...").

**Fix:** `EmployeeService.Create` now falls back to the system default
`"Employee"` role (seeded with `auth:login`, carries
`leave_requests:read/create`, `attendance:read`) when the admin supplies no
`role_ids`. Assignment stays inside the same creation transaction. Explicit
`role_ids` are still honoured unchanged.
File: `internal/services/employee_service.go` (new `defaultEmployeeRoleName`
const + role fallback block in `Create`).

Proof after fix: new employee login → HTTP 200, `GET /users/me` shows
`"roles": ["Employee"]`.

## Happy path

### 0. Swagger sanity — `GET /swagger/doc.json`
All Phase 2 paths present:
```
/api/v1/employees
/api/v1/employees/me
/api/v1/employees/me/avatar
/api/v1/employees/{id}
/api/v1/employees/{id}/avatar
/api/v1/employees/{id}/dependents
/api/v1/employees/{id}/dependents/{dependentID}
/api/v1/employees/{id}/leave-quota
/api/v1/users/me
/api/v1/users/me/change-email
/api/v1/users/me/change-password
/api/v1/users/me/device-tokens
/api/v1/users/me/device-tokens/{token}
/api/v1/users/me/notification-settings
/api/v1/users/{id}
/api/v1/users/{id}/change-password
/api/v1/users/{id}/roles
```
Health: `{"success":true,"data":{"status":"ok","service":"exnodes-hrm-api"}}`

### 1. Admin login — `POST /api/v1/auth/login` → 200, access_token issued.
`{"email":"admin@exnodes.vn","password":"Admin@12345"}` → token prefix
`eyJhbGciOiJIUzI1NiIsInR5cCI6Ik...`

### 2. Create employee — `POST /api/v1/employees` → **HTTP 201**
Request: `{"email":"e2e_1779080151@exnodes.vn","password":"E2eUser1234","full_name":"E2E User","basic_salary":15000,"phone":"0900000000"}`
```json
{"success":true,"data":{"id":"333544ae-900d-4be6-856b-32056db4f1d9",
"user_id":"f151fc48-7978-4b47-96a2-b9cf2eeff7ea","email":"e2e_1779080151@exnodes.vn",
"full_name":"E2E User","basic_salary":15000}}
```
EMP_ID=`333544ae-900d-4be6-856b-32056db4f1d9` USER_ID=`f151fc48-7978-4b47-96a2-b9cf2eeff7ea`

### 3. List employees — `GET /api/v1/employees?search=...` → 200
`{"success":true,"total":1,"names":["E2E User"]}`

### 4. Get employee — `GET /api/v1/employees/:id` → **HTTP 200**
`{"success":true,"full_name":"E2E User","phone":"0900000000","basic_salary":15000}`

### 5. Admin patch — `PATCH /api/v1/employees/:id` → 200
Request `{"full_name":"E2E Updated","basic_salary":22000}` →
`{"success":true,"full_name":"E2E Updated","basic_salary":22000}`

### 6. New-employee login — `POST /api/v1/auth/login` → **HTTP 200** (post-fix)
`{"success":true,"has_token":true}`

### 7. GET /users/me — embedded employee summary + auto role
`{"success":true,"email":"e2e_1779080151@exnodes.vn",
"employee_full_name":"E2E Updated","roles":["Employee"]}`

### 8. GET /employees/me → 200
`{"success":true,"full_name":"E2E Updated","phone":"0900000000"}`

### 9. Self-update whitelist — `PATCH /api/v1/employees/me` → **HTTP 200**
Request DELIBERATELY includes forbidden fields:
`{"phone":"0123-456-789","personal_email":"e2e.personal@example.com","basic_salary":999,"department_id":"11111111-1111-1111-1111-111111111111"}`
Response: `{"success":true,"phone":"0123-456-789","personal_email":"e2e.personal@example.com"}`

**SQL PROOF (the single most important assertion) — forbidden fields ignored:**
```
SELECT basic_salary, department_id IS NULL, phone, personal_email
FROM employees WHERE id='333544ae-900d-4be6-856b-32056db4f1d9';

 basic_salary=22000.00 | department_id IS NULL=true | phone=0123-456-789 | personal_email=e2e.personal@example.com
```
`basic_salary` stayed at the admin-set `22000.00` (NOT 999); `department_id`
stayed NULL. The DTO-boundary whitelist holds.

### 10. Change own password — `POST /api/v1/users/me/change-password` → **HTTP 200**
`{"current_password":"E2eUser1234","new_password":"E2eUser5678"}` → `{"success":true}`
Re-login with NEW password → **HTTP 200** `{"success":true,"has_token":true}`
Login with OLD password → **HTTP 401** `{"success":false,"code":"unauthorized"}`

### 11. Dependents CRUD (self-owner) under `/employees/:id/dependents`
- POST → **HTTP 201** `{"id":"216e5764-...","full_name":"Child One","relationship":"child"}`
- GET list → `{"success":true,"total":1,"names":["Child One"]}`
- PATCH rename → `{"success":true,"full_name":"Child Renamed"}`
- DELETE → **HTTP 200** `{"success":true}`

### 12. DB spot-check
```
SELECT email FROM users;       -> e2e_1779080151@exnodes.vn
SELECT full_name,phone ...     -> E2E Updated | 0123-456-789
SELECT full_name,is_deleted    -> Child Renamed | t   (soft-deleted)
```

## Error cases

> Note: the JWT access TTL is 60 min. An expired admin token returns
> HTTP 401 `"Could not validate credentials"` (token validation, not the
> business rule). Steps E8–E11 were re-run with a fresh admin token; the
> outputs below are the authoritative results.

### E1. Create employee WITHOUT token — `POST /api/v1/employees` → **HTTP 401**
`{"success":false,"code":"unauthorized","message":"Could not validate credentials"}`

### E2. Non-admin LIST — `GET /api/v1/employees` (Employee token) → **HTTP 403**
`{"success":false,"code":"forbidden"}` (lacks `employees:read`)

### E3. Non-admin CREATE — `POST /api/v1/employees` (Employee token) → **HTTP 403**
`{"success":false,"code":"forbidden"}` (lacks `employees:create`)

### E4. Duplicate email — `POST /api/v1/employees` (admin, existing email) → **HTTP 409**
`{"success":false,"code":"conflict","message":"A user with this email already exists"}`

### E5. Stranger accesses another employee's dependents → **HTTP 403**
`GET /api/v1/employees/<SECOND_ID>/dependents` with the first employee's token:
`{"success":false,"code":"forbidden","message":"You may only manage your own dependents"}`

### E6. Not-found employee — `GET /api/v1/employees/000...000` (admin) → **HTTP 404**
`{"success":false,"code":"not_found"}`

### E7. Unauthenticated /employees/me — no token → **HTTP 401**
`{"success":false,"code":"unauthorized"}`

### E8. Wrong-password admin delete — `DELETE /api/v1/users/:id` `{"current_password":"WrongPass"}` → **HTTP 400**
`{"success":false,"code":"bad_request","message":"Incorrect password. Please try again."}`

### E9–E11. Soft-delete + cascade — `DELETE /api/v1/employees/:id` (admin) → **HTTP 200**

Pre-delete DB state: `emp_deleted=f | user_active=t`

`{"success":true}`

**SQL PROOF of soft-delete cascade:**
```
SELECT e.is_deleted, e.deleted_at IS NOT NULL, u.is_active
FROM employees e JOIN users u ON u.id=e.user_id WHERE e.id='333544ae-...';

 emp_deleted | emp_del_at_set | user_active
-------------+----------------+-------------
 t           | t              | f
```
`employees.is_deleted=t`, `deleted_at` set, **linked `users.is_active=f`**
(cascade deactivation).

- `GET /api/v1/employees/:id` after delete → **HTTP 404** `{"success":false,"code":"not_found"}`
- Deleted user re-login → **HTTP 401**
  `{"success":false,"code":"unauthorized","message":"Your account has been deactivated. Contact your administrator."}`

## Result

All happy-path and error-path assertions pass against a live server +
real Postgres. One bug found and fixed (default-role-on-create, see top).
The self-update whitelist and soft-delete cascade are both proven by direct SQL.
