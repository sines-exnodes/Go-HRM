---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
status: draft
version: "1.0"
last_updated: "2026-03-27"
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

# Requirements: Leave Requests

**Epic:** EP-002 (Leave Management)
**Story:** US-001-leave-requests
**Status:** Draft

---

## User Stories

### US-001-01: View Leave Requests List
As an **administrator/manager**, I want to **view all leave requests in a searchable, filterable, paginated list**, so that **I can monitor and manage employee leave across the organization**.

### US-001-02: Create Leave Request
As an **authorized user**, I want to **submit a new leave request with dates, leave type, and period**, so that **my time off is formally recorded and routed for approval**.

### US-001-03: Edit Leave Request
As an **authorized user**, I want to **edit a pending leave request**, so that **I can correct or update details before it is reviewed**.

### US-001-04: Approve/Reject Leave Request
As a **manager/administrator**, I want to **approve or reject a leave request**, so that **the employee is informed and the leave balance is updated accordingly**.

### US-001-05: Cancel Leave Request
As an **authorized user**, I want to **cancel a leave request that is no longer needed**, so that **my leave balance is restored and the request is removed from active tracking**.

---

## Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-001-01 | Leave requests list view | Critical | Table with 9 columns: Full Name, Department & Position, From Date, To Date, Total Days, Leave Type, Leave Period, Status, Action |
| FR-US-001-02 | Search | High | Multi-field: name, email, phone number. Case-insensitive, partial match, 300ms debounce |
| FR-US-001-03 | Filter: Department | High | Multi-select dropdown with in-dropdown search |
| FR-US-001-04 | Filter: Position | High | Multi-select dropdown with in-dropdown search |
| FR-US-001-05 | Filter: Status | High | Multi-select dropdown; shows count of selected values |
| FR-US-001-06 | Reset filters | Medium | Clears all filters and search; visible only when filters/search active |
| FR-US-001-07 | Pagination | Medium | Default 10 rows per page; options: 10, 25, 50 |
| FR-US-001-08 | Export | Medium | Export currently filtered list |
| FR-US-001-09 | Create leave request | Critical | Full-page form with leave details |
| FR-US-001-10 | Edit leave request | High | Edit pending requests only |
| FR-US-001-11 | Approve/reject leave request | Critical | Manager/admin action on pending requests |
| FR-US-001-12 | Cancel leave request | High | User can cancel own pending/approved requests |
| FR-US-001-13 | Action menu per row | High | Gear icon with context-sensitive options based on status and role |

---

## Business Rules

- **BR-US-001-01:** Leave requests have a status lifecycle: Pending → Approved/Rejected; Pending/Approved → Cancelled
- **BR-US-001-02:** Only pending leave requests can be edited
- **BR-US-001-03:** Only pending leave requests can be approved or rejected
- **BR-US-001-04:** Users can cancel their own pending or approved requests; managers/admins can cancel any
- **BR-US-001-05:** Access to the leave requests list is controlled by role permission (US-004)

---

## Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | List loads with all filters and search within 2 seconds | Standard HRM page load |
| Data Integrity | Leave balance updated atomically on approve/cancel | No race conditions |
| Access Control | Permission-based visibility of actions | Enforced via US-004 |

---

## 11. UI Specifications [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-27.

### Screen: Leave Requests List

**Figma Reference:** [Leave Requests](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7904) (node `3123:7904`)

**Layout:**
- Sidebar (200px) + Content area (1694px)
- Page title: "Leave Requests" (Geist Semibold 24px)
- Action bar: Search (320px) + 3 filter chips (Department, Position, Status) + Reset button | Export + "+ Add New"
- Table: 9 columns with placeholder data
- Pagination: Rows per page [10], Page 1 of 10

**Table Columns:**

| Column | Width | Format | Notes |
|--------|-------|--------|-------|
| Full Name | ~201px | Text | Employee name |
| Department & Position | ~201px | Two-line: dept name (14px bold) + position (12px muted) | Unique two-line format |
| From Date | ~160px | Date text | Leave start date |
| To Date | ~160px | Date text | Leave end date |
| Total Days | ~126px | Number text | Calculated duration |
| Leave Type | ~201px | Text | Type of leave (e.g., Annual, Sick) |
| Leave Period | ~201px | Text | Period description |
| Status | ~201px | Colored badge (TBD) | Status values TBD |
| Action | ~201px | Gear icon | Context-sensitive actions |

**Filter Chips:**
- Department, Position: multi-select dropdowns with in-dropdown search
- Status: multi-select with count indicator (e.g., "Status (2)")
- Reset: icon button, visible only when filters/search active

**Design Gaps:**
- Status badge colors not defined in design (placeholder "TBD")
- Leave sidebar navigation section not visible — "Menu Section" placeholder
- Gear icon dropdown actions not specified in design
- Status values not confirmed (expected: Pending, Approved, Rejected, Cancelled)

---

**Document Version:** 1.0
**Last Updated:** 2026-03-27
**Author:** BA Agent
