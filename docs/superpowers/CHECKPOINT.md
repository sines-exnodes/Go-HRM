# Resume Checkpoint — Monthly Workdays API MERGED to main (PR #18, no migration)

**Last updated:** 2026-06-22
**Stopped at:** Monthly Workdays API **merged to `main` via PR #18** (merge commit `9a1b511`). 1 endpoint live at `GET /api/v1/workdays?year=<year>`. Post-merge revision **DR-009-004-01 v1.1 (AC-04)**: holidays are informational only — **Workdays = TotalDays − Weekends** (was `−holidays`). `main` is in sync with `origin/main`. Next: Request Tickets module (EP-003/US-003) or Attendance G7 (holiday "H" cells, now unblocked by the holidays calendar).
**Branch:** `feat/monthly-workdays` — **merged to `main` (PR #18); no longer in flight.**

> **2026-06-22 session (docs/infra only — no migration, no business logic change):**
> - **Scalar replaces Stoplight Elements** at `/docs` (`internal/handlers/router.go`, constant `scalarDocsHTML`). Stoplight Elements `@8.0.1` injected literal `\n` bytes into the Try It request body causing `"invalid character '\n' in string literal"` 400s. Scalar reads the same `/swagger/doc.json` spec; BearerAuth is auto-wired via `data-configuration`. Token entered once per session, persists across all endpoints.
> - **Serena memories updated**: `resume_protocol` step 0 now reads `CLAUDE.md` + `AGENTS.md` first; `core` has a new "Authoritative behavior files" section at the top of the memory graph.
**DB migration version:** **23** (applied). Next free: **000024**.
**See:** [Post-migration parity work](#post-migration-parity-work-python--go-api-parity) and [User Contracts](#user-contracts-module--done-migration-000022) below.

> **Roles & Permissions parity — DONE (PR #14, on `main`).** Full role CRUD at `/api/v1/roles` (was catalog/seed-only); role `level` (1–100) + assignment authority (`user_service.AssignRoles` → 403 if granting above your max level); soft-delete frees the name (partial-unique on `LOWER(name)`); `is_system` guards; `GET /roles/permissions` gated `roles:read`. Brief role embed renamed `dto.RoleRead`→`dto.RoleRef`. Seed levels: Super Admin 100 / Admin 90 / HR Manager 80 / Manager 50 / Employee 10. Migrations 000020 + 000021.

> **User Contracts — DONE (migration 000022, on `main`).** Full contract CRUD at `/api/v1/users/:id/contracts` — 6 endpoints (list, get, create, update, delete, upload-attachment). `user_contracts` table with `trg_user_contracts_set_updated_at`. `PermUsersContractsView` + `PermUsersContractsManage`; view seeded to all 5 roles, manage to Admin + HR Manager. 13 integration tests green. Swagger regenerated (14 contract refs in swagger.json). Commits: 9d6bef2, 5a8780e, e9f6987, 65c0eaf, 52dafe1, c0378ea, 6f76a29, f09fde5.

## Monthly Workdays API — DONE (no migration, `feat/monthly-workdays`)

Computed endpoint that returns workday counts for each month of a given year. No new DB table — reads from existing `holidays` table (migration 000023). All computation is pure Go. **Formula (DR-009-004-01 v1.1 AC-04, revised post-merge): `workdays = total_days − weekends`** — holidays are still counted and returned per month (range-clamped, cross-month split) but are **informational only**, NOT subtracted from workdays. (The original v1.0 formula `− holidays` was changed in commit `01f1ac6`.)

- **No migration** — reads from existing `holidays` table.
- **DTOs** (`internal/dto/workday.go`): `WorkdayQuery`, `MonthWorkdaysRead`, `WorkdayTotalRead`, `WorkdayYearRead`.
- **Repository extension** (`internal/repositories/holiday_repo.go`): `FindByYear(ctx, year int)` added to `HolidayRepository` interface + implementation.
- **Service** (`internal/services/workday_service.go`): `WorkdayService.GetYear` — iterates 12 months, counts weekends via `time.Weekday()`, clamps holiday ranges per month (cross-month split), sums total row.
- **Permissions**: `PermOrgWorkdaysView = "organization:workdays_view"`. Group `organization_workdays` "Monthly Workdays". Seeded to all 5 roles (Super Admin via `*`, Admin, HR Manager, Manager, Employee).
- **Handler** (`internal/handlers/workday_handler.go`): `WorkdayHandler.GetYear` — `ShouldBindQuery` + `AppError` propagation + `dto.Response` envelope.
- **Route**: `authed.GET("/workdays", middleware.RequirePerms(authSvc, permissions.PermOrgWorkdaysView), workdayH.GetYear)` in `cmd/server/main.go`.
- **Tests** (`internal/services/workday_service_test.go`): 7 integration tests — no holidays, holiday on weekday, holiday on weekend (AC-05), cross-month holiday (AC-06), leap year, non-leap year, total row.
- **Swagger**: regenerated after handler completion.
- **Key commits**: `3256b6e` (DTOs), `445408d` (FindByYear), `ee6a219` (perms+seed), `9bc7dad` (service+7 tests), `e344f18` (handler+route+swag), `0463790` (verification log), `01f1ac6` (**v1.1 fix — holidays informational only**), `9a1b511` (**merge PR #18 → `main`**), `55ecf6f` (v1.1 DR doc), `809b2ef` (GitNexus reindex).

**Verified (at merge):** `go build ./...` clean, 7 tests skip cleanly without DB, Docker container rebuilt, curl smoke: 12 months correct shape + 400/401 error cases. ⚠️ **The committed verification log [`docs/superpowers/verification/phase-workdays.md`](verification/phase-workdays.md) is STALE** — it still shows the pre-v1.1 formula (`workdays = total_days − weekends − holidays`) and example numbers computed with `−holidays` (e.g. Feb workdays=13 instead of the v1.1 value 28−8=20). The merged code is v1.1 (`− weekends` only). Re-run the live curl and refresh the log before treating it as proof.

---

## Holiday Management module — DONE (migration 000023)

New feature module implementing company holiday calendar with leave-recalculation integration.

- **Migration 000023** (`migrations/000023_holidays.up/down.sql`): `holidays` table (UUID PK, `year`, `name`, `from_date`, `to_date`, audit cols + `trg_holidays_set_updated_at`) + `holiday_templates` table (Vietnamese public holiday presets, seeded via migration).
- **Models** (`internal/models/holiday.go`): `Holiday` + `HolidayTemplate`.
- **DTOs** (`internal/dto/holiday.go`): `HolidayCreate`, `HolidayUpdate`, `HolidayListQuery`, `HolidayRead`, `HolidayTemplateRead`, `HolidayImportRequest`, `HolidayImportResult`.
- **Permissions**: `PermOrgHolidaysView` (`org:holidays_view`) + `PermOrgHolidaysManage` (`org:holidays_manage`). View seeded to all 5 roles; Manage to Admin + HR Manager.
- **Repository** (`internal/repositories/holiday_repo.go`): `HolidayRepository` interface + GORM impl — List (paginated/year-scoped), Get, Create, Update, Delete (soft), FindInRange, YearsWithHolidays, ExistsByNameAndYear. `HolidayTemplateRepository` with ListByYear.
- **Service** (`internal/services/holiday_service.go`): Create, Update, Delete, List, GetYears, ListTemplates, Import. `recalculateAffectedLeaves` recalculates `total_days` on all Approved leave requests overlapping a mutated holiday range via `leaveRepo.FindApprovedOverlapping` + `BulkUpdateTotalDays`. Leave-aware from day 1.
- **Handler** (`internal/handlers/holiday_handler.go`): 7 endpoints — `GET /holidays`, `POST /holidays`, `PATCH /holidays/:id`, `DELETE /holidays/:id`, `GET /holidays/years`, `GET /holidays/templates`, `POST /holidays/import`.
- **Routes**: wired in `cmd/server/main.go` under `/api/v1/holidays`.
- **Tests** (`internal/services/holiday_service_test.go`): integration tests covering Create (duplicate 409), Update (name change, date change, not-found), Delete (recalc, not-found), List (pagination, year scoping), GetYears (inject current), ListTemplates, Import (skip duplicates). All PASS.
- **Swagger**: regenerated after handler completion.
- **Bug fixed during verification**: typed-nil `*AppError` in `error` interface in `Update` and `Delete` handlers. `parseIDParam` returns `(uuid.UUID, error)` (interface). Reassigning the same `aerr` variable from a service call returning `*apperrors.AppError` boxed a nil `*AppError` into a non-nil `error` interface, causing a panic in error middleware when accessing `ae.HTTP` on a nil `*AppError`. Fix: rename `parseIDParam` result to `err` (matching all other handlers), so the service `:=` freshly declares `aerr` as `*apperrors.AppError`. Verified: PATCH 200 + DELETE 200 + error paths 404.
- **Key commits**: `04b4167`, `dc22fb9`, `bcd2dcf`, `1ec11cb`, `174233a` (holiday repo, service, injection, permissions, handler+routes).

**Verified:** `go build ./...` clean, `go vet ./...` clean, migration 000023 applied, Swagger regenerated, all 7 endpoints curl-verified end-to-end, bug found + fixed + re-verified. Verification log: [`docs/superpowers/verification/phase-holidays.md`](verification/phase-holidays.md).

## User Contracts module — DONE (migration 000022)

New feature module (not a Python-parity item) — adds full contract lifecycle to user profiles.

- **Migration 000022** (`migrations/000022_user_contracts.up/down.sql`): `user_contracts` table — UUID PK, `user_id → users(id)`, `contract_type` (FULL_TIME/PART_TIME/FREELANCE/INTERNSHIP/PROBATION), `start_date/end_date`, `attachment_url`, audit cols + `trg_user_contracts_set_updated_at`.
- **Model** (`internal/models/user_contract.go`): `UserContract` + `ContractType` enum.
- **DTOs** (`internal/dto/user_contract.go`): `UserContractCreate`, `UserContractUpdate`, `UserContractListQuery`, `UserContractRead`, `UserContractAttachmentResponse`.
- **Repository** (`internal/repositories/user_contract_repo.go`): interface + GORM impl — List (paginated), Get, Create, Update, Delete (soft).
- **Service** (`internal/services/user_contract_service.go`): Create, Get, Delete, List, Update, UploadAttachment.
- **Permissions**: `PermUsersContractsView` (`users:contracts_view`) + `PermUsersContractsManage` (`users:contracts_manage`). View seeded to all 5 roles; Manage to Admin + HR Manager only.
- **Handler** (`internal/handlers/user_contract_handler.go`): 6 endpoints — `GET /users/:id/contracts`, `POST /users/:id/contracts`, `GET /users/:id/contracts/:contract_id`, `PATCH /users/:id/contracts/:contract_id`, `DELETE /users/:id/contracts/:contract_id`, `POST /users/:id/contracts/:contract_id/attachment`.
- **Routes**: wired in `cmd/server/main.go` at `/api/v1/users/:id/contracts`.
- **Tests** (`internal/services/user_contract_service_test.go`): 13 integration tests — all PASS (`make test` 0 fail / 0 skip).
- **Swagger**: regenerated — 14 contract refs in `docs/swagger/swagger.json`.
- **Commits**: 9d6bef2 (migration), 5a8780e (model+DTO), e9f6987 (repo), 65c0eaf (service), 52dafe1 (permissions+seed), c0378ea (handler+routes), 6f76a29 (tests), f09fde5 (fmt+swag).

**Verified:** `go build ./...` clean, `go vet ./...` clean, `make test` 0 fail / 0 skip, `make swag` regenerated (14 contracts refs). Dev DB at migration 21 — **`make migrate-up` required** to apply 000022.

## Announcements parity — Plans A+B (committed to `main`; push pending)

Python↔Go announcements parity audit → 3 gaps locked → Plans A (trivial) + B (dispatch), executed subagent-driven. Full suite green. **No migration.**

- **Audit:** [`specs/2026-06-09-announcements-parity-audit.md`](specs/2026-06-09-announcements-parity-audit.md) — 8 gaps, 3 actionable (G1/G2/G3), 5 intentional divergences.
- **Plans:** [`plans/2026-06-09-announcements-parity-plan-a.md`](plans/2026-06-09-announcements-parity-plan-a.md) · [`plans/2026-06-09-announcements-parity-plan-b.md`](plans/2026-06-09-announcements-parity-plan-b.md).
- **Plan A — G3 + G1** (`a988f29` + `bfbaf35` + `23daf5c`):
  - G3: 409 guard in `AnnouncementService.Update` for published/archived rows — matches Python's ConflictException. Fires before ownership check.
  - G1: `POST /api/v1/mobile/announcements/:id/read` route alias → existing `MarkViewed` handler. Zero new logic.
  - Swagger `@Failure 409` added to Update handler.
- **Plan B — G2: push + email dispatch on publish** (`97c4af8` + `f525dd8` + `8952277` + `6601f51` + `3f7fd4c` + `222b979`):
  - `EmployeeRepository`: `FindAllActive` / `FindByIDs` / `FindByDepartmentIDs` — used to resolve `target_audience` → `[]uuid.UUID` user IDs.
  - `EmailService.SendAnnouncementNotification` — gomail pattern, 10s timeout. `sendMessage` + `fromAddress` private helpers extracted to deduplicate vs `SendInvite`.
  - `AnnouncementNotifier` interface (nil-safe) + `broadcastPublished` now launches `go dispatchNotifications` after SSE. `resolveRecipientUserIDs` handles all/department/custom audience modes.
  - Concrete `announcementNotifier` in `internal/services/announcement_notifier.go` — calls `push.SendToUser` + `email.SendAnnouncementNotification` per user; per-user errors logged, loop continues.
  - Wired in `main.go`: `annNotifier := services.NewAnnouncementNotifier(pushSvc, emailSvc, userRepo)`.
- **Verification:** `go test ./...` all PASS, `go fmt ./...` + `go vet ./...` clean. HEAD is `222b979`.
- **Deploy pending:** push + `docker restart exnodes-hrm-app`.

## Leave Requests parity — Plan A + B (committed locally; push + PR pending)

Python↔Go leave-request audit → locked decisions D1–D7 → two plans, executed subagent-driven (fresh implementer per task + two-stage spec/quality review per plan). Full suite green (`make test`, all PASS or SKIP). **No migration** (leave was migration 000008, already live).

- **Audit + decisions:** [`specs/2026-06-08-leave-requests-parity-audit.md`](specs/2026-06-08-leave-requests-parity-audit.md) — 11-endpoint inventory, Gaps G1–G10, Decisions D1–D7 all LOCKED 2026-06-08.
- **Plans:** [`plans/2026-06-08-leave-requests-parity-plan-a.md`](plans/2026-06-08-leave-requests-parity-plan-a.md) (permission split + seed), [`plans/2026-06-08-leave-requests-parity-plan-b.md`](plans/2026-06-08-leave-requests-parity-plan-b.md) (bug fixes + export).
- **Plan A (`0132dd8`):** `leave_requests:approve_team` + `leave_requests:approve_all` permission constants; `ApproveScope` type + `checkCanApproveOrReject` BFS enforcement in service (mirrors Python `check_can_approve_or_reject`, reuses `emps.SubordinateIDs`); `resolveApproveScope` handler helper; perm gate moved out of router middleware into handler; seed: Admin/HR Manager → `approve_all`, Manager → `approve_team + update + delete`. Also: `PermLeaveApprove` removed from `AllPermissions()` (kept as constant for backward-compat runtime code path) to satisfy `TestPermissionGroupsContainsAll`.
- **Plan B (`d632a69`):** G3 empty-PATCH no-op guard (no more Approved→Pending revert); G4 DOCX MIME detection via ZIP-sniff + `.docx` extension fallback; G5 attachment limit 10 MB → 5 MB; G6 dashboard limit 5 → 10; G7 `Cancel` returns `(read, wasApproved, error)` + handler embeds `was_approved` in response; G10 `GET /api/v1/leave-requests/export` → xlsx via excelize.
- **Decisions delivered:** D1=A (full split + legacy compat) · D2=B (keep Go MIME types + DOCX + 5 MB) · D3/D5/D7 = recommended (fix dashboard, cancel wasApproved, Manager seed) · D6=A (Excel export).
- **Deferred:** G9 leave-quota carry-forward endpoints (not in Python source per audit).
- **FE docs — DONE:** `docs/superpowers/handoff-2026-06-08-leave-requests-fe-api-changes.md` (verbose handoff) + web repo `exnodes-hrm-web-nextjs/api_info_go/leave_requests.md` (concise developer reference, same style as `attendance.md`).
- **Pushed to `origin/main`** — `a5b6a69` is the HEAD on remote.

## Attendance parity — Plan A + B (MERGED to `main`, PR #16 `0a42ff1`)

Python↔Go attendance audit → locked decisions D1–D6 → two plans, executed subagent-driven (fresh implementer per task + two-stage spec/quality review). Full suite green (`go test ./...`, services ~181s, 0 fail / 0 skip). **No migration** (reads existing `leave_requests` + uses the existing `is_auto_checkout` column).

- **Audit + plans:** [`specs/2026-06-05-attendance-parity-audit.md`](specs/2026-06-05-attendance-parity-audit.md), [`plans/2026-06-05-attendance-parity-a-matrix-leave.md`](plans/2026-06-05-attendance-parity-a-matrix-leave.md), [`plans/2026-06-05-attendance-parity-b-export-jobs.md`](plans/2026-06-05-attendance-parity-b-export-jobs.md).
- **Plan A — leave-integrated matrix** (`cc60537`): approved leave rendered in matrix cells (G1, AC-016); combined half-day cells with `worked_half_status` + AM/PM thresholds 09:00/13:15/12:00/18:00 (G2, AC-026–031); leave-aware SR-011 Total Late/Early summaries (G6); `on_leave` status filter + combined multi-match (G4); **root `GET /api/v1/attendance` now returns the matrix** (D1, `43fdda6`) — flat list moved to `/attendance/records`, `/matrix` kept as alias.
- **Plan B** — Excel export `GET /attendance/export` + `/export/{employee_id}` via `xuri/excelize/v2`, reusing the leave-aware `buildAllRows` so totals match the matrix (G3, AC-011/012/025); check-in/out now return `TodayStatusRead` (D2); `is_half_day` hours-based auto-flip removed — half-day is leave-driven (D5); `AutoCheckOut(cutoff)` service + `POST /attendance/auto-checkout` admin trigger gated `attendance:manage_data` (G5).
- **Decisions:** D1 root=matrix · D2 TodayStatusRead · D3 excelize · D4 service+admin-endpoint now · D5 follow-BA (drop flip) · D6 keep admin CRUD.
- **Deferred / follow-ups:** **G7 holiday "H" cells + streak-excludes-holidays — BLOCKED** (no holiday-calendar source; BA open question). Real **23:00 scheduler** trigger for `AutoCheckOut` (admin endpoint lands the logic now). **BA back-fill DR** for the Go-only admin-CRUD surface (kept per D6, currently unspecced).
- **FE doc — DONE:** `worktree`/repo handoff `docs/superpowers/handoff-2026-06-08-attendance-fe-api-changes.md` + web repo `exnodes-hrm-web-nextjs/api_info_go/attendance.md` (created, untracked there). Leads with the breaking changes (root=matrix, check-in/out→TodayStatusRead, `is_half_day` no longer hours-driven).
- **Deployed to dev (`:8080`) — DONE:** the app container runs **Air hot-reload, bind-mounted from the `E:\Work\Go-HRM` checkout** (now on `main`). After merge, the main checkout was switched to `main` and the container restarted; live smoke verified root matrix (leave cells + combined half-day + `worked_half_status`), `?status=on_leave`, `GET /export` (valid xlsx), `/records`, boot at migration 21. (A WIP stash from `fix/skill-icon-cleanup-2mb-cap` was set aside on that branch — `git stash list`.)
- **Demo data — SEEDED** (dev DB `exnodes_hrm`): April + May 2026 across **6 employees** (added Mai/Long/Huong/Tuan), covering on_time/late/absent/weekend/multi-session, all 5 full-day leave types, and half-day-off with all 3 `worked_half_status` variants (on_time/late/absent) for both morning & afternoon halves; June 1–5 light fill for the default view. Idempotent script: `seed-attendance-demo.sql` (currently in the redundant worktree; not yet moved to `scripts/`).

## How to resume next session

### ✅ DONE: Monthly Workdays merged to `main`

Merged via **PR #18** (`9a1b511`); `main` in sync with `origin/main`. No further action — `feat/monthly-workdays` is no longer in flight. (Follow-up: the stale verification log noted above.)

### IMMEDIATE next & subsequent priorities (in descending value):

1. **Request Tickets module (EP-003/US-003) — NEW, biggest remaining gap.** Entirely unbuilt: no `request_tickets:*` perms, no model/migration/repo/service/handler/routes. FE matrix P11/P12/P13 have no backing. Full vertical slice: ticket model + migration (**000024**) + CRUD + status transitions (In Progress/On Hold/Resume/Resolve) + row-level own-records scoping + submitter-exclusive Close/Reopen. Read EP-003 DRs in `ba-requirements/` first.
2. **Attendance G7 — holidays now available.** Holiday calendar (migration 000023) unlocks the blocked "H" cell type in the attendance matrix + streak-excludes-holidays. Can now proceed.
3. **Announcement view-permission tier.** FE matrix wants P23 (Announcement View, read-only) + P24 (Management); Go has only `announcements:manage`. Decide with BA.
4. **Attendance follow-ups**: real **23:00 scheduler** for `AutoCheckOut`; switch thresholds to `system_config` lookup; move `seed-attendance-demo.sql` into `scripts/`.
5. **Bundled code review** — Phases 4-9 have not been formally reviewed.
6. **Phase 7 attachment-upload HTTP handler** — model + repo in place; route is the missing piece.
7. **Production env wiring** — `FIREBASE_CREDENTIALS_PATH` + real SMTP host.

Latest taken migration = **000023** (holidays); next is **000024**.

### Resume entry points

1. **`docs/superpowers/CHECKPOINT.md`** (this file).
2. **`.serena/memories/project_overview.md`** — code-map + boot protocol.
3. **`CLAUDE.md`** — auto-loaded into every Claude session.

## Post-migration parity work (Python ↔ Go API parity)

The migration is done; ongoing work reconciles Go's API shape with the Python
source, audited module-by-module: audit → locked decisions → PR(s) →
verification log → FE doc → handoff for deferred items.

### Announcements — DONE (merged)
PR #5 (`body`→`description`, `send_now`, brief mobile widget) + PR #6 (hybrid
per-user targeting `target_audience:"custom"` + `recipient_ids[]`, CORS fix).
Migrations 000013–000016, both on `main`. 13-decision audit.

### Employees — DONE (merged)
19-decision audit (Python `users` module ↔ Go `users`⟂`employees`⟂`dependents`).
All three layers merged to `main`:

- **A `feat/employees-parity`** (PR #7) — emergency-contact list (new
  `employee_emergency_contacts` table), leave-quota/skills/cv on read, widened
  self-edit (name/gender/dob), self & destructive guards, admin change-email,
  marital/education enums. **Migration 000017.**
- **B `feat/employees-salary-banking`** (PR #9) — salary/banking field-level
  perms (`users:{salary,banking}_{view,manage}`) + account masking + write-gate
  (#6). No migration.
- **C `feat/employees-line-manager`** (PR #10) — line-manager suite: assignment
  validation (self / cycle via subordinate-BFS / inactive, **advisory-locked
  in-tx re-check**), `GET /employees/manager-candidates`, `GET /employees/{id}/
  direct-reports`, rich `manager` brief. No migration.

Verified: [`verification/employees-parity-pr-a.md`](verification/employees-parity-pr-a.md),
[`-pr-b.md`](verification/employees-parity-pr-b.md),
[`employees-line-manager.md`](verification/employees-line-manager.md) — build/vet,
full integration suite, migration up/down round-trip, live HTTP smoke
(`scripts/smoke-employees-parity.sh`, 34/34), + a 4-lens adversarial review on C
(10 confirmed findings fixed).

**Still deferred / follow-ups** → [`handoff-2026-05-29-employee-parity.md`](handoff-2026-05-29-employee-parity.md):
- **#11** cv/id-card upload endpoints (URLs accepted for now).
- **#15** role-level assignment authority (N/A — Go RBAC has no role `level`).
- **Found during C — both RESOLVED in parity round 2** (`feat/employees-parity-2`):
  (a) the `GET /employees` uuid list-filters now bind as repeated `[]string` →
  `uuid.Parse` → SQL `IN` (also made multi-select per BA DR-001-005-01); (b) the
  employee's **own** `department`/`position` names are now resolved on `EmployeeRead`.

FE: web repo PR **#5** (`feat/go-employees-parity` → main) carries the full FE
wiring + `api_info_go/employee.md` + `me.md`. The web repo is self-managed.

### Employees — ROUND 2 — DONE (merged to `main`, PR #12 / `de83970`)

Parity round 2, 7 commits, grounded in a fresh 5-dimension Python parity audit.
Verified: [`verification/employees-parity-2.md`](verification/employees-parity-2.md)
— build/vet, full integration suite **220 tests / 0 skip / 0 fail**, migration
000018 + 000019 up/down round-trips, **live HTTP smoke 38/38**, DB spot-check;
two-stage subagent review per task.

- **Multi-select list filters** (`5ae3b6d`) — dept/position/role/manager repeated
  params → `IN`; fixes the 400-on-any-value bug. **BA over Python** here (Python
  single-values dept/position; BA DR-005-01 wants all multi-select).
- **experience_year as a career-start year** (`ddf0d94`) — validate `>1900 &&
  ≤ current year`; **migration 000019** normalizes legacy counts → years.
- **Inline `skill_ids`** on Create/Update (`56f53ef`) — Python parity; standalone
  `PUT /employees/:id/skills` kept.
- **Own department/position resolved** on read (`b5bd158`) — Phase-3 gap closed.
- **Name split** (`c52d66a`) — drop `full_name`, add `first_name`/`last_name`
  (**migration 000018**); `Employee.FullName()` method composes the display name
  for briefs. Confirmed by the Python audit (Python stores first/last separately).
- **Swagger regen + smoke update** (`5c986d8`).
- **Direct reports**: kept the standalone endpoint (Python parity — NOT embedded).

Latest taken migration **000019**; next is **000020**. Still deferred (unchanged):
#11 cv/id-card upload, #15 role levels (N/A); avatar/CV/ID-image upload endpoints;
name-split FE wiring (web repo self-managed).

**Done:** merged to `main` via PR #12 (`de83970`) on 2026-06-04 — `main` is now at migration 19. No parity branch remains in flight.

### Roles & Permissions — DONE (implemented + verified on `feat/roles-permissions-parity`, not yet merged)

Parity audit: [`specs/2026-06-04-roles-permissions-parity-audit.md`](specs/2026-06-04-roles-permissions-parity-audit.md).
Plan: [`plans/2026-06-04-roles-permissions-parity.md`](plans/2026-06-04-roles-permissions-parity.md).
Verification: [`verification/roles-permissions-parity.md`](verification/roles-permissions-parity.md).
**Closed the headline gap:** Go had *no role-management API* (catalog only). Now has
full CRUD (list/get/create/update/delete) + a role-`level` authority hierarchy.

Implemented subagent-driven (fresh implementer per task + two-stage review).
Verified: build/vet, **full repo test suite 0 fail / 0 skip** (services 172s),
**live HTTP smoke 16/16** (incl. level-authority 403 over HTTP + soft-delete name
reuse + catalog gate-403), DB spot-check. **Migrations 000020 (role `level`) +
000021 (`uq_roles_name_active ON roles(LOWER(name)) WHERE is_deleted=FALSE`).**
Latest taken migration **000021**; next free **000022**.

**Locked decisions — all delivered:**
- **D1 role `level` (1–100) + assignment-authority** — done. **Reopened-and-RESOLVED
  deferred #15.** Seed levels: Super Admin 100 / Admin 90 / HR Manager 80 /
  Manager 50 / Employee 10. `user_service.AssignRoles` enforces "assigner may only
  grant roles ≤ their own max level" (ported from Python `check_role_assignment_authority`).
- **D2 soft delete, name freed** — partial-unique on `LOWER(name)` allows reuse.
- **D3 sort by `level` ASC** (then name).
- Also delivered: gate `GET /roles/permissions` behind `roles:read`; one list endpoint
  with full `permissions[]` + `permission_count`; is_system rename/level/delete guards;
  role-name regex + perm-string validation. Registry delta (`approve_team`/`approve_all`)
  left to the leave-requests pass (out of scope, as audited).

Also fixed: test harness migration source made cross-platform (iofs) so the suite
runs on Windows (commit `6d58ad4`).

**Status:** final whole-branch review APPROVED; **PR [#14](https://github.com/sines-exnodes/Go-HRM/pull/14) open** (`feat/roles-permissions-parity` → `main`) — awaiting merge.
The brief role embed in user/employee responses was renamed `dto.RoleRead`→`dto.RoleRef`
(JSON wire format unchanged) to free `RoleRead` for the full role API shape.
**Dev env:** `exnodes_hrm` migrated to **000021**; `:8080` app container now runs the branch via the
dev compose override (Air hot-reload, source-mounted) — `docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d app`.
FE fix (web repo): `role.api.ts` list/create dropped the trailing slash (`/api/v1/roles/`→`/api/v1/roles`) that caused a CORS-less 301.

## Phase Summary (final)

| # | Module | Migration | Verification | Commits |
|---|---|---|---|---|
| 0 | Foundation | 000001 | [phase-00.md](verification/phase-00.md) | — |
| 1 | Auth + RBAC | 000002 | [phase-01.md](verification/phase-01.md) | — |
| 2 | Users + Employees + Dependents | 000003 + 000004 | [phase-02.md](verification/phase-02.md) | — |
| 3 | Departments + Positions | 000005 | [phase-03.md](verification/phase-03.md) | — |
| 4 | Skills + Labels | 000006 + 000007 | [phase-04.md](verification/phase-04.md) | — |
| 5 | Leave Requests + Quota | 000008 | [phase-05.md](verification/phase-05.md) | ~10 |
| 6 | Attendance + Matrix | 000009 | [phase-06.md](verification/phase-06.md) | 14 |
| 7 | Announcements + Mobile + SSE | 000010 | [phase-07.md](verification/phase-07.md) | 15 |
| 8 | Organization Settings | 000011 | [phase-08.md](verification/phase-08.md) | 8 |
| 9 | Email + Invite + Push | 000012 | [phase-09.md](verification/phase-09.md) | 12 |

### Phase 9 — DONE ✅ (2026-05-22)

Email + Invite + Push notifications. 7 new endpoints (5 admin invite +
1 public accept + 1 push test), invite table (migration **000012**),
embedded HTML/plain email templates, SMTP via gomail, FCM HTTP v1 push
client with no-op fallback. Live verification:
[docs/superpowers/verification/phase-09.md](verification/phase-09.md)
(23 e2e steps + DB spot-check, all green, real email delivery
verified via Mailpit). Highlights:

- **Migration 000012** — `invites` table with `invited_by →
  employees(id) RESTRICT` (schema split) + `accepted_user_id → users(id)
  SET NULL` (auth-level marker populated on accept). Two partial unique
  indexes: `(token) WHERE is_deleted=FALSE` and `(email) WHERE
  accepted_at IS NULL AND is_deleted=FALSE`. `role_ids UUID[]` column
  for role-on-accept assignment (custom `UUIDArray` scanner/valuer).
- **Diverges from Python source** intentionally: invites are first-class
  rows; user creation is deferred to `/invites/accept`. Avoids partially-
  provisioned users sitting in the `users` table.
- **`POST /invites/accept` is genuinely public** — registered outside
  the JWT-protected group. Token in body is the credential. Atomically
  creates user + employee via `empSvc.Create` + assigns roles via
  `users.ReplaceRoles` + stamps `accepted_at` + `accepted_user_id`.
  Verified live (user logs in immediately after accept).
- **SMTP graceful degradation** — when `SMTP_HOST=""`, the EmailService
  returns `ErrEmailDisabled` and the InviteService records the message
  on `invites.last_email_error`. Invite creation never returns 500 on
  SMTP misconfig (REVISION NOTES #11). Verified via service test.
  Mailpit pipeline proves the real-delivery path end-to-end.
- **FCM HTTP v1 push client** with JWT-based service-account auth
  (`google.JWTConfigFromJSON`). When `FIREBASE_CREDENTIALS_PATH=""` or
  the file is unreadable, returns a no-op logger client that satisfies
  the interface but reports `IsConfigured()=false`. Boot never fails
  on FCM misconfig. `/notifications/test` returns the diagnostic
  `{sent, skipped, errors}` envelope.
- **`PermInviteManage` constant added** to registry + seeded to Admin
  and HR Manager via merge-seed (the seed service appends missing
  perms to existing system roles on every boot — idempotent).
- **Token format**: 32 random bytes → URL-safe base64 (no padding) →
  43 chars. Verified at the Mailpit boundary.
- **Composite signal — accept flow validation**: replay (409),
  unknown token (404), short password (400), revoked token (404),
  duplicate-pending (409), existing-user-email (409). All verified
  live in steps 12-17.

#### Phase 9 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `4a388f6` | – | Plan REVISION NOTES at the top |
| `2b520f8` | T1+T2 | Dependencies (gomail/oauth2) + env keys |
| `3bf9f51` | T3 | Migration 000012 — invites table |
| `8e70793` | T4+T5 | Invite model + UUIDArray scanner + DTOs |
| `0153a53` | T6 | InviteRepository |
| `7923375` | T7+T8 | EmailService + embedded HTML/text templates |
| `7d9dc35` | T9+T10 | InviteService (CRUD + Accept) |
| `0ede002` | T11+T12 | PushClient interface + FCM impl + PushNotificationService |
| `ae0a871` | – | PermInviteManage + seed |
| `2214f14` | T13+T14 | InviteHandler + NotificationHandler |
| `f470410` | T15 | Wire in main.go (public /accept + 5 admin /invites + /notifications/test) |
| `6ccd8e9` | T16+T17 | 17 service tests + truncateAll extension + Swagger regen |
| `14e872e` | T19-T22 | E2E verification log (Mailpit pipeline + DB spot-check) |
| _this_ | T18 | README + CHECKPOINT close (migration complete) |

## TOOLING NOTE

~~Subagent dispatch (`Agent` with `subagent_type`) is structurally unavailable in
the VSCode-extension SDK runtime.~~ **STALE — corrected 2026-06-04.** Subagent
dispatch now works in this runtime (verified with a read-only probe). The roles &
permissions parity work is being executed subagent-driven (fresh subagent per task
+ review between). Phases 0–10 + earlier parity were done inline by the
project-owner (commit-per-task); that remains a valid fallback.

## Code review status

Phases 0–3: review applied, fixes committed.

Phases 4–9: **review not yet requested.** Recommendation — one final bundled review covering:

- **Multipart upload pattern** (avatar P2, skill icon P4, leave attachment P5; will add announcement attachments + logo at follow-up). The `http.DetectContentType` + MIME allowlist sniff is now at four sites — extract into a shared helper.
- **Two-layer access control** (`RequirePerms` + `asAdmin bool` ownership branch — Phases 5/6/7/8/9). Pattern works; document it in `.serena/memories`.
- **Composite-PK reactivation pattern** (`AnnouncementLabel` + `AnnouncementTargetDepartment` + `EmployeeSkill`). Three sites share this; extract a small generic helper.
- **Singleton repo pattern** (`SystemConfigRepository`) — 4-method interface, DB-level CHECK as the last line of defense.
- **SSE hub design** (single-process, drop-on-full-buffer) — fine for a single replica; document the scaling boundary clearly.
- **Public endpoints surface** — `/auth/login`, `/auth/refresh`, `/invites/accept`. Worth a security pass before production exposure (rate-limiting in particular).

## Local environment notes

- **Postgres**: Docker at `localhost:5432`, user `postgres` / pass `devpassword` (verified working 2026-06-01; the earlier `ennam/ennam_dev_2026` note was stale). Main DB `exnodes_hrm`, test DB `exnodes_hrm_test`. The integration suite auto-migrates the test DB via `TestMain` (golang-migrate library); `migrate`/`psql`/`docker` CLIs are NOT on PATH here, so migration round-trips were verified via the library. `swag` lives in `GOPATH/bin` (Makefile references it by full path).
- **Port 8080 conflict**: `ennam-kg-server` container holds host port 8080 in this dev environment. Phases 7-9 live verifications ran on `PORT=8082`. CI default stays 8080.
- **Mailpit for SMTP verification**: `docker run -d --rm -p 11025:1025 -p 18025:8025 --name mailpit-phase09 axllent/mailpit`. UI at `http://localhost:18025`.
- **FCM disabled in dev** — `FIREBASE_CREDENTIALS_PATH=""`. PushClient is the no-op logger. Production rollout: set the env var to a service-account JSON + `FIREBASE_PROJECT_ID`.
- `.env` is git-ignored. `.env.example` has every key the project reads.
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** every cross-aggregate FK from Phase 2 onward targets `employees(id)`, NOT `users(id)`. Exceptions: `users.id` is the FK target for auth-level surfaces (`device_tokens`, `user_notification_settings`, `announcement_views`, `invites.accepted_user_id`).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at` + `BEFORE UPDATE` trigger. Singletons keep the columns for schema parity even when soft-delete is meaningless.
- **Singleton tables:** PK is a fixed sentinel UUID + `CHECK (id = '…')` constraint at the DB level. Repo exposes only `Get / EnsureExists / UpdateFields(map)`.
- **Composite-PK join models:** declare an explicit Go model; replace-set logic uses snapshot-diff-reactivate.
- **Repo joins MUST qualify `is_deleted`** — `models.NotDeleted` becomes ambiguous after a JOIN to a table that carries `is_deleted`.
- **SSE broadcast is a "refresh hint"** — FE refetches via GET on receipt; visibility is enforced on the read path.
- **Idempotent markers** — `Clauses(clause.OnConflict{DoNothing:true})` preserves first-occurrence semantics.
- **Partial PATCH semantics:** pointer-typed DTO fields → only write when non-nil. Empty PATCH = no-op success. Stamp `updated_by/at` only when "real" content fields change.
- **Public endpoints with token-as-credential** (Phase 9 `/invites/accept`) live OUTSIDE the JWT group. Token validated in the service, not the middleware.
- **External integrations degrade gracefully** — empty SMTP_HOST / FIREBASE_CREDENTIALS_PATH disable the respective integration without crashing the boot. Record errors on the relevant row (e.g. `invites.last_email_error`) rather than rolling back the request.

## Outstanding micro-items

- Untracked (intentional): `.claude/`, `AGENTS.md`, `CLAUDE.md`.
- Phase 5/6/7/8/9 plan files have unticked `- [ ]` checkboxes in draft task bodies — superseded by REVISION NOTES blocks; not worth churn commits.
- Phase 5 manager-role completeness gap — **FIXED in leave-requests parity Plan A** (Manager now has `approve_team + update + delete`).
- Phase 6 attendance service still reads thresholds from env vars — should switch to `system_config` lookup now that the row exists.
- Phase 7 attachment-upload HTTP handler is deferred.
- Phase 7 `target_audience='custom'` deferred (no backing table).
- Phase 7 scheduled-publish cron deferred.
- Phase 8 logo upload deferred.
- Phase 9 password-reset email flow deferred (needs BA confirmation).
- Phase 9 `/invites/accept` rate-limiting (consider IP-based at the reverse proxy).
- Phase 9 FCM topic-based fanout deferred (per-device only at present).
