---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-002
story_name: "Dashboard"
status: draft
version: "1.0"
last_updated: "2026-07-09"
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
  - path: "./details/DR-001-002-01-dashboard-page.md"
    relationship: child
revision_history: []
---

# Requirements: Dashboard

**Epic:** EP-001 (Foundation)
**Story:** US-002-dashboard
**Status:** Draft

---

## 1. User Stories

| Story ID | As a... | I want to... | So that... | Priority |
|----------|---------|-------------|------------|----------|
| US-002-01 | Authenticated user | See a permission-appropriate dashboard after sign in | I can understand what needs my attention | Critical |
| US-002-02 | HR/Admin user | See organization-wide workforce, attendance, and action summaries | I can monitor HR operations quickly | Critical |
| US-002-03 | Manager | See team-scoped attendance and pending actions | I can manage my team's HR work efficiently | High |
| US-002-04 | Employee | See my own attendance, leave, tickets, and announcements | I can track my personal HR items without searching multiple pages | High |

---

## 2. Functional Requirements

| ID | Requirement | Priority | Detail Requirement |
|----|-------------|----------|--------------------|
| FR-US-002-01 | Dashboard page loads after sign in and displays a common layout with permission-scoped widgets | Critical | DR-001-002-01 |
| FR-US-002-02 | Dashboard displays summary metric cards from permitted modules | Critical | DR-001-002-01 |
| FR-US-002-03 | Dashboard displays action queues for pending leave, request tickets, and announcements where permitted | High | DR-001-002-01 |
| FR-US-002-04 | Dashboard displays latest announcements targeted to the user or manageable by the user | High | DR-001-002-01 |
| FR-US-002-05 | Dashboard displays upcoming holiday/workday context | Medium | DR-001-002-01 |
| FR-US-002-06 | Dashboard provides quick links into existing module pages | Medium | DR-001-002-01 |
| FR-US-002-07 | Dashboard supports loading, empty, partial error, and full error states | High | DR-001-002-01 |

---

## 3. Business Rules

| ID | Rule |
|----|------|
| BR-US-002-01 | Dashboard is available to all authenticated active users after sign in. |
| BR-US-002-02 | Widget visibility and values are controlled by the source module permissions configured in US-004. |
| BR-US-002-03 | Dashboard widgets are read-only summaries; users perform record-changing actions only in source module screens. |
| BR-US-002-04 | Counts and lists respect the same user, team, and organization scope as the source module. |
| BR-US-002-05 | A widget hidden because of missing permission is not shown as an empty widget. |

---

## 4. Detail Requirement Coverage

| DR ID | Name | Status | File |
|-------|------|--------|------|
| DR-001-002-01 | Dashboard Page | Draft | [DR-001-002-01-dashboard-page.md](details/DR-001-002-01-dashboard-page.md) |

---

## 5. Acceptance Summary

- Dashboard opens after successful sign in.
- All users share the same dashboard layout.
- HR/Admin users see permitted organization-wide summaries.
- Managers see permitted team-scoped summaries.
- Employees see personal summaries only.
- Widgets and actions hidden by permission are not displayed.
- Quick links navigate to existing module pages.
- Dashboard does not expose hidden module data.

---

**Document Version:** 1.0
**Last Updated:** 2026-07-09
**Author:** BA Agent
