# Project Knowledge Base

**Purpose:** Centralized repository of confirmed patterns, rules, and decisions from completed detail requirements. This file is the single source of truth for established behaviors across the HRM platform.

**Usage:** The dr-agent reads this file during Phase 2 (Interactive Detail Writing) to auto-suggest answers for questions that have already been resolved in previous DRs. Instead of re-asking the user, the agent references this knowledge and presents suggestions for confirmation.

**Maintenance:** Updated after each new DR is completed or when existing patterns change. Add new entries, update existing ones, or mark deprecated patterns.

**Last Updated:** 2026-05-06 (DR-010-001-04 v1.2 — Update Announcement form excludes Receiver Info; recipient editing for an existing Draft is exclusively on the Send Announcement screen DR-010-001-05; §4.5 announcement-specific edit form rules updated accordingly)

---

## Table of Contents

1. [List View Pattern](#1-list-view-pattern)
2. [Create Form Pattern](#2-create-form-pattern)
3. [Detail/Profile View Pattern](#3-detailprofile-view-pattern)
4. [Edit Form Pattern](#4-edit-form-pattern)
5. [Delete Action Pattern](#5-delete-action-pattern)
6. [Entity Relationships](#6-entity-relationships)
7. [Permission & Access Control](#7-permission--access-control)
8. [Common UI Behaviors](#8-common-ui-behaviors)
9. [Status & State Definitions](#9-status--state-definitions)
10. [Token-Based Setup Pattern](#10-token-based-setup-pattern)
11. [Leave Quota Management Pattern](#11-leave-quota-management-pattern)
12. [Mobile Authentication Pattern](#12-mobile-authentication-pattern)
13. [Mobile Dashboard Pattern](#13-mobile-dashboard-pattern)
14. [Mobile Create Form Pattern](#14-mobile-create-form-pattern)
15. [Mobile Check-In/Out Pattern](#15-mobile-check-inout-pattern)
16. [Calendar Matrix View Pattern](#16-calendar-matrix-view-pattern)
17. [Mobile Announcement Widget Pattern](#17-mobile-announcement-widget-pattern)
18. [Open/Pending Decisions](#18-openpending-decisions)

---

## 1. List View Pattern

**Applies to:** Any page that displays a collection of records in a table format.
**Confirmed in:** User List (DR-001-005-01), Department List (DR-008-001-01), Position List (DR-008-002-01), Role List (DR-001-004-01), Skill List (DR-008-003-01), Leave Requests List (DR-002-001-01), Request Tickets List (DR-003-001-01), Announcement List (DR-009-001-01)

### 1.1 Search

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Debounce | ~300ms delay before triggering search | All list DRs |
| Matching | Case-insensitive, partial/contains matching | All list DRs |
| Scope | Search applies to specific fields per feature (defined per DR) | All list DRs |
| No Results | Display "No results found" message when search returns zero | All list DRs |
| Reset | Clearing search restores full list | All list DRs |

**Search field count varies by feature:**
- Simple entities (Department, Position, Role, Skill): 1 field (name only)
- Complex entities (User): multiple fields (name, email, phone, address)
- Transactional (Leave Requests): multiple fields (employee name, leave type)
- Transactional (Request Tickets): multiple fields (employee name, email, phone number)

### 1.2 Filters

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Type | Multi-select dropdown chips | User List, Leave Requests, Request Tickets |
| Built-in Search | Filter dropdowns have client-side search (no server round-trip) | User List, Leave Requests, Request Tickets |
| Chip Display | Shows count of selected values when active (e.g., "Status (2)") | User List, Leave Requests, Request Tickets |
| Logic | OR within same filter, AND across different filters | User List, Leave Requests, Request Tickets |
| Reset Button | Only visible when filters or search are active | User List, Leave Requests, Request Tickets |
| No Filters | Simple entity lists (Department, Position, Role, Skill) have no filter chips | DR-008 series |

**Note:** Not all list views have filters. Filters are used when the entity has meaningful categorical attributes to filter by.

### 1.3 Pagination

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Default | 10 rows per page | All list DRs |
| Options | 10, 25, 50 rows per page | All list DRs |
| Reset | Resets to page 1 when search/filters applied or cleared | All list DRs |
| Hidden | Pagination controls hidden when total results <= current rows per page | All list DRs |

### 1.4 Export

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Scope | Downloads currently filtered/searched result set (not full list) | All export-enabled lists |
| Visibility | Export button visible to all users with view permission | All export-enabled lists |
| Loading | Export button shows loading state while preparing file | All export-enabled lists |
| Format | Pending PO decision (CSV or .xlsx) | All lists |

### 1.5 Gear Icon (Row Actions)

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Visibility | Hidden for view-only users (no management permission) | All list DRs |
| Content | Context-sensitive actions based on row status and feature type | All list DRs |
| Position | Last column of the table | All list DRs |

**Common gear actions by feature type:**
- User List: Edit, Deactivate/Activate (toggles based on current status)
- Department/Position/Skill: Edit, Delete
- Role: Edit, Delete
- Leave Requests: Approve, Reject, Cancel (varies by status and user role)
- Request Tickets: Edit, In Progress, On Hold, Resume, Resolve, Close, Reopen, Cancel, Delete (varies by status and user role; 6-status lifecycle)
- Announcements: Draft status = Edit, Send, Delete; Sent status = Delete only (2-status lifecycle; immutable after send)

### 1.6 Table Display

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Header Style | Muted background (#f5f5f5) | User List, Skill List |
| Sort Order | Alphabetical by name (default) OR newest first (transactional) | All list DRs |
| Loading State | Skeleton rows to prevent layout shift | All list DRs |
| Empty State - No Data | Specific message with CTA to create first record | All list DRs |
| Empty State - No Results | "No results found" message (different from no data) | All list DRs |
| Empty State - No Filter Results | "No matching results" with option to clear filters | Filtered lists |

### 1.7 "+ Add New" Button

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Position | Top-right area of the list page | All list DRs |
| Visibility | Only visible to users with management permission (exception: Request Tickets visible to all authorized users including employees) | User List, Department, Position, Role, Skill, Leave Requests, Request Tickets |
| Navigation | Redirects to create form page | All list DRs |

---

## 2. Create Form Pattern

**Applies to:** Any page for creating a new record.
**Confirmed in:** Create User (DR-001-005-02), Create Department (DR-008-001-02), Create Leave Request (DR-002-001-02), Create Request Ticket (DR-003-001-02)

### 2.1 Layout

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Page Type | Full-page form (not modal) | Create User, Create Department, Create Leave Request, Create Request Ticket |
| Card Structure | Single or multiple cards grouping related fields | Create User (3 cards), Create Department (1 card), Create Leave Request (2 cards), Create Request Ticket (2 cards) |
| Button Position | Cancel (secondary) and Save (primary) in top-right header area | Create User, Create Department, Create Leave Request, Create Request Ticket |
| Dual Save (exception) | Full-width Save button at bottom of form in addition to top-right Save — both trigger same action | Create Leave Request (DR-002-001-02), Create Request Ticket (DR-003-001-02) |

### 2.2 Field Behavior

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Auto-focus | First/primary field auto-focused on page load | Create User, Create Department |
| Whitespace | Leading/trailing whitespace trimmed before saving | Create User, Create Department, Create Request Ticket |
| Whitespace-only | Name consisting only of whitespace treated as empty/invalid | Create Department, Create Request Ticket (Subject, Description) |

### 2.3 Validation

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Client-side | Mandatory fields, format, file types/sizes, date ranges | Create User, Create Department, Create Request Ticket |
| Server-side | Uniqueness checks (email, department name, etc.) | Create User, Create Department |
| Error Display | Inline below field (not toast), multiple errors shown simultaneously | Create User, Create Department, Create Request Ticket |
| Mandatory Error | "[Field name] is required" inline error message | Create User, Create Department, Create Request Ticket |
| Duplicate Error | Inline error + toast notification for uniqueness violations | Create User, Create Department |
| Email Format | "Please enter a valid email address" for invalid format | Create User |
| Trimming Order | Trim whitespace BEFORE uniqueness check | Create User, Create Department |

### 2.4 Dropdown Fields

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Searchable | Client-side search when options are dynamic/large | Create User (Role, Department, Position dropdowns) |
| Non-searchable | Simple static dropdowns for small option sets | Create User (Gender), Create Request Ticket (Request Type — 8 options) |

### 2.5 File Upload Fields

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Client-side Validation | Format + size validation before upload | Create User (Avatar, CV), Create Leave Request (Attachment), Create Request Ticket (Attachment) |
| Accepted Formats | PDF, PNG, JPG, DOCX (varies by field) | Create Leave Request, Create Request Ticket |
| Max Size | 5MB per file | Create Leave Request, Create Request Ticket |
| Preview | Preview shown on file selection | Create User |

### 2.6 Cancel/Discard Behavior

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Dirty Form | Shows "Discard unsaved changes?" confirmation dialog | Create User, Create Department, Create Request Ticket |
| Clean Form | Redirects immediately without confirmation | Create User, Create Department, Create Request Ticket |

### 2.7 Success Path

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Notification | Success toast message displayed | Create User, Create Department, Create Request Ticket |
| Redirect | Navigate back to list page | Create User, Create Department, Create Request Ticket |

### 2.8 Dual-Role Form Pattern

**Applies to:** Forms where employees create their own records AND admin/managers can create on behalf of others.

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Employee Role | Employee section hidden; record auto-assigned to logged-in user | Create Leave Request (DR-002-001-02), Create Request Ticket (DR-003-001-02) |
| Admin/Manager Role | Employee section visible with searchable dropdown | Create Leave Request, Create Request Ticket |
| Employee Info Card | On dropdown selection, shows profile: Full Name, Email, Phone, Department, Position | Create Leave Request, Create Request Ticket |
| Info Hint | Before selection, muted banner: "Select an employee to view their info" | Create Leave Request, Create Request Ticket |
| Success Toast (employee) | "[Record type] has been submitted" | Create Leave Request, Create Request Ticket |
| Success Toast (admin) | "[Record type] for '[employee name]' has been submitted" | Create Leave Request, Create Request Ticket |
| Active Employees Only | Employee dropdown loads active employees only — inactive/deactivated excluded | Create Request Ticket (DR-003-001-02) |

**Note:** The employee info card may show additional context-specific fields (e.g., Leave Days Remaining for leave requests). Each feature decides which extra fields to include.

### 2.9 Dual-Action Save Pattern

**Applies to:** Forms with two distinct save actions that create different statuses.
**Confirmed in:** Create Announcement (DR-010-001-02)

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Two Save Actions | Form has "Save As Draft" (secondary) and "Save & Send" (primary) buttons | Create Announcement |
| Dual Button Placement | Both actions appear in header AND bottom of form | Create Announcement |
| Header Layout | Cancel + Save As Draft + Save & Send (right-aligned) | Create Announcement |
| Bottom Layout | Save As Draft (left, 50%) + Save & Send (right, 50%) side-by-side | Create Announcement |
| Primary Action Style | "Save & Send" uses dark background (#010101), white text | Create Announcement |
| Secondary Action Style | "Save As Draft" uses white background with border | Create Announcement |
| Different Statuses | Draft action creates status=Draft; Send action creates status=Sent | Create Announcement |
| Immutable After Send | Records saved via "Save & Send" cannot be edited | Create Announcement |
| Success Toast (Draft) | "[Record type] has been saved as draft" | Create Announcement |
| Success Toast (Sent) | "[Record type] has been sent" | Create Announcement |

### 2.10 Mutually Exclusive Toggle Pattern

**Applies to:** Forms with toggle switches that function as radio buttons (only one can be ON).
**Confirmed in:** Create Announcement (DR-010-001-02)

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Toggle Behavior | Turning one toggle ON automatically turns the other OFF | Create Announcement |
| Default State | One toggle is ON by default (the most common use case) | Create Announcement |
| Conditional Fields | Dependent fields enabled/disabled based on toggle state | Create Announcement |
| Disabled Visual | Dependent fields shown at 50% opacity when disabled | Create Announcement |
| Validation | If toggle requires dependent field, validation enforces selection | Create Announcement |

**Announcement-specific implementation:**
- "Everyone" toggle (default ON) = send to all employees
- "Specific one" toggle (default OFF) = enable Employee multi-select dropdown
- Employee dropdown disabled at 50% opacity when "Everyone" is selected

---

## 3. Detail/Profile View Pattern

**Applies to:** Read-only pages displaying a single record's full information.
**Confirmed in:** User Details (DR-001-005-03), Request Ticket Details (DR-003-001-05), Announcement Details (DR-010-001-03)

### 3.1 Layout

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Structure | Two-panel: action menu (left) + content (right) | User Details, Request Ticket Details |
| Left Panel | Stacked buttons — may include navigation, status actions, and danger actions depending on feature | User Details, Request Ticket Details |
| Content | Read-only display of all data fields grouped into cards | User Details, Request Ticket Details |
| Button Grouping | Left panel buttons grouped into tiers: Navigation (Back, Edit) / Status Actions / Danger Actions (Cancel, Delete) — visually separated | Request Ticket Details |

### 3.2 Data Display

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Field Format | Icon + label + value per field | User Details |
| Empty Fields | Display "—" placeholder (never hidden) | User Details, Request Ticket Details |
| No Inline Editing | All modifications via separate action views | User Details, Request Ticket Details |
| Card Grouping | Data grouped into logical sections (Personal, Work, Account, etc.) | User Details, Request Ticket Details |
| No Truncation | Detail page shows full text for all fields — in contrast to list view truncation | Request Ticket Details (Subject, Description) |
| Attachment Display | Clickable filename linking to file; "No attachment" placeholder when absent | Request Ticket Details |

### 3.3 Navigation

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Back Button | Returns to list page | User Details, Request Ticket Details |
| State Preservation | Back navigation preserves list state (search, filters, page) | User Details, Request Ticket Details |

### 3.4 Status Management on Detail Page

**Applies to:** Detail pages for transactional entities with a status lifecycle (e.g., Request Tickets, Announcements).

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Dual Purpose | Detail page serves both as read-only data view AND as primary location for status transitions | Request Ticket Details, Announcement Details |
| Contextual Buttons | Left panel renders only the action buttons applicable to the current user role and record status | Request Ticket Details, Announcement Details |
| Status Button Visibility | Buttons render after data loads — prevents acting before current status is known | Request Ticket Details, Announcement Details |
| Disabled-Visible Pattern | For status-locked actions (e.g., Update on Sent announcement), button stays visible at 50% opacity rather than hidden — gives users awareness of full action set | Announcement Details (DR-010-001-03) |
| Confirmation Required | All status transitions require a confirmation dialog before proceeding | Request Ticket Details, Announcement Details |
| Action-Specific Dialogs | Each status action has its own dialog title and message (not a generic "Are you sure?") | Request Ticket Details, Announcement Details |
| In-Place Refresh | After a successful status change, page refreshes in place — status badge and action buttons update without full navigation | Request Ticket Details, Announcement Details |
| Redirect on Delete | Soft-delete redirects to the parent list (not in-place refresh, since record is gone) | Announcement Details |
| Concurrent Protection | Server verifies record is still in expected status before applying transition; if changed, error toast + page reload (or redirect if deleted) | Request Ticket Details, Announcement Details |
| Submitter-Exclusive Actions | Close and Reopen actions are available to the submitting employee only — managers/admins cannot act on behalf of the submitter | Request Ticket Details |

### 3.6 Dedicated Status-Transition Screen Pattern

**Applies to:** Standalone screens whose sole purpose is to execute a single irreversible status transition on an existing record (e.g., Send Announcement). Distinct from the Detail page transitions (§3.4) where the transition is triggered via a button + confirmation dialog directly from the read-only Details view.
**Confirmed in:** Send Announcement (DR-010-001-05)

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Layout | Same two-panel layout as Detail/Update views (status-aware action panel + content card) — full visual parity across the entity's status-managed screens | Send Announcement (DR-010-001-05) |
| Active Sibling Button | The screen's own action button on the action panel is highlighted with #f5f5f5 accent background to indicate the current view | Send Announcement (DR-010-001-05) |
| Sibling Buttons Always Interactive | Other action panel buttons (Details, Update, Delete) remain interactive at all times — clicking them performs immediate navigation with no confirmation | Send Announcement (DR-010-001-05) |
| Focused Form Content | The form card displays ONLY the fields relevant to the transition (e.g., Receiver Info for Send) — content authored elsewhere (Title/Description/Label) is NOT shown on this screen and NOT editable here | Send Announcement (DR-010-001-05) |
| No Cancel/Discard Button | Same no-Cancel-button deviation as §4.1 — back arrow and sibling action buttons handle navigation; unsaved changes silently discarded | Send Announcement (DR-010-001-05), inherits from Update Announcement (DR-010-001-04) |
| Single Primary Action Button | Full-width 600px, 40px tall, placed below the form card. **The button label SHOULD match the transition verb (e.g., "Send" for the Send Announcement screen, "Delete" for the Delete Announcement screen).** Avoid generic labels like "Save" on a single-purpose transition screen because they create a semantic mismatch with the screen's intent. | Send Announcement (DR-010-001-05) labeled "Send"; Delete Announcement (DR-010-001-06) labeled "Delete" |
| Primary Action Button Style — Neutral Variant | #010101 background + white text for non-destructive transitions (e.g., Send) | Send Announcement (DR-010-001-05) |
| Primary Action Button Style — Destructive Variant | Danger red background + white text for destructive transitions (e.g., Delete). Visually distinct from the neutral variant to reinforce the gravity of irreversible loss-of-data actions | Delete Announcement (DR-010-001-06) |
| Primary Action Opens Confirmation Dialog | The primary action button does not directly execute the transition — it opens an explicit confirmation dialog with action-specific wording and a recipient/scope summary (or record-identity summary for destructive actions) | Send Announcement (DR-010-001-05), Delete Announcement (DR-010-001-06) |
| Confirmation Default Focus | "Cancel" button has default keyboard focus to prevent accidental confirmation via Enter key | Send Announcement (DR-010-001-05), Delete Announcement (DR-010-001-06) |
| Recipient/Scope/Identity Summary in Dialog | Dialog includes a computed summary of who/what is affected — recipient summary for dispatch actions (e.g., "Everyone" or "X employee(s) selected") or record identity for destructive actions (e.g., the announcement title) — so the user has one final visual confirmation before the irreversible action | Send Announcement (DR-010-001-05), Delete Announcement (DR-010-001-06) |
| Redirect on Success — Destination Depends on Post-Transition State | After successful transition, redirect to wherever the record is reachable in its new state. **For non-destructive transitions** (e.g., Send) where the record still exists post-transition, redirect to the entity's Details page (which renders the post-transition variant). **For destructive transitions** (e.g., Delete) where the record is gone from active views, redirect to the parent List with preserved list state. The transition screen itself is never the success destination, because the transition is no longer applicable on the same record once executed | Send Announcement (DR-010-001-05) -> Details; Delete Announcement (DR-010-001-06) -> List |
| Concurrent Protection | Server re-verifies the record state before applying. **409 Conflict** if status changed for non-destructive transitions; **404 Not Found** if record was soft-deleted | Send Announcement (DR-010-001-05) |
| Concurrent Idempotency (Destructive) | For destructive transitions where two users attempt the same action, the second user's 404 is treated as success (informational toast + redirect to the success destination), not as an error — the desired end state is achieved either way | Delete Announcement (DR-010-001-06) |
| Direct URL Guard | Direct URL to the transition screen for a record no longer in the eligible status (or already in the post-transition state for destructive actions) shows an empty state or redirects to read-only Details with an explanatory toast | Send Announcement (DR-010-001-05), Delete Announcement (DR-010-001-06) |
| Recoverable Server Error | Generic 5xx from the server leaves the record in its pre-transition state (atomic transition) — error toast + user remains on the transition screen (or in the still-open confirmation dialog for destructive actions) with state intact and Retry available | Send Announcement (DR-010-001-05), Delete Announcement (DR-010-001-06) |
| No-Form-State Variant | When the transition screen has only a static warning paragraph (no editable fields), navigation away (back arrow, sibling action buttons) is immediate without any discard confirmation — there is no form state to discard. Distinct from the editable-form variant (e.g., Send with Receiver Info), which also has no Cancel button but does silently discard unsaved input on navigation | Delete Announcement (DR-010-001-06) |

**When to use this pattern vs. Detail-page-only transitions (§3.4):**

| Use this pattern when... | Use §3.4 detail-page transitions when... |
|--------------------------|------------------------------------------|
| The transition needs additional input from the user (e.g., finalize Receiver Info before Send) | The transition is a simple one-click action with no additional input (e.g., Approve / Reject) |
| The transition is irreversible and high-stakes (e.g., dispatching notifications) | The transition is reversible or low-stakes (e.g., In Progress -> On Hold) |
| The transition deserves its own URL for deep-linking, navigation history, and sibling-screen consistency | The transition is best executed in-context from the Details view |

### 3.5 Recipient/Audience Display Pattern

**Applies to:** Detail pages for records with a target audience (e.g., announcements).
**Confirmed in:** Announcement Details (DR-010-001-03)

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Everyone Indicator | When `everyone = true`, display literal "Everyone" | Announcement Details |
| Single Recipient | When 1 specific recipient, display employee full name | Announcement Details |
| Few Recipients | When 2-3 specific recipients, display comma-separated names | Announcement Details |
| Many Recipients | When 4+ specific recipients, display "Name1, Name2 and N others" with hover tooltip listing all | Announcement Details |
| Tooltip Accessibility | Tooltip content accessible via keyboard focus, not hover-only | Announcement Details |

---

## 4. Edit Form Pattern

**Applies to:** Forms for modifying existing records.
**Confirmed in:** Update Leave Request (DR-002-001-03), Update Request Ticket (DR-003-001-03, superseded by DR-003-001-06), Update Announcement (DR-010-001-04)

### 4.1 Confirmed Rules

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Pre-fill | All fields pre-filled with current values on page load | Update Leave Request, Update Request Ticket, Update Announcement |
| Same Validation | Same validation rules as create form | Update Leave Request, Update Request Ticket, Update Announcement |
| Cancel Behavior (default) | Same dirty/clean form behavior as create (back arrow or Cancel triggers discard dialog if dirty) | Update Leave Request, Update Request Ticket |
| Cancel Behavior (no-Cancel-button variant) | Some edit forms (e.g. Update Announcement) do not include a Cancel/Discard button at all. In that variant: navigation away (back arrow, sibling action buttons, browser nav) is immediate with **no confirmation** and unsaved changes are silently discarded. Use only when explicitly designed without a Cancel button. | Update Announcement (DR-010-001-04) |
| Success (default) | Toast "[Record type] has been updated" + redirect to list | Update Leave Request, Update Request Ticket |
| Success (stay-on-page variant) | Toast "[Record type] has been updated" — user stays on the edit form, saved values become the new pre-fill baseline, form returns to clean state (Save disabled until next change). Used for standalone-tab edit forms where rapid multi-pass editing is expected. | Update Announcement (DR-010-001-04) |
| Locked Employee | Employee/submitter field is read-only — cannot change who submitted the record | Update Leave Request, Update Request Ticket |
| Employee Info Card | Always visible for all roles in edit form (unlike create where employee section is hidden for employees) — shows submitter profile read-only, pre-filled on load without requiring selection | Update Leave Request, Update Request Ticket (DR-003-001-06) |
| No Hint Banner | "Select an employee to view their info" hint is never shown in edit context — submitter is always pre-filled | Update Request Ticket (DR-003-001-06) |

### 4.2 Status-on-Edit Rules

| Feature | Rule | Confirmed In |
|---------|------|-------------|
| Leave Requests | Editing reverts status: Approved → Pending | Update Leave Request (DR-002-001-03) |
| Request Tickets | Editing does NOT change status — no re-approval workflow; In Progress stays In Progress | Update Request Ticket (DR-003-001-03, DR-003-001-06) |
| Announcements | Editing does NOT change status — Draft remains Draft; Sent records are immutable and cannot reach the edit form | Update Announcement (DR-010-001-04) |

**Note:** Status-on-edit behavior is feature-specific. Leave requests use an approval lifecycle where edits require re-approval. Request tickets and announcements have no approval workflow, so status is preserved on edit. Announcements additionally enforce immutability after Send (Sent records cannot be edited at all).

### 4.3 Concurrent Edit Protection

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Status Check on Save | Server checks record is still in editable status before saving | Update Leave Request, Update Request Ticket, Update Announcement |
| Conflict Toast (status changed) | If status changed by another user: error toast + redirect to read-only Details (or list) | Update Leave Request, Update Request Ticket, Update Announcement |
| Conflict Toast (record deleted) | If record was soft-deleted by another user: 404 from server > error toast + redirect to list | Update Announcement (DR-010-001-04) |

### 4.4 Attachment Handling in Edit Forms

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Keep | No change to file input → existing attachment preserved | Update Request Ticket (DR-003-001-03, DR-003-001-06) |
| Replace | User selects new file → new file replaces existing | Update Request Ticket (DR-003-001-03, DR-003-001-06) |
| Remove | User clears file input → attachment removed from record | Update Request Ticket (DR-003-001-03, DR-003-001-06) |
| Display | Existing attachment filename shown on page load | Update Request Ticket (DR-003-001-03, DR-003-001-06) |

### 4.5 Edit Form Layout Variants

**Note:** Not all edit forms use the same layout as their corresponding create form. Layout is feature-specific.

| Feature | Edit Form Layout | Create Form Layout | Confirmed In |
|---------|------------------|--------------------|-------------|
| Leave Requests | Full-page form with header Cancel + Save buttons | Same | Update Leave Request (DR-002-001-03) |
| Request Tickets | Two-panel layout: left action panel (162px) + right form card (600px); no header Cancel/Save; back arrow + single bottom Save only | Full-page form with header Cancel + Save + bottom Save | Update Request Ticket (DR-003-001-06) |
| Announcements | Two-panel layout: left action panel (207px) + right form card (600px); no header Cancel/Save; back arrow + single bottom Save only | Full-page form with header Cancel + Save As Draft + Save & Send + bottom dual save | Update Announcement (DR-010-001-04) |

**Request Ticket edit form specific rules:**
- Left action panel (162px) contains 3 buttons: Request Details, Update Request (active), Delete Request
- "Update Request" button highlighted/active to indicate current view
- "Update Request" button only visible when ticket is in an editable status per the user's role
- Page title row includes a ticket reference breadcrumb: `#[ticketId] - [RequestType]`
- Navigation back to list is via back arrow (ArrowLeft) in title row, not a Cancel button
- Single Save button: full-width 600px at bottom of form (background #010101)

**Announcement edit form specific rules:**
- Left action panel (207px) contains 4 buttons: Details, Update Announcement (active), Send Announcement, Delete Announcement (same vertical stack as Announcement Details DR-010-001-03)
- "Update Announcement" button highlighted with #f5f5f5 accent background to indicate current view
- "Update Announcement" is only reachable when status = Draft (Sent announcements are immutable per §9.4); direct URL to edit a Sent record redirects to read-only Details with toast
- **No Cancel/Discard button on this form** — only Save exists. Sibling action buttons (Details, Send, Delete) and the back arrow are interactive at all times; clicking them performs **immediate navigation with no confirmation** and silently discards unsaved changes. This deviates from the §4.1 default Cancel Behavior (no overlay context to dismiss in a standalone-tab form).
- Page title is "Announcement Details" with the announcement title shown in a secondary breadcrumb (consistent with DR-010-001-03)
- Navigation back is via back arrow (ArrowLeft) in title row, not a Cancel button
- Single Save button: full-width 600px at bottom of form card (background #010101, "Save" text)
- Save button enabled only when form is **valid AND dirty** (no-op submissions blocked)
- **Stay-on-page Save:** on successful save the user remains on the Update form; system shows toast "Announcement has been updated"; saved values become the new pre-fill baseline; form returns to clean state with Save disabled until the next change. Deviates from §4.1 default Success behavior — chosen to support rapid multi-pass editing of a single Draft.
- No status change on save — Draft stays Draft (per §4.2)
- No notifications dispatched on edit — notifications fire only on Send (per DR-010-001-02 Rule 10)
- **Edit form contains ONLY the three content fields (Title, Description, Label)** — Receiver Info is intentionally NOT on this form. Audience editing for an existing Draft happens exclusively on the Send Announcement screen (DR-010-001-05). The Update payload submitted to the server contains only the three content fields and never mutates the persisted recipient configuration. Validation rules for the three included fields match the corresponding fields on Create (DR-010-001-02 §3).

---

## 5. Delete Action Pattern

**Applies to:** Deletion of records.
**Confirmed in:** Department List (DR-008-001-01), Position List (DR-008-002-01), Role List (DR-001-004-01), Delete Leave Request (DR-002-001-04), Delete Request Ticket (DR-003-001-04)

### 5.1 Deletion Rules

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Dependency Check | System checks for related records before allowing delete | Department, Position, Role |
| Live Count | Employee/user count is live (not cached) at time of check | Department, Position, Role |
| Blocking Message | "Cannot delete — [X] [entity] assigned" with exact count | Department, Position, Role |
| Confirmation | Confirmation dialog before deletion proceeds | All delete DRs |
| Delete Type | Hard delete (record permanently removed) for organizational entities | Department, Position |
| Soft Delete (accounts) | User accounts are deactivated, not deleted (preserved in system) | User management |
| Soft Delete (transactional) | Record preserved in database with "deleted" flag; removed from all active views, filters, searches, exports | Leave Requests (DR-002-001-04), Request Tickets (DR-003-001-04) |

### 5.2 Transactional Delete Pattern

**Applies to:** Deleting employee-submitted records (leave requests, request tickets, etc.)

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Employee Scope | Employees can only delete their own records in the initial/unactioned status | Leave Request (Pending), Request Ticket (Open) |
| Admin Scope | Admin/managers can delete any record regardless of status | Leave Request, Request Ticket |
| Dialog Summary | Shows record summary (employee name, key fields) — more detailed than simple entity delete | Leave Request, Request Ticket |
| Default Focus | Cancel button (not Delete) — prevents accidental deletion | Leave Request, Request Ticket |
| Danger Style | Delete button uses red/danger style | Leave Request, Request Ticket |
| Focus Trap | Keyboard focus trapped inside dialog | Leave Request, Request Ticket |
| Escape Key | Closes dialog (same as Cancel) | Leave Request, Request Ticket |
| Error Handling | API error: toast + dialog stays open. Already deleted: toast + list refreshes | Leave Request, Request Ticket |
| List State | Active search/filters preserved after deletion | Leave Request, Request Ticket |

**Employee delete scope by feature:**
- Leave Requests: Employees can delete own Pending requests only
- Request Tickets: Employees can delete own Open tickets only
- Pattern: Employee delete is restricted to the initial status (before any management action has been taken)

### 5.3 Resource Restoration on Delete

| Feature | Rule | Confirmed In |
|---------|------|-------------|
| Leave Requests | Deleting Approved request restores leave days to balance | DR-002-001-04 |
| Request Tickets | No resource restoration — tickets have no consumable resource at any status | DR-003-001-04 |

**Note:** Resource restoration is feature-specific. Only applies when the deleted record had consumed a trackable resource (e.g., leave days). Request tickets confirmed as having no consumable resource.

---

## 6. Entity Relationships

### 6.1 Core Organizational Structure

```
Department (EP-008 US-001)
  ├── Has multiple Positions
  └── Has multiple Employees

Position (EP-008 US-002)
  ├── Belongs to organizational structure
  └── Has multiple Employees

Skill (EP-008 US-003)
  └── Assigned to Employees (many-to-many)

User (EP-001 US-005)
  ├── Assigned single Role (1:1)
  ├── Linked to Department (1:1 or 0..1)
  ├── Linked to Position (1:1 or 0..1)
  └── Assigned multiple Skills (many-to-many)

Role (EP-001 US-004)
  ├── Has multiple Permissions
  └── Assigned to multiple Users
```

### 6.2 Leave Management Structure

```
Leave Request (EP-002 US-001)
  ├── Submitted by Employee (references User)
  ├── Has Leave Type (Annual, Sick, Personal, Maternity, Unpaid)
  ├── Has Leave Period (Full Day, Morning Half, Afternoon Half)
  ├── Has Status (Pending, Approved, Rejected, Cancelled)
  └── Calculated Total Days (supports half-day: 0.5, 1.0, 1.5, etc.)
```

### 6.3 Request Ticket Management Structure

```
Request Ticket (EP-003 US-001)
  ├── Submitted by Employee (references User)
  ├── Has Request Type (IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other)
  ├── Has Subject/Title (displayed as "Request" column, truncated)
  ├── Has Status (Open, In Progress, On Hold, Resolved, Closed, Cancelled)
  └── Data Visibility: employees see own only; managers/admins see all
```

### 6.4 Filter Data Sources

| Filter | Source | Values |
|--------|--------|--------|
| Department | EP-008 US-001 | Dynamic from department list |
| Position | EP-008 US-002 | Dynamic from position list |
| Skill | EP-008 US-003 | Dynamic from skill list |
| System Role | EP-001 US-004 | Dynamic from role list |
| Status (User) | Internal | "Active", "Inactive" |
| Status (Leave) | Internal | "Pending", "Approved", "Rejected", "Cancelled" |
| Leave Type | Internal | "Annual", "Sick", "Personal", "Maternity", "Unpaid" |
| Status (Request Ticket) | Internal | "Open", "In Progress", "On Hold", "Resolved", "Closed", "Cancelled" |
| Request Type | Internal | "IT Support", "Facility", "HR Inquiry", "Office Supplies", "Access Request", "Travel & Expense", "Training", "Other" |

---

## 7. Permission & Access Control

**Confirmed across all DRs:**

| Rule | Detail |
|------|--------|
| Permission Source | All access controlled by permissions configured in US-004 (Role & Permission Management) |
| View Permission | Required to access any list/detail page |
| Management Permission | Required for Add, Edit, Delete, Status change actions |
| Button Visibility | "+ Add New" and gear icon hidden for view-only users |
| Direct URL Access | Unauthorized users redirected to fallback page |
| Data Visibility Scoping | Employees see only own records; managers/admins see all (enforced server-side) — applies to Request Tickets (DR-003-001-01) |

---

## 8. Common UI Behaviors

### 8.1 Loading States

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| List Loading | Skeleton rows to prevent layout shift | All list DRs |
| Detail Loading | Skeleton content blocks | User Details |
| Button Loading | Loading spinner on action buttons during processing | Create forms, Export |

### 8.2 Toast Notifications

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Success | Shown after successful create/edit/delete/status change | All action DRs |
| Error (Server) | Shown for server-side validation failures (e.g., duplicate) | Create forms |
| Position | Top-right of screen (standard toast position) | All DRs |

### 8.3 Confirmation Dialogs

| Trigger | Dialog Message | Confirmed In |
|---------|---------------|-------------|
| Cancel dirty form | "Discard unsaved changes?" | Create User, Create Department, Create Request Ticket |
| Delete record | "Are you sure you want to delete [name]?" | Department, Position, Role |
| Status change | Varies by feature | User (Deactivate/Activate) |
| Cancel ticket | Confirmation before cancelling request ticket | Request Tickets (DR-003-001-01) |
| Delete ticket | Confirmation before soft-deleting request ticket | Request Tickets (DR-003-001-01) |

### 8.4 Text Truncation & Tooltip

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Truncation | Text truncated with ellipsis ("...") when exceeding column width | Request Tickets (DR-003-001-01) — Request column |
| Tooltip | On hover, show full untruncated text in a tooltip | Request Tickets (DR-003-001-01) — Request column |
| Max Lines | Single line with text-overflow ellipsis | Request Tickets (DR-003-001-01) |
| Accessibility | Tooltip content accessible via keyboard focus (not hover-only) | Request Tickets (DR-003-001-01) |

---

## 9. Status & State Definitions

### 9.1 User Account Status

| Status | Badge Color | Meaning |
|--------|-------------|---------|
| Active | Green | User can log in and use the system |
| Inactive | Gray | User account disabled, cannot log in |

**Transition:** Active <-> Inactive (toggle via management permission)

### 9.2 Leave Request Status

| Status | Badge Color | Meaning |
|--------|-------------|---------|
| Pending | Amber | Awaiting approval |
| Approved | Green | Leave approved by manager |
| Rejected | Red | Leave rejected by manager |
| Cancelled | Gray | Cancelled by employee or manager |

**Transitions:**
- Pending -> Approved (by manager)
- Pending -> Rejected (by manager)
- Pending -> Cancelled (by employee or manager)
- Approved -> Cancelled (by employee or manager)
- Rejected: terminal state
- Cancelled: terminal state

**Rules:**
- Only Pending and Approved requests can be edited (Approved reverts to Pending)
- Half-day support: Total Days can be decimal (0.5, 1.0, 1.5, etc.)

### 9.3 Request Ticket Status

| Status | Badge Color | Meaning |
|--------|-------------|---------|
| Open | Blue | New ticket, awaiting action |
| In Progress | Amber | Actively being worked on |
| On Hold | Orange | Work paused temporarily |
| Resolved | Green | Solution provided, awaiting submitter confirmation |
| Closed | Gray | Submitter confirmed resolution; finalized |
| Cancelled | Red | Ticket cancelled by submitter or admin |

**Transitions:**
- Open -> In Progress (by manager/admin)
- In Progress -> On Hold (by manager/admin)
- On Hold -> In Progress (by manager/admin — resume)
- In Progress -> Resolved (by manager/admin)
- Resolved -> Closed (by submitter — confirms resolution)
- Resolved -> Open (by submitter — reopens, resolution unsatisfactory)
- Open -> Cancelled (by submitter or manager/admin)
- In Progress -> Cancelled (by submitter or manager/admin)
- On Hold -> Cancelled (by submitter or manager/admin)
- Closed: terminal state
- Cancelled: terminal state

**Rules:**
- Only Open tickets can be edited by the submitting employee
- Managers/admins can edit tickets in Open or In Progress status
- Employees can cancel own tickets in Open, In Progress, On Hold
- Managers/admins can cancel any non-terminal ticket
- Resolved -> Closed requires submitter confirmation
- Submitter can reopen Resolved tickets (reverts to Open)
- Managers/admins can delete any ticket regardless of status (soft delete)
- Data visibility: employees see own tickets only; managers/admins see all

### 9.4 Announcement Status

| Status | Badge Color | Meaning |
|--------|-------------|---------|
| Draft | Gray | Announcement created but not yet sent |
| Sent | Green | Announcement has been sent to employees |

**Transitions:**
- Draft -> Sent (via Send action; also sets Last Announce Date timestamp)
- Draft -> Deleted (via Delete action — soft delete)
- Sent -> Deleted (via Delete action — soft delete, admin cleanup only)

**Rules:**
- Only 2 statuses: Draft and Sent
- Once Sent, announcements are immutable (cannot be edited)
- Delete of Sent announcement is administrative cleanup only; employees retain received copy
- Permission-controlled access via US-004 Permission Management
- Last Announce Date is null for Draft (displayed as "—"), populated only when Sent
- Sending dispatches notifications to recipients via two channels (email + push) asynchronously

**Action Availability by Status (on Detail page):**

| Status | Update | Send | Delete |
|--------|--------|------|--------|
| Draft | Enabled | Enabled | Enabled |
| Sent | Disabled (50% opacity) | Disabled (50% opacity) | Enabled |

**Confirmed in:** DR-009-001-01 (list view), DR-010-001-02 (create), DR-010-001-03 (details), DR-010-001-04 (update), DR-010-001-05 (send), DR-010-001-06 (delete)

---

## 10. Token-Based Setup Pattern

**Applies to:** Pages accessed via secure email links for one-time actions (Set Password, Password Reset).
**Confirmed in:** Set Password (DR-001-001-02)

### 10.1 Token Validation

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Token in URL | Secure token passed as URL parameter from email link | Set Password (DR-001-001-02) |
| Validation on Load | Token validated immediately on page load before displaying form | Set Password |
| Invalid Token | Display error page: "This link is invalid. Please contact your administrator." | Set Password |
| Expired Token | Display error page: "This link has expired. Please contact your administrator." | Set Password |
| Already Used | Display error page with guidance to sign in or use forgot password | Set Password |
| Single-Use | Token invalidated after successful action — cannot be reused | Set Password |

### 10.2 Form Layout

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Card Layout | Same 450px centered card as Sign In page | Set Password |
| Personalization | Welcome message with user's full name from token data | Set Password |
| Pre-filled Email | Email field shown read-only (50% opacity), pre-filled from token | Set Password |
| Password Fields | Password + Confirm Password with independent visibility toggles | Set Password |
| Button Disabled | Submit button disabled until all required fields have input | Set Password |

### 10.3 Validation & Success

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| On Blur | Field-level validation (password policy, password match) | Set Password |
| Inline Errors | Validation errors shown inline below respective field | Set Password |
| Loading State | Button shows loading spinner, form disabled during submission | Set Password |
| Success Redirect | Redirect to Sign In page with success toast | Set Password |
| Success Toast | "[Action] successfully. Please sign in." | Set Password |

### 10.4 Error States (Full-Page)

| Error | Message | Confirmed In |
|-------|---------|-------------|
| Token Invalid | "This link is invalid. Please contact your administrator for a new invitation." | Set Password |
| Token Expired | "This link has expired. Please contact your administrator for a new invitation." | Set Password |
| Already Completed | "Your password has already been set. Please sign in or use 'Forgot Password' if you need to reset it." | Set Password |

---

## 11. Leave Quota Management Pattern

**Applies to:** Admin actions for adjusting employee leave allocations.
**Confirmed in:** Change Leave Quota (DR-002-001-06)

### 11.1 Entry Point & Layout

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Entry Point | User Details page (EP-001 US-005) > Left action panel > "Leave Quota" button | DR-002-001-06 |
| Layout | Two-panel: left action panel (189px) + form card (600px) | DR-002-001-06 |
| Active State | "Leave Quota" button highlighted with accent background when active | DR-002-001-06 |
| Cross-Epic | Feature belongs to EP-002 (Leave Management) but entry point is in EP-001 (User Details) | DR-002-001-06 |

### 11.2 Quota Fields

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Leave Types | Two types: Annual Leave Quota, Sick Leave Quota | DR-002-001-06 |
| Field Type | Number input (supports decimals for half-day precision) | DR-002-001-06 |
| Pre-fill | Both fields pre-filled with current values on page load | DR-002-001-06 |
| Mandatory | Both fields mandatory (marked with asterisk) | DR-002-001-06 |
| Validation | Non-negative, max 365, numeric only | DR-002-001-06 |

### 11.3 Save Behavior

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Button Style | Full-width (600px), black background (#010101), white text | DR-002-001-06 |
| Immediate Effect | Quota changes apply immediately upon save | DR-002-001-06 |
| Stay on Page | After save, form remains on same page with updated values | DR-002-001-06 |
| Success Toast | "Leave quota has been updated" | DR-002-001-06 |
| Dirty Form | Discard confirmation when navigating with unsaved changes | DR-002-001-06 |

### 11.4 Balance Calculation

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Formula | Leave Days Remaining = Quota - Sum of Approved Leave Days | DR-002-001-06 |
| Negative Allowed | If quota < used days, balance becomes negative (allowed) | DR-002-001-06 |
| No Retroactive | Changing quota does NOT affect already-approved requests | DR-002-001-06 |

---

## 12. Mobile Authentication Pattern

**Applies to:** Mobile app authentication screens (Sign In, Forgot Password, Session Management).
**Confirmed in:** MOBILE-APP Sign In (DR-001-001-01)
**Platform:** MOBILE-APP only

### 12.1 Sign In Screen

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Layout | Dark gradient header (logo, heading, tagline) + light form card overlay | DR-001-001-01 |
| Form Card | Rounded corners (40px radius), shadow-md, #f5f5f5 background | DR-001-001-01 |
| Input Icons | @ icon for email, lock icon for password | DR-001-001-01 |
| Password Toggle | Eye-off/eye icon on right side of password field | DR-001-001-01 |
| Remember Me | Checkbox, unchecked by default, extends session token validity | DR-001-001-01 |
| Sign In Button | Full-width, dark (#171717), disabled until both fields have input | DR-001-001-01 |
| Biometric Placeholder | Fingerprint icon visible below Sign In button (non-functional in v1.0) | DR-001-001-01 |
| Forgot Password Link | Blue text (#1d4ed8) below form card | DR-001-001-01 |

### 12.2 Mobile-Specific Behaviors

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Keyboard Optimization | Email field shows email keyboard (with @ and .com keys) | DR-001-001-01 |
| Auto-focus | Email field auto-focused on screen load | DR-001-001-01 |
| Touch Targets | Minimum 44x44 points for all interactive elements | DR-001-001-01 |
| Password Autofill | Support iOS Keychain / Android Autofill | DR-001-001-01 |
| Haptic Feedback | Optional haptic on error (device-dependent) | DR-001-001-01 |

### 12.3 Session & Security

| Rule | Detail | Confirmed In |
|------|--------|-------------|
| Token Storage | Platform-native secure storage (iOS Keychain / Android Keystore) | DR-001-001-01 |
| Session Persistence | Session persists across app restarts until timeout or logout | DR-001-001-01 |
| Concurrent Sessions | Multiple sessions allowed across devices (mobile + web) | DR-001-001-01 |
| Remember Me Effect | Extended session token validity in secure storage | DR-001-001-01 |

### 12.4 Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Forgot Password | Email link with token | OTP code entered in-app |
| Session Storage | Browser cookies/localStorage | Secure device storage (Keychain/Keystore) |
| Remember Me | Checkbox option | Checkbox option (affects secure storage duration) |
| Biometric Auth | Not supported | Future enhancement (fingerprint icon placeholder) |

### 12.5 Error Handling

| Error Type | Display Method | Message |
|------------|----------------|---------|
| Invalid credentials | Toast | "Invalid email or password" |
| Account deactivated | Toast | "Your account has been deactivated. Contact your administrator." |
| No login permission | Toast | "You do not have permission to access this system." |
| Account locked | Toast | "Account temporarily locked. Try again in X minutes." |
| Network failure | Toast | "Unable to connect. Please check your network connection." |
| Invalid email format | Inline (below field) | "Please enter a valid email address" |

---

## 13. Mobile Dashboard Pattern

**Applies to:** Mobile app dashboard screens that summarize data with quick actions.
**Confirmed in:** MOBILE-APP Leave Requests Dashboard (DR-002-001-01)
**Platform:** MOBILE-APP only

### 13.1 Dashboard Layout Structure

| Component | Position | Purpose | Confirmed In |
|-----------|----------|---------|--------------|
| Header | Top (70px) | User avatar + department/position badges + company name + notification bell | DR-002-001-01 |
| Page Title | Below header | Icon prefix + title text (e.g., "[Cal] Leave Requests") | DR-002-001-01 |
| Hero Card | Below title | Promotional CTA with heading, subtitle, illustration, and primary action button | DR-002-001-01 |
| Metric Cards | Below hero | 2x2 grid of summary metrics with large values and decorative icons | DR-002-001-01 |
| Tab Switcher | Below metrics | Two-tab toggle for filtering content (e.g., Upcoming/History) | DR-002-001-01 |
| Item List | Below tabs | Scrollable list of recent items | DR-002-001-01 |
| Bottom Navigation | Fixed bottom (76px) | 4-icon persistent app navigation | DR-002-001-01 |

### 13.2 Hero Card Pattern

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Layout | Icon + heading + subtitle on left, decorative illustration on right | DR-002-001-01 |
| CTA Button | Primary dark button (144x36) with icon prefix | DR-002-001-01 |
| Purpose | Promote primary action (e.g., "Apply for Leave") | DR-002-001-01 |
| Dimensions | Full width (369px), height ~146px | DR-002-001-01 |

### 13.3 Balance/Metric Card Grid

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Layout | 2x2 grid (179.5px per card, 69px height, 10px gap) | DR-002-001-01 |
| Card Content | Label (12px muted) + value (24px semibold) + decorative icon (bottom-right) | DR-002-001-01 |
| Data Source | Server-calculated values (e.g., Remaining = Quota - Used) | DR-002-001-01 |

### 13.4 Tab-Based Filtering

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Tab Count | Two tabs (binary toggle) | DR-002-001-01 |
| Active State | Filled background vs outlined | DR-002-001-01 |
| Content Switching | Tapping tab updates list content below | DR-002-001-01 |
| State Persistence | Tab state persists during session | DR-002-001-01 |

### 13.5 Mobile List Item Format

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Row Height | 89px for comfortable touch targets | DR-002-001-01 |
| Layout | Type indicator (left bar) + main content + status icon + duration (right) | DR-002-001-01 |
| Date Range | "F [date]" and "T [date]" format for compact display | DR-002-001-01 |
| Touch Action | Tap entire row to navigate to details | DR-002-001-01 |

### 13.6 Pull-to-Refresh

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Gesture | Standard iOS/Android pull-down gesture | DR-002-001-01 |
| Indicator | Refresh spinner at top during reload | DR-002-001-01 |
| Data Update | Fresh data fetched from backend | DR-002-001-01 |
| Offline | Cached data displayed; refresh fetches new | DR-002-001-01 |

### 13.7 Bottom Navigation

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Position | Fixed at bottom (76px height) | DR-002-001-01 |
| Icons | 4 icons: Home, Calendar (Leave), Documents, Settings | DR-002-001-01 |
| Active State | Yellow/highlighted background on active icon | DR-002-001-01 |
| Touch Targets | Minimum 44x44 points per icon | DR-002-001-01 |

### 13.8 Mobile vs WEB-APP Differences

| Aspect | WEB-APP Pattern | MOBILE-APP Pattern |
|--------|-----------------|-------------------|
| Primary View | Full table list with filters | Dashboard summary + list |
| Metrics Display | Not on list page | Prominent card grid |
| Filtering | Multi-select chip filters | Tab-based (binary) |
| Create Action | "+ Add New" button | Hero card CTA |
| Row Actions | Gear icon dropdown | Tap row for details |
| Navigation | Sidebar | Bottom navigation |
| Refresh | Page reload / F5 | Pull-to-refresh gesture |

---

## 14. Mobile Create Form Pattern

**Applies to:** Mobile app screens for creating new records.
**Confirmed in:** MOBILE-APP Create Leave Request (DR-002-001-02)
**Platform:** MOBILE-APP only

### 14.1 Layout Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP | Confirmed In |
|--------|---------|------------|--------------|
| Form Layout | Full-page with multiple cards | Single form card | DR-002-001-02 |
| Submit Button | Header Save + Bottom Save (dual) | Header "Done" button only | DR-002-001-02 |
| Cancel | Header Cancel button | Back arrow with discard check | DR-002-001-02 |
| Date Picker | Custom calendar component | Native iOS/Android date picker | DR-002-001-02 |
| Dropdowns | Custom searchable dropdown | Native ActionSheet/BottomSheet | DR-002-001-02 |
| Employee Selection | Admin can select employee | Employees submit own requests only (v1.0) | DR-002-001-02 |

### 14.2 Mobile Form Field Layout

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Field Row Structure | Label (100px) + Value/Placeholder + Icon | DR-002-001-02 |
| Field Separator | 1px horizontal line between rows (except last) | DR-002-001-02 |
| Icon Indicators | CaretDown for dropdowns, CalendarDots for dates, CloudArrowUp for file upload | DR-002-001-02 |
| Placeholder Style | 14px regular, muted foreground (#64748b) | DR-002-001-02 |
| Value Style | 14px regular, text foreground (#020817) | DR-002-001-02 |

### 14.3 Mobile Form Header

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Back Arrow | 18x18 ArrowLeft icon, navigates back with discard check if dirty | DR-002-001-02 |
| Title | Screen title (14px medium), left of center | DR-002-001-02 |
| Done Button | Pill-shaped, white bg, shadow, disabled until all mandatory fields filled | DR-002-001-02 |
| Done Button Position | Header right | DR-002-001-02 |

### 14.4 Mobile Form Validation

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Done Button State | Disabled until all mandatory fields completed | DR-002-001-02 |
| Error Display | Inline below affected field | DR-002-001-02 |
| Date Validation | From Date >= today; To Date >= From Date | DR-002-001-02 |
| File Validation | Format + size check before upload (PDF, PNG, JPG, DOCX; max 5MB) | DR-002-001-02 |

### 14.5 Mobile Form Success Path

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Success Toast | "[Record type] has been submitted" | DR-002-001-02 |
| Redirect | Navigate back to Dashboard (not list) | DR-002-001-02 |
| Data Refresh | Dashboard refreshes to show new record | DR-002-001-02 |

### 14.6 Mobile Form Cancel/Discard

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Clean Form | Back arrow navigates immediately without confirmation | DR-002-001-02 |
| Dirty Form | Back arrow shows "Discard unsaved changes?" confirmation | DR-002-001-02 |

---

## 15. Mobile Check-In/Out Pattern

**Applies to:** Mobile app GPS-verified check-in/out functionality.
**Confirmed in:** MOBILE-APP Check-In/Out (DR-003-001-01)
**Platform:** MOBILE-APP only

### 15.1 Dashboard Widget Layout

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Entry Point | Embedded card/widget on Home Dashboard (not separate screen) | DR-003-001-01 |
| Widget Content | Location status + action button + status text + metrics row | DR-003-001-01 |
| Button Toggle | Check-In button shown when not checked in; Check-Out button when checked in | DR-003-001-01 |
| Metrics Row | Streak count + Monthly days count at bottom of widget | DR-003-001-01 |

### 15.2 GPS Location Verification

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Verification Method | GPS coordinates compared to configured office location | DR-003-001-01 |
| Radius Tolerance | 50 meters (configurable by admin) | DR-003-001-01 |
| Accuracy Threshold | GPS accuracy must be <= 50m; otherwise treated as unavailable | DR-003-001-01 |
| Distance Formula | Haversine formula for coordinate distance calculation | DR-003-001-01 |
| Permission Required | Location permission must be granted; prompt to settings if denied | DR-003-001-01 |

### 15.3 Location Status Indicators

| Status | Badge Color | Condition | Button State |
|--------|-------------|-----------|--------------|
| At Office | Green (#22c55e) | Within 50m radius | Enabled |
| Not at Office | Gray (#6b7280) | Outside 50m radius | Disabled |
| Location Unavailable | Yellow (#eab308) | GPS unavailable/denied | Disabled |
| Acquiring Location | N/A (spinner) | GPS in progress | Disabled |

### 15.4 Check-In/Out Rules

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Check-In Availability | 24/7 (no time window restriction) | DR-003-001-01 |
| Late Threshold | Check-in after 9:00 AM marked as "Late" (configurable) | DR-003-001-01 |
| Check-Out Optional | Employees are not required to check out | DR-003-001-01 |
| Auto Check-Out | System auto-closes session at 11:00 PM if not checked out | DR-003-001-01 |
| Auto Check-Out Flag | Auto check-outs marked distinctly from manual | DR-003-001-01 |
| Multiple Sessions | Allowed per day; monthly count increments once per calendar day | DR-003-001-01 |

### 15.5 Attendance Streak Calculation

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Definition | Consecutive workdays (Mon-Fri) with at least one check-in | DR-003-001-01 |
| Weekends | Do NOT break streak (Sat-Sun excluded from calculation) | DR-003-001-01 |
| Holidays | Do NOT break streak (company holidays excluded) | DR-003-001-01 |
| Missed Workday | Resets streak to zero | DR-003-001-01 |
| Milestone Celebrations | Confetti animation at 5, 10, 20, 30, 50, 100 days | DR-003-001-01 |
| Server-Side Calculation | Streak calculated on server for consistency across devices | DR-003-001-01 |

### 15.6 UX Behaviors

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Button Size | Minimum 56x56 points (larger than standard 44x44) | DR-003-001-01 |
| Haptic Feedback | Triggered on successful check-in/out | DR-003-001-01 |
| Real-Time Location | Button enables/disables as employee moves in/out of radius | DR-003-001-01 |
| Manual Refresh | Refresh button allows re-acquiring GPS location | DR-003-001-01 |
| Skeleton Loading | Widget shows skeleton during GPS acquisition | DR-003-001-01 |
| Actionable Errors | Error messages include action (e.g., "tap to open settings") | DR-003-001-01 |

### 15.7 Data Captured

| Data Point | Purpose | Confirmed In |
|------------|---------|--------------|
| Timestamp (UTC) | When check-in/out occurred | DR-003-001-01 |
| GPS Coordinates | Employee location at check-in/out (audit trail) | DR-003-001-01 |
| GPS Accuracy | Device-reported accuracy in meters | DR-003-001-01 |
| Late Flag | Whether check-in was after threshold | DR-003-001-01 |
| Auto-Checkout Flag | Whether system auto-closed the session | DR-003-001-01 |

---

## 16. Calendar Matrix View Pattern

**Applies to:** Pages displaying data in a calendar-style matrix (rows = entities, columns = dates).
**Confirmed in:** WEB-APP Attendance List (DR-004-001-01)
**Platform:** WEB-APP

### 16.1 Matrix Layout

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Structure | Rows = entities (e.g., employees), Columns = dates (1-31) | DR-004-001-01 |
| Fixed Column | First column (entity name) fixed, doesn't scroll horizontally | DR-004-001-01 |
| Scrollable Dates | Date columns horizontally scrollable | DR-004-001-01 |
| Column Header | Day of week (Mon, Tue...) + date number | DR-004-001-01 |
| Cell Width | 44px minimum for date columns | DR-004-001-01 |
| Row Height | 44px for comfortable interaction | DR-004-001-01 |
| No Summary Columns | Keep table clean; summary shown elsewhere if needed | DR-004-001-01 |

### 16.2 Cell Status Display

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Icon-Based | Single character/icon per cell (compact for narrow columns) | DR-004-001-01 |
| Color-Coded | Background color indicates status category | DR-004-001-01 |
| Hover Tooltip | Details shown on hover, not in cell | DR-004-001-01 |
| Click: No Action | Cells are hover-only; click does nothing | DR-004-001-01 |

**Attendance-specific status icons:**

| Icon | Meaning | Cell Background | Confirmed In |
|------|---------|-----------------|--------------|
| ✓ | Present (on-time) | Green | DR-004-001-01 |
| L | Late | Amber | DR-004-001-01 |
| A | Absent | Red | DR-004-001-01 |
| — | Weekend | Gray (muted) | DR-004-001-01 |
| H | Holiday | Blue (muted) | DR-004-001-01 |
| AL | Annual Leave | Purple | DR-004-001-01 |
| SL | Sick Leave | Orange | DR-004-001-01 |
| ½ | Half-day Leave | Light blue | DR-004-001-01 |

### 16.3 Tooltip Content

| Context | Tooltip Shows | Confirmed In |
|---------|---------------|--------------|
| Attendance Day | Date, check-in time, check-out time, hours worked, status | DR-004-001-01 |
| Leave Day | Date, leave type, approver name | DR-004-001-01 |
| Weekend/Holiday | Date, "Weekend" or holiday name | DR-004-001-01 |
| Absent Day | Date, "No check-in recorded", "Absent" | DR-004-001-01 |

### 16.4 Matrix Filters

| Filter | Type | Behavior | Confirmed In |
|--------|------|----------|--------------|
| Date Range | Month/Year picker | Selects which month to display | DR-004-001-01 |
| Entity Filter | Multi-select (e.g., Department) | Filters rows | DR-004-001-01 |
| Status Filter | Multi-select | Shows entities with ≥1 day matching selected status(es) | DR-004-001-01 |
| Search | Text input | Filters by entity name | DR-004-001-01 |

### 16.5 Matrix Export

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Bulk Export | Top-right button exports all visible rows | DR-004-001-01 |
| Individual Export | Gear icon per row for single-entity export | DR-004-001-01 |
| Format | Excel (.xlsx) | DR-004-001-01 |
| Scope | Respects active filters | DR-004-001-01 |

### 16.6 Weekend/Holiday Display

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Show Weekends | Weekend columns shown (not hidden) | DR-004-001-01 |
| Muted Style | Gray background to distinguish from workdays | DR-004-001-01 |
| Weekly Rhythm | Keeping weekends visible helps pattern spotting | DR-004-001-01 |

### 16.7 Data Integration

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Cross-Module | Matrix can integrate data from multiple sources (e.g., attendance + leave) | DR-004-001-01 |
| Read-Only | Matrix view is display-only; no inline editing | DR-004-001-01 |
| Real-Time | Data refreshed on page load; no live updates | DR-004-001-01 |

---

## 17. Mobile Announcement Widget Pattern

**Applies to:** Mobile app widgets that display read-only data from WEB-APP sources.
**Confirmed in:** MOBILE-APP Top 5 Latest Announcements (DR-004-001-01)
**Platform:** MOBILE-APP only

### 17.1 Widget Layout

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Entry Point | Embedded on home screen dashboard (not a separate screen) | DR-004-001-01 |
| Widget Header | Title + "View All" link (right-aligned) | DR-004-001-01 |
| Content Area | Vertical stack of item cards (up to 5) | DR-004-001-01 |
| Card Content | Unread indicator (left) + Title + Date (right) + Preview text (below) | DR-004-001-01 |

### 17.2 Unread/Read Indicator Pattern

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Indicator Type | Small colored dot (blue or primary accent) | DR-004-001-01 |
| Unread State | Dot visible on left side of card | DR-004-001-01 |
| Read State | Dot hidden (no indicator shown) | DR-004-001-01 |
| Mark as Read | Triggered when user taps card (not when widget is viewed) | DR-004-001-01 |
| Persistence | Read status tracked per user per item (server-side) | DR-004-001-01 |
| No Reverse | Once read, cannot transition back to unread | DR-004-001-01 |
| Cross-Device Sync | Read status synced across mobile and web | DR-004-001-01 |

### 17.3 Cross-Platform Data Consumer Pattern

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Data Source | WEB-APP creates data, MOBILE-APP consumes | DR-004-001-01 |
| Status Filter | Only items with "Sent" status displayed (Draft excluded) | DR-004-001-01 |
| Targeting Filter | Only items targeting current user (Everyone OR specific) | DR-004-001-01 |
| Sort Order | Newest first (by sent_date descending) | DR-004-001-01 |
| Display Limit | Widget shows limited count (e.g., 5); full list via "View All" | DR-004-001-01 |

### 17.4 Widget Display States

| State | Condition | What User Sees | Confirmed In |
|-------|-----------|----------------|--------------|
| Loading | Initial fetch or refresh | Skeleton cards (3-5 placeholders) | DR-004-001-01 |
| Empty | No items for this user | Icon + "No [items] yet" message | DR-004-001-01 |
| Populated | 1-5 items available | Item cards in vertical stack | DR-004-001-01 |
| Error | Network/server failure | Error message + "Tap to retry" | DR-004-001-01 |
| Refreshing | Pull-to-refresh in progress | Spinner at top of widget | DR-004-001-01 |

### 17.5 Widget Interactions

| Interaction | Behavior | Confirmed In |
|-------------|----------|--------------|
| Tap Card | Navigate to detail screen + mark as read | DR-004-001-01 |
| Tap View All | Navigate to full list screen | DR-004-001-01 |
| Pull-to-Refresh | Fetch fresh data from server | DR-004-001-01 |
| Offline | Show cached data with "Last updated" timestamp | DR-004-001-01 |

### 17.6 Mobile Full List Screen Pattern

**Applies to:** Full-screen list views accessed via "View All" from home widgets.
**Confirmed in:** MOBILE-APP Announcement List (DR-004-001-02)

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Entry Point | "View All" link from home widget OR back navigation from detail | DR-004-001-02 |
| Header Layout | Back arrow (ArrowLeft 18x18) + Screen title (left-aligned) | DR-004-001-02 |
| No Display Limit | Shows all items (unlike widget which limits to 5) | DR-004-001-02 |
| Same Card Format | Cards identical to widget cards (title, preview, date, badge, unread dot) | DR-004-001-02 |
| Scrollable Content | Full-screen scrollable list with 20px gap between cards | DR-004-001-02 |
| Bottom Navigation | Persistent 4-icon nav bar (76px) at bottom | DR-004-001-02 |
| Back Navigation | ArrowLeft returns to previous screen (home or detail) | DR-004-001-02 |
| No Search/Filter | Full list has no search or filter UI in v1.0 | DR-004-001-02 |
| Category Badges | Dynamic badges from data (Holiday, Announcement, Reminder, Meeting, etc.) | DR-004-001-02 |

**Layout Dimensions (from Figma):**

| Component | Height | Details |
|-----------|--------|---------|
| Header | 70px | Back arrow + title, padding 12px horizontal |
| Content Area | Flexible | Scrollable list |
| Bottom Nav | 76px | Fixed at bottom |

### 17.7 Mobile Announcement Detail Pattern

**Applies to:** Full-screen detail view for a single announcement.
**Confirmed in:** MOBILE-APP Announcement Detail (DR-004-001-03)

| Rule | Detail | Confirmed In |
|------|--------|--------------|
| Entry Point | Tap announcement card from widget (DR-004-001-01) or list (DR-004-001-02) | DR-004-001-03 |
| Header Layout | Back arrow (ArrowLeft 18x18) + truncated title (14px medium) | DR-004-001-03 |
| Title Section | Full title (24px semibold) + date badge + category badge | DR-004-001-03 |
| Date Badge | White background, calendar icon, "DDth MMM YYYY" format | DR-004-001-03 |
| Category Badge | Dark background (#171717), white text, pill shape | DR-004-001-03 |
| Content Card | White background, 10px radius, 20px padding, rich text content | DR-004-001-03 |
| Rich Text Support | Bold, paragraphs (12px spacing), bulleted lists, 16px line height | DR-004-001-03 |
| Mark as Read | Triggered when detail screen is viewed (not on list view) | DR-004-001-03 |
| No Edit Actions | Read-only view; announcements immutable after send | DR-004-001-03 |
| Bottom Navigation | Persistent 4-icon nav bar (76px), Home icon active | DR-004-001-03 |

**Layout Dimensions (from Figma):**

| Component | Dimensions | Details |
|-----------|------------|---------|
| Header | 70px height | Back arrow + truncated title, 12px horizontal padding |
| Content Area | Flexible | Title + badges + content card, scrollable |
| Content Card | Auto height | White bg, 10px radius, 20px internal padding |
| Bottom Nav | 76px height | Fixed at bottom |

---

## 18. Open/Pending Decisions

These items are NOT yet confirmed and should still be asked during DR creation:

| Question | Modules Affected | Status |
|----------|-----------------|--------|
| Export format | All lists with export | Pending PO |
| Status badge exact color tokens | User List, Leave Requests, Request Tickets (6 statuses) | Pending Design |
| Gender dropdown values | Create User | Pending PO |
| Phone number format restrictions | Create User | Pending PO |
| Avatar dimensions / auto-resize | Create User, User Details | Pending PO |
| Department-Position dependency (should selecting Dept filter Positions?) | Create User | Pending PO |
| Skills selector UX pattern | Create User | Pending Design |
| Cover image per-user customization | User Details | Pending Design |
| Invitation token expiration duration | Set Password | Pending PO |
| Password policy (min length, complexity) | Set Password, Password Reset, Mobile Sign In | Pending PO |
| Session timeout duration (standard + extended) | Mobile Sign In, WEB-APP Sign In | Pending PO |
| Failed login attempts threshold | Mobile Sign In, WEB-APP Sign In | Pending PO |
| Account lockout duration | Mobile Sign In, WEB-APP Sign In | Pending PO |
| Mobile tagline text (currently placeholder) | Mobile Sign In | Pending PO |
| Biometric authentication timeline | Mobile Sign In | Pending PO |

---

## Changelog

| Date | Change | Reason |
|------|--------|--------|
| 2026-03-27 | Initial creation | Consolidated patterns from existing DRs across EP-001, EP-002, EP-008 |
| 2026-04-01 | Removed EP-003 Service Ticket patterns | Service Ticket module removed from knowledge base |
| 2026-04-01 | Added Request Tickets List (DR-003-001-01) patterns | New 6-status lifecycle, data visibility scoping, submitter confirmation, Request Type values, gear action matrix |
| 2026-04-01 | Resolved "Request column tooltip on hover" decision; added Section 8.4 | Tooltip on hover confirmed for truncated Request column text in DR-003-001-01 |
| 2026-04-01 | Added Create Request Ticket (DR-003-001-02) patterns | Dual-role form, file upload, cancel/discard, non-searchable dropdown for Request Type, active-employees-only dropdown rule |
| 2026-04-03 | Added Update Request Ticket (DR-003-001-03) patterns | Edit form: status NOT reverted (no re-approval), employee info card always visible for all roles, attachment keep/replace/remove handling, concurrent edit conflict toast |
| 2026-04-03 | Added Delete Request Ticket (DR-003-001-04) patterns | Extended §5 transactional delete to cover Request Tickets; confirmed employee delete scope = Open only; confirmed no resource restoration for tickets; updated §5.2 to generalize employee scope rule across features |
| 2026-04-03 | Added Request Ticket Details (DR-003-001-05) patterns | Extended §3 Detail/Profile View pattern: added §3.4 Status Management on Detail Page (dual-purpose detail+status page, contextual buttons, in-place refresh, submitter-exclusive Close/Reopen, concurrent protection); added no-truncation and attachment display rules to §3.2; updated §3.1 button grouping rule |
| 2026-04-03 | Updated Update Request Ticket patterns (DR-003-001-06 supersedes DR-003-001-03) | Added §4.5 Edit Form Layout Variants: Request Ticket edit uses two-panel layout (left action panel with Request Details/Update Request/Delete Request, no header Cancel/Save, back arrow + single bottom Save, ticket reference breadcrumb in title). Added "No Hint Banner" rule to §4.1. Updated §4.1 Employee Info Card rule to clarify pre-filled-on-load behavior. |
| 2026-04-10 | Added Token-Based Setup Pattern (§10) from Set Password (DR-001-001-02) | New pattern for email-triggered one-time actions: token validation, single-use tokens, personalized welcome, pre-filled read-only email, password confirmation, full-page error states for invalid/expired/used tokens |
| 2026-04-10 | Added Leave Quota Management rules from Change Leave Quota (DR-002-001-06) | New feature: quota editing via User Details page; two leave types (Annual, Sick); decimal support; immediate effect; negative balance allowed; cross-epic entry point pattern |
| 2026-04-16 | Added Mobile Authentication Pattern (Section 12) from MOBILE-APP Sign In (DR-001-001-01) | First MOBILE-APP DR: mobile sign-in layout, keyboard optimization, touch targets, secure device storage (Keychain/Keystore), OTP-based forgot password difference from WEB-APP, biometric placeholder, mobile-specific error handling |
| 2026-04-16 | Added Mobile Dashboard Pattern (Section 13) from MOBILE-APP Leave Requests Dashboard (DR-002-001-01) | First mobile dashboard: hero card CTA pattern, 2x2 balance card grid, tab-based filtering (Upcoming/History), request row format (type indicator + status + duration + dates), pull-to-refresh, bottom navigation |
| 2026-04-17 | Added Mobile Create Form Pattern (Section 14) from MOBILE-APP Create Leave Request (DR-002-001-02) | Mobile form layout (single card, header Done button, back arrow navigation), native pickers, field row structure with icons, mobile-specific validation and success path |
| 2026-04-18 | Added Mobile Check-In/Out Pattern (Section 15) from MOBILE-APP Check-In/Out (DR-003-001-01) | Dashboard widget layout, GPS location verification (50m radius, Haversine formula), location status indicators, check-in/out rules (late threshold, auto check-out at 11 PM, multiple sessions), attendance streak calculation (excludes weekends/holidays), UX behaviors (56x56 button, haptic feedback, real-time location updates) |
| 2026-04-21 | Added Calendar Matrix View Pattern (Section 16) from WEB-APP Attendance List (DR-004-001-01) | New pattern for calendar-style matrix views: rows=entities, columns=dates, fixed first column, cell status icons with color coding, hover tooltips (no click action), status filter logic (≥1 day matching), weekend display (shown but muted), cross-module data integration (attendance + leave), bulk + individual Excel export |
| 2026-04-23 | Added Late Arrival Threshold Configuration (DR-004-001-02) notes | Extends Leave Quota Management Pattern (§11) — single-field org settings page with time picker, full-width save button, stay-on-page behavior, dirty form check; late threshold affects attendance calculation in DR-004-001-01 |
| 2026-04-25 | Added Announcement List (DR-009-001-01) patterns | New EP-009 Organization Settings module: 2-status lifecycle (Draft/Sent), immutable-after-send rule, status-based gear actions (Draft: Edit/Send/Delete; Sent: Delete only), permission-controlled access; added §9.4 Announcement Status; updated §1 List View Pattern confirmed list |
| 2026-04-25 | Added Create Announcement (DR-010-001-02) patterns | New §2.9 Dual-Action Save Pattern (Save As Draft vs Save & Send with different statuses, dual button placement header+bottom); New §2.10 Mutually Exclusive Toggle Pattern (radio-button behavior for toggles, conditional field enabling at 50% opacity); Announcement receiver selection (Everyone vs Specific one) |
| 2026-04-25 | Added Mobile Announcement Widget Pattern (§17) from MOBILE-APP Top 5 Latest Announcements (DR-004-001-01) | New pattern for mobile read-only widgets: widget layout (header + View All + cards), unread/read indicator (dot, mark on tap, server-side tracking, cross-device sync), cross-platform data consumer (WEB-APP creates, MOBILE-APP consumes, status/targeting filters), widget display states (loading/empty/populated/error/refreshing), widget interactions (tap card, View All, pull-to-refresh, offline caching) |
| 2026-04-25 | Added Mobile Full List Screen Pattern (§17.6) from MOBILE-APP Announcement List (DR-004-001-02) | Full-screen list accessed via "View All": back arrow header, no display limit, same card format as widget, scrollable content with 20px gap, persistent bottom nav, no search/filter in v1.0, dynamic category badges |
| 2026-04-25 | Added Mobile Announcement Detail Pattern (§17.7) from MOBILE-APP Announcement Detail (DR-004-001-03) | Detail view layout: header with truncated title, full title + date/category badges, white content card with rich text (bold, lists, paragraphs), mark as read on view, read-only (no edit actions) |
| 2026-05-05 | Added Announcement Details (DR-010-001-03) patterns | Extended §3 Detail/Profile View Pattern: confirmed Announcement Details applies the two-panel layout with status-aware action panel (Details/Update/Send/Delete); added "Disabled-Visible Pattern" rule to §3.4 (status-locked actions render at 50% opacity rather than hidden); added "Redirect on Delete" rule to §3.4 distinguishing in-place refresh (status change) from list redirect (delete); added new §3.5 Recipient/Audience Display Pattern (Everyone / single name / comma-separated / "Name1, Name2 and N others" with tooltip); extended §9.4 Announcement Status with action-availability matrix and Last Announce Date timestamp rule |
| 2026-05-06 | Added Update Announcement (DR-010-001-04 v1.1) deviations | Extended §4 Edit Form Pattern: added §4.1 no-Cancel-button variant of Cancel Behavior (immediate navigation with silent discard, no confirmation) and §4.1 stay-on-page Success variant (toast only, saved values become new pre-fill baseline, form returns to clean state); extended §4.5 with announcement-specific edit form rules (Update Announcement only reachable for Draft, Save enabled only when valid AND dirty, no status change on save) |
| 2026-05-06 | Added Send Announcement (DR-010-001-05) patterns | New §3.6 Dedicated Status-Transition Screen Pattern: standalone screen whose sole purpose is executing a single irreversible status transition (e.g., Send) — same two-panel layout as Detail/Update views, focused form content (only the fields relevant to the transition), no Cancel/Discard button (back arrow + sibling action buttons handle navigation), single Save button that opens an explicit confirmation dialog with action-specific wording and recipient/scope summary, default keyboard focus on Cancel in the dialog, redirect to Details page on success (does NOT stay on the transition screen because the transition is no longer applicable on the same record once executed), concurrent protection (409 if status changed, 404 if deleted), direct URL guard, recoverable 5xx (atomic transition leaves record in pre-transition state). Documented when to use this pattern vs. §3.4 detail-page transitions. Updated §9.4 confirmed-in list to include DR-010-001-05. |
| 2026-05-06 | DR-010-001-05 v1.1 — corrected Send Announcement primary button label | The bottom action button is **"Send"**, not "Save". §3.6 generalized: renamed "Single Save Button" -> "Single Primary Action Button" and "Save Opens Confirmation Dialog" -> "Primary Action Opens Confirmation Dialog". Added the rule that the primary action button label SHOULD match the transition verb (e.g., "Send" for the Send Announcement screen) — generic labels like "Save" create a semantic mismatch on a single-purpose transition screen. The original Figma frame showed "Save" but is treated as a design artifact to be corrected; the DR is the authoritative source. |
| 2026-05-06 | Added Delete Announcement (DR-010-001-06) patterns | Extended §3.6 Dedicated Status-Transition Screen Pattern with: (1) **Primary Action Button Style — Neutral vs Destructive Variants** (neutral #010101 for Send-style transitions, danger red for Delete-style destructive transitions); (2) **Redirect on Success — Destination Depends on Post-Transition State** rule: non-destructive transitions redirect to Details (record still exists post-transition), destructive transitions redirect to parent List (record gone from active views); (3) **Concurrent Idempotency (Destructive)** rule: for destructive transitions, a second-user 404 is treated as success (informational toast + redirect to success destination), not an error; (4) **No-Form-State Variant** rule: when the transition screen has only a static warning (no editable fields), navigation away is immediate without any discard logic. Generalized the §3.6 dialog summary rule to cover both Recipient/Scope summaries (dispatch actions) and Identity summaries (destructive actions). Updated §9.4 confirmed-in list to include DR-010-001-06. Confirms that Delete is the only action available for both Draft AND Sent announcements (per §9.4) — administrative cleanup of Sent records is allowed; recipients retain their delivered copy regardless of deletion. Soft-delete semantics (per §5.1) — `deleted` flag + `deleted_at` timestamp, removed from all active views (lists, search, filters, exports, cross-platform consumers). |
| 2026-05-06 | DR-010-001-04 v1.2 — Update Announcement scope correction | Receiver Info is **not** on the Update form. The Figma design shows only the Announcement Details card (Title/Description/Label); audience editing for an existing Draft is exclusively on the Send Announcement screen (DR-010-001-05). Updated §4.5 announcement edit form specific rules — the previous bullet stating Update preserves "Title, Description, Label, Receiver Info" was incorrect and is now narrowed to the three content fields, with an explicit note that the Update payload never mutates the persisted recipient configuration. This split makes the announcement lifecycle two-screen: Update for content, Send for audience finalization. |
