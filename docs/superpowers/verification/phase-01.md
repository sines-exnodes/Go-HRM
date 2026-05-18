# Phase 1 Verification Log

Date: 2026-05-18
Phase: 1 — Auth & RBAC
Agent: Claude Code (worker)
Spec: docs/superpowers/specs/2026-05-15-go-migration-design.md
Plan: docs/superpowers/plans/2026-05-15-phase-01-auth-rbac.md

All outputs below were captured live against the local Postgres
(`ennam-ecom-postgres`, db `exnodes_hrm`) and a running `make run`
server on :8080. JWT values are replaced with `<REDACTED>`.

## 1. Migration state

Command:
    make migrate-version

Output:
    /Users/sines/go/bin/migrate -path migrations -database "postgres://ennam:...@localhost:5432/exnodes_hrm?sslmode=disable" version
    3

Command:
    make migrate-up

Output:
    /Users/sines/go/bin/migrate -path migrations -database "postgres://ennam:...@localhost:5432/exnodes_hrm?sslmode=disable" up
    no change

Result: DB already at version 3 (000001 init, 000002 roles/users,
000003 employees/dependents). `migrate-up` is a no-op.

## 2. Server boot + seed (`make run`)

Relevant boot log lines (verbatim):

    2026/05/18 09:44:33 seed: created role "Super Admin"
    2026/05/18 09:44:33 seed: created role "Admin"
    2026/05/18 09:44:33 seed: created role "HR Manager"
    2026/05/18 09:44:33 seed: created role "Manager"
    2026/05/18 09:44:33 seed: created role "Employee"
    2026/05/18 09:44:33 seed: created super admin user "admin@exnodes.vn"
    2026/05/18 09:44:33 seed: created super admin employee profile for "admin@exnodes.vn"
    [GIN-debug] POST   /api/v1/auth/login        --> ...AuthHandler).Login-fm (5 handlers)
    [GIN-debug] POST   /api/v1/auth/refresh      --> ...AuthHandler).Refresh-fm (5 handlers)
    [GIN-debug] POST   /api/v1/auth/logout       --> ...AuthHandler).Logout-fm (6 handlers)
    [GIN-debug] GET    /api/v1/roles/permissions --> ...RoleHandler).ListPermissions-fm (6 handlers)
    2026/05/18 09:44:33 exnodes-hrm-api listening on :8080 (env=development, swagger=true)

The seed ran AFTER the migration check and BEFORE serving. It is
idempotent — the boot log shows `record not found` → `INSERT` only
because this was the first seed against this DB; the lookup-before-write
path (FindByName / GetByEmail / GetByUserID) makes a re-run a no-op.

## 3. `GET /health`

Command:
    curl -s http://localhost:8080/health

Response:
    {"success":true,"data":{"status":"ok","service":"exnodes-hrm-api"}}

HTTP status: 200

## 4. Login as super admin (step 17.3)

Command:
    curl -s -X POST http://localhost:8080/api/v1/auth/login \
         -H 'Content-Type: application/json' \
         -d '{"email":"admin@exnodes.vn","password":"<REDACTED>"}'

HTTP status: 200

Response body (tokens redacted):

    {
      "success": true,
      "message": "Login successful",
      "data": {
        "access_token": "<REDACTED>",
        "refresh_token": "<REDACTED>",
        "token_type": "Bearer",
        "user": {
          "id": "3c8bc5ea-f7d5-4ffa-b02f-cfcdacd74e1e",
          "email": "admin@exnodes.vn",
          "is_active": true,
          "employee": {
            "id": "ab07e787-3ffe-4758-83f2-729c3daa59f7",
            "full_name": "Super Admin"
          },
          "roles": [
            {
              "id": "53db4bb1-93c7-4579-9160-03f398850ea1",
              "name": "Super Admin",
              "is_system": true,
              "permissions": ["*"]
            }
          ]
        }
      }
    }

Confirmed: `data.user.employee.full_name == "Super Admin"` (matches
`SUPER_ADMIN_NAME`). Roles carry the `["*"]` wildcard.

## 5. Protected endpoint WITH token (step 17.4)

Command:
    curl -s -H "Authorization: Bearer <REDACTED>" \
         http://localhost:8080/api/v1/roles/permissions

HTTP status: 200

`data` is an array of 11 PermissionGroup entries. Resource keys:

    ['auth', 'users', 'roles', 'departments', 'positions', 'skills',
     'leave_requests', 'leave_quota', 'attendance',
     'organization_settings', 'announcements']

Count = 11, matches `len(permissions.PermissionGroups)`.

First ~400 chars of body:

    {"success":true,"data":[{"resource":"auth","label":"Authentication",
    "permissions":[{"key":"auth:login","label":"Login","description":
    "Sign in to the system"}]},{"resource":"users","label":"Users",
    "permissions":[{"key":"users:read","label":"View Users",...

## 6. Protected endpoint with NO token (step 17.5)

Command:
    curl -s -o /dev/null -w '%{http_code}\n' \
         http://localhost:8080/api/v1/roles/permissions

HTTP status: 401

## 7. Protected endpoint with tampered token (step 17.6)

Command:
    curl -s -o /dev/null -w '%{http_code}\n' \
         -H "Authorization: Bearer <REDACTED>xxxx" \
         http://localhost:8080/api/v1/roles/permissions

HTTP status: 401

## 8. Wrong password (step 17.7)

Command:
    curl -s -X POST http://localhost:8080/api/v1/auth/login \
         -H 'Content-Type: application/json' \
         -d '{"email":"admin@exnodes.vn","password":"definitely-wrong"}' \
         -o /dev/null -w '%{http_code}\n'

HTTP status: 401

## 9. Non-existent email (step 17.8)

Command:
    curl -s -X POST http://localhost:8080/api/v1/auth/login \
         -H 'Content-Type: application/json' \
         -d '{"email":"ghost@nowhere.com","password":"anything"}' \
         -o /dev/null -w '%{http_code}\n'

HTTP status: 401

## 10. Refresh flow (step 17.9)

Command:
    curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
         -H 'Content-Type: application/json' \
         -d '{"refresh_token":"<REDACTED>"}'

HTTP status: 200
Result: `success: true`, a brand-new access+refresh pair returned.
The new access token differs from the original access token.

Negative case — using an access token as a refresh token:

    curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
         -H 'Content-Type: application/json' \
         -d '{"refresh_token":"<REDACTED access token>"}' \
         -o /dev/null -w '%{http_code}\n'

HTTP status: 400 (token-type mismatch rejected).

## 11. Logout (step 17.10)

Command:
    curl -s -X POST http://localhost:8080/api/v1/auth/logout \
         -H "Authorization: Bearer <REDACTED>" \
         -o /dev/null -w '%{http_code}\n'

HTTP status: 200

## 12. Swagger paths (step 17.11)

Command:
    curl -s http://localhost:8080/swagger/doc.json | jq '.paths | keys'

Paths emitted:

    ['/api/v1/auth/login',
     '/api/v1/auth/logout',
     '/api/v1/auth/refresh',
     '/api/v1/roles/permissions',
     '/health']

All 4 new Phase 1 endpoints are listed, plus the Phase 0 /health.

## 13. DB spot-check

Command:
    psql -c "SELECT email, is_active FROM users;"
Output:
     admin@exnodes.vn | t

Command:
    psql -c "SELECT full_name, user_id FROM employees;"
Output:
     Super Admin | 3c8bc5ea-f7d5-4ffa-b02f-cfcdacd74e1e

Command:
    psql -c "SELECT count(*) roles, (SELECT count(*) FROM users) users,
             (SELECT count(*) FROM employees) employees FROM roles;"
Output:
     5 | 1 | 1

Tables present (`\dt`): roles, users, user_roles, employees,
dependents, schema_migrations.

The employee `user_id` (3c8bc5ea-...) matches the user id returned in
the login response, confirming the seeded user↔employee link.

## 14. Test pass

Command:
    go test ./...

Result: all green.
    ok  github.com/exnodes/hrm-api/internal/permissions
    ok  github.com/exnodes/hrm-api/internal/services   0.592s
    ok  github.com/exnodes/hrm-api/pkg/utils
Remaining packages report `[no test files]`. `go build ./...` and
`go vet ./...` both exit 0.

## 15. Limitations

- **Logout is stateless.** `POST /api/v1/auth/logout` returns 200 but
  does not server-side invalidate the access token; the token remains
  valid until its natural expiry. This is a documented Phase 1
  limitation — a token denylist / session store is deferred to a later
  phase. The middleware does already honour the per-user
  `email_changed_at` / `password_reset_at` invalidation timestamps.

## 16. Sign-off

All Phase 1 acceptance criteria met:

- [x] Migrations at version 3; `make migrate-up` is a clean no-op
- [x] Server boots with no panic; seed runs on boot (5 roles, 1 user, 1 employee)
- [x] Login returns 200 with embedded `user.employee.full_name == "Super Admin"`
- [x] Protected endpoint returns 200 with 11 permission groups when authed
- [x] 401 for missing token, tampered token, wrong password, unknown email
- [x] Refresh returns a new token pair; access-as-refresh rejected (400)
- [x] Logout returns 200 (stateless — noted as a limitation)
- [x] Swagger lists all 4 new endpoints
- [x] DB spot-check confirms seeded super admin user + employee link
- [x] `go build`, `go vet`, `go test` all pass
