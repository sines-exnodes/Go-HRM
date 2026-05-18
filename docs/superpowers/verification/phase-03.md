# Phase 3 — End-to-end verification (Departments + Positions + real FK)

Date: 2026-05-18
Server: `make run` against local Postgres (`localhost:5432`, db `exnodes_hrm`,
migration **v5**). All output below is **real captured** curl / psql output;
only the bearer token value is redacted (length shown).

A genuine bug was found in step 6 and fixed before the final run — see
**Bug found & fixed** at the end. After the fix the full flow passes and the
service suite is **55/55**.

---

## 1. Boot — migration v5, schema not dirty, default org tree seeded

First boot of a fresh DB logged the seed (idempotent):

```
seed: created default org tree (4 departments, 3 positions)
[GIN-debug] GET    /api/v1/departments       --> ...DepartmentHandler.List-fm
[GIN-debug] POST   /api/v1/departments       --> ...DepartmentHandler.Create-fm
[GIN-debug] GET    /api/v1/departments/:id   --> ...DepartmentHandler.Get-fm
[GIN-debug] PATCH  /api/v1/departments/:id   --> ...DepartmentHandler.Update-fm
[GIN-debug] DELETE /api/v1/departments/:id   --> ...DepartmentHandler.Delete-fm
[GIN-debug] GET    /api/v1/positions         --> ...PositionHandler.List-fm
[GIN-debug] POST   /api/v1/positions         --> ...PositionHandler.Create-fm
[GIN-debug] GET    /api/v1/positions/:id     --> ...PositionHandler.Get-fm
[GIN-debug] PATCH  /api/v1/positions/:id     --> ...PositionHandler.Update-fm
[GIN-debug] DELETE /api/v1/positions/:id     --> ...PositionHandler.Delete-fm
exnodes-hrm-api listening on :8080 (env=development, swagger=true)
```

Subsequent boots: seed line absent (org tree already exists — idempotent,
correct). Migration version check:

```
psql> SELECT version||' dirty='||dirty FROM schema_migrations;
 5 dirty=false
```

FK constraints exist (deferred from 000003, added in 000005), both
`ON DELETE SET NULL`:

```
psql> SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conname LIKE 'fk_employees_%';
         conname         |                       pg_get_constraintdef
-------------------------+---------------------------------------------------------------------------
 fk_employees_department | FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL
 fk_employees_position   | FOREIGN KEY (position_id)   REFERENCES positions(id)   ON DELETE SET NULL
```

(`pg_constraint.confdeltype = 'n'` is the catalog code for SET NULL — not
"no action", which is `'a'`. Matches migration `000005`.)

## 2. Login as super admin

```
POST /api/v1/auth/login {"email":"admin@exnodes.vn","password":"<redacted>"}
-> success=true, user=admin@exnodes.vn, access_token = 209 chars (value redacted)
```

## 3. Create departments

```
POST /api/v1/departments {"name":"Verify Root Dept","description":"e2e root"}
-> HTTP 201
   {"id":"b0b9e847-6208-4e9d-b481-ced5f05d93f3","name":"Verify Root Dept","parent_id":null,
    "created_at":"2026-05-18T16:54:28.128846+07:00"}

POST /api/v1/departments {"name":"Verify Child Dept","parent_id":"b0b9e847-...d93f3"}
-> HTTP 201
   {"id":"4a43b48a-4fb3-4196-8705-1e15590a4014","name":"Verify Child Dept",
    "parent_id":"b0b9e847-6208-4e9d-b481-ced5f05d93f3"}   # parent_id == ROOT_ID
```

## 4. Create position in child dept

```
POST /api/v1/positions {"name":"Verify Engineer","department_id":"4a43b48a-...4014"}
-> HTTP 201
   {"id":"a70e4c75-0386-47ad-b41a-e599fd21adb9","name":"Verify Engineer",
    "department_id":"4a43b48a-4fb3-4196-8705-1e15590a4014"}   # == CHILD_ID

POST /api/v1/positions {"name":"Bad Pos","department_id":"11111111-2222-3333-4444-555555555555"}
-> HTTP 400  {"success":false,"message":"Department not found"}
```

## 5. List departments — `?parent_id=root&search=Verify`

```
GET /api/v1/departments?parent_id=root&search=Verify  -> HTTP 200
{"total":1,"items":[{"name":"Verify Root Dept","parent_id":null}]}
```

Includes the root, excludes the child (child has a non-null parent). ✔

## 6. List positions — `?department_id=<CHILD_ID>`  (was the bug; now fixed)

```
GET /api/v1/positions?department_id=4a43b48a-4fb3-4196-8705-1e15590a4014 -> HTTP 200
{"total":1,"items":[{"name":"Verify Engineer",
                     "department_id":"4a43b48a-4fb3-4196-8705-1e15590a4014"}]}
```

## 7. PATCH child department

```
PATCH /api/v1/departments/4a43b48a-...4014 {"description":"updated via e2e"}
-> HTTP 200  {"id":"4a43b48a-...4014","name":"Verify Child Dept","description":"updated via e2e"}
```

## 8. FK exercise — assign an employee to the created dept + position

```
POST /api/v1/employees
  {"email":"fkprobe@example.com","password":"<redacted>","full_name":"FK Probe",
   "department_id":"4a43b48a-...4014","position_id":"a70e4c75-...adb9"}
-> HTTP 201  employee id = 98298e3b-a8bf-4bd4-b96b-215446e0e690
```

**SQL proof the FK assignment landed in the row** (the single most important
assertion). Note: the `EmployeeRead` response intentionally omits dept/pos
refs (see "Observation" below) — the DB row is the source of truth:

```
psql> SELECT id, department_id, position_id FROM employees WHERE id = '98298e3b-...e690';
                  id                  |            department_id             |             position_id
--------------------------------------+--------------------------------------+--------------------------------------
 98298e3b-a8bf-4bd4-b96b-215446e0e690 | 4a43b48a-4fb3-4196-8705-1e15590a4014 | a70e4c75-0386-47ad-b41a-e599fd21adb9

psql> SELECT (e.department_id = '4a43b48a-...4014') AS dept_fk_match,
             (e.position_id   = 'a70e4c75-...adb9') AS pos_fk_match
      FROM employees e WHERE e.id = '98298e3b-...e690';
 dept_fk_match | pos_fk_match
---------------+--------------
 t             | t
```

## 9. Delete the child dept while it has a position + employee → **409**

```
DELETE /api/v1/departments/4a43b48a-...4014  -> HTTP 409
{"success":false,
 "message":"Cannot delete — 1 position is assigned to this department. Delete or reassign them first.",
 "code":"conflict"}
```

SQL spot-check — dept STILL present, NOT soft-deleted:

```
psql> SELECT id, name, is_deleted, deleted_at FROM departments WHERE id = '4a43b48a-...4014';
                  id                  |       name        | is_deleted | deleted_at
--------------------------------------+-------------------+------------+------------
 4a43b48a-4fb3-4196-8705-1e15590a4014 | Verify Child Dept | f          |
```

Also — delete the **root** dept while it has a child → **409** (child guard):

```
DELETE /api/v1/departments/b0b9e847-...d93f3  -> HTTP 409
{"success":false,
 "message":"Cannot delete department — it has child departments. Move or delete them first.",
 "code":"conflict"}
```

## 10. Delete the position while the employee references it → **409**

```
DELETE /api/v1/positions/a70e4c75-...adb9  -> HTTP 409
{"success":false,
 "message":"Cannot delete — 1 employee is assigned to this position. Reassign all employees before deleting.",
 "code":"conflict"}

psql> SELECT id, name, is_deleted, deleted_at FROM positions WHERE id = 'a70e4c75-...adb9';
 a70e4c75-0386-47ad-b41a-e599fd21adb9 | Verify Engineer | f |    -- still present, not soft-deleted
```

## 11. Clear the employee, then delete position + dept → **200**, soft-deleted

```
PATCH /api/v1/employees/98298e3b-...e690 {"clear_department":true,"clear_position":true}
-> HTTP 200
psql> SELECT department_id IS NULL AS dept_cleared, position_id IS NULL AS pos_cleared
      FROM employees WHERE id='98298e3b-...e690';
 dept_cleared | pos_cleared
--------------+-------------
 t            | t

DELETE /api/v1/positions/a70e4c75-...adb9     -> HTTP 200 {"success":true,"message":"Position deleted"}
DELETE /api/v1/departments/4a43b48a-...4014   -> HTTP 200 {"success":true,"message":"Department deleted"}

psql> soft-delete columns set on BOTH:
     t      | is_deleted | has_deleted_at
------------+------------+----------------
 position   | t          | t
 department | t          | t
```

## 12. Error paths

```
POST /api/v1/departments {"name":"Orphan","parent_id":"99999999-...555"}
-> HTTP 400  {"success":false,"message":"Parent department not found"}

POST /api/v1/departments {"name":"Verify Root Dept"}      # duplicate root name
-> HTTP 409  {"success":false,"message":"Department name already exists","code":"conflict"}
```

## 13. Permission enforcement

```
POST /api/v1/departments  (NO Authorization header)
-> HTTP 401  {"success":false,"message":"Could not validate credentials","code":"unauthorized"}

# create a role-less employee, log in as it, retry:
POST /api/v1/employees {"email":"plainuser@example.com",...}  -> HTTP 201
POST /api/v1/auth/login (plainuser)                            -> token (209 chars)

POST /api/v1/departments  (Bearer <plain-user token>)
-> HTTP 403  {"success":false,"message":"Insufficient permissions","code":"forbidden"}
GET  /api/v1/departments  (Bearer <plain-user token>)
-> HTTP 403  {"success":false,"message":"Insufficient permissions","code":"forbidden"}
```

## 14. Cleanup

```
DELETE /api/v1/departments/b0b9e847-...d93f3 (root)  -> HTTP 200
DELETE /api/v1/employees/98298e3b-...e690 (fkprobe)  -> HTTP 200
DELETE /api/v1/employees/d55b646b-...8a03 (plain)    -> HTTP 200
psql> SELECT is_deleted, deleted_at IS NOT NULL FROM departments WHERE id='b0b9e847-...d93f3';
 t | t
```

Server stopped cleanly (0 listeners on :8080).

---

## Bug found & fixed (during step 6)

**Symptom:** `GET /api/v1/positions?department_id=<uuid>` returned **HTTP 400**:

```
{"success":false,"message":"[\"4a43b48a-...4014\"] is not valid value for uuid.UUID","code":"bad_request"}
```

**Root cause:** `dto.PositionListQuery.DepartmentID` was typed `*uuid.UUID`
with `form:"department_id"`. Gin's query binding cannot populate a
`*uuid.UUID` from a raw query string (it passes the `[]string` slice through),
so every filtered list 400'd. `DepartmentListQuery.ParentID` already avoided
this by being a `string` parsed in the service.

**Fix (mirrors the established department pattern):**
- `internal/dto/position.go` — `DepartmentID` changed to `string`
  (`form:"department_id"`).
- `internal/services/position_service.go` — `List` now `strings.TrimSpace`
  + `uuid.Parse` the value (empty = no filter; bad format = `400 Invalid
  department_id`), exactly like `DepartmentService.List` does for `parent_id`.
- `internal/services/position_service_test.go` — two `PositionListQuery`
  literals updated from `&dept.ID` to `dept.ID.String()`.

Post-fix: `go build ./...` exit 0; `go test ./internal/services/...` = **55
PASS / 0 FAIL**; step 6 returns HTTP 200 with the expected single item.

## Observation (not a Phase 3 plan task — out of scope)

`EmployeeService.toRead`/`toSummary` deliberately leave the nested
`department` / `position` ref objects nil ("intentionally nil until then" /
"refs filled in Phase 3"), so the employee create/get JSON omits them even
though the DB FK columns are correctly set (proven by SQL in step 8). The FK
guard logic operates on DB state and is fully correct. Populating the
employee read view with dept/position refs is not part of the Phase 3 plan
task list (Tasks 1–18) and is left for a follow-up; flagged here for
visibility.
