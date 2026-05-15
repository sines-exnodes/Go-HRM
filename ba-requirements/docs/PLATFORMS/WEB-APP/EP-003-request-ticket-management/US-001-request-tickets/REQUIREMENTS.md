---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
status: draft
version: "1.0"
last_updated: "2026-04-01"
add_on_sections: ["UI Specifications"]
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
---

# Requirements: Request Tickets

**Epic:** EP-003 (Request Ticket Management)
**Story:** US-001-request-tickets
**Status:** Draft

---

## User Stories

### US-001-01: View Request Tickets List
As an **administrator/manager**, I want to **view all request tickets in a searchable, filterable, paginated list**, so that **I can monitor and manage employee requests across the organization**.

### US-001-02: Create Request Ticket
As an **authorized user**, I want to **submit a new request ticket with type, subject, and details**, so that **my request is formally recorded and routed for handling**.

### US-001-03: Edit Request Ticket
As an **authorized user**, I want to **edit a request ticket that is still in an editable status**, so that **I can correct or update details before it is resolved**.

### US-001-04: Manage Request Ticket Status
As a **manager/administrator**, I want to **change the status of a request ticket through its lifecycle**, so that **progress is tracked and employees are informed**.

### US-001-05: Delete Request Ticket
As an **authorized user**, I want to **delete a request ticket that is no longer needed**, so that **the list remains clean and relevant**.

---

## Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-001-01 | Request tickets list view | Critical | Table with 7 columns: Full Name, Department & Position, Request Date, Request Type, Request, Status, Action |
| FR-US-001-02 | Search | High | Multi-field: employee name, email, phone number. Case-insensitive, partial match, 300ms debounce |
| FR-US-001-03 | Filter: Department | High | Multi-select dropdown with in-dropdown search |
| FR-US-001-04 | Filter: Position | High | Multi-select dropdown with in-dropdown search |
| FR-US-001-05 | Filter: Status | High | Multi-select dropdown; shows count of selected values |
| FR-US-001-06 | Reset filters | Medium | Clears all filters and search; visible only when filters/search active |
| FR-US-001-07 | Pagination | Medium | Default 10 rows per page; options: 10, 25, 50 |
| FR-US-001-08 | Export | Medium | Export currently filtered list |
| FR-US-001-09 | Create request ticket | Critical | Full-page form with request details |
| FR-US-001-10 | Edit request ticket | High | Edit tickets in editable statuses only |
| FR-US-001-11 | Status management | Critical | Context-sensitive status transitions based on current status and role |
| FR-US-001-12 | Delete request ticket | High | Soft delete with role-dependent scope |
| FR-US-001-13 | Action menu per row | High | Gear icon with context-sensitive options based on status and role |
| FR-US-001-14 | Data visibility | High | Employees see only own tickets; managers/admins see all |
---

## Business Rules

- **BR-US-001-01:** Request tickets have a status lifecycle: Open -> In Progress -> Resolved -> Closed; Open -> Cancelled; In Progress -> On Hold -> In Progress
- **BR-US-001-02:** Employees can only view and manage their own request tickets
- **BR-US-001-03:** Managers/administrators can view and manage all request tickets
- **BR-US-001-04:** Only tickets in Open status can be edited by the submitting employee
- **BR-US-001-05:** Managers/admins can edit tickets in Open or In Progress status
- **BR-US-001-06:** Employees can delete only their own tickets in Open status (soft delete)
- **BR-US-001-07:** Managers/admins can delete any ticket regardless of status (soft delete)
- **BR-US-001-08:** Access to the request tickets list is controlled by role permission (US-004)
- **BR-US-001-09:** Request Type must include an "Other" option for uncategorized requests

---

## Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | List loads with all filters and search within 2 seconds | Standard HRM page load |
| Data Integrity | Status transitions enforced server-side | No invalid state changes |
| Access Control | Permission-based visibility of actions and data | Enforced via US-004 |

---

## 11. UI Specifications [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-04-01.

### Screen: Request Tickets List

**Figma Reference:** [Request Tickets](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3171-3062) (node `3171:3062`)

**Layout:**
- Sidebar (200px) + Content area (1694px)
- Page title: "Request Tickets" (Geist Semibold 24px)
- Action bar: Search (320px) + 3 filter chips (Department, Position, Status) + Reset button | Export + "+ Add Request"
- Table: 7 columns with placeholder data
- Pagination: Rows per page [10], Page 1 of 10

**Table Columns:**

| Column | Width | Format | Notes |
|--------|-------|--------|-------|
| Full Name | ~201px | Text | Employee name |
| Department & Position | ~201px | Two-line: dept name (14px bold) + position (12px muted) | Same two-line format as Leave Requests |
| Request Date | ~160px | Date text | Date request was submitted |
| Request Type | ~160px | Text | Category of request |
| Request | ~201px | Text | Request subject/title (truncated if long) |
| Status | ~201px | Colored badge (TBD) | Status values TBD |
| Action | ~50px | Gear icon | Context-sensitive actions |

**Filter Chips:**
- Department, Position: multi-select dropdowns with in-dropdown search
- Status: multi-select with count indicator (e.g., "Status (2)")
- Reset: icon button, visible only when filters/search active

**Design Gaps:**
- Status badge colors not defined in design (placeholder "TBD")
- Sidebar "Operation" group not visible — "Menu Section" placeholder only
- Gear icon dropdown actions not specified in design
- Request Type values not confirmed
- "Request" column content format unclear (title vs. description)

---

**Document Version:** 1.0
**Last Updated:** 2026-04-01
**Author:** BA Agent
