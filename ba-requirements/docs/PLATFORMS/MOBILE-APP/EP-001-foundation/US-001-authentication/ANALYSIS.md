---
document_type: ANALYSIS
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-001
story_id: US-001
story_name: "Authentication"
status: draft
version: "1.0"
last_updated: "2026-04-16"
add_on_sections: []
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
  - path: "../../../WEB-APP/EP-001-foundation/US-001-authentication/ANALYSIS.md"
    relationship: reference
revision_history: []
input_sources: []
---

# Analysis: Authentication

**Epic:** EP-001 (Foundation)
**Story:** US-001-authentication
**Platform:** MOBILE-APP
**Status:** Draft

---

## 1. Business Context

### Problem Statement

Mobile users need secure, convenient access to the Exnodes HRM platform from their smartphones and tablets. The authentication experience must be optimized for mobile interaction patterns while maintaining security standards. Users expect quick sign-in, seamless password recovery, and persistent sessions that don't require frequent re-authentication.

### Stakeholders

- **Primary Users:** All employees accessing HRM via mobile app
- **Secondary Users:** IT support staff (troubleshooting auth issues)
- **Business Owner:** HR department leadership

### Business Goals

- Enable secure mobile access to HR functions
- Provide frictionless authentication experience
- Support self-service password recovery without IT intervention
- Maintain security through appropriate session management

---

## 2. Scope Definition

### In Scope

- Sign In (email + password)
- Forgot Password (OTP-based via email)
- Session management (secure token storage)
- Logout functionality
- Error handling and user feedback

### Out of Scope

- Biometric authentication (Face ID, fingerprint) — future enhancement
- Social login / SSO — future enhancement
- Multi-factor authentication — future enhancement
- Account registration (admin-managed via WEB-APP)

### Dependencies

- **WEB-APP Backend:** Authentication API (shared with web platform)
- **EP-001 US-004 (WEB-APP):** Role & Permission Management
- **Infrastructure:** Email delivery service for OTP

---

## 3. Mobile-Specific Considerations

### Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Sign In | Same (email + password) | Same (email + password) |
| Forgot Password | Email link with token | Email with OTP code (entered in-app) |
| Remember Me | Checkbox option | Default behavior (secure storage) |
| Session Storage | Browser cookies | Secure device storage (Keychain/Keystore) |
| Password Visibility | Eye icon toggle | Eye icon toggle |

### Mobile UX Patterns

- Large touch targets for form inputs
- Keyboard type optimization (email keyboard for email field)
- Password autofill support
- Loading states with activity indicators
- Haptic feedback on errors (optional)

---

## 4. User Journey Mapping

### Primary Flow: Sign In

```
App Launch → Sign In Screen → Enter Email → Enter Password → 
Tap "Sign In" → Validate Credentials → Navigate to Home/Dashboard
```

### Alternative Flow: Forgot Password (OTP)

```
Sign In Screen → Tap "Forgot Password?" → Enter Email Screen →
Submit Email → Check Email for OTP → Enter OTP Screen →
Enter 6-digit Code → Verify OTP → Set New Password Screen →
Enter New Password + Confirm → Submit → Success → Sign In Screen
```

### Error Flows

- Invalid credentials → Inline error, remain on Sign In
- Invalid/expired OTP → Inline error, option to resend
- Network error → Toast notification, retry option

---

## 5. Success Criteria

### Functional Success Criteria

- [ ] User can sign in with valid email and password
- [ ] User can request password reset via email
- [ ] User receives OTP code in email within 60 seconds
- [ ] User can enter OTP and set new password
- [ ] Session persists across app restarts
- [ ] User can sign out from any screen

### Business Metrics

- Sign-in success rate > 95%
- Password reset completion rate > 80%
- Average sign-in time < 10 seconds

---

## 6. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| OTP email delivery delays | Medium | High | Document retry flow, 10-minute expiry window |
| Brute force OTP attempts | Low | Critical | Rate limiting (3 attempts per 15 min) |
| Token storage vulnerability | Low | Critical | Use platform-native secure storage |
| Network connectivity issues | Medium | Medium | Offline error handling, retry mechanisms |

---

## 7. Assumptions & Notes

### Assumptions

1. Backend authentication API is shared with WEB-APP
2. Email address is the unique identifier for login
3. OTP codes are 6 digits, expire after 10 minutes
4. Users have access to their registered email on mobile

### Open Questions

- [ ] OTP expiration duration — confirm 10 minutes with PO
- [ ] Rate limiting thresholds — confirm with security team
- [ ] Biometric auth timeline — when to add as enhancement?

---

**Document Version:** 1.0
**Last Updated:** 2026-04-16
**Author:** BA Team
