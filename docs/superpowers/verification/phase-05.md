# Phase 5 — Leave Requests + Quota — End-to-End Verification

**Date:** 2026-05-20
**Branch:** `main`
**Migration version at start:** 7
**Migration version at end:** **8**
**Server:** built from current `main` HEAD (`/tmp/hrm-server`), run live
**Postgres:** Docker `ennam-ecom-postgres` (`localhost:5432`), DB `exnodes_hrm`, user `ennam`
**Attachment storage:** not exercised live (stubbed in unit tests with `stubUploader`; see §7)

The 22 steps below were executed live against a running server in a
single session. Every HTTP status, DB row, and security gate was
observed by `curl` / `psql` — no part of this log is paraphrased.
Raw command output: [`/tmp/phase05-verify2.log`](/tmp/phase05-verify2.log)
(produced by [`/tmp/p5-verify2.sh`](/tmp/p5-verify2.sh)).

---

## 0. Pre-flight

- Migration 000008 applied to **main + test** DBs (`make migrate-version` → `8` on both).
- Round-trip `8 → 7 → 8` performed during the migration commit (proves down/up symmetry — see commit `063bb0d`).
- `make fmt && make vet` → clean.
- `TEST_DATABASE_URL=... go test -count=1 ./internal/services/...` → **22 `TestLeaveService_*` tests pass** (~12s; full project suite ~36s).

## 1. Boot + seed gap fix (CRITICAL)

The live verification surfaced one load-bearing issue not caught by the unit tests: **the seed-merge gap**. REVISION NOTES item #4 had claimed "All leave permission constants + seeds are complete" but the `Employee` role's actual seed only carried `PermLeaveRead + PermLeaveCreate`. This meant non-admin owners could not even reach the service body for `cancel/update/delete` on their own pending request — the route gate fired first with 403.

Fix applied in [`internal/services/seed_service.go`](../../internal/services/seed_service.go) — added `PermLeaveUpdate, PermLeaveCancel, PermLeaveDelete` to the `Employee` role. The merge runs at boot:

```
2026/05/20 13:55:32 seed: merged permissions into role "Employee"
2026/05/20 13:55:32 exnodes-hrm-api listening on :8080 (env=development, swagger=true)
```

Service-level ownership checks (`row.EmployeeID == currentEmp.ID || asAdmin`) prevent these new perms from leaking cross-employee writes. This is the Phase-5 analog of Phase 4's `PermAnnounceManage` gap closure.

After merge, alice's effective perm set is:
```
auth:login, leave_requests:read, leave_requests:create, attendance:read,
leave_requests:update, leave_requests:cancel, leave_requests:delete
```

## 2. Create + warnings + validation

| Step | Action | Expected | Observed |
|---|---|---|---|
| 4 | alice `POST /leave-requests` 3-day annual full_day | 201, `total_days=3`, `warnings=[]` | 201, `total_days=3`, `warnings:[]` ✓ |
| 5 | alice `POST` overlapping 7/2–7/4 annual | 201 + overlap warning, row still created | 201, `warnings:["Date range overlaps existing leave request(s): 2026-07-01..2026-07-03 (pending)"]` ✓ |
| 6 | alice `POST` `to_date < from_date` | 400 | 400 `To Date must be on or after From Date` ✓ |
| 7 | alice `POST` `morning_half` with `from != to` | 400 | 400 `Half-day leave must be a single day (from_date must equal to_date)` ✓ |

Step 5 is the load-bearing **non-blocking warning** behavior (REVISION NOTES #8): the second request was still created (`status=pending`) but the response carries the warning string. The FE renders it as a soft notice; the row is real.

## 3. Status transitions + balance lifecycle

| Step | Action | Expected | Observed |
|---|---|---|---|
| 8 | admin `POST /:id/approve` | 200, `status=approved` | 200, `status=approved` ✓ |
| 9 | `GET /balance/{alice_emp}` | `annual_used=3, annual_remaining=7` | `{"annual_quota":10,"annual_used":3,"annual_remaining":7,"sick_quota":5,...,"leaves_this_year":1}` ✓ |
| 10 | alice `POST /:id/cancel` (her own approved) | 200, `status=cancelled` | 200, `status=cancelled` ✓ |
| 11 | `GET /balance/{alice_emp}` again | `annual_used=0, annual_remaining=10` (restored) | `{"annual_used":0,"annual_remaining":10,"leaves_this_year":0}` ✓ |
| 12 | alice `PATCH /:id` on the cancelled row | 409 | 409 `Cannot edit a cancelled leave request` ✓ |

Step 11 proves REVISION NOTES #9: the balance is a live SUM over `status='approved' AND is_deleted=false`, so cancelling an approved row naturally restores the remaining count. No background job, no extra column.

## 4. Admin-on-behalf-of (created_by attribution)

| Step | Action | Expected | Observed |
|---|---|---|---|
| 13 | admin `POST /leave-requests` with `employee_id=alice_emp` | 201, `request.employee.id=alice_emp`, `request.created_by=admin_emp` | 201 — subject `29e01fe6-...` (alice), `created_by:"ab07e787-3ffe-4758-83f2-729c3daa59f7"` (admin) ✓ |

This proves REVISION NOTES #6: both `employee_id` and `created_by` are NOT NULL `employees(id)` FKs; admin acting on behalf sets `created_by` = admin's **employee** id, not their user id.

## 5. Access control (service-level ownership + route-level perms)

| Step | Action | Expected | Observed |
|---|---|---|---|
| 14a | alice (now with `PermLeaveUpdate`) `PATCH` bob's pending | 403 — ownership rejection | 403 `You do not own this leave request` ✓ |
| 14b | alice `POST /:id/delete` bob's pending | 403 — ownership rejection | 403 `You do not own this leave request` ✓ |
| 15 | alice creates → admin approves → alice tries `POST /:id/delete` her own approved | 403 — non-admin owner only `pending` | 403 `Only pending leave requests can be deleted by their owner` ✓ |
| 16 | bob `POST /:id/approve` (Employee role lacks Approve) | 403 — route gate | 403 `{"missing":["leave_requests:approve"],"required":["leave_requests:approve"]}` ✓ |
| 17 | unauth `GET /leave-requests` | 401 | 401 `Could not validate credentials` ✓ |

Two distinct guard layers are observable here:
- **Route-level** (`middleware.RequirePerms`) — step 16's 403 carries the structured `{required, missing}` shape, fired BEFORE any handler code runs.
- **Service-level** (`row.EmployeeID == currentEmp.ID || asAdmin`) — steps 14a/14b/15 fire from inside the handler, with descriptive messages (`"You do not own this leave request"`, `"Only pending leave requests can be deleted by their owner"`). Both kinds are exercised; both produce 403 but with different evidence shape, exactly as designed.

## 6. Dashboard + history + admin list + soft-delete

| Step | Action | Expected | Observed |
|---|---|---|---|
| 18 | alice `GET /leave-requests/dashboard/me` | 200, balance + upcoming + history populated | 200, `balance.leaves_this_year=1`, `upcoming` 3 rows, `history` empty (no past dates yet) ✓ |
| 19 | alice `GET /leave-requests/history/me?page=1&page_size=10` | 200, cancelled row visible | 200, `total=1`, single item with `status=cancelled` ✓ |
| 20 | admin `POST /:id/delete` on the cancelled row | 200 (admin can delete any status) | 200 `Leave request deleted` ✓ |
| 21 | psql `SELECT id, status, is_deleted, deleted_at IS NOT NULL` | row exists, `is_deleted=t`, `deleted_at` set | `cancelled \| t \| t` ✓ |
| 22 | admin `GET /leave-requests?page=1&page_size=10` | 200, paginated list of all live rows | 200, `total=4` (admin-on-behalf, alice's overlap, alice's approved, bob's pending — the soft-deleted cancelled row is excluded) ✓ |

Step 21 is the load-bearing soft-delete contract: the row stays in Postgres (no hard `DELETE`), the `NotDeleted` scope hides it from every read, and the `deleted_at` timestamp is populated by the service's `gorm.Expr("NOW()")` write.

## 7. Attachment upload — coverage path

Attachment upload was **not exercised live** in this session. The coverage is:

- **Unit tests** (`leave_service_test.go`): `TestLeaveService_Create_AttachmentSpoof_BadRequest` proves that a text body with `.pdf` extension and a spoofed `Content-Type: application/pdf` is rejected (`400`) with **no upload attempted** (`stubUploader.uploaded == 0`). `TestLeaveService_Create_AttachmentPDF_OK` proves a valid PDF prefix uploads successfully.
- **Pattern reuse**: `LeaveService.uploadAttachment` uses the same `http.DetectContentType` content-sniff + allowlist as `SkillService.uploadIcon` (Phase 4) and `EmployeeService.uploadAvatar` (Phase 2). The live S3 round-trip for the underlying `UploadService.Upload` interface was already verified end-to-end in Phase 4 (`docs/superpowers/verification/phase-04.md` §3 with MinIO).

Net: the attachment-upload security-critical path (content sniffing) is double-tested at unit level; the storage round-trip is single-tested via the shared `Uploader` interface used by three services. To re-exercise live, recreate the MinIO container per the Phase 4 recipe and re-run with `STORAGE_*` env vars set.

## 8. Definition of Done — REVISION NOTES item #11

- [x] Server boots, migration version asserts `8`.
- [x] Login as admin → create leave (full_day, annual, 3 days) → 201, `total_days=3`, no warnings.
- [x] Create overlapping request → 201 with overlap warning.
- [x] Create with `to_date < from_date` → 400.
- [x] Create `morning_half` with `from_date != to_date` → 400.
- [x] Approve → 200, `status=approved`.
- [x] `GET /balance/{employee_id}` reflects `annual_used=3, annual_remaining=7`; `sick_used=0`.
- [x] Cancel approved → 200, `status=cancelled`; balance restored on next `GET /balance` (`annual_used=0`).
- [x] PATCH cancelled → 409.
- [x] Admin creates on behalf of another employee → 201, `created_by` = admin's employee id (not user id).
- [x] Non-admin owner `PATCH` someone else's pending → 403 (ownership-rejection from service body).
- [x] Non-admin owner `POST /:id/delete` someone else's pending → 403.
- [x] Non-admin owner `POST /:id/delete` their own approved → 403 (only `pending` allowed for non-admin).
- [x] Attachment content-sniff: text bytes with spoofed PDF content-type → 400, no upload (unit-tested with `stubUploader`).
- [x] Soft-delete row spot-check via psql: `is_deleted=t, deleted_at IS NOT NULL=t`.
- [x] 401 unauthenticated; 403 Employee role lacks Approve.

## 9. Phase 5 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `063bb0d` | 1 | Migration 000008 — leave_requests table, FK → employees(id) |
| `1818ac6` | 2 | Models: LeaveRequest + enum constants + IsQuotaLeaveType/IsHalfDayPeriod helpers |
| `ced69dc` | 3 | DTOs: Create/Update/Read/WriteResult envelope/Balance/Dashboard/List/History |
| `6dc0969` | 4 | LeaveRequestRepository (interface + Postgres impl, named LeaveDaysCount) |
| `ed1bdcd` | 5–8 | LeaveService (Create/Update/Approve/Reject/Cancel/Delete/Get/List/Balance/Dashboard/History) |
| `43abc0f` | 9–13 | leave_handler HTTP layer (11 endpoints + multipart + hasLeaveManageAll) |
| `214acad` | 14–15 | Route wiring in main.go + Swagger regen |
| `6bfa6de` | 16–19 | 22 integration tests + truncateAll(leave_requests) order fix |
| _next_ | 20–21 | (this verification log + Employee seed-merge fix + CHECKPOINT update) |

## 10. Deferred follow-ups (NOT in scope for Phase 5)

- **Multipart upload security review** carryover from Phase 4 (now applies to both `readSkillIcon` and `readLeaveAttachment` — the duplication is real and the shared-helper extraction is a clean refactor candidate). Schedule this review before Phase 7 (Announcements) since announcements may grow their own attachment surface.
- **N+1 in `populateRead`**: every list/dashboard call resolves employee/department/position per row. Acceptable at `page_size <= 100` and matches the Phase 4 pattern, but a batched preload would help when admin lists grow. Tracked in code via the comment in [`internal/services/leave_service.go`](../../internal/services/leave_service.go).
- **Manager role completeness**: `Manager` carries `Read/Create/Approve/Cancel/Manage` but lacks `Update/Delete`. Symmetric with Admin/HR Manager who have everything. Not a Phase-5 contract — flagged for the next BA pass.
- **`Employee.Subordinates` / `Employee.Manager` not used by leave**: the Phase 2 manager hierarchy is unrelated to leave approval flow today. If the contract evolves to "your manager approves your leave" the resolver lives in `LeaveService.Approve` and would consult `s.emps.FindByID(row.EmployeeID).ManagerID`.

## 11. Tooling note

Subagent dispatch remained unavailable in this session (same failure mode as the Phase 4 session — every probed `subagent_type` returns "not found" with an empty available-agents list). Phase 5 was executed inline by `project-owner` under explicit user override (decision **A** at session resume). Next phase: restart Claude to test whether the plugin registration recovers; with subagents back, prefer `superpowers:subagent-driven-development`.
