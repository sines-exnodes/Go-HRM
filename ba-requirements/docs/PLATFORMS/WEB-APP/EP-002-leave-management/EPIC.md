---
document_type: EPIC
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
epic_name: "Leave Management"
status: approved
version: "1.0"
created_date: "2026-03-27"
last_updated: "2026-03-27"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
  - path: "../EP-008-organization-data/EPIC.md"
    relationship: dependency
user_stories:
  - id: US-001
    name: "Leave Requests"
    status: in_progress
    description: "Leave request list, create, edit, approve/reject, cancel workflows"
---

# Epic: Leave Management

**Epic ID:** EP-002
**Platform:** Exnodes HRM (WEB-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Provide a comprehensive leave management system that allows employees to request time off and administrators/managers to review, approve, or reject leave requests. The system tracks leave balances, leave types, and provides visibility into team availability.

### Scope

- Leave request submission and management
- Leave approval/rejection workflows
- Leave type configuration
- Leave balance tracking
- Team leave calendar visibility

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Leave Requests | Leave request list view with search, filters, pagination; create, edit, approve/reject, cancel workflows | In Progress |

### Dependencies

- **EP-001 (Foundation):** Authentication (US-001), Role & Permission Management (US-004), User Management (US-005)
- **EP-008 (Organization Data):** Department Management (US-001), Position Management (US-002) — used in leave request filters and employee context

### Success Criteria

- Employees can submit leave requests with all required details
- Managers/administrators can review and approve/reject requests efficiently
- Leave balances are accurately tracked and enforced
- Team availability is visible to prevent scheduling conflicts

---

**Document Version:** 1.0
**Last Updated:** 2026-03-27
**Author:** BA Team
