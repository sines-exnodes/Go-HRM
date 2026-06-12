---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-02
detail_name: "Create User"
parent_requirement: FR-US-005-10
status: draft
version: "1.3"
created_date: "2026-03-24"
last_updated: "2026-05-20"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../TODO.yaml"
    relationship: sibling
  - path: "./DR-001-005-01-user-list.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Create A New User screen"
    node_id: "3050:3782"
    extraction_date: "2026-03-24"
---

# Detail Requirement: Create User

**Detail ID:** DR-001-005-02
**Parent Requirement:** FR-US-005-10
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.3

---

## 1. Use Case Description

As a **user with user management permission**, I want to **create a new user account by entering personal information, assigning a role, and linking to a department/position**, so that **the new user can access the HRM system with appropriate permissions and their employee profile is initialized**.

**Purpose:** Enable administrators to onboard new users into the HRM platform. Creating a user sets up their authentication credentials, assigns a system role (which governs their access), and links them to the organization structure (department + position). Once created, the user can immediately log in and access the system per their role.

**Target Users:** Any user with user management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Multi-section form: Personal Information (extended with Marital status, Nationality, Permanent address, Temporary address), User Account (role + status toggle + send login URL toggle), Employee Profile (dept/position/experience/education level + optional Line Manager + CV + skills), ID Cards (front/back images + ID number + issue date), Emergency Contact (repeatable rows), Salary (permission-gated), Banking (permission-gated)
- 14 mandatory fields, plus optional fields across all sections (including optional Line Manager assignment for reporting hierarchy)
- Two-column layout extended into 4 rows — Personal Info + User Account, Personal Info cont. + Employee Profile, ID Cards + Emergency Contact, Salary + Banking (permission-gated)
- Email uniqueness validation (server-side on submit — both inline error and error toast)
- Dropdown selectors with built-in search for User Role, Department, Position, Nationality, Bank
- File uploads for Avatar (.png/.jpeg/.jpg/.webp, max 5MB), CV/Resumé (PDF/DOCX, max 5MB), and ID Card front/back images (.png/.jpeg/.jpg/.webp, max 5MB each)
- Skills multi-select via "+ Add skills" button; Emergency Contacts repeatable via "+ Add contact" button
- Salary and Banking sections hidden entirely without `user.salary.view` / `user.banking.view` permissions
- Returns to User List on success or cancel

---

## 2. User Workflow

**Entry Point:** User List → click "+ Add New" button (visible only to users with management permission)

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has user management permission (US-004)

**Main Flow:**
1. User clicks "+ Add New" on the User List page
2. System navigates to the Create A New User page
3. System loads dropdown data from APIs: roles (US-004), departments (EP-008 US-001), positions (EP-008 US-002), gender options, marital status options, nationality (ISO country list), education levels, skills (EP-008 US-003), bank list (hardcoded VN banks)
4. System evaluates current user's permissions: `user.salary.view`/`user.salary.manage` and `user.banking.view`/`user.banking.manage` — Salary and Banking cards are hidden if their `view` permission is absent
5. Form displays with cards laid out in 4 rows: Row 1 (Personal Info top + User Account), Row 2 (Personal Info cont. + Employee Profile), Row 3 (ID Cards + Emergency Contact), Row 4 (Salary + Banking, if permission allows). Nationality defaults to "Vietnam"; Account status toggle ON by default
6. User fills **Personal Information**: enters first name, last name, email, phone number, date of birth, selects gender, confirms or changes Nationality (mandatory); optionally selects Marital status, uploads avatar, enters Permanent address and Temporary address
7. User fills **User Account**: selects a user role from dropdown (with search), sets account status toggle (default: Active), confirms "Send login URL" toggle (default: ON; disabled if Account status is Inactive)
8. User fills **Employee Profile**: selects department, position, and Education level (mandatory) from dropdowns, enters experience year; optionally selects a Line Manager from the searchable user dropdown (or leaves as "No line manager"); optionally uploads CV/Resumé and adds skills
9. User optionally fills **ID Cards**: uploads front/back images, enters ID number, picks issue date
10. User optionally fills **Emergency Contact**: clicks "+ Add contact" to add 0..N rows, each containing full name, relationship, phone number; rows can be removed via X
11. If visible — User optionally fills **Salary**: enters base salary and insurance salary in VND (thousand separators auto-formatted)
12. If visible — User optionally fills **Banking**: selects bank, enters bank account number, account name, selects transfer method
13. User clicks **Save**
14. System trims whitespace from all text fields (including new fields)
15. System validates client-side: all 14 mandatory fields filled, email format valid, date of birth is valid past date, experience year is valid and ≤ current year, file types/sizes valid, bank account number is 6-20 digits, salary values are non-negative numerics, ID card issue date is a valid past date
16. System submits to server; server validates email uniqueness and re-validates permission-gated sections
17. If all validations pass: system creates the user account with all provided data including ID cards, emergency contacts, salary, and banking
18. If "Send login URL" toggle is ON: system sends a first-login invitation email to the user's email address
19. System displays success toast: "User '[first name] [last name]' has been created"
20. System redirects to User List — new user visible in list with Active (or Inactive) status badge

**Alternative Flows:**

- **Alt 1 — Empty mandatory field:** Inline error below the empty field: "[Field name] is required". Form is not submitted. User corrects and retries.
- **Alt 2 — Invalid email format:** Inline error: "Please enter a valid email address". Form is not submitted.
- **Alt 3 — Duplicate email (server-side):** Server returns duplicate error. Both inline error "This email is already in use" AND error toast "This email is already in use" displayed. User enters a different email.
- **Alt 4 — Invalid avatar file:** Inline error below Upload Avatar: "Only PNG, JPEG, JPG, and WEBP files are accepted" or "File size must not exceed 5MB".
- **Alt 5 — Invalid CV file:** Inline error below CV/Resumé: "Only PDF and DOCX files are accepted" or "File size must not exceed 5MB".
- **Alt 6 — Cancel (form modified):** System shows "Discard unsaved changes?" dialog. Confirm → redirect to User List. Cancel → stay on form.
- **Alt 7 — Cancel (form untouched):** System redirects to User List immediately without confirmation.
- **Alt 8 — Dropdown API fails:** Error message in affected section with retry option. Other sections remain functional.
- **Alt 9 — Nationality not selected:** Inline error "Nationality is required" below the Nationality dropdown. Default "Vietnam" prevents this in typical flow.
- **Alt 10 — Education level not selected:** Inline error "Education level is required" below the Education level dropdown.
- **Alt 11 — Invalid ID Card image:** Inline error below the affected front/back upload: "Only PNG, JPEG, JPG, and WEBP files are accepted" or "File size must not exceed 5MB".
- **Alt 12 — ID Card issue date in future:** Inline error: "Issue date cannot be in the future".
- **Alt 13 — Invalid bank account number:** Inline error below Bank account number: "Bank account number must be 6 to 20 digits".
- **Alt 14 — Negative salary value:** Inline error: "Salary cannot be negative".
- **Alt 15 — Emergency contact remove:** Clicking X on an emergency contact row removes that row immediately — no confirmation. If only one empty row remains, the section returns to its empty state.
- **Alt 16 — Salary/Banking section hidden:** If current user lacks `user.salary.view`, the entire Salary card is omitted from the layout (not greyed out). Same for Banking.
- **Alt 17 — Selected Line Manager became inactive:** If the selected Line Manager was deactivated between dropdown load and form save, server returns an error. Both inline error "Selected line manager is no longer active" below the Line Manager field AND error toast displayed. User selects a different line manager or chooses "No line manager".

**Exit Points:**
- **Success:** User created → toast → redirect to User List
- **Cancel:** Redirect to User List (with or without confirmation depending on form state)
- **Error:** Validation errors shown inline (+ error toast for server-side duplicate email); user corrects and retries

---

## 3. Field Definitions

### Input Fields — Personal Information

| Field Name | Field Type | Validation Rule | Mandatory | Default | Max Length | Description |
|------------|------------|-----------------|-----------|---------|------------|-------------|
| Avatar | Image upload | PNG, JPEG, JPG, WEBP; max 5MB; validated client-side on file selection | No | No avatar | N/A | Profile photo. Upload area with "Upload Avatar" button and format hint ".png/.jpeg/.jpg/.webp" |
| First name | Text input | Not empty; trimmed | Yes (*) | Empty | 100 chars | Placeholder: "Enter first name" (278px, left half) |
| Last name | Text input | Not empty; trimmed | Yes (*) | Empty | 100 chars | Placeholder: "Enter last name" (278px, right half) |
| Email | Text input | Not empty; valid email format; unique (**server-side on submit**); trimmed | Yes (*) | Empty | 255 chars | Placeholder: "Enter email" (278px, left half) |
| Phone number | Text input | Not empty; trimmed | Yes (*) | Empty | 20 chars | Placeholder: "Enter phone number" (278px, right half) |
| Date of birth | Date picker | Not empty; must be a valid date in the past | Yes (*) | Empty | N/A | Placeholder: "Enter date of birth" with calendar icon (278px, left half) |
| Gender | Dropdown | Not empty; must select a value | Yes (*) | "Select gender" | N/A | Options: Male, Female, Other, Prefer not to say (pending PO). No dropdown search (3-4 options). (278px, right half) |
| Marital status | Dropdown | Trimmed; no other validation | No | "Select marital status" | N/A | Options: Single, Married, Other. No dropdown search (3 options). (278px, left half) |
| Nationality | Searchable dropdown | Not empty; must select a value | Yes (*) | "Vietnam" | N/A | Options: full ISO country list. Includes search field. Placeholder: "Select nationality". (278px, right half) |
| Permanent address | Text input | Trimmed; no other validation | No | Empty | 500 chars | Placeholder: "Enter permanent address" (576px, full width). Previously named "Address". |
| Temporary address | Text input | Trimmed; no other validation | No | Empty | 500 chars | Placeholder: "Enter temporary address" (576px, full width) |

### Input Fields — User Account

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| User Role | Dropdown with search | Not empty; must select a role | Yes (*) | "Select user role" | Options loaded dynamically from US-004. Dropdown includes search field. (576px) |
| Account status | Toggle switch | Always has a value | Yes (*) | Active (on) | On = Active (user can log in), Off = Inactive (user blocked). Label: "Activate/Deactivate" |
| Send login URL | Toggle switch | Always has a value; **disabled when Account status is Inactive** | Yes (*) | ON (send) | On = system sends first-login invitation email to user's email after save. Off = no email sent. Disabled + forced OFF when Account status is Inactive. |

### Input Fields — Employee Profile

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Department | Dropdown with search | Not empty; must select a department | Yes (*) | "Select department" | Options loaded from EP-008 US-001. Dropdown includes search field. (278px, left half) |
| Position | Dropdown with search | Not empty; must select a position | Yes (*) | "Select position" | Options loaded from EP-008 US-002. Dropdown includes search field. (278px, right half) |
| Experience from (year) | Text/number input | Not empty; valid 4-digit year; must be ≤ current year | Yes (*) | Empty | Placeholder: "Enter year" (576px) |
| Education level | Dropdown | Not empty; must select a value | Yes (*) | "Select education level" | Options: High school, College, Bachelor's degree, Master's degree, Doctorate. No dropdown search (5 options). (576px) |
| Line Manager | Searchable user dropdown | Server-side validates selected user is still active at save time | No | "No line manager" | Options: active users loaded dynamically from API (inactive users excluded server-side). Top-of-list special option: "No line manager". Option display format: `Full Name — Position, Department` (e.g., `Sarah Le — CTO, Engineering`). Search matches name, position, or department (case-insensitive). Placeholder: "Select line manager (optional)". (576px, full width) |
| CV/Resumé | File input | PDF, DOCX; max 5MB; validated client-side on file selection | No | No file chosen | "Choose File / No file chosen" (576px) |
| Skills | Multi-select | No minimum required | No | None selected | "+ Add skills" button opens selector. Skills loaded from EP-008 US-003. |

### Input Fields — ID Cards

| Field Name | Field Type | Validation Rule | Mandatory | Default | Max Length | Description |
|------------|------------|-----------------|-----------|---------|------------|-------------|
| Front image | Image upload | PNG, JPEG, JPG, WEBP; max 5MB; validated client-side on file selection | No | No file chosen | N/A | Upload area for ID card front side. Format hint ".png/.jpeg/.jpg/.webp". (left half) |
| Back image | Image upload | PNG, JPEG, JPG, WEBP; max 5MB; validated client-side on file selection | No | No file chosen | N/A | Upload area for ID card back side. Format hint ".png/.jpeg/.jpg/.webp". (right half) |
| ID number | Text input | Trimmed; plain text (no format validation) | No | Empty | 50 chars | Placeholder: "Enter ID number" (278px, left half) |
| Issue date | Date picker | Valid date in the past | No | Empty | N/A | Placeholder: "Enter issue date" with calendar icon (278px, right half) |

### Input Fields — Emergency Contact (repeatable rows, 0..N)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Max Length | Description |
|------------|------------|-----------------|-----------|---------|------------|-------------|
| Full name | Text input | Trimmed; no other validation | No | Empty | 100 chars | Placeholder: "Enter full name" |
| Relationship | Text input | Trimmed; no other validation | No | Empty | 50 chars | Placeholder: "Enter relationship" (e.g., Spouse, Parent, Sibling) |
| Phone number | Text input | Trimmed; no other validation | No | Empty | 20 chars | Placeholder: "Enter phone number" |

- Section is entirely optional — saving with zero contacts is allowed
- Unlimited rows; "+ Add contact" button below last row
- Each row has remove (X) button — removing a row deletes it immediately, no confirmation

### Input Fields — Salary (permission-gated: `user.salary.view` / `user.salary.manage`)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Base salary | Numeric input | Non-negative number; min 0 | No | Empty | VND with thousand separators auto-formatted (e.g., "15,000,000"). Placeholder: "Enter base salary" (278px, left half) |
| Insurance salary | Numeric input | Non-negative number; min 0 | No | Empty | VND with thousand separators auto-formatted. Placeholder: "Enter insurance salary" (278px, right half) |

- Section hidden entirely without `user.salary.view`
- Section visible but read-only without `user.salary.manage`

### Input Fields — Banking (permission-gated: `user.banking.view` / `user.banking.manage`)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Max Length | Description |
|------------|------------|-----------------|-----------|---------|------------|-------------|
| Bank | Searchable dropdown | Must select from list | No | "Select bank" | N/A | Options: Vietcombank, BIDV, Techcombank, MB Bank, VPBank, Agribank, ACB, VIB, Sacombank, TPBank, SHB, HDBank, Eximbank, OCB, SeABank, MSB, LienVietPostBank, NCB, ABBank, NamABank, PVcomBank. Includes search field. (278px, left half) |
| Bank account number | Text input | Numeric; 6 to 20 digits | No | Empty | 20 chars | Placeholder: "Enter bank account number" (278px, right half) |
| Account name | Text input | Trimmed; no other validation | No | Empty | 100 chars | Placeholder: "Enter account name" (278px, left half) |
| Transfer method | Dropdown | Must select if not empty | No | "Select transfer method" | N/A | Options: Bank transfer, Cash. (278px, right half) |

- Section hidden entirely without `user.banking.view`
- Section visible but read-only without `user.banking.manage`

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Left in action bar | Always visible | If dirty → "Discard changes?" dialog; if clean → redirect | Discard and return to User List |
| Save | Button (primary) | Right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → redirect | Save and return to User List |
| Upload Avatar | Button | Inside avatar upload area | Always visible | Opens file picker for image | Upload profile photo |
| + Add skills | Button (secondary) | Inside Employee Profile card | Always visible | Opens skills selector | Add skills to user profile |
| Upload Front (ID Card) | Button | Inside ID Cards card (left) | Always visible | Opens file picker for image | Upload ID card front image |
| Upload Back (ID Card) | Button | Inside ID Cards card (right) | Always visible | Opens file picker for image | Upload ID card back image |
| + Add contact | Button (secondary) | Inside Emergency Contact card, below last row | Always visible | Adds a new empty contact row | Add an emergency contact row |
| X (remove contact) | Icon button | Inside each Emergency Contact row | Always visible per row | Removes that row immediately | Remove an emergency contact row |

### Validation Error Messages

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Any mandatory field empty | "[Field name] is required" | Inline, below the field |
| Invalid email format | "Please enter a valid email address" | Inline, below Email field |
| Duplicate email (server-side) | "This email is already in use" | **Both:** inline below Email + error toast |
| Invalid avatar file type | "Only PNG, JPEG, JPG, and WEBP files are accepted" | Inline, below Upload Avatar |
| Avatar file too large | "File size must not exceed 5MB" | Inline, below Upload Avatar |
| Invalid CV file type | "Only PDF and DOCX files are accepted" | Inline, below CV/Resumé |
| CV file too large | "File size must not exceed 5MB" | Inline, below CV/Resumé |
| Invalid date of birth | "Please enter a valid date of birth" | Inline, below Date of birth |
| Future date of birth | "Date of birth cannot be in the future" | Inline, below Date of birth |
| Experience year invalid | "Please enter a valid year" | Inline, below Experience from |
| Experience year in future | "Year cannot be in the future" | Inline, below Experience from |
| Nationality empty | "Nationality is required" | Inline, below Nationality |
| Education level empty | "Education level is required" | Inline, below Education level |
| Invalid ID Card image type | "Only PNG, JPEG, JPG, and WEBP files are accepted" | Inline, below the affected upload |
| ID Card image too large | "File size must not exceed 5MB" | Inline, below the affected upload |
| ID Card issue date in future | "Issue date cannot be in the future" | Inline, below Issue date |
| Invalid bank account number | "Bank account number must be 6 to 20 digits" | Inline, below Bank account number |
| Negative salary value | "Salary cannot be negative" | Inline, below the affected salary field |
| Selected line manager became inactive (server-side, on save) | "Selected line manager is no longer active" | **Both:** inline below Line Manager field + error toast |

---

## 4. Data Display

### Information Shown to User

**Personal Information card (left column, 600×615px):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Avatar upload area | Image zone | "Upload Avatar" button + ".png/.jpeg/.jpg/.webp" hint | Centered in card | Profile photo |
| First name | Text input | Placeholder: "Enter first name" | 278px, left half | User's first name |
| Last name | Text input | Placeholder: "Enter last name" | 278px, right half | User's last name |
| Email | Text input | Placeholder: "Enter email" | 278px, left half | User's login email |
| Phone number | Text input | Placeholder: "Enter phone number" | 278px, right half | Contact number |
| Date of birth | Date picker | Placeholder: "Enter date of birth" + calendar icon | 278px, left half | Birth date |
| Gender | Dropdown | Placeholder: "Select gender" | 278px, right half | Gender |
| Marital status | Dropdown | Placeholder: "Select marital status" | 278px, left half | Marital status (optional) |
| Nationality | Searchable dropdown | Default: "Vietnam" | 278px, right half | Country of citizenship (mandatory) |
| Permanent address | Text input | Placeholder: "Enter permanent address" | 576px, full width | Permanent residence address (optional) |
| Temporary address | Text input | Placeholder: "Enter temporary address" | 576px, full width | Current temporary address (optional) |

**User Account card (right column top, 600×200px):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| User Role | Dropdown with search | Placeholder: "Select user role" | 576px | Role assignment |
| Account status | Toggle switch | Default: Active (on) | Switch + "Activate/Deactivate" label | Login access control |
| Send login URL | Toggle switch | Default: ON (send) | Switch + label; disabled when Account status is Inactive | Controls whether first-login invitation email is sent |

**Employee Profile card (right column bottom, 600×380px):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Department | Dropdown with search | Placeholder: "Select department" | 278px, left half | Org assignment |
| Position | Dropdown with search | Placeholder: "Select position" | 278px, right half | Job assignment |
| Experience from (year) | Text/number | Placeholder: "Enter year" | 576px | Career start year |
| Education level | Dropdown | Placeholder: "Select education level" | 576px | Highest education completed (mandatory) |
| Line Manager | Searchable user dropdown | Placeholder: "Select line manager (optional)" | 576px, full width | Who this user reports to (optional) |
| CV/Resumé | File input | "Choose File / No file chosen" | 576px | Resume document |
| Skills | Multi-select | "+ Add skills" button only | Button → selector → chips | Competency tags |

**ID Cards card (row 3 left column):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Front image | Image upload | "Upload Front" button + ".png/.jpeg/.jpg/.webp" hint | 278px, left half | ID card front side photo |
| Back image | Image upload | "Upload Back" button + ".png/.jpeg/.jpg/.webp" hint | 278px, right half | ID card back side photo |
| ID number | Text input | Placeholder: "Enter ID number" | 278px, left half | Government ID number |
| Issue date | Date picker | Placeholder: "Enter issue date" + calendar icon | 278px, right half | Date ID was issued |

**Emergency Contact card (row 3 right column, repeatable):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Contact row | Composite (3 fields + remove icon) | No rows initially; "+ Add contact" button visible | Each row: full name, relationship, phone + X remove | One emergency contact entry |
| + Add contact | Button | Always visible below last row | Secondary button | Adds an empty contact row |

**Salary card (row 4 left column, permission-gated):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Base salary | Numeric input | Placeholder: "Enter base salary" | 278px, left half; VND thousand separators | Monthly base salary (optional) |
| Insurance salary | Numeric input | Placeholder: "Enter insurance salary" | 278px, right half; VND thousand separators | Salary used for insurance calculation (optional) |

Card entirely hidden when current user lacks `user.salary.view`. Read-only when user has `view` but lacks `manage`.

**Banking card (row 4 right column, permission-gated):**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Bank | Searchable dropdown | Placeholder: "Select bank" | 278px, left half | Bank for salary disbursement |
| Bank account number | Text input | Placeholder: "Enter bank account number" | 278px, right half | Account number (numeric, 6-20 digits) |
| Account name | Text input | Placeholder: "Enter account name" | 278px, left half | Name on the bank account |
| Transfer method | Dropdown | Placeholder: "Select transfer method" | 278px, right half | How salary is paid (Bank transfer / Cash) |

Card entirely hidden when current user lacks `user.banking.view`. Read-only when user has `view` but lacks `manage`.

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads, APIs loaded | 3 cards with empty fields; Account status ON (Active); Send login URL ON; Save and Cancel enabled |
| Account Inactive selected | Admin toggles Account status OFF | Send login URL toggle is disabled and forced OFF (greyed out) |
| Account Active re-selected | Admin toggles Account status back ON | Send login URL toggle is re-enabled and defaults back to ON |
| Loading dropdowns | API calls in progress | Loading indicators in Role, Department, Position dropdowns |
| Dropdown API error | API fails for a dropdown | Error in affected dropdown with retry option; other fields functional |
| Avatar selected | Valid image chosen | Preview thumbnail replaces upload area |
| CV selected | Valid file chosen | Filename displayed next to "Choose File" |
| Skills added | Skills selected | Skill tags/chips displayed below "+ Add skills" button; each removable via X |
| Validation error — mandatory | Save with empty mandatory fields | Inline error(s) below each empty mandatory field (all shown simultaneously) |
| Validation error — email format | Invalid email | Inline error: "Please enter a valid email address" |
| Validation error — duplicate email | Server returns duplicate | Inline error + error toast: "This email is already in use" |
| Validation error — file | Invalid file type or size | Inline error below the relevant upload field |
| Validation error — date/year | Invalid date or future year | Inline error below the relevant field |
| Saving | Save clicked, request in progress | Save button shows spinner + disabled; Cancel disabled |
| Success | User created | Toast: "User '[First] [Last]' has been created" → redirect to User List |
| Discard confirmation | Cancel with modified form | Modal: "Discard unsaved changes?" with Confirm and Cancel |
| Nationality default | Page loads | Nationality dropdown pre-filled with "Vietnam" |
| Emergency Contact empty | No rows added | Card shows only the "+ Add contact" button |
| Emergency Contact row added | "+ Add contact" clicked | New empty row appears with Full name, Relationship, Phone + X remove icon |
| Emergency Contact row removed | X icon clicked on a row | That row is removed immediately, no confirmation |
| Salary section hidden | User lacks `user.salary.view` | Salary card not rendered; surviving Row 4 section spans full width |
| Salary section read-only | User has `view` but lacks `manage` | Salary fields visible but disabled — cannot be edited |
| Banking section hidden | User lacks `user.banking.view` | Banking card not rendered; surviving Row 4 section spans full width |
| Banking section read-only | User has `view` but lacks `manage` | Banking fields visible but disabled — cannot be edited |
| Both Salary + Banking hidden | User lacks both `view` permissions | Row 4 omitted entirely |
| Salary thousand separator | User types "15000000" in Base salary | Field auto-formats to "15,000,000" |
| ID Card image selected | Valid image uploaded | Preview thumbnail replaces upload area for that side |

### Page Layout (from Figma)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Users Management / Users / Create A New User                  │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A New User                       [Cancel]    [Save] │
│              │                                                             │
│              │  Row 1                                                      │
│              │  ┌─────────────────────────┐  ┌─────────────────────────┐   │
│              │  │ Personal Information    │  │ User Account            │   │
│              │  │  ┌───────────────────┐  │  │ * User Role    [▾ 🔍]  │   │
│              │  │  │  Upload Avatar    │  │  │ * Account status  [⊙]  │   │
│              │  │  │  .png/.jpeg/...   │  │  │   Activate/Deactivate  │   │
│              │  │  └───────────────────┘  │  │ * Send login URL  [⊙]  │   │
│              │  │ *First name  *Last name │  └─────────────────────────┘   │
│              │  │ *Email     *Phone number│                                │
│              │  │ *DOB [📅]  *Gender [▾] │  Row 2                         │
│              │  │  Marital status [▾]     │  ┌─────────────────────────┐   │
│              │  │  *Nationality [▾ 🔍]    │  │ Employee Profile        │   │
│              │  │  Permanent address      │  │ *Dept [▾ 🔍] *Pos [▾ 🔍]│   │
│              │  │  Temporary address      │  │ *Experience from (year) │   │
│              │  └─────────────────────────┘  │ *Education level [▾]    │   │
│              │                               │  Line Manager [▾ 🔍]    │   │
│              │                               │  CV/Resumé [Choose File]│   │
│              │                               │  Skills [+ Add skills]  │   │
│              │                               └─────────────────────────┘   │
│              │                                                             │
│              │  Row 3                                                      │
│              │  ┌─────────────────────────┐  ┌─────────────────────────┐   │
│              │  │ ID Cards                │  │ Emergency Contact       │   │
│              │  │ [Upload Front][UpldBack]│  │  Full name              │   │
│              │  │  ID number  Issue date  │  │  Relationship  Phone X  │   │
│              │  │                         │  │  [+ Add contact]        │   │
│              │  └─────────────────────────┘  └─────────────────────────┘   │
│              │                                                             │
│              │  Row 4  (permission-gated — hidden if user lacks `view`)    │
│              │  ┌─────────────────────────┐  ┌─────────────────────────┐   │
│              │  │ Salary 🔒               │  │ Banking 🔒              │   │
│              │  │  Base salary (VND)      │  │  Bank [▾ 🔍]            │   │
│              │  │  Insurance salary (VND) │  │  Account number         │   │
│              │  │                         │  │  Account name           │   │
│              │  │                         │  │  Transfer method [▾]    │   │
│              │  └─────────────────────────┘  └─────────────────────────┘   │
└──────────────┴──────────────────────────────────────────────────────────────┘

Layout note: If one Row 4 section is hidden by permission, the surviving
section spans full width. If both are hidden, Row 4 is omitted entirely.
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Create User page displays with page title "Create A New User" in Geist Semibold 24px
- **AC-02:** Form uses a two-column, 4-row layout: Row 1 (Personal Information top + User Account), Row 2 (Personal Information cont. + Employee Profile), Row 3 (ID Cards + Emergency Contact), Row 4 (Salary + Banking, permission-gated). When permission-gated sections are hidden, the surviving Row 4 section spans full width; if both are hidden, Row 4 is omitted entirely.
- **AC-03:** Two buttons visible: Cancel and Save (no "Save & Create Another")

**Personal Information:**
- **AC-04:** Avatar upload area accepts PNG, JPEG, JPG, WEBP files up to 5MB
- **AC-05:** After selecting a valid avatar, a preview thumbnail replaces the upload area
- **AC-06:** First name and Last name are displayed side-by-side (278px each)
- **AC-07:** Email and Phone number are displayed side-by-side (278px each)
- **AC-08:** Date of birth shows a date picker (calendar icon) and Gender shows a dropdown — side-by-side
- **AC-09:** Permanent address (previously "Address") and Temporary address are each full-width (576px) and optional (no asterisk). Both accept up to 500 characters and are trimmed on save.

**User Account:**
- **AC-10:** User Role dropdown loads available roles from US-004 dynamically and includes a search field for quick option finding
- **AC-11:** Account status toggle defaults to Active (on) when creating a new user
- **AC-12:** Toggle label shows "Activate/Deactivate" with current state clear
- **AC-44:** "Send login URL" toggle defaults to ON (send) when creating a new user
- **AC-45:** When Account status is toggled to Inactive, the "Send login URL" toggle is automatically disabled and forced to OFF (greyed out)
- **AC-46:** When Account status is toggled back to Active, the "Send login URL" toggle is re-enabled and defaults back to ON
- **AC-47:** If "Send login URL" is ON at save time, the system sends a first-login invitation email to the user's email address after successful creation
- **AC-48:** If "Send login URL" is OFF at save time, no invitation email is sent — the user account is created but the user is not notified

**Employee Profile:**
- **AC-13:** Department and Position dropdowns are side-by-side; options loaded from EP-008 US-001 and US-002 respectively; each includes a search field
- **AC-14:** Experience from (year) accepts a 4-digit year that is ≤ current year
- **AC-15:** CV/Resumé file input accepts PDF and DOCX files up to 5MB
- **AC-16:** After selecting a valid CV file, the filename is displayed next to the input
- **AC-17:** "+ Add skills" button opens a skills selector; selected skills display as removable tags/chips
- **AC-18:** Skills are loaded from EP-008 US-003 (Skill Management)

**Mandatory Field Validation (client-side):**
- **AC-19:** User cannot save without filling all 14 mandatory fields — inline error "[Field name] is required" shown for each empty field
- **AC-20:** Mandatory fields: First name, Last name, Email, Phone number, Date of birth, Gender, Nationality, User Role, Account status, Send login URL, Department, Position, Experience from (year), Education level
- **AC-21:** All mandatory fields are marked with asterisk (*)

**Email Validation:**
- **AC-22:** Client-side: invalid email format shows inline error "Please enter a valid email address"
- **AC-23:** Server-side: duplicate email shows both inline error "This email is already in use" AND error toast "This email is already in use"
- **AC-24:** Email uniqueness check is server-side on submit only — cannot be checked before submission

**File Upload Validation:**
- **AC-25:** Invalid avatar file type shows inline error "Only PNG, JPEG, JPG, and WEBP files are accepted"
- **AC-26:** Avatar file exceeding 5MB shows inline error "File size must not exceed 5MB"
- **AC-27:** Invalid CV file type shows inline error "Only PDF and DOCX files are accepted"
- **AC-28:** CV file exceeding 5MB shows inline error "File size must not exceed 5MB"

**Date & Year Validation:**
- **AC-29:** Invalid date of birth shows inline error "Please enter a valid date of birth"
- **AC-30:** Experience year that is not a valid 4-digit year shows inline error "Please enter a valid year"
- **AC-31:** Experience year in the future shows inline error "Year cannot be in the future"

**Save Behavior:**
- **AC-32:** Save creates the user, shows success toast "User '[First] [Last]' has been created", and redirects to User List
- **AC-33:** Newly created user appears in the User List with Active (or Inactive) status badge
- **AC-34:** Leading and trailing whitespace is trimmed from all text fields before saving
- **AC-35:** User can be saved without avatar, address, CV, or skills (optional fields)

**Cancel Behavior:**
- **AC-36:** Cancel on an untouched form redirects to User List without confirmation
- **AC-37:** Cancel on a modified form shows "Discard unsaved changes?" dialog — Confirm discards and redirects, Cancel stays on form
- **AC-38:** "Form modified" includes any change to any field, dropdown selection, file upload, toggle, or skill selection

**Dropdown Search:**
- **AC-39:** User Role, Department, and Position dropdowns include a search field at the top that filters options as the user types
- **AC-40:** Dropdown search is case-insensitive with partial match
- **AC-41:** When dropdown search returns no options, "No results found" is displayed within the dropdown

**Access Control:**
- **AC-42:** Create User page is accessible only to users with user management permission
- **AC-43:** Direct URL access by unauthorized users redirects to an appropriate fallback page

**Personal Information — New Fields:**
- **AC-49:** Marital status dropdown is optional and displays options: Single, Married, Other. No dropdown search.
- **AC-50:** Nationality dropdown is mandatory, defaults to "Vietnam", includes a search field, and accepts the full ISO country list.
- **AC-51:** Permanent address (renamed from "Address") and Temporary address are independent optional fields, each accepting up to 500 trimmed characters.

**Employee Profile — Education Level:**
- **AC-52:** Education level dropdown is mandatory and displays options: High school, College, Bachelor's degree, Master's degree, Doctorate. Inline error "Education level is required" if empty on save.

**ID Cards Section:**
- **AC-53:** ID Cards card displays as Row 3 left column, with Front image and Back image upload areas (each accepting PNG/JPEG/JPG/WEBP, max 5MB) and ID number + Issue date fields below.
- **AC-54:** All ID Cards fields are optional — the user can save without any ID Card data.
- **AC-55:** ID number is plain text (no format validation), trimmed, max 50 characters.
- **AC-56:** Issue date must be a valid past date when provided; inline error "Issue date cannot be in the future" otherwise.

**Emergency Contact Section:**
- **AC-57:** Emergency Contact card displays as Row 3 right column. Initial state shows only the "+ Add contact" button — no rows.
- **AC-58:** Clicking "+ Add contact" adds an empty row with Full name, Relationship, Phone number, and an X (remove) icon.
- **AC-59:** User can add unlimited contact rows; all fields per row are optional.
- **AC-60:** Clicking X on a contact row removes it immediately without confirmation.
- **AC-61:** Saving with zero emergency contacts is allowed.

**Salary Section (permission-gated):**
- **AC-62:** Salary card is hidden entirely from the layout if the current user lacks `user.salary.view`. The surviving Row 4 section spans full width.
- **AC-63:** Salary card is read-only (visible but disabled) if the user has `user.salary.view` but lacks `user.salary.manage`.
- **AC-64:** Both Base salary and Insurance salary are optional and accept non-negative numeric values; inline error "Salary cannot be negative" otherwise.
- **AC-65:** Salary input auto-formats with thousand separators (e.g., entering "15000000" displays as "15,000,000").
- **AC-66:** Currency is VND only (no currency selector).

**Banking Section (permission-gated):**
- **AC-67:** Banking card is hidden entirely from the layout if the current user lacks `user.banking.view`. The surviving Row 4 section spans full width.
- **AC-68:** Banking card is read-only (visible but disabled) if the user has `user.banking.view` but lacks `user.banking.manage`.
- **AC-69:** Bank dropdown is searchable and displays the hardcoded VN bank list (Vietcombank, BIDV, Techcombank, MB Bank, VPBank, Agribank, ACB, VIB, Sacombank, TPBank, SHB, HDBank, Eximbank, OCB, SeABank, MSB, LienVietPostBank, NCB, ABBank, NamABank, PVcomBank).
- **AC-70:** Bank account number accepts 6 to 20 numeric digits; inline error "Bank account number must be 6 to 20 digits" otherwise.
- **AC-71:** Account name is optional text, max 100 characters, trimmed.
- **AC-72:** Transfer method dropdown displays options: Bank transfer, Cash.
- **AC-73:** All Banking fields are optional — the user can save without any Banking data (when section is visible and editable).

**Layout Behavior:**
- **AC-74:** If both Salary and Banking sections are hidden (user has neither `view` permission), Row 4 is omitted entirely from the page.
- **AC-75:** Form-dirty detection includes all new fields (Marital status, Nationality, Permanent address, Temporary address, Education level, all ID Cards fields, all Emergency Contact rows, all Salary fields, all Banking fields). Any change triggers the discard confirmation on Cancel.

**Line Manager (Employee Profile):**
- **AC-76:** Line Manager field appears in the Employee Profile section as a full-width (576px) row positioned after Education level and before CV/Resumé.
- **AC-77:** Line Manager is optional and defaults to "No line manager" (top-of-list special option).
- **AC-78:** Line Manager dropdown displays each option in the format `Full Name — Position, Department` (e.g., `Sarah Le — CTO, Engineering`).
- **AC-79:** Line Manager dropdown is searchable; search is case-insensitive and matches across name, position, and department.
- **AC-80:** Line Manager dropdown options are loaded dynamically from the active user list; inactive users are excluded server-side and do not appear as options.
- **AC-81:** If the selected Line Manager was deactivated between dropdown load and form save, the server rejects the save and the form displays both inline error "Selected line manager is no longer active" below the Line Manager field AND an error toast.
- **AC-82:** User can be saved with no line manager (the "No line manager" option selected) — supports the top-of-hierarchy case (e.g., CEO, org founder).
- **AC-83:** Mandatory field count remains 14 — Line Manager is optional and does NOT change the mandatory field count.

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — all fields | Fill all 15 fields + avatar + CV + skills | User created, toast, redirect to list | High |
| Happy path — mandatory only | Fill 11 mandatory fields only | User created, toast, redirect | High |
| Empty mandatory field | Leave First name blank, click Save | Inline error "First name is required" | High |
| Multiple empty fields | Leave 3 mandatory fields blank | 3 inline errors shown simultaneously | High |
| Invalid email format | Enter "notanemail" | Inline error "Please enter a valid email address" | High |
| Duplicate email (server) | Enter existing email | Inline error + error toast "This email is already in use" | High |
| Invalid avatar type | Upload .gif file | Inline error about accepted formats | Medium |
| Avatar too large | Upload 10MB PNG | Inline error "File size must not exceed 5MB" | Medium |
| Valid avatar | Upload 2MB JPG | Preview thumbnail shown | Medium |
| Invalid CV type | Upload .txt file | Inline error about PDF/DOCX | Medium |
| CV too large | Upload 10MB PDF | Inline error "File size must not exceed 5MB" | Medium |
| Future experience year | Enter "2030" | Inline error "Year cannot be in the future" | Medium |
| Invalid year format | Enter "abcd" | Inline error "Please enter a valid year" | Medium |
| Add skills | Click "+ Add skills", select 3 skills | 3 skill chips displayed, removable | Medium |
| Dropdown search | Type "Eng" in Department dropdown | Only departments containing "Eng" shown | Medium |
| Dropdown search no match | Type "zzz" in Role dropdown | "No results found" in dropdown | Medium |
| Cancel dirty form | Fill some fields, click Cancel | "Discard unsaved changes?" dialog | Medium |
| Cancel clean form | No changes, click Cancel | Redirect to User List | Low |
| Account status default | Page loads | Toggle is ON (Active) by default | Medium |
| Send login URL default | Page loads | Send login URL toggle ON by default | Medium |
| Send login URL with Active | Save with Account Active + Send login URL ON | User created, invitation email sent | High |
| Send login URL OFF | Save with Account Active + Send login URL OFF | User created, no email sent | Medium |
| Send login URL disabled | Toggle Account status to Inactive | Send login URL disabled + forced OFF | High |
| Send login URL re-enabled | Toggle Account status back to Active | Send login URL re-enabled, defaults ON | Medium |
| Dropdown API fails | Role API unavailable | Error in Role dropdown with retry | Medium |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |
| Nationality default | Page loads | Nationality dropdown pre-filled with "Vietnam" | High |
| Nationality empty save | Clear Nationality, click Save | Inline error "Nationality is required" | High |
| Education level mandatory | Leave Education level empty, click Save | Inline error "Education level is required" | High |
| Marital status optional | Save without selecting marital status | User created successfully | Medium |
| Permanent + Temporary address | Fill both addresses | Both saved independently, each trimmed | Medium |
| ID Card optional skip | Save without any ID Card data | User created successfully | High |
| ID Card front image upload | Upload 3MB PNG to Front image | Preview thumbnail shown | Medium |
| ID Card image too large | Upload 7MB JPG | Inline error "File size must not exceed 5MB" | Medium |
| ID Card issue date future | Enter future date | Inline error "Issue date cannot be in the future" | Medium |
| Emergency contact add | Click "+ Add contact" 3 times | 3 empty rows appear with X icons | Medium |
| Emergency contact remove | Click X on middle of 3 rows | That row removed immediately, 2 remain | Medium |
| Emergency contact save empty | Save with zero contacts | User created successfully | High |
| Salary thousand separator | Type "15000000" in Base salary | Auto-formats to "15,000,000" | High |
| Salary negative | Enter "-1000" in Base salary | Inline error "Salary cannot be negative" | Medium |
| Salary hidden (no view) | User without `user.salary.view` opens form | Salary card not rendered; Banking spans full width if visible | High |
| Salary read-only | User with `view` but no `manage` | Salary fields visible but disabled | High |
| Banking hidden (no view) | User without `user.banking.view` opens form | Banking card not rendered; Salary spans full width if visible | High |
| Banking read-only | User with `view` but no `manage` | Banking fields visible but disabled | High |
| Both Row 4 hidden | User lacks both `salary.view` and `banking.view` | Row 4 omitted entirely | High |
| Bank dropdown search | Type "Vietco" in Bank dropdown | Only "Vietcombank" appears | Medium |
| Bank account number invalid | Enter "abc123" or "12345" (5 digits) | Inline error "Bank account number must be 6 to 20 digits" | Medium |
| Bank account number valid | Enter "12345678901234" (14 digits) | Accepted, no error | Medium |
| Transfer method Cash | Select Cash | Bank/account fields can remain empty; user saved | Medium |
| Form dirty new field | Modify only Temporary address, click Cancel | "Discard unsaved changes?" dialog appears | Medium |
| Line Manager — happy path with selection | Select a line manager from dropdown, save | User created with line manager assigned | High |
| Line Manager — happy path no manager | Leave as "No line manager", save | User created with no line manager (top-of-hierarchy case) | High |
| Line Manager — search by department | Type a department name in Line Manager dropdown search | Only users in that department appear | Medium |
| Line Manager — search by position | Type a position name in Line Manager dropdown search | Only users with that position appear | Medium |
| Line Manager — selected user becomes inactive | Line manager deactivated between load and save | Server error + inline error "Selected line manager is no longer active" + error toast | High |
| Line Manager — dropdown excludes inactive | Open Line Manager dropdown | Only active users listed; inactive users absent | High |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Only users with user management permission can access the Create User page
- **SR-02:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-03:** Email must be unique organization-wide — checked **server-side on form submission only**. If duplicate, both inline error and error toast displayed.
- **SR-04:** All text fields (first name, last name, email, phone, permanent address, temporary address, ID number, emergency contact full name/relationship/phone, banking account name) are trimmed of leading/trailing whitespace before saving
- **SR-05:** Account status defaults to Active (toggle ON) when creating a new user. Administrator can switch to Inactive before saving if needed.
- **SR-06:** A newly created user with Active status can log in immediately — no additional activation step required
- **SR-07:** A newly created user with Inactive status cannot log in until activated by an administrator
- **SR-08:** User Role dropdown options are loaded dynamically from US-004 — no hardcoded role list
- **SR-09:** Department dropdown options are loaded dynamically from EP-008 US-001 (Department Management)
- **SR-10:** Position dropdown options are loaded dynamically from EP-008 US-002 (Position Management)
- **SR-11:** Skills options are loaded dynamically from EP-008 US-003 (Skill Management)
- **SR-12:** Avatar accepted formats: PNG, JPEG, JPG, WEBP. Max file size: 5MB. Validated client-side on file selection.
- **SR-13:** CV/Resumé accepted formats: PDF, DOCX. Max file size: 5MB. Validated client-side on file selection.
- **SR-14:** Experience from (year) must be a valid 4-digit year and cannot be in the future (≤ current year)
- **SR-15:** Gender dropdown options: pending PO confirmation. Expected: Male, Female, Other, Prefer not to say.
- **SR-16:** Date of birth must be a valid date in the past — future dates are not allowed
- **SR-17:** If "Send login URL" is ON at save time, the system sends a first-login invitation email containing a unique login URL to the user's email address. The email is sent asynchronously after the user record is created — form save does not wait for email delivery.
- **SR-18:** If "Send login URL" is OFF at save time, the user account is created but no invitation email is sent. The administrator must manually share login information with the user.
- **SR-19:** Dropdown search fields (User Role, Department, Position) filter options client-side — no server round-trip for dropdown filtering. Consistent with User List filter dropdown pattern.
- **SR-20:** When Account status is set to Inactive, the "Send login URL" toggle is automatically disabled and forced to OFF — since an inactive user cannot log in, sending a login URL would be misleading. When Account status is toggled back to Active, the "Send login URL" toggle is re-enabled and defaults back to ON.
- **SR-21:** "Form dirty" detection compares all fields against their initial values — including all new fields (Marital status, Nationality, Permanent address, Temporary address, Education level, all ID Cards fields, all Emergency Contact rows, all Salary fields, all Banking fields). Any change triggers the discard confirmation on Cancel.
- **SR-22:** "Address" field is renamed to "Permanent address" everywhere in the form, validation messages, and saved data. A new "Temporary address" field is added with the same validation rules (optional, trimmed, max 500 chars).
- **SR-23:** Nationality is a mandatory field, defaults to "Vietnam", and is sourced from the full ISO country list. The dropdown includes a search field.
- **SR-24:** Marital status is an optional field with options: Single, Married, Other. No dropdown search (only 3 options).
- **SR-25:** Education level is a mandatory field with options: High school, College, Bachelor's degree, Master's degree, Doctorate. No dropdown search (only 5 options).
- **SR-26:** ID Cards section is entirely optional — all four fields (Front image, Back image, ID number, Issue date) can be left empty. ID number is plain text with no format validation (max 50 chars). Issue date must be a valid past date when provided.
- **SR-27:** ID Card front and back images accept PNG, JPEG, JPG, WEBP. Max file size: 5MB each. Validated client-side on file selection.
- **SR-28:** Emergency Contact section is entirely optional (0 contacts allowed) and supports unlimited rows. Each row has Full name, Relationship, Phone number — all optional. The X icon removes a row immediately without confirmation.
- **SR-29:** Salary section is visibility-gated by `user.salary.view`. Without this permission, the entire Salary card is omitted from the page (not greyed out). With `view` but without `user.salary.manage`, the fields are visible but disabled (read-only).
- **SR-30:** Salary fields accept non-negative numeric values in VND, auto-formatted with thousand separators (e.g., "15,000,000"). Both Base salary and Insurance salary are optional.
- **SR-31:** Banking section is visibility-gated by `user.banking.view`. Without this permission, the entire Banking card is omitted from the page. With `view` but without `user.banking.manage`, the fields are visible but disabled (read-only).
- **SR-32:** Banking bank list is hardcoded (Vietnamese banks): Vietcombank, BIDV, Techcombank, MB Bank, VPBank, Agribank, ACB, VIB, Sacombank, TPBank, SHB, HDBank, Eximbank, OCB, SeABank, MSB, LienVietPostBank, NCB, ABBank, NamABank, PVcomBank.
- **SR-33:** Banking bank account number must be 6 to 20 numeric digits when provided.
- **SR-34:** When a Row 4 permission-gated section is hidden, the surviving section spans full width. When both are hidden, Row 4 is omitted entirely from the page.
- **SR-35:** Permission-gating is enforced both in the UI (hide / disable) and on the server (forbid). UI hiding alone is not security — the API must reject any payload that contains salary or banking data from a user without the corresponding `manage` permission.
- **SR-36:** Line Manager is OPTIONAL. Top-of-hierarchy users (e.g., CEO, org founder) may legitimately have no line manager — the "No line manager" option supports this case.
- **SR-37:** Line Manager candidate pool is any ACTIVE user — no role restriction, no department restriction. Cross-department assignment is allowed (e.g., an Engineering user can report to an Operations manager).
- **SR-38:** Line Manager is stored as a reference (user ID), NOT a snapshot. Display values (full name, position, department) resolve at query time, so they remain accurate as the referenced user's profile changes.
- **SR-39:** Line Manager dropdown options are loaded dynamically from the active user list. Inactive users are excluded server-side and do not appear in the dropdown. The user being created is not applicable on Create (the user does not exist yet, so self-assignment and cycle validation are n/a on Create).
- **SR-40:** Line Manager option display format: `Full Name — Position, Department` (e.g., `Sarah Le — CTO, Engineering`). Dropdown search is case-insensitive and matches across name, position, OR department.
- **SR-41:** Server-side, on save, the system re-validates that the selected Line Manager is still active. If not (the user was deactivated between dropdown load and save submission), the save is rejected with both inline error "Selected line manager is no longer active" below the Line Manager field and an error toast.
- **SR-42:** Line Manager interacts with downstream permissions via the `.team` / `.all` scope split documented in the design spec (`docs/superpowers/specs/2026-05-20-user-management-line-manager-design.md`). When downstream stories add `<action>.team` permissions (e.g., `leave.approve.team`), holders can only act on requests from users in their subordinate chain (resolved via Line Manager). Holders of `<action>.all` can act on anyone. This DR does NOT introduce new permissions itself.
- **SR-43:** Edit access for the Line Manager field is gated by the existing `user.edit` permission — no new permission is introduced for Line Manager management.

**State Transitions:**
```
[User List] → "+ Add New" click → [Create User Form (empty, Account status: Active)]
[Create User Form] → Save (valid, server OK) → [Toast] → [User List + new user visible]
[Create User Form] → Save (valid, server: duplicate email) → [Inline error + Error toast]
[Create User Form] → Save (invalid, client-side) → [Inline error(s) below affected fields]
[Create User Form] → Cancel (clean) → [User List]
[Create User Form] → Cancel (dirty) → [Confirmation Dialog]
[Confirmation Dialog] → Confirm discard → [User List]
[Confirmation Dialog] → Cancel → [Create User Form]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in; handles login credentials for new user
- **Depends on:** US-004 (Role & Permission Management) — provides role dropdown options; controls access to this page
- **Depends on:** EP-008 US-001 (Department Management) — provides department dropdown options
- **Depends on:** EP-008 US-002 (Position Management) — provides position dropdown options
- **Depends on:** EP-008 US-003 (Skill Management) — provides skills selector options
- **Depends on:** US-004 (Role & Permission Management) — provides the new `user.salary.view`, `user.salary.manage`, `user.banking.view`, `user.banking.manage` permissions used to gate the Salary and Banking sections (catalog update handled separately in US-004)
- **Depends on:** Existing user list API (DR-001-005-01) — provides the active user list used to populate the Line Manager dropdown (with name, position, department for display formatting; inactive users filtered server-side)
- **Consumed by:** User List (DR-001-005-01) — new user appears in list after creation
- **Consumed by:** User Details (DR-001-005-03) — displays all new fields and sections read-only with the same permission gating
- **Consumed by:** Update User Information (DR-001-005-04) — mirrors this extended structure for editing
- **Consumed by:** All modules — new user's role-permission data used for access control

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** First name field is auto-focused on page load — user can begin typing immediately
- **UX-02:** Save button shows loading spinner while request is in progress; Cancel button also disabled — prevents double submission
- **UX-03:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-04:** Error toast (duplicate email from server) persists until user dismisses manually — does not auto-dismiss
- **UX-05:** Avatar upload shows preview thumbnail immediately after file selection — user can confirm correct image before saving
- **UX-06:** CV/Resumé file input shows filename after selection — user can confirm correct file
- **UX-07:** Selected skills display as removable tag chips — user can click X on each chip to remove a skill without reopening the selector
- **UX-08:** "Discard unsaved changes?" dialog uses default focus on Cancel (stay on form) — prevents accidental data loss
- **UX-09:** Inline validation errors appear directly below the relevant field — not as popups. Multiple errors shown simultaneously for all invalid fields.
- **UX-10:** User Role, Department, and Position dropdowns include a search field for quick option finding — critical for organizations with many departments/positions. Gender dropdown does not include search (only 3-4 options).
- **UX-11:** Date of birth field uses a native date picker (calendar popup) — no manual date format required
- **UX-12:** Two-column layout groups related fields logically: personal data (left), system/org assignments (right) — reduces cognitive load
- **UX-13:** Account status toggle provides immediate visual feedback (on = Active green, off = Inactive gray) — consistent with status badge colors on User List
- **UX-14:** Tab order follows visual layout: left column top-to-bottom → right column top-to-bottom → Save → Cancel

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Two-column layout: left card (600px) + right cards (600px) |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields in logical order, Enter/Space to open dropdowns and file pickers
- [x] Screen reader compatible — labels associated with all inputs, error messages announced, toggle state announced
- [x] Date picker accessible via keyboard — arrow keys to navigate, Enter to select
- [x] Dropdown search accessible via keyboard — type to filter, arrow keys to navigate options
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible on all interactive elements
- [x] Form errors linked to fields via aria-describedby
- [x] Toggle switch state communicated via aria-checked

**Design References:**
- Figma: [Create A New User](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3050-3782) (node `3050:3782`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Create Role (DR-001-004-02) for action bar; unique two-column layout for User

---

## 8. Additional Information

### Out of Scope
- Edit User — separate detail requirement (DR-001-005-03, planned)
- Deactivate/Activate User — separate detail requirement (DR-001-005-08)
- User self-registration (admin-only creation in this module)
- Invitation email template customization (email content/design is system-defined)
- Bulk user creation or CSV import
- Avatar cropping or image editing within the upload flow
- Employee HR profile detail page (covered by EP-002 Employee Management)
- Assigning multiple roles to a single user (one role per user)
- Department-Position dependency (filtering positions based on selected department — all positions shown regardless; pending PO confirmation)
- User duplication / cloning
- Permission catalog update for new permissions (`user.salary.view`, `user.salary.manage`, `user.banking.view`, `user.banking.manage`) — handled separately in US-004 (Role & Permission Management)
- Mobile/responsive layout for the new sections (Row 3 and Row 4)
- Bank list management UI — hardcoded for now per PO decision
- Insurance salary calculation logic — admin manual input per existing business rule
- Audit logging for Salary and Banking field changes — separate logging story
- ID Card image storage and retention policy — handled by file storage service
- Bulk line manager assignment / reassignment tool (e.g., reassigning all subordinates of a departing manager) — separate story
- Org-chart visualization (graphical reporting hierarchy view) — separate story
- Acting manager / temporary delegation (assigning a stand-in manager while the primary is unavailable) — out of scope for this release
- Multiple / matrix line managers (a user reporting to more than one manager) — single primary line manager only

### Open Questions
- [ ] **Gender dropdown options:** Male, Female, Other, Prefer not to say assumed — confirm exact values. — **Owner:** Product Owner — **Status:** Pending
- [x] **Login credentials mechanism:** Resolved — "Send login URL" toggle controls whether first-login invitation email is sent. Default ON for Active users, disabled for Inactive users. — **Owner:** Product Owner — **Status:** Resolved ✅
- [ ] **Phone number format:** Any format restrictions (country code, digits only, min/max digits)? Or free text? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Skills selector UX:** What opens when user clicks "+ Add skills"? Modal with search? Dropdown multi-select? — **Owner:** Design Team — **Status:** Pending
- [ ] **Department-Position dependency:** Should selecting a department filter the Position dropdown to only positions in that department? Or are they independent? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Avatar dimensions:** Should uploaded images be auto-resized/cropped to a standard size (e.g., 200×200px)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Edit User screen:** When will Design Team deliver the Edit User Figma screen? — **Owner:** Design Team — **Status:** Pending
- [ ] **Bank list:** Confirm exact set of Vietnamese banks supported (current draft list of 21 banks needs PO sign-off). — **Owner:** Product Owner — **Status:** Pending
- [ ] **Nationality dropdown source:** Should it use the full ISO 3166 country list or a curated subset relevant to the company's workforce? — **Owner:** Product Owner — **Status:** Pending
- [ ] **ID Card permission gating:** Are stored ID card images sensitive enough to require a dedicated `user.idcards.view` permission similar to Salary/Banking? Currently they are gated only by base `user.view`. — **Owner:** Product Owner / Security — **Status:** Pending
- [ ] **Salary currency:** VND only, or is multi-currency support needed for any cross-border employee scenarios? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Emergency Contact phone format:** Should phone numbers be validated against Vietnam mobile format or remain free text? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Banking account number validation:** Beyond 6-20 digits, should the system run a Luhn check or other structural validation? — **Owner:** Product Owner / Finance — **Status:** Pending
- [ ] **First-user Line Manager default:** Should "No line manager" be the default for the very first user created (org founder), or should the form force a selection? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Cross-department warning:** Should the form display a warning when selecting a Line Manager from a different department, in case the assignment is unintentional? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-01: User List | Entry point — Create User accessed via "+ Add New"; new user appears in list |
| DR-001-005-03: Edit User (planned) | Expected to mirror Create User with pre-filled data |
| DR-001-005-08: Activate/Deactivate | Status toggle behavior defined here; deactivation flow is separate DR |
| US-001: Authentication | Handles login credentials and invitation flow for newly created users |
| US-004: Role & Permission Management | Provides User Role dropdown options; controls access to this page |
| EP-008 US-001: Department Management | Provides Department dropdown options |
| EP-008 US-002: Position Management | Provides Position dropdown options |
| EP-008 US-003: Skill Management | Provides Skills selector options |
| DR-001-005-03: User Details | Read-only display of all fields and sections created here, with the same permission gating for Salary and Banking |
| DR-001-005-04: Update User Information | Edit form mirrors this Create form's extended structure (ID Cards, Emergency Contact, Salary, Banking) |
| Design Spec: User Management Extended Fields (2026-05-18) | Authoritative spec for the 4 new sections + Personal/Work Profile field extensions across all 3 User DRs |

### Notes
- This is the **most complex create form** in the entire HRM platform — 16 fields across 3 card sections in a two-column layout. All other create forms have 1–3 fields in a single centered card.
- The form has **5 different input types**: text input, dropdown (with and without search), date picker, file upload, and toggle switch (×2: Account status + Send login URL), plus the skills multi-select. This is the highest variety of input types in any single form.
- **Email uniqueness** is the only server-side validation — all other validations (mandatory fields, formats, file types/sizes) are client-side. When server returns duplicate email, both inline error AND error toast are shown (consistent with Skill name duplicate pattern from DR-008-003-02).
- **Account status defaults to Active** — this is intentional so that the most common flow (create an active user) requires no extra action. Administrator can toggle to Inactive before saving for pre-provisioning scenarios.
- The **Figma design shows 2 buttons** (Cancel + Save) — no "Save & Create Another" for user creation since each user requires unique data (unlike skills or departments which may be batch-created).
- The **Department-Position dependency** is flagged as an open question. If confirmed, selecting a department would filter the Position dropdown. Currently, all positions show regardless of department selection.
- **Dropdown search** is included in User Role, Department, and Position dropdowns (dynamic data with potentially many options) but NOT in Gender dropdown (only 3-4 static options).
- The **"Send login URL" toggle** is a conditional field — it is disabled and forced OFF when Account status is Inactive, because sending a login URL to a user who cannot log in would be misleading. This is the first field in the platform with a dependency on another field's value.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-24 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-24 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context; most complex create form in platform |
| 1.1 | 2026-03-24 | BA Agent | Added "Send login URL" toggle field (AC-44 through AC-48, SR-17/18/20); resolved login credentials open question |
| 1.2 | 2026-05-18 | BA Agent | Added ID Cards, Emergency Contact, Salary (permission-gated), Banking (permission-gated) sections; added Nationality (mandatory) + Marital status + Temporary address + Education level (mandatory) fields; renamed Address → Permanent address |
| 1.3 | 2026-05-20 | BA Agent | Added optional Line Manager field to Employee Profile (searchable user dropdown showing "Name — Position, Department"); excludes inactive users; field is optional for top-of-hierarchy users; documented two-variant `.team`/`.all` permission pattern for downstream approval stories (see design spec 2026-05-20-user-management-line-manager-design.md); no new permissions added in this DR |
