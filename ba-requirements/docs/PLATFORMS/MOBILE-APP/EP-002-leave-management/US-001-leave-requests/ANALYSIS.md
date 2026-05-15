---
document_type: ANALYSIS
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
status: draft
version: "1.0"
last_updated: "2026-04-16"
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
  - path: "../../../WEB-APP/EP-002-leave-management/US-001-leave-requests/ANALYSIS.md"
    relationship: reference
revision_history: []
input_sources:
  - type: figma
    file_id: "YEHeFgVZau7wmo9BZBVuZC"
    node_id: "3250:3755"
    frame_name: "Leave Requests"
    extraction_date: "2026-04-16"
add_on_sections:
  - "8. Design Context [ADD-ON]"
---

# Analysis: Leave Requests

**Epic:** EP-002 (Leave Management)
**Story:** US-001-leave-requests
**Platform:** MOBILE-APP
**Status:** Draft

---

## 1. Business Context

### Problem Statement

Employees need to manage their leave requests while away from their desks. The mobile experience should provide quick access to leave balances, pending requests, and the ability to submit new requests without requiring desktop access. Managers need mobile visibility into team leave for quick approvals.

### Stakeholders

- **Primary Users:** All employees submitting and tracking leave requests
- **Secondary Users:** Managers approving/rejecting team leave requests
- **Business Owner:** HR department leadership

### Business Goals

- Enable leave request submission from anywhere
- Provide instant visibility into leave balances
- Reduce time-to-approval through mobile notifications
- Improve employee self-service adoption

---

## 2. Scope Definition

### In Scope

- Leave Requests Dashboard (summary view with key metrics)
- Leave request list with basic filtering
- Create new leave request
- View leave request details
- Cancel pending leave request
- Leave balance display

### Out of Scope

- Manager approval actions (v1.0 — future enhancement)
- Leave type configuration (admin function, WEB-APP only)
- Export functionality (not applicable for mobile)
- Team calendar view (future enhancement)

### Dependencies

- **EP-001 US-001 (Authentication):** User must be signed in
- **WEB-APP Backend:** Leave management API, leave types, approval workflows
- **WEB-APP EP-002:** Leave business logic defined in web platform

---

## 3. Mobile-Specific Considerations

### Dashboard vs List Approach

Mobile users benefit from a **dashboard-first** approach:
1. **Dashboard** — Summary metrics, quick actions, recent activity
2. **List** — Full request history with filters (accessed from dashboard)

### Key Dashboard Elements

- Leave balance summary (by type: Annual, Sick, etc.)
- Pending requests count
- Recent requests (last 3-5)
- Quick action: "New Request" button
- Upcoming approved leave

### Mobile UX Patterns

- Pull-to-refresh for data updates
- Floating Action Button (FAB) for new request
- Swipe-to-cancel on pending requests (optional)
- Bottom sheet for filters
- Native date range picker

---

## 4. User Journey Mapping

### Primary Flow: View Dashboard

```
Home Screen → Tap "Leave" tab → Dashboard loads →
View balance summary → View pending count → View recent requests
```

### Alternative Flow: Create Request

```
Dashboard → Tap FAB "+" → Select leave type → 
Pick dates → Add reason (optional) → Submit → 
Success confirmation → Return to dashboard (updated)
```

### Alternative Flow: View History

```
Dashboard → Tap "View All" → Request list loads →
Apply filters (optional) → Tap request → View details
```

---

## 5. Success Criteria

### Functional Success Criteria

- [ ] Dashboard displays leave balance by type
- [ ] Dashboard shows pending request count
- [ ] User can view list of all their requests
- [ ] User can create new leave request
- [ ] User can view request details
- [ ] User can cancel pending requests
- [ ] Pull-to-refresh updates data

### Business Metrics

- Mobile leave request submission rate
- Time from request to approval (with mobile notifications)
- Mobile app adoption for leave management

---

## 6. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Stale balance data | Medium | Medium | Show last-updated timestamp, pull-to-refresh |
| Date picker usability | Low | Medium | Use native platform date picker |
| Offline submission failure | Medium | High | Clear error messaging, retry option |

---

## 7. Assumptions & Notes

### Assumptions

1. Leave types and quotas are configured in WEB-APP
2. Backend API is shared with WEB-APP
3. Push notifications will be implemented for status changes (separate story)
4. Manager approval actions are out of scope for v1.0

### Open Questions

- [ ] Should dashboard show team members' upcoming leave? (for managers)
- [ ] Half-day leave selection UX — segmented control or separate fields?
- [ ] Should cancelled requests appear in history or be hidden?

---

## 8. Design Context [ADD-ON]

### Source Information

| Property | Value |
|----------|-------|
| Figma File | YEHeFgVZau7wmo9BZBVuZC |
| Node ID | 3250:3755 |
| Frame Name | Leave Requests |
| Extraction Date | 2026-04-16 |
| Figma URL | [Leave Requests Dashboard](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3250-3755) |

### Layout Overview

```
┌─────────────────────────────────────────┐
│ [Avatar] [Dept] [Position]    Exnodes | │  <- Header (70px)
├─────────────────────────────────────────┤
│ [Cal] Leave Requests                    │  <- Page Title
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ [Palm] Need a Break?      [Image]   │ │  <- Hero Card (146px)
│ │ Submit your leave...                │ │
│ │ [Apply for Leave]                   │ │
│ └─────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│ Leave Balance                           │  <- Section Title
├─────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐         │
│ │Annual Leave │ │Sick Leave   │         │  <- Balance Cards
│ │     8       │ │     4       │         │     2x2 Grid
│ └─────────────┘ └─────────────┘         │
│ ┌─────────────┐ ┌─────────────┐         │
│ │Leaves This  │ │TBD          │         │
│ │Year   4     │ │     4       │         │
│ └─────────────┘ └─────────────┘         │
├─────────────────────────────────────────┤
│ [Upcoming Leave] [     History     ]    │  <- Tab Switcher (36px)
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ |Sick           [✓]    [Cal] 3 days│ │  <- Request Rows
│ │   F 24th Apr       T 26th Apr      │ │
│ └─────────────────────────────────────┘ │
│ ┌─────────────────────────────────────┐ │
│ │ |Vacation       [✓]    [Cal] 3 days│ │
│ │   F 24th Apr       T 26th Apr      │ │
│ └─────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│ [Home] [Calendar] [Docs] [Settings]     │  <- Bottom Nav (76px)
└─────────────────────────────────────────┘
```

### Component Inventory

| Component | Node ID | Purpose | Dimensions |
|-----------|---------|---------|------------|
| Header Frame | 3250:3756 | User info, company, notifications | 393x70 |
| Avatar | 3250:3758 | User profile photo | 30x30 |
| Department Badge | 3250:3759 | Shows user department | 52x24 |
| Position Badge | 3250:3760 | Shows user position | 72x24 |
| Notification Bell | 3250:3764 | BellSimpleRinging icon | 20x20 |
| Hero Card | 3250:3771 | Promotional CTA card | 369x146 |
| Apply for Leave Button | 3250:3778 | Primary action button | 144x36 |
| Balance Grid | 3250:3780 | 2x2 leave balance cards | 369x148 |
| Annual Leave Card | 3250:3782 | Annual leave remaining | 179.5x69 |
| Sick Leave Card | 3250:3786 | Sick leave remaining | 179.5x69 |
| Leaves This Year Card | 3250:3791 | Total leaves taken | 179.5x69 |
| TBD Card | 3250:3795 | Placeholder metric | 179.5x69 |
| Tab Buttons | 3250:3801 | Upcoming/History switcher | 369x36 |
| Upcoming Leave Tab | 3250:3802 | Active tab button | 182x36 |
| History Tab | 3250:3804 | Inactive tab button | 182x36 |
| Request Row | 3250:3951 | Leave request item | 369x89 |
| Bottom Navigation | 3250:4054 | App navigation bar | 369x76 |

### Design Tokens Used

| Token | Value | Usage |
|-------|-------|-------|
| Primary | #171717 | Buttons, text |
| Background | #ffffff | Page background |
| Foreground | #0a0a0a | Primary text |
| Muted Foreground | #737373 | Secondary text |
| Accent | #f5f5f5 | Card backgrounds |
| Exnodes Green | #27ae60 | Company name |
| Border | #e5e5e5 | Card borders |
| Green/600 | #16a34a | Success/approved |
| Blue/500 | #3b82f6 | Links/active state |
| Font Family | Geist | All text |
| Heading 4 | 20px Semibold | Card titles |
| Paragraph Small | 14px Regular | Body text |
| Paragraph Mini | 12px Regular | Labels |
| Rounded LG | 8px | Card corners |
| Rounded Full | 9999px | Avatar |

### Key Design Patterns

1. **Dashboard-First Approach**: Summary metrics and quick actions before detailed lists
2. **Hero Card CTA**: Prominent promotional card with primary action button
3. **2x2 Balance Grid**: Compact metric cards with decorative icons
4. **Tab-Based Filtering**: Simple toggle between Upcoming and History views
5. **Request Row Format**: Type indicator bar + status + duration + date range
6. **Bottom Navigation**: Persistent 4-icon navigation bar with active state

### Design Gaps / Questions

- 4th balance card labeled "TBD" — awaiting final metric definition
- Leave type color indicators not fully specified — need mapping for each type
- Exact behavior of "View All" for request lists not shown in design

---

**Document Version:** 1.0
**Last Updated:** 2026-04-16
**Author:** BA Team
