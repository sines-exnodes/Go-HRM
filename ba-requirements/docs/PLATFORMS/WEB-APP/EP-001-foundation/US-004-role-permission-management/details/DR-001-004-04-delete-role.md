---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
detail_id: DR-001-004-04
detail_name: "Delete Role"
parent_requirement: FR-US-004-06
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
  - path: "./DR-001-004-01-role-list.md"
    relationship: sibling
  - path: "./DR-001-004-02-create-role.md"
    relationship: sibling
  - path: "./DR-001-004-03-edit-role.md"
    relationship: sibling
input_sources: []
---

# Detail Requirement: Delete Role

**Detail ID:** DR-001-004-04
**Parent Requirement:** FR-US-004-06
**Story:** US-004-role-permission-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with role management permission**, I want to **delete a role that is no longer in use** so that **the role list stays clean and free of obsolete entries**.

**Purpose:** Allow authorized administrators to permanently remove unused roles from the system. To prevent access disruption, the system blocks deletion of any role that still has users assigned — the administrator must first reassign those users to a different role before the role can be deleted. This is a hard delete; the role is permanently removed, not deactivated.

**Target Users:** Any user with role management permission (not limited to a specific role such as Super Admin).

**Key Functionality:**
- Delete triggered from gear icon dropdown on the Role List
- System checks whether users are assigned to the role before proceeding
- If no users assigned → confirmation dialog → permanent deletion
- If users assigned → blocked dialog showing the assigned user count
- On success, role is permanently removed and the Role List refreshes

---

## 2. User Workflow

**Entry Point:** Role List → gear icon on the target row → select "Delete"

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has role management permission (US-004)
- The role to be deleted exists in the Role List

**Main Flow (no users assigned):**
1. User locates the role in the Role List
2. User clicks the gear icon on the role's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Delete"
5. System checks the number of users currently assigned to this role
6. System finds 0 users assigned
7. System displays a confirmation dialog: "Are you sure you want to delete role '[role name]'? This action cannot be undone."
8. User clicks **Confirm**
9. System permanently deletes the role
10. System displays success toast: "Role '[role name]' has been deleted"
11. System refreshes the Role List — the deleted role is no longer visible

**Alternative Flows:**

- **Alt 1 — Users assigned (deletion blocked):** At step 6, system finds ≥1 user assigned. System displays a blocked dialog: "Cannot delete role '[role name]'. There are [N] user(s) currently assigned to this role. Please reassign them to a different role before deleting." User clicks **OK** to dismiss and returns to the Role List.
- **Alt 2 — User cancels confirmation:** At step 8, user clicks **Cancel** in the confirmation dialog. Dialog closes and user returns to the Role List. No deletion occurs.
- **Alt 3 — Delete API fails:** At step 9, the server returns an error. System displays error toast: "Failed to delete role. Please try again." User remains on the Role List.
- **Alt 4 — Role deleted by another user:** Between step 4 and step 9, another administrator deletes the same role. The API returns a "not found" error. System displays error toast: "Role not found — it may have already been deleted." Role List refreshes automatically.

**Exit Points:**
- **Success:** Role deleted → toast "Role '[name]' has been deleted" → Role List refreshes
- **Blocked:** Dialog informs user of assigned users → user dismisses → returns to Role List
- **Cancel:** User cancels confirmation → returns to Role List
- **Error:** Error toast displayed → user retries or investigates

---

## 3. Field Definitions

### Input Fields

This feature has no input fields — it is a delete action triggered from a dropdown menu.

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Delete option | Dropdown menu item | Gear icon dropdown on Role List row | Visible only to users with role management permission | Initiates deletion check (user count) | Entry point for delete flow |
| Confirm button | Button (destructive) | Confirmation dialog — right side | Visible only when 0 users assigned | Permanently deletes the role | Confirms the irreversible action |
| Cancel button | Button (secondary) | Confirmation dialog — left side | Always visible in confirmation dialog | Closes dialog, no deletion | Allows user to abort |
| OK button | Button (primary) | Blocked dialog | Visible only when ≥1 users assigned | Dismisses the blocked dialog | Acknowledges the block reason |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Role name (in confirmation dialog) | Text | N/A — always populated | "Are you sure you want to delete role '[role name]'?" | Identifies which role will be deleted |
| Assigned user count (in blocked dialog) | Number | N/A — shown only when ≥1 | "There are [N] user(s) currently assigned to this role." | Explains why deletion is blocked |
| Success toast message | Text | N/A | "Role '[role name]' has been deleted" | Confirms successful deletion |
| Error toast message | Text | N/A | "Failed to delete role. Please try again." | Informs user of failure |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Confirmation dialog | Gear → Delete clicked AND 0 users assigned | Modal dialog with role name, warning text, Cancel and Confirm buttons |
| Blocked dialog | Gear → Delete clicked AND ≥1 users assigned | Modal dialog with role name, user count, explanation text, and OK button |
| Deleting | Confirm clicked, API request in progress | Confirm button shows loading state (disabled + spinner); Cancel button disabled |
| Success | Role deleted successfully | Toast: "Role '[role name]' has been deleted" → Role List refreshes (deleted role removed) |
| Error | Delete API returns error | Error toast: "Failed to delete role. Please try again." User remains on Role List |
| Not found | Role no longer exists (deleted by another user) | Error toast: "Role not found — it may have already been deleted." Role List refreshes |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Deletion Blocking:**
- **AC-01:** System checks the number of users assigned to a role before displaying the confirmation dialog
- **AC-02:** If ≥1 user is assigned, system displays a blocked dialog with the message "Cannot delete role '[role name]'. There are [N] user(s) currently assigned to this role. Please reassign them to a different role before deleting."
- **AC-03:** Blocked dialog has an OK button that dismisses the dialog and returns to the Role List
- **AC-04:** Deletion blocking is enforced server-side — even if the client bypasses the UI check, the API rejects the request

**Confirmation Flow:**
- **AC-05:** If 0 users are assigned, system displays a confirmation dialog: "Are you sure you want to delete role '[role name]'? This action cannot be undone."
- **AC-06:** Confirmation dialog has Cancel (left, secondary) and Confirm (right, destructive style) buttons
- **AC-07:** Clicking Cancel closes the dialog without deleting — user returns to Role List
- **AC-08:** Clicking Confirm permanently deletes the role (hard delete)

**Success Behavior:**
- **AC-09:** On successful deletion, system displays success toast: "Role '[role name]' has been deleted"
- **AC-10:** Role List refreshes and the deleted role is no longer visible
- **AC-11:** Pagination adjusts if the deleted role causes a page to become empty (e.g., user is moved to previous page)

**Error Handling:**
- **AC-12:** If the delete API fails, system displays error toast: "Failed to delete role. Please try again."
- **AC-13:** If the role no longer exists (deleted by another user concurrently), system displays: "Role not found — it may have already been deleted." and refreshes the list

**Access Control:**
- **AC-14:** Delete option in the gear icon dropdown is visible only to users with role management permission
- **AC-15:** Users without role management permission cannot trigger deletion, even via direct API call

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — delete unused role | Role with 0 users → Confirm | Role deleted, toast shown, list refreshed | High |
| Blocked — role has users | Role with 3 users → Delete | Blocked dialog: "3 user(s) currently assigned" | High |
| Cancel confirmation | Role with 0 users → Cancel | Dialog closes, role still exists | High |
| Server-side block enforcement | API call to delete role with assigned users | API returns 409 Conflict / rejection | High |
| Concurrent deletion | Two admins delete same role simultaneously | Second admin sees "Role not found" error | Medium |
| Delete API failure | Server error during deletion | Error toast: "Failed to delete role. Please try again." | Medium |
| Pagination adjustment | Delete last role on page 3 (only item) | User moved to page 2 | Low |
| Unauthorized user | User without permission attempts delete | Delete option not visible; API rejects if called directly | High |
| Delete role then search | Delete role "Manager", search "Manager" | No results — role no longer exists | Low |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Deletion is blocked if ≥1 user is assigned to the role. The check is performed server-side — the API must reject deletion requests for roles with assigned users, returning the current user count (BR-US-004-03).
- **SR-02:** Deletion is a hard delete — the role record is permanently removed from the system. There is no soft-delete, archive, or inactive status.
- **SR-03:** Only users with role management permission can trigger the delete action. The gear icon dropdown hides the Delete option for unauthorized users. The API also enforces this check.
- **SR-04:** System logs role deletion events — records the deleting user, the deleted role name, the deleted role's permissions at time of deletion, and timestamp.
- **SR-05:** If the role is deleted while another administrator has the Edit Role form open for the same role, the Edit form's Save action should return an appropriate error (role not found).
- **SR-06:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Role List] → [Gear → Delete] → [System checks user count]
[User count = 0] → [Confirmation Dialog]
[User count ≥ 1] → [Blocked Dialog] → [OK] → [Role List]
[Confirmation Dialog] → Cancel → [Role List]
[Confirmation Dialog] → Confirm → [Deleting] → [Success Toast] → [Role List (refreshed)]
[Confirmation Dialog] → Confirm → [Deleting] → [API Error] → [Error Toast] → [Role List]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 permission data — role management permission must exist to see the Delete option
- **Depends on:** User-to-role assignment data — system needs the current assigned user count to determine if deletion is allowed
- **Consumed by:** Role List (DR-001-004-01) — list refreshes after deletion
- **Consumed by:** Edit Role (DR-001-004-03) — if a role is deleted while Edit is open, Edit should handle the missing role gracefully

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Confirmation dialog uses destructive styling for the Confirm button (red background) to clearly signal an irreversible action
- **UX-02:** Blocked dialog clearly explains why deletion is not possible and what the user needs to do (reassign users) — actionable guidance, not just an error
- **UX-03:** Success toast auto-dismisses after 5 seconds (with manual close option), consistent with Create and Edit role toasts
- **UX-04:** Confirm button shows loading spinner while the delete request is in progress, preventing double-click deletion attempts
- **UX-05:** Role name is displayed in both the confirmation and blocked dialogs so the user is certain which role they are acting on
- **UX-06:** After deletion, the Role List maintains the user's current search query and pagination position (adjusted if needed) — no loss of browsing context

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Centered modal dialog, 400–480px wide |
| Tablet (768–1024px) | Centered modal dialog, same width with slight padding adjustment |
| Mobile (<768px) | Full-width modal dialog with padding, stacked buttons (Cancel above Confirm) |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab between Cancel and Confirm buttons; Escape closes dialog (same as Cancel)
- [x] Screen reader compatible — dialog announced as modal, role name and warning text read aloud
- [x] Sufficient color contrast — destructive button meets WCAG 2.1 AA contrast ratio
- [x] Focus indicators visible — clear focus ring on dialog buttons
- [x] Focus trapped within dialog — Tab does not escape to background content while dialog is open
- [x] Focus returns to the gear icon (or next row) after dialog closes

**Design References:**
- No dedicated Figma design for Delete Role — follows standard confirmation/blocked dialog pattern used across the platform
- Destructive action styling: red Confirm button (consistent with platform conventions)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]

---

## 8. Additional Information

### Out of Scope
- Soft delete / deactivation of roles (roles are hard-deleted only)
- Bulk deletion of multiple roles at once
- Undo deletion (once confirmed, the role is permanently removed)
- Viewing which users are assigned to the role from the delete dialog (only the count is shown)
- Reassigning users to a different role from within the delete flow (user must go to User Management first)
- Archiving roles for historical reference

### Open Questions
- [ ] **Blocked dialog — link to assigned users:** Should the blocked dialog include a link or button to navigate to the User List filtered by the assigned role, so the administrator can quickly reassign users? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Gear icon dropdown design:** The gear icon dropdown (with Edit and Delete options) has not been designed in Figma. Confirm visual spec. — **Owner:** Design Team — **Status:** Pending

### Related Features
- **DR-001-004-01:** Role List — Delete is triggered from the gear icon; on success, the list refreshes with the role removed
- **DR-001-004-02:** Create Role — Unrelated workflow, but a deleted role name becomes available for reuse
- **DR-001-004-03:** Edit Role — Accessible from the same gear icon dropdown; if a role is deleted while Edit is open, Edit should handle gracefully
- **US-001:** Authentication — user must be signed in
- **User Management (EP-001):** Users must be reassigned before a role can be deleted

### Notes
- Delete is the simplest of the four role management actions (List, Create, Edit, Delete). It has no form, no input fields — just a confirmation or blocking dialog.
- The business rule BR-US-004-03 (block deletion if users assigned) is the critical safeguard. It must be enforced server-side, not just in the UI, to prevent data integrity issues from direct API calls.
- A deleted role name becomes immediately available for reuse — a new role can be created with the same name after deletion.
- Unlike Edit Role, there is no "dirty form" detection needed — the user either confirms or cancels; there is no intermediate state to protect.

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
| 1.0 | 2026-03-24 | BA Agent | Initial draft — deletion flow with blocking rule (BR-US-004-03) and confirmation dialog |
