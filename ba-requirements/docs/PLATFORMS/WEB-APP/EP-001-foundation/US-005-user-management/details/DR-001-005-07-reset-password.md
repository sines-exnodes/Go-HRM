---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-07
detail_name: "Reset Password"
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
  - path: "./DR-001-005-05-change-user-role.md"
    relationship: sibling
  - path: "./DR-001-005-06-change-email.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Reset Password screen (User Details sub-page)"
    node_id: "3123:7334"
    extraction_date: "2026-03-26"
---

# Detail Requirement: Reset Password

**Detail ID:** DR-001-005-07
**Parent Requirement:** FR-US-005-11
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **trigger a password reset for another user**, so that **they can create a new password via a secure, time-limited reset link sent to their registered email**.

**Purpose:** Enable administrators to help users who have forgotten their password or need a credential refresh. The admin does not set the new password directly — the system sends a secure reset link to the user's registered email, and the user creates their own new password. When triggered, the user's current password stops working immediately and their session is invalidated.

**Target Users:**
- Any user with user management permission (including self-reset — admin can trigger a reset for their own account)

**Key Functionality:**
- Pure action screen — no input fields
- Display user's registered email so admin can verify before triggering
- Descriptive text explaining the reset link flow (time-limited, single-use)
- Confirmation dialog before sending
- Immediate effect — current password stops working, session invalidated
- Reset link expires after 24 hours

---

## 2. User Workflow

**Entry Point:** User Details page → "Reset Password" in the left action panel.

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has user management permission (US-004)
- User is viewing a user's details (DR-001-005-03)

**Main Flow:**
1. User is on the User Details page
2. User clicks "Reset Password" in the left action panel
3. System displays the Reset Password screen with descriptive text and the user's registered email (read-only)
4. Admin verifies the email address is correct
5. Admin clicks "Reset Password" button
6. System shows confirmation dialog: "Are you sure you want to send a password reset link to [email]?"
7. Admin clicks Confirm
8. System sends a password reset link to the user's registered email
9. The user's current password stops working immediately
10. The user's session is invalidated immediately (forced logout on next page load/API fetch)
11. System shows success toast "Password reset link sent successfully"
12. Admin stays on the Reset Password page
13. The affected user receives an email with a time-limited (24 hours), single-use reset link
14. The affected user clicks the link, accesses the reset page, and creates a new password

**Alternative Flows:**

- **Alt 1 — Confirmation cancelled:** Admin clicks Cancel in the confirmation dialog → returns to page, no action taken.
- **Alt 2 — Multiple resets:** Admin clicks Reset Password again (confirmed) → previous reset link is invalidated, new link sent, new 24-hour window starts.
- **Alt 3 — Self-reset:** Admin triggers reset for their own account → reset link sent to their own email, their password stops working, their session is invalidated (they are logged out).
- **Alt 4 — Server error:** Send fails → error toast with retry suggestion, no password invalidation occurs.
- **Alt 5 — Expired link:** User clicks the reset link after 24 hours → link is invalid, user must request a new reset through the admin.

**Exit Points:**
- **Success:** Toast shown, stays on page, reset link sent
- **Cancel:** No action, stays on page

---

## 3. Field Definitions

### Input Fields

None — this is a pure action screen with no editable fields.

### Read-Only Display

| Element | Type | Value | Description |
|---------|------|-------|-------------|
| Email | Read-only label with mail icon | User's registered email (e.g., "henry@exnodes.vn") | Shows where the reset link will be sent — admin verifies before triggering |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Reset Password | Button (primary, full-width 600px) | Below the card | Always visible; disabled only during API call | Confirmation dialog → send reset link | Triggers password reset email. Label: "Reset Password" (not "Save") |
| Confirmation dialog | Modal dialog | Center of screen | Appears when button clicked | Confirm → send; Cancel → return to page | Safety net before triggering credential invalidation |

**Notes:**
- No Cancel button on the page — consistent with other User Details sub-pages
- Button is always enabled except during the API call (loading state with spinner)
- No cooldown period after sending — admin can trigger multiple resets if needed

---

## 4. Data Display

### Information Shown to User

| Data Element | Data Type | Format | Business Meaning |
|-------------|-----------|--------|------------------|
| Card title | Static text | "Reset Password" — Geist Medium 14px | Section context |
| Descriptive text (paragraph 1) | Static text | "Use this option to reset a user's password. When triggered, the system will send a secure password reset link to the user's registered email address. The email will include a login URL that allows the user to access the reset page and create a new password." | Explains the mechanism |
| Descriptive text (paragraph 2) | Static text | "Please ensure the user's email address is correct before proceeding. The reset link is time-limited for security purposes and can only be used once." | Security warning |
| Email (with icon) | Read-only label | Mail icon + "Email" label + email value (e.g., "henry@exnodes.vn") | Where the reset link will be sent |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads | Descriptive text + email display + Reset Password button |
| Confirmation dialog | Admin clicks Reset Password | "Are you sure you want to send a password reset link to [email]?" with Confirm + Cancel |
| Sending | API call in progress | Button shows loading state (disabled + spinner) |
| Success | Reset link sent | Success toast: "Password reset link sent successfully" |
| Server error | Send fails | Error toast with retry suggestion |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Reset Password is accessed from the User Details left action panel
- **AC-02:** "Reset Password" card displays with two paragraphs of descriptive text explaining the reset link flow
- **AC-03:** User's registered email is displayed read-only with a mail icon — admin can verify before triggering
- **AC-04:** No input fields — this is a pure action screen
- **AC-05:** Single button labeled "Reset Password" (not "Save"), full-width (600px), primary style

**Action Behavior:**
- **AC-06:** Clicking "Reset Password" shows a confirmation dialog: "Are you sure you want to send a password reset link to [email]?"
- **AC-07:** Confirming the dialog triggers the system to send a password reset link to the user's registered email
- **AC-08:** After successful send, toast "Password reset link sent successfully" is shown and admin stays on the page
- **AC-09:** Cancelling the dialog returns to the page with no action taken
- **AC-10:** Button shows loading state (disabled + spinner) during API call, re-enabled after response

**Reset Link Behavior:**
- **AC-11:** The reset link expires after 24 hours — if not used within that period, the link becomes invalid and a new reset must be triggered
- **AC-12:** The reset link can only be used once — after the user creates a new password, the link is invalidated
- **AC-13:** Triggering a new reset invalidates any previously active reset link for that user
- **AC-14:** The reset email includes a login URL that allows the user to access the reset page and create a new password

**Credential & Session Impact:**
- **AC-15:** When a reset is triggered, the user's current password stops working immediately — they cannot log in with the old password
- **AC-16:** The affected user's session is invalidated immediately — they are forced to log out on their next page load or API fetch
- **AC-17:** The user can only regain access by using the reset link to create a new password

**Self-Reset:**
- **AC-18:** Admin can reset their own password from this screen — reset link sent to their own email, their session is invalidated

**Access Control:**
- **AC-19:** Reset Password action is visible only to users with user management permission
- **AC-20:** Direct URL access by unauthorized users redirects to an appropriate fallback

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path | Click Reset Password → Confirm | Toast, reset link sent to user's email | High |
| Confirmation cancel | Click Reset Password → Cancel | Returns to page, no email sent | High |
| Multiple resets | Click Reset Password twice (both confirmed) | Second link invalidates first | High |
| Self-reset | Admin views own profile → Reset Password → Confirm | Reset link sent, admin session invalidated | High |
| Button loading | Click → Confirm → API processing | Button disabled with spinner | Medium |
| Password stops working | Reset triggered for active user | User cannot log in with old password | High |
| Session invalidated | Reset triggered for active user | User forced out on next page load | High |
| Expired link | User clicks link after 24 hours | Link invalid, must request new reset | Medium |
| Single-use link | User uses link, then clicks again | Second click shows link already used | Medium |
| Server error | API fails | Error toast, no password invalidation | High |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Reset Password does not allow the admin to set a new password directly — the system sends a reset link to the user's registered email
- **SR-02:** The reset link expires after 24 hours
- **SR-03:** The reset link can only be used once — after the user creates a new password, the link is invalidated
- **SR-04:** Triggering a new reset invalidates any previously active reset link for that user
- **SR-05:** The reset email contains a login URL to access the reset page and create a new password
- **SR-06:** Only users with user management permission can access the Reset Password action
- **SR-07:** Self-password-reset is allowed — admin can trigger a reset for their own account
- **SR-08:** Clicking Reset Password requires confirmation before sending
- **SR-09:** When a password reset is triggered, the user's current password stops working immediately — they cannot log in with the old password
- **SR-10:** The affected user's session is invalidated immediately when the reset is triggered — they are forced to log out on their next page load or API fetch
- **SR-11:** The user can only regain access by using the reset link to create a new password
- **SR-12:** If the server fails to send the reset email, the user's password is NOT invalidated — atomic operation (either both happen or neither)
- **SR-13:** Audit logging will be handled by a separate logging story — this DR does not define logging behavior

**State Transitions:**
```
[User Details] → "Reset Password" click → [Reset Password page (email shown)]
[Reset Password] → Click button → [Confirmation dialog]
[Confirmation dialog] → Confirm → [API call → password invalidated + session invalidated + email sent → success toast]
[Confirmation dialog] → Cancel → [Reset Password (no action)]
[Affected user] → Next page load → [Forced logout → login page]
[Affected user] → Uses reset link (within 24h) → [Reset page → create new password → login]
[Affected user] → Uses reset link (after 24h) → [Link expired error]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — reset link generation, reset page, password creation flow, session invalidation
- **Depends on:** US-004 (Role & Permission Management) — access control enforcement
- **Depends on:** DR-001-005-03 (User Details) — parent page providing the action panel
- **Related:** DR-001-005-06 (Change Email) — email must be correct before resetting password; admin should verify/update email first if needed

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Email displayed prominently with mail icon — admin can visually verify the destination before triggering the reset
- **UX-02:** Descriptive text explains the full flow upfront (link sent, time-limited, single-use) — no surprises for the admin
- **UX-03:** Confirmation dialog provides a safety net — triggering a reset immediately locks the user out
- **UX-04:** Button labeled "Reset Password" (not generic "Save") — clearly communicates the action
- **UX-05:** Button shows loading spinner during API call, preventing double submission
- **UX-06:** Success toast auto-dismisses after 5 seconds (with manual close option)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Action panel (189px) + card (600px), side-by-side |
| Tablet (768-1024px) | Action panel collapses, card full-width |
| Mobile (<768px) | Action panel becomes top menu, card full-width |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab to Reset Password button, Enter to activate
- [x] Screen reader compatible — descriptive text, email display, confirmation dialog, toast announcements
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on button

**Design References:**
- Figma: [Reset Password](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7334) (node `3123:7334`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Change User Role (DR-001-005-05), Change Email (DR-001-005-06) — same action panel pattern

---

## 8. Additional Information

### Out of Scope
- Admin setting a new password directly (reset is link-based only)
- Password complexity rules (defined in US-001 Authentication)
- The reset page UI itself (part of US-001 Authentication flow)
- Email template design/content (handled by system configuration)
- Bulk password reset (multiple users at once)
- Password reset history/audit log (separate logging story)
- Password expiry policies (future enhancement)

### Open Questions
- None remaining. All questions resolved during requirement writing session.

### Related Features
- **DR-001-005-03:** User Details — parent page providing the action panel
- **DR-001-005-06:** Change Email — email must be correct before resetting password; admin should verify/update email first if needed
- **US-001:** Authentication — reset link generation, reset page, password creation flow
- **US-004:** Role & Permission Management — access control

### Notes
- This is the only User Details sub-page with **no input fields** — it's a pure action screen. The button label is "Reset Password" (action-specific) rather than the generic "Save" used on other sub-pages.
- The design text explicitly warns: "Please ensure the user's email address is correct before proceeding." This creates a natural dependency on Change Email (DR-001-005-06) — if the email is wrong, the admin should update it first.
- SR-12 defines an atomic operation: if the email fails to send, the password should NOT be invalidated. This prevents a scenario where the user is locked out with no way to receive the reset link.
- When an admin resets their own password, they will be logged out and must use the reset link to create a new password and log back in.

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
