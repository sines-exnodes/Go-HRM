---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-002
story_name: "Position Management"
status: draft
version: "1.0"
created_date: 2026-03-05
last_updated: 2026-03-05
add_on_sections: []
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../US-001-department-management/REQUIREMENTS.md"
    relationship: related
---

# Requirements: Position Management

**Story ID:** US-002
**Epic:** EP-008 (Organization Data)
**Platform:** Exnodes HRM (Web App)
**Status:** Draft
**Version:** 1.0

---

## 1. Executive Summary

Position Management allows authorized administrators to create, edit, and delete organizational positions. Positions are foundational reference data used across HR modules — primarily for employee assignments. This story delivers a complete CRUD interface for position data with permission-based access control.

**Scope:** Position list view with search, pagination, and export; create/edit/delete position forms; name uniqueness enforcement; hard delete with blocking rule when employees are assigned.

**Out of Scope:** Optional department association (deferred to future iteration); position hierarchies; position-to-role mapping.

---

## 2. User Stories

### US-002-01: View Position List

> As an **administrator**, I want to **see all positions in a searchable, paginated list**, so I can **browse and manage the organization's position structure**.

| Attribute | Value |
|-----------|-------|
| **Priority** | Critical |
| **Complexity** | Small |
| **Status** | Not Started |

**Acceptance Criteria:**
- [ ] AC-001-01-01: Position list displays all positions in a table with Name, Employee Count, and Action columns
- [ ] AC-001-01-02: User can search positions by name (partial, case-insensitive)
- [ ] AC-001-01-03: List paginates at 10 rows per page (configurable)
- [ ] AC-001-01-04: Export downloads currently filtered list

---

### US-002-02: Create Position

> As an **administrator**, I want to **create a new position**, so that **it becomes available for selection across all HR modules (e.g., employee profiles)**.

| Attribute | Value |
|-----------|-------|
| **Priority** | Critical |
| **Complexity** | Small |
| **Status** | Not Started |

**Acceptance Criteria:**
- [ ] AC-001-02-01: "Add New" button opens create position form
- [ ] AC-001-02-02: Form requires position name (mandatory)
- [ ] AC-001-02-03: System validates position name is unique (case-insensitive)
- [ ] AC-001-02-04: System validates position name max 100 characters
- [ ] AC-001-02-05: Successfully created position appears in list immediately

---

### US-002-03: Edit Position

> As an **administrator**, I want to **edit an existing position's name**, so that **the position information stays accurate as the business changes**.

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Complexity** | Small |
| **Status** | Not Started |

**Acceptance Criteria:**
- [ ] AC-001-03-01: Edit option available from position's action menu
- [ ] AC-001-03-02: Edit form pre-filled with current position name
- [ ] AC-001-03-03: Saving with unchanged name succeeds without duplicate error
- [ ] AC-001-03-04: Updated name immediately reflected across all HR modules

---

### US-002-04: Delete Position

> As an **administrator**, I want to **delete a position that is no longer in use**, so that **the position list stays accurate and free of obsolete entries**.

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Complexity** | Small |
| **Status** | Not Started |

**Acceptance Criteria:**
- [ ] AC-001-04-01: Delete option available from position's action menu
- [ ] AC-001-04-02: System shows confirmation dialog before deleting
- [ ] AC-001-04-03: Deletion blocked if position has one or more employees assigned
- [ ] AC-001-04-04: When blocked, system shows employee count and instructs to reassign first
- [ ] AC-001-04-05: Confirmed deletion removes position from list permanently

---

## 3. Use Cases

### UC-1: View and Browse Position List

**Actor:** Any user with position view permission
**Precondition:** User is signed in with view permission
**Flow:**
1. User navigates to Organization Data > Positions in sidebar
2. System displays Position List page with table
3. User browses, searches, paginates, or exports

**Postcondition:** User has located the desired position(s)

---

### UC-2: Delete Position

**Actor:** User with position management permission
**Precondition:** User is signed in with management permission; target position exists
**Main Flow:**
1. User clicks gear icon on a position row and selects Delete
2. System checks employee count for this position in real time
3. If count = 0: System shows Confirmation Dialog → User confirms → Position permanently deleted
4. System refreshes Position List

**Exception Flow:**
- If count ≥ 1: System shows Blocked Dialog with employee count; no deletion performed
- User must reassign all employees before attempting to delete again

---

## 4. Functional Requirements

### 4.1 List View

| FR ID | Requirement | Description | Priority |
|-------|-------------|-------------|----------|
| **FR-US-002-01** | Position list | Display positions in table with name, employee count, and action | Critical |
| **FR-US-002-02** | Search | Filter list by position name (partial, case-insensitive, debounced) | High |
| **FR-US-002-03** | Pagination | Navigate list with configurable rows per page (default 10) | Medium |
| **FR-US-002-04** | Export | Export currently filtered position list to file | Medium |

### 4.2 Position Management

| FR ID | Requirement | Description | Priority |
|-------|-------------|-------------|----------|
| **FR-US-002-05** | Create position | Full-page form to add new position with position name field | Critical |
| **FR-US-002-06** | Edit position | Full-page form to update existing position name (pre-filled) | High |
| **FR-US-002-07** | Delete position | Hard delete with confirmation dialog; blocked when employees assigned | High |
| **FR-US-002-08** | Name uniqueness validation | Prevent duplicate position names (case-insensitive) | Critical |
| **FR-US-002-09** | Action menu | Gear icon per row with available actions (Edit, Delete) | High |

---

## 5. Non-Functional Requirements

| NFR ID | Category | Requirement |
|--------|----------|-------------|
| **NFR-US-002-ACC01** | Access & Security | Position list accessible to all roles with view permission |
| **NFR-US-002-ACC02** | Access & Security | Create, Edit, Delete actions require management permission (configured via US-004) |
| **NFR-US-002-U01** | Usability | Administrator can create a new position in under 60 seconds |
| **NFR-US-002-U02** | Usability | Search results update as user types (not on submit), with debounce |
| **NFR-US-002-DI01** | Data Integrity | Positions with employees assigned cannot be deleted; system blocks deletion |
| **NFR-US-002-DI02** | Data Integrity | Deleted positions are permanently removed; all employee assignments must be cleared first |

---

## 6. Business Rules

| BR ID | Business Rule | Description |
|-------|---------------|-------------|
| **BR-US-002-01** | Unique Position Names | No two positions can share the same name; validated case-insensitively on create and edit |
| **BR-US-002-02** | Flat Position Structure | Positions are a flat list — no parent-child hierarchy; create/edit forms have position name field only |
| **BR-US-002-03** | Deletion Blocked When Employees Assigned | A position with ≥1 employee assigned cannot be deleted; system shows: "Cannot delete — [X] employees are assigned to this position. Reassign all employees before deleting." |
| **BR-US-002-04** | Access Controlled by Role Permission | Position management (create, edit, delete) is only available to authorized roles configured via US-004; buttons hidden for unauthorized roles |

---

## 7. Data Requirements

### Business Data Elements

| Field | Type | Constraints | Usage |
|-------|------|-------------|-------|
| Position Name | Text | Required; max 100 chars; unique (case-insensitive) | Identifies the position; used in employee profile selectors |
| Employee Count | Number (computed) | Non-negative integer; calculated live | Displayed in Position List; determines delete eligibility |

### Information Display

The Position List page displays:
- Position Name (sortable/searchable)
- No. of Employees (live count from EP-002)
- Action column with gear icon (Edit/Delete — management permission only)

---

## 8. Interface Requirements

### Navigation
- Position List accessible from sidebar: Organization Data > Positions
- Create Position: full-page form (not modal)
- Edit Position: full-page form (not modal), pre-filled
- Delete Position: modal dialog overlaid on Position List

### Action Triggers
- Create: "+ Add New" button on Position List header
- Edit: Gear icon → "Edit" on Position List row
- Delete: Gear icon → "Delete" on Position List row

---

## 9. Constraints & Limitations

- Position name is the only field — no department association in this release
- Hard delete only — no soft delete, deactivation, or status flag
- No bulk operations (create/edit/delete multiple at once)
- No import from file (CSV or otherwise)
- Desktop-first UI — no mobile/tablet responsive adaptation in this release

---

## 10. Assumptions & Dependencies

### Assumptions
- Positions are independent reference data (no required department link)
- Position names are globally unique across all positions
- Administrators will manually reassign employees before deleting a position

### Dependencies

| Dependency | Type | Description |
|------------|------|-------------|
| **EP-001: Foundation** | Upstream | Authentication must be operational |
| **US-004: Role & Permission Management** | Upstream | Controls which roles can view/manage positions |
| **EP-002: Employee Management** | Consumer | Employee records reference position; employee count drives delete blocking |
| **US-001: Department Management** | Sibling | Departments must be available before extending to cross-reference (future) |

---

## 11. Acceptance Criteria

### Definition of Done — All must be met before story is considered complete:

- [ ] Position list displays all positions with correct employee counts
- [ ] Search filters in real time with debounce; case-insensitive partial match
- [ ] Export downloads currently filtered list
- [ ] Create position form validates: required, unique (case-insensitive), max 100 chars
- [ ] Edit position form pre-fills with current name; uniqueness check excludes current position
- [ ] Delete flow: employee count checked before dialog; blocked if ≥1 employee
- [ ] Confirmed delete permanently removes position; list refreshes
- [ ] All management actions hidden/disabled for view-only roles
- [ ] Permissions controlled via US-004 — no hardcoded role names

---

## 12. Sign-Off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |
