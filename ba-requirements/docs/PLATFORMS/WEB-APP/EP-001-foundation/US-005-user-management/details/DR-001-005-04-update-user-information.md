---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-04
detail_name: "Update User Information"
parent_requirement: FR-US-005-11
status: draft
version: "1.3"
created_date: "2026-03-26"
last_updated: "2026-06-29"
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
input_sources:
  - type: figma
    description: "Update Information screen (User Details sub-page)"
    node_id: "3122:6323"
    extraction_date: "2026-03-26"
---

# Detail Requirement: Update User Information

**Detail ID:** DR-001-005-04
**Parent Requirement:** FR-US-005-11
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.3

---

## 1. Use Case Description

As a **user with user management permission**, I want to **update a user's extended profile** — including personal information, work profile, ID cards, emergency contacts, and (with the right permissions) salary and banking details — so that **employee records stay accurate and complete as personal details, organizational assignments, identification, or compensation data change**.

**Purpose:** Enable administrators to keep the full user profile current — updating names, contact details, department/position assignments, education level, identification documents, emergency contacts, salary, and banking information — without needing to recreate the user account. This is an edit form for existing users, accessed from the User Details page, and mirrors the Create User form minus Email, Role, and Account Status, which remain separate dedicated actions. Salary and Banking sections are permission-gated for sensitive data protection.

**Target Users:**
- Any user with user management permission — full edit access to Personal, Work, ID Cards, Emergency Contact sections
- Users with `user.salary.view` / `user.salary.manage` — view-only or editable Salary section
- Users with `user.banking.view` / `user.banking.manage` — view-only or editable Banking section

**Key Functionality:**
- Pre-filled edit form with existing user data across all sections
- Personal Information editing (name, DOB, phone, gender, permanent address, temporary address, marital status, nationality, social insurance number, tax identification number, avatar)
- Work Profile editing (department, position, experience year, education level, line manager, CV, skills)
- ID Cards editing (front/back images, ID number, issue date)
- Emergency Contact editing (repeatable rows — full name, relationship, phone)
- Salary editing (base salary, insurance salary) — permission-gated by `user.salary.view` / `user.salary.manage`
- Banking editing (bank, account number, account name, transfer method) — permission-gated by `user.banking.view` / `user.banking.manage`
- Single Save button — stays on page after saving
- Unsaved changes protection via confirmation dialog with smart dirty check across all new fields

---

## 2. User Workflow

**Entry Point:** User Details page → "Update Information" in the left action panel.

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has user management permission (US-004)
- User is viewing a specific user's details (DR-001-005-03)

**Main Flow:**
1. User is on the User Details page (Overview tab)
2. User clicks "Update Information" in the left action panel
3. System displays the Update Information form pre-filled with the user's existing data across all visible sections (Personal Information, Work Profile, ID Cards, Emergency Contact, plus Salary and Banking if permitted)
4. System hides Salary section if user lacks `user.salary.view`; disables Salary inputs if user has `user.salary.view` but not `user.salary.manage`
5. System hides Banking section if user lacks `user.banking.view`; disables Banking inputs if user has `user.banking.view` but not `user.banking.manage`
6. User modifies any fields in any visible/editable section (including the Line Manager field in Work Profile, which is editable for users with `user.edit`)
7. User clicks Save (full-width button at bottom of form)
8. System validates all mandatory fields (Nationality and Education level included)
9. System saves changes — read-only sections (Salary/Banking when view-only) are preserved unchanged
10. System shows success toast "User information updated successfully"
11. User stays on the Update Information page with updated data

**Alternative Flows:**

- **Alt 1 — Validation fails:** Inline errors shown below invalid fields. Save button re-enables. Form is not submitted.
- **Alt 2 — Navigate away (form dirty):** User clicks another action (Overview, Change Email, back arrow, etc.) while form has unsaved changes → "Discard unsaved changes?" confirmation dialog appears → Confirm: navigates to target page, changes lost / Cancel: stays on form, changes preserved.
- **Alt 3 — Navigate away (form clean):** User clicks another action with no unsaved changes → navigates immediately, no dialog.
- **Alt 4 — Save with no changes:** User clicks Save without modifying anything → saves silently, no error message.
- **Alt 5 — Server error:** Save fails due to server error → error toast with retry suggestion, form data preserved.
- **Alt 6 — Current line manager is inactive (pre-fill):** Form loads with an inactive user as the current Line Manager → yellow/amber info banner shown above the Line Manager field reading *"Current line manager is inactive. Reassign to an active user or approval routing will fall back to [HR for leave / CEO for OT]."* → banner is informational and does NOT block Save → admin may keep the inactive manager or select a different active one (banner disappears immediately on selecting an active user).
- **Alt 7 — Cycle attempt on Line Manager selection:** Admin picks a user from the Line Manager dropdown whose own Line Manager is the current user (direct cycle) → inline error shown below the field: *"Cannot assign — selected user is already in this user's reporting chain (would create a cycle)"* → Save button is disabled until admin chooses a different option.
- **Alt 8 — Server-side cycle on save:** Save submitted with a line manager that would create a cycle (defense-in-depth, e.g., chain changed between load and save) → server rejects → error toast *"Line manager change creates a cycle — please choose a different user"* → form stays editable, no data lost.

**Exit Points:**
- **Success:** Toast shown, stays on page with updated data
- **Navigate away (clean):** Direct navigation to target
- **Navigate away (dirty + confirm discard):** Navigation to target, changes lost
- **Navigate away (dirty + cancel discard):** Stay on form, changes preserved

---

## 3. Field Definitions

### Input Fields — Personal Information Card

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Upload Avatar | Image upload | .png/.jpeg/.jpg/.webp only; max 2MB; max 500×500px | No | Existing avatar or placeholder | — | User profile photo |
| First name | Text input (278px, left) | Not empty | Yes (*) | Pre-filled | "Enter first name" | User's given name |
| Last name | Text input (278px, right) | Not empty | Yes (*) | Pre-filled | "Enter last name" | User's family name |
| Date of birth | Date picker (278px, left) | Not empty; valid date | Yes (*) | Pre-filled | "Enter date of birth" | User's birth date |
| Phone number | Text input (278px, right) | Not empty; Vietnam format: 10 digits starting with 0 | Yes (*) | Pre-filled | "Enter phone number" | Contact phone number |
| Gender | Dropdown (576px, full width) | Not empty | Yes (*) | Pre-filled | "Select gender" | User's gender |
| Marital status | Dropdown (576px, full width) | Optional | No | Pre-filled or empty | "Select marital status" | Single / Married / Other |
| Nationality | Searchable dropdown (576px, full width) | Not empty; must select a value from ISO country list | Yes (*) | Pre-filled with user's existing nationality | "Select nationality" | User's nationality (full ISO country list) |
| Permanent address | Text input (576px, full width) | Trimmed, max 500 chars | No | Pre-filled or empty | "Enter permanent address" | User's permanent residential address (renamed from "Address") |
| Temporary address | Text input (576px, full width) | Trimmed, max 500 chars | No | Pre-filled or empty | "Enter temporary address" | User's temporary residential address |
| Social Insurance Number | Text input (576px, full width) | Trimmed, max 50 chars; no format validation | No | Pre-filled or empty | "Enter social insurance number" | Employee's social insurance number (optional free-text) |
| Tax Identification Number | Text input (576px, full width) | Trimmed, max 50 chars; no format validation | No | Pre-filled or empty | "Enter tax identification number" | Employee's tax identification number (optional free-text) |

### Input Fields — Work Profile Card

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Department | Dropdown (278px, left) | Not empty; active departments only | Yes (*) | Pre-filled | "Select department" | Organizational unit assignment |
| Position | Dropdown (278px, right) | Not empty; active positions only | Yes (*) | Pre-filled | "Select position" | Job title assignment |
| Experience from (year) | Text/year input (576px) | Not empty; 4-digit year > 1900 and ≤ current year | Yes (*) | Pre-filled | "Enter year" | Career start year |
| Education level | Dropdown (576px, full width) | Not empty; must select | Yes (*) | Pre-filled from user data | "Select education level" | High school / College / Bachelor's degree / Master's degree / Doctorate |
| Line Manager | Searchable user dropdown (576px, full width) | Optional; must reference an existing user; cannot be self; cannot create a direct cycle (selected user's Line Manager is the current user) | No | Pre-filled with current Line Manager (or "No line manager" if empty) | "Select line manager (optional)" | The user this employee reports to. Options: active users only, EXCEPT (a) self is excluded, (b) entire subordinate chain is excluded transitively to prevent cycles, (c) the currently assigned manager is shown with `(Inactive)` suffix if deactivated so admin can keep the historical assignment. Option display format: `Full Name — Position, Department`. Search is case-insensitive across name, position, and department. A special "No line manager" option appears at top of list. |
| CV/Resumé | File input (576px) | Accepted formats TBD (aligned with Create User) | No | Existing file name or "No file chosen" | — | Resume document |
| Skills | "+ Add skills" button | None | No | Existing skills displayed | — | Employee competencies |

### Input Fields — ID Cards Card [NEW]

All fields pre-filled with existing user data on load.

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Front image | Image upload (full width) | .png/.jpeg/.jpg/.webp only; max 5MB | No | Existing image displayed; user can replace | — | ID card front photo |
| Back image | Image upload (full width) | .png/.jpeg/.jpg/.webp only; max 5MB | No | Existing image displayed; user can replace | — | ID card back photo |
| ID number | Text input (full width) | Plain text, max 50 chars | No | Pre-filled or empty | "Enter ID number" | National ID or identification number |
| Issue date | Date picker (full width) | Valid past date | No | Pre-filled or empty | "Select issue date" | Date the ID card was issued |

### Input Fields — Emergency Contact Card [NEW, repeatable]

Pre-filled with existing contacts on load. Section is entirely optional — user may remove all contacts. Unlimited rows; "+ Add contact" button below last row; each row has a remove (X) button.

| Field per row | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|---------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Full name | Text input | Trimmed, max 100 chars | No | Pre-filled per existing row | "Enter full name" | Contact's full name |
| Relationship | Text input | Trimmed, max 50 chars | No | Pre-filled per existing row | "Enter relationship" | Relationship to user (e.g., Spouse, Parent) |
| Phone number | Text input | Trimmed, max 20 chars | No | Pre-filled per existing row | "Enter phone number" | Contact's phone number |

### Input Fields — Salary Card [NEW, permission-gated]

Section visibility/editability is permission-gated:
- Without `user.salary.view`: section is hidden entirely (no placeholder card rendered)
- With `user.salary.view` only: section is visible but inputs are disabled (read-only); values are pre-filled
- With `user.salary.manage`: section is visible and fully editable

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Base salary | Numeric input (full width) | VND, min 0; display with thousand separators (e.g., 15,000,000) | No | Pre-filled or empty | "Enter base salary" | Monthly base salary in VND |
| Insurance salary | Numeric input (full width) | VND, min 0; display with thousand separators | No | Pre-filled or empty | "Enter insurance salary" | Salary basis for insurance calculations |

### Input Fields — Banking Card [NEW, permission-gated]

Section visibility/editability is permission-gated:
- Without `user.banking.view`: section is hidden entirely (no placeholder card rendered)
- With `user.banking.view` only: section is visible but inputs are disabled (read-only); bank account number is shown in full (NOT masked — this is the edit form where editors need to see the complete value)
- With `user.banking.manage`: section is visible and fully editable

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Bank | Searchable dropdown (full width) | Must select from hardcoded VN bank list | No | Pre-filled | "Select bank" | Bank name |
| Bank account number | Text input (full width) | Numeric only, 6-20 digits; shown in full (unmasked) | No | Pre-filled | "Enter bank account number" | Bank account number |
| Account name | Text input (full width) | Trimmed, max 100 chars | No | Pre-filled | "Enter account name" | Account holder name |
| Transfer method | Dropdown (full width) | Bank transfer / Cash | No | Pre-filled | "Select transfer method" | Payment transfer method |

### Validation Error Messages (Line Manager)

| Condition | Message | Display Location |
|-----------|---------|------------------|
| Direct cycle detected (selected user's Line Manager is the current user) | "Cannot assign — selected user is already in this user's reporting chain (would create a cycle)" | Inline below the Line Manager field; Save disabled until corrected |
| Self-selection (defense-in-depth — typically impossible since self is filtered from dropdown) | "Cannot set line manager to self" | Inline below the Line Manager field |
| Server-side cycle violation on save (e.g., chain changed between load and save) | "Line manager change creates a cycle — please choose a different user" | Error toast; form stays editable, no data lost |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Save | Button (primary, full-width 600px) | Bottom of form | Always visible | Validate → save → success toast → stay on page | Submit changes |
| Upload Avatar button | Button (secondary) | Within avatar upload area | Always visible | Opens file picker for image selection | Replace or set avatar |
| "+ Add skills" button | Button (secondary, 110×32px) | Below Skills label | Always visible | Opens skill selector | Add skills to user |
| Date picker calendar icon | Icon button | Inside Date of birth / Issue date fields | Always visible | Opens calendar date picker | Select date |
| ID front/back image upload buttons | Button (secondary) | Within each image upload area | Always visible | Opens file picker; replaces existing image | Replace ID card images |
| "+ Add contact" button | Button (secondary) | Below last Emergency Contact row | Always visible | Adds a new empty contact row | Add another emergency contact |
| Remove contact (X) button | Icon button | Right side of each Emergency Contact row | Always visible per row | Removes that row from the form | Remove an emergency contact |
| Line Manager dropdown | Searchable dropdown | Work Profile card, after Education level | Always visible (editable with `user.edit`) | Opens searchable list of selectable users + "No line manager" option | Assign / change / clear the line manager |
| Inactive-manager warning banner | Inline banner (yellow/amber) | Above the Line Manager field | Visible only when pre-filled Line Manager is a deactivated user | Disappears when admin selects a different active manager | Informational notice; does NOT block Save |

**Notes:**
- No Cancel button in the design — users navigate away via left action panel or back arrow
- Email, User Role, and Account Status fields are NOT present on this form
- Salary and Banking cards are conditionally rendered based on permissions — see card-level notes

---

## 4. Data Display

### Information Shown to User

| Data Element | Data Type | Display When Empty | Format | Business Meaning |
|-------------|-----------|-------------------|--------|------------------|
| Avatar | Image | Default placeholder silhouette | Thumbnail in upload area | User's profile photo |
| First name | Text | Input with placeholder | Geist Regular 14px in input field | Given name |
| Last name | Text | Input with placeholder | Geist Regular 14px in input field | Family name |
| Date of birth | Date | Input with placeholder | Date format in date picker | Birth date |
| Phone number | Text | Input with placeholder | Geist Regular 14px in input field | Contact number |
| Gender | Text | Dropdown with placeholder | Selected option in dropdown | Gender |
| Marital status | Text | Dropdown with placeholder | Selected option in dropdown | Marital status |
| Nationality | Text | Searchable dropdown with placeholder | Selected option in dropdown | Nationality |
| Permanent address | Text | Input with placeholder | Geist Regular 14px in input field | Permanent residential address |
| Temporary address | Text | Input with placeholder | Geist Regular 14px in input field | Temporary residential address |
| Social Insurance Number | Text | Input with placeholder | Geist Regular 14px in input field | Employee's social insurance number |
| Tax Identification Number | Text | Input with placeholder | Geist Regular 14px in input field | Employee's tax identification number |
| Department | Text | Dropdown with placeholder | Selected option in dropdown | Assigned department |
| Position | Text | Dropdown with placeholder | Selected option in dropdown | Assigned position |
| Experience from (year) | Text | Input with placeholder | 4-digit year in input field | Career start year |
| Education level | Text | Dropdown with placeholder | Selected option in dropdown | Highest education attained |
| Line Manager | Text (selected option) | Searchable dropdown showing "No line manager" selection | Selected option as `Full Name — Position, Department` (or `(Inactive)` suffix when the currently assigned manager is deactivated) | The user this employee reports to; used as the first-level approver for downstream leave/OT/request flows |
| CV/Resumé | File name | "No file chosen" | File name text next to Choose File button | Resume document |
| Skills | Tags/badges | "+ Add skills" button only | Skill tags with remove option | Assigned competencies |
| ID Card front image | Image | Empty upload area with prompt | Thumbnail with replace option | ID card front photo |
| ID Card back image | Image | Empty upload area with prompt | Thumbnail with replace option | ID card back photo |
| ID number | Text | Input with placeholder | Plain text in input field | National ID number |
| Issue date | Date | Date picker with placeholder | Date format in date picker | ID issue date |
| Emergency contacts | List of rows | Empty section with "+ Add contact" only | Repeatable rows with Full name / Relationship / Phone | Emergency contact list |
| Base salary | Numeric | Input with placeholder (or hidden if no permission) | VND with thousand separators | Monthly base salary |
| Insurance salary | Numeric | Input with placeholder (or hidden if no permission) | VND with thousand separators | Insurance salary basis |
| Bank | Text | Searchable dropdown with placeholder (or hidden) | Selected option in dropdown | Bank name |
| Bank account number | Text | Input with placeholder (or hidden) | Full digits, unmasked in edit form | Account number |
| Account name | Text | Input with placeholder (or hidden) | Plain text in input field | Account holder name |
| Transfer method | Text | Dropdown with placeholder (or hidden) | Selected option in dropdown | Payment method |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads | All visible fields pre-filled with existing user data |
| Loading | Data being fetched | Loading indicator / skeleton fields |
| Section hidden by permission | User lacks `user.salary.view` or `user.banking.view` | The corresponding card is not rendered at all (other cards remain stacked) |
| Section read-only by permission | User has `user.salary.view` but not `user.salary.manage` (likewise banking) | Card is visible, fields pre-filled, all inputs disabled; Save does not modify these values |
| Emergency Contact list (populated) | User has existing emergency contacts | Each contact rendered as a row with prefilled fields and a remove (X) button; "+ Add contact" below last row |
| Emergency Contact list (empty) | User had no contacts or removed them all | Empty card body with only "+ Add contact" button visible |
| Validation error (empty mandatory) | User clears a mandatory field and clicks Save | Inline error below field: "[Field name] is required" |
| Validation error (nationality empty) | User clears Nationality and clicks Save | Inline error: "Nationality is required" |
| Validation error (education level empty) | User clears Education level and clicks Save | Inline error: "Education level is required" |
| Validation error (phone format) | Invalid phone number format | Inline error: "Please enter a valid Vietnam phone number (10 digits starting with 0)" |
| Validation error (year range) | Year ≤ 1900 or > current year | Inline error: "Please enter a valid year (after 1900)" |
| Validation error (avatar size) | File exceeds 2MB | Inline error: "File size must not exceed 2MB" |
| Validation error (avatar format) | Unsupported file type | Inline error: "Accepted formats: .png, .jpeg, .jpg, .webp" |
| Validation error (ID image size) | ID card image exceeds 5MB | Inline error: "File size must not exceed 5MB" |
| Validation error (ID image format) | Unsupported ID card image type | Inline error: "Accepted formats: .png, .jpeg, .jpg, .webp" |
| Validation error (bank account format) | Bank account number not 6-20 numeric digits | Inline error: "Bank account number must be 6-20 digits" |
| Saving | User clicks Save, request in progress | Save button shows loading state (disabled + spinner) |
| Success | Changes saved successfully | Success toast: "User information updated successfully" |
| Discard confirmation | User navigates away with unsaved changes (in any visible section) | Modal: "Discard unsaved changes?" with Confirm + Cancel options |
| Server error | Save fails | Error toast with retry suggestion, form data preserved |
| Line manager assigned (pre-filled, active) | Form loads with an active current Line Manager | Field shows current manager as `Full Name — Position, Department`; no banner |
| Line manager assigned (pre-filled, inactive) | Form loads with a deactivated current Line Manager | Field shows current manager with `(Inactive)` suffix; yellow/amber warning banner shown above the field |
| Line manager not assigned (pre-filled) | Form loads with no Line Manager set | "No line manager" selected; no banner |
| Cycle attempted (selection invalid) | Admin selects a user whose Line Manager is the current user | Inline error below field; Save disabled until corrected |
| Line manager changed from inactive to active | Admin selects an active user replacing an inactive pre-fill | Warning banner disappears immediately; dirty state set |
| Server-side cycle on save | Server rejects save due to cycle (defense-in-depth) | Error toast; form stays editable; no data lost |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Update Information form displays within the User Details page when "Update Information" is clicked in the left action panel
- **AC-02:** Form uses a single-column layout (600px) with the following cards stacked vertically: (1) Personal Information, (2) Work Profile, (3) ID Cards, (4) Emergency Contact, (5) Salary (permission-gated), (6) Banking (permission-gated), then the full-width Save button. When permission-gated sections are hidden, remaining cards remain stacked in the same order.
- **AC-03:** All fields in every visible section are pre-filled with the user's existing data on load (including ID Cards images, Emergency Contact rows, Salary, and Banking when permitted)
- **AC-04:** Mandatory fields marked with asterisk (*) — 10 total: First name, Last name, Date of birth, Phone number, Gender, Nationality, Department, Position, Experience from (year), Education level
- **AC-05:** Email, User Role, and Account Status fields are NOT present on this form

**Validation:**
- **AC-06:** User cannot save with any mandatory field empty
- **AC-07:** Phone number must be 10 digits starting with 0 (Vietnam format)
- **AC-08:** Experience from (year) must be a valid 4-digit year, greater than 1900 and not exceeding the current year
- **AC-09:** Avatar upload accepts only .png/.jpeg/.jpg/.webp, max 2MB, max 500×500px
- **AC-10:** Validation errors display inline below the relevant field
- **AC-11:** Validation is triggered on Save (not on blur)

**Save Behavior:**
- **AC-12:** Save button saves changes, shows toast "User information updated successfully", and stays on the Update Information page with updated data
- **AC-13:** Save button shows loading state (disabled + spinner) while the request is in progress
- **AC-14:** Save with no changes is allowed silently (no error)

**Navigation & Unsaved Changes:**
- **AC-15:** No Cancel button on the form — users navigate via left action panel or back arrow
- **AC-16:** If form has unsaved changes and user navigates away, a "Discard unsaved changes?" confirmation dialog is shown
- **AC-17:** If form has no unsaved changes, navigation proceeds immediately without dialog
- **AC-18:** Smart dirty check — reverting a field to its original value makes the form "clean" again (no false positives)

**Avatar & Files:**
- **AC-19:** Existing avatar is displayed in the upload area on load; user can replace it
- **AC-20:** Existing CV/Resumé file name is shown; user can replace or remove it
- **AC-21:** Existing skills are displayed; user can add or remove skills via "+ Add skills"

**Access Control:**
- **AC-22:** Update Information action is visible only to users with user management permission
- **AC-23:** Direct URL access by unauthorized users redirects to an appropriate fallback

**Extended Profile Fields:**
- **AC-24:** Nationality field is pre-filled with the user's existing nationality on load and is mandatory — saving with an empty Nationality shows inline error "Nationality is required"
- **AC-25:** Education level field is pre-filled with the user's existing education level on load and is mandatory — saving with an empty Education level shows inline error "Education level is required"
- **AC-26:** ID Cards front and back images are pre-loaded if previously uploaded; user can replace either image with a new file (≤5MB, .png/.jpeg/.jpg/.webp)
- **AC-27:** Emergency Contact rows are pre-filled with existing contacts; user can add new rows via "+ Add contact" or remove any row via its X button — the section may be left entirely empty

**Permission-Gated Sections:**
- **AC-28:** Salary card is hidden entirely (no placeholder rendered) when user lacks `user.salary.view`
- **AC-29:** Salary card is read-only (visible with disabled inputs, pre-filled values) when user has `user.salary.view` but not `user.salary.manage`; saving does not modify Salary values in this state
- **AC-30:** Banking card is hidden entirely (no placeholder rendered) when user lacks `user.banking.view`
- **AC-31:** Banking card is read-only when user has `user.banking.view` but not `user.banking.manage`; bank account number is shown in full (unmasked) in the edit form because editors need to see the complete value — this deliberately differs from User Details which masks it; saving does not modify Banking values in read-only state

**Line Manager Field:**
- **AC-32:** Line Manager field appears in Work Profile (between Education level and CV/Resumé), pre-filled with the user's current Line Manager value on load
- **AC-33:** Line Manager field is optional — admin can clear it by selecting "No line manager" from the top of the dropdown
- **AC-34:** Dropdown excludes self AND the entire subordinate chain (transitive cycle prevention computed server-side); these users do not appear in the picker at all
- **AC-35:** Dropdown displays each option in the format `Full Name — Position, Department`
- **AC-36:** Dropdown search is case-insensitive and matches against full name, position, OR department
- **AC-37:** When the currently assigned manager is deactivated, that user is still shown in the dropdown options with `(Inactive)` suffix and remains selectable, so admin can preserve the historical assignment
- **AC-38:** A yellow/amber inactive-manager warning banner appears above the Line Manager field if AND only if the pre-filled current Line Manager is a deactivated user; banner copy is *"Current line manager is inactive. Reassign to an active user or approval routing will fall back to [HR for leave / CEO for OT]."*
- **AC-39:** Inactive-manager warning banner disappears immediately when admin selects a different active manager from the dropdown
- **AC-40:** Inactive-manager warning banner does NOT block Save — admin may intentionally keep the inactive manager
- **AC-41:** Direct-cycle attempt (selecting a user whose Line Manager is the current user) shows inline error *"Cannot assign — selected user is already in this user's reporting chain (would create a cycle)"* and disables Save until corrected
- **AC-42:** Server-side cycle violation on save shows error toast *"Line manager change creates a cycle — please choose a different user"*; form stays editable, no data lost
- **AC-43:** Line Manager field participates in the smart dirty check — changing the value sets dirty state; reverting to the originally loaded value clears the field's contribution to dirty state
- **AC-44:** Mandatory field count remains 10 (Line Manager is optional and does NOT increase the mandatory count)
- **AC-45:** No new permission is introduced for Line Manager editing — access is gated by the existing `user.edit` permission, same as other Work Profile fields

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path | Modify first name, click Save | Toast, stays on page, data updated | High |
| Multiple field changes | Modify name + department + phone | All changes saved, toast shown | High |
| Empty mandatory field | Clear first name, click Save | Inline error "First name is required" | High |
| Invalid phone | Enter "12345", click Save | Inline error about Vietnam format | High |
| Valid phone | Enter "0912345678" | Accepted, no error | High |
| Invalid year (too old) | Enter "1800" | Inline error "Please enter a valid year (after 1900)" | Medium |
| Invalid year (future) | Enter year > current year | Inline error "Year cannot exceed current year" | Medium |
| Avatar too large | Upload 5MB image | Inline error "File size must not exceed 2MB" | Medium |
| Avatar wrong format | Upload .gif file | Inline error about accepted formats | Medium |
| No changes + Save | Click Save without editing | Saves silently, no error | Medium |
| Navigate away (dirty) | Modify field, click Overview | "Discard unsaved changes?" dialog | High |
| Navigate away (clean) | No changes, click Overview | Navigates immediately | Medium |
| Confirm discard | Dirty form → dialog → Confirm | Navigates to Overview, changes lost | Medium |
| Cancel discard | Dirty form → dialog → Cancel | Stays on form, changes preserved | Medium |
| Smart dirty check | Change name, then revert to original | Form is clean, no dialog on navigate | Medium |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |
| Nationality mandatory empty | Clear Nationality, click Save | Inline error "Nationality is required" | High |
| Education level mandatory empty | Clear Education level, click Save | Inline error "Education level is required" | High |
| Add emergency contact | Click "+ Add contact" | New empty row appended, dirty state set | Medium |
| Remove emergency contact | Click X on a contact row | Row removed, dirty state set | Medium |
| Remove all emergency contacts | Remove every existing contact row, click Save | Saves successfully, no contacts persisted | Medium |
| Salary read-only without manage | User has `user.salary.view` only | Salary card visible, inputs disabled, values displayed; Save does not modify Salary | High |
| Salary hidden without view | User lacks `user.salary.view` | Salary card not rendered at all | High |
| Banking hidden without view | User lacks `user.banking.view` | Banking card not rendered at all | High |
| Banking unmasked in edit form | User opens form with banking visible | Bank account number shows full digits (not masked) | High |
| Banking read-only without manage | User has `user.banking.view` only | Banking card visible, inputs disabled, full account number shown; Save does not modify Banking | High |
| ID card image replace | Upload a new front image over existing one | New image replaces existing; dirty state set; saves on Save | Medium |
| ID card image too large | Upload 8MB image to ID front | Inline error "File size must not exceed 5MB" | Medium |
| Save preserves read-only sections | User without `user.salary.manage` saves form with other changes | Salary values unchanged; other changes persisted | High |
| Smart dirty check (new fields) | Change Temporary address, then revert | Form is clean, no dialog on navigate | Medium |
| Line Manager pre-fill (active) | Open form for a user with an active line manager | Field populated with `Full Name — Position, Department`; no warning banner | High |
| Line Manager pre-fill (inactive) | Open form for a user whose current line manager is deactivated | Field populated with `(Inactive)` suffix; yellow/amber warning banner shown above field | High |
| Line Manager pre-fill (none) | Open form for a user with no line manager set | "No line manager" selected; no banner | Medium |
| Change to active manager | Select an active user from dropdown (banner was shown) | Banner disappears immediately; dirty state set | High |
| Attempt to select self | Open Line Manager dropdown | Current user's own record is NOT in the option list (filtered) | High |
| Attempt to select a subordinate | Open Line Manager dropdown for a user who has direct or indirect reports | None of those subordinates appear in the option list (filtered transitively) | High |
| Direct cycle attempt | Select a user whose Line Manager is the current user | Inline error "Cannot assign — selected user is already in this user's reporting chain (would create a cycle)"; Save disabled until corrected | High |
| Save with inactive manager kept | Open form with inactive line manager; do not change; click Save | Save succeeds; on re-render, banner persists since manager is still inactive | Medium |
| Save with line manager changed to active | Open form with inactive line manager; pick active user; click Save | Save succeeds; banner cleared on re-render | High |
| Server-side cycle on save | Submit Save with a payload that would create a cycle (chain changed between load and save) | Error toast "Line manager change creates a cycle — please choose a different user"; form stays editable | High |
| Revert line manager change | Change line manager, then change back to the originally loaded value | Smart dirty check clears the field's contribution; if no other dirty fields, no discard dialog on navigate | Medium |
| Search dropdown by name | Type "sar" in dropdown search | Options matching name "sar" (case-insensitive) returned | Medium |
| Search dropdown by department | Type "engineering" in dropdown search | Options whose department matches "engineering" returned | Medium |
| Search dropdown by position | Type "CTO" in dropdown search | Options whose position matches "CTO" returned | Medium |
| Clear line manager | Select "No line manager" option for a user who currently has a line manager | Field shows "No line manager"; dirty state set; Save persists the cleared value | Medium |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** All fields are pre-filled with existing user data when the form loads — this is an edit form, not a create form
- **SR-02:** Email, User Role, and Account Status are NOT editable on this form — they are managed via separate dedicated actions (Change Email, Change User Role, Activate/Deactivate)
- **SR-03:** Phone number must follow Vietnam format: 10 digits, starting with 0 (e.g., 0912345678)
- **SR-04:** Experience from (year) must be > 1900 and ≤ current year
- **SR-05:** Avatar upload: .png/.jpeg/.jpg/.webp only, max 2MB, max 500×500px
- **SR-06:** Only users with user management permission can access the Update Information form
- **SR-07:** Saving with no changes is allowed — system processes it silently
- **SR-08:** Audit logging will be handled by a separate logging story — this DR does not define logging behavior
- **SR-09:** Department and Position dropdowns are populated from active departments/positions only (EP-008 data)
- **SR-10:** Skills are populated from the active skill catalog (EP-008 US-003)
- **SR-11:** Concurrent editing uses last-save-wins strategy — no conflict warning or optimistic locking
- **SR-12:** Smart dirty check compares current field values to originally loaded values — reverting to original values makes the form "clean"
- **SR-13:** Salary card is gated by `user.salary.view` (visibility) and `user.salary.manage` (editability); Banking card is gated by `user.banking.view` (visibility) and `user.banking.manage` (editability) — UI hiding/disabling is enforced both client-side and server-side (UI alone is not security)
- **SR-14:** "Permanent address" replaces the former "Address" field — same data column, renamed label only
- **SR-15:** All new fields (Marital status, Nationality, Permanent address, Temporary address, Social Insurance Number, Tax Identification Number, Education level, ID Cards fields, Emergency Contact rows, Salary fields, Banking fields) participate in the smart dirty check — modifying any of them sets the form dirty; reverting to the original loaded value clears their contribution to the dirty state
- **SR-16:** Bank account number is shown in full (unmasked) on this Update form — this deliberately differs from DR-001-005-03 (User Details), which masks the account number with only the last 4 digits visible to view-only users. Editors require the complete value to verify and modify the account.
- **SR-17:** Saving the form when the user lacks `user.salary.manage` preserves existing Salary values unchanged; saving when the user lacks `user.banking.manage` preserves existing Banking values unchanged — the system never writes empty/null values for permission-disabled sections
- **SR-18:** All new text fields trim whitespace on save (consistent with existing fields)
- **SR-19:** Nationality is mandatory — saving with an empty Nationality is rejected client-side and server-side
- **SR-20:** Education level is mandatory — saving with an empty Education level is rejected client-side and server-side
- **SR-21:** Line Manager field is editable by users with `user.edit` — no new permission is introduced by this field; access is identical to other Work Profile fields
- **SR-22:** Line Manager dropdown options exclude self AND the entire subordinate chain (transitive — computed server-side via the inverse Line Manager relationship); these users do not appear in the picker at all, providing a stronger cycle-prevention guarantee than direct-cycle checks alone
- **SR-23:** Inactive users may appear in the Line Manager dropdown options ONLY if they are the currently assigned manager — this preserves historical assignment editability while keeping the active candidate pool clean
- **SR-24:** Cycle validation runs both client-side (preview, immediate inline error on selection) and server-side (authoritative, on save) — UI prevention is not security; server is the source of truth
- **SR-25:** Inactive-manager warning is purely informational — it does NOT block Save; admin may intentionally keep the historical assignment with a deactivated user (downstream approval routes fall back to HR for leave / CEO for OT per design spec)
- **SR-26:** Smart dirty check includes the Line Manager field — changing the value sets dirty state; reverting to the originally loaded value clears its contribution
- **SR-27:** Permission scope model: this DR introduces NO scoped permissions. The two-variant `.team`/`.all` permission pattern documented in the design spec (`docs/superpowers/specs/2026-05-20-user-management-line-manager-design.md`) applies to downstream approval permissions (`leave.approve.*`, `ot.approve.*`, `request.approve.*`) owned by their respective downstream stories — not to user-management permissions, which remain global

**State Transitions:**
```
[User Details - Overview] → "Update Information" click → [Update Information form (pre-filled)]
[Update Information] → Save (valid) → [Update Information (updated data + success toast)]
[Update Information] → Save (invalid) → [Update Information (inline errors shown)]
[Update Information] → Navigate away (clean) → [Target page]
[Update Information] → Navigate away (dirty) → [Confirmation dialog]
[Confirmation dialog] → Confirm discard → [Target page]
[Confirmation dialog] → Cancel → [Update Information (changes preserved)]
[Update Information] → Select Line Manager option that creates direct cycle → [Inline error + Save disabled]
[Update Information] → Form pre-filled with inactive Line Manager → [Form shown + inactive-manager warning banner above field]
[Update Information] → Select active Line Manager (banner shown) → [Banner disappears + dirty state set]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — access control enforcement
- **Depends on:** DR-001-005-03 (User Details) — parent page providing the action panel and entry point
- **Depends on:** EP-008 US-001 (Department Management) — department dropdown data
- **Depends on:** EP-008 US-002 (Position Management) — position dropdown data
- **Depends on:** EP-008 US-003 (Skill Management) — skill catalog data
- **Depends on:** User list API — populates Line Manager dropdown options (active users + currently assigned manager if inactive)
- **Depends on:** Subordinate chain resolver API (server-side) — computes the transitive exclusion set for the Line Manager dropdown to prevent cycles

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** All fields pre-filled on load — user immediately sees current data and can identify what needs changing
- **UX-02:** Save button shows loading spinner while request is in progress, preventing double submission
- **UX-03:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-04:** "Discard unsaved changes?" dialog only appears when form is actually dirty — prevents unnecessary interruptions
- **UX-05:** Smart dirty check: form tracks dirty state by comparing current values to original loaded values — reverting a field to its original value makes the form "clean" again
- **UX-06:** Keyboard navigation supported — Tab through fields, Enter to submit
- **UX-07:** Date of birth uses a date picker with calendar icon for easy selection
- **UX-08:** Department and Position dropdowns support type-ahead search for quick selection

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Action panel (189px) + form (600px), side-by-side field pairs |
| Tablet (768-1024px) | Action panel collapses, form full-width, side-by-side pairs preserved |
| Mobile (<768px) | Action panel becomes top menu, fields stack vertically (single column) |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields and Save button
- [x] Screen reader compatible — field labels, error messages, toast announcements
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements

**Design References:**
- Figma: [Update Information](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3122-6323) (node `3122:6323`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Create User (DR-001-005-02) — shares field definitions and validation rules

---

## 8. Additional Information

### Out of Scope
- Change Email — separate action in User Details (planned)
- Change User Role — separate action in User Details (planned)
- Reset Password — separate action in User Details (planned)
- Activate/Deactivate — separate action in User Details (planned)
- Delete User — separate action in User Details (planned)
- User self-service profile editing (admin-only in this scope)
- Audit logging implementation (separate logging story)
- Concurrent editing conflict detection (last-save-wins strategy chosen)
- Permission catalog update in US-004 to register `user.salary.view`, `user.salary.manage`, `user.banking.view`, `user.banking.manage` — handled in a separate sync task
- Bank list management UI — bank list is hardcoded for now per PO decision
- Mobile/responsive layout for the new sections (desktop only in this scope)
- Audit logging for Salary/Banking field changes (separate logging story)
- Insurance salary calculation logic (admin manual input per existing business rule)
- Bulk reassignment of subordinates when a line manager leaves the organization (separate story)
- Acting / temporary delegation of line manager during vacation (separate story)
- Multiple / matrix line managers — v1 supports a single primary line manager only
- Automatic reassign picker on user delete — for now, deleting a user simply clears the Line Manager field on all affected subordinates (deferred to delete-user flow)
- Org-chart picker UI for visual line-manager selection (separate story)

### Open Questions
- **Bank list confirmation:** Final list of Vietnamese banks pending PO confirmation
- **Nationality source:** Full ISO 3166 country list or a curated subset for Vietnam-centric use?
- **ID Card permission gating:** Are stored ID card images sensitive enough to require a dedicated `user.idcards.view` permission like Salary/Banking? (Currently assumed under base `user.view`)
- **Salary currency:** VND only or multi-currency support needed later?
- **Emergency Contact phone format:** Vietnam phone format validation or free text?
- **Banking account number:** Luhn check or any structural validation beyond 6-20 digits numeric?
- **Subordinate-chain preview on Line Manager edit:** Should the form show a preview of the subordinate chain (e.g., "This user has 5 direct reports") for context before reassigning their line manager? — Owner: PO, Status: Pending
- **Reassign picker on delete:** When deleting a user with N subordinates, should the confirmation dialog offer a reassign picker instead of just clearing the Line Manager field on all affected subordinates? — Owner: PO, Status: Pending
- **Cross-department warning:** Should the form warn when selecting a line manager from a different department, in case it's unintentional? — Owner: PO, Status: Pending
- **Audit log scope for Line Manager changes:** Do we log every line manager change, or only when it affects approval routing for in-flight requests? — Owner: PO, Status: Pending

### Related Features
- **DR-001-005-03:** User Details — parent page providing the action panel and entry point
- **DR-001-005-02:** Create User — shares most fields (Update Information is a subset without Email, Role, Account Status)
- **DR-001-005-01:** User List — navigation origin (User List → User Details → Update Information)
- **EP-008 US-001:** Department Management — department dropdown data source
- **EP-008 US-002:** Position Management — position dropdown data source
- **EP-008 US-003:** Skill Management — skill catalog data source
- **US-001:** Authentication — user must be signed in
- **US-004:** Role & Permission Management — access control enforcement

### Notes
- The Figma design does not include a Cancel button — this is intentional. Users navigate away via the left action panel or back arrow, with a confirmation dialog protecting unsaved changes.
- The Update Information form is a **subset** of Create User: it excludes Email (Change Email action), User Role (Change User Role action), and Account Status (Activate/Deactivate action). These are deliberately separated into individual actions for audit clarity and permission granularity.
- The form layout is **single-column** (600px) unlike Create User's two-column layout, because the User Account card (Role + Status) is absent.
- The Last name placeholder in Figma says "Enter first name" — this is a copy-paste error. The correct placeholder is "Enter last name".
- **Permission gating is more granular than DR-02 (Create User) and DR-03 (User Details):** This DR explicitly distinguishes between `view` (visibility, read-only) and `manage` (editability) for both Salary and Banking sections. The form supports four states per section: hidden, read-only, editable, and fully absent (when no permission at all).
- **Banking account number is shown UNMASKED on this Update form**, deliberately differing from DR-001-005-03 (User Details) which masks the account number with only the last 4 digits visible. Rationale: editors need the complete value to verify and modify the account; view-only display elsewhere protects against shoulder-surfing of non-actionable views.
- New mandatory fields (Nationality, Education level) bring the mandatory count from 8 to 10. Both are pre-filled from existing user data on load, so this rarely affects the user experience except when underlying data is missing (legacy records).
- **Line Manager edit is a regular profile change** — it is NOT a separate dedicated action like Change Email, Change User Role, or Activate/Deactivate. It is edited inline on the Update Information form alongside other Work Profile fields, gated by the same `user.edit` permission. The mandatory field count remains 10 because Line Manager is optional (top-of-hierarchy users have none).

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-26 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-26 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context |
| 1.1 | 2026-05-18 | BA Agent | Added ID Cards, Emergency Contact, Salary (permission-gated), Banking (permission-gated) sections; added Nationality (mandatory) + Marital status + Temporary address + Education level (mandatory) fields; renamed Address → Permanent address; mandatory field count 8 → 10 |
| 1.2 | 2026-05-20 | BA Agent | Added editable Line Manager field to Work Profile (searchable user dropdown showing "Name — Position, Department"); dropdown excludes self AND subordinate chain to prevent cycles; client-side cycle preview + server enforcement; inactive-manager warning banner (informational, non-blocking); smart dirty check coverage; mandatory field count unchanged at 10; no new permissions added — gated by existing `user.edit`; two-variant `.team`/`.all` permission pattern documented in design spec for downstream approval stories |
| 1.3 | 2026-06-29 | BA Agent | Added Social Insurance Number and Tax Identification Number as optional free-text fields (max 50 chars, no format validation) to Personal Information card; updated Key Functionality, SR-15 (dirty check coverage); mandatory field count unchanged at 10 |
