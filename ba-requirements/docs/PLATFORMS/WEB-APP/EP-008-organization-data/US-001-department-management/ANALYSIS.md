---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-001
story_name: "Department Management"
status: draft
version: "1.0"
last_updated: "2026-03-03"
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
    description: "Department List screen"
    node_id: "3059:1722"
    extraction_date: "2026-03-03"
---

# Analysis: Department Management

**Epic:** EP-008 (Organization Data)
**Story:** US-001-department-management
**Status:** 🟡 In Progress

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

Organizations need a structured way to define and maintain their department hierarchy. Without a centralized department registry, employee records, approval workflows, and reports become inconsistent — different modules may reference the same department by different names or not at all. This story delivers the tools for administrators to create and manage the organization's department structure as authoritative reference data.

### Stakeholders

- **Primary Users**: Administrators with department management permission (configurable via US-004)
- **Secondary Users**: Managers (read access — view their department context); all other modules that reference department data
- **Business Owner**: HR department leadership

### Business Goals

- Goal 1: Provide a single source of truth for all departments used across HR modules
- Goal 2: Allow administrators to maintain department records as the organization evolves (add, edit, deactivate) without disrupting historical employee records

---

## 2. Scope Definition

### In Scope

- Department list view (search, table, pagination)
- Create new department
- Edit existing department
- Deactivate department (not permanent deletion)
- Export department list
- Access controlled by role permission (US-004)

### Out of Scope

- Org chart / visual hierarchy tree (future enhancement)
- Headcount planning by department (future reporting epic)
- Bulk import of departments from external files
- Automatic department merges or restructuring

### Dependencies

- **Internal**: US-004 (Role & Permission Management) — access control for create/edit/deactivate
- **Internal**: US-002-position-management — positions may reference departments
- **Downstream**: EP-002 Employee Management — employees assigned to departments

---

## 3. Requirements Analysis

### Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-001-01 | Department list view | Critical | Table with search, pagination, export |
| FR-US-001-02 | Create department | Critical | Single-field form — department name only |
| FR-US-001-03 | Edit department | High | Update department name |
| FR-US-001-04 | Deactivate department | High | Soft deactivation, not deletion |
| FR-US-001-05 | Search departments | High | Search by department name |
| FR-US-001-06 | Export department list | Medium | Export visible list data |
| FR-US-001-07 | Pagination | Medium | Configurable rows per page |
| FR-US-001-08 | Name uniqueness validation | Critical | No duplicate department names allowed |

### Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | Administrators can add a department in under 60 seconds | Single-form workflow |
| Data Integrity | Departments in use by employees cannot be deleted | Deactivation only |
| Access Control | Only authorized roles can create/edit/deactivate | Enforced via US-004 |

---

## 4. Data Flow Analysis

### Data Entities

- **Department**: Name, status (active/inactive), employee count
- **Employee** (downstream): References department as assignment

### Data Relationships

- Department → Employees: One-to-many (one department has many employees)

### Data Flow

- [Administrator creates department] → [System saves with active status] → [Department available for employee assignment]
- [Administrator deactivates department] → [System marks inactive] → [Historical employee records preserved, department removed from active selections]

---

## 5. User Journey Mapping

### Primary User Flow: View and Manage Departments

```
Administrator navigates to Organization Data → Departments →
Department List displayed → Administrator searches/browses →
Clicks Add New → Fills form → Saves → Department appears in list

OR

Administrator finds existing department → Clicks gear icon →
Selects Edit → Updates details → Saves changes

OR

Administrator finds department to deactivate → Clicks gear icon →
Selects Deactivate → Confirms → Department marked inactive
```

### Key Touch Points

1. Department list (first view — browse, search, export)
2. Add New form (create department)
3. Gear icon actions (edit / deactivate per row)
4. Pagination (navigate large datasets)

---

## 6. Business Rules & Constraints

### Business Rules

- BR-US-001-01: Department names must be unique within the organization
- BR-US-001-02: Departments are a flat list — no parent-child hierarchy (confirmed by Product Owner)
- BR-US-001-03: A department with active employee assignments can only be deactivated, not deleted
- BR-US-001-04: Deactivated departments remain visible in historical employee records

### Constraints

- Single sign-in page serves all authorized roles (access configured via US-004)
- No self-service department creation by employees or managers

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-03.
> Three screens extracted: Department List, Create A New Department, Edit A Department.

### Source Information

| Frame | Node ID | Size | Extraction Date |
|-------|---------|------|-----------------|
| Department List | 3059:1722 | 1920 × 1080 | 2026-03-03 |
| Create A New Department | 3059:1793 | 1920 × 1080 | 2026-03-03 |
| Edit A Department | 3060:22578 | 1920 × 1080 | 2026-03-03 |

---

### Screen 1: Department List

### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Logo]      │  Department List                                 │
│              │                                                  │
│  Richard R.  │  [Search 320px]         [Export] [+ Add New]    │
│  Super Admin │                                                  │
│  ──────────  │  ┌────────────────┬────────────────┬──────────┐  │
│  Users Mgmt  │  │ Department     │ No. Employees  │ Action   │  │
│  > Users     │  ├────────────────┼────────────────┼──────────┤  │
│  > Roles     │  │ Text           │ Text           │ ⚙        │  │
│  > Perms     │  │ Text           │ Text           │ ⚙        │  │
│  ──────────  │  │ Text           │ Text           │ ⚙        │  │
│  Org Data    │  │ Text           │ Text           │ ⚙        │  │
│  > Depts ◄   │  │ Text           │ Text           │ ⚙        │  │
│  > Positions │  │ Text           │ Text           │ ⚙        │  │
│  ──────────  │  └────────────────┴────────────────┴──────────┘  │
│  Menu Sect.  │                                                  │
│  > Menu 1   │  Rows per page [10▼]  Page 1 of 10  1…2 [3] 4…5 >│
│  > Menu 2   │                                                  │
│  > Menu 3   │                                                  │
└──────────────┴──────────────────────────────────────────────────┘
Sidebar: 200px │ Content: 1694px
```

### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Sidebar navigation | 3059:1723 | Primary navigation with collapsible sections | Designed |
| Logo (horizontal) | 3059:1725 | Brand identity in sidebar top | Designed |
| User profile card | 3059:1726 | Shows current user name + role | Designed |
| Nav section: Users Management | 3059:1733 | Collapsible nav group (Users, Roles, Permissions) | Designed |
| Nav section: Organization Data | 3059:2430 | Collapsible nav group (Departments, Positions) | Designed |
| Nav item: Departments | 3059:2434 | Active nav item — current page | Designed |
| Nav item: Positions | 3059:2437 | Links to Position Management (US-002) | Designed |
| Breadcrumb bar | 3059:1760 | Page location path | Placeholder text |
| Page title "Department List" | 3059:1769 | Page heading | Designed |
| Search input | 3059:1773 | Filter list by department name | Designed |
| Export button | 3059:1780 | Export department data | Designed |
| Add New button | 3059:1781 | Open create department form | Designed |
| Table: Department column | 3059:1783 | Lists department names | Placeholder data |
| Table: Employees column | 3059:1784 | Shows employee count per dept | Placeholder data |
| Table: Action column | 3059:1791 | Per-row gear icon for actions | Designed |
| Pagination | 3059:1792 | Navigate large department lists | Designed |

### Design Constraints

- **Sidebar width:** 200px fixed left
- **Content area width:** 1694px (remaining from 1920px)
- **Table column widths:** Department 775.5px | Employees 775.5px | Action 103px
- **Search input width:** 320px
- **Button layout:** Export + Add New right-aligned in action bar
- **Pagination:** Rows per page selector (default 10) + numbered page navigation

### Design Tokens Referenced

| Token | Value | Used For |
|-------|-------|----------|
| general/primary | #171717 | Add New button background |
| general/primary foreground | #fafafa | Add New button text |
| general/background | #ffffff | Page + table background |
| general/border | #e5e5e5 | Table borders, dividers |
| general/muted | #f5f5f5 | Table header background |
| general/muted foreground | #737373 | Placeholder text, secondary labels |
| general/foreground | #0a0a0a | Primary text |
| border/border-border | #e4e4e7 | Input borders |
| Font | Geist | All text |
| rounded-md | 6px | Button, input radius |

### Gaps Identified from Design — Department List

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| No status column | Cannot distinguish active vs inactive departments at a glance | Add Status column or filter to toggle active/inactive view |
| Gear icon actions undefined | Users don't know available actions per row | Design dropdown with Edit / Deactivate options |
| Breadcrumb is placeholder text | Navigation context unclear | Replace with real path: Organization Data / Departments |
| "Menu Section" in sidebar is placeholder | Future navigation items not yet defined | Confirm with Product Owner what modules will appear here |
| Empty state not designed | No guidance when no departments exist | Design empty state with "Add your first department" CTA |

---

### Screen 2: Create A New Department

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A New Department    [Cancel] [💾 Save]   │
│              │                                                  │
│              │         ┌──────────────────────────────┐        │
│              │         │  Department Information       │        │
│              │         │  ─────────────────────────   │        │
│              │         │  * Department name            │        │
│              │         │  [Enter department name     ] │        │
│              │         └──────────────────────────────┘        │
│              │              Card: 600px centered                │
└──────────────┴──────────────────────────────────────────────────┘
```

#### Component Inventory — Create Form

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Create A New Department" | 3059:1851 | Context heading | Designed |
| Cancel button | 3059:1853 | Discard and return to list | Designed |
| Save button (primary) | 3059:1854 | Submit form | Designed |
| Card container "Department Information" | 3059:1856 | Groups form fields | Designed |
| Section header "Department Information" | 3059:1858 | Section label | Designed |
| Field: Department name (Vertical Field) | 3059:1865 | Single input — required | Designed |

#### Design Constraints — Create Form

- **Form style:** Full page (not a modal dialog)
- **Card width:** 600px, horizontally centered in content area (1654px)
- **Card position:** Centered at offset 527px from content left edge
- **Field width:** 576px (within 600px card, 12px padding each side)
- **Button layout:** Cancel + Save right-aligned in page header row (x=1458, total 196px)
- **Required indicator:** Asterisk (*) before "Department name" label

#### ✅ Design Confirmed — Flat Structure, Name Field Only

> **The Create form contains only ONE field: Department Name.**
> No parent department selector — confirmed flat department structure.

**Resolved:** Product Owner confirmed departments are a flat list (no parent-child hierarchy). The single-field form is correct by design. No additional fields required.

---

### Screen 3: Edit A Department

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Edit A Department          [Cancel] [💾 Save]   │
│              │                                                  │
│              │         ┌──────────────────────────────┐        │
│              │         │  Department Information       │        │
│              │         │  ─────────────────────────   │        │
│              │         │  * Department name            │        │
│              │         │  [Enter department name     ] │        │
│              │         └──────────────────────────────┘        │
│              │              (pre-filled with current value)     │
└──────────────┴──────────────────────────────────────────────────┘
```

#### Component Inventory — Edit Form

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Edit A Department" | 3060:22636 | Context heading | Designed |
| Cancel button | 3060:22638 | Discard and return to list | Designed |
| Save button (primary) | 3060:22639 | Submit changes | Designed |
| Card container "Department Information" | 3060:22641 | Groups form fields | Designed |
| Section header "Department Information" | 3060:22643 | Section label | Designed |
| Field: Department name (Vertical Field) | 3060:22645 | Single input — pre-filled | Designed |

#### Key Observation — Create vs Edit Forms

Both forms are **identical in structure** — same layout, same card, same single field. The only difference is the page title. This confirms the design intent is a minimal, single-field form for both operations.

---

## 8. Success Criteria & Metrics

### Functional Success Criteria

- [ ] Administrator can create a department in under 60 seconds
- [ ] Department list displays correctly with search and pagination
- [ ] Deactivated departments do not appear in active selections (employee assignment)
- [ ] Export produces accurate department list data
- [ ] Only roles with permission can access create/edit/deactivate actions

### Business Metrics

- Department data consistency: 100% of employee records reference valid department entries
- Administrator task completion: Org structure setup before first employee is onboarded

---

## 9. Risk Assessment

### Identified Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Department deactivation breaks downstream references | Medium | High | Show warning listing affected employees before confirming deactivation |
| Duplicate department names entered | Low | Medium | System validates uniqueness on create and edit; shows error if duplicate found |

---

## 10. Assumptions & Notes

### Assumptions

1. Department names are unique organization-wide
2. Departments are a flat list — no parent-child hierarchy (confirmed by Product Owner)
3. Deactivating a department does not auto-reassign its employees
4. The "Number of employees" column counts direct employees only
5. "Menu Section" in the sidebar is a placeholder for future modules — not part of this story

### Open Questions

- [x] **Department hierarchy:** Confirmed flat list — no parent-child hierarchy. Single-field form (name only). — **Owner:** Product Owner — **Status:** Resolved ✅
- [x] **Create/Edit form style:** Full-page confirmed by Figma design (not a modal). — **Owner:** Design Team — **Status:** Resolved ✅
- [x] **Gear icon actions:** Edit + Delete confirmed. No deactivate — departments are hard-deleted (blocked if employees assigned). — **Owner:** Product Owner — **Status:** Resolved ✅
- [x] **Inactive departments in list:** No status — departments have no active/inactive state. List shows all departments. — **Owner:** Product Owner — **Status:** Resolved ✅
- [x] **Employee count scope:** Total employees assigned (flat list — all are direct). — **Owner:** Product Owner — **Status:** Resolved ✅
- [ ] **Export format:** What file format for export? CSV? Excel? — **Owner:** Product Owner — **Status:** Pending

---

**Document Version:** 1.0
**Last Updated:** 2026-03-03
**Author:** BA Team
**Reviewer:** Pending
