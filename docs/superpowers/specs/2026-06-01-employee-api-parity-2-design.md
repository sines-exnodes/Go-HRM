# Employee API Parity Round 2 — Design

**Date:** 2026-06-01
**Branch:** `feat/employees-parity-2`
**Status:** Approved (brainstorming) — pending implementation plan
**Migrations:** `000018_employee_name_split`, `000019_employee_experience_year_to_start_year`
**Source of intent:** BA docs (DR-001-005-01/02/03/04, EP-008 US-003) + Python parity
(`/Users/panda/Documents/Work/exnodes-hrm-api`)

---

## 1. Purpose

Reconcile the Go employee/user API with the Python source and the current BA
requirements across five areas the FE depends on:

1. **List filters** (DR-001-005-01) — make the department/position/role/status
   filters actually work (today they `400` on any value) and support multi-select.
2. **Experience year** (DR-001-005-02/03/04) — treat `experience_year` as a
   4-digit *career-start year* (≤ current year), not a count of years.
3. **Skills** (DR-001-005-02/03, EP-008 US-003) — assign skills inline on user
   create/update.
4. **Direct reports** (DR-001-005-03) — keep the standalone endpoint (Python parity).
5. **Names** — split the stored `full_name` into `first_name` + `last_name`.

Plus an adjacent fix surfaced by 1 & 3: an employee's **own** department/position
names are never resolved on the employee read (a Phase-3 gap → currently `null`).

This is post-migration parity work, following the established pattern: audit →
locked decisions → PR → verification log → FE handoff. Branch convention matches
`feat/employees-parity`, `feat/employees-salary-banking`, `feat/employees-line-manager`.

## 2. Parity findings (evidence-backed)

A 5-dimension audit of the Python source produced these verdicts:

| Dimension | Python's actual shape | Decision |
|---|---|---|
| **Names** | `first_name` + `last_name` stored separately (both required, min 1/max 100); **no stored `full_name`** — derived `f"{first} {last}".strip()` only in briefs. Search = first OR last OR email OR phone. Sort = `first_name`, then `last_name`. (`app/models/user.py:88-89`, `app/schemas/user.py:122-123,170-171,223-224`, `app/services/user.py:75,274-279,298`) | **MATCH** → drop `full_name`, add first/last. |
| **Experience** | `experience_year: int\|None` = career start year (test fixture `2018`); BA mandates `>1900 && ≤ current year`. (`app/models/user.py:110`, `tests/test_user_line_manager.py:51`) | **MATCH** → from-year + validate + migrate. |
| **Skills** | `skill_ids` array on user; accepted **inline** on `AdminUserCreate` (`list[str]=[]`) and `AdminUserUpdate` (`list[str]\|None`); read returns `skills:[{id,name}]`; **no separate endpoint** in Python. (`app/models/user.py:109`, `app/schemas/user.py:137,184,237`, `app/services/user.py:115-118,342-343,467-468`) | **MATCH** → inline `skill_ids`; keep Go's `PUT .../skills` as a superset. |
| **Direct reports** | Returned **only** via `GET /users/{id}/direct-reports`; `UserRead` has **no** `direct_reports` field; computed live from inverse `line_manager_id`, sorted by full name, incl. inactive. (`app/routers/users.py:211-221`, `app/services/user.py:698-733`) | Standalone endpoint **only** (no embed). |
| **List filters** | `role_ids` multi-select (`$in`); **`department_id` & `position_id` single-value**; `is_active` single bool; search first/last/email/phone; sort `first_name`. (`app/routers/users.py:104-114`, `app/services/user.py:273-299`) | BA (DR-005-01) overrides Python here → **all of dept/position/role multi-select**. |

The filters row is a deliberate **BA-over-Python** divergence: DR-001-005-01
(AC-09/10/11/14, SR-05) requires Department, Position, and System Role to all be
multi-select chips. The user confirmed the BA is authoritative for filters.

## 3. Migrations

Two files, one concern each, for clean review/rollback. Next versions: 18, 19.

### `000018_employee_name_split`
- **Up:** add `first_name TEXT NOT NULL DEFAULT ''`, `last_name TEXT NOT NULL DEFAULT ''`;
  backfill `first_name = split_part(full_name,' ',1)`,
  `last_name = btrim(substr(full_name, length(split_part(full_name,' ',1))+1))`;
  `DROP COLUMN full_name`. Single-token legacy names land as `first_name=token,
  last_name=''` (the `DEFAULT ''` tolerates them; the `min=1` rule applies only to
  *new* writes through the DTO).
- **Down:** re-add `full_name TEXT NOT NULL DEFAULT ''`; backfill
  `full_name = btrim(first_name||' '||last_name)`; drop the two columns. Lossless.

### `000019_employee_experience_year_to_start_year`
- **Up:** `UPDATE employees SET experience_year = EXTRACT(YEAR FROM CURRENT_DATE)::int - experience_year WHERE experience_year IS NOT NULL AND experience_year < 1900;`
  — converts a count (`7` → `2019`); leaves already-year-shaped values (`≥1900`) alone.
- **Down:** inverse arithmetic + `RAISE NOTICE` documenting the round-trip is
  approximate (depends on the apply-date year). Honours the data-loss-guard
  convention (migration 000016 pattern).

No migration is needed for filters, skills, direct reports, or the dept/position
display fix — those are read-path/DTO changes only.

## 4. Models — `internal/models/employee.go`

- Remove `FullName string`; add `FirstName string` + `LastName string`
  (`gorm:"type:text;not null;default:''"`).
- Add `func (e *Employee) DisplayName() string { return strings.TrimSpace(e.FirstName + " " + e.LastName) }`
  for the briefs that compose a display name.

## 5. DTOs — `internal/dto/employee.go`, `internal/dto/auth.go`

- **`EmployeeCreate`**: `FullName` → `FirstName` + `LastName` (`required,min=1,max=100`);
  add `SkillIDs []uuid.UUID` (optional).
- **`EmployeeUpdate`**: `FullName *string` → `FirstName *string` + `LastName *string`
  (`omitempty,min=1,max=100`); add `SkillIDs *[]uuid.UUID` (pointer-to-slice PATCH:
  `nil`=leave, `[]`=clear, non-empty=replace).
- **`EmployeeSelfUpdate`**: `FullName` → `FirstName` + `LastName` (self may edit own
  name, per Python `UserUpdate`). Skills remain **out** of self-update (admin-managed).
- **`EmployeeRead`**: `FullName` → `FirstName` + `LastName` (no `full_name` field —
  Python parity); `Department`/`Position` refs now **populated**.
- **`ManagerBrief` / `DirectReportRead` / `ManagerCandidateRead`**: keep their composed
  `full_name` (Python `LineManagerRead`/`DirectReportRead` do exactly this).
- **`EmployeeSummary`** (auth login + `/users/me`): `FullName` → `FirstName` +
  `LastName` (cascades to `auth_service` / `user_service` summary builders).
- **`EmployeeListQuery`**: `DepartmentID/PositionID/ManagerID/RoleID` `*uuid.UUID` →
  `[]string` (`form:"department_id"` … — repeated query params); `Search` unchanged.
  The BA's 2-value **Status** chip (Active/Inactive) collapses to the single optional
  `IsActive *bool` — the FE sends it only when exactly one status is selected
  (both selected, or none, = omit = show all). No separate `status[]` param.

## 6. Repository — `internal/repositories/employee_repo.go`

- **`List`**:
  - Filters: `employees.department_id IN ?`, `employees.position_id IN ?`,
    `employees.manager_id IN ?`, and `role_id` via
    `EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=employees.user_id AND ur.role_id IN ? AND ur.is_deleted=false)`. `users.is_active = ?` unchanged.
  - **Preload `Department` + `Position`** (the dept/pos fix).
  - Search → `employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR
    employees.phone ILIKE ? OR employees.personal_email ILIKE ? OR users.email ILIKE ?`.
  - Order → `employees.first_name ASC, employees.last_name ASC`.
- **`FindByIDWithFull` / `FindByUserIDWithFull`**: add
  `Preload("Department", notDeleted)` + `Preload("Position", notDeleted)`.
- **`ListDirectReports`** & **`ListManagerCandidates`**: order/search switch from
  `full_name` to `first_name, last_name` (candidate search keeps the
  position/department LEFT JOINs).

## 7. Service — `internal/services/employee_service.go`

- **`toRead`**: set `FirstName/LastName`; resolve `out.Department`/`out.Position`
  into `RefRead{ID,Name}` from the preloaded relations.
- **Experience validation**: `validateExperienceYear(*int) error` → if non-nil, must
  be `1901..time.Now().UTC().Year()` else `ErrBadRequest`. Called in Create + Update.
- **Skill assignment wiring**: inject a narrow interface
  `type skillAssigner interface { ReplaceForEmployee(ctx, employeeID uuid.UUID, skillIDs []uuid.UUID) ([]dto.SkillRead, error) }`
  backed by the existing `*SkillService` (acyclic — SkillService depends only on the
  employee *repo*). Validate skill IDs **before** the create-tx; apply the replace
  **after** the employee row commits (mirrors the emergency-contacts ordering).
  Update applies the pointer-to-slice PATCH semantics.
- **`Create`/`Update`/`SelfUpdate`**: first/last handling (trim); briefs use
  `DisplayName()`.
- **Wiring** (`cmd/server/main.go`): construct `skillSvc` before `empSvc` and pass it
  into `NewEmployeeService` as the `skillAssigner`.

## 8. Handlers / routes / seed / docs

- **`List` handler** (`employee_handler.go`): parse each `[]string` filter →
  `[]uuid.UUID` (invalid uuid → `400`); update Swagger `@Param` to multi.
- **`Create`/`Update`/`SelfUpdate` handlers**: bind the new DTO fields (first/last,
  `skill_ids`). No new field-perm gate (skills are under base `employees:update`).
- **Routes** (`cmd/server/main.go`): unchanged — the direct-reports endpoint stays;
  no embed.
- **Seed** (`internal/services/seed_service.go`): admin user `FullName` →
  `FirstName`/`LastName`.
- **Swagger**: regenerate via `make swag` (never hand-edited).

## 9. Tests & verification (never skipped — AGENTS.md Rule 12)

- Update fixtures/seed/smoke (`full_name`→first/last; `experience_year` now a year).
- New tests: multi-select `IN` filters (incl. invalid-uuid `400`); experience-year
  rejection (future, `≤1900`) + acceptance; inline `skill_ids` on create + update
  (replace/clear/leave); dept/position resolved on read; name-split round-trip;
  **migration up/down round-trip** for both 18 & 19.
- Live HTTP smoke + DB spot-check → commit
  `docs/superpowers/verification/employees-parity-2.md`.
- Update `docs/superpowers/CHECKPOINT.md` + serena `code_map.md` / `conventions.md`.

## 10. Commit sequence

1. Migrations 000018 + 000019 (+ model field swap).
2. DTOs + repo (filters, preloads, ordering).
3. Service (toRead/dept-pos, experience validation, skill wiring).
4. Handlers + routes + seed.
5. Swagger regen.
6. Tests + smoke.
7. Verification log + checkpoint.

## 11. Out of scope

- Avatar / CV / ID-card image **upload** endpoints (URLs still accepted — deferred
  follow-up #11).
- Name-split **frontend** work (the web repo is self-managed).
- Embedding `direct_reports[]` in the detail read (rejected for Python parity).
- Any non-employee module.

## 12. Open risks / notes

- **Name backfill heuristic** is best-effort for multi-word names; real data is
  minimal (seed + smoke fixtures), so impact is low. Documented in the migration.
- **Experience down-migration** is arithmetically approximate across a year boundary
  — acceptable and noted.
- **Search field set** is a slight Go superset of Python (adds `personal_email`);
  preserved intentionally rather than narrowed.
