---
document_type: DETAIL_REQUIREMENT
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-01
detail_name: "Leave Requests Dashboard"
parent_requirement: FR-US-001-01
status: draft
version: "1.0"
created_date: 2026-04-16
last_updated: 2026-04-16
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../../../../WEB-APP/EP-002-leave-management/US-001-leave-requests/details/DR-002-001-01-leave-requests-list.md"
    relationship: reference
input_sources:
  - type: figma
    file_id: "YEHeFgVZau7wmo9BZBVuZC"
    node_id: "3250:3755"
    frame_name: "Leave Requests"
    extraction_date: "2026-04-16"
---

# Detail Requirement: Leave Requests Dashboard

**Detail ID:** DR-002-001-01
**Parent Requirement:** FR-US-001-01
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Platform:** MOBILE-APP
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee**, I want to **view a dashboard summarizing my leave balance, pending requests, and recent leave activity**, so that **I can quickly understand my leave status and take action without navigating to multiple screens**.

**Purpose:** The Leave Requests Dashboard is the primary entry point for leave management on mobile. It provides a glanceable summary of leave balances, a promotional call-to-action for submitting new requests, and quick access to upcoming and historical leave requests. The dashboard is optimized for mobile interaction patterns — prioritizing at-a-glance information and one-tap actions.

**Target Users:**
- All authenticated employees with access to the mobile app
- Users checking leave balance before planning time off
- Users tracking status of submitted requests

**Key Functionality:**
- Leave balance summary cards (Annual Leave, Sick Leave, Leaves This Year)
- Hero promotional card with "Apply for Leave" quick action button
- Tab-based view switching between "Upcoming Leave" and "History"
- Recent requests list showing leave type, status, duration, and date range
- Pull-to-refresh to update dashboard data
- Bottom navigation for app-wide navigation

---

## 2. User Workflow

**Entry Point:** Bottom navigation bar > Calendar icon (Leave tab) OR Home screen > Leave section

**Preconditions:**
- User is signed in to the mobile app (EP-001 US-001)
- User has employee role with leave access

**Main Flow:**
1. User taps the Calendar icon in the bottom navigation bar
2. System loads the Leave Requests Dashboard
3. System displays the header with user avatar, department/position badges, company name, and notification bell
4. System displays the page title "Leave Requests" with calendar icon
5. System displays the hero card "Need a Break?" with "Apply for Leave" button
6. System displays the Leave Balance section with 4 balance cards in a 2x2 grid
7. System displays the tab switcher with "Upcoming Leave" (default active) and "History" tabs
8. System displays recent leave requests under the active tab
9. User browses dashboard content or takes an action

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Apply for Leave** | User taps "Apply for Leave" button on hero card > Navigate to Create Leave Request screen |
| **View Upcoming** | User taps "Upcoming Leave" tab (default) > Shows approved/pending future leaves |
| **View History** | User taps "History" tab > Shows past leave requests (all statuses) |
| **Tap Request** | User taps a request in the list > Navigate to Leave Request Details screen |
| **Pull to Refresh** | User pulls down on the screen > Dashboard data refreshes > Loading indicator shown > Data updated |
| **Tap Notification Bell** | User taps notification bell icon > Navigate to Notifications screen |
| **Tap Home** | User taps Home icon in bottom nav > Navigate to Home Dashboard |
| **Tap Settings** | User taps Settings icon in bottom nav > Navigate to Settings screen |

**Exit Points:**
- **Apply for Leave** > Navigate to Create Leave Request screen
- **Tap Request** > Navigate to Leave Request Details screen
- **Bottom Navigation** > Navigate to respective screen (Home, Documents, Settings)
- **Notification Bell** > Navigate to Notifications screen

---

## 3. Field Definitions

### Display-Only Fields (No Input)

This is a dashboard screen with no user input fields. All content is read-only display.

### Interaction Elements

| Element | Type | Location | Trigger Action | Description |
|---------|------|----------|----------------|-------------|
| User Avatar | Image (30x30) | Header left | None (display only) | Shows user's profile photo |
| Department Badge | Badge (52x24) | Header left | None (display only) | Shows user's department (e.g., "Front End") |
| Position Badge | Badge (72x24) | Header left | None (display only) | Shows user's position (e.g., "Developer") |
| Company Name | Text link | Header right | Navigate to company info (optional) | "Exnodes" in green (#27ae60) |
| Notification Bell | Icon button (20x20) | Header right | Navigate to Notifications | Bell icon with optional badge count |
| Apply for Leave | Primary button (144x36) | Hero card | Navigate to Create Leave Request | Dark button (#171717) with calendar icon |
| Upcoming Leave Tab | Tab button (182x36) | Tab section | Switch to upcoming view | Shows future approved/pending requests |
| History Tab | Tab button (182x36) | Tab section | Switch to history view | Shows all past requests |
| Request Row | Touchable row (369x89) | Request list | Navigate to Request Details | Shows leave type, status, duration, dates |
| Home Icon | Nav icon (26x26) | Bottom nav | Navigate to Home | House icon, yellow highlight when active |
| Calendar Icon | Nav icon (26x26) | Bottom nav | Current screen (active) | Calendar icon, indicates Leave tab |
| Documents Icon | Nav icon (26x26) | Bottom nav | Navigate to Documents | Scroll/document icon |
| Settings Icon | Nav icon (26x26) | Bottom nav | Navigate to Settings | Gear icon |

---

## 4. Data Display

### Header Section

| Data | Format | Source | Description |
|------|--------|--------|-------------|
| User Avatar | 30x30 circular image | User profile | Placeholder if no photo |
| Department | Badge text | User profile | e.g., "Front End" |
| Position | Badge text | User profile | e.g., "Developer" |
| Company Name | Text | System config | "Exnodes" |

### Hero Card ("Need a Break?")

| Element | Format | Description |
|---------|--------|-------------|
| Icon | Palm tree emoji/icon (24x24) | Vacation theme indicator |
| Title | "Need a Break?" (Heading 4, 20px semibold) | Promotional headline |
| Subtitle | "Submit your leave request in just a few presses" (14px regular, muted) | Supporting text |
| Illustration | Decorative image (193x96) | Person relaxing in hammock illustration |
| CTA Button | "Apply for Leave" (144x36, primary dark) | Quick action to create request |

### Leave Balance Section

| Card | Label | Value Format | Icon | Description |
|------|-------|--------------|------|-------------|
| Annual Leave Remaining | "Annual Leave Remaining" | Number (e.g., "8") | Calendar icon | Days available for annual leave |
| Sick Leave Remaining | "Sick Leave Remaining" | Number (e.g., "4") | Syringe icon | Days available for sick leave |
| Leaves This Year | "Leaves This Year" | Number (e.g., "4") | Calendar star icon | Total leaves taken this year |
| TBD | "TBD" | Number (e.g., "4") | Calendar star icon | Placeholder for additional metric |

**Balance Card Layout:**
- 2x2 grid (179.5px width per card, 69px height)
- 10px gap between cards
- Label: 12px regular, muted foreground (#737373)
- Value: 24px semibold, primary foreground (#0a0a0a)
- Icon: Decorative, positioned bottom-right, partially clipped

### Request List Items

| Data | Format | Position | Description |
|------|--------|----------|-------------|
| Leave Type Indicator | 5px colored bar + type text | Left | Color bar + "Sick", "Vacation", etc. |
| Status Icon | Checkmark icon (18x18) | Center-right | Green checkmark for approved |
| Duration | Icon + "X days" | Right | Calendar icon + day count |
| From Date | "F" prefix + "24th Apr 2026" | Bottom left | "F" = From |
| To Date | "T" prefix + "26th Apr 2026" | Bottom right | "T" = To |

**Leave Type Colors (from design):**
- Sick: Color indicator bar (specific color TBD from design system)
- Vacation/Annual: Color indicator bar (specific color TBD from design system)

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Dashboard first opens or refreshing | Skeleton cards for balance, skeleton rows for requests |
| Populated | User has leave balance and requests | Full dashboard with all sections populated |
| Empty Balance | New user with no leave history | Balance cards show "0" or "--" for each type |
| Empty Requests (Upcoming) | No upcoming approved/pending leaves | Empty state message: "No upcoming leaves" |
| Empty Requests (History) | No past leave requests | Empty state message: "No leave history yet" with CTA to apply |
| Pull-to-Refresh | User pulls down | Refresh spinner at top, data reloads |
| Error | API failure | Toast error message, retry option via pull-to-refresh |

### Page Layout (ASCII Diagram)

```
┌─────────────────────────────────────────┐
│ [Avatar] [Dept] [Position]    Exnodes | │  <- Header (70px)
├─────────────────────────────────────────┤
│ [Cal] Leave Requests                    │  <- Page Title (20px)
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ [Palm] Need a Break?      [Image]   │ │  <- Hero Card (146px)
│ │ Submit your leave...                │ │
│ │ [Apply for Leave]                   │ │
│ └─────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│ Leave Balance                           │  <- Section Title (16px)
├─────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐         │
│ │Annual Leave │ │Sick Leave   │         │  <- Balance Cards
│ │     8       │ │     4       │         │     Row 1 (69px)
│ └─────────────┘ └─────────────┘         │
│ ┌─────────────┐ ┌─────────────┐         │
│ │Leaves This  │ │TBD          │         │  <- Balance Cards
│ │Year   4     │ │     4       │         │     Row 2 (69px)
│ └─────────────┘ └─────────────┘         │
├─────────────────────────────────────────┤
│ [Upcoming Leave] [     History     ]    │  <- Tab Switcher (36px)
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ |Sick           [✓]    [Cal] 3 days│ │  <- Request Row 1 (89px)
│ │   F 24th Apr       T 26th Apr      │ │
│ └─────────────────────────────────────┘ │
│ ┌─────────────────────────────────────┐ │
│ │ |Vacation       [✓]    [Cal] 3 days│ │  <- Request Row 2 (89px)
│ │   F 24th Apr       T 26th Apr      │ │
│ └─────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│ [Home] [Calendar] [Docs] [Settings]     │  <- Bottom Nav (76px)
└─────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

**Header:**
- **AC-01:** Header displays user avatar (30x30), department badge, and position badge on the left side
- **AC-02:** Header displays company name "Exnodes" in green (#27ae60) and notification bell icon on the right side
- **AC-03:** Notification bell icon is tappable and navigates to Notifications screen

**Page Title:**
- **AC-04:** Page title "Leave Requests" is displayed with calendar icon prefix
- **AC-05:** Title uses Geist font, 16px or appropriate heading size

**Hero Card:**
- **AC-06:** Hero card displays palm tree icon, "Need a Break?" heading, and promotional subtitle
- **AC-07:** Hero card displays decorative illustration (person in hammock) on the right side
- **AC-08:** "Apply for Leave" button is prominently displayed with calendar icon
- **AC-09:** Tapping "Apply for Leave" navigates to Create Leave Request screen

**Leave Balance:**
- **AC-10:** "Leave Balance" section title is displayed below the hero card
- **AC-11:** Four balance cards are displayed in a 2x2 grid layout
- **AC-12:** Each balance card shows label (muted text), value (large number), and decorative icon
- **AC-13:** Balance values are fetched from the backend and reflect current user's leave quota
- **AC-14:** Annual Leave Remaining and Sick Leave Remaining display accurate values from user's quota

**Tab Switcher:**
- **AC-15:** Two tabs are displayed: "Upcoming Leave" (default active) and "History"
- **AC-16:** Active tab is visually distinguished (filled background vs outlined)
- **AC-17:** Tapping a tab switches the content below to show relevant requests
- **AC-18:** Tab state persists during the session until user navigates away

**Request List:**
- **AC-19:** Each request row displays: leave type (with color indicator), status icon, duration ("X days"), and date range (From/To)
- **AC-20:** Leave type is shown with a colored vertical bar indicator on the left
- **AC-21:** Approved requests show a green checkmark icon
- **AC-22:** Duration shows calendar icon + "X days" format
- **AC-23:** Date range shows "F [date]" and "T [date]" format (e.g., "F 24th Apr 2026")
- **AC-24:** Tapping a request row navigates to Leave Request Details screen
- **AC-25:** "Upcoming Leave" tab shows only future Approved and Pending requests
- **AC-26:** "History" tab shows all past requests regardless of status

**Bottom Navigation:**
- **AC-27:** Bottom navigation bar displays 4 icons: Home, Calendar (Leave), Documents, Settings
- **AC-28:** Calendar icon is highlighted (active state) when on Leave Requests Dashboard
- **AC-29:** Tapping other icons navigates to respective screens

**Data Refresh:**
- **AC-30:** Pull-to-refresh gesture triggers dashboard data reload
- **AC-31:** Loading indicator is shown during refresh
- **AC-32:** Data updates reflect latest values from backend after refresh

**Empty States:**
- **AC-33:** When no upcoming leaves exist, "No upcoming leaves" message is displayed
- **AC-34:** When no history exists, "No leave history yet" message is displayed with option to apply

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Dashboard load | Navigate to Leave tab | Dashboard displays with all sections populated | High |
| Apply for Leave tap | Tap "Apply for Leave" button | Navigate to Create Leave Request screen | High |
| View upcoming | Default tab state | Shows Approved/Pending future requests | High |
| View history | Tap "History" tab | Shows all past requests | High |
| Tap request | Tap on a request row | Navigate to Leave Request Details | High |
| Pull to refresh | Pull down gesture | Data reloads, spinner shown, values updated | Medium |
| Empty upcoming | No future leaves | "No upcoming leaves" empty state | Medium |
| Empty history | No past requests | "No leave history yet" with CTA | Medium |
| Balance display | User with 8 annual, 4 sick days | Cards show "8" and "4" respectively | High |
| Notification tap | Tap bell icon | Navigate to Notifications | Medium |
| Bottom nav - Home | Tap Home icon | Navigate to Home Dashboard | Medium |
| Bottom nav - Settings | Tap Settings icon | Navigate to Settings | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Dashboard data is fetched on screen load and on pull-to-refresh
- **SR-02:** Leave balance values are calculated server-side: Remaining = Quota - Approved Leave Days
- **SR-03:** "Leaves This Year" counts all approved leaves within the current calendar year
- **SR-04:** "Upcoming Leave" tab filters requests where: status IN (Pending, Approved) AND from_date >= today
- **SR-05:** "History" tab filters requests where: to_date < today (all statuses)
- **SR-06:** Requests are sorted by date: Upcoming = nearest first, History = most recent first
- **SR-07:** Request list shows maximum 10 items per tab on dashboard; "View All" navigates to full list (if implemented)
- **SR-08:** Dashboard data is cached locally for offline viewing; pull-to-refresh fetches fresh data
- **SR-09:** User avatar falls back to initials or placeholder if no profile photo exists
- **SR-10:** Leave types displayed match WEB-APP configuration: Annual, Sick, Personal, Maternity, Unpaid

**State Transitions:**
```
[Screen loads] → [Skeleton state] → [Data fetched] → [Populated state]
[User pulls down] → [Refresh spinner] → [API call] → [Data updated]
[User taps tab] → [Tab switches] → [Content updates]
[User taps Apply for Leave] → [Navigate to Create screen]
[User taps request row] → [Navigate to Details screen]
```

**Leave Balance Calculation:**
- Annual Leave Remaining = Annual Quota - Sum of Approved Annual Leave Days
- Sick Leave Remaining = Sick Quota - Sum of Approved Sick Leave Days
- Leaves This Year = Count of Approved Leave Requests in current year

**Dependencies:**
- **EP-001 US-001 (Authentication):** User must be signed in
- **WEB-APP EP-002:** Leave management API, leave types, quotas
- **EP-001 US-005 (User Profile):** User avatar, department, position data

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Dashboard loads with skeleton state to prevent layout shift
- **UX-02:** Leave balance cards use large typography (24px) for quick scanning
- **UX-03:** Color-coded leave type indicators enable instant recognition
- **UX-04:** Hero card provides clear, prominent CTA for the primary action (apply for leave)
- **UX-05:** Tab switcher allows quick context switching between upcoming and past leaves
- **UX-06:** Pull-to-refresh follows native iOS/Android gesture conventions
- **UX-07:** Request rows have sufficient touch target height (89px) for easy tapping
- **UX-08:** Bottom navigation provides persistent access to key app sections
- **UX-09:** Active tab/nav item is visually highlighted (yellow background for nav, filled for tab)
- **UX-10:** Date format uses ordinal suffixes (24th, 26th) for readability
- **UX-11:** "F" and "T" prefixes for From/To dates are space-efficient for mobile

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Standard mobile (393px width) | Design as specified |
| Small mobile (< 360px) | Balance cards may wrap; text truncated with ellipsis |
| Large mobile (> 414px) | Additional horizontal padding; cards expand |
| Tablet | Out of scope for this release |

**Accessibility Requirements:**
- [x] All interactive elements have minimum 44x44 touch target
- [x] Screen reader labels for all icons and buttons
- [x] Balance values announced with context (e.g., "Annual Leave Remaining: 8 days")
- [x] Tab state announced (e.g., "Upcoming Leave, selected")
- [x] Request rows announced with full context (type, status, duration, dates)
- [x] Color is not the only indicator for status (checkmark icon accompanies color)
- [x] Sufficient color contrast for text (meets WCAG 2.1 AA)
- [x] Pull-to-refresh has haptic feedback (device-dependent)

**Design References:**
- Figma: [Leave Requests Dashboard](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3250-3755) (node `3250:3755`)
- Design tokens: 
  - Primary: `#171717`
  - Background: `#ffffff`
  - Muted foreground: `#737373`
  - Accent (Exnodes green): `#27ae60`
  - Border: `#e5e5e5`
  - Font: Geist (Regular, Medium, Semibold)
  - Radius: `rounded-lg` (8px), `rounded-full` (9999px for avatars)

---

## 8. Additional Information

### Out of Scope
- Manager approval actions (viewing team leave, approving/rejecting requests) — future enhancement
- Team calendar view showing multiple employees' leaves
- Leave type configuration (admin function, WEB-APP only)
- Export functionality (not applicable for mobile)
- Push notifications for status changes (separate story)
- Offline request submission (future enhancement)
- "TBD" balance card functionality (placeholder in design)
- Swipe-to-cancel gesture on request rows (future enhancement)
- Search/filter within request list on dashboard

### Open Questions

| Question | Status | Owner |
|----------|--------|-------|
| What should the 4th balance card ("TBD") display? | Pending | Product Owner |
| Should cancelled requests appear in History tab? | Suggested: Yes, all statuses | Product Owner |
| Maximum number of requests to show in each tab on dashboard? | Suggested: 5-10 with "View All" | Product Owner |
| Should "Upcoming Leave" include pending requests or only approved? | Suggested: Both Pending and Approved | Product Owner |
| Leave type color mapping (Sick, Vacation, Personal, etc.)? | Pending | Design Team |

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-02: Create Leave Request (planned) | Triggered from "Apply for Leave" button |
| DR-002-001-03: Leave Request Details (planned) | Triggered from tapping request row |
| DR-002-001-04: Leave Request List (planned) | Full list view from "View All" |
| EP-001 US-001: Authentication | User must be signed in |
| WEB-APP DR-002-001-01: Leave Requests List | Web equivalent with table view |

### Notes
- This is the **first dashboard-style screen** for mobile — establishes the pattern for other mobile dashboards (e.g., Request Tickets Dashboard)
- The **hero card with CTA** pattern could be reused across other modules to promote primary actions
- The **2x2 balance card grid** pattern can be reused for other numeric summaries
- The **tab switcher** pattern can be reused for other list views with categorical filtering
- **Request row format** (type indicator + status + duration + dates) can be standardized across mobile list items
- The **bottom navigation** pattern is shared across all screens in the mobile app

### Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Primary View | Full table with 9 columns | Dashboard summary + request list |
| Leave Balance | Not shown on list page | Prominent 2x2 card grid |
| Filters | Multi-select chip filters | Tab-based (Upcoming/History) |
| Create Action | "+ Add New" button | Hero card "Apply for Leave" CTA |
| Actions | Gear icon dropdown per row | Tap row to view details |
| Navigation | Sidebar | Bottom navigation bar |
| Refresh | Page reload | Pull-to-refresh gesture |

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-16 | BA Agent | Initial draft — Figma design context from node 3250:3755 |
