---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
status: draft
version: "0.1"
last_updated: "2026-03-24"
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

# Requirements: Role & Permission Management

**Epic:** EP-001 (Foundation)
**Story:** US-004-role-permission-management
**Status:** 🔴 Draft — Stub (pending full requirements elaboration)

---

> **Note:** This is a stub document. Full requirements will be elaborated after:
> 1. Open questions from ANALYSIS.md are resolved with Product Owner
> 2. Create/Edit Role Figma screens are delivered by Design Team
> 3. Permission matrix design is confirmed

---

## 1. User Stories

| Story ID | As a... | I want to... | So that... | Priority |
|----------|---------|-------------|------------|----------|
| US-004-01 | User with role management permission | View all roles in a searchable list | I can browse and manage the organization's role catalog | Critical |
| US-004-02 | User with role management permission | Create a new role with assigned permissions | New access patterns can be defined as the organization grows | Critical |
| US-004-03 | User with role management permission | Edit an existing role's name and permissions | Role definitions stay accurate as business needs change | High |
| US-004-04 | User with role management permission | Delete a role that is no longer in use | The role list stays clean and free of obsolete entries | High |

---

## 2. Functional Requirements (Preliminary)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-US-004-01 | Role list view with Role Name, Permission, and Action columns | Critical | From Figma |
| FR-US-004-02 | Search by role name | High | From Figma |
| FR-US-004-03 | Pagination with configurable rows per page | Medium | From Figma |
| FR-US-004-04 | Create role (name + permission assignment) | Critical | Pending design |
| FR-US-004-05 | Edit role (name and/or permissions) | High | Pending design |
| FR-US-004-06 | Delete role (blocked if users assigned) | High | Pattern from US-001 |
| FR-US-004-07 | Permission matrix (assign/revoke per role) | Critical | Pending design |
| FR-US-004-08 | Role name uniqueness (case-insensitive) | Critical | Business rule |
| FR-US-004-09 | Gear icon action menu (Edit / Delete) | High | From Figma |

---

## 3. Business Rules (Preliminary)

| ID | Rule |
|----|------|
| BR-US-004-01 | Role names must be unique within the organization (case-insensitive) |
| BR-US-004-02 | Roles are a flat list — no role inheritance or hierarchy |
| BR-US-004-03 | A role with ≥1 user assigned cannot be deleted |
| BR-US-004-04 | Permission changes take effect immediately for all assigned users |

---

## 5. UI Specifications [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-19.
> Source: [Role List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3081-1680) (node `3081:1680`)

### 5.1 Role List — Page Structure

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page title | "Role List" — Geist Semibold 24px, line-height 28.8px, letter-spacing -1, color `#0a0a0a` | `3081:1727` |
| Breadcrumb | Users Management / Roles (placeholder in design — needs real path) | `3081:1718` |
| Sidebar toggle | SidebarSimple icon, 16×16, top-left of content area | `3081:1719` |

### 5.2 Role List — Action Bar

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Search input | Width: 320px, left-aligned. Placeholder: "Search by role name.." Border: `#e4e4e7`, bg: `#ffffff`, radius: 6px | `3081:1731` |
| Add New button | Right-aligned. Label: "+ Add New". Primary style: bg `#171717`, text `#fafafa`, radius: 6px. Visible only to users with role management permission. | `3081:1739` |

**Note:** No Export button is present on the Role List (unlike Department/Position lists). Pending confirmation from Product Owner.

### 5.3 Role List — Table

| Column | Width | Content | Figma Node |
|--------|-------|---------|------------|
| Role Name | 181px | Role name text — Geist Regular 14px, color `#09090b` | `3081:1741` |
| Permission | 1380px | Permission display (format TBD — tags, badges, or text) — placeholder "Text" in design | `3081:1742` |
| Action | 93px | Gear icon (⚙) — opens dropdown with Edit / Delete. Visible only to users with role management permission. | `3081:1749` |

**Table header:** Geist Medium 14px, bg `#f5f5f5`, color `#737373`
**Table border:** `#e5e5e5`
**Table background:** `#ffffff`

### 5.4 Role List — Pagination

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Rows per page | Dropdown, default: 10. Options: 10, 25, 50. Geist Regular 12px. | `3081:1750` |
| Page indicator | "Page 1 of 10" — Geist Regular 12px, color `#737373` | `3081:1750` |
| Page navigation | Numbered buttons with prev/next arrows. Active page: outlined. | `3081:1750` |

### 5.5 Role List — Sidebar Navigation Context

| Nav Section | Items | Active Item |
|-------------|-------|-------------|
| Users Management | Users, Roles, Permissions | **Roles** (current page) |
| Menu Section (placeholder) | Menu 1, Menu 2, Menu 3 | — |

**Sidebar icons:**
- Users → `UsersThree` (node `3081:1696`)
- Roles → `UserGear` (node `3081:1699`)
- Permissions → `Key` (node `3081:1702`)

### 5.6 Acceptance Criteria from Design

- **AC-UI-01:** Role List page displays with page title "Role List" in Geist Semibold 24px
- **AC-UI-02:** Search input is 320px wide, left-aligned, with placeholder text
- **AC-UI-03:** "+ Add New" button is right-aligned, primary style (dark bg, white text)
- **AC-UI-04:** Table has 3 columns: Role Name (181px), Permission (1380px), Action (93px)
- **AC-UI-05:** Table header row uses muted background (`#f5f5f5`) with medium-weight text
- **AC-UI-06:** Each data row has a gear icon in the Action column
- **AC-UI-07:** Pagination displays below the table with rows-per-page selector and page navigation
- **AC-UI-08:** Sidebar shows "Roles" as the active navigation item under "Users Management"
- **AC-UI-09:** "+ Add New" button and gear icon are hidden for users without role management permission

### 5.7 Design Gaps Requiring Resolution

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 1 | Permission column display format unknown | Cannot implement Permission column | Design Team: provide permission rendering spec (tags/badges/text) |
| 2 | No Export button | Inconsistency with other list screens | Product Owner: confirm if intentional |
| 3 | No user count column | Cannot see role usage at a glance | Product Owner: confirm if "No. of Users" column needed |
| ~~4~~ | ~~Create Role screen not designed~~ | ~~Cannot define create form~~ | ✅ Resolved 2026-03-24 — see Section 5.8 |
| 5 | Edit Role screen not designed | Cannot define edit form | Design Team: deliver Edit Role screen |
| 6 | Gear icon dropdown not designed | Actions not visually defined | Design Team: confirm Edit + Delete in dropdown |
| 7 | Empty state not designed | No guidance when no roles exist | Design Team: provide empty state design |

### 5.8 Create A New Role — Page Structure

> Extracted from Figma on 2026-03-24.
> Source: [Create A New Role](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3083-1837) (node `3083:1837`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page title | "Create A New Role" — Geist Semibold 24px, line-height 28.8px, letter-spacing -1, color `#0a0a0a` | `3083:1895` |
| Breadcrumb | Placeholder (expected: Users Management / Roles / Create A New Role) | `3083:1887` |
| Page layout | Full-page form (not modal), consistent with Department/Position create forms | `3083:1893` |

### 5.9 Create A New Role — Action Bar

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Cancel button | Left button in action bar. Secondary style. Returns to Role List. | `3083:1897` |
| Save button | Right button in action bar. Primary style: bg `#171717`, text `#fafafa`, radius: 6px. Submits form. | `3083:1898` |

### 5.10 Create A New Role — Role Information Card

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600×141px, centered (x=527), bg `#ffffff`, radius: 8px, shadow-2xs | `3083:1900` |
| Section title | "Role Information" — Geist Medium 14px, color `#0a0a0a` | `3083:1902` |
| Role name field | Vertical Field instance. Label: "Role name" (mandatory — marked with *). Placeholder: "Enter role name". Width: 576px (12px padding each side). | `3083:1904` |

### 5.11 Create A New Role — Permissions Card

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600×504px, below Role Information card, same centering and styling | `3083:1957` |
| Section title | "Permissions" — Geist Medium 14px, color `#0a0a0a` | `3083:1959` |
| Group separators | Horizontal lines (`#e5e5e5`) between each permission group | various |

### 5.12 Create A New Role — Permission Groups

| Module Group | Permissions | Layout | Figma Node |
|-------------|-------------|--------|------------|
| **Users** | View Users, Create Users, Edit Users, Delete Users | 4 checkboxes in 1 row | `3089:1785` |
| **Roles** | View Roles, Create Roles, Edit Roles, Delete Roles | 4 checkboxes in 1 row | `3089:1791` |
| **Module** (placeholder) | 8× "Permission" placeholder | 4 checkboxes × 2 rows | `3089:1812` |
| **Module** (placeholder) | 8× "Permission" placeholder | 4 checkboxes × 2 rows | `3089:1851` |

**Permission group pattern:**
- Module label (Geist Regular 14px, 20px line-height) at top
- Checkbox Group instances arranged in 4-column grid
- If >4 permissions per module, wraps to additional rows
- Each checkbox: Checkbox Group component instance with label
- Horizontal separator line between groups

### 5.13 Acceptance Criteria from Design — Create Role

- **AC-UI-10:** Create Role page displays with page title "Create A New Role" in Geist Semibold 24px
- **AC-UI-11:** Cancel and Save buttons are right-aligned in the title action bar
- **AC-UI-12:** Role Information card is 600px wide, centered in content area
- **AC-UI-13:** Role name field is mandatory (marked with *) with placeholder "Enter role name"
- **AC-UI-14:** Permissions card displays below Role Information with grouped checkboxes
- **AC-UI-15:** Permission checkboxes are grouped by module with a module label header
- **AC-UI-16:** Each module group shows permission checkboxes in a 4-column grid layout
- **AC-UI-17:** Users module shows: View Users, Create Users, Edit Users, Delete Users
- **AC-UI-18:** Roles module shows: View Roles, Create Roles, Edit Roles, Delete Roles
- **AC-UI-19:** Horizontal separator lines divide each permission module group
- **AC-UI-20:** All checkboxes are unchecked by default (new role starts with no permissions)
- **AC-UI-21:** Create Role page is accessible only to users with role management permission

### 5.14 Design Gaps — Create Role

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 8 | Placeholder "Module" labels (×2) | Cannot determine full module list | Product Owner: confirm which modules appear in permission matrix |
| 9 | Placeholder "Permission" labels (×16) | Cannot determine permission names for other modules | Product Owner: confirm permission names per module |
| 10 | No validation error states | Cannot define error presentation for duplicate/empty name | Design Team: provide inline error state |
| 11 | No success confirmation | Cannot define post-save behavior | Product Owner: redirect to Role List with toast, or stay on form? |
| 12 | No "Select All" per module | Administrator must check individually | Product Owner: consider adding Select All for efficiency |
| 13 | Cancel behavior undefined | May lose unsaved data without warning | Product Owner: immediate redirect or "discard changes?" dialog? |

---

## 6. Next Steps

- [ ] Resolve open questions with Product Owner (see ANALYSIS.md Section 10)
- [ ] Obtain Create/Edit Role Figma screens from Design Team
- [ ] Elaborate full requirements including acceptance criteria, NFRs, and data requirements
- [ ] Write detail requirements (DRs) for each feature

---

**Document Version:** 0.1
**Last Updated:** 2026-03-19
**Author:** BA Agent
**Reviewer:** Pending
