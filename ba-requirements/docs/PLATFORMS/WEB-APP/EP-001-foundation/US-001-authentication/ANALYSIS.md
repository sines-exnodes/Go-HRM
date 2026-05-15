---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-001
story_name: "User Authentication"
status: draft
version: "1.0"
last_updated: "2026-02-26"
add_on_sections:
  - "Design Context [ADD-ON]"
approved_by: null
related_documents:
  - path: "./REQUIREMENTS.md"
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
  - type: figma
    description: "Set Password screen design"
    node_id: "3223:36"
    extraction_date: "2026-04-10"
---

# Analysis: User Authentication

**Epic:** EP-001 (Foundation)
**Story:** US-001-authentication
**Status:** 🟡 In Progress

---

## Table of Contents

1. [Business Context](#1-business-context)
2. [Scope Definition](#2-scope-definition)
3. [Requirements Analysis](#3-requirements-analysis)
4. [Data Flow Analysis](#4-data-flow-analysis)
5. [User Journey Mapping](#5-user-journey-mapping)
6. [Business Rules & Constraints](#6-business-rules--constraints)
7. [Design Context [ADD-ON]](#7-design-context-add-on)
8. [Success Criteria & Metrics](#8-success-criteria--metrics)
9. [Risk Assessment](#9-risk-assessment)
10. [Assumptions & Notes](#10-assumptions--notes)

---

## 1. Business Context

### Problem Statement

SMB organizations need a secure way for authorized users to access the Exnodes HRM platform. Without proper authentication, sensitive HR data (employee records, payroll information, performance reviews) is at risk. The system must support configurable user roles with different access levels through a single sign-in experience. Roles are managed via the Role & Permission module (US-004) — any role can be granted or denied login permission.

### Stakeholders

- **Primary Users**: Any user whose role has login permission (roles are configurable via US-004)
- **Secondary Users**: IT support staff (password resets, account management)
- **Business Owner**: HR department leadership

### Business Goals

- Goal 1: All authorized users can securely access the platform with their credentials
- Goal 2: User sessions are managed securely with appropriate timeout policies
- Goal 3: Users can recover access independently via password reset without IT intervention

---

## 2. Scope Definition

### In Scope

- Sign-in form (email + password)
- "Remember me" session persistence
- Password visibility toggle
- Password reset / forgot password flow
- Session management (timeout, single-session)
- Logout functionality
- Redirect to Dashboard after successful login (login permission verified per role)

### Out of Scope

- User registration / self sign-up (managed by administrator via EP-001/US-004)
- Multi-factor authentication (future enhancement)
- Social login / SSO (future enhancement)
- Biometric authentication
- Account lockout policy administration (part of US-005 Organization Settings)

### Dependencies

- **Internal**: US-004 (Role & Permission Management) — defines which roles exist
- **External**: Authentication service (backend)

---

## 3. Requirements Analysis

### Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-001-01 | Sign-in form with email and password fields | Critical | From Figma design |
| FR-US-001-02 | "Remember me" checkbox to persist session | High | From Figma design |
| FR-US-001-03 | Password visibility toggle (show/hide) | High | From Figma design — eye icon |
| FR-US-001-04 | Forgot password / password reset flow | High | Not in current design — needs design |
| FR-US-001-05 | Session timeout and automatic logout | High | |
| FR-US-001-06 | Manual logout from any page | High | |
| FR-US-001-07 | Redirect to Dashboard after successful login | Medium | Login permission verified per role |
| FR-US-001-08 | Account lockout after failed attempts | Medium | |
| FR-US-001-09 | "Invalid credentials" error messaging | High | Not yet designed — needs error states |

### Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Security | Credentials transmitted securely | Encrypted transport |
| Usability | Sign-in completes within minimal steps | 2 fields + 1 click |
| Usability | Password reset completes without IT support | Self-service flow |
| Reliability | Authentication service available during business hours | 99.5% uptime |

---

## 4. Data Flow Analysis

### Data Entities

- **User Credentials**: Email address + password
- **User Session**: Active session state, role, expiry
- **Password Reset Token**: Temporary token for password recovery

### Data Flow

- [User enters credentials] → [System validates] → [Session created] → [User redirected to dashboard]
- [User clicks "Forgot Password"] → [System sends reset link] → [User resets password] → [Redirected to sign-in]

---

## 5. User Journey Mapping

### Primary User Flow: Sign In

```
User opens Exnodes HRM → Sign In page displayed → User enters email →
User enters password → (Optional: checks "Remember me") → Clicks "Sign In" →
System validates credentials → Role-based dashboard displayed
```

### Alternative Flows

- **Forgot Password**: Sign In page → Click "Forgot Password" → Enter email → Receive reset link → Set new password → Return to Sign In
- **Invalid Credentials**: Sign In page → Enter wrong credentials → Error message displayed → User retries
- **Session Expired**: User accesses page → Session check fails → Redirected to Sign In with message

### Key Touch Points

1. Sign In form (first interaction with platform)
2. Error messaging (failed login attempts)
3. Password reset email (recovery experience)
4. Dashboard redirect (successful entry point)

---

## 6. Business Rules & Constraints

### Business Rules

- BR-US-001-01: Email address must be a valid, registered user email
- BR-US-001-02: Password must meet organization's password policy
- BR-US-001-03: Account must be active (not deactivated by administrator)
- BR-US-001-04: After N consecutive failed attempts, account temporarily locks
- BR-US-001-05: "Remember me" extends session duration beyond standard timeout
- BR-US-001-06: Password reset links expire after a defined period

### Constraints

- Single sign-in page serves all user roles (roles are configurable via US-004)
- System verifies login permission after authentication and redirects to Dashboard
- No self-registration — accounts are created by administrators via the Role & Permission module

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-02-26.

### Source Information

| Attribute | Value |
|-----------|-------|
| **Frame** | Frame 1 (Sign In page) |
| **Node ID** | `3016:2711` |
| **Dimensions** | 1920 x 1080 |
| **Extraction Date** | 2026-02-26 |

### Component Inventory

| Component | Figma Node ID | Business Purpose | Design Status |
|-----------|---------------|------------------|---------------|
| Background Image | `3021:2964` | Visual branding (abstract gradient) | Designed |
| Logo Area (dual logos) | `3017:2948` | Brand identity display | Designed |
| Tagline Text | `3017:2748` | Welcome message / brand messaging | Placeholder text |
| "Sign In" Heading | `3021:2968` | Form section title | Designed |
| Email Input | `3017:2816` | Email address entry (with @ icon) | Designed |
| Password Input | `3017:2827` | Password entry (lock icon + eye-off toggle) | Designed |
| Remember Me Checkbox | `3017:2879` | Session persistence preference | Designed |
| Sign In Button | `3017:2917` | Submit credentials | Designed |

### Design Layout

```
┌────────────────────────── 1920 x 1080 ──────────────────────────┐
│                  (Abstract gradient background)                  │
│                                                                  │
│                 ┌──── 450 x 500 Card (rounded 20px) ────┐       │
│                 │  [Dark Header - #363636]                │       │
│                 │  🟢 Exnodes Logo  🔵 Partner Logo       │       │
│                 │  "Welcome tagline text"                 │       │
│                 │                                        │       │
│                 │  ┌─ Light Form Area (#F5F5F5) ───────┐ │       │
│                 │  │  Sign In                           │ │       │
│                 │  │                                    │ │       │
│                 │  │  [@] Enter email address           │ │       │
│                 │  │  [🔒] Enter password          [👁]  │ │       │
│                 │  │  ☐ Remember me                     │ │       │
│                 │  │                                    │ │       │
│                 │  │  [████████ Sign In ████████]       │ │       │
│                 │  └────────────────────────────────────┘ │       │
│                 └────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────────────┘
```

### Design Tokens Referenced

| Token | Value | Usage |
|-------|-------|-------|
| Exnodes color | `#27AE60` | Brand green |
| general/primary | `#171717` | Sign In button background |
| general/primary-foreground | `#fafafa` | Button text color |
| general/secondary | `#f5f5f5` | Form area background |
| general/muted-foreground | `#737373` | Placeholder text, "Sign In" heading |
| general/input | `#ffffff` | Input field background |
| general/border | `#e5e5e5` | Input field border |
| Font family | Geist | All text |
| Heading 3 | Geist Semibold, 24px, LH 28.8px | "Sign In" heading, tagline |
| Paragraph Small | Geist Regular, 14px, LH 20px | Input placeholders, checkbox label |
| Button text | Geist Medium, 14px | "Sign In" button label |

### Design Observations

1. **Card-centered layout** — Sign-in form in a centered 450px card over full-bleed gradient background
2. **Two-zone card** — Dark header (branding) + light form area with 40px rounded top corners
3. **Minimal form** — Only email + password + "Remember me" + submit button
4. **Password toggle** — Eye-off icon for show/hide password
5. **Dual-logo branding** — Exnodes logo + partner/product logo

### Gaps Identified from Design

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| No "Forgot Password" link | Users cannot self-recover access | Add link below form or near password field |
| No error states designed | Users won't know why login failed | Design inline error messages for invalid credentials |
| No loading state | Users may click multiple times | Design button loading/disabled state |
| No sign-up or registration path | New users have no entry point | Clarify: Administrator creates accounts via US-004 (no self-registration needed) |
| Tagline is placeholder text | Branding incomplete | Replace "Lorem ipsum" with actual welcome message |

---

### Set Password Screen Design Context

> Extracted from Figma design via `/figma-extract` on 2026-04-10.

#### Source Information

| Attribute | Value |
|-----------|-------|
| **Frame** | Set Password |
| **Node ID** | `3223:36` |
| **Dimensions** | 1920 x 1080 |
| **Extraction Date** | 2026-04-10 |

#### Component Inventory

| Component | Figma Node ID | Business Purpose | Design Status |
|-----------|---------------|------------------|---------------|
| Background Image | `3223:37` | Visual branding (abstract gradient) | Designed |
| Logo Area (dual logos) | `3223:44` | Brand identity display | Designed |
| Welcome Message | `3223:55` | Personalized greeting with user's full name | Designed (placeholder: "Welcome {{full name}},") |
| "Set Your Password" Heading | `3223:57` | Form section title | Designed |
| Email Input (disabled) | `3223:59` | Pre-filled email display (read-only, 50% opacity) | Designed |
| Password Input | `3223:60` | New password entry (lock icon + eye-off toggle) | Designed |
| Confirm Password Input | `3223:94` | Password confirmation (lock icon + eye-off toggle) | Designed |
| Submit Button | `3223:62` | Form submission | Designed (label: "Sign In" - may need update to "Set Password") |

#### Design Layout

```
+--------------------------- 1920 x 1080 ----------------------------+
|                   (Abstract gradient background)                    |
|                                                                     |
|                  +---- 450 x 500 Card (rounded 20px) ----+          |
|                  |  [Dark Header - #363636]               |          |
|                  |  [Exnodes Logo]  [Partner Logo]        |          |
|                  |  "Welcome {{full name}},"              |          |
|                  |                                        |          |
|                  |  +- Light Form Area (#F5F5F5) -------+ |          |
|                  |  |  Set Your Password                 | |          |
|                  |  |                                    | |          |
|                  |  |  [@] {{email}}          (disabled) | |          |
|                  |  |  [lock] Enter password       [eye] | |          |
|                  |  |  [lock] Re-Enter password    [eye] | |          |
|                  |  |                                    | |          |
|                  |  |  [########## Sign In ##########]   | |          |
|                  |  +------------------------------------+ |          |
|                  +----------------------------------------+          |
+---------------------------------------------------------------------+
```

#### Design Tokens Referenced

| Token | Value | Usage |
|-------|-------|-------|
| Exnodes color | `#27AE60` | Brand green |
| general/primary | `#171717` | Submit button background |
| general/primary-foreground | `#fafafa` | Button text color |
| general/secondary | `#f5f5f5` | Form area background |
| general/muted-foreground | `#737373` | "Set Your Password" heading, placeholders |
| general/input | `#ffffff` | Input field background |
| general/border | `#e5e5e5` | Input field border |
| Font family | Geist | All text |
| Heading 3 | Geist Semibold, 24px, LH 28.8px | Headings, welcome message |
| Paragraph Small | Geist Regular, 14px, LH 20px | Input placeholders |
| Button text | Geist Medium, 14px | Button label |

#### Design Observations

1. **Same card layout as Sign In** — 450px centered card, same visual treatment
2. **Personalized welcome** — Dynamic greeting with user's full name from invitation
3. **Email is read-only** — Pre-filled from token, 50% opacity indicates non-editable
4. **Dual password fields** — Password and Confirm Password with independent visibility toggles
5. **Button label discrepancy** — Shows "Sign In" but should be "Set Password" for clarity

#### Gaps Identified from Set Password Design

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Button label says "Sign In" | User confusion — this is password setup, not sign-in | Change button text to "Set Password" |
| No error states designed | Users won't see password policy failures or mismatch errors | Design inline error states for validation |
| No loading state | Users may click multiple times | Design button loading/disabled state |
| No error page for invalid/expired tokens | Users with bad links see nothing useful | Design error page for token validation failures |

---

## 8. Success Criteria & Metrics

### Functional Success Criteria

- [ ] User can sign in with valid email and password
- [ ] "Remember me" extends session appropriately
- [ ] Password visibility toggle works correctly
- [ ] Invalid credentials show clear error message
- [ ] Forgot password flow allows self-service recovery
- [ ] Successful login (with login permission) redirects to Dashboard
- [ ] Logout clears session and redirects to sign-in page

### Business Metrics

- Sign-in success rate > 95% (valid credential attempts)
- Password reset completion rate > 80%
- Zero IT tickets for routine password resets

---

## 9. Risk Assessment

### Identified Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Forgot password flow missing from design | High | High | Request design for password reset screens |
| Error states not designed | High | Medium | Define error messages in requirements, request error state designs |
| Session management complexity | Medium | Medium | Define clear session policies early |
| Account lockout frustration | Low | Medium | Clear messaging on lockout duration and recovery |

---

## 10. Assumptions & Notes

### Assumptions

1. All user accounts are created by administrators via the Role & Permission module (no self-registration)
2. Email address is the unique identifier for login
3. Single sign-in page serves all roles (roles are dynamic and configurable via US-004)
4. Password policies will be configurable in Organization Settings (US-005)

### Open Questions

- [ ] What is the session timeout duration? (e.g., 30 min idle, 8 hours absolute)
- [ ] How long does "Remember me" extend the session? (e.g., 7 days, 30 days)
- [ ] How many failed attempts before account lockout? (e.g., 5 attempts)
- [ ] How long does account lockout last? (e.g., 15 minutes, until admin unlock)
- [ ] Should the forgot password email include the user's name for personalization?

---

## Appendices

### A. Related Documents

- [REQUIREMENTS.md](./REQUIREMENTS.md) — Detailed requirements specification
- [FLOWCHART.md](./FLOWCHART.md) — Authentication process flowcharts
- [EPIC.md](../EPIC.md) — EP-001 Foundation Epic

---

**Document Version:** 1.0
**Last Updated:** 2026-02-26
**Author:** BA Team
