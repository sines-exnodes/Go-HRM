# Verification — Employees salary/banking access control (PR B)

**Date:** 2026-05-29
**Branch:** `feat/employees-salary-banking` (stacked on `feat/employees-parity` / PR A)
**Scope:** audit decision **#6** — full parity for salary/banking field-level access control.
**Migration:** none (permissions are in-code + seeded; no schema change).

## What changed

| Concern | Behavior |
|---|---|
| New perms | `users:salary_view`, `users:salary_manage`, `users:banking_view`, `users:banking_manage` added to the registry + picker catalog. |
| Seed | Admin + HR Manager receive all four (they manage payroll); the merge-seed adds them to existing roles on boot. Super Admin keeps `*`. |
| Read gating | A read returns the salary section only if the caller has `salary_view` (or `salary_manage`/`*`); the banking section only with `banking_view` (or `banking_manage`/`*`). Otherwise the section is `null`. |
| Account masking | On reads, `bank_account` is masked to `•••• <last4>` even for banking viewers. On write echoes (POST/PATCH responses) it is returned in full. |
| Write gating | Setting any salary field without `salary_manage` → **403**; any banking field without `banking_manage` → **403**. Enforced at the handler (the codebase's handler-level authorization pattern, like `RequirePerms`). |

**Parity-faithful nuance:** gating is purely permission-based (like Python's `populate_user_read(viewer=…)`), so a plain employee does **not** see their own salary/banking on `/employees/me` unless they hold the view perm. Confirmed intentional with the project owner.

## Design / layering

- `EmployeeFieldPerms` (service) + `ApplyEmployeeFieldVisibility(view, perms, unmask)` (read gate + mask) + `GuardSalaryWrite` / `GuardBankingWrite` (write gate) are **pure functions** — unit-tested without a DB.
- Service `Create`/`Update` signatures are unchanged from PR A; the write-gate is invoked by the handler before the service call (mirrors route-level `RequirePerms`). This keeps the ~20 existing service-test call sites untouched.
- Every employee-read-emitting handler (`List`, `Get`, `GetMe`, and the `Create`/`Update`/`SelfUpdate`/avatar/quota echoes) applies visibility: reads with `unmask=false`, write echoes with `unmask=true`.

## 1. Build / vet / tests — green

```
go build ./...   → clean
go vet ./...     → clean
go test ./internal/permissions/  → ok   (perm set/count assertions still pass with +4 perms)
go test ./internal/services/     → ok   (full integration suite, 100s)
```

7 new pure unit tests (`employee_field_perms_test.go`), all passing:
`StripsBothWhenNoPerms`, `SalaryViewKeepsSalaryStripsBanking`, `BankingViewMasksAccountOnRead`,
`UnmaskedOnWriteEcho`, `ManageImpliesViewButStillMasksOnRead`, `GuardSalaryWrite`, `GuardBankingWrite`.

## 2. End-to-end HTTP smoke (server on :8082 → test DB)

Super admin (`*`) and a custom `FieldViewer` role (SQL-seeded — this build has no role-create API) exercised the live paths:

| Step | Actor | Request | Result |
|---|---|---|---|
| Create w/ salary+bank | super admin (manage) | `POST /employees` | 201; echo `bank_account:"190233445566"` **unmasked**, salary present ✅ |
| Read | super admin | `GET /employees/{id}` | `bank_account:"•••• 5566"` **masked**, salary present ✅ |
| Read | `FieldViewer` (read, no salary/banking) | `GET /employees/{id}` | `full_name:"Payroll One"` visible; `basic_salary:null`, `bank_account:null`, `payment_method:null` — **stripped** (HTTP 200) ✅ |
| Write no salary | `FieldViewer` (+`employees:create`) | `POST /employees` (no salary) | **201** ✅ |
| Write w/ salary | `FieldViewer` | `POST /employees` `basic_salary:777` | **403** "You do not have permission to set salary fields" ✅ |
| Write w/ bank | `FieldViewer` | `POST /employees` `bank_account:"123"` | **403** "You do not have permission to set banking fields" ✅ |

(The write-gate 403 is distinct from the route-level `employees:create` 403 "Insufficient permissions" — verified by granting `employees:create` and observing my guard's message fire.)

**Status: PR B verified end-to-end (build + unit tests + full integration suite + live HTTP gating/masking/write-gate). Ready to commit + push; PR open awaits user OK.**
