---
document_type: ANALYSIS
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
  - type: text
    description: "Dashboard proposal inferred from all existing DRs and Project Knowledge Base"
    extraction_date: "2026-07-09"
  - type: text
    description: "Approved brainstorming spec for common Dashboard layout and permission-hidden widgets"
    extraction_date: "2026-07-09"
---

# Analysis: Dashboard

**Epic:** EP-001 (Foundation)
**Story:** US-002-dashboard
**Status:** Draft

---

## 1. Business Context

### Problem Statement

After sign in, users need one trusted place to understand what needs attention across the HRM platform. Existing modules provide strong list and detail pages, but users must navigate module by module to see pending leave requests, attendance exceptions, open request tickets, latest announcements, and basic workforce context. The Dashboard solves this by presenting permission-appropriate summaries and action queues immediately after authentication.

### Stakeholders

- **Primary Users:** HR administrators, HR staff, managers, employees
- **Secondary Users:** Product Owner, leadership, BA team, support team
- **Business Owner:** HR department leadership

### Business Goals

- Give every authenticated user a useful landing page after sign in
- Surface urgent work without requiring users to open each module
- Respect the same permission and data-scope rules already defined in existing DRs
- Provide a scalable dashboard pattern for current and future HR modules

---

## 2. Scope Definition

### In Scope

- Common dashboard layout for all users with permission-controlled widget visibility
- Summary metric cards based on existing modules
- Pending action queues for leave requests, request tickets, and announcements
- Latest announcements widget
- Upcoming holidays and workday context
- Quick action links into existing module pages
- Loading, empty, error, and permission-hidden states

### Out of Scope

- Configurable dashboards or user-personalized widget layout
- Advanced analytics, charts, forecasting, or executive reporting
- Creating, editing, approving, or deleting records directly inside dashboard widgets
- Payroll, recruitment, performance, or training widgets until those modules are documented
- Mobile native dashboard behavior, which is covered separately by MOBILE-APP DRs

### Dependencies

- **US-001 Authentication:** Dashboard is shown after successful sign in
- **US-004 Role & Permission Management:** Controls dashboard widget visibility and data access
- **US-005 User Management:** Provides employee and profile summary data
- **EP-002 Leave Management:** Provides leave balances, pending approvals, and upcoming leave
- **EP-003 Request Tickets:** Provides open ticket and user action counts
- **EP-004 Attendance Management:** Provides attendance status and exception summaries
- **EP-009 Organization Settings:** Provides holiday and monthly workday context
- **EP-010 Announcements:** Provides latest sent announcements and draft/sent admin summaries

---

## 3. Requirements Analysis

### Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-US-002-01 | Dashboard landing page after sign in | Critical | First screen after successful authentication |
| FR-US-002-02 | Common widget composition with permission-controlled visibility | Critical | All users share the same layout; widgets and actions hide by permission and data scope |
| FR-US-002-03 | Workforce summary cards | High | Active employees, new joiners, departments, positions where permitted |
| FR-US-002-04 | Attendance today summary | High | Present, late, absent, on leave, own status depending on scope |
| FR-US-002-05 | Leave action queue | High | Pending approvals for managers/admins; own requests and balance for employees |
| FR-US-002-06 | Request ticket action queue | High | Open/in-progress tickets, own resolved tickets awaiting close/reopen |
| FR-US-002-07 | Latest announcements widget | Medium | Latest sent announcements for all users; draft/send cleanup cues for managers/admins |
| FR-US-002-08 | Upcoming holidays/workdays widget | Medium | Next holiday and current month workday context |
| FR-US-002-09 | Quick action links | Medium | Navigate to create/view flows without performing actions inline |
| FR-US-002-10 | Display states | High | Loading, empty, permission-hidden, partial error, full error |

### Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Usability | Users can identify top pending actions without opening module lists | First viewport contains action summary |
| Consistency | Widgets reuse existing module naming, statuses, and permission rules | No dashboard-only status meanings |
| Access Control | Widgets never reveal data outside the user's permission scope | Server and UI enforce same scope |
| Maintainability | New modules can add widgets later without changing current dashboard rules | Widget-based layout |

---

## 4. Data Flow Analysis

### Data Sources

| Source Module | Dashboard Data Used |
|---------------|---------------------|
| User Management | Active employee count, profile summary, department/position context |
| Attendance | Today's attendance status, late/absent counts, own attendance status |
| Leave Requests | Pending approvals, upcoming leave, own leave balance |
| Request Tickets | Open/action-needed tickets, own ticket status counts |
| Announcements | Latest sent announcements, draft/sent counts for management users |
| Holidays/Workdays | Next holiday, current month workday count |

### Data Scope

- Users with management or all-record access see organization-wide or team-wide summary data as allowed by the source module.
- Employees see only their own personal data and announcements targeted to them.
- If a user lacks the permission for a source module, the dashboard hides that widget rather than showing zero values.

---

## 5. User Journey Mapping

### Primary Flow

1. User signs in successfully.
2. System routes the user to the Dashboard.
3. Dashboard loads the common layout and permitted summary widgets.
4. User reviews metric cards and action queues.
5. User clicks a widget row, View All link, or quick action.
6. System navigates to the relevant module page with normal module permissions applied.

### Key Touch Points

- First page after authentication
- Metric cards for quick scanning
- Short actionable lists for pending work
- Latest announcements widget
- Quick action links to commonly used flows

---

## 6. Business Rules & Constraints

### Business Rules

- BR-US-002-01: Dashboard is available to all authenticated active users with login permission.
- BR-US-002-02: Dashboard widgets use underlying module permissions; there is no independent dashboard management permission in v1.
- BR-US-002-03: Widgets hidden by permission do not count as empty states.
- BR-US-002-04: Dashboard widgets are read-only summaries; record-changing actions occur in the source module screens.
- BR-US-002-05: All counts must respect ownership, team, or organization scope from the source module.
- BR-US-002-06: Dashboard uses fixed time windows in v1: today, current month, next 7 days, and latest 5 records.

### Constraints

- Web admin experience is desktop-first.
- No Figma design is currently available for this story.
- The dashboard must not introduce technical implementation details or new source-module behavior.

---

## 7. Success Criteria & Metrics

### Functional Success Criteria

- [ ] Dashboard opens after successful sign in for all active users
- [ ] Users see only widgets and data they are authorized to view
- [ ] HR/Admin users see organization-wide summaries where permissions allow
- [ ] Managers see team-scoped pending actions where permissions allow
- [ ] Employees see personal summaries and targeted announcements
- [ ] Each widget has loading, empty, and error behavior
- [ ] Quick links route users to existing module pages

### Business Metrics

- Fewer clicks to find pending leave/request ticket actions
- Reduced missed announcements after sign in
- Faster HR/Admin review of daily attendance exceptions

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Dashboard exposes data outside user's scope | Medium | High | Reuse source module permission and data-scope rules |
| Dashboard becomes too crowded | Medium | Medium | Limit v1 to current modules and fixed widget order |
| Counts differ from source module lists | Medium | High | Widgets navigate to source module and use same filtering definitions |
| Future modules change dashboard needs | High | Low | Widget-based layout allows later additions |

---

## 9. Assumptions & Notes

### Assumptions

1. Dashboard is the post-login landing page and fallback route for unauthorized module access.
2. HR/Admin means users with organization-wide permissions through US-004, not hardcoded role names.
3. Manager data scope is team-based where the source module supports manager scope.
4. Dashboard v1 uses no customizable widgets.
5. Dashboard v1 does not require a page-level or widget-level "last updated" timestamp.

### Open Questions

No unresolved v1 questions. Future configurable/reorderable widgets remain a future consideration outside Dashboard v1.

---

**Document Version:** 1.0
**Last Updated:** 2026-07-09
**Author:** BA Agent
**Reviewer:** Pending
