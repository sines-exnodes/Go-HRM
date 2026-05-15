---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
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

# Requirements: Skill Management

**Epic:** EP-008 (Organization Data)
**Story:** US-003-skill-management
**Status:** Draft

---

## 1. User Stories

| Story ID | As a... | I want to... | So that... | Priority |
|----------|---------|-------------|------------|----------|
| US-003-01 | User with org data permission | View all skills in a searchable list | I can browse and manage the organization's skill catalog | Critical |
| US-003-02 | User with org data permission | Create a new skill | New competencies can be tracked as the organization's needs evolve | Critical |
| US-003-03 | User with org data permission | Edit an existing skill's name | Skill definitions stay accurate as business terminology changes | High |
| US-003-04 | User with org data permission | Delete a skill that is no longer in use | The skill list stays clean and free of obsolete entries | High |

---

## 2. Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-US-003-01 | Skill list view with Icon, Skill Name, Description, Action columns | Critical | From Figma |
| FR-US-003-02 | Search by skill name | High | Pending design |
| FR-US-003-03 | Pagination with configurable rows per page | Medium | Pending design |
| FR-US-003-04 | Create skill (name field) | Critical | Pending design |
| FR-US-003-05 | Edit skill (update name) | High | Pending design |
| FR-US-003-06 | Delete skill (blocked if employees assigned) | High | Pattern from US-001 |
| FR-US-003-07 | Export skill list | Medium | Format TBD |

---

## 3. Business Rules

| ID | Rule |
|----|------|
| BR-US-003-01 | Skill names must be unique within the organization (case-insensitive) |
| BR-US-003-02 | Skills are a flat list — no categories or hierarchy |
| BR-US-003-03 | A skill with ≥1 employee assigned cannot be deleted |
| BR-US-003-04 | Access controlled by US-004 permission; action buttons hidden for unauthorized roles |

---

## 4. Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | Administrator can create a skill in under 1 minute | Single-field form |
| Data Integrity | Skills in use by employees cannot be deleted | Deletion blocked server-side |
| Access Control | Only authorized roles can manage skills | Permission enforcement via US-004 |

---

## 5. UI Specifications [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-24.
> Source: [Skill List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3089-1888) (node `3089:1888`)

### 5.1 Skill List — Page Structure

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page title | "Skill List" — Geist Semibold 24px, line-height 28.8px, letter-spacing -1, color `#0a0a0a` | `3089:1935` |
| Breadcrumb | Placeholder (expected: Organization Data / Skills) | `3089:1926` |
| Sidebar toggle | SidebarSimple icon, 16×16, top-left of content area | `3089:1927` |

### 5.2 Skill List — Action Bar

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Search input | Width: 320px, left-aligned. Placeholder: "Search by skill name.." (design has "role name" — copy-paste error). Border: `#e4e4e7`, bg: `#ffffff`, radius: 6px | `3089:1939` |
| Add New button | Right-aligned. Label: "+ Add New". Primary style: bg `#171717`, text `#fafafa`, radius: 6px. Visible only to users with org data permission. | `3089:1940` |

**Note:** No Export button is present on the Skill List (same as Role List, unlike Department/Position lists).

### 5.3 Skill List — Table

| Column | Width | Content | Figma Node |
|--------|-------|---------|------------|
| Icon | 85px | Skill icon/color indicator — placeholder colored squares in design | `3089:1942` |
| Skill Name | 181px | Skill name text — Geist Regular 14px, color `#09090b` | `3089:2343` |
| Description | 1295px | Skill description text — Geist Regular 14px, color `#09090b` | `3089:1943` |
| Action | 93px | Gear icon (⚙) — opens dropdown with Edit / Delete. Visible only to users with org data permission. | `3089:1944` |

**Table header:** Geist Medium 14px, bg `#f5f5f5`, color `#737373`
**Table border:** `#e5e5e5`
**Table background:** `#ffffff`

### 5.4 Skill List — Pagination

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Rows per page | Dropdown, default: 10. Options: 10, 25, 50. Geist Regular 12px. | `3089:1945` |
| Page indicator | "Page 1 of 10" — Geist Regular 12px, color `#737373` | `3089:1945` |
| Page navigation | Numbered buttons with prev/next arrows. Active page: outlined. | `3089:1945` |

### 5.5 Acceptance Criteria from Design

- **AC-UI-01:** Skill List page displays with page title "Skill List" in Geist Semibold 24px
- **AC-UI-02:** Search input is 320px wide, left-aligned, with placeholder "Search by skill name.."
- **AC-UI-03:** "+ Add New" button is right-aligned, primary style (dark bg, white text)
- **AC-UI-04:** Table has 4 columns: Icon (85px), Skill Name (181px), Description (1295px), Action (93px)
- **AC-UI-05:** Table header row uses muted background (`#f5f5f5`) with medium-weight text
- **AC-UI-06:** Each data row has a gear icon in the Action column
- **AC-UI-07:** Pagination displays below the table with rows-per-page selector and page navigation
- **AC-UI-08:** "+ Add New" button and gear icon are hidden for users without org data permission

### 5.6 Design Gaps Requiring Resolution

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 1 | Icon column mechanism unknown | Cannot implement icon selection | Product Owner: confirm if icons are user-selectable, fixed set, or color-coded |
| 2 | Search placeholder says "role name" | Incorrect label | Design Team: fix to "skill name" |
| 3 | No "No. of Employees" column | Cannot see skill usage at a glance | Product Owner: confirm if intentional |
| 4 | No Export button | Inconsistency with Department/Position lists | Product Owner: confirm if intentional |
| 5 | No Organization Data nav section | Skills not shown in sidebar navigation | Product Owner: confirm sidebar placement |
| ~~6~~ | ~~Create/Edit Skill screens not designed~~ | ~~Cannot define form layout~~ | ✅ Create Skill extracted 2026-03-24 — see Section 5.8. Edit Skill still pending. |
| 7 | Gear icon dropdown not designed | Actions not visually defined | Confirm: Edit + Delete |
| 8 | Empty state not designed | No guidance when no skills exist | Design Team: provide empty state |

### 5.8 Create A Skill — Page Structure

> Extracted from Figma on 2026-03-24.
> Source: [Create A Skill](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3089-1946) (node `3089:1946`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page title | "Create A Skill" — Geist Semibold 24px, line-height 28.8px, letter-spacing -1, color `#0a0a0a` | `3089:2004` |
| Breadcrumb | Placeholder (expected: Organization Data / Skills / Create A Skill) | `3089:1996` |
| Page layout | Full-page form (not modal), consistent with Department/Position create forms | `3089:2002` |

### 5.9 Create A Skill — Action Bar

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Cancel button | Left button in action bar. Secondary style. Returns to Skill List. | `3089:2006` |
| Save button | Right button in action bar. Primary style: bg `#171717`, text `#fafafa`, radius: 6px. Submits form. | `3089:2007` |

**Note:** Only 2 buttons (Cancel + Save). No "Save & Create Another" button.

### 5.10 Create A Skill — Skill Information Card

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600×302px, centered (x=527), bg `#ffffff`, radius: 8px, shadow-2xs | `3089:2009` |
| Section title | "Skill Information" — Geist Medium 14px, color `#0a0a0a` | `3089:2011` |
| Skill name field | Vertical Field instance. Label: "Skill name" (mandatory — marked with *). Placeholder: "Enter skill name". Width: 576px. | `3089:2013` |
| Description field | Vertical Field instance. Label: "Description" (optional — no asterisk). Placeholder: "Enter skill description". Width: 576px. | `3089:2383` |
| Skill Icon field | Label: "Skill Icon". Input File instance: "Choose File / No file chosen". Width: 576px. | `3089:2524` |

### 5.11 Acceptance Criteria from Design — Create Skill

- **AC-UI-09:** Create Skill page displays with page title "Create A Skill" in Geist Semibold 24px
- **AC-UI-10:** Cancel and Save buttons are right-aligned in the title action bar
- **AC-UI-11:** Skill Information card is 600px wide, centered in content area
- **AC-UI-12:** Skill name field is mandatory (marked with *) with placeholder "Enter skill name"
- **AC-UI-13:** Description field is optional (no asterisk) with placeholder "Enter skill description"
- **AC-UI-14:** Skill Icon field uses a file input ("Choose File / No file chosen")
- **AC-UI-15:** Create Skill page is accessible only to users with org data management permission

### 5.12 Design Gaps — Create Skill

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 9 | File upload constraints unknown | Cannot validate uploaded files | Product Owner: confirm accepted formats (PNG, JPG, SVG?), max file size, and max dimensions |
| 10 | Description max length unknown | Cannot validate field | Product Owner: confirm character limit for description |
| 11 | No validation error states | Cannot define error presentation | Assume: inline error below field (same as Department/Position) |
| 12 | No success confirmation | Cannot define post-save behavior | Assume: redirect to Skill List with updated entry |
| 13 | No "Save & Create Another" | Only one save path | Confirm: intentional (2 buttons) or add third button? |
| 14 | Skills not in sidebar nav | Cannot navigate to Skills | Add "Skills" under Organization Data in sidebar |

---

## 6. Next Steps

- [ ] Extract Figma design for Skill List screen
- [ ] Elaborate full requirements with acceptance criteria
- [ ] Write detail requirements (DRs) for each feature

---

**Document Version:** 0.1
**Last Updated:** 2026-03-24
**Author:** BA Agent
**Reviewer:** Pending
