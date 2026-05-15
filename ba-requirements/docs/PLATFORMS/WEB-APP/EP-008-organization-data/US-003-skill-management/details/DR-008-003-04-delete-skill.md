---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
detail_id: DR-008-003-04
detail_name: "Delete Skill"
parent_requirement: FR-US-003-06
status: draft
version: "1.0"
created_date: "2026-03-24"
last_updated: "2026-03-24"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../TODO.yaml"
    relationship: sibling
  - path: "./DR-008-003-01-skill-list.md"
    relationship: sibling
  - path: "./DR-008-003-02-create-skill.md"
    relationship: sibling
  - path: "./DR-008-003-03-edit-skill.md"
    relationship: sibling
input_sources: []
---

# Detail Requirement: Delete Skill

**Detail ID:** DR-008-003-04
**Parent Requirement:** FR-US-003-06
**Story:** US-003-skill-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with organization data management permission**, I want to **delete a skill that is no longer in use**, so that **the skill catalog stays accurate and free of obsolete entries**.

**Purpose:** Allow authorized administrators to permanently remove skills that are no longer relevant. This keeps the skill catalog clean and prevents outdated skills from appearing in HR module selections (e.g., employee competency assignment). Deletion is hard delete — permanent and irreversible.

**Target Users:** Any user with org data management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Triggered from the gear icon dropdown on the Skill List — no separate page
- System checks employee count before any dialog is shown
- If employees are assigned: blocked immediately with an informational message
- If no employees assigned: confirmation dialog shown before permanent deletion

---

## 2. User Workflow

**Entry Point:** Skill List → gear icon on the target skill row → select "Delete"

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has org data management permission (US-004)
- The skill to be deleted exists in the list

**Main Flow (Happy Path — No Employees Assigned):**
1. User locates the skill in the Skill List
2. User clicks the gear icon on the skill's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Delete"
5. System checks the assigned employee count for this skill in real time
6. Employee count = 0 → system shows Confirmation Dialog with skill name and irreversibility warning
7. User clicks "Delete" (Confirm)
8. System shows loading state on the Delete button; buttons are disabled
9. System permanently deletes the skill (including its icon file if one exists)
10. Dialog closes; skill is no longer visible in the Skill List

**Alternative Flows:**
- **Alt 1 — Employees Assigned (Blocked):** At step 5, employee count ≥ 1 → system shows Blocked Dialog displaying the skill name and exact employee count. No Confirmation Dialog is shown. User must close the dialog and reassign all employees before attempting to delete again.
- **Alt 2 — Cancel Confirmation:** At step 7, user clicks "Cancel" → Confirmation Dialog closes. No changes made. Skill remains in the list.
- **Alt 3 — Close via X:** At any step with a dialog open, user clicks the X button → dialog closes. No changes made.
- **Alt 4 — Race Condition:** Between step 6 and step 9, an employee is assigned to the skill → system blocks the deletion at execution, closes the Confirmation Dialog, and shows the Blocked Dialog with the updated employee count.
- **Alt 5 — Last Skill:** If the deleted skill is the last one in the list, deletion follows the same happy path; the empty state is displayed after deletion.

**Exit Points:**
- **Success:** Skill permanently deleted → list refreshes, skill removed
- **Blocked:** Blocked Dialog shown, no changes made, user closes dialog
- **Cancel:** Confirmation Dialog closed, no changes made

---

## 3. Field Definitions

### Input Fields

No input fields — this feature does not use a form. The delete action is confirmed via a modal dialog only.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Delete (Confirm) | Danger/Destructive button | Shown in Confirmation Dialog; disabled while processing | Permanently deletes the skill | Styled in red/danger to signal destructive action |
| Cancel | Secondary button | Shown in Confirmation Dialog | Closes dialog without deleting | Returns user to the list with no changes |
| Close (X) | Icon button | Shown in both Confirmation and Blocked dialogs | Dismisses the dialog | Same effect as Cancel — no changes made |
| Close (Blocked) | Secondary button | Shown in Blocked Dialog only | Dismisses the Blocked Dialog | Returns user to the list with no changes |

---

## 4. Data Display

### Information Shown to User

**Confirmation Dialog (employee count = 0):**

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Skill Name | Text | Quoted: "[Skill Name]" | Identifies which skill will be deleted |
| Irreversibility warning | Static text | "This action cannot be undone." | Communicates permanence of the action |
| Dialog title | Text | "Delete Skill" | Labels the dialog |

**Blocked Dialog (employee count ≥ 1):**

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Skill Name | Text | Quoted: "[Skill Name]" | Identifies the skill that cannot be deleted |
| Employee count | Number | "[X] employees are assigned to this skill" | Shows exactly why deletion is blocked |
| Instruction | Static text | "Reassign all employees before deleting this skill." | Tells user what to do next |
| Dialog title | Text | "Cannot Delete Skill" | Labels the dialog |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Confirmation Dialog | Skill has 0 assigned employees | Modal overlay with skill name, warning, Cancel + Delete buttons |
| Blocked Dialog | Skill has ≥1 assigned employee | Modal overlay with skill name, employee count, Close button |
| Deleting | After Delete clicked, system processing | Delete button shows loading spinner; both buttons disabled |
| Success | Skill deleted | Dialog closes; skill no longer visible in list |
| Dismissed | User clicks Cancel, X, or Close | Dialog closes; list unchanged |

### Dialog Layouts

**Confirmation Dialog:**
```
┌─────────────────────────────────────────┐
│  Delete Skill                   [X]     │
├─────────────────────────────────────────┤
│                                         │
│  Are you sure you want to delete        │
│  "[Skill Name]"?                        │
│                                         │
│  This action cannot be undone.          │
│                                         │
│              [Cancel]  [Delete]         │
└─────────────────────────────────────────┘
```

**Blocked Dialog:**
```
┌─────────────────────────────────────────┐
│  Cannot Delete Skill            [X]     │
├─────────────────────────────────────────┤
│                                         │
│  [X] employees are assigned to          │
│  "[Skill Name]".                        │
│                                         │
│  Reassign all employees before          │
│  deleting this skill.                   │
│                                         │
│                           [Close]       │
└─────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** Delete option is available in the gear icon dropdown for each skill row
- **AC-02:** Delete option is visible only to users with org data management permission
- **AC-03:** When a skill with ≥1 assigned employee is selected for deletion, the Blocked Dialog is shown immediately — no Confirmation Dialog is shown
- **AC-04:** Blocked Dialog displays the exact number of employees currently assigned to the skill
- **AC-05:** Blocked Dialog displays the skill name
- **AC-06:** Blocked Dialog instructs the user to reassign all employees before deleting
- **AC-07:** When a skill with 0 assigned employees is selected for deletion, the Confirmation Dialog is shown
- **AC-08:** Confirmation Dialog displays the skill name to be deleted
- **AC-09:** Confirmation Dialog includes the warning "This action cannot be undone"
- **AC-10:** Clicking Cancel in the Confirmation Dialog closes it without deleting the skill
- **AC-11:** Clicking the X button in either dialog closes it without making any changes
- **AC-12:** Confirmed deletion permanently removes the skill from the Skill List
- **AC-13:** Deleted skill no longer appears in HR module skill selectors (e.g., employee competency assignment)
- **AC-14:** Deleted skill's icon file is also removed from storage
- **AC-15:** Delete button in the Confirmation Dialog shows a loading state while the system processes the deletion (prevents double-click)
- **AC-16:** After deleting the last skill in the list, the empty state is displayed
- **AC-17:** Employee count in the Blocked Dialog reflects the count at the time of the delete attempt (real time, not cached)
- **AC-18:** Both dialogs are modal overlays — the Skill List behind them is not interactive while a dialog is open
- **AC-19:** After successful deletion, any active search filter applied to the list remains in effect

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — skill with 0 employees | Click Delete on unused skill | Confirmation Dialog shown with skill name | High |
| Confirm deletion | Confirmation Dialog → click Delete | Skill permanently deleted; removed from list | High |
| Cancel deletion | Confirmation Dialog → click Cancel | Dialog closes; skill still in list | High |
| Dismiss with X | Any dialog → click X | Dialog closes; no changes | High |
| Blocked — 1 employee | Click Delete on skill with 1 employee | Blocked Dialog shows "1 employee assigned" | High |
| Blocked — multiple employees | Click Delete on skill with 5 employees | Blocked Dialog shows "5 employees assigned" | High |
| Unauthorized user | User without management permission | Gear icon not visible; no Delete option | High |
| Delete last skill | Delete the only skill in the list | Skill removed; empty state shown | Medium |
| Delete with active search | Search active, delete filtered skill | Skill removed; search filter remains applied | Medium |
| Race condition | Employee assigned between dialog open and Confirm | System blocks at execution; Blocked Dialog shown | Medium |
| Double-click Delete | Click Delete twice rapidly | Only one deletion processed; second click ignored (loading state) | Medium |
| Icon cleanup | Delete skill that has a custom icon | Skill + icon file both removed | Medium |

---

## 6. System Rules

**Business Logic:**

- **SR-01:** Only users with org data management permission can access the Delete action. The gear icon with Delete option is hidden for users without this permission.
- **SR-02:** Employee count is checked in real time at the moment the Delete action is triggered — not from a cached value from when the list was loaded.
- **SR-03:** If employee count ≥ 1 at the time of the check, deletion is blocked server-side. The Blocked Dialog is shown and no deletion is attempted.
- **SR-04:** Confirmed deletion is permanent and irreversible. There is no soft delete, no recycle bin, and no undo mechanism.
- **SR-05:** After deletion, the skill is removed from all HR module skill selection lists immediately — no additional sync required.
- **SR-06:** If an employee is assigned to the skill between the time the Confirmation Dialog opens and the time the user clicks Confirm (race condition), the system blocks the deletion at the point of execution and shows the Blocked Dialog with the updated employee count.
- **SR-07:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-08:** When a skill is deleted, its associated icon file (if any) is also permanently removed from storage.

**State Transitions:**
```
[Skill List] → [Gear → Delete] → [Employee Count Check (real time)]
[Check: count ≥ 1] → [Blocked Dialog] → [User Closes (X or Close)] → [Skill List]
[Check: count = 0] → [Confirmation Dialog] → [User Cancels (Cancel or X)] → [Skill List]
[Confirmation Dialog] → [User Confirms Delete] → [Deleting state] → [Skill List (skill removed)]
[Deleting state: race condition detected] → [Blocked Dialog] → [User Closes] → [Skill List]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — org data management permission required
- **Consumed by:** EP-002 (Employee Management) — employee count is sourced from employee-skill assignments

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Delete button in the Confirmation Dialog uses a danger/destructive style (red background) to visually signal the irreversible nature of the action
- **UX-02:** Skill name is displayed in quotes within both dialogs to make it unambiguous which skill is affected
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
- [x] Keyboard focus trapped inside dialog while open (Tab cycles within dialog only)
- [x] Escape key closes dialog (same as Cancel/X)
- [x] Default focus on Cancel button when Confirmation Dialog opens
- [x] Screen reader announces dialog title and content on open (ARIA role="dialog")
- [x] Delete button marked as destructive via ARIA (aria-describedby pointing to warning text)

**Design References:**
- Figma design for Delete Skill dialogs is **pending delivery** from the Design Team
- Pattern reference: Delete Department (DR-008-001-04), Delete Position (DR-008-002-04) — same two-dialog pattern with employee count check
- Dialogs should follow the existing HRM modal/dialog design system component

---

## 8. Additional Information

### Out of Scope
- Soft delete or deactivation (hard delete only — confirmed by BR-US-003-03)
- Bulk deletion of multiple skills at once
- Undoing or restoring a deleted skill
- Auto-reassigning employees when a skill is deleted (user must manually reassign first)
- Audit log or history of deleted skills
- Deleting a skill's icon independently (icon is only removed when the entire skill is deleted)

### Open Questions
- None — all requirements follow established Delete Department/Position pattern and BR-US-003-03.

### Related Features
- **DR-008-003-01** (Skill List) — Delete is triggered from the gear icon on the list; on success, the list refreshes
- **DR-008-003-02** (Create Skill) — Completes the CRUD lifecycle for skill management
- **DR-008-003-03** (Edit Skill) — Shares the same gear icon dropdown as Delete
- **DR-008-001-04** (Delete Department) — Pattern reference for the two-dialog delete flow
- **DR-008-002-04** (Delete Position) — Pattern reference for the two-dialog delete flow
- **US-004** (Role & Permission Management) — Controls Delete action and gear icon visibility
- **EP-002** (Employee Management) — Employee count determines whether deletion is blocked or allowed

### Notes
- This is the only skill management feature that does **not** navigate to a separate page. The entire flow is handled via modal dialogs overlaid on the Skill List.
- The employee count check is performed **before** the confirmation dialog is shown (not after). This is intentional — to avoid showing a confirmation dialog for an action that will be immediately blocked.
- The "race condition" scenario (Alt 4) is a server-side protection rule (SR-06) ensuring data integrity even if the check and the delete happen slightly apart in time.
- When a skill is deleted, its icon file is also cleaned up from storage (SR-08). This is a skill-specific addition — Department and Position do not have icon files.
- This delete pattern is **consistent across all organization data modules**: Department (DR-008-001-04), Position (DR-008-002-04), and Skill (this document) all use the same two-dialog flow with employee count pre-check.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-24 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-24 | BA Agent | Initial draft — consistent with Delete Department/Position pattern; adds icon cleanup (SR-08) |
