---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-08
detail_name: "Activate/Deactivate User Account"
parent_requirement: FR-US-005-08
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
---

# Detail Requirement: Activate/Deactivate User Account

**Detail ID:** DR-001-005-08
**Parent Requirement:** FR-US-005-08
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **activate or deactivate a user's account**, so that **I can control whether the user can log in and access the system — without deleting their data**.

**Purpose:** Allow administrators to toggle a user's account status between Active and Inactive. Activating an account grants the user login access; deactivating prevents any login or system access. This is a soft control — deactivation does not delete user data and can be reversed at any time. The feature is accessed from the User Details left action panel.

**Target Users:** Any role with user management permission (configured via US-004). Users with view-only permission cannot see this action button.

**Key Functionality:**
- Single toggle switch to activate or deactivate
- Takes effect immediately on Save — no additional step required
- Reversible — can be toggled back at any time
- No confirmation dialog (simple, non-destructive action)
- Displayed inline within User Details page (not a separate page or modal)

---

## 2. User Workflow

**Entry Point:** User Details page → left panel → click "Activate/Deactivate" button

**Preconditions:**
- User is signed in (US-001)
- User has user management permission (US-004)
- The target user exists and User Details page is loaded

**Main Flow (Deactivate an active user):**
1. Administrator is on the User Details page for a user
2. Administrator clicks "Activate/Deactivate" in the left action panel
3. Right panel loads the Activate/Deactivate view — toggle shows current status (ON = Active)
4. Description text explains the impact of the action
5. Administrator toggles the switch to OFF (Inactive)
6. Administrator clicks Save
7. System updates the account status immediately
8. Success toast: "User '[name]' has been deactivated"
9. Toggle remains in the new state (OFF); user can navigate away or toggle back

**Main Flow (Activate an inactive user):**
- Same as above but toggle starts OFF → user toggles to ON → Save → toast: "User '[name]' has been activated"

**Alternative Flows:**
- **Alt 1 — Save without change:** Administrator clicks Save without toggling → no-op; no server call, no toast, no error.
- **Alt 2 — Navigate away without saving:** Administrator clicks another left panel button or back arrow without saving → change is discarded; no confirmation dialog (toggle is trivially reversible).
- **Alt 3 — Self-deactivation:** Administrator attempts to deactivate their own account → system blocks with error: "You cannot deactivate your own account". Toggle reverts to ON.
- **Alt 4 — API error:** Save fails → error toast: "Failed to update account status. Please try again." Toggle reverts to previous state.

**Exit Points:**
- **Success:** Status updated → toast → stays on Activate/Deactivate view
- **Navigate away:** Click another panel button or back arrow → returns to that view; unsaved toggle change discarded
- **Error:** Save fails → error toast; toggle reverts

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Account status | Toggle switch (33×18px) | Always has a value (ON or OFF); self-deactivation blocked | Yes (*) | Pre-filled with current status (ON = Active, OFF = Inactive) | Controls whether the user can log in and access the system |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Account status toggle | Switch | Right-aligned in status row, within 600px card | Pre-filled with current status | Toggles between Active (ON) and Inactive (OFF) | Visual: ON = dark/filled, OFF = light/empty |
| Save | Button (primary, full-width) | Below card content, 600×40px | Always visible; disabled + spinner while saving | Submits current toggle state to server | Dark background `#171717`, white text |

### Validation Error Messages

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Self-deactivation | "You cannot deactivate your own account" | Inline below toggle + error toast |
| API failure | "Failed to update account status. Please try again." | Error toast |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Format | Business Meaning |
|-----------|-----------|--------|------------------|
| Card title | Static text | "Activate/Deactivate User Account" (Geist Semibold 18px) | Identifies the action |
| Description paragraph 1 | Static text | "Use this option to control whether a user account is active or inactive. Activating an account allows the user to log in and access the system, while deactivating an account prevents any login or system access." | Explains what the toggle does |
| Description paragraph 2 | Static text | "This change takes effect immediately and does not require any further action. Deactivating an account does not delete user data and can be reversed at any time by reactivating the account." | Explains impact and reversibility |
| Account status label | Label + sublabel | "* Account status" (bold 14px) / "Activate/Deactivate" (muted 12px) | Labels the toggle |
| Account status toggle | Switch | ON = Active (dark/filled), OFF = Inactive (light/empty) | Current account state |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Active user | Target user is currently Active | Toggle is ON; description text visible; Save enabled |
| Inactive user | Target user is currently Inactive | Toggle is OFF; description text visible; Save enabled |
| Saving | Save clicked, request in progress | Save button shows spinner + disabled; toggle disabled |
| Success (deactivated) | Account deactivated | Toast: "User '[name]' has been deactivated"; toggle stays OFF |
| Success (activated) | Account activated | Toast: "User '[name]' has been activated"; toggle stays ON |
| Self-deactivation blocked | Admin toggles own account OFF and clicks Save | Error: "You cannot deactivate your own account"; toggle reverts to ON |
| API error | Save fails | Error toast: "Failed to update account status. Please try again."; toggle reverts to previous state |

### Page Layout (Design Reference)

```
┌──────────────────────────────────────────────────────────────────┐
│ Breadcrumb / Breadcrumb / Breadcrumb                  [Top Bar]  │
├──────────┬───────────────────────────────────────────────────────┤
│ [Sidebar]│ ← User Details > Henry Tran                          │
│  200px   │                                                       │
│          │ ┌─────────────┬──────────────────────────────────────┐│
│          │ │  Overview   │ ┌──────────────────────────────────┐ ││
│          │ │  Update Info│ │ Activate/Deactivate User Account │ ││
│          │ │  Change Role│ │                                  │ ││
│          │ │  Change Mail│ │ Use this option to control       │ ││
│          │ │  Reset Pass │ │ whether a user account is active │ ││
│          │ │ [Act/Deact] │ │ or inactive...                   │ ││
│          │ │  Delete User│ │                                  │ ││
│          │ │             │ │ * Account status          [===]  │ ││
│          │ │  189px      │ │   Activate/Deactivate            │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ [          Save           ]      │ ││
│          │ │             │ └──────────────────────────────────┘ ││
│          │ └─────────────┴──────────────────────────────────────┘│
└──────────┴───────────────────────────────────────────────────────┘
```

> **Note:** Figma design available at node `3123:7524`. Card is 600px wide, positioned to the right of the left action panel.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Clicking "Activate/Deactivate" in the left panel loads the Activate/Deactivate view in the right panel
- **AC-02:** Card title displays "Activate/Deactivate User Account"
- **AC-03:** Two description paragraphs explain the impact and reversibility of the action
- **AC-04:** Account status toggle is pre-filled with the user's current status (ON = Active, OFF = Inactive)
- **AC-05:** Save button is full-width (600px), primary/dark style, below the card content

**Toggle Behavior:**
- **AC-06:** Toggling from ON to OFF represents deactivating the account
- **AC-07:** Toggling from OFF to ON represents activating the account
- **AC-08:** Toggle state change is visual only until Save is clicked — no server call on toggle alone

**Save Behavior:**
- **AC-09:** Clicking Save submits the current toggle state to the server
- **AC-10:** Change takes effect immediately — no additional activation step required
- **AC-11:** After successful deactivation, toast displays: "User '[name]' has been deactivated"
- **AC-12:** After successful activation, toast displays: "User '[name]' has been activated"
- **AC-13:** Save button shows loading spinner and is disabled while request is in progress; toggle also disabled
- **AC-14:** After successful save, user stays on the Activate/Deactivate view (not redirected)
- **AC-15:** Saving without changing the toggle is a no-op — no error, no toast

**Impact of Status Change:**
- **AC-16:** A deactivated user cannot log in to the system — login attempts are rejected immediately
- **AC-17:** A deactivated user's data is fully preserved — no data deletion occurs
- **AC-18:** An activated user can log in immediately after the status change
- **AC-19:** Status change is reflected on the User Details Overview page (status badge updates to match)
- **AC-20:** Status change is reflected on the User List page (Status column badge updates)

**Self-Deactivation Protection:**
- **AC-21:** An administrator cannot deactivate their own account
- **AC-22:** If an administrator toggles their own account to OFF and clicks Save, the system blocks with error: "You cannot deactivate your own account"
- **AC-23:** Toggle reverts to ON after self-deactivation is blocked

**Navigation:**
- **AC-24:** Navigating away without saving discards the unsaved toggle change — no confirmation dialog
- **AC-25:** The left panel "Activate/Deactivate" button shows active/highlighted state while on this view

**Error Handling:**
- **AC-26:** If Save fails due to API error, error toast displays: "Failed to update account status. Please try again."
- **AC-27:** On API error, toggle reverts to the previous (server-confirmed) state

**Access Control:**
- **AC-28:** Activate/Deactivate button is visible only to users with user management permission
- **AC-29:** Users with view-only permission cannot see or access this action

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Deactivate active user | Toggle ON→OFF, Save | Toast: "deactivated"; toggle stays OFF; user can't log in | High |
| Activate inactive user | Toggle OFF→ON, Save | Toast: "activated"; toggle stays ON; user can log in | High |
| Save without change | Don't toggle, click Save | No-op; no toast, no error | Medium |
| Self-deactivation | Admin toggles own account OFF, Save | Blocked; error message; toggle reverts to ON | High |
| API error | Save fails | Error toast; toggle reverts to previous state | Medium |
| Navigate away unsaved | Toggle, then click Overview | Toggle change discarded; Overview loads | Medium |
| Pre-filled state (active) | Open view for active user | Toggle shows ON | High |
| Pre-filled state (inactive) | Open view for inactive user | Toggle shows OFF | High |
| Status reflected on Overview | Deactivate, go to Overview | Badge shows "Deactivated" (gray) | Medium |
| Status reflected on User List | Deactivate, go back to list | Status column shows "Inactive" | Medium |
| Unauthorized | View-only user | Activate/Deactivate button not visible | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with user management permission can access the Activate/Deactivate action. The button is hidden for view-only users both visually and at the API level.
- **SR-02:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-03:** The toggle is pre-filled with the user's current account status fetched from the server on view load — not cached from a previous page.
- **SR-04:** Status change takes effect **immediately** on Save — no additional activation step, no email confirmation, no delay.
- **SR-05:** A deactivated user's active sessions are invalidated immediately — any current login is terminated on status change.
- **SR-06:** Deactivation does **not** delete any user data — profile, role, department, position, skills, CV, and all assignments are fully preserved.
- **SR-07:** Deactivation is fully reversible — reactivating restores full login access with all data intact.
- **SR-08:** An administrator **cannot deactivate their own account** — the system blocks at the API level (not just UI). This prevents accidental lockout.
- **SR-09:** Saving without changing the toggle is a no-op — no server call, no toast, no error.
- **SR-10:** On API failure, the toggle reverts to the server-confirmed state (the state before the failed save attempt).
- **SR-11:** After successful status change, the new status is immediately reflected on: (a) User Details Overview page status badge, (b) User List Status column badge, (c) User's login ability.
- **SR-12:** Navigating away from the view without saving discards the unsaved toggle change — no server state is modified until Save is clicked.

**State Transitions:**
```
[User Details: Overview] → [Click Activate/Deactivate button] → [Activate/Deactivate view loaded]
[View loaded] → [Toggle unchanged, Save] → [No-op]
[View loaded] → [Toggle ON→OFF, Save] → [Saving] → [User deactivated; toast; sessions invalidated]
[View loaded] → [Toggle OFF→ON, Save] → [Saving] → [User activated; toast; login enabled]
[View loaded] → [Toggle own account OFF, Save] → [Blocked: "Cannot deactivate own account"; toggle reverts]
[Saving] → [API error] → [Error toast; toggle reverts to previous state]
[View loaded] → [Navigate away without saving] → [Toggle change discarded]
```

**Dependencies:**
- **US-001 (Authentication):** Deactivation invalidates active sessions; activation enables login
- **US-004 (Role & Permission Management):** Controls access to this action
- **DR-001-005-01 (User List):** Status column reflects the change; gear icon Activate/Deactivate mirrors this feature
- **DR-001-005-03 (User Details):** Status badge on Overview reflects the change; left panel is the entry point

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Description text clearly explains the impact before the user interacts with the toggle — no ambiguity about what activating/deactivating means
- **UX-02:** Toggle state provides immediate visual feedback — ON (dark/filled) = Active, OFF (light/empty) = Inactive
- **UX-03:** Save button is full-width (600px) — large, prominent target; hard to miss
- **UX-04:** Save button shows loading spinner while request is in progress — prevents double-click
- **UX-05:** No confirmation dialog — action is non-destructive and reversible, keeping the flow fast and frictionless
- **UX-06:** Success toast auto-dismisses after 5 seconds (with manual close option) — consistent with other success toasts
- **UX-07:** Error toast persists until dismissed manually — ensures admin notices the failure
- **UX-08:** Self-deactivation block is immediate and clear — toggle reverts automatically, no ambiguous state
- **UX-09:** No Cancel button in the design — the toggle is trivially reversible, and navigating away discards changes naturally
- **UX-10:** Left panel "Activate/Deactivate" button shows active/highlighted state — clear context of current view

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Two-panel layout: left action panel (189px) + right card (600px) |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Toggle switch accessible via keyboard (Space/Enter to toggle)
- [x] Toggle state communicated via aria-checked attribute
- [x] Screen reader announces "Account status: Active" or "Account status: Inactive"
- [x] Save button keyboard accessible (Enter to submit)
- [x] Sufficient color contrast on toggle states
- [x] Focus indicators visible on toggle and Save button

**Design References:**
- Figma: [Activate/Deactivate](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7524) (node `3123:7524`)
- Toggle component: Switch (33×18px) — same as Create User Account status toggle
- Design tokens: primary `#171717` (Save button), background `#ffffff`, muted foreground `#737373` (sublabel)

---

## 8. Additional Information

### Out of Scope
- Bulk activate/deactivate (multiple users at once)
- Scheduled deactivation (e.g., "deactivate on date X")
- Automatic deactivation after inactivity period
- Deactivation reason or comment field
- Notification email to the affected user when deactivated/activated
- Audit log of status changes on this page (future enhancement)
- Delete User — separate DR (planned)

### Open Questions
- [ ] **Session invalidation timing:** Is deactivation enforced on the user's very next API call, or are active sessions terminated immediately (forced logout)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Notification to user:** Should the deactivated/activated user receive an email notification about their status change? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Save without change behavior:** Confirmed as no-op (no toast). Or should it show an info toast "No changes made"? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-01: User List | Status column badge reflects the change; gear icon shows Activate/Deactivate |
| DR-001-005-02: Create User | Account status toggle on create form defaults to Active; same toggle pattern |
| DR-001-005-03: User Details | Status badge on Overview reflects the change; entry point via left panel |
| Delete User (planned) | Destructive alternative — permanent removal vs. soft status toggle |
| US-001: Authentication | Deactivation invalidates sessions; activation enables login |
| US-004: Role & Permission Management | Controls access to this action |

### Notes
- This is the **simplest action page** in User Details — single toggle + Save button. No form fields, no dropdowns, no file uploads.
- The design intentionally has **no Cancel button and no confirmation dialog** — the action is non-destructive (data preserved) and fully reversible (toggle back and save again). This keeps the flow fast.
- The **self-deactivation protection** (SR-08) is critical — prevents an administrator from locking themselves out of the system. This must be enforced at the API level, not just the UI.
- The **Activate/Deactivate button label** on the User Details left panel is static — it always says "Activate/Deactivate" regardless of the user's current status. The toggle inside the view reflects the actual state.
- This feature replaces the **gear icon → Deactivate/Activate** action on the User List. Both paths achieve the same result, but User Details provides the descriptive context and a more deliberate workflow.

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
| 1.0 | 2026-03-27 | BA Agent | Initial draft — Figma design context from node 3123:7524 |
