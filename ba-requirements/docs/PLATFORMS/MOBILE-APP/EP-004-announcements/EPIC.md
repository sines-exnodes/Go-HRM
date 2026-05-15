---
document_type: EPIC
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-004
epic_name: "Announcements"
status: approved
version: "1.0"
created_date: "2026-04-28"
last_updated: "2026-04-28"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
  - path: "../../WEB-APP/EP-010-announcements/EPIC.md"
    relationship: cross-platform
user_stories:
  - id: US-001
    name: "Home Announcements"
    status: draft
    description: "Display latest announcements on the mobile home screen"
---

# Epic: Announcements (Mobile)

**Epic ID:** EP-004
**Platform:** Exnodes HRM Mobile (MOBILE-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Enable employees to view company announcements on their mobile devices. Announcements created by HR/Admin via WEB-APP (EP-010) are displayed to employees on the mobile home screen, ensuring important communications reach all staff.

### Scope

- Display latest announcements on home screen
- View announcement details
- Receive push notifications for new announcements
- Mark announcements as read

### Out of Scope

- Creating or editing announcements (handled by WEB-APP EP-010)
- Announcement management (admin functions)
- Replying to announcements

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Home Announcements | Display top 5 latest announcements on mobile home screen | Draft |

### Dependencies

- **Depends on:** EP-001 (Foundation) — Authentication required
- **Cross-platform:** WEB-APP EP-010 (Announcements) — Announcement data source

---

## Stories Overview

### US-001: Home Announcements

**Business Purpose:** Keep employees informed of company announcements by displaying the latest announcements prominently on the mobile home screen.

**Key Deliverables:**
- Top 5 latest announcements widget on home screen
- Tap to view full announcement details
- Visual indicator for unread announcements

**User Value:** Employees stay informed about company news without needing to check a separate section.

**Status:** Draft

**Documentation:** [US-001-home-announcements/](./US-001-home-announcements/)

---

## Success Criteria

### Business Acceptance Criteria

- [ ] **View Announcements**: Employees see top 5 latest announcements on home screen
- [ ] **View Details**: Tapping an announcement opens full details
- [ ] **Unread Indicator**: New/unread announcements are visually distinct
- [ ] **Refresh**: Pull-to-refresh updates the announcement list

---

## Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-28 | Initial epic creation | BA Agent |

---

**Document Version:** 1.0
**Last Updated:** 2026-04-28
**Author:** BA Agent
