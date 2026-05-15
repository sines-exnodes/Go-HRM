---
document_type: EPIC
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-001
epic_name: "Foundation"
status: approved
version: "1.0"
created_date: "2026-04-16"
last_updated: "2026-04-16"
approved_by: "Product Owner"
related_documents:
  - path: "../../WEB-APP/EP-001-foundation/EPIC.md"
    relationship: reference
user_stories:
  - id: US-001
    name: "Authentication"
    status: planned
    description: "Mobile authentication flows including login, forgot password, and session management"
---

# Epic: Foundation

**Epic ID:** EP-001
**Platform:** Exnodes HRM Mobile (MOBILE-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Establish the foundational capabilities for the Exnodes HRM mobile application, enabling users to securely access the platform from their mobile devices. This epic covers authentication, core navigation, and essential app infrastructure.

### Scope

- Mobile authentication (login, forgot password, session management)
- App navigation structure
- Push notification infrastructure
- Offline data handling foundation
- App settings and preferences

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Authentication | Login, forgot password (OTP-based), logout, session management | Planned |
| US-002 | Navigation & Layout | Bottom tab navigation, header, profile menu | Planned |
| US-003 | App Settings | Notification preferences, biometric settings, language | Planned |

### Dependencies

- **WEB-APP Backend Services:** Authentication API, user management API
- **Infrastructure:** Push notification service (Firebase/APNs)

### Success Criteria

- Users can authenticate via email/password on mobile
- Forgot password flow works end-to-end with OTP via email
- Session persists appropriately with secure token storage
- App navigation provides access to all planned modules

---

## Mobile-Specific Considerations

### Authentication Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Forgot Password | Email link with token | Email with OTP code (entered in-app) |
| Session Storage | Browser cookies/localStorage | Secure device storage (Keychain/Keystore) |
| Biometric Auth | Not supported | Future enhancement (US-001 or separate story) |
| Remember Me | Checkbox option | Default behavior with secure storage |

### Security Requirements

- Secure token storage using platform-native solutions
- Certificate pinning for API communication
- Biometric authentication as optional enhancement
- Session timeout with background app detection

---

## Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| OTP delivery delays | Medium | High | Document retry/resend flow, set appropriate timeout |
| Token storage security | Low | Critical | Use platform-native secure storage (Keychain/Keystore) |
| Offline authentication | Medium | Medium | Define clear offline behavior in requirements |

---

## Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-16 | Initial EPIC created | BA Team |

---

**Document Status:** Approved
**Approval Required:** No - EPIC approved, stories can be created
**Last Updated:** 2026-04-16
