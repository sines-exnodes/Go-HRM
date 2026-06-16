---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-003
story_name: "Holiday Management"
detail_id: DR-009-003-01
detail_name: "Holiday Management"
parent_requirement: FR-US-003-01
status: draft
version: "1.0"
created_date: 2026-06-15
last_updated: 2026-06-15
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
  - path: "../../../EP-002-leave-management/US-001-leave-requests/details/"
    relationship: cross-reference
  - path: "../../../EP-004-attendance-management/US-001-attendance-list/details/"
    relationship: cross-reference
  - path: "../../../EP-001-foundation/US-005-user-management/details/DR-001-005-02-create-user.md"
    relationship: pattern-reference
input_sources:
  - type: text
    description: "Approved design spec — docs/superpowers/specs/2026-06-15-holiday-management-design.md"
    extraction_date: "2026-06-15"
  - type: figma
    description: "Holidays list frame — nodeId 3537:3821"
    extraction_date: "2026-06-15"
---

# Detail Requirement: Holiday Management

**Detail ID:** DR-009-003-01
**Parent Requirement:** FR-US-003-01
**Story:** US-003-holiday-management
**Epic:** EP-009 (Organization Settings)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **HR administrator or authorized manager**, I want to **create and maintain a company holiday calendar per year** so that **leave request day counts automatically exclude public holidays, giving employees accurate leave balance calculations across the entire year**.

**Purpose:** The Exnodes HRM system references company holidays in two places — leave request calculations and the attendance calendar matrix — but currently has no admin interface to define them. Without confirmed holiday records, the system cannot accurately calculate how many leave days an employee consumed (holidays must be excluded from calendar-day counts), and the attendance matrix cannot display confirmed holiday names or exempt employees from late-arrival flags on those dates. This module closes that gap by providing a full CRUD interface for the company holiday calendar, organized by year, with an import shortcut for Vietnamese public holiday presets.

**Target Users:**
- HR administrators with `organization.holidays.manage` permission — primary users who create, edit, delete, and import holidays
- All roles with `organization.holidays.view` permission — read-only access to see the holiday calendar

**Key Functionality:**
- Browse holidays by year with search filtering
- Create a new holiday (date range with auto-calculated total days)
- Edit an existing holiday's name and dates
- Delete a holiday with automatic recalculation of affected approved leave requests
- Import Vietnamese public holiday preset for a selected year via a two-step modal

---

## 2. User Workflow

**Entry Point:** Left sidebar navigation — "Holidays" link under Organization Settings section. Accessible to all users with `organization.holidays.view` or `organization.holidays.manage` permission.

**Preconditions:**
- User is authenticated and has at least `organization.holidays.view` permission
- The system is operating and holiday data is accessible

---

### 2A. Browse Holiday List

**Main Flow:**
1. User clicks "Holidays" in the left sidebar.
2. System navigates to the Holidays list page (breadcrumb: Organization Settings / Holidays).
3. System defaults the year filter to the current calendar year and loads the holiday table.
4. Table displays holidays for the selected year sorted by From Date ascending. Each row shows: Holiday Name, From Date, To Date, Total Days, and Action (gear icon).
5. User may change the year dropdown to view a different year — table reloads, search resets.
6. User may type in the search input to filter holidays by name (debounced ~300ms, case-insensitive, partial match, scoped to selected year).
7. User scrolls or paginates through results (10 rows default; options 10/25/50).

**Alternative Flows:**
- **No holidays for year (no search active):** Page shows empty state "No holidays set for [year]" with Import From Template and Add New call-to-action buttons visible to manage-permission users.
- **Search returns no results:** Table shows "No results found" with a clear search option.
- **View-only user:** Import From Template and Add New buttons are hidden. Gear icon column is hidden. List and search remain accessible.

**Exit Points:**
- User clicks Add New → navigates to Create Holiday page (Flow 2B).
- User clicks Import From Template → opens Import modal (Flow 2D).
- User clicks gear icon on a row → sees Edit and Delete options.

---

### 2B. Create Holiday

**Entry Point:** "+ Add New" button on the holiday list page (visible only to `organization.holidays.manage` users).

**Main Flow:**
1. System navigates to the Create Holiday full-page form (breadcrumb: Organization Settings / Holidays / Create Holiday).
2. The Year field displays the currently filtered year as a read-only label — this locks the holiday to the correct year.
3. HR enters Holiday Name (required, max 100 characters).
4. HR selects From Date using the date picker.
5. HR selects To Date using the date picker (must be on or after From Date). Total Days auto-calculates live: (To Date − From Date) + 1.
6. HR clicks Save. System validates all required fields and business rules.
7. On success: toast notification "Holiday has been created" appears; system redirects to the holiday list with the same year filter active.

**Alternative Flows:**
- **Validation failure (inline):** Error message appears below the failing field. Save button remains enabled for re-attempt.
  - Missing Holiday Name or whitespace-only: "Holiday name is required"
  - To Date before From Date: "To date must be on or after from date"
  - Duplicate name in same year: "A holiday with this name already exists for [year]"
- **Cancel with unsaved changes (dirty form):** Dialog appears — "Discard unsaved changes?" with Cancel and Discard buttons. Confirming Discard navigates back to the list.
- **Cancel with no changes (clean form):** Navigates immediately back to the holiday list, no dialog.

**Recalculation on Create:**
- After a new holiday is saved, the system automatically checks all Approved leave requests whose date range overlaps the new holiday's dates.
- Any overlapping Approved leave request has its leave days recalculated: the newly added holiday days are excluded, reducing the leave days consumed (restoring balance to the employee).
- Recalculation is silent — no additional dialog or step required from HR. Only Approved requests are recalculated; Pending, Rejected, and Cancelled are not affected.

**Exit Points:**
- **Success:** Redirect to holiday list, same year filter active.
- **Cancel:** Return to holiday list (with or without discard dialog depending on form state).

---

### 2C. Edit Holiday

**Entry Point:** Gear icon on a holiday row → Edit option (visible only to `organization.holidays.manage` users).

**Main Flow:**
1. System navigates to the Update Holiday full-page form, pre-filled with existing values (breadcrumb: Organization Settings / Holidays / Update Holiday).
2. Year field is a read-only label — locked, cannot be changed via edit.
3. HR may update Holiday Name, From Date, and/or To Date. Total Days auto-recalculates live.
4. HR clicks Save. System validates as per Create Holiday rules (duplicate name check excludes the current record).
5. On success: toast notification "Holiday has been updated" appears; system redirects to the holiday list with the same year filter active.

**Alternative Flows:**
- Same inline validation flows as Create Holiday apply.
- Cancel dirty / Cancel clean behavior is identical to Create Holiday.

**Recalculation on Edit:**
- After dates are changed and saved, the system automatically re-evaluates all Approved leave requests that overlapped the holiday's old or new date range and adjusts their leave day counts to reflect the updated holiday dates.
- Recalculation scope: Approved requests only.

**Exit Points:**
- **Success:** Redirect to holiday list, same year filter active.
- **Cancel:** Return to holiday list.

---

### 2D. Import From Template

**Entry Point:** "Import From Template" button on the holiday list page (visible only to `organization.holidays.manage` users).

**Main Flow:**
1. Import modal opens: "Import Holidays From Template".
2. **Step 1 — Year Selection:** Year dropdown defaults to the currently filtered year. If the selected year already has existing holidays, the modal shows a warning: "This year already has [N] holidays. Importing will add to the existing list. Duplicates (same name) will be skipped."
3. HR selects or confirms the target year and proceeds to the preview step.
4. **Step 2 — Preview:** System loads the Vietnamese public holiday preset for the selected year. A preview table shows all preset holidays (Holiday Name, From Date, To Date, Total Days) with all rows pre-checked.
5. HR may uncheck individual rows to exclude them, or use "Select all / Deselect all" toggle.
6. Footer shows the live count: "Import [N] Holidays" button updates as rows are checked/unchecked.
7. HR clicks "Import [N] Holidays". System imports all checked holidays.
8. Duplicate detection: any checked holiday whose name already exists in the target year is silently skipped (not imported, no error).
9. Modal closes. List refreshes to show the imported year.
10. Toast notification: "12 holidays imported for 2026" or "10 holidays imported, 2 skipped (already exist)" depending on skip count.

**Alternative Flows:**
- **No preset available for selected year:** Preview step shows "No template available for [year]. Add holidays manually."
- **Cancel at any step:** Modal closes, no changes made.

**Recalculation on Import:**
- Each successfully imported holiday triggers the same recalculation logic as Create Holiday — Approved leave requests overlapping the new holiday date are recalculated automatically.

**Exit Points:**
- **Success:** Modal closes, list refreshes to imported year.
- **Cancel:** Modal closes, no changes.

---

### 2E. Delete Holiday

**Entry Point:** Gear icon on a holiday row → Delete option (visible only to `organization.holidays.manage` users).

**Main Flow:**
1. Confirmation dialog appears: "Delete Holiday" / "Are you sure you want to delete **[Holiday Name]** ([From Date] – [To Date])? This will recalculate any approved leave requests that overlap this period."
2. Dialog buttons: Cancel (default focus, secondary style) and Delete (danger style).
3. Focus is trapped within the dialog. Pressing Escape closes the dialog (equivalent to Cancel).
4. HR clicks Delete.
5. System soft-deletes the holiday record (record is retained in system with a deleted marker, removed from all active views).
6. System recalculates all Approved leave requests that overlapped the deleted holiday's date range — those holiday days are added back as consumed leave days (employee balance decreases accordingly, since the days are no longer excluded from the count).
7. Toast notification appears:
   - If recalculation affected one or more requests: "Holiday deleted. [N] leave request(s) recalculated."
   - If no requests were affected: "Holiday has been deleted."

**Alternative Flows:**
- **Cancel dialog:** No action taken.
- **Escape key:** No action taken; dialog closes.

**Exit Points:**
- **Confirmed delete:** Holiday removed from list; toast shown.
- **Cancelled:** Dialog closes; no change.

---

## 3. Field Definitions

### Input Fields — Create / Edit Holiday Form

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Year | Read-only label | N/A — locked to currently filtered year | Display only | Current year filter value | Shows which year this holiday belongs to; cannot be changed via form |
| Holiday Name | Text input | Required; max 100 characters; whitespace-only treated as empty; must be unique within the same year | Yes | Empty | The official name of the holiday (e.g., "Independence Day") |
| From Date | Date picker | Required; any calendar date | Yes | Empty | The first day of the holiday period |
| To Date | Date picker | Required; must be on or after From Date (same day = single-day holiday) | Yes | Empty | The last day of the holiday period |
| Total Days | Read-only display | N/A — auto-calculated; not user-editable | Display only | Blank until From Date and To Date are both set | Displays (To Date − From Date) + 1; updates live as dates change |

### Input Fields — Import From Template Modal

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Import Year | Dropdown | Required; must select a year | Yes | Currently filtered year on the list page | The year for which the Vietnamese public holiday preset is loaded |
| Row Selection | Checkbox per row | At least one row must be checked to enable Import button | Yes | All rows pre-checked | Allows HR to exclude specific holidays from the import |

### List Screen Filters

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Year Filter | Dropdown | Must select a year | Yes (always has a value) | Current calendar year | Scopes the holiday list to the selected year; dropdown shows all years with at least one holiday plus the current year |
| Search | Text input | Optional; debounced ~300ms | No | Empty | Case-insensitive partial match on holiday name; scoped to the selected year |

### Interaction Elements

| Element Name | Type | State / Condition | Trigger Action | Description |
|--------------|------|-------------------|----------------|-------------|
| + Add New | Primary button | Visible only to `organization.holidays.manage` users | Navigate to Create Holiday page | Opens the full-page Create Holiday form locked to the currently filtered year |
| Import From Template | Secondary button | Visible only to `organization.holidays.manage` users | Open Import From Template modal | Launches the two-step import modal |
| Gear icon | Icon button | Visible only to `organization.holidays.manage` users; one per table row | Reveal Edit and Delete options | Opens inline dropdown with Edit and Delete actions for the row's holiday |
| Save | Primary button (with spinner) | Always visible on form pages; spinner shown during save request | Submit form | Validates and saves the holiday record |
| Cancel | Secondary button | Always visible on form pages | Navigate back | Returns to holiday list; prompts discard dialog if form has unsaved changes |
| Select all / Deselect all | Toggle link | Visible in import modal Step 2 | Toggle all checkboxes | Checks or unchecks all preview rows simultaneously |
| Import [N] Holidays | Primary button | Count updates live; disabled when 0 rows selected | Confirm import | Imports all checked holiday rows |
| Pagination controls | Pagination component | Hidden when total rows ≤ page size | Change page / page size | Standard 10/25/50 page size options; resets to page 1 on year or search change |

---

## 4. Data Display

### Information Shown — Holiday List Table

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Holiday Name | Text | — | Plain text, max 100 chars | The official name of the holiday |
| From Date | Date | — | DD/MM/YYYY | The start date of the holiday period |
| To Date | Date | — | DD/MM/YYYY | The end date of the holiday period |
| Total Days | Integer | — | Numeric (e.g., "3") | The total number of calendar days in the holiday period, inclusive |
| Action | Icon | Hidden (no gear for view-only users) | Gear icon | Reveals Edit and Delete options for manage-permission users |

### Information Shown — Create / Edit Form

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Year | Integer | N/A (always pre-filled) | Plain number (e.g., "2026") | Indicates which year calendar this holiday belongs to |
| Holiday Name | Text | Empty input field | Plain text | The holiday's name as entered by HR |
| From Date | Date | Empty date picker | DD/MM/YYYY | Start of holiday |
| To Date | Date | Empty date picker | DD/MM/YYYY | End of holiday |
| Total Days | Integer | Blank / "—" | Numeric | Live calculation of (To Date − From Date) + 1; updates as dates are selected |

### Information Shown — Import Preview Table

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Row checkbox | Boolean | N/A | Checked by default | Controls whether this preset row is included in the import |
| Holiday Name | Text | — | Plain text | Name of the Vietnamese public holiday |
| From Date | Date | — | DD/MM/YYYY | Start date of the preset holiday |
| To Date | Date | — | DD/MM/YYYY | End date of the preset holiday |
| Total Days | Integer | — | Numeric | Calendar days in this preset holiday |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading (list) | Holidays are being fetched for the selected year/search | Skeleton rows in the table area |
| Empty — no holidays (list) | Selected year has no holidays and no active search | "No holidays set for [year]" message with Import From Template and Add New CTAs (manage users only) |
| Empty — search no results (list) | Search query returns no matches for selected year | "No results found" with a clear search link |
| Error (list) | System cannot retrieve holiday data | Inline error message with a retry option |
| Success — create | Holiday saved successfully | Toast "Holiday has been created"; redirect to list |
| Success — update | Holiday updated successfully | Toast "Holiday has been updated"; redirect to list |
| Success — delete (with recalc) | Holiday deleted, affected leave requests recalculated | Toast "Holiday deleted. [N] leave request(s) recalculated." |
| Success — delete (no recalc) | Holiday deleted, no overlapping approved requests | Toast "Holiday has been deleted." |
| Success — import | Import completed | Toast "X holidays imported for [year]" or "X holidays imported, Y skipped (already exist)"; list refreshes |
| Import — empty preset | Selected year has no Vietnamese public holiday preset in the system | "No template available for [year]. Add holidays manually." within the modal |
| Loading (import preview) | System is fetching the preset for the selected year | Loading indicator within modal |
| Discard dialog | User cancels a dirty form | Dialog: "Discard unsaved changes?" with Cancel and Discard buttons |
| Delete confirmation | User selects Delete from gear menu | Dialog: "Delete Holiday / Are you sure..." with Cancel (default focus) and Delete (danger) buttons |

---

## 5. Acceptance Criteria

**Definition of Done — all criteria must be met:**

- **AC-01:** Holiday list defaults to the current calendar year on page load; the year dropdown includes all years that have at least one holiday plus the current year; changing the year reloads the table and resets the search field.
- **AC-02:** Users with only `organization.holidays.view` permission see the holiday list and can search and paginate, but the Add New button, Import From Template button, and gear icon column are all hidden.
- **AC-03:** Creating a holiday with a name that already exists in the same year displays the error "A holiday with this name already exists for [year]" inline below the Holiday Name field; the form does not submit.
- **AC-04:** Total Days auto-calculates live on the form as (To Date − From Date) + 1 when both dates are selected; selecting a single day (From = To) shows Total Days = 1.
- **AC-05:** Selecting a To Date earlier than From Date displays the error "To date must be on or after from date" inline and prevents form submission.
- **AC-06:** Creating a new holiday that overlaps an existing Approved leave request automatically recalculates the affected request's leave day count by excluding the new holiday days; no manual step is required from HR.
- **AC-07:** Deleting a holiday that overlaps one or more Approved leave requests recalculates those requests (adding back the holiday days as consumed leave days); the post-delete toast shows the count: "Holiday deleted. [N] leave request(s) recalculated." Deleting a holiday with no overlapping Approved requests shows "Holiday has been deleted."
- **AC-08:** Editing a holiday's dates triggers automatic recalculation of all Approved leave requests that overlapped either the old or the new date range; the success toast shows "Holiday has been updated" and the list reloads with the same year filter.
- **AC-09:** The Import From Template modal pre-checks all rows; HR can uncheck individual rows or use Select all / Deselect all; the Import button label updates live with the selected count; holidays with names already existing in the target year are silently skipped on confirm.
- **AC-10:** Cancelling a dirty Create or Edit form (any field modified) shows the "Discard unsaved changes?" dialog; confirming returns to the holiday list. Cancelling a clean form (no changes) navigates back immediately with no dialog.
- **AC-11:** The Delete confirmation dialog traps focus between its Cancel and Delete buttons; pressing Escape closes the dialog without deleting; Delete button uses danger styling; Cancel has default focus.
- **AC-12:** Half-day leave requests that overlap a holiday date have 0.5 days excluded from their leave day count, consistent with the leave day formula.
- **AC-13:** Recalculation triggered by any holiday mutation (create, edit, delete, import) applies only to Approved leave requests; Pending, Rejected, and Cancelled requests are not modified.
- **AC-14:** The holiday list table is sorted by From Date ascending for the selected year; pagination defaults to 10 rows per page with options 10/25/50; pagination controls are hidden when total results fit within one page; changing page size or year or search resets to page 1.

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Create holiday — happy path | Holiday Name "Tet Holiday", From Date 2026-01-27, To Date 2026-01-31 | Holiday saved; Total Days shows 5; toast "Holiday has been created"; list reloads filtered to 2026 | High |
| Create holiday — duplicate name | Holiday Name "Tet Holiday" (already exists in 2026) | Inline error: "A holiday with this name already exists for 2026"; form not submitted | High |
| Create holiday — To Date before From Date | From Date 2026-05-10, To Date 2026-05-09 | Inline error: "To date must be on or after from date"; form not submitted | High |
| Create holiday — whitespace-only name | "   " in Holiday Name field | Inline error: "Holiday name is required"; form not submitted | Medium |
| Create holiday triggers leave recalc | HR creates holiday 2026-04-30; employee has Approved leave request covering 2026-04-29 to 2026-05-02 | Leave request recalculated: 2026-04-30 excluded; leave days consumed reduced by 1 | High |
| Delete holiday with overlapping Approved leave | Delete "Tet Holiday" (2026-01-27 to 2026-01-31); 2 employees have overlapping Approved leave requests | Toast: "Holiday deleted. 2 leave request(s) recalculated."; those 2 leave requests have holiday days added back as consumed | High |
| Delete holiday — no overlapping requests | Delete a holiday with no overlapping Approved leave requests | Toast: "Holiday has been deleted." | Medium |
| Edit holiday dates — recalculation | Change "Tet Holiday" From Date from Jan 27 to Jan 28 | System re-evaluates overlapping Approved requests against new dates; toast "Holiday has been updated" | High |
| Import From Template — full import | Select 2026; preset has 12 holidays; all rows checked | Toast: "12 holidays imported for 2026"; all 12 appear in list | High |
| Import From Template — partial import with skips | Select 2026; 12 preset rows; 2 already exist in system; 10 checked, 2 skipped | Toast: "10 holidays imported, 2 skipped (already exist)" | High |
| Import From Template — uncheck rows | 12 preset rows; HR unchecks 3 | Import button shows "Import 9 Holidays"; only 9 imported | Medium |
| View-only user | User with only `organization.holidays.view` | No Add New, no Import From Template, no gear column visible | High |
| Year filter change | User changes year dropdown from 2026 to 2025 | Table reloads with 2025 holidays; search field cleared; pagination resets to page 1 | Medium |
| Empty year — no active search | Year 2030 selected; no holidays recorded | "No holidays set for 2030" with Import and Add New CTAs | Medium |
| Half-day overlap with holiday | Employee has half-day Approved leave on 2026-01-27 (also a holiday) | Leave day count: 0.5 days excluded for that holiday date | Medium |

---

## 6. System Rules

**Business Logic:**

- **Rule 1 — Holiday ownership:** Each holiday record belongs to a specific year (integer). The year is set at creation from the currently filtered year and cannot be changed via the edit form.
- **Rule 2 — Name uniqueness per year:** Holiday names must be unique within the same year. The same name may exist in different years (e.g., "Tet Holiday" can exist in both 2025 and 2026).
- **Rule 3 — Total days calculation:** Total Days = (To Date − From Date) + 1 using calendar days (includes weekends and other holidays within the range). This is a display value computed at query time; it is not stored as a separate field.
- **Rule 4 — Leave day formula:** `Leave Days Deducted = Calendar Days in request range − Company Holiday Days in request range`. This formula is applied at leave creation time and re-applied during recalculation. Holiday exclusion applies to all leave types: Annual, Sick, Personal, Maternity, and Unpaid.
- **Rule 5 — Half-day proportional exclusion:** If an employee takes a half-day leave on a date that is also a company holiday, 0.5 days are excluded from the leave day count for that date (rather than 1.0 day for a full-day holiday exclusion).
- **Rule 6 — Recalculation scope:** Any mutation to the holiday calendar (create, edit, delete, or import) that affects dates automatically triggers recalculation of all Approved leave requests whose date range overlaps the affected holiday dates. Pending, Rejected, and Cancelled leave requests are never recalculated.
- **Rule 7 — Recalculation direction:**
  - Adding a holiday (create or import): affected Approved leave requests have those holiday days excluded → leave days consumed decreases → balance restored to employee.
  - Removing a holiday (delete): affected Approved leave requests have those holiday days added back as consumed → leave days consumed increases → balance decreases.
  - Editing dates: system re-evaluates against both old and new date ranges and adjusts accordingly.
- **Rule 8 — Soft delete:** Deleted holiday records are retained in the system with a deletion marker and are excluded from all active views. Historical leave calculation data that referenced the deleted holiday may still exist in audit logs.
- **Rule 9 — Import duplicate handling:** During an import, any preset holiday whose name exactly matches an existing holiday in the target year is silently skipped. Skipped holidays are not treated as errors. The success toast reports both the imported count and the skipped count when skips occur.
- **Rule 10 — Year filter dropdown population:** The year dropdown on the list page always includes the current calendar year. It also includes any other year for which at least one holiday record exists. Empty years (except current) are not shown.
- **Rule 11 — Attendance integration (existing behavior, formalized here):** Holiday dates do not break employee attendance streaks. The attendance calendar matrix displays "H" with a tooltip showing the holiday name on holiday cells. The late-arrival threshold configured in Attendance Settings does not apply on confirmed holiday dates.
- **Rule 12 — Permission enforcement:** All create, edit, delete, and import operations require `organization.holidays.manage` permission. Users with only `organization.holidays.view` can browse and search. Any manage action attempted without the permission is rejected.

**Calculations / Formulas:**

- **Total Days (display):** (To Date − From Date) + 1 calendar days (inclusive of weekends)
- **Leave Days Deducted:** Calendar Days in leave range − Company Holiday Days within leave range
- **Half-day on holiday:** 0.5 days excluded (not 1.0)

**Dependencies:**

- **EP-001 US-004 (Role & Permissions):** `organization.holidays.view` and `organization.holidays.manage` permissions must be defined and assignable via the permission management system.
- **EP-002 Leave Management:** Leave request records must store date ranges in a format that allows overlap comparison with holiday date ranges. Recalculation must be able to update the leave days consumed field on Approved requests.
- **EP-004 Attendance Management:** Attendance streak and matrix display logic must reference the holiday table when determining whether a date is a holiday.
- **Vietnamese Public Holiday Preset:** A system-maintained dataset of Vietnamese public holidays per year must be available internally. The number of available years is to be confirmed by the development team.

---

## 7. UX Optimizations

**Usability Considerations:**

- **Year filter as primary context anchor:** The year filter is positioned to the left of the search bar (left action area of the action bar), making it visually prominent as the primary scoping control. Changing the year immediately reloads the table and clears search, eliminating stale results.
- **Live Total Days:** On the Create/Edit form, Total Days updates as the user selects or changes dates (live calculation), giving HR immediate feedback on the scope of the holiday without requiring a form submit.
- **Pre-checked import rows:** The import preview defaults all rows to checked, minimizing clicks for the common case (import everything). HR only needs to uncheck rows they want to exclude.
- **Live import count in button label:** The "Import [N] Holidays" button reflects the current selection count in real time, so HR always knows exactly how many holidays will be added before confirming.
- **Silent duplicate skipping:** Duplicate holidays during import are skipped without error dialogs or interruptions. The toast reports the skip count post-import so HR is informed without being blocked.
- **Debounced search (300ms):** Search input waits 300ms after the user stops typing before querying, preventing excessive requests on each keystroke.
- **Skeleton loading:** While the holiday list is loading, skeleton placeholder rows are shown in the table area, preventing layout shift and signaling that content is incoming.
- **Recalculation transparency:** The delete confirmation dialog explicitly informs HR that approved leave requests will be recalculated. The post-delete toast confirms how many requests were affected, giving HR visibility without requiring them to navigate elsewhere.
- **Focus management on dialogs:** The delete confirmation dialog traps focus between its Cancel and Delete buttons, preventing keyboard users from accidentally interacting with background elements. Escape key maps to Cancel. Cancel has default focus to guard against accidental deletes.
- **Dirty form protection:** The discard dialog on cancel prevents accidental data loss when HR has made edits but changes their mind before saving.
- **Gear icon last column:** Action gear icon is placed in the rightmost column of the table, consistent with all other list screens in the system.
- **Empty state CTAs:** When a year has no holidays and no active search, the empty state includes actionable Import From Template and Add New buttons (for manage users), guiding HR directly to the resolution without navigating back to the action bar.

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full layout — action bar (year filter + search + buttons), table with 5 columns, pagination |
| Tablet (768–1024px) | Out of scope (web admin is desktop-only per EP-009 scope) |
| Mobile (<768px) | Out of scope (web admin is desktop-only per EP-009 scope) |

**Accessibility Requirements:**
- [x] Keyboard navigable — all interactive elements (year filter, search, table gear icons, form fields, dialog buttons) are reachable via Tab; Enter activates buttons
- [x] Screen reader compatible — table column headers are labelled; dialog has appropriate ARIA role; toast notifications are announced
- [x] Sufficient color contrast — danger Delete button meets WCAG AA contrast; skeleton loading uses system design tokens
- [x] Focus indicators visible — focus ring shown on all interactive elements
- [x] Focus trap in dialogs — delete confirmation and discard dialogs trap focus; Escape closes

**Design References:**
- Holidays list frame: Figma nodeId `3537:3821` (deep-link URL pending Figma file ID — see §8 Open Questions)
- Create/Edit form pattern reference: DR-001-005-02 (Create User — 600px centered card, header Cancel + Save, dirty/clean discard dialog)

---

## 8. Additional Information

### Out of Scope
- Payslip and payroll calculation (future DR when payroll module is built; holiday calendar will serve as an input reference at that time)
- Recurring / auto-repeating holidays across years (e.g., fixed-date holidays auto-copied to each new year)
- Per-department or per-employee holiday exceptions (all holidays apply organization-wide)
- Mobile view (this module is web admin only; mobile app out of scope for EP-009)
- Public holiday API integration (the Vietnamese public holiday preset is a system-maintained internal dataset, not fetched live from an external service)
- Overlap validation between two holidays in the same year (not blocked at the business rule level for this story)
- Status filter on the holiday list (only year filter and search are scoped)
- Bulk delete of multiple holidays

### Open Questions
- [ ] Which roles should receive `organization.holidays.manage` permission by default? — Owner: Product Owner — Status: Pending
- [ ] When a holiday is added, edited, or deleted and leave requests are recalculated, should the affected employees receive an in-app notification or email? — Owner: Product Owner — Status: Pending
- [ ] How many years of Vietnamese public holiday presets are available in the system? — Owner: Development Team — Status: Pending
- [ ] What is the exact Figma file ID for constructing deep-link URLs to the design? (Placeholder used in §7 Design References) — Owner: Design Team — Status: Pending

### Related Features
- DR-001-005-02 (Create User) — pattern reference for the full-page create/edit form layout
- EP-002 Leave Management — US-001-leave-requests: leave day calculation engine that consumes the holiday calendar
- EP-004 Attendance Management — US-001-attendance-list: attendance matrix "H" display and streak exclusion logic
- EP-001 US-004 Role & Permissions — defines `organization.holidays.view` and `organization.holidays.manage`

### Notes
- The Figma frame for the Create/Edit form has not yet been designed as of 2026-06-15. The form follows the Create User pattern (DR-001-005-02) per design spec decision.
- The Year field on the form is a read-only label (not an editable dropdown) to enforce the constraint that a holiday cannot be moved between years via the edit form. This is a deliberate design decision in the approved spec.
- The `total_days` value is computed at display time from `from_date` and `to_date`; it is not stored as a separate database field. Business documents reference it as a display value only.
- Leave day formula is defined in EP-002 and referenced here. Any change to the formula is owned by EP-002; this module only consumes the formula result during recalculation.
- The design spec is authoritative for this DR. Approved 2026-06-15.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-06-15 | BA Agent (DR-009-003-01) | Initial draft — full Holiday Management module (list, create/edit, delete, import, leave recalculation) |
