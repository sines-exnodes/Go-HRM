---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
status: draft
version: "1.0"
last_updated: "2026-03-27"
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
    description: "Leave Requests List screen"
    node_id: "3123:7904"
    extraction_date: "2026-03-27"
  - type: figma
    description: "Change Leave Quota screen"
    node_id: "3224:2523"
    extraction_date: "2026-04-10"
---

# Analysis: Leave Requests

**Epic:** EP-002 (Leave Management)
**Story:** US-001-leave-requests
**Status:** Draft

---

## 1. Business Context

### Problem Statement

Organizations need a structured way to manage employee leave requests — from submission through approval to tracking. Without a centralized leave management system, time-off tracking becomes inconsistent, approvals are delayed, and team scheduling is unreliable.

### Stakeholders

- **Primary Users:** Administrators and managers who review/approve leave requests
- **Secondary Users:** All employees who submit leave requests
- **Business Owner:** HR department leadership

### Business Goals

- Provide a centralized system for leave request submission and approval
- Enable managers to make informed approval decisions with visibility into team availability
- Maintain accurate leave balance tracking across the organization

---

## 2. Scope Definition

### In Scope

- Leave requests list view (search, filters, table, pagination, export)
- Create new leave request
- Edit pending leave requests
- Approve/reject leave requests (manager/admin)
- Cancel leave requests
- Status-based workflow (Pending → Approved/Rejected; Pending/Approved → Cancelled)

### Out of Scope

- Leave balance configuration and management (future epic)
- Leave type configuration/CRUD (future epic)
- Team leave calendar view (future enhancement)
- Leave accrual rules and automatic balance calculations
- Integration with payroll systems
- Mobile-specific leave request interface

### Dependencies

- **EP-001 US-001 (Authentication):** User must be signed in
- **EP-001 US-004 (Role & Permission Management):** Controls access to list, approve/reject actions
- **EP-001 US-005 (User Management):** Employee data displayed in list (name, department, position)
- **EP-008 US-001/002 (Department/Position):** Filter dropdown options

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-27.

### Source Information

| Attribute | Value |
|-----------|-------|
| **Figma File** | [Exnodes HRM](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC) |
| **Frame** | [Leave Requests](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7904) |
| **Node ID** | `3123:7904` |
| **Dimensions** | 1920 × 1080 |
| **Extraction Date** | 2026-03-27 |

### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Leave Requests" | `3123:7951` | Page heading | Designed |
| Search input | `3123:7955` | Multi-field search (320px) | Designed |
| Department filter | `3123:7956` | Filter by department (135px chip) | Designed |
| Position filter | `3123:7957` | Filter by position (109px chip) | Designed |
| Status filter | `3123:7959` | Filter by status with count (120px chip) | Designed |
| Reset button | `3123:7960` | Clear all filters (95px) | Designed |
| Export button | `3123:7962` | Export filtered list (100px) | Designed |
| Add New button | `3123:7963` | Create leave request (116px) | Designed |
| Table: Full Name | `3123:7965` | Employee name column (~201px) | Placeholder data |
| Table: Department & Position | `3124:2763` | Two-line: dept + position (~201px) | Placeholder data |
| Table: From Date | `3123:7969` | Leave start date (~160px) | Placeholder data |
| Table: To Date | `3123:7966` | Leave end date (~160px) | Placeholder data |
| Table: Total Days | `3123:7967` | Calculated duration (~126px) | Placeholder data |
| Table: Leave Type | `3123:7970` | Type of leave (~201px) | Placeholder data |
| Table: Leave Period | `3124:3729` | Period description (~201px) | Placeholder data |
| Table: Status | `3123:7972` | Status badge (~201px) | Placeholder (TBD badges) |
| Table: Action | `3123:7973` | Gear icon per row (~201px) | Designed |
| Pagination | `3123:7974` | Rows per page + page navigation | Designed |

### Design Constraints

- Sidebar: 200px | Content: 1694px
- Search: 320px left-aligned
- 3 filter chips + Reset button in action bar left section
- Export + Add New in action bar right section
- Table: 9 columns spanning full content width
- Department & Position column uses unique two-line format (dept bold + position muted below)

### Gaps Identified

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Status badge colors not defined | Cannot distinguish status types at a glance | Define colors: Pending (yellow/amber), Approved (green), Rejected (red), Cancelled (gray) |
| Gear icon actions not specified | Users won't know available actions per row | Define context-sensitive actions based on status + role |
| Leave sidebar navigation missing | No "Leave Management" section in sidebar | Add navigation section for Leave Management module |
| Status values not confirmed | Cannot implement status filter | Confirm: Pending, Approved, Rejected, Cancelled |

---

### Screen: Change Leave Quota

> Extracted from Figma design via `/figma-extract` on 2026-04-10.

**Source Information:**

| Attribute | Value |
|-----------|-------|
| **Figma File** | [Exnodes HRM](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC) |
| **Frame** | [Change Leave Quota](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3224-2523) |
| **Node ID** | `3224:2523` |
| **Dimensions** | 1920 x 1080 |
| **Extraction Date** | 2026-04-10 |

**Layout Overview:**

This screen is accessed via the User Details page (EP-001 US-005) left action panel. It uses a two-panel layout consistent with other User Details sub-views.

```
┌─────────────────────────────────────────────────────────────────┐
│  ← User Details  > [Employee Name]                              │
├─────────────────────┬───────────────────────────────────────────┤
│  Overview           │  Change Leave Quota                       │
│  Update Information │                                           │
│  Change User Role   │  [Description text]                       │
│  Change Email       │                                           │
│  Reset Password     │  * Annual Leave Quota                     │
│  Activate/Deactivate│  ┌─────────────────────────────────────┐  │
│ [Leave Quota]       │  │ 12                                  │  │
│  Delete User        │  └─────────────────────────────────────┘  │
│                     │                                           │
│                     │  * Sick Leave Quota                       │
│                     │  ┌─────────────────────────────────────┐  │
│                     │  │ 6                                   │  │
│                     │  └─────────────────────────────────────┘  │
│                     │                                           │
│                     │  ┌─────────────────────────────────────┐  │
│                     │  │              Save                   │  │
│                     │  └─────────────────────────────────────┘  │
├─────────────────────┴───────────────────────────────────────────┤
│      189px                         600px                        │
└─────────────────────────────────────────────────────────────────┘
```

**Component Inventory:**

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "User Details" | `3224:2583` | Page heading with back arrow | Designed |
| Employee breadcrumb | `3224:2586` | Shows target employee name | Designed |
| Left action panel | `3224:2588` | Navigation buttons (8 actions) | Designed |
| Leave Quota button | `3224:2681` | Active state with accent bg | Designed |
| Form card | `3224:2597` | Contains quota fields | Designed |
| Card title | `3224:2599` | "Change Leave Quota" (18px semibold) | Designed |
| Description text | `3224:2600` | Two paragraphs explaining impact | Designed |
| Annual Leave Quota field | `3224:2601` | Number input (pre-filled: 12) | Designed |
| Sick Leave Quota field | `3224:2688` | Number input (pre-filled: 6) | Designed |
| Save button | `3224:2602` | Full-width black button | Designed |

**Design Constraints:**

- Left panel: 189px width with 8 stacked action buttons
- Form card: 600px width with 12px horizontal padding
- Save button: Full-width (600px), black background (#010101), white text
- Both quota fields marked mandatory with asterisk (*)
- Form card has light border and subtle shadow

**Key Observations:**

- Entry point is User Details page (EP-001), not Leave Requests list
- "Leave Quota" button highlighted with accent background when active
- Simple form with only 2 fields — no complex validation UI shown
- Description text provides context before editing
- Consistent with other User Details sub-views (Update Information, Change Email, etc.)

---

**Document Version:** 1.0
**Last Updated:** 2026-04-10
**Author:** BA Agent
