# Verification — Roles & Permissions API Parity

**Date:** 2026-06-05
**Branch:** `feat/roles-permissions-parity`
**Spec:** [`specs/2026-06-04-roles-permissions-parity-audit.md`](../specs/2026-06-04-roles-permissions-parity-audit.md)
**Plan:** [`plans/2026-06-04-roles-permissions-parity.md`](../plans/2026-06-04-roles-permissions-parity.md)
**Migrations added:** **000020** (role `level` column) + **000021** (partial-unique
`uq_roles_name_active ON roles(LOWER(name)) WHERE is_deleted=FALSE`).
Latest applied = **000021**; next free = **000022**.

Executed subagent-driven (fresh implementer per task + two-stage review). Each
task committed separately; commit list at the bottom.

## 1. Build / vet / format

| Check | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | clean |
| `gofmt` (changed Go files) | clean (pre-existing CRLF noise on unrelated files only) |
| `make swag` (swagger regen) | regenerated `docs/swagger/{docs.go,swagger.json,swagger.yaml}`; role paths present (`/roles`, `/roles/{id}`, `/roles/permissions`) |

## 2. Automated tests — full repo suite (0 fail, 0 skip)

`go test ./... -count=1` with `TEST_DATABASE_URL` set (real Postgres, all
migrations incl. 000020/000021 applied by `TestMain`):

```
ok  github.com/exnodes/hrm-api/internal/middleware    0.159s
ok  github.com/exnodes/hrm-api/internal/permissions   0.425s
ok  github.com/exnodes/hrm-api/internal/services      172.223s
ok  github.com/exnodes/hrm-api/internal/sse           0.806s
ok  github.com/exnodes/hrm-api/pkg/utils              1.265s
```

New role tests (in `internal/services/role_service_test.go`, 9 cases) +
`user_service_test.go` level-authority cases all pass. The pre-existing
`TestUserService_AssignRoles` was correctly updated for the new authority gate
(a role-less admin can no longer grant a level-100 role — matches Python).

Also fixed during setup: the test harness built a `file://E:\...` migration URL
unparseable on Windows; switched to an in-process `iofs` source so the suite runs
cross-platform (commit `6d58ad4`).

## 3. Live HTTP end-to-end smoke

Server booted on `PORT=8082` against the migrated test DB (`exnodes_hrm_test`,
truncated to a clean slate; boot `Seed()` created the 5 system roles with levels +
the super admin). Boot clean: `exnodes-hrm-api listening on :8082`, no Gin route
panic. Token = super-admin access token from `/auth/login`.

| # | Step | Request | Result |
|---|---|---|---|
| 1 | Login | `POST /auth/login` admin | 200, access_token issued |
| 2 | List (level-asc + permission_count) | `GET /roles` | 200; order **10 Employee, 50 Manager, 80 HR Manager, 90 Admin, 100 Super Admin**; `permission_count` populated (7/12/36/44/1) |
| 3 | Catalog (super admin) | `GET /roles/permissions` | 200 |
| 4 | Create | `POST /roles {QA Lead, level 40, [users:read, roles:read]}` | **201**, level 40, permission_count 2, is_system false |
| 5 | Duplicate name (case-insensitive) | `POST /roles {qa lead}` | **409** "Role name already exists" |
| 6a | Invalid name | `POST /roles {bad@name!}` | **400** "can only contain letters, numbers, spaces, hyphens, and ampersands" |
| 6b | Unknown permission | `POST /roles {permissions:[made:up]}` | **400** "Unknown permissions: made:up" |
| 7 | Update (rename + perms) | `PATCH /roles/{id}` | **200**, name "QA Director", permission_count 1 |
| 8a | System role rename | `PATCH /roles/{Admin}` name | **400** "Cannot rename a system role" |
| 8b | System role level change | `PATCH /roles/{Admin}` level | **400** "Cannot change the level of a system role" |
| 9a | Delete (0 users) | `DELETE /roles/{QA Director}` | **200** |
| 9b | Get deleted | `GET /roles/{id}` | **404** "Role not found" |
| 9c | Recreate same name | `POST /roles {QA Director}` | **201** (name reusable after soft-delete) |
| 10 | Delete blocked by assigned users | assign role to user → `DELETE` | **409** "Cannot delete role 'Temp Assigned' — 1 user is currently assigned. Please reassign them before deleting." |
| 11a | **Level authority** — Admin(90) grants Super Admin(100) | `PUT /users/{id}/roles` as admin-level actor | **403** "Cannot assign role 'Super Admin' (level 100): exceeds your authority level (90)" |
| 11b | Level authority — Admin(90) grants Manager(50) | `PUT /users/{id}/roles` | **200** (within authority) |
| G1 | Gate-tightening — catalog without `roles:read` | `GET /roles/permissions` as Employee | **403** "Insufficient permissions" |
| G2 | Gate — list without `roles:read` | `GET /roles` as Employee | **403** |

### DB spot-check (`SELECT name, level, is_system, is_deleted FROM roles ORDER BY level, name`)

```
    name     | level | is_system | is_deleted
-------------+-------+-----------+------------
 Employee    |    10 | t         | f
 QA Director |    40 | f         | f      <- recreated (live)
 QA Director |    40 | f         | t      <- soft-deleted original
 Manager     |    50 | t         | f
 HR Manager  |    80 | t         | f
 Admin       |    90 | t         | f
 Super Admin |   100 | t         | f
```

The two `QA Director` rows (one live, one soft-deleted) prove D2 — soft-delete with
name reuse enforced by the `LOWER(name) WHERE is_deleted=FALSE` partial-unique index.

## 4. Decisions verified against the locked audit

- **D1 (level + authority)** — column present + CHECK(1..100); seed levels 100/90/80/50/10; assignment-authority enforced at the service AND proven over HTTP (step 11). Reopened-and-resolved deferred **#15**.
- **D2 (soft-delete, name freed)** — step 9 + DB spot-check.
- **D3 (sort by level asc)** — step 2 + the `ORDER BY level ASC, LOWER(name) ASC` repo query.
- **Catalog gated by `roles:read`** — G1 (was auth-only before this branch).
- **is_system guards / name regex / permission validation** — steps 8, 6a, 6b.
- **One list endpoint with full `permissions[]` + `permission_count`** — step 2.

## 5. Out of scope (unchanged, as audited)

`leave_requests:approve_team`/`approve_all` registry split (left to the
leave-requests parity pass). Go-only `employees:*`/`dependents:manage`/`invites:manage`
registry additions untouched.

## Commits (this branch)

| Commit | Task |
|---|---|
| `6dfd851` | docs: parity audit + plan |
| `6d58ad4` | test: cross-platform iofs migration source (Windows fix) |
| `a628470` | T1 migration 000020 + model level |
| `bb95bb4` | T2 role DTOs (+ RoleRead→RoleRef brief rename) |
| `a0862c6` | T3 repo List/SoftDelete/CountUsersWithRole |
| `736ef5d` | T4 RoleService CRUD + migration 000021 partial-unique |
| `4ca6df3` | T4 review fixes (uq_ prefix + LOWER(name); FindByName doc) |
| `79b1a59` | T5 level-based assignment authority |
| `7c28472` | T6 wire routes, gate catalog, seed levels |
| `cd1a53c` | T7 swagger regen |
