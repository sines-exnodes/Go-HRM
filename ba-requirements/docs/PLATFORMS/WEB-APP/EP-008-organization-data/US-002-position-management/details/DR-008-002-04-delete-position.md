---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-002
story_name: "Position Management"
detail_id: DR-008-002-04
detail_name: "Delete Position"
parent_requirement: FR-US-002-07
status: draft
version: "1.0"
created_date: 2026-03-05
last_updated: 2026-03-05
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-008-002-01-position-list.md"
    relationship: sibling
  - path: "./DR-008-002-02-create-position.md"
    relationship: sibling
  - path: "./DR-008-002-03-edit-position.md"
    relationship: sibling
---

# Detail Requirement: Delete Position

**Detail ID:** DR-008-002-04
**Parent Requirement:** FR-US-002-07
**Story:** US-002-position-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with position management permission**, I want to **delete a position that is no longer in use**, so that **the position list stays accurate and free of obsolete entries**.

**Purpose:** Allow authorized administrators to permanently remove positions that are no longer needed. This keeps the position catalog clean and prevents outdated positions from appearing in HR module selections (e.g., employee profile position dropdowns). Deletion is hard delete — permanent and irreversible.

**Target Users:** Any role with position management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Triggered from the gear icon dropdown on the Position List — no separate page
- System checks employee count before any dialog is shown
- If employees are assigned: blocked immediately with an informational message
- If no employees assigned: confirmation dialog shown before permanent deletion

---

## 2. User Workflow

**Entry Point:** Position List → gear icon on the target position row → select "Delete"

**Preconditions:**
- User is signed in
- User's role has position management permission (configured via US-004)
- The position to be deleted exists in the list

**Main Flow (Happy Path — No Employees Assigned):**
1. User locates the position in the Position List
2. User clicks the gear icon on the position's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Delete"
5. System checks the assigned employee count for this position in real time
6. Employee count = 0 → system shows Confirmation Dialog with position name and irreversibility warning
7. User clicks "Delete" (Confirm)
8. System shows loading state on the Delete button; buttons are disabled
9. System permanently deletes the position
10. Dialog closes; position is no longer visible in the Position List

**Alternative Flows:**
- **Alt 1 — Employees Assigned (Blocked):** At step 5, employee count ≥ 1 → system shows Blocked Dialog displaying the position name and exact employee count. No Confirmation Dialog is shown. User must close the dialog and reassign all employees before attempting to delete again.
- **Alt 2 — Cancel Confirmation:** At step 7, user clicks "Cancel" → Confirmation Dialog closes. No changes made. Position remains in the list.
- **Alt 3 — Close via X:** At any step with a dialog open, user clicks the X button → dialog closes. No changes made.
- **Alt 4 — Race Condition:** Between step 6 and step 9, an employee is assigned to the position → system blocks the deletion at execution, closes the Confirmation Dialog, and shows the Blocked Dialog with the updated employee count.
- **Alt 5 — Last Position:** If the deleted position is the last one in the list, deletion follows the same happy path; the empty state is displayed after deletion.

**Exit Points:**
- **Success:** Position permanently deleted → list refreshes, position removed
- **Blocked:** Blocked Dialog shown, no changes made, user closes dialog
- **Cancel:** Confirmation Dialog closed, no changes made

---

## 3. Field Definitions

### Input Fields

No input fields — this feature does not use a form. The delete action is confirmed via a modal dialog only.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Delete (Confirm) | Danger/Destructive button | Shown in Confirmation Dialog; disabled while processing | Permanently deletes the position | Styled in red/danger to signal destructive action |
| Cancel | Secondary button | Shown in Confirmation Dialog | Closes dialog without deleting | Returns user to the list with no changes |
| Close (X) | Icon button | Shown in both Confirmation and Blocked dialogs | Dismisses the dialog | Same effect as Cancel — no changes made |
| Close (Blocked) | Secondary button | Shown in Blocked Dialog only | Dismisses the Blocked Dialog | Returns user to the list with no changes |

---

## 4. Data Display

### Information Shown to User

**Confirmation Dialog (employee count = 0):**

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Position Name | Text | Quoted: "[Position Name]" | Identifies which position will be deleted |
| Irreversibility warning | Static text | "This action cannot be undone." | Communicates permanence of the action |
| Dialog title | Text | "Delete Position" | Labels the dialog |

**Blocked Dialog (employee count ≥ 1):**

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Position Name | Text | Quoted: "[Position Name]" | Identifies the position that cannot be deleted |
| Employee count | Number | "[X] employees are assigned to this position" | Shows exactly why deletion is blocked |
| Instruction | Static text | "Reassign all employees before deleting this position." | Tells user what to do next |
| Dialog title | Text | "Cannot Delete Position" | Labels the dialog |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Confirmation Dialog | Position has 0 assigned employees | Modal overlay with position name, warning, Cancel + Delete buttons |
| Blocked Dialog | Position has ≥1 assigned employee | Modal overlay with position name, employee count, Close button |
| Deleting | After Delete clicked, system processing | Delete button shows loading spinner; both buttons disabled |
| Success | Position deleted | Dialog closes; position no longer visible in list |
| Dismissed | User clicks Cancel, X, or Close | Dialog closes; list unchanged |

### Dialog Layouts

**Confirmation Dialog:**
```
┌─────────────────────────────────────────┐
│  Delete Position                [X]     │
├─────────────────────────────────────────┤
│                                         │
│  Are you sure you want to delete        │
│  "[Position Name]"?                     │
│                                         │
│  This action cannot be undone.          │
│                                         │
│              [Cancel]  [Delete]         │
└─────────────────────────────────────────┘
```

**Blocked Dialog:**
```
┌─────────────────────────────────────────┐
│  Cannot Delete Position         [X]     │
├─────────────────────────────────────────┤
│                                         │
│  [X] employees are assigned to          │
│  "[Position Name]".                     │
│                                         │
│  Reassign all employees before          │
│  deleting this position.                │
│                                         │
│                           [Close]       │
└─────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** Delete option is available in the gear icon dropdown for each position row
- **AC-02:** Delete option is visible only to users with position management permission
- **AC-03:** When a position with ≥1 assigned employee is selected for deletion, the Blocked Dialog is shown immediately — no Confirmation Dialog is shown
- **AC-04:** Blocked Dialog displays the exact number of employees currently assigned to the position
- **AC-05:** Blocked Dialog displays the position name
- **AC-06:** Blocked Dialog instructs the user to reassign all employees before deleting
- **AC-07:** When a position with 0 assigned employees is selected for deletion, the Confirmation Dialog is shown
- **AC-08:** Confirmation Dialog displays the position name to be deleted
- **AC-09:** Confirmation Dialog includes the warning "This action cannot be undone"
- **AC-10:** Clicking Cancel in the Confirmation Dialog closes it without deleting the position
- **AC-11:** Clicking the X button in either dialog closes it without making any changes
- **AC-12:** Confirmed deletion permanently removes the position from the Position List
- **AC-13:** Deleted position no longer appears in HR module position selectors
- **AC-14:** Delete button in the Confirmation Dialog shows a loading state while the system processes the deletion (prevents double-click)
- **AC-15:** After deleting the last position in the list, the empty state is displayed
- **AC-16:** Employee count in the Blocked Dialog reflects the count at the time of the delete attempt (real time, not cached)
- **AC-17:** Both dialogs are modal overlays — the Position List behind them is not interactive while a dialog is open
- **AC-18:** After successful deletion, any active search filter applied to the list remains in effect

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — position with 0 employees | Click Delete on empty position | Confirmation Dialog shown with position name | High |
| Confirm deletion | Confirmation Dialog → click Delete | Position permanently deleted; removed from list | High |
| Cancel deletion | Confirmation Dialog → click Cancel | Dialog closes; position still in list | High |
| Dismiss with X | Any dialog → click X | Dialog closes; no changes | High |
| Blocked — 1 employee | Click Delete on position with 1 employee | Blocked Dialog shows "1 employee assigned" | High |
| Blocked — multiple employees | Click Delete on position with 5 employees | Blocked Dialog shows "5 employees assigned" | High |
| Unauthorized user | User without management permission | Gear icon not visible; no Delete option | High |
| Delete last position | Delete the only position in the list | Position removed; empty state shown | Medium |
| Delete with active search | Search active, delete filtered position | Position removed; search filter remains applied | Medium |
| Race condition | Employee assigned between dialog open and Confirm | System blocks at execution; Blocked Dialog shown | Medium |
| Double-click Delete | Click Delete twice rapidly | Only one deletion processed; second click ignored (loading state) | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with position management permission can access the Delete action. The gear icon with Delete option is hidden for users without this permission.
- **SR-02:** Employee count is checked in real time at the moment the Delete action is triggered — not from a cached value from when the list was loaded.
- **SR-03:** If employee count ≥ 1 at the time of the check, deletion is blocked server-side. The Blocked Dialog is shown and no deletion is attempted.
- **SR-04:** Confirmed deletion is permanent and irreversible. There is no soft delete, no recycle bin, and no undo mechanism.
- **SR-05:** After deletion, the position is removed from all HR module position selection lists immediately — no additional sync required.
- **SR-06:** If an employee is assigned to the position between the time the Confirmation Dialog opens and the time the user clicks Confirm (race condition), the system blocks the deletion at the point of execution and shows the Blocked Dialog with the updated employee count.
- **SR-07:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Position List] → [Gear → Delete] → [Employee Count Check (real time)]
[Check: count ≥ 1] → [Blocked Dialog] → [User Closes (X or Close)] → [Position List]
[Check: count = 0] → [Confirmation Dialog] → [User Cancels (Cancel or X)] → [Position List]
[Confirmation Dialog] → [User Confirms Delete] → [Deleting state] → [Position List (position removed)]
[Deleting state: race condition detected] → [Blocked Dialog] → [User Closes] → [Position List]
```

**Dependencies:**
- **US-004 (Role & Permission Management):** Controls which roles can access the Delete action and gear icon
- **EP-002 (Employee Management):** Employee count is sourced from employee records assigned to this position

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Delete button in the Confirmation Dialog uses a danger/destructive style (red background) to visually signal the irreversible nature of the action
- **UX-02:** Position name is displayed in quotes within both dialogs to make it unambiguous which position is affected
- **UX-03:** Delete button shows a loading spinner after click and is disabled until the system responds — prevents accidental double submission
- **UX-04:** Keyboard focus is trapped inside the dialog while it is open — Tab key cycles only through dialog elements
- **UX-05:** Pressing Escape closes either dialog (same behavior as clicking Cancel or X)
- **UX-06:** When the Confirmation Dialog opens, default focus is placed on the Cancel button (not Delete) — prevents accidental deletion if user presses Enter immediately
- **UX-07:** The Blocked Dialog contains only a Close button — no Delete option — making it impossible to override the blocking rule from this dialog
- **UX-08:** Employee count in the Blocked Dialog is shown as a precise number (e.g., "3 employees") — not a range or approximation

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Centered modal dialog, fixed width ~400px, vertically centered |
| Tablet (768–1024px) | Centered modal dialog, full-width with horizontal padding |
| Mobile (<768px) | Bottom sheet or full-width modal anchored to bottom of screen |

**Accessibility Requirements:**
- [ ] Keyboard focus trapped inside dialog while open (Tab cycles within dialog only)
- [ ] Escape key closes dialog (same as Cancel/X)
- [ ] Default focus on Cancel button when Confirmation Dialog opens
- [ ] Screen reader announces dialog title and content on open (ARIA role="dialog")
- [ ] Delete button marked as destructive via ARIA (aria-describedby pointing to warning text)

**Design References:**
- Figma design for Delete Position dialogs is **pending delivery** from the Design Team
- Pattern reference: Confirmation and Blocked dialogs should follow the existing HRM modal/dialog design system component (consistent with Department Delete dialogs from US-001)

---

## 8. Additional Information

### Out of Scope
- Soft delete or deactivation (hard delete only — confirmed by Product Owner)
- Bulk deletion of multiple positions at once
- Undoing or restoring a deleted position
- Auto-reassigning employees when a position is deleted (user must manually reassign first)
- Audit log or history of deleted positions

### Open Questions
- None — all requirements confirmed consistent with Department Management pattern.

### Related Features
- **DR-008-002-01** (Position List) — Delete is triggered from the gear icon on the list; on success, the list refreshes
- **DR-008-002-02** (Create Position) — Completes the CRUD lifecycle for position management
- **DR-008-002-03** (Edit Position) — Shares the same gear icon dropdown as Delete
- **US-004** (Role & Permission Management) — Controls Delete action and gear icon visibility
- **EP-002** (Employee Management) — Employee count determines whether deletion is blocked or allowed

### Notes
- This is the only position management feature that does **not** navigate to a separate page. The entire flow is handled via modal dialogs overlaid on the Position List.
- The employee count check is performed **before** the confirmation dialog is shown (not after). This is intentional — to avoid showing a confirmation dialog for an action that will be immediately blocked.
- The "race condition" scenario (Alt 4) is a server-side protection rule (SR-06) ensuring data integrity even if the check and the delete happen slightly apart in time.

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
| 1.0 | 2026-03-05 | BA Agent | Initial draft |
