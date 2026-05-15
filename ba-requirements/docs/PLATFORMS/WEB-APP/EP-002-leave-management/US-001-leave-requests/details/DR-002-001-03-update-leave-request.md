---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-03
detail_name: "Update Leave Request"
parent_requirement: FR-US-001-10
status: draft
version: "1.0"
created_date: 2026-03-27
last_updated: 2026-03-27
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-002-001-01-leave-requests-list.md"
    relationship: sibling
  - path: "./DR-002-001-02-create-leave-request.md"
    relationship: sibling
---

# Detail Requirement: Update Leave Request

**Detail ID:** DR-002-001-03
**Parent Requirement:** FR-US-001-10
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with leave management permission**, I want to **edit a pending or approved leave request's dates, leave type, period, reason, and attachment**, so that **I can correct or update details before or after initial approval**.

**Purpose:** Allow authorized users to modify an existing leave request. The employee who submitted can edit their own; admin/managers can edit any eligible request. Only Pending and Approved requests can be edited — Rejected and Cancelled are locked. Editing an Approved request **reverts it to Pending** status, requiring re-approval.

**Target Users:**
- **Employees** — edit their own Pending or Approved requests
- **Admin/Managers** — edit any Pending or Approved request

**Key Functionality:**
- Pre-filled form with current leave request values
- Employee info card displayed read-only (no dropdown — employee cannot be changed)
- Editable for Pending and Approved statuses only — Rejected/Cancelled are locked
- Editing an Approved request reverts status to Pending (re-approval required)
- Leave balance warning and overlapping dates warning (both non-blocking)
- Overlapping dates check excludes the current request (self-exclusion)
- Returns to Leave Requests List on success

---

## 2. User Workflow

**Entry Point:** Leave Requests List → gear icon on a Pending/Approved request → select "Edit"

**Preconditions:**
- User is signed in (US-001)
- User has leave management permission (US-004)
- The leave request status is Pending or Approved (not Rejected or Cancelled)
- Employee editing own request, OR admin/manager editing any request

**Main Flow:**
1. User clicks gear icon on a leave request row → selects "Edit"
2. System validates request status is Pending or Approved
3. System navigates to "Update Leave Request" page
4. Employee info card displays read-only: Full name, Email, Phone, Department, Position, Leave Days Remaining
5. Leave Info fields pre-filled with current values: From Date, To Date, Leave Period, Leave Type, Reason, Attachment
6. User modifies one or more fields
7. User clicks Save
8. System validates all mandatory fields and date logic
9. System checks leave balance and overlapping dates (excluding self) — shows warnings if applicable (non-blocking)
10. System updates the leave request
11. If request was **Approved** → status reverts to **Pending** (re-approval required)
12. If request was **Pending** → status remains **Pending**
13. Success toast: "Leave request has been updated" (+ "Status reverted to Pending" if was Approved)
14. System redirects to Leave Requests List — updated request visible

**Alternative Flows:**
- **Alt 1 — Empty mandatory field:** Inline error "[Field name] is required". Form not submitted.
- **Alt 2 — To Date before From Date:** Inline error: "To Date must be on or after From Date". Form not submitted.
- **Alt 3 — Insufficient leave balance (warning):** Yellow warning — non-blocking.
- **Alt 4 — Overlapping dates (warning):** Yellow warning — non-blocking. Excludes current request from overlap check.
- **Alt 5 — Invalid attachment file:** Inline error (PDF, PNG, JPG, DOCX; max 5MB).
- **Alt 6 — Cancel (form modified):** "Discard unsaved changes?" dialog.
- **Alt 7 — Cancel (form untouched):** Redirect to Leave Requests List immediately.
- **Alt 8 — Request status changed:** Between page load and Save, another user rejected/cancelled the request → error toast: "This leave request has been updated by another user." → redirect to list.
- **Alt 9 — Save without changes:** No-op — no error, no status change, no toast.

**Exit Points:**
- **Success:** Request updated → toast → redirect to list
- **Cancel:** Redirect to list (with or without confirmation)
- **Error:** Validation errors inline; user corrects and retries

---

## 3. Field Definitions

### Employee Section (read-only — no input)

| Data | Format | Description |
|------|--------|-------------|
| Full name | Label + value | Fixed — cannot be changed |
| Email | Label + value | Fixed |
| Phone number | Label + value | Fixed |
| Department | Label + value | Fixed |
| Position | Label + value | Fixed |
| Leave Days Remaining | Label + decimal value | Real-time balance |

### Input Fields — Leave Info Section

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| From Date | Date picker (278px) | Not empty; must be ≤ To Date; past dates allowed | Yes (*) | Pre-filled with current From Date | Leave start date |
| To Date | Date picker (278px) | Not empty; must be ≥ From Date | Yes (*) | Pre-filled with current To Date | Leave end date |
| Leave Period | Dropdown (278px) | Must select a value | Yes (*) | Pre-filled with current period | Full Day, Morning Half, Afternoon Half |
| Leave Type | Dropdown (278px) | Must select a value | Yes (*) | Pre-filled with current type | Annual, Sick, Personal, Maternity, Unpaid |
| Reason | Text area (576px) | Not empty; trimmed | Yes (*) | Pre-filled with current reason | Explanation for leave |
| Attachment | File input (576px) | PDF, PNG, JPG, DOCX; max 5MB | No | Pre-filled with current file (if any) | Supporting document |

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Header action bar | Always visible | If dirty → discard dialog; if clean → redirect | Returns to list |
| Save (header) | Button (primary) | Header action bar | Disabled + spinner while saving | Validates → updates → reverts status if needed → toast → redirect | Submit |
| Save (bottom) | Button (primary, full-width) | Below form, 600×40px | Same as header Save | Identical to header Save | Convenience duplicate |

### Validation Error Messages

| Condition | Error Message | Display | Blocks Submit? |
|-----------|--------------|---------|---------------|
| Mandatory field empty | "[Field name] is required" | Inline, below field | Yes |
| To Date before From Date | "To Date must be on or after From Date" | Inline, below To Date | Yes |
| Insufficient leave balance | "Insufficient leave balance. [X] days remaining, [Y] days requested." | Yellow inline + warning toast | **No** |
| Overlapping dates | "This employee already has a leave request for overlapping dates." | Yellow inline + warning toast | **No** |
| Invalid attachment type | "Only PDF, PNG, JPG, and DOCX files are accepted" | Inline, below Attachment | Yes (file rejected) |
| Attachment too large | "File size must not exceed 5MB" | Inline, below Attachment | Yes (file rejected) |
| Request status changed | "This leave request has been updated by another user." | Error toast → redirect to list | Yes (redirect) |

---

## 4. Data Display

### Information Shown to User

**Employee Section (always read-only):**

| Data Name | Format | Business Meaning |
|-----------|--------|-----------------|
| Employee info card | 6 fields in 2×3 grid — always visible, no dropdown | Full name, Email, Phone (left) / Dept, Position, Leave Days Remaining (right) |

**Leave Info Section (pre-filled):**

| Data Name | Format | Pre-filled With | Business Meaning |
|-----------|--------|----------------|-----------------|
| From Date | Date picker (278px) | Current From Date | Leave start |
| To Date | Date picker (278px) | Current To Date | Leave end |
| Leave Period | Dropdown (278px) | Current period value | Full Day / Morning Half / Afternoon Half |
| Leave Type | Dropdown (278px) | Current type value | Annual / Sick / Personal / Maternity / Unpaid |
| Reason | Text area (576px) | Current reason text | Justification |
| Attachment | File input (576px) | Current filename (if any) | Supporting document |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads | Employee info card (read-only) + Leave Info pre-filled; Save enabled |
| Leave balance warning | Modified dates/period exceed remaining | Yellow inline warning + toast — still submittable |
| Overlapping dates warning | Modified dates overlap with another request (excluding self) | Yellow inline warning + toast — still submittable |
| Validation error | Save with invalid/empty fields | Inline red error(s) below affected fields |
| Saving | Save clicked, in progress | Both Save buttons spinner + disabled; fields disabled |
| Success (Pending stayed) | Pending request updated | Toast: "Leave request has been updated" → redirect |
| Success (Approved reverted) | Approved request updated | Toast: "Leave request has been updated. Status reverted to Pending." → redirect |
| No changes | Save without modifying anything | No-op — no error, no status change |
| Race condition | Status changed by another user | Error toast: "This leave request has been updated by another user." → redirect |
| Discard confirmation | Cancel with modified form | Modal: "Discard unsaved changes?" |

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Update Leave Request         [Cancel] [Save]   │
│  200px       │                                                  │
│              │     ┌────────────────────────────────────┐       │
│              │     │ Employee (read-only)               │       │
│              │     │ ┌─────────────┬─────────────────┐  │       │
│              │     │ │ Full name   │ Department      │  │       │
│              │     │ │ Email       │ Position        │  │       │
│              │     │ │ Phone       │ Leave Remaining │  │       │
│              │     │ └─────────────┴─────────────────┘  │       │
│              │     ├────────────────────────────────────┤       │
│              │     │ Leave Info (pre-filled)            │       │
│              │     │ *From Date      *To Date           │       │
│              │     │ *Leave Period   *Leave Type        │       │
│              │     │ *Reason                            │       │
│              │     │  Attachment [current file]         │       │
│              │     ├────────────────────────────────────┤       │
│              │     │ [            Save              ]   │       │
│              │     └────────────────────────────────────┘       │
└──────────────┴──────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Page title displays "Update Leave Request" in Geist Semibold 24px
- **AC-02:** Form has two sections: Employee (read-only card) + Leave Info (pre-filled) — centered 600px card
- **AC-03:** Cancel + Save in header; Save full-width at bottom — both identical
- **AC-04:** Employee info card always visible — no dropdown, no ability to change employee

**Employee Section:**
- **AC-05:** Employee info card displays 6 read-only fields: Full name, Email, Phone (left) | Department, Position, Leave Days Remaining (right)
- **AC-06:** Leave Days Remaining shows current real-time balance (decimal)

**Editable Statuses:**
- **AC-07:** Only Pending and Approved requests can be edited
- **AC-08:** Rejected requests cannot be edited — gear icon does not show Edit
- **AC-09:** Cancelled requests cannot be edited — gear icon does not show Edit

**Status Revert:**
- **AC-10:** Editing a Pending request keeps status as Pending
- **AC-11:** Editing an Approved request **reverts status to Pending** — re-approval required
- **AC-12:** When Approved → Pending revert occurs, toast includes: "Status reverted to Pending."
- **AC-13:** The reverted request appears in the list with Pending (amber) badge

**Pre-filled Fields:**
- **AC-14:** From Date pre-filled with current From Date value
- **AC-15:** To Date pre-filled with current To Date value
- **AC-16:** Leave Period pre-filled with current Leave Period selection
- **AC-17:** Leave Type pre-filled with current Leave Type selection
- **AC-18:** Reason pre-filled with current reason text
- **AC-19:** Attachment shows current filename if a file was previously uploaded

**Date Validation:**
- **AC-20:** Past dates are allowed (consistent with Create)
- **AC-21:** To Date must be on or after From Date — inline error if violated
- **AC-22:** Total Days recalculated automatically when dates or period change

**Leave Balance Warning:**
- **AC-23:** If updated request days exceed Leave Days Remaining, yellow inline warning + toast shown
- **AC-24:** Warning message: "Insufficient leave balance. [X] days remaining, [Y] days requested."
- **AC-25:** Warning does NOT block submission

**Overlapping Dates Warning:**
- **AC-26:** If updated dates overlap with another request for the same employee, yellow inline warning + toast shown
- **AC-27:** Overlapping check **excludes the current request** (self-exclusion) — the request does not overlap with itself
- **AC-28:** Warning does NOT block submission

**Mandatory Field Validation:**
- **AC-29:** All mandatory fields must be filled: From Date, To Date, Leave Period, Leave Type, Reason
- **AC-30:** Empty mandatory fields show inline error: "[Field name] is required"
- **AC-31:** Reason trimmed; whitespace-only rejected

**Attachment:**
- **AC-32:** Attachment is optional — can be removed or replaced
- **AC-33:** New file must be PDF, PNG, JPG, or DOCX; max 5MB
- **AC-34:** Invalid file shows inline error

**Save Behavior:**
- **AC-35:** Save updates the leave request with validated data
- **AC-36:** Toast: "Leave request has been updated" (+ status revert message if applicable)
- **AC-37:** Redirects to Leave Requests List after save
- **AC-38:** Save buttons show spinner + disabled while processing
- **AC-39:** Save without changes is a no-op — no error, no status change, no toast

**Cancel Behavior:**
- **AC-40:** Cancel on untouched form redirects without confirmation
- **AC-41:** Cancel on modified form shows "Discard unsaved changes?" dialog

**Race Condition:**
- **AC-42:** If request status was changed by another user between page load and Save, error toast: "This leave request has been updated by another user." → redirect to list
- **AC-43:** Race condition check is performed server-side at save time

**Access Control:**
- **AC-44:** Edit action visible only for Pending/Approved requests to authorized users
- **AC-45:** Direct URL access to edit a Rejected/Cancelled request redirects to list with error

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Edit Pending — change dates | Modify From/To, Save | Updated; stays Pending; toast; redirect | High |
| Edit Approved — change reason | Modify Reason, Save | Updated; **reverted to Pending**; toast with revert msg | High |
| Edit Approved — revert badge | Edit Approved request | List shows Pending (amber) badge after save | High |
| Pre-filled values | Open edit form | All fields show current values | High |
| Employee locked | Edit form loads | Info card read-only; no dropdown | High |
| To Date before From | Change To Date before From | Inline error | High |
| Past dates OK | From Date in past | Saves successfully | High |
| Balance warning | Increase to exceed remaining | Yellow warning; still saveable | High |
| Overlap warning (self excluded) | Dates overlap own request | No warning (self excluded) | Medium |
| Overlap warning (other) | Dates overlap different request | Yellow warning; still saveable | Medium |
| No changes — no-op | Open form, Save immediately | Nothing happens | Medium |
| Race condition | Another user rejects while editing | Error toast → redirect | Medium |
| Replace attachment | Upload new PDF | New filename shown; old replaced | Medium |
| Remove attachment | Clear file | Saved without attachment | Medium |
| Cancel dirty | Modify field, Cancel | Discard dialog | Medium |
| Rejected request URL | Direct URL to edit rejected | Redirect to list with error | Medium |
| Unauthorized | User without permission | Edit not visible in gear menu | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with leave management permission can access the Edit Leave Request page
- **SR-02:** Permissions configured via US-004. No role names hardcoded.
- **SR-03:** **Editable statuses:** Only Pending and Approved requests can be edited. Rejected and Cancelled requests are locked — the Edit action is not available.
- **SR-04:** **Status revert rule:** When an **Approved** request is edited, the status automatically reverts to **Pending**. This requires re-approval from a manager/admin.
- **SR-05:** When a **Pending** request is edited, the status remains **Pending** — no change.
- **SR-06:** **Save without changes** is a no-op — no server call, no status change, no toast.
- **SR-07:** Employee info card is read-only — populated from the original request's employee profile. Employee cannot be changed on edit.
- **SR-08:** **Leave Days Remaining** is fetched real-time on page load — reflects current balance.
- **SR-09:** **Past dates are allowed** — consistent with Create flow.
- **SR-10:** **To Date must be ≥ From Date.** Requests with invalid date order are rejected.
- **SR-11:** **Total Days calculation:** Inclusive counting. Leave Period determines count: Full Day = 1.0, Morning Half / Afternoon Half = 0.5.
- **SR-12:** **Leave balance warning only** — non-blocking. If updated days > remaining → yellow warning.
- **SR-13:** **Overlapping dates warning only** — non-blocking. Check **excludes the current request** from comparison. Compares against Pending and Approved requests only.
- **SR-14:** Reason field trimmed. Whitespace-only rejected as empty.
- **SR-15:** Attachment: PDF, PNG, JPG, DOCX; max 5MB. User can replace or remove existing attachment.
- **SR-16:** Both Save buttons trigger identical logic.
- **SR-17:** **Race condition protection:** At save time, system checks the request's current status. If status has been changed by another user (e.g., rejected or cancelled since page load), save is blocked and user is redirected to list.
- **SR-18:** "Form dirty" detection covers all Leave Info fields. Changes trigger discard confirmation on Cancel.
- **SR-19:** Direct URL access to edit a Rejected or Cancelled request returns an error and redirects to list.

**State Transitions:**
```
[Leave Requests List] → [Gear → Edit (Pending/Approved)] → [Update Form (pre-filled)]
[Update Form] → [Save, valid, Pending request] → [Updated; stays Pending] → [Toast] → [List]
[Update Form] → [Save, valid, Approved request] → [Updated; reverted to Pending] → [Toast + revert msg] → [List]
[Update Form] → [Save, no changes] → [No-op]
[Update Form] → [Save, invalid] → [Inline errors] → [Stay on form]
[Update Form] → [Save, balance/overlap warning] → [Yellow warning] → [User saves again] → [Updated]
[Update Form] → [Save, race condition] → [Error toast] → [List]
[Update Form] → [Cancel (clean)] → [List]
[Update Form] → [Cancel (dirty)] → [Discard dialog]
```

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls edit access; determines who can edit which requests
- **US-005 (User Management):** Employee profile for read-only info card
- **Leave Balance System:** Leave Days Remaining calculation
- **DR-002-001-01 (Leave Requests List):** Entry point via gear → Edit; return destination after save
- **DR-002-001-02 (Create Leave Request):** Shares form layout and field definitions; Edit mirrors Create with pre-filled data
- **DR-002-001-04 (Approve/Reject):** Reverted Approved requests require re-approval via this flow

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** All Leave Info fields pre-filled with current values — user can see existing data and change only what's needed
- **UX-02:** Employee info card always visible — user confirms they're editing the correct request
- **UX-03:** Leave Days Remaining shown for context — helps assess impact of date/period changes
- **UX-04:** Balance and overlap warnings use yellow styling — communicates caution without hard error
- **UX-05:** Overlap check excludes self — editing dates within the same range doesn't trigger false overlap warning
- **UX-06:** Both Save buttons (header + bottom) for convenience on longer forms
- **UX-07:** Save buttons show spinner while processing — prevents double submission
- **UX-08:** Status revert toast explicitly states "Status reverted to Pending" — admin is aware re-approval is needed
- **UX-09:** No-op on save without changes — no confusing toast or status change
- **UX-10:** "Discard unsaved changes?" defaults focus on Cancel — prevents accidental data loss
- **UX-11:** Race condition handled gracefully — toast explains what happened and redirects

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Centered form card 600px |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields
- [x] Screen reader compatible — labels, pre-filled values announced
- [x] Date pickers accessible via keyboard
- [x] Balance/overlap warnings announced (role="alert")
- [x] Color contrast meets WCAG 2.1 AA
- [x] Focus indicators visible

**Design References:**
- Figma: [Update Leave Request](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC) — extracted from screenshot
- Pattern: Mirrors Create Leave Request (DR-002-001-02) with pre-filled data and locked employee

---

## 8. Additional Information

### Out of Scope
- Changing the employee on an existing request (employee is locked)
- Editing Rejected or Cancelled requests
- Bulk editing multiple requests
- Version history or audit trail of edits
- Notification to approver when a request is edited
- Notification to employee when admin edits their request
- Mobile or tablet layout

### Open Questions
- [ ] **Notification on edit:** Should the approver be notified when a Pending request is edited? Should the employee be notified when an admin edits their request? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Edit count limit:** Is there a maximum number of times a request can be edited? Or unlimited? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-01: Leave Requests List | Entry point via gear → Edit; return destination |
| DR-002-001-02: Create Leave Request | Shares form layout; Edit mirrors Create with pre-filled data |
| DR-002-001-04: Approve/Reject (planned) | Reverted Approved requests need re-approval |
| DR-002-001-05: Cancel Leave Request (planned) | Alternative action for requests the user no longer wants |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls edit access |
| US-005: User Management | Employee profile for info card |

### Notes
- **Key difference from Create:** No employee dropdown; employee info card is always read-only. All Leave Info fields are pre-filled with current values.
- **Approved → Pending revert** is the most important workflow rule in this DR. It prevents stale approvals — if the dates or type change, the approver must review again. The toast explicitly communicates this.
- **Overlapping dates self-exclusion** is critical — without it, every edit would trigger a false overlap warning because the request overlaps with itself.
- **Race condition protection** (SR-17) prevents conflicts when multiple admins manage the same request concurrently.
- **No-op on save without changes** (SR-06) is intentional — prevents unnecessary status reverts and confusing toasts when user opens the form and saves immediately.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-27 | BA Agent | Initial draft — mirrors Create with pre-filled data, status revert rule, self-exclusion overlap |
