---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
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
    description: "Skill List screen"
    node_id: "3089:1888"
    extraction_date: "2026-03-24"
  - type: figma
    description: "Create A Skill screen"
    node_id: "3089:1946"
    extraction_date: "2026-03-24"
---

# Analysis: Skill Management

**Epic:** EP-008 (Organization Data)
**Story:** US-003-skill-management
**Status:** Draft

---

## 1. Business Context

### Problem Statement

Organizations need a standardized catalog of skills (competencies, certifications, technical abilities) that can be referenced across HR modules. Without a centralized skill list, employee competency tracking relies on free-text entries, leading to inconsistent data and unreliable skill-based reporting.

### Stakeholders

- **Primary Users**: Administrators with organization data management permission
- **Secondary Users**: Managers (view skills for team context), HR (skill-based reporting)
- **Downstream Consumers**: Employee Management (EP-002), Performance Management (EP-006), Training & Development (EP-007)

### Business Goals

- Goal 1: Provide a centralized, administrator-managed skill catalog as reference data
- Goal 2: Ensure consistent skill definitions used across all HR modules

---

## 2. Scope Definition

### In Scope

- Skill list view (search, table, pagination)
- Create new skill
- Edit existing skill
- Delete skill (blocked if employees are assigned)
- Access controlled by role permissions (US-004)

### Out of Scope

- Employee-to-skill assignment (managed in Employee Management — EP-002)
- Skill proficiency levels (future enhancement)
- Skill categories / grouping (future enhancement)
- Skill import from external systems

### Dependencies

- **Internal**: US-001 (Authentication) — user must be signed in; US-004 (Role & Permission) — access control
- **Downstream**: EP-002 (Employee Management) — employee profiles reference skills

---

## 3. Requirements Analysis

### Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-003-01 | Skill list view | Critical | Table with Icon, Skill Name, Description, Action columns |
| FR-US-003-02 | Search skills | High | Filter by skill name |
| FR-US-003-03 | Pagination | Medium | Configurable rows per page |
| FR-US-003-04 | Create skill | Critical | Skill name field |
| FR-US-003-05 | Edit skill | High | Update skill name |
| FR-US-003-06 | Delete skill | High | Blocked if employees are assigned |
| FR-US-003-07 | Export skill list | Medium | Export format TBD |

---

## 4. Data Flow Analysis

### Data Entities

- **Skill**: Icon, Name (unique), Description

### Data Relationships

- Skill → Employees: One-to-many (one skill can be assigned to many employees)

### Data Flow

- [Administrator creates skill] → [Skill available for employee assignment]
- [Administrator edits skill name] → [Updated name reflected in employee records]
- [Administrator attempts delete] → [System checks employee count] → [If 0: Confirmation → delete | If ≥1: Blocked]

---

## 5. User Journey Mapping

### Primary User Flow

```
Administrator navigates to Organization Data → Skills →
Skill List displayed with Skill Name, No. of Employees, and Action columns →
Administrator searches/browses skills →
Clicks "+ Add New" to create a new skill →
OR clicks gear icon → Edit to modify skill name →
OR clicks gear icon → Delete to remove unused skill
```

---

## 6. Business Rules & Constraints

### Business Rules

- BR-US-003-01: Skill names must be unique within the organization (case-insensitive)
- BR-US-003-02: Skills are a flat list — no categories or hierarchy
- BR-US-003-03: A skill with ≥1 employee assigned cannot be deleted; system blocks deletion and shows employee count
- BR-US-003-04: Access controlled by US-004 permission; action buttons hidden for unauthorized roles

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-24.
> One screen extracted: Skill List.

### Source Information

| Attribute | Value |
|-----------|-------|
| **Figma File** | [Exnodes HRM](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC) |
| **Frame** | [Skill List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3089-1888) |
| **Node ID** | `3089:1888` |
| **Dimensions** | 1920 × 1080 |
| **Extraction Date** | 2026-03-24 |

> URLs are clickable — open directly in Figma

---

### Screen: Skill List

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Logo]      │  Skill List                                      │
│              │                                                  │
│  Richard R.  │  [Search 320px]                    [+ Add New]  │
│  Super Admin │                                                  │
│  ──────────  │  ┌──────┬──────────┬──────────────┬──────────┐  │
│  Users Mgmt  │  │ Icon │Skill Name│ Description  │ Action   │  │
│  > Users     │  ├──────┼──────────┼──────────────┼──────────┤  │
│  > Roles     │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│  > Perms     │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│  ──────────  │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│  Menu Sect.  │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│  > Menu 1   │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│  > Menu 2   │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│  > Menu 3   │  │ [■]  │ Text     │ Text         │ ⚙        │  │
│              │                                                  │
│              │  Rows per page [10▼]  Page 1 of 10  1…2 [3] 4…5 >│
└──────────────┴──────────────────────────────────────────────────┘
Sidebar: 200px │ Content: 1694px
```

#### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Skill List" | `3089:1935` | Page heading (Heading 3: Geist Semibold 24px) | Designed |
| Search input | `3089:1939` | Filter list by skill name — 320px width. Placeholder: "Search by role name.." (likely copy-paste error — should be "Search by skill name..") | Designed |
| Add New button | `3089:1940` | Create new skill — primary button, right-aligned | Designed |
| Table: Icon column | `3089:1942` | Skill icon/color indicator — 85px width | Placeholder |
| Table: Skill Name column | `3089:2343` | Skill name text — 181px width | Placeholder data |
| Table: Description column | `3089:1943` | Skill description text — 1295px width | Placeholder data |
| Table: Action column | `3089:1944` | Per-row gear icon for actions — 93px width | Designed |
| Pagination | `3089:1945` | Navigate large skill lists (rows per page + page numbers) | Designed |
| Sidebar navigation | `3089:1889` | Primary navigation — no Organization Data section visible | Designed |
| Breadcrumb bar | `3089:1926` | Page location path — placeholder text | Placeholder |

#### Design Constraints

- **Sidebar width:** 200px fixed left
- **Content area width:** 1694px
- **Table column widths:** Icon 85px | Skill Name 181px | Description 1295px | Action 93px
- **Search input width:** 320px, left-aligned
- **Button layout:** Only "+ Add New" right-aligned (no Export button)
- **Pagination:** Rows per page selector (default 10) + numbered page navigation
- **Page title:** "Skill List" — Geist Semibold 24px

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
| `heading 3` | Geist Semibold 24px / 28.8px LH / -1 LS | Page title |
| `paragraph small/medium` | Geist Medium 14px / 20px LH | Table header text |
| `paragraph small/regular` | Geist Regular 14px / 20px LH | Table body text |
| `paragraph mini/regular` | Geist Regular 12px / 16px LH | Pagination text |
| `rounded-md` | 6px | Button, input border-radius |
| Font family | Geist | All text |

#### Key Differences from Department/Position Lists

| Aspect | Department/Position | Skill List |
|--------|-------------------|------------|
| Table columns | Name, No. of Employees, Action | **Icon**, Name, **Description**, Action |
| Fields | Name only | Icon + Name + Description |
| No. of Employees column | Present | **Absent** |
| Export button | Present | **Absent** |

#### Gaps Identified from Design — Skill List

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Search placeholder says "Search by role name.." | Incorrect label — copy-paste from Role List | Fix to "Search by skill name.." |
| Icon column content is placeholder (colored squares) | Cannot determine how skill icons work — user-selectable? Fixed set? Color-coded? | Confirm with Product Owner: icon selection mechanism |
| Description column is placeholder ("Text") | Cannot determine max length, truncation rules | Confirm description character limit and truncation behavior |
| No "No. of Employees" column | Cannot see skill usage at a glance | Confirm with PO: intentional omission or should be added? |
| No Export button | Inconsistency with Department/Position lists | Confirm with PO: intentional or should be added? |
| No Organization Data nav section in sidebar | Skills not shown under Organization Data in navigation | Confirm where Skills appears in sidebar navigation |
| Gear icon actions undefined | Actions per skill not visually defined | Expected: Edit + Delete (confirm with PO) |
| Breadcrumb is placeholder | Navigation path unclear | Expected: Organization Data / Skills |
| ~~Create/Edit skill screens not designed~~ | ~~Cannot define form layout~~ | ✅ Create Skill extracted 2026-03-24 (node `3089:1946`). Edit Skill still pending. |
| Empty state not designed | No guidance when no skills exist | Request empty state design |

---

### Screen: Create A Skill

> Extracted from Figma on 2026-03-24.
> Source: [Create A Skill](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3089-1946) (node `3089:1946`)

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A Skill                 [Cancel]  [Save] │
│              │                                                  │
│  Users Mgmt  │  ┌──────────────────────────────────────────┐    │
│  > Users     │  │ Skill Information                        │    │
│  > Roles     │  │                                          │    │
│  > Perms     │  │  * Skill name                            │    │
│  ──────────  │  │  ┌──────────────────────────────────────┐│    │
│  Org Data    │  │  │ Enter skill name                     ││    │
│  > Depts     │  │  └──────────────────────────────────────┘│    │
│  > Positions │  │                                          │    │
│  ──────────  │  │  Description                             │    │
│  Menu Sect.  │  │  ┌──────────────────────────────────────┐│    │
│  > Menu 1    │  │  │ Enter skill description              ││    │
│  > Menu 2    │  │  └──────────────────────────────────────┘│    │
│  > Menu 3    │  │                                          │    │
│              │  │  Skill Icon                              │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ Choose File   No file chosen         ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  └──────────────────────────────────────────┘    │
└──────────────┴──────────────────────────────────────────────────┘
```

#### Component Inventory — Create Form

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Create A Skill" | `3089:2004` | Context heading (Geist Semibold 24px) | Designed |
| Cancel button | `3089:2006` | Discard and return to Skill List | Designed |
| Save button (primary) | `3089:2007` | Submit form | Designed |
| Card "Skill Information" | `3089:2009` | Groups form fields (600×302px, centered at x=527) | Designed |
| Section header "Skill Information" | `3089:2011` | Section label | Designed |
| Field: Skill name (Vertical Field) | `3089:2013` | Single text input — required (*) | Designed |
| Field: Description (Vertical Field) | `3089:2383` | Single text input — optional (no asterisk) | Designed |
| Field: Skill Icon (Input File) | `3089:2524` | File upload — "Choose File / No file chosen" | Designed |

#### Design Constraints — Create Form

- **Form style:** Full page (not a modal)
- **Card width:** 600px, horizontally centered in content area
- **Card position:** Centered at offset 527px from content left edge
- **Card height:** 302px
- **Field width:** 576px (within 600px card, 12px padding each side)
- **Button layout:** Cancel + Save right-aligned in page header row
- **Required indicator:** Asterisk (*) before "Skill name" label only
- **Page title:** "Create A Skill" (NOT "Create A New Skill" — differs from Department/Position naming)

#### Key Differences from Department/Position Create Forms

| Aspect | Department/Position | Skill |
|--------|-------------------|-------|
| Page title | "Create A New Department/Position" | **"Create A Skill"** (no "New") |
| Fields | 1 (Name only) | **3** (Name + Description + Icon upload) |
| File upload | None | **Skill Icon** (image file) |
| Card height | ~141px (1 field) | **302px** (3 fields) |
| "Save & Create Another" | Not present | **Not present** |

#### Sidebar Navigation — Create Skill Screen

| Nav Section | Items |
|-------------|-------|
| Users Management | Users, Roles, Permissions |
| **Organization Data** | **Departments, Positions** (Skills NOT listed) |
| Menu Section (placeholder) | Menu 1, Menu 2, Menu 3 |

> **Note:** The Organization Data section now appears in the sidebar (it was absent in the Skill List screen) but still does NOT include "Skills". This is a navigation gap — Skills should be listed under Organization Data.

#### Gaps Identified from Design — Create Skill

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Skills not in sidebar navigation | Users cannot navigate to Skills via sidebar | Add "Skills" under Organization Data, below Positions |
| No validation error states shown | Cannot define inline error presentation | Assume same pattern as Department/Position: inline error below field |
| No success confirmation shown | Cannot define post-save behavior | Assume: redirect to Skill List (same as Department/Position) |
| File upload constraints unknown | Cannot define accepted formats, max size | Confirm with PO: accepted image formats (PNG, JPG, SVG?), max file size |
| Description max length unknown | Cannot validate field input | Confirm character limit for description field |
| No "Save & Create Another" button | Only one save path | Confirm: intentional (2 buttons only) or should match Role pattern (3 buttons)? |

---

## 8. Success Criteria & Metrics (renumbered from 7)

### Functional Success Criteria

- [ ] Administrator can create, view, edit, and delete skills
- [ ] Skill list displays correctly with search and pagination
- [ ] Delete is blocked when employees are assigned to a skill
- [ ] Only authorized roles can access skill management actions

---

## 8. Assumptions & Notes

### Assumptions

1. Skills are a flat list — no categories or hierarchy (follows Department/Position pattern)
2. Hard delete for skills (not deactivate) — blocked if employees assigned
3. Skills have 3 fields: Icon, Name, Description — richer than Department/Position (name-only)
4. Skill list follows same UI pattern as Department and Position lists (sidebar + table + pagination)

### Open Questions

- [ ] **Export format:** CSV or Excel for skill list export? — **Owner:** Product Owner — **Status:** Pending

---

**Document Version:** 1.0
**Last Updated:** 2026-03-24
**Author:** BA Agent
**Reviewer:** Pending
