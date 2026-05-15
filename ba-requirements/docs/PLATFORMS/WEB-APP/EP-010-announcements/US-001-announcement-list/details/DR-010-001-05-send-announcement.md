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
detail_id: DR-010-001-05
detail_name: "Send Announcement"
# Status & Version
status: draft
version: "1.1"
created_date: 2026-05-06
last_updated: 2026-05-06
# Document linking
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "./DR-010-001-01-announcement-list.md"
    relationship: sibling
  - path: "./DR-010-001-02-create-announcement.md"
    relationship: sibling
  - path: "./DR-010-001-03-announcement-details.md"
    relationship: sibling
  - path: "./DR-010-001-04-update-announcement.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
# Input sources
input_sources:
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3385:22326"
    description: "Send Announcement screen design"
    extraction_date: 2026-05-06
---

# Detail Requirement: Send Announcement

**Detail ID:** DR-010-001-05
**Story:** US-001-announcement-list
**Epic:** EP-010 (Announcements)
**Status:** Draft
**Version:** 1.1

---

## 1. Use Case Description

As an **HR Manager or Admin**, I want to **finalize the recipient configuration of a Draft announcement and dispatch it to employees** so that **the announcement is delivered to the intended audience via email and push notification, and the record transitions to the Sent (immutable) state**.

**Purpose:** Provide the dedicated screen for executing the **Draft -> Sent** status transition. The screen presents only the **Receiver Info** for last-mile confirmation (Everyone / Specific one + Employee selection); the announcement's title, description, and label are not editable here — those are managed through Update Announcement (DR-010-001-04). On Send (the bottom action button), the system transitions the record's status to Sent, sets the Last Announce Date timestamp, dispatches notifications (email + push) asynchronously, and redirects to the now-immutable Announcement Details page. Send is permanent — once executed, the announcement cannot be edited or recalled (per Knowledge Base §9.4).

**Target Users:**
- HR Managers with announcement management permission
- Admins with full system access

**Key Functionality:**
- Pre-fill the Receiver Info with the Draft's current recipient configuration (Everyone or Specific one + employee list)
- Allow last-minute adjustment of the audience before sending
- The bottom **Send** button confirms the receiver configuration AND triggers the Draft -> Sent status transition with notification dispatch (via an explicit confirmation dialog as the irreversibility safety gate)
- Two-panel layout consistent with Announcement Details (DR-010-001-03) and Update Announcement (DR-010-001-04)
- Status-aware sibling action panel (Details, Update, Send active, Delete)
- No Cancel/Discard button — back arrow and sibling action buttons navigate immediately with silent discard of unsaved Receiver changes

---

## 2. User Workflow

**Entry Point:**
- Announcement Details page (DR-010-001-03) > Left action panel > "Send Announcement" button (only enabled when status = Draft)
- Update Announcement page (DR-010-001-04) > Left action panel > "Send Announcement" button (always interactive in Update view; Update is only reachable for Draft)
- Direct URL navigation (e.g., bookmarked or deep-linked send URL); access still gated by status check + permission

**Preconditions:**
- User is authenticated and logged in
- User has announcement management permission (via US-004 Permission Management)
- The target announcement exists, is not soft-deleted, and is currently in **Draft** status

**Main Flow:**
1. User clicks "Send Announcement" on the Announcement Details page (Draft only) or on the Update Announcement action panel
2. System navigates to the Send Announcement page (URL pattern includes announcement ID)
3. System displays the page header: back arrow + "Announcement Details" page title + breadcrumb showing the announcement title (e.g., "Hung King's Festival")
4. System renders the two-panel layout:
   - **Left action panel (207px):** Details, Update Announcement, Send Announcement (active with #f5f5f5 accent), Delete Announcement
   - **Right form card (600px):** Card title "Receiver Info" with Everyone toggle + description, Specific one toggle + description, and an Employee multi-select field (visible at all times, enabled only when Specific one is ON)
5. System fetches the announcement record and pre-fills the Receiver Info with the Draft's current recipient configuration during a brief skeleton loading state
6. User reviews the pre-filled audience; if needed, the user can adjust:
   - Toggle "Everyone" ON to send to all employees (disables Specific one + Employee dropdown at 50% opacity)
   - Toggle "Specific one" ON to send only to selected employees (enables the Employee dropdown; at least one employee required)
7. User clicks **Send** at the bottom of the form (600px-wide, 40px-tall, #010101 background, white "Send" text)
8. System opens a confirmation dialog: "Send this announcement to recipients? This announcement will be sent via email and push notification. This action cannot be undone."
9. User confirms by clicking "Send" (primary, dark) in the dialog or backs out by clicking "Cancel" (default focus)
10. On confirm, system validates the receiver configuration client-side; on success, sends the request to the server with the final receiver configuration
11. Server transitions the record from Draft -> Sent, sets `last_announce_date` to the current UTC timestamp, persists the receiver configuration, and enqueues the notification job
12. Server returns success; system shows toast "Announcement has been sent" and **redirects to the Announcement Details page** (DR-010-001-03), which now renders the Sent variant (status badge = Sent (green), Update + Send buttons at 50% opacity, Delete still available)
13. Notification job dispatches email to each recipient's registered email and push notification to each recipient's mobile device asynchronously in the background — the UI does not block waiting for notification delivery

**Alternative Flows:**
- **Alt 1 - Validation Error (no employee on Specific one):** User toggles Specific one ON but leaves the Employee dropdown empty, then clicks Send > inline error "Please select at least one employee" appears below the Employee field; the Send action is blocked (no confirmation dialog opens)
- **Alt 2 - Cancel Confirmation Dialog:** User clicks Send > confirmation dialog opens > user clicks "Cancel" > dialog closes, no status change, user remains on the Send Announcement page with the form intact
- **Alt 3 - Back-Navigation:** User clicks the back arrow at any time > immediate navigation to Announcement Details; any unsaved Receiver changes are discarded silently (no confirmation, no Cancel/Discard button exists on this screen)
- **Alt 4 - Click Sibling Action Buttons:** User clicks "Details", "Update Announcement", or "Delete Announcement" on the action panel at any time > immediate navigation to that view/flow; any unsaved Receiver changes are discarded silently (no confirmation)
- **Alt 5 - Concurrent Status Change (Sent):** While on the Send page, another user sends the announcement; on Send (after dialog confirm), server returns 409 Conflict; system shows error toast "This announcement has already been sent" and redirects to the (now read-only) Announcement Details page
- **Alt 6 - Concurrent Delete:** While on the Send page, another user deletes the announcement; on Send (after dialog confirm), server returns 404; system shows error toast "This announcement no longer exists" and redirects to Announcement List
- **Alt 7 - Permission Lost Mid-Session:** User's management permission is revoked while on the Send page; on Send (after dialog confirm), server returns 403; toast "You no longer have permission to send this announcement" and redirect to fallback page
- **Alt 8 - Direct URL to Sent Announcement:** User opens the Send URL of a Sent announcement directly; system blocks load and redirects to the read-only Announcement Details with toast "This announcement has already been sent and can no longer be sent again"
- **Alt 9 - Direct URL to Deleted/Invalid Announcement:** Empty state "Announcement not found or no longer available" + "Back to List" button
- **Alt 10 - Server Error During Send:** Network or server failure on the Send submission > error toast "Failed to send announcement. Please try again." > user remains on the Send page with the receiver configuration intact; the announcement remains in Draft (no partial transition)
- **Alt 11 - Notification Dispatch Failure (Background):** The status transition succeeds and the user sees the success toast + redirect; if some recipients' email or push delivery fails downstream, those failures are handled by the notification job and do not affect the UI flow (no retry from this screen)

**Exit Points:**
- **Success:** Toast "Announcement has been sent" + redirect to Announcement Details page (DR-010-001-03), which now renders the Sent variant; recipients receive email + push notifications asynchronously in the background
- **Cancel Confirmation Dialog:** Dialog closes, no status change, user remains on the Send Announcement page
- **Back / Sibling Action Navigation:** Immediate navigation to the chosen target — no confirmation, unsaved Receiver changes silently discarded (there is no Cancel/Discard button on this screen)
- **Validation Error:** Inline error displayed under Employee field; user remains on form (no confirmation dialog opened)
- **Server Error:** Error toast + form remains; user can retry; announcement stays Draft

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Everyone | Toggle Switch | Mutually exclusive with "Specific one" | Yes (one must be selected) | Pre-filled with current Draft value | Send to all employees |
| Specific one | Toggle Switch | Mutually exclusive with "Everyone" | Yes (one must be selected) | Pre-filled with current Draft value | Send to selected employees only |
| Employee | Searchable Multi-select Dropdown | At least 1 employee required when "Specific one" is ON | Conditional | Pre-filled with current Draft selection (if any) | Specific recipient(s) when Specific one is ON |

**Note:** This screen edits ONLY the Receiver Info — Title, Description, and Label are not editable here. To change those, the user navigates to Update Announcement (DR-010-001-04) via the action panel. The data shape for Receiver Info is identical to Create Announcement (DR-010-001-02 §3) and Update Announcement (DR-010-001-04 §3) — the same `everyone` boolean + employee ID list model.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back Arrow | Icon Button | Always enabled | Navigate to Announcement Details immediately (unsaved Receiver changes discarded silently) | Top-left of header (24x24 ArrowLeft icon) |
| Details Button | Tertiary Button | Always interactive | Navigate to Announcement Details immediately (unsaved Receiver changes discarded silently) | First button in action panel |
| Update Announcement Button | Tertiary Button | Always interactive (Send is only reachable for Draft, and Update is also enabled for Draft) | Navigate to Update Announcement form immediately (unsaved Receiver changes discarded silently) | Second button in action panel |
| Send Announcement Button | Tertiary Button | Active (current view) | No-op (visual indicator only) | Highlighted with #f5f5f5 accent background |
| Delete Announcement Button | Tertiary Button (danger) | Always interactive | Open Delete confirmation immediately (unsaved Receiver changes discarded silently) | Bottom of action panel |
| Send Button | Primary Button | Enabled when Receiver Info is valid (any required selection satisfied) | Open Send confirmation dialog | Full-width 600px, 40px tall, dark background (#010101), white **"Send"** text |
| Everyone Toggle | Toggle Switch | Mutually exclusive with Specific one | Turn Everyone ON, Specific one OFF, disable Employee dropdown | Pre-filled state from Draft record |
| Specific one Toggle | Toggle Switch | Mutually exclusive with Everyone | Turn Specific one ON, Everyone OFF, enable Employee dropdown | Pre-filled state from Draft record |
| Employee Dropdown | Searchable Multi-select | Disabled at 50% opacity when Everyone is ON; enabled at 100% opacity when Specific one is ON | Open employee selection list (active employees only) | Pre-filled with current Draft recipients |

### Status-Aware Action Panel (Send view)

| Status | Details | Update Announcement | Send Announcement | Delete Announcement |
|--------|---------|---------------------|-------------------|---------------------|
| Draft (only state Send is reachable in) | Interactive | Interactive | Active (current view) | Interactive |

**Note:** Send Announcement is only enabled for Draft status (per Knowledge Base §9.4). Sent announcements never reach this view — the Send action is at 50% opacity on the Details and Update sibling pages once the record is Sent. If a user attempts direct URL access to the send URL of a Sent record, the system redirects to the read-only Details page (Alt 8).

### Receiver Selection Behavior (parity with Create — DR-010-001-02 §3 and Update — DR-010-001-04 §3)

| State | Everyone Toggle | Specific one Toggle | Employee Dropdown |
|-------|-----------------|---------------------|-------------------|
| Pre-filled "Everyone" (Draft default) | ON | OFF | Disabled (50% opacity) |
| Pre-filled "Specific" | OFF | ON | Enabled (100% opacity, mandatory, populated with current employees) |
| User toggles to "Everyone" | ON | OFF (auto-toggled OFF) | Disabled (50% opacity); current Specific selection retained in memory but not used on Send |
| User toggles to "Specific one" | OFF (auto-toggled OFF) | ON | Enabled (100% opacity, must contain >=1 employee at Send time) |

### Confirmation Dialog (triggered by clicking Send)

| Element | Detail |
|---------|--------|
| Dialog Title | "Send Announcement?" |
| Dialog Message | "This announcement will be sent to the selected recipients via email and push notification. This action cannot be undone." |
| Primary Button | "Send" (primary, dark #010101 background, white text) |
| Cancel Button | "Cancel" (default keyboard focus) |
| Recipient Summary in Dialog | Display the resolved audience: "Everyone" or "X employee(s) selected" so the user has one final visual confirmation of who will receive the announcement |

---

## 4. Data Display

### Layout Structure

| Region | Position | Dimensions | Content |
|--------|----------|------------|---------|
| Header Row | Top | Full width, 29px tall | Back arrow + "Announcement Details" title + breadcrumb separator + announcement title |
| Action Panel | Left | 207px wide, 192px tall (4 buttons x 36px + gaps) | Vertical stack: Details, Update Announcement, Send Announcement (active), Delete Announcement |
| Form Card | Right | 600px wide, ~259px tall | Card title "Receiver Info" + Everyone toggle row + Specific one toggle row + Employee field (Vertical Field instance) |
| Send Button | Below form card | 600px wide, 40px tall | Single full-width Send button (#010101 background, white **"Send"** text) |

### Form Card Layout (from Figma)

| Card | Title | Content |
|------|-------|---------|
| Form Card | Receiver Info | Everyone toggle row (label + description "Everyone will receive this announcement" + 33x18 toggle switch) + Specific one toggle row (label + description "Select employee to receive the announcement" + 33x18 toggle switch) + Employee field (Vertical Field, label "* Employee" + multi-select dropdown placeholder "Select employee") |

**Note on the Figma design:** The Send Announcement Figma frame focuses exclusively on the "Receiver Info" card — there is **no "Announcement Details" content card** on this screen. This is the deliberate UX choice for the Send action: by the time the user reaches Send, the title/description/label have already been authored (via Create or Update), and the only remaining decision is the audience. Title, description, and label are NOT shown on this screen and are NOT editable from here. To preview the announcement content before sending, the user navigates to the Details sibling button.

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch | Skeleton placeholders for action panel buttons and Receiver Info card |
| Loaded (Pre-filled) | Announcement loaded with status = Draft | Receiver Info card pre-filled with current configuration (Everyone or Specific one + employees); Send button enabled (assuming valid pre-filled state) |
| Editing Receiver | User toggles between Everyone / Specific one or modifies employee selection | Toggles update visually; Employee dropdown enables/disables at 50% / 100% opacity; Send remains enabled if valid |
| Validation Error | Specific one ON but employee list empty | Inline error "Please select at least one employee" below Employee field; clicking Send does not open confirmation dialog |
| Send Confirmation Dialog | User clicked Send with valid configuration | Modal dialog "Send Announcement?" with recipient summary, "Send" and "Cancel" buttons (Cancel has default focus) |
| Sending | User confirmed in dialog | Dialog "Send" button shows loading spinner, disabled to prevent double-submit; form fields and the page-level Send button disabled |
| Success | Server returned success | Toast "Announcement has been sent" + redirect to Announcement Details (Sent variant) |
| Conflict (Sent) | Server returns 409 because status changed to Sent | Error toast "This announcement has already been sent" + redirect to read-only Details |
| Conflict (Deleted) | Server returns 404 | Error toast "This announcement no longer exists" + redirect to Announcement List |
| Server Error | Generic 5xx from server | Error toast "Failed to send announcement. Please try again." + remain on Send page with form intact; status remains Draft |
| Not Found | Direct URL to invalid/deleted ID | Empty state "Announcement not found or no longer available" + "Back to List" button |
| Permission Denied | User loses management permission | Toast "You no longer have permission to send this announcement" + redirect to fallback |
| Direct URL to Sent | User opens Send URL of a Sent announcement | Redirect to read-only Details with toast "This announcement has already been sent and can no longer be sent again" |

### Validation Error Messages (parity with Create — DR-010-001-02 §4)

| Field | Error Condition | Error Message |
|-------|----------------|---------------|
| Employee | "Specific one" ON but no employee selected at Send click | "Please select at least one employee" |

**Note:** Title, Description, and Label validation are not relevant on this screen — those fields are not displayed or editable here. The Receiver Info card has only one validation rule (employee required when Specific one is ON), and one of Everyone / Specific one must always be ON (the toggle pair has exactly one selected at all times because it is mutually exclusive and one is pre-filled).

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Users with announcement management permission can access the Send Announcement page; users without permission are redirected to fallback with permission-denied toast
- **AC-02:** Send Announcement is only reachable for announcements in Draft status; attempting to load the send URL of a Sent announcement redirects to the read-only Details page with toast "This announcement has already been sent and can no longer be sent again"
- **AC-03:** Page renders the two-panel layout: 207px-wide action panel on the left, 600px-wide Receiver Info card on the right
- **AC-04:** Header shows back arrow + "Announcement Details" title + breadcrumb displaying the announcement title
- **AC-05:** Action panel renders four buttons in order: Details, Update Announcement, Send Announcement (active with #f5f5f5 accent background), Delete Announcement
- **AC-06:** The Receiver Info card is the only content card on this screen — Title, Description, and Label fields are NOT displayed and NOT editable here
- **AC-07:** The Receiver Info card pre-fills with the Draft's current recipient configuration on load
- **AC-08:** Everyone and Specific one toggles are mutually exclusive — turning one ON automatically turns the other OFF
- **AC-09:** Employee dropdown is disabled at 50% opacity when Everyone is ON, and enabled at 100% opacity when Specific one is ON
- **AC-10:** When Specific one is ON, at least one employee must be selected — inline error "Please select at least one employee" if empty when the user clicks Send
- **AC-11:** Employee dropdown is searchable (client-side) and lists active employees only (deactivated/inactive users are excluded — per DR-010-001-02 §2.8)
- **AC-12:** The page-level Send button is full-width 600px, 40px tall, with dark background (#010101) and white **"Send"** text (the visible label is "Send", not "Save")
- **AC-13:** Clicking Send with a valid configuration opens a confirmation dialog "Send Announcement?" with message "This announcement will be sent to the selected recipients via email and push notification. This action cannot be undone."
- **AC-14:** Confirmation dialog includes a recipient summary (Everyone OR "X employee(s) selected") so the user has one final visual confirmation
- **AC-15:** Confirmation dialog default keyboard focus is "Cancel" (not "Send") — prevents accidental confirmation
- **AC-16:** Confirming "Send" in the dialog transitions the record's status from Draft to Sent, sets the Last Announce Date to the current UTC timestamp, persists the final receiver configuration, and enqueues the notification job
- **AC-17:** Notifications dispatch via two channels asynchronously: email to each recipient's registered email and push notification to each recipient's mobile device — UI does not block waiting for delivery
- **AC-18:** On successful send, system shows toast "Announcement has been sent" and redirects to the Announcement Details page (DR-010-001-03), which now renders the Sent variant
- **AC-19:** Cancelling the confirmation dialog closes the dialog without changing status; user remains on the Send page with the form intact
- **AC-20:** Clicking the back arrow at any time navigates immediately to the Announcement Details page; unsaved Receiver changes are discarded silently — there is no confirmation dialog and no Cancel/Discard button on this screen
- **AC-21:** Clicking sibling action buttons (Details, Update Announcement, Delete Announcement) at any time navigates immediately to that target; unsaved Receiver changes are discarded silently (no confirmation dialog)
- **AC-22:** Loading state shows skeleton placeholders for action panel buttons and the Receiver Info card
- **AC-23:** If the announcement is sent by another user during the session, server returns 409 Conflict when the user confirms Send; system shows error toast "This announcement has already been sent" + redirects to read-only Details
- **AC-24:** If the announcement is deleted by another user during the session, server returns 404; system shows error toast "This announcement no longer exists" + redirects to Announcement List
- **AC-25:** If a generic server error occurs during Send, system shows error toast "Failed to send announcement. Please try again." and the announcement remains in Draft; user can retry from the same screen
- **AC-26:** Direct URL to a deleted/invalid announcement ID shows "Announcement not found or no longer available" empty state with "Back to List" button
- **AC-27:** Send Announcement is one of two screens that trigger the Draft -> Sent status transition + notification dispatch — the other being **Save & Send** on the Create Announcement form (per DR-010-001-02 §2.9). For already-Drafted records, this screen is the canonical path

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Open Send for Draft (Everyone pre-filled) | Click Send Announcement on Details page (Draft, Everyone) | Receiver Info card pre-filled with Everyone ON + Specific one OFF; Employee dropdown disabled at 50% opacity | High |
| Open Send for Draft (Specific pre-filled) | Click Send Announcement on Details page (Draft, Specific with 3 employees) | Receiver Info pre-filled with Specific one ON + Employee dropdown showing 3 employees enabled at 100% opacity | High |
| Send to Everyone (no changes) | Click Send | Confirmation dialog with recipient summary "Everyone"; on confirm, toast + redirect to Sent Details | High |
| Send to Specific (no changes) | Click Send | Confirmation dialog with recipient summary "X employee(s) selected"; on confirm, toast + redirect to Sent Details | High |
| Toggle Everyone -> Specific | Toggle Specific one ON | Everyone toggles OFF, Employee dropdown enables at 100% opacity, dropdown empty (or pre-filled if record had prior Specific selection) | High |
| Toggle Specific -> Everyone | Toggle Everyone ON | Specific one toggles OFF, Employee dropdown disables at 50% opacity (current Specific selection retained but ignored on Send) | Medium |
| Validation - Specific with no employee | Toggle Specific ON, leave dropdown empty, click Send | Inline error "Please select at least one employee"; confirmation dialog does NOT open | High |
| Validation - Specific with employee | Toggle Specific ON, select 2 employees, click Send | Confirmation dialog opens with "2 employee(s) selected" summary | High |
| Cancel confirmation dialog | Click Send > Cancel in dialog | Dialog closes, no status change, remain on Send page, form intact | High |
| Successful send (Everyone) | Send > confirm | Status -> Sent, Last Announce Date set, notifications enqueued, toast "Announcement has been sent", redirect to Sent Details | High |
| Successful send (Specific) | Send > confirm | Status -> Sent for selected employees only, notifications enqueued, toast + redirect to Sent Details | High |
| Email dispatch | Send > confirm with employee having registered email | Email sent to employee's registered address asynchronously | High |
| Push dispatch | Send > confirm with employee having MOBILE-APP push enabled | Push notification delivered to employee's mobile device asynchronously | High |
| Email skip - no registered email | Send > confirm with employee lacking email | Email skipped silently (no error); push still dispatched if applicable | Medium |
| Push skip - push disabled | Send > confirm with employee who disabled push | Push skipped silently (no error); email still dispatched if applicable | Medium |
| Back arrow with no changes | Open Send, immediately click back arrow | Direct navigation to Details (no confirmation, no toast) | Medium |
| Back arrow with changed receivers | Toggle Specific, click back arrow | Direct navigation to Details (no confirmation); Receiver changes discarded silently; record remains Draft with original Receiver config | High |
| Click Details sibling with changes | Toggle Specific, click Details button | Direct navigation to Details; Receiver changes discarded silently | Medium |
| Click Update sibling with changes | Toggle Specific, click Update Announcement | Direct navigation to Update form; Receiver changes discarded silently | Medium |
| Click Delete sibling with changes | Toggle Specific, click Delete Announcement | Direct navigation to Delete confirmation; Receiver changes discarded silently | Medium |
| Concurrent send conflict | User A on Send page; User B sends from Details; User A clicks Send -> confirm | Server 409 -> Error toast "This announcement has already been sent" + redirect to read-only Details | High |
| Concurrent delete conflict | User A on Send page; User B deletes; User A clicks Send -> confirm | Server 404 -> Error toast "This announcement no longer exists" + redirect to List | High |
| Direct URL to Sent announcement | Visit /announcements/:id/send when status=Sent | Redirect to Details with toast "This announcement has already been sent and can no longer be sent again" | High |
| Direct URL to deleted announcement | Visit /announcements/:id/send when soft-deleted | Empty state "Announcement not found or no longer available" + Back to List button | Medium |
| Permission revoked mid-session | User loses permission; clicks Send -> confirm | 403 from server -> toast "You no longer have permission to send this announcement" + redirect to fallback | Medium |
| Generic server error during Send | Network/server failure on the Send submission | Toast "Failed to send announcement. Please try again."; status remains Draft; user can retry | Medium |
| Double-click Send (race) | Click Send rapidly twice | Confirmation dialog opens once; on confirm, only one Send request issued (button disabled during in-flight request) | Medium |
| Skeleton loading | Slow network on first load | Skeleton for action panel + Receiver Info card until data arrives | Low |
| Active employees only in dropdown | Toggle Specific, open Employee dropdown | Dropdown lists active employees only; deactivated users not present | Medium |
| Recipient summary in dialog (Everyone) | Click Send with Everyone selected | Dialog summary reads "Everyone" | Low |
| Recipient summary in dialog (1 employee) | Click Send with 1 employee selected | Dialog summary reads "1 employee selected" | Low |
| Recipient summary in dialog (5 employees) | Click Send with 5 employees selected | Dialog summary reads "5 employees selected" | Low |
| Cancel default focus in dialog | Open confirmation dialog | "Cancel" button has default keyboard focus (not "Send") | Medium |
| Escape closes dialog | Open dialog, press Escape | Dialog closes (same as Cancel); remain on Send page | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only users with announcement management permission (configured via US-004) can access the Send Announcement page; access verified at route load and on Send submission
- **Rule 2:** Send is only allowed when the announcement is in **Draft** status — server rejects any send attempt on a Sent or soft-deleted record
- **Rule 3:** Sending transitions status from Draft -> Sent atomically with setting `last_announce_date` to the current UTC timestamp (server-generated)
- **Rule 4:** The receiver configuration submitted from this screen replaces the Draft's stored receiver configuration if changed; the final configuration is what is used for notification dispatch
- **Rule 5:** Receiver configuration is stored in the same shape as Create and Update (per DR-010-001-02 Rule 5):
  - `everyone = true` when Everyone toggle is ON (Employee list ignored on save even if populated in memory)
  - `everyone = false` + list of employee IDs when Specific one is ON
- **Rule 6:** Same Specific-one validation rule as Create form applies: at least one employee required when Specific one is ON
- **Rule 7:** Sending dispatches two notification channels asynchronously (per DR-010-001-02 Rule 10):
  - **Email notification** to each recipient's registered email address
  - **Push notification** to each recipient's mobile device via MOBILE-APP push service
- **Rule 8:** Notifications run as a background job — the UI does not block waiting for delivery; the success toast appears immediately on status update
- **Rule 9:** If a recipient has no registered email, email is skipped for that recipient (no error) — per DR-010-001-02 Rule 12
- **Rule 10:** If a recipient has push disabled on MOBILE-APP, push is skipped for that recipient (no error) — per DR-010-001-02 Rule 13
- **Rule 11:** Email subject format: "[Company Name] Announcement: {Title}" — per DR-010-001-02 Rule 14
- **Rule 12:** Push notification title: "New Announcement"; body: "{Title}" (truncated to 100 chars if needed) — per DR-010-001-02 Rule 15
- **Rule 13:** Sending is irreversible (per Knowledge Base §9.4 Announcement Status); once Sent, the record is immutable and cannot be edited or recalled. Recipients who already received the email/push retain their copy regardless of any later admin action (delete still allowed but is administrative cleanup only — per DR-010-001-03 Rule 11)
- **Rule 14:** Concurrent send protection: server verifies the announcement is still in Draft status (and not soft-deleted) before applying the transition (per Knowledge Base §4.3)
  - If status changed to Sent: return 409 Conflict; UI redirects to read-only Details with toast
  - If record was soft-deleted: return 404 Not Found; UI redirects to list with toast
- **Rule 15:** All status transitions require an explicit confirmation dialog (per Knowledge Base §3.4 Confirmation Required) — clicking the page-level Send button does not directly dispatch; it opens the confirmation dialog, and the actual Draft -> Sent transition happens only after the user confirms inside the dialog
- **Rule 16:** Generic server failure (5xx) during Send is recoverable: status transition is atomic, so partial failure leaves the announcement in Draft; the UI shows an error toast and the user can retry without data loss
- **Rule 17:** Permission verification is performed both on initial page load and on Send submission — revocation mid-session is handled gracefully (per DR-010-001-04 Rule 16)
- **Rule 18:** No Cancel/Discard button exists on this screen — the page has only one bottom action button labeled **"Send"**. Users navigate away via the back arrow or sibling action buttons (Details, Update, Delete); all such navigation is **immediate with no confirmation dialog**, and any unsaved Receiver changes are silently discarded. This is the **same deliberate deviation** from the modal/dialog Edit Form Pattern (Knowledge Base §4.1 Cancel Behavior) confirmed for Update Announcement (DR-010-001-04 Rule 13) — Send Announcement is also a standalone full-page form with no overlay context to "cancel" out of, and the explicit Send confirmation dialog provides the safety gate against accidental dispatch
- **Rule 19:** On successful send, the user is **redirected to the Announcement Details page** (DR-010-001-03), which now renders the Sent variant. This is **not** the stay-on-page deviation used by Update Announcement (DR-010-001-04 Rule 14) — for Send, the Send action is no longer applicable on the same record once the transition completes (Send button at 50% opacity), so staying on the Send page would leave the user on a now-unreachable view. Redirecting to Details lets the user verify the new Sent state and access the still-applicable Delete action
- **Rule 20:** Sibling action buttons (Details, Update Announcement, Delete Announcement) on the action panel are interactive at all times while on the Send page; clicking them performs immediate navigation to the corresponding view/flow with no confirmation, even if the Receiver Info has been modified
- **Rule 21:** Send Announcement is one of two entry points for the Draft -> Sent transition; the other is "Save & Send" on the Create Announcement form (per DR-010-001-02 §2.9 Dual-Action Save Pattern). Both paths produce identical end states (Sent record + notifications dispatched). Send Announcement is the canonical path for already-Drafted records
- **Rule 22:** Employee dropdown loads active employees only — inactive/deactivated users are excluded (per DR-010-001-02 §2.8 — Active Employees Only)
- **Rule 23:** Title, Description, and Label are read from the existing Draft record and are NOT editable on this screen; to change them, the user must navigate to Update Announcement (DR-010-001-04). The screen displays no preview of these fields (the design intent is that the user has already authored them in Create or Update, and uses Send only for audience finalization and dispatch)

**State Transitions:**

```
[Viewing Draft (Details)]            -> [Click Send Announcement]   -> [Send Form (pre-filled Receiver Info)]
[Viewing Draft (Update)]              -> [Click Send Announcement]   -> [Send Form (pre-filled Receiver Info)]
[Send Form (valid Receiver Info)]     -> [Click page-level Send button] -> [Confirmation Dialog]
[Confirmation Dialog]                 -> [Click Cancel / Escape]      -> [Back to Send Form (intact)]
[Confirmation Dialog]                 -> [Click Send]                 -> [Sending (spinner)] -> [Server: Draft -> Sent + notifications enqueued] -> [Toast + Redirect to Sent Details]
[Send Form]                           -> [Click Back arrow]           -> [Announcement Details]                    // immediate, no confirmation
[Send Form]                           -> [Click Details / Update / Delete buttons] -> [Target view/flow]           // immediate, no confirmation
[Send Form]                           -> [Click Send during Sent conflict] -> [Server 409] -> [Redirect to read-only Details + Error Toast]
[Send Form]                           -> [Click Send during Deleted conflict] -> [Server 404] -> [Redirect to List + Error Toast]
[Send Form]                           -> [Click Send during 5xx]      -> [Error Toast + remain on Send Form (Draft preserved)]
```

**Permission Model:**

| Action | Required Permission | Status Constraint |
|--------|---------------------|-------------------|
| Access Send Form | Announcement Management | Draft only |
| Click Send (open confirmation) | Announcement Management | Draft only |
| Confirm Send in dialog (Draft -> Sent transition) | Announcement Management | Draft only (server re-verifies) |

**Confirmation Dialogs:**

| Action | Dialog Title | Dialog Message | Primary Button | Cancel Button |
|--------|--------------|----------------|----------------|---------------|
| Page-level Send | "Send Announcement?" | "This announcement will be sent to the selected recipients ({recipient summary}) via email and push notification. This action cannot be undone." | "Send" (primary, dark #010101) | "Cancel" (default focus, secondary) |

**Recipient Summary in Confirmation Dialog (computed when the user clicks Send):**

| Configuration | Summary Text |
|---------------|--------------|
| Everyone ON | "Everyone" |
| Specific one ON, 1 employee selected | "1 employee selected" |
| Specific one ON, N employees selected (N > 1) | "{N} employees selected" |

**Dependencies:**

- US-004 Permission Management — provides view and management permission gates
- DR-010-001-01 Announcement List — navigation target on conflict (deleted concurrently)
- DR-010-001-02 Create Announcement — defines Receiver Info data shape, validation rules, notification dispatch behavior, and email/push subject/body formats; also provides the alternative Save & Send entry point for the Draft -> Sent transition
- DR-010-001-03 Announcement Details — entry point (Send Announcement button), back-navigation target via the back arrow / Details button, and **success-redirect destination** (Sent variant)
- DR-010-001-04 Update Announcement — alternative entry point for the Send Announcement action panel button; defines the no-Cancel-button deviation that this DR also adopts
- DR-010-001-06 Delete Announcement Confirmation (planned) — destination of Delete Announcement action from action panel
- Authentication system — user must be logged in
- Email service — for sending email notifications on Send (per DR-010-001-02 dependencies)
- Push notification service — for mobile push delivery (per DR-010-001-02 dependencies)
- Employee/User data — for Employee multi-select dropdown (active employees only)

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Two-panel layout matches Announcement Details (DR-010-001-03) and Update Announcement (DR-010-001-04) — same spatial mental model across the entity's status-managed entry points, reducing learning cost
- **Optimization 2:** Active "Send Announcement" button uses the established #f5f5f5 accent background (per Knowledge Base §3.4 Status Management on Detail Page) so users always know which view they're on
- **Optimization 3:** All sibling actions (Details, Update, Delete) remain visible and interactive in the action panel — users can pivot to a different action immediately if they realize the content also needs editing or the announcement should be removed instead of sent
- **Optimization 4:** Pre-filled Receiver Info reflects the Draft's current configuration — users see immediately who will receive the announcement, with the option to adjust last-minute
- **Optimization 5:** Receiver-only screen reduces cognitive load at the moment of dispatch — the user is not re-authoring the message, only confirming the audience
- **Optimization 6:** Explicit confirmation dialog with recipient summary provides the final visual gate against accidental sending — the user sees exactly who will receive the message before the irreversible Draft -> Sent transition
- **Optimization 7:** Confirmation dialog default keyboard focus is "Cancel" (not "Send") — prevents accidental confirmation via Enter key, aligned with destructive/irreversible action defaults (per Knowledge Base §5.2 Default Focus and DR-010-001-03 Optimization 6)
- **Optimization 8:** Inline validation error appears immediately under the Employee field when Specific one is ON without selection (not via toast) so users can correct without losing context (per Knowledge Base §2.3)
- **Optimization 9:** Success redirect to the Sent variant of Announcement Details closes the loop — the user lands on a screen that confirms the new state (Sent badge, Last Announce Date populated, Update + Send disabled at 50% opacity) without any additional navigation
- **Optimization 10:** Single bottom Send button (no header Save, no Cancel button) keeps the visual hierarchy simple. The button label "Send" matches the action it performs — eliminating the conceptual mismatch a generic "Save" label would create on a screen whose entire purpose is dispatch. Layout is otherwise consistent with Update Request Ticket (DR-003-001-06) and Update Announcement (DR-010-001-04). Navigation away is via the back arrow or sibling action buttons
- **Optimization 11:** Disabled Employee dropdown at 50% opacity (when Everyone is ON) communicates that the field is intentionally inert without hiding it — users see the full audience-selection model at all times
- **Optimization 12:** Notification dispatch happens asynchronously in the background — the user does not wait on email/push delivery; this keeps the perceived send latency low even when recipient counts are large

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full two-panel layout: 207px action panel + 600px Receiver Info card |
| Below Desktop | Out of scope — web admin is desktop-only |

**Accessibility Requirements:**

- [x] Keyboard navigable — Tab order: Back arrow > action panel buttons (top to bottom) > Receiver Info toggles (Everyone, Specific one) > Employee dropdown (when enabled) > page-level Send button
- [x] Screen reader compatible — toggle states announced ("Everyone: on/off", "Specific one: on/off"), Employee dropdown announces its enabled/disabled state, page-level Send button labeled "Send"
- [x] Sufficient color contrast — disabled toggle/dropdown states still meet WCAG AA when combined with aria-disabled
- [x] Focus indicators visible — clear focus ring on action buttons, toggles, dropdown, Send button, and dialog buttons
- [x] Confirmation dialog traps focus inside the dialog and supports Escape to cancel (per Knowledge Base §5.2)
- [x] Recipient summary in confirmation dialog announced via screen reader so non-sighted users know exactly who will receive the announcement before confirming
- [x] Error messages announced via aria-live region when validation fails

**Design References:**

- Figma Node: 3385:22326 (Send Announcement screen)
- Follows Detail/Profile View Pattern (Knowledge Base §3) and §3.4 Status Management on Detail Page (status-aware action panel, in-place navigation, confirmation required for transitions)
- Follows the no-Cancel-button variant of Edit Form Pattern (Knowledge Base §4.1 — Cancel Behavior, no-Cancel-button variant) confirmed for Update Announcement (DR-010-001-04)
- Receiver Info card model and validation per DR-010-001-02 §3 (Create Announcement)
- Page-level Send button styled per Knowledge Base §2.9 primary action style (#010101 background, white text), labeled **"Send"**
- Confirmation dialog pattern per Knowledge Base §3.4 (Confirmation Required) and DR-010-001-03 §6 (Send confirmation message wording)

---

## 8. Additional Information

### Out of Scope

- Editing Title, Description, or Label from this screen (those are managed via Update Announcement — DR-010-001-04)
- Re-sending a Sent announcement (Sent is immutable per Knowledge Base §9.4)
- Scheduled send (sending at a future date/time) — out of scope at epic level
- Per-recipient delivery status display (read receipts, email bounces) — handled by notification job, not surfaced in UI
- Retry / resend failed notifications from this screen — failures are handled by the notification job's own retry policy
- Preview the email/push body before sending — out of scope at epic level
- Bulk send (sending multiple Drafts in one action)
- Recall / unsend after a Sent announcement (announcements are immutable)
- Department-specific or role-specific targeting beyond the Everyone / Specific one model — out of scope at epic level (deferred to a future targeting enhancement)
- Mobile responsive layout (web admin desktop-only)
- Versioning / "Save a copy as new draft" from this screen
- Audit log UI showing send history (no audit trail UI in v1.0)

### Open Questions

- None — all decisions resolved against the Detail/Profile View Pattern (Knowledge Base §3 and §3.4), the Create Announcement DR (DR-010-001-02), the Announcement Details DR (DR-010-001-03), and the Update Announcement DR (DR-010-001-04 v1.1).

### Related Features

- DR-010-001-01: Announcement List (parent list view; navigation target only on delete-conflict)
- DR-010-001-02: Create Announcement (defines Receiver Info data shape, validation, notification dispatch behavior; also provides the alternative Save & Send entry point for the Draft -> Sent transition)
- DR-010-001-03: Announcement Details (entry point — Send Announcement button on action panel; success-redirect destination for the Sent variant)
- DR-010-001-04: Update Announcement (alternative entry point for the Send Announcement action panel button; defines the no-Cancel-button deviation pattern adopted here)
- DR-010-001-06: Delete Announcement Confirmation (planned — destination of Delete Announcement action from action panel)
- DR-003-001-06: Update Request Ticket (WEB-APP) — sibling Edit Form layout precedent (two-panel + single bottom Save)
- US-004: Permission Management (provides management permission source)

### Notes

- The Figma frame is named "Send Announcement" (id 3385:22326) but reuses the page title "Announcement Details" with the announcement title as breadcrumb — same header treatment as DR-010-001-03 (Details) and DR-010-001-04 (Update), keeping navigation consistent across the entity's screens.
- The bottom action button is labeled **"Send"** (this corrects the original Figma frame, which showed the same button labeled "Save"). The "Send" label is the authoritative requirement: it matches the screen's purpose — dispatching the announcement — and avoids the semantic mismatch a generic "Save" label would create on an irreversible-action screen. Implementation should follow this DR; the Figma label is treated as a design artifact to be corrected.
- **Receiver-only screen design rationale:** Unlike Update Announcement (which presents the full editable record), Send Announcement focuses exclusively on the Receiver Info card. This separates the two distinct concerns — content authoring (Create/Update) vs. audience finalization + dispatch (Send) — and reduces the cognitive load at the irreversible action. If the user realizes the content also needs editing, they can pivot to the Update Announcement sibling button.
- **No Cancel/Discard button:** This DR adopts the same deliberate deviation from the §4.1 Cancel Behavior pattern as DR-010-001-04 v1.1. The Send Announcement screen has only one bottom button labeled "Send". Sibling action buttons (Details, Update, Delete) and the back arrow are interactive at all times; clicking them performs immediate navigation with no confirmation, silently discarding any unsaved Receiver changes. The explicit Send confirmation dialog provides the safety gate against accidental dispatch — there is no need for an additional discard-unsaved-changes confirmation, because the only "unsaved" state is a Receiver Info change which is itself non-destructive (the underlying Draft is unchanged until Send is confirmed).
- **Redirect-to-Details on success (does NOT adopt the stay-on-page deviation from DR-010-001-04 v1.1):** Send Announcement intentionally does NOT use the stay-on-page success behavior introduced for Update Announcement. The reason is that on Send the action is no longer applicable on the same record after success (Send button at 50% opacity for Sent), so staying on the Send page would leave the user on a now-unreachable view. Redirecting to the Details page lets the user immediately verify the new Sent state (Sent badge, populated Last Announce Date, disabled Update + Send buttons) and access the still-applicable Delete action.
- The page-level Send button uses the dark primary style (#010101 / white text) consistent with Update Request Ticket (DR-003-001-06) and Update Announcement (DR-010-001-04) — single bottom button, no header action, no Cancel/Discard button at all.
- Send Announcement is the canonical screen for executing the Draft -> Sent transition for already-Drafted records; the alternative path is "Save & Send" on the Create Announcement form (per DR-010-001-02 §2.9). Both produce identical end states.
- Notifications fire only on Send (per DR-010-001-02 Rule 10 and DR-010-001-03 Rule 7). This screen is the only place from which notifications can be dispatched for an existing Draft; editing the Draft's content via Update Announcement (DR-010-001-04) does not trigger notifications by design.

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
| 1.0 | 2026-05-06 | Claude | Initial draft from Figma "Send Announcement" frame extraction (node 3385:22326); applies Detail/Profile View Pattern (Knowledge Base §3 and §3.4) with status-aware action panel; adopts the no-Cancel-button deviation (Knowledge Base §4.1 no-Cancel-button variant) confirmed in DR-010-001-04 v1.1; preserves Receiver Info data shape and validation from DR-010-001-02 Create Announcement; preserves notification dispatch behavior (email + push asynchronous) from DR-010-001-02 Rule 10; redirect-to-Details on success (does NOT adopt the stay-on-page Save deviation from DR-010-001-04 v1.1, because Send is no longer applicable on the same record after success). Single Save button (Figma label "Save"), explicit Send confirmation dialog with recipient summary as the irreversibility safety gate. |
| 1.1 | 2026-05-06 | Claude | Stakeholder feedback: the bottom action button label is **"Send"**, not "Save" — corrects the original Figma frame. Replaced "Save" with "Send" throughout: §1 Use Case, §2 Workflow main flow + alt flows + exit points, §3 Interaction Elements (renamed Save Button -> Send Button) + Confirmation Dialog header, §4 Layout (Save Button row -> Send Button row) + Display States ("Save Confirmation Dialog" -> "Send Confirmation Dialog") + Validation Errors, §5 ACs (AC-10/12/13/23/27 reworded; testing scenarios updated), §6 Rules (Rule 1, 15, 17, 18 reworded; State Transitions diagram updated; Permission Model rows updated; Confirmation Dialogs row label updated), §7 Optimization 10 rationale rewritten + Accessibility tab order updated, §8 Notes block rewritten — the "Figma showed Save" rationale is now flipped: the requirement is "Send" and the Figma label is treated as a design artifact to be corrected. Compound terms preserved unchanged: "Save & Send" (Create form), "Save a copy" (out-of-scope reference), Knowledge Base section names. |
