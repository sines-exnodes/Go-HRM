---
document_type: DETAIL_REQUIREMENT
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-004
story_id: US-001
story_name: "Home Announcements"
detail_id: DR-004-001-03
detail_name: "Announcement Detail"
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
  - path: "./DR-004-001-02-announcement-list.md"
    relationship: sibling
  - path: "../../../../WEB-APP/EP-010-announcements/US-001-announcement-list/details/DR-010-001-02-create-announcement.md"
    relationship: cross-platform
input_sources:
  - type: figma
    node_id: "3370:18484"
    description: "Announcement Detail screen design"
    extraction_date: "2026-04-25"
---

# Detail Requirement: Announcement Detail

**Detail ID:** DR-004-001-03
**Parent Story:** US-001-home-announcements
**Epic:** EP-004 (Announcements)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee**, I want to **view the full content of an announcement** so that **I can read the complete message and understand all details of the company communication**.

**Purpose:** Provide employees with a dedicated screen to read the full announcement content. While the home widget and list screens show truncated previews (title + 3 lines), this detail screen displays the complete announcement text with proper formatting, allowing employees to fully consume the information.

**Target Users:** All employees using the Exnodes HRM Mobile app

**Key Functionality:**
- Display full announcement title prominently
- Show sent date and category badge
- Display complete announcement content with rich text formatting (bold, lists, paragraphs)
- Mark announcement as "read" when viewed
- Back navigation to return to the list screen
- Scrollable content area for long announcements
- Persistent bottom navigation for app-wide navigation

---

## 2. User Workflow

**Entry Point:** 
- Tap on an announcement card from the home screen widget (DR-004-001-01)
- Tap on an announcement card from the full list screen (DR-004-001-02)

**Preconditions:**
- User is authenticated and logged into the mobile app
- User has an active employee account
- The announcement exists and targets this user (status = "Sent")

**Main Flow:**
1. User taps on an announcement card from widget or list
2. System navigates to the Announcement Detail screen
3. System displays header with back arrow and announcement title (truncated if long)
4. System loads and displays the full announcement content
5. Content area shows: title (large), date badge, category badge, and full text
6. System marks the announcement as "read" for this user (server-side)
7. User scrolls through the content if it exceeds screen height
8. User taps back arrow to return to the previous screen
9. Previous screen reflects updated read status (unread dot removed)

**Alternative Flows:**
- **Alt 1 - Long Content:** If announcement content exceeds visible area, user can scroll vertically to read all content
- **Alt 2 - Network Error on Load:** If content fails to load, display error state with retry option
- **Alt 3 - Offline Mode:** If cached content is available, display cached version; otherwise show offline error
- **Alt 4 - Bottom Navigation:** User taps a bottom nav icon to navigate to another app section (leaves detail screen)

**Exit Points:**
- **Back:** User taps back arrow to return to list or home widget
- **Bottom Navigation:** User taps bottom nav icon to switch to another app section

---

## 3. Field Definitions

### Display Elements (Read-Only)

| Element Name | Type | Display Format | Validation | Description |
|--------------|------|----------------|------------|-------------|
| Header Title | Text | 14px medium, single line, truncated with ellipsis | N/A | Announcement title in header (compact version) |
| Main Title | Text | 24px semibold, left-aligned, multi-line allowed | N/A | Full announcement title prominently displayed |
| Sent Date | Badge | "DDth MMM YYYY" (e.g., "24th Apr 2026") in white pill badge with calendar icon | N/A | Date announcement was sent |
| Category Badge | Badge | 12px semibold, dark background (#171717), white text, pill shape | N/A | Category/type of announcement |
| Content Body | Rich Text | 12px regular, black text, supports bold, lists, paragraphs | N/A | Full announcement content with formatting |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back Arrow | Icon Button | Always enabled | Navigate to previous screen | ArrowLeft icon (18x18) in header |
| Content Area | Scrollable | Scrollable when content exceeds visible area | Vertical scroll | Container for announcement content |
| Bottom Navigation | Tab Bar | Always visible, Home icon active | Switch to other app sections | 4-icon persistent navigation |

### Layout Structure (from Figma)

| Component | Position | Dimensions | Details |
|-----------|----------|------------|---------|
| Header | Top, fixed | 70px height | Back arrow + title, padding 12px horizontal, 20px vertical |
| Content Area | Below header, scrollable | Flexible | Title + badges + content card, padding 12px |
| Title Section | Top of content | Auto height | 24px title + badge row (date + category) |
| Content Card | Below title section | Auto height, white background | Full text with 20px padding, rounded corners (10px) |
| Bottom Navigation | Bottom, fixed | 76px height | 4-icon navigation bar (Home highlighted) |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Announcement Title | Text | N/A (title is required) | 24px semibold, multi-line | What the announcement is about |
| Sent Date | Date | N/A (date is required) | "DDth MMM YYYY" in badge | When announcement was published |
| Category | Text/Badge | Show badge with category name | Dark pill badge | Type/category of announcement |
| Content Body | Rich Text | N/A (content is required) | Formatted text (bold, lists, paragraphs) | Full announcement message |

### Content Formatting (from Figma Example)

The content body supports rich text formatting:
- **Bold text:** Section headers and emphasis (e.g., "Hello everyone,", "Notes:")
- **Paragraphs:** Standard paragraph spacing (12px between paragraphs)
- **Bulleted lists:** Indented bullet points for list items
- **Line height:** 16px for body text

### Category Badge Values

| Category | Badge Style | Usage |
|----------|-------------|-------|
| Public Holiday | Dark background (#171717), white text | Holiday announcements |
| Announcement | Dark background (#171717), white text | General announcements |
| Reminder | Dark background (#171717), white text | Reminder notifications |
| Meeting | Dark background (#171717), white text | Meeting-related announcements |
| (Others) | Dynamic from backend | Additional categories |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial content fetch in progress | Skeleton layout (title placeholder, content card skeleton) |
| Populated | Content loaded successfully | Full announcement detail with all elements |
| Error | Network failure or server error | Error message with "Tap to retry" action |
| Offline (cached) | No network, cached content available | Cached content displayed normally |
| Offline (no cache) | No network, no cached content | "Unable to load. Please check your connection." |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** System displays the full announcement title prominently (24px semibold) at the top of the content area
- **AC-02:** System displays the sent date in a white pill badge with calendar icon in "DDth MMM YYYY" format
- **AC-03:** System displays the category badge with dark background (#171717) and white text next to the date badge
- **AC-04:** System displays the complete announcement content with proper rich text formatting (bold, paragraphs, bullet lists)
- **AC-05:** Announcement is marked as "read" when user navigates to the detail screen (unread dot removed on list/widget)
- **AC-06:** Read status is synced across devices (if read on mobile, web shows as read too)
- **AC-07:** Back arrow in header navigates user to the previous screen (list or home widget)
- **AC-08:** Header title shows announcement title truncated with ellipsis if exceeds available width
- **AC-09:** Content area is scrollable when announcement content exceeds visible screen height
- **AC-10:** Bottom navigation remains visible and functional throughout the screen
- **AC-11:** Loading state displays skeleton layout matching the content structure
- **AC-12:** Error state displays appropriate message with retry action
- **AC-13:** Content card has white background with 10px rounded corners and 20px padding
- **AC-14:** Screen background color is #f0f1f3 (same as list screen)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| View full content | Tap announcement card | Detail screen with full title, date, category, and content | High |
| Mark as read | Navigate to detail screen | Announcement marked read; dot removed on list return | High |
| Long content scroll | Announcement with >500 words | Content scrollable, header and bottom nav fixed | High |
| Rich text formatting | Content with bold, lists, paragraphs | Formatting rendered correctly | High |
| Back navigation | Tap back arrow | Return to previous screen (list or home) | High |
| Cross-device sync | Read on mobile | Web shows same announcement as read | Medium |
| Network error | Offline without cache | Error message with retry option | Medium |
| Cached content | Offline with cache | Cached content displayed | Medium |
| Header title truncation | Title longer than header width | Truncated with ellipsis | Medium |
| Category badge display | Different category types | Correct category text in dark badge | Medium |

---

## 6. System Rules

**Business Logic:**

- **Rule 1:** Only announcements with status = "Sent" can be viewed (Draft announcements are never accessible on mobile)
- **Rule 2:** Only announcements targeting this user are accessible (either "Everyone" was selected, or user was specifically included)
- **Rule 3:** Viewing the detail screen triggers the "mark as read" action (not just loading the content)
- **Rule 4:** Read status is tracked per user per announcement (server-side)
- **Rule 5:** Read status cannot be reverted to unread (one-way transition: Unread -> Read)
- **Rule 6:** Read status is synced across all devices and platforms for the same user
- **Rule 7:** Announcement content may contain rich text formatting (bold, lists, paragraphs) created via WEB-APP
- **Rule 8:** If user navigates away before content fully loads, read status is NOT triggered

**State Transitions:**

```
[Unread] -> [User views detail screen] -> [Read]
[Read] -> (no reverse transition - stays read permanently)
```

**Data Source:**

- Announcement detail fetched from API by announcement ID
- Content includes: title, sent_date, category, full_content (rich text)
- API call triggers "mark as read" action on successful content load

**Dependencies:**

- **EP-001 Foundation:** User authentication required to view announcements
- **WEB-APP EP-010:** Announcement data created and sent via web admin
- **DR-010-001-02:** Create Announcement (defines content structure and targeting)
- **DR-004-001-01:** Home widget provides entry point to this screen
- **DR-004-001-02:** Full list screen provides entry point to this screen

**API Considerations:**

- Endpoint: `GET /api/v1/announcements/{id}` returns full announcement details
- Mark as read: `POST /api/v1/announcements/{id}/read` or combined with GET
- Cache announcement content locally for offline viewing
- Track read status changes and sync to server

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Header title uses same font style as list card title for visual consistency
- **Optimization 2:** Skeleton loading matches exact layout (title block, badge placeholders, content card)
- **Optimization 3:** Content area scrolls smoothly with momentum scrolling on iOS/Android
- **Optimization 4:** Back arrow touch target extends beyond icon bounds (minimum 44x44 points)
- **Optimization 5:** Content card has subtle white background to distinguish from screen background
- **Optimization 6:** Rich text preserves formatting from WEB-APP (bold, lists, paragraphs)
- **Optimization 7:** Date badge includes calendar icon for quick visual recognition

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Small phone (<375px) | Title may wrap to 2 lines; content text size unchanged |
| Standard phone (375-428px) | Default layout as per Figma |
| Large phone (>428px) | Wider content area; same proportions |

**Accessibility Requirements:**

- [x] Touch targets minimum 44x44 points for back arrow
- [x] Screen reader announces: "[Title], sent [date], category [category], [content]"
- [x] Content text has sufficient contrast (black #000000 on white #ffffff)
- [x] Header text has sufficient contrast (#020817 on #f0f1f3)
- [x] Focus indicators visible for keyboard/switch control navigation
- [x] Back button announces "Go back" or "Return to previous screen"
- [x] Rich text formatting conveyed to screen readers (headings, lists)

**Mobile-Specific Behaviors:**

| Behavior | Detail |
|----------|--------|
| Touch Targets | Minimum 44x44 points for back arrow |
| Scroll | Momentum scrolling for content area |
| Skeleton Loading | Shimmer animation during content fetch |
| Offline Handling | Cache announcement content; show cached version offline |
| Memory Management | No special concerns (single item view) |

**Design References (from Figma):**

- Screen Node ID: 3370:18484
- Screen background: #f0f1f3
- Header height: 70px
- Content padding: 12px
- Content card: white background, 10px radius, 20px internal padding
- Title: 24px semibold, #020817
- Date badge: white background, 8px radius, calendar icon + 12px text
- Category badge: #171717 background, white text, 8px radius
- Content text: 12px regular, black, 16px line height
- Bottom navigation: 76px height, white background, 500px radius pill

---

## 8. Additional Information

### Out of Scope

- Editing or deleting announcements (WEB-APP only)
- Replying to or commenting on announcements
- Sharing announcements externally
- Downloading attachments (announcements do not have attachments)
- Printing or exporting announcement content
- Announcement search from detail screen
- Navigate to next/previous announcement
- "Mark as unread" functionality

### Open Questions

- None (all questions resolved via knowledge base, Figma design, and existing patterns)

### Related Features

- **DR-004-001-01:** Top 5 Latest Announcements (home widget - provides entry point)
- **DR-004-001-02:** Announcement List (full screen - provides entry point)
- **WEB-APP EP-010:** Announcement management (create, send, delete)
- **DR-010-001-02:** Create Announcement (defines content structure)

### Design Pattern Alignment

This screen follows the **Mobile Detail View Pattern**:

| Pattern Element | Implementation |
|-----------------|----------------|
| Header with back arrow | ArrowLeft icon + truncated title |
| Read-only content | Title, badges, full content (no edit actions) |
| Scrollable content area | Flexible height content card |
| Fixed bottom navigation | 4-icon nav bar (consistent with other screens) |
| Mark as read on view | Automatic server-side status update |

### Notes

- This is the final screen in the announcement viewing flow: Widget -> List -> Detail
- Content formatting must match what was created in WEB-APP (rich text preserved)
- Read status tracking ensures employees can see which announcements they've viewed
- Cross-platform sync ensures consistent read status between mobile and web
- The example in Figma shows a Public Holiday announcement with holiday schedule details

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
