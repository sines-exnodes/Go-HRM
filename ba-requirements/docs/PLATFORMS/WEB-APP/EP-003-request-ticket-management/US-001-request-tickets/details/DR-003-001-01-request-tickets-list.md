---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
detail_id: DR-003-001-01
detail_name: "Request Tickets List"
parent_requirement: FR-US-001-01
status: draft
version: "1.0"
created_date: 2026-04-01
last_updated: 2026-04-01
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
---

# Detail Requirement: Request Tickets List

**Detail ID:** DR-003-001-01
**Parent Requirement:** FR-US-001-01
**Story:** US-001-request-tickets
**Epic:** EP-003 (Request Ticket Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **manager or administrator with request ticket management permission**, I want to **view all request tickets in a searchable, filterable, paginated list showing employee details, request type, subject, and current status**, so that **I can monitor employee requests across the organization, track resolution progress, and take appropriate actions based on each ticket's status**.

As an **employee with request ticket access**, I want to **view my own request tickets in the same list format**, so that **I can track the progress of my submitted requests and take actions such as closing or reopening resolved tickets**.

**Purpose:** Provide a centralized view of request tickets with role-based data visibility. The Request Tickets list is the primary entry point for request ticket management — from here, authorized users can search, filter, review request details at a glance, and access lifecycle actions (create, edit, status transitions, cancel, delete) via the gear icon. Employees see only their own tickets; managers/admins see all tickets across the organization.

**Target Users:**
- Any role with **request ticket view permission** — can browse, search, filter, paginate, and export the list
- Any role with **request ticket management permission** — can additionally access gear icon actions (Edit, In Progress, On Hold, Resume, Resolve, Close, Reopen, Cancel, Delete) based on ticket status and role context
- **Employees** — can view their own tickets, create new tickets, and perform submitter-exclusive actions (Close, Reopen) on their own tickets
- **Managers/Administrators** — can view all tickets and perform administrative actions including status transitions and deletion

**Key Functionality:**
- Searchable table with 7 columns including a two-line Department & Position column
- 3 multi-select filter chips (Department, Position, Status) with in-dropdown search — no Request Type filter chip (follows Figma design)
- Status-driven color badges for 6 statuses (Open, In Progress, On Hold, Resolved, Closed, Cancelled)
- Context-sensitive gear icon actions per row (actions vary by ticket status and user's role, with submitter-exclusive actions)
- Request column shows subject/title text, truncated with tooltip on hover for long text
- Export of currently filtered list
- "+ Add Request" to create a new request ticket (visible to all authorized users including employees)
- Role-based data visibility: employees see only their own tickets; managers/admins see all

---

## 2. User Workflow

**Entry Point:** Sidebar navigation → Operation > Request Tickets

**Preconditions:**
- User is signed in (US-001)
- User's role has request ticket view permission (US-004)

**Main Flow:**
1. User clicks "Request Tickets" under Operation section in the sidebar
2. System loads the Request Tickets page
3. System applies data visibility scoping: employees see own tickets only; managers/admins see all
4. System displays the request tickets table with 7 columns, paginated (default 10 rows)
5. Tickets are listed by most recent submission date first (newest at top)
6. User browses or takes one of the available actions

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Search** | User types in search box → list filters by employee name, email, or phone number with debounce (300ms) → pagination resets to page 1 |
| **Clear Search** | User clears search box → full list restored (respecting active filters and data visibility) → pagination resets to page 1 |
| **Filter: Department** | User clicks Department chip → dropdown with search → selects one or more → list filters → page 1 |
| **Filter: Position** | Same pattern as Department |
| **Filter: Status** | Same pattern; chip shows count e.g., "Status (2)"; options: Open, In Progress, On Hold, Resolved, Closed, Cancelled |
| **Reset** | User clicks Reset → all filters and search cleared → full list (within data visibility scope) → page 1 |
| **Combine** | Filters are additive: OR within same filter, AND across filters, AND with search |
| **Paginate** | User changes page or rows per page → list updates |
| **Export** | User clicks Export → downloads currently filtered list |
| **+ Add Request** | User clicks "+ Add Request" → navigates to Create Request Ticket page *(DR-003-001-02)* |
| **Gear → Edit** | Submitter: Open only; Manager/Admin: Open or In Progress → navigates to Edit page *(DR-003-001-03)* |
| **Gear → In Progress** | Manager/Admin only, Open tickets → status changes to In Progress *(DR-003-001-04)* |
| **Gear → On Hold** | Manager/Admin only, In Progress tickets → status changes to On Hold *(DR-003-001-04)* |
| **Gear → Resume** | Manager/Admin only, On Hold tickets → status changes back to In Progress *(DR-003-001-04)* |
| **Gear → Resolve** | Manager/Admin only, In Progress tickets → status changes to Resolved *(DR-003-001-04)* |
| **Gear → Close** | Submitter only, Resolved tickets → confirms resolution, status changes to Closed *(DR-003-001-04)* |
| **Gear → Reopen** | Submitter only, Resolved tickets → resolution unsatisfactory, status reverts to Open *(DR-003-001-04)* |
| **Gear → Cancel** | Submitter: Open, In Progress, On Hold; Manager/Admin: any non-terminal → confirmation dialog → status Cancelled *(DR-003-001-05)* |
| **Gear → Delete** | Manager/Admin only, Resolved tickets only → confirmation dialog → soft delete *(DR-003-001-06)* |

**Gear Icon Action Matrix (Status x Role):**

| Ticket Status | Submitter (own ticket) | Manager/Admin |
|---------------|----------------------|---------------|
| Open | Edit, Cancel | Edit, In Progress, Cancel |
| In Progress | Cancel | On Hold, Resolve, Cancel |
| On Hold | Cancel | Resume, Cancel |
| Resolved | Close, Reopen | Delete |
| Closed | — (no actions) | — (no actions) |
| Cancelled | — (no actions) | — (no actions) |

**Key Design Decisions:**
- **Close and Reopen are submitter-exclusive actions** — only the employee who submitted the ticket can confirm resolution (Close) or reject it (Reopen). Managers/admins cannot close or reopen on behalf of the submitter.
- **Delete is admin-only on Resolved tickets** — managers/admins can only delete tickets that are in Resolved status. This prevents deletion of active tickets while allowing cleanup of resolved-but-unclosed tickets.
- **Submitters cannot delete** — deletion is not available to the submitting employee in any status.

**Exit Points:**
- **+ Add Request** → Navigate to Create Request Ticket page
- **Edit** → Navigate to Edit Request Ticket page
- **In Progress / On Hold / Resume / Resolve / Close / Reopen** → Stays on list; status updated inline; list refreshes
- **Cancel** → Confirmation dialog → status updated inline; list refreshes
- **Delete** → Confirmation dialog → row removed from list; list refreshes
- **Export** → File downloads; stays on list

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Search On | Trigger | Mandatory | Placeholder | Description |
|------------|------------|-----------|---------|-----------|-------------|-------------|
| Search | Text input (320px) | Employee name, email, phone number | On change with 300ms debounce | No | "Search by name, email, phone number..." | Multi-field search — case-insensitive, partial match |

### Filter Elements

| Filter Name | Type | Options Source | Multi-select | Default | Description |
|-------------|------|---------------|-------------|---------|-------------|
| Department | Dropdown chip | Departments from EP-008 US-001 | Yes | None (show all) | With in-dropdown search |
| Position | Dropdown chip | Positions from EP-008 US-002 | Yes | None (show all) | With in-dropdown search |
| Status | Dropdown chip | Open, In Progress, On Hold, Resolved, Closed, Cancelled | Yes | None (show all) | Shows count e.g., "(2)"; with in-dropdown search |

**Note:** No Request Type filter chip is included. The Figma design specifies 3 filter chips only (Department, Position, Status). This was confirmed as the intended design.

### Interaction Elements

| Element | Type | Visible To | State/Condition | Trigger Action | Description |
|---------|------|-----------|-----------------|----------------|-------------|
| Reset | Icon button | All with view permission | Visible when any filter/search active; hidden otherwise | Clears all filters and search | Refresh icon |
| Export | Button (secondary) | All with view permission | Always visible | Downloads currently filtered list | Exports result set |
| + Add Request | Button (primary) | All authorized users (including employees) | Always visible for authorized users | Navigate to Create Request Ticket | Creates new ticket |
| Gear icon (per row) | Icon button | Context-sensitive | Actions vary by status x role (see matrix above) | Opens action dropdown | Hidden for Closed and Cancelled rows |
| Rows per page | Dropdown | All with view permission | Default: 10 | Changes page size (10, 25, 50) | Pagination control |
| Page navigation | Button group | All with view permission | Always visible | Navigate between pages | Numbered page buttons |

---

## 4. Data Display

### Information Shown to User

| Data | Column | Width | Format | Empty State | Business Meaning |
|------|--------|-------|--------|-------------|-----------------|
| Full Name | Column 1 | ~201px | Text | — | Employee who submitted the ticket |
| Department & Position | Column 2 | ~201px | Two-line: dept (14px bold) + position (12px muted) | "—" / "—" | Organizational context of the submitter |
| Request Date | Column 3 | ~160px | Date (DD/MM/YYYY) | — | Date the ticket was submitted |
| Request Type | Column 4 | ~160px | Text | — | Category: IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other |
| Request | Column 5 | ~201px | Text (truncated with tooltip on hover) | — | Request subject/title — shows full text via tooltip when truncated |
| Status | Column 6 | ~201px | Colored badge | — | Current lifecycle status |
| Action | Column 7 | ~50px | Gear icon (context-sensitive) | — | Actions vary by status x role |

### Request Type Values (Confirmed)

| Request Type | Description |
|-------------|-------------|
| IT Support | Technology, hardware, software requests |
| Facility | Building, workspace, maintenance requests |
| HR Inquiry | Human resources related questions and requests |
| Office Supplies | Stationery, equipment, supply requests |
| Access Request | System access, permissions, credentials |
| Travel & Expense | Business travel and expense reimbursement |
| Training | Training programs, courses, certifications |
| Other | Any request not covered by above categories |

### Status Badge Colors (Confirmed)

| Status | Badge Color | Text |
|--------|------------|------|
| Open | Blue background | "Open" |
| In Progress | Amber background | "In Progress" |
| On Hold | Orange background | "On Hold" |
| Resolved | Green background | "Resolved" |
| Closed | Gray background | "Closed" |
| Cancelled | Red background | "Cancelled" |

### Request Column Display Rules

| Rule | Detail |
|------|--------|
| Content | Displays the request subject/title text |
| Truncation | Text truncated with ellipsis ("...") when exceeding column width |
| Tooltip | On hover, show full untruncated text in a tooltip |
| Max lines | Single line with text-overflow ellipsis |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens or data being fetched | Skeleton rows in table |
| Populated | Request tickets exist (within visibility scope) | Table with data rows, filters, pagination |
| Empty | No request tickets in the system (or none owned by employee) | "No request tickets found" with "+ Add Request" CTA |
| No Results (search) | Search returns zero matches | "No request tickets match your search" with "Clear search" link |
| No Results (filter) | Filters return zero matches | "No request tickets match the selected filters" with "Reset filters" link |
| No Results (combined) | Search + filters return zero | "No request tickets match your search and filters" with "Reset all" link |
| Filter active | One or more filters selected | Active chips highlighted; Reset visible; Status chip shows count |

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Request Tickets                                  │
│  200px       │                                                  │
│  Users Mgmt  │  [Search 320px] [Dept] [Position] [Status(2)]   │
│  > Users     │  [Reset]                  [Export] [+ Add Request]│
│  > Roles     │                                                  │
│  > Perms     │  ┌────────┬──────────┬──────┬──────┬─────┬─────┐│
│              │  │FullName│Dept&Pos  │Req   │Req   │Req  │Stat ││
│  HRM         │  │        │          │Date  │Type  │     │us   ││
│  > Leave     │  ├────────┼──────────┼──────┼──────┼─────┼─────┤│
│              │  │ Name   │Frontend  │ Date │IT Sup│Subj.│[Open]││
│  Operation   │  │        │Team Lead │      │port  │     │      ││
│  > Request ◄ │  │ Name   │Backend   │ Date │Facil-│Subj.│[In  ]││
│    Tickets   │  │        │Developer │      │ity   │     │Prog. ││
│              │  └────────┴──────────┴──────┴──────┴─────┴─────┘│
│              │  (cont: Action ⚙)                                │
│              │                                                  │
│              │  Rows per page [10▼]  Page 1 of 10  1 2 [3] 4 > │
└──────────────┴──────────────────────────────────────────────────┘
```

> **Note:** Figma design at node `3171:3062`. See ANALYSIS.md Design Context [ADD-ON] for full component inventory.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Request Tickets page displays with page title "Request Tickets" in Geist Semibold 24px
- **AC-02:** Table has 7 columns: Full Name, Department & Position, Request Date, Request Type, Request, Status, Action
- **AC-03:** Tickets are listed by most recent submission date first (newest at top)
- **AC-04:** Department & Position column displays two lines: department name (14px bold) on top, position (12px muted) below
- **AC-05:** Status column displays colored badges: Open (blue), In Progress (amber), On Hold (orange), Resolved (green), Closed (gray), Cancelled (red)
- **AC-06:** Request column displays subject/title text, truncated with ellipsis when exceeding column width
- **AC-07:** Hovering over a truncated Request cell shows the full text in a tooltip

**Data Visibility:**
- **AC-08:** Employees see only their own submitted tickets
- **AC-09:** Managers and administrators see all tickets across the organization
- **AC-10:** Data visibility scoping is enforced server-side (not just UI filtering)

**Search:**
- **AC-11:** Search field filters across employee name, email, and phone number simultaneously with ~300ms debounce
- **AC-12:** Search is case-insensitive and uses partial/contains matching
- **AC-13:** When search returns no results, "No request tickets match your search" is displayed with a clear search option
- **AC-14:** Clearing the search field restores the full list (respecting active filters and data visibility) and resets to page 1

**Filters:**
- **AC-15:** Department filter chip opens a multi-select dropdown with in-dropdown search
- **AC-16:** Position filter chip opens a multi-select dropdown with in-dropdown search
- **AC-17:** Status filter chip opens a multi-select dropdown listing: Open, In Progress, On Hold, Resolved, Closed, Cancelled
- **AC-18:** Only 3 filter chips are displayed (Department, Position, Status) — no Request Type filter chip
- **AC-19:** Active filter chips show the count of selected values (e.g., "Status (2)")
- **AC-20:** Filters are additive: OR within same filter, AND across filters, AND with search
- **AC-21:** When filters return no results, "No request tickets match the selected filters" is displayed with a "Reset filters" link
- **AC-22:** Reset button clears all filters and search; hidden when no filters/search are active
- **AC-23:** Each filter dropdown includes a search field to filter options (client-side, no server round-trip)

**Pagination:**
- **AC-24:** List paginates with default 10 rows per page; user can change to 25 or 50
- **AC-25:** Pagination controls are hidden when total results <= current rows per page
- **AC-26:** When search or filters are applied or cleared, pagination resets to page 1

**Export:**
- **AC-27:** Export button downloads the currently filtered/searched result set (within data visibility scope)
- **AC-28:** Export button shows a brief loading state while file is being prepared

**Context-Sensitive Gear Icon Actions:**
- **AC-29:** Gear icon dropdown actions vary based on ticket status AND user's role (submitter vs. manager/admin)
- **AC-30:** For Open tickets — submitter sees: Edit, Cancel
- **AC-31:** For Open tickets — manager/admin sees: Edit, In Progress, Cancel
- **AC-32:** For In Progress tickets — submitter sees: Cancel
- **AC-33:** For In Progress tickets — manager/admin sees: On Hold, Resolve, Cancel
- **AC-34:** For On Hold tickets — submitter sees: Cancel
- **AC-35:** For On Hold tickets — manager/admin sees: Resume, Cancel
- **AC-36:** For Resolved tickets — submitter sees: Close, Reopen
- **AC-37:** For Resolved tickets — manager/admin sees: Delete
- **AC-38:** For Closed tickets — no actions available (gear icon hidden)
- **AC-39:** For Cancelled tickets — no actions available (gear icon hidden)
- **AC-40:** Edit navigates to Edit Request Ticket page
- **AC-41:** In Progress, On Hold, Resume, Resolve actions update status inline — list refreshes with new status badge
- **AC-42:** Close and Reopen actions update status inline — list refreshes
- **AC-43:** Cancel triggers confirmation dialog → status updated inline → list refreshes
- **AC-44:** Delete triggers confirmation dialog → row removed from list (soft delete) → list refreshes

**Permissions & Visibility:**
- **AC-45:** "+ Add Request" button is visible to all authorized users including employees (not restricted to management permission)
- **AC-46:** Users with view-only permission can browse, search, filter, paginate, and export but cannot see "+ Add Request" or gear icon actions
- **AC-47:** Export button is visible to all users with view permission
- **AC-48:** Direct URL access by unauthorized users redirects to fallback page

**Empty & Loading States:**
- **AC-49:** When no request tickets exist (or employee has none), empty state shows "No request tickets found" with "+ Add Request" CTA
- **AC-50:** Skeleton rows are displayed while data is loading

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Default load (admin) | Navigate to Request Tickets as admin | All tickets shown, newest first, page 1 | High |
| Default load (employee) | Navigate to Request Tickets as employee | Only own tickets shown, newest first | High |
| Search by name | Type "Henry" | Tickets from employees matching "Henry" (within visibility scope) | High |
| Search no match | Type "zzz" | "No request tickets match your search" | High |
| Filter by status | Select "Open" | Only Open tickets shown; chip shows "(1)" | High |
| Filter multiple statuses | Select "Open" + "In Progress" | Open OR In Progress tickets | High |
| Combined search + filter | Search "john" + Status=Open | Open tickets matching "john" | High |
| Reset all | Filters + search active, click Reset | All cleared, full list restored (within visibility scope) | Medium |
| Export filtered | Filters active, click Export | Only filtered results downloaded | High |
| Gear — Open (submitter) | Click gear on own Open ticket | Shows Edit, Cancel | High |
| Gear — Open (admin) | Click gear on Open ticket | Shows Edit, In Progress, Cancel | High |
| Gear — In Progress (submitter) | Click gear on own In Progress ticket | Shows Cancel | High |
| Gear — In Progress (admin) | Click gear on In Progress ticket | Shows On Hold, Resolve, Cancel | High |
| Gear — On Hold (submitter) | Click gear on own On Hold ticket | Shows Cancel | High |
| Gear — On Hold (admin) | Click gear on On Hold ticket | Shows Resume, Cancel | High |
| Gear — Resolved (submitter) | Click gear on own Resolved ticket | Shows Close, Reopen | High |
| Gear — Resolved (admin) | Click gear on Resolved ticket | Shows Delete | High |
| Gear — Closed | Click gear on Closed ticket | No actions / gear hidden | Medium |
| Gear — Cancelled | Click gear on Cancelled ticket | No actions / gear hidden | Medium |
| Dept & Position column | Ticket from Frontend Team Leader | Two lines: "Frontend" (bold) + "Team Leader" (muted) | Medium |
| Request column truncation | Ticket with long subject text | Text truncated with "...", tooltip on hover shows full text | Medium |
| Empty state (employee) | Employee with no tickets | Empty state with "+ Add Request" CTA | Medium |
| Pagination reset | On page 3, apply filter | Returns to page 1 | Medium |
| Unauthorized | View-only user | No Add Request, no gear icons | High |
| Data visibility enforcement | Employee tries API with other user's ticket ID | Server rejects; only own data returned | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Access to the Request Tickets page is controlled by view permission configured in US-004 — users without view permission cannot access this page
- **SR-02:** Data visibility is enforced server-side: employees see only their own submitted tickets; managers/admins see all tickets organization-wide
- **SR-03:** The "+ Add Request" button is visible to all authorized users including employees — this is an exception to the typical pattern where only management permission sees the Add button
- **SR-04:** Gear icon actions are only displayed to roles with management permission, with the exception of submitter-exclusive actions (Close, Reopen, Cancel on own tickets)
- **SR-05:** Search applies across 3 fields simultaneously: employee name, email, phone number. A match in ANY field qualifies the row.
- **SR-06:** Search matching is case-insensitive with partial/contains matching
- **SR-07:** Filters are multi-select — OR logic within the same filter, AND logic across different filters. Search is ANDed with all filters.
- **SR-08:** When search or any filter is applied or cleared, pagination resets to page 1
- **SR-09:** The default display order is most recent submission date first (newest at top)
- **SR-10:** Export produces the currently filtered/searched result set, scoped by the user's data visibility
- **SR-11:** Each filter dropdown dynamically loads options: Department from EP-008 US-001, Position from EP-008 US-002, Status from system-defined values (Open, In Progress, On Hold, Resolved, Closed, Cancelled)
- **SR-12:** Filter dropdown options include a search field — filtered client-side (no server round-trip)
- **SR-13:** The Reset button is only visible when at least one filter or search is active
- **SR-14:** Gear icon actions are context-sensitive based on both ticket status AND user's role:

| Status | Submitter (own ticket) | Manager/Admin |
|--------|----------------------|---------------|
| Open | Edit, Cancel | Edit, In Progress, Cancel |
| In Progress | Cancel | On Hold, Resolve, Cancel |
| On Hold | Cancel | Resume, Cancel |
| Resolved | Close, Reopen | Delete |
| Closed | (none) | (none) |
| Cancelled | (none) | (none) |

- **SR-15:** For rows with no available actions (Closed, Cancelled), the gear icon is hidden
- **SR-16:** In Progress, On Hold, Resume, Resolve, Close, and Reopen actions update status inline on the list — no page navigation; list refreshes to reflect new badge
- **SR-17:** Cancel action requires a confirmation dialog before executing
- **SR-18:** Delete action requires a confirmation dialog before executing; performs soft delete (record flagged as deleted, removed from all active views)
- **SR-19:** Close and Reopen are submitter-exclusive — only the ticket submitter can perform these actions. Managers/admins cannot close or reopen tickets.
- **SR-20:** Delete is restricted to managers/admins and only available on Resolved tickets
- **SR-21:** Request ticket status lifecycle:
  - `Open → In Progress` (manager/admin)
  - `In Progress → On Hold` (manager/admin)
  - `On Hold → In Progress` (manager/admin — resume)
  - `In Progress → Resolved` (manager/admin)
  - `Resolved → Closed` (submitter — confirms resolution)
  - `Resolved → Open` (submitter — reopens, resolution unsatisfactory)
  - `Open → Cancelled` (submitter or manager/admin)
  - `In Progress → Cancelled` (submitter or manager/admin)
  - `On Hold → Cancelled` (submitter or manager/admin)
  - Closed: terminal state
  - Cancelled: terminal state
- **SR-22:** Only Open tickets can be edited by the submitting employee; managers/admins can edit Open or In Progress tickets
- **SR-23:** Only 3 filter chips are displayed (Department, Position, Status) — no Request Type filter chip per Figma design

**State Transitions:**
```
[List loads] → [Skeleton] → [Data visibility applied] → [Data fetched] → [Table displayed (newest first)]
[User types in search] → [Debounce 300ms] → [List filters across 3 fields] → [Page 1]
[User clears search] → [List restores (respecting filters + visibility)] → [Page 1]
[User selects filter value] → [List filters immediately] → [Page 1]
[User clicks Reset] → [All filters + search cleared] → [Full list (within visibility)] → [Page 1]
[User clicks Export] → [Export loading] → [File downloaded (filtered set, visibility-scoped)]
[User clicks Gear → In Progress] → [Status updated to In Progress] → [List refreshes; badge amber]
[User clicks Gear → On Hold] → [Status updated to On Hold] → [List refreshes; badge orange]
[User clicks Gear → Resume] → [Status updated to In Progress] → [List refreshes; badge amber]
[User clicks Gear → Resolve] → [Status updated to Resolved] → [List refreshes; badge green]
[User clicks Gear → Close] → [Status updated to Closed] → [List refreshes; badge gray]
[User clicks Gear → Reopen] → [Status updated to Open] → [List refreshes; badge blue]
[User clicks Gear → Cancel] → [Confirmation dialog] → [Confirm → status Cancelled; list refreshes; badge red]
[User clicks Gear → Delete] → [Confirmation dialog] → [Confirm → row removed (soft delete); list refreshes]
[User clicks Gear → Edit] → [Navigate to Edit Request Ticket page]
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
- **UX-05:** Each filter dropdown includes a search field at the top for quick option finding (client-side)
- **UX-06:** Reset button only appears when filters/search are active — keeps interface clean
- **UX-07:** Status badges are color-coded for 6 statuses — Open (blue) and In Progress (amber) draw attention to tickets needing action
- **UX-08:** Department & Position column uses two-line format (dept bold + position muted) — packs two data points into one column, consistent with Leave Requests pattern
- **UX-09:** Gear icon is hidden for Closed and Cancelled tickets — no empty dropdown shown
- **UX-10:** Status transition actions (In Progress, On Hold, Resume, Resolve, Close, Reopen) update inline without page navigation — enables rapid workflow processing
- **UX-11:** Cancel and Delete actions require confirmation dialogs — prevents accidental data loss
- **UX-12:** Export button shows brief loading/spinner state — prevents double-click
- **UX-13:** Pagination controls hidden when total results fit on one page
- **UX-14:** Default sort (newest first) ensures recently submitted tickets are immediately visible
- **UX-15:** Request column truncates long text with ellipsis and shows full text via tooltip on hover — maintains consistent column width while preserving access to full content
- **UX-16:** Data visibility scoping ensures employees only see relevant tickets — reduces noise and maintains privacy
- **UX-17:** "+ Add Request" visible to all authorized users including employees — encourages self-service ticket submission

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>=1280px) | Full 7-column layout: sidebar 200px + content area 1694px |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through search, filter chips, Reset, Export, Add Request, gear icons, pagination
- [x] Screen reader compatible — table column headers, button labels, filter state announced
- [x] Filter dropdowns accessible via keyboard — arrow keys to navigate, Enter to select, Escape to close
- [x] Status badges have accessible labels (not color-only — text included in badge)
- [x] Sufficient color contrast — badge colors meet WCAG 2.1 AA standards
- [x] Focus indicators visible on all interactive elements
- [x] Tooltip content accessible via keyboard focus (not hover-only)

**Design References:**
- Figma: [Request Tickets](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3171-3062) (node `3171:3062`)
- Design tokens: primary `#171717`, background `#ffffff`, border `#e5e5e5`, muted foreground `#737373`, font: Geist
- Pattern reference: Leave Requests List (DR-002-001-01) for filter + search + status workflow pattern

---

## 8. Additional Information

### Out of Scope
- Request ticket detail/drill-down view (future DR)
- Column sorting by user interaction (newest-first default only)
- Advanced search syntax or field-specific search operators
- Bulk actions (select multiple and change status at once)
- Request Type filter chip (not in Figma design — only 3 chips: Department, Position, Status)
- Request assignment/routing to specific handlers (future enhancement)
- SLA tracking and escalation rules (future enhancement)
- Comments/thread discussion on tickets (future enhancement)
- Automated notifications/email alerts on status changes (future enhancement)
- Request ticket templates per type (future enhancement)
- Mobile or tablet responsive layout

### Open Questions (Resolved)

| Question | Answer | Confirmed By |
|----------|--------|-------------|
| Status values | Open, In Progress, On Hold, Resolved, Closed, Cancelled (6 statuses) | Product Owner |
| Status badge colors | Blue, Amber, Orange, Green, Gray, Red (exact tokens pending Design Team) | Product Owner |
| Request Type values | IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other | Product Owner |
| Filter chips | 3 chips only: Department, Position, Status — no Request Type chip (follows Figma) | Product Owner / Figma |
| Request column content | Subject/title text, truncated with tooltip on hover | Product Owner |
| Data visibility | Employees see own only; managers/admins see all (server-side enforced) | Product Owner |
| Who can create tickets | All authorized users including employees (not restricted to management permission) | Product Owner |
| Close/Reopen ownership | Submitter-exclusive — only the ticket submitter can Close or Reopen | Product Owner |
| Delete scope | Admin/manager only, Resolved tickets only | Product Owner |
| Submitter Cancel scope | Submitters can cancel own tickets in Open, In Progress, On Hold | Product Owner |
| Sidebar navigation | Operation is menu group; Request Tickets is menu item under it | Product Owner |

### Open Questions (Pending)

- [ ] **Export format:** CSV or Excel (.xlsx)? Which columns included? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Status badge exact tokens:** Confirm exact color hex values for each of the 6 status badges. — **Owner:** Design Team — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-003-001-02: Create Request Ticket (planned) | Triggered from "+ Add Request" button |
| DR-003-001-03: Edit Request Ticket (planned) | Triggered from gear → Edit |
| DR-003-001-04: Manage Request Ticket Status (planned) | Triggered from gear → In Progress / On Hold / Resume / Resolve / Close / Reopen |
| DR-003-001-05: Cancel Request Ticket (planned) | Triggered from gear → Cancel |
| DR-003-001-06: Delete Request Ticket (planned) | Triggered from gear → Delete |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls view/management permissions; role-based actions |
| US-005: User Management | Employee data (Full Name, Department, Position) |
| EP-008 US-001/002: Department & Position | Filter dropdown options |

### Notes
- This is the **first module under the Operation sidebar section** — introduces the Operation menu group in EP-003.
- The **6-status lifecycle** is the most complex workflow in the HRM system so far. Leave Requests has 4 statuses; Request Tickets has 6 with branching paths.
- **Submitter-exclusive actions** (Close, Reopen) are a new pattern — previous features (Leave Requests) do not have actions restricted to the submitter only.
- **Data visibility scoping** (employees see own only) is shared with the Leave Requests pattern and must be enforced server-side.
- The **Request column tooltip on hover** for truncated text is a new UI behavior not present in other list views.
- **Delete restricted to Resolved status** is a new constraint — previous delete patterns (Leave Requests) allowed admin deletion regardless of status.
- The **"+ Add Request" button visible to all authorized users** (including employees) is an exception to the standard pattern where only management permission sees the Add button.

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
| 1.0 | 2026-04-01 | BA Agent | Initial draft — Figma design context from node 3171:3062; 6-status lifecycle; submitter-exclusive Close/Reopen; data visibility scoping |
