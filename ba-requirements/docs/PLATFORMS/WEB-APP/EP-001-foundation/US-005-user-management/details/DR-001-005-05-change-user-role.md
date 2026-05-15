---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-05
detail_name: "Change User Role"
parent_requirement: FR-US-005-11
status: draft
version: "1.0"
created_date: "2026-03-26"
last_updated: "2026-03-26"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-001-005-01-user-list.md"
    relationship: sibling
  - path: "./DR-001-005-02-create-user.md"
    relationship: sibling
  - path: "./DR-001-005-03-user-details.md"
    relationship: sibling
  - path: "./DR-001-005-04-update-user-information.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Change User Role screen (User Details sub-page)"
    node_id: "3123:6810"
    extraction_date: "2026-03-26"
---

# Detail Requirement: Change User Role

**Detail ID:** DR-001-005-05
**Parent Requirement:** FR-US-005-11
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **change another user's system role**, so that **their access rights and permissions are updated to match their current responsibilities**.

**Purpose:** Allow administrators to reassign roles as staff responsibilities change — promotions, department transfers, or role restructuring. Changes take effect instantly without requiring the affected user to log out or take any action. The affected user's permissions refresh on their next page load.

**Target Users:**
- Any user with user management permission (excluding self-role-change)

**Key Functionality:**
- Single-field form to change a user's assigned role
- Descriptive text explaining the impact and immediacy of role changes
- Confirmation dialog before applying changes
- Self-role-change blocked to prevent accidental lockout
- Instant permission update for the affected user (refreshed on next page load)

---

## 2. User Workflow

**Entry Point:** User Details page → "Change User Role" in the left action panel.

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has user management permission (US-004)
- User is viewing another user's details (not their own — self-role-change is blocked)

**Main Flow:**
1. User is on the User Details page
2. User clicks "Change User Role" in the left action panel
3. System displays the Change User Role form with "User Account" card
4. System shows descriptive text explaining that role changes take effect immediately
5. User Role dropdown is pre-filled with the user's current role
6. User selects a new role from the dropdown (shows role name + description per option)
7. User clicks Save
8. System shows confirmation dialog: "Are you sure you want to change this user's role to [new role]? Changes take effect immediately."
9. User clicks Confirm
10. System saves the new role — changes take effect instantly
11. System shows success toast "User role updated successfully"
12. User stays on the Change User Role page with the updated role displayed

**Alternative Flows:**

- **Alt 1 — Confirmation cancelled:** User clicks Cancel in the confirmation dialog → returns to form, no changes applied.
- **Alt 2 — Same role selected:** User clicks Save without changing the role → saves silently, no confirmation dialog needed.
- **Alt 3 — Navigate away (form dirty):** User selects a different role then clicks another action or back arrow → "Discard unsaved changes?" confirmation dialog → Confirm: navigates away, changes lost / Cancel: stays on form.
- **Alt 4 — Navigate away (form clean):** User navigates away without changing role → navigates immediately, no dialog.
- **Alt 5 — Self-role-change attempt:** Admin views their own profile → "Change User Role" action is disabled or hidden in the action panel.
- **Alt 6 — Validation fails:** User clears the dropdown and clicks Save → inline error "User role is required".
- **Alt 7 — Server error:** Save fails → error toast with retry suggestion, form data preserved.

**Exit Points:**
- **Success:** Toast shown, stays on page with updated role
- **Navigate away (clean):** Direct navigation to target
- **Navigate away (dirty + confirm discard):** Navigation to target, changes lost
- **Navigate away (dirty + cancel discard):** Stay on form, changes preserved

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| User Role | Dropdown (576px, full width) | Not empty; must select from available active roles | Yes (*) | Pre-filled with current role | "Select user role" | The system role to assign to this user. Dropdown shows all active roles with role name and description per option. |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Save | Button (primary, full-width 600px) | Below the card | Always visible | If role changed → confirmation dialog → save on confirm; if same role → save silently | Submit role change |
| Confirmation dialog | Modal dialog | Center of screen | Appears when Save is clicked with a different role | Confirm → save and apply; Cancel → return to form | Safety net against accidental role changes |

**Notes:**
- No Cancel button on the form — users navigate away via left action panel or back arrow
- This is the simplest User Details sub-page — single dropdown field only

---

## 4. Data Display

### Information Shown to User

| Data Element | Data Type | Format | Business Meaning |
|-------------|-----------|--------|------------------|
| Card title | Static text | "User Account" — Geist Medium 14px | Section context |
| Descriptive text (paragraph 1) | Static text | Geist Regular 14px | "Use this option to update the user's role and permissions within the system. Selecting a new role will immediately apply the corresponding access rights and restrictions" |
| Descriptive text (paragraph 2) | Static text | Geist Regular 14px | "Please ensure the correct role is assigned, as this determines what the user can view and manage in the system. Changes take effect instantly and do not require any further action." |
| User Role dropdown | Selected option | Role name + description | Currently assigned role (pre-filled) |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads | Descriptive text + User Role dropdown pre-filled with current role |
| Loading | Data being fetched | Skeleton/loading indicator for dropdown |
| Validation error | User clears dropdown and clicks Save | Inline error below dropdown: "User role is required" |
| Confirmation dialog | User clicks Save with a different role selected | Modal: "Are you sure you want to change this user's role to [new role]? Changes take effect immediately." with Confirm + Cancel buttons |
| Saving | Request in progress after confirmation | Save button shows loading state (disabled + spinner) |
| Success | Role changed | Success toast: "User role updated successfully" |
| Self-role-change blocked | Admin views their own profile | "Change User Role" action disabled or hidden in action panel |
| Server error | Save fails | Error toast with retry suggestion |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Change User Role is accessed from the User Details left action panel
- **AC-02:** "User Account" card displays with two paragraphs of descriptive text explaining role change impact and immediacy
- **AC-03:** Single mandatory field: User Role dropdown (576px), pre-filled with the user's current role
- **AC-04:** Dropdown shows all active roles with role name and description per option

**Validation:**
- **AC-05:** User cannot save with an empty User Role selection
- **AC-06:** Validation error displays inline below the dropdown: "User role is required"

**Save Behavior:**
- **AC-07:** When a different role is selected and user clicks Save, a confirmation dialog appears: "Are you sure you want to change this user's role to [new role]? Changes take effect immediately."
- **AC-08:** Confirming the dialog saves the role change, shows toast "User role updated successfully", and stays on the page with updated role
- **AC-09:** Cancelling the dialog returns to the form with no changes applied
- **AC-10:** Save with the same role (no change) saves silently — no confirmation dialog needed
- **AC-11:** Save button shows loading state (disabled + spinner) while request is in progress
- **AC-12:** Role change takes effect instantly — the affected user's permissions update immediately in the system

**Permission Refresh:**
- **AC-13:** The affected user's permissions refresh on their next page load — not mid-action or via forced logout

**Navigation & Unsaved Changes:**
- **AC-14:** No Cancel button on the form — users navigate via left action panel or back arrow
- **AC-15:** If a different role is selected and user navigates away, "Discard unsaved changes?" confirmation dialog is shown
- **AC-16:** If role is unchanged, navigation proceeds immediately without dialog
- **AC-17:** Smart dirty check — selecting the same role as the original makes the form "clean" again

**Access Control:**
- **AC-18:** Change User Role action is visible only to users with user management permission
- **AC-19:** Admins cannot change their own role — "Change User Role" action is disabled or hidden when viewing their own profile
- **AC-20:** Direct URL access by unauthorized users redirects to an appropriate fallback

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path | Select new role → Save → Confirm | Toast, role updated, stays on page | High |
| Confirmation cancel | Select new role → Save → Cancel dialog | Returns to form, no change | High |
| Same role + Save | Click Save without changing role | Saves silently, no dialog | Medium |
| Empty role | Clear dropdown, click Save | Inline error "User role is required" | High |
| Self-role-change | Admin views own profile | Change User Role action disabled/hidden | High |
| Navigate away (dirty) | Select new role, click Overview | "Discard unsaved changes?" dialog | Medium |
| Navigate away (clean) | No change, click Overview | Navigates immediately | Medium |
| Instant effect | Change role for active user | Permissions update on their next page load | High |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |
| Role list content | Open dropdown | All active roles shown with name + description | Medium |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** User Role dropdown is pre-filled with the user's current role on load
- **SR-02:** Dropdown shows all active roles with role name and description — including the current role
- **SR-03:** Role change takes effect instantly — the affected user's permissions are updated immediately in the system
- **SR-04:** The affected user's permissions refresh on their next page load — not mid-action or via forced logout. If they are currently using the system, they continue with existing permissions until they navigate to a new page.
- **SR-05:** Self-role-change is blocked — administrators cannot change their own role to prevent accidental lockout
- **SR-06:** Only users with user management permission can access the Change User Role action
- **SR-07:** Saving with the same role (no actual change) is allowed silently — no confirmation dialog, no error
- **SR-08:** Saving with a different role requires explicit confirmation from the admin before applying
- **SR-09:** No notification is sent to the affected user when their role is changed
- **SR-10:** Audit logging will be handled by a separate logging story — this DR does not define logging behavior
- **SR-11:** Last-save-wins for concurrent editing — no conflict detection
- **SR-12:** Smart dirty check compares selected role to originally loaded role — selecting the original role makes the form "clean"

**State Transitions:**
```
[User Details] → "Change User Role" click → [Change User Role form (pre-filled)]
[Change User Role] → Save (same role) → [Change User Role (saved silently)]
[Change User Role] → Save (different role) → [Confirmation dialog]
[Confirmation dialog] → Confirm → [Change User Role (updated role + success toast)]
[Confirmation dialog] → Cancel → [Change User Role (no change)]
[Change User Role] → Navigate away (clean) → [Target page]
[Change User Role] → Navigate away (dirty) → [Discard changes dialog]
[Discard changes dialog] → Confirm discard → [Target page]
[Discard changes dialog] → Cancel → [Change User Role (preserved)]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — access control + role data source (all active roles)
- **Depends on:** DR-001-005-03 (User Details) — parent page providing the action panel and entry point

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Descriptive text prominently explains the impact and immediacy of role changes — sets user expectations before they act
- **UX-02:** Confirmation dialog for role changes provides a safety net against accidental changes, reinforcing the "instant effect" warning
- **UX-03:** Save button shows loading spinner while request is in progress, preventing double submission
- **UX-04:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-05:** Role dropdown shows name + description to help admins choose the correct role without needing to navigate to Role Management
- **UX-06:** Self-role-change prevention avoids accidental lockout scenarios
- **UX-07:** Smart dirty check — selecting the same role as the original makes the form "clean" again (consistent with Update Information)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Action panel (189px) + card (600px), side-by-side |
| Tablet (768-1024px) | Action panel collapses, card full-width |
| Mobile (<768px) | Action panel becomes top menu, card full-width |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab to dropdown and Save button, Enter to open dropdown and submit
- [x] Screen reader compatible — field label, descriptive text, confirmation dialog, toast announcements
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements

**Design References:**
- Figma: [Change User Role](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-6810) (node `3123:6810`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Update Information (DR-001-005-04) — same navigation pattern, action panel, no Cancel button

---

## 8. Additional Information

### Out of Scope
- Bulk role change (changing multiple users' roles at once)
- Role change history/audit log (separate logging story)
- Notification to the affected user when their role is changed
- Forced session refresh or logout of the affected user
- Creating or editing roles (managed in US-004 Role & Permission Management)
- Changing email, password, status, or profile information (separate User Details actions)

### Open Questions
- None remaining. All questions resolved during requirement writing session.

### Related Features
- **DR-001-005-03:** User Details — parent page providing the action panel and entry point
- **DR-001-005-04:** Update Information — sibling action, same navigation pattern
- **DR-001-004-02:** Create Role — roles created here appear in the dropdown
- **DR-001-004-03:** Edit Role — role permission changes propagate to users assigned that role
- **US-001:** Authentication — user must be signed in
- **US-004:** Role & Permission Management — access control + role data source

### Notes
- This is the simplest User Details sub-page — a single dropdown field with descriptive text. The design deliberately separates role changes from profile editing (Update Information) for audit clarity and permission granularity.
- The descriptive text in the design explicitly states that "Changes take effect instantly and do not require any further action" — this is a business decision, not a technical detail. The affected user's permissions update on their next page load.
- Self-role-change is blocked at the UI level (action disabled/hidden) and should also be enforced at the backend level to prevent API-level circumvention.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-26 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-26 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context |
