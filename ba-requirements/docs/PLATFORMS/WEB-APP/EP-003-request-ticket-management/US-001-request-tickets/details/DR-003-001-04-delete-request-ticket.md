---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
detail_id: DR-003-001-04
detail_name: "Delete Request Ticket"
parent_requirement: FR-US-001-12
status: draft
version: "1.0"
created_date: 2026-04-03
last_updated: 2026-04-03
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-003-001-01-request-tickets-list.md"
    relationship: sibling
  - path: "./DR-003-001-02-create-request-ticket.md"
    relationship: sibling
  - path: "./DR-003-001-03-update-request-ticket.md"
    relationship: sibling
---

# Detail Requirement: Delete Request Ticket

**Detail ID:** DR-003-001-04
**Parent Requirement:** FR-US-001-12
**Story:** US-001-request-tickets
**Epic:** EP-003 (Request Ticket Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **authorized user**, I want to **delete a request ticket that is no longer needed**, so that **the request tickets list remains clean and irrelevant records are removed from active views**.

**Purpose:** Allow authorized users to soft-delete request tickets. This removes the ticket from all active views (Request Tickets List, filters, searches, exports) but preserves the data in the database for audit purposes. Employees can delete only their own tickets in Open status; admin/managers can delete any ticket regardless of status.

**Target Users:**
- **Employees** — delete their own **Open** tickets only
- **Admin/Managers** — delete **any** ticket regardless of status (Open, In Progress, On Hold, Resolved, Closed, or Cancelled)

**Key Functionality:**
- Triggered from the gear icon dropdown on the Request Tickets List — no separate page
- Confirmation dialog before deletion, showing a summary of the ticket being deleted
- Soft delete — record hidden from all active views, data preserved in the database
- Permission-based: employees limited to own Open tickets; admin/managers unrestricted
- No resource restoration — request tickets have no consumable resource associated with any status

---

## 2. User Workflow

**Entry Point:** Request Tickets List → gear icon on a ticket row → select "Delete"

**Preconditions:**
- User is signed in (EP-001 US-001)
- User has request ticket management permission (EP-001 US-004)
- For employees: ticket must be their own AND status must be Open
- For admin/managers: any ticket, any status

**Main Flow:**
1. User clicks the gear icon on a request ticket row
2. User selects "Delete" from the dropdown menu
3. System displays a Confirmation Dialog with a ticket summary (employee name, request type, subject)
4. Dialog text: "Are you sure you want to delete this request ticket? This action cannot be undone."
5. User clicks "Delete" (confirm)
6. System shows a loading state on the Delete button; all dialog buttons are disabled
7. System soft-deletes the request ticket (marks as deleted in the database)
8. Dialog closes; Request Tickets List refreshes — deleted ticket is no longer visible
9. Success toast: "Request ticket has been deleted"

**Alternative Flows:**
- **Alt 1 — Cancel confirmation:** User clicks "Cancel" in the dialog → dialog closes, no changes made.
- **Alt 2 — Close via X:** User clicks the X button → dialog closes, no changes made.
- **Alt 3 — Escape key:** User presses Escape → dialog closes, no changes made.
- **Alt 4 — Employee attempts non-Open ticket:** Employee tries to delete their own In Progress, On Hold, Resolved, Closed, or Cancelled ticket → Delete option not visible in the gear menu.
- **Alt 5 — Employee attempts another employee's ticket:** Employee cannot see Delete for tickets belonging to other employees (gear icon actions filtered by ownership).
- **Alt 6 — API error:** Delete fails → error toast: "Failed to delete request ticket. Please try again." Dialog remains open for retry.
- **Alt 7 — Ticket already deleted:** Another admin deleted the ticket between dialog open and confirm → error toast: "Request ticket not found." → list refreshes.

**Exit Points:**
- **Success:** Ticket soft-deleted → success toast → list refreshes
- **Cancel:** Dialog closed → no changes
- **Error:** Error toast → dialog stays open or list refreshes (depending on error type)

---

## 3. Field Definitions

### Input Fields

No input fields — deletion is confirmed via a modal dialog only. No data entry is required from the user.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Delete (gear menu item) | Menu item | Visible per role and status rules (see Section 4) | Opens Confirmation Dialog | Entry point for deletion |
| Delete (confirm) | Danger/Destructive button | Shown in Confirmation Dialog; disabled while processing | Soft-deletes the request ticket | Red/danger style to signal destructive action |
| Cancel | Secondary button | Shown in Confirmation Dialog | Closes dialog without deleting | Returns user to the list with no changes |
| Close (X) | Icon button | Shown in dialog header | Dismisses dialog | Same effect as Cancel |

---

## 4. Data Display

### Confirmation Dialog Content

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Dialog title | Text | "Delete Request Ticket" | Labels the dialog action |
| Employee name | Text | "[Employee Full Name]" | Identifies whose ticket is being deleted |
| Request type | Text | "[Request Type]" | Identifies the category of the request |
| Subject | Text | "[Ticket Subject]" | Identifies the specific ticket content |
| Warning text | Static text | "Are you sure you want to delete this request ticket? This action cannot be undone." | Communicates the significance and permanence of the action |

### Dialog Layout

```
+------------------------------------------+
|  Delete Request Ticket             [X]   |
+------------------------------------------+
|                                          |
|  Are you sure you want to delete         |
|  this request ticket?                    |
|                                          |
|  Employee: [Full Name]                   |
|  Type:     [Request Type]                |
|  Subject:  [Ticket Subject]              |
|                                          |
|  This action cannot be undone.           |
|                                          |
|                 [Cancel]  [Delete]       |
+------------------------------------------+
```

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Confirmation Dialog | User selects Delete from gear menu | Modal dialog with ticket summary, Cancel + Delete buttons |
| Deleting | Delete confirmed, processing | Delete button shows spinner and is disabled; Cancel also disabled |
| Success | Ticket deleted successfully | Dialog closes; list refreshes; success toast: "Request ticket has been deleted" |
| Dismissed | User clicks Cancel, X, or presses Escape | Dialog closes; no changes; list unchanged |
| API error | Delete request fails | Error toast: "Failed to delete request ticket. Please try again."; dialog stays open |
| Already deleted | Ticket removed by another user before confirm | Error toast: "Request ticket not found."; dialog closes; list refreshes |

### Gear Icon Delete Visibility

| Status | Employee (own ticket) | Admin/Manager (any ticket) |
|--------|-----------------------|---------------------------|
| Open | Delete visible | Delete visible |
| In Progress | Delete NOT visible | Delete visible |
| On Hold | Delete NOT visible | Delete visible |
| Resolved | Delete NOT visible | Delete visible |
| Closed | Delete NOT visible | Delete visible |
| Cancelled | Delete NOT visible | Delete visible |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Gear Icon Visibility:**
- **AC-01:** Delete option is available in the gear icon dropdown based on user role and ticket status per the visibility matrix in Section 4
- **AC-02:** Employees see Delete only for their own Open tickets
- **AC-03:** Admin/managers see Delete for any ticket regardless of status
- **AC-04:** Employees do NOT see Delete for their own tickets in In Progress, On Hold, Resolved, Closed, or Cancelled status
- **AC-05:** Employees do NOT see Delete for tickets belonging to other employees

**Confirmation Dialog:**
- **AC-06:** Clicking Delete in the gear menu opens a Confirmation Dialog as a modal overlay
- **AC-07:** Dialog displays the ticket summary: employee full name, request type, and subject
- **AC-08:** Dialog includes warning text: "Are you sure you want to delete this request ticket? This action cannot be undone."
- **AC-09:** Dialog includes a Cancel button, a Delete button (danger style), and an X close button in the header
- **AC-10:** Delete button uses a danger/red style to visually signal a destructive action
- **AC-11:** Default focus is placed on the Cancel button (not Delete) when the dialog opens

**Delete Behavior:**
- **AC-12:** Confirmed deletion soft-deletes the ticket — record is preserved in the database with a deleted flag
- **AC-13:** Deleted ticket is removed from the Request Tickets List immediately after deletion
- **AC-14:** Deleted ticket is excluded from all active filters, searches, and exports
- **AC-15:** Success toast displays: "Request ticket has been deleted"
- **AC-16:** After deletion, any active search or filter state on the list is preserved (no disruptive reset)
- **AC-17:** No resource or balance adjustment is made for any ticket deletion (request tickets have no consumable resource)

**Dialog Interactions:**
- **AC-18:** Clicking Cancel closes the dialog without deleting the ticket
- **AC-19:** Clicking X closes the dialog without deleting the ticket
- **AC-20:** Pressing Escape closes the dialog (same effect as Cancel)
- **AC-21:** Delete button shows a loading spinner and is disabled while the deletion is processing
- **AC-22:** Dialog is a modal overlay — the list behind is not interactive while the dialog is open
- **AC-23:** Keyboard focus is trapped inside the dialog (Tab key cycles through dialog elements only)

**Error Handling:**
- **AC-24:** If deletion fails, error toast: "Failed to delete request ticket. Please try again." Dialog remains open.
- **AC-25:** If the ticket was already deleted by another user, error toast: "Request ticket not found." Dialog closes and list refreshes.

**Access Control:**
- **AC-26:** Delete action is visible only to users with request ticket management permission
- **AC-27:** Users with view-only permission do not see the Delete option in the gear menu

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee deletes own Open ticket | Gear → Delete on own Open ticket | Confirmation dialog → deleted; success toast; removed from list | High |
| Admin deletes an Open ticket | Gear → Delete on any Open ticket | Confirmation dialog → deleted; success toast | High |
| Admin deletes an In Progress ticket | Gear → Delete on In Progress ticket | Confirmation dialog → deleted; success toast | High |
| Admin deletes an On Hold ticket | Gear → Delete on On Hold ticket | Confirmation dialog → deleted; success toast | Medium |
| Admin deletes a Resolved ticket | Gear → Delete on Resolved ticket | Confirmation dialog → deleted; success toast | Medium |
| Admin deletes a Closed ticket | Gear → Delete on Closed ticket | Confirmation dialog → deleted; success toast | Medium |
| Admin deletes a Cancelled ticket | Gear → Delete on Cancelled ticket | Confirmation dialog → deleted; success toast | Medium |
| Employee — In Progress ticket | Employee views own In Progress ticket | Delete NOT visible in gear menu | High |
| Employee — other employee's Open ticket | Employee views another employee's ticket | Delete NOT visible in gear menu | High |
| Cancel confirmation | Click Cancel in dialog | Dialog closes; no changes | High |
| Close via X | Click X in dialog header | Dialog closes; no changes | Medium |
| Escape closes dialog | Press Escape | Dialog closes; no changes | Medium |
| Delete with active search/filter | Search active, delete a ticket | Ticket removed from list; search/filter remains in effect | Medium |
| API error on delete | Server returns error | Error toast: "Failed to delete..."; dialog stays open | Medium |
| Ticket already deleted by another user | Concurrent deletion scenario | Error toast: "Request ticket not found."; list refreshes | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with request ticket management permission can access the Delete action. Permissions are configured via EP-001 US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-02:** **Employee delete scope:** Employees can only delete their own tickets with Open status. The Delete option is hidden in the gear menu for all other employee scenarios (own non-Open tickets or other employees' tickets).
- **SR-03:** **Admin/manager delete scope:** Admin/managers can delete any request ticket regardless of current status.
- **SR-04:** Deletion is a **soft delete** — the record is preserved in the database with a deleted flag. The ticket is removed from all active views immediately.
- **SR-05:** Soft-deleted tickets are excluded from: Request Tickets List, all filter dropdowns, search results, and exports.
- **SR-06:** No resource restoration is performed on deletion. Request tickets have no consumable resource (no leave days, no balance) associated with any status. This applies to all statuses including Resolved and Closed.
- **SR-07:** Deletion requires user confirmation via the dialog — no direct single-click delete.
- **SR-08:** Default focus on the Cancel button in the dialog prevents accidental deletion via keyboard Enter.
- **SR-09:** After deletion, any active search or filter state on the Request Tickets List is preserved.
- **SR-10:** If the ticket was already deleted by another user between dialog open and confirm, the system returns a "not found" error and refreshes the list to reflect the current state.
- **SR-11:** Data visibility rules remain in effect for Delete: employees can only see Delete for their own records; the server enforces ownership check on the delete request.

**State Transitions:**
```
[Request Tickets List] → [Gear → Delete] → [Confirmation Dialog]
[Confirmation Dialog] → [Cancel / X / Escape] → [List unchanged]
[Confirmation Dialog] → [Confirm Delete] → [Deleting] → [Soft delete] → [List refreshes; success toast]
[Deleting] → [API error] → [Error toast; dialog stays open]
[Deleting] → [Already deleted] → [Error toast; dialog closes; list refreshes]
```

**Dependencies:**
- **EP-001 US-001 (Authentication):** User must be signed in
- **EP-001 US-004 (Role & Permission Management):** Controls delete access and distinguishes employee vs admin/manager permissions
- **DR-003-001-01 (Request Tickets List):** Provides the gear icon trigger and the list view that refreshes after deletion
- **DR-003-001-02 (Create Request Ticket):** Creates tickets that this feature can delete
- **DR-003-001-03 (Update Request Ticket):** Edit operations on tickets that may later be deleted

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Confirmation dialog displays a ticket summary (employee name, request type, subject) — user can confirm they are deleting the correct ticket before proceeding
- **UX-02:** Delete button uses danger/red style — visually communicates that this is a destructive, irreversible action
- **UX-03:** Default focus is on the Cancel button — prevents accidental deletion if the user presses Enter immediately after the dialog opens
- **UX-04:** Delete button shows a loading spinner while processing — prevents double-click and provides visual feedback
- **UX-05:** Pressing Escape closes the dialog — consistent keyboard behavior across all dialogs in the application
- **UX-06:** Keyboard focus is trapped inside the dialog — Tab key cycles only through dialog elements while the dialog is open
- **UX-07:** Success toast auto-dismisses after 5 seconds
- **UX-08:** Error toast persists until dismissed by the user — ensures the failure notice is not missed
- **UX-09:** Active search and filter state is preserved after deletion — no disruptive list reset interrupts the user's workflow

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Centered modal dialog ~400px wide, vertically centered on screen |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard focus trapped inside dialog (Tab cycles within dialog only)
- [x] Escape key closes dialog
- [x] Default focus on Cancel button when dialog opens
- [x] Screen reader announces dialog title and content (ARIA role="dialog")
- [x] Delete button marked as destructive via ARIA
- [x] Sufficient color contrast on danger/red button meets WCAG AA standard

**Design References:**
- No dedicated Figma design frame for the delete dialog — pattern follows the established confirmation dialog style used across the HRM platform (Cancel + Delete buttons, X close, Escape key, default focus on Cancel)
- Dialog summary format adapted from Delete Leave Request (DR-002-001-04), replacing leave dates/type with request type and subject

---

## 8. Additional Information

### Out of Scope
- Hard delete (permanent removal from the database) — soft delete only
- Restoration or undoing of soft-deleted tickets (future enhancement)
- Bulk delete of multiple tickets at once
- Auto-delete after a data retention period
- Email or in-app notification to the employee when an admin deletes their ticket
- Audit log of individual deletions visible to end users
- Mobile or tablet layout

### Open Questions
- [ ] **Notification on admin delete:** Should the submitting employee be notified when an admin deletes their ticket? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Retention policy:** How long are soft-deleted records kept in the database before permanent removal? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-003-001-01: Request Tickets List | Gear icon trigger; list refreshes after deletion |
| DR-003-001-02: Create Request Ticket | Creates tickets that this feature can delete |
| DR-003-001-03: Update Request Ticket | Edit operations on tickets; edit and delete are separate actions |
| EP-001 US-001: Authentication | User must be signed in to perform deletion |
| EP-001 US-004: Role & Permission Management | Controls delete permission and scope (employee vs admin) |

### Notes
- **Delete vs Cancel:** These are distinct actions. "Cancel" changes a ticket's status to Cancelled — the ticket remains visible in the list with a red badge and is part of the normal status lifecycle. "Delete" removes the ticket from all active views entirely via soft delete. Cancel is a status transition; Delete is a record removal.
- **No resource restoration:** Unlike leave requests where deleting an Approved request restores leave days, request tickets have no consumable resource at any status. No balance or quota adjustment is made for any ticket deletion.
- **Employee permission is intentionally limited to Open tickets.** This prevents employees from deleting tickets that have already been acted upon by management (In Progress, On Hold, Resolved) or formally closed/cancelled. The restriction ensures management actions on a ticket are always traceable.
- **Confirmation dialog includes a ticket summary** — unlike simpler entity delete confirmations (e.g., Delete Department), the request ticket dialog shows employee name, request type, and subject to help admin/managers confirm they are deleting the intended ticket.
- Pattern is consistent with the established HRM confirmation dialog design: Cancel + Delete buttons, X close button, Escape key, focus on Cancel, focus trap.

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
| 1.0 | 2026-04-03 | BA Agent | Initial draft — soft delete with role-based permissions and no resource restoration |
