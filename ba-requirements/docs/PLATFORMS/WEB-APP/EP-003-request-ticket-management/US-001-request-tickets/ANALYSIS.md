---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
status: draft
version: "1.0"
last_updated: "2026-04-01"
add_on_sections: ["Design Context"]
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
    description: "Request Tickets List screen"
    node_id: "3171:3062"
    extraction_date: "2026-04-01"
  - type: figma
    description: "Update Request Ticket screen"
    node_id: "3181:1976"
    extraction_date: "2026-04-03"
---

# Analysis: Request Tickets

**Epic:** EP-003 (Request Ticket Management)
**Story:** US-001-request-tickets
**Status:** Draft

---

## 1. Business Context

### Problem Statement

Organizations need a centralized, general-purpose internal request system that allows employees to submit any kind of request — IT support, office supplies, facility maintenance, HR inquiries, and more. Without such a system, requests are scattered across emails, chat messages, and verbal communication, leading to lost requests, unclear ownership, and no visibility into resolution progress.

### Stakeholders

- **Primary Users:** Administrators and managers who review, track, and resolve request tickets
- **Secondary Users:** All employees who submit request tickets
- **Business Owner:** Operations department leadership

### Business Goals

- Provide a single point of entry for all internal employee requests
- Enable managers/administrators to track, prioritize, and resolve requests efficiently
- Maintain clear audit trail of request status changes
- Categorize requests by type for reporting and resource planning

---

## 2. Scope Definition

### In Scope

- Request tickets list view (search, filters, table, pagination, export)
- Create new request ticket
- Edit request tickets (status-dependent)
- Status management workflow (Open through Closed/Cancelled)
- Delete request tickets (soft delete, role-dependent)
- Role-based data visibility (employees see own; managers/admins see all)
- Sidebar navigation under "Operation" menu group

### Out of Scope

- Request assignment/routing to specific departments or individuals (future enhancement)
- SLA tracking and escalation rules (future enhancement)
- Request ticket comments/thread discussion (future enhancement)
- Automated notifications/email alerts on status changes (future enhancement)
- Request ticket templates per type (future enhancement)
- Mobile-specific request ticket interface
- Reporting dashboard for request ticket analytics

### Dependencies

- **EP-001 US-001 (Authentication):** User must be signed in
- **EP-001 US-004 (Role & Permission Management):** Controls access to list, manage actions
- **EP-001 US-005 (User Management):** Employee data displayed in list (name, department, position)
- **EP-008 US-001/002 (Department/Position):** Filter dropdown options

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-04-01.

### Source Information

| Attribute | Value |
|-----------|-------|
| **Figma File** | [Exnodes HRM](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC) |
| **Frame** | [Request Tickets](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3171-3062) |
| **Node ID** | `3171:3062` |
| **Dimensions** | 1920 x 1080 |
| **Extraction Date** | 2026-04-01 |

### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Request Tickets" | `3171:3065` | Page heading | Designed |
| Search input | `3171:3069` | Multi-field search (320px) | Designed |
| Department filter | `3171:3070` | Filter by department (chip) | Designed |
| Position filter | `3171:3071` | Filter by position (chip) | Designed |
| Status filter | `3171:3072` | Filter by status with count (chip) | Designed |
| Reset button | `3171:3073` | Clear all filters | Designed |
| Export button | `3171:3074` | Export filtered list | Designed |
| Add Request button | `3171:3075` | Create request ticket | Designed |
| Table: Full Name | `3171:3078` | Employee name column (~201px) | Placeholder data |
| Table: Department & Position | `3171:3079` | Two-line: dept + position (~201px) | Placeholder data |
| Table: Request Date | `3171:3080` | Date request was submitted (~160px) | Placeholder data |
| Table: Request Type | `3171:3081` | Category of request (~160px) | Placeholder data |
| Table: Request | `3171:3082` | Request subject/description (~201px) | Placeholder data |
| Table: Status | `3171:3083` | Status badge (~201px) | Placeholder (TBD badges) |
| Table: Action | `3171:3084` | Gear icon per row (~50px) | Designed |
| Pagination | `3171:3085` | Rows per page + page navigation | Designed |

### Design Constraints

- Sidebar: 200px | Content: 1694px
- Search: 320px left-aligned, placeholder "Search by name, email, phone number..."
- 3 filter chips (Department, Position, Status) + Reset button in action bar left section
- Export + "+ Add Request" in action bar right section
- Table: 7 columns spanning full content width
- Department & Position column uses two-line format (dept bold + position muted below) — same pattern as Leave Requests
- Sidebar: "Menu Section" placeholder — to be placed under "Operation" menu group

### Gaps Identified

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Status badge colors not defined | Cannot distinguish status types at a glance | Define colors per status (see Knowledge Confirmation Summary) |
| Gear icon actions not specified | Users won't know available actions per row | Define context-sensitive actions based on status + role |
| Request Type values not defined | Cannot populate filter or form dropdown | Define standard request types with "Other" option |
| Sidebar "Operation" group not visible | No navigation path for Request Tickets | Add "Request Tickets" under "Operation" group in sidebar |
| "Request" column content unclear | Unclear if this is title/subject or full description | Recommend: short subject/title, truncated if needed |

### Screen: Update Request Ticket

**Figma Reference:** [Update Request Ticket](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3181-1976) (node `3181:1976`)
**Extraction Date:** 2026-04-03

**Layout:**
- Sidebar (200px) + Content area (1694px)
- Page title: "Update Request Ticket" (Geist Semibold 24px)
- Header action bar: Cancel (secondary) + Save (primary) — top-right
- Two cards centered 600px: Employee card + Request Info card
- Bottom full-width Save button (600x40, bg #010101)

**Component Inventory:**

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Update Request Ticket" | `3181:2034` | Page heading | Designed |
| Cancel button | `3181:2036` | Return to list | Designed |
| Save button (header) | `3181:2037` | Submit changes | Designed |
| Employee card | `3181:2040` | Employee section (600px) | Designed |
| Employee dropdown | `3181:2044` | Submitter selection field | Designed (empty state shown) |
| Employee hint banner | `3181:2045` | "Select an employee to view their info" | Designed |
| Request Info card | `3181:2048` | Request info section (600px) | Designed |
| "Service type" dropdown | `3181:2052` | Request Type field (576px) | Designed (naming gap) |
| "Request details" textarea | `3181:2053` | Description field (576x76px) | Designed (naming gap) |
| Attachment file input | `3181:2054` | Optional file upload (full width) | Designed |
| Save button (bottom) | `3181:2057` | Submit changes (full-width convenience) | Designed |

**Layout Diagram:**
```
+---------------------------------------------------------------+
|  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]  |
+--------------+------------------------------------------------+
|  [Sidebar]   |  Update Request Ticket         [Cancel] [Save]  |
|  200px       |                                                 |
|              |     +------------------------------------+      |
|              |     | Employee                           |      |
|              |     | * Employee [dropdown, 576px]       |      |
|              |     | [hint: Select an employee to...]   |      |
|              |     +------------------------------------+      |
|              |     | Service Ticket Info [naming gap]   |      |
|              |     | * Service type [dropdown, 576px]   |      |
|              |     | * Request details [textarea, 576px]|      |
|              |     |   Attachment [Choose File]         |      |
|              |     +------------------------------------+      |
|              |     | [             Save               ]  |      |
|              |     +------------------------------------+      |
+--------------+------------------------------------------------+
```

**Design Gaps:**

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Section header "Service Ticket Info" (old name) | Confusing after module rename to "Request Ticket" | Rename to "Request Info" to match Create form |
| Field label "Service type" (old name) | Inconsistent with Create form "Request Type" | Rename to "Request Type" to match Create form |
| Field label "Request details" (unclear scope) | Subject field missing — cannot update short title shown in list | Clarify: "Request details" = Description only; add separate Subject field consistent with Create form |
| Employee shown in empty/unselected state | Edit form should show pre-filled, locked submitter | Design the locked info-card state (employee pre-filled, read-only) |
| No visual for employee-only view | Design only shows manager view | Design the hidden-employee-section variant for employee edit flow |

---

**Document Version:** 1.0
**Last Updated:** 2026-04-03
**Author:** BA Agent
