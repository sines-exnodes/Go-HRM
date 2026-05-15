---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-002
story_name: "Position Management"
status: draft
version: "1.0"
last_updated: "2026-03-05"
add_on_sections: []
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
input_sources: []
---

# Analysis: Position Management

**Epic:** EP-008 (Organization Data)
**Story:** US-002-position-management
**Status:** 🔴 Draft

---

## Table of Contents

1. [Business Context](#1-business-context)
2. [Scope Definition](#2-scope-definition)
3. [Requirements Analysis](#3-requirements-analysis)
4. [Data Flow Analysis](#4-data-flow-analysis)
5. [User Journey Mapping](#5-user-journey-mapping)
6. [Business Rules & Constraints](#6-business-rules--constraints)
7. [Success Criteria & Metrics](#7-success-criteria--metrics)
8. [Risk Assessment](#8-risk-assessment)
9. [Implementation Approach](#9-implementation-approach)
10. [Timeline & Milestones](#10-timeline--milestones)
11. [Quality & Testing Strategy](#11-quality--testing-strategy)
12. [Documentation & Handoff](#12-documentation--handoff)
13. [Assumptions & Notes](#13-assumptions--notes)

---

## 1. Business Context

### Problem Statement

Organizations need a structured way to define and maintain the positions that exist within the business. Without a centralized position registry, employee records become inconsistent — different modules may reference the same position by different names or not at all. This story delivers the tools for administrators to create and manage the organization's position catalog as authoritative reference data available across all HR modules.

### Stakeholders

- **Primary Users**: Administrators with position management permission (configurable via US-004)
- **Secondary Users**: All other HR modules that reference position data (e.g., employee profiles)
- **Business Owner**: HR department leadership

### Business Goals

- Goal 1: Provide a single source of truth for all positions used across HR modules
- Goal 2: Allow administrators to maintain position records as the organization evolves (add, edit, delete) without disrupting employee records

---

## 2. Scope Definition

### In Scope

- Position list view (search, table, pagination, export)
- Create new position
- Edit existing position name
- Delete position (hard delete, blocked if employees are assigned)
- Access controlled by role permission (US-004)

### Out of Scope

- Optional department association (deferred to future iteration)
- Position hierarchies or parent-child structure
- Position-to-role mapping
- Bulk import of positions from external files
- Automatic position merges or restructuring

### Dependencies

- **Internal**: US-004 (Role & Permission Management) — access control for create/edit/delete
- **Downstream**: EP-002 Employee Management — employees assigned to positions; employee count drives delete blocking

---

## 3. Requirements Analysis

### Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-002-01 | Position list view | Critical | Table with Name, Employee Count, Action columns |
| FR-US-002-02 | Search | High | Filter by position name (partial, case-insensitive, debounced) |
| FR-US-002-03 | Pagination | Medium | Configurable rows per page, default 10 |
| FR-US-002-04 | Export | Medium | Export currently filtered list |
| FR-US-002-05 | Create position | Critical | Single-field form — position name only |
| FR-US-002-06 | Edit position | High | Update position name; pre-filled form |
| FR-US-002-07 | Delete position | High | Hard delete, blocked if employees assigned |
| FR-US-002-08 | Name uniqueness validation | Critical | No duplicate position names (case-insensitive) |
| FR-US-002-09 | Action menu | High | Gear icon per row with Edit / Delete options |

### Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | Administrators can add a position in under 60 seconds | Single-form workflow |
| Data Integrity | Positions in use by employees cannot be deleted | Deletion blocked server-side |
| Access Control | Only authorized roles can create/edit/delete | Enforced via US-004 |

---

## 4. Data Flow Analysis

### Data Entities

- **Position**: Name (unique), employee count (computed)
- **Employee** (downstream): References position as assignment

### Data Relationships

- Position → Employees: One-to-many (one position can have many employees)

### Data Flow

- [Administrator creates position] → [System saves] → [Position immediately available for employee assignment]
- [Administrator edits position] → [System updates record] → [Updated name reflected across all HR modules]
- [Administrator attempts delete] → [System checks employee count] → [If 0: Confirmation → hard delete | If ≥1: Blocked]

---

## 5. User Journey Mapping

### Primary User Flow: View and Manage Positions

```
Administrator navigates to Organization Data → Positions →
Position List displayed → Administrator searches/browses →
Clicks Add New → Fills form → Saves → Position appears in list

OR

Administrator finds existing position → Clicks gear icon →
Selects Edit → Updates name → Saves changes

OR

Administrator finds position to delete → Clicks gear icon →
Selects Delete → System checks employee count →
If 0 employees: Confirmation dialog → Confirms → Position permanently deleted
If ≥1 employees: Blocked dialog → User must reassign employees first
```

### Key Touch Points

1. Position list (first view — browse, search, export)
2. Add New form (create position — full page)
3. Edit form (edit position — full page, pre-filled)
4. Gear icon actions (Edit / Delete per row)
5. Delete dialogs (Confirmation or Blocked)

---

## 6. Business Rules & Constraints

### Business Rules

- BR-US-002-01: Position names must be unique within the organization (case-insensitive)
- BR-US-002-02: Positions are a flat list — no parent-child hierarchy; single-field form (name only)
- BR-US-002-03: A position with ≥1 employee assigned cannot be deleted; system blocks deletion and shows employee count
- BR-US-002-04: Access to create/edit/delete is controlled by role permission (US-004); buttons hidden for unauthorized roles

### Constraints

- Position name is the only field — no department association in this release
- Hard delete only — no soft delete, deactivation, or status flag
- No bulk operations

---

## 7. Success Criteria & Metrics

### Functional Success Criteria

- [ ] Administrator can create a position in under 60 seconds
- [ ] Position list displays correctly with search and pagination
- [ ] Export produces accurate position list data
- [ ] Delete is blocked when employees are assigned to a position
- [ ] Only roles with permission can access create/edit/delete actions

### Business Metrics

- Position data consistency: 100% of employee records reference valid position entries
- Administrator task completion: Position catalog set up before first employee is onboarded

---

## 8. Risk Assessment

### Identified Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Position deletion breaks downstream employee references | Low | High | System blocks deletion if employees assigned; user must reassign first |
| Duplicate position names entered | Low | Medium | System validates uniqueness (case-insensitive) on create and edit |
| Race condition on delete (employee assigned between check and execution) | Low | Medium | Employee count re-checked at execution time (server-side guard) |

---

## 9. Implementation Approach

### Recommended Approach

Position Management follows the same pattern as Department Management (US-001) — a flat reference data CRUD module with a list view and single-field forms. The implementation is a direct parallel of US-001 with "position" in place of "department".

### Key Implementation Notes

- Identical UI pattern: list + full-page create/edit form + modal dialogs for delete
- Hard delete with pre-delete employee count check (not soft delete)
- No Figma design available yet — pending design delivery for position screens
- All permission checks via US-004; no hardcoded role names

---

## 10. Timeline & Milestones

| Milestone | Description | Status |
|-----------|-------------|--------|
| Requirements complete | REQUIREMENTS.md written and reviewed | Pending review |
| Design available | Figma screens for Position List, Create, Edit delivered | Pending |
| Detail requirements | DR-008-002-01 through DR-008-002-04 written | In Progress |
| Stakeholder sign-off | REQUIREMENTS.md approved by Product Owner | Pending |
| UAT scenarios defined | Test plan created from acceptance criteria | Pending |

---

## 11. Quality & Testing Strategy

### Testing Approach

- Unit validation tests: name empty, name too long, name duplicate (case-insensitive, exact, self-exclusion on edit)
- Integration: position delete blocking when employees assigned; position available in employee profile selectors after creation
- Permission tests: management actions hidden for view-only roles; page accessible only with view permission

### Key Test Scenarios

| Scenario | Type | Priority |
|----------|------|----------|
| Create position with valid name | Functional | High |
| Duplicate name blocked (case-insensitive) | Validation | High |
| Edit: saving unchanged name succeeds | Validation | High |
| Delete blocked when employees assigned | Business Rule | High |
| Delete succeeds when 0 employees | Functional | High |
| Race condition protection on delete | Edge Case | Medium |
| Management actions hidden for view-only users | Permission | High |

---

## 12. Documentation & Handoff

### Documents Produced

| Document | Status | Notes |
|----------|--------|-------|
| REQUIREMENTS.md | Draft | Full FR/BR/NFR content complete |
| ANALYSIS.md | Draft | This document — stub |
| FLOWCHART.md | Draft | Stub — pending full flowchart |
| TODO.yaml | Draft | Task tracking |
| DR-008-002-01 (Position List) | Draft | Full detail requirement |
| DR-008-002-02 (Create Position) | Draft | Full detail requirement |
| DR-008-002-03 (Edit Position) | Draft | Full detail requirement |
| DR-008-002-04 (Delete Position) | Draft | Full detail requirement |

### Handoff Notes

- No Figma reference available at time of writing — design team must provide Position List, Create Position, and Edit Position screens
- Design expected to be identical in structure to Department Management screens (US-001) with "Position" in place of "Department"
- Delete flow does not have a separate page — handled via modal dialogs on the Position List

---

## 13. Assumptions & Notes

### Assumptions

1. Position names are unique organization-wide
2. Positions are a flat list — no parent-child hierarchy (consistent with Department Management approach)
3. Hard delete only; no active/inactive status for positions
4. Administrators will manually reassign employees before deleting a position
5. The "Number of employees" column counts employees directly assigned to the position

### Open Questions

- [ ] **Export format:** What file format should the export produce — CSV or Excel (.xlsx)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Design delivery:** When will Figma screens for Position List, Create Position, and Edit Position be available? — **Owner:** Design Team — **Status:** Pending

---

**Document Version:** 1.0
**Last Updated:** 2026-03-05
**Author:** BA Agent
**Reviewer:** Pending
