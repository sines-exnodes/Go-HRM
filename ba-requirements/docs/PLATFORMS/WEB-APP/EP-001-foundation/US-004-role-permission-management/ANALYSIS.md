---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
status: draft
version: "1.0"
last_updated: "2026-03-24"
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
    description: "Role List screen"
    node_id: "3081:1680"
    extraction_date: "2026-03-19"
  - type: figma
    description: "Create A New Role screen"
    node_id: "3083:1837"
    extraction_date: "2026-03-24"
---

# Analysis: Role & Permission Management

**Epic:** EP-001 (Foundation)
**Story:** US-004-role-permission-management
**Status:** 🔴 Draft

---

## Table of Contents

1. [Business Context](#1-business-context)
2. [Scope Definition](#2-scope-definition)
3. [Requirements Analysis](#3-requirements-analysis)
4. [Data Flow Analysis](#4-data-flow-analysis)
5. [User Journey Mapping](#5-user-journey-mapping)
6. [Business Rules & Constraints](#6-business-rules--constraints)
7. [Design Context [ADD-ON]](#7-design-context-add-on)
8. [Success Criteria & Metrics](#8-success-criteria--metrics)
9. [Risk Assessment](#9-risk-assessment)
10. [Assumptions & Notes](#10-assumptions--notes)

---

## 1. Business Context

### Problem Statement

Organizations need a flexible way to define user roles and assign permissions so that each user accesses only the features they are authorized for. Without centralized role and permission management, access control is hardcoded and inflexible — adding a new role or adjusting what a role can do requires a code change instead of an administrator action. This story delivers the tools for administrators to define roles, assign permissions to roles, and control what each role can see and do across the entire HRM platform.

### Stakeholders

- **Primary Users**: Administrators with role management permission
- **Secondary Users**: All HRM users (affected by role assignments); all other modules (depend on US-004 for permission enforcement)
- **Business Owner**: HR department leadership / IT administration

### Business Goals

- Goal 1: Provide a centralized, administrator-managed system for defining roles and their permissions — no hardcoded role names
- Goal 2: Ensure every feature across all HRM modules enforces access control based on the roles and permissions defined here

---

## 2. Scope Definition

### In Scope

- Role list view (search, table, pagination)
- Create new role
- Edit existing role (name and permissions)
- Delete role (blocked if users are assigned)
- Permission matrix (assign/revoke permissions per role)
- Access controlled by role management permission (self-referential — US-004 manages its own access)

### Out of Scope

- User-to-role assignment (managed in User Management — US within EP-001)
- Predefined system roles (Super Admin, etc.) — these may be seeded but are still editable
- Audit log of permission changes (future enhancement)
- Hierarchical role inheritance (roles are flat — each role has its own permission set)

### Dependencies

- **Internal**: US-001 (Authentication) — user must be signed in
- **Downstream**: All other stories and epics — every permission check across the platform references US-004's role-permission data

---

## 3. Requirements Analysis

### Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-004-01 | Role list view | Critical | Table with Role Name, Permission, Action columns |
| FR-US-004-02 | Search roles | High | Filter by role name |
| FR-US-004-03 | Pagination | Medium | Configurable rows per page |
| FR-US-004-04 | Create role | Critical | Role name + permission assignment |
| FR-US-004-05 | Edit role | High | Update role name and/or permissions |
| FR-US-004-06 | Delete role | High | Blocked if users are assigned to this role |
| FR-US-004-07 | Permission matrix | Critical | Assign/revoke individual permissions per role |
| FR-US-004-08 | Role name uniqueness | Critical | No duplicate role names (case-insensitive) |
| FR-US-004-09 | Action menu per row | High | Gear icon with Edit / Delete options |

### Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | Administrator can create a role and assign permissions in under 5 minutes | Clear permission grouping |
| Data Integrity | Roles in use by users cannot be deleted | Deletion blocked server-side |
| Access Control | Only authorized roles can manage other roles | Self-referential permission enforcement |

---

## 4. Data Flow Analysis

### Data Entities

- **Role**: Name (unique), list of assigned permissions
- **Permission**: Individual access right (e.g., "view departments", "manage users")
- **User** (external): References role as assignment

### Data Relationships

- Role → Permissions: Many-to-many (one role has many permissions; one permission can be assigned to many roles)
- Role → Users: One-to-many (one role can have many users assigned)

### Data Flow

- [Administrator creates role] → [Administrator assigns permissions] → [Role available for user assignment]
- [Administrator edits role permissions] → [All users with that role immediately gain/lose access accordingly]
- [Administrator attempts delete] → [System checks user count] → [If 0 users: Confirmation → delete | If ≥1 users: Blocked]

---

## 5. User Journey Mapping

### Primary User Flow: View and Manage Roles

```
Administrator navigates to Users Management → Roles →
Role List displayed with Role Name, Permission, and Action columns →
Administrator searches/browses roles →
Clicks "+ Add New" to create a new role →
OR clicks gear icon → Edit to modify role name/permissions →
OR clicks gear icon → Delete to remove unused role
```

### Key Touch Points

1. Role list (first view — browse, search)
2. Add New form/page (create role + assign permissions)
3. Edit form/page (modify role name and permissions)
4. Gear icon actions (Edit / Delete per row)
5. Delete dialog (Confirmation or Blocked)

---

## 6. Business Rules & Constraints

### Business Rules

- BR-US-004-01: Role names must be unique within the organization (case-insensitive)
- BR-US-004-02: Roles are a flat list — no role inheritance or hierarchy
- BR-US-004-03: A role with ≥1 user assigned cannot be deleted; system blocks deletion and shows user count
- BR-US-004-04: Permission changes to a role take effect immediately for all users assigned to that role
- BR-US-004-05: The permission to manage roles is itself a permission — an administrator can lock themselves out if they remove their own role management permission (confirm with PO if this should be prevented)

### Constraints

- No hardcoded role names — all roles are administrator-defined
- Permissions are system-defined (developers define the permission keys; administrators assign them to roles)

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract`.
> Two screens extracted: Role List (2026-03-19), Create A New Role (2026-03-24).

### Source Information

| Screen | Figma Frame | Node ID | Dimensions | Extraction Date |
|--------|------------|---------|------------|-----------------|
| Role List | [Role List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3081-1680) | `3081:1680` | 1920 × 1080 | 2026-03-19 |
| Create A New Role | [Create A New Role](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3083-1837) | `3083:1837` | 1920 × 1080 | 2026-03-24 |

**Figma File:** [Exnodes HRM](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC)

> URLs are clickable — open directly in Figma

---

### Screen: Role List

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Logo]      │  Role List                                       │
│              │                                                  │
│  Richard R.  │  [Search 320px]                    [+ Add New]  │
│  Super Admin │                                                  │
│  ──────────  │  ┌──────────┬─────────────────────┬──────────┐  │
│  Users Mgmt  │  │Role Name │ Permission          │ Action   │  │
│  > Users     │  ├──────────┼─────────────────────┼──────────┤  │
│  > Roles  ◄  │  │ Text     │ Text                │ ⚙        │  │
│  > Perms     │  │ Text     │ Text                │ ⚙        │  │
│  ──────────  │  │ Text     │ Text                │ ⚙        │  │
│  Menu Sect.  │  │ Text     │ Text                │ ⚙        │  │
│  > Menu 1   │  │ Text     │ Text                │ ⚙        │  │
│  > Menu 2   │  │ Text     │ Text                │ ⚙        │  │
│  > Menu 3   │  └──────────┴─────────────────────┴──────────┘  │
│              │                                                  │
│              │  Rows per page [10▼]  Page 1 of 10  1…2 [3] 4…5 >│
└──────────────┴──────────────────────────────────────────────────┘
Sidebar: 200px │ Content: 1694px
```

#### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Sidebar navigation | `3081:1681` | Primary navigation with collapsible sections | Designed |
| Logo (horizontal) | `3081:1683` | Brand identity in sidebar top | Designed |
| User profile card | `3081:1684` | Shows current user name + role | Designed |
| Nav section: Users Management | `3081:1691` | Collapsible nav group (Users, Roles, Permissions) | Designed |
| Nav item: Roles | `3081:1700` | Active page — current page | Designed |
| Nav section: Menu Section | `3081:1704` | Placeholder collapsible group (Menu 1, Menu 2, Menu 3) | Placeholder |
| Breadcrumb bar | `3081:1718` | Page location path — placeholder text | Placeholder |
| Sidebar toggle | `3081:1719` | Toggle sidebar visibility (SidebarSimple icon) | Designed |
| Page title "Role List" | `3081:1727` | Page heading (Heading 3: Geist Semibold 24px) | Designed |
| Search input | `3081:1731` | Filter list by role name — 320px width | Designed |
| Add New button | `3081:1739` | Create new role — primary button, right-aligned | Designed |
| Table: Role Name column | `3081:1741` | Lists role names — 181px width | Placeholder data |
| Table: Permission column | `3081:1742` | Shows permissions per role — 1380px width | Placeholder data |
| Table: Action column | `3081:1749` | Per-row gear icon for actions — 93px width | Designed |
| Pagination | `3081:1750` | Navigate large role lists (rows per page + page numbers) | Designed |

#### Design Constraints

- **Sidebar width:** 200px fixed left
- **Content area width:** 1694px (remaining from 1920px)
- **Table column widths:** Role Name 181px | Permission 1380px | Action 93px
- **Search input width:** 320px, left-aligned in action bar
- **Button layout:** Only "+ Add New" right-aligned (no Export button — unlike Department/Position lists)
- **Pagination:** Rows per page selector (default 10) + numbered page navigation
- **Page title:** "Role List" — Geist Semibold, 24px, line-height 28.8px, letter-spacing -1

#### Design Tokens Referenced

| Token | Value | Used For |
|-------|-------|----------|
| `general/primary` | `#171717` | Add New button background |
| `general/primary foreground` | `#fafafa` | Add New button text |
| `general/background` | `#ffffff` | Page + table background |
| `general/border` | `#e5e5e5` | Table borders, dividers |
| `general/muted` | `#f5f5f5` | Table header background |
| `general/muted foreground` | `#737373` | Placeholder text, secondary labels |
| `general/foreground` | `#0a0a0a` | Primary text |
| `text/text-foreground` | `#09090b` | Table cell text |
| `border/border-input` | `#e4e4e7` | Search input border |
| `general/input` | `#ffffff` | Search input background |
| `heading 3` | Geist Semibold 24px / 28.8px LH / -1 LS | Page title |
| `paragraph small/medium` | Geist Medium 14px / 20px LH | Table header text |
| `paragraph small/regular` | Geist Regular 14px / 20px LH | Table body text |
| `paragraph mini/regular` | Geist Regular 12px / 16px LH | Pagination text |
| `rounded-md` | 6px | Button, input border-radius |
| Font family | Geist | All text |

#### Gaps Identified from Design — Role List

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| No Export button | Unlike Department/Position lists, Role List has no Export — deliberate or omission? | Confirm with Product Owner whether role export is needed |
| Permission column is very wide (1380px) | Permissions may be displayed as tags, badges, or comma-separated list | Confirm display format — tags, chips, or plain text? |
| Permission column content is placeholder ("Text") | Cannot determine how permissions are displayed | Request detailed design showing permission rendering (tags, badge list, truncation rules) |
| Gear icon actions undefined | Users don't know available actions per role | Design dropdown — expected: Edit, Delete (confirm with PO) |
| Breadcrumb is placeholder text | Navigation context unclear | Replace with real path: Users Management / Roles |
| "Menu Section" in sidebar is placeholder | Future navigation items not yet defined | Confirm with Product Owner what modules appear here |
| Search placeholder not visible | Cannot confirm search hint text | Expected: "Search by role name..." (confirm) |
| No user count column | Cannot tell how many users are assigned to each role at a glance | Consider adding a "No. of Users" column (aids in delete-blocking awareness) |
| ~~Create/Edit role screens not designed~~ | ~~Cannot define the form layout for role creation or permission assignment~~ | ✅ Resolved — Create Role screen extracted 2026-03-24 (Edit Role still pending) |

---

### Screen: Create A New Role

> Extracted 2026-03-24 from Figma node `3083:1837`.

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Logo]      │  Create A New Role            [Cancel] [💾 Save] │
│              │                                                  │
│  Richard R.  │  ┌────────────────────────────────────────┐      │
│  Super Admin │  │ Role Information                        │      │
│  ──────────  │  │                                        │      │
│  Users Mgmt  │  │ * Role name                            │      │
│  > Users     │  │ ┌──────────────────────────────┐       │      │
│  > Roles  ◄  │  │ │ Enter role name              │       │      │
│  > Perms     │  │ └──────────────────────────────┘       │      │
│  ──────────  │  └────────────────────────────────────────┘      │
│  Org Data    │                                                  │
│  > Depts     │  ┌────────────────────────────────────────┐      │
│  > Positions │  │ Permissions                             │      │
│  ──────────  │  │ ──────────────────────────────────────  │      │
│  Menu Sect.  │  │ Users                                   │      │
│  > Menu 1   │  │ ☐ View Users  ☐ Create  ☐ Edit  ☐ Del │      │
│  > Menu 2   │  │ ──────────────────────────────────────  │      │
│  > Menu 3   │  │ Roles                                   │      │
│              │  │ ☐ View Roles  ☐ Create  ☐ Edit  ☐ Del │      │
│              │  │ ──────────────────────────────────────  │      │
│              │  │ Module                                  │      │
│              │  │ ☐ Perm  ☐ Perm  ☐ Perm  ☐ Perm       │      │
│              │  │ ☐ Perm  ☐ Perm  ☐ Perm  ☐ Perm       │      │
│              │  │ ──────────────────────────────────────  │      │
│              │  │ Module                                  │      │
│              │  │ ☐ Perm  ☐ Perm  ☐ Perm  ☐ Perm       │      │
│              │  │ ☐ Perm  ☐ Perm  ☐ Perm  ☐ Perm       │      │
│              │  └────────────────────────────────────────┘      │
└──────────────┴──────────────────────────────────────────────────┘
Sidebar: 200px │ Content: 1694px │ Form card: 600px centered
```

#### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Create A New Role" | `3083:1895` | Page heading (Heading 3: Geist Semibold 24px) | Designed |
| Cancel button | `3083:1897` | Discard changes and return to Role List | Designed |
| Save button | `3083:1898` | Submit form to create the new role | Designed |
| Role Information card | `3083:1900` | Groups role identity fields | Designed |
| Role name input (Vertical Field) | `3083:1904` | Text input for role name — mandatory (*) | Designed |
| Permissions card | `3083:1957` | Groups all permission checkboxes by module | Designed |
| Permission group: Users | `3089:1785` | CRUD checkboxes for Users module | Designed |
| Permission group: Roles | `3089:1791` | CRUD checkboxes for Roles module | Designed |
| Permission group: Module (placeholder 1) | `3089:1812` | 8 placeholder permission checkboxes (4×2 grid) | Placeholder |
| Permission group: Module (placeholder 2) | `3089:1851` | 8 placeholder permission checkboxes (4×2 grid) | Placeholder |
| Checkbox Group instances | various | Individual permission checkboxes within each group | Designed |
| Separator lines | `3089:1789`, `3089:1787`, `3089:1798`, `3089:1850` | Horizontal dividers between permission groups | Designed |

#### Design Constraints

- **Page layout:** Full-page form (not modal) — consistent with Department/Position create forms
- **Form card width:** 600px, horizontally centered in content area (offset x=527)
- **Role Information card:** 600×141px — contains single mandatory field
- **Permissions card:** 600×504px — scrollable if more permission groups added
- **Role name field:** Vertical Field instance, 576px within card (12px padding each side)
- **Permission group layout:** Module label → horizontal row of 4 checkbox groups, separated by lines
- **Checkbox grid:** 4 checkboxes per row; if >4 permissions per module, wraps to 2nd row (as shown in Module placeholders: 8 checkboxes = 4+4)
- **Action buttons:** Cancel (secondary) + Save (primary), right-aligned in title bar area
- **Breadcrumb:** Placeholder text (expected: Users Management / Roles / Create A New Role)

#### Permission Matrix Structure (from design)

| Module Group | Permissions Shown | Pattern |
|-------------|-------------------|---------|
| **Users** | View Users, Create Users, Edit Users, Delete Users | CRUD (4 checkboxes) |
| **Roles** | View Roles, Create Roles, Edit Roles, Delete Roles | CRUD (4 checkboxes) |
| **Module** (placeholder) | 8× "Permission" placeholder | Extended CRUD (4+4 checkboxes in 2 rows) |
| **Module** (placeholder) | 8× "Permission" placeholder | Extended CRUD (4+4 checkboxes in 2 rows) |

> **Key insight:** The permission matrix uses a **module-grouped checkbox pattern** where each module has its own set of permissions. The confirmed modules (Users, Roles) follow a standard View/Create/Edit/Delete pattern. Placeholder modules show up to 8 permissions per module in a 4-column grid.

#### Design Tokens Referenced (Create Role specific)

| Token | Value | Used For |
|-------|-------|----------|
| `heading 3` | Geist Semibold 24px / 28.8px LH / -1 LS | Page title "Create A New Role" |
| `general/secondary` | `#f5f5f5` | Page background |
| `general/background` | `#ffffff` | Form card background |
| `general/border` | `#e5e5e5` | Card borders, permission group separators |
| `general/foreground` | `#0a0a0a` | Section titles, field labels |
| `paragraph small/medium` | Geist Medium 14px / 20px LH | Section titles ("Role Information", "Permissions") |
| `paragraph small/regular` | Geist Regular 14px / 20px LH | Module group labels ("Users", "Roles"), checkbox labels |
| `rounded-lg` | 8px | Form card border-radius |
| `shadow-2xs` | 0px 1px 0px transparent | Card shadow |
| `general/primary` | `#171717` | Save button background |
| `general/primary foreground` | `#fafafa` | Save button text |

#### Gaps Identified from Design — Create Role

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Placeholder "Module" labels (×2) | Cannot determine which actual modules appear | Confirm full module list with Product Owner (e.g., Departments, Positions, Attendance, Leave, Payroll, etc.) |
| Placeholder "Permission" checkbox labels (×16) | Cannot determine actual permission names for non-Users/Roles modules | Confirm permission names per module |
| No validation error states shown | Cannot define inline error presentation for role name | Design Team: provide error state for duplicate name, empty name |
| No success confirmation after save | Cannot define post-save behavior | Confirm: redirect to Role List with toast? Or stay on form? |
| Edit Role screen still not extracted | Cannot confirm if Edit Role mirrors Create Role | Request Edit Role screen (expected: identical layout with pre-filled data) |
| No "Select All" option per module | Administrator must check each permission individually | Consider adding "Select All" checkbox per module group for efficiency |
| Breadcrumb is placeholder | Navigation path unclear | Expected: Users Management / Roles / Create A New Role |
| Cancel button behavior undefined | User may lose unsaved data | Confirm: redirect to Role List immediately, or show "discard changes?" dialog? |

---

## 8. Success Criteria & Metrics

### Functional Success Criteria

- [ ] Administrator can create a role and assign permissions in under 5 minutes
- [ ] Role list displays correctly with search and pagination
- [ ] Delete is blocked when users are assigned to a role
- [ ] Permission changes to a role take effect immediately for all assigned users
- [ ] Only roles with role management permission can access create/edit/delete actions

### Business Metrics

- Access control consistency: 100% of features across all modules enforce role-based permissions defined in US-004
- Zero hardcoded role names: All roles are administrator-defined and configurable

---

## 9. Risk Assessment

### Identified Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Role deletion breaks user access | Low | High | System blocks deletion if users assigned; user must reassign first |
| Administrator locks themselves out | Medium | High | Confirm with PO: prevent removing own role management permission? |
| Permission model too granular | Medium | Medium | Group permissions by module for clear organization |
| Permission changes have immediate effect | Low | Medium | Show warning when editing permissions of a role with many assigned users |

---

## 10. Assumptions & Notes

### Assumptions

1. Role names are unique organization-wide
2. Roles are a flat list — no inheritance or hierarchy
3. Hard delete for roles; no active/inactive status
4. Permissions are system-defined keys; administrators assign them to roles but cannot create new permission keys
5. The "Permission" column in the Role List shows a summary or preview of assigned permissions (display format TBD)
6. ~~Create/Edit role screens will follow the same full-page form pattern as Department/Position management~~ **Confirmed** — Create Role screen uses full-page form pattern (extracted 2026-03-24)

### Open Questions

- [ ] **Permission display format:** How should the Permission column in the Role List display permissions — tags/badges, comma-separated text, count only, or truncated list? — **Owner:** Design Team — **Status:** Pending
- [ ] **No Export button:** Is omitting the Export button from the Role List intentional, or should it be added for consistency with other list screens? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Self-lock prevention:** Should the system prevent an administrator from removing their own role management permission? — **Owner:** Product Owner — **Status:** Pending
- [x] **Create Role design:** ~~When will Figma screens for Create Role be available?~~ — **Resolved 2026-03-24:** Create Role screen extracted (node `3083:1837`). Edit Role screen still pending.
- [ ] **Edit Role design:** When will Figma screen for Edit Role be available? Expected to mirror Create Role with pre-filled data. — **Owner:** Design Team — **Status:** Pending
- [ ] **Cancel button behavior:** Should Cancel redirect to Role List immediately, or show a "discard changes?" confirmation dialog? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Post-save behavior:** After saving a new role, should the system redirect to Role List with success toast, or stay on the form? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Select All per module:** Should each permission module group have a "Select All" checkbox for efficiency? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Full module list:** Which modules appear in the permission matrix? Only Users and Roles are confirmed in design; placeholders suggest more. — **Owner:** Product Owner — **Status:** Pending
- [ ] **User count column:** Should a "No. of Users" column be added to the Role List (similar to "No. of Employees" in Department/Position lists)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Gear icon actions:** Confirmed as Edit + Delete? Any other options (e.g., Duplicate role)? — **Owner:** Product Owner — **Status:** Pending

---

**Document Version:** 1.0
**Last Updated:** 2026-03-19
**Author:** BA Agent
**Reviewer:** Pending
