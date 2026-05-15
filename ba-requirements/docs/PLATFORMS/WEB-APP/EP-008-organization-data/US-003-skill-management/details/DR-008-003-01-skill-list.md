---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
detail_id: DR-008-003-01
detail_name: "Skill List"
parent_requirement: FR-US-003-01
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
    description: "Skill List screen"
    node_id: "3089:1888"
    extraction_date: "2026-03-24"
---

# Detail Requirement: Skill List

**Detail ID:** DR-008-003-01
**Parent Requirement:** FR-US-003-01
**Story:** US-003-skill-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with organization data permission**, I want to **view all skills in a searchable, paginated list** so that **I can browse and manage the organization's skill catalog**.

**Purpose:** Provide a centralized, searchable catalog of skills (competencies, certifications, technical abilities) that supports employee competency tracking, training planning, and recruitment matching across the organization. Without a standardized skill list, employee competency data relies on free-text entries, leading to inconsistent data and unreliable reporting.

**Target Users:**
- Any user with organization data management permission — full access (browse, search, add, edit, delete)
- Any role granted view permission — can browse, search, and paginate (read-only)

**Key Functionality:**
- Searchable, paginated table of all skills with icon, name, and description
- Entry point for creating, editing, and deleting skills
- Access-controlled actions based on user permissions

---

## 2. User Workflow

**Entry Point:** Organization Data → Skills in sidebar navigation.

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has at least view permission for organization data (US-004)

**Main Flow:**
1. User navigates to Organization Data → Skills in sidebar
2. System loads skill data and displays the Skill List page
3. System shows table with all skills: Icon, Skill Name, Description, Action columns
4. Skills are sorted alphabetically by name by default
5. User browses the list, paginating as needed

**Search Flow:**
6. User types in the search box → list filters by skill name with debounce (300ms) → pagination resets to page 1
7. Search is case-insensitive and uses partial/contains matching
8. User clears the search box → full skill list restored → pagination resets to page 1

**Action Flows:**
9. User clicks **"+ Add New"** → system navigates to Create Skill page
10. User clicks **gear icon → Edit** on a row → system navigates to Edit Skill page for that skill
11. User clicks **gear icon → Delete** on a row → system triggers delete flow (blocked if employees assigned)

**Alternative Flows:**

- **Alt 1 — No skills exist:** System displays empty state: "No skills have been created yet" with prompt to add first skill
- **Alt 2 — Search returns no results:** System displays "No skills match your search" with "Clear search" link to reset
- **Alt 3 — Data load fails:** System displays error message with retry option
- **Alt 4 — Unauthorized user:** User with view-only permission sees the list but without "+ Add New" button or gear icons

**Exit Points:**
- **Navigate to Create Skill:** via "+ Add New" button
- **Navigate to Edit Skill:** via gear → Edit
- **Trigger Delete flow:** via gear → Delete
- **Navigate away:** via sidebar or breadcrumb navigation

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Search On | Trigger | Mandatory | Placeholder | Description |
|------------|------------|-----------|---------|-----------|-------------|-------------|
| Search | Text input | Skill name | On change with 300ms debounce | No | "Search by skill name.." | Filters table by skill name; case-insensitive; partial match supported |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| + Add New | Button (primary) | Right-aligned in action bar | Visible only to users with org data management permission | Navigate to Create Skill page | Create a new skill |
| Gear icon | Icon button (per row) | Action column | Visible only to users with org data management permission | Opens dropdown: Edit, Delete | Row-level actions |
| Rows per page | Dropdown | Left side of pagination bar | Always visible, default: 10 | Changes page size (10, 25, 50) | Controls how many rows per page |
| Page navigation | Button group | Right side of pagination bar | Always visible | Navigate between pages | Numbered page buttons with prev/next arrows |

**Note:** No Export button — intentionally omitted for Skill List (confirmed).

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Icon | Uploaded image | Default placeholder square | Small image thumbnail (within 85px column) | Custom icon uploaded by administrator — visual identifier for the skill |
| Skill Name | Text | N/A — always populated (mandatory field) | Geist Regular 14px, color `#09090b` | Name of the skill |
| Description | Text | "—" | Geist Regular 14px, single-line truncated with "..." if overflow; full text on hover tooltip | Brief description of what the skill entails |
| Gear icon | Icon button | N/A | ⚙ icon per row | Access Edit/Delete actions |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads with data | Table with all skills sorted alphabetically, search empty, pagination showing |
| Loading | Data being fetched | Skeleton rows to maintain layout stability |
| Empty (no skills) | No skills exist in the system | "No skills have been created yet" message with "+ Add New" prompt |
| No Results | Search returns zero matches | "No skills match your search" message with "Clear search" link to reset |
| Error | Data fetch fails | Error message with retry option |
| Description hover | User hovers over truncated description | Tooltip shows full description text |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Skill List page displays with page title "Skill List" in Geist Semibold 24px
- **AC-02:** Table has 4 columns: Icon (85px), Skill Name (181px), Description (1295px), Action (93px)
- **AC-03:** Table header row uses muted background (`#f5f5f5`) with medium-weight text

**Search:**
- **AC-04:** Search field filters the list by skill name as the user types, with a debounce delay (approximately 300ms)
- **AC-05:** Search is case-insensitive and matches partial names (contains match — "java" matches "JavaScript")
- **AC-06:** When search returns no results, the message "No skills match your search" is displayed with a clear search option
- **AC-07:** Clearing the search field immediately restores the full skill list and resets to page 1

**Pagination:**
- **AC-08:** Rows per page dropdown offers 10, 25, 50 with default 10
- **AC-09:** Page indicator shows "Page X of Y"
- **AC-10:** When a search is applied or cleared, pagination resets to page 1

**Data Display:**
- **AC-11:** Each row displays a custom icon image, skill name, and description
- **AC-12:** Description truncates to a single line with "..." if text overflows the column width
- **AC-13:** Hovering over a truncated description shows the full text in a tooltip
- **AC-14:** Skills are sorted alphabetically by name by default
- **AC-15:** Empty state shows "No skills have been created yet" when no skills exist in the system

**Actions:**
- **AC-16:** "+ Add New" button navigates to Create Skill page
- **AC-17:** Gear icon opens dropdown with Edit and Delete options
- **AC-18:** Gear icon Edit navigates to Edit Skill page
- **AC-19:** Gear icon Delete triggers the delete flow

**Access Control:**
- **AC-20:** "+ Add New" button and gear icon are hidden for users without org data management permission
- **AC-21:** Users with view-only permission can browse, search, and paginate

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Default load | Navigate to Skill List | All skills displayed alphabetically, page 1 | High |
| Search with match | Type "java" | Only skills containing "java" displayed | High |
| Search no match | Type "zzz" | "No skills match your search" with clear option | High |
| Search case-insensitive | Type "PYTHON" | Matches "Python", "python" | Medium |
| Clear search | Clear search box | Full list restored, page 1 | Medium |
| Pagination | Click page 2 | Page 2 results shown | Medium |
| Pagination reset on search | On page 3, type search | Returns to page 1 with filtered results | Medium |
| Description truncation | Skill with 200+ char description | Single-line truncated with "..." | Medium |
| Description tooltip | Hover over truncated description | Full text shown in tooltip | Medium |
| Empty state | No skills in system | Empty state message with Add New prompt | Medium |
| Unauthorized user | User without management permission | No Add New button, no gear icons, can browse | High |
| Custom icon display | Skill with uploaded icon | Icon image displayed in Icon column | Medium |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Search applies to the skill name field only — description and icon are not searched
- **SR-02:** Search matching is case-insensitive (e.g., "java" matches "Java", "JAVA", "JavaScript")
- **SR-03:** Search uses partial/contains matching (not exact match)
- **SR-04:** When a search is applied or cleared, pagination resets to page 1 to avoid showing an empty page
- **SR-05:** Skills are sorted alphabetically by name by default
- **SR-06:** No user-controlled column sorting (consistent with Department/Position lists)
- **SR-07:** Gear icon actions (Edit/Delete) are only rendered for users with org data management permission — not just hidden via CSS but excluded from the response for unauthorized users
- **SR-08:** The icon column displays a custom image uploaded by the administrator when creating/editing the skill
- **SR-09:** If no custom icon is uploaded, a default placeholder is shown

**State Transitions:**
```
[Skill List] → "+ Add New" click → [Create Skill page]
[Skill List] → Gear → Edit → [Edit Skill page]
[Skill List] → Gear → Delete → [Delete flow]
[Skill List] → Search input → [Filtered list, page 1]
[Skill List] → Clear search → [Full list, page 1]
[Skill List] → Page navigation → [Selected page]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — access control enforcement
- **Consumed by:** Create Skill (DR-008-003-02) — navigated to via "+ Add New"
- **Consumed by:** Edit Skill (DR-008-003-03) — navigated to via gear → Edit
- **Consumed by:** Delete Skill (DR-008-003-04) — triggered via gear → Delete

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Search input uses a debounce delay (~300ms) so the list does not filter on every single keystroke — reduces unnecessary load and improves perceived performance
- **UX-02:** Description column truncates with "..." on a single line — full text visible via tooltip on hover
- **UX-03:** Custom icon provides visual scanning aid — users can quickly identify skills at a glance
- **UX-04:** Pagination preserves the current rows-per-page selection during the session
- **UX-05:** Loading state uses skeleton rows to maintain layout stability while data is fetched

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full 4-column table, sidebar visible |
| Tablet (768-1024px) | Description column truncates more aggressively, sidebar collapsible |
| Mobile (<768px) | Icon and Description columns hidden, card layout alternative |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through search, Add New, gear icons, pagination
- [x] Screen reader compatible — table column headers, button labels, tooltip content
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements

**Design References:**
- Figma: [Skill List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3089-1888) (node `3089:1888`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Department List (DR-008-001-01), Position List (DR-008-002-01) — same list page pattern

---

## 8. Additional Information

### Out of Scope
- Create Skill — separate detail requirement (DR-008-003-02, planned)
- Edit Skill — separate detail requirement (DR-008-003-03, planned)
- Delete Skill — separate detail requirement (DR-008-003-04, planned)
- Employee-to-skill assignment (managed in Employee Management — EP-002)
- Skill proficiency levels (future enhancement)
- Skill categories / grouping management (future enhancement)
- Advanced filtering beyond name search (e.g., filter by category)
- Column sorting by user interaction (alphabetical default only)
- Export functionality (intentionally omitted)

### Open Questions
- None remaining for Skill List. All questions resolved during requirement writing session.

### Related Features
- **DR-008-003-02:** Create Skill (planned) — accessed via "+ Add New" button
- **DR-008-003-03:** Edit Skill (planned) — accessed via gear → Edit
- **DR-008-003-04:** Delete Skill (planned) — accessed via gear → Delete
- **DR-008-001-01:** Department List — same list page pattern (reference for consistency)
- **DR-008-002-01:** Position List — same list page pattern (reference for consistency)
- **US-001:** Authentication — user must be signed in
- **US-004:** Role & Permission Management — access control

### Notes
- The Figma design search placeholder says "Search by role name.." — this is a copy-paste error from the Role List screen. The correct placeholder is "Search by skill name..".
- The icon column uses **custom uploaded images**, not a fixed icon set. Administrators upload an icon when creating or editing a skill. A default placeholder is shown if no icon is uploaded.
- The Skill List does not include a "No. of Employees" column (unlike Department/Position lists) and does not include an Export button. Both are confirmed as intentional omissions.
- The sidebar navigation in the Figma design does not show an Organization Data section — the navigation placement for Skills needs to be confirmed with the Product Owner, but this is a navigation concern outside the scope of this detail requirement.

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
