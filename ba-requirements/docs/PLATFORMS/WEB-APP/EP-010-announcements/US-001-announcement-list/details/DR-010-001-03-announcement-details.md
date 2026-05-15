---
document_type: DETAIL_REQUIREMENT
# Platform Information
platform: web-app
platform_display: "WEB-APP"
# Epic & Story Identification
epic_id: EP-010
story_id: US-001
story_name: "Announcement List"
# Detail Requirement Identification
detail_id: DR-010-001-03
detail_name: "Announcement Details"
# Status & Version
status: draft
version: "1.0"
created_date: 2026-05-05
last_updated: 2026-05-05
# Document linking
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "./DR-010-001-01-announcement-list.md"
    relationship: sibling
  - path: "./DR-010-001-02-create-announcement.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
# Input sources
input_sources:
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3385:18633"
    description: "Announcement Details (Sent) screen design"
    extraction_date: 2026-05-05
---

# Detail Requirement: Announcement Details

**Detail ID:** DR-010-001-03
**Story:** US-001-announcement-list
**Epic:** EP-010 (Announcements)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **HR Manager or Admin**, I want to **view the full details of a specific announcement and access status-appropriate management actions** so that **I can review what was published, follow up with edits or delivery actions while still in Draft, and remove obsolete records once Sent**.

**Purpose:** Provide a single, read-only consolidated view of one announcement, including its full description, recipient configuration, lifecycle dates, and current status. The Details page also serves as the **primary location for status-changing actions** (Update, Send, Delete) — buttons render based on the announcement's current status and the user's permission, in line with the Detail/Profile View Pattern (Knowledge Base §3.4).

**Target Users:**
- HR Managers with announcement management permission
- Admins with full system access

**Key Functionality:**
- Read-only display of all announcement metadata in a summary card
- Read-only display of the full description with preserved formatting (line breaks, headings, bullet lists)
- Status-aware action panel: Update, Send, Delete buttons enabled/disabled per current status
- Navigate back to list with preserved list state (search, filters, page)
- Trigger status transitions (Send) and destructive actions (Delete) with confirmation
- Trigger edit flow (Update Announcement) for Draft records only

---

## 2. User Workflow

**Entry Point:**
- Announcement List page > click a row OR click gear icon > "View Details" / "Details"
- Direct URL navigation (e.g., from a notification or shared link)

**Preconditions:**
- User is authenticated and logged in
- User has announcement view permission (via US-004 Permission Management)
- Announcement record exists and is not soft-deleted

**Main Flow:**
1. User clicks an announcement row (or gear icon > Details) on the Announcement List
2. System navigates to the Announcement Details page (URL pattern includes announcement ID)
3. System fetches the announcement record and displays a skeleton loading state during fetch
4. Once loaded, system renders:
   - Page header with back arrow, "Announcement Details" title, and breadcrumb showing the announcement title
   - Left action panel (207px) with status-aware buttons: Details (active), Update Announcement, Send Announcement, Delete Announcement
   - Card 1 "Announcement Details" — summary fields (Title, Created Date, Last Announce Date, Recipients, Status badge)
   - Card 2 "Description" — full description content with preserved formatting
5. Action button states are determined by the current status:
   - **Draft:** Update enabled, Send enabled, Delete enabled
   - **Sent:** Update disabled (50% opacity), Send disabled (50% opacity), Delete enabled
6. User chooses one of the available actions:
   - **Click back arrow** > navigate to Announcement List with preserved state
   - **Click "Details"** > no-op (already on this view, button reflects active state)
   - **Click "Update Announcement"** (Draft only) > navigate to Edit Announcement form pre-filled with current data
   - **Click "Send Announcement"** (Draft only) > open Send confirmation dialog
   - **Click "Delete Announcement"** > open Delete confirmation dialog

**Alternative Flows:**
- **Alt 1 - Send Announcement (Draft):** User clicks Send > confirmation dialog "Send this announcement to recipients?" > on confirm, system updates status to Sent, distributes via email + push, shows success toast, refreshes the page in place (status badge updates to Sent, action buttons re-render with Update and Send disabled)
- **Alt 2 - Delete Announcement:** User clicks Delete > confirmation dialog "Are you sure you want to delete this announcement?" > on confirm, system soft-deletes the record, shows success toast, redirects to Announcement List
- **Alt 3 - Concurrent Status Change:** While viewing a Draft, another user sends or deletes the same announcement; on attempting Send/Update, server rejects with conflict and system shows error toast + reloads the page (or redirects to list if deleted)
- **Alt 4 - Permission Lost Mid-View:** If the user's permission is revoked while viewing, attempting any action returns a permission error toast and redirects to a fallback page
- **Alt 5 - Record Not Found:** Direct URL to a deleted/invalid announcement ID > display "Announcement not found or no longer available" with a button to return to the list
- **Alt 6 - Long Description:** Description content scrolls within its card or extends the page; full content is rendered without truncation

**Exit Points:**
- **Success (Send):** Toast "Announcement has been sent" + page refreshes in place with updated status; recipients receive email + push notifications asynchronously
- **Success (Delete):** Toast "Announcement has been deleted" + redirect to Announcement List with prior list state preserved
- **Success (Update navigation):** Navigate to Edit Announcement form (no toast)
- **Back Navigation:** Redirect to Announcement List with preserved search/filter/page state
- **Error:** Error toast displayed; user remains on Details page (or redirected to list if record no longer exists)

---

## 3. Field Definitions

### Display-Only Fields (read-only)

This screen has no user input fields — it is a read-only view. All "fields" below are display elements.

| Field Name | Display Type | Source | Format | Description |
|------------|--------------|--------|--------|-------------|
| Title | Text | Announcement record | Plain text, no truncation | Announcement subject/headline |
| Created Date | Date | Announcement record (auto-generated on save) | DD/MM/YYYY | When the announcement was first created (Draft saved or Sent directly) |
| Last Announce Date | Date | Announcement record (set when sent) | DD/MM/YYYY | When the announcement was actually sent to recipients; "—" if still Draft |
| Recipients | Text | Announcement record | "Everyone" OR comma-separated list of employee names | Who received / will receive the announcement |
| Status | Badge | Announcement record | Gray (Draft) or Green (Sent) badge | Current lifecycle state |
| Description | Rich text | Announcement record | Multi-line text, preserves line breaks, bold headings, bullet lists | Full announcement content |

### Interaction Elements (Action Panel — Left)

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back Arrow | Icon Button | Always enabled | Navigate to Announcement List with preserved state | Top-left of header (24x24 ArrowLeft icon) |
| Details Button | Tertiary Button | Always active (current view) | No-op (visual indicator only) | First button in action panel; uses #f5f5f5 background to indicate active state |
| Update Announcement Button | Tertiary Button | Enabled only when status = Draft AND user has management permission | Navigate to Edit Announcement form | Disabled state at 50% opacity for Sent status |
| Send Announcement Button | Tertiary Button | Enabled only when status = Draft AND user has management permission | Open Send confirmation dialog | Disabled state at 50% opacity for Sent status |
| Delete Announcement Button | Tertiary Button (danger) | Enabled when user has management permission (regardless of status) | Open Delete confirmation dialog | Available for both Draft and Sent records |

### Action Button Visibility/State Matrix

| Status | Details | Update Announcement | Send Announcement | Delete Announcement |
|--------|---------|---------------------|-------------------|---------------------|
| Draft | Active | Enabled | Enabled | Enabled |
| Sent | Active | Disabled (50% opacity) | Disabled (50% opacity) | Enabled |

### Recipients Display Logic

| Receiver Configuration | Displayed Value |
|------------------------|------------------|
| `everyone = true` | "Everyone" |
| `everyone = false`, 1 employee | Employee full name (e.g., "John Doe") |
| `everyone = false`, 2-3 employees | Comma-separated names (e.g., "John Doe, Jane Smith, Alice Lee") |
| `everyone = false`, 4+ employees | First 2 names + "and N others" (e.g., "John Doe, Jane Smith and 5 others") with hover tooltip showing full list |

---

## 4. Data Display

### Layout Structure

| Region | Position | Dimensions | Content |
|--------|----------|------------|---------|
| Header Row | Top | Full width, 29px tall | Back arrow + "Announcement Details" title + breadcrumb separator + announcement title |
| Action Panel | Left | 207px wide, 192px tall (4 buttons × 36px + gaps) | Vertical stack: Details, Update, Send, Delete |
| Content Card 1 | Right (next to action panel) | 600px wide, ~260px tall | "Announcement Details" — 5 summary field rows |
| Content Card 2 | Below Card 1 | 600px wide, auto-sized to content | "Description" — full content body |

### Card 1 Field Rows (Announcement Details)

| Row | Label (152px wide) | Value (360px wide) |
|-----|--------------------|---------------------|
| 1 | Title | [Announcement title] |
| 2 | Created Date | DD/MM/YYYY |
| 3 | Last Announce Date | DD/MM/YYYY or "—" if Draft |
| 4 | Recipients | "Everyone" or comma-separated names |
| 5 | Status | [Status badge] |

Field rows separated by 1px horizontal lines (#e5e5e5).

### Card 2 — Description

- Card title: "Description"
- Body: full announcement description text
- Preserves formatting: line breaks, headings (bold), bullet lists
- No truncation; long content scrolls naturally with the page
- Empty content not possible (Description is mandatory at create time)

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch | Skeleton placeholders for both cards and action buttons |
| Loaded - Draft | Announcement loaded with status = Draft | All 4 action buttons interactive; Status badge gray "Draft" |
| Loaded - Sent | Announcement loaded with status = Sent | Update + Send buttons at 50% opacity (disabled); Delete enabled; Status badge green "Sent" |
| Not Found | Announcement ID invalid or soft-deleted | Empty state: "Announcement not found or no longer available" + "Back to List" button |
| Error | Data fetch fails | Error message + "Retry" button |
| Action Processing | Send/Delete confirmation submitted | Confirmation dialog shows loading spinner; primary button disabled |
| Permission Denied | User lost view permission mid-session | Redirect to fallback page with toast "You no longer have permission to view this announcement" |

### Empty Field Behavior

| Field | When Empty | Display |
|-------|-----------|---------|
| Last Announce Date | Status = Draft (never sent) | "—" |
| Recipients | Cannot be empty (validated at create) | N/A |
| All other fields | Cannot be empty (mandatory at create) | N/A |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Users with announcement view permission can access the Announcement Details page; users without permission are redirected to fallback with permission-denied toast
- **AC-02:** Page renders two-panel layout: 207px-wide action panel on the left, 600px-wide content area on the right
- **AC-03:** Header shows back arrow, "Announcement Details" title, and a breadcrumb displaying the announcement title
- **AC-04:** Card 1 "Announcement Details" displays Title, Created Date, Last Announce Date, Recipients, and Status badge as five labeled rows separated by horizontal lines
- **AC-05:** Card 2 "Description" displays the full announcement content with preserved formatting (line breaks, bold headings, bullet lists)
- **AC-06:** Status badge color matches: Draft = gray, Sent = green (per Knowledge Base §9.4)
- **AC-07:** When status = Draft, all four action buttons (Details, Update Announcement, Send Announcement, Delete Announcement) are interactive
- **AC-08:** When status = Sent, Update Announcement and Send Announcement buttons render at 50% opacity and are non-interactive; Details and Delete remain available
- **AC-09:** Clicking back arrow navigates to Announcement List with preserved search, filter, and page state
- **AC-10:** Clicking "Update Announcement" (Draft only) navigates to the Edit Announcement form pre-filled with current values
- **AC-11:** Clicking "Send Announcement" (Draft only) opens confirmation dialog "Send this announcement to recipients?"
- **AC-12:** Confirming Send transitions status from Draft to Sent, triggers email + push notifications asynchronously, shows toast "Announcement has been sent", and refreshes the page in place (status badge and action buttons update without full navigation)
- **AC-13:** Clicking "Delete Announcement" opens confirmation dialog "Are you sure you want to delete this announcement?"
- **AC-14:** Confirming Delete soft-deletes the record, shows toast "Announcement has been deleted", and redirects to Announcement List
- **AC-15:** Recipients field displays "Everyone" when targeting all employees, individual name(s) when 1-3 specific recipients, or "Name1, Name2 and N others" with tooltip when 4+ specific recipients
- **AC-16:** Last Announce Date displays "—" when status = Draft (never sent)
- **AC-17:** Loading state shows skeleton placeholders for both cards and the action panel
- **AC-18:** Direct URL access to a deleted or invalid announcement ID shows "Announcement not found or no longer available" empty state with a "Back to List" button
- **AC-19:** All status transitions require confirmation dialog (no direct/silent transitions)
- **AC-20:** Server-side concurrency check: if status changed by another user (e.g., already Sent or already Deleted) at the moment of action, system shows error toast and refreshes/redirects appropriately
- **AC-21:** Description text preserves multi-line formatting and is fully visible without truncation regardless of length

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| View Draft announcement | Click row in list with status=Draft | All 4 action buttons enabled; gray Draft badge | High |
| View Sent announcement | Click row in list with status=Sent | Update and Send buttons at 50% opacity; green Sent badge | High |
| Send Draft announcement | Click Send > confirm dialog | Status changes to Sent in place, badge updates, buttons re-render disabled, toast "Announcement has been sent" | High |
| Delete Draft announcement | Click Delete > confirm | Soft-deleted, redirect to list, toast "Announcement has been deleted" | High |
| Delete Sent announcement | Click Delete > confirm | Soft-deleted, redirect to list, toast shown | High |
| Update Draft announcement | Click Update Announcement | Navigate to Edit form pre-filled with current values | High |
| Update Sent announcement | Try clicking Update (50% opacity) | Button non-interactive — no navigation occurs | High |
| Send Sent announcement | Try clicking Send (50% opacity) | Button non-interactive — no dialog opens | High |
| Back navigation preserves list state | Apply filter on list, click row, click back | Returns to list with same filter applied | Medium |
| Recipients = Everyone | View announcement targeted at everyone | Recipients row shows "Everyone" | Medium |
| Recipients = 1 specific employee | View announcement targeted at 1 employee | Recipients row shows "John Doe" | Medium |
| Recipients = 5 specific employees | View announcement targeted at 5 employees | Recipients row shows "John Doe, Jane Smith and 3 others" with hover tooltip | Medium |
| Last Announce Date for Draft | View Draft announcement | Last Announce Date shows "—" | Medium |
| Last Announce Date for Sent | View Sent announcement | Last Announce Date shows actual send date in DD/MM/YYYY | Medium |
| Long description renders fully | View announcement with 1000+ char description | Full content rendered, no truncation, formatting preserved | Medium |
| Concurrent send conflict | User A views Draft; User B sends it; User A clicks Send | Error toast "This announcement has already been sent" + page refreshes showing Sent state | High |
| Concurrent delete conflict | User A views announcement; User B deletes it; User A clicks any action | Error toast "This announcement no longer exists" + redirect to list | High |
| Direct URL — invalid ID | Navigate to /announcements/99999 | Empty state "Announcement not found or no longer available" + Back to List button | Medium |
| Permission revoked mid-view | User loses permission; clicks Send | Permission error toast + redirect to fallback page | Medium |
| Skeleton loading | Slow network / first load | Skeleton rows for cards and action buttons until data arrives | Low |
| Cancel send confirmation | Click Send > Cancel in dialog | Dialog closes, no status change, remain on Details page | Medium |
| Cancel delete confirmation | Click Delete > Cancel in dialog | Dialog closes, record preserved, remain on Details page | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only users with announcement view permission (configured via US-004) can access the Announcement Details page; access controlled both at the route level and on direct URL access
- **Rule 2:** Action buttons render based on a combination of (a) current announcement status and (b) user's management permission — view-only users (no management permission) see no action buttons (action panel hidden or showing only Details)
- **Rule 3:** "Update Announcement" is enabled only when announcement status = Draft and user has management permission; otherwise rendered at 50% opacity and non-interactive
- **Rule 4:** "Send Announcement" is enabled only when announcement status = Draft and user has management permission; otherwise rendered at 50% opacity and non-interactive
- **Rule 5:** "Delete Announcement" is available for both Draft and Sent statuses provided the user has management permission
- **Rule 6:** Sending transitions status from Draft → Sent and sets the Last Announce Date to the current UTC timestamp (server-generated)
- **Rule 7:** Sending triggers two notification channels asynchronously (per DR-010-001-02 Rule 10):
  - **Email notification** to each recipient's registered email
  - **Push notification** to each recipient's mobile device via MOBILE-APP
- **Rule 8:** Notifications are sent in a background job — the UI does not block waiting for delivery; success toast appears immediately on status update
- **Rule 9:** After a successful Send, the Details page refreshes in place — the URL does not change, but the status badge, Last Announce Date, and action button states all re-render (Update and Send become disabled)
- **Rule 10:** Deleting an announcement performs a soft delete: record preserved in database with `deleted = true` flag, removed from all active list views, filters, searches, and exports (per Knowledge Base §5.1 Soft Delete Transactional)
- **Rule 11:** Deleting a Sent announcement is administrative cleanup only — recipients who already received the email/push retain their copy; no retraction occurs
- **Rule 12:** All status transitions require a confirmation dialog before proceeding (per Knowledge Base §3.4 Confirmation Required)
- **Rule 13:** Concurrent action protection: server verifies the announcement is still in the expected status before applying any transition; on mismatch, server returns a conflict response and the UI shows an error toast and reloads (or redirects if record was deleted)
- **Rule 14:** Back navigation preserves list state via stored query parameters or session-scoped state (per Knowledge Base §3.3 State Preservation)
- **Rule 15:** All dates are stored in UTC and displayed in DD/MM/YYYY format using the user's timezone for display
- **Rule 16:** Last Announce Date is null (displayed as "—") for Draft announcements; populated only after successful Send
- **Rule 17:** The Details page shows the announcement description in full — no truncation regardless of length (in contrast to list view truncation per Knowledge Base §3.2)
- **Rule 18:** Description preserves formatting authored at create time: line breaks, bold (markdown-style), bullet lists; rendering is consistent with what employees see in MOBILE-APP announcement detail
- **Rule 19:** Recipients field displays a derived summary string from the stored receiver configuration:
  - `everyone = true` → "Everyone"
  - `everyone = false` with employee list → comma-separated names; if more than 3, show first 2 + "and N others" with hover tooltip listing all names
- **Rule 20:** Permission verification is performed both on initial page load and on each action click — revocation mid-session is handled gracefully with a toast and redirect

**State Transitions (from Details page):**

```
[Viewing Draft] → [Click Send] → [Confirm Dialog] → [Confirm] → [Sent + Notifications dispatched + Page refreshes]
[Viewing Draft] → [Click Update] → [Edit Form (pre-filled)]
[Viewing Draft] → [Click Delete] → [Confirm Dialog] → [Confirm] → [Soft-Deleted + Redirect to List]
[Viewing Sent] → [Click Delete] → [Confirm Dialog] → [Confirm] → [Soft-Deleted + Redirect to List]
[Viewing Any] → [Click Back Arrow] → [Announcement List with preserved state]
```

**Permission Model:**

| Action | Required Permission | Status Constraint |
|--------|---------------------|-------------------|
| View Details Page | Announcement View | Any non-deleted status |
| Update Announcement | Announcement Management | Draft only |
| Send Announcement | Announcement Management | Draft only |
| Delete Announcement | Announcement Management | Draft or Sent |

**Confirmation Dialogs:**

| Action | Dialog Title | Dialog Message | Primary Button | Cancel Button |
|--------|--------------|----------------|----------------|---------------|
| Send | "Send Announcement?" | "This announcement will be sent to the selected recipients via email and push notification. This action cannot be undone." | "Send" (primary, dark bg) | "Cancel" (default focus) |
| Delete | "Delete Announcement?" | "Are you sure you want to delete this announcement? This action cannot be undone." | "Delete" (danger, red) | "Cancel" (default focus) |

**Dependencies:**

- US-004 Permission Management — provides view and management permission gates
- DR-010-001-01 Announcement List — entry point and back-navigation target
- DR-010-001-02 Create Announcement — establishes data shape, recipient model, notification flow
- Future DR (Edit Announcement) — destination of Update Announcement action; pre-fills from this record
- Email service — for sending email notifications on Send (per DR-010-001-02 dependencies)
- Push notification service — for mobile push delivery (per DR-010-001-02 dependencies)
- Authentication system — user must be logged in

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Two-panel layout keeps the most common actions (Update, Send, Delete) consistently in the same screen position regardless of description length, reducing cognitive load (per Knowledge Base §3.4 Status Management on Detail Page)
- **Optimization 2:** Disabled buttons remain visible at 50% opacity rather than hidden — gives users clear feedback that the action exists but is not currently allowed, and signals what the workflow looks like at other lifecycle stages
- **Optimization 3:** Active "Details" button uses a subtle accent background (#f5f5f5) to anchor the user — visually consistent with the active-state pattern used on Update Request Ticket (DR-003-001-06)
- **Optimization 4:** In-place page refresh after Send keeps the user on the same screen with updated context, avoiding unnecessary navigation hops
- **Optimization 5:** Confirmation dialog for Send explicitly mentions the irreversibility ("This action cannot be undone") and the channels used (email + push) so the user understands the magnitude before confirming
- **Optimization 6:** Cancel is the default keyboard focus in confirmation dialogs to prevent accidental destructive actions (per Knowledge Base §5.2 Default Focus)
- **Optimization 7:** Description rendering preserves the author's formatting verbatim — no re-flow or summarization — so what HR sees here matches what employees receive
- **Optimization 8:** Recipients summary uses a "first 2 + N others" pattern with hover tooltip for large recipient lists, balancing scan-ability with completeness
- **Optimization 9:** Back arrow + breadcrumb give two visible navigation cues (single icon for quick exit, contextual breadcrumb for spatial awareness)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full two-panel layout: 207px action panel + 600px content card stacked vertically (Card 1 above Card 2) |
| Below Desktop | Out of scope — web admin is desktop-only |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab order: Back arrow → Action panel buttons (top to bottom) → Card content
- [x] Disabled buttons announce disabled state to screen readers (aria-disabled)
- [x] Screen reader compatible — all field labels associated with values; status badge has aria-label
- [x] Sufficient color contrast — disabled state at 50% opacity still meets WCAG AA when combined with aria-disabled
- [x] Focus indicators visible — clear focus ring on action buttons and back arrow
- [x] Confirmation dialogs trap focus inside the dialog and support Escape to cancel (per Knowledge Base §5.2)
- [x] Recipients tooltip accessible via keyboard focus, not hover-only

**Design References:**
- Figma Node: 3385:18633 (Announcement Details — Sent variant)
- Follows Detail/Profile View Pattern from Knowledge Base §3 (two-panel layout with status-aware action panel)
- Follows Status Management on Detail Page pattern from Knowledge Base §3.4
- Action panel layout consistent with Request Ticket Details (DR-003-001-05) and Update Request Ticket (DR-003-001-06)
- Status badges follow Knowledge Base §9.4 Announcement Status definitions

---

## 8. Additional Information

### Out of Scope

- Inline editing of fields on the Details page (all edits go through the Update Announcement form)
- Bulk actions (multi-select, bulk delete)
- Read receipts / per-recipient delivery status display
- Notification delivery status tracking (which recipients received email vs push)
- Resend / retry failed notifications from the Details page
- Announcement comments or replies
- Announcement scheduling (send at future date) — out of scope at epic level
- Announcement versioning / revision history
- Mobile responsive layout (web admin desktop-only)
- Print / export announcement to PDF
- Sharing / copying a link to the announcement (would require deep-link permission model)
- Scheduling unsend / recall after Sent

### Open Questions

- None — all decisions resolved against established patterns and the Sent variant Figma frame.

### Related Features

- DR-010-001-01: Announcement List (parent list view; back-navigation target)
- DR-010-001-02: Create Announcement (defines data model and notification flow)
- DR-010-001-04: Edit Announcement (planned — destination of Update Announcement action)
- DR-010-001-05: Send Announcement Confirmation (planned — confirmation dialog details)
- DR-010-001-06: Delete Announcement Confirmation (planned — confirmation dialog details)
- US-004: Permission Management (view + management permission source)
- DR-004-001-03 (MOBILE-APP): Announcement Detail (employee-facing read view; should render description identically)

### Notes

- The Figma frame analyzed is explicitly named "Announcement Details (Sent)" — the "(Sent)" suffix in the frame name signals that a status-specific variant exists. The 50% opacity on Update and Send buttons confirms the immutability rule for Sent announcements (per DR-010-001-02 Rule 4).
- Although Card 1 in the design shows "Last Announce Date" with the same value as "Created Date" (08/06/1991), in production these will diverge: Created Date = save timestamp; Last Announce Date = send timestamp (only set after Send action).
- The action panel order matches the natural lifecycle flow: Details (current view) → Update (modify) → Send (publish) → Delete (remove). Buttons are stacked vertically with 16px gap.
- Per Knowledge Base §5.1, transactional records use soft delete — once "deleted", an announcement is hidden from all active views but preserved in the database for audit purposes.
- "Last Announce Date" matches the column name used on the Announcement List (per ANALYSIS.md Design Context) — terminology stays consistent across list and detail views.
- The Details page is the canonical location for status transitions on this entity, mirroring the Request Ticket Details pattern (DR-003-001-05) and adhering to Knowledge Base §3.4 Status Management on Detail Page.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | | | Pending |
| Product Owner | | | Pending |
| UX Designer | | | Pending |
| Tech Lead | | | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-05 | Claude | Initial draft from Figma "Announcement Details (Sent)" frame extraction; applies Detail/Profile View Pattern (Knowledge Base §3 and §3.4) with status-aware action panel for Draft/Sent statuses |
