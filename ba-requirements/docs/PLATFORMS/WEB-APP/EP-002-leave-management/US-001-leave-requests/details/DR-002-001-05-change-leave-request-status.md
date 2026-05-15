---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-05
detail_name: "Change Leave Request Status"
parent_requirement: FR-US-001-11
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
  - path: "./DR-002-001-04-delete-leave-request.md"
    relationship: sibling
---

# Detail Requirement: Change Leave Request Status

**Detail ID:** DR-002-001-05
**Parent Requirement:** FR-US-001-11 (Approve/Reject), FR-US-001-12 (Cancel)
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with the appropriate leave management permission**, I want to **approve, reject, or cancel a leave request**, so that **the request's status is updated and the employee's leave balance is adjusted accordingly**.

**Purpose:** This DR covers all 3 status-changing actions for leave requests — Approve, Reject, and Cancel. These are inline actions triggered from either the Leave Requests List (gear icon) or a future Leave Request Details page. Each action updates the status, may affect the employee's leave balance, and provides immediate visual feedback via badge color change and toast notification.

**Target Users:**
- Users with **approve/reject leave request permission** — can Approve or Reject Pending requests
- Users with **cancel leave request permission** — can Cancel Pending or Approved requests
- All permissions controlled via US-004 — no hardcoded roles

**Key Functionality:**
- 3 actions combined in one DR: Approve, Reject, Cancel
- All are inline actions — no separate page, just confirmation dialogs
- Available from Leave Requests List gear icon AND future Leave Request Details page
- Status transitions: Pending → Approved, Pending → Rejected, Pending → Cancelled, Approved → Cancelled
- Leave balance deducted on Approve, restored on Cancel (if was Approved)
- No reason required for any action (simple confirmation)

---

## 2. User Workflow

**Entry Points:**
1. Leave Requests List → gear icon on a request row → select Approve / Reject / Cancel
2. Leave Request Details page (future) → action button → Approve / Reject / Cancel

**Preconditions:**
- User is signed in (US-001)
- User has the appropriate permission via US-004:
  - Approve/Reject: requires "approve/reject leave request" permission
  - Cancel: requires "cancel leave request" permission
- The target request is in a valid status for the action

### Action 1: Approve (Pending → Approved)

**Flow:**
1. User clicks Approve on a Pending request
2. System shows Confirmation Dialog: "Are you sure you want to approve this leave request?"
3. Dialog displays request summary: employee name, dates, leave type, total days
4. User clicks "Approve" (confirm)
5. System updates status to Approved
6. System deducts Total Days from employee's Leave Days Remaining
7. List refreshes — badge changes from Pending (amber) to Approved (green)
8. Success toast: "Leave request has been approved"

### Action 2: Reject (Pending → Rejected)

**Flow:**
1. User clicks Reject on a Pending request
2. System shows Confirmation Dialog: "Are you sure you want to reject this leave request?"
3. Dialog displays request summary: employee name, dates, leave type, total days
4. User clicks "Reject" (confirm)
5. System updates status to Rejected
6. No balance change (leave was not yet deducted)
7. List refreshes — badge changes from Pending (amber) to Rejected (red)
8. Success toast: "Leave request has been rejected"

### Action 3: Cancel (Pending/Approved → Cancelled)

**Flow:**
1. User clicks Cancel on a Pending or Approved request
2. System shows Confirmation Dialog: "Are you sure you want to cancel this leave request?"
3. Dialog displays request summary: employee name, dates, leave type, total days
4. If request was Approved: dialog includes note "Leave balance will be restored."
5. User clicks "Cancel Request" (confirm)
6. System updates status to Cancelled
7. If was Approved → system restores Total Days to employee's Leave Days Remaining
8. If was Pending → no balance change
9. List refreshes — badge changes to Cancelled (gray)
10. Success toast: "Leave request has been cancelled" (+ "Leave balance restored." if was Approved)

### Alternative Flows (apply to all 3 actions)

- **Alt 1 — Dismiss dialog:** User clicks "Cancel" button or X or Escape → dialog closes, no changes.
- **Alt 2 — Race condition:** Between dialog open and confirm, another user already changed the status → error toast: "This leave request has already been updated." → list refreshes.
- **Alt 3 — API error:** Action fails → error toast: "Failed to update leave request status. Please try again." Dialog stays open.
- **Alt 4 — Request deleted:** Request was soft-deleted since list loaded → error toast: "Leave request not found." → list refreshes.

**Exit Points:**
- **Success:** Status updated → toast → list refreshes inline (no page navigation)
- **Dismissed:** Dialog closed → no changes
- **Error:** Error toast → dialog stays open or list refreshes

---

## 3. Field Definitions

### Input Fields

No input fields — all actions use confirmation dialogs only. No reason or comment field required.

### Interaction Elements

**Approve Dialog:**

| Element | Type | Style | Trigger Action |
|---------|------|-------|----------------|
| Approve (confirm) | Primary button | Green/success style | Approves the request |
| Cancel (dismiss) | Secondary button | Default | Closes dialog |
| Close (X) | Icon button | Default | Closes dialog |

**Reject Dialog:**

| Element | Type | Style | Trigger Action |
|---------|------|-------|----------------|
| Reject (confirm) | Danger button | Red/danger style | Rejects the request |
| Cancel (dismiss) | Secondary button | Default | Closes dialog |
| Close (X) | Icon button | Default | Closes dialog |

**Cancel Dialog:**

| Element | Type | Style | Trigger Action |
|---------|------|-------|----------------|
| Cancel Request (confirm) | Warning button | Amber/warning style | Cancels the request |
| Go Back (dismiss) | Secondary button | Default | Closes dialog |
| Close (X) | Icon button | Default | Closes dialog |

---

## 4. Data Display

### Confirmation Dialog Content

**All 3 dialogs share the same request summary:**

| Data Name | Format | Business Meaning |
|-----------|--------|------------------|
| Dialog title | "Approve Leave Request" / "Reject Leave Request" / "Cancel Leave Request" | Identifies the action |
| Employee name | "[Employee Name]" | Whose request |
| Leave dates | "[From Date] — [To Date]" | Period |
| Leave type | "[Leave Type]" | Type of leave |
| Total days | "[X] days" | Duration |
| Balance note (Cancel only) | "Leave balance will be restored." | Shown only when cancelling an Approved request |
| Warning text | "Are you sure you want to [approve/reject/cancel] this leave request?" | Confirmation prompt |

### Dialog Layouts

**Approve Dialog:**
```
┌─────────────────────────────────────────┐
│  Approve Leave Request          [X]     │
├─────────────────────────────────────────┤
│                                         │
│  Are you sure you want to approve       │
│  this leave request?                    │
│                                         │
│  Employee: [Name]                       │
│  Dates: [From] — [To]                  │
│  Type: [Leave Type]                     │
│  Total: [X] days                        │
│                                         │
│              [Cancel]  [Approve]        │
└─────────────────────────────────────────┘
```

**Reject Dialog:**
```
┌─────────────────────────────────────────┐
│  Reject Leave Request           [X]     │
├─────────────────────────────────────────┤
│                                         │
│  Are you sure you want to reject        │
│  this leave request?                    │
│                                         │
│  Employee: [Name]                       │
│  Dates: [From] — [To]                  │
│  Type: [Leave Type]                     │
│  Total: [X] days                        │
│                                         │
│              [Cancel]  [Reject]         │
└─────────────────────────────────────────┘
```

**Cancel Dialog (when Approved):**
```
┌─────────────────────────────────────────┐
│  Cancel Leave Request           [X]     │
├─────────────────────────────────────────┤
│                                         │
│  Are you sure you want to cancel        │
│  this leave request?                    │
│                                         │
│  Employee: [Name]                       │
│  Dates: [From] — [To]                  │
│  Type: [Leave Type]                     │
│  Total: [X] days                        │
│                                         │
│  Leave balance will be restored.        │
│                                         │
│            [Go Back]  [Cancel Request]  │
└─────────────────────────────────────────┘
```

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Approve Dialog | User clicks Approve on Pending request | Dialog with green Approve button |
| Reject Dialog | User clicks Reject on Pending request | Dialog with red Reject button |
| Cancel Dialog (Pending) | User clicks Cancel on Pending request | Dialog without balance note |
| Cancel Dialog (Approved) | User clicks Cancel on Approved request | Dialog with "Leave balance will be restored." note |
| Processing | Confirm clicked, request in progress | Confirm button spinner + disabled; dismiss disabled |
| Success (Approved) | Request approved | Dialog closes; badge → green; toast |
| Success (Rejected) | Request rejected | Dialog closes; badge → red; toast |
| Success (Cancelled) | Request cancelled | Dialog closes; badge → gray; toast (+ balance restored msg if was Approved) |
| Race condition | Status already changed | Error toast; list refreshes |
| API error | Server fails | Error toast; dialog stays open |

### Gear Icon Action Visibility

| Request Status | Approve | Reject | Cancel |
|---------------|---------|--------|--------|
| Pending | ✅ (approve/reject perm) | ✅ (approve/reject perm) | ✅ (cancel perm) |
| Approved | ❌ | ❌ | ✅ (cancel perm) |
| Rejected | ❌ | ❌ | ❌ |
| Cancelled | ❌ | ❌ | ❌ |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Action Visibility:**
- **AC-01:** Approve and Reject actions are visible only for Pending requests to users with approve/reject permission
- **AC-02:** Cancel action is visible for Pending requests to users with cancel permission
- **AC-03:** Cancel action is visible for Approved requests to users with cancel permission
- **AC-04:** No status-change actions visible for Rejected or Cancelled requests
- **AC-05:** Actions are available from both Leave Requests List gear icon AND future Leave Request Details page

**Approve Flow:**
- **AC-06:** Clicking Approve shows a Confirmation Dialog with request summary
- **AC-07:** Approve button uses green/success style
- **AC-08:** Confirming Approve changes status from Pending to Approved
- **AC-09:** On Approve, Total Days is deducted from the employee's Leave Days Remaining
- **AC-10:** After Approve, list refreshes with green "Approved" badge
- **AC-11:** Success toast: "Leave request has been approved"

**Reject Flow:**
- **AC-12:** Clicking Reject shows a Confirmation Dialog with request summary
- **AC-13:** Reject button uses red/danger style
- **AC-14:** Confirming Reject changes status from Pending to Rejected
- **AC-15:** On Reject, no leave balance change (leave was not deducted)
- **AC-16:** After Reject, list refreshes with red "Rejected" badge
- **AC-17:** Success toast: "Leave request has been rejected"
- **AC-18:** No reason field is required for rejection

**Cancel Flow:**
- **AC-19:** Clicking Cancel shows a Confirmation Dialog with request summary
- **AC-20:** Cancel Request button uses amber/warning style
- **AC-21:** Confirm button label is "Cancel Request" (not "Cancel" — to distinguish from dismiss)
- **AC-22:** Dismiss button label is "Go Back" (not "Cancel" — to avoid confusion with the action)
- **AC-23:** Confirming Cancel changes status from Pending or Approved to Cancelled
- **AC-24:** If cancelling an Approved request, dialog shows note: "Leave balance will be restored."
- **AC-25:** If cancelling an Approved request, Total Days is restored to employee's Leave Days Remaining
- **AC-26:** If cancelling a Pending request, no leave balance change
- **AC-27:** After Cancel, list refreshes with gray "Cancelled" badge
- **AC-28:** Success toast: "Leave request has been cancelled" (+ "Leave balance restored." if was Approved)

**Dialog Behavior (all 3):**
- **AC-29:** All dialogs display request summary: employee name, dates, leave type, total days
- **AC-30:** Dismiss button / X / Escape closes dialog without changes
- **AC-31:** Confirm button shows loading spinner and is disabled while processing
- **AC-32:** Dialogs are modal overlays — background is not interactive
- **AC-33:** Default focus on dismiss button (not confirm) — prevents accidental action

**Error Handling:**
- **AC-34:** If status was already changed by another user, error toast: "This leave request has already been updated." List refreshes.
- **AC-35:** If request was deleted, error toast: "Leave request not found." List refreshes.
- **AC-36:** If API fails, error toast: "Failed to update leave request status. Please try again." Dialog stays open.

**Access Control:**
- **AC-37:** Approve/Reject require "approve/reject leave request" permission (US-004)
- **AC-38:** Cancel requires "cancel leave request" permission (US-004)
- **AC-39:** Users without the relevant permission do not see the action in the gear menu

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Approve Pending | Approve → confirm | Status → Approved (green); balance deducted; toast | High |
| Reject Pending | Reject → confirm | Status → Rejected (red); no balance change; toast | High |
| Cancel Pending | Cancel → confirm | Status → Cancelled (gray); no balance change; toast | High |
| Cancel Approved | Cancel → confirm | Status → Cancelled; balance restored; toast + restore msg | High |
| Cancel Approved — balance | 5-day Approved request cancelled | Employee balance +5 days | High |
| Approve — balance deduction | 3-day request approved | Employee balance −3 days | High |
| Dismiss Approve dialog | Click Cancel in dialog | Dialog closes; status unchanged | High |
| Dismiss via Escape | Press Escape | Dialog closes | Medium |
| Race condition — already approved | Approve → another user rejected | Error toast; list refreshes | Medium |
| API error | Server fails on approve | Error toast; dialog stays open | Medium |
| Request deleted | Approve → request was deleted | "Not found" toast; list refreshes | Medium |
| No approve permission | User without approve perm | Approve/Reject not visible | High |
| No cancel permission | User without cancel perm | Cancel not visible | High |
| Rejected — no actions | Check gear on Rejected request | No status-change actions | Medium |
| Half-day balance | Approve 0.5-day request | Balance −0.5 | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Approve and Reject require "approve/reject leave request" permission configured in US-004.
- **SR-02:** Cancel requires "cancel leave request" permission configured in US-004.
- **SR-03:** No role names are hardcoded — all permission checks use US-004 configuration.
- **SR-04:** **Valid status transitions:**
  - Pending → Approved (Approve action)
  - Pending → Rejected (Reject action)
  - Pending → Cancelled (Cancel action)
  - Approved → Cancelled (Cancel action)
  - No other transitions are allowed.
- **SR-05:** **Leave balance on Approve:** Total Days is deducted from employee's Leave Days Remaining. Formula: Leave Days Remaining -= Total Days.
- **SR-06:** **Leave balance on Reject:** No change — leave was not deducted for Pending requests.
- **SR-07:** **Leave balance on Cancel (was Approved):** Total Days is restored. Formula: Leave Days Remaining += Total Days.
- **SR-08:** **Leave balance on Cancel (was Pending):** No change.
- **SR-09:** All 3 actions require confirmation dialog — no direct execution on click.
- **SR-10:** Default focus on dismiss button in all dialogs — prevents accidental action.
- **SR-11:** After any successful status change, the list refreshes inline — no page navigation.
- **SR-12:** Active search/filter state is preserved after status change.
- **SR-13:** **Race condition protection:** At execution time, system verifies the request's current status matches the expected pre-action status. If another user already changed it, the action is blocked.
- **SR-14:** Actions are available from both the Leave Requests List gear icon and the future Leave Request Details page — identical behavior in both contexts.
- **SR-15:** No reason or comment is required for any action (Approve, Reject, or Cancel).

**State Transitions:**
```
[Pending] → [Approve confirmed] → [Approved; balance deducted]
[Pending] → [Reject confirmed] → [Rejected; no balance change]
[Pending] → [Cancel confirmed] → [Cancelled; no balance change]
[Approved] → [Cancel confirmed] → [Cancelled; balance restored]
[Any dialog] → [Dismiss / X / Escape] → [No change]
[Any action] → [Race condition] → [Error toast; list refreshes]
[Any action] → [API error] → [Error toast; dialog stays open]
```

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls approve/reject and cancel permissions separately
- **Leave Balance System:** Balance deducted on Approve, restored on Cancel-from-Approved
- **DR-002-001-01 (Leave Requests List):** Primary entry point; gear icon actions; list refreshes after change
- **DR-002-001-03 (Update Leave Request):** Editing an Approved request also reverts to Pending (related workflow)

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Each dialog shows the full request summary (name, dates, type, days) — user confirms the right request before acting
- **UX-02:** Each action has a **distinct button style:** Approve (green), Reject (red), Cancel Request (amber) — instant visual differentiation
- **UX-03:** Cancel dialog uses "Cancel Request" as confirm label and "Go Back" as dismiss — avoids confusion between the Cancel action and the Cancel button
- **UX-04:** When cancelling an Approved request, the dialog explicitly states "Leave balance will be restored." — no surprise about balance impact
- **UX-05:** Default focus on dismiss button in all dialogs — prevents accidental approval, rejection, or cancellation
- **UX-06:** Confirm buttons show spinner while processing — prevents double-click
- **UX-07:** Pressing Escape closes any dialog — consistent keyboard shortcut
- **UX-08:** Status badge updates inline immediately — no page reload needed; rapid batch processing supported
- **UX-09:** Toast messages include balance context when relevant (e.g., "Leave balance restored.")
- **UX-10:** All 3 actions work from both List and Details — consistent behavior regardless of entry point

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Centered modal dialog ~400px |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard focus trapped inside dialog
- [x] Escape closes dialog
- [x] Default focus on dismiss button
- [x] Screen reader announces dialog title and content (ARIA role="dialog")
- [x] Confirm buttons have descriptive ARIA labels ("Approve leave request", "Reject leave request", "Cancel leave request")
- [x] Color-coded buttons also have text labels — not color-only

**Design References:**
- No Figma design available — follows established HRM confirmation dialog pattern
- Button styles: Approve (green/success), Reject (red/danger), Cancel Request (amber/warning)
- Dialog layout consistent with Delete Department, Delete Skill, Delete Leave Request patterns

---

## 8. Additional Information

### Out of Scope
- Rejection reason or comment field (simple rejection — no reason required)
- Approval reason or comment field
- Bulk approve/reject/cancel (multiple requests at once)
- Approval chain or multi-level approval workflow
- Auto-approval rules (e.g., auto-approve if ≤ 1 day)
- Notification email to employee on status change (future enhancement)
- Undo after status change
- Mobile or tablet layout

### Open Questions
- [ ] **Notification on status change:** Should the employee receive an email or in-app notification when their request is approved/rejected/cancelled? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Approval delegation:** Can an approver delegate their approval authority to another user when out of office? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-01: Leave Requests List | Primary entry point via gear icon; list refreshes after status change |
| DR-002-001-02: Create Leave Request | Creates requests processed by this feature |
| DR-002-001-03: Update Leave Request | Editing Approved → reverts to Pending (related workflow) |
| DR-002-001-04: Delete Leave Request | Separate action — removes record vs. changes status |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls approve/reject and cancel as separate permissions |
| Leave Balance System | Deducted on Approve, restored on Cancel-from-Approved |

### Notes
- **Combined DR rationale:** Approve, Reject, and Cancel are combined into one DR because they share the same pattern: gear icon → confirmation dialog → inline status update → toast. Separating them would create 3 nearly identical documents.
- **"Cancel Request" vs "Cancel" naming:** The confirm button in the Cancel dialog is deliberately labeled "Cancel Request" (not "Cancel") to distinguish the action from the dismiss button. The dismiss button is labeled "Go Back" for the same reason.
- **Two separate permissions:** Approve/Reject is one permission; Cancel is a separate permission. This allows organizations to grant cancel rights to more users (e.g., team leads) while restricting approve/reject to senior managers.
- **No reason for rejection:** Confirmed by Product Owner — rejections are simple status changes. If a reason-for-rejection feature is needed later, it can be added as a text field in the Reject dialog without changing the overall pattern.
- **Balance impact summary:**
  - Approve: balance deducted (−Total Days)
  - Reject: no change
  - Cancel from Pending: no change
  - Cancel from Approved: balance restored (+Total Days)

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
| 1.0 | 2026-03-27 | BA Agent | Initial draft — combined Approve/Reject/Cancel with permission-based access and balance management |
