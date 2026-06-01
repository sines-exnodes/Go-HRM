# Verification — Employees API Parity Round 2

**Date:** 2026-06-01
**Branch:** `feat/employees-parity-2`
**Spec:** [specs/2026-05-15 … ](../specs/2026-06-01-employee-api-parity-2-design.md) · **Plan:** [plans/2026-06-01-employee-api-parity-2.md](../plans/2026-06-01-employee-api-parity-2.md)
**DB migration version:** **19** (000018 name split, 000019 experience-year normalize)

## What shipped (7 commits)

| Commit | Task | Summary |
|---|---|---|
| `5ae3b6d` | 1 | Multi-select list filters — `department_id`/`position_id`/`manager_id`/`role_id` repeated query params → `uuid.Parse` → SQL `IN` (fixes the 400-on-any-value bug); `is_active` stays a single bool. |
| `ddf0d94` | 2 | `experience_year` validated as a career-start year (`>1900 && ≤ current year`) in Create + Update. |
| `56f53ef` | 3 | Inline `skill_ids` on Create/Update (Python parity) via a narrow `skillAssigner` interface backed by `*SkillService`; standalone `PUT /employees/:id/skills` retained. |
| `b5bd158` | 4 | Resolve the employee's OWN department/position names on the read (preload + `RefRead`); closes the Phase-3 null gap. |
| `c52d66a` | 5 | **Name split** — drop `employees.full_name`, add `first_name`/`last_name` (migration **000018**); `models.Employee.FullName()` method composes the display name for briefs; Create/Update/SelfUpdate/Read/Summary DTOs carry first/last; repo search/order use the split columns; seed + invite-accept split adapted. |
| `68ddb37` | 6 | Migration **000019** — normalize any `experience_year` stored as a count (`<1900`) to `currentYear − count`; reversible (approximate) down with a `RAISE NOTICE`. |
| `5c986d8` | 7a | Swagger regenerated (`make swag`) + smoke script updated for first/last + experience-as-year + new skill_ids/multi-filter blocks. |

Decisions grounded in a 5-dimension Python parity audit (`/Users/panda/Documents/Work/exnodes-hrm-api`): names, experience, and skills MATCH the Python source; **filters follow the BA (DR-001-005-01)** over Python (Python only multi-selects `role_ids`; the BA wants dept/position/role all multi-select); **direct reports stay a standalone endpoint** (Python parity — no embed).

## Verification environment

- Local Postgres (docker) at `localhost:5432`, user `postgres` / pass `devpassword`. Test DB `exnodes_hrm_test`.
- Integration suite auto-migrates the test DB via `TestMain` (golang-migrate library: Drop + Up from `migrations/`).
- `migrate` / `psql` / `docker` CLIs are NOT on PATH in this environment; `swag` is (in `GOPATH/bin`, referenced by the Makefile). Migration up/down round-trips were therefore exercised via the golang-migrate **library** (throwaway `go run` harnesses) rather than the CLI.

## Evidence

### 1. Build / vet / format
```
go build ./...   → exit 0
go vet ./...      → clean
gofmt -l internal/ cmd/  → (empty)
```

### 2. Full integration suite (real Postgres)
```
TEST_DATABASE_URL=…/exnodes_hrm_test go test ./...
ok  github.com/exnodes/hrm-api/internal/services   123.188s
ok  github.com/exnodes/hrm-api/internal/middleware  …
ok  github.com/exnodes/hrm-api/internal/permissions …
ok  github.com/exnodes/hrm-api/internal/sse         …
ok  github.com/exnodes/hrm-api/pkg/utils            …
```
`-v` run: **220 tests RUN, 0 SKIP, 0 FAIL.** (The `internal/services` real duration — not `(cached)` — confirms tests genuinely ran against the DB, not skipped via `skipIfNoDB`.)

New behavior covered by added tests: multi-select dept filter (`IN`) + single-value filter (the previously-broken path); experience-year rejection (future / `≤1900`) on Create AND Update + acceptance; inline `skill_ids` on create (2 skills) + update (replace→1, clear→0) + invalid-skill-on-create returns 400 with no orphan user; department/position names resolved on read; name-split round-trip across all fixtures.

### 3. Migration 000018 (name split) — up/down round-trip
golang-migrate library against the test DB:
```
at latest version=18
down one ok -> version=17 (000018 down ran)
re-up ok -> version=18 dirty=false
MIGRATION_ROUNDTRIP_OK
```

### 4. Migration 000019 (experience count→year) — data transform + round-trip
Seeded an employee with `experience_year = 7` at version 18, then:
```
at version=18 (pre-transform)
UP transform ok: 7 -> 2019 (= 2026 - 7)
DOWN reverse ok: 2019 -> 7
MIGRATION_000019_DATA_ROUNDTRIP_OK
```

### 5. Live HTTP smoke (`scripts/smoke-employees-parity.sh`)
Builds the server, boots it on `:8082` against the test DB (storage/SMTP/FCM degrade gracefully), seeds the admin, exercises every audited behavior with real HTTP requests:
```
==================== SMOKE SUMMARY ====================
  PASS: 38    FAIL: 0
=======================================================
```
Includes the round-2 additions:
```
== parity-2: inline skill_ids on create + multi-select department filter ==
  ✓ #parity2 create skill (201)
  ✓ #parity2 create employee with skill_ids (201)
  ✓ #parity2 skill echoed on create (length=1)
  ✓ #parity2 multi-select department_id filter returns 200
```
plus `#7 self can edit first_name`, `#8 experience_year echoed as a year`, and all prior parity assertions (emergency contacts, leave quota, salary/banking strip+mask+write-gate, change-email, self-guards, line-manager suite).

### 6. DB schema spot-check (test DB, post-smoke)
```
employees name/exp columns: experience_year first_name last_name   (no full_name)
  row: first="Smoke" last="3164732683" experience_year=2018
```

### 7. Swagger
`make swag` regenerated `docs/swagger/`. `dto.EmployeeCreate` / `EmployeeUpdate` / `EmployeeSelfUpdate` / `EmployeeSummary` now carry `first_name`/`last_name` (no `full_name`) + `skill_ids`; list params show the multi-select `[]string` filters. Remaining `full_name` definitions are the legitimately-unchanged `Dependent*`, `Invite*`, `EmergencyContactInput`, and the manager/direct-report briefs.

## Review

Every task passed a two-stage subagent review (spec compliance → code quality). Notable review-driven fixes folded into the task commits: a single-value filter test (Task 1); experience-year `Update` validation moved before the self-deactivate guard + an `Update` rejection test (Task 2); an orphan-user assertion + `callerUserID` rename + post-commit-failure-mode comment (Task 3); the invite-accept single-token name now yields an empty last_name instead of a doubled surname, + a stale `LOWER(full_name)` comment fixed (Task 5).

## Known follow-ups / notes

- **Name backfill heuristic** (split on first space) is best-effort for multi-word legacy names; real data was the seed admin + smoke fixtures, so impact is nil.
- **000019 down** is arithmetically approximate across a year boundary (uses the apply-date year) — documented in the migration + spec.
- **`EmployeeRead` has no Swagger definition** — employee handler `@Success` annotations use `map[string]interface{}`, so swag never expands it. Pre-existing doc gap, not introduced here.
- **Single-token invite names** become `first_name=token, last_name=""` (valid: column is `NOT NULL DEFAULT ''` and the accept path is an internal `empSvc.Create`, where the `min=1` last-name *binding* tag does not fire).
- Deferred (unchanged from before): avatar/CV/ID-image upload endpoints (URLs still accepted); name-split FE wiring (web repo self-managed).
