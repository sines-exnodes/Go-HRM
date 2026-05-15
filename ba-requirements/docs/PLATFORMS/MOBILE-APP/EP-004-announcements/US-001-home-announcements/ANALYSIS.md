---
document_type: ANALYSIS
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-004
story_id: US-001
story_name: "Home Announcements"
status: draft
version: "1.0"
last_updated: "2026-04-28"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../../../WEB-APP/EP-010-announcements/US-001-announcement-list/ANALYSIS.md"
    relationship: cross-platform
revision_history: []
input_sources:
  - type: figma
    node_id: "3370:18068"
    description: "Announcements full list screen - full screen view with all announcements"
    extraction_date: "2026-04-25"
  - type: figma
    node_id: "3370:18484"
    description: "Announcement Detail screen - full content view of single announcement"
    extraction_date: "2026-04-25"
---

# Analysis: Home Announcements

**Epic:** EP-004 (Announcements)
**Story:** US-001-home-announcements
**Status:** Draft

---

## Business Context

Employees need to stay informed about company announcements without actively seeking out information. By displaying the latest announcements on the mobile home screen, important communications are immediately visible when employees open the app.

This feature consumes announcement data created via WEB-APP EP-010 (Announcements). When HR/Admin sends an announcement, it appears on all targeted employees' mobile home screens.

---

## Scope

### In Scope

- Top 5 latest announcements widget on home screen
- Announcement card display (title, date, preview)
- Tap to view full announcement details
- Unread/read status indicator
- Pull-to-refresh to update list

### Out of Scope

- Creating or editing announcements (WEB-APP only)
- Announcement search or filters
- Archiving announcements
- Announcement categories/labels display
- Pagination beyond top 5 (separate "View All" screen)

---

## Open Questions

- [ ] What information shows in the announcement card preview? (Title only? Title + truncated content?) — Owner: Product Owner
- [ ] Should there be a "View All" link to see more than 5 announcements? — Owner: Product Owner

---

## Notes

This feature is the mobile consumer side of WEB-APP EP-010 Announcements. Data flows from WEB-APP (create/send) to MOBILE-APP (view).

Push notifications for new announcements are handled as part of the "Save & Send" action in WEB-APP DR-010-001-02.

---

## Design Context [ADD-ON]

### Source Information

- **Figma Node ID:** 3370:18068
- **Screen Name:** Announcements (Full List)
- **Extraction Date:** 2026-04-25

### Layout Overview

```
+---------------------------------------+
| [<] Announcements                     |  <- Header (70px)
+---------------------------------------+
|                                       |
| +-----------------------------------+ |
| | [*] Hung King's Festival          | |  <- Announcement Card
| | Lorem ipsum dolor sit amet...     | |     - Unread indicator (blue dot)
| +-----------------------------------+ |     - Title (14px medium)
| | [Cal] 24th Apr 2026 | Public Hol. | |     - Preview text (12px, 3 lines max)
| +-----------------------------------+ |     - Date + Category badge row
|                                       |
| +-----------------------------------+ |
| | Scheduled System Maintenance...   | |  <- Read announcement (no dot)
| | We would like to inform all...    | |
| +-----------------------------------+ |
| | [Cal] 24th Apr 2026 | Announcement| |
| +-----------------------------------+ |
|                                       |
| +-----------------------------------+ |
| | Important Security Update         | |
| | All users are required to...      | |
| +-----------------------------------+ |
| | [Cal] 31st Mar 2026 | Reminder    | |
| +-----------------------------------+ |
|                                       |
| +-----------------------------------+ |
| | [*] Quarterly Financial Review... | |  <- Unread
| | The next quarterly financial...   | |
| +-----------------------------------+ |
| | [Cal] 15th May 2026 | Meeting     | |
| +-----------------------------------+ |
|                                       |
+---------------------------------------+
| [Home] [Cal] [Scroll] [Settings]     |  <- Bottom Navigation (76px)
+---------------------------------------+
```

### Component Inventory

| Component | Node ID | Purpose |
|-----------|---------|---------|
| Header | 3370:18069 | Back arrow + "Announcements" title |
| ArrowLeft | 3370:18071 | Back navigation icon (18x18) |
| Announcement Card | 3370:18143, 3370:18417, 3370:18435, 3370:18451 | Individual announcement display |
| Unread Indicator | 3370:18433, 3370:18467 | Blue dot (6x6 ellipse) for unread items |
| Title | 3370:18148, 3370:18419, etc. | Announcement title (14px medium) |
| Preview Text | 3370:18410, 3370:18420, etc. | Content preview (12px, truncated) |
| Date Row | 3370:18154, 3370:18421, etc. | Calendar icon + date + category badge |
| CalendarDots Icon | 3370:18399 | Date indicator icon (18x18) |
| Category Badge | 3370:18406, 3370:18429, etc. | Pill badge showing category |
| Bottom Navigation | 3370:18470 | 4-icon navigation bar (Home, Calendar, Scroll, Settings) |

### Design Tokens

| Token | Value | Usage |
|-------|-------|-------|
| background/bg-background | #ffffff | Card backgrounds |
| background/bg-secondary | #f1f5f9 | Category badge background |
| text/text-foreground | #020817 | Primary text color |
| text/text-muted-foreground | #64748b | Date text color |
| spacing/3 | 12px | Card padding, screen margins |
| spacing/5 | 20px | Header vertical padding |
| radius/rounded-xl | 8.8px | Card corner radius |
| radius/rounded-full | 9999px | Badge pill shape |
| paragraph/small/font-size | 14px | Title text size |
| paragraph/mini/font-size | 12px | Preview and date text size |

### Key Design Observations

1. **Full-Screen List View:** Unlike the home widget (top 5), this is a dedicated full-screen with scrollable list
2. **No "View All" Link:** This IS the full list — no pagination or "load more" visible in design
3. **Card Structure:** Two-part card (content area + metadata row with date and category)
4. **Unread Indicator:** Blue dot (6x6) positioned to the left of the title
5. **Category Badges:** Different categories shown (Public Holiday, Announcement, Reminder, Meeting)
6. **Date Format:** "DDth MMM YYYY" format (e.g., "24th Apr 2026")
7. **Back Navigation:** ArrowLeft icon returns to previous screen (likely home)
8. **Bottom Navigation:** Persistent 4-icon nav bar (same as other screens)
9. **Scrollable Content Area:** Cards stack vertically with 20px gap between them
10. **No Search/Filter:** Full list view has no search or filter functionality in this design
