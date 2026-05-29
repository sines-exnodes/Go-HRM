# Verification — Employees Python-shape parity (PR A)

**Date:** 2026-05-29
**Branch:** `feat/employees-parity` (off `main` @ `bcb4c0b`, which includes merged PR #6)
**Migration:** `000017_employee_parity` (latest on disk → applied version **17**)
**Scope:** audit decisions **#4, #5, #7, #8, #12, #13, #17, #17b** (the "Python-shape parity" group).
PR B (salary/banking access control, #6) and the deferred features (#10, #11, #15) are tracked separately.

## What changed

| # | Decision | Change |
|---|---|---|
| #4 | Emergency contacts → list | New `employee_emergency_contacts` child table; dropped the 3 flat `emergency_contact_*` columns; replace-set on create/update/self-update; `emergency_contacts[]` on read. |
| #5 | Leave quota in read | `EmployeeRead.annual_leave_quota` / `sick_leave_quota` hydrated from `employee_leave_quotas` (defaults 12/6 when no row). |
| #7 | Self-edit widen | `/employees/me` now also accepts `full_name`, `gender`, `dob` (+ self-managed emergency contacts). Salary/dept/position/manager remain admin-only (whitelist intact). |
| #8 | Missing read fields | Added `employees.experience_year` + `employees.cv_url` columns; embedded `skills[]` (from `employee_skills`) on the read. |
| #12 | Self/destructive guards | Can't deactivate self (`PATCH /employees/{id}`, `PATCH /users/{id}`); can't delete self (`DELETE /employees/{id}`); deleting a manager NULLs `manager_id` on direct reports. (No reauth on employee delete — per locked decision.) |
| #13 | Admin change-email | New `POST /users/{id}/change-email` (`users:update`); stamps `email_changed_at`. |
| #17 | marital_status enum | `single/married/other` (DTO validators on create/update/self). |
| #17b | education enum | `high_school/college/bachelor/master/doctorate` (DTO validators). |

## 1. Build / format / vet — green

```
go build ./...   → clean
make fmt         → gofmt -s -w .   (no diff after)
make vet         → clean
make swag        → regenerated docs/swagger (AdminChangeEmailRequest + new DTO shapes present)
```

## 2. Migration 000017 up + down round-trip — green

`TestMain` drops + re-applies all migrations from scratch on every test run, so the full suite passing (below) already proves `000017` up applies cleanly on a virgin DB. Down/up round-trip verified explicitly via the golang-migrate library:

```
start version=17 dirty=false
after down version=16
after up version=17
emergency_contacts table present: true
experience_year+cv_url columns present: true
ROUND-TRIP OK
```

Down-migration carries the data-loss guard (refuses to revert if any employee has >1 emergency contact), mirroring migration 000016.

## 3. Integration tests (real Postgres) — all green

`TEST_DATABASE_URL=…/exnodes_hrm_test`, `go test ./... -count=1`:

```
ok  github.com/exnodes/hrm-api/internal/middleware
ok  github.com/exnodes/hrm-api/internal/permissions
ok  github.com/exnodes/hrm-api/internal/services      (89s — full integration suite)
ok  github.com/exnodes/hrm-api/internal/sse
ok  github.com/exnodes/hrm-api/pkg/utils
```

12 new behavior tests added (`internal/services/employee_parity_test.go`), all passing — each encodes *why* the behavior matters (AGENTS Rule 9):

- `Create_WithEmergencyContacts`, `Update_ReplaceEmergencyContacts` (replace / leave-on-nil / clear-on-`[]`)
- `Read_HydratesLeaveQuota` (defaults 12/6 + explicit), `Read_EmbedsSkills`, `Read_ExposesExperienceAndCV`
- `SelfUpdate_AllowsIdentityFields` (name/gender/dob change **but salary stays put**), `SelfUpdate_ManagesOwnEmergencyContacts`
- `SoftDelete_RejectsSelf`, `SoftDelete_ClearsSubordinateManager`, `Update_RejectsSelfDeactivate`
- `UserParity_AdminPatch_RejectsSelfDeactivate`, `UserParity_AdminChangeEmail` (no-op / conflict / happy + `email_changed_at` stamp)

> Bug caught by the real-DB test (not by compile): the subordinate `manager_id` clear initially used `Update("manager_id", nil)`, which GORM dropped as a zero value. Switched to the codebase's proven `Updates(map{"manager_id": nil})` pattern; test then passed.

## 4. End-to-end HTTP smoke (server on :8082 → test DB)

Booted the built binary against `exnodes_hrm_test` (migrated to 17). `GET /health` → `{"status":"ok"}`. Logged in as the seeded super admin.

| Step | Request | Result |
|---|---|---|
| `/employees/me` shape | `GET /employees/me` | `full_name:"Local Admin"`, `emergency_contacts:[]`, `skills:[]`, `annual_leave_quota:12`, `sick_leave_quota:6` ✅ |
| Create + contacts | `POST /employees` w/ 2 contacts + `experience_year:7` + `cv_url` | 201; `emergency_contacts` has 2 (with ids), `experience_year:7`, `cv_url` echoed, quota defaults 12/6 ✅ |
| Replace contacts | `PATCH /employees/{id}` `emergency_contacts:[1]` | list replaced to 1 ✅ |
| Clear contacts | `PATCH /employees/{id}` `emergency_contacts:[]` | count 0 ✅ |
| **Admin change-email** | `POST /users/{id}/change-email` | 200; email updated ✅ |
| Self-guard delete | `DELETE /employees/{ownEmp}` | **400** "You cannot delete your own account" ✅ |
| Self-guard deactivate | `PATCH /users/{ownUid}` `is_active:false` | **400** "You cannot deactivate your own account" ✅ |
| Enum validation | `PATCH /employees/{id}` `marital_status:"divorced"` | **400** (rejected per #17) ✅ |

## Notes / parity nuances

- Emergency-contact `relationship` is free text (Python doesn't constrain it); only the dependents relationship enum stays constrained.
- `skills[]` + quota + emergency contacts are preloaded in `List` too (GORM batches the has-many preloads — no N+1).
- The single-contact → list migration carries existing data forward (only rows with a non-empty name).

**Status: PR A verified end-to-end (build + migration round-trip + integration suite + live HTTP). Ready to commit + push; PR open awaits user OK.**
