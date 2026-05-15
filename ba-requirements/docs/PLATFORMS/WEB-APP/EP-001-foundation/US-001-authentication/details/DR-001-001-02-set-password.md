---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-001
story_name: "User Authentication"
detail_id: DR-001-001-02
detail_name: "Set Password"
parent_requirement: FR-US-001-08
status: draft
version: "1.0"
created_date: "2026-04-10"
last_updated: "2026-04-10"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-US-001-01.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Set Password screen design"
    node_id: "3223:36"
    extraction_date: "2026-04-10"
---

# Detail Requirement: Set Password

**Detail ID:** DR-001-001-02
**Parent Requirement:** FR-US-001-08
**Story:** US-001-authentication
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **new user who received an account invitation**, I want to **set my password using the secure link sent to my email** so that **I can complete my account setup and access the platform**.

**Purpose:** Enable new users to securely establish their password after an administrator creates their account. The admin creates the user account (via US-005 User Management), and the system sends an invitation email with a secure link. The user clicks this link to set their password for the first time, completing the account activation process.

**Target Users:** Any new user whose account has been created by an administrator. This includes employees, managers, and any other role that can be assigned in the system.

**Key Functionality:**
- Personalized welcome message with user's full name
- Read-only email display (pre-filled from invitation link)
- Password entry with visibility toggle
- Password confirmation (re-enter password)
- Password policy validation
- Secure token verification from invitation link
- Redirect to Sign In page after successful password setup

---

## 2. User Workflow

**Entry Point:** User clicks the "Set Password" link in the account invitation email sent by the system after admin creates their account.

**Preconditions:**
- Administrator has created the user's account via User Management (US-005)
- System has sent an invitation email to the user's registered email address
- Invitation link contains a valid, unexpired security token
- User has not yet set their password (first-time setup)

**Main Flow:**
1. User receives account invitation email from Exnodes HRM
2. User clicks the "Set Password" link in the email
3. System validates the security token in the URL
4. System displays the Set Password page with:
   - Personalized welcome message ("Welcome [Full Name],")
   - Pre-filled, read-only email address field
   - Empty password field
   - Empty confirm password field
   - "Set Password" button (disabled until both password fields have input)
5. User enters desired password in the password field
6. User re-enters password in the confirm password field
7. User optionally clicks eye icon to verify entered password
8. User clicks "Set Password" button
9. Button shows loading state; form fields become disabled
10. System validates password against policy requirements
11. System verifies passwords match
12. System saves the new password
13. System invalidates the invitation token (single-use)
14. System redirects user to Sign In page with success message: "Password set successfully. Please sign in."

**Alternative Flows:**
- **Alt 1 — Show/Hide Password:** User clicks eye icon on password field to toggle between masked and visible characters
- **Alt 2 — Show/Hide Confirm Password:** User clicks eye icon on confirm password field to toggle visibility
- **Alt 3 — Password Mismatch (On Submit):** User enters non-matching passwords, clicks submit, system shows inline error below confirm password field

**Exit Points:**
- **Success:** Redirect to Sign In page with success toast: "Password set successfully. Please sign in."
- **Cancel:** N/A — No cancel button; user can close browser tab or navigate away
- **Error — Token Invalid:** Display error page: "This link is invalid. Please contact your administrator for a new invitation."
- **Error — Token Expired:** Display error page: "This link has expired. Please contact your administrator for a new invitation."
- **Error — Password Already Set:** Display error page: "Your password has already been set. Please sign in or use 'Forgot Password' if you need to reset it."
- **Error — Password Policy:** Inline validation error below password field with specific requirement not met
- **Error — Password Mismatch:** Inline error below confirm password field: "Passwords do not match"

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | When Validated | Mandatory | Default Value | Description |
|------------|------------|-----------------|----------------|-----------|---------------|-------------|
| Email Address | Text input (with @ icon) | N/A (read-only, pre-filled) | N/A | Yes (display only) | Pre-filled from token | User's email address, shown for context |
| Password | Password input (with lock icon) | Must meet password policy (min length, complexity) | On blur + on submit | Yes | Empty | New password to set |
| Confirm Password | Password input (with lock icon) | Must match Password field | On blur + on submit | Yes | Empty | Re-entry to confirm password |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Password visibility toggle | Icon button (eye-off / eye) | Always visible on password field | Toggle password between masked (*****) and visible characters | Helps user verify entered password |
| Confirm Password visibility toggle | Icon button (eye-off / eye) | Always visible on confirm password field | Toggle confirm password between masked and visible | Helps user verify entered confirmation |
| Set Password | Primary button (full-width, dark) | **Disabled** when either password field empty; **Enabled** when both fields have input; **Loading** after click | Submit form for password setup | Main form submission action |

### Validation Behavior

| Trigger | Scope | Behavior |
|---------|-------|----------|
| On blur (password) | Field-level | Validate against password policy; show inline error if requirements not met |
| On blur (confirm password) | Field-level | Validate match with password field; show inline error if mismatch |
| On submit | Server-side | Validate token, password policy, password match; errors shown as inline messages or toast |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Brand logos | Images | N/A (always present) | Exnodes logo + partner logo in dark header | Brand identity and trust |
| Welcome message | Text | N/A (always present) | "Welcome [Full Name]," — Geist Semibold 24px, white | Personalized greeting using user's full name from account |
| "Set Your Password" heading | Text | N/A (always present) | Geist Semibold 24px, muted gray (#737373) | Form section title |
| Email address | Text | N/A (always pre-filled) | Pre-filled in disabled input with @ icon, 50% opacity | User's email for reference |
| Password policy hint | Text | Hidden until error | Inline text below password field | Password requirements not met |
| Password mismatch error | Text | Hidden until mismatch | Inline text below confirm password field, error color | Passwords do not match |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Initial | Page loads with valid token | Set Password form with welcome message, pre-filled email, empty password fields, button disabled |
| Token Invalid | Invalid or tampered token | Error page with message: "This link is invalid" |
| Token Expired | Token past expiration time | Error page with message: "This link has expired" |
| Already Set | Password already configured | Error page with message: "Password already set" |
| Field error (blur) | User leaves field with invalid input | Inline hint/error text below the specific field |
| Ready | Both password fields have input | Set Password button becomes enabled |
| Loading | User clicked Set Password | Button shows loading spinner, form fields and button disabled |
| Success | Password saved successfully | Redirect to Sign In page with success toast |
| Server error | Unexpected error | Toast notification with error message |

### Error Message Placement

| Error Type | Display Method | Example Message |
|------------|----------------|-----------------|
| Token invalid (page load) | Full-page error state | "This link is invalid. Please contact your administrator for a new invitation." |
| Token expired (page load) | Full-page error state | "This link has expired. Please contact your administrator for a new invitation." |
| Password already set (page load) | Full-page error state | "Your password has already been set. Please sign in or use 'Forgot Password' if you need to reset it." |
| Password policy (on blur/submit) | Inline below password field | "Password must be at least 8 characters with 1 uppercase, 1 number" |
| Password mismatch (on blur/submit) | Inline below confirm password field | "Passwords do not match" |
| Server error | Toast notification | "An error occurred. Please try again." |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Token & Access:**
- **AC-01:** Accessing the Set Password page with a valid, unexpired token displays the password form with personalized welcome message
- **AC-02:** Accessing the Set Password page with an invalid token displays an error page: "This link is invalid"
- **AC-03:** Accessing the Set Password page with an expired token displays an error page: "This link has expired"
- **AC-04:** Accessing the Set Password page after password is already set displays an error page with sign-in guidance

**Form & UI:**
- **AC-05:** Set Password page displays email field (read-only, pre-filled), password field, confirm password field, and "Set Password" button
- **AC-06:** Welcome message displays user's full name from their account: "Welcome [Full Name],"
- **AC-07:** Email field is visually disabled (50% opacity) and cannot be edited
- **AC-08:** Set Password button is disabled until both password and confirm password fields have input
- **AC-09:** Password fields mask characters by default; eye icon toggles visibility independently on each field

**Validation:**
- **AC-10:** Password field shows inline validation error on blur if it does not meet password policy requirements
- **AC-11:** Confirm password field shows inline error "Passwords do not match" on blur if it differs from password field
- **AC-12:** Set Password button shows loading state and form disables while processing

**Success Path:**
- **AC-13:** Valid password that meets policy and matches confirmation saves successfully
- **AC-14:** After successful password setup, user is redirected to Sign In page with toast: "Password set successfully. Please sign in."
- **AC-15:** Invitation token is invalidated after successful password setup (single-use)

**Keyboard & Accessibility:**
- **AC-16:** User can Tab between password fields and press Enter to submit
- **AC-17:** Screen readers announce error messages when they appear

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — valid setup | Valid token, matching passwords meeting policy | Password saved, redirect to Sign In with success message | High |
| Invalid token | Tampered or wrong token | Error page: "This link is invalid" | High |
| Expired token | Token past expiration | Error page: "This link has expired" | High |
| Password already set | Valid token but password exists | Error page: "Password already set" | High |
| Password policy failure (blur) | Password "123" (too short) | Inline error: password requirements | High |
| Password mismatch (blur) | Password: "Abc123!", Confirm: "Abc123" | Inline error: "Passwords do not match" | High |
| Empty fields submit | Click Set Password with empty fields | Button disabled, cannot submit | Medium |
| Password toggle | Click eye icon on password field | Password characters toggle visible/masked | Low |
| Keyboard navigation | Tab through fields, Enter to submit | Focus moves correctly, form submits on Enter | Low |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Token validation — verify token exists, is not expired, and has not been used
- **SR-02:** Token contains user identifier — used to retrieve user's email and full name for display
- **SR-03:** Password hashed before storage — never stored in plain text
- **SR-04:** Token invalidation — mark token as used after successful password setup (single-use token)
- **SR-05:** Password policy enforcement — server-side validation (same rules as client-side)
- **SR-06:** Account status update — user account becomes fully active after password is set
- **SR-07:** Audit logging — record timestamp of password setup for security audit

**State Transitions:**

```
[User receives invitation email] → [Clicks link] → [Token validated] → [Set Password page displayed]
[Token invalid] → [Error page: invalid link]
[Token expired] → [Error page: expired link]
[Password already set] → [Error page: already configured]
[User submits valid password] → [Password saved] → [Token invalidated] → [Redirect to Sign In]
[User submits invalid password] → [Inline error] → [User corrects] → [Resubmit]
```

**Token Lifecycle:**
```
[Token created (admin creates user)] → Active
Active → [User accesses Set Password page] → Active (validated)
Active → [User successfully sets password] → Used (invalidated)
Active → [Expiration time reached] → Expired
```

**Dependencies:**
- User Management module (US-005) — admin creates user account and triggers invitation email
- Email service — sends invitation email with secure token link
- Authentication service (backend) — validates token, stores password, manages account activation

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Auto-focus on password field when page loads — email is pre-filled, so user starts typing password immediately
- **UX-02:** Press Enter to submit form from any field
- **UX-03:** Tab order: Password → Confirm Password → Set Password button
- **UX-04:** Loading spinner on Set Password button during processing (prevents double-click)
- **UX-05:** Independent password visibility toggles — each field has its own eye icon
- **UX-06:** Clear success feedback — redirect to Sign In with visible toast message
- **UX-07:** Personalized welcome — user sees their name, confirming they're setting up the correct account
- **UX-08:** Password field uses `type="password"` for mobile keyboard optimization
- **UX-09:** Real-time validation on blur — immediate feedback before submission

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Centered 450px card over full-bleed gradient background (as per Figma) |
| Tablet (768-1024px) | Centered card with slightly reduced padding |
| Mobile (<768px) | Full-width card, background image scales, form fills screen width |

**Accessibility Requirements:**
- [x] Keyboard navigable (Tab, Enter)
- [x] Screen reader compatible (labels for icons, error announcements)
- [x] Sufficient color contrast (WCAG AA compliance)
- [x] Focus indicators visible on all interactive elements
- [x] Error messages announced to screen readers (aria-live region)
- [x] Password visibility toggle has accessible label ("Show password" / "Hide password")

**Design References:**
- Figma Node: `3223:36` (Set Password screen)
- Figma URL: https://www.figma.com/design/[file_id]?node-id=3223-36
- Font Family: Geist (Regular 14px for body, Medium 14px for button, Semibold 24px for headings)
- Primary button: `#171717` background, `#fafafa` text
- Form background: `#f5f5f5`
- Input background: `#ffffff` with `#e5e5e5` border
- Disabled input: 50% opacity
- Card: 450px wide, 20px border radius, dark header `#363636`

---

## 8. Additional Information

### Out of Scope

- Password reset flow (separate feature — covered by Forgot Password DR-US-001-03)
- Self-registration / user sign-up (accounts created by admin only)
- Password change for existing users (separate feature)
- Multi-factor authentication setup (future enhancement)
- Password strength meter / visual indicator (future enhancement)
- "Remember this device" option (not applicable for first-time setup)
- Resend invitation email functionality (handled in User Management)

### Open Questions

- [ ] Token expiration duration — how long is the invitation link valid? (e.g., 24 hours, 7 days) — **Owner:** Product Owner — **Status:** Pending
- [ ] Password policy — minimum length, complexity requirements? (e.g., 8 chars, 1 uppercase, 1 number) — **Owner:** Product Owner — **Status:** Pending
- [ ] Button text — Figma shows "Sign In" but context suggests "Set Password" — confirm correct label — **Owner:** Design Team — **Status:** Pending (suggested: "Set Password")

### Related Features

- **US-005** (User Management) — admin creates user account and triggers invitation
- **DR-US-001-01** (Sign In) — destination page after successful password setup
- **FR-US-001-07** (Password reset email) — related email-triggered password flow
- **DR-US-001-03** (Forgot Password) — similar flow for existing users, to be created

### Notes

- The email field is intentionally read-only with 50% opacity per Figma design — user cannot change the email associated with their invitation
- The welcome message uses the user's full name from the account created by the administrator
- Token-based access ensures only the intended recipient can set the password
- Single-use tokens prevent replay attacks — once used, the link becomes invalid
- Design shows "Sign In" as button text — this appears to be a design inconsistency; recommend "Set Password" to match page context

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | --- | --- | Pending |
| Product Owner | --- | --- | Pending |
| UX Designer | --- | --- | Pending |
| Tech Lead | --- | --- | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-10 | BA Team | Initial draft — all 8 sections completed via dr-agent with Figma extraction |
