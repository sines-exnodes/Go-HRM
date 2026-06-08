# Leave Requests — Python ↔ Go API Parity Audit

**Date:** 2026-06-08
**Module:** Leave Requests (EP-002 US-001 — web list/create/edit/status + mobile dashboard + quota)
**Status:** AUDIT COMPLETE — decisions **D1–D7 LOCKED 2026-06-08**. Implementation plan in `plans/`.
**Method:** Same pipeline as attendance/employees/roles:
audit → locked decisions → plans → subagent-driven implementation → PR → FE doc → deploy → seed

**Sources read**
- Python: `app/routers/leave_requests.py`, `app/services/leave_request.py`,
  `app/schemas/leave_request.py`, `app/models/leave_request.py`, `app/core/permissions.py`
  (at `E:\Work\exnodes-hrm-api`)
- Go: `internal/handlers/leave_handler.go`, `internal/services/leave_service.go`,
  `internal/repositories/leave_request_repo.go`, `internal/repositories/leave_quota_repo.go`,
  `internal/models/leave_request.go`, `internal/dto/leave.go`,
  `internal/permissions/registry.go`, `cmd/server/main.go` (route wiring),
  `migrations/000008_create_leave_requests.up.sql`
- BA intent (authoritative):
  - `ba-requirements/.../WEB-APP/EP-002-leave-management/US-001-leave-requests/details/` (DR-002-001-01 through -06)
  - `ba-requirements/.../MOBILE-APP/EP-002-leave-management/US-001-leave-requests/details/DR-002-001-01-leave-requests-dashboard.md`

---

## 0. TL;DR

Go's leave-requests module is **at near-endpoint parity with Python** — all 11 routes present,
matching path/method/permission patterns, identical `LeaveRequestRead` wire format, identical
write DTOs. This is the best-shape module entering parity: no missing endpoints, no missing
response fields, no missing business logic at the headline level.

The gaps are narrower than attendance:

- **REGISTRY GAP (permission split)** — Go has only the legacy `leave_requests:approve`
  (approve anyone). Python has `approve_team` (subordinate chain only) + `approve_all`
  (anyone). Go service has no scope check. **D1** (the biggest contested point;
  explicitly deferred to this pass from the roles audit).

- **Go BUG — empty PATCH reverts Approved→Pending** — Update service has no no-op guard.
  An empty-body `PATCH /:id` on an Approved request silently reverts it to Pending.
  Python exits early. **D4** (not contested — a bug fix).

- **Minor Go gaps**: dashboard limit hardcoded at 5 (BA wants 10), cancel response missing
  "balance restored" message, DOCX not recognised as valid attachment MIME (Go standard
  library limitation), attachment max 10 MB vs BA-specified 5 MB.

- **Manager seed gap** — Manager role missing `leave_requests:update` and
  `leave_requests:delete`. CHECKPOINT follow-up item. **D7**.

- **Neither has list export** — BA AC-21/22 asks for it; format is an open BA question.
  **G10 / D6**.

No migration required for G3–G8 fixes. D1 (approve_team/approve_all) requires a registry
seed update (new perm strings added to existing roles — an idempotent boot-seed addition,
not a new migration).

---

## 1. Endpoint inventory

| # | Python (`/leave-requests`) | Go (`/api/v1/leave-requests`) | Middleware gate | Gap |
|---|---|---|---|---|
| 1 | `GET /` → paginated list | `GET ""` → paginated list | `leave_requests:read` | ✅ parity |
| 2 | `GET /balance/{id}` | `GET /balance/:employee_id` | `leave_requests:read` | ✅ parity |
| 3 | `GET /dashboard/me` | `GET /dashboard/me` | JWT only | ✅ parity |
| 4 | `GET /history/me` | `GET /history/me` | JWT only | ✅ parity |
| 5 | `POST /` | `POST ""` | `leave_requests:create` | ✅ parity |
| 6 | `GET /{id}` | `GET /:id` | `leave_requests:read` | ✅ parity |
| 7 | `PATCH /{id}` | `PATCH /:id` | `leave_requests:update` | ✅ parity |
| 8 | `POST /{id}/approve` | `POST /:id/approve` | Py: JWT; Go: `leave_requests:approve` | **scope gap — D1** |
| 9 | `POST /{id}/reject` | `POST /:id/reject` | same as approve | **scope gap — D1** |
| 10 | `POST /{id}/cancel` | `POST /:id/cancel` | `leave_requests:cancel` | ✅ parity |
| 11 | `POST /{id}/delete` | `POST /:id/delete` | `leave_requests:delete` | ✅ parity |

**Approve/Reject permission architecture diverges:**
- Python: route uses `Depends(get_current_user)` (no perm at route level);
  service `check_can_approve_or_reject(admin, lr)` checks `*`, `approve_all`,
  legacy `approve`, or `approve_team`+subordinate-chain BFS.
- Go: route uses `RequirePerms(PermLeaveApprove)` (anyone with `leave_requests:approve` passes);
  service `Approve/Reject` calls `transitionStatus` with no caller identity — **no scope check**.

`leave_requests:manage` (used by `hasLeaveManageAll` in both Create/Update/Get/Delete/Cancel)
behaves identically in Python (`_can_manage`) and Go — both check for the manage perm or `*`.

---

## 2. Data model comparison

| Field | Python (`LeaveRequest` Beanie) | Go (`models.LeaveRequest`) | Notes |
|---|---|---|---|
| `id` | Mongo `ObjectId` | UUID `gen_random_uuid()` | different DB; wire format both `string` |
| `employee_id` | `PydanticObjectId` → users coll | UUID → `employees(id)` | **Go schema split** (expected) |
| `created_by` | `PydanticObjectId` → users | UUID → `employees(id)` | Go split: creator = employee, not user |
| `from_date` / `to_date` | Python `date` | Postgres `DATE` (via `time.Time`) | parity on wire (date-only string) |
| `leave_period` | `LeavePeriod` StrEnum | `text CHECK (3 values)` | enum values identical |
| `leave_type` | `LeaveType` StrEnum | `text CHECK (5 values)` | enum values identical |
| `total_days` | `float` | `numeric(5,1)` | parity |
| `reason` | `str` | `text NOT NULL` | parity |
| `attachment_url` | `str \| None` | `*string` | parity |
| `status` | `LeaveStatus` StrEnum | `text CHECK (4 values)` | enum values identical |
| soft-delete | `is_deleted bool` only | full 4 audit cols + trigger | Go convention |

**Read shape** (`LeaveRequestRead` / `dto.LeaveRequestRead`) — **wire format identical**:
```json
{
  "id", "employee": {"id","name"}, "department": {"id","name"}, "position": {"id","name"},
  "from_date", "to_date", "leave_period", "leave_type", "total_days", "reason",
  "attachment_url", "status", "created_by", "created_at", "updated_at"
}
```

**Write DTOs** — **identical** (Create: employee_id?, from_date, to_date, leave_period,
leave_type, reason; Update: all pointer types). Both sides handle multipart+JSON transparently.

**Quota model structural difference** (transparent to the API):
- Python: `annual_leave_quota` / `sick_leave_quota` fields directly on the `User` document.
- Go: separate `employee_leave_quotas` table (`EmployeeLeaveQuota` — migration 000004).
- Both expose the same `LeaveBalanceSummary` wire shape; the structural difference is invisible
  to callers.

**Balance computation — both implicit:** Neither Python nor Go has an explicit debit/credit
ledger. Balance = `Quota − SUM(total_days WHERE status='approved' AND year=N)`. When a request
is approved, it enters the sum; when cancelled/deleted, it leaves. This is parity.

---

## 3. Business-rule comparison (vs BA acceptance criteria)

| Rule | BA backing | Python | Go today | Verdict |
|---|---|---|---|---|
| New requests start Pending | DR-02-02 SR-07 | ✅ | ✅ | parity |
| Admin can create on behalf | DR-02-02 SR-03/04 | ✅ (`can_manage`) | ✅ (`hasLeaveManageAll`) | parity |
| To Date ≥ From Date | DR-02-02 SR-09 | ✅ | ✅ | parity |
| Half-day: must be a single day | DR-02-02 SR-10 (implied) | ❌ not enforced | ✅ Go validates `from==to` | Go stricter; BA silent on this |
| Past dates allowed | DR-02-02 SR-08 | ✅ | ✅ | parity |
| Total days = inclusive × period (0.5 for half) | DR-02-02 SR-10 | ✅ | ✅ | parity |
| Insufficient balance: warning only, non-blocking | DR-02-02 SR-11 | ✅ | ✅ | parity |
| Overlapping dates: warning only, non-blocking | DR-02-02 SR-12 | ✅ | ✅ | parity |
| Overlap checks Pending+Approved only | DR-02-02 SR-18 | ✅ | ✅ | parity |
| Overlap self-excludes on update | DR-02-03 AC-27 | ✅ | ✅ | parity |
| Approved → Pending revert on edit | DR-02-03 SR-04 | ✅ | ✅ | parity |
| **No-op on empty PATCH body** | **DR-02-03 SR-06** | **✅ early exit** | **❌ always saves; reverts Approved→Pending** | **G3 — Go bug** |
| Edit locked for Rejected / Cancelled | DR-02-03 SR-03 | ✅ (400) | ✅ (409) | parity (409 more correct) |
| Status transitions: P→A, P→R, P→C, A→C only | DR-02-05 SR-04 | ✅ | ✅ | parity |
| Balance implicitly restored on Cancel-from-Approved | DR-02-05 SR-07 | ✅ (implicit) | ✅ (same implicit) | parity |
| **"Balance restored" message when cancelling Approved** | **DR-02-05 AC-28** | **✅ `was_approved` flag** | **❌ no flag, no message variant** | **G7 — Go gap** |
| Employee delete: own Pending only | DR-02-04 SR-03 | ✅ | ✅ | parity |
| Admin delete: any status | DR-02-04 SR-04 | ✅ | ✅ | parity |
| Balance restored on Approved delete | DR-02-04 SR-07 | ✅ (implicit) | ✅ (implicit) | parity |
| **Approve scope: approve_team (subordinate chain)** | **Python SR-27; not in BA DRs** | **✅ service-enforced** | **❌ absent** | **G1 — Go gap** |
| **Approve scope: approve_all vs legacy approve** | **Python registry split** | **✅** | **❌ single legacy perm** | **G2 — Go gap** |
| Quota: annual + sick only | DR-02-06 SR-02 | ✅ | ✅ | parity |
| Quota: immediate effect (no approval workflow) | DR-02-06 SR-03 | ✅ | ✅ | parity |
| Attachment: PDF, PNG, JPG accepted | BA DR-02-02 SR-14 | ✅ | ✅ | parity |
| **Attachment: DOCX accepted** | **BA DR-02-02 SR-14** | **✅ (route says PDF/PNG/JPG/DOCX)** | **❌ DOCX not in `allowedAttachmentMIME`** | **G4 — Go gap** |
| **Attachment max 5 MB** | **BA DR-02-02 SR-14** | **✅ (presumed 5 MB)** | **❌ Go allows 10 MB** | **G5 — Go vs BA** |
| **Dashboard limit: default 10 items/tab** | **Mobile BA SR-07** | **✅ `?limit` (default 10, max 50)** | **❌ hardcoded 5** | **G6 — Go gap** |
| Multi-select dept/position filter on list | BA DR-02-01 §3 Filters | ❌ single only | ❌ single only | **G9 — neither** |
| Leave list export | BA DR-02-01 AC-21/22 | ❌ not implemented | ❌ not implemented | **G10 — neither (BA open question)** |
| **Manager role: update + delete perms in seed** | CHECKPOINT follow-up | — | ❌ seed gap | **G8 — Go seed gap** |

---

## 4. Gaps to close (priority order)

### G1 + G2 — Approval permission split + scope enforcement (largest; blocks D1)

Go has only `PermLeaveApprove = "leave_requests:approve"` (pre-split legacy).
Python implemented the `.team`/`.all` split post-roles-parity:
- `leave_requests:approve_team` — approver may only approve employees in their transitive
  subordinate chain (BFS via `line_manager_id`)
- `leave_requests:approve_all` — may approve any request
- Legacy `leave_requests:approve` treated as `approve_all` at runtime for backward compat

Go service `Approve/Reject` calls `transitionStatus` with no identity check — any bearer of
`leave_requests:approve` can approve anyone's request.

**Changes required (subject to D1 locked):**
1. Add `PermLeaveApproveTeam = "leave_requests:approve_team"` + `PermLeaveApproveAll = "leave_requests:approve_all"` to `registry.go`; keep `PermLeaveApprove = "leave_requests:approve"` as legacy (still valid perm, treated as all).
2. Add approve_team + approve_all entries to `LeavePermGroup` in the catalog. Remove `approve` from catalog items (so new roles can't assign it via the UI; existing grants still work).
3. Change middleware gate on approve/reject from `PermLeaveApprove` → accept any of {`PermLeaveApprove`, `PermLeaveApproveAll`, `PermLeaveApproveTeam`} — or gate on a sentinel that means "any approve perm". (Simplest: add an `PermLeaveApproveAny` OR use `RequireAnyPerm`.)
4. Service: add `approverUserID uuid.UUID` param to `Approve` and `Reject`.
5. Implement `checkCanApproveOrReject(ctx, approverUserID, lr)` in the service — mirrors Python's function; needs `emps.FindByUserID` + employee subordinate-chain query (BFS on `line_manager_id`). The BFS helper already exists from the line-manager parity work.
6. Handler: pass `u.ID` to `svc.Approve` / `svc.Reject`.
7. Seed: Admin/HR Manager → `approve_all`; Manager → `approve_team`. Remove `approve` from Manager seed (it would continue working as `approve_all`, which is too broad for a Manager).

*Blocked on D1. Unblocking G8 (seed) is a dependent step.*

---

### G3 — Bug: empty PATCH body reverts Approved → Pending

Go's `Update` service does NOT guard the no-op case. When a caller sends `PATCH /:id` with an
empty JSON body `{}` (all pointer fields nil, no attachment), the service still:
1. Applies no field patches (correct)
2. Recomputes `totalDays` (same value, harmless)
3. **Executes `if row.Status == Approved { row.Status = Pending }`** — status reverts!
4. Calls `s.leaves.Update(ctx, row)` — writes the row

Python exits early: `if not update_data and not (attachment and attachment.filename): return lr, warnings`.

**Fix** (3 lines in `leave_service.go`, `Update` method, before the status-revert block):
```go
hasChanges := in.FromDate != nil || in.ToDate != nil || in.LeavePeriod != nil ||
    in.LeaveType != nil || in.Reason != nil || attachment != nil
if !hasChanges {
    read, _ := s.populateRead(ctx, row)
    return &dto.LeaveRequestWriteResult{Request: read, Warnings: []string{}}, nil
}
```

---

### G4 — DOCX attachment type not accepted

`allowedAttachmentMIME` in `leave_service.go` has `image/jpeg`, `image/png`, `image/gif`,
`image/webp`, `application/pdf`. Missing: `application/vnd.openxmlformats-officedocument.wordprocessingml.document`.

Problem: `http.DetectContentType` cannot detect DOCX — DOCX is a ZIP archive; sniff returns
`application/zip`. An extension-based secondary check is required.

**Fix** (subject to D2): After the MIME sniff gate, add an extension fallback:
```go
if !allowedAttachmentMIME[sniffed] && strings.ToLower(att.Ext) == ".docx" {
    sniffed = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    // allow through — DOCX cannot be sniffed, only detected by extension
}
```
Add `"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true` to the map.

---

### G5 — Attachment max size: 10 MB vs BA 5 MB

Both constants (`maxLeaveAttachmentBytes` in handler + `leaveAttachmentMaxBytes` in service)
reference 10 MB. BA SR-14 specifies 5 MB.

**Fix** (subject to D2): change both constants to `5 * 1024 * 1024`.

---

### G6 — Dashboard limit hardcoded at 5

`GetMyDashboard` service calls `s.leaves.Upcoming(ctx, emp.ID, today, 5)` and
`s.leaves.History(ctx, emp.ID, today, 5)`. Python exposes `?limit` (default 10, max 50).
Mobile BA SR-07 says "maximum 10 items per tab".

**Fix** (subject to D3): change `5` → `10`, or expose query param. See D3.

---

### G7 — Cancel: "balance restored" message missing

Python `cancel_leave_request` returns `(lr, was_approved: bool)`. The router conditionally
appends "Leave balance restored." to the message. BA AC-28 specifies this.

Go `Cancel` service calls `transitionStatus` which re-fetches the row — at that point the status
is already changed; `was_approved` is lost.

**Fix** (subject to D5): capture `wasApproved := row.Status == models.LeaveStatusApproved`
*before* calling `transitionStatus`. Return via an extended struct or separate bool:
```go
// In service Cancel:
row, _ := s.leaves.FindByID(ctx, id)
wasApproved := row.Status == models.LeaveStatusApproved
read, err := s.transitionStatus(...)
// return read, wasApproved, err
```
Handler constructs message from the flag.

---

### G8 — Manager role seed: update + delete permissions missing

Manager role seed lacks `leave_requests:update` and `leave_requests:delete`. Manager also
needs `leave_requests:approve_team` once D1 is locked.

Without these in the seed, Manager can't reach the `PATCH /:id` or `POST /:id/delete`
endpoints (403 at middleware), even to manage their own requests (which the service guards
correctly via ownership check + `asAdmin=false`).

**Fix**: Add to Manager's permission list in the idempotent boot-seed:
- `leave_requests:update`
- `leave_requests:delete`
- (conditional on D1=A) swap `leave_requests:approve` → `leave_requests:approve_team`

---

### G9 — Multi-select dept/position filter on list (NEITHER — defer)

BA DR-02-01 §3 specifies multi-select Department and Position filter chips. Both Python and
Go accept single `department_id` / `position_id` values. The employee-list query (from
employees-parity-2) already supports `DepartmentIDs []uuid.UUID` under the hood — the gap
is only in the leave list filter plumbing.

This is a FE-driven enhancement. Defer. Flag to BA that the API currently accepts only a
single value for each filter.

---

### G10 — Leave list export (NEITHER — BA open question)

BA AC-21/22: export button; BA open question on format (CSV vs Excel). Python has no export
route. Go has no export route. See D6 for decision.

---

## 5. Decisions (LOCKED 2026-06-08)

> All decisions locked by user sign-off on 2026-06-08. Implementation proceeds per the locked choices below.

---

### D1 — Permission split: `approve_team` / `approve_all` — **LOCKED: A (full split)**

**The question:** Go has only the legacy `leave_requests:approve`. Should we add the full
Python split, a partial split, or keep the single perm?

**The stakes:** `approve_team` enforces that a Manager can only approve their own
subordinates' requests — without it, every Manager can approve across the whole company.
The subordinate BFS chain is already implemented (line-manager parity).

**Options:**

**A — Full split (Python parity) ← Recommended**
- Registry: add `PermLeaveApproveTeam = "leave_requests:approve_team"` + `PermLeaveApproveAll = "leave_requests:approve_all"`. Keep `PermLeaveApprove = "leave_requests:approve"` as a legacy constant (not assignable via UI catalog).
- Service `checkCanApproveOrReject(ctx, approverUserID, lr)`: `*` or `approve_all` or legacy `approve` → pass; `approve_team` → BFS check via `line_manager_id`.
- Handler: pass `u.ID` to `svc.Approve/Reject`.
- Seed: Admin, HR Manager → `approve_all`; Manager → `approve_team`. No migration — boot-seed idempotent addition.
- Catalog: add approve_team + approve_all entries; remove `approve` from catalog (existing grants still work at runtime as `approve_all`).

**B — Partial split (approve_all only, no team scoping)**
- Add `approve_all` only; `approve` treated as alias.
- No subordinate chain logic.
- Managers can still approve anyone (no improvement over current state for scope).

**C — Keep as-is (document only)**
- No code change; document that scope split is planned.
- Does not resolve the CHECKPOINT item.

**Recommended: A.** The user prompt explicitly defers this here. The BFS helper already exists.

**Implication if A:**
- `Approve` and `Reject` service methods get a new `approverUserID uuid.UUID` parameter.
- `LeaveHandler.Approve` and `LeaveHandler.Reject` need to pass `u.ID`.
- A new `checkCanApproveOrReject` function in the service (mirrors Python exactly).
- Seed update: idempotent boot-time perm addition for Admin/HR Manager/Manager roles.

---

### D2 — Attachment file types and size limit — **LOCKED: B (keep Go types + add DOCX + fix size to 5 MB)**

**The question:**
- BA: PDF, PNG, JPG, DOCX; max **5 MB**
- Go today: PDF, JPEG, PNG, GIF, WEBP (no DOCX); max **10 MB**
- Sniffing DOCX is not possible via `http.DetectContentType` (ZIP-based format)

**Options:**

**A — Follow BA strictly**
- Accepted: PDF, PNG, JPG, DOCX (extension-based detection for DOCX)
- Remove GIF, WEBP
- Max: 5 MB

**B — Keep Go types, add DOCX, fix size ← Recommended**
- Keep PDF, JPEG, PNG, GIF, WEBP (harmless — useful for evidence photos / screenshots)
- Add DOCX via extension fallback (`.docx` + sniff is `application/zip` → accept as DOCX MIME)
- Max: **5 MB** (mandatory — BA is explicit)
- Flag to BA: "Go additionally accepts GIF and WEBP (harmless; useful for photo evidence)"

**C — Skip DOCX, fix size only**
- Change max to 5 MB; keep current types; skip DOCX
- Flag to BA that DOCX is not feasible via standard byte-sniff

**Recommended: B.** GIF/WEBP are harmless extras and useful. DOCX is in the BA spec and
the extension-based fallback is safe (Go will accept only `.docx` extension, not all ZIPs).

---

### D3 — Dashboard limit (hardcoded 5 vs BA 10) — **LOCKED: A (fix to 10)**

**The question:** Fix the hardcoded `5` to `10`, or expose as a query param like Python?

**Options:**

**A — Fix to 10 (BA conformant) ← Recommended**
Change hardcoded `5` → `10` in `GetMyDashboard`. No query param added.

**B — Expose `?limit` query param (Python parity)**
Add `limit int` to the dashboard endpoint (default 10, max 50). Update service signature.

**C — Keep at 5** (non-conformant with BA)

**Recommended: A.** Mobile BA SR-07 says max 10; a configurable limit is overkill for a
dashboard summary widget. Keep it simple.

---

### D4 — Empty PATCH no-op guard — **LOCKED: mandatory bug fix**

This is a correctness bug, not a design choice. An empty PATCH body on an Approved request
**must not** revert it to Pending. Fix is mandatory and unambiguous. Implement per G3 above.
No sign-off needed — ship it.

---

### D5 — Cancel: "balance restored" message tracking — **LOCKED: A (add was_approved flag)**

**The question:** Should the cancel endpoint return a `was_approved` flag so the FE can
show "Leave balance restored." per BA AC-28?

**Options:**

**A — Add `was_approved` to service + handler ← Recommended**
- Service `Cancel` captures `wasApproved` before `transitionStatus`.
- Returns `(*dto.LeaveRequestRead, bool, error)` or a new `LeaveRequestCancelResult` struct.
- Handler constructs message: base + conditional "Leave balance restored."

**B — FE reads previous status (no server change)**
- FE already has the row loaded before clicking Cancel.
- FE shows the message on its own.
- No API change required.

**Recommended: A.** Server-side authority for the message is cleaner; Python parity; message
correctness guaranteed even if FE state is stale.

---

### D6 — Leave list export — **LOCKED: A (Excel via excelize)**

**The question:** Add export now (BA AC-21/22) or defer until BA resolves the format question?

**Background:** BA open question: "Export format: CSV or Excel (.xlsx)? Which columns included?"
Python has no export route. `excelize` is already in `go.mod` from attendance parity.

**Options:**

**A — Add Excel export now ← Recommended**
- `GET /leave-requests/export` — gated `leave_requests:read`; no manage perm → own records
  only (same scoping as the list endpoint)
- Filter params: same as list (`search`, `status`, `department_id`, `position_id`)
- Columns: Full Name, Department, Position, From Date, To Date, Leave Period, Leave Type,
  Total Days, Status
- Format: `.xlsx` via `xuri/excelize/v2` (already in `go.mod`) — consistent with attendance
- No new migration

**B — Defer until BA closes format question**
No export until BA confirms CSV vs Excel. Track in CHECKPOINT outstanding items.

**C — Add CSV export (simpler)**
`text/csv` via stdlib `encoding/csv`. No new dependency. Less rich than Excel.

**Recommended: A.** `excelize` is already present from attendance. Excel matches the attendance
export pattern. If BA later requests CSV, a second format can be added. The columns are
self-evident from the list page spec.

---

### D7 — Manager role seed: update / delete / approve_team — **LOCKED: A (all three)**

**The question:** What permissions should Manager role have for leave requests?

**Current Manager seed (Phase 5):** `leave_requests:read`, `leave_requests:create`,
`leave_requests:cancel`, `leave_requests:approve` (legacy).

**Proposed additions:**

| Perm | Who benefits | Service-side scope |
|---|---|---|
| `leave_requests:update` | Manager editing own request | `asAdmin=false` → owns it + Pending/Approved only |
| `leave_requests:delete` | Manager deleting own Pending | `asAdmin=false` → owns it + Pending only |
| `leave_requests:approve_team` | Manager approving subordinates' requests | service `checkCanApproveOrReject` + BFS |

**Options:**

**A — Add all three (conditional on D1=A) ← Recommended**
Manager: read, create, update, delete, cancel, approve_team
Drop `approve` from Manager seed (covered by `approve_team`; `approve` would grant approve_all
scope which is too broad for a Manager).

**B — Only update + delete; defer approve change to D1 finalization**
Manager: read, create, update, delete, cancel, approve (unchanged)
Defer: swap `approve` → `approve_team` when D1 is locked.

**Recommended: A contingent on D1=A being locked.** Both decisions should be locked together.

---

## 6. Out of scope for this audit / follow-ups

- **Multi-select dept/position list filter (G9)** — needs FE-side work too; defer.
- **Leave list export format** — BA open question; D6 proposes to proceed with Excel anyway.
- **Notification on status change** — BA AC open question; out of scope for this pass.
- **Approval delegation / multi-level workflow** — BA explicitly out of scope (DR-02-05 §8).
- **FE wiring** — web repo is self-managed (`api_info_go/`); produce FE doc after PRs land.
- **Quota audit log UI** — BA open question; server-side logging is out of scope here.
- **Mobile Create Leave Request (DR-002-001-02)** — this DR was read; it shares the same
  `POST /leave-requests` API. No additional endpoints required.

---

## Appendix: Gap summary

| Gap | Description | Severity | Decision |
|---|---|---|---|
| G1+G2 | `approve_team`/`approve_all` split + scope check | High | D1 |
| G3 | Empty PATCH reverts Approved→Pending (bug) | High | D4 (bug fix, no choice) |
| G4 | DOCX attachment not accepted | Medium | D2 |
| G5 | Attachment max 10 MB vs BA 5 MB | Medium | D2 |
| G6 | Dashboard limit hardcoded 5 vs BA 10 | Low | D3 |
| G7 | Cancel "balance restored" message missing | Low | D5 |
| G8 | Manager seed lacks update/delete perms | Medium | D7 |
| G9 | Multi-select dept/pos filter (neither) | Low | defer |
| G10 | Leave list export (neither) | Medium | D6 |
