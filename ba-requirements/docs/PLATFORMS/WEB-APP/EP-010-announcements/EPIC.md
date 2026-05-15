---
document_type: EPIC
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-010
epic_name: "Announcements"
status: approved
version: "1.0"
created_date: "2026-04-25"
last_updated: "2026-04-25"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
user_stories:
  - id: US-001
    name: "Announcement List"
    status: draft
    description: "View, create, edit, and manage company announcements"
---

# Epic: Announcements

**Epic ID:** EP-010
**Platform:** Exnodes HRM (WEB-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Provide administrators and HR with a centralized communication tool to publish company-wide announcements. Employees can view announcements relevant to them, ensuring important information reaches the entire organization efficiently.

### Scope

- Announcement list view with search and filters
- Create, edit, publish announcements
- Target audience selection (all employees, specific departments, roles)
- Announcement scheduling (publish now or schedule for later)
- Read status tracking

### Out of Scope

- Push notifications to mobile devices (handled by MOBILE-APP)
- Email distribution of announcements
- Rich media attachments (images, videos) - future enhancement
- Announcement analytics/reporting - future enhancement

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Announcement List | View, create, edit, and manage company announcements | Draft |

### Dependencies

- **Depends on:** EP-001 (Foundation) — Authentication and permissions required
- **Depends on:** EP-008 (Organization Data) — Department data for targeting

---

## Stories Overview

### US-001: Announcement List

**Business Purpose:** Enable administrators and HR to manage company announcements from a centralized list view, with capabilities to create, edit, publish, and archive announcements.

**Key Deliverables:**
- Announcement list with search and filters
- Create new announcement form
- Edit existing announcement
- Publish/unpublish actions
- Delete/archive announcements

**User Value:** HR and administrators can efficiently communicate with the organization through a structured announcement system.

**Status:** Draft

**Documentation:** [US-001-announcement-list/](./US-001-announcement-list/)

---

## Success Criteria

### Business Acceptance Criteria

- [ ] **View Announcements**: Users can view a list of announcements
- [ ] **Create Announcement**: Admin/HR can create new announcements
- [ ] **Edit Announcement**: Admin/HR can edit draft announcements
- [ ] **Publish Announcement**: Admin/HR can publish announcements
- [ ] **Access Control**: Only authorized roles can manage announcements

---

## Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-25 | Initial epic creation | BA Agent |

---

**Document Version:** 1.0
**Last Updated:** 2026-04-25
**Author:** BA Agent
