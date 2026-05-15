---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-01
detail_name: "User List"
parent_requirement: FR-US-005-01
status: draft
version: "1.0"
created_date: "2026-03-24"
last_updated: "2026-03-24"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../TODO.yaml"
    relationship: sibling
input_sources:
  - type: figma
    description: "User List screen"
    node_id: "3048:6011"
    extraction_date: "2026-03-24"
---

# Detail Requirement: User List

**Detail ID:** DR-001-005-01
**Parent Requirement:** FR-US-005-01 (List View category: FR-US-005-01 through FR-US-005-09)
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **view all users in a searchable, filterable, paginated list showing their profile, role, and status**, so that **I can browse, find, and manage all user accounts in the system**.

**Purpose:** Provide a centralized view of all user accounts in the HRM platform. The User List is the main entry point for user management — administrators can search across name/email/phone, filter by department/position/role/status, export the list, and access Create, Edit, and Deactivate/Activate actions.

**Target Users:**
- Any role with **view permission** on user management — can browse, search, filter, paginate, and export
- Any role with **management permission** — can additionally access "+ Add New", Edit, and Deactivate/Activate via gear icon

**Key Functionality:**
- 9-column table: First Name, Last Name, Department, Position, System Role, Email, Phone Number, Status, Action
- Multi-field search across first name, last name, email, and phone number
- 4 filter chips (Department, Position, System Role, Status) — each with multi-select dropdown and in-dropdown search
- Reset button to clear all filters and search
- Export of the current filtered/searched result set
- Per-row gear icon with context-sensitive actions (Edit + Deactivate for active users, Edit + Activate for inactive users)
- Permission-controlled visibility of management actions

---

## 2. User Workflow

**Entry Point:** Sidebar navigation → Users Management > Users

**Preconditions:**
- User is signed in (US-001 Authentication)
- User's role has user view permission (configured via US-004)

**Main Flow:**
1. User clicks "Users" under Users Management in the sidebar
2. System loads the User List page
3. System displays the user table with 9 columns, all users paginated (default 10 rows per page)
4. Users are listed alphabetically by first name by default
5. User browses or takes one of the available actions

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Search** | User types in search box → list filters by first name, last name, email, or phone number with debounce (300ms) → pagination resets to page 1 |
| **Clear Search** | User clears search box → full list restored (respecting active filters) → pagination resets to page 1 |
| **Filter: Department** | User clicks Department chip → dropdown opens with search field + department list → user selects one or more → list filters → pagination resets to page 1 |
| **Filter: Position** | Same as Department filter but for positions |
| **Filter: System Role** | Same pattern for system roles |
| **Filter: Status** | Same pattern for status values (Active, Inactive) — chip shows count of selected values e.g., "(2)" |
| **Filter dropdown search** | User types in filter dropdown search field → dropdown options filter instantly (client-side) |
| **Combine filters** | Filters are additive — search + multiple filters can be active simultaneously; OR within same filter, AND across different filters |
| **Reset** | User clicks Reset → all filters and search cleared → full list restored → pagination resets to page 1 |
| **Paginate** | User changes page number or rows per page → list updates to show selected page |
| **Export** | User clicks Export → system downloads the currently filtered/searched list |
| **Add New** | User clicks "+ Add New" → navigates to Create User page *(DR-001-005-02)* |
| **Gear → Edit** | User clicks gear icon → selects Edit → navigates to Edit User page *(DR-001-005-03)* |
| **Gear → Deactivate** | User clicks gear icon on active user → selects Deactivate → confirmation dialog → status updated inline *(DR-001-005-08)* |
| **Gear → Activate** | User clicks gear icon on inactive user → selects Activate → confirmation dialog → status updated inline *(DR-001-005-08)* |

**Exit Points:**
- **Add New** → Navigates to Create User page
- **Edit** → Navigates to Edit User page
- **Deactivate/Activate** → Stays on list; status badge updates inline after confirmation
- **Export** → File downloads; stays on list

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Search On | Trigger | Mandatory | Placeholder | Description |
|------------|------------|-----------|---------|-----------|-------------|-------------|
| Search | Text input | First name, last name, email, phone number | On change with 300ms debounce | No | "Search by name, email, phone number..." | Multi-field search — case-insensitive, partial match. A match in ANY of the 4 fields qualifies the row. |

### Filter Elements

| Filter Name | Type | Options Source | Multi-select | Has Dropdown Search | Default | Description |
|-------------|------|---------------|-------------|---------------------|---------|-------------|
| Department | Dropdown chip | Departments from EP-008 US-001 | Yes | Yes | None selected (show all) | Filters by assigned department |
| Position | Dropdown chip | Positions from EP-008 US-002 | Yes | Yes | None selected (show all) | Filters by assigned position |
| System Role | Dropdown chip | Roles from US-004 | Yes | Yes | None selected (show all) | Filters by assigned system role |
| Status | Dropdown chip | Active, Inactive (values pending PO) | Yes | Yes | None selected (show all) | Filters by account status; chip shows count of selected values |

### Interaction Elements

| Element | Type | Visible To | State/Condition | Trigger Action | Description |
|---------|------|-----------|-----------------|----------------|-------------|
| Reset | Button (icon) | All with view permission | Visible only when any filter or search is active; hidden otherwise | Clears all filters and search | Refresh icon button |
| Export | Button (secondary) | All with view permission | Always visible | Downloads currently filtered/searched list | Exports current result set |
| + Add New | Button (primary) | Management permission only | Always visible for authorized users | Navigate to Create User page | Opens DR-001-005-02 flow |
| Gear icon (per row) | Icon button | Management permission only | Hidden for view-only roles | Opens context-sensitive dropdown | Edit + Deactivate (active user) or Edit + Activate (inactive user) |
| Rows per page | Dropdown | All with view permission | Default: 10 | Changes page size (10, 25, 50) | Pagination control |
| Page navigation | Button group | All with view permission | Hidden when total ≤ rows per page | Navigate between pages | Numbered page buttons with prev/next arrows |

---

## 4. Data Display

### Information Shown to User

| Data | Column | Format | Empty State | Business Meaning |
|------|--------|--------|-------------|-----------------|
| First Name | Column 1 (~184px) | Text, alphabetical default sort | — | User's first name |
| Last Name | Column 2 (~184px) | Text | — | User's last name |
| Department | Column 3 (~184px) | Text | "—" (unassigned) | Department the user belongs to |
| Position | Column 4 (~184px) | Text | "—" (unassigned) | Position the user holds |
| System Role | Column 5 (~184px) | Text | — (always required) | Role assigned to the user |
| Email | Column 6 (~184px) | Text | — (always required) | User's email address |
| Phone Number | Column 7 (~184px) | Text | "—" (if not provided) | User's phone number |
| Status | Column 8 (~184px) | Colored badge | — (always has status) | Active (green badge) or Inactive (gray badge) |
| Action | Column 9 (~184px) | Gear icon (management only) | — | Context-sensitive actions per row |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens or data being fetched | Skeleton rows in table while data loads |
| Populated | Users exist | Table with data rows, filter chips, pagination controls |
| Empty | No users in the system | "No users found" message with "+ Add New" CTA (management permission only) |
| No Results (search) | Search returns zero matches | "No users match your search" with "Clear search" link |
| No Results (filter) | Filters return zero matches | "No users match the selected filters" with "Reset filters" link |
| No Results (combined) | Search + filters return zero | "No users match your search and filters" with "Reset all" link |
| Filter active | One or more filters selected | Active filter chips highlighted; Reset button visible; chips show count |

### Status Badge Colors (Pending PO Confirmation)

| Status | Badge Color | Text |
|--------|------------|------|
| Active | Green background | "Active" |
| Inactive | Gray background | "Inactive" |

### Page Layout (from Figma)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Users Management > Users                                        [Top Bar] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  User List                                                  │
│  200px       │                                                             │
│              │  [Search 320px] [Department▾] [Position▾] [System Role▾]    │
│              │                 [Status(2)▾]  [Reset ↻]   [Export] [+AddNew]│
│              │                                                             │
│              │  ┌──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──┐│
│              │  │First │Last  │Dept  │Posit │System│Email │Phone │Status│⚙ ││
│              │  │Name  │Name  │      │ion   │Role  │      │Number│      │  ││
│              │  ├──────┼──────┼──────┼──────┼──────┼──────┼──────┼──────┼──┤│
│              │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[Act] │⚙ ││
│              │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[Ina] │⚙ ││
│              │  │...   │...   │...   │...   │...   │...   │...   │...   │  ││
│              │  └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┴──┘│
│              │                                                             │
│              │  Rows per page [10▼]          Page 1 of 10  1…2 [3] 4…5 >  │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** User List page displays with page title "User List" in Geist Semibold 24px
- **AC-02:** Table has 9 columns: First Name, Last Name, Department, Position, System Role, Email, Phone Number, Status, Action
- **AC-03:** Users are listed alphabetically by first name by default
- **AC-04:** Status column displays colored badges (Active = green, Inactive = gray)

**Search:**
- **AC-05:** Search field filters the list across first name, last name, email, and phone number simultaneously with ~300ms debounce
- **AC-06:** Search is case-insensitive and uses partial/contains matching ("rog" matches "Rogers" in last name, "rogers@" in email)
- **AC-07:** When search returns no results, "No users match your search" is displayed with a clear search option
- **AC-08:** Clearing the search field restores the full list (respecting any active filters) and resets to page 1

**Filters:**
- **AC-09:** Department filter chip opens a dropdown listing all departments; user can select multiple values
- **AC-10:** Position filter chip opens a dropdown listing all positions; user can select multiple values
- **AC-11:** System Role filter chip opens a dropdown listing all roles; user can select multiple values
- **AC-12:** Status filter chip opens a dropdown listing status values (Active, Inactive); user can select multiple values
- **AC-13:** Active filter chips show the count of selected values (e.g., "Status (2)")
- **AC-14:** Filters are additive — search + multiple filters can be active simultaneously; results match ALL active criteria (OR within same filter, AND across different filters)
- **AC-15:** When filters return no results, "No users match the selected filters" is displayed with a "Reset filters" link
- **AC-16:** Reset button clears all filters and search, restoring the full list; Reset button is hidden when no filters or search are active

**Filter Dropdown Search:**
- **AC-31:** Each filter dropdown (Department, Position, System Role, Status) includes a search field at the top that filters the dropdown options as the user types
- **AC-32:** Filter dropdown search is case-insensitive with partial match
- **AC-33:** When filter dropdown search returns no options, "No results found" is displayed within the dropdown

**Pagination:**
- **AC-17:** List paginates with default 10 rows per page; user can change to 25 or 50
- **AC-18:** Pagination controls are hidden when total results ≤ current rows per page setting
- **AC-19:** When search or filters are applied or cleared, pagination resets to page 1

**Export:**
- **AC-20:** Export button downloads the currently filtered/searched result set (not the full list if filters are active)
- **AC-21:** Export button shows a brief loading state while the file is being prepared

**Permissions & Visibility:**
- **AC-22:** "+ Add New" button and gear icon are visible only to users with user management permission
- **AC-23:** Users with view-only permission can browse, search, filter, paginate, and export but cannot see "+ Add New" or gear icon
- **AC-24:** Export button is visible to all users with view permission

**Empty & Loading States:**
- **AC-25:** When no users exist, empty state shows "No users found" with "+ Add New" CTA (management permission only)
- **AC-26:** Skeleton rows are displayed while data is loading

**Gear Icon Actions:**
- **AC-27:** Gear icon dropdown shows "Edit" and "Deactivate" for active users
- **AC-28:** Gear icon dropdown shows "Edit" and "Activate" for inactive users
- **AC-29:** Edit navigates to the Edit User page
- **AC-30:** Deactivate/Activate triggers the status change flow (confirmation dialog)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Default load | Navigate to User List | All users displayed alphabetically by first name, page 1 | High |
| Search by name | Type "Rogers" | Users with "Rogers" in first or last name | High |
| Search by email | Type "admin@" | Users with "admin@" in email | High |
| Search by phone | Type "555" | Users with "555" in phone number | Medium |
| Search no match | Type "zzz" | "No users match your search" | High |
| Filter by department | Select "Engineering" | Only Engineering users shown | High |
| Filter by multiple departments | Select "Engineering" + "HR" | Users in Engineering OR HR shown | Medium |
| Filter by status | Select "Active" | Only active users; chip shows "(1)" | High |
| Combined search + filter | Search "john" + filter Status=Active | Active users matching "john" | High |
| Filter dropdown search | Type "Eng" in Department dropdown | Only departments containing "Eng" shown in dropdown | Medium |
| Filter dropdown no match | Type "zzz" in Department dropdown | "No results found" in dropdown | Medium |
| Reset all | Filters + search active, click Reset | All cleared, full list restored | Medium |
| Export filtered | Search active, click Export | Only filtered results downloaded | High |
| Pagination reset | On page 3, apply filter | Returns to page 1 | Medium |
| Empty state | No users in system | Empty state with Add New CTA | Medium |
| Gear — active user | Click gear on active user | Shows Edit + Deactivate | High |
| Gear — inactive user | Click gear on inactive user | Shows Edit + Activate | High |
| Unauthorized | View-only user | No Add New, no gear icons, can browse/search/filter/export | High |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Access to the User List page is controlled by view permission configured in US-004 — users without view permission cannot access this page
- **SR-02:** The "+ Add New" button, gear icon (Edit / Deactivate / Activate actions) are only displayed to roles granted user management permission via US-004
- **SR-03:** Search applies across 4 fields simultaneously: first name, last name, email, phone number. A match in ANY of these fields qualifies the row.
- **SR-04:** Search matching is case-insensitive with partial/contains matching (e.g., "rog" matches "Rogers", "rogers@company.com")
- **SR-05:** Filters are multi-select — selecting multiple values within one filter uses OR logic (e.g., Department = "Engineering" OR "HR"). Across different filters, AND logic applies (e.g., Department = "Engineering" AND Status = "Active").
- **SR-06:** Search and filters are additive — results must match the search term AND all active filter criteria
- **SR-07:** When search or any filter is applied or cleared, pagination resets to page 1
- **SR-08:** The default display order is alphabetical by first name (A → Z)
- **SR-09:** Export produces the currently filtered/searched result set — if filters or search are active, only matching users are included
- **SR-10:** Each filter dropdown dynamically loads its options from the relevant data source: Department from EP-008 US-001, Position from EP-008 US-002, System Role from US-004, Status from system-defined values
- **SR-11:** Filter dropdown options include a search field — options are filtered client-side as the user types (no server round-trip for dropdown filtering)
- **SR-12:** Status badge values: "Active" (green) for active accounts, "Inactive" (gray) for deactivated accounts. Additional status values pending PO confirmation.
- **SR-13:** Gear icon dropdown shows context-sensitive actions based on user status: "Edit" + "Deactivate" for active users, "Edit" + "Activate" for inactive users
- **SR-14:** The Reset button is only visible when at least one filter or search is active — hidden when all filters are cleared and search is empty

**State Transitions:**
```
[List loads] → [Skeleton] → [Data fetched] → [Table displayed]
[User types in search] → [Debounce 300ms] → [List filters across 4 fields] → [Page 1]
[User clears search] → [List restores (respecting active filters)] → [Page 1]
[User selects filter value] → [List filters immediately] → [Page 1]
[User selects multiple filter values] → [OR within filter, AND across filters] → [Page 1]
[User clicks Reset] → [All filters + search cleared] → [Full list] → [Page 1]
[User clicks Export] → [Export loading] → [File downloaded (filtered set)]
[User clicks Gear → Edit] → [Navigate to Edit User page]
[User clicks Gear → Deactivate] → [Confirmation dialog] → [Status updated inline]
[User clicks Gear → Activate] → [Confirmation dialog] → [Status updated inline]
```

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls view and management permissions; provides System Role filter options
- **EP-008 US-001 (Department Management):** Provides Department filter options and user-department assignments
- **EP-008 US-002 (Position Management):** Provides Position filter options and user-position assignments

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Search input uses a debounce delay (~300ms) to reduce unnecessary filtering on every keystroke
- **UX-02:** Skeleton rows are shown during page load to prevent layout shift and indicate content is loading
- **UX-03:** Filter chips visually indicate active state (highlighted/filled) when values are selected — clear distinction from inactive chips
- **UX-04:** Active filter chips show the count of selected values (e.g., "Status (2)") so the user knows how many values are filtering
- **UX-05:** Each filter dropdown includes a search field at the top — critical for long lists (e.g., 50+ departments). Search filters options client-side instantly.
- **UX-06:** Reset button only appears when filters or search are active — no clutter when nothing is applied
- **UX-07:** The gear icon action menu closes automatically when the user clicks anywhere outside it
- **UX-08:** Gear icon shows context-sensitive actions (Deactivate for active users, Activate for inactive users) — no irrelevant options displayed
- **UX-09:** Export button shows a brief loading/spinner state while the file is being prepared — prevents double-click
- **UX-10:** Pagination controls are hidden when total results fit on one page — no unnecessary UI clutter
- **UX-11:** When filters narrow results to zero, the empty state message is specific to the active context ("No users match your search", "No users match the selected filters", or "No users match your search and filters") with a clear action to reset

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Full 9-column layout: sidebar 200px + content area |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through search, filter chips, Reset, Export, Add New, gear icons, pagination
- [x] Screen reader compatible — table column headers, button labels, filter state announced
- [x] Filter dropdowns accessible via keyboard — arrow keys to navigate options, Enter to select, Escape to close
- [x] Status badges have accessible labels (not color-only — text "Active"/"Inactive" included in badge)
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible on all interactive elements

**Design References:**
- Figma: [User List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3048-6011) (node `3048:6011`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Role List (DR-001-004-01) for basic list structure; User List extends with filters and multi-field search

---

## 8. Additional Information

### Out of Scope
- User detail/profile drill-down view (clicking a user row does not navigate to a detail page)
- Column sorting by user interaction (alphabetical by first name default only; no user-controlled sort toggle)
- Advanced search syntax (e.g., "email:john@" or field-specific search operators)
- Bulk actions (e.g., select multiple users and deactivate/export)
- User self-registration (admin-only creation)
- Password reset from the User List (handled in US-001 Authentication)
- Mobile or tablet responsive layout (desktop-first application)
- Inline editing of user fields on the list page

### Open Questions
- [ ] **Status values:** Confirmed as Active and Inactive? Or additional values like Suspended, Pending? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Status badge colors:** Green for Active, Gray for Inactive assumed — confirm exact colors/tokens. — **Owner:** Design Team — **Status:** Pending
- [ ] **Export format:** CSV or Excel (.xlsx)? Which columns included? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Gear icon actions:** Confirmed as Edit + Deactivate/Activate? Any others (e.g., Reset Password, View Profile)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Filter chip behavior:** Dropdown with multi-select + search confirmed — pending visual design for the dropdown component. — **Owner:** Design Team — **Status:** Pending
- [ ] **Default sort:** Alphabetical by first name assumed — or by creation date (newest first)? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-02: Create User (planned) | Triggered from "+ Add New" button on this list |
| DR-001-005-03: Edit User (planned) | Triggered from gear icon → Edit on a list row |
| DR-001-005-08: Activate/Deactivate User | Triggered from gear icon → Deactivate or Activate |
| US-001: Authentication | Upstream — user must be signed in; manages login/password flows |
| US-004: Role & Permission Management | Upstream — controls access; provides System Role filter data |
| EP-008 US-001: Department Management | Provides Department filter data and user-department assignments |
| EP-008 US-002: Position Management | Provides Position filter data and user-position assignments |

### Notes
- This is the **most complex list screen** in the HRM platform — 9 columns, 4 multi-select filter chips with in-dropdown search, and multi-field text search across 4 data fields. Other list screens (Department, Position, Skill, Role) have at most 4 columns and a single-field search with no filters.
- The **filter logic** follows a clear pattern: OR within the same filter (Department = "Engineering" OR "HR"), AND across different filters (Department = "Engineering" AND Status = "Active"). Search is ANDed with all filters.
- The **gear icon is context-sensitive** — showing "Deactivate" for active users and "Activate" for inactive users. This is unique to User List; other modules show the same options for all rows.
- The **Status column** uses colored badges — this is the first list screen with a status indicator. The exact values and colors need PO/Design confirmation.
- Users cannot be **deleted** — only deactivated. This preserves historical data and differs from Department/Position/Skill which use hard delete.
- The **filter dropdown search** (AC-31 through AC-33) is essential for usability — the Department list alone may have 50+ entries, making a scrollable-only dropdown impractical.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-24 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-24 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context |
