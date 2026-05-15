---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-001
story_name: "User Authentication"
status: draft
version: "1.0"
last_updated: "2026-02-26"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
input_sources:
  - type: figma
    description: "Sign In screen design"
    node_id: "3016:2711"
    extraction_date: "2026-02-26"
---

# Requirements Specification: User Authentication

**Epic:** EP-001 (Foundation)
**Story:** US-001-authentication
**Status:** Draft
**Version:** 1.0
**Last Updated:** 2026-02-26

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [User Stories](#2-user-stories)
3. [Use Cases](#3-use-cases)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Business Rules](#6-business-rules)
7. [Business Data Elements](#7-business-data-elements)
8. [Information Display Requirements](#8-information-display-requirements)
9. [Dependencies](#9-dependencies)
10. [Acceptance Criteria](#10-acceptance-criteria)
11. [Open Questions](#11-open-questions)

---

## 1. Executive Summary

### Overview

This story delivers the authentication system for Exnodes HRM, enabling authorized users to securely sign in to the platform. Access is controlled through configurable roles — any user whose role has been granted login permission can sign in. The sign-in experience uses email and password credentials with a "remember me" option, and includes password reset for self-service recovery.

### Objectives

- Provide secure, role-based access to the HRM platform
- Enable self-service password recovery without IT intervention
- Support session persistence via "remember me" functionality

---

## 2. User Stories

### US-001-01: Sign In

```
As a platform user with login permission,
I want to sign in with my email and password,
so that I can access the features authorized for my role.
```

**Priority:** Critical
**Complexity:** Medium

**Acceptance Criteria:**
- [ ] AC-001-01-01: Sign-in form displays email and password fields
- [ ] AC-001-01-02: User can toggle password visibility (show/hide)
- [ ] AC-001-01-03: User can check "Remember me" to extend session
- [ ] AC-001-01-04: Successful sign-in (valid credentials + login permission) redirects to Dashboard
- [ ] AC-001-01-05: Invalid credentials display a clear error message

---

### US-001-02: Password Reset

```
As a user who forgot my password,
I want to reset it via email,
so that I can regain access without contacting IT support.
```

**Priority:** High
**Complexity:** Medium

**Acceptance Criteria:**
- [ ] AC-001-02-01: "Forgot password" link is accessible from sign-in page
- [ ] AC-001-02-02: User enters email to receive reset link
- [ ] AC-001-02-03: Reset link sent to registered email
- [ ] AC-001-02-04: User can set a new password via the reset link
- [ ] AC-001-02-05: After reset, user is redirected to sign-in page with success message

---

### US-001-03: Logout

```
As a signed-in user,
I want to log out of the platform,
so that my session is terminated securely.
```

**Priority:** High
**Complexity:** Small

**Acceptance Criteria:**
- [ ] AC-001-03-01: Logout option available from user profile menu
- [ ] AC-001-03-02: Clicking logout clears the session
- [ ] AC-001-03-03: User is redirected to sign-in page after logout

---

### US-001-04: Session Management

```
As a platform administrator,
I want user sessions to time out after inactivity,
so that unattended sessions do not pose a security risk.
```

**Priority:** High
**Complexity:** Medium

**Acceptance Criteria:**
- [ ] AC-001-04-01: Session expires after defined idle timeout
- [ ] AC-001-04-02: User is redirected to sign-in page on session expiry
- [ ] AC-001-04-03: "Remember me" sessions have a longer expiry period
- [ ] AC-001-04-04: Session expiry message is displayed to user

---

## 3. Use Cases

### UC-1: Sign In

**Primary Actor:** Platform user with login-permitted role
**Precondition:** User has an active account and their assigned role has login permission

**Main Flow:**
1. User navigates to the Exnodes HRM sign-in page
2. System displays sign-in form with email, password, "Remember me", and "Sign In" button
3. User enters email address
4. User enters password
5. User optionally checks "Remember me"
6. User clicks "Sign In"
7. System validates credentials
8. System checks if user's assigned role has login permission
9. System creates user session
10. System redirects user to Dashboard

**Postcondition:** User is authenticated and viewing their dashboard

**Alternate Flows:**
- **A1 - Forgot Password:** At step 6, user clicks "Forgot password" link → redirected to password reset flow
- **A2 - Show Password:** At step 4, user clicks eye icon → password characters become visible

**Exception Flows:**
- **E1 - Invalid Credentials:** At step 7, credentials invalid → system shows error "Invalid email or password", user remains on sign-in page
- **E2 - Account Locked:** At step 7, account is locked → system shows "Account temporarily locked" message with lockout duration
- **E3 - Account Deactivated:** At step 7, account is inactive → system shows "Your account has been deactivated. Contact your administrator."
- **E4 - No Login Permission:** At step 8, role lacks login permission → system shows "You do not have permission to access this system."

---

### UC-2: Password Reset

**Primary Actor:** User who forgot password
**Precondition:** User has an active account with a registered email

**Main Flow:**
1. User clicks "Forgot password" link on sign-in page
2. System displays password reset form (email field)
3. User enters registered email address
4. User clicks "Send reset link"
5. System sends password reset email with secure link
6. User opens email and clicks reset link
7. System displays new password form
8. User enters and confirms new password
9. User clicks "Reset password"
10. System updates password
11. System redirects to sign-in page with "Password reset successful" message

**Postcondition:** User's password is updated, user can sign in with new password

**Exception Flows:**
- **E1 - Email Not Found:** At step 5, system still shows "If this email exists, a reset link has been sent" (no indication of whether email exists)
- **E2 - Link Expired:** At step 7, reset link has expired → system shows "Link expired, please request a new one"

---

## 4. Functional Requirements

### Category: Sign-In

| Req ID | Requirement | Description | Priority |
|--------|-------------|-------------|----------|
| FR-US-001-01 | Sign-in form | Display email and password fields with sign-in button | Critical |
| FR-US-001-02 | Remember me | Checkbox to extend session duration | High |
| FR-US-001-03 | Password toggle | Eye icon to show/hide password characters | High |
| FR-US-001-04 | Credential validation | Validate email and password against stored credentials | Critical |
| FR-US-001-05 | Login permission check | Verify user's role has login permission before granting access | Critical |

### Category: Password Recovery

| Req ID | Requirement | Description | Priority |
|--------|-------------|-------------|----------|
| FR-US-001-06 | Forgot password link | Link on sign-in page to initiate password reset | High |
| FR-US-001-07 | Password reset email | Send secure reset link to registered email | High |
| FR-US-001-08 | New password form | Form to enter and confirm new password | High |
| FR-US-001-09 | Reset link expiry | Reset links expire after defined period | High |

### Category: Session Management

| Req ID | Requirement | Description | Priority |
|--------|-------------|-------------|----------|
| FR-US-001-10 | Session creation | Create authenticated session after successful sign-in | Critical |
| FR-US-001-11 | Session timeout | Expire session after idle timeout period | High |
| FR-US-001-12 | Logout | Clear session and redirect to sign-in page | High |
| FR-US-001-13 | Account lockout | Lock account after N consecutive failed sign-in attempts | Medium |

---

## 5. Non-Functional Requirements

### Security
- **NFR-US-001-SEC01:** Credentials must be transmitted over encrypted connection
- **NFR-US-001-SEC02:** Passwords must be stored in hashed format (never plain text)
- **NFR-US-001-SEC03:** Password reset links must be single-use and time-limited

### Usability
- **NFR-US-001-U01:** Sign-in form must be completable in under 30 seconds
- **NFR-US-001-U02:** Error messages must be clear and actionable
- **NFR-US-001-U03:** Form must support keyboard navigation (Tab between fields, Enter to submit)

### Performance
- **NFR-US-001-P01:** Sign-in response time under 3 seconds
- **NFR-US-001-P02:** Password reset email delivered within 60 seconds

---

## 6. Business Rules

### BR-US-001-01: Single Credential Type
**Description:** All users sign in with email + password. No social login or SSO at this stage.
**Enforcement:** Sign-in form only accepts email and password fields.

### BR-US-001-02: Account Creation by Administrator Only
**Description:** Users cannot self-register. Accounts are created by administrators through the Role & Permission module (US-004).
**Enforcement:** No "Sign up" or registration link on sign-in page.

### BR-US-001-03: Consistent Sign-In Page for All Roles
**Description:** All user roles use the same sign-in page. Roles are configurable and managed via the Role & Permission module (US-004).
**Enforcement:** Role determination and permission check happen after authentication, not before.

### BR-US-001-04: Security-First Error Messages
**Description:** Error messages must not reveal whether an email is registered.
**Enforcement:** Invalid credential errors show generic "Invalid email or password" message.

---

## 7. Business Data Elements

### Input Data Requirements

| Data Element | Business Purpose | Required? | Business Validation | Example Value |
|--------------|------------------|-----------|---------------------|---------------|
| Email Address | User identification | Yes | Must be valid email format, registered in system | john@company.com |
| Password | User authentication | Yes | Must match stored credentials | ******** |
| Remember Me | Session preference | No | Boolean toggle | Checked / Unchecked |

---

## 8. Information Display Requirements

### Display Specifications

| Information | Display Context | Empty State | Format | Business Meaning |
|-------------|-----------------|-------------|--------|------------------|
| Sign-in form | Sign-in page, always | N/A (always shows form) | Card layout per Figma | Entry point to platform |
| Error message | Below form on failed attempt | Hidden (no error) | Inline text, error color | Credential validation failed |
| Success message | After password reset | Hidden | Banner / toast | Password has been updated |
| Session expired message | On redirect to sign-in | Hidden | Info banner | Session ended, please sign in again |

---

## 9. Dependencies

### Upstream Dependencies
- **Authentication Service:** Backend service to validate credentials and manage sessions
- **Email Service:** Transactional email for password reset links

### Downstream Dependencies
- **US-002 (Dashboard):** Receives authenticated user after sign-in redirect
- **US-004 (Role & Permission Management):** Defines configurable roles and determines which roles have login permission

---

## 10. Acceptance Criteria

### Definition of Done

- [ ] All sign-in functional requirements implemented (FR-US-001-01 through FR-US-001-13)
- [ ] Sign-in form matches Figma design (layout, colors, typography)
- [ ] Password reset flow works end-to-end
- [ ] Session management enforces timeout policies
- [ ] Error states display clear, actionable messages
- [ ] Keyboard navigation works on sign-in form
- [ ] User acceptance testing completed with multiple role configurations

---

## 11. Open Questions

- [ ] **Session timeout duration:** What is the idle timeout? (e.g., 30 min) — **Owner:** Product Owner — **Status:** Pending
- [ ] **Remember me duration:** How long does extended session last? (e.g., 7 days) — **Owner:** Product Owner — **Status:** Pending
- [ ] **Lockout threshold:** How many failed attempts before lockout? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Lockout duration:** How long does lockout last? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Password policy:** Minimum length, complexity rules? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Forgot password design:** Are Figma screens for password reset coming? — **Owner:** Design Team — **Status:** Pending

---

## Appendices

### A. Related Documents

- [ANALYSIS.md](./ANALYSIS.md) — Business analysis with design context
- [FLOWCHART.md](./FLOWCHART.md) — Authentication process flowcharts
- [TODO.yaml](./TODO.yaml) — BA task tracking

---

**Document Control:**
- **Version:** 1.0
- **Status:** Draft
- **Last Updated:** 2026-02-26
