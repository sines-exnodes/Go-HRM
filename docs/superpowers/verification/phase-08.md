# Phase 8 — Organization Settings: End-to-End Verification Log

**Date:** 2026-05-22
**Branch:** `main`
**Migration version:** `11`
**Server:** `PORT=8082 go run ./cmd/server` (port 8080 still occupied by
`ennam-kg-server` container — same workaround as Phase 7)
**Base URL:** `http://localhost:8082/api/v1`

---

## Summary

16 end-to-end steps + DB-level singleton constraint check. All green.

Highlights:

- `system_config` singleton boots correctly via `seedSystemConfig()` —
  GET returns DB defaults (9:00 late, 18:00 checkout) right after boot.
- Pointer-based partial PATCH works: supplying only `late_threshold_hour`
  doesn't reset the other three fields.
- `company_address_updated_by` references `employees(id)` per REVISION
  NOTES #2; updated_by_name resolves to `Employee.FullName` ("Super Admin"
  in the seeded environment).
- Permission gating verified: non-admin user gets `403 forbidden` with
  `missing: ["organization_settings:manage"]` on every admin route.
  `GET /company-profile` is the only open read (200 for non-admins).
- Binding-tag bounds enforce DB invariants at the API edge:
  `late_threshold_hour=25` → 400, `company_latitude=200` → 400.
- DB-level `system_config_singleton` CHECK constraint blocks any second
  INSERT attempt at the Postgres layer.

---

## Endpoints exercised

| # | Endpoint | Auth | Status | Notes |
|---|---|---|---|---|
| 1 | `POST /auth/login` (super admin) | – | 200 | seeded |
| 2 | `GET /organization-settings/attendance` | admin | 200 | defaults: 9/0/18/0 |
| 3 | `PATCH /organization-settings/attendance` (hour=10) | admin | 200 | partial patch |
| 4 | `GET /organization-settings/attendance` | admin | 200 | shows new hour=10, others unchanged |
| 5 | `PATCH /organization-settings/attendance` (hour=25) | admin | **400** | binding-tag max=23 |
| 6 | `POST /employees` (non-admin) | admin | 201 | Reader for perm tests |
| 7 | `GET /organization-settings/attendance` (non-admin) | reader | **403** | `missing: ["organization_settings:manage"]` |
| 8 | `PATCH /organization-settings/attendance` (non-admin) | reader | **403** | same RequirePerms gate |
| 9 | `GET /organization-settings/company-profile` (non-admin) | reader | **200** | open read by design |
| 10 | `PATCH /organization-settings/company-profile` (non-admin) | reader | **403** | admin-only write |
| 11 | `PATCH /organization-settings/company-profile` (admin, full) | admin | 200 | stamps updated_at + updated_by |
| 12 | `GET /organization-settings/company-profile` (admin) | admin | 200 | `updated_by_name="Super Admin"` |
| 13 | `PATCH /organization-settings/company-profile` (lat=200) | admin | **400** | binding-tag max=90 |
| 14 | unauthenticated request | – | **401** | `code: unauthorized` |
| 15 | psql spot-check | – | – | one row, sentinel UUID, expected values |
| 16 | psql INSERT with different UUID | – | **CHECK** | `system_config_singleton` constraint rejects |

---

## DB-level singleton enforcement

```
INSERT INTO system_config (id) VALUES ('22222222-2222-2222-2222-222222222222');

ERROR:  new row for relation "system_config" violates check constraint "system_config_singleton"
DETAIL:  Failing row contains (22222222-2222-2222-2222-222222222222, 9, 0, 18, 0, null, null, null, null, null, ...).
```

The migration's `CHECK (id = '00000000-0000-0000-0000-000000000001')` is
the last line of defense against a duplicate row even if the service
layer were bypassed.

## DB spot-check

```
                  id                  | late_threshold_hour | late_threshold_minute | checkout_threshold_hour |   company_address   | company_latitude | company_longitude | has_updated_by
--------------------------------------+---------------------+-----------------------+-------------------------+---------------------+------------------+-------------------+----------------
 00000000-0000-0000-0000-000000000001 |                  10 |                     0 |                      18 | 123 HRM Lane, Hanoi |          21.0285 |          105.8542 | t
```

- Sentinel UUID matches `models.SystemConfigSingletonID`.
- `late_threshold_hour=10` after the step-3 PATCH; minute stayed 0
  (partial patch worked).
- `company_address_updated_by` is non-NULL — the FK reference resolved
  to the admin's employee row.

## Test summary

```
$ go test ./internal/services -run 'TestOrgSettings_' -count=1
ok  github.com/exnodes/hrm-api/internal/services   2.709s
```

8 service tests, all PASS. Full project test suite remains green:

```
$ go test ./... -count=1
ok  github.com/exnodes/hrm-api/internal/permissions  0.339s
ok  github.com/exnodes/hrm-api/internal/services    56.108s
ok  github.com/exnodes/hrm-api/internal/sse          1.214s
ok  github.com/exnodes/hrm-api/pkg/utils             1.374s
```

## Operational note: port 8080 conflict (carried over from Phase 7)

`ennam-kg-server` Docker container holds host port 8080 in this dev
environment. Phase 8 verification ran on `PORT=8082` again. CI default
stays 8080.

## What's deferred

- **Attendance service integration with `system_config`** — Phase 6's
  attendance code currently reads thresholds from `LATE_THRESHOLD_*`
  env vars. The Phase 6 plan flagged this as a Phase-8 follow-up: now
  that the row exists, refactor the attendance service to read it via
  the repo (with env vars as a fallback when the row is unreachable).
  Small change; deferred to a follow-up commit to keep Phase 8 scope
  tight.
- **Logo upload** — not in the Python source, not in the BA brief.
  Add when needed; column can land in a future migration without
  breaking the contract.
- **Per-resource fine-grained perms** — currently the four routes share
  `PermOrgSettings` (plus the open read). If BA splits attendance vs
  company-profile management, add new permission constants in a future
  pass.
