---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-01
detail_name: "Leave Requests List"
parent_requirement: FR-US-001-01
status: draft
version: "1.0"
created_date: 2026-03-27
last_updated: 2026-03-27
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
---

# Detail Requirement: Leave Requests List

**Detail ID:** DR-002-001-01
**Parent Requirement:** FR-US-001-01
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **manager or administrator with leave management permission**, I want to **view all leave requests in a searchable, filterable, paginated list showing employee details, dates, leave type, and approval status**, so that **I can monitor team leave activity, identify pending requests requiring action, and manage leave approvals efficiently**.

**Purpose:** Provide a centralized view of all leave requests across the organization. The Leave Requests list is the primary entry point for leave management — from here, authorized users can search for requests, filter by department/position/status, review request details at a glance, and access actions (create, edit, approve/reject, cancel) via the gear icon.

**Target Users:**
- Any role with **leave view permission** — can browse, search, filter, paginate, and export the list
- Any role with **leave management permission** — can additionally access "+ Add New" and gear icon actions (Edit, Approve, Reject, Cancel) based on request status and role context
- **Employees** — can create their own leave requests; admin/managers can create on behalf of others

**Key Functionality:**
- Searchable table with 9 columns including a unique two-line Department & Position column
- 3 multi-select filter chips (Department, Position, Status) with in-dropdown search
- Status-driven color badges (Pending, Approved, Rejected, Cancelled)
- Context-sensitive gear icon actions per row (actions vary by request status and user's role)
- Export of currently filtered list
- "+ Add New" to create a leave request

---

## 2. User Workflow

**Entry Point:** Sidebar navigation → HRM > Leave Requests

**Preconditions:**
- User is signed in (US-001)
- User's role has leave view permission (US-004)

**Main Flow:**
1. User clicks "Leave Requests" under HRM section in the sidebar
2. System loads the Leave Requests page
3. System displays the leave requests table with 9 columns, paginated (default 10 rows)
4. Requests are listed by most recent submission date first (newest at top)
5. User browses or takes one of the available actions

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Search** | User types in search box → list filters by full name, email, or phone number with debounce (300ms) → pagination resets to page 1 |
| **Clear Search** | User clears search box → full list restored (respecting active filters) → pagination resets to page 1 |
| **Filter: Department** | User clicks Department chip → dropdown with search → selects one or more → list filters → page 1 |
| **Filter: Position** | Same pattern as Department |
| **Filter: Status** | Same pattern; chip shows count e.g., "Status (2)" |
| **Reset** | User clicks Reset → all filters and search cleared → full list → page 1 |
| **Combine** | Filters are additive: OR within same filter, AND across filters, AND with search |
| **Paginate** | User changes page or rows per page → list updates |
| **Export** | User clicks Export → downloads currently filtered list |
| **+ Add New** | User clicks "+ Add New" → navigates to Create Leave Request page *(DR-002-001-02)* |
| **Gear → Edit** | Visible only for Pending requests (own or admin) → navigates to Edit page *(DR-002-001-03)* |
| **Gear → Approve** | Visible only for Pending requests, manager/admin only → triggers approval *(DR-002-001-04)* |
| **Gear → Reject** | Visible only for Pending requests, manager/admin only → triggers rejection *(DR-002-001-04)* |
| **Gear → Cancel** | Visible for Pending/Approved requests (own or admin) → triggers cancellation *(DR-002-001-05)* |

**Gear Icon Action Matrix (Status × Role):**

| Request Status | Requester (own request) | Manager/Admin |
|---------------|------------------------|---------------|
| Pending | Edit, Cancel | Edit, Approve, Reject, Cancel |
| Approved | Cancel | Cancel |
| Rejected | — (no actions) | — (no actions) |
| Cancelled | — (no actions) | — (no actions) |

**Exit Points:**
- **+ Add New** → Navigate to Create Leave Request page
- **Edit** → Navigate to Edit Leave Request page
- **Approve/Reject/Cancel** → Stays on list; status updated inline; list refreshes
- **Export** → File downloads; stays on list

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Search On | Trigger | Mandatory | Placeholder | Description |
|------------|------------|-----------|---------|-----------|-------------|-------------|
| Search | Text input (320px) | Full name, email, phone number | On change with 300ms debounce | No | "Search by name, email, phone number..." | Multi-field search — case-insensitive, partial match |

### Filter Elements

| Filter Name | Type | Options Source | Multi-select | Default | Description |
|-------------|------|---------------|-------------|---------|-------------|
| Department | Dropdown chip (135px) | Departments from EP-008 US-001 | Yes | None (show all) | With in-dropdown search |
| Position | Dropdown chip (109px) | Positions from EP-008 US-002 | Yes | None (show all) | With in-dropdown search |
| Status | Dropdown chip (120px) | Pending, Approved, Rejected, Cancelled | Yes | None (show all) | Shows count e.g., "(2)"; with in-dropdown search |

### Interaction Elements

| Element | Type | Visible To | State/Condition | Trigger Action | Description |
|---------|------|-----------|-----------------|----------------|-------------|
| Reset | Icon button (95px) | All with view permission | Visible when any filter/search active; hidden otherwise | Clears all filters and search | Refresh icon |
| Export | Button (secondary, 100px) | All with view permission | Always visible | Downloads currently filtered list | Exports result set |
| + Add New | Button (primary, 116px) | Management permission only | Always visible for authorized users | Navigate to Create Leave Request | Creates new request |
| Gear icon (per row) | Icon button | Context-sensitive | Actions vary by status × role (see matrix above) | Opens action dropdown | Hidden/disabled for Rejected and Cancelled |
| Rows per page | Dropdown | All with view permission | Default: 10 | Changes page size (10, 25, 50) | Pagination control |
| Page navigation | Button group | All with view permission | Always visible | Navigate between pages | Numbered page buttons |

---

## 4. Data Display

### Information Shown to User

| Data | Column | Width | Format | Empty State | Business Meaning |
|------|--------|-------|--------|-------------|-----------------|
| Full Name | Column 1 | ~201px | Text | — | Employee who submitted the request |
| Department & Position | Column 2 | ~201px | Two-line: dept (14px medium) + position (12px muted) | "—" / "—" | Organizational context of the requester |
| From Date | Column 3 | ~160px | Date (DD/MM/YYYY) | — | Leave start date |
| To Date | Column 4 | ~160px | Date (DD/MM/YYYY) | — | Leave end date |
| Total Days | Column 5 | ~126px | Number (supports decimals: 0.5, 1.5, etc.) | — | Calculated duration with half-day support |
| Leave Type | Column 6 | ~201px | Text | — | Annual, Sick, Personal, Maternity, or Unpaid |
| Leave Period | Column 7 | ~201px | Text | — | Full Day, Morning Half, or Afternoon Half |
| Status | Column 8 | ~201px | Colored badge | — | Current approval status |
| Action | Column 9 | ~201px | Gear icon (context-sensitive) | — | Actions vary by status × role |

### Status Badge Colors (Confirmed)

| Status | Badge Color | Text |
|--------|------------|------|
| Pending | Amber/yellow background | "Pending" |
| Approved | Green background | "Approved" |
| Rejected | Red background | "Rejected" |
| Cancelled | Gray background | "Cancelled" |

### Leave Type Values (Confirmed)

| Leave Type | Description |
|------------|-------------|
| Annual | Regular annual leave / vacation |
| Sick | Sick leave |
| Personal | Personal leave / day off |
| Maternity | Maternity leave |
| Unpaid | Unpaid leave |

### Leave Period Values (Confirmed)

| Leave Period | Description | Day Count |
|-------------|-------------|-----------|
| Full Day | Entire day off | 1.0 per day |
| Morning Half | Morning half-day off | 0.5 per day |
| Afternoon Half | Afternoon half-day off | 0.5 per day |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens or data being fetched | Skeleton rows in table |
| Populated | Leave requests exist | Table with data rows, filters, pagination |
| Empty | No leave requests in the system | "No leave requests found" with "+ Add New" CTA (management only) |
| No Results (search) | Search returns zero matches | "No leave requests match your search" with "Clear search" link |
| No Results (filter) | Filters return zero matches | "No leave requests match the selected filters" with "Reset filters" link |
| No Results (combined) | Search + filters return zero | "No leave requests match your search and filters" with "Reset all" link |
| Filter active | One or more filters selected | Active chips highlighted; Reset visible; Status chip shows count |

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Leave Requests                                  │
│  200px       │                                                  │
│  Users Mgmt  │  [Search 320px] [Dept] [Position] [Status(2)]   │
│  > Users     │  [Reset]                    [Export] [+ Add New] │
│  > Roles     │                                                  │
│  > Perms     │  ┌────────┬──────────┬──────┬──────┬─────┬─────┐│
│              │  │FullName│Dept&Pos  │ From │ To   │Total│Leave ││
│  HRM        │  │        │          │ Date │ Date │Days │Type  ││
│  > Leave ◄  │  ├────────┼──────────┼──────┼──────┼─────┼─────┤│
│              │  │ Name   │Frontend  │ Date │ Date │ 2.5 │Annual││
│  Menu Sect.  │  │        │Team Lead │      │      │     │      ││
│  > ...      │  │ Name   │Frontend  │ Date │ Date │ 1.0 │Sick  ││
│              │  │        │Member    │      │      │     │      ││
│              │  └────────┴──────────┴──────┴──────┴─────┴─────┘│
│              │  (cont: Leave Period | Status [badge] | Action ⚙)│
│              │                                                  │
│              │  Rows per page [10▼]  Page 1 of 10  1 2 [3] 4 > │
└──────────────┴──────────────────────────────────────────────────┘
```

> **Note:** Figma design at node `3123:7904`. See ANALYSIS.md Design Context [ADD-ON] for full component inventory.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Leave Requests page displays with page title "Leave Requests" in Geist Semibold 24px
- **AC-02:** Table has 9 columns: Full Name, Department & Position, From Date, To Date, Total Days, Leave Type, Leave Period, Status, Action
- **AC-03:** Requests are listed by most recent submission date first (newest at top)
- **AC-04:** Department & Position column displays two lines: department name (14px medium) on top, position (12px muted) below
- **AC-05:** Status column displays colored badges: Pending (amber), Approved (green), Rejected (red), Cancelled (gray)

**Search:**
- **AC-06:** Search field filters across full name, email, and phone number simultaneously with ~300ms debounce
- **AC-07:** Search is case-insensitive and uses partial/contains matching
- **AC-08:** When search returns no results, "No leave requests match your search" is displayed with a clear search option
- **AC-09:** Clearing the search field restores the full list (respecting active filters) and resets to page 1

**Filters:**
- **AC-10:** Department filter chip opens a multi-select dropdown with in-dropdown search
- **AC-11:** Position filter chip opens a multi-select dropdown with in-dropdown search
- **AC-12:** Status filter chip opens a multi-select dropdown listing: Pending, Approved, Rejected, Cancelled
- **AC-13:** Active filter chips show the count of selected values (e.g., "Status (2)")
- **AC-14:** Filters are additive: OR within same filter, AND across filters, AND with search
- **AC-15:** When filters return no results, "No leave requests match the selected filters" is displayed with a "Reset filters" link
- **AC-16:** Reset button clears all filters and search; hidden when no filters/search are active
- **AC-17:** Each filter dropdown includes a search field to filter options

**Pagination:**
- **AC-18:** List paginates with default 10 rows per page; user can change to 25 or 50
- **AC-19:** Pagination controls are hidden when total results ≤ current rows per page
- **AC-20:** When search or filters are applied or cleared, pagination resets to page 1

**Export:**
- **AC-21:** Export button downloads the currently filtered/searched result set
- **AC-22:** Export button shows a brief loading state while file is being prepared

**Context-Sensitive Gear Icon Actions:**
- **AC-23:** Gear icon dropdown actions vary based on request status AND user's role
- **AC-24:** For Pending requests — requester sees: Edit, Cancel
- **AC-25:** For Pending requests — manager/admin sees: Edit, Approve, Reject, Cancel
- **AC-26:** For Approved requests — requester sees: Cancel
- **AC-27:** For Approved requests — manager/admin sees: Cancel
- **AC-28:** For Rejected requests — no actions available (gear icon hidden or disabled)
- **AC-29:** For Cancelled requests — no actions available (gear icon hidden or disabled)
- **AC-30:** Edit navigates to Edit Leave Request page
- **AC-31:** Approve/Reject triggers inline status update — list refreshes with new status badge
- **AC-32:** Cancel triggers confirmation dialog → status updated inline → list refreshes

**Permissions & Visibility:**
- **AC-33:** "+ Add New" button is visible only to users with leave management permission
- **AC-34:** Users with view-only permission can browse, search, filter, paginate, and export but cannot see "+ Add New" or gear icon actions
- **AC-35:** Export button is visible to all users with view permission

**Empty & Loading States:**
- **AC-36:** When no leave requests exist, empty state shows "No leave requests found" with "+ Add New" CTA (management permission only)
- **AC-37:** Skeleton rows are displayed while data is loading

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Default load | Navigate to Leave Requests | All requests shown, newest first, page 1 | High |
| Search by name | Type "Henry" | Requests from employees matching "Henry" | High |
| Search no match | Type "zzz" | "No leave requests match your search" | High |
| Filter by status | Select "Pending" | Only pending requests shown; chip shows "(1)" | High |
| Filter multiple statuses | Select "Pending" + "Approved" | Pending OR Approved requests | High |
| Combined search + filter | Search "john" + Status=Pending | Pending requests matching "john" | High |
| Reset all | Filters + search active, click Reset | All cleared, full list restored | Medium |
| Export filtered | Filters active, click Export | Only filtered results downloaded | High |
| Gear — Pending (requester) | Click gear on own pending request | Shows Edit, Cancel | High |
| Gear — Pending (admin) | Click gear on pending request | Shows Edit, Approve, Reject, Cancel | High |
| Gear — Approved (requester) | Click gear on own approved request | Shows Cancel only | High |
| Gear — Rejected | Click gear on rejected request | No actions / gear hidden | Medium |
| Gear — Cancelled | Click gear on cancelled request | No actions / gear hidden | Medium |
| Dept & Position column | Request from Frontend Team Leader | Two lines: "Frontend" (bold) + "Team Leader" (muted) | Medium |
| Half-day Total Days | Morning Half request, 1 day | Total Days shows "0.5" | Medium |
| Empty state | No requests in system | Empty state with Add New CTA | Medium |
| Pagination reset | On page 3, apply filter | Returns to page 1 | Medium |
| Unauthorized | View-only user | No Add New, no gear icons | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Access to the Leave Requests page is controlled by view permission configured in US-004 — users without view permission cannot access this page
- **SR-02:** The "+ Add New" button and gear icon actions are only displayed to roles granted leave management permission via US-004
- **SR-03:** Search applies across 3 fields simultaneously: full name, email, phone number. A match in ANY field qualifies the row.
- **SR-04:** Search matching is case-insensitive with partial/contains matching
- **SR-05:** Filters are multi-select — OR logic within the same filter, AND logic across different filters. Search is ANDed with all filters.
- **SR-06:** When search or any filter is applied or cleared, pagination resets to page 1
- **SR-07:** The default display order is most recent submission date first (newest at top)
- **SR-08:** Export produces the currently filtered/searched result set
- **SR-09:** Each filter dropdown dynamically loads options: Department from EP-008 US-001, Position from EP-008 US-002, Status from system-defined values (Pending, Approved, Rejected, Cancelled)
- **SR-10:** Filter dropdown options include a search field — filtered client-side (no server round-trip)
- **SR-11:** The Reset button is only visible when at least one filter or search is active
- **SR-12:** Gear icon actions are context-sensitive based on both request status AND user's role:

| Status | Requester (own request) | Manager/Admin |
|--------|------------------------|---------------|
| Pending | Edit, Cancel | Edit, Approve, Reject, Cancel |
| Approved | Cancel | Cancel |
| Rejected | (none) | (none) |
| Cancelled | (none) | (none) |

- **SR-13:** For rows with no available actions (Rejected, Cancelled), the gear icon is hidden or disabled
- **SR-14:** Approve and Reject actions update the status inline on the list — no page navigation; list refreshes to reflect new badge
- **SR-15:** Cancel action requires a confirmation dialog before executing
- **SR-16:** Leave request status lifecycle: `Pending → Approved`, `Pending → Rejected`, `Pending → Cancelled`, `Approved → Cancelled`. No other transitions allowed.
- **SR-17:** Only pending leave requests can be edited
- **SR-18:** Total Days is calculated using inclusive counting (From Date to To Date including both days) with half-day support. Leave Period determines the day count: Full Day = 1.0 per day, Morning Half / Afternoon Half = 0.5 per day.

**State Transitions:**
```
[List loads] → [Skeleton] → [Data fetched] → [Table displayed (newest first)]
[User types in search] → [Debounce 300ms] → [List filters across 3 fields] → [Page 1]
[User clears search] → [List restores (respecting active filters)] → [Page 1]
[User selects filter value] → [List filters immediately] → [Page 1]
[User clicks Reset] → [All filters + search cleared] → [Full list] → [Page 1]
[User clicks Export] → [Export loading] → [File downloaded (filtered set)]
[User clicks Gear → Approve] → [Status updated to Approved] → [List refreshes; badge green]
[User clicks Gear → Reject] → [Status updated to Rejected] → [List refreshes; badge red]
[User clicks Gear → Cancel] → [Confirmation dialog] → [Confirm → status Cancelled; list refreshes]
[User clicks Gear → Edit] → [Navigate to Edit Leave Request page]
```

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls view and management permissions; determines role-based gear icon actions
- **US-005 (User Management):** Employee data displayed (Full Name, Department, Position)
- **EP-008 US-001 (Department Management):** Department filter options and employee-department assignments
- **EP-008 US-002 (Position Management):** Position filter options and employee-position assignments

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Search input uses a debounce delay (~300ms) to reduce unnecessary filtering on every keystroke
- **UX-02:** Skeleton rows shown during page load to prevent layout shift
- **UX-03:** Filter chips visually indicate active state when values are selected
- **UX-04:** Active filter chips show count of selected values (e.g., "Status (2)")
- **UX-05:** Each filter dropdown includes a search field at the top for quick option finding
- **UX-06:** Reset button only appears when filters/search are active
- **UX-07:** Status badges are color-coded for instant visual scanning — Pending (amber) stands out to draw attention to requests needing action
- **UX-08:** Department & Position column uses two-line format (dept bold + position muted) — packs two data points into one column
- **UX-09:** Gear icon is hidden/disabled for Rejected and Cancelled requests — no empty dropdown
- **UX-10:** Approve/Reject actions update status inline without page navigation — rapid batch processing
- **UX-11:** Cancel action requires confirmation dialog — prevents accidental cancellation
- **UX-12:** Export button shows brief loading/spinner state — prevents double-click
- **UX-13:** Pagination controls hidden when total results fit on one page
- **UX-14:** Default sort (newest first) ensures recently submitted requests are immediately visible

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Full 9-column layout: sidebar 200px + content area |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through search, filter chips, Reset, Export, Add New, gear icons, pagination
- [x] Screen reader compatible — table column headers, button labels, filter state announced
- [x] Filter dropdowns accessible via keyboard — arrow keys to navigate, Enter to select, Escape to close
- [x] Status badges have accessible labels (not color-only — text included)
- [x] Sufficient color contrast — badge colors meet WCAG 2.1 AA standards
- [x] Focus indicators visible on all interactive elements

**Design References:**
- Figma: [Leave Requests](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7904) (node `3123:7904`)
- Design tokens: primary `#171717`, background `#ffffff`, border `#e5e5e5`, muted foreground `#737373`, font: Geist
- Pattern reference: User List (DR-001-005-01) for filter + search pattern

---

## 8. Additional Information

### Out of Scope
- Leave request detail/drill-down view
- Column sorting by user interaction (newest-first default only)
- Advanced search syntax or field-specific search operators
- Bulk actions (select multiple and approve/reject at once)
- Leave balance display on the list page (future enhancement)
- Leave type configuration/CRUD (future epic)
- Team leave calendar view (future enhancement)
- Notification system for status changes (future enhancement)
- Mobile or tablet responsive layout

### Open Questions (Resolved)

| Question | Answer | Confirmed By |
|----------|--------|-------------|
| Status values | Pending, Approved, Rejected, Cancelled | Product Owner |
| Status badge colors | Amber, Green, Red, Gray (exact tokens pending Design Team) | Product Owner |
| Leave Type values | Annual, Sick, Personal, Maternity, Unpaid | Product Owner |
| Leave Period values | Full Day, Morning Half, Afternoon Half | Product Owner |
| Total Days calculation | Inclusive counting + half-day support | Product Owner |
| Sidebar navigation | HRM is menu group; Leave Requests is menu item under it; other HRM items TBD | Product Owner |
| Who can create requests | Employees create own; admin/managers can create on behalf of others | Product Owner |

### Open Questions (Pending)

- [ ] **Export format:** CSV or Excel (.xlsx)? Which columns included? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Status badge exact tokens:** Confirm exact color hex values for each status badge. — **Owner:** Design Team — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-02: Create Leave Request (planned) | Triggered from "+ Add New" button |
| DR-002-001-03: Edit Leave Request (planned) | Triggered from gear → Edit (Pending only) |
| DR-002-001-04: Approve/Reject (planned) | Triggered from gear → Approve/Reject (Pending, manager/admin) |
| DR-002-001-05: Cancel Leave Request (planned) | Triggered from gear → Cancel (Pending/Approved) |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls view/management permissions; role-based actions |
| US-005: User Management | Employee data (Full Name, Department, Position) |
| EP-008 US-001/002: Department & Position | Filter dropdown options |

### Notes
- This is the **first module outside Foundation and Organization Data** — introduces the HRM sidebar section and EP-002 epic.
- The **status-driven workflow** is a new pattern. Other lists have uniform gear actions; Leave Requests varies actions per row based on status × role.
- The **two-line Department & Position column** is unique — packing two data points into one column.
- **Approve/Reject execute inline** on the list — no separate page. Enables rapid batch processing of pending requests.
- **Employees can create their own requests**, and admin/managers can create on behalf of others — the "+ Add New" flow will need to handle both scenarios.
- **Half-day support** means Total Days can be decimal (0.5, 1.5, 2.0, etc.) — the Leave Period column determines the day count per day.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-27 | BA Agent | Initial draft — Figma design context from node 3123:7904 |
