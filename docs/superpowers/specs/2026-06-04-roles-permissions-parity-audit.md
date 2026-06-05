# Roles & Permissions — Python ↔ Go API Parity Audit

**Date:** 2026-06-04
**Module:** Roles & Permissions (RBAC management)
**Status:** AUDIT — decisions LOCKED 2026-06-04 (no code yet)
**Method:** module-by-module parity audit (same pipeline as announcements/employees:
audit → locked decisions → PR(s) → verification log → FE doc → handoff)

**Sources read**
- Python: `app/routers/roles.py`, `app/services/role.py`, `app/schemas/role.py`,
  `app/models/role.py`, `app/core/permissions.py` (at `E:\Work\exnodes-hrm-api`)
- Go: `internal/handlers/role_handler.go`, `internal/repositories/role_repo.go`,
  `internal/models/role.go`, `internal/permissions/registry.go`,
  `internal/services/seed_service.go`, `cmd/server/main.go` (route wiring),
  `internal/services/auth_service.go` (`ResolveUserPermissions`),
  `internal/services/user_service.go` + `user_handler.go` (`AssignRoles`)
- BA intent: `requirements/EP-001-foundation/US-004-role-permission-management/details/`
  DR-001-004-01 (list), -02 (create), -03 (edit), -04 (delete); and
  US-005/DR-001-005-05 (change-user-role)

---

## 0. TL;DR

**The Go side has no role-management API.** It ships exactly one endpoint —
`GET /api/v1/roles/permissions` (the permission catalog for the FE picker) — and
roles are otherwise *seed-only* (`seedRoles` in `seed_service.go` creates the 5
system roles and merges perms on boot). Python ships a full role CRUD suite
(list ×2, get, create, update, delete) plus a role-level authority hierarchy.

To reach parity (and satisfy BA US-004), Go needs: **list, get, create, update,
delete** role endpoints, a paginated/searchable repo layer, a "users assigned"
count for the delete guard, role-name validation, and `is_system` write guards.

Three decisions were contested (Python vs BA vs Go conventions disagreed) and are
now **LOCKED** (2026-06-04) — see **§5 Decisions**:
- **D1** → **ADD role `level` (1–100) + assignment-authority** (full Python parity;
  needs migration 000020; reopens the previously-deferred #15)
- **D2** → **soft delete** (Go convention; name freed for reuse via the live index)
- **D3** → **sort by `level` ascending** (Python parity; coherent now level stays)

---

## 1. Endpoint inventory

| # | Endpoint | Python | Go | Gap |
|---|---|---|---|---|
| 1 | `GET /roles/permissions` (grouped catalog) | ✅ gated `roles:read` | ✅ **auth-only, no perm** | perm-gate mismatch (D6) |
| 2 | `GET /roles/role-and-permission` (paginated, full perms) | ✅ gated `roles:read` | ❌ missing | **add** |
| 3 | `GET /roles/` (paginated list, `permission_count`) | ✅ gated `roles:read` | ❌ missing | **add** |
| 4 | `GET /roles/{id}` (single, full perms) | ✅ gated `roles:read` | ❌ missing | **add** |
| 5 | `POST /roles/` (create, 201) | ✅ gated `roles:create` | ❌ missing | **add** |
| 6 | `PATCH /roles/{id}` (update) | ✅ gated `roles:update` | ❌ missing | **add** |
| 7 | `DELETE /roles/{id}` (delete) | ✅ gated `roles:delete` | ❌ missing | **add** |

Permission constants `roles:{read,create,update,delete}` **already exist** in the
Go registry (`permissions/registry.go`) and are already seeded (Admin holds
read+create+update; HR Manager holds read; see `seed_service.go`). So the perm
plumbing is ready — only the handlers/service/repo are absent.

> Note: Python exposes **two** list endpoints — `/role-and-permission` (returns
> full `permissions[]` per row) and `/` (returns `permission_count` only). The BA
> Role-List screen (DR-001-004-01 §4) shows the permission *names* comma-separated
> in the list, i.e. it needs the full-permissions variant. See D5.

---

## 2. Data model comparison

| Field | Python (`Role` Document) | Go (`models.Role`) | Notes |
|---|---|---|---|
| id | Mongo ObjectId | UUID (`BaseModel`) | — |
| name | `Indexed(str, unique=True)` | `text unique` | parity |
| description | `str = ""` | `text default ''` | parity |
| **level** | `int` (1–100) | **absent** | **D1** |
| permissions | `list[str]` | `StringSlice` (jsonb) | parity |
| is_system | `bool = False` | `bool default false` | parity |
| created_at / updated_at | yes | yes (`BaseModel` + trigger) | parity |
| **is_deleted / deleted_at** | **absent** (hard delete) | yes (soft-delete cols) | **D2** |

Go's `Role` already carries the four audit columns and the `NotDeleted` scope
(repo `notDeleted`), so the table *supports* soft delete today; Python hard-deletes
via `role.delete()`.

---

## 3. Business-rule comparison

| Rule | Python | BA requirement | Go today |
|---|---|---|---|
| Name uniqueness (case-insensitive) | ✅ regex `^name$` /i on create + update-excl-self | ✅ AC-06 create, AC-07 edit (excl self) | repo `FindByName` exists; no service |
| Name format validation | `^[a-zA-Z0-9 &-]+$`, trim, 1–100 | ✅ letters/digits/space/`-`/`&`, trim, ≤100 | none |
| Permission-string validation | reject unknown (allow `*`) | implied (picker is API-driven) | `permissions.IsValid` exists; unused for roles |
| Create sets `is_system=false` | ✅ | implied | n/a |
| System role: no rename | ✅ `BadRequest` | implied (system roles protected) | n/a |
| System role: no level change | ✅ | n/a (no level in BA) | n/a |
| System role: no delete | ✅ `BadRequest` | implied | n/a |
| Delete blocked if users assigned | ✅ counts `User.role_ids == oid`, returns `user_count` | ✅ SR-01/AC-04 server-side, returns N | **no count method** |
| Delete is hard | ✅ | ✅ SR-02 "hard delete, no soft-delete/archive" | n/a (**D2**) |
| Partial update (only sent fields) | ✅ `exclude_unset` | ✅ edit form | n/a |
| Permission changes take effect immediately | ✅ (perms resolved per request) | ✅ AC-14 / SR-05 | ✅ already — `ResolveUserPermissions` reads live |
| Default sort | `+level` ascending | **alphabetical A→Z** (AC-02, SR-07) | n/a (**D3**) |
| Search by name (case-insensitive, partial) | ✅ `$regex /i` | ✅ AC-06/07 | `BuildILIKEPattern` util exists |
| Pagination | ✅ page/page_size (≤100) | ✅ default 10, options 25/50 | use existing paginated envelope |

### Role-assignment authority (the `level` system)

Python `role.py` has `get_user_max_role_level` + `check_role_assignment_authority`:
an assigner can only grant roles whose `level` ≤ the assigner's own max level.
This is invoked during user create / change-role.

- **BA never mentions role levels.** DR-001-004-02 (create) field list has no
  `level` field; DR-001-005-05 (change-user-role) is a plain dropdown with no
  authority-tier logic. The level system is a Python-internal construct not
  surfaced in the confirmed requirements.
- Go's `AssignRoles` (`PUT /users/:id/roles`, gated `users:manage_roles`) has a
  self-guard ("You cannot change your own role" — matches DR-001-005-05 SR-05)
  but **no level check** and does not validate that the role IDs exist.
- Prior parity work already logged this as **deferred #15 — "N/A, Go RBAC has no
  role level"** (CHECKPOINT). D1 revisits whether to keep that stance.

---

## 4. Permission-registry parity (catalog content)

The grouped catalog (`PERMISSION_GROUPS` py / `PermissionGroups` go) is the payload
of endpoint #1. Content differs because of the Go schema split + leave-phase choices:

**Go-only groups/keys (intentional, from the migration):**
- `employees:*` (read/create/update/delete) — Go's `employees` aggregate is split
  out of Python's monolithic `users`.
- `dependents:manage` — Go-side dependent management perm.
- `invites:manage` — Phase 9 invites.

**Python-only keys:**
- `leave_requests:approve_team` **and** `leave_requests:approve_all` (a split).
  Go collapses these into a single `leave_requests:approve` (+ a recognised legacy
  `leave_requests:approve` string in Python for back-compat). Known leave-phase
  divergence — **out of scope for this audit**, flag only.

Everything else (users incl. salary/banking ×4, roles, departments, positions,
skills, leave_quota, attendance, organization_settings, announcements) matches
key-for-key. → **D7**: decide whether the registry delta is in-scope here or left
to the leave-requests parity pass.

---

## 5. Decisions

### Contested — LOCKED 2026-06-04

**D1 — Role `level` hierarchy → ADD level + authority** (full Python parity).
Go gains a `level int` (1–100) column and the assignment-authority rule: an
assigner may only grant roles whose `level` ≤ the assigner's own max level.
Implications: **migration 000020** adds `level` (default 100) to `roles`;
`RoleCreate`/`RoleUpdate` carry `level` (validated 1–100); system roles cannot
have `level` changed (Python guard); seed roles get explicit levels (Super Admin
highest → Employee lowest); `users.AssignRoles` gains a level-authority check
ported from Python's `check_role_assignment_authority` /
`get_user_max_role_level`. This **reopens the previously-deferred #15** — update
CHECKPOINT's "N/A" note. Seed level assignment (proposal, descending authority):
Super Admin 100, Admin 90, HR Manager 80, Manager 50, Employee 10 — confirm at
implementation.

**D2 — Delete semantics → SOFT delete, name freed** (Go convention kept).
Delete writes `is_deleted=true, deleted_at=NOW()` via the repo; the row vanishes
from all reads (`NotDeleted` scope) and the name becomes reusable through the live
partial-unique index. Honours BA SR-02's observable contract ("permanently
removed, name reusable") without breaking the no-AutoMigrate/soft-delete house
rule. The delete-blocked-if-users-assigned guard still applies first.

**D3 — Default sort → by `level` ascending** (Python parity; coherent now that
D1 keeps level). Search-by-name (case-insensitive, partial via `BuildILIKEPattern`)
layers on top. Note: this diverges from BA AC-02/SR-07 ("alphabetical A→Z") — the
FE may still sort client-side; flag in the FE handoff. Secondary sort by name
within equal levels for stable ordering.

### Proposed (low-risk, will implement unless objected)

- **D4 — Endpoints to add:** list (full perms, paginated, name-search),
  `GET /{id}`, `POST` (201), `PATCH /{id}` (partial), `DELETE /{id}`. Mirror the
  Python route shapes under `/api/v1/roles`.
- **D5 — List shape:** ship **one** list endpoint returning full `permissions[]`
  **and** a `permission_count` (superset of both Python lists) so the BA Role-List
  permission column works without a second call. Keep the Python path name
  `GET /roles/` ; skip the redundant `/role-and-permission` alias (or add it as a
  thin alias if FE already calls it — confirm with web repo).
- **D6 — Gate `GET /roles/permissions` behind `roles:read`** to match Python
  (today it's auth-only). Low risk: every actor who needs the picker (create/edit
  role) already holds `roles:read`.
- **D7 — Registry delta** (`approve_team`/`approve_all`): **out of scope**, leave
  to the leave-requests parity pass; note only.
- **D8 — DTOs:** `RoleCreate{name, description, permissions[]}`,
  `RoleUpdate{*pointer fields}` (partial), `RoleRead{id,name,description,
  permissions[],permission_count,is_system,created_at,updated_at}`,
  `RoleListQuery{page,page_size,search}`. Validate name regex + permission strings
  (`permissions.IsValid`) at the DTO/service boundary.
- **D9 — `is_system` guards** in the service: block rename/delete of system roles
  (`BadRequest`), as Python does.
- **D10 — Repo additions:** `List(ctx, q) ([]Role,total)`, `SoftDelete`/`Delete`,
  `CountUsersWithRole(ctx, roleID) (int64)` (join `user_roles`). Reuse
  `BuildILIKEPattern` for search.

---

## 6. Proposed work breakdown (after decisions locked)

1. **migration `000020_roles_add_level`** — add `level int NOT NULL DEFAULT 100`
   (D1); backfill seed roles to their assigned levels in the migration or seed.
2. `internal/dto/role.go` — Create/Update/Read/ListQuery + validators (name regex,
   permission-string check, `level` 1–100).
3. `role_repo.go` — add `List` (sort `+level`, name search), soft-delete,
   `CountUsersWithRole`.
4. `role_service.go` (new) — CRUD + uniqueness + is_system guards (no rename / no
   level-change / no delete) + delete-blocked-if-assigned; port level-authority
   into `user_service.AssignRoles`.
5. `role_handler.go` — wire 5 new handlers; gate catalog behind `roles:read`.
6. `cmd/server/main.go` — register routes with `RequirePerms`; update `seedRoles`
   to set `level` on the 5 system roles.
7. `make swag` regen; integration tests; live HTTP smoke; verification log
   `docs/superpowers/verification/roles-permissions-parity.md`.
8. FE handoff doc (`api_info_go/roles.md` in web repo) + CHECKPOINT update.

---

## 7. Open questions for the web repo / PO

- Does the FE call `GET /roles/role-and-permission`, `GET /roles/`, or both?
  (decides D5 alias).
- Is `level` surfaced anywhere in the live web UI? (decides D1).
- Confirm BA "hard delete" is satisfied by soft-delete-that-frees-the-name (D2).
