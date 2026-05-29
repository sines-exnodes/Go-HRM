# Verification — Line-manager suite (deferred decision #10)

**Date:** 2026-05-29
**Branch:** `feat/employees-line-manager` (off `feat/employees-salary-banking` / PR B)
**Migration:** none (uses existing `employees.manager_id` + departments/positions FKs).
**Scope:** the deferred line-manager feature from the employees parity audit (#10).

## What changed

| Area | Change |
|---|---|
| Assignment validation | `validateManagerAssignment` (exists + active + not-self + no-cycle via transitive subordinate BFS), wired into `Create` (existence/active only — no target yet) and `Update` (full self/cycle check). |
| Picker | `GET /employees/manager-candidates?for_employee_id=&search=&limit=` (`employees:read`) — active employees, excluding the target + its transitive subordinate chain; keeps a currently-assigned but deactivated manager visible; search matches name/position/department. |
| Direct reports | `GET /employees/{id}/direct-reports` (`employees:read`) — live reports (active AND inactive), sorted by name. |
| Rich manager brief | `EmployeeRead.manager` upgraded from `{id,name}` to `{id, full_name, position, department, is_active}` (matches Python's LineManagerRead). |
| Model/repo | `Employee.Department`/`Position` relations; nested `Manager.User/Department/Position` preloads; `SubordinateIDs` (BFS), `ListManagerCandidates`, `ListDirectReports`, `FindByIDWithOrg`. |

## 1. Build / vet / tests — green

```
go build ./... · go vet ./...  → clean
go test ./internal/services/ ./internal/permissions/  → ok (full integration suite, ~102s)
```
8 new integration tests (`employee_line_manager_test.go`), all passing:
`RejectsSelfAssignment`, `RejectsCycle`, `RejectsMissingManager`, `RejectsInactiveManager` (update + create),
`RichManagerBriefOnRead`, `Candidates_ExcludesSelfAndSubordinates`, `Candidates_KeepsInactiveCurrentManager`,
`DirectReports_IncludesInactive`.

## 2. End-to-end HTTP smoke — green

`scripts/smoke-employees-parity.sh` extended with an 11-check #10 section; full script **34/34 PASS**:
valid-manager create (201), rich manager brief on read (id + is_active), self-as-manager (400), cycle (400),
nonexistent manager (400), manager-candidates (200) excluding self + subordinate, direct-reports (200) including the report.

## 3. Two real bugs found by e2e (fixed)

- **Zero-UUID manager → 500.** `validateManagerAssignment` treated `uuid.Nil` as "no manager" and skipped validation, so a client-supplied `manager_id` of `000…0` reached the DB and tripped the FK (500). Fix: callers (`Create`/`Update`) invoke validation only when `manager_id` is actually supplied; the validator no longer early-returns on Nil, so any supplied-but-bogus id (including the zero UUID) gets a clean **400 "does not exist"**.
- **`for_employee_id` query → 400.** gin cannot bind a `uuid.UUID` from a query param. Fix: bind `for_employee_id` as a string and `uuid.Parse` it in the handler.

## 4. Pre-existing bug flagged (NOT fixed here — out of #10 scope)

The same gin limitation breaks the **existing** `GET /employees` list filters: `department_id`, `position_id`, `manager_id`, `role_id` are all `*uuid.UUID form:"…"` and **400** with `"[…] is not valid value for uuid.UUID"` whenever a value is supplied. This means audit decision #16 ("keep Go list filters") rests on filters that never worked with a value. Recommend a small follow-up: bind those as strings + parse (same fix pattern), or a custom gin binding. Tracked here for a separate PR.

## 5. Adversarial review + fixes

Ran a 4-lens adversarial review workflow (cycle-correctness / security / GORM-data / parity), each finding independently verified: **13 raised, 10 confirmed, 3 refuted** (the 3 refuted were correctly self-described as non-bugs: limit-bound is spec-sanctioned, clear-manager precedence is deterministic, manager-brief is_active is unreachable-nil).

**Fixed in this branch:**
- **Soft-delete scoping (4 findings, 1 med + 3 low):** all new preloads + search joins now use the `NotDeleted` scope — `Manager.User/Department/Position` (3 read paths), `FindByIDWithOrg`, `ListManagerCandidates`/`ListDirectReports` Department/Position, and the search `LEFT JOIN`s now require `is_deleted = false`. A soft-deleted org row can no longer leak into a manager brief / candidate / report row or drive a search match. Regression test: `TestLineManager_SoftDeletedManagerOrgNotLeaked`.
- **Cycle-check TOCTOU (medium):** `validateManagerAssignment` now runs **inside** the Update write-tx under a Postgres transaction advisory lock (`pg_advisory_xact_lock`, key `managerTreeLockKey`), re-reading committed state. Concurrent reparents serialize, so two requests can no longer each commit half a cycle. (Create needs no cycle check — a not-yet-created employee has no reports.) The exists/active TOCTOU (low) is narrowed to the same in-tx window.
- **Ordering (2 low):** `ListManagerCandidates` + `ListDirectReports` now `ORDER BY LOWER(full_name)` (matches the repo convention in department/skill/label/position repos); the kept-inactive legacy manager is re-sorted into its alphabetical slot after append.

**Documented as intentional / deferred (no code change):**
- **Search-before-limit (medium, parity):** Go applies the candidate `search` in SQL *before* `LIMIT`; Python applies it in memory *after* limiting a page. Go's behavior is the better picker UX (returns up to `limit` matching rows rather than search-pruning a truncated page) — kept intentionally; recorded here as a deliberate deviation.
- **Employee's own `department`/`position` still nil on `EmployeeRead` (low):** a pre-existing Phase-3 gap (`toRead` never populated them); #10 only surfaces it by contrast (the manager brief now shows org names). Out of #10 scope — follow-up to preload + map the employee's own org refs.
- **Pre-existing list-filter binding bug:** `GET /employees` uuid filters (`department_id`/`position_id`/`manager_id`/`role_id`) 400 on any value (same gin `*uuid.UUID` query-binding limitation). Tracked for a separate follow-up.

**Status: #10 verified end-to-end, reviewed, and review findings addressed. Pushed; PR open awaits owner OK.**
