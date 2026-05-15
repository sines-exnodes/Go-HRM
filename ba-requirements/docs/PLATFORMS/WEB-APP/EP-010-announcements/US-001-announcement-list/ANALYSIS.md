---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-010
story_id: US-001
story_name: "Announcement List"
status: draft
version: "1.0"
last_updated: "2026-04-25"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
input_sources:
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3333:6039"
    extraction_date: "2026-04-25"
    description: "Announcement List screen design"
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3385:18633"
    extraction_date: "2026-05-05"
    description: "Announcement Details (Sent) screen design"
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3385:22326"
    extraction_date: "2026-05-06"
    description: "Send Announcement screen design"
---

# Analysis: Announcement List

**Epic:** EP-010 (Announcements)
**Story:** US-001-announcement-list
**Status:** Draft

---

## Business Context

Organizations need a formal channel to communicate important information to employees. Currently, announcements may be scattered across emails, chat tools, or bulletin boards with no centralized tracking. A dedicated announcement system provides:

- Single source of truth for company communications
- Structured publishing workflow (draft → publish)
- Targeted distribution to specific audiences
- Read tracking for compliance-sensitive announcements

---

## Scope

### In Scope

- Announcement list view (table with search, filters, pagination)
- Create announcement form
- Edit announcement (draft only)
- Publish/unpublish actions
- Delete announcement
- Status management (Draft, Published, Archived)

### Out of Scope

- Mobile push notifications
- Email distribution
- Rich media attachments (images, videos)
- Announcement analytics/read tracking dashboard
- Scheduled publishing (future enhancement)
- Department-specific targeting (future enhancement)

---

## Open Questions

- [ ] What fields are required for an announcement? (Title, Content, Priority?) — Owner: Product Owner
- [ ] Can published announcements be edited? — Owner: Product Owner
- [ ] Is there an archive vs delete distinction? — Owner: Product Owner

---

## Notes

This is the first story in the Announcements epic. The dr-agent will derive field definitions and workflows from Figma design if available.

---

## Design Context [ADD-ON]

**Source Information:**
- Figma File: exn-hr-design
- Node ID: 3333:6039
- Extraction Date: 2026-04-25

### Layout Overview

```
+------------------+------------------------------------------------+
|                  |  Breadcrumb / Breadcrumb / Breadcrumb          |
|     Sidebar      +------------------------------------------------+
|     (200px)      |  Announcement                                  |
|                  +------------------------------------------------+
|  - Logo          |  [Search...] [Last Announce Date] [Status (2)] |
|  - User Profile  |  [Reset]                     [Export] [+Add New]|
|  - Navigation    +------------------------------------------------+
|    - Users Mgmt  |  Title | Description | Created | Last Announce |
|    - Menu        |         | Date       | Date    | Recipient     |
|                  |         |            |         | Status|Action |
|                  +------------------------------------------------+
|                  |  Rows per page [10]    Page 1 of 10  [< 1 2 >] |
+------------------+------------------------------------------------+
```

### Component Inventory

| Component | Node ID | Purpose |
|-----------|---------|---------|
| Search Input | 3333:6090 | Search by title, description |
| Last Announce Date Filter | 3333:6091 | Date range filter button |
| Status Filter | 3333:6094 | Multi-select status filter (shows count) |
| Reset Button | 3333:6095 | Clear all filters |
| Export Button | 3333:6097 | Export filtered data |
| Add New Button | 3333:6098 | Create new announcement |
| Table | 3333:6099 | 7-column data table |
| Pagination | 3333:6109 | Page navigation with rows-per-page |

### Table Columns (from design)

| Column | Width | Content Type |
|--------|-------|--------------|
| Title | 263px | Text |
| Description | 263px | Text |
| Created Date | 263px | Date |
| Last Announce Date | 263px | Date |
| Recipient | 263px | Number (count) |
| Status | 263px | Badge (TBD shown) |
| Action | 76px | Gear icon |

### Design Constraints

- Page dimensions: 1920x1080 (desktop)
- Sidebar width: 200px fixed
- Content area: 1694px
- Table column widths: 263px each (Action: 76px)
- Pagination: Shows "Rows per page" dropdown + "Page X of Y" + page numbers

### Key Design Observations

1. **Status Filter** shows "(2)" indicating multi-select with count display
2. **Recipient** column shows numeric values (10, 20) suggesting target audience count
3. **Status badges** show "TBD" placeholder - actual values to be confirmed
4. **Action column** uses gear icon pattern consistent with other list views
5. **Date filters** include "Last Announce Date" as a dedicated filter

---

### Announcement Details Screen (Sent variant)

**Source Information:**
- Figma File: exn-hr-design
- Node ID: 3385:18633
- Variant: Announcement Details (Sent)
- Extraction Date: 2026-05-05

**Layout Overview:**

```
+------------------+----------------------------------------------------+
|                  |  Breadcrumb / Breadcrumb / Breadcrumb              |
|     Sidebar      +----------------------------------------------------+
|     (200px)      |  [<-] Announcement Details > Hung King's Festival  |
|                  +-----+----------------------------------------------+
|                  |Acts | [Card 1: Announcement Details]               |
|                  |207px|  Title           Hung King's Festival        |
|  - Logo          |     |  Created Date    08/06/1991                  |
|  - Profile       |Det  |  Last Announce   08/06/1991                  |
|  - Navigation    |Upd  |  Recipients      Everyone                    |
|                  |Send |  Status          [Sent]                      |
|                  |Del  |                                              |
|                  |     | [Card 2: Description]                        |
|                  |     |  Hello everyone, ... (rich body text)        |
+------------------+-----+----------------------------------------------+
```

**Component Inventory (Details screen):**

| Component | Node ID | Purpose |
|-----------|---------|---------|
| Back Arrow | 3385:19672 | Navigate back to Announcement List |
| Title Header | 3385:19673 | "Announcement Details" page title |
| Breadcrumb (announcement) | 3385:19678 | Shows current announcement title (e.g., "Hung King's Festival") |
| Action Panel | 3385:19696 | Vertical button stack (left side, 207px wide) |
| Details Button | 3385:19697 | Active button (current view) |
| Update Announcement Button | 3385:19698 | Edit action (50% opacity for Sent status) |
| Send Announcement Button | 3385:20533 | Send action (50% opacity for Sent status) |
| Delete Announcement Button | 3385:19701 | Delete action (always available based on status) |
| Details Card | 3385:18698 | First card with summary fields |
| Description Card | 3385:18705 | Second card with full content |
| Status Badge | 3385:19097 | Shows status (Sent in this variant) |

**Data Fields Shown (from design):**

| Field | Value (sample) | Source |
|-------|----------------|--------|
| Title | Hung King's Festival | Card 1 |
| Created Date | 08/06/1991 | Card 1 |
| Last Announce Date | 08/06/1991 | Card 1 |
| Recipients | Everyone | Card 1 |
| Status | Sent (badge) | Card 1 |
| Description | Long-form rich text body | Card 2 |

**Design Constraints:**

- Two-panel layout: action panel (207px) + content card (600px)
- Action panel: 4 buttons stacked vertically with 16px gap
- Active button uses #f5f5f5 (general/accent) background; inactive use ghost (transparent) + 50% opacity
- Buttons height: 36px each
- Details card: 600px wide, ~260px tall (5 field rows + title + padding)
- Description card: 600px wide, ~408px tall (auto-sized to content)
- Field rows: Label (152px) + Value (360px) layout
- Field rows separated by 1px horizontal lines

**Key Design Observations (Details screen):**

1. **Two-panel layout** matches Request Ticket Details and User Details patterns
2. **Sent variant** — Update and Send buttons rendered at 50% opacity (disabled), confirming "immutable after send" rule
3. **Action button order** (top to bottom): Details, Update Announcement, Send Announcement, Delete Announcement
4. **Status-conditional visibility/state** — for Sent: only Details (active) and Delete are interactive; for Draft: Details (active) + Update + Send + Delete all interactive
5. **Breadcrumb pattern** — shows the announcement title in a secondary breadcrumb after the page title
6. **No inline editing** — all modifications via separate Update Announcement view
7. **Description preserves formatting** — newlines, bold headings, bullet lists rendered as authored
8. **"Last Announce Date"** — appears alongside Created Date, indicating the actual send/distribution timestamp

---

### Send Announcement Screen

**Source Information:**
- Figma File: exn-hr-design
- Node ID: 3385:22326
- Extraction Date: 2026-05-06

**Layout Overview:**

```
+------------------+----------------------------------------------------+
|                  |  Breadcrumb / Breadcrumb / Breadcrumb              |
|     Sidebar      +----------------------------------------------------+
|     (200px)      |  [<-] Announcement Details > Hung King's Festival  |
|                  +-----+----------------------------------------------+
|                  |Acts | [Card: Receiver Info]                        |
|                  |207px|  Everyone           [Toggle ON ]             |
|  - Logo          |     |    Everyone will receive this announcement   |
|  - Profile       |Det  |  Specific one      [Toggle OFF]              |
|  - Navigation    |Upd  |    Select employee to receive the announce.. |
|                  |Send |  * Employee                                  |
|                  |Del  |    [Select employee...        v]             |
|                  |     |                                              |
|                  |     | [Save] (full-width, dark, 600x40)            |
+------------------+-----+----------------------------------------------+
```

**Component Inventory (Send screen):**

| Component | Node ID | Purpose |
|-----------|---------|---------|
| Back Arrow | 3385:22385 | Navigate back to Announcement Details |
| Title Header | 3385:22386 | "Announcement Details" page title |
| Breadcrumb (announcement) | 3385:22387 | Shows current announcement title (e.g., "Hung King's Festival") |
| Action Panel | 3385:22391 | Vertical button stack (left side, 207px wide) |
| Details Button | 3385:22392 | Navigate to Details (interactive) |
| Update Announcement Button | 3385:22393 | Navigate to Update form (interactive) |
| Send Announcement Button | 3385:22394 | Active button (current view) |
| Delete Announcement Button | 3385:22395 | Open Delete confirmation (interactive) |
| Receiver Info Card | 3385:22489 | Single card containing Everyone/Specific one toggles + Employee field |
| Everyone Toggle | 3385:22496 | Send to all employees (default ON) |
| Specific one Toggle | 3385:22501 | Send to selected employees only (default OFF) |
| Employee Field | 3385:22502 | Vertical Field with multi-select dropdown |
| Save Button | 3385:22405 | Full-width 600x40, #010101 background, white "Save" text |

**Design Constraints:**

- Two-panel layout: action panel (207px) + content area (600px)
- Action panel: 4 buttons stacked vertically with 16px gap
- Active button (Send Announcement) uses #f5f5f5 (general/accent) background; sibling buttons use ghost (transparent) and remain interactive
- Buttons height: 36px each
- Receiver Info card: 600px wide, ~259px tall (title + 2 toggle rows + Employee field)
- Save button: 600px wide, 40px tall (separate from card, sits below)
- Toggle dimensions: 33x18 px
- No "Announcement Details" content card on this screen — only Receiver Info

**Key Design Observations (Send screen):**

1. **Receiver-only screen** — unlike Update Announcement (which shows Title/Description/Label + Receiver Info), Send Announcement focuses exclusively on the Receiver Info card. Title, description, and label are not editable here.
2. **Save button labeled "Save"** (not "Send") in Figma — the explicit Send confirmation dialog conveys the irreversible Send semantics, while the button itself uses the same "Save" wording as other commit buttons across the entity's screens.
3. **Two-panel layout reused** from Announcement Details (DR-010-001-03) and Update Announcement (DR-010-001-04) — same spatial mental model across the entity.
4. **Action panel order** (top to bottom): Details, Update Announcement, Send Announcement (active), Delete Announcement.
5. **No Cancel/Discard button** on this screen — same standalone-tab deviation as Update Announcement (DR-010-001-04 v1.1). Back arrow + sibling action buttons handle navigation; the Send confirmation dialog is the irreversibility safety gate.
6. **Send is only reachable for Draft** (per Knowledge Base §9.4); on Sent records the Send button on Details / Update is at 50% opacity. Direct URL access to /:id/send for Sent records redirects to read-only Details.
7. **Success path:** Draft -> Sent transition + notification dispatch, then redirect to Announcement Details (Sent variant) — the user does NOT stay on this page after success (different from DR-010-001-04's stay-on-page Save deviation, because Send is no longer applicable once the transition completes).
