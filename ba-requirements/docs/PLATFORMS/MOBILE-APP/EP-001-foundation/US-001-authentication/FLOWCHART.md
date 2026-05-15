---
document_type: FLOWCHART
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-001
story_id: US-001
story_name: "Authentication"
status: draft
version: "1.0"
last_updated: "2026-04-16"
---

# Flowcharts: Authentication

**Epic:** EP-001 (Foundation)
**Story:** US-001-authentication
**Platform:** MOBILE-APP

---

## 1. Sign In Flow

```mermaid
flowchart TD
    A[App Launch] --> B{Session Valid?}
    B -->|Yes| C[Navigate to Home]
    B -->|No| D[Sign In Screen]
    
    D --> E[Enter Email]
    E --> F[Enter Password]
    F --> G[Tap Sign In]
    
    G --> H{Validate Locally}
    H -->|Invalid| I[Show Inline Error]
    I --> D
    
    H -->|Valid| J[Show Loading]
    J --> K{API Authentication}
    
    K -->|Success| L[Store Token Securely]
    L --> C
    
    K -->|Invalid Credentials| M[Show Error Message]
    M --> D
    
    K -->|Network Error| N[Show Retry Option]
    N --> G
```

---

## 2. Forgot Password Flow (OTP)

```mermaid
flowchart TD
    A[Sign In Screen] --> B[Tap Forgot Password]
    B --> C[Enter Email Screen]
    
    C --> D[Enter Email Address]
    D --> E[Tap Submit]
    
    E --> F{Email Registered?}
    F -->|No| G[Show Error: Email not found]
    G --> C
    
    F -->|Yes| H[Send OTP to Email]
    H --> I[Navigate to Enter OTP Screen]
    
    I --> J[User Checks Email]
    J --> K[Enter 6-Digit OTP]
    K --> L[Tap Verify]
    
    L --> M{OTP Valid?}
    M -->|Invalid| N[Show Error: Invalid Code]
    N --> I
    
    M -->|Expired| O[Show Error: Code Expired]
    O --> P[Tap Resend OTP]
    P --> H
    
    M -->|Valid| Q[Navigate to Set Password Screen]
    
    Q --> R[Enter New Password]
    R --> S[Confirm Password]
    S --> T[Tap Submit]
    
    T --> U{Passwords Match & Valid?}
    U -->|No| V[Show Validation Error]
    V --> Q
    
    U -->|Yes| W[Update Password]
    W --> X[Show Success Message]
    X --> Y[Navigate to Sign In]
```

---

## 3. Session Management Flow

```mermaid
flowchart TD
    A[App Backgrounded] --> B[Token Stored Securely]
    
    C[App Foregrounded] --> D{Token Exists?}
    D -->|No| E[Navigate to Sign In]
    D -->|Yes| F{Token Expired?}
    
    F -->|Yes| G{Refresh Token Valid?}
    G -->|Yes| H[Refresh Access Token]
    H --> I[Continue to Home]
    G -->|No| E
    
    F -->|No| I
```

---

## 4. Logout Flow

```mermaid
flowchart TD
    A[User in App] --> B[Navigate to Settings/Profile]
    B --> C[Tap Logout]
    
    C --> D{Confirm Logout?}
    D -->|Cancel| E[Return to Settings]
    D -->|Confirm| F[Clear Stored Token]
    
    F --> G[Clear Local Data]
    G --> H[Navigate to Sign In]
```

---

**Document Version:** 1.0
**Last Updated:** 2026-04-16
**Author:** BA Team
