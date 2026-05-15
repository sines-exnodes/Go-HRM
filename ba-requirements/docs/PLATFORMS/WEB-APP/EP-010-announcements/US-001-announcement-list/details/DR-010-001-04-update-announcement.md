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
detail_id: DR-010-001-04
detail_name: "Update Announcement"
# Status & Version
status: draft
version: "1.2"
created_date: 2026-05-05
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
  - path: "../../EPIC.md"
    relationship: parent
# Input sources
input_sources:
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3385:21783"
    description: "Update Announcement screen design"
    extraction_date: 2026-05-05
---

# Detail Requirement: Update Announcement

**Detail ID:** DR-010-001-04
**Story:** US-001-announcement-list
**Epic:** EP-010 (Announcements)
**Status:** Draft
**Version:** 1.2

---

## 1. Use Case Description

As an **HR Manager or Admin**, I want to **edit the title, description, and label of a Draft announcement** so that **I can refine the content before sending it to employees**.

**Purpose:** Provide an edit form for **Draft** announcements so authors can correct typos, refine the message, or update the categorization label before pushing the announcement out. Sent announcements are immutable (per Knowledge Base §9.4) and therefore cannot reach this screen via the Update Announcement action — the action is only enabled when status = Draft. **Recipient configuration is NOT edited on this screen** — audience adjustments are made on the Send Announcement screen (DR-010-001-05) immediately before dispatch.

**Target Users:**
- HR Managers with announcement management permission
- Admins with full system access

**Key Functionality:**
- Pre-fill the form with the current announcement content fields: Title, Description, Label
- Allow editing of those fields with the same validation rules as the corresponding fields on the Create form
- Save changes — record stays in Draft status (no status change on edit), success toast shown, user stays on the form
- Navigate away freely via back arrow or sibling action buttons (no Cancel/Discard button exists; unsaved changes are simply lost on navigation)
- Two-panel layout consistent with Update Request Ticket (DR-003-001-06) and Announcement Details (DR-010-001-03)
- **Receiver Info is not editable here** — to change the audience the user navigates to the Send Announcement screen (DR-010-001-05)

---

## 2. User Workflow

**Entry Point:**
- Announcement Details page (DR-010-001-03) > Left action panel > "Update Announcement" button (only enabled when status = Draft)
- Direct URL navigation (e.g., bookmarked or deep-linked edit URL); access still gated by status check + permission

**Preconditions:**
- User is authenticated and logged in
- User has announcement management permission (via US-004 Permission Management)
- The target announcement exists, is not soft-deleted, and is currently in **Draft** status

**Main Flow:**
1. User clicks "Update Announcement" on the Announcement Details page (Draft only)
2. System navigates to the Update Announcement page (URL pattern includes announcement ID)
3. System displays the page header: back arrow + "Announcement Details" title + breadcrumb showing the announcement title (e.g., "Hung King's Festival")
4. System renders the two-panel layout:
   - **Left action panel (207px):** Details, Update Announcement (active), Send Announcement, Delete Announcement
   - **Right form card (600px):** Single card titled "Announcement Details" with three fields (Title, Description, Label) and a full-width Save button at the bottom — no Receiver Info card on this screen
5. System fetches the announcement record and pre-fills the three content fields with current values during a brief skeleton loading state
6. User edits any combination of fields:
   - **Title** (text input, mandatory)
   - **Description** (textarea, mandatory, resizable)
   - **Label** (single-select dropdown with create-on-type, optional — same behavior as Create)
7. User clicks **Save** at the bottom of the form
8. System validates the three content fields client-side; on success, sends the update to the server
9. Server saves the record (status remains Draft, recipient configuration unchanged), system shows a success toast "Announcement has been updated"; user stays on the Update form with the now-saved values pre-filled (form transitions back to clean state, Save button disabled until next change)

**Alternative Flows:**
- **Alt 1 - Validation Error:** Mandatory field empty or invalid value > inline error appears below the field; Save is blocked
- **Alt 2 - Back-Navigation:** User clicks the back arrow at any time > immediate navigation to Announcement Details; unsaved changes are discarded silently (no confirmation, no Cancel/Discard button exists)
- **Alt 3 - Concurrent Status Change (Sent):** While editing, another user sends the announcement; on Save, server returns conflict; system shows error toast "This announcement has already been sent and can no longer be edited" + redirects to the (now read-only) Announcement Details page
- **Alt 4 - Concurrent Delete:** While editing, another user deletes the announcement; on Save, server returns 404; system shows error toast "This announcement no longer exists" + redirects to Announcement List
- **Alt 5 - Permission Lost Mid-Session:** User's management permission is revoked while editing; on Save, server returns 403; toast "You no longer have permission to edit this announcement" + redirect to fallback page
- **Alt 6 - Direct URL to Sent Announcement:** User opens the edit URL of a Sent announcement directly; system blocks load and redirects to the read-only Announcement Details with toast "This announcement has been sent and can no longer be edited"
- **Alt 7 - Direct URL to Deleted/Invalid Announcement:** Empty state "Announcement not found or no longer available" + "Back to List" button
- **Alt 8 - Click Sibling Action Buttons:** User clicks "Details", "Send Announcement", or "Delete Announcement" on the action panel at any time > immediate navigation to that action; unsaved changes are discarded silently (no confirmation)
- **Alt 9 - Label Created During Edit:** User types a label name that does not exist + presses Enter > new label is created on save and assigned to this announcement; new label becomes available system-wide for future announcements

**Exit Points:**
- **Success:** Toast "Announcement has been updated" — user **stays on the Update form**; saved values become the new pre-fill baseline; form returns to clean state (Save disabled until next change)
- **Back / Sibling Action Navigation:** Immediate navigation to the chosen target — no confirmation, unsaved changes silently discarded (there is no Cancel/Discard button on this screen)
- **Validation Error:** Inline errors displayed; user remains on form
- **Server Error:** Error toast + form remains; user can retry

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Title | Text Input | Min 1 char, Max 100 chars, whitespace-only invalid | Yes | Pre-filled with current value | Announcement subject/headline |
| Description | Textarea (resizable) | Min 1 char, Max 2000 chars, whitespace-only invalid | Yes | Pre-filled with current value | Full announcement content |
| Label | Single-select Dropdown (create-on-type) | Max 50 chars for new label name | No | Pre-filled with current value (or empty) | Optional categorization tag |

**Note on the Figma design:** The Figma frame shows three fields (Title, Description, Label) marked with asterisks. The asterisk on Label is a design rendering artifact — Label remains **optional** for consistency with the Create form (DR-010-001-02 §3) and the underlying data model.

**Receiver Info is NOT on this form.** Audience adjustments (Everyone / Specific one + Employee selection) are made on the dedicated Send Announcement screen (DR-010-001-05), which is reached via the "Send Announcement" sibling button in the action panel. The underlying receiver configuration is preserved unchanged when the Update form is saved.

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back Arrow | Icon Button | Always enabled | Navigate to Announcement Details immediately (unsaved changes discarded silently) | Top-left of header (24x24 ArrowLeft icon) |
| Details Button | Tertiary Button | Always interactive | Navigate to Announcement Details immediately (unsaved changes discarded silently) | First button in action panel |
| Update Announcement Button | Tertiary Button | Active (current view) | No-op (visual indicator only) | Highlighted with #f5f5f5 accent background |
| Send Announcement Button | Tertiary Button | Always interactive (Draft only — Update is only reachable for Draft) | Open Send confirmation immediately (unsaved changes discarded silently) | Vertical button stack |
| Delete Announcement Button | Tertiary Button (danger) | Always interactive | Open Delete confirmation immediately (unsaved changes discarded silently) | Bottom of action panel |
| Save Button | Primary Button | Enabled when form is valid AND form is dirty | Save changes, status remains Draft, show success toast, stay on the Update form (form returns to clean state) | Full-width 600px, dark background (#010101), white text |
| Title Field | Text Input | Always enabled | Update value | Pre-filled, auto-focus on page load |
| Description Field | Textarea | Always enabled | Update value | Pre-filled, resizable |
| Label Dropdown | Single-select Dropdown (create-on-type) | Always enabled | Select existing or type to create new | Pre-filled |

### Status-Aware Action Panel (Update view)

| Status | Details | Update Announcement | Send Announcement | Delete Announcement |
|--------|---------|---------------------|-------------------|---------------------|
| Draft (only state Update is reachable in) | Interactive | Active (current view) | Interactive | Interactive |

**Note:** Update Announcement is only enabled for Draft status (per Knowledge Base §9.4). Sent announcements never reach this view; if a user attempts direct URL access to the edit URL of a Sent record, the system redirects to the read-only Details page (Alt 7).

### Label Dropdown Behavior (Create-on-Type — same as DR-010-001-02 §3)

| User Action | System Behavior | Result |
|-------------|-----------------|--------|
| Click dropdown | Opens dropdown showing all existing labels | User sees available labels |
| Type text matching existing label | Filters dropdown to matching labels | User can select from filtered list |
| Type text NOT matching any label | Shows filtered results + "Press Enter to add new" help text | User informed they can create |
| Press Enter with non-matching text | Creates new label with typed name, selects it | New label saved system-wide and selected |
| Select existing label | Label selected and dropdown closes | Label value updated |
| Clear selection (X icon) | Removes selected label | Field returns to empty state |

---

## 4. Data Display

### Layout Structure

| Region | Position | Dimensions | Content |
|--------|----------|------------|---------|
| Header Row | Top | Full width, 29px tall | Back arrow + "Announcement Details" title + breadcrumb separator + announcement title |
| Action Panel | Left | 207px wide, 192px tall (4 buttons × 36px + gaps) | Vertical stack: Details, Update Announcement (active), Send Announcement, Delete Announcement |
| Form Card | Right | 600px wide, ~343px tall | Single card titled "Announcement Details" containing the three fields (Title, Description, Label) |
| Save Button | Below form card | 600px wide, 40px tall | Single full-width Save button (dark background) |

### Form Card Layout (from Figma)

| Card | Title | Content |
|------|-------|---------|
| Form Card | Announcement Details | Title input, Description textarea (76px tall, resizable), Label dropdown |

**Note on Receiver Info:** The Update screen does NOT include a Receiver Info card. The Figma frame intentionally shows only the "Announcement Details" card; this DR honours that design. Audience changes are made on the dedicated Send Announcement screen (DR-010-001-05), which is the canonical place to confirm/adjust recipients immediately before dispatch.

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch | Skeleton placeholders for action panel buttons and form fields |
| Loaded (Clean) | Announcement loaded with status = Draft, no edits yet | Form pre-filled with current values, Save button disabled |
| Dirty | User has changed any field from the pre-filled value | Save button enabled (assuming form is valid) |
| Validation Error | Mandatory field empty or invalid | Inline error in red below the field; Save remains disabled |
| Saving | Save clicked | Save button shows loading spinner; form fields disabled to prevent further edits |
| Success | Save completed | Toast "Announcement has been updated" — user stays on the Update form; saved values become the new pre-fill baseline; form returns to clean state (Save disabled) |
| Conflict (Sent) | Server returns conflict because status changed to Sent | Error toast + redirect to Announcement Details (now read-only) |
| Conflict (Deleted) | Server returns 404 | Error toast + redirect to Announcement List |
| Not Found | Direct URL to invalid/deleted ID | Empty state "Announcement not found or no longer available" + "Back to List" button |
| Permission Denied | User loses management permission | Toast "You no longer have permission to edit this announcement" + redirect to fallback |

### Validation Error Messages (parity with Create — DR-010-001-02 §4)

| Field | Error Condition | Error Message |
|-------|----------------|---------------|
| Title | Empty or whitespace-only | "Title is required" |
| Title | Exceeds 100 characters | "Title must not exceed 100 characters" |
| Description | Empty or whitespace-only | "Description is required" |
| Description | Exceeds 2000 characters | "Description must not exceed 2000 characters" |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Users with announcement management permission can access the Update Announcement page; users without permission are redirected to fallback with permission-denied toast
- **AC-02:** Update Announcement is only reachable for announcements in Draft status; attempting to load the edit URL of a Sent announcement redirects to the read-only Details page with toast
- **AC-03:** Page renders the two-panel layout: 207px-wide action panel on the left, 600px-wide form card area on the right
- **AC-04:** Header shows back arrow + "Announcement Details" title + breadcrumb displaying the announcement title
- **AC-05:** Action panel renders four buttons in order: Details, Update Announcement (active with #f5f5f5 accent background), Send Announcement, Delete Announcement
- **AC-06:** Form card "Announcement Details" displays exactly three fields pre-filled with the current values: Title, Description, Label — no Receiver Info card is rendered on this screen
- **AC-07:** Title field is mandatory (min 1 char, max 100 chars, whitespace-only invalid); inline error "Title is required" if empty
- **AC-08:** Description field is mandatory (min 1 char, max 2000 chars, whitespace-only invalid); inline error "Description is required" if empty
- **AC-09:** Label field is optional and supports the create-on-type pattern from DR-010-001-02 (filter, "Press Enter to add new", clear selection)
- **AC-10:** Save button is disabled while the form is invalid (any required field empty or invalid value) OR while the form is clean (no changes from pre-filled values)
- **AC-11:** Save button uses full-width 600px layout with dark background (#010101) and white "Save" text
- **AC-12:** Clicking Save on a valid, modified form persists the three content fields and **keeps the status as Draft** (no status change on edit); recipient configuration is left untouched on the server
- **AC-13:** On successful save, system shows toast "Announcement has been updated" and the user **stays on the Update form**; the saved values become the new pre-fill baseline; the form returns to clean state with the Save button disabled until the next change (no redirect)
- **AC-14:** Clicking the back arrow at any time (clean or dirty form) navigates immediately to the Announcement Details page; unsaved changes are discarded silently — there is no confirmation dialog and no Cancel/Discard button on this screen
- **AC-15:** Clicking sibling action buttons (Details, Send Announcement, Delete Announcement) at any time navigates immediately to that action; unsaved changes are discarded silently (no confirmation dialog)
- **AC-16:** Loading state shows skeleton placeholders for action panel buttons and the three content fields
- **AC-17:** Whitespace-only input is treated as empty (validation fails) and leading/trailing whitespace is trimmed before saving
- **AC-18:** If the announcement is Sent by another user during edit, server returns conflict; system shows error toast "This announcement has already been sent and can no longer be edited" + redirects to read-only Details
- **AC-19:** If the announcement is deleted by another user during edit, server returns 404; system shows error toast "This announcement no longer exists" + redirects to Announcement List
- **AC-20:** Direct URL to a deleted/invalid announcement ID shows "Announcement not found or no longer available" empty state with "Back to List" button
- **AC-21:** Auto-focus is applied to the Title field on page load (after data is fetched)
- **AC-22:** Newly created labels (via create-on-type during update) are immediately available system-wide for future announcements
- **AC-23:** Audience adjustments (Everyone / Specific one + Employee) cannot be made on this screen — to change recipients the user must navigate to Send Announcement (DR-010-001-05) via the action panel

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Open Update for Draft announcement | Click Update Announcement on Details page (Draft) | Form loaded with pre-filled values, Update button active, status remains Draft | High |
| Edit title and save | Change title, click Save | Toast "Announcement has been updated", redirect to list, status still Draft | High |
| Edit description and save | Change description, click Save | Persisted, redirect to list | High |
| Edit label (select existing) | Select existing label, save | Label updated on record; recipient config unchanged | Medium |
| Edit label (create new) | Type new name, press Enter, save | New label created system-wide and assigned; recipient config unchanged | Medium |
| Clear label | Remove label via X icon, save | Label removed from record; recipient config unchanged | Medium |
| Receiver Info absent on Update | Render Update form for any Draft | No Receiver Info card visible; only the "Announcement Details" card with Title/Description/Label is shown | High |
| Recipient config preserved through edit | Save with content changes only | Server-side recipient configuration is unchanged after save (no toggle/employee fields submitted) | High |
| Validation - empty title | Clear title, click Save | Inline error "Title is required", Save disabled | High |
| Validation - empty description | Clear description, click Save | Inline error "Description is required", Save disabled | High |
| Validation - title over 100 chars | Type 101+ chars | Error "Title must not exceed 100 characters" | Medium |
| Validation - description over 2000 chars | Type 2001+ chars | Error "Description must not exceed 2000 characters" | Medium |
| Whitespace trimming | Pre-pend/append spaces in title, save | Saved value has trimmed whitespace | Medium |
| Whitespace-only title | Replace title with spaces only, save | Inline error "Title is required" | Medium |
| Back arrow with clean form | Open Update, immediately click back | Direct navigation to Details (no confirmation) | Medium |
| Back arrow with dirty form | Edit title, click back | Direct navigation to Details (no confirmation); unsaved changes discarded silently | High |
| Click Details with dirty form | Edit, click Details button | Direct navigation to Details (read-only); unsaved changes discarded silently | Medium |
| Click Send with dirty form | Edit, click Send Announcement | Direct navigation to Send confirmation flow; unsaved changes discarded silently | Medium |
| Click Delete with dirty form | Edit, click Delete Announcement | Direct navigation to Delete confirmation flow; unsaved changes discarded silently | Medium |
| Save success - stays on page | Edit title, click Save | Toast "Announcement has been updated", form remains visible with new values pre-filled, Save disabled (clean state) | High |
| Save then re-edit | After save, change another field | Save button re-enables once the form is dirty again | Medium |
| Concurrent Send conflict | Edit form open, another user sends; current user clicks Save | Error toast + redirect to read-only Details | High |
| Concurrent Delete conflict | Edit form open, another user deletes; current user clicks Save | Error toast + redirect to list | High |
| Direct URL to Sent announcement | Visit /announcements/:id/edit when status=Sent | Redirect to Details with toast "This announcement has been sent and can no longer be edited" | High |
| Direct URL to deleted announcement | Visit /announcements/:id/edit when soft-deleted | Empty state "Announcement not found or no longer available" + Back to List button | Medium |
| Permission revoked mid-edit | User loses permission; clicks Save | 403 from server > toast + redirect to fallback | Medium |
| Save with no changes | Click Save without editing | Save disabled (form is not dirty) — button remains disabled | Low |
| Save while loading | Network slow, double-click Save | Button shows spinner, disabled to prevent duplicate submission | Low |
| Skeleton loading | First load on slow network | Skeleton for action panel + form fields until data arrives | Low |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only users with announcement management permission (configured via US-004) can access the Update Announcement page; access is verified at route load and on Save
- **Rule 2:** Update Announcement is only allowed when the announcement is in **Draft** status — server rejects any update to a Sent announcement (per Knowledge Base §9.4 immutability)
- **Rule 3:** Saving an edit **does not change the status** — Draft remains Draft (per Knowledge Base §4.2: feature-specific status-on-edit; announcements have no approval workflow, so no status reset is needed)
- **Rule 4:** Created Date is preserved (never updated by an edit); only the record's content fields and the implicit `last_updated` timestamp change on save
- **Rule 5:** Last Announce Date remains null after edit (still Draft) — it is only set when the user invokes Send (per DR-010-001-02 Rule 9 and DR-010-001-03 Rule 6)
- **Rule 6:** Same field-level validation rules as the Create form apply for the three editable fields (Title, Description, Label) — see Knowledge Base §4.1 "Same Validation"
- **Rule 7:** Leading/trailing whitespace is trimmed from Title and Description before saving (per DR-010-001-02 Rule 6)
- **Rule 8:** Whitespace-only input is treated as empty and fails validation (per DR-010-001-02 Rule 7)
- **Rule 9:** Receiver configuration is **NOT** edited on this screen — the existing recipient configuration on the record is preserved unchanged when Update saves. The server payload submitted from this form contains only the three content fields (Title, Description, Label); the recipient configuration is left as-is on the persisted record. To change the audience, the user must use the dedicated Send Announcement screen (DR-010-001-05)
- **Rule 10:** Label is optional; saving with an empty Label clears the existing label association
- **Rule 11:** New labels created via create-on-type are persisted system-wide and become available for all announcements (per DR-010-001-02 Rule 21); duplicate label creation (case-insensitive) selects the existing label instead
- **Rule 12:** Concurrent edit protection: server verifies the announcement is still in Draft status (and not soft-deleted) before applying the update (per Knowledge Base §4.3)
  - If status changed to Sent: return 409 Conflict; UI redirects to read-only Details with toast
  - If record was soft-deleted: return 404 Not Found; UI redirects to list with toast
- **Rule 13:** No Cancel/Discard button exists on this screen — the Update form has only a Save button. Users navigate away via the back arrow or sibling action buttons (Details, Send, Delete); all such navigation is **immediate with no confirmation dialog**, and any unsaved changes are silently discarded. This is a **deliberate deviation** from the modal/dialog Edit Form Pattern (Knowledge Base §4.1 Cancel Behavior) because Update Announcement is a standalone full-page form with no overlay context to "cancel" out of
- **Rule 14:** On successful save, the user **stays on the Update form** — no redirect occurs. The system shows a success toast, the saved values become the new pre-fill baseline, and the form transitions back to the clean state (Save button disabled until the next change). This deviates from the Knowledge Base §4.1 Edit Form Success pattern (which redirects to the list) and is feature-specific to announcements
- **Rule 15:** No notifications are dispatched on update — notifications fire only on Send (per DR-010-001-02 Rule 10). Editing a Draft has zero side-effects on recipients
- **Rule 16:** Permission verification is performed both on initial page load and on Save submission — revocation mid-session is handled gracefully (per DR-010-001-03 Rule 20)
- **Rule 17:** All dates remain stored in UTC; display uses DD/MM/YYYY format in the user's timezone (per DR-010-001-03 Rule 15)
- **Rule 18:** Sibling action buttons (Details, Send Announcement, Delete Announcement) on the action panel are interactive at all times while editing; clicking them performs immediate navigation to the corresponding view/flow with no confirmation, even if the form is dirty
- **Rule 19:** The Update Announcement page is the only entry point for editing announcement content fields (no inline editing on the Details page, per Knowledge Base §3.2 "No Inline Editing"); the Send Announcement page (DR-010-001-05) is the only entry point for editing recipient configuration on an existing Draft

**State Transitions:**

```
[Viewing Draft (Details)] → [Click Update Announcement] → [Update Form (pre-filled, clean)]
[Update Form (clean)] → [Edit Fields] → [Update Form (dirty, Save enabled)]
[Update Form (dirty)] → [Click Save (valid)] → [Persisted (still Draft)] → [Update Form (clean, new pre-fill baseline) + Success Toast]   // stays on page
[Update Form] → [Click Back arrow] → [Announcement Details]                         // immediate, no confirmation
[Update Form] → [Click Details / Send / Delete buttons] → [Target view/flow]        // immediate, no confirmation
[Update Form] → [Click Save during Sent conflict] → [Server 409] → [Redirect to read-only Details + Error Toast]
[Update Form] → [Click Save during Deleted conflict] → [Server 404] → [Redirect to List + Error Toast]
```

**Permission Model:**

| Action | Required Permission | Status Constraint |
|--------|---------------------|-------------------|
| Access Update Form | Announcement Management | Draft only |
| Save Update | Announcement Management | Draft only |

**Confirmation Dialogs:**

None — this screen has no Cancel/Discard button and no discard-unsaved-changes confirmation. All navigation away from the form (back arrow, sibling action buttons, browser navigation) is immediate; unsaved changes are silently discarded. Success Save shows a toast only and keeps the user on the form.

**Dependencies:**

- US-004 Permission Management — provides view and management permission gates
- DR-010-001-01 Announcement List — navigation target only on conflict (deleted concurrently); not the success destination
- DR-010-001-02 Create Announcement — defines the data shape, validation rules, label create-on-type pattern, and receiver model that this form preserves and edits
- DR-010-001-03 Announcement Details — entry point (Update Announcement button) and navigation target via the back arrow / Details button
- DR-010-001-05 Send Announcement Confirmation (planned) — destination of Send Announcement action from action panel
- DR-010-001-06 Delete Announcement Confirmation (planned) — destination of Delete Announcement action from action panel
- Authentication system — user must be logged in
- Label data — for label dropdown options and create-on-type behavior

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Two-panel layout matches the Update Request Ticket pattern (DR-003-001-06) and Announcement Details (DR-010-001-03) — same spatial mental model for status-managed entities, reducing learning cost
- **Optimization 2:** Active "Update Announcement" button uses the established #f5f5f5 accent background (per Knowledge Base §3.4 Status Management on Detail Page) so users always know which view they're on
- **Optimization 3:** All sibling actions (Details, Send, Delete) remain visible and interactive in the action panel — users can pivot to a different action immediately
- **Optimization 4:** Pre-filled form with all current values lets users see exactly what will change — no need to re-enter unchanged fields
- **Optimization 5:** Auto-focus on Title field on page load supports immediate typing for quick title corrections (most common edit case)
- **Optimization 6:** Successful Save keeps the user on the form (no redirect) so they can verify the saved values immediately and iterate on subsequent changes without re-navigating — supports rapid multi-pass editing of a single Draft
- **Optimization 7:** Single Save button at the bottom (no header Save, no Cancel button) keeps the visual hierarchy simple and matches the Update Request Ticket form (DR-003-001-06) — only one place to commit changes; navigation away is via the back arrow or sibling action buttons
- **Optimization 8:** Inline validation errors appear immediately below the field (not via toast) so users can correct without losing context (per Knowledge Base §2.3)
- **Optimization 9:** Save button disabled while form is invalid OR clean OR while saving — prevents invalid submissions, no-op submissions, and double-submission
- **Optimization 10:** No status change on edit means a Draft author can iterate freely without triggering downstream effects (no re-approval, no notifications) — consistent with the Request Ticket edit pattern (Knowledge Base §4.2)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full two-panel layout: 207px action panel + 600px form card |
| Below Desktop | Out of scope — web admin is desktop-only |

**Accessibility Requirements:**

- [x] Keyboard navigable — Tab order: Back arrow > action panel buttons (top to bottom) > form fields (top to bottom) > Save button
- [x] Screen reader compatible — all field labels associated with inputs; toggle states announced; Save button labeled
- [x] Sufficient color contrast — disabled button states still meet WCAG AA when combined with aria-disabled
- [x] Focus indicators visible — clear focus ring on action buttons, fields, and Save button
- [x] Auto-focus on Title respects user's keyboard navigation preferences (no focus theft mid-form)
- [x] Error messages announced via aria-live region when validation fails
- [x] Success toast announced via aria-live="polite" so screen reader users are informed of save success even though they remain on the same page

**Design References:**

- Figma Node: 3385:21783 (Update Announcement screen)
- Follows Edit Form Pattern from Knowledge Base §4 (pre-fill, same validation as Create, same dirty/clean cancel behavior)
- Follows Edit Form Layout Variant from Knowledge Base §4.5 — same two-panel structure as Update Request Ticket (DR-003-001-06)
- Follows Detail/Profile View Pattern §3.4 for the action panel (status-aware, action-button visibility)
- Save button styled per Knowledge Base §2.9 primary action style (#010101 background, white text)
- Validation messages and error display per Knowledge Base §2.3
- Label dropdown create-on-type pattern per DR-010-001-02 §3

---

## 8. Additional Information

### Out of Scope

- Editing Sent announcements (immutable per Knowledge Base §9.4 — not allowed at any time)
- **Editing recipient configuration (Everyone / Specific one + Employee) on this screen — that is exclusively done on the Send Announcement screen (DR-010-001-05)**
- Status change from Update form (no Send button inside the form — Send is via the action panel which navigates to the dedicated Send confirmation flow)
- Bulk edit (multi-select edit on the list)
- Inline editing on the Details page (all content edits go through this Update form per Knowledge Base §3.2)
- Optimistic UI / auto-save — saves are explicit and manual
- Edit history / revision tracking (no audit trail UI in v1.0)
- Undo after save (no redo / undo flow once saved)
- Mobile responsive layout (web admin desktop-only)
- Notification preview (email/push body preview before save) — out of scope at epic level
- Resending notifications when a Sent announcement is updated (Sent is immutable, so this case never arises)
- Versioning / "Save a copy as new draft" from edit — out of scope

### Open Questions

- None — all decisions resolved against the Edit Form Pattern (Knowledge Base §4), the Create Announcement DR (DR-010-001-02), and the Update Request Ticket DR (DR-003-001-06).

### Related Features

- DR-010-001-01: Announcement List (parent list view; back-navigation target on Save)
- DR-010-001-02: Create Announcement (defines field shape, validation, label create-on-type, and receiver model)
- DR-010-001-03: Announcement Details (entry point — Update Announcement button; back-navigation target on cancel)
- DR-010-001-05: Send Announcement Confirmation (planned — destination of Send Announcement action)
- DR-010-001-06: Delete Announcement Confirmation (planned — destination of Delete Announcement action)
- DR-003-001-06: Update Request Ticket (WEB-APP) — sibling Edit Form layout precedent (two-panel + single bottom Save)
- US-004: Permission Management (provides management permission source)

### Notes

- The Figma frame is named "Update Announcement" (id 3385:21783) but reuses the page title "Announcement Details" with the announcement title as breadcrumb — same header treatment as DR-010-001-03 to keep navigation consistent across the entity's screens.
- The Figma frame depicts three required-marked fields (Title, Description, Label), but Label was confirmed optional in DR-010-001-02. To preserve consistency with Create (where Label is optional) and the underlying data model, this Update form keeps **Label as optional**. The asterisk on Label in the design is treated as a rendering artifact.
- **Receiver Info is intentionally NOT on this form.** The Figma frame deliberately shows only the "Announcement Details" card; this DR follows that design. Editing audience and editing content are split across two screens for a single Draft: Update (this DR) handles the message content; Send (DR-010-001-05) handles the audience finalization immediately before dispatch. This separation means the Update payload contains only the three content fields and never mutates the recipient configuration on the persisted record.
- The Save button uses the dark primary style (#010101 / white text) consistent with Update Request Ticket (DR-003-001-06) — single bottom Save, no header Save, **no Cancel/Discard button** at all.
- **Stay-on-page Save behavior:** This DR deliberately deviates from the Knowledge Base §4.1 Edit Form Success pattern (which redirects to the list). On successful save the user remains on the Update form, the success toast confirms the save, and saved values become the new pre-fill baseline. This supports rapid multi-pass editing of a Draft and aligns with the standalone-tab UX (no overlay context to dismiss).
- **No discard-unsaved-changes confirmation:** Because this screen has no Cancel/Discard button, navigating away (back arrow, sibling action buttons, browser navigation) is immediate and silently discards unsaved changes. This deliberately deviates from the Knowledge Base §4.1 Cancel Behavior (which assumes a Cancel button + dirty confirmation). Authors are expected to either Save or accept the loss of unsaved edits.
- Update Announcement is the **only** way to modify a Draft's content; no inline editing on Details (per Knowledge Base §3.2 / §3.4) and no edit affordance for Sent records (per Knowledge Base §9.4).
- Edits do **not** trigger notifications — notifications fire only on Send (per DR-010-001-02 Rule 10 and DR-010-001-03 Rule 7). This means a Draft author can iterate freely without side-effects on employees' inboxes or devices.
- Per Knowledge Base §4.2, status-on-edit behavior is feature-specific. Announcements behave like Request Tickets here (status preserved on edit) — not like Leave Requests (which revert Approved → Pending). This is appropriate because announcements have no approval workflow.

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
| 1.0 | 2026-05-05 | Claude | Initial draft from Figma "Update Announcement" frame extraction; applies Edit Form Pattern (Knowledge Base §4) and §4.5 Layout Variant (two-panel + single bottom Save) consistent with DR-003-001-06 Update Request Ticket; preserves data shape and validation from DR-010-001-02 Create Announcement; status-aware action panel from DR-010-001-03 Announcement Details |
| 1.1 | 2026-05-06 | Claude | Stakeholder feedback: (1) successful Save no longer redirects — user stays on the Update form, toast only, saved values become new pre-fill baseline; (2) removed all discard/cancel confirmation logic — there is no Cancel/Discard button, so back-arrow and sibling action navigation is immediate with silent discard of unsaved changes. Updated §1, §2 (workflow + alt flows + exit points), §3 (interaction elements), §4 (display states), §5 (AC-15..17, scenarios), §6 (Rule 13 rewritten, Rule 14 added stay-on-page, Rule 18 removed dirty discard, state transitions, confirmation dialogs section now N/A), §7 (Optimization 6 rewritten, Optimization 9 includes clean-state guard, accessibility note for toast), §8 (Notes block clarifies the two deviations from Knowledge Base §4.1) |
| 1.2 | 2026-05-06 | Claude | Stakeholder feedback: Receiver Info is NOT on the Update form — the Figma design shows only the Announcement Details card, and audience adjustments now happen exclusively on the Send Announcement screen (DR-010-001-05). Removed the Receiver Info card and all associated content: §1 Use Case + Key Functionality narrowed to three content fields; §2 Main Flow step removed (former step 7), step numbering updated; §3 Input Fields table reduced to Title/Description/Label (removed Everyone/Specific one/Employee rows); §3 Interaction Elements table removed Everyone Toggle / Specific one Toggle / Employee Dropdown rows; §3 Receiver Selection Behavior subsection deleted; §4 Form Card Layout note rewritten (single card only); §4 Validation Errors removed Employee row; §5 ACs removed AC-10 + AC-11 (renumbered the rest 10..22) and added AC-23 stating recipient edits live on Send Announcement; §5 Test Scenarios removed Specific/Everyone toggle and Specific-with-no-employee rows, added scenarios verifying Receiver Info absence and that recipient config is preserved server-side through edit; §6 Rule 9 rewritten (recipient config preserved unchanged on Update server payload); §6 Rule 19 split entry-points (Update = content only; Send = recipients only); §8 Out of Scope now explicitly excludes recipient editing on this form; §8 Notes rewritten to explain the deliberate split between content editing and audience editing across the two screens. |
