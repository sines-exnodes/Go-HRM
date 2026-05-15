---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
detail_id: DR-001-004-01
detail_name: "Role List"
parent_requirement: FR-US-004-01
status: draft
version: "1.0"
created_date: 2026-03-19
last_updated: 2026-03-19
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../TODO.yaml"
    relationship: sibling
---

# Detail Requirement: Role List

**Detail ID:** DR-001-004-01
**Parent Requirement:** FR-US-004-01 (List View category: FR-US-004-01 through FR-US-004-03, FR-US-004-09)
**Story:** US-004-role-permission-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with role management permission**, I want to **view all roles in a searchable, paginated list showing each role's name and assigned permissions**, so that **I can browse the organization's role catalog and manage roles as needed**.

**Purpose:** Provide a centralized, browsable view of all roles defined in the system. The Role List is the main entry point for role management — from here, authorized administrators can search for roles, review their permissions at a glance, and access Create, Edit, and Delete actions.

**Target Users:**
- Any role with **view permission** on role management — can browse, search, and paginate
- Any role with **management permission** — can additionally access "+ Add New", Edit, and Delete actions via the gear icon

**Key Functionality:**
- Searchable, paginated table with 3 columns: Role Name, Permission, Action
- Per-row gear icon (Edit / Delete) for authorized users
- No Export button (differs from Department/Position lists — pending PO confirmation)
- Permission-controlled visibility of management actions

---

## 2. User Workflow

**Entry Point:** Sidebar navigation → Users Management > Roles

**Preconditions:**
- User is signed in to Exnodes HRM
- User's role has role view permission (configured via US-004)

**Main Flow:**
1. User clicks "Roles" under Users Management in the sidebar
2. System loads the Role List page
3. System displays the role table with columns: Role Name | Permission | Action
4. Roles are listed alphabetically by name, paginated (default 10 rows per page)
5. User browses or takes one of the available actions (see Alternative Flows below)

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Search** | User types in the search box → list filters by role name with debounce (300ms) → pagination resets to page 1 |
| **Clear Search** | User clears the search box → full role list restored → pagination resets to page 1 |
| **Paginate** | User changes page number or rows per page → list updates to show selected page |
| **Add New** | User clicks "+ Add New" → navigates to Create Role page *(DR-001-004-02)* |
| **Gear → Edit** | User clicks gear icon on a row → selects Edit → navigates to Edit Role page *(DR-001-004-03)* |
| **Gear → Delete** | User clicks gear icon on a row → selects Delete → system checks user count → proceeds or blocks *(DR-001-004-04)* |

**Exit Points:**
- **Add New** → Navigates away to Create Role page
- **Edit (via gear)** → Navigates away to Edit Role page
- **Delete (via gear)** → Stays on list page; either deletion completes and list refreshes, or system shows blocking message

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Search On | Trigger | Mandatory | Placeholder | Description |
|------------|------------|-----------|---------|-----------|-------------|-------------|
| Search | Text input | Role name | On change with 300ms debounce | No | "Search by role name…" | Filters table by role name; case-insensitive; partial match supported |

### Interaction Elements

| Element | Type | Visible To | State / Condition | Trigger Action | Description |
|---------|------|-----------|-------------------|----------------|-------------|
| + Add New button | Primary button | Management permission only | Hidden for view-only roles | Navigates to Create Role page | Opens DR-001-004-02 flow |
| Gear icon (per row) | Icon button | Management permission only | Hidden for view-only roles | Opens action menu | Shows Edit and Delete options for the row |
| Rows per page | Dropdown | All with view permission | Always visible | Updates rows shown per page | Options: 10, 25, 50 (default: 10) |
| Page navigation | Pagination control | All with view permission | Hidden when total ≤ rows per page | Moves between pages | Shows current page and total pages |

---

## 4. Data Display

### Information Shown to User

| Data | Column | Format | Empty State | Business Meaning |
|------|--------|--------|-------------|-----------------|
| Role Name | Column 1 (181px) | Text, alphabetical order | — | The name of the role |
| Permission | Column 2 (1380px) | Comma-separated text, single-line with ellipsis overflow; full list on hover (tooltip) | — | Permissions assigned to this role |
| Action | Column 3 (93px) | Gear icon (management permission only) | — | Opens Edit / Delete action menu for this row |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens or data is being fetched | Skeleton rows displayed in table while data loads |
| Populated | Roles exist | Table with data rows, pagination controls below |
| Empty | No roles in the system | "No roles found" message; "Get started by creating your first role" sub-text; "+ Add New" button (management permission only) |
| No Results | Search returns zero matches | "No roles match your search" message; "Clear search" link to reset |
| Delete Blocked | User attempts to delete a role with assigned users | Dialog: "Cannot delete — [X] users are assigned to this role. Reassign all users before deleting." |

### Permission Column Detail

- Permissions displayed as comma-separated text (e.g., "View Users, Manage Roles, Edit Departments, View Reports")
- Single-line display with CSS text-overflow ellipsis when content exceeds column width
- Full permission list visible on hover (tooltip)
- Full permission detail accessible when user navigates to Edit Role

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────────────────┐
│  Users Management > Roles                           [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Role List                                       │
│  200px       │                                                  │
│              │  [Search 320px]                    [+ Add New]   │
│              │                                                  │
│              │  ┌──────────┬─────────────────────┬──────────┐  │
│              │  │Role Name │ Permission          │ Action   │  │
│              │  ├──────────┼─────────────────────┼──────────┤  │
│              │  │ Admin    │ View Users, Manage…  │ ⚙        │  │
│              │  │ Manager  │ View Users, View R…  │ ⚙        │  │
│              │  │ Employee │ View Profile          │ ⚙        │  │
│              │  └──────────┴─────────────────────┴──────────┘  │
│              │                                                  │
│              │  Rows per page [10▼]  Page 1 of N   1 2 [3] 4 > │
└──────────────┴──────────────────────────────────────────────────┘
```

> **Figma Reference:** [Role List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3081-1680) (node `3081:1680`)

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** Role List page displays a table with 3 columns — Role Name, Permission, and Action
- **AC-02:** Roles are listed alphabetically by role name (A → Z)
- **AC-03:** Permission column displays assigned permissions as comma-separated text, truncated with ellipsis when exceeding one row
- **AC-04:** Full permission list is visible on hover (tooltip) and on keyboard focus
- **AC-05:** Page breadcrumb displays "Users Management / Roles"
- **AC-06:** Search field filters the list by role name as the user types, with ~300ms debounce
- **AC-07:** Search is case-insensitive and matches partial names ("adm" matches "Admin")
- **AC-08:** When search returns no results, "No roles match your search" is displayed with a clear search option
- **AC-09:** Clearing the search field restores the full role list and resets to page 1
- **AC-10:** List paginates with default 10 rows per page; user can change to 25 or 50
- **AC-11:** Pagination controls are hidden when total roles ≤ current rows per page setting
- **AC-12:** When search is applied or cleared, pagination resets to page 1
- **AC-13:** "+ Add New" button and gear icon are visible only to users with role management permission
- **AC-14:** Users with view-only permission can browse, search, and paginate but cannot see "+ Add New" or gear icon
- **AC-15:** When no roles exist, empty state shows "No roles found" with "Get started by creating your first role" and an "+ Add New" CTA (management permission only)
- **AC-16:** Skeleton rows are displayed while data is loading
- **AC-17:** If a user attempts to delete a role with assigned users, the system blocks and shows: "Cannot delete — [X] users are assigned to this role. Reassign all users before deleting."
- **AC-18:** After successful deletion, any active search filter remains in effect

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| List loads with roles | Navigate to Roles page | Table displays all roles alphabetically, paginated | High |
| Search with match | Type "adm" | Only roles containing "adm" displayed | High |
| Search with no match | Type "zzz" | No results state with clear search option | High |
| Search case-insensitive | Type "ADMIN" | Matches "Admin", "admin" | Medium |
| Permission ellipsis | Role has 10+ permissions | Single line, truncated with ellipsis | High |
| Permission tooltip | Hover over truncated permission text | Full permission list shown in tooltip | Medium |
| Add New hidden (view-only) | User with view-only role | "+ Add New" button not visible | High |
| Gear icon hidden (view-only) | User with view-only role | No gear icon in Action column | High |
| Empty state | No roles in system | Empty state with message and Add New CTA | Medium |
| Delete blocked | Delete role with 5 users | Blocking message shows "5 users assigned" | High |
| Pagination reset on search | User on page 3, applies search | Returns to page 1 with filtered results | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Access to the Role List page is controlled by view permission configured in US-004 — users without view permission cannot access this page
- **SR-02:** The "+ Add New" button, gear icon (Edit / Delete actions) are only displayed to roles granted role management permission via US-004
- **SR-03:** Deletion is blocked server-side if the role has ≥1 user assigned — the system checks user count before allowing the delete action to proceed
- **SR-04:** User count per role is a live count calculated from current user-role assignments, not a cached value
- **SR-05:** Search applies to the Role Name column only — the Permission column is not searched
- **SR-06:** Search matching is case-insensitive (e.g., "adm" matches "Admin", "ADMIN", "admin")
- **SR-07:** The default display order for roles is alphabetical by role name (A → Z)
- **SR-08:** When a search is applied or cleared, pagination resets to page 1 to avoid showing an empty page
- **SR-09:** Permission data displayed in the Permission column is read from the role-permission assignments — the same source of truth used for access enforcement across all modules
- **SR-10:** Users without role view permission who attempt to access the Role List URL directly are redirected to the Dashboard (or shown a 403 Forbidden page)

**State Transitions:**
```
[List loads] → [Skeleton shown] → [Data fetched] → [Table displayed]
[User types in search] → [Debounce 300ms] → [List filters] → [Pagination resets to page 1]
[User clears search] → [Debounce 300ms] → [Full list restored] → [Pagination resets to page 1]
[User clicks Delete (0 users)] → [Confirmation] → [Role deleted] → [List refreshes]
[User clicks Delete (>0 users)] → [System blocks] → [Blocked message shown] → [No changes]
```

**Dependencies:**
- **US-004 (self-referential):** Role management permissions control access to this very page and its management actions
- **US-001 (Authentication):** User must be signed in to access the Role List
- **All other HRM modules:** Roles and permissions defined here govern access across the entire platform

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Search input uses a debounce delay (~300ms) so the list does not filter on every single keystroke — reduces unnecessary load and improves perceived performance
- **UX-02:** Skeleton rows are shown during page load to prevent layout shift and indicate that content is loading
- **UX-03:** The gear icon action menu closes automatically when the user clicks anywhere outside it
- **UX-04:** Permission text truncated with ellipsis shows full list on hover (tooltip) — no click required for quick review
- **UX-05:** Pagination controls are hidden when total roles fit on one page — no unnecessary UI clutter
- **UX-06:** When the Confirmation Dialog opens for delete, default focus is placed on the Cancel button (not Delete) — prevents accidental deletion if user presses Enter immediately
- **UX-07:** Delete button in the Confirmation Dialog uses a danger/destructive style (red background) to visually signal the irreversible nature of the action

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Full layout: sidebar 200px + content area 1694px |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [ ] Keyboard navigable (Tab through search, Add New, gear icons, pagination)
- [ ] Screen reader compatible (table column headers, button labels)
- [ ] Permission tooltip accessible via keyboard focus (not hover-only)
- [ ] Sufficient color contrast (design tokens: foreground `#0a0a0a` on background `#ffffff`)
- [ ] Focus indicators visible on all interactive elements

**Design Reference:**
- Figma: [Role List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3081-1680) (node `3081:1680`)
- Design tokens: primary `#171717`, background `#ffffff`, border `#e5e5e5`, muted `#f5f5f5`, font: Geist, border-radius: 6px

---

## 8. Additional Information

### Out of Scope

- Role detail/drill-down view (clicking a role name does not navigate to a detail page)
- Column sorting by user interaction (alphabetical default only; no user-controlled sort toggle)
- Advanced filtering beyond name search (e.g., filter by permission, filter by user count)
- Export functionality (not present in design — pending PO confirmation)
- Permission column click-to-expand (full list shown via hover tooltip only)
- Mobile or tablet responsive layout (desktop-first application)

### Open Questions

- [ ] **Export button:** Is omitting Export from Role List intentional, or should it be added for consistency with Department/Position lists? — **Owner:** Product Owner — **Status:** Pending
- [ ] **User count column:** Should a "No. of Users" column be added (aids delete-blocking awareness)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Gear icon actions:** Confirmed as Edit + Delete only? Any other options (e.g., Duplicate Role)? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-004-02: Create Role | Triggered from "+ Add New" button on this list |
| DR-001-004-03: Edit Role | Triggered from gear icon → Edit on a list row |
| DR-001-004-04: Delete Role | Triggered from gear icon → Delete on a list row |
| US-001: Authentication | Upstream — user must be signed in |
| US-004: Role & Permission Management | Self — this DR is part of US-004 |
| All HRM modules | Downstream — roles defined here govern access platform-wide |

### Notes

- This is a **self-referential** module — US-004 controls its own access permissions.
- Unlike Department/Position lists, Role List has **no Export button** and shows a **Permission column** instead of an employee/user count column.
- The Permission column display format (comma-separated with ellipsis + hover tooltip) is a working decision — may be revised when Design Team delivers the final spec.

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
| 1.0 | 2026-03-19 | BA Agent | Initial draft |
