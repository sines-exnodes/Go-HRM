---
document_type: DETAIL_REQUIREMENT
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-004
story_id: US-001
story_name: "Home Announcements"
detail_id: DR-004-001-02
detail_name: "Announcement List (Full Screen)"
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
  - path: "./DR-004-001-01-top-5-latest-announcements.md"
    relationship: sibling
  - path: "../../../../WEB-APP/EP-010-announcements/US-001-announcement-list/details/DR-010-001-02-create-announcement.md"
    relationship: cross-platform
input_sources:
  - type: figma
    node_id: "3370:18068"
    description: "Announcements full list screen design"
    extraction_date: "2026-04-25"
---

# Detail Requirement: Announcement List (Full Screen)

**Detail ID:** DR-004-001-02
**Parent Story:** US-001-home-announcements
**Epic:** EP-004 (Announcements)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee**, I want to **view all company announcements in a full-screen list** so that **I can browse and read any company communication beyond the top 5 shown on the home screen widget**.

**Purpose:** Provide employees with complete access to all announcements targeted at them. While the home screen widget (DR-004-001-01) shows only the 5 most recent announcements, this full-screen list allows employees to scroll through the complete history of announcements they have received.

**Target Users:** All employees using the Exnodes HRM Mobile app

**Key Functionality:**
- Full-screen list view of all announcements (not limited to 5)
- Same card format as home widget (title, preview, date, category, unread indicator)
- Scrollable list with all targeted announcements
- Tap to view announcement details
- Back navigation to return to previous screen
- Pull-to-refresh to update the list

---

## 2. User Workflow

**Entry Point:** 
- "View All" link from the home screen announcement widget (DR-004-001-01)
- Back navigation from announcement detail screen
- Direct navigation from app menu (if applicable)

**Preconditions:**
- User is authenticated and logged into the mobile app
- User has an active employee account
- At least one announcement has been sent that targets this user

**Main Flow:**
1. User taps "View All" on the home screen announcement widget
2. System navigates to the Announcements full-screen list
3. System displays header with back arrow and "Announcements" title
4. System loads and displays all announcements targeted at this user
5. Each announcement card shows: unread indicator (if applicable), title, preview text, date, and category badge
6. User scrolls through the list to browse announcements
7. User taps on an announcement card to view full details
8. System navigates to the announcement detail screen
9. System marks the announcement as "read" for this user
10. User taps back arrow on detail screen to return to the list
11. List reflects updated read status (unread dot removed)

**Alternative Flows:**
- **Alt 1 - Back Navigation:** User taps back arrow in header to return to previous screen (home or wherever they came from)
- **Alt 2 - No Announcements:** If no announcements exist for this user, system displays empty state message
- **Alt 3 - Pull to Refresh:** User pulls down on the list to refresh announcement data from server
- **Alt 4 - Network Error:** If network fails, system displays error state with retry option
- **Alt 5 - Offline Mode:** If offline, system displays cached announcements with "Last updated" timestamp

**Exit Points:**
- **Back:** User taps back arrow to return to previous screen
- **Detail View:** User taps announcement card to view full details
- **Bottom Navigation:** User taps bottom nav icon to switch to another app section

---

## 3. Field Definitions

### Display Elements (Read-Only)

| Element Name | Type | Display Format | Validation | Description |
|--------------|------|----------------|------------|-------------|
| Announcement Title | Text | Single line, 14px medium, truncated with ellipsis if exceeds width | N/A | Title of the announcement |
| Preview Text | Text | 12px regular, max 3 lines, truncated with ellipsis | N/A | First portion of announcement content |
| Sent Date | Date | "DDth MMM YYYY" (e.g., "24th Apr 2026") | N/A | Date announcement was sent |
| Category Badge | Pill Badge | 12px semibold, gray background (#f1f5f9), rounded-full | N/A | Category/type of announcement |
| Unread Indicator | Dot | Blue dot (6x6 pixels), visible for unread only | N/A | Visual marker for unread announcements |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back Arrow | Icon Button | Always enabled | Navigate to previous screen | ArrowLeft icon (18x18) in header |
| Announcement Card | Touchable Area | Always enabled | Navigate to detail screen + mark as read | Entire card is tappable |
| Pull-to-Refresh | Gesture | Always enabled | Refresh announcement data from server | Standard pull-down gesture |
| Bottom Navigation | Tab Bar | Always visible | Switch to other app sections | 4-icon persistent navigation |

### Card Layout Structure

| Component | Position | Dimensions |
|-----------|----------|------------|
| Content Area | Top of card | Full width, padding 12px |
| Unread Dot | Left of title | 6x6 pixels, 4px gap from title |
| Title | Top-left (after dot if present) | 14px medium, single line |
| Preview Text | Below title | 12px regular, max 3 lines |
| Metadata Row | Below content area | Full width, padding-left 20px |
| Date Section | Left of metadata row | CalendarDots icon + date text |
| Category Badge | Right of metadata row | Pill shape, padding 10px horizontal |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Announcement Title | Text | N/A (card not shown) | Truncated single line | What the announcement is about |
| Preview Text | Text | N/A | Max 3 lines truncated | Brief summary of content |
| Sent Date | Date | N/A | "DDth MMM YYYY" | When announcement was published |
| Category | Text/Badge | Show badge with category name | Pill badge | Type/category of announcement |
| Unread Status | Visual | Dot hidden | Blue dot visible | Whether user has viewed this announcement |

### Category Badge Values (from Figma)

| Category | Badge Text | Usage |
|----------|-----------|-------|
| Public Holiday | "Public Holiday" | Holiday announcements |
| Announcement | "Announcement" | General announcements |
| Reminder | "Reminder" | Reminder notifications |
| Meeting | "Meeting" | Meeting-related announcements |
| (Others) | Dynamic | Additional categories from backend |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch in progress | Skeleton cards (placeholder rows matching card layout) |
| Empty | No announcements exist for this user | Empty state: icon + "No announcements yet" message |
| Populated | 1 or more announcements available | Scrollable list of announcement cards |
| Error | Network failure or server error | Error message with "Tap to retry" action |
| Refreshing | Pull-to-refresh in progress | Refresh spinner at top of list |
| Offline | No network connection | Cached announcements + "Last updated: [timestamp]" banner |

### Screen Layout

| Component | Position | Height | Details |
|-----------|----------|--------|---------|
| Header | Top, fixed | 70px | Back arrow + "Announcements" title, padding 12px horizontal, 20px vertical |
| Content Area | Below header, scrollable | Flexible | List of announcement cards with 20px gap between cards |
| Bottom Navigation | Bottom, fixed | 76px | 4-icon navigation bar (Home, Calendar, Scroll, Settings) |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** System displays all announcements targeted at the current user (not limited to 5), sorted by sent date (newest first)
- **AC-02:** Each announcement card displays: title, preview text (max 3 lines), sent date, and category badge
- **AC-03:** Unread announcements display a visible blue dot indicator to the left of the title; read announcements do not show the dot
- **AC-04:** Tapping an announcement card navigates user to the announcement detail screen
- **AC-05:** Announcement is marked as "read" when user taps to view it (dot indicator removed on next list load/refresh)
- **AC-06:** Back arrow in header navigates user to the previous screen
- **AC-07:** Pull-to-refresh gesture fetches fresh data from the server and updates the list
- **AC-08:** Empty state displays appropriate message and icon when no announcements exist for the user
- **AC-09:** Loading state displays skeleton cards matching the card layout to prevent layout shift
- **AC-10:** Only announcements with status "Sent" and targeting this user (Everyone OR specifically included) are displayed
- **AC-11:** Bottom navigation remains visible and functional throughout the screen
- **AC-12:** Read status is synced across devices (if user reads on mobile, web shows as read too)
- **AC-13:** Date format follows "DDth MMM YYYY" pattern (e.g., "24th Apr 2026")
- **AC-14:** List scrolls smoothly with no pagination controls (continuous scroll)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Display all announcements | User with 15 announcements | All 15 cards displayed in scrollable list | High |
| Unread indicator | User has not viewed 3 announcements | Blue dot visible on those 3 cards only | High |
| Read marking | User taps announcement card | Navigate to detail; dot removed on return | High |
| Empty state | User with no announcements | Empty state message and icon displayed | Medium |
| Pull-to-refresh | User pulls down on list | Data refreshes, new announcements appear at top | Medium |
| Back navigation | User taps back arrow | Return to previous screen (home widget or detail) | High |
| Targeted announcements | Announcement sent to "Everyone" | All employees see it in their list | High |
| Specific targeting | Announcement sent to specific employees | Only targeted employees see it | High |
| Sort order | Multiple announcements with different dates | Newest (most recent sent_date) appears at top | High |
| Category badges | Announcements with different categories | Correct badge displayed for each (Holiday, Meeting, etc.) | Medium |
| Network error | Offline or server error | Error message with "Tap to retry" option | Medium |
| Offline cache | No network, cached data available | Cached announcements shown with "Last updated" | Low |
| Cross-device sync | Read on mobile, check web | Web shows same announcement as read | Medium |

---

## 6. System Rules

**Business Logic:**

- **Rule 1:** Display all announcements with status = "Sent" (Draft announcements are never shown on mobile)
- **Rule 2:** Display only announcements where the user is in the target audience (either "Everyone" was selected, or user was specifically included in recipient list)
- **Rule 3:** Sort announcements by sent_date descending (newest first)
- **Rule 4:** No limit on number of announcements displayed (unlike home widget which limits to 5)
- **Rule 5:** Read status is tracked per user per announcement (server-side)
- **Rule 6:** Marking as "read" occurs when user taps the card to navigate to detail (not when list is viewed)
- **Rule 7:** Read status cannot be reverted to unread (one-way transition)
- **Rule 8:** Read status is synced across all devices and platforms for the same user

**State Transitions:**

```
[Unread] -> [User taps card] -> [Read]
[Read] -> (no reverse transition - stays read permanently)
```

**Data Source:**

- Announcements API endpoint returns announcements filtered by:
  - status = "Sent"
  - target_audience includes current user OR target_audience = "Everyone"
  - Sorted by sent_date DESC
  - No limit (returns all matching announcements)

**Dependencies:**

- **EP-001 Foundation:** User authentication required to view announcements
- **WEB-APP EP-010:** Announcement data created and sent via web admin
- **DR-010-001-02:** Create Announcement (defines data structure and targeting)
- **DR-004-001-01:** Home widget provides "View All" entry point to this screen

**API Considerations:**

- Same endpoint as home widget but without the `limit=5` parameter
- Consider pagination or infinite scroll if announcement volume becomes large
- Cache announcements locally for offline access
- Track read status changes and sync to server

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Skeleton loading cards match exact layout of real cards (title placeholder, preview lines, date/badge row) to prevent layout shift
- **Optimization 2:** Unread indicator uses a distinct blue color visible against white card background
- **Optimization 3:** Touch target for entire card (not just text) for easier tapping on mobile
- **Optimization 4:** Pull-to-refresh uses native iOS/Android gesture with familiar spinner indicator
- **Optimization 5:** Smooth scrolling with no janky behavior during rapid scroll
- **Optimization 6:** Cards have subtle shadows or borders to visually separate them in the list
- **Optimization 7:** Back arrow touch target extends beyond icon for easier tapping

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Small phone (<375px) | Cards full width, preview text truncation more aggressive (2 lines) |
| Standard phone (375-428px) | Default layout as per Figma, 3 lines preview text |
| Large phone (>428px) | Slightly more preview text visible, same card structure |

**Accessibility Requirements:**

- [x] Touch targets minimum 44x44 points for all interactive elements
- [x] Screen reader announces: "Announcement: [title], [date], [category], [unread/read]"
- [x] Unread indicator has non-color alternative (screen reader announces "unread" or "new")
- [x] Pull-to-refresh announced by screen reader ("Refreshing announcements")
- [x] Sufficient color contrast (text #020817 on white #ffffff background)
- [x] Focus indicators visible for keyboard/switch control navigation
- [x] Back button announces "Go back" or "Return to previous screen"

**Mobile-Specific Behaviors:**

| Behavior | Detail |
|----------|--------|
| Touch Targets | Minimum 44x44 points for all interactive elements |
| Pull-to-Refresh | Standard iOS/Android gesture with spinner indicator at top |
| Skeleton Loading | Cards show shimmer animation during initial load |
| Haptic Feedback | Optional haptic on error states (device-dependent) |
| Offline Handling | Cache last-fetched announcements; show cached data with "Last updated: [timestamp]" banner |
| Scroll Performance | Hardware-accelerated scrolling, lazy load cards if needed |
| Memory Management | Recycle card views for large lists to prevent memory issues |

**Design References:**

- Figma Node ID: 3370:18068
- Card background: #ffffff
- Screen background: #f0f1f3
- Card radius: 8.8px (rounded-xl)
- Card padding: 12px
- Card gap: 20px
- Badge background: #f1f5f9
- Muted text (date): #64748b
- Primary text: #020817

---

## 8. Additional Information

### Out of Scope

- Creating or editing announcements (WEB-APP only)
- Announcement search functionality
- Announcement filtering by date, category, or read status
- Archiving announcements
- Announcement categories/labels management
- Push notification handling (separate feature)
- Announcement detail screen content (separate DR)
- Bulk mark as read functionality
- Delete or hide individual announcements

### Open Questions

- None (all questions resolved via knowledge base, Figma design, and existing patterns)

### Related Features

- **DR-004-001-01:** Top 5 Latest Announcements (home widget - provides entry point)
- **WEB-APP EP-010:** Announcement management (create, send, delete)
- **DR-010-001-02:** Create Announcement (defines targeting and send behavior)
- **Mobile Announcement Detail:** View full announcement content (future DR)

### Design Differences from Home Widget (DR-004-001-01)

| Aspect | Home Widget (DR-004-001-01) | Full List (This DR) |
|--------|----------------------------|---------------------|
| Display Limit | Top 5 only | All announcements |
| Header | "Announcements" + "View All" link | Back arrow + "Announcements" |
| Location | Embedded widget on home screen | Dedicated full screen |
| Navigation Out | Tap card or "View All" | Tap card or back arrow |
| Scroll | Limited widget area | Full screen scroll |

### Notes

- This feature extends the home widget (DR-004-001-01) to provide complete announcement access
- Entry point is primarily "View All" link from home widget
- Data structure and display format are consistent with home widget for familiarity
- Category badges are dynamic based on data; Figma shows examples (Holiday, Announcement, Reminder, Meeting)
- Read status synced across devices ensures consistent experience on mobile and web

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
