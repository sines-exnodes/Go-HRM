---
document_type: EPIC
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-002
epic_name: "Leave Management"
status: approved
version: "1.0"
created_date: "2026-04-16"
last_updated: "2026-04-16"
approved_by: "Product Owner"
related_documents:
  - path: "../../WEB-APP/EP-002-leave-management/EPIC.md"
    relationship: reference
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
user_stories:
  - id: US-001
    name: "Leave Requests"
    status: in_progress
    description: "Mobile leave request dashboard, list, create, and status views"
---

# Epic: Leave Management

**Epic ID:** EP-002
**Platform:** Exnodes HRM Mobile (MOBILE-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Enable employees to manage their leave requests on-the-go through the mobile application. The mobile experience prioritizes quick actions, glanceable information, and streamlined workflows optimized for touch interaction.

### Scope

- Leave requests dashboard (summary view)
- Leave request list with filtering
- Create new leave request
- View leave request details
- Leave balance visibility
- Push notifications for status changes

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Leave Requests | Dashboard, list view, create request, view details | In Progress |

### Dependencies

- **EP-001 (Foundation):** Authentication (US-001) — user must be signed in
- **WEB-APP Backend:** Leave management API (shared with web platform)
- **WEB-APP EP-002:** Leave types, approval workflows defined in web platform

### Success Criteria

- Employees can view their leave balance and history on mobile
- Employees can submit leave requests from mobile
- Users receive push notifications for request status changes
- Dashboard provides at-a-glance summary of pending/approved requests

---

## Mobile-Specific Considerations

### Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Primary View | Full table list | Dashboard summary + list |
| Filters | Multi-select chip filters | Simplified filter sheet |
| Create Flow | Full-page form | Step-by-step or bottom sheet |
| Notifications | In-app only | Push notifications |
| Export | CSV/Excel export | Not supported (mobile) |

### Mobile UX Patterns

- Pull-to-refresh for list updates
- Swipe actions on list items
- Bottom sheet for quick actions
- Calendar date picker optimized for touch
- Floating action button for "New Request"

---

## Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Offline request submission | Medium | High | Queue requests for sync when online |
| Calendar picker usability | Low | Medium | Use native date picker components |
| Push notification delivery | Low | Medium | Fallback to in-app notifications |

---

## Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-16 | Initial EPIC created | BA Team |

---

**Document Status:** Approved
**Approval Required:** No - EPIC approved, stories can be created
**Last Updated:** 2026-04-16
