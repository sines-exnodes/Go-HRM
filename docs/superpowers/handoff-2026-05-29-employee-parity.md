# Handoff — Employees parity, deferred follow-ups

**Date:** 2026-05-29
**Audit:** Python `users` module ↔ Go `employees`/`users`/`dependents` split. 19 decisions captured + locked.
**Shipped this session (two PRs, both pushed, not yet opened — awaiting owner OK):**

| Branch | PR | Covers |
|---|---|---|
| `feat/employees-parity` (off `main`@`bcb4c0b`) | PR A | #4 emergency-contact list · #5 leave-quota on read · #7 widened self-edit · #8 experience_year/cv_url/skills · #12 self/destructive guards · #13 admin change-email · #17/#17b enums |
| `feat/employees-salary-banking` (stacked on PR A) | PR B | #6 salary/banking field-level perms + masking + write-gate |

Verification logs: `docs/superpowers/verification/employees-parity-pr-a.md`, `…-pr-b.md`.
FE docs (uncommitted, web repo): `exnodes-hrm-web-next/api_info_go/employee.md` (new) + `me.md` (refreshed).

The remaining 3 decisions were **deferred** — specs below so a future session can pick them up cleanly.

---

## The 19 locked decisions — DO NOT re-litigate

| # | Decision | Outcome |
|---|---|---|
| 1 | Profile architecture | Keep Go split |
| 2 | Name field | Keep `full_name` |
| 3 | Sub-objects | Keep Go flat columns |
| 4 | Emergency contacts | List ✅ PR A |
| 5 | Leave quota on read | Hydrate ✅ PR A |
| 6 | Salary/banking access | Granular perms + mask ✅ PR B |
| 7 | Self-edit set | Widen to name/gender/dob ✅ PR A |
| 8 | Missing read fields | experience_year/cv_url/skills ✅ PR A |
| 9 | Mandatory-on-create | Keep optional |
| **10** | **Line-manager suite** | **DEFERRED (see below)** |
| **11** | **CV/ID-card upload** | **DEFERRED (see below)** |
| 12 | Self/destructive guards | ✅ PR A |
| 13 | Admin change-email | ✅ PR A |
| 14 | Reset-password/send-invite | Keep Go invites; reset-pw stays deferred |
| **15** | **Role-assignment authority** | **DEFERRED / N/A (see below)** |
| 16 | List filters | Keep Go |
| 17/17b | marital_status / education enums | ✅ PR A |
| 18 | Device-token endpoint | Keep Go split |
| 19 | Dependents | Keep Go-only |

---

## Decision #10 — Line-manager management suite (a feature, not a tweak)

Python has a full line-manager subsystem the Go bare `manager_id` doesn't replicate. Build as its own PR after PR A/B merge.

### What Python has (source: `exnodes-hrm-api/app/services/user.py`)
- **Rich resolved manager** on the profile: `{ id, full_name, position, department, is_active }` (Go returns slim `{ id, name }`).
- **`GET /users/line-manager-candidates`** — picker options. Filters: excludes deleted; excludes self (when `for_user_id` given); excludes the **entire transitive subordinate chain** (cycle prevention, BFS over the inverse edge); active-only **except** keeps a currently-assigned-but-deactivated manager visible. `search` matches name/position/department; `limit` 1..200.
- **`GET /users/{id}/direct-reports`** — all live users whose `line_manager_id == id` (active + inactive), sorted by name, with resolved position/department.
- **Assignment validation** (on create + update): manager must exist, be non-deleted **and active**; cannot be self; cannot be anyone in the target's transitive subordinate chain (cycle).
- **On delete**: clears `line_manager_id` on direct reports. *(Go PR A already does the equivalent — NULLs `manager_id` on direct reports when a manager is deleted.)*

### Go build spec (suggested)
- New endpoints under the employees tree (FK targets `employees(id)`, not users):
  - `GET /api/v1/employees/manager-candidates?for_employee_id=&search=&limit=` (`employees:read`)
  - `GET /api/v1/employees/{id}/direct-reports` (`employees:read`)
- Service: `subordinateChain(employeeID)` BFS over `employees.manager_id`; `validateManagerAssignment(managerID, targetID)` enforcing exists+active+non-self+no-cycle. Call it inside `Create` (target=nil) and `Update` (target=id) **before** persisting `manager_id`.
- Enrich the read `manager` ref to the rich object (or add a `manager_detail`), resolving position+department+is_active. Watch the N+1 — batch-resolve in `List`.
- Migration: none (uses existing `employees.manager_id`). DTO: `ManagerCandidateRead`, `DirectReportRead` (id, full_name, avatar_url, position, department, is_active).
- Tests: cycle rejection, self rejection, inactive-manager rejection, candidate exclusion of the subordinate chain, direct-reports includes inactive.

---

## Decision #11 — Server-side CV / ID-card upload

Python's multipart create/update accept `avatar` + `cv` + `id_card_front` + `id_card_back` and store them; file uploads override any URL in the JSON. Go currently:
- Has avatar upload (`PATCH /employees/{id}/avatar`, `PATCH /employees/me/avatar`) — multipart, `http.DetectContentType` sniff + allowlist.
- Accepts `cv_url`, `id_front_image`, `id_back_image` only as **URL strings** (the FE uploads elsewhere and passes the URL). `cv_url` column shipped in PR A.

### Go build spec (suggested)
- New multipart endpoints reusing the avatar sniff/allowlist helper (extract a shared `readUpload(c, field, maxBytes, allowedMIME)` — the CHECKPOINT already flagged consolidating the 4 sniff sites):
  - `PATCH /api/v1/employees/{id}/cv` — PDF/DOCX allowlist, max 5MB → sets `cv_url` (`employees:update`); self variant `PATCH /employees/me/cv` if desired.
  - `PATCH /api/v1/employees/{id}/id-card` — `front` and/or `id_card_back` image parts → sets `id_front_image`/`id_back_image`; sending only one doesn't clear the other; delete old file on replace.
- Sniff the bytes (client `Content-Type` is a hint only — mandatory per `mem:conventions`). Best-effort delete of the previous object on replace.
- Tests: content-type sniff rejects spoofed type; partial id-card update preserves the other side.

---

## Decision #15 — Role-assignment authority (N/A without role levels)

Python's `check_role_assignment_authority` blocks assigning a role whose `level` exceeds the assigner's max role level. **Go's `Role` model has no `level` field** — the RBAC model is flat (permission strings only). Porting this requires:
1. Adding a `level INT` column to `roles` (migration) + seeding sensible levels for the 5 system roles.
2. A `maxRoleLevel(user)` helper + the assignment check in `UserService.AssignRoles` and `EmployeeService.Create` (role_ids path).

This is a **roles-module** change, out of scope for the employees aggregate. Only do it if the BA confirms level-based authority is required; otherwise leave `AssignRoles` as-is (it already blocks changing your own role).

---

## Loose ends

- **FE docs are on disk, uncommitted** in `exnodes-hrm-web-next/api_info_go/` (`employee.md` new, `me.md` updated). Commit on the web side (amend the existing FE parity branch/PR or a new docs branch), same as the announcement.md handoff dance.
- **PR B is stacked on PR A.** Merge order: PR A → main, then rebase PR B onto main (or merge A into B first). If opening PR B against `main` before A merges, its diff will include A's commits.
- **`docs/superpowers/handoff-2026-05-27-announce-targeting.md`** is still untracked (prior session). Unrelated to this work.
- Migration numbering: PR A added **000017**. Next free migration is **000018** (the line-manager suite needs none; CV/ID upload needs none).
- The seeded salary/banking perms (PR B) reach existing Admin/HR Manager roles via the **merge-seed on boot** — no manual migration of role rows needed.

## Pointers

- **Audit + decisions:** this session's conversation (search "Parity audit table — 19 decisions").
- **Python source of truth:** `exnodes-hrm-api/app/{routers,schemas,services,models}/users.py` + `user.py` (no `employee*.py`/`dependent*.py` exist — that's the split).
- **Go source of truth:** `internal/{handlers,services,dto,models,repositories}/{employee,dependent,user}*.go`.
- **Verification:** `docs/superpowers/verification/employees-parity-pr-{a,b}.md`.

When resuming the deferred work, start with:
> "Pick up the deferred employees-parity follow-ups — read `docs/superpowers/handoff-2026-05-29-employee-parity.md` and build decision #10 (line-manager suite) first."
