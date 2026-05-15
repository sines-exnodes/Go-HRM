---
document_type: EPIC
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
epic_name: "Request Ticket Management"
status: approved
version: "1.0"
created_date: "2026-04-01"
last_updated: "2026-04-01"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
  - path: "../EP-008-organization-data/EPIC.md"
    relationship: dependency
user_stories:
  - id: US-001
    name: "Request Tickets"
    status: in_progress
    description: "Request ticket list, create, edit, status management, and delete workflows"
---

# Epic: Request Ticket Management

**Epic ID:** EP-003
**Platform:** Exnodes HRM (WEB-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Provide a general-purpose internal request system that allows employees to submit any kind of request — from IT support to office supplies to facility maintenance — and gives administrators/managers visibility, tracking, and resolution workflows across all request types.

### Scope

- Request ticket submission and management
- Request type categorization
- Status-driven workflow (submission through resolution)
- Role-based visibility (employees see own tickets; managers/admins see all)
- Sidebar placement under "Operation" menu group

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Request Tickets | Request ticket list view with search, filters, pagination; create, edit, status management, and delete workflows | In Progress |

### Dependencies

- **EP-001 (Foundation):** Authentication (US-001), Role & Permission Management (US-004), User Management (US-005)
- **EP-008 (Organization Data):** Department Management (US-001), Position Management (US-002) — used in request ticket filters and employee context

### Success Criteria

- Employees can submit request tickets with all required details
- Managers/administrators can view, track, and manage all request tickets
- Request types are categorized for reporting and routing
- Status lifecycle provides clear visibility into request progress
- Sidebar navigation under "Operation" provides easy access

---

**Document Version:** 1.0
**Last Updated:** 2026-04-01
**Author:** BA Team
