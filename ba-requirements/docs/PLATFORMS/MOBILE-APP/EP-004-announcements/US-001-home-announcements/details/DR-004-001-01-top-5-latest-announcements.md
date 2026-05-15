---
document_type: DETAIL_REQUIREMENT
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-004
story_id: US-001
story_name: "Home Announcements"
detail_id: DR-004-001-01
detail_name: "Top 5 Latest Announcements"
status: draft
version: "1.0"
created_date: 2026-04-25
last_updated: 2026-04-25
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
  - path: "../../../../WEB-APP/EP-010-announcements/US-001-announcement-list/details/DR-010-001-02-create-announcement.md"
    relationship: cross-platform
input_sources:
  - type: text
    description: "User-confirmed answers for tap behavior, View All link, and unread indicator"
    extraction_date: "2026-04-25"
---

# Detail Requirement: Top 5 Latest Announcements

**Detail ID:** DR-004-001-01
**Parent Story:** US-001-home-announcements
**Epic:** EP-004 (Announcements)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee**, I want to **see the latest company announcements on my mobile home screen** so that **I stay informed about important communications without needing to navigate to a separate section**.

**Purpose:** Ensure employees are immediately aware of company announcements when they open the mobile app. This passive information delivery reduces the chance of missing important communications and keeps all staff aligned with company news.

**Target Users:** All employees using the Exnodes HRM Mobile app

**Key Functionality:**
- Display the 5 most recent announcements on the home screen dashboard
- Visual indicator for unread announcements
- Tap to view full announcement details
- Quick access to view all announcements

---

## 2. User Workflow

**Entry Point:** Home screen dashboard (widget embedded on main screen after login)

**Preconditions:**
- User is authenticated and logged into the mobile app
- User has an active employee account
- At least one announcement has been sent that targets this user

**Main Flow:**
1. User opens the mobile app and lands on the home screen dashboard
2. System displays the "Announcements" widget showing up to 5 latest announcements
3. Each announcement card shows: title, sent date, preview text, and unread indicator (if applicable)
4. User scans the announcement list to see recent communications
5. User taps on an announcement card to view full details
6. System navigates to the announcement detail screen (separate DR)
7. System marks the announcement as "read" for this user
8. User reads the full announcement content
9. User navigates back to the home screen

**Alternative Flows:**
- **Alt 1 - View All:** User taps "View All" link to navigate to the full announcements list screen
- **Alt 2 - No Announcements:** If no announcements exist for this user, system displays empty state message
- **Alt 3 - Pull to Refresh:** User pulls down on the widget to refresh announcement data

**Exit Points:**
- **Success:** User views announcement detail or returns to home screen
- **Navigation:** User taps "View All" to see full list
- **Dismiss:** User scrolls past announcements widget to other home screen content

---

## 3. Field Definitions

### Display Elements (Read-Only)

| Element Name | Type | Display Format | Description |
|--------------|------|----------------|-------------|
| Announcement Title | Text | Single line, truncated with ellipsis if >2 lines | Title of the announcement |
| Sent Date | Date | "DD MMM YYYY" (e.g., "25 Apr 2026") | Date announcement was sent |
| Preview Text | Text | 2-3 lines max, truncated with ellipsis | First portion of announcement content |
| Unread Indicator | Dot | Small colored dot (e.g., blue) | Visible for unread announcements only |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Announcement Card | Touchable Area | Always enabled | Navigate to announcement detail screen | Entire card is tappable |
| View All Link | Text Link | Always visible when announcements exist | Navigate to full announcements list | Link text: "View All" |
| Pull-to-Refresh | Gesture | Always enabled | Refresh announcement data from server | Standard pull-down gesture |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Announcement Title | Text | N/A (card not shown) | Truncated if long | What the announcement is about |
| Sent Date | Date | N/A | "DD MMM YYYY" | When announcement was published |
| Preview Text | Text | N/A | 2-3 lines truncated | Brief summary of content |
| Unread Indicator | Visual | Hidden | Colored dot | Whether user has viewed this announcement |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch or refresh in progress | Skeleton cards (3-5 placeholder rows) |
| Empty | No announcements exist for this user | Empty state: icon + "No announcements yet" message |
| Populated | 1-5 announcements available | Announcement cards in list format |
| Error | Network failure or server error | Error message with "Tap to retry" action |
| Refreshing | Pull-to-refresh in progress | Refresh spinner at top of widget |

### Widget Layout

| Component | Position | Details |
|-----------|----------|---------|
| Widget Header | Top | "Announcements" title + "View All" link (right-aligned) |
| Announcement Cards | Below header | Vertical stack of up to 5 cards |
| Card Content | Per card | Unread dot (left) + Title + Date (right) + Preview text (below) |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** System displays up to 5 most recent announcements on the home screen, sorted by sent date (newest first)
- **AC-02:** Each announcement card displays: title, sent date, and preview text (truncated to 2-3 lines)
- **AC-03:** Unread announcements display a visible dot indicator; read announcements do not show the dot
- **AC-04:** Tapping an announcement card navigates user to the announcement detail screen
- **AC-05:** Announcement is marked as "read" when user taps to view it (dot indicator removed on next load)
- **AC-06:** "View All" link is visible and navigates to the full announcements list screen
- **AC-07:** Pull-to-refresh gesture fetches fresh data from the server and updates the widget
- **AC-08:** Empty state displays appropriate message when no announcements exist for the user
- **AC-09:** Loading state displays skeleton cards to prevent layout shift
- **AC-10:** Only announcements with status "Sent" and targeting this user are displayed

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Display announcements | User with 3 sent announcements | 3 announcement cards displayed | High |
| Display limit | User with 10 announcements | Only 5 most recent shown, "View All" visible | High |
| Unread indicator | User has not viewed announcement | Blue dot indicator visible on card | High |
| Read marking | User taps announcement card | Navigate to detail; dot removed on return | High |
| Empty state | User with no announcements | Empty state message displayed | Medium |
| Pull-to-refresh | User pulls down on widget | Data refreshes, new announcements appear | Medium |
| View All navigation | User taps "View All" | Navigate to full announcements list screen | Medium |
| Targeted announcements | Announcement sent to "Everyone" | All employees see it | High |
| Specific targeting | Announcement sent to specific employees | Only targeted employees see it | High |
| Network error | Offline or server error | Error message with retry option | Medium |

---

## 6. System Rules

**Business Logic:**

- **Rule 1:** Only display announcements with status = "Sent" (Draft announcements are never shown on mobile)
- **Rule 2:** Only display announcements where the user is in the target audience (either "Everyone" was selected, or user was specifically included)
- **Rule 3:** Sort announcements by sent_date descending (newest first)
- **Rule 4:** Limit display to 5 announcements maximum; additional announcements viewable via "View All"
- **Rule 5:** Read status is tracked per user per announcement (server-side)
- **Rule 6:** Marking as "read" occurs when user taps the card (not when widget is viewed)

**State Transitions:**

```
[Unread] -> [User taps card] -> [Read]
[Read] -> (no reverse transition - stays read)
```

**Dependencies:**

- **EP-001 Foundation:** User authentication required to view announcements
- **WEB-APP EP-010:** Announcement data created and sent via web admin
- **DR-010-001-02:** Create Announcement (defines data structure and targeting)

**Data Source:**

- Announcements API endpoint returns announcements filtered by:
  - status = "Sent"
  - target_audience includes current user OR target_audience = "Everyone"
  - Sorted by sent_date DESC
  - Limited to 5 records for widget (full list uses pagination)

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Skeleton loading cards match exact layout of real cards to prevent layout shift
- **Optimization 2:** Unread indicator uses a distinct color (blue or primary accent) visible against card background
- **Optimization 3:** Touch target for entire card (not just text) for easier tapping
- **Optimization 4:** Pull-to-refresh uses native iOS/Android gesture with familiar spinner
- **Optimization 5:** "View All" link uses accent color to indicate interactivity

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Small phone (<375px) | Cards stack vertically, text truncation more aggressive |
| Standard phone (375-428px) | Default layout, 2-3 lines preview text |
| Large phone (>428px) | Slightly more preview text visible |

**Accessibility Requirements:**

- [x] Touch targets minimum 44x44 points
- [x] Screen reader announces: "Announcement: [title], [date], [unread/read status]"
- [x] Unread indicator has non-color alternative (screen reader announces "unread")
- [x] Pull-to-refresh announced by screen reader
- [x] Sufficient color contrast for text and indicators

**Mobile-Specific Behaviors:**

| Behavior | Detail |
|----------|--------|
| Touch Targets | Minimum 44x44 points for all interactive elements |
| Pull-to-Refresh | Standard iOS/Android gesture with spinner indicator |
| Skeleton Loading | 3-5 skeleton card placeholders during initial load |
| Haptic Feedback | Optional haptic on error states (device-dependent) |
| Offline Handling | Cache last-fetched announcements; show cached data with "Last updated" timestamp |

---

## 8. Additional Information

### Out of Scope

- Creating or editing announcements (WEB-APP only)
- Announcement search functionality
- Announcement filtering by date or type
- Archiving announcements
- Announcement categories or labels display
- Push notification handling (separate feature)
- Announcement detail screen (separate DR)
- Full announcements list screen (separate DR)

### Open Questions

- None (all questions resolved via user confirmation and knowledge base)

### Related Features

- **WEB-APP EP-010:** Announcement management (create, send, delete)
- **DR-010-001-02:** Create Announcement (defines targeting and send behavior)
- **Mobile Announcement Detail:** View full announcement content (future DR)
- **Mobile Announcements List:** View all announcements with pagination (future DR)

### Notes

- This feature is the mobile consumer side of WEB-APP EP-010 Announcements
- Data flows from WEB-APP (create/send) to MOBILE-APP (view)
- Push notifications for new announcements are triggered by WEB-APP "Save & Send" action
- Read status is synced across devices (if user reads on mobile, web shows as read too)

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-04-25 | Pending |
| Product Owner | | | Pending |
| UX Designer | | | Pending |
| Tech Lead | | | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-25 | BA Agent | Initial draft |
