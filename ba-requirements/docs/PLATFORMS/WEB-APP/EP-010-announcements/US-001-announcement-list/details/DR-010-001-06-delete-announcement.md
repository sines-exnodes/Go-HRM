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
detail_id: DR-010-001-06
detail_name: "Delete Announcement"
# Status & Version
status: draft
version: "1.0"
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
  - path: "./DR-010-001-05-send-announcement.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
# Input sources
input_sources:
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3385:22522"
    description: "Delete Announcement screen design"
    extraction_date: 2026-05-06
---

# Detail Requirement: Delete Announcement

**Detail ID:** DR-010-001-06
**Story:** US-001-announcement-list
**Epic:** EP-010 (Announcements)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **HR Manager or Admin**, I want to **permanently remove an announcement from the active system** so that **obsolete Draft announcements are cleaned up before reaching employees, and Sent announcements that are no longer relevant are removed from administrative views**.

**Purpose:** Provide the dedicated screen for executing the destructive **soft-delete** action on an existing announcement. The screen presents a clear warning paragraph explaining the consequences of deletion (irreversibility, removal from all active views, loss of access to associated content). On Delete (the bottom action button), the system opens an explicit confirmation dialog as the irreversibility safety gate; on confirm the record is soft-deleted (preserved in the database with a `deleted` flag, removed from all active list views, filters, searches, and exports per Knowledge Base §5.1) and the user is redirected to the Announcement List. Delete is permitted for **both Draft and Sent** announcements (per Knowledge Base §9.4 — Sent records are immutable for content but admin cleanup via delete is still allowed; recipients who already received the announcement retain their copy regardless of deletion).

**Target Users:**
- HR Managers with announcement management permission
- Admins with full system access

**Key Functionality:**
- Display a clear warning explaining deletion is permanent and irreversible
- Confirm-and-execute soft-delete via the bottom **Delete** action button (with an explicit confirmation dialog as the safety gate)
- Two-panel layout consistent with Announcement Details (DR-010-001-03), Update Announcement (DR-010-001-04), and Send Announcement (DR-010-001-05)
- Status-aware sibling action panel: Update + Send at 50% opacity for Sent records (per §9.4); always-interactive for Draft records
- No Cancel/Discard button — back arrow and sibling action buttons handle navigation; no form state to discard
- Available for both Draft and Sent records (no status restriction on Delete)

---

## 2. User Workflow

**Entry Point:**
- Announcement Details page (DR-010-001-03) > Left action panel > "Delete Announcement" button (always interactive for users with management permission, regardless of Draft/Sent status)
- Update Announcement page (DR-010-001-04) > Left action panel > "Delete Announcement" button (always interactive)
- Send Announcement page (DR-010-001-05) > Left action panel > "Delete Announcement" button (always interactive)
- Direct URL navigation (e.g., bookmarked or deep-linked delete URL); access still gated by permission and existence checks

**Preconditions:**
- User is authenticated and logged in
- User has announcement management permission (via US-004 Permission Management)
- The target announcement exists and is not already soft-deleted
- The announcement may be in Draft OR Sent status (Delete is permitted in both states per Knowledge Base §9.4)

**Main Flow:**
1. User clicks "Delete Announcement" on any of the entity's screens (Details, Update, or Send action panel)
2. System navigates to the Delete Announcement page (URL pattern includes announcement ID, e.g., `/announcements/:id/delete`)
3. System displays the page header: back arrow + "Announcement Details" page title + breadcrumb showing the announcement title (e.g., "Hung King's Festival")
4. System renders the two-panel layout:
   - **Left action panel (207px):** Details, Update Announcement, Send Announcement, Delete Announcement (active with #f5f5f5 accent background)
   - **Right warning card (600px):** Card title "Delete Announcement" with the full warning paragraph: "This option permanently removes an announcement from the system. Once deleted, it will no longer be visible to users and cannot be recovered. Please proceed with caution, as this action is irreversible and will result in the loss of all content associated with the announcement. Confirm only if you are certain you want to continue."
5. System fetches the announcement record to confirm it still exists and resolves the announcement title for the breadcrumb during a brief skeleton loading state
6. System renders the sibling action panel buttons in their status-appropriate state:
   - For **Draft** announcements: Details, Update Announcement, Send Announcement all interactive
   - For **Sent** announcements: Details interactive; Update Announcement and Send Announcement at 50% opacity (non-interactive, per §9.4)
7. User reviews the warning paragraph
8. User clicks **Delete** at the bottom of the warning card (full-width 600px-wide, 40px-tall, danger red background, white "Delete" text)
9. System opens a confirmation dialog: "Delete Announcement?" with message "Are you sure you want to delete this announcement? This action cannot be undone."
10. User confirms by clicking "Delete" (danger red button) in the dialog or backs out by clicking "Cancel" (default keyboard focus)
11. On confirm, system sends the soft-delete request to the server
12. Server marks the record as soft-deleted (sets `deleted = true` flag and `deleted_at` timestamp), removes it from all active list views, filters, searches, and exports
13. Server returns success; system shows toast "Announcement has been deleted" and **redirects to the Announcement List** (DR-010-001-01) with the previously preserved list state (search, filter, page) restored

**Alternative Flows:**
- **Alt 1 - Cancel Confirmation Dialog:** User clicks Delete > confirmation dialog opens > user clicks "Cancel" (or presses Escape) > dialog closes, no deletion occurs, user remains on the Delete Announcement page
- **Alt 2 - Back-Navigation:** User clicks the back arrow at any time > immediate navigation to the Announcement Details page; no confirmation needed (the warning page is informational only — there is no form state to discard)
- **Alt 3 - Click Sibling Action Button (Details):** User clicks "Details" on the action panel > immediate navigation to the read-only Details view
- **Alt 4 - Click Sibling Action Button (Update or Send) on Draft:** User clicks "Update Announcement" or "Send Announcement" (both interactive only when status = Draft) > immediate navigation to the corresponding flow
- **Alt 5 - Click Sibling Action Button (Update or Send) on Sent:** Buttons render at 50% opacity and are non-interactive; clicks have no effect (per §9.4)
- **Alt 6 - Concurrent Delete:** While on the Delete page, another user deletes the same announcement first; on Delete confirm, server returns 404; system shows informational toast "This announcement has already been deleted" and redirects to the Announcement List (the desired outcome — record gone — is achieved either way)
- **Alt 7 - Concurrent Status Change (Draft -> Sent during delete flow):** While viewing the Delete page for a Draft, another user sends the announcement; the page does not auto-refresh, but Delete remains valid (Delete is allowed for Sent too). On Delete confirm, the record is soft-deleted regardless of the Sent transition that just happened — server enforces idempotency on the delete action
- **Alt 8 - Permission Lost Mid-Session:** User's management permission is revoked while on the Delete page; on Delete confirm, server returns 403; system shows toast "You no longer have permission to delete this announcement" and redirects to a fallback page
- **Alt 9 - Direct URL to Already-Deleted Announcement:** User opens the Delete URL of a record that was already soft-deleted; system blocks load and shows empty state "Announcement not found or no longer available" with a "Back to List" button
- **Alt 10 - Direct URL to Invalid Announcement ID:** Empty state "Announcement not found or no longer available" + "Back to List" button (same as Alt 9)
- **Alt 11 - Server Error During Delete:** Network or server failure on the Delete submission > error toast "Failed to delete announcement. Please try again." > confirmation dialog stays open with the Delete button re-enabled; the announcement remains undeleted; user can retry from the dialog or click Cancel to dismiss

**Exit Points:**
- **Success:** Toast "Announcement has been deleted" + redirect to Announcement List (DR-010-001-01) with preserved list state; record is soft-deleted (preserved in database for audit, hidden from active views)
- **Cancel Confirmation Dialog:** Dialog closes, no deletion, user remains on the Delete Announcement page
- **Back / Sibling Action Navigation:** Immediate navigation to the chosen target — no confirmation needed (no form state to discard)
- **Server Error:** Error toast in dialog; record remains undeleted; user can retry or cancel
- **Concurrent Delete (already deleted):** Informational toast + redirect to list (desired end state already achieved)

---

## 3. Field Definitions

### Input Fields

This screen has no editable input fields. It is a **destructive action confirmation screen** — the only "input" is the user's intent to proceed, captured via the bottom Delete button + the explicit confirmation dialog.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back Arrow | Icon Button | Always enabled | Navigate to Announcement Details immediately (no confirmation, no state to discard) | Top-left of header (24x24 ArrowLeft icon) |
| Details Button | Tertiary Button | Always interactive | Navigate to Announcement Details immediately | First button in action panel |
| Update Announcement Button | Tertiary Button | Interactive only when status = Draft; at 50% opacity for Sent | Navigate to Update Announcement form (Draft only); no-op for Sent (per §9.4) | Second button in action panel |
| Send Announcement Button | Tertiary Button | Interactive only when status = Draft; at 50% opacity for Sent | Navigate to Send Announcement form (Draft only); no-op for Sent (per §9.4) | Third button in action panel |
| Delete Announcement Button (action panel) | Tertiary Button (danger) | Active (current view) | No-op (visual indicator only) | Highlighted with #f5f5f5 accent background |
| Delete Button (page-level, bottom) | Primary Button (danger) | Always enabled (the screen has no validation gate; user intent is the only requirement) | Open Delete confirmation dialog | Full-width 600px, 40px tall, danger red background, white **"Delete"** text |

### Status-Aware Action Panel (Delete view)

| Announcement Status | Details | Update Announcement | Send Announcement | Delete Announcement |
|---------------------|---------|---------------------|-------------------|---------------------|
| Draft | Interactive | Interactive | Interactive | Active (current view) |
| Sent | Interactive | Disabled (50% opacity) | Disabled (50% opacity) | Active (current view) |

**Note:** Delete is the only action available for both Draft and Sent announcements (per Knowledge Base §9.4). When opening the Delete page from a Sent announcement, the Update and Send sibling buttons render at 50% opacity to communicate that those actions are no longer applicable, while Delete remains permitted as administrative cleanup.

### Confirmation Dialog (triggered by clicking the page-level Delete button)

| Element | Detail |
|---------|--------|
| Dialog Title | "Delete Announcement?" |
| Dialog Message | "Are you sure you want to delete this announcement? This action cannot be undone." |
| Dialog Summary (optional context) | Display the announcement title (e.g., "Hung King's Festival") so the user has one final visual confirmation of which record they are about to delete |
| Primary Button | "Delete" (danger red background, white text) |
| Cancel Button | "Cancel" (default keyboard focus, secondary style) |
| Escape Key | Closes dialog (same as Cancel) |
| Focus Trap | Keyboard focus trapped inside the dialog while open |

**Note on the page-level Delete button label:** Per Knowledge Base §3.6 (generalized rule from DR-010-001-05 v1.1), the primary action button label SHOULD match the transition verb. This screen's transition verb is "Delete", so the bottom button is labeled **"Delete"** — not "Save". The Figma frame may show this button labeled as "Save"; that label is treated as a design artifact to be corrected. The DR is the authoritative source: the visible label must be "Delete".

---

## 4. Data Display

### Layout Structure

| Region | Position | Dimensions | Content |
|--------|----------|------------|---------|
| Header Row | Top | Full width, 29px tall | Back arrow + "Announcement Details" title + breadcrumb separator + announcement title |
| Action Panel | Left | 207px wide, 192px tall (4 buttons x 36px + gaps) | Vertical stack: Details, Update Announcement, Send Announcement, Delete Announcement (active) |
| Warning Card | Right | 600px wide, ~194px tall | Card title "Delete Announcement" + warning paragraph (114px tall) |
| Delete Button | Below warning card | 600px wide, 40px tall | Single full-width Delete button (danger red background, white **"Delete"** text) |

### Warning Card Content (from Figma)

| Card | Title | Body |
|------|-------|------|
| Warning Card | Delete Announcement | "This option permanently removes an announcement from the system. Once deleted, it will no longer be visible to users and cannot be recovered. Please proceed with caution, as this action is irreversible and will result in the loss of all content associated with the announcement. Confirm only if you are certain you want to continue." |

**Note on the Figma design:** The Delete Announcement Figma frame focuses exclusively on the warning paragraph. There is **no preview of the announcement's content** (Title, Description, Last Announce Date, Recipients, Status) on this screen — the deliberate UX choice is that by the time the user reaches Delete, they have already viewed those details on the Details page (DR-010-001-03) or are deleting based on context (e.g., from the list gear menu via Details). To preview the announcement before deleting, the user navigates to the Details sibling button. The breadcrumb showing the announcement title in the header gives a minimal visual confirmation of which record is targeted, and the optional dialog summary (announcement title) provides a second confirmation gate before deletion.

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch | Skeleton placeholders for action panel buttons and warning card |
| Loaded (Draft) | Announcement loaded with status = Draft | All four action panel buttons interactive (Update + Send enabled); warning card rendered; Delete button enabled |
| Loaded (Sent) | Announcement loaded with status = Sent | Details + Delete interactive; Update + Send at 50% opacity; warning card rendered; Delete button enabled |
| Delete Confirmation Dialog | User clicked the page-level Delete button | Modal dialog "Delete Announcement?" with announcement title summary, Delete (danger red) and Cancel buttons (Cancel has default focus) |
| Deleting | User confirmed in dialog | Dialog "Delete" button shows loading spinner, disabled to prevent double-submit; page-level Delete button also disabled |
| Success | Server returned success | Toast "Announcement has been deleted" + redirect to Announcement List (with preserved list state) |
| Already Deleted (Concurrent) | Server returns 404 (someone else deleted it first) | Informational toast "This announcement has already been deleted" + redirect to Announcement List (the desired end state — record gone — is achieved) |
| Server Error | Generic 5xx from server | Error toast "Failed to delete announcement. Please try again." within the dialog; dialog stays open; Delete button re-enabled for retry; record remains undeleted |
| Permission Denied | User loses management permission | Toast "You no longer have permission to delete this announcement" + redirect to fallback |
| Not Found | Direct URL to invalid/already-deleted ID on initial load | Empty state "Announcement not found or no longer available" + "Back to List" button |

### Validation

There is no user-input validation on this screen. The page-level Delete button is always enabled when the page is loaded with a valid (non-deleted) announcement. The single safety gate is the explicit confirmation dialog — Delete cannot proceed without an explicit "Delete" confirmation click in the dialog (with default focus on Cancel to prevent accidental Enter-key confirmation).

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Users with announcement management permission can access the Delete Announcement page; users without permission are redirected to a fallback page with permission-denied toast
- **AC-02:** Delete Announcement is reachable for announcements in **both Draft and Sent** statuses (per Knowledge Base §9.4) — there is no status restriction on the Delete action
- **AC-03:** Page renders the two-panel layout: 207px-wide action panel on the left, 600px-wide warning card on the right
- **AC-04:** Header shows back arrow + "Announcement Details" title + breadcrumb displaying the announcement title (e.g., "Hung King's Festival")
- **AC-05:** Action panel renders four buttons in order: Details, Update Announcement, Send Announcement, Delete Announcement (active with #f5f5f5 accent background)
- **AC-06:** For Draft records, all sibling action panel buttons (Details, Update Announcement, Send Announcement) are interactive
- **AC-07:** For Sent records, Update Announcement and Send Announcement render at 50% opacity and are non-interactive; Details remains interactive (per Knowledge Base §9.4 Action Availability matrix)
- **AC-08:** The warning card displays the title "Delete Announcement" and the full warning paragraph: "This option permanently removes an announcement from the system. Once deleted, it will no longer be visible to users and cannot be recovered. Please proceed with caution, as this action is irreversible and will result in the loss of all content associated with the announcement. Confirm only if you are certain you want to continue."
- **AC-09:** No preview of the announcement's content (Title, Description, Last Announce Date, Recipients, Status) is displayed on this screen — the breadcrumb showing the announcement title is the only inline reference to the targeted record
- **AC-10:** The page-level Delete button is full-width 600px, 40px tall, with danger red background and white **"Delete"** text (the visible label is "Delete", not "Save", per the §3.6 generalized rule from DR-010-001-05 v1.1)
- **AC-11:** Clicking the page-level Delete button opens a confirmation dialog "Delete Announcement?" with message "Are you sure you want to delete this announcement? This action cannot be undone."
- **AC-12:** The confirmation dialog includes the announcement title (as a summary) so the user has one final visual confirmation of which record will be deleted
- **AC-13:** The confirmation dialog's default keyboard focus is on the **Cancel** button (not Delete) — prevents accidental confirmation via Enter key (per Knowledge Base §5.2)
- **AC-14:** The confirmation dialog's primary "Delete" button uses the danger red style; "Cancel" uses the secondary style
- **AC-15:** Escape key closes the confirmation dialog (same as clicking Cancel)
- **AC-16:** Keyboard focus is trapped inside the confirmation dialog while it is open (per Knowledge Base §5.2)
- **AC-17:** Confirming "Delete" in the dialog soft-deletes the record (sets `deleted = true` flag and `deleted_at` timestamp on the server; record remains in the database for audit purposes per Knowledge Base §5.1)
- **AC-18:** Soft-deleted announcements are removed from all active list views, filters, searches, and exports (per Knowledge Base §5.1 — Soft Delete Transactional)
- **AC-19:** On successful delete, system shows toast "Announcement has been deleted" and **redirects to the Announcement List** (DR-010-001-01) with the previously preserved list state (search, filter, page) restored
- **AC-20:** Cancelling the confirmation dialog closes the dialog without any deletion; user remains on the Delete Announcement page
- **AC-21:** Clicking the back arrow at any time navigates immediately to the Announcement Details page; no confirmation dialog (no form state to discard)
- **AC-22:** Clicking sibling action buttons (Details, Update Announcement when Draft, Send Announcement when Draft) at any time navigates immediately to the corresponding view/flow; no confirmation
- **AC-23:** Loading state shows skeleton placeholders for the action panel buttons and the warning card
- **AC-24:** If the announcement is already deleted by another user during the session (server returns 404 on Delete confirm), system shows informational toast "This announcement has already been deleted" and redirects to the Announcement List (the desired end state is reached either way — this is graceful idempotency, not an error)
- **AC-25:** If a generic server error (5xx) occurs during Delete, system shows error toast "Failed to delete announcement. Please try again." within the dialog; the dialog stays open with the Delete button re-enabled for retry; the announcement remains undeleted
- **AC-26:** Direct URL access to a deleted or invalid announcement ID shows empty state "Announcement not found or no longer available" with a "Back to List" button
- **AC-27:** Deleting a Sent announcement is administrative cleanup only — recipients who already received the email/push notifications retain their copy regardless of the deletion (per Knowledge Base §9.4 and DR-010-001-03 Rule 11)
- **AC-28:** Permission verification is performed both on initial page load and on Delete submission — revocation mid-session is handled gracefully with a toast and redirect

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Open Delete for Draft announcement | Click Delete Announcement on Details page (Draft) | Page renders with all 4 sibling buttons interactive; warning card visible; Delete button enabled | High |
| Open Delete for Sent announcement | Click Delete Announcement on Details page (Sent) | Page renders with Details + Delete interactive; Update + Send at 50% opacity; warning card visible; Delete button enabled | High |
| Open Delete from Update page (Draft) | Click Delete Announcement on Update page action panel | Direct navigation to Delete page with Draft sibling state | High |
| Open Delete from Send page (Draft) | Click Delete Announcement on Send page action panel | Direct navigation to Delete page with Draft sibling state | High |
| View warning paragraph | Open Delete page | Full warning paragraph (114px) renders inside the warning card with title "Delete Announcement" | High |
| Click Delete (Draft) | Click page-level Delete button | Confirmation dialog opens with title "Delete Announcement?", message, announcement title summary, Cancel (default focus) and Delete (danger red) | High |
| Click Delete (Sent) | Click page-level Delete button on Sent record | Same confirmation dialog opens — Delete is allowed for Sent | High |
| Confirm delete (Draft) | Open dialog > click Delete | Record soft-deleted, toast "Announcement has been deleted", redirect to Announcement List | High |
| Confirm delete (Sent) | Open dialog > click Delete | Record soft-deleted, toast "Announcement has been deleted", redirect to Announcement List; recipients retain their copy | High |
| Cancel confirmation dialog | Open dialog > click Cancel | Dialog closes, no deletion, remain on Delete page | High |
| Press Escape in dialog | Open dialog > press Escape | Dialog closes (same as Cancel), no deletion | High |
| Default focus on Cancel | Open confirmation dialog | Cancel button has default keyboard focus (not Delete) | High |
| Pressing Enter without focusing Delete | Open dialog, do not move focus, press Enter | Cancel triggers (default focus) — no deletion occurs (intentional safety gate) | High |
| List state preserved on success | Apply filters on list, click row, click Delete > confirm | Returns to list with same filters/page applied; deleted row no longer visible | High |
| Soft-delete removes from active views | Confirm delete, return to list | Deleted record absent from list, search results, filter results, and exports | High |
| Soft-delete preserves audit | Database check after delete | Record remains in DB with `deleted = true` and `deleted_at` timestamp set | Medium |
| Back arrow with no action | Open Delete page, click back arrow | Direct navigation to Details (no confirmation, no toast) | Medium |
| Click Details sibling | Open Delete page (Draft), click Details | Direct navigation to read-only Details | Medium |
| Click Update sibling (Draft) | Open Delete page (Draft), click Update Announcement | Direct navigation to Update form | Medium |
| Click Send sibling (Draft) | Open Delete page (Draft), click Send Announcement | Direct navigation to Send Announcement page | Medium |
| Click Update sibling (Sent) | Open Delete page (Sent), click Update Announcement (50% opacity) | Button non-interactive — no navigation occurs | Medium |
| Click Send sibling (Sent) | Open Delete page (Sent), click Send Announcement (50% opacity) | Button non-interactive — no navigation occurs | Medium |
| Concurrent delete (already deleted) | User A on Delete page; User B deletes; User A clicks Delete > confirm | Server 404 -> informational toast "This announcement has already been deleted" + redirect to list | High |
| Concurrent send during delete flow (Draft -> Sent) | User A viewing Delete (Draft); User B sends; User A clicks Delete > confirm | Record soft-deleted regardless of the Sent transition (Delete is allowed for Sent too); toast + redirect to list | Medium |
| Direct URL to already-deleted announcement | Visit /announcements/:id/delete when soft-deleted | Empty state "Announcement not found or no longer available" + Back to List button | Medium |
| Direct URL to invalid announcement ID | Visit /announcements/99999/delete | Empty state "Announcement not found or no longer available" + Back to List button | Medium |
| Permission revoked mid-session | User loses permission; clicks Delete > confirm | 403 from server -> toast "You no longer have permission to delete this announcement" + redirect to fallback | Medium |
| Generic server error during Delete | Network/server failure on the Delete submission | Toast "Failed to delete announcement. Please try again." inside the dialog; dialog stays open; Delete button re-enabled; record undeleted | Medium |
| Double-click Delete (race) | Click page-level Delete rapidly twice | Confirmation dialog opens once; on confirm, only one Delete request issued (button disabled during in-flight request) | Medium |
| Skeleton loading | Slow network on first load | Skeleton for action panel + warning card until data arrives | Low |
| Announcement title in dialog summary | Open dialog | Dialog body includes the announcement title (e.g., "Hung King's Festival") for one final visual confirmation | Medium |
| Page-level Delete button label | Inspect button | Visible label is "Delete" — not "Save" (per §3.6 generalized rule) | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only users with announcement management permission (configured via US-004) can access the Delete Announcement page; access is verified at route load and on Delete submission
- **Rule 2:** Delete is allowed for announcements in **both Draft and Sent** statuses (per Knowledge Base §9.4) — there is no status-based restriction; the only access control is the management permission
- **Rule 3:** Deletion is **soft-delete** (per Knowledge Base §5.1 Soft Delete Transactional): the record is preserved in the database with a `deleted = true` flag and a `deleted_at` UTC timestamp; the record is not physically removed
- **Rule 4:** Soft-deleted announcements are excluded from all active views: list views, search results, filter results, exports, and any cross-platform consumers (e.g., MOBILE-APP announcement widget per §17 — only Sent + non-deleted records reach mobile)
- **Rule 5:** Deleting a Sent announcement is **administrative cleanup only** — recipients who already received the email and push notifications retain their copy regardless of the deletion (per Knowledge Base §9.4 and DR-010-001-03 Rule 11). The deletion does not retract or recall the notifications already dispatched
- **Rule 6:** The Delete action is **destructive and irreversible from the UI perspective** — there is no Undo / Restore flow available to users in v1.0. Audit / restoration would require database-level access by a super-admin (out of scope at the UI layer)
- **Rule 7:** All Delete actions require an explicit confirmation dialog as the irreversibility safety gate (per Knowledge Base §3.4 Confirmation Required and §5 Delete Action Pattern) — clicking the page-level Delete button does not directly delete; it opens the confirmation dialog, and the actual soft-delete happens only after the user confirms inside the dialog
- **Rule 8:** The confirmation dialog's default keyboard focus is on **Cancel** (not Delete) — prevents accidental destructive confirmation via the Enter key (per Knowledge Base §5.2 Default Focus on destructive actions)
- **Rule 9:** Concurrent delete protection (idempotent): if the announcement is already soft-deleted by another user when the current user confirms Delete, the server returns 404; the UI shows an informational toast ("This announcement has already been deleted") and redirects to the Announcement List. This is **not treated as an error** — the desired end state (record gone from active views) is achieved either way
- **Rule 10:** Concurrent status change (Draft -> Sent during the delete flow) does not block deletion — Delete is permitted for both Draft and Sent (per Rule 2). The server applies the soft-delete idempotently regardless of the latest status
- **Rule 11:** Generic server failure (5xx) during Delete is recoverable: the soft-delete operation is atomic, so partial failure leaves the announcement undeleted; the UI shows an error toast inside the still-open confirmation dialog and re-enables the Delete button so the user can retry without re-opening the dialog
- **Rule 12:** Permission verification is performed both on initial page load and on Delete submission — revocation mid-session is handled gracefully with a 403 response, a permission-denied toast, and a redirect to a fallback page
- **Rule 13:** No Cancel/Discard button exists on this screen — the page has only one bottom action button labeled **"Delete"**. Users navigate away via the back arrow or sibling action buttons (Details, Update, Send); all such navigation is **immediate with no confirmation dialog**, because the warning page has no editable form state to "discard". This is the **same standalone-screen layout** as Send Announcement (DR-010-001-05) and Update Announcement (DR-010-001-04 v1.1), and consistent with the §3.6 Dedicated Status-Transition Screen Pattern
- **Rule 14:** On successful soft-delete, the user is **redirected to the Announcement List** (DR-010-001-01) — not to the Details page, because the deleted record is no longer accessible in active views. This redirect destination differs from Send Announcement (DR-010-001-05 Rule 19), which redirects to Details because the Sent record still exists. Delete redirects to List because the record is gone from active views
- **Rule 15:** The page-level Delete button label is **"Delete"** — it matches the transition verb per Knowledge Base §3.6 generalized rule from DR-010-001-05 v1.1 (the primary action button label SHOULD match the transition verb; generic labels like "Save" create a semantic mismatch on a single-purpose transition screen). The Figma frame may render the button labeled "Save"; that is treated as a design artifact to be corrected, with the DR being the authoritative source
- **Rule 16:** The page-level Delete button uses the **danger red style** (not the standard #010101 dark used by Update Save and Send) — destructive actions are visually distinct from neutral commit actions to reinforce the gravity of the operation (per Knowledge Base §5 Delete Action Pattern, danger style)
- **Rule 17:** The confirmation dialog's "Delete" button also uses the danger red style; "Cancel" uses the secondary style; default focus is on Cancel
- **Rule 18:** Sibling action buttons (Details, Update Announcement, Send Announcement) on the action panel render their status-appropriate state at all times while on the Delete page:
  - For **Draft** records: all three sibling buttons are interactive
  - For **Sent** records: Details is interactive; Update Announcement and Send Announcement are at 50% opacity and non-interactive (per Knowledge Base §9.4 Action Availability matrix)
- **Rule 19:** Direct URL guard: if a user opens the Delete URL of an already-soft-deleted record (or an invalid ID), the system shows an empty state "Announcement not found or no longer available" with a "Back to List" button — does NOT open the warning page or the confirmation dialog
- **Rule 20:** No notifications are dispatched on delete — deletion is silent from a recipient's perspective (no email or push informing recipients that an announcement was deleted; recipients of Sent announcements simply retain their previously delivered copy)
- **Rule 21:** List state preservation: on successful redirect to the Announcement List, the previously preserved list state (search query, filter selections, current page) is restored — consistent with Knowledge Base §3.3 State Preservation
- **Rule 22:** No audit log UI is exposed in v1.0 — the soft-delete `deleted_at` timestamp and acting user are persisted in the database for future audit features, but there is no in-app surface to view the deletion history

**State Transitions:**

```
[Viewing Draft (Details / Update / Send)] -> [Click Delete Announcement]   -> [Delete Page (warning card visible)]
[Viewing Sent (Details)]                   -> [Click Delete Announcement]   -> [Delete Page (warning card visible, Update + Send at 50%)]
[Delete Page]                              -> [Click page-level Delete]    -> [Confirmation Dialog]
[Confirmation Dialog]                      -> [Click Cancel / Escape]      -> [Back to Delete Page (intact)]
[Confirmation Dialog]                      -> [Click Delete]               -> [Deleting (spinner)] -> [Server: soft-delete + remove from active views] -> [Toast + Redirect to Announcement List]
[Delete Page]                              -> [Click Back arrow]           -> [Announcement Details]                            // immediate, no confirmation
[Delete Page]                              -> [Click Details / Update (Draft) / Send (Draft) buttons] -> [Target view/flow]      // immediate, no confirmation
[Delete Page]                              -> [Click Delete during already-deleted (404)] -> [Informational Toast + Redirect to List]
[Delete Page]                              -> [Click Delete during 5xx]    -> [Error Toast in dialog; dialog stays open; Delete button re-enabled (record undeleted)]
[Delete Page]                              -> [Click Delete during 403]    -> [Toast + Redirect to fallback]
```

**Permission Model:**

| Action | Required Permission | Status Constraint |
|--------|---------------------|-------------------|
| Access Delete Page | Announcement Management | Any non-deleted status (Draft OR Sent) |
| Click Delete (open confirmation) | Announcement Management | Any non-deleted status |
| Confirm Delete in dialog (soft-delete) | Announcement Management | Any non-deleted status (server re-verifies) |

**Confirmation Dialogs:**

| Action | Dialog Title | Dialog Message | Primary Button | Cancel Button |
|--------|--------------|----------------|----------------|---------------|
| Page-level Delete | "Delete Announcement?" | "Are you sure you want to delete this announcement? This action cannot be undone." (with announcement title summary, e.g., "Hung King's Festival") | "Delete" (danger red) | "Cancel" (default focus, secondary) |

**Dependencies:**

- US-004 Permission Management — provides view and management permission gates
- DR-010-001-01 Announcement List — **success-redirect destination** after deletion + back-navigation target via the list state
- DR-010-001-03 Announcement Details — primary entry point (Delete Announcement button on action panel) and back-navigation target via the back arrow / Details button
- DR-010-001-04 Update Announcement — alternative entry point (Delete Announcement button on its action panel)
- DR-010-001-05 Send Announcement — alternative entry point (Delete Announcement button on its action panel)
- Authentication system — user must be logged in
- Knowledge Base §5.1 Soft Delete Transactional — defines the `deleted` flag + `deleted_at` timestamp persistence model
- Knowledge Base §3.6 Dedicated Status-Transition Screen Pattern — establishes the screen layout, no-Cancel-button deviation, primary-action-button-label-matches-verb rule, and confirmation-dialog safety gate
- Knowledge Base §9.4 Announcement Status — establishes that Delete is permitted in both Draft and Sent states

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Two-panel layout matches Announcement Details (DR-010-001-03), Update Announcement (DR-010-001-04), and Send Announcement (DR-010-001-05) — same spatial mental model across the entity's status-managed entry points, reducing learning cost (per Knowledge Base §3.6)
- **Optimization 2:** Active "Delete Announcement" button uses the established #f5f5f5 accent background (per Knowledge Base §3.4 Status Management on Detail Page) so users always know which view they're on
- **Optimization 3:** All sibling actions (Details, Update Announcement when Draft, Send Announcement when Draft) remain visible and interactive in the action panel — users can pivot to a different action immediately if they realize they came to the wrong screen and don't actually want to delete
- **Optimization 4:** The full warning paragraph is displayed prominently in the warning card (114px tall, 600px wide) — the user reads it before reaching the bottom Delete button. The wording explicitly mentions irreversibility ("cannot be recovered", "permanently removes", "this action is irreversible") and consequences ("loss of all content associated with the announcement") so the user understands the gravity before proceeding
- **Optimization 5:** Two-step confirmation: the page-level Delete button does NOT directly delete — it opens an explicit confirmation dialog. The dialog provides a second safety gate with its own destructive-action default focus (Cancel) — the user must take two deliberate actions to delete a record (page button + dialog confirm)
- **Optimization 6:** Confirmation dialog default keyboard focus is "Cancel" (not "Delete") — prevents accidental confirmation via the Enter key, aligned with destructive action defaults (per Knowledge Base §5.2 and DR-010-001-05 Optimization 7)
- **Optimization 7:** Page-level Delete button and dialog Delete button both use **danger red** styling — distinct from the neutral #010101 dark style used by Send (DR-010-001-05) and Update Save (DR-010-001-04). The visual contrast reinforces the gravity of the destructive action and helps prevent the user from confusing it with the dispatch action on the sibling Send screen
- **Optimization 8:** Delete button label is **"Delete"** (matches the transition verb) — eliminates the conceptual mismatch a generic "Save" label would create on a destructive screen, and aligns with the §3.6 generalized rule from DR-010-001-05 v1.1
- **Optimization 9:** No editable form state means no "discard unsaved changes" confirmation is needed when navigating away — the back arrow and sibling action buttons can navigate immediately without friction. This is a deliberate simplification: the warning page is pure information + a single action gate
- **Optimization 10:** Dialog summary includes the announcement title (e.g., "Hung King's Festival") — gives the user one final visual confirmation of which record they're about to delete, especially important if the user navigated to this screen via direct URL or a bookmark
- **Optimization 11:** Concurrent-delete idempotency (already deleted -> informational toast, not an error) treats the second delete attempt gracefully — the user's intent (record removed from active views) is satisfied either way, so surfacing a hard error would be confusing and unnecessary
- **Optimization 12:** Status-aware sibling buttons (Update + Send at 50% opacity for Sent records) communicate the limited remaining action set without hiding the buttons — the user sees the full action panel at all times and understands which actions are no longer applicable for this record's lifecycle stage

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full two-panel layout: 207px action panel + 600px warning card |
| Below Desktop | Out of scope — web admin is desktop-only |

**Accessibility Requirements:**

- [x] Keyboard navigable — Tab order: Back arrow > action panel buttons (top to bottom) > page-level Delete button
- [x] Screen reader compatible — warning paragraph announced as static text; page-level Delete button labeled "Delete"; sibling action button states (interactive vs disabled) announced via aria-disabled
- [x] Sufficient color contrast — danger red Delete button meets WCAG AA against white text; disabled sibling buttons (50% opacity) still meet WCAG AA when combined with aria-disabled
- [x] Focus indicators visible — clear focus ring on action buttons, page-level Delete button, and dialog buttons
- [x] Confirmation dialog traps focus inside the dialog and supports Escape to cancel (per Knowledge Base §5.2)
- [x] Dialog default focus on Cancel announced via screen reader so non-sighted users land on the safe choice
- [x] Announcement title in the dialog summary announced via screen reader so non-sighted users know exactly which record is targeted before confirming
- [x] Toast messages (success, error, already-deleted, permission-denied) announced via aria-live regions
- [x] Disabled sibling buttons for Sent records announce their disabled state to screen readers (aria-disabled)

**Design References:**

- Figma Node: 3385:22522 (Delete Announcement screen)
- Follows Detail/Profile View Pattern (Knowledge Base §3) and §3.4 Status Management on Detail Page (status-aware action panel, in-place navigation, confirmation required for transitions)
- Follows the §3.6 Dedicated Status-Transition Screen Pattern (single primary action button, primary-action-button-label-matches-verb rule, no Cancel/Discard button, confirmation dialog as safety gate, redirect-on-success) confirmed in DR-010-001-05
- Follows §5 Delete Action Pattern + §5.1 Soft Delete Transactional + §5.2 Transactional Delete Pattern (default focus on Cancel, danger red Delete button style, focus trap, escape closes dialog)
- Page-level Delete button styled in danger red (per §5 Delete Action Pattern), labeled **"Delete"** (per §3.6 generalized rule)
- Confirmation dialog wording aligned with DR-010-001-03 §6 Confirmation Dialogs Delete row

---

## 8. Additional Information

### Out of Scope

- Restoring (undeleting) a soft-deleted announcement from the UI — there is no Restore / Undo flow in v1.0; soft-deleted records can only be restored via direct database access by a super-admin
- Audit log UI showing deletion history (no in-app surface to view who deleted what and when in v1.0)
- Bulk delete (multi-select delete on the list) — out of scope at story level
- Hard delete / permanent purge of soft-deleted records — out of scope at story level (would be a separate retention policy / data lifecycle feature)
- Recall / unsend notifications when a Sent announcement is deleted — recipients retain their delivered copy regardless (per §9.4)
- Post-delete notification to recipients ("This announcement has been removed") — out of scope; deletion is silent from the recipient's perspective
- "Reason for deletion" field or audit comment — out of scope at story level
- Confirmation dialog input ("Type DELETE to confirm") for extra-strict deletion gating — out of scope; the default-focus-on-Cancel + danger button styling is sufficient for v1.0
- Editing the warning paragraph wording per-tenant or per-language — out of scope; the warning paragraph is a static string per Figma
- Mobile responsive layout (web admin desktop-only)

### Open Questions

- None — all decisions resolved against the Detail/Profile View Pattern (Knowledge Base §3 and §3.4), the Dedicated Status-Transition Screen Pattern (§3.6), the Delete Action Pattern (§5), the Announcement Status rules (§9.4), and the established sibling DRs (DR-010-001-03, DR-010-001-04 v1.1, DR-010-001-05 v1.1).

### Related Features

- DR-010-001-01: Announcement List (parent list view; **success-redirect destination** after deletion)
- DR-010-001-02: Create Announcement (defines the entity that this DR deletes; not a direct entry point)
- DR-010-001-03: Announcement Details (primary entry point — Delete Announcement button on action panel; back-navigation target via back arrow / Details sibling button)
- DR-010-001-04: Update Announcement (alternative entry point — Delete Announcement button on its action panel; defines the no-Cancel-button deviation pattern this DR adopts)
- DR-010-001-05: Send Announcement (alternative entry point; defines the §3.6 Dedicated Status-Transition Screen Pattern this DR follows; sets the §3.6 generalized rule that the primary action button label SHOULD match the transition verb — "Delete" here, not "Save")
- DR-003-001-04: Delete Request Ticket (WEB-APP) — sibling Soft Delete Transactional pattern precedent (per Knowledge Base §5.2)
- DR-002-001-04: Delete Leave Request (WEB-APP) — sibling Soft Delete Transactional pattern precedent (per Knowledge Base §5.2)
- US-004: Permission Management (provides management permission source)

### Notes

- The Figma frame is named "Delete Announcement" (id 3385:22522) but reuses the page title "Announcement Details" with the announcement title as breadcrumb — same header treatment as DR-010-001-03 (Details), DR-010-001-04 (Update), and DR-010-001-05 (Send), keeping navigation consistent across the entity's screens.
- The bottom action button is labeled **"Delete"** (this corrects the original Figma frame, which may show the same button labeled "Save"). The "Delete" label is the authoritative requirement: it matches the screen's purpose — soft-deleting the announcement — and avoids the semantic mismatch a generic "Save" label would create on a destructive-action screen. Implementation should follow this DR; the Figma label is treated as a design artifact to be corrected. This rule is established in §3.6 (generalized in DR-010-001-05 v1.1).
- **Warning-only screen design rationale:** Unlike Send Announcement (which presents the Receiver Info card for last-mile audience confirmation), Delete Announcement focuses exclusively on the warning paragraph. There are no editable fields and no preview of the announcement's content on this screen. This is the deliberate UX choice for the destructive action: by the time the user reaches Delete, they have already viewed the announcement on the Details page (or are deleting from a list-level gear menu via Details). The screen's purpose is to surface the consequences of deletion clearly, give the user a moment to reconsider via the breadcrumb-shown title and the warning paragraph, and then capture an intentional Delete confirmation through the bottom button + the explicit confirmation dialog.
- **Two-step confirmation:** Delete cannot be triggered with a single click — the page-level Delete button opens a confirmation dialog, and the actual soft-delete only happens after the user clicks "Delete" inside the dialog. The dialog also defaults focus to "Cancel" (per §5.2), so even pressing Enter without intentional focus would NOT delete the record. This intentional friction is appropriate for an irreversible destructive action.
- **Danger red button style:** Both the page-level Delete button and the dialog's Delete button use the danger red style — visually distinct from the #010101 dark used by Update Save (DR-010-001-04) and Send (DR-010-001-05). This contrast helps the user perceive the gravity of the action and avoids any accidental confusion between Send and Delete on screens that share the same two-panel layout.
- **Available for both Draft and Sent:** Per Knowledge Base §9.4 Action Availability matrix, Delete is the only action available on Sent announcements (Update and Send are at 50% opacity for Sent records). The Delete page renders correctly for both statuses; only the sibling action panel state differs (Update + Send interactive for Draft, at 50% opacity for Sent). This allows administrative cleanup of Sent records while still preserving recipient-side delivered copies.
- **Soft-delete semantics:** The record is marked with `deleted = true` and a `deleted_at` UTC timestamp. The record remains in the database for audit purposes (per Knowledge Base §5.1) but is excluded from all active views, search results, filter results, exports, and cross-platform consumers (e.g., MOBILE-APP announcement widget per §17 — only Sent + non-deleted records reach mobile). Sent announcements that are subsequently deleted do not retract notifications already dispatched; recipients retain their delivered copy.
- **Redirect-to-List on success (does NOT adopt the redirect-to-Details from DR-010-001-05 or the stay-on-page from DR-010-001-04 v1.1):** Delete Announcement intentionally redirects to the Announcement List on success because the deleted record is no longer accessible in active views — staying on the Delete page or redirecting to Details would leave the user on a now-unreachable view. List redirect lets the user immediately verify the record is gone and resume their list-level workflow.
- **Concurrent-delete idempotency:** If two users try to delete the same record, the second user's request returns 404, but this is treated as success (informational toast + redirect to list) rather than as a hard error — the desired end state (record gone) is achieved either way, and surfacing an "error" would be confusing UX.
- **No notifications on delete:** Unlike Send (which dispatches email + push to recipients), Delete dispatches NO notifications. Recipients of a previously-Sent announcement are not informed that the announcement was deleted; they retain their delivered copy and the deletion is invisible from their perspective. This matches Knowledge Base §17 cross-platform consumer rules — only Sent + non-deleted records reach mobile, so mobile widgets and detail screens will simply stop showing the deleted record on next refresh.
- **No discard-unsaved-changes confirmation:** Because the screen has no editable form state, navigation away (back arrow, sibling action buttons, browser navigation) is immediate without any confirmation. This is consistent with the §3.6 Dedicated Status-Transition Screen Pattern but simpler than Send Announcement (DR-010-001-05) — Send had a Receiver Info form that could be modified, while Delete has only a static warning paragraph and the page-level Delete button. There is nothing to "discard".

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
| 1.0 | 2026-05-06 | Claude | Initial draft from Figma "Delete Announcement" frame extraction (node 3385:22522); applies Detail/Profile View Pattern (Knowledge Base §3 and §3.4) with status-aware action panel; follows the §3.6 Dedicated Status-Transition Screen Pattern (single primary action button labeled with the transition verb, no Cancel/Discard button, explicit confirmation dialog as the irreversibility safety gate, redirect-on-success — but to the Announcement List instead of Details, because the deleted record is no longer accessible in active views); adopts the Soft Delete Transactional pattern (§5.1) — `deleted` flag + `deleted_at` timestamp, removed from all active views; available for both Draft and Sent statuses (per §9.4) with Update + Send sibling buttons at 50% opacity for Sent records; danger red Delete button style (per §5 Delete Action Pattern) distinct from the neutral #010101 used by Update Save and Send; default keyboard focus on Cancel in the confirmation dialog (per §5.2); concurrent-delete idempotency (already-deleted -> informational toast, not an error); no notifications dispatched on delete (recipients retain their delivered copy of Sent announcements). The bottom button label is "Delete" — the Figma frame may show "Save" but that is treated as a design artifact per the §3.6 generalized rule from DR-010-001-05 v1.1. |
