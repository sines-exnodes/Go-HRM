---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-001
story_name: "Department Management"
status: draft
version: "1.0"
last_updated: "2026-03-03"
add_on_sections: []
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
input_sources:
  - type: figma
    description: "Department List screen"
    node_id: "3059:1722"
    extraction_date: "2026-03-03"
---

# Requirements Specification: Department Management

**Epic:** EP-008 (Organization Data)
**Story:** US-001-department-management
**Status:** Draft
**Version:** 1.0
**Last Updated:** 2026-03-03

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [User Stories](#2-user-stories)
3. [Use Cases](#3-use-cases)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Business Rules](#6-business-rules)
7. [Business Data Elements](#7-business-data-elements)
8. [Information Display Requirements](#8-information-display-requirements)
9. [Dependencies](#9-dependencies)
10. [Acceptance Criteria](#10-acceptance-criteria)
11. [Open Questions](#11-open-questions)

---

## 1. Executive Summary

### Overview

This story delivers department management for Exnodes HRM — the ability for authorized administrators to create, view, edit, and delete organizational departments. Departments are foundational reference data used across all HR modules (employee profiles, leave management, payroll, performance reviews). This story establishes the single source of truth for the organization's department structure.

### Objectives

- Enable administrators to define and maintain the organization's department structure
- Provide search, browse, and export capabilities for the department list
- Enforce data integrity — departments with employees assigned cannot be deleted until all employees are reassigned

---

## 2. User Stories

### US-001-01: View Department List

```
As an administrator with department view permission,
I want to see all departments in a searchable, paginated list,
so that I can browse and manage the organization's department structure.
```

**Priority:** Critical
**Complexity:** Small

**Acceptance Criteria:**
- [ ] AC-001-01-01: Department list displays department name, employee count, and action controls
- [ ] AC-001-01-02: List supports search by department name
- [ ] AC-001-01-03: List is paginated with configurable rows per page (default 10)
- [ ] AC-001-01-04: List can be exported

---

### US-001-02: Create Department

```
As an administrator with department management permission,
I want to create a new department with a name,
so that the organization structure reflects our actual business units.
```

**Priority:** Critical
**Complexity:** Small

**Acceptance Criteria:**
- [ ] AC-001-02-01: "Add New" button opens a create department form
- [ ] AC-001-02-02: Form requires department name (mandatory)
- [ ] AC-001-02-03: System validates department name is unique
- [ ] AC-001-02-05: Successfully created department appears in the list

---

### US-001-03: Edit Department

```
As an administrator with department management permission,
I want to edit an existing department's name,
so that the department information stays up to date as the business changes.
```

**Priority:** High
**Complexity:** Small

**Acceptance Criteria:**
- [ ] AC-001-03-01: Edit option available from the department's action menu
- [ ] AC-001-03-02: Edit form pre-fills with current values
- [ ] AC-001-03-03: Name uniqueness validated on save
- [ ] AC-001-03-04: Changes reflected immediately in the list

---

### US-001-04: Delete Department

```
As an administrator with department management permission,
I want to delete a department that is no longer in use,
so that the department list stays accurate and free of obsolete entries.
```

**Priority:** High
**Complexity:** Small

**Acceptance Criteria:**
- [ ] AC-001-04-01: Delete option available from the department's action menu
- [ ] AC-001-04-02: System shows a confirmation dialog before deleting
- [ ] AC-001-04-03: Deletion is blocked if the department has one or more employees assigned
- [ ] AC-001-04-04: When deletion is blocked, system shows the number of assigned employees and instructs the administrator to reassign them first
- [ ] AC-001-04-05: Confirmed deletion removes the department from the list permanently

---

## 3. Use Cases

### UC-1: Create Department

**Primary Actor:** Administrator with department management permission
**Precondition:** Administrator is signed in and has create permission for departments

**Main Flow:**
1. Administrator navigates to Organization Data → Departments
2. System displays Department List with existing departments
3. Administrator clicks "Add New"
4. System displays create department form
5. Administrator enters department name
6. Administrator clicks Save
7. System validates name is unique
8. System creates department
9. System returns to Department List; new department appears

**Exception Flows:**
- **E1 - Duplicate Name:** At step 7, name already exists → system shows "Department name already exists"

---

### UC-2: Delete Department

**Primary Actor:** Administrator with department management permission
**Precondition:** Department exists in the department list

**Main Flow:**
1. Administrator locates department in the list
2. Administrator clicks gear icon → selects Delete
3. System checks whether any employees are assigned to this department
4. If no employees assigned: system shows confirmation dialog
5. Administrator confirms deletion
6. System permanently removes the department from the list

**Exception Flows:**
- **E1 - Employees Assigned:** At step 3, department has employees assigned → system blocks deletion and shows: "Cannot delete — [X] employees are assigned to this department. Reassign all employees before deleting." No changes made.
- **E2 - Administrator Cancels:** At step 5, administrator cancels → no changes made

---

## 4. Functional Requirements

### Category: List View

| Req ID | Requirement | Description | Priority |
|--------|-------------|-------------|----------|
| FR-US-001-01 | Department list | Display departments in a table with name, employee count, and action | Critical |
| FR-US-001-02 | Search | Filter list by department name | High |
| FR-US-001-03 | Pagination | Navigate list with configurable rows per page | Medium |
| FR-US-001-04 | Export | Export department list to file | Medium |

### Category: Department Management

| Req ID | Requirement | Description | Priority |
|--------|-------------|-------------|----------|
| FR-US-001-05 | Create department | Form to add new department with name field | Critical |
| FR-US-001-06 | Edit department | Update existing department name | High |
| FR-US-001-07 | Delete department | Hard delete with confirmation; blocked when employees are assigned | High |
| FR-US-001-08 | Name uniqueness validation | Prevent duplicate department names | Critical |
| FR-US-001-09 | Action menu | Gear icon per row with available actions (Edit, Delete) | High |

---

## 5. Non-Functional Requirements

### Access & Security
- **NFR-US-001-ACC01:** Department list is accessible to all roles with view permission
- **NFR-US-001-ACC02:** Create, Edit, and Delete actions require specific permission (configured via US-004)

### Usability
- **NFR-US-001-U01:** Administrator can create a new department in under 60 seconds
- **NFR-US-001-U02:** Search results update as the user types (not on submit)

### Data Integrity
- **NFR-US-001-DI01:** Departments with employees assigned cannot be deleted — the system blocks deletion and requires all employees to be reassigned first
- **NFR-US-001-DI02:** Deleted departments are permanently removed; ensure all employee assignments are cleared before deletion is permitted

---

## 6. Business Rules

### BR-US-001-01: Unique Department Names
**Description:** No two departments can share the same name.
**Enforcement:** System validates on create and edit; shows error if duplicate found.

### BR-US-001-02: Flat Department Structure
**Description:** Departments are a flat list — no parent-child hierarchy is supported. Each department is a standalone organizational unit.
**Enforcement:** Create and Edit forms contain only a department name field. No parent selector is available.

### BR-US-001-03: Deletion Blocked When Employees Assigned
**Description:** A department with one or more employees currently assigned cannot be deleted. All employees must be reassigned to another department before deletion is permitted.
**Enforcement:** System checks employee count before allowing delete. If count ≥ 1, deletion is blocked and a message is shown: "Cannot delete — [X] employees are assigned. Reassign all employees before deleting."

### BR-US-001-04: Access Controlled by Role Permission
**Description:** Department management actions (create, edit, delete) are only available to roles granted the relevant permission via US-004.
**Enforcement:** Action buttons and menu items hidden/disabled for roles without permission.

---

## 7. Business Data Elements

### Department Data

| Data Element | Business Purpose | Required? | Business Validation | Example Value |
|--------------|------------------|-----------|---------------------|---------------|
| Department Name | Identifies the department | Yes | Unique, non-empty | "Engineering" |
| Employee Count | Shows staffing level | System | Calculated from employee records | 12 |

---

## 8. Information Display Requirements

### List View Display

| Information | Display Context | Empty State | Format | Business Meaning |
|-------------|-----------------|-------------|--------|------------------|
| Department name | Table row, column 1 | "No departments found" | Text | The organizational unit name |
| Number of employees | Table row, column 2 | 0 | Number | Direct employees in this dept |
| Action menu | Table row, column 3 | — | Gear icon | Opens Edit/Delete options |
| Search field | Above table | Placeholder: "Search by department name" | Text input | Filter by name |
| Pagination | Below table | Hidden if ≤ page size | Row count + page numbers | Navigate large lists |

---

## 9. Dependencies

### Upstream Dependencies
- **US-004 (Role & Permission Management):** Determines which roles can view, create, edit, or delete departments

### Downstream Dependencies
- **US-002 (Position Management):** Positions may be optionally linked to departments
- **EP-002 (Employee Management):** Employee profiles reference department assignment

---

## 10. Acceptance Criteria

### Definition of Done

- [ ] Department list view displays correctly with search, pagination, and export
- [ ] Administrator can create a department with name (single-field form — flat structure)
- [ ] Administrator can edit an existing department (full-page form, pre-filled)
- [ ] Administrator can delete a department with confirmation; deletion blocked when employees are assigned
- [ ] Name uniqueness validation works correctly on create and edit
- [ ] Access control enforced — unauthorized roles cannot see management actions (Add New, gear icon)
- [ ] Delete blocking message shows correct employee count when deletion is attempted on a department with assigned employees

---

## 11. Open Questions

- [x] **Department hierarchy:** Confirmed flat list — no parent-child hierarchy. Single-field form (name only). — **Owner:** Product Owner — **Status:** Resolved ✅
- [x] **Create/Edit form style:** Full-page (not modal). — **Owner:** Design Team — **Status:** Resolved ✅
- [x] **Gear icon actions:** Edit + Delete confirmed. No status/deactivate — departments are hard-deleted (blocked if employees assigned). — **Owner:** Product Owner — **Status:** Resolved ✅
- [x] **Inactive departments in list:** No status — departments have no active/inactive state. List always shows all departments. — **Owner:** Product Owner — **Status:** Resolved ✅
- [x] **Employee count scope:** Total employees assigned to the department (flat list — all are direct). — **Owner:** Product Owner — **Status:** Resolved ✅
- [ ] **Export format:** CSV or Excel? — **Owner:** Product Owner — **Status:** Pending

---

## Appendices

### A. Related Documents

- [ANALYSIS.md](./ANALYSIS.md) — Business analysis with design context
- [FLOWCHART.md](./FLOWCHART.md) — Department management process flowcharts
- [TODO.yaml](./TODO.yaml) — BA task tracking

---

**Document Control:**
- **Version:** 1.0
- **Status:** Draft
- **Last Updated:** 2026-03-03
