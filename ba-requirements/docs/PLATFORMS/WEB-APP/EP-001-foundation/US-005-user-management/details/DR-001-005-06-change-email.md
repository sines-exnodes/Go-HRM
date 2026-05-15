---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-06
detail_name: "Change Email"
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
input_sources:
  - type: figma
    description: "Change Email screen (User Details sub-page)"
    node_id: "3123:7139"
    extraction_date: "2026-03-26"
---

# Detail Requirement: Change Email

**Detail ID:** DR-001-005-06
**Parent Requirement:** FR-US-005-11
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **change another user's email address**, so that **their login credential and communication contact stay up to date when their email changes**.

**Purpose:** Email is the user's login credential in Exnodes HRM. When a user's email changes (e.g., name change, domain migration, correction of a typo), the administrator updates it here. The change takes effect immediately — the old email can no longer be used for login, and the affected user must re-authenticate with the new email.

**Target Users:**
- Any user with user management permission (including self-email-change — unlike role change, admins can change their own email)

**Key Functionality:**
- Display current email as read-only reference
- Single-field form to enter new email address
- Email format and uniqueness validation
- Confirmation dialog before applying (login credential change)
- Immediate effect — old email stops working, affected user forced to re-authenticate
- Descriptive text explaining the impact

---

## 2. User Workflow

**Entry Point:** User Details page → "Change Email" in the left action panel.

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has user management permission (US-004)
- User is viewing a user's details (DR-001-005-03)

**Main Flow:**
1. User is on the User Details page
2. User clicks "Change Email" in the left action panel
3. System displays the Change Email form with current email (read-only) and empty New Email input
4. System shows descriptive text explaining that email is a login credential and changes take effect immediately
5. User enters the new email address
6. User clicks Save
7. System validates: not empty, valid email format, unique across organization, differs from current email
8. System shows confirmation dialog: "Are you sure you want to change this user's email to [new email]? The user will need to log in with the new email immediately."
9. User clicks Confirm
10. System saves the new email — change takes effect immediately
11. System shows success toast "User email updated successfully"
12. User stays on the Change Email page with updated current email displayed
13. The affected user's session is invalidated on their next page load or API fetch — forced to re-authenticate with new email

**Alternative Flows:**

- **Alt 1 — Confirmation cancelled:** User clicks Cancel in the confirmation dialog → returns to form, no changes applied.
- **Alt 2 — Empty field:** User clicks Save with empty New Email → inline error: "New email is required".
- **Alt 3 — Invalid format:** User enters invalid email format → inline error: "Please enter a valid email address".
- **Alt 4 — Duplicate email:** Email already used by another user → inline error: "This email is already in use".
- **Alt 5 — Same email:** User enters the current email → inline error: "New email must be different from the current email".
- **Alt 6 — Navigate away (form dirty):** User has typed in the New Email field then clicks another action or back arrow → "Discard unsaved changes?" confirmation dialog → Confirm: navigates away / Cancel: stays on form.
- **Alt 7 — Navigate away (form clean):** User navigates away with empty/unchanged field → navigates immediately, no dialog.
- **Alt 8 — Self-email-change:** Admin changes their own email → after save, they are also forced to re-authenticate with the new email.
- **Alt 9 — Server error:** Save fails → error toast with retry suggestion, form data preserved.

**Exit Points:**
- **Success:** Toast shown, stays on page with updated current email
- **Navigate away (clean):** Direct navigation to target
- **Navigate away (dirty + confirm discard):** Navigation to target, input lost
- **Navigate away (dirty + cancel discard):** Stay on form, input preserved

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Current Email | Read-only text label | N/A — display only | N/A | User's current email | — | Shows which email is being changed from. Not in original Figma design — added per BA/PO decision. |
| New Email | Text input (576px, full width) | Not empty; valid email format (user@domain.tld); unique across organization (case-insensitive); must differ from current email. Checked on Save. | Yes (*) | Empty | "Enter email" | The new email address to assign to this user |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Save | Button (primary, full-width 600px) | Below the card | Always visible | Validate → if valid, show confirmation dialog → save on confirm | Submit email change |
| Confirmation dialog | Modal dialog | Center of screen | Appears when Save clicked with valid new email | Confirm → save and apply; Cancel → return to form | Safety net for login credential change |

**Notes:**
- No Cancel button on the form — users navigate away via left action panel or back arrow
- Single input field only — simplest form alongside Change User Role

---

## 4. Data Display

### Information Shown to User

| Data Element | Data Type | Format | Business Meaning |
|-------------|-----------|--------|------------------|
| Card title | Static text | "Change Email" — Geist Medium 14px | Section context |
| Descriptive text | Static text | Explains that email is a login credential and changes take effect immediately (pending design fix — see Notes) | Sets expectations about impact |
| Current Email | Read-only label | User's current email address | Reference — shows what is being changed |
| New Email | Text input | Empty, placeholder "Enter email" | The new email to assign |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads | Current email (read-only) + empty New Email input + descriptive text |
| Loading | Data being fetched | Skeleton/loading indicator |
| Validation error (empty) | User clicks Save with empty field | Inline error: "New email is required" |
| Validation error (format) | Invalid email format | Inline error: "Please enter a valid email address" |
| Validation error (duplicate) | Email already used by another user | Inline error: "This email is already in use" |
| Validation error (same email) | User enters the current email | Inline error: "New email must be different from the current email" |
| Confirmation dialog | Save clicked with valid new email | "Are you sure you want to change this user's email to [new email]? The user will need to log in with the new email immediately." with Confirm + Cancel |
| Saving | Request in progress after confirmation | Save button shows loading state (disabled + spinner) |
| Success | Email changed | Success toast: "User email updated successfully". Current email label updates to new email. |
| Server error | Save fails | Error toast with retry suggestion |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Change Email is accessed from the User Details left action panel
- **AC-02:** "Change Email" card displays with descriptive text explaining that email is a login credential and changes take effect immediately
- **AC-03:** Current email is displayed as a read-only label above the New Email input
- **AC-04:** Single mandatory field: New Email text input (576px), empty by default, placeholder "Enter email"

**Validation:**
- **AC-05:** User cannot save with an empty New Email field
- **AC-06:** New Email must be a valid email format (user@domain.tld)
- **AC-07:** New Email must be unique across the organization (case-insensitive) — duplicate email shows inline error "This email is already in use"
- **AC-08:** New Email must differ from current email — same email shows inline error "New email must be different from the current email"
- **AC-09:** Validation is triggered on Save (not on blur)
- **AC-10:** Validation errors display inline below the New Email field

**Save Behavior:**
- **AC-11:** When valid new email is entered and user clicks Save, a confirmation dialog appears: "Are you sure you want to change this user's email to [new email]? The user will need to log in with the new email immediately."
- **AC-12:** Confirming the dialog saves the email change, shows toast "User email updated successfully", and stays on the page with updated current email
- **AC-13:** Cancelling the dialog returns to the form with no changes applied
- **AC-14:** Save button shows loading state (disabled + spinner) while request is in progress

**Session & Login Impact:**
- **AC-15:** Email change takes effect immediately — the old email can no longer be used for login
- **AC-16:** The affected user's session is invalidated on their next page load or API fetch — they are forced to re-authenticate with the new email
- **AC-17:** If the admin changes their own email, they are also forced to re-authenticate with the new email after save

**Navigation & Unsaved Changes:**
- **AC-18:** No Cancel button — users navigate via left action panel or back arrow
- **AC-19:** If New Email field has been modified and user navigates away, "Discard unsaved changes?" confirmation dialog is shown
- **AC-20:** If New Email field is empty/unchanged, navigation proceeds immediately

**Access Control:**
- **AC-21:** Change Email action is visible only to users with user management permission
- **AC-22:** Direct URL access by unauthorized users redirects to an appropriate fallback

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path | Enter valid new email → Save → Confirm | Toast, email updated, current email label updated | High |
| Confirmation cancel | Enter email → Save → Cancel dialog | Returns to form, no change | High |
| Empty email | Click Save with empty field | Inline error "New email is required" | High |
| Invalid format | Enter "notanemail" → Save | Inline error "Please enter a valid email address" | High |
| Duplicate email | Enter email used by another user → Save | Inline error "This email is already in use" | High |
| Same email | Enter current email → Save | Inline error "New email must be different from the current email" | High |
| Case-insensitive duplicate | "Admin@test.com" exists, enter "admin@test.com" | Inline error "This email is already in use" | Medium |
| Affected user forced logout | Email changed for active user | User forced to re-auth on next page load | High |
| Self-email-change | Admin changes own email → Save → Confirm | Admin forced to re-authenticate with new email | High |
| Navigate away (dirty) | Type in field, click Overview | "Discard unsaved changes?" dialog | Medium |
| Navigate away (clean) | No input, click Overview | Navigates immediately | Medium |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Email is the user's login credential — changing it directly affects authentication
- **SR-02:** New email must be unique across the entire organization (case-insensitive — "User@test.com" = "user@test.com")
- **SR-03:** New email must differ from the current email
- **SR-04:** Email change takes effect immediately — the old email can no longer be used for login
- **SR-05:** The affected user's session is invalidated on their next page load or API fetch — they must re-authenticate with the new email
- **SR-06:** If the admin changes their own email, they are also forced to re-authenticate after save
- **SR-07:** Only users with user management permission can access the Change Email action
- **SR-08:** Self-email-change is allowed (unlike self-role-change which is blocked)
- **SR-09:** Validation (format, uniqueness, differs from current) is checked on Save, not on blur
- **SR-10:** Saving requires explicit confirmation via dialog before applying
- **SR-11:** No email notifications are sent to either the old or new email address when the email is changed
- **SR-12:** Audit logging will be handled by a separate logging story — this DR does not define logging behavior
- **SR-13:** Last-save-wins for concurrent editing — no conflict detection

**State Transitions:**
```
[User Details] → "Change Email" click → [Change Email form (current email shown, New Email empty)]
[Change Email] → Save (valid) → [Confirmation dialog]
[Confirmation dialog] → Confirm → [Change Email (updated email + success toast + session invalidated for affected user)]
[Confirmation dialog] → Cancel → [Change Email (no change)]
[Change Email] → Save (invalid) → [Change Email (inline errors)]
[Change Email] → Navigate away (clean) → [Target page]
[Change Email] → Navigate away (dirty) → [Discard changes dialog]
[Discard changes dialog] → Confirm discard → [Target page]
[Discard changes dialog] → Cancel → [Change Email (preserved)]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — email is the login credential; session invalidation on change
- **Depends on:** US-004 (Role & Permission Management) — access control enforcement
- **Depends on:** DR-001-005-03 (User Details) — parent page providing the action panel and entry point

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Current email displayed as read-only label above the input — admin can see what they're changing from without needing to navigate back
- **UX-02:** Descriptive text explains the impact (login credential change, immediate effect) — sets expectations before the admin acts
- **UX-03:** Confirmation dialog reinforces the severity of the action — email is a login credential, and the affected user will be forced to re-authenticate
- **UX-04:** Save button shows loading spinner while request is in progress, preventing double submission
- **UX-05:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-06:** Smart dirty check — if the user types then clears the New Email field back to empty, the form is "clean" again

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Action panel (189px) + card (600px), side-by-side |
| Tablet (768-1024px) | Action panel collapses, card full-width |
| Mobile (<768px) | Action panel becomes top menu, card full-width |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab to input and Save button, Enter to submit
- [x] Screen reader compatible — field label, current email label, descriptive text, confirmation dialog, toast announcements
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements

**Design References:**
- Figma: [Change Email](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7139) (node `3123:7139`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Change User Role (DR-001-005-05) — same navigation pattern, single-field form

---

## 8. Additional Information

### Out of Scope
- Email verification / confirmation link sent to new email
- Notification to old or new email address
- Bulk email change (multiple users at once)
- Email change history/audit log (separate logging story)
- Self-service email change by the user themselves (this is admin-managed)
- Email domain restrictions (e.g., company domain only)

### Open Questions
- None remaining. All questions resolved during requirement writing session.

### Related Features
- **DR-001-005-03:** User Details — parent page providing the action panel and entry point
- **DR-001-005-04:** Update Information — sibling action; deliberately excludes email (managed here)
- **DR-001-005-05:** Change User Role — sibling action, same single-field pattern
- **US-001:** Authentication — email is the login credential; session invalidation on change
- **US-004:** Role & Permission Management — access control

### Design Issues to Resolve
- **Descriptive text copy error:** The Figma design (node `3123:7518`) currently shows text copied from the Reset Password screen: "Use this option to reset a user's password..." This must be replaced by the Design Team with appropriate text describing email change behavior and impact. The requirement assumes corrected text will explain that email is a login credential and changes take effect immediately.
- **Current email not shown in design:** The Figma design does not include a read-only label for the current email. This was identified as a gap during requirements writing. BA/PO decided the current email should be displayed above the New Email input so the admin knows what they're changing from. Design Team should update the Figma screen accordingly.

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
