---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-002
story_name: "Dashboard"
detail_id: DR-001-002-01
detail_name: "Dashboard Page"
parent_requirement: FR-US-002-01
status: draft
version: "1.1"
created_date: "2026-07-09"
last_updated: "2026-07-09"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: grandparent
  - path: "../../US-004-role-permission-management/details/DR-001-004-01-role-list.md"
    relationship: permission-reference
  - path: "../../../../../superpowers/specs/2026-07-09-dashboard-page-design.md"
    relationship: approved-design-input
input_sources:
  - type: text
    description: "Approved brainstorming design spec: common operational action center with permission-hidden widgets"
    extraction_date: "2026-07-09"
  - type: text
    description: "Dashboard proposal generated from existing DR library and Project Knowledge Base"
    extraction_date: "2026-07-09"
---

# Detail Requirement: Dashboard Page

**Detail ID:** DR-001-002-01
**Parent Requirement:** FR-US-002-01
**Story:** US-002-dashboard
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.1

---

## 1. Use Case Description

As an **authenticated HRM user**, I want to **see one consistent dashboard after sign in that shows only the widgets and actions I am permitted to access** so that **I can quickly understand my current HRM priorities without searching through each module**.

**Purpose:** The Dashboard is the web application's operational landing page. It provides a common layout for all users, then uses the user's source-module permissions and data scope to decide which widgets, counts, rows, links, and actions are visible. The Dashboard is not a reporting suite, does not replace source module pages, and does not perform record-changing actions inline.

**Target Users:**

| User Group | Dashboard Need |
|------------|----------------|
| HR/Admin users | Organization-wide summaries and action queues where permissions allow |
| Managers | Team-scoped summaries and pending actions where source modules support team scope |
| Employees | Personal attendance, leave, request ticket, announcement, and calendar context |

**Key Functionality:**

- Common Dashboard layout for all users
- Permission-controlled widget and action visibility
- Summary counts plus short actionable lists inside visible widgets
- Navigation to source module pages for full details and record actions
- Loading, empty, restricted, partial error, and full error states
- No page-level or widget-level "last updated" timestamp requirement in v1

---

## 2. User Workflow

**Entry Point:** Successful sign in from WEB-APP US-001 Authentication. Unauthorized module redirects may also use the Dashboard as the fallback destination.

**Preconditions:**

- User is authenticated.
- User account is active.
- User has login permission.
- Source-module permissions are available through Role & Permission Management.

**Main Flow:**

1. User signs in successfully.
2. System routes the user to the Dashboard.
3. Dashboard loads the common layout and stable widget order.
4. Dashboard checks source-module permissions and user data scope.
5. Dashboard hides widgets, counts, rows, links, and actions the user cannot access.
6. Dashboard displays loading placeholders for visible widgets while summaries load.
7. Dashboard shows permitted widgets with summary counts and short actionable lists.
8. User reviews the visible widgets and selects a row, "View All" link, or quick action.
9. System navigates the user to the relevant source module page, where the source module applies its normal permissions, filters, confirmations, and status rules.

**Alternative Flows:**

| Flow | Description |
|------|-------------|
| HR/Admin scope | User sees organization-wide data only in widgets backed by permissions that allow organization-wide visibility. |
| Manager scope | User sees team-scoped attendance, leave, request ticket, or approval data where source modules support team scope. |
| Employee scope | User sees personal data and targeted announcements only. |
| Widget hidden by permission | Widget is omitted entirely and does not appear as a locked or empty card. |
| Summary visible, action restricted | Widget may show permitted read-only summary information while restricted buttons or action links are hidden. |
| No visible widgets | Dashboard still loads and shows a neutral empty state without exposing unavailable modules. |
| Partial widget error | One widget fails to load; that widget shows an inline retry state while other widgets remain visible. |
| Full dashboard error | No dashboard summaries can load; user sees a page-level retry state. |

**Exit Points:**

- User opens a source module list, detail, or workflow page.
- User starts an existing create flow, such as Create Leave Request or Create Request Ticket, when permitted.
- User navigates using the standard sidebar or header.

---

## 3. Field Definitions

This is a read-only dashboard page. It has no data-entry fields.

### Interaction Elements

| Element | Type | State/Condition | Trigger Action | Description |
|---------|------|-----------------|----------------|-------------|
| Summary Count | Read-only value | Visible when the widget is permitted | None | Shows the permitted count for a source-module summary |
| Actionable List Row | Row/link | Visible when row data exists and the user may open the target destination | Navigate to source record or filtered list | Shows a short item that needs attention or review |
| View All Link | Link | Visible when the user can access the source module list | Navigate to full source module list | Opens the complete source-module view |
| Quick Action Button | Button/link | Visible when user can access the target flow | Navigate to existing flow | Examples: Create Leave Request, Create Request Ticket |
| Retry Button | Button | Visible in widget or page error states | Reload failed summary area | Gives the user a recovery action |

### V1 Widget Set

| Widget | Count Example | Short Actionable List Example | Visibility Source |
|--------|---------------|-------------------------------|------------------|
| Attendance Overview | Present, late, absent, or my attendance today | Current-day exceptions or missing actions | Attendance permissions and data scope |
| Pending Approvals | Pending items requiring review | Pending leave approvals or other accessible approval work | Source-module approval permissions |
| Request Tickets | Open or assigned ticket count | Open, in-progress, or submitter follow-up tickets | Request Ticket permissions and ownership scope |
| Leave Summary | Balance, pending requests, or upcoming leave count | Upcoming leave or requests awaiting action | Leave permissions and ownership/team scope |
| Announcements | Latest active announcements count | Latest sent announcements visible to the user | Announcement visibility and audience targeting |
| Holidays & Workdays | Next holiday or current-period workday context | Upcoming holiday/workday items | Organization Settings or calendar visibility |
| Workforce Summary | Active employees or current-month workforce changes | New joiners or employee records needing review | User/Workforce permissions |

Each visible widget shows a compact list only. The Dashboard does not show the full source module dataset.

---

## 4. Data Display

### Page Structure

| Area | Content |
|------|---------|
| Header/Greeting | Dashboard title and user greeting |
| Summary Cards | Stable top area for permitted high-level counts |
| Action Widgets | Short lists of pending work, exceptions, or recent items |
| Awareness Widgets | Announcements, holidays, workdays, and workforce context where permitted |
| Quick Actions | Navigation shortcuts into existing module flows, hidden when target access is not allowed |

### Common Layout Rule

The Dashboard has the same widget order for every user. Permission rules decide what is shown. Hidden widgets collapse out of the page, and the remaining widgets keep their relative order. The Dashboard does not create separate HR/Admin, Manager, and Employee layouts in v1.

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Dashboard first opens or visible widgets refresh | Skeleton cards and rows that match final widget dimensions |
| Populated | One or more permitted widgets have values or rows | Common layout with permitted widgets, counts, and short lists |
| Empty Widget | User has permission but no relevant records | Widget-specific empty message such as "No pending approvals" or "No open tickets" |
| Permission Hidden | User lacks the source permission or data scope | Widget, count, row, link, or action is not rendered |
| All Widgets Hidden | User has login access but no visible source widgets | Neutral empty state without naming restricted modules |
| Partial Error | One widget fails while others load | Inline widget error with Retry; other widgets remain usable |
| Full Error | No dashboard summary can load | Page-level error with Retry |

### Fixed Time Windows for v1

| Context | Window |
|---------|--------|
| Attendance | Today |
| Workforce changes | Current month |
| Upcoming leave and holidays | Next 7 days |
| Announcements | Latest 5 sent announcements |
| Request tickets | Current active statuses |

### No Timestamp Requirement

The Dashboard does not need a page-level or widget-level "last updated" timestamp in v1. Users rely on the source module pages for authoritative detail review.

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Dashboard is shown after successful sign in for active authenticated users.
- **AC-02:** Dashboard uses one common widget layout and widget order for all users.
- **AC-03:** Dashboard does not require a separate Dashboard permission in v1; access follows successful authentication and login permission.
- **AC-04:** Widget visibility is controlled by source-module permissions from Role & Permission Management.
- **AC-05:** Users do not see widgets, counts, rows, links, or quick actions for modules or data scopes they cannot access.
- **AC-06:** A hidden widget is omitted entirely and is not shown as an empty or locked card.
- **AC-07:** If a user has summary visibility but lacks action permission, the widget can show read-only summary information while restricted actions remain hidden.
- **AC-08:** Visible widgets show summary counts plus short actionable lists where the source module has relevant items.
- **AC-09:** Actionable lists are limited summaries and provide navigation to the source module for full review.
- **AC-10:** HR/Admin users see organization-wide values only where their permissions allow organization-wide access.
- **AC-11:** Managers see team-scoped values only where the source module supports manager/team scope.
- **AC-12:** Employees see only their own personal data and announcements targeted to them.
- **AC-13:** Dashboard values match the same status, ownership, and scope definitions used by source module list/detail pages.
- **AC-14:** Dashboard widgets are read-only; create, edit, delete, approve, reject, send, close, reopen, and cancel actions are not performed inline.
- **AC-15:** Quick actions navigate to existing source-module flows and remain hidden when target access is not allowed.
- **AC-16:** Loading state uses skeleton cards/rows and prevents layout shift.
- **AC-17:** Empty widget states appear only when the user has permission but no relevant records.
- **AC-18:** If all widgets are hidden by permissions, the Dashboard loads a neutral empty state without exposing restricted module names.
- **AC-19:** A failed widget shows an inline error and retry action without hiding other successfully loaded widgets.
- **AC-20:** If no dashboard summaries can load, a page-level retry state is displayed.
- **AC-21:** Latest announcements show only sent, non-deleted announcements visible to the current user.
- **AC-22:** Dashboard does not display a page-level or widget-level "last updated" timestamp in v1.

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Common layout | Compare HR/Admin, Manager, and Employee dashboards | Same widget order is used; only permitted widgets differ | High |
| HR/Admin scope | User with organization-wide permissions | Organization-wide permitted counts and lists shown | High |
| Manager scope | User with team-scoped permissions | Team-scoped counts and pending actions shown where supported | High |
| Employee scope | Employee with self-service access | Personal attendance, leave, tickets, announcements, and calendar context shown | High |
| Widget hidden | User lacks Announcement permission | Announcement widget and announcement actions are not displayed | High |
| Read-only summary | User can view a source area but cannot act | Summary is visible; restricted action button is hidden | High |
| Empty permitted widget | Manager has no pending approvals | Widget shows a friendly empty message | Medium |
| No visible widgets | Authenticated user has no source-module visibility | Neutral empty state shown without restricted module names | High |
| Partial error | Ticket widget fails, attendance loads | Ticket widget shows Retry; attendance remains visible | Medium |
| Full error | Dashboard summaries cannot load | Page-level retry state appears | High |
| Quick action navigation | Employee clicks Create Request Ticket | User navigates to existing Create Request Ticket flow | High |
| Count consistency | Pending leave count displayed | Count matches Leave Requests list using same scope and status | High |
| No timestamp | Dashboard loads normally | No "last updated" timestamp is shown for page or widgets | Medium |

---

## 6. System Rules

**Business Logic & Behavior:**

- **SR-01:** Dashboard is the default landing page after sign in.
- **SR-02:** Dashboard access requires an active authenticated account with login permission.
- **SR-03:** Dashboard uses one common widget layout and order for all users in v1.
- **SR-04:** Widget visibility is derived from source-module permissions; no dashboard-specific widget management permission exists in v1.
- **SR-05:** Dashboard values must respect source-module data scope: own, team, or organization-wide.
- **SR-06:** Hidden widgets are omitted entirely and are not treated as empty states.
- **SR-07:** Restricted actions inside otherwise visible widgets are hidden when the user lacks the target action permission.
- **SR-08:** Dashboard does not mutate records. All record-changing actions navigate to source module pages.
- **SR-09:** Counts and rows exclude soft-deleted records.
- **SR-10:** Transactional status counts use existing source-module status definitions:
  - Leave: Pending, Approved, Rejected, Cancelled
  - Request Tickets: Open, In Progress, On Hold, Resolved, Closed, Cancelled
  - Announcements: Draft, Sent
- **SR-11:** Latest announcements display only Sent announcements that target the current user unless the user is viewing a permitted announcement management summary.
- **SR-12:** Fixed v1 time windows are today, current month, next 7 days, latest 5, and current active statuses.
- **SR-13:** Dashboard v1 has no page-level or widget-level "last updated" timestamp requirement.

**State Transitions:**

```
[Sign in success] -> [Dashboard loading] -> [Permission evaluation] -> [Visible widgets loading] -> [Dashboard populated]
[Visible widget loading] -> [Widget populated]
[Visible widget loading] -> [Widget empty]
[Visible widget loading] -> [Widget error]
[Widget error] -> [Retry] -> [Widget loading]
[No visible widgets] -> [Neutral empty state]
```

**Dependencies:**

| Dependency | Purpose |
|------------|---------|
| US-001 Authentication | Provides post-login routing and active account check |
| US-004 Role & Permission Management | Provides permissions for widget visibility and action access |
| US-005 User Management | Provides workforce and profile context |
| EP-002 Leave Management | Provides leave counts, balances, and pending approvals |
| EP-003 Request Tickets | Provides ticket status counts and action queues |
| EP-004 Attendance Management | Provides attendance summary data |
| EP-009 Organization Settings | Provides holiday/workday context |
| EP-010 Announcements | Provides latest announcement and announcement management data |

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Keep widget order stable so users can learn where information appears.
- **UX-02:** Hide restricted widgets completely rather than showing locked modules.
- **UX-03:** Use compact summary counts and short rows for quick scanning.
- **UX-04:** Keep actionable lists short; use "View All" for full source-module lists.
- **UX-05:** Use the same status names and badge meanings as source modules.
- **UX-06:** Use skeleton loaders that match final card and row dimensions.
- **UX-07:** Display partial errors inline so one failed widget does not block the whole dashboard.
- **UX-08:** Avoid advanced charts in v1 unless a chart directly supports an existing operational decision.

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>= 1280px) | Primary supported layout: multi-column dashboard grid |
| Tablet / below desktop | Out of scope for this story |
| Mobile | Out of scope; native mobile dashboards are documented separately |

**Accessibility Requirements:**

- [x] All widget links and buttons are keyboard navigable
- [x] Summary counts include accessible labels, not number-only announcements
- [x] Status meaning is conveyed by text, not color alone
- [x] Error states include actionable recovery text
- [x] Focus order follows the visible page layout
- [x] Hidden widgets are absent from keyboard and screen reader navigation

**Design References:**

- Approved design spec: `docs/superpowers/specs/2026-07-09-dashboard-page-design.md`
- No Dashboard Figma frame is available for this DR revision.
- Use existing WEB-APP design patterns: sidebar, content header, compact cards, short lists, muted table/list headers, and primary dark action buttons.

---

## 8. Additional Information

### Out of Scope

| Item | Reason |
|------|--------|
| User-configurable widget order | Future personalization; v1 uses a common fixed layout |
| Separate role-specific layouts | Replaced by common layout plus permission-controlled visibility |
| Page-level or widget-level "last updated" timestamp | Not required for v1 |
| Advanced analytics/charts | Belongs to future reporting module |
| Inline approvals/status changes | Source module pages own confirmation and status workflows |
| Payroll/recruitment/performance/training widgets | These modules are not documented yet |
| Mobile behavior | Covered by MOBILE-APP dashboard/widget DRs |
| Export/download from dashboard | Source module lists own export behavior |

### Open Questions

None.

### Related Features

| Feature | Relationship |
|---------|-------------|
| WEB-APP US-001 Authentication | Dashboard is the post-login landing page |
| WEB-APP US-004 Role & Permission Management | Controls widget visibility and action access |
| WEB-APP US-005 User Management | Provides workforce and profile summaries |
| WEB-APP EP-002 Leave Management | Provides leave summaries and pending actions |
| WEB-APP EP-003 Request Tickets | Provides ticket summaries and action queues |
| WEB-APP EP-004 Attendance Management | Provides attendance summaries |
| WEB-APP EP-009 Organization Settings | Provides holiday/workday context |
| WEB-APP EP-010 Announcements | Provides announcements summary |

### Notes

- This is the first WEB-APP dashboard DR and establishes the reusable web dashboard widget pattern.
- The approved v1 decision is a common dashboard layout for everyone, with permission-controlled widget visibility.
- Dashboard counts are operational summaries; source module list/detail pages remain authoritative for record review and action.
- The Dashboard should feel like a working action center, not a marketing page or executive reporting page.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-07-09 | Draft |
| Product Owner | - | - | Pending |
| UX Designer | - | - | Pending |
| Tech Lead | - | - | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-07-09 | BA Agent | Initial draft generated from existing DR library and Project Knowledge Base |
| 1.1 | 2026-07-09 | BA Agent | Revised from approved brainstorming spec: common layout for all users, permission-hidden widgets/actions, summary counts plus short actionable lists, and no Dashboard timestamp requirement |
