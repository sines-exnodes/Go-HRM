---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-09
detail_name: "Delete User Account"
parent_requirement: FR-US-005-09
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
  - path: "./DR-001-005-01-user-list.md"
    relationship: sibling
  - path: "./DR-001-005-02-create-user.md"
    relationship: sibling
  - path: "./DR-001-005-03-user-details.md"
    relationship: sibling
  - path: "./DR-001-005-08-activate-deactivate-user.md"
    relationship: sibling
---

# Detail Requirement: Delete User Account

**Detail ID:** DR-001-005-09
**Parent Requirement:** FR-US-005-09
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **permanently delete a user account from the system**, so that **obsolete or mistakenly created accounts are removed, keeping the user registry clean and accurate**.

**Purpose:** Allow administrators to remove a user account from all active views. This is a **soft delete** — the user record is retained in the database with a "deleted" status flag, but hidden from all active lists, selectors, and assignment counts. Unlike Activate/Deactivate (which blocks login but keeps the user visible), deletion removes the user from all active views entirely. To prevent accidental or unauthorized deletion, the administrator must confirm by entering their own password before the action executes.

**Target Users:** Any role with user management permission (configured via US-004). Users with view-only permission cannot see this action button.

**Key Functionality:**
- Password confirmation required — administrator must enter their own current password to authorize the deletion
- Soft delete — user data preserved in database with "deleted" status flag
- User hidden from all active views (User List, selectors, filters, assignment counts)
- Login blocked and active sessions invalidated immediately
- No undo mechanism in this release — restoration is a future enhancement
- Displayed inline within User Details page (same two-panel layout as other actions)

---

## 2. User Workflow

**Entry Point:** User Details page → left panel → click "Delete User" button

**Preconditions:**
- User is signed in (US-001)
- User has user management permission (US-004)
- The target user exists and User Details page is loaded

**Main Flow:**
1. Administrator is on the User Details page for a user
2. Administrator clicks "Delete User" in the left action panel
3. Right panel loads the Delete User Account view
4. Description text warns that deletion removes all access and is significant
5. Administrator enters their own current password in the "Your Password" field
6. Administrator clicks Delete
7. System validates the password against the administrator's account (server-side)
8. If password is correct: system soft-deletes the target user account (sets "deleted" status flag)
9. System invalidates any active sessions for the deleted user
10. Success toast: "User '[name]' has been permanently deleted"
11. System redirects to User List — deleted user no longer visible

**Alternative Flows:**
- **Alt 1 — Wrong password:** At step 7, password does not match → inline error: "Incorrect password. Please try again." + error toast. User stays on form; password field cleared.
- **Alt 2 — Empty password:** Administrator clicks Delete without entering password → inline error: "Password is required". Form not submitted.
- **Alt 3 — Self-deletion:** Administrator attempts to delete their own account → system blocks with inline error + error toast: "You cannot delete your own account". Form not submitted.
- **Alt 4 — Navigate away:** Administrator clicks another left panel button or back arrow without submitting → no deletion occurs; no confirmation dialog needed.
- **Alt 5 — API error:** Delete fails for a server reason → error toast: "Failed to delete user account. Please try again." User stays on form.
- **Alt 6 — User already deleted:** Between loading User Details and clicking Delete, the user was deleted by another admin → "User not found" error toast → redirect to User List.

**Exit Points:**
- **Success:** User soft-deleted → toast → redirect to User List
- **Navigate away:** Click another panel button or back arrow → no action taken
- **Error:** Password wrong, API error, or self-deletion blocked → stays on form

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Your Password | Password input (576px) with show/hide eye icon | Not empty (client-side); validated server-side against admin's current password | Yes (*) | Empty | Administrator's own password to authorize the deletion. Placeholder: "Enter password" |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Password visibility toggle | Eye icon button | Right side of password input | Toggle show/hide | Switches between masked (•••) and plain text | Default: masked |
| Delete | Button (danger, full-width) | Below card content, 600×40px | Always visible; disabled + spinner while processing | Validates password → soft-deletes user → toast → redirect | Danger/red style to signal destructive action |

### Validation Error Messages

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Password empty | "Password is required" | Inline, below password field |
| Password incorrect (server-side) | "Incorrect password. Please try again." | Inline below field + error toast |
| Self-deletion | "You cannot delete your own account" | Inline below field + error toast |
| API failure | "Failed to delete user account. Please try again." | Error toast |
| User already deleted | "User not found" | Error toast → redirect to User List |

**Design Gaps Identified:**
1. Figma shows button label as **"Save"** — recommend changing to **"Delete"** or **"Delete User"** with danger/red style
2. Figma shows field label as **"Super Admin Password"** — recommend changing to **"Your Password"** (no hardcoded role names per US-004)

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Card title | Static text | "Delete User Account" (Geist Semibold 18px) | Identifies the action |
| Description paragraph 1 | Static text | "Use this option to permanently delete a user account from the system. Once deleted, the user will no longer be able to log in, and all associated access will be removed." | Warns about permanence |
| Description paragraph 2 | Static text | "Please proceed with caution, as this action may be irreversible and could result in the loss of user-related data. Ensure this is the intended action before confirming." | Warns about significance |
| Password label | Label | "* Your Password" (bold 14px, mandatory) | Labels the confirmation field |
| Password input | Password field | Masked (•••) by default, 576px with eye icon | Administrator's own password for authorization |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | View loads | Empty password field; Delete button enabled; description text visible |
| Password entered | Admin types password | Masked characters (•••) unless eye icon toggled |
| Password visible | Eye icon toggled | Plain text password shown; eye icon changes to "hide" state |
| Deleting | Delete clicked, request in progress | Delete button shows spinner + disabled; password field disabled |
| Success | User deleted | Toast: "User '[name]' has been permanently deleted" → redirect to User List |
| Wrong password | Server rejects password | Inline error + error toast; password field cleared |
| Empty password | Delete clicked with blank field | Inline error: "Password is required" |
| Self-deletion blocked | Admin attempts to delete own account | Inline error + error toast: "You cannot delete your own account" |
| API error | Server fails | Error toast: "Failed to delete user account. Please try again." |
| User already deleted | Target user removed by another admin | Error toast: "User not found" → redirect to User List |

### Page Layout (Design Reference)

```
┌──────────────────────────────────────────────────────────────────┐
│ Breadcrumb / Breadcrumb / Breadcrumb                  [Top Bar]  │
├──────────┬───────────────────────────────────────────────────────┤
│ [Sidebar]│ ← User Details > Henry Tran                          │
│  200px   │                                                       │
│          │ ┌─────────────┬──────────────────────────────────────┐│
│          │ │  Overview   │ ┌──────────────────────────────────┐ ││
│          │ │  Update Info│ │ Delete User Account              │ ││
│          │ │  Change Role│ │                                  │ ││
│          │ │  Change Mail│ │ Use this option to permanently   │ ││
│          │ │  Reset Pass │ │ delete a user account...         │ ││
│          │ │  Act/Deact  │ │                                  │ ││
│          │ │ [Delete Usr]│ │ * Your Password                  │ ││
│          │ │             │ │ ┌────────────────────────── 👁 ┐ │ ││
│          │ │  189px      │ │ │ Enter password                │ │ ││
│          │ │             │ │ └──────────────────────────────┘ │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ [        Delete User        ]    │ ││
│          │ │             │ └──────────────────────────────────┘ ││
│          │ └─────────────┴──────────────────────────────────────┘│
└──────────┴───────────────────────────────────────────────────────┘
```

> **Note:** Figma design at node `3123:7686`. Button should use danger/red style; field label should read "Your Password" (not "Super Admin Password").

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Clicking "Delete User" in the left panel loads the Delete User Account view in the right panel
- **AC-02:** Card title displays "Delete User Account"
- **AC-03:** Description text warns that deletion is significant and removes all access
- **AC-04:** Password field is empty by default with placeholder "Enter password"
- **AC-05:** Password field is masked (•••) by default with an eye icon to toggle visibility

**Password Input:**
- **AC-06:** Password field label displays "* Your Password" (no hardcoded role name)
- **AC-07:** Eye icon toggles between masked and plain text display
- **AC-08:** Password field accepts any characters (validated server-side against admin's actual password)

**Delete Button:**
- **AC-09:** Delete button is full-width (600px) and uses danger/red style to signal destructive action
- **AC-10:** Delete button label reads "Delete" or "Delete User" (not "Save")
- **AC-11:** Delete button shows loading spinner and is disabled while request is in progress; password field also disabled

**Validation — Client-side:**
- **AC-12:** Clicking Delete with an empty password field shows inline error: "Password is required"
- **AC-13:** Form is not submitted if password field is empty

**Validation — Server-side:**
- **AC-14:** If password is incorrect, inline error displays: "Incorrect password. Please try again." AND error toast is shown
- **AC-15:** After incorrect password, the password field is cleared — admin must re-enter
- **AC-16:** Password validation is performed against the administrator's own account (the logged-in user), not the target user's account

**Self-Deletion Protection:**
- **AC-17:** An administrator cannot delete their own account
- **AC-18:** If an administrator attempts to delete their own account, the system blocks with inline error + error toast: "You cannot delete your own account"
- **AC-19:** Self-deletion is blocked at the API level, not just the UI

**Success Behavior:**
- **AC-20:** On successful deletion, success toast displays: "User '[name]' has been permanently deleted"
- **AC-21:** After successful deletion, system redirects to User List — deleted user is no longer visible
- **AC-22:** Deleted user's active sessions are invalidated immediately
- **AC-23:** Deleted user can no longer log in — login attempts are rejected

**Impact of Soft Delete:**
- **AC-24:** Deletion is a soft delete — user data is preserved in the database with a "deleted" status flag; no data is permanently removed
- **AC-25:** Deleted user is removed from the active User List, all department/position/role/skill assignment counts are updated
- **AC-26:** Deleted user no longer appears in any active HR module selectors or filters
- **AC-27:** Historical references to the deleted user (audit trails, past assignments) are preserved

**Error Handling:**
- **AC-28:** If Delete fails due to API error, error toast displays: "Failed to delete user account. Please try again."
- **AC-29:** If the target user was already deleted by another admin, error toast: "User not found" → redirect to User List

**Navigation:**
- **AC-30:** Navigating away without clicking Delete takes no action — no confirmation dialog needed
- **AC-31:** Left panel "Delete User" button shows active/highlighted state while on this view

**Access Control:**
- **AC-32:** Delete User button is visible only to users with user management permission
- **AC-33:** Users with view-only permission cannot see or access this action

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — correct password | Enter admin password, click Delete | User soft-deleted, toast, redirect to list | High |
| Wrong password | Enter wrong password, click Delete | Inline error + toast; field cleared; stays on form | High |
| Empty password | Click Delete without entering password | Inline error: "Password is required" | High |
| Self-deletion | Admin tries to delete own account | Blocked: "You cannot delete your own account" | High |
| Password visibility toggle | Click eye icon | Password switches between masked and plain text | Medium |
| API error | Server fails on delete | Error toast; stays on form | Medium |
| User already deleted | Another admin deleted the user | "User not found" toast → redirect to list | Medium |
| Navigate away | Enter password, click Overview | No deletion; navigates to Overview | Medium |
| Deleted user sessions | Delete active user | Their sessions invalidated immediately | Medium |
| Deleted user login | Deleted user attempts login | Login rejected | Medium |
| Data preserved | Soft-delete a user | Record retained in database with "deleted" flag | Medium |
| Historical references | Soft-delete a user with past assignments | Historical records still reference the user | Medium |
| Unauthorized | View-only user | Delete User button not visible | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with user management permission can access the Delete User action. The button is hidden for view-only users both visually and at the API level.
- **SR-02:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-03:** Deletion requires the administrator to enter **their own current password** — validated server-side against the logged-in administrator's account, not the target user's account.
- **SR-04:** Password validation happens server-side only — no client-side password comparison.
- **SR-05:** If password is incorrect, the server returns an authentication error. The password field is cleared and the admin must re-enter. No indication of whether the password was "close" or how it was wrong.
- **SR-06:** An administrator **cannot delete their own account** — blocked at the API level (not just UI). Prevents accidental lockout.
- **SR-07:** Deletion is a **soft delete** — the user record is retained in the database with a "deleted" status flag. No data is permanently removed.
- **SR-08:** A soft-deleted user **cannot log in** — login attempts are rejected. Any active sessions are invalidated immediately on deletion.
- **SR-09:** A soft-deleted user is **excluded from all active views**: User List, department/position/role/skill assignment counts, HR module selectors, and any active filters or searches.
- **SR-10:** A soft-deleted user's **historical references are preserved** — any audit trails, past assignments, or historical records that reference this user remain intact.
- **SR-11:** Restoration of soft-deleted users is **out of scope** for this release but architecturally possible — the data is retained for this purpose.
- **SR-12:** After successful deletion, the system redirects to the User List — the deleted user is no longer visible in the active list.
- **SR-13:** If the target user was already deleted by another administrator between page load and delete action, the system returns a "User not found" error and redirects to User List.
- **SR-14:** The password field label must read "Your Password" (or equivalent generic label) — not "Super Admin Password" or any role-specific name. Consistent with US-004 convention.

**State Transitions:**
```
[User Details] → [Click Delete User button] → [Delete User view loaded (empty password)]
[Delete User view] → [Enter password, click Delete] → [Server validates password]
[Password valid] → [Soft delete executed] → [Sessions invalidated] → [Toast] → [User List]
[Password invalid] → [Inline error + toast; password cleared] → [Delete User view]
[Password empty] → [Client-side error: "Password is required"] → [Delete User view]
[Self-deletion attempt] → [Blocked: "Cannot delete own account"] → [Delete User view]
[Server error] → [Error toast] → [Delete User view]
[User already deleted] → [Error toast: "User not found"] → [User List]
[Navigate away without submitting] → [No action taken]
```

**Dependencies:**
- **US-001 (Authentication):** Password validation against admin's account; session invalidation for deleted user; login rejection
- **US-004 (Role & Permission Management):** Controls access to this action; role/permission assignment counts updated
- **DR-001-005-01 (User List):** Deleted user removed from active list; redirected here on success
- **DR-001-005-03 (User Details):** Entry point via left panel
- **DR-001-005-08 (Activate/Deactivate):** Soft delete is a stronger action — deactivated users are still visible in User List; deleted users are not

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Description text clearly warns about the significance of the action before the password field — admin reads the consequences first
- **UX-02:** Password field is masked by default — prevents shoulder-surfing in shared office environments
- **UX-03:** Eye icon provides show/hide toggle for password — admin can verify what they typed without retyping
- **UX-04:** Delete button uses **danger/red style** to visually signal destructive action — distinct from Save buttons on other User Details action views
- **UX-05:** Delete button shows loading spinner while request is in progress — prevents double-click
- **UX-06:** After incorrect password, the field is cleared automatically — forces admin to re-enter deliberately, reducing copy-paste mistakes
- **UX-07:** Success toast auto-dismisses after 5 seconds — consistent with other success toasts
- **UX-08:** Error toasts (wrong password, self-deletion, API error) persist until dismissed manually — ensures admin notices the failure
- **UX-09:** No separate confirmation dialog — the **password entry IS the confirmation**. This is a stronger guard than a simple "Are you sure?" dialog, because it proves the admin has credentials authority.
- **UX-10:** Left panel "Delete User" button shows active/highlighted state — clear context of current view
- **UX-11:** Navigating away without submitting requires no confirmation dialog — nothing has been executed yet, and there's no valuable data to lose

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Two-panel layout: left action panel (189px) + right card (600px) |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Password field keyboard accessible — Tab to focus, Enter to submit
- [x] Eye icon toggle keyboard accessible — Space/Enter to toggle visibility
- [x] Screen reader announces "Your Password" label and masked state
- [x] Delete button marked as destructive (aria-describedby pointing to warning text)
- [x] Sufficient color contrast — danger button meets WCAG 2.1 AA
- [x] Focus indicators visible on password field, eye icon, and Delete button

**Design References:**
- Figma: [Delete User Account](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7686) (node `3123:7686`)
- **Design gaps to address:**
  1. Button label: "Save" → should be "Delete" or "Delete User" with danger/red style
  2. Field label: "Super Admin Password" → should be "Your Password" (no hardcoded role names)

---

## 8. Additional Information

### Out of Scope
- Hard delete (permanent data removal) — soft delete only; data retained in database
- Restoration of soft-deleted users — future enhancement; architecturally possible but not in this release
- Bulk delete (multiple users at once)
- Scheduled deletion (e.g., "delete after 30 days of inactivity")
- Deletion reason or comment field
- Notification email to the deleted user
- Audit log of deletions on this page (future enhancement)
- Admin password change or reset from this view

### Open Questions
- [ ] **Button label and style:** Figma shows "Save" with primary/dark style. Recommend "Delete" or "Delete User" with danger/red style. — **Owner:** Design Team — **Status:** Pending
- [ ] **Field label:** Figma shows "Super Admin Password". Recommend "Your Password" to avoid hardcoded role names. — **Owner:** Design Team — **Status:** Pending
- [ ] **Data retention policy:** How long is soft-deleted user data retained in the database? Indefinitely, or auto-purged after a period (e.g., 90 days)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Restoration mechanism:** Will there be a future admin tool to restore soft-deleted users? If so, where — a separate "Deleted Users" list or a flag in the existing User List? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Impact on department/position/skill counts:** When a user is soft-deleted, are they removed from the "No. of Employees" counts on Department, Position, and Skill lists? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-01: User List | Deleted user removed from active list; redirected here on success |
| DR-001-005-02: Create User | Creates the user that this feature can delete |
| DR-001-005-03: User Details | Entry point via left panel "Delete User" button |
| DR-001-005-08: Activate/Deactivate | Lighter alternative — blocks login but keeps user visible in active lists. Delete is stronger: blocks login AND hides from all views |
| US-001: Authentication | Password validation, session invalidation, login rejection |
| US-004: Role & Permission Management | Controls access; assignment counts may be affected |
| EP-008 US-001/002: Dept & Position Mgmt | Employee counts may decrease when user is soft-deleted |
| EP-008 US-003: Skill Management | Employee counts may decrease when user is soft-deleted |

### Notes
- **Soft delete vs. Deactivate** — these are two distinct levels of "removing" a user:
  - **Deactivate** (DR-001-005-08): Blocks login, but user remains visible in User List with "Inactive" badge. Easily reversible via toggle.
  - **Delete** (this DR): Blocks login AND hides user from all active views (User List, filters, selectors, counts). Data retained in database. Requires password confirmation. Restoration is a future feature.
- **Password confirmation** replaces the typical "Are you sure?" dialog. This is a **stronger security pattern** — it proves the admin has credential authority, not just the ability to click a button.
- The **two design gaps** (button label + field label) should be addressed by the Design Team before development. Both are usability and convention issues, not blockers.
- **Self-deletion protection** (SR-06) is critical and must be enforced at the API level — consistent with the self-deactivation protection in DR-001-005-08.

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
| 1.0 | 2026-03-27 | BA Agent | Initial draft — soft delete with password confirmation. Figma node 3123:7686. |
