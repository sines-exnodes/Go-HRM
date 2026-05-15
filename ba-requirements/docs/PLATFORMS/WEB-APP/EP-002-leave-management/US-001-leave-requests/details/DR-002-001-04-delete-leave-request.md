---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-04
detail_name: "Delete Leave Request"
parent_requirement: FR-US-001-12
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
  - path: "./DR-002-001-03-update-leave-request.md"
    relationship: sibling
---

# Detail Requirement: Delete Leave Request

**Detail ID:** DR-002-001-04
**Parent Requirement:** FR-US-001-12
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with leave management permission**, I want to **delete a leave request that is no longer needed**, so that **the leave requests list stays clean and irrelevant records are removed from active views**.

**Purpose:** Allow authorized users to soft-delete leave requests. This removes the request from all active views (Leave Requests List, filters, exports) but preserves the data in the database for audit purposes. Employees can delete their own Pending requests; admin/managers can delete any request regardless of status.

**Target Users:**
- **Employees** — delete their own **Pending** requests only
- **Admin/Managers** — delete **any** request (Pending, Approved, Rejected, or Cancelled)

**Key Functionality:**
- Triggered from gear icon dropdown on the Leave Requests List — no separate page
- Confirmation dialog before deletion ("Are you sure?")
- Soft delete — record hidden from active views, data preserved in database
- Permission-based: employees limited to own Pending; admin/managers unrestricted
- If deleting an Approved request, the leave balance is restored

---

## 2. User Workflow

**Entry Point:** Leave Requests List → gear icon on a request row → select "Delete"

**Preconditions:**
- User is signed in (US-001)
- User has leave management permission (US-004)
- For employees: request must be their own AND status must be Pending
- For admin/managers: any request, any status

**Main Flow:**
1. User clicks gear icon on a leave request row
2. User selects "Delete" from the dropdown
3. System shows Confirmation Dialog with request summary (employee name, dates, leave type)
4. Dialog text: "Are you sure you want to delete this leave request? This action cannot be undone."
5. User clicks "Delete" (confirm)
6. System shows loading state on Delete button; buttons disabled
7. System soft-deletes the leave request
8. If the request was **Approved** → system restores the leave days to the employee's balance
9. Dialog closes; Leave Requests List refreshes — deleted request no longer visible
10. Success toast: "Leave request has been deleted"

**Alternative Flows:**
- **Alt 1 — Cancel confirmation:** User clicks "Cancel" in dialog → dialog closes, no changes made.
- **Alt 2 — Close via X:** User clicks X button → dialog closes, no changes made.
- **Alt 3 — Employee attempts non-Pending:** Employee tries to delete an Approved/Rejected/Cancelled request → Delete option not visible in gear menu.
- **Alt 4 — Employee attempts other's request:** Employee cannot see Delete for requests belonging to other employees (gear icon actions filtered).
- **Alt 5 — Race condition:** Between dialog open and confirm, the request status changes → system checks at execution and proceeds (soft delete works for any current status for admin).
- **Alt 6 — API error:** Delete fails → error toast: "Failed to delete leave request. Please try again." Dialog remains open.
- **Alt 7 — Request already deleted:** Another admin deleted the request → error toast: "Leave request not found." → list refreshes.

**Exit Points:**
- **Success:** Request soft-deleted → toast → list refreshes
- **Cancel:** Dialog closed → no changes
- **Error:** Error toast → dialog stays open or list refreshes

---

## 3. Field Definitions

### Input Fields

No input fields — deletion is confirmed via a modal dialog only.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Delete (Confirm) | Danger/Destructive button | Shown in Confirmation Dialog; disabled while processing | Soft-deletes the leave request | Red/danger style |
| Cancel | Secondary button | Shown in Confirmation Dialog | Closes dialog without deleting | Returns to list |
| Close (X) | Icon button | Shown in dialog | Dismisses dialog | Same as Cancel |

---

## 4. Data Display

### Confirmation Dialog

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Dialog title | Text | "Delete Leave Request" | Labels the dialog |
| Employee name | Text | "[Employee Name]" | Identifies whose request |
| Leave dates | Text | "[From Date] — [To Date]" | Identifies the period |
| Leave type | Text | "[Leave Type]" | Identifies the type |
| Warning text | Static text | "Are you sure you want to delete this leave request? This action cannot be undone." | Communicates significance |

### Dialog Layout

```
┌─────────────────────────────────────────┐
│  Delete Leave Request           [X]     │
├─────────────────────────────────────────┤
│                                         │
│  Are you sure you want to delete        │
│  this leave request?                    │
│                                         │
│  Employee: [Name]                       │
│  Dates: [From] — [To]                  │
│  Type: [Leave Type]                     │
│                                         │
│  This action cannot be undone.          │
│                                         │
│              [Cancel]  [Delete]         │
└─────────────────────────────────────────┘
```

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Confirmation Dialog | User clicks Delete in gear menu | Modal with request summary, Cancel + Delete buttons |
| Deleting | Delete clicked, processing | Delete button spinner + disabled; Cancel disabled |
| Success | Request deleted | Dialog closes; list refreshes; toast: "Leave request has been deleted" |
| Dismissed | User clicks Cancel or X | Dialog closes; no changes |
| API error | Delete fails | Error toast; dialog stays open for retry |
| Already deleted | Request removed by another user | Error toast: "Leave request not found"; list refreshes |

### Gear Icon Delete Visibility

| Status | Employee (own) | Admin/Manager |
|--------|---------------|---------------|
| Pending | ✅ Delete visible | ✅ Delete visible |
| Approved | ❌ Not visible | ✅ Delete visible |
| Rejected | ❌ Not visible | ✅ Delete visible |
| Cancelled | ❌ Not visible | ✅ Delete visible |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Gear Icon Visibility:**
- **AC-01:** Delete option is available in the gear icon dropdown based on user role and request status
- **AC-02:** Employees see Delete only for their own Pending requests
- **AC-03:** Admin/managers see Delete for any request regardless of status
- **AC-04:** Employees do NOT see Delete for their own Approved, Rejected, or Cancelled requests
- **AC-05:** Employees do NOT see Delete for other employees' requests

**Confirmation Dialog:**
- **AC-06:** Clicking Delete in gear menu shows a Confirmation Dialog (modal overlay)
- **AC-07:** Dialog displays request summary: employee name, dates (From — To), leave type
- **AC-08:** Dialog includes warning: "Are you sure you want to delete this leave request? This action cannot be undone."
- **AC-09:** Dialog has Cancel button, Delete button (danger style), and X close button
- **AC-10:** Delete button uses danger/red style to signal destructive action

**Delete Behavior:**
- **AC-11:** Confirmed deletion soft-deletes the request — record preserved in database with "deleted" flag
- **AC-12:** Deleted request is removed from the Leave Requests List immediately
- **AC-13:** Deleted request is excluded from all filters, searches, and exports
- **AC-14:** Success toast displays: "Leave request has been deleted"
- **AC-15:** After deletion, any active search/filter on the list remains in effect

**Leave Balance Restoration:**
- **AC-16:** If a Pending request is deleted, no balance change (leave was not yet deducted)
- **AC-17:** If an Approved request is deleted by admin, the approved leave days are restored to the employee's balance
- **AC-18:** If a Rejected or Cancelled request is deleted by admin, no balance change

**Dialog Interactions:**
- **AC-19:** Clicking Cancel closes the dialog without deleting
- **AC-20:** Clicking X closes the dialog without deleting
- **AC-21:** Delete button shows loading spinner and is disabled while processing
- **AC-22:** Dialog is a modal overlay — list behind is not interactive while dialog is open
- **AC-23:** Default focus is on Cancel button (not Delete) when dialog opens
- **AC-24:** Pressing Escape closes the dialog (same as Cancel)

**Error Handling:**
- **AC-25:** If delete fails, error toast: "Failed to delete leave request. Please try again." Dialog stays open.
- **AC-26:** If request was already deleted by another user, error toast: "Leave request not found." List refreshes.

**Access Control:**
- **AC-27:** Delete action is visible only to users with leave management permission
- **AC-28:** Users with view-only permission cannot see Delete in gear menu

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee deletes own Pending | Gear → Delete on own Pending | Confirmation → deleted; toast; removed from list | High |
| Admin deletes any Pending | Gear → Delete on Pending | Confirmation → deleted | High |
| Admin deletes Approved | Gear → Delete on Approved | Confirmation → deleted; balance restored | High |
| Admin deletes Rejected | Gear → Delete on Rejected | Confirmation → deleted; no balance change | Medium |
| Admin deletes Cancelled | Gear → Delete on Cancelled | Confirmation → deleted; no balance change | Medium |
| Employee — Approved hidden | Employee checks own Approved | Delete not visible in gear | High |
| Employee — other's request | Employee checks other's Pending | Delete not visible | High |
| Cancel confirmation | Click Cancel in dialog | Dialog closes; no changes | High |
| Close via X | Click X in dialog | Dialog closes; no changes | Medium |
| Escape closes dialog | Press Escape | Dialog closes | Medium |
| Balance restored | Delete Approved (5 days) | Employee balance increases by 5 days | High |
| API error | Server fails | Error toast; dialog stays open | Medium |
| Already deleted | Another admin deleted | "Leave request not found" toast; list refreshes | Medium |
| Delete with active filter | Search active, delete request | Request removed; filter remains | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with leave management permission can access the Delete action.
- **SR-02:** Permissions configured via US-004. No role names hardcoded.
- **SR-03:** **Employee delete permission:** Employees can only delete their own requests with Pending status. The Delete option is hidden in the gear menu for all other scenarios.
- **SR-04:** **Admin/manager delete permission:** Admin/managers can delete any leave request regardless of status (Pending, Approved, Rejected, Cancelled).
- **SR-05:** Deletion is a **soft delete** — the record is preserved in the database with a "deleted" flag. Removed from all active views.
- **SR-06:** Soft-deleted requests are excluded from: Leave Requests List, all filters, searches, exports, and overlapping dates checks.
- **SR-07:** **Leave balance restoration:** When an Approved request is deleted, the approved leave days are restored to the employee's Leave Days Remaining. Formula: Leave Days Remaining += Total Days of the deleted request.
- **SR-08:** Deleting a Pending, Rejected, or Cancelled request does NOT affect the leave balance (leave was not deducted for these statuses).
- **SR-09:** Deletion requires confirmation via dialog — no direct delete on click.
- **SR-10:** Default focus on Cancel button in dialog — prevents accidental deletion.
- **SR-11:** After deletion, any active search/filter state on the list is preserved.
- **SR-12:** If the request was already deleted by another user between dialog open and confirm, system returns "not found" error and refreshes the list.

**State Transitions:**
```
[Leave Requests List] → [Gear → Delete] → [Confirmation Dialog]
[Confirmation Dialog] → [Cancel / X / Escape] → [List unchanged]
[Confirmation Dialog] → [Confirm Delete] → [Deleting] → [Soft delete + balance restore if Approved] → [List refreshes]
[Deleting] → [API error] → [Error toast; dialog stays open]
[Deleting] → [Already deleted] → [Error toast; list refreshes]
```

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls delete access; determines employee vs admin permissions
- **Leave Balance System:** Balance restored when Approved request is deleted
- **DR-002-001-01 (Leave Requests List):** Gear icon trigger; list refreshes after delete
- **DR-002-001-02 (Create Leave Request):** Creates requests that this feature can delete
- **DR-002-001-03 (Update Leave Request):** Overlapping dates check excludes soft-deleted requests

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Confirmation dialog shows request summary (name, dates, type) — user confirms they're deleting the right request
- **UX-02:** Delete button uses danger/red style — visually signals destructive action
- **UX-03:** Default focus on Cancel button — prevents accidental deletion if user presses Enter
- **UX-04:** Delete button shows spinner while processing — prevents double-click
- **UX-05:** Pressing Escape closes the dialog — consistent keyboard shortcut across all dialogs
- **UX-06:** Keyboard focus trapped inside dialog — Tab cycles through dialog elements only
- **UX-07:** Toast auto-dismisses after 5 seconds on success
- **UX-08:** Error toast persists until dismissed — ensures user notices the failure
- **UX-09:** List search/filter state preserved after deletion — no disruptive reset

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Centered modal dialog ~400px, vertically centered |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard focus trapped inside dialog (Tab cycles within dialog only)
- [x] Escape key closes dialog
- [x] Default focus on Cancel button
- [x] Screen reader announces dialog title and content (ARIA role="dialog")
- [x] Delete button marked as destructive via ARIA
- [x] Sufficient color contrast on danger button

**Design References:**
- No Figma design available — pattern follows Delete Department (DR-008-001-04) and Delete Skill (DR-008-003-04) confirmation dialog pattern
- Adapted with request summary display (name, dates, type) unique to leave requests

---

## 8. Additional Information

### Out of Scope
- Hard delete (permanent removal) — soft delete only
- Restoration of soft-deleted requests (future enhancement)
- Bulk delete (multiple requests at once)
- Auto-delete after a retention period
- Notification to employee when admin deletes their request
- Audit log of deletions
- Mobile or tablet layout

### Open Questions
- [ ] **Notification on delete:** Should the employee be notified when an admin deletes their Approved request (since balance is restored)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Retention policy:** How long are soft-deleted records kept in the database? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-01: Leave Requests List | Gear icon trigger; list refreshes after delete |
| DR-002-001-02: Create Leave Request | Creates requests this feature can delete |
| DR-002-001-03: Update Leave Request | Overlapping check excludes soft-deleted requests |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls delete permissions |
| Leave Balance System | Balance restored when Approved request deleted |

### Notes
- **Delete vs Cancel:** These are distinct actions. Cancel changes status to "Cancelled" (request remains visible in list with grey badge). Delete removes the request from all active views entirely (soft delete). Cancel is a status change; Delete is a record removal.
- **Leave balance restoration** only applies when deleting Approved requests — because only Approved requests have had leave days deducted from the balance. Pending/Rejected/Cancelled requests have no balance impact.
- **Employee permission is intentionally limited** to own Pending requests. This prevents employees from deleting requests that have already been acted upon by management (Approved/Rejected) or formally cancelled.
- **Confirmation dialog includes request summary** — unlike simpler delete confirmations (Department, Skill), the leave request dialog shows employee name, dates, and type to help the admin identify the exact request being deleted.
- Pattern follows the established HRM confirmation dialog design (Cancel + Delete buttons, X close, Escape key, focus on Cancel).

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
| 1.0 | 2026-03-27 | BA Agent | Initial draft — soft delete with role-based permissions and balance restoration |
