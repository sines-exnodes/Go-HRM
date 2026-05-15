---
document_type: FLOWCHART
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
  - path: "./REQUIREMENTS.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
---

# Business Process Flowcharts: User Authentication

**Epic:** EP-001 (Foundation)
**Story:** US-001-authentication
**Last Updated:** 2026-02-26

---

## Table of Contents

1. [Sign-In Process Flow](#1-sign-in-process-flow)
2. [Password Reset Flow](#2-password-reset-flow)
3. [Session Management Flow](#3-session-management-flow)
4. [Actor Interactions](#4-actor-interactions)
5. [State Diagram](#5-state-diagram)
6. [Error Handling](#6-error-handling)
7. [Notes & Assumptions](#7-notes--assumptions)

---

## 1. Sign-In Process Flow

### Overall Sign-In Flow

```mermaid
flowchart TD
    A[User opens Exnodes HRM] --> B[Display Sign-In Page]
    B --> C[User enters email]
    C --> D[User enters password]
    D --> E{Remember me?}
    E -->|Checked| F[Set extended session flag]
    E -->|Unchecked| G[Standard session]
    F --> H[User clicks Sign In]
    G --> H
    H --> I{Validate credentials}
    I -->|Valid| J{Check account status}
    I -->|Invalid| K[Increment failed attempts]
    J -->|Active| L{Role has login permission?}
    J -->|Deactivated| M[Show: Account deactivated]
    K --> N{Lockout threshold reached?}
    N -->|Yes| O[Lock account temporarily]
    N -->|No| P[Show: Invalid email or password]
    O --> Q[Show: Account temporarily locked]
    L -->|Yes| R[Create user session]
    L -->|No| S[Show: No permission to access system]
    R --> T[Redirect to Dashboard]
    P --> B
    Q --> B
    S --> B
    M --> V[Show: Contact HR administrator]
```

### Key Steps

1. **Display Sign-In Page** — Centered card with email, password, remember me, and sign-in button
2. **Credential Validation** — System checks email/password against stored credentials
3. **Account Status Check** — Verifies account is active (not deactivated or locked)
4. **Login Permission Check** — Verifies user's assigned role has login permission (roles are configurable via US-004)
5. **Dashboard Redirect** — Authenticated user redirected to Dashboard

---

## 2. Password Reset Flow

### Forgot Password Flow

```mermaid
flowchart TD
    A[User clicks Forgot Password] --> B[Display email input form]
    B --> C[User enters email address]
    C --> D[User clicks Send Reset Link]
    D --> E{Email registered?}
    E -->|Yes| F[Generate reset token]
    E -->|No| G[Show generic success message]
    F --> H[Send reset email with link]
    H --> G
    G --> I[Show: If this email exists, a reset link has been sent]
    I --> J[User opens email]
    J --> K[User clicks reset link]
    K --> L{Token valid and not expired?}
    L -->|Valid| M[Display new password form]
    L -->|Expired| N[Show: Link expired, request new one]
    M --> O[User enters new password]
    O --> P[User confirms new password]
    P --> Q{Passwords match and meet policy?}
    Q -->|Yes| R[Update password]
    Q -->|No| S[Show validation errors]
    S --> O
    R --> T[Invalidate reset token]
    T --> U[Redirect to Sign-In with success message]
    N --> A
```

### Key Steps

1. **Initiate Reset** — User clicks "Forgot Password" on sign-in page
2. **Email Verification** — System checks if email exists (but never reveals this to user)
3. **Token Generation** — Secure, single-use, time-limited reset link
4. **Password Update** — User sets new password meeting policy requirements
5. **Redirect** — User returns to sign-in page with confirmation message

---

## 3. Session Management Flow

### Session Lifecycle

```mermaid
flowchart TD
    A[User signs in successfully] --> B{Remember me checked?}
    B -->|Yes| C[Create extended session]
    B -->|No| D[Create standard session]
    C --> E[User uses platform]
    D --> E
    E --> F{User action detected?}
    F -->|Yes| G[Reset idle timer]
    G --> E
    F -->|No - Idle| H{Idle timeout reached?}
    H -->|No| F
    H -->|Yes| I[Expire session]
    I --> J[Redirect to Sign-In]
    J --> K[Show: Session expired message]

    E --> L{User clicks Logout?}
    L -->|Yes| M[Clear session]
    M --> J
```

---

## 4. Actor Interactions

### Sign-In Sequence

```mermaid
sequenceDiagram
    participant User
    participant SignInPage as Sign-In Page
    participant AuthService as Auth Service
    participant Dashboard

    User->>SignInPage: Open Exnodes HRM
    SignInPage-->>User: Display sign-in form

    User->>SignInPage: Enter email + password
    User->>SignInPage: Click Sign In

    SignInPage->>AuthService: Validate credentials
    AuthService->>AuthService: Check role login permission
    AuthService-->>SignInPage: Credentials valid + permission granted

    SignInPage->>Dashboard: Redirect to Dashboard
    Dashboard-->>User: Display dashboard
```

### Password Reset Sequence

```mermaid
sequenceDiagram
    participant User
    participant SignInPage as Sign-In Page
    participant AuthService as Auth Service
    participant EmailService as Email Service

    User->>SignInPage: Click Forgot Password
    SignInPage-->>User: Display email form

    User->>SignInPage: Enter email, click Send
    SignInPage->>AuthService: Request password reset

    AuthService->>EmailService: Send reset email
    EmailService-->>User: Reset link email

    User->>AuthService: Click reset link
    AuthService-->>User: Display new password form

    User->>AuthService: Enter new password
    AuthService-->>User: Password updated
    AuthService-->>SignInPage: Redirect with success message
```

---

## 5. State Diagram

### User Authentication States

```mermaid
stateDiagram-v2
    [*] --> Unauthenticated

    Unauthenticated --> Authenticating: Enter credentials
    Authenticating --> Authenticated: Valid credentials
    Authenticating --> FailedAttempt: Invalid credentials
    Authenticating --> AccountLocked: Lockout threshold reached
    Authenticating --> AccountDeactivated: Account inactive
    Authenticating --> NoLoginPermission: Role lacks login permission

    FailedAttempt --> Unauthenticated: Retry
    NoLoginPermission --> Unauthenticated: Return to sign-in
    AccountLocked --> Unauthenticated: Lockout expires

    Authenticated --> SessionActive: Session created
    SessionActive --> SessionExpired: Idle timeout
    SessionActive --> Unauthenticated: User logs out

    SessionExpired --> Unauthenticated: Redirect to sign-in

    Unauthenticated --> ResettingPassword: Forgot password
    ResettingPassword --> Unauthenticated: Password reset complete
```

**States:**
- **Unauthenticated:** User is on sign-in page, no active session
- **Authenticating:** Credentials being validated
- **Authenticated:** Credentials valid, session being created
- **SessionActive:** User is using the platform
- **SessionExpired:** Idle timeout reached, session invalidated
- **FailedAttempt:** Invalid credentials, user can retry
- **AccountLocked:** Too many failed attempts, temporarily locked
- **AccountDeactivated:** Account disabled by administrator
- **NoLoginPermission:** User's role does not have login permission
- **ResettingPassword:** User is in password reset flow

---

## 6. Error Handling

### Error Recovery Flow

```mermaid
flowchart TD
    A[Sign-In Attempt] --> B{Error Type?}
    B -->|Invalid Credentials| C[Show: Invalid email or password]
    B -->|Account Locked| D[Show: Account locked for X minutes]
    B -->|Account Deactivated| E[Show: Contact your administrator]
    B -->|No Login Permission| F[Show: No permission to access system]
    B -->|Network Error| G[Show: Unable to connect, try again]
    B -->|Session Expired| H[Show: Session expired, please sign in]

    C --> I[User corrects and retries]
    D --> J[User waits for lockout to expire]
    E --> K[User contacts administrator]
    F --> K
    G --> L[User retries after checking connection]
    H --> I
```

**Error Types:**
- **Recoverable:** Invalid credentials, network error, session expired — user can retry
- **Requires Action:** Account locked (wait), Account deactivated (contact administrator), No login permission (contact administrator)

---

## 7. Notes & Assumptions

### Assumptions

1. Authentication service is available as a backend dependency
2. Email service is configured for transactional emails (password reset)
3. All user roles use the same sign-in page and flow (roles are configurable via US-004)
4. Password policies are configurable but have sensible defaults

### Future Enhancements

- [ ] Multi-factor authentication (MFA)
- [ ] Social login / SSO integration
- [ ] Biometric authentication (mobile)
- [ ] "Sign in as" feature for administrators (impersonation for support)

---

**Document Control:**
- **Version:** 1.0
- **Status:** Draft
- **Last Updated:** 2026-02-26
