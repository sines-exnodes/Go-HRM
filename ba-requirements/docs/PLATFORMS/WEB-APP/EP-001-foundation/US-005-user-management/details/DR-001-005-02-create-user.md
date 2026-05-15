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
version: "1.1"
created_date: "2026-03-24"
last_updated: "2026-03-24"
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
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **create a new user account by entering personal information, assigning a role, and linking to a department/position**, so that **the new user can access the HRM system with appropriate permissions and their employee profile is initialized**.

**Purpose:** Enable administrators to onboard new users into the HRM platform. Creating a user sets up their authentication credentials, assigns a system role (which governs their access), and links them to the organization structure (department + position). Once created, the user can immediately log in and access the system per their role.

**Target Users:** Any user with user management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- 3-section form: Personal Information (7 fields + avatar), User Account (role + status toggle + send login URL toggle), Employee Profile (dept/position/experience + CV + skills)
- 12 mandatory fields, 4 optional fields
- Two-column layout — Personal Information (left), User Account + Employee Profile (right)
- Email uniqueness validation (server-side on submit — both inline error and error toast)
- Dropdown selectors with built-in search for User Role, Department, Position
- File uploads for Avatar (.png/.jpeg/.jpg/.webp, max 5MB) and CV/Resumé (PDF/DOCX, max 5MB)
- Skills multi-select via "+ Add skills" button
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
3. System loads dropdown data from APIs: roles (US-004), departments (EP-008 US-001), positions (EP-008 US-002), gender options, skills (EP-008 US-003)
4. Form displays with 3 empty cards: Personal Information (left), User Account (right top — Account status toggle ON by default), Employee Profile (right bottom)
5. User fills **Personal Information**: enters first name, last name, email, phone number, date of birth, selects gender; optionally uploads avatar and enters address
6. User fills **User Account**: selects a user role from dropdown (with search), sets account status toggle (default: Active), confirms "Send login URL" toggle (default: ON; disabled if Account status is Inactive)
7. User fills **Employee Profile**: selects department and position from dropdowns (with search), enters experience year; optionally uploads CV/Resumé and adds skills
8. User clicks **Save**
9. System trims whitespace from all text fields
10. System validates client-side: all mandatory fields filled, email format valid, date of birth is valid past date, experience year is valid and ≤ current year, file types/sizes valid
11. System submits to server; server validates email uniqueness
12. If all validations pass: system creates the user account
13. If "Send login URL" toggle is ON: system sends a first-login invitation email to the user's email address
14. System displays success toast: "User '[first name] [last name]' has been created"
15. System redirects to User List — new user visible in list with Active (or Inactive) status badge

**Alternative Flows:**

- **Alt 1 — Empty mandatory field:** Inline error below the empty field: "[Field name] is required". Form is not submitted. User corrects and retries.
- **Alt 2 — Invalid email format:** Inline error: "Please enter a valid email address". Form is not submitted.
- **Alt 3 — Duplicate email (server-side):** Server returns duplicate error. Both inline error "This email is already in use" AND error toast "This email is already in use" displayed. User enters a different email.
- **Alt 4 — Invalid avatar file:** Inline error below Upload Avatar: "Only PNG, JPEG, JPG, and WEBP files are accepted" or "File size must not exceed 5MB".
- **Alt 5 — Invalid CV file:** Inline error below CV/Resumé: "Only PDF and DOCX files are accepted" or "File size must not exceed 5MB".
- **Alt 6 — Cancel (form modified):** System shows "Discard unsaved changes?" dialog. Confirm → redirect to User List. Cancel → stay on form.
- **Alt 7 — Cancel (form untouched):** System redirects to User List immediately without confirmation.
- **Alt 8 — Dropdown API fails:** Error message in affected section with retry option. Other sections remain functional.

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
| Address | Text input | Trimmed; no other validation | No | Empty | 500 chars | Placeholder: "Enter address" (576px, full width) |

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
| CV/Resumé | File input | PDF, DOCX; max 5MB; validated client-side on file selection | No | No file chosen | "Choose File / No file chosen" (576px) |
| Skills | Multi-select | No minimum required | No | None selected | "+ Add skills" button opens selector. Skills loaded from EP-008 US-003. |

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Left in action bar | Always visible | If dirty → "Discard changes?" dialog; if clean → redirect | Discard and return to User List |
| Save | Button (primary) | Right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → redirect | Save and return to User List |
| Upload Avatar | Button | Inside avatar upload area | Always visible | Opens file picker for image | Upload profile photo |
| + Add skills | Button (secondary) | Inside Employee Profile card | Always visible | Opens skills selector | Add skills to user profile |

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
| Address | Text input | Placeholder: "Enter address" | 576px, full width | Address (optional) |

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
| CV/Resumé | File input | "Choose File / No file chosen" | 576px | Resume document |
| Skills | Multi-select | "+ Add skills" button only | Button → selector → chips | Competency tags |

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

### Page Layout (from Figma)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Users Management / Users / Create A New User                  │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A New User                       [Cancel]    [Save] │
│              │                                                             │
│              │  ┌─────────────────────────┐  ┌─────────────────────────┐   │
│              │  │ Personal Information    │  │ User Account            │   │
│              │  │                         │  │                         │   │
│              │  │  ┌───────────────────┐  │  │ * User Role    [▾ 🔍]  │   │
│              │  │  │  Upload Avatar    │  │  │ * Account status  [⊙]  │   │
│              │  │  │  .png/.jpeg/...   │  │  │   Activate/Deactivate  │   │
│              │  │  └───────────────────┘  │  └─────────────────────────┘   │
│              │  │                         │                                │
│              │  │ *First name  *Last name │  ┌─────────────────────────┐   │
│              │  │ *Email     *Phone number│  │ Employee Profile        │   │
│              │  │ *DOB [📅]  *Gender [▾] │  │                         │   │
│              │  │  Address                │  │ *Dept [▾ 🔍] *Pos [▾ 🔍]│   │
│              │  └─────────────────────────┘  │ *Experience from (year) │   │
│              │                               │  CV/Resumé [Choose File]│   │
│              │                               │  Skills [+ Add skills]  │   │
│              │                               └─────────────────────────┘   │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Create User page displays with page title "Create A New User" in Geist Semibold 24px
- **AC-02:** Form uses a two-column layout: Personal Information (left), User Account + Employee Profile (right)
- **AC-03:** Two buttons visible: Cancel and Save (no "Save & Create Another")

**Personal Information:**
- **AC-04:** Avatar upload area accepts PNG, JPEG, JPG, WEBP files up to 5MB
- **AC-05:** After selecting a valid avatar, a preview thumbnail replaces the upload area
- **AC-06:** First name and Last name are displayed side-by-side (278px each)
- **AC-07:** Email and Phone number are displayed side-by-side (278px each)
- **AC-08:** Date of birth shows a date picker (calendar icon) and Gender shows a dropdown — side-by-side
- **AC-09:** Address is full-width (576px) and optional (no asterisk)

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
- **AC-19:** User cannot save without filling all 12 mandatory fields — inline error "[Field name] is required" shown for each empty field
- **AC-20:** Mandatory fields: First name, Last name, Email, Phone number, Date of birth, Gender, User Role, Account status, Send login URL, Department, Position, Experience from (year)
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

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Only users with user management permission can access the Create User page
- **SR-02:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-03:** Email must be unique organization-wide — checked **server-side on form submission only**. If duplicate, both inline error and error toast displayed.
- **SR-04:** All text fields (first name, last name, email, phone, address) are trimmed of leading/trailing whitespace before saving
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
- **SR-21:** "Form dirty" detection compares all 16 fields against their initial values. Any change triggers the discard confirmation on Cancel.

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
- **Consumed by:** User List (DR-001-005-01) — new user appears in list after creation
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

### Open Questions
- [ ] **Gender dropdown options:** Male, Female, Other, Prefer not to say assumed — confirm exact values. — **Owner:** Product Owner — **Status:** Pending
- [x] **Login credentials mechanism:** Resolved — "Send login URL" toggle controls whether first-login invitation email is sent. Default ON for Active users, disabled for Inactive users. — **Owner:** Product Owner — **Status:** Resolved ✅
- [ ] **Phone number format:** Any format restrictions (country code, digits only, min/max digits)? Or free text? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Skills selector UX:** What opens when user clicks "+ Add skills"? Modal with search? Dropdown multi-select? — **Owner:** Design Team — **Status:** Pending
- [ ] **Department-Position dependency:** Should selecting a department filter the Position dropdown to only positions in that department? Or are they independent? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Avatar dimensions:** Should uploaded images be auto-resized/cropped to a standard size (e.g., 200×200px)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Edit User screen:** When will Design Team deliver the Edit User Figma screen? — **Owner:** Design Team — **Status:** Pending

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
