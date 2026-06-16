---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-10
detail_name: "Contracts Management"
parent_requirement: FR-US-005-10
status: draft
version: "1.2"
created_date: 2026-06-12
last_updated: 2026-06-12
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-001-005-01-user-list.md"
    relationship: sibling
  - path: "./DR-001-005-02-create-user.md"
    relationship: sibling
  - path: "./DR-001-005-03-user-details.md"
    relationship: sibling
  - path: "./DR-001-005-04-update-user-information.md"
    relationship: sibling
  - path: "./DR-001-005-09-delete-user.md"
    relationship: sibling
input_sources:
  - type: text
    description: "Feature brief — Contracts submenu inside User Details page, full LIST + CRUD, Labour Contract only, endless contracts disable expiry date"
    extraction_date: "2026-06-12"
  - type: text
    description: "User feedback round 1 — (1) contract status Active/Expired, (2) list filters by type / signed date from-to / expiry date from-to, (3) create/edit forms move to a separate page following the Add New User pattern"
    extraction_date: "2026-06-12"
  - type: figma
    description: "Contracts list screen (User Details sub-page)"
    file_id: "YEHeFgVZau7wmo9BZBVuZC"
    node_id: "3513:5418"
    extraction_date: "2026-06-12"
  - type: figma
    description: "Create Contract screen (separate full page)"
    file_id: "YEHeFgVZau7wmo9BZBVuZC"
    node_id: "3518:4594"
    extraction_date: "2026-06-12"
---

# Detail Requirement: Contracts Management

**Detail ID:** DR-001-005-10
**Parent Requirement:** FR-US-005-10
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.2

---

## 1. Use Case Description

As a **user with contracts management permission**, I want to **record, view, filter, update, and remove the employment contracts attached to a user's profile**, so that **the organization maintains an accurate, centralized employment-contract history per employee and can verify each contract's validity period at a glance**.

**Purpose:** Today there is no place in the HRM platform to track which employment contract an employee is on, when it was signed, or when it expires. HR must keep contract records in spreadsheets or physical files, making it hard to anticipate renewals or verify employment terms. This feature adds a **Contracts submenu inside the User Details page** — a per-user contract register with a filterable, paginated list and full create/read/update/delete capability. Each contract record captures the contract type, the signed date, the expiry date (with support for **endless/indefinite contracts** that have no expiry date), and an optional scan of the signed document. Create and edit happen on **dedicated full pages** (following the same pattern as Add New User), keeping the in-panel list focused on browsing and filtering.

**Target Users:**
- Any role with **contracts view permission** (`user.contracts.view`, configured via US-004) — can see the Contracts button on User Details, view the contract list, and use the filters
- Any role with **contracts management permission** (`user.contracts.manage`, configured via US-004) — can additionally add, edit, and delete contracts

**Key Functionality:**
- New **"Contracts" button** in the User Details left action panel (position 3, directly after Update Information) — opens the Contracts list in the right panel
- **Contract list** per user with columns: Contract type, Signed date, Expiry date (or "Endless"), Attachment link, and gear-icon row actions; standard pagination below the table
- **Filter bar** above the list: Signed date from–to range, Expiry date from–to range, with a Reset control
- **Create contract** on a **separate full page** titled "Create Contract" — contract type (Labour Contract only in this release), signed date, expiry date with endless toggle, optional attachment; header Cancel + Save buttons (Add New User pattern)
- **Edit contract** on a **separate full page** — same fields pre-filled; attachment keep/replace/remove supported
- **Delete contract** — confirmation dialog with contract summary, triggered from the list; soft delete
- **Endless contract rule:** when the "Endless contract" toggle is ON, the Expiry date field is disabled and cleared — no expiry date is stored

---

## 2. User Workflow

**Entry Point:** User Details page (DR-001-005-03) → left action panel → click "Contracts" button (position 3 of 8)

**Preconditions:**
- User is signed in (US-001)
- User has contracts view permission (US-004); management actions additionally require contracts management permission
- The target user exists and the User Details page is loaded

**Main Flow (List & Filter):**
1. Viewer is on the User Details page for a user
2. Viewer clicks "Contracts" in the left action panel — the button shows the active/highlighted state
3. Right panel loads the Contracts list view: an action bar (filters left, "+ Add New" right) and a table of the user's contracts with pagination below
4. Contracts are listed newest first (by Signed date descending), each row showing Contract type, Signed date, Expiry date (or "Endless"), Attachment link, and a gear icon in the Action column
5. Viewer optionally narrows the list using the filter bar: sets a Signed date from–to range and/or an Expiry date from–to range — the table updates to show only matching contracts and pagination resets to page 1
6. A "Reset" control (visible only while any filter is active) clears all filters and restores the full list
7. Viewer reviews the contract history; validity is read directly from the Signed date and Expiry date columns

**Main Flow (Create — separate page):**
1. Manager clicks "+ Add New" (top-right of the Contracts action bar — visible only with management permission)
2. System navigates to the **Create Contract page** — a dedicated full page (same pattern as Add New User) with a breadcrumb identifying the user, page title "Create Contract", and **Cancel** (secondary) + **Save** (primary) buttons in the top-right header area
3. The form shows a single 600px "Contract Information" card with fields in this order: Contract type, Endless contract toggle, Signed date + Expiry date (side by side), Attachment
4. Manager selects the Contract type ("Labour Contract" — the only option in this release; the dropdown opens with the placeholder "Select contract type")
5. Manager selects the Signed date
6. Manager either selects an Expiry date, or turns ON the "Endless contract" toggle — turning it ON clears and disables the Expiry date field (shown at 50% opacity)
7. Manager optionally uploads the signed contract file (Attachment)
8. Manager clicks Save in the header
9. System validates inputs; on success the contract is created, toast "Contract has been created" is shown, and the system redirects back to the User Details page with the Contracts list open — the new record is visible and any previously active filters are preserved

**Main Flow (Edit — separate page):**
1. Manager clicks the gear icon on a contract row → selects "Edit"
2. System navigates to the **Update Contract page** — same full-page layout as Create Contract, with all fields pre-filled (Endless toggle ON and Expiry date disabled if the contract is endless; existing attachment filename displayed). *(No dedicated edit frame exists in Figma yet — the Create Contract frame is the layout reference.)*
3. Manager modifies values and clicks Save
4. System validates; on success toast "Contract has been updated" is shown and the system redirects back to the Contracts list with refreshed data

**Main Flow (Delete — from the list):**
1. Manager clicks the gear icon on a contract row → selects "Delete"
2. Confirmation dialog opens showing a contract summary (contract type and validity period); Cancel button has default keyboard focus; Delete button uses danger/red style
3. Manager clicks Delete → system soft-deletes the contract
4. Toast "Contract has been deleted" is shown; dialog closes; list refreshes without the deleted record (active filters preserved)

**Alternative Flows:**
- **Alt 1 — Endless toggle turned OFF:** On create/edit, turning the toggle OFF re-enables the Expiry date field and makes it mandatory again
- **Alt 2 — Validation failure:** Missing mandatory field or Expiry date not after Signed date → inline errors below the affected fields; form not submitted
- **Alt 3 — Cancel a dirty form:** Manager clicks Cancel (or navigates away) on the Create/Update Contract page with unsaved changes → "Discard unsaved changes?" confirmation dialog; confirming returns to the Contracts list; a clean form returns immediately without confirmation
- **Alt 4 — Delete dialog cancelled:** Manager clicks Cancel or presses Escape → dialog closes, no action taken
- **Alt 5 — Contract already deleted:** Another manager deleted the contract between page load and action → "Contract not found" error toast; manager is returned to the Contracts list (refreshed); on the list itself, the list simply refreshes
- **Alt 6 — API error:** Create/update/delete fails server-side → error toast "Failed to save contract. Please try again." (or delete equivalent); user stays on the form/dialog with input intact
- **Alt 7 — View-only viewer:** Viewer with view permission only sees the list, filters, and pagination; "+ Add New" and gear icons are hidden
- **Alt 8 — Filters match nothing:** Active filters return zero contracts → "No matching results" message with an option to clear filters (distinct from the no-data empty state)
- **Alt 9 — Invalid date range:** In a from–to filter, the "To" picker does not allow choosing a date earlier than the selected "From" date — invalid ranges cannot be entered

**Exit Points:**
- **Success:** Contract created/updated/deleted → toast → return to the Contracts list (filters preserved)
- **Cancel:** Dirty-form discard confirmed, or delete dialog dismissed → no changes persisted
- **Navigate away:** From the list — click another left-panel button or the back arrow; from a form page — Cancel or browser back (discard check applies if the form is dirty)
- **Error:** Validation or server error → user stays in place with error feedback

---

## 3. Field Definitions

### Input Fields (Create / Update Contract page — "Contract Information" card, fields in design order)

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Contract type | Dropdown (576px, non-searchable) | Must be one of the available types | Yes (*) | Empty — placeholder "Select contract type" | Contract category. **This release offers only "Labour Contract"** — the dropdown is retained so additional types can be added later without redesign. Not pre-selected (per Figma) |
| Endless contract | Toggle switch (below Contract type, above the dates) | — | No | OFF | When ON: contract has no expiry — Expiry date field is cleared, disabled, and shown at 50% opacity. When turned OFF: Expiry date re-enables and becomes mandatory |
| Signed date | Date picker (278px, left half) | Valid date; required | Yes (*) | Empty — placeholder "Select signed date" | Date the contract was signed; serves as the contract's effective/reference date for sorting and filtering |
| Expiry date | Date picker (278px, right half) | Required when "Endless contract" is OFF; must be **after** Signed date; disabled and cleared when "Endless contract" is ON | Conditional (*) | Empty — placeholder "Select expiry date" | Date the contract ends. Not applicable for endless contracts |
| Attachment | File input (576px, "Choose File / No file chosen") | PDF, PNG, JPG, DOCX; max 5MB; validated client-side before upload | No | Empty (edit: existing filename shown) | Scan of the signed contract document |

> **Removed in v1.2 (not present in the Figma design):** the optional **Contract Number** field and the optional **Note** field from v1.1 — the form holds exactly the five elements above.

### Filter Controls (Contracts list action bar, above the table)

| Filter | Type | Behavior | Default |
|--------|------|----------|---------|
| Signed date | Filter button (calendar icon) opening a from–to date range | Matches contracts whose Signed date falls within the range, inclusive of both bounds; either bound may be left empty (open-ended range); the "To" picker cannot select a date before "From" | Empty (no restriction) |
| Expiry date | Filter button (calendar icon) opening a from–to date range | Matches contracts whose Expiry date falls within the range, inclusive of both bounds; either bound may be left empty; **endless contracts have no expiry date and never match an active Expiry date filter**; the "To" picker cannot select a date before "From" | Empty (no restriction) |
| Reset | Button (refresh icon) | Clears all active filters and restores the full list; visible only while at least one filter is active (the static frame renders it for illustration) | Hidden |

- The two filters combine with **AND** (Signed date range AND Expiry date range) — consistent with the platform filter pattern
- Active filter buttons display the selected range (e.g., "Signed date: 01/01/26 – 31/12/26")
- **Removed in v1.2:** the Contract Type multi-select filter from v1.1 — not present in the Figma design (moot while only one contract type exists; see Open Questions)
- Applying or clearing a filter resets pagination to page 1

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Contracts | Button | User Details left action panel, position 3 of 8 (directly after Update Information) | Visible with `user.contracts.view`; active/highlighted (#f5f5f5) while the Contracts submenu is open | Loads Contracts list in the right panel | Confirmed by Figma node 3513:5418 |
| + Add New | Button (primary, dark) | Top-right of the Contracts action bar | Visible only with `user.contracts.manage` | Navigates to the Create Contract page | Hidden for view-only viewers |
| Signed date / Expiry date | Filter buttons (calendar icon) | Action bar, left side | Visible to all viewers with view permission | Open a from–to date range picker | Active state shows the selected range |
| Reset | Button (refresh icon) | Action bar, right of the filter buttons | Visible only while a filter is active | Clears all filters | — |
| Gear icon | Dropdown trigger | "Action" column (last) of each contract row | Visible only with `user.contracts.manage` | Opens row actions: Edit, Delete | Hidden for view-only viewers |
| Edit | Gear menu item | Gear dropdown | Always available | Navigates to the Update Contract page pre-filled | No status-based lock — any contract can be corrected |
| Delete | Gear menu item (danger) | Gear dropdown | Always available | Opens delete confirmation dialog | Red/danger text style |
| Attachment link | Clickable filename | Attachment column / form | Present when a file is attached | Downloads or opens the file in a new tab | "—" shown when absent |
| Pagination | Rows-per-page selector + page controls | Below the table | Hidden when total results ≤ current rows per page | Changes page / page size | Default 10; options 10/25/50 |
| Cancel (form page) | Button (secondary, #f5f5f5, square-x icon) | Top-right header of Create/Update Contract page, left of Save | Always visible | Dirty form → "Discard unsaved changes?" dialog; clean form → immediate return to the Contracts list | Add New User pattern |
| Save (form page) | Button (primary, #171717, save icon) | Top-right header of Create/Update Contract page | Disabled + spinner while submitting | Validates → saves → toast → redirects back to the Contracts list | Add New User pattern |
| Delete (dialog) | Button (danger) | Confirmation dialog | Cancel has default keyboard focus; focus trapped in dialog; Escape closes | Soft-deletes the contract | Danger/red style |

### Validation Error Messages

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Contract type empty | "Contract type is required" | Inline, below field |
| Signed date empty | "Signed date is required" | Inline, below field |
| Expiry date empty (toggle OFF) | "Expiry date is required" | Inline, below field |
| Expiry date not after Signed date | "Expiry date must be after signed date" | Inline, below field |
| Attachment wrong format | "File must be PDF, PNG, JPG, or DOCX" | Inline, below field |
| Attachment too large | "File must not exceed 5MB" | Inline, below field |
| Contract already deleted | "Contract not found" | Error toast; return to/refresh the list |
| API failure (save) | "Failed to save contract. Please try again." | Error toast; form stays with input intact |
| API failure (delete) | "Failed to delete contract. Please try again." | Error toast; dialog stays open |

---

## 4. Data Display

### Contracts List — Information Shown to User (columns per Figma)

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Contract type | Text | — (always populated) | "Labour Contract" | Category of the contract |
| Signed date | Date | — (always populated) | DD/MM/YYYY | When the contract was signed / takes effect |
| Expiry date | Date / literal | — | DD/MM/YYYY, or literal **"Endless"** for endless contracts | When the contract ends; "Endless" signals an indefinite contract |
| Attachment | Clickable filename | "—" | Filename link → download / new tab | Scan of the signed contract |
| Action | Gear icon menu | Hidden for view-only viewers | Last column | Edit / Delete row actions |

> **Removed in v1.2:** the derived **Status badge column** (Active/Expired) and the **Contract Number column** from v1.1 — neither appears in the Figma design. The list communicates validity through the Signed date and Expiry date columns directly. Whether an Active/Expired badge should be added back to the design is flagged as an open question (it was requested in user feedback round 1 but is absent from the design).

**Sort order:** Signed date descending (newest contract first).

**Pagination (added in v1.2 per Figma):** standard platform pagination below the table — "Rows per page" selector (default 10; options 10, 25, 50) with numbered page controls. Pagination resets to page 1 when filters are applied or cleared, and is hidden when total results ≤ the current rows per page. The list has **no free-text search** (consistent with the design — only the two date-range filters).

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Contracts list fetching | Skeleton rows in the table (prevents layout shift) |
| Empty (no data) | User has no contracts | "No contracts" message centered in the panel, with "+ Add New" CTA (CTA visible only with management permission); filter bar hidden or inert |
| Empty (no filter results) | Active filters match zero contracts | "No matching results" message with an option to clear filters |
| Populated | 1+ contracts (after filters) | Action bar + table (newest first) + pagination when more rows than the page size |
| Filters active | Any filter applied | Filter buttons show active state (selected range); Reset control visible |
| Error | List fails to load | Error message with retry button; left panel remains visible |
| Form — endless ON | Toggle switched ON | Expiry date cleared, disabled, 50% opacity |
| Form — endless OFF | Toggle switched OFF | Expiry date enabled and marked mandatory |
| Form — saving | Save clicked, request in progress | Save button disabled with spinner; Cancel disabled; fields disabled |
| Success (create) | Contract saved | Toast "Contract has been created" → redirect back to the Contracts list |
| Success (update) | Contract saved | Toast "Contract has been updated" → redirect back to the list with refreshed data |
| Success (delete) | Contract removed | Toast "Contract has been deleted" → dialog closes, list refreshes |
| Conflict | Contract deleted by another manager | "Contract not found" error toast → return to/refresh the list |
| View-only mode | Viewer lacks management permission | List, filters, and pagination visible; "+ Add New" and gear icons hidden |

### Page Layout (Design Reference — Figma)

```
Contracts list (right panel of User Details) — node 3513:5418:
┌──────────────────────────────────────────────────────────────────┐
│ Breadcrumb / Breadcrumb / Breadcrumb                  [Top Bar]  │
├──────────┬───────────────────────────────────────────────────────┤
│ [Sidebar]│ ← User Details > Henry Tran                          │
│  200px   │                                                       │
│          │ ┌─────────────┐ [📅Signed date][📅Expiry date][Reset⟳]│
│          │ │  Overview   │                          [+ Add New] │
│          │ │  Update Info│ ┌──────────────────────────────────┐ │
│          │ │ [Contracts] │ │Contract type│Signed date│Expiry  │ │
│          │ │  Change Role│ │             │           │date    │ │
│          │ │  Change Mail│ │─────────────┼───────────┼────────│ │
│          │ │  Reset Pass │ │ Labour      │ 01/01/26  │Endless │ │
│          │ │  Act/Deact  │ │ Contract    │           │ 📎  ⚙ │ │
│          │ │  Delete Usr │ │ Labour      │ 01/01/25  │31/12/25│ │
│          │ │   181px     │ │ Contract    │           │ 📎  ⚙ │ │
│          │ └─────────────┘ └──────────────────────────────────┘ │
│          │            Rows per page [10]   Page 1 of N  ‹ 1…N ›  │
└──────────┴───────────────────────────────────────────────────────┘

Create Contract (separate full page) — node 3518:4594:
┌──────────────────────────────────────────────────────────────────┐
│ Breadcrumb / Breadcrumb / Breadcrumb                  [Top Bar]  │
├──────────┬───────────────────────────────────────────────────────┤
│ [Sidebar]│ Create Contract                    [Cancel]   [Save] │
│          │                                                       │
│          │        ┌──────────────────────────────────┐           │
│          │        │ Contract Information       600px │           │
│          │        │ * Contract type                  │           │
│          │        │ [Select contract type        ▾]  │           │
│          │        │ (toggle) Endless contract        │           │
│          │        │ * Signed date     * Expiry date  │           │
│          │        │ [Select date 📅] [Select date 📅]│           │
│          │        │ Attachment                       │           │
│          │        │ [Choose File  No file chosen]    │           │
│          │        └──────────────────────────────────┘           │
└──────────┴───────────────────────────────────────────────────────┘
```

> **Note:** Figma frames now exist for the list and create screens (section "Contracts", node 3518:4831). See ANALYSIS.md Design Context [ADD-ON] for the full component inventory. No frames yet for the Update Contract page and the delete confirmation dialog — the Create Contract frame and the platform delete-dialog pattern are the references in the meantime.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Entry & List Display:**
- **AC-01:** A "Contracts" button appears in the User Details left action panel at position 3 (directly after Update Information) for viewers with contracts view permission, and shows the active/highlighted state while the Contracts submenu is open
- **AC-02:** Clicking "Contracts" loads the Contracts list in the right panel without leaving the User Details page
- **AC-03:** The list shows exactly the columns: Contract type, Signed date, Expiry date, Attachment, Action (gear icon)
- **AC-04:** Contracts are ordered by Signed date descending (newest first)
- **AC-05:** Endless contracts display the literal "Endless" in the Expiry date column instead of a date
- **AC-06:** Empty attachment displays "—"
- **AC-07:** When the user has zero contracts, the panel shows a "No contracts" empty state with an "+ Add New" CTA (CTA only with management permission)
- **AC-08:** Skeleton rows are shown while the list loads

**Pagination:**
- **AC-09:** Standard pagination renders below the table: rows-per-page selector (default 10; options 10, 25, 50) and numbered page controls
- **AC-10:** Pagination resets to page 1 when a filter is applied or cleared
- **AC-11:** Pagination controls are hidden when total results ≤ the current rows per page

**Filters:**
- **AC-12:** The action bar above the table contains two filter buttons — Signed date and Expiry date — each opening a from–to date range picker
- **AC-13:** The Signed date filter shows only contracts whose Signed date falls within the selected from–to range, inclusive of both bounds; either bound may be left empty for an open-ended range
- **AC-14:** The Expiry date filter shows only contracts whose Expiry date falls within the selected from–to range, inclusive of both bounds; endless contracts (no expiry date) never match an active Expiry date filter
- **AC-15:** In each date-range filter, the "To" picker cannot select a date earlier than the chosen "From" date
- **AC-16:** The two filters combine with AND logic when both are active
- **AC-17:** A Reset control is visible only while at least one filter is active; activating it clears all filters and restores the full list
- **AC-18:** When active filters match zero contracts, a "No matching results" message is shown with an option to clear filters (distinct from the no-data empty state)
- **AC-19:** Filters are available to all viewers with contracts view permission (including view-only viewers)

**Create Contract (separate page):**
- **AC-20:** "+ Add New" navigates to the dedicated Create Contract page; the button is visible only to viewers with contracts management permission
- **AC-21:** The Create Contract page shows a breadcrumb identifying the target user, the page title "Create Contract", and Cancel (secondary) + Save (primary) buttons in the top-right header area — consistent with the Add New User pattern
- **AC-22:** The form is a single 600px "Contract Information" card with fields in this order: Contract type, Endless contract toggle, Signed date + Expiry date side by side, Attachment
- **AC-23:** The contract being created is always attached to the user whose User Details page launched the flow — there is no employee selector on the form
- **AC-24:** Contract type is mandatory, opens with the placeholder "Select contract type" (not pre-selected), and offers "Labour Contract" as the only option in this release
- **AC-25:** Signed date is mandatory; submitting without it shows inline error "Signed date is required"
- **AC-26:** With the Endless toggle OFF, Expiry date is mandatory and must be strictly after Signed date; violations show the corresponding inline error
- **AC-27:** Turning the Endless toggle ON clears the Expiry date value and disables the field (50% opacity); no expiry date is stored for the contract
- **AC-28:** Turning the Endless toggle OFF re-enables Expiry date and makes it mandatory again
- **AC-29:** Attachment accepts PDF, PNG, JPG, DOCX up to 5MB, validated client-side before upload
- **AC-30:** On successful save, toast "Contract has been created" is shown and the system redirects back to the User Details page with the Contracts list open and the new record visible

**Edit Contract (separate page):**
- **AC-31:** Gear icon → "Edit" navigates to a dedicated Update Contract page (same layout as Create Contract) with all fields pre-filled with current values
- **AC-32:** For an endless contract, the form opens with the Endless toggle ON and Expiry date disabled
- **AC-33:** The existing attachment filename is displayed on load; the manager can keep (no change), replace (select new file), or remove (clear input) the attachment
- **AC-34:** Edit applies the same validation rules as create
- **AC-35:** On successful save, toast "Contract has been updated" is shown and the system redirects back to the Contracts list with refreshed data
- **AC-36:** Any contract can be edited — there is no status-based edit lock

**Delete Contract:**
- **AC-37:** Gear icon → "Delete" opens a confirmation dialog summarizing the contract (contract type and validity period)
- **AC-38:** The dialog's Cancel button has default keyboard focus; the Delete button uses danger/red style; focus is trapped inside the dialog; Escape closes it
- **AC-39:** Confirming deletes the contract (soft delete), shows toast "Contract has been deleted", closes the dialog, and refreshes the list
- **AC-40:** Deletion is a soft delete — the record is flagged as deleted and removed from all active views, but retained in the database

**Cancel / Navigation:**
- **AC-41:** Cancel on a modified Create/Update Contract form shows the "Discard unsaved changes?" confirmation; confirming returns to the Contracts list; Cancel on an untouched form returns immediately without confirmation
- **AC-42:** After returning from the Create/Update Contract page (save or cancel), the Contracts list restores its previous state, including any active filters
- **AC-43:** Leaving the Contracts submenu via another left-panel button or the back arrow is immediate (the list itself holds no unsaved form state)

**Error Handling & Concurrency:**
- **AC-44:** If a save fails server-side, error toast "Failed to save contract. Please try again." is shown and the form retains the entered values
- **AC-45:** If a delete fails server-side, the error toast is shown and the dialog stays open
- **AC-46:** Saving an edit for a contract that was deleted by another manager shows "Contract not found" and returns the manager to the refreshed Contracts list

**Access Control:**
- **AC-47:** The Contracts button is visible only to viewers with contracts view permission; without it, the button does not render and direct access is rejected
- **AC-48:** "+ Add New", the gear icons, and direct URL access to the Create/Update Contract pages are available only with contracts management permission; management actions are also rejected at the API level for unauthorized users
- **AC-49:** Permission checks are enforced server-side — hiding controls in the UI alone is not sufficient

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — fixed-term create | Type=Labour Contract, Signed=01/07/2026, Expiry=30/06/2027, Save | Contract created, toast, redirect to list, new row at top | High |
| Happy path — endless create | Signed=01/01/2026, Endless ON, Save | Contract created with no expiry; list shows "Endless" | High |
| Endless toggle ON | Expiry filled, then toggle ON | Expiry cleared + disabled at 50% opacity; form saves without expiry | High |
| Endless toggle OFF again | Toggle ON → OFF | Expiry re-enabled and mandatory; save blocked until provided | High |
| Type not pre-selected | Open Create Contract page | Contract type shows "Select contract type" placeholder; save blocked until selected | Medium |
| Expiry before signed | Signed=01/06/2026, Expiry=01/05/2026 | Inline error "Expiry date must be after signed date" | High |
| Expiry equals signed | Signed=01/06/2026, Expiry=01/06/2026 | Inline error — expiry must be strictly after signed date | Medium |
| Missing mandatory fields | Save with empty Signed date | Inline error; form not submitted | High |
| Filter by signed date range | Signed date from 01/01/2026 to 31/12/2026 | Only contracts signed within 2026 shown, bounds inclusive; pagination resets to page 1 | High |
| Filter by expiry date range | Expiry date from 01/01/2027 to 31/12/2027 | Only contracts expiring within 2027 shown; endless contracts excluded | High |
| Open-ended date filter | Signed date from 01/01/2026, To empty | All contracts signed on/after 01/01/2026 shown | Medium |
| Combined filters | Signed range + Expiry range both active | Rows matching BOTH conditions shown (AND) | High |
| Invalid range prevented | From=01/06/2026, open To picker | Dates before 01/06/2026 not selectable | Medium |
| Filter no results | Range matching no contracts | "No matching results" + clear-filters option | Medium |
| Reset filters | Filters active, click Reset | All filters cleared; full list restored; Reset hidden | Medium |
| Filter state preserved | Apply filters, open Update page, save | Return to list with same filters still active | High |
| Pagination default | User with 15 contracts | 10 rows on page 1; page controls shown; "Rows per page 10" | Medium |
| Pagination hidden | User with 4 contracts, page size 10 | No pagination controls rendered | Medium |
| Rows per page change | Select 25 rows per page | Up to 25 rows shown; resets to page 1 | Low |
| Attachment replace | Edit, choose new file, Save | New file replaces old; new filename in list | Medium |
| Attachment remove | Edit, clear file input, Save | Attachment removed; list shows "—" | Medium |
| Oversized attachment | Upload 6MB file | Inline error "File must not exceed 5MB"; no upload | Medium |
| Delete confirmed | Gear → Delete → confirm | Soft-deleted; toast; row removed from list | High |
| Delete cancelled | Gear → Delete → Escape | Dialog closes; contract unchanged | Medium |
| Concurrent delete | Edit a contract deleted by another manager, Save | "Contract not found" toast; return to refreshed list | Medium |
| Dirty form discard | Enter data on Create Contract page, click Cancel | "Discard unsaved changes?" dialog | High |
| Clean form cancel | Open Create Contract page, no changes, click Cancel | Immediate return to Contracts list, no dialog | Medium |
| View-only viewer | Viewer with view permission only | List + filters + pagination visible; no "+ Add New", no gear icons | High |
| No view permission | Viewer without contracts view permission | Contracts button not rendered; direct access rejected | High |
| Unauthorized form URL | View-only viewer opens Create Contract page URL directly | Access rejected / fallback page | High |
| Empty state | User with zero contracts | "No contracts" message + CTA (management only) | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Contracts are a per-user sub-entity — every contract belongs to exactly one user. The contract **list** lives inside that user's User Details → Contracts submenu; the **Create/Update forms** are dedicated full pages reached from that list and always return to it. There is no global cross-user contract list in this release.
- **SR-02:** Access is controlled by two fine-grained permissions configured via US-004: **`user.contracts.view`** (see the Contracts button, list, filters, and pagination) and **`user.contracts.manage`** (add, edit, delete — including direct URL access to the Create/Update pages). This mirrors the Salary/Banking fine-grained permission pattern on User Details. No role names are hardcoded.
- **SR-03:** Permission gating is enforced at both the UI level (button/controls/pages not rendered) and the API level (requests rejected) — UI hiding alone is not security.
- **SR-04:** **Contract type** is restricted to **"Labour Contract"** in this release. The field is stored as a typed value (not free text) so additional types (e.g., Probation, Internship, Service) can be introduced later without data migration of existing records. The dropdown opens with a placeholder ("Select contract type") and is not pre-selected, per the Figma design.
- **SR-05:** **The contract record stores exactly: contract type, signed date, expiry date (nullable), endless flag, and optional attachment.** The Contract Number and Note fields from v1.1 were removed — they do not exist in the design and are not stored. *(Changed in v1.2 per Figma.)*
- **SR-06:** **Endless contracts store no expiry date** — the expiry value is empty/null, not a sentinel far-future date. The "Endless" literal in the list, the disabled Expiry date field on forms, and the exclusion from the Expiry date filter are all driven by this flag.
- **SR-07:** **No status field exists — stored or derived — in this release.** The list has no Status column (per Figma); contract validity is communicated through the Signed date and Expiry date columns. The v1.1 derived Active/Expired badge is removed from scope pending the PO/Design decision recorded in Open Questions. If reinstated, it returns as a pure date-derived display value (never stored, never manually set). *(Changed in v1.2 per Figma.)*
- **SR-08:** Expiry date must be **strictly after** Signed date (a contract signed and expiring on the same day is not a valid record in this model).
- **SR-09:** Multiple contracts per user are allowed, including records with overlapping date ranges — the system does not block overlaps in this release (a renewal is often signed before the prior contract ends). Overlap warning/validation is flagged as an open question for the PO.
- **SR-10:** **Filters operate within the single user's contract list only.** Two date-range filters exist: Signed date and Expiry date. AND across the two filters; bounds inclusive; either bound of a range may be empty (open-ended). Endless contracts never match an active Expiry date filter because they hold no expiry date. Clearing/resetting filters restores the full list. Applying or clearing a filter resets pagination to page 1. *(The v1.1 Type filter was removed — not in the design and moot with a single contract type.)*
- **SR-11:** **Pagination follows the standard platform pattern** (per Figma): default 10 rows per page, options 10/25/50, reset to page 1 on filter changes, hidden when total results ≤ the current rows per page. The list has no free-text search. *(Changed in v1.2 — v1.1 omitted pagination.)*
- **SR-12:** The date previously modeled as "Start Date" is the **Signed date** — a single stored date that serves as the contract's signing/effective reference for display, sorting, and filtering. No separate start-date or signed-date pair exists. *(Resolves the v1.1 open question — confirmed by the Figma field naming.)*
- **SR-13:** Editing a contract does not pass through any approval workflow — changes apply immediately on save. Any contract can be edited or deleted at any time (no status-based locks).
- **SR-14:** Attachment handling on edit follows the established keep/replace/remove rule: no change preserves the existing file, selecting a new file replaces it, clearing the input removes it.
- **SR-15:** Deletion is a **soft delete** — the record is flagged as deleted with a deletion timestamp and removed from all active views (including filtered views), but retained in the database. Contracts are legal employment documents; retention supports audit and future restoration tooling.
- **SR-16:** Server re-verifies the record exists before applying an update or delete; if it was already deleted by another manager, the server returns "not found", the client shows the "Contract not found" toast, and the manager is returned to the refreshed Contracts list.
- **SR-17:** Deleting or deactivating the **user** does not require contracts to be removed first — contracts follow the user record (a soft-deleted user's contracts are retained with the user data, hidden from active views along with the profile).
- **SR-18:** The Create/Update Contract pages follow the established full-page create-form pattern (Add New User): header Cancel + Save buttons, dirty-form discard confirmation, clean-form immediate cancel, success toast + redirect back to the originating list. The target user is fixed by the navigation context — the form never offers an employee selector.
- **SR-19:** Returning to the Contracts list from a Create/Update page (after save or cancel) restores the list's previous state, including any active filters — consistent with the platform's list-state preservation rule.

**State Transitions:**
```
Contracts have no managed or displayed status lifecycle in this release.
[Record exists] → [Edit]   → record updated in place (immediate, no approval)
[Record exists] → [Delete] → soft-deleted (removed from active views, retained in DB)
```

**Dependencies:**
- **US-001 (Authentication):** Viewer must be signed in
- **US-004 (Role & Permission Management):** Defines the new `user.contracts.view` and `user.contracts.manage` permissions — permission catalog (docs/FE-PERMISSION-MATRIX.md) must be extended
- **DR-001-005-03 (User Details):** Host page — the Contracts button sits at position 3 of the left action panel (8 buttons total) and the list renders in the right panel
- **DR-001-005-02 (Create User):** Pattern reference — the Create/Update Contract pages follow the same full-page form pattern (header Cancel + Save, dirty/clean cancel, toast + redirect)
- **DR-001-005-01 (User List):** Upstream entry point to User Details
- **DR-001-005-09 (Delete User):** A soft-deleted user's contracts are hidden along with the profile

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** The Contracts list reuses the familiar User Details two-panel pattern, while the Create/Update forms reuse the familiar full-page create-form pattern (Add New User) — both navigation models are already known to managers, so nothing new to learn
- **UX-02:** The two date-range filters (Signed date, Expiry date) match how HR actually thinks about contracts — "what was signed this year?", "what expires next quarter?" — and the "To" picker blocking dates before "From" makes invalid ranges impossible rather than producing confusing empty results
- **UX-03:** The "Endless" literal in the Expiry date column is more meaningful than a blank or "—": it states positively that the contract has no end, rather than implying missing data
- **UX-04:** Disabling (not hiding) the Expiry date field at 50% opacity when Endless is ON keeps the form layout stable and shows the manager exactly what the toggle controls — consistent with the established conditional-field pattern
- **UX-05:** Turning Endless ON clears any entered expiry value immediately — prevents a stale hidden date from being saved accidentally
- **UX-06:** Placing the Endless toggle directly between Contract type and the date pair (per the design) means the manager decides "fixed-term or endless?" before reaching the date fields, so the disabled Expiry date never comes as a surprise
- **UX-07:** Newest-first ordering (Signed date descending) puts the current/most recent contract at the top, where managers look first
- **UX-08:** The Reset control appears only when filters are active — no dead control cluttering the default view; active filter buttons display their selected range so the manager always sees why rows are hidden
- **UX-09:** Moving create/edit to a dedicated page gives the form room to breathe, a stable header with Cancel/Save always visible, and a shareable/bookmarkable location — while the list keeps the quick in-panel browsing experience
- **UX-10:** Standard pagination keeps the in-panel list compact for long contract histories while staying consistent with every other table on the platform
- **UX-11:** Delete confirmation summarizes the specific contract (type and validity period) so the manager verifies they are removing the right record; Cancel holds default focus to prevent accidental Enter-key deletion
- **UX-12:** Save button shows a loading spinner and disables during submission — prevents duplicate contract records from double-clicks
- **UX-13:** Dirty-form discard confirmation protects in-progress data entry on the Create/Update pages
- **UX-14:** Skeleton rows during list load prevent layout shift and signal that data is coming
- **UX-15:** Preserving filter state when returning from a Create/Update page keeps the manager's working context intact — no need to re-apply filters after every change

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | List: two-panel layout (left action panel 181px + right content area with action bar, table, pagination). Create/Update: full-page form with single centered card (600px), header Cancel + Save |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through panel buttons, filter controls, table rows, gear menus, pagination, form fields, toggle, header Cancel/Save
- [x] Filter buttons and date-range pickers operable via keyboard; active filter state announced to screen readers
- [x] Endless toggle operable via keyboard (Space/Enter) with its state announced to screen readers
- [x] Disabled Expiry date field announced as disabled/not applicable when Endless is ON
- [x] Delete dialog: focus trap, Escape to close, default focus on Cancel, danger button meets WCAG 2.1 AA contrast
- [x] Attachment link uses the filename as descriptive link text
- [x] Inline validation errors associated with their fields for screen reader announcement
- [x] Focus indicators visible on all interactive elements

**Design References:**
- Figma: [Contracts list](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3513-5418) (node `3513:5418`) and [Create Contract](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3518-4594) (node `3518:4594`), inside section "Contracts" (node `3518:4831`)
- Design tokens: Geist typography, heading 3 (24px Semibold) page title, paragraph small (14/20) labels and cells, `general/primary` #171717 (Save button), `general/secondary` #f5f5f5 (Cancel button, active panel button accent), `general/border` #e5e5e5, `general/muted foreground` #737373, rounded-lg 8px inputs/buttons, form card 600px with 576px inputs
- **Missing frames:** Update Contract page and delete confirmation dialog — Design Team to produce; the Create Contract frame and the platform delete-dialog pattern are the interim references

---

## 8. Additional Information

### Out of Scope
- A global, cross-user Contracts list page (e.g., "all contracts expiring this quarter") — this release is per-user only
- Contract types other than Labour Contract (Probation, Internship, Service, etc.) — dropdown is forward-compatible but offers one option
- A derived Active/Expired status badge on the list — present in v1.1, removed in v1.2 because the design has no Status column (pending PO/Design decision; see Open Questions)
- Contract Number and Note fields — removed in v1.2 (not in the design)
- A Contract Type filter — removed in v1.2 (not in the design; moot with a single type)
- Free-text search on the contract list — the two date-range filters cover the narrowing need
- Contract renewal automation or reminders/notifications before expiry — future enhancement
- Approval workflow for contract changes — edits apply immediately
- Contract document generation/templating — only upload of externally signed documents
- E-signature integration
- Salary/compensation terms inside the contract record — compensation lives on the user profile (Salary card, DR-001-005-03)
- Restoration of soft-deleted contracts — data retained, tooling deferred
- Overlap prevention between contract periods — allowed in v1, pending PO decision
- Bulk operations (multi-delete, import/export of contracts)
- Contract change history / audit log view
- Mobile or tablet layout

### Open Questions
- [ ] **Status badge vs design:** User feedback round 1 requested a derived Active/Expired status, but the Figma list has **no Status column**. v1.2 follows the design (no status display). Confirm: should the Design Team add an Active/Expired badge column, or is validity-via-dates the final intent? — **Owner:** Product Owner + Design Team — **Status:** Pending
- [ ] **Type filter vs design:** The v1.1 Type multi-select filter is absent from the Figma action bar (only Signed date + Expiry date + Reset). Presumed intentional while only "Labour Contract" exists — confirm the Type filter should be (re)introduced when additional contract types arrive. — **Owner:** Product Owner — **Status:** Pending
- [ ] **Contract overlap rule:** Should the system warn or block when a new contract's period overlaps an existing non-deleted contract for the same user? v1 allows overlaps silently. — **Owner:** Product Owner — **Status:** Pending
- [ ] **Permission granularity:** Confirm dedicated `user.contracts.view` / `user.contracts.manage` permissions (mirroring Salary/Banking) vs. folding contracts under general user view/management permission. v1 assumes dedicated permissions because contracts are sensitive employment documents. — **Owner:** Product Owner — **Status:** Pending
- [ ] **Expiry reminder:** Is an "expiring soon" indicator (e.g., within 30 days of expiry) wanted in a future iteration? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Missing Figma frames:** Update Contract page and delete confirmation dialog frames do not exist yet (list + Create Contract frames delivered 2026-06-12). — **Owner:** Design Team — **Status:** Pending

**Resolved in v1.2 (by the Figma design):**
- ~~Signed date vs Start Date~~ — the field **is** "Signed date"; a single stored date serves as the contract's signing/effective reference (no separate start date)
- ~~Contract Number field~~ — not in the design; field removed
- ~~Left panel button position~~ — confirmed at position 3, directly after Update Information (8 buttons total)
- ~~List/create Figma frames~~ — delivered (nodes 3513:5418, 3518:4594)

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-03: User Details | Host page — Contracts button at position 3 of the left action panel; list renders in the right panel |
| DR-001-005-02: Create User | Pattern reference — Create/Update Contract pages follow the same full-page form pattern |
| DR-001-005-01: User List | Upstream entry point to User Details |
| DR-001-005-04: Update User Information | Sibling sub-page on the same action panel (Contracts sits directly after it) |
| DR-001-005-09: Delete User | Soft-deleting a user hides their contracts along with the profile |
| US-004: Role & Permission Management | Defines the new contracts view/manage permissions |
| US-001: Authentication | Viewer must be signed in |

### Notes
- **v1.2 is the design-reconciliation revision** — the Figma frames (delivered 2026-06-12) superseded several v1.1 assumptions: the date field is **Signed date** (not Start Date), the list has **no Status column and no Contract Number column**, the form has **no Contract Number or Note fields**, the filter bar holds **only the two date-range filters**, the list **does paginate**, the add button is **"+ Add New"**, and the create page is titled **"Create Contract"**.
- **The derived-status concept is parked, not rejected:** v1.1's user-confirmed Active/Expired badge does not appear in the design. v1.2 documents what the design shows; the conflict is logged as the top open question for PO + Design. If reinstated, the previous rules apply unchanged (pure function of expiry vs today, inclusive expiry, never stored).
- **Endless contracts** model Vietnamese indefinite-term labour contracts: no expiry date exists, so the field is disabled and stored empty — not faked with a far-future date. Endless contracts display the "Endless" literal and never match the Expiry date filter.
- The **Contract type dropdown with a single option** is deliberate: it sets the data shape for future contract types while keeping this release's scope to Labour Contract only, per the PO brief. Per the design it opens with a placeholder and is not pre-selected.
- **v1.1 hybrid pattern confirmed by design:** the per-user sub-entity keeps its **in-panel list** (quick browsing inside User Details) and **create/edit on dedicated full pages** following the Add New User pattern. The delete confirmation remains a dialog on the list.
- The list **omits free-text search** but **includes standard pagination** (per the design) — v1.1's "no pagination" assumption was overturned by the Figma frame, which renders the standard pagination component.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-06-12 | BA Agent | Initial draft — per-user Contracts submenu on User Details with full LIST + CRUD; Labour Contract only; endless contracts disable/clear expiry date; derived Upcoming/Active/Expired status. No Figma frame available yet. |
| 1.1 | 2026-06-12 | BA Agent | User feedback round 1: (1) status simplified to two derived values — Active/Expired (Upcoming removed; future-dated contracts are Active); (2) added list filter bar — Contract Type multi-select + Start Date from–to + Expiry Date from–to ranges, AND across filters, inclusive bounds, endless contracts excluded from Expiry filter, Reset control, filter state preserved across navigation ("signed date" interpreted as Start Date — open question for PO); (3) create/edit forms moved from the User Details right panel to dedicated full pages following the Add New User pattern (header Cancel + Save, dirty/clean discard, toast + redirect back to the list). |
| 1.2 | 2026-06-12 | BA Agent | **Figma design reconciliation** (nodes 3513:5418 list, 3518:4594 create): (1) "Start Date" renamed to **Signed date** everywhere — resolves the signed-date open question (single stored date, no separate start date); (2) **Status badge column removed** — design has no Status column; derived Active/Expired display parked as open question for PO/Design; (3) **Contract Number and Note fields removed** — not in the design; (4) **Type filter removed** — action bar holds only Signed date + Expiry date range filters + Reset; (5) **standard pagination added** (default 10, options 10/25/50, reset on filter change, hidden when few rows) — overturns v1.1 "no pagination"; (6) add button is **"+ Add New"**; create page titled **"Create Contract"**; (7) Contract type **not pre-selected** (placeholder "Select contract type"); (8) form field order per design: Type → Endless toggle → Signed + Expiry → Attachment; (9) panel position confirmed: Contracts at position 3 of 8, after Update Information; (10) design references and tokens updated; Update Contract page + delete dialog frames still missing. |
