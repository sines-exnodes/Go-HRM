---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-001
story_name: "Department Management"
detail_id: DR-008-001-01
detail_name: "Department List"
parent_requirement: FR-US-001-01
status: draft
version: "1.0"
created_date: 2026-03-03
last_updated: 2026-03-03
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
---

# Detail Requirement: Department List

**Detail ID:** DR-008-001-01
**Parent Requirement:** FR-US-001-01 (List View category: FR-US-001-01 through FR-US-001-07)
**Story:** US-001-department-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

> ⚠️ **Requirements Change Note:** During detail requirement gathering, the Product Owner confirmed that departments have **no active/inactive status**. The original Deactivate (soft-delete) approach has been replaced with **hard Delete**, blocked when employees are assigned. REQUIREMENTS.md must be updated to reflect this change (US-001-04: Deactivate → Delete).

---

## 1. Use Case Description

As a **user with department view permission**, I want to **see all departments in a searchable, paginated list**, so that **I can browse the organization's department structure and authorized administrators can manage it as reference data for other HR modules**.

**Purpose:** Provide a centralized, browsable view of all departments so the data stays accurate and available as a selection source across other HR features — employee records, payroll, and any other module that requires a department reference.

**Target Users:**
- Any role granted **view permission** — can browse, search, paginate, and export the list
- Any role granted **management permission** — can additionally access Add New, Edit, and Delete actions

**Key Functionality:**
- Searchable, paginated table of all departments
- Per-row action menu (Edit / Delete) for authorized users
- Export of the current filtered list
- Permission-controlled visibility of management actions

---

## 2. User Workflow

**Entry Point:** Sidebar navigation → Organization Data > Departments

**Preconditions:**
- User is signed in to Exnodes HRM
- User's role has department view permission (configured via US-004)

**Main Flow:**
1. User clicks "Departments" under Organization Data in the sidebar
2. System loads the Department List page
3. System displays the department table with columns: Department Name | No. of Employees | Action
4. Departments are listed alphabetically by name, paginated (default 10 rows per page)
5. User browses or takes one of the available actions (see Alternative Flows below)

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Search** | User types in the search box → list filters by department name with debounce (300ms) → pagination resets to page 1 |
| **Clear Search** | User clears the search box → full department list restored → pagination resets to page 1 |
| **Paginate** | User changes page number or rows per page → list updates to show selected page |
| **Export** | User clicks Export button → system prepares and downloads the currently filtered list |
| **Add New** | User clicks "+ Add New" → navigates to Create Department page *(DR-008-001-02)* |
| **Gear → Edit** | User clicks gear icon on a row → selects Edit → navigates to Edit Department page *(DR-008-001-03)* |
| **Gear → Delete** | User clicks gear icon on a row → selects Delete → system checks employee count → proceeds or blocks *(DR-008-001-04)* |

**Exit Points:**
- **Add New** → Navigates away to Create Department page
- **Edit (via gear)** → Navigates away to Edit Department page
- **Delete (via gear)** → Stays on list page; either deletion completes and list refreshes, or system shows blocking message

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Search On | Trigger | Mandatory | Placeholder | Description |
|------------|------------|-----------|---------|-----------|-------------|-------------|
| Search | Text input | Department name | On change with 300ms debounce | No | "Search by department name" | Filters table by department name; case-insensitive; partial match supported |

### Interaction Elements

| Element | Type | Visible To | State / Condition | Trigger Action | Description |
|---------|------|-----------|-------------------|----------------|-------------|
| Export button | Button | All with view permission | Always visible | Downloads currently filtered list | Exports the current result set (filtered or full) |
| Add New button | Button | Management permission only | Hidden for view-only roles | Navigates to Create Department page | Opens DR-008-001-02 flow |
| Rows per page | Dropdown | All with view permission | Always visible | Updates rows shown per page | Options: 10, 25, 50 (default: 10) |
| Page navigation | Pagination control | All with view permission | Hidden when total ≤ rows per page | Moves between pages | Shows current page and total pages |
| Gear icon (per row) | Icon button | Management permission only | Hidden for view-only roles | Opens action menu | Shows Edit and Delete options for the row |

---

## 4. Data Display

### Information Shown to User

| Data | Column | Format | Empty State | Business Meaning |
|------|--------|--------|-------------|------------------|
| Department Name | Column 1 | Text, alphabetical order | — | The name of the organizational unit |
| No. of Employees | Column 2 | Number (integer) | 0 | Total employees currently assigned to this department |
| Action | Column 3 | Gear icon (management permission only) | — | Opens Edit / Delete action menu for this row |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens or data is being fetched | Skeleton rows displayed in table while data loads |
| Populated | Departments exist | Table with data rows, pagination controls below |
| Empty | No departments in the system | "No departments found" message; "Get started by adding your first department" sub-text; "+ Add New" button (management permission only) |
| No Results | Search returns zero matches | "No departments match your search" message; "Clear search" link to reset |
| Delete Blocked | User attempts to delete a department with assigned employees | Inline error or dialog: "Cannot delete — [X] employees are assigned to this department. Reassign all employees before deleting." |

### Page Layout (from Figma — node 3059:1722)

```
┌─────────────────────────────────────────────────────────────────┐
│  Organization Data > Departments                    [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Department List                                 │
│  200px       │                                                  │
│              │  [Search 320px]            [Export] [+ Add New]  │
│              │                                                  │
│              │  ┌──────────────┬──────────────┬──────────────┐  │
│              │  │ Department   │ No.Employees │ Action       │  │
│              │  ├──────────────┼──────────────┼──────────────┤  │
│              │  │ Name text    │ 0            │ ⚙            │  │
│              │  │ Name text    │ 12           │ ⚙            │  │
│              │  └──────────────┴──────────────┴──────────────┘  │
│              │                                                  │
│              │  Rows per page [10▼]  Page 1 of N   1 2 [3] 4 > │
└──────────────┴──────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** Department list displays a table with 3 columns — Department Name, No. of Employees, and Action
- **AC-02:** No. of Employees shows the total count of all employees assigned to that department
- **AC-03:** Search field filters the list by department name as the user types, with a debounce delay (approximately 300ms)
- **AC-04:** Search is case-insensitive and matches partial names (contains match — "eng" matches "Engineering")
- **AC-05:** When search returns no results, the message "No departments match your search" is displayed with a clear search option
- **AC-06:** Clearing the search field immediately restores the full department list and resets to page 1
- **AC-07:** List paginates with a default of 10 rows per page; user can change rows per page
- **AC-08:** Pagination controls are hidden when the total number of departments is ≤ the current rows per page setting
- **AC-09:** When a search is applied or cleared, pagination resets to page 1
- **AC-10:** Export button downloads the currently visible/filtered list (if search is active, only matching departments are exported)
- **AC-11:** Add New button and gear icon are visible only to users with department management permission
- **AC-12:** Users with view-only permission can browse, search, paginate, and export but cannot see the Add New button or gear icon
- **AC-13:** When no departments exist, the empty state shows "No departments found" with an Add New CTA (visible to management permission users only)
- **AC-14:** If a user attempts to delete a department with employees assigned, the system blocks the deletion and shows: "Cannot delete — [X] employees are assigned to this department. Reassign all employees before deleting."
- **AC-15:** Page breadcrumb displays "Organization Data > Departments" to indicate current location

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| List loads with departments | User navigates to Departments page | Table displays all departments alphabetically, paginated | High |
| Search with match | User types "eng" | Only departments containing "eng" displayed | High |
| Search with no match | User types "zzz" | No results state shown with clear search option | High |
| Search case-insensitive | User types "ENGINEERING" | Matches "Engineering", "engineering" | Medium |
| Export filtered list | Search active, user clicks Export | Only filtered results downloaded | High |
| Add New hidden (view-only) | User with view-only role | Add New button not visible | High |
| Gear icon hidden (view-only) | User with view-only role | No gear icon shown in Action column | High |
| Empty state (no departments) | No departments in system | Empty state with message and Add New CTA (if permitted) | Medium |
| Delete blocked | User deletes dept with 5 employees | Blocking message shows "5 employees assigned" | High |
| Delete allowed | User deletes dept with 0 employees | Confirmation → department removed → list refreshes | High |
| Pagination reset on search | User on page 3, applies search | Returns to page 1 with filtered results | Medium |

---

## 6. System Rules

- **SR-01:** Deletion is blocked if the department has one or more employees assigned — the system checks the employee count before allowing the delete action to proceed
- **SR-02:** Employee count per department is a live count calculated from current employee-department assignments, not a cached value
- **SR-03:** Access to the Department List page is controlled by view permission configured in US-004 — users without view permission cannot access this page
- **SR-04:** The Add New button, gear icon (Edit / Delete actions) are only displayed to roles granted department management permission via US-004
- **SR-05:** Search applies to the department name field only — no other fields (e.g., employee count) are searched
- **SR-06:** Export produces the currently filtered result set — if a search is active, only matching departments are included in the export
- **SR-07:** Search matching is case-insensitive (e.g., "eng" matches "Engineering", "ENGINEERING", "engineering")
- **SR-08:** The default display order for departments is alphabetical by department name (A → Z)
- **SR-09:** When a search is applied or cleared, pagination resets to page 1 to avoid showing an empty page

**State Transitions:**
```
[List loads] → [Skeleton shown] → [Data fetched] → [Table displayed]
[User types in search] → [Debounce 300ms] → [List filters] → [Pagination resets to page 1]
[User clears search] → [Debounce 300ms] → [Full list restored] → [Pagination resets to page 1]
[User clicks Export] → [Export loading state] → [File downloaded]
[User clicks Delete (0 employees)] → [Confirmation] → [Department deleted] → [List refreshes]
[User clicks Delete (>0 employees)] → [System blocks] → [Error message shown] → [No changes made]
```

**Dependencies:**
- **US-004 (Role & Permission Management):** Controls which roles can view and manage departments; all permission checks on this page depend on US-004 being operational
- **EP-002 (Employee Management):** Employee count per department is sourced from employee-department assignments in EP-002

---

## 7. UX Optimizations

- **UX-01:** Search input uses a debounce delay (~300ms) so the list does not filter on every single keystroke — reduces unnecessary load and improves perceived performance
- **UX-02:** Skeleton rows are shown during page load to prevent layout shift and indicate that content is loading
- **UX-03:** The gear icon action menu closes automatically when the user clicks anywhere outside it
- **UX-04:** The Export button shows a brief loading state (spinner or disabled state) while the file is being prepared to prevent double-clicks
- **UX-05:** The delete blocking message clearly states the exact number of assigned employees and the required action: "Cannot delete — [X] employees are assigned. Reassign all employees before deleting."
- **UX-06:** Pagination controls are hidden when the total number of departments fits on one page — no unnecessary UI clutter
- **UX-07:** This screen is designed for desktop use (1920×1080 reference resolution) — no mobile or tablet adaptation required for this release

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Full layout: sidebar 200px + content 1694px |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [ ] Keyboard navigable (Tab through search, Export, Add New, pagination)
- [ ] Screen reader compatible (table column headers, button labels)
- [ ] Sufficient color contrast (design tokens: foreground #0a0a0a on background #ffffff)
- [ ] Focus indicators visible on all interactive elements

**Design Reference:**
- Figma node `3059:1722` — Department List screen
- Design tokens: primary `#171717`, background `#ffffff`, border `#e5e5e5`, muted `#f5f5f5`, font: Geist, border-radius: 6px

---

## 8. Additional Information

### Out of Scope

- Department detail/drill-down view (clicking a department name — this list does not navigate to a detail page)
- Column sorting by user interaction (alphabetical default only; no user-controlled sort toggle)
- Advanced filtering beyond name search (e.g., filter by employee count range)
- Bulk import of departments from external files
- Org chart or visual hierarchy representation
- Mobile or tablet responsive layout (desktop-first application)

### Open Questions

- [ ] **Export format:** What file format should the export produce — CSV or Excel (.xlsx)? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-008-001-02: Create Department | Triggered from "+ Add New" button on this list |
| DR-008-001-03: Edit Department | Triggered from gear icon → Edit on a list row |
| DR-008-001-04: Delete Department | Triggered from gear icon → Delete on a list row |
| US-004: Role & Permission Management | Upstream dependency — controls all access on this page |
| EP-002: Employee Management | Source of employee count per department |

### Notes

- **Requirements change:** During detail requirement gathering, the Product Owner confirmed departments have no active/inactive status. The Deactivate (soft-delete) flow from the original REQUIREMENTS.md (US-001-04) is replaced by a **hard Delete** that is blocked when employees are assigned. REQUIREMENTS.md, FLOWCHART.md, and ANALYSIS.md must be updated to reflect this change.
- **Employee count** is a total count of all employees assigned to the department (flat list — no sub-department hierarchy).
- **Gear icon actions** confirmed as: Edit + Delete (pending original open question now resolved by this DR).

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Team | 2026-03-03 | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-03 | BA Team | Initial draft |
